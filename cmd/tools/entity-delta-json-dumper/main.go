package main

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

type SimpleEntityDeltaSettings struct {
	ModifiedComponents map[string][]datalib.ComponentModificationDelta `json:"modified_components,omitempty"`
}

type SimpleComponentEntityDeltaStorage map[string]SimpleEntityDeltaSettings

func main() {
	knownHashes := app.ParseHashes(hashes.Hashes)
	knownDatalibHashes := app.ParseHashes(hashes.DLTypeNames)

	hashesMap := make(map[stingray.Hash]string)
	for _, h := range knownHashes {
		hashesMap[stingray.Sum(h)] = h
	}

	lookupHash := func(hash stingray.Hash) string {
		if name, ok := hashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	datalibHashesMap := make(map[datalib.DLHash]string)
	for _, h := range knownDatalibHashes {
		datalibHashesMap[datalib.Sum(h)] = h
	}
	lookupDLHash := func(hash datalib.DLHash) string {
		if name, ok := datalibHashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	componentIndicesToHashes, err := datalib.ParseComponentIndices()
	if err != nil {
		panic(err)
	}

	entityDeltas, err := datalib.ParseEntityDeltas()
	if err != nil {
		panic(err)
	}

	simpleEntityDeltas := make(SimpleComponentEntityDeltaStorage)
	for name, deltaSettings := range entityDeltas {
		simpleModifiedComponents := make(map[string][]datalib.ComponentModificationDelta)
		for _, modifiedComponent := range deltaSettings.ModifiedComponents {
			componentHash := componentIndicesToHashes[modifiedComponent.ComponentIndex]
			simpleModifiedComponents[lookupDLHash(componentHash)] = modifiedComponent.Deltas
		}
		simpleDeltaSettings := SimpleEntityDeltaSettings{
			ModifiedComponents: simpleModifiedComponents,
		}
		simpleEntityDeltas[lookupHash(name)] = simpleDeltaSettings
	}

	output, err := json.MarshalIndent(simpleEntityDeltas, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
