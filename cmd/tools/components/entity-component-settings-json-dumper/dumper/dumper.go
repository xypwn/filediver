package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
)

func Dump(a components.HashLookup, lookupDLHash func(hash datalib.DLHash) string) {
	entityHashmap, err := datalib.ParseEntityComponentSettings()
	if err != nil {
		panic(err)
	}

	result := make(map[string]datalib.SimpleEntity)
	for name, entity := range entityHashmap {
		result[a.LookupHash(name)] = entity.ToSimple(a.LookupHash, a.LookupThinHash, lookupDLHash, a.LookupString)
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
