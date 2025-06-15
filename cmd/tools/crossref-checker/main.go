package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func sortedMapThinHashKeys[V any](m map[stingray.ThinHash]V) []stingray.ThinHash {
	res := make([]stingray.ThinHash, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Value < res[j].Value
	})
	return res
}

func printUsage() {
	fmt.Println(`Usage:
  crossref-checker [options] GAME_FILES_TO_SEARCH_GLOB [ADDITIONAL_INPUT_FILES_GLOB...]
  or
  crossref-checker [options] -h HASH [ADDITIONAL_INPUT_FILES_GLOB...]

Finds all occurrences of existing game file hashes in other game files.

options:
  -o OUTPUT_FILE  --  output file path (default: "fd_crossrefs.txt")
  -S              --  exclude references by the file to itself
  -h HASH         --  list of hashes to search for, separated by commas, spaces or newlines; prefix with 0x for big endian, no prefix for little endian
  -H              --  only search for supplied hashes
  -i              --  only search through supplied input files

examples:
  crossref-checker "*.material" material_crossrefs.txt  --  search all game files with the "material" extension for references to other game files and output result to "material_crossrefs.txt"`)
}

func parseFreeArg(args *[]string) string {
	if len(*args) > 0 {
		res := (*args)[0]
		*args = (*args)[1:]
		return res
	} else {
		return ""
	}
}

func parseFlag(args *[]string, optionName string) bool {
	if len(*args) > 0 && (*args)[0] == optionName {
		*args = (*args)[1:]
		return true
	} else {
		return false
	}
}

func parseArgWithParam(args *[]string, optionName string) (string, bool) {
	if len(*args) > 0 && (*args)[0] == optionName {
		*args = (*args)[1:]
		var param string
		if len(*args) > 0 {
			param = (*args)[0]
			*args = (*args)[1:]
			return param, true
		} else {
			return "", false
		}
	} else {
		return "", false
	}
}

