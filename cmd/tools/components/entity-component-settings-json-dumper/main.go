package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)
	knownDLHashes := app.ParseHashes(hashes.DLTypeNames)

	hashesMap := make(map[stingray.Hash]string)
	for _, name := range knownHashes {
		hashesMap[stingray.Sum(name)] = name
	}

	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, name := range knownThinHashes {
		thinHashesMap[stingray.Sum(name).Thin()] = name
	}

	dlHashesMap := make(map[datalib.DLHash]string)
	for _, name := range knownDLHashes {
		dlHashesMap[datalib.Sum(name)] = name
	}

	ctx := context.Background()

	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Helldivers 2 Steam installation path not found: %v", err)
	}

	dataDir, err := stingray.OpenDataDir(ctx, filepath.Join(gameDir, "data"), func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		prt.Fatalf("Could not open data dir: %v", err)
	}
	mapping := stingray_strings.LoadLanguageMap(dataDir, stingray_strings.LanguageFriendlyNameToHash["English (US)"])

	lookupHash := func(hash stingray.Hash) string {
		if name, ok := hashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	lookupThinHash := func(hash stingray.ThinHash) string {
		if name, ok := thinHashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	lookupDLHash := func(hash datalib.DLHash) string {
		if name, ok := dlHashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	lookupString := func(stringId uint32) string {
		if name, ok := mapping[stringId]; ok {
			return name
		}
		return fmt.Sprintf("String ID not found: %v", stringId)
	}

	entityHashmap, err := datalib.ParseEntityComponentSettings()
	if err != nil {
		panic(err)
	}

	result := make(map[string]datalib.SimpleEntity)
	for name, entity := range entityHashmap {
		result[lookupHash(name)] = entity.ToSimple(lookupHash, lookupThinHash, lookupDLHash, lookupString)
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
