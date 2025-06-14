package app

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"maps"

	"github.com/gobwas/glob"
	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	extr_animation "github.com/xypwn/filediver/extractor/animation"
	extr_bik "github.com/xypwn/filediver/extractor/bik"
	extr_bones "github.com/xypwn/filediver/extractor/bones"
	extr_geogroup "github.com/xypwn/filediver/extractor/geometry_group"
	extr_material "github.com/xypwn/filediver/extractor/material"
	extr_package "github.com/xypwn/filediver/extractor/package"
	extr_state_machine "github.com/xypwn/filediver/extractor/state_machine"
	extr_strings "github.com/xypwn/filediver/extractor/strings"
	extr_texture "github.com/xypwn/filediver/extractor/texture"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	extr_wwise "github.com/xypwn/filediver/extractor/wwise"
	"github.com/xypwn/filediver/steampath"
	"github.com/xypwn/filediver/stingray"
	dlbin "github.com/xypwn/filediver/stingray/dl_bin"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
	stingray_wwise "github.com/xypwn/filediver/stingray/wwise"
	"github.com/xypwn/filediver/wwise"
)

func parseWwiseDep(ctx context.Context, f *stingray.File) (string, error) {
	r, err := f.Open(ctx, stingray.DataMain)
	if err != nil {
		return "", err
	}
	var magicNum [4]byte
	if _, err := io.ReadFull(r, magicNum[:]); err != nil {
		return "", err
	}
	if magicNum != [4]byte{0xd8, '/', 'v', 'x'} {
		return "", errors.New("invalid magic number")
	}
	var textLen uint32
	if err := binary.Read(r, binary.LittleEndian, &textLen); err != nil {
		return "", err
	}
	text := make([]byte, textLen-1)
	if _, err := io.ReadFull(r, text); err != nil {
		return "", err
	}
	return string(text), nil
}

// Returns error if steam path couldn't be found.
func DetectGameDir() (string, error) {
	return steampath.GetAppPath("553850", "Helldivers 2")
}

func VerifyGameDir(path string) error {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return fmt.Errorf("invalid game directory: %v: not a directory", path)
	}
	if info, err := os.Stat(filepath.Join(path, "settings.ini")); err == nil && info.Mode().IsRegular() {
		// We were given the "data" directory => go back
		path = filepath.Dir(path)
	}
	if info, err := os.Stat(filepath.Join(path, "data", "settings.ini")); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("invalid game directory: %v: valid data directory not found", path)
	}
	return nil
}

func ParseHashes(str string) []string {
	var res []string
	sc := bufio.NewScanner(strings.NewReader(str))
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s != "" && !strings.HasPrefix(s, "//") {
			res = append(res, s)
		}
	}
	return res
}

type App struct {
	Hashes     map[stingray.Hash]string
	ThinHashes map[stingray.ThinHash]string
	ArmorSets  map[stingray.Hash]dlbin.ArmorSet
	DataDir    *stingray.DataDir
}

// Automatically gets most wwise-related hashes by reading the game files
func getWwiseHashes(ctx context.Context, dataDir *stingray.DataDir) (map[stingray.Hash]string, error) {
	hashes := make(map[stingray.Hash]string)
	// Read wwise_dep files to figure out the names of all wwise_bank files
	for id, file := range dataDir.Files {
		if id.Type == stingray.Sum64([]byte("wwise_dep")) {
			h, err := parseWwiseDep(ctx, file)
			if err != nil {
				return nil, fmt.Errorf("wwise_dep: %w", err)
			}
			hashes[stingray.Sum64([]byte(h))] = h
		}
	}
	// Read wwise_bank files to figure out the names of most wwise_stream files
	for id, file := range dataDir.Files {
		if id.Type == stingray.Sum64([]byte("wwise_bank")) {
			name, ok := hashes[id.Name]
			if !ok {
				// It seems the wwise banks no longer all have an according wwise_dep (https://github.com/xypwn/filediver/issues/35).
				// Hopefully these banks missing won't become a problem.
				//return nil, fmt.Errorf("expected all wwise banks to have a known name, but cannot find name for hash %v", id.Name)
				continue
			}
			dir := path.Dir(name)
			if err := func() error {
				r, err := file.Open(ctx, stingray.DataMain)
				if err != nil {
					return err
				}
				defer r.Close()
				bnk, err := stingray_wwise.OpenBnk(r)
				if err != nil {
					return err
				}
				for i := 0; i < bnk.NumFiles(); i++ {
					id := bnk.FileID(i)
					streamPath := path.Join(dir, fmt.Sprint(id))
					hashes[stingray.Sum64([]byte(streamPath))] = streamPath
				}
				for _, obj := range bnk.HircObjects {
					if obj.Header.Type == wwise.BnkHircObjectSound {
						streamPath := path.Join(dir, fmt.Sprint(obj.Sound.SourceID))
						hashes[stingray.Sum64([]byte(streamPath))] = streamPath
					}
				}
				return nil
			}(); err != nil {
				return nil, err
			}
		}
	}
	// 6 hashes were still missing when tested
	/*// Validate that all wwise_stream file names are known
	{
		numWithoutName := 0
		for id := range dataDir.Files {
			if id.Type == stingray.Sum64([]byte("wwise_stream")) {
				if _, ok := hashes[id.Name]; !ok {
					fmt.Println(id.Name)
					numWithoutName++
				}
			}
		}
		if numWithoutName > 0 {
			return nil, fmt.Errorf("expected all wwise streams to have a known name, but cannot find name for %v hashes", numWithoutName)
		}
	}*/

	return hashes, nil
}

