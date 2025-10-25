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

	lookupString := func(val uint32) string {
		return fmt.Sprintf("%x", val)
	}

	_, components, err := datalib.ParseHealthComponentsArray()
	if err != nil {
		panic(err)
	}

	fmt.Println("[")
	for _, component := range components {
		output, err := json.MarshalIndent(component.ToSimple(lookupHash, lookupThinHash, lookupString), "    ", "    ")
		if err != nil {
			fmt.Printf("    \"Error: %v\",\n", err)
			continue
		}
		fmt.Printf("    %v,\n", string(output))
	}
	fmt.Println("]")

	// output, err := json.MarshalIndent(result, "", "    ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Print(string(output))
}
