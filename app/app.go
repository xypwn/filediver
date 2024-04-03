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
					PossibleValues: []string{"ogg", "wav", "aac", "mp3", "source"},
				},
			},
		},
		"wwise_bank": {
			Category: "audio",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"ogg", "wav", "aac", "mp3", "source"},
				},
			},
		},
		"bik": {
			Category: "video",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"mp4", "source"},
				},
			},
		},
		"texture": {
			Category: "image",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"png", "source"},
				},
			},
		},
		"unit": {
			Category: "model",
			Options: map[string]ConfigTemplateOption{
				"format": {
					PossibleValues: []string{"glb", "source"},
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

func New(printer *Printer) (*App, error) {
	a := &App{
		Hashes: make(map[stingray.Hash]string),
	}

	return a, nil
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

func (a *App) matchFileID(id stingray.FileID, glb glob.Glob) bool {
	nameVariations := []string{
		id.Name.StringEndian(binary.LittleEndian),
		id.Name.StringEndian(binary.BigEndian),
	}
	if name, ok := a.Hashes[id.Name]; ok {
		nameVariations = append(nameVariations, name)
	}

	typeVariations := []string{
		id.Type.StringEndian(binary.LittleEndian),
		id.Type.StringEndian(binary.BigEndian),
	}
	if typ, ok := a.Hashes[id.Type]; ok {
		typeVariations = append(typeVariations, typ)
	}

	for _, name := range nameVariations {
		for _, typ := range typeVariations {
			if glb.Match(name + "." + typ) {
				return true
			}
		}
	}

	return false
}

func (a *App) AllFiles() map[stingray.FileID]*stingray.File {
	return a.dataDir.Files
}

func (a *App) MatchingFiles(includeGlob, excludeGlob string, cfgTemplate ConfigTemplate, cfg map[string]extractor.Config) (map[stingray.FileID]*stingray.File, error) {
	var inclGlob glob.Glob
	if includeGlob != "" {
		if !strings.Contains(includeGlob, ".") {
			includeGlob += ".*"
		}
		var err error
		inclGlob, err = glob.Compile(includeGlob)
		if err != nil {
			return nil, err
		}
	}
	var exclGlob glob.Glob
	if excludeGlob != "" {
		if !strings.Contains(excludeGlob, ".") {
			excludeGlob += ".*"
		}
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
			shouldIncl = a.matchFileID(id, inclGlob)
		}
		if excludeGlob != "" {
			if a.matchFileID(id, exclGlob) {
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

func (a *App) ExtractFile(id stingray.FileID, outDir string, extrCfg map[string]extractor.Config, runner *exec.Runner) error {
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
		cfg = make(extractor.Config)
	}

	justExtract := cfg["format"] == "source"

	var extr extractor.ExtractFunc
	switch typ {
	case "bik":
		if justExtract {
			extr = extr_bik.Extract
		} else {
			extr = extr_bik.Convert
		}
	case "wwise_stream":
		if justExtract {
			extr = extr_wwise.ExtractWem
		} else {
			extr = extr_wwise.ConvertWem
		}
	case "wwise_bank":
		if justExtract {
			extr = extr_wwise.ExtractBnk
		} else {
			extr = extr_wwise.ConvertBnk
		}
	case "unit":
		if justExtract {
			extr = extractor.ExtractFuncRaw("unit")
		} else {
			extr = extr_unit.Convert
		}
	case "texture":
		if justExtract {
			extr = extr_texture.Extract
		} else {
			extr = extr_texture.Convert
		}
	default:
		extr = extractor.ExtractFuncRaw(typ)
	}

	var readers [3]io.ReadSeeker
	foundDataTypes := 0
	for dataType := stingray.DataType(0); dataType < stingray.NumDataType; dataType++ {
		if !file.Exists(dataType) {
			continue
		}
		r, err := file.Open(dataType)
		if err != nil {
			return err
		}
		defer r.Close()
		readers[dataType] = r
		foundDataTypes++
	}
	if foundDataTypes == 0 {
		return fmt.Errorf("extract %v.%v: no data", name, typ)
	}
	outPath := filepath.Join(outDir, name)
	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := extr(
		outPath,
		readers,
		cfg,
		runner,
		func(name, typ stingray.Hash) *stingray.File {
			return a.dataDir.Files[stingray.FileID{
				Name: name,
				Type: typ,
			}]
		},
	); err != nil {
		return fmt.Errorf("extract %v.%v: %w", name, typ, err)
	}

	return nil
}
