package main

import (
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

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

	visibilityMasks, err := datalib.ParseVisibilityMasks()
	if err != nil {
		panic(err)
	}

	for _, component := range visibilityMasks {
		for _, info := range component.MaskInfos {
			if info.Name.Value == 0 {
				break
			}
			fmt.Println(lookupThinHash(info.Name))
		}
	}
}
