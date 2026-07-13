package main

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/extractor/entity"
	"github.com/xypwn/filediver/stingray"
)

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	parser := argparse.NewParser("entity-json-parser", "", nil)
	input := parser.Strings("i", "input", &argparse.Option{
		Positional: false,
	})
	output := parser.String("o", "output", &argparse.Option{
		Positional: false,
	})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("%v", err)
	}

	archive := stingray.Archive{}
	marshalledFiles := make(map[stingray.FileID][]byte)

	outputPath := filepath.Clean(*output)
	for _, name := range *input {
		prt.Infof("%v", name)
		inputPath := filepath.Clean(name)
		if _, err := os.Stat(inputPath); err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}
		base := filepath.Base(inputPath)
		fileId := stingray.FileID{
			Name: stingray.Hash{},
			Type: stingray.Sum("entity"),
		}
		if strings.Contains(base, "0x") {
			filename, _, _ := strings.Cut(base, ".")
			prt.Infof("%v\n%v", filename, base)
			nameHash, err := stingray.ParseHash(filename)
			if err != nil {
				prt.Fatalf("%v", err)
			}
			fileId.Name = nameHash
		} else if strings.Contains(inputPath, "content") {
			pathList := filepath.SplitList(inputPath)
			start := slices.Index(pathList, "content")
			filename, _, _ := strings.Cut(path.Join(pathList[start:]...), ".")
			prt.Infof("%v", filename)
			fileId.Name = stingray.Sum(filename)
		}

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

		archive.AddFile(fileId, uint32(len(entityData)), 16)
		marshalledFiles[fileId] = entityData
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		prt.Fatalf("%v: %v", outputPath, err)
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		prt.Fatalf("%v: %v", outputPath, err)
	}
	defer outputFile.Close()

	{
		archiveData, err := archive.MarshalBinary()
		if err != nil {
			prt.Fatalf("Marshal archive: %v", err)
		}
		_, err = outputFile.Write(archiveData)
		if err != nil {
			prt.Fatalf("Write archive: %v", err)
		}
	}

	for _, fileInfo := range archive.Files {
		offset, err := outputFile.Seek(0, io.SeekCurrent)
		if err != nil {
			prt.Fatalf("Seek current: %v", err)
		}
		if offset < int64(fileInfo.Offsets[stingray.DataMain]) {
			_, err := outputFile.Write(make([]byte, int64(fileInfo.Offsets[stingray.DataMain])-offset))
			if err != nil {
				prt.Fatalf("Write padding: %v", err)
			}
		} else if offset > int64(fileInfo.Offsets[stingray.DataMain]) {
			prt.Fatalf("offset exceeded expected offset of file by %v bytes", offset-int64(fileInfo.Offsets[stingray.DataMain]))
		}
		_, err = outputFile.Write(marshalledFiles[fileInfo.ID])
		if err != nil {
			prt.Fatalf("Write file %v.entity: %v", fileInfo.ID.Name.String(), err)
		}
	}
}
