package app

import (
	"bufio"
	"encoding/binary"
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
	*Printer
	Hashes  map[stingray.Hash]string
	gameDir string
	dataDir *stingray.DataDir
}

func New(printer *Printer) (*App, error) {
	a := &App{
		Printer: printer,
		Hashes:  make(map[stingray.Hash]string),
	}

	// HACK: We don't know this hash's source string yet
	a.Hashes[stingray.Hash{Value: 0xeac0b497876adedf}] = "material"

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
	return nil
}

func (a *App) AddHashesFromString(str string) {
	sc := bufio.NewScanner(strings.NewReader(str))
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s != "" && !strings.HasPrefix(s, "//") {
			a.Hashes[stingray.Sum64([]byte(s))] = s
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
