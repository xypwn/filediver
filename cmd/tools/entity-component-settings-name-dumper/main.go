package main

import (
	"fmt"

	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

func main() {
	knownHashes := app.ParseHashes(hashes.Hashes)

	hashesMap := make(map[stingray.Hash]string)
	for _, name := range knownHashes {
		hashesMap[stingray.Sum(name)] = name
	}

	lookupHash := func(hash stingray.Hash) string {
		if name, ok := hashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	entityHashmap, err := datalib.ParseEntityComponentSettings()
	if err != nil {
		panic(err)
	}

	for name := range entityHashmap {
		fmt.Println(lookupHash(name))
	}
}
