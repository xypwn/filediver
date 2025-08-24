package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/strings"
)

func main() {
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
		prt.Fatalf("Command line option for installation path not implemented. Please open an issue on GitHub.")
	}

	ctx := context.Background() // no need to exit cleanly since we're only reading
	knownHashes := app.ParseHashes(hashes.Hashes)
	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, []string{}, stingray.Hash{}, func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}

	files, err := a.MatchingFiles("", "", nil, nil)
	if err != nil {
		prt.Fatalf("Error matching files: %v", err)
	}

	langs := make(map[stingray.ThinHash]struct{})
	for id, file := range files {
		if id.Type != stingray.Sum("strings") {
			continue
		}
		r, err := file.Open(ctx, stingray.DataMain)
		if err != nil {
			prt.Errorf("Error opening file: %v", err)
		}
		strings, err := strings.Load(r)
		if err != nil {
			prt.Errorf("Error reading %v.strings: %v", a.LookupHash(id.Name), err)
		}
		langs[strings.Language] = struct{}{}
	}
	for _, lang := range slices.SortedFunc(maps.Keys(langs), stingray.ThinHash.Cmp) {
		fmt.Println(lang)
	}
}
