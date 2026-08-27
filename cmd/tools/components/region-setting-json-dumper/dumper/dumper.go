package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/datalibrary/enum"
)

func Dump(a components.HashLookup) {
	regionSettings, err := datalib.LoadRegionSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	regionGroups, err := datalib.LoadRegionGroups(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simpleSettings := make(map[enum.LevelGenerationRegionVariantType]datalib.SimpleGenerationRegionSettings, 0)
	for i, setting := range regionSettings {
		simpleSettings[enum.LevelGenerationRegionVariantType(i+1)] = setting.ToSimple(a.LookupHash, a.LookupThinHash, a.LookupString)
	}

	simpleGroups := make([]datalib.SimpleGenerationRegionGroup, 0)
	for _, group := range regionGroups {
		simpleGroups = append(simpleGroups, group.ToSimple(a.LookupHash, a.LookupThinHash, a.LookupString))
	}

	result := map[string]any{
		"region_settings": simpleSettings,
		"region_groups":   simpleGroups,
	}

	output, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
