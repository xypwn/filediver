package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/datalibrary/enum"
)

func Dump(a components.HashLookup) {
	zoneSettings, err := datalib.LoadZoneSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simple := make(map[enum.ZoneId]datalib.SimpleZoneSettings, 0)
	for i, zone := range zoneSettings {
		simple[enum.ZoneId(i+1)] = zone.ToSimple(a.LookupHash, a.LookupThinHash, a.LookupString)
	}

	output, err := json.MarshalIndent(simple, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