func main() {
	var inclGlob string
	outFilePath := "fd_crossrefs.txt"
	excludeSelfReferences := false
	specifiedHashesAsBytes := make(map[[8]byte]struct{})
	specifiedThinHashesAsBytes := make(map[[4]byte]struct{})
	var additionalInputFiles []string
	onlySearchForSpecifiedHashes := false
	onlySearchThroughAdditionalInputFiles := false
	{
		var specifiedHashes []string
		var freeArgs []string
		args := os.Args[1:]
		for len(args) > 0 {
			if p, ok := parseArgWithParam(&args, "-o"); ok {
				outFilePath = p
			} else if parseFlag(&args, "-S") {
				excludeSelfReferences = true
			} else if parseFlag(&args, "-H") {
				onlySearchForSpecifiedHashes = true
			} else if parseFlag(&args, "-i") {
				onlySearchThroughAdditionalInputFiles = true
			} else if p, ok := parseArgWithParam(&args, "-h"); ok {
				specifiedHashes = strings.FieldsFunc(p, func(r rune) bool {
					return strings.ContainsRune(", \n\r\t", r)
				})
			} else {
				freeArgs = append(freeArgs, parseFreeArg(&args))
			}
		}
		if len(specifiedHashes) == 0 {
			if len(freeArgs) < 1 {
				printUsage()
				os.Exit(0)
			}
			inclGlob = freeArgs[0]
			freeArgs = freeArgs[1:]
		}
		for _, freeArg := range freeArgs {
			paths, err := filepath.Glob(freeArg)
			if err != nil {
				printUsage()
				os.Exit(0)
			}
			additionalInputFiles = append(additionalInputFiles, paths...)
		}
		if len(specifiedHashes) != 0 {
			for _, k := range specifiedHashes {
				k, isBigEndian := strings.CutPrefix(k, "0x")
				rawBytes, err := hex.DecodeString(k)
				if err != nil || (len(rawBytes) != 8 && len(rawBytes) != 4) {
					fmt.Println("Invalid hash string. Must be in hexadecimal and 4 or 8 bytes (8 or 16 digits) long. 0x prefix to indicate big endian.")
					printUsage()
					os.Exit(0)
				}
				if isBigEndian {
					// Files are encoded in little endian, so reverse byte order
					for i := 0; i < len(rawBytes)/2; i++ {
						j := len(rawBytes) - 1 - i
						rawBytes[i], rawBytes[j] = rawBytes[j], rawBytes[i]
					}
				}
				switch len(rawBytes) {
				case 8:
					specifiedHashesAsBytes[[8]byte(rawBytes)] = struct{}{}
				case 4:
					specifiedThinHashesAsBytes[[4]byte(rawBytes)] = struct{}{}
				default:
					panic("unreachable")
				}
			}
		}
	}

	prt := app.NewConsolePrinter(
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
	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, []string{}, stingray.Hash{}, func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}

	searchForHashesAsBytes := make(map[[8]byte]struct{})
	if len(specifiedHashesAsBytes) == 0 {
		if !onlySearchForSpecifiedHashes {
			for k := range a.DataDir.Files {
				var b [8]byte
				binary.LittleEndian.PutUint64(b[:], k.Name.Value)
				searchForHashesAsBytes[b] = struct{}{}
			}
		}
	} else {
		searchForHashesAsBytes = specifiedHashesAsBytes
	}
	searchForThinHashesAsBytes := specifiedThinHashesAsBytes

	files, err := a.MatchingFiles(inclGlob, "", nil, nil)
	if err != nil {
		prt.Fatalf("Error matching files: %v", err)
	}

	if len(files) == 0 {
		prt.Fatalf("Glob \"%v\" doesn't match any game files. Use `filediver -c \"enable:all\" -l` to list game files.", inclGlob)
	}

	var numFilesToSearch int
	if !onlySearchThroughAdditionalInputFiles {
		numFilesToSearch += len(files)
	}
	numFilesToSearch += len(additionalInputFiles)

	{
		tail := ""
		if excludeSelfReferences {
			tail += " (excluding itself)"
		} else {
			tail += " (including itself)"
		}
		prt.Infof("Going to search %v files for cross-references to any other files%v", numFilesToSearch, tail)
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		prt.Fatalf("Opening output file: %v", err)
	}
	defer outFile.Close()

	searchedFileCounter := 0
	crossrefCounter := 0

	findMatchingBytes := func(
		bytesToSearchIn []byte, excludeHash *stingray.Hash,
	) (
		byteOffsetsByHash map[stingray.Hash][]int,
		byteOffsetsByThinHash map[stingray.ThinHash][]int,
	) {
		byteOffsetsByHash = make(map[stingray.Hash][]int) // where each match (that was found) was found
		for offset := 0; offset <= len(bytesToSearchIn)-8; offset++ {
			match := bytesToSearchIn[offset : offset+8]
			if _, ok := searchForHashesAsBytes[[8]byte(match)]; ok {
				foundHash := stingray.Hash{Value: binary.LittleEndian.Uint64(match)}
				if excludeSelfReferences && excludeHash != nil && foundHash.Value == excludeHash.Value {
					continue
				}
				byteOffsetsByHash[foundHash] = append(byteOffsetsByHash[foundHash], offset)
			}
		}
		byteOffsetsByThinHash = make(map[stingray.ThinHash][]int)
		for offset := 0; offset <= len(bytesToSearchIn)-4; offset++ {
			match := bytesToSearchIn[offset : offset+4]
			if _, ok := searchForThinHashesAsBytes[[4]byte(match)]; ok {
				foundHash := stingray.ThinHash{Value: binary.LittleEndian.Uint32(match)}
				byteOffsetsByThinHash[foundHash] = append(byteOffsetsByThinHash[foundHash], offset)
			}
		}
		return byteOffsetsByHash, byteOffsetsByThinHash
	}

	var filebuf bytes.Buffer
	if !onlySearchThroughAdditionalInputFiles {
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
				byteOffsetsByHash, byteOffsetsByThinHash := findMatchingBytes(filebuf.Bytes(), &fileID.Name)
				for _, hash := range sortedMapHashKeys(byteOffsetsByHash) {
					byteOffsets := byteOffsetsByHash[hash]
					fmt.Fprintf(outFile, "gamefile %v.%v (%v) -> %v %v time(s), offsets: %v\n", a.LookupHash(file.ID().Name), a.LookupHash(file.ID().Type), dataType, a.LookupHash(hash), len(byteOffsets), byteOffsets)
					crossrefCounter++
				}
				for _, hash := range sortedMapThinHashKeys(byteOffsetsByThinHash) {
					byteOffsets := byteOffsetsByThinHash[hash]
					fmt.Fprintf(outFile, "gamefile %v.%v (%v) -> %v %v time(s), offsets: %v\n", a.LookupHash(file.ID().Name), a.LookupHash(file.ID().Type), dataType, hash, len(byteOffsets), byteOffsets)
					crossrefCounter++
				}
			}
			prt.Statusf("Searched file %v/%v (found %v cross-references)", searchedFileCounter+1, numFilesToSearch, crossrefCounter)
			searchedFileCounter++
		}
	}
	for _, filePath := range additionalInputFiles {
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			prt.Errorf("Error reading custom file: %v", err)
			continue
		}
		byteOffsetsByHash, byteOffsetsByThinHash := findMatchingBytes(bytes, nil)
		for _, hash := range sortedMapHashKeys(byteOffsetsByHash) {
			byteOffsets := byteOffsetsByHash[hash]
			fmt.Fprintf(outFile, "customfile %v -> %v %v time(s), offsets: %v\n", filePath, a.LookupHash(hash), len(byteOffsets), byteOffsets)
			crossrefCounter++
		}
		for _, hash := range sortedMapThinHashKeys(byteOffsetsByThinHash) {
			byteOffsets := byteOffsetsByThinHash[hash]
			fmt.Fprintf(outFile, "customfile %v -> %v %v time(s), offsets: %v\n", filePath, hash, len(byteOffsets), byteOffsets)
			crossrefCounter++
		}
		prt.Statusf("Searched file %v/%v (found %v cross-references)", searchedFileCounter+1, numFilesToSearch, crossrefCounter)
		searchedFileCounter++
	}
	prt.NoStatus()
	prt.Infof("Saved references to file %v", outFilePath)
}
