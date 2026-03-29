package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
)

type SimpleSkySettings struct {
	ID       string               `json:"id"`
	Settings []datalib.SkySetting `json:"sky_settings"`
}

func Dump(a components.HashLookup) {
	skySettingsArray, err := datalib.LoadSkySettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simpleSkySettingsArray := make([]SimpleSkySettings, 0)

	for _, skySetting := range skySettingsArray {
		simpleSkySettingsArray = append(simpleSkySettingsArray, SimpleSkySettings{
			ID:       a.LookupHash(skySetting.ID),
			Settings: skySetting.Settings,
		})
	}

	output, err := json.MarshalIndent(simpleSkySettingsArray, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
