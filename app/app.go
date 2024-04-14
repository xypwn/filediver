package app

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	extr_bik "github.com/xypwn/filediver/extractor/bik"
	extr_texture "github.com/xypwn/filediver/extractor/texture"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	extr_wwise "github.com/xypwn/filediver/extractor/wwise"
	"github.com/xypwn/filediver/steampath"
	"github.com/xypwn/filediver/stingray"
)

var ConfigFormat = ConfigTemplate{
	Extractors: map[string]ConfigTemplateExtractor{
		"wwise_stream": {
			Category: "audio",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"ogg", "wav", "aac", "mp3", "wem", "source"},
				},
			},
		},
		"wwise_bank": {
			Category: "audio",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"ogg", "wav", "aac", "mp3", "bnk", "source"},
				},
			},
		},
		"bik": {
			Category: "video",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"mp4", "bik", "source"},
				},
			},
		},
		"texture": {
			Category: "image",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"png", "dds", "source"},
				},
			},
		},
		"unit": {
			Category: "model",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"glb", "source"},
				},
				"meshes": {
					PossibleValues: []string{"highest_detail", "all"},
				},
			},
		},
		"raw": {
			Category: "",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"source"},
				},
			},
			DefaultDisabled: true,
		},
	},
	Fallback: "raw",
}

type App struct {
	Hashes  map[stingray.Hash]string
	gameDir string
	dataDir *stingray.DataDir
}

func New() *App {
	a := &App{
		Hashes: make(map[stingray.Hash]string),
	}

	return a
}

func (a *App) SetGameDir(path string) error {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return fmt.Errorf("invalid game directory: %v: not a directory", path)
	}
	if info, err := os.Stat(filepath.Join(path, "settings.ini")); err == nil && info.Mode().IsRegular() {
		// We were given the "data" directory => go back
		path = filepath.Dir(path)
	}
	if info, err := os.Stat(filepath.Join(path, "data", "settings.ini")); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("invalid game directory: %v: no valid data directory not found", path)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	a.gameDir = path
	return nil
}

func (a *App) DetectGameDir() (string, error) {
	path, err := steampath.GetAppPath("553850", "Helldivers 2")
	if err != nil {
		return "", err
	}
	if err := a.SetGameDir(path); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) OpenGameDir() error {
	dataDir, err := stingray.OpenDataDir(filepath.Join(a.gameDir, "data"))
	if err != nil {
		return err
	}
	a.dataDir = dataDir

	// wwise_dep files let us know the string of many of the wwise_banks
	for id, file := range dataDir.Files {
		if id.Type == stingray.Sum64([]byte("wwise_dep")) {
			if err := a.addHashFromWwiseDep(*file); err != nil {
				return fmt.Errorf("wwise_dep: %w", err)
			}
		}
	}

	return nil
}

func (a *App) addHashFromWwiseDep(f stingray.File) error {
	r, err := f.Open(stingray.DataMain)
	if err != nil {
		return err
	}
	var magicNum [4]byte
	if _, err := io.ReadFull(r, magicNum[:]); err != nil {
		return err
	}
	if magicNum != [4]byte{0xd8, '/', 'v', 'x'} {
		return errors.New("invalid magic number")
	}
	var textLen uint32
	if err := binary.Read(r, binary.LittleEndian, &textLen); err != nil {
		return err
	}
	text := make([]byte, textLen-1)
	if _, err := io.ReadFull(r, text); err != nil {
		return err
	}
	a.AddHashFromString(string(text))
	return nil
}

func (a *App) AddHashFromString(str string) {
	a.Hashes[stingray.Sum64([]byte(str))] = str
}

func (a *App) AddHashesFromString(str string) {
	sc := bufio.NewScanner(strings.NewReader(str))
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s != "" && !strings.HasPrefix(s, "//") {
			a.AddHashFromString(s)
		}
	}
}

func (a *App) AddHashesFromFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	a.AddHashesFromString(string(b))
	return nil
}

func (a *App) File(id stingray.FileID) (f *stingray.File, exists bool) {
	f, exists = a.dataDir.Files[id]
	return
}

