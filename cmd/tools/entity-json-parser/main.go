package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"

	"os/signal"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"syscall"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/extractor/entity"
	extr_level "github.com/xypwn/filediver/extractor/level"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/level"
)

var prt app.Printer

func getFileId(inputPath, filetype string) (stingray.FileID, error) {
	base := filepath.Base(inputPath)
	fileId := stingray.FileID{
		Name: stingray.Hash{},
		Type: stingray.Sum(filetype),
	}
	if strings.Contains(base, "0x") {
		filename, _, _ := strings.Cut(base, ".")
		nameHash, err := stingray.ParseHash(filename)
		if err != nil {
			return stingray.FileID{}, err
		}
		fileId.Name = nameHash
	} else if strings.Contains(inputPath, "content") {
		pathList := strings.Split(inputPath, string(os.PathSeparator))
		start := slices.Index(pathList, "content")
		prt.Infof("%v", path.Join(pathList[start:]...))
		filename, _, _ := strings.Cut(path.Join(pathList[start:]...), ".")
		fileId.Name = stingray.Sum(filename)
	}
	return fileId, nil
}

func main() {
	prt = app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	parser := argparse.NewParser("entity-json-parser", "", nil)
	entities := parser.Strings("e", "entities", &argparse.Option{
		Positional: false,
	})
	levels := parser.Strings("l", "levels", &argparse.Option{
		Positional: false,
	})
	output := parser.String("o", "output", &argparse.Option{
		Positional: false,
	})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Unable to detect game install directory.")
	}

	a, err := app.OpenGameDir(ctx, gameDir, nil, nil, stingray.ThinHash{}, func(curr int, total int) {
		prt.Statusf("Opening game directory %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			prt.NoStatus()
			prt.Warnf("Game load canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	archive := stingray.Archive{}
	marshalledFiles := make(map[stingray.FileID][]byte)

	outputPath := filepath.Clean(*output)
	for _, name := range *entities {
		prt.Infof("%v", name)
		inputPath := filepath.Clean(name)
		if _, err := os.Stat(inputPath); err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}
		fileId, err := getFileId(inputPath, "entity")
		if err != nil {
			prt.Fatalf("getFileId: %v")
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

	for _, name := range *levels {
		prt.Infof("%v", name)
		inputPath := filepath.Clean(name)
		if _, err := os.Stat(inputPath); err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}
		fileId, err := getFileId(inputPath, "level")
		if err != nil {
			prt.Fatalf("getFileId: %v")
		}

		var fileInfo []stingray.FileInfo
		var ok bool
		if fileInfo, ok = a.DataDir.Files[fileId]; !ok || !fileInfo[0].Exists(stingray.DataMain) {
			prt.Errorf("%v.level not found, skipping", fileId.Name.String())
			continue
		}

		levelData, err := a.DataDir.Read(fileId, stingray.DataMain)
		if err != nil {
			prt.Fatalf("%v.level: %v", fileId.Name.String(), err)
		}

		prt.Infof("Opening '%v'", inputPath)
		inputFile, err := os.Open(inputPath)
		if err != nil {
			prt.Fatalf("%v: %v", inputPath, err)
		}
		defer inputFile.Close()

		var simple extr_level.SimpleLevel
		data, err := io.ReadAll(inputFile)
		if err != nil {
			prt.Fatalf("reading file: %v", err)
		}
		err = json.Unmarshal(data, &simple)
		if err != nil {
			prt.Fatalf("unmarshal: %v", err)
		}

		levelEntity, err := simple.Entity.ToEntity()
		if err != nil {
			prt.Fatalf("level entity toEntity: %v", err)
		}
		levelEntityData, err := levelEntity.MarshalBinary()
		if err != nil {
			prt.Fatalf("level entity MarshalBinary: %v", err)
		}

		var gameLevelHeader level.RawLevel
		if _, err := binary.Decode(levelData, binary.LittleEndian, &gameLevelHeader); err != nil {
			prt.Fatalf("level decode: %v", err)
		}

		prt.Infof("level entity offset: %#08x", gameLevelHeader.EntityOffset)
		result := make([]byte, 0)
		result = append(result, levelData[:gameLevelHeader.EntityOffset]...)
		result = append(result, levelEntityData...)
		result = append(result, levelData[gameLevelHeader.EntityOffset+uint32(len(levelEntityData)):]...)
		archive.AddFile(fileId, uint32(len(result)), 16)
		marshalledFiles[fileId] = result
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
