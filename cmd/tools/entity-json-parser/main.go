package main

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/extractor/entity"
)

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	parser := argparse.NewParser("entity-json-parser", "", nil)
	input := parser.String("i", "input", &argparse.Option{
		Positional: true,
	})
	output := parser.String("o", "output", &argparse.Option{
		Positional: true,
	})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("%v", err)
	}

	inputPath := path.Clean(*input)
	if _, err := os.Stat(inputPath); err != nil {
		prt.Fatalf("%v: %v", inputPath, err)
	}
	outputPath := path.Clean(*output)

	prt.Infof("Opening '%v'", inputPath)
	inputFile, err := os.Open(inputPath)
	if err != nil {
		prt.Fatalf("%v: %v", inputPath, err)
	}
	defer inputFile.Close()

	var simple entity.SimpleEntity
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

	entityData, err := entity.MarshalBinary()
	if err != nil {
		prt.Fatalf("MarshalBinary: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		prt.Fatalf("%v: %v", outputPath, err)
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		prt.Fatalf("%v: %v", outputPath, err)
	}
	defer outputFile.Close()

	_, err = outputFile.Write(entityData)
	if err != nil {
		prt.Fatalf("Write: %v", err)
	}
}
