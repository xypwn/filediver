package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
)

func Dump(a components.HashLookup) {
	zoneSettings, err := datalib.LoadZoneSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simple := make([]datalib.SimpleZoneSettings, 0)
	for _, zone := range zoneSettings {
		simple = append(simple, zone.ToSimple(a.LookupHash, a.LookupThinHash, a.LookupString))
	}

	output, err := json.MarshalIndent(simple, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
