package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

func sortedMapFileIDKeys[V any](m map[stingray.FileID]V) []stingray.FileID {
	res := make([]stingray.FileID, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Name.Value < res[j].Name.Value {
			return true
		} else if res[i].Name.Value > res[j].Name.Value {
			return false
		}
		return res[i].Type.Value < res[j].Type.Value
	})
	return res
}

func sortedMapHashKeys[V any](m map[stingray.Hash]V) []stingray.Hash {
	res := make([]stingray.Hash, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Value < res[j].Value
	})
	return res
}

func printUsage() {
	fmt.Println(`Usage: crossref-checker [options] GAME_FILES_TO_SEARCH_GLOB OUTPUT_FILE

Finds all occurrences of existing game file hashes in other game files.

options:
  -S  --  exclude references by the file to itself

examples:
  crossref-checker "*.material" material_crossrefs.txt  --  search all game files with the "material" extension for references to other game files and output result to "material_crossrefs.txt"`)
}

func main() {
	var inclGlob string
	var outFilePath string
	excludeSelfReferences := false
	{
		var nonFlagArgs []string
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "-") {
				switch arg {
				case "-S":
					excludeSelfReferences = true
				default:
					printUsage()
					os.Exit(0)
				}
			} else {
				nonFlagArgs = append(nonFlagArgs, arg)
			}
		}
		if len(nonFlagArgs) < 2 {
			printUsage()
			os.Exit(0)
		}
		inclGlob, outFilePath = nonFlagArgs[0], nonFlagArgs[1]
	}

	prt := app.NewPrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	gameDir, err := app.DetectGameDir()
	if err == nil {
		prt.Infof("Using game found at: \"%v\"", gameDir)
	} else {
		prt.Errorf("Helldivers 2 Steam installation path not found: %v", err)
		prt.Fatalf("Command line option for installation path not implemented in crossref-checker. Please open an issue on GitHub.")
	}

	ctx := context.Background() // no need to exit cleanly since we're only reading
	knownHashes := app.ParseHashes(hashes.Hashes)
	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, []string{}, nil, stingray.Hash{}, func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}

	appConfig, err := app.ParseExtractorConfig(app.ConfigFormat, "enable:all")
	if err != nil {
		panic(err) // should never fail
	}

	allExistingFileNamesAsBytes := make(map[[8]byte]struct{})
	for k := range a.DataDir.Files {
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], k.Name.Value)
		allExistingFileNamesAsBytes[b] = struct{}{}
	}

	files, err := a.MatchingFiles(inclGlob, "", nil, app.ConfigFormat, appConfig)
	if err != nil {
		prt.Fatalf("Error matching files: %v", err)
	}

	if len(files) == 0 {
		prt.Fatalf("Glob \"%v\" doesn't match any game files. Use `filediver -c \"enable:all\" -l` to list game files.", inclGlob)
	}

	{
		tail := ""
		if excludeSelfReferences {
			tail += " (excluding itself)"
		} else {
			tail += " (including itself)"
		}
		prt.Infof("Going to search %v files for cross-references to any other files%v", len(files), tail)
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		prt.Fatalf("Opening output file: %v", err)
	}
	defer outFile.Close()

	searchedFileCounter := 0
	crossrefCounter := 0

	var filebuf bytes.Buffer
	for _, fileID := range sortedMapFileIDKeys(files) {
		file := files[fileID]
		for dataType := stingray.DataType(0); dataType < stingray.NumDataType; dataType++ {
			if !file.Exists(dataType) {
				continue
			}
			{
				rd, err := file.Open(ctx, dataType)
				if err != nil {
					prt.Errorf("Error opening file: %v", err)
					continue
				}
				filebuf.Reset()
				_, err = io.Copy(&filebuf, rd)
				rd.Close()
				if err != nil {
					prt.Errorf("Error reading file: %v", err)
					continue
				}
			}
			bytesToSearchIn := filebuf.Bytes()
			byteOffsetsByHash := make(map[stingray.Hash][]int) // where each match (that was found) was found
			for offset := 0; offset <= len(bytesToSearchIn)-8; offset++ {
				match := bytesToSearchIn[offset : offset+8]
				if _, ok := allExistingFileNamesAsBytes[[8]byte(match)]; ok {
					foundHash := stingray.Hash{Value: binary.LittleEndian.Uint64(match)}
					if excludeSelfReferences && foundHash.Value == fileID.Name.Value {
						continue
					}
					byteOffsetsByHash[foundHash] = append(byteOffsetsByHash[foundHash], offset)
					crossrefCounter++
				}
			}
			for _, hash := range sortedMapHashKeys(byteOffsetsByHash) {
				byteOffsets := byteOffsetsByHash[hash]
				fmt.Fprintf(outFile, "%v.%v (%v) -> %v %v time(s), offsets: %v\n", a.LookupHash(file.ID().Name), a.LookupHash(file.ID().Type), dataType, a.LookupHash(hash), len(byteOffsets), byteOffsets)
			}
		}
		prt.Statusf("Searched file %v/%v (found %v cross-references)", searchedFileCounter+1, len(files), crossrefCounter)
		searchedFileCounter++
	}
	fmt.Println()
}