func (a *App) matchFileID(id stingray.FileID, glb glob.Glob, nameOnly bool) bool {
	nameVariations := []string{
		id.Name.StringEndian(binary.LittleEndian),
		id.Name.StringEndian(binary.BigEndian),
		"0x" + id.Name.StringEndian(binary.BigEndian),
	}
	if name, ok := a.Hashes[id.Name]; ok {
		nameVariations = append(nameVariations, name)
	}

	var typeVariations []string
	if !nameOnly {
		typeVariations = []string{
			id.Type.StringEndian(binary.LittleEndian),
			id.Type.StringEndian(binary.BigEndian),
			"0x" + id.Name.StringEndian(binary.BigEndian),
		}
		if typ, ok := a.Hashes[id.Type]; ok {
			typeVariations = append(typeVariations, typ)
		}
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

func (a *App) AllFiles() map[stingray.FileID]*stingray.File {
	return a.dataDir.Files
}

func (a *App) MatchingFiles(includeGlob, excludeGlob string, cfgTemplate ConfigTemplate, cfg map[string]map[string]string) (map[stingray.FileID]*stingray.File, error) {
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

	res := make(map[stingray.FileID]*stingray.File)
	for id, file := range a.dataDir.Files {
		shouldIncl := true
		if includeGlob != "" {
			shouldIncl = a.matchFileID(id, inclGlob, inclGlobNameOnly)
		}
		if excludeGlob != "" {
			if a.matchFileID(id, exclGlob, exclGlobNameOnly) {
				shouldIncl = false
			}
		}
		if !shouldIncl {
			continue
		}

		typ, ok := a.Hashes[id.Type]
		if !ok {
			typ = id.Type.String()
		}
		_, knownType := cfgTemplate.Extractors[typ]
		if !knownType {
			typ = cfgTemplate.Fallback
		}

		shouldExtract := true
		if extrTempl, ok := cfgTemplate.Extractors[typ]; ok {
			shouldExtract = !extrTempl.DefaultDisabled
		}
		if cfg["enable"] != nil {
			shouldExtract = cfg["enable"][typ] == "true"
		}
		if cfg["disable"] != nil {
			if cfg["disable"][typ] == "true" {
				shouldExtract = false
			}
		}
		if !shouldExtract {
			continue
		}

		res[id] = file
	}

	return res, nil
}

type extractContext struct {
	app     *App
	file    *stingray.File
	runner  *exec.Runner
	config  map[string]string
	outPath string
}

func newExtractContext(app *App, file *stingray.File, runner *exec.Runner, config map[string]string, outPath string) *extractContext {
	return &extractContext{
		app:     app,
		file:    file,
		runner:  runner,
		config:  config,
		outPath: outPath,
	}
}

func (c *extractContext) File() *stingray.File      { return c.file }
func (c *extractContext) Runner() *exec.Runner      { return c.runner }
func (c *extractContext) Config() map[string]string { return c.config }
func (c *extractContext) GetResource(name, typ stingray.Hash) (file *stingray.File, exists bool) {
	file, exists = c.app.AllFiles()[stingray.FileID{Name: name, Type: typ}]
	return
}
func (c *extractContext) CreateFile(suffix string) (io.WriteCloser, error) {
	return os.Create(c.outPath + suffix)
}
func (c *extractContext) CreateFileDir(dirSuffix, filename string) (io.WriteCloser, error) {
	dir := c.outPath + dirSuffix
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(filepath.Join(dir, filename))
}
func (c *extractContext) OutPath() (string, error) { return c.outPath, nil }

func (a *App) ExtractFile(id stingray.FileID, outDir string, extrCfg map[string]map[string]string, runner *exec.Runner) error {
	name, ok := a.Hashes[id.Name]
	if !ok {
		name = id.Name.String()
	}
	typ, ok := a.Hashes[id.Type]
	if !ok {
		typ = id.Type.String()
	}

	file, ok := a.dataDir.Files[id]
	if !ok {
		return fmt.Errorf("extract %v.%v: file does not found", name, typ)
	}

	cfg := extrCfg[typ]
	if cfg == nil {
		cfg = make(map[string]string)
	}

	var extr extractor.ExtractFunc
	switch typ {
	case "bik":
		if cfg["format"] == "source" {
			extr = extractor.ExtractFuncRaw(".stingray_bik", stingray.DataMain, stingray.DataStream, stingray.DataGPU)
		} else if cfg["format"] == "bik" {
			extr = extr_bik.ExtractBik
		} else {
			extr = extr_bik.ConvertToMP4
		}
	case "wwise_stream":
		if cfg["format"] == "source" {
			extr = extractor.ExtractFuncRaw(".stingray_wem", stingray.DataMain, stingray.DataStream, stingray.DataGPU)
		} else if cfg["format"] == "wem" {
			extr = extractor.ExtractFuncRaw(".wem", stingray.DataStream)
		} else {
			extr = extr_wwise.ConvertWem
		}
	case "wwise_bank":
		if cfg["format"] == "source" {
			extr = extractor.ExtractFuncRaw(".bnk", stingray.DataMain, stingray.DataStream, stingray.DataGPU)
		} else if cfg["format"] == "bnk" {
			extr = extr_wwise.ExtractBnk
		} else {
			extr = extr_wwise.ConvertBnk
		}
	case "unit":
		if cfg["format"] == "source" {
			extr = extractor.ExtractFuncRaw(".unit", stingray.DataMain, stingray.DataStream, stingray.DataGPU)
		} else {
			extr = extr_unit.Convert
		}
	case "texture":
		if cfg["format"] == "source" {
			extr = extractor.ExtractFuncRaw(".texture", stingray.DataMain, stingray.DataStream, stingray.DataGPU)
		} else if cfg["format"] == "dds" {
			extr = extr_texture.ExtractDDS
		} else {
			extr = extr_texture.ConvertToPNG
		}
	default:
		extr = extractor.ExtractFuncRaw(typ)
	}

	outPath := filepath.Join(outDir, name)
	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := extr(newExtractContext(
		a,
		file,
		runner,
		cfg,
		outPath,
	)); err != nil {
		return fmt.Errorf("extract %v.%v: %w", name, typ, err)
	}

	return nil
}
