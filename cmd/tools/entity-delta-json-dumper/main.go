package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

type SimpleComponentModificationDelta struct {
	Offset uint32 `json:"offset"`
	Data   string `json:"data,omitempty"`
}

type SimpleEntityDeltaSettings struct {
	ModifiedComponents map[string][]SimpleComponentModificationDelta `json:"modified_components,omitempty"`
}

type SimpleComponentEntityDeltaStorage struct {
	ModifiedTypes []string                             `json:"modified_types"`
	Deltas        map[string]SimpleEntityDeltaSettings `json:"deltas"`
}

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

	modifiedTypesSet := make(map[string]bool)
	modifiedTypesSlice := make([]string, 0)
	simpleEntityDeltas := make(map[string]SimpleEntityDeltaSettings)
	for name, deltaSettings := range entityDeltas {
		simpleModifiedComponents := make(map[string][]SimpleComponentModificationDelta)
		for _, modifiedComponent := range deltaSettings.ModifiedComponents {
			componentHash := componentIndicesToHashes[modifiedComponent.ComponentIndex]
			simpleDeltas := make([]SimpleComponentModificationDelta, 0)
			for _, delta := range modifiedComponent.Deltas {
				simpleDeltas = append(simpleDeltas, SimpleComponentModificationDelta{
					Offset: delta.Offset,
					Data:   hex.EncodeToString(delta.Data),
				})
			}
			simpleModifiedComponents[lookupDLHash(componentHash)] = simpleDeltas
			if _, ok := modifiedTypesSet[lookupDLHash(componentHash)]; !ok {
				modifiedTypesSet[lookupDLHash(componentHash)] = true
				modifiedTypesSlice = append(modifiedTypesSlice, lookupDLHash(componentHash))
			}
		}
		simpleDeltaSettings := SimpleEntityDeltaSettings{
			ModifiedComponents: simpleModifiedComponents,
		}
		simpleEntityDeltas[lookupHash(name)] = simpleDeltaSettings
	}
	slices.Sort(modifiedTypesSlice)

	result := SimpleComponentEntityDeltaStorage{
		ModifiedTypes: modifiedTypesSlice,
		Deltas:        simpleEntityDeltas,
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