// Open game dir and read metadata.
func OpenGameDir(ctx context.Context, gameDir string, hashStrings []string, thinhashes []string, armorStrings stingray.Hash, onProgress func(curr, total int)) (*App, error) {
	dataDir, err := stingray.OpenDataDir(ctx, filepath.Join(gameDir, "data"), onProgress)
	if err != nil {
		return nil, err
	}

	hashesMap := make(map[stingray.Hash]string)
	if wwiseHashes, err := getWwiseHashes(ctx, dataDir); err == nil {
		for h, n := range wwiseHashes {
			hashesMap[h] = n
		}
	} else {
		return nil, err
	}
	for _, h := range hashStrings {
		hashesMap[stingray.Sum64([]byte(h))] = h
	}
	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range thinhashes {
		thinHashesMap[stingray.Sum64([]byte(h)).Thin()] = h
	}

	stringsFile, ok := dataDir.Files[stingray.FileID{
		Name: armorStrings,
		Type: stingray.Sum64([]byte("strings")),
	}]

	var stringMap *stingray_strings.StingrayStrings = nil
	if ok {
		stringsReader, err := stringsFile.Open(ctx, stingray.DataMain)
		if err == nil {
			defer stringsReader.Close()
			stringMap, err = stingray_strings.LoadStingrayStrings(stringsReader)
			if err != nil {
				stringMap = nil
			}
		}
	}

	var mapping map[uint32]string = make(map[uint32]string)
	if stringMap != nil {
		mapping = stringMap.Strings
	}

	armorSets, err := dlbin.LoadArmorSetDefinitions(mapping)
	if err != nil {
		return nil, err
	}

	return &App{
		Hashes:     hashesMap,
		ThinHashes: thinHashesMap,
		ArmorSets:  armorSets,
		DataDir:    dataDir,
	}, nil
}

func (a *App) hashNameVariationsForMatch(h stingray.Hash) []string {
	res := []string{
		h.StringEndian(binary.LittleEndian),
		h.StringEndian(binary.BigEndian),
		"0x" + h.StringEndian(binary.BigEndian),
		"0x" + strings.TrimLeft(h.StringEndian(binary.BigEndian), "0"),
		strings.ToUpper(h.StringEndian(binary.LittleEndian)),
		strings.ToUpper(h.StringEndian(binary.BigEndian)),
		"0x" + strings.ToUpper(h.StringEndian(binary.BigEndian)),
		"0x" + strings.ToUpper(strings.TrimLeft(h.StringEndian(binary.BigEndian), "0")),
	}
	if name, ok := a.Hashes[h]; ok {
		res = append(res, name)
	}
	return res
}

func (a *App) matchFileID(id stingray.FileID, glb glob.Glob, nameOnly bool) bool {
	nameVariations := a.hashNameVariationsForMatch(id.Name)

	var typeVariations []string
	if !nameOnly {
		typeVariations = a.hashNameVariationsForMatch(id.Type)
	}

	for _, name := range nameVariations {
		if nameOnly {
			if glb.Match(name) {
				return true
			}
		} else {
			for _, typ := range typeVariations {
				if glb.Match(name + "." + typ) {
					return true
				}
			}
		}
	}

	return false
}

