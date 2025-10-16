package main

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

func main() {
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

	entityHashmap, err := datalib.ParseEntityComponentSettings()
	if err != nil {
		panic(err)
	}

	result := make(map[string]datalib.SimpleEntity)
	for name, entity := range entityHashmap {
		result[lookupHash(name)] = entity.ToSimple(lookupHash, lookupThinHash, lookupDLHash)
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
