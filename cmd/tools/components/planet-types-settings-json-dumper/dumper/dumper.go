package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/datalibrary/enum"
)

type SimpleSolarSystemObjectTypeSettings struct {
	UnitId                string                     `json:"unit_id"`
	MaterialId            string                     `json:"material_id"`
	RingUnitId            string                     `json:"ring_unit_id"`
	RingMaterialId        string                     `json:"ring_material_id"`
	RingScaleMult         float32                    `json:"ring_scale_mult"`
	BillboardUnitId       string                     `json:"billboard_unit_id"`
	BillboardMaterialId   string                     `json:"billboard_material_id"`
	BillboardScaleMult    float32                    `json:"billboard_scale_mult"`
	SurfaceColor          mgl32.Vec3                 `json:"surface_color"`
	EmissiveColor         mgl32.Vec3                 `json:"emissive_color"`
	EmissiveIntensity     float32                    `json:"emissive_intensity"`
	RingSortingType       enum.RingSortingTypes      `json:"ring_sorting_type"`
	SolarSystemPlanetType enum.SolarSystemPlanetType `json:"solar_system_planet_type"`
}

type SimpleSolarSystemObjectType struct {
	Type           string                                `json:"type"`
	PlanetSettings []SimpleSolarSystemObjectTypeSettings `json:"planet_settings"`
}

type SimpleSolarSystemObjectTypes struct {
	PlanetTypes []SimpleSolarSystemObjectType `json:"planet_types"`
}

func Dump(a components.HashLookup) {
	planetTypesSettingsArray, err := datalib.LoadPlanetTypesSettings(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simplePlanetTypesSettingsArray := make([]SimpleSolarSystemObjectTypes, 0)

	for _, planetTypesSettings := range planetTypesSettingsArray {
		planetTypes := make([]SimpleSolarSystemObjectType, 0)
		for _, planetType := range planetTypesSettings.PlanetTypes {
			planetSettings := make([]SimpleSolarSystemObjectTypeSettings, 0)
			for _, planetSetting := range planetType.PlanetSettings {
				planetSettings = append(planetSettings, SimpleSolarSystemObjectTypeSettings{
					UnitId:                a.LookupHash(planetSetting.UnitId),
					MaterialId:            a.LookupHash(planetSetting.MaterialId),
					RingUnitId:            a.LookupHash(planetSetting.RingUnitId),
					RingMaterialId:        a.LookupHash(planetSetting.RingMaterialId),
					RingScaleMult:         planetSetting.RingScaleMult,
					BillboardUnitId:       a.LookupHash(planetSetting.BillboardUnitId),
					BillboardMaterialId:   a.LookupHash(planetSetting.BillboardMaterialId),
					BillboardScaleMult:    planetSetting.BillboardScaleMult,
					SurfaceColor:          planetSetting.SurfaceColor,
					EmissiveColor:         planetSetting.EmissiveColor,
					EmissiveIntensity:     planetSetting.EmissiveIntensity,
					RingSortingType:       planetSetting.RingSortingType,
					SolarSystemPlanetType: planetSetting.SolarSystemPlanetType,
				})
			}
			planetTypes = append(planetTypes, SimpleSolarSystemObjectType{
				Type:           a.LookupHash(planetType.Type),
				PlanetSettings: planetSettings,
			})
		}

		simplePlanetTypesSettingsArray = append(simplePlanetTypesSettingsArray, SimpleSolarSystemObjectTypes{
			PlanetTypes: planetTypes,
		})
	}

	output, err := json.MarshalIndent(simplePlanetTypesSettingsArray[0], "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