func (a *App) MatchingFiles(
	includeGlob string,
	excludeGlob string,
	includeOnlyTypes []string,
	includeTriadIDs []stingray.Hash,
) (
	map[stingray.FileID]*stingray.File,
	error,
) {
	var inclGlob glob.Glob
	inclGlobNameOnly := !strings.Contains(includeGlob, ".")
	if includeGlob != "" {
		var err error
		inclGlob, err = glob.Compile(includeGlob)
		if err != nil {
			return nil, err
		}
	}
	var exclGlob glob.Glob
	exclGlobNameOnly := !strings.Contains(excludeGlob, ".")
	if excludeGlob != "" {
		var err error
		exclGlob, err = glob.Compile(excludeGlob)
		if err != nil {
			return nil, err
		}
	}

	var includeTriadFiles map[stingray.FileID]*stingray.File = make(map[stingray.FileID]*stingray.File)
	for _, includeTriadID := range includeTriadIDs {
		files, ok := a.DataDir.FilesByTriad[includeTriadID]
		if !ok {
			return nil, fmt.Errorf("triad %v does not exist", includeTriadID.String())
		}
		maps.Copy(includeTriadFiles, files)
	}

	res := make(map[stingray.FileID]*stingray.File)
	for id, file := range a.DataDir.Files {
		shouldIncl := true
		if len(includeTriadIDs) != 0 {
			if _, ok := includeTriadFiles[id]; !ok {
				shouldIncl = false
			}
		}
		if len(includeOnlyTypes) != 0 {
			typeVariations := a.hashNameVariationsForMatch(id.Type)
			if slices.ContainsFunc(includeOnlyTypes, func(includedType string) bool {
				return !slices.Contains(typeVariations, includedType)
			}) {
				continue
			}
		}
		if includeGlob != "" {
			// Include all files in triad even if they don't match the includeGlob - includeGlob will only add files to read
			shouldIncl = (len(includeTriadIDs) != 0 && shouldIncl) || a.matchFileID(id, inclGlob, inclGlobNameOnly)
		}
		if excludeGlob != "" {
			if a.matchFileID(id, exclGlob, exclGlobNameOnly) {
				shouldIncl = false
			}
		}
		if !shouldIncl {
			continue
		}

		res[id] = file
	}

	return res, nil
}

// Prints hash if human-readable name is unknown.
func (a *App) LookupHash(hash stingray.Hash) string {
	if name, ok := a.Hashes[hash]; ok {
		return name
	}
	return hash.String()
}

// Prints hash if human-readable name is unknown.
func (a *App) LookupThinHash(hash stingray.ThinHash) string {
	if name, ok := a.ThinHashes[hash]; ok {
		return name
	}
	return hash.String()
}

type extractContext struct {
	ctx      context.Context
	app      *App
	file     *stingray.File
	runner   *exec.Runner
	config   appconfig.Config
	outPath  string
	files    []string
	triadIDs []stingray.Hash
	printer  Printer
}

func newExtractContext(ctx context.Context, app *App, file *stingray.File, runner *exec.Runner, config appconfig.Config, outPath string, triadIDs []stingray.Hash, printer Printer) *extractContext {
	return &extractContext{
		ctx:      ctx,
		app:      app,
		file:     file,
		runner:   runner,
		config:   config,
		outPath:  outPath,
		triadIDs: triadIDs,
		printer:  printer,
	}
}

func (c *extractContext) OutPath() (string, error) { return c.outPath, nil }
func (c *extractContext) File() *stingray.File     { return c.file }
func (c *extractContext) Runner() *exec.Runner     { return c.runner }
func (c *extractContext) Config() appconfig.Config { return c.config }
func (c *extractContext) GetResource(name, typ stingray.Hash) (file *stingray.File, exists bool) {
	file, exists = c.app.DataDir.Files[stingray.FileID{Name: name, Type: typ}]
	return
}
func (c *extractContext) CreateFile(suffix string) (io.WriteCloser, error) {
	path, err := c.AllocateFile(suffix)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}
