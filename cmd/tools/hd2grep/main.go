package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

func printUsage() {
	fmt.Println("hd2grep STRING GAME_FILES_TO_SEARCH_GLOB")
}

func main() {
	if len(os.Args) != 3 {
		printUsage()
		os.Exit(1)
	}
	query, inclGlob := []byte(os.Args[1]), os.Args[2]

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
		prt.Fatalf("Command line option for installation path not implemented in hd2grep. Please open an issue on GitHub.")
	}

	ctx := context.Background() // no need to exit cleanly since we're only reading
	knownHashes := app.ParseHashes(hashes.Hashes)
	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, []string{}, stingray.ThinHash{}, func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}

	files, err := a.MatchingFiles(inclGlob, "", nil, nil, "")
	if err != nil {
		prt.Fatalf("Error matching files: %v", err)
	}

	numSearchedFiles := 0
	for id := range files {
		prt.Statusf("%v/%v", numSearchedFiles, len(files))
		name := a.LookupHash(id.Name) + "." + a.LookupHash(id.Type)
		for dataType := range stingray.NumDataType {
			bs, err := a.DataDir.Read(id, dataType)
			if err == stingray.ErrFileDataTypeNotExist {
				continue
			}
			if err != nil {
				prt.Errorf("Error reading file: %v", err)
			}

			currentOffset := 0
			var offsets []int
			for {
				idx := bytes.Index(bs[currentOffset:], query)
				if idx == -1 {
					break
				} else {
					offsets = append(offsets, currentOffset+idx)
					currentOffset += idx + len(query)
				}
			}
			if len(offsets) > 0 {
				prt.Infof("Match in %v (offsets: %v)", name, offsets)
			}
		}
		numSearchedFiles++
	}
}
