package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	extr_entity "github.com/xypwn/filediver/extractor/entity"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

var lookupThinHash func(stingray.ThinHash) string
var lookupHash func(stingray.Hash) string
var prt app.Printer

func main() {
	prt = app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	parser := argparse.NewParser("entity-json-parser", "", nil)
	entities := parser.Strings("e", "entities", &argparse.Option{
		Positional: false,
		Required:   true,
	})
	outputDirectory := parser.String("o", "output", &argparse.Option{
		Positional: false,
	})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("%v", err)
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	hashesMap := make(map[stingray.Hash]string)
	for _, h := range knownHashes {
		hashesMap[stingray.Sum(h)] = h
	}
	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range knownThinHashes {
		thinHashesMap[stingray.Sum(h).Thin()] = h
	}

	lookupThinHash = func(t stingray.ThinHash) string {
		if res, found := thinHashesMap[t]; found {
			return res
		}
		return t.String()
	}

	lookupHash = func(h stingray.Hash) string {
		if res, found := hashesMap[h]; found {
			return res
		}
		return h.String()
	}

	if err := os.MkdirAll(filepath.Dir(*outputDirectory), os.ModePerm); err != nil {
		prt.Fatalf("%v: %v", *outputDirectory, err)
	}

	for _, name := range *entities {
		prt.Infof("%v", name)
		inputPath := filepath.Clean(name)

		if _, err := os.Stat(inputPath); err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}

		inputFile, err := os.Open(inputPath)
		if err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}
		defer inputFile.Close()

		var simple extr_entity.SimpleEntity
		data, err := io.ReadAll(inputFile)
		if err != nil {
			prt.Fatalf("reading file: %v", err)
		}
		err = json.Unmarshal(data, &simple)
		if err != nil {
			prt.Fatalf("unmarshal: %v", err)
		}

		entity, err := simple.ToEntity()
		if err != nil {
			prt.Fatalf("ToEntity: %v", err)
		}

		filename, _, _ := strings.Cut(filepath.Base(name), ".")
		outputPath := filepath.Join(*outputDirectory, filename) + ".asset_grading_lut.dds"

		out, err := os.Create(outputPath)
		if err != nil {
			prt.Fatalf("create %v: %v", outputPath, err)
		}
		defer out.Close()
		err = extr_entity.WriteDDSColorGradingLut(out, entity)
		if err != nil {
			prt.Errorf("Writing color grading lut: %v", err)
		}
	}
}
