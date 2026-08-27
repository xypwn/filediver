package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/datalibrary/enum"
)

type SimplePlanetRegionSettings struct {
	Template                 string              `json:"template"`
	UnkThinHash0             string              `json:"unk_thin_hash0"`
	UnkThinHash1             string              `json:"unk_thin_hash1"`
	RegionPreviewTexturePath string              `json:"region_preview_texture_path"`
	RegionRaceType           enum.RaceType       `json:"region_race_type"`
	OperationTags            []enum.OperationTag `json:"operation_tags"`
	UnkInt0                  uint32              `json:"unk_int0"`
	UnkInt1                  uint32              `json:"unk_int1"`
	RegionOperationType      enum.OperationType  `json:"region_operation_type"`
	UnkThinHashes0           []string            `json:"unk_thin_hashes0"`
	UnkEnumType              uint32              `json:"unk_enum_type"`
	UnkThinHashes1           []string            `json:"unk_thin_hashes1"`
	UnkEnumItems             []uint32            `json:"unk_enum_items"`
	UnkInt2                  uint32              `json:"unk_int2"`
}

func Dump(a components.HashLookup) {
	planetRegionSettingsMap, err := datalib.LoadPlanetRegionSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simplePlanetRegionSettingsMap := make(map[string]SimplePlanetRegionSettings)

	for key, planetTypesSettings := range planetRegionSettingsMap {
		unkThinHashes0 := make([]string, 0)
		for _, item := range planetTypesSettings.UnkThinHashes0 {
			unkThinHashes0 = append(unkThinHashes0, a.LookupThinHash(item))
		}
		unkThinHashes1 := make([]string, 0)
		for _, item := range planetTypesSettings.UnkThinHashes1 {
			unkThinHashes1 = append(unkThinHashes1, a.LookupThinHash(item))
		}
		simplePlanetRegionSettingsMap[a.LookupHash(key)] = SimplePlanetRegionSettings{
			Template:                 planetTypesSettings.Template,
			UnkThinHash0:             a.LookupThinHash(planetTypesSettings.UnkThinHash0),
			UnkThinHash1:             a.LookupThinHash(planetTypesSettings.UnkThinHash1),
			RegionPreviewTexturePath: a.LookupHash(planetTypesSettings.RegionPreviewTexturePath),
			RegionRaceType:           planetTypesSettings.RegionRaceType,
			OperationTags:            planetTypesSettings.OperationTags,
			UnkInt0:                  planetTypesSettings.UnkInt0,
			UnkInt1:                  planetTypesSettings.UnkInt1,
			RegionOperationType:      planetTypesSettings.RegionOperationType,
			UnkThinHashes0:           unkThinHashes0,
			UnkEnumType:              planetTypesSettings.UnkEnumType,
			UnkThinHashes1:           unkThinHashes1,
			UnkEnumItems:             planetTypesSettings.UnkEnumItems,
			UnkInt2:                  planetTypesSettings.UnkInt2,
		}
	}

	output, err := json.MarshalIndent(simplePlanetRegionSettingsMap, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