func (c *extractContext) AllocateFile(suffix string) (string, error) {
	path := c.outPath + suffix
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return "", err
	}
	c.files = append(c.files, path)
	return path, nil
}
func (c *extractContext) Ctx() context.Context { return c.ctx }
func (c *extractContext) Files() []string {
	return c.files
}
func (c *extractContext) Hashes() map[stingray.Hash]string {
	return c.app.Hashes
}
func (c *extractContext) ThinHashes() map[stingray.ThinHash]string {
	return c.app.ThinHashes
}
func (c *extractContext) TriadIDs() []stingray.Hash {
	return c.triadIDs
}
func (c *extractContext) ArmorSets() map[stingray.Hash]dlbin.ArmorSet {
	return c.app.ArmorSets
}
func (c *extractContext) Warnf(f string, a ...any) {
	name, typ := c.app.LookupHash(c.file.ID().Name), c.app.LookupHash(c.file.ID().Type)
	c.printer.Warnf("extract %v.%v: %v", name, typ, fmt.Sprintf(f, a...))
}
func (c *extractContext) LookupHash(hash stingray.Hash) string {
	return c.app.LookupHash(hash)
}
func (c *extractContext) LookupThinHash(hash stingray.ThinHash) string {
	return c.app.LookupThinHash(hash)
}

func getSourceExtractFunc(extrCfg appconfig.Config, typ string) (extr extractor.ExtractFunc) {
	switch extrCfg.Raw.Format {
	case "main":
		extr = extractor.ExtractFuncRawSingleType(typ, stingray.DataMain)
	case "stream":
		extr = extractor.ExtractFuncRawSingleType(typ, stingray.DataStream)
	case "gpu":
		extr = extractor.ExtractFuncRawSingleType(typ, stingray.DataGPU)
	case "combined":
		extr = extractor.ExtractFuncRawCombined(typ)
	default:
		extr = extractor.ExtractFuncRaw(typ)
	}
	return
}

// Returns path to extracted file/directory.
func (a *App) ExtractFile(ctx context.Context, id stingray.FileID, outDir string, extrCfg appconfig.Config, runner *exec.Runner, gltfDoc *gltf.Document, triadIDs []stingray.Hash, printer Printer) ([]string, error) {
	name, typ := a.LookupHash(id.Name), a.LookupHash(id.Type)

	file, ok := a.DataDir.Files[id]
	if !ok {
		return nil, fmt.Errorf("extract %v.%v: file does not exist", name, typ)
	}

	typeFormats := appconfig.GetTypeFormats(extrCfg)
	extrFormat := typeFormats[typ]

	var extr extractor.ExtractFunc
	if extrFormat == "raw" {
		extr = getSourceExtractFunc(extrCfg, typ)
	} else {
		switch typ {
		case "animation":
			extr = extr_animation.ExtractAnimationJson
		case "bik":
			if extrFormat == "bik" {
				extr = extr_bik.ExtractBik
			} else {
				extr = extr_bik.ConvertToMP4
			}
		case "wwise_stream":
			if extrFormat == "wwise" {
				extr = extr_wwise.ExtractWem
			} else {
				extr = extr_wwise.ConvertWem
			}
		case "wwise_bank":
			if extrFormat == "wwise" {
				extr = extr_wwise.ExtractBnk
			} else {
				extr = extr_wwise.ConvertBnk
			}
		case "material":
			if extrFormat == "textures" {
				extr = extr_material.ConvertTextures
			} else {
				extr = extr_material.Convert(gltfDoc)
			}
		case "unit":
			extr = extr_unit.Convert(gltfDoc)
		case "geometry_group":
			extr = extr_geogroup.Convert(gltfDoc)
		case "texture":
			if extrFormat == "dds" {
				extr = extr_texture.ExtractDDS
			} else {
				extr = extr_texture.ConvertToPNG
			}
		case "state_machine":
			extr = extr_state_machine.ExtractStateMachineJson
		case "strings":
			extr = extr_strings.ExtractStringsJSON
		case "package":
			extr = extr_package.ExtractPackageJSON
		case "bones":
			extr = extr_bones.ExtractBonesJSON
		default:
			extr = getSourceExtractFunc(extrCfg, typ)
		}
	}

	outPath := filepath.Join(outDir, name)
	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return nil, err
	}
	extrCtx := newExtractContext(
		ctx,
		a,
		file,
		runner,
		extrCfg,
		outPath,
		triadIDs,
		printer,
	)
	if err := extr(extrCtx); err != nil {
		{
			var err error
			var errPath string
			for _, path := range extrCtx.Files() {
				if e := os.Remove(path); e != nil && !errors.Is(e, os.ErrNotExist) && err == nil {
					err = e
					errPath = path
				}
			}
			if err != nil {
				return nil, fmt.Errorf("cleanup %v: %w", errPath, err)
			}
		}
		return nil, fmt.Errorf("extract %v.%v: %w", name, typ, err)
	}

	return extrCtx.Files(), nil
}
