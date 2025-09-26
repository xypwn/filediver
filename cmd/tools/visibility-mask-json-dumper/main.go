package main

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

type SimpleVisibilityMaskInfo struct {
	Name        string `json:"name"`
	Index       uint16 `json:"index"`
	StartHidden bool   `json:"default_hidden"`
}

type SimpleVisibilityRandomization struct {
	Identifier     string   `json:"id"`
	MaskIndexNames []string `json:"mask_index_names,omitempty"`
}

type SimpleVisibilityMaskComponent struct {
	MaskInfos      []SimpleVisibilityMaskInfo      `json:"mask_infos,omitempty"`
	Randomizations []SimpleVisibilityRandomization `json:"randomizations,omitempty"`
}

func main() {
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range knownThinHashes {
		thinHashesMap[stingray.Sum(h).Thin()] = h
	}

	lookupThinHash := func(hash stingray.ThinHash) string {
		if name, ok := thinHashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	knownHashes := app.ParseHashes(hashes.Hashes)

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

	visibilityMasks, err := datalib.ParseVisibilityMasks()
	if err != nil {
		panic(err)
	}

	result := make(map[string]SimpleVisibilityMaskComponent)
	for name, component := range visibilityMasks {
		simpleCmp := SimpleVisibilityMaskComponent{
			MaskInfos:      make([]SimpleVisibilityMaskInfo, 0),
			Randomizations: make([]SimpleVisibilityRandomization, 0),
		}
		for _, info := range component.MaskInfos {
			if info.Name.Value == 0 {
				break
			}
			simpleCmp.MaskInfos = append(simpleCmp.MaskInfos, SimpleVisibilityMaskInfo{
				Name:        lookupThinHash(info.Name),
				Index:       info.Index,
				StartHidden: info.StartHidden != 0,
			})
		}
		for _, rand := range component.Randomizations {
			if rand.Identifier.Value == 0 {
				break
			}
			maskIndexNames := make([]string, 0)
			for _, maskIndexName := range rand.MaskIndexNames {
				if maskIndexName.Value == 0 {
					break
				}
				maskIndexNames = append(maskIndexNames, lookupThinHash(maskIndexName))
			}
			simpleCmp.Randomizations = append(simpleCmp.Randomizations, SimpleVisibilityRandomization{
				Identifier:     lookupThinHash(rand.Identifier),
				MaskIndexNames: maskIndexNames,
			})
		}
		result[lookupHash(name)] = simpleCmp
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
