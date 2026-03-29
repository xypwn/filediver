package dumper

import (
	"encoding/json"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/components"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/datalibrary/enum"
)

type SimpleResourceRegionOverride struct {
	ID         string          `json:"id"`
	RegionFlag enum.RegionFlag `json:"region_flag"`
}

type SimpleLevelGenerationPaletteGroup struct {
	Palette                 string                  `json:"palette"`
	AssetGrading            string                  `json:"asset_grading"`
	SkySettingsGroup        string                  `json:"sky_settings_group"`
	DayGrading              string                  `json:"day_grading"`
	NightGrading            string                  `json:"night_grading"`
	SunsetGrading           string                  `json:"sunset_grading"`
	WeatherColorSet         string                  `json:"weather_color_set"`
	WeatherColorSetInternal datalib.WeatherColorSet `json:"weather_color_set_internal"`
}

type SimplePlanetData struct {
	Inherits                         string                            `json:"inherits"`
	PlanetNameLoc                    string                            `json:"planet_name_loc"`
	PlanetDescriptionLoc             string                            `json:"planet_description_loc"`
	PlanetDescriptionShortLoc        string                            `json:"planet_description_short_loc"`
	PlanetSystemNameLoc              string                            `json:"planet_system_name_loc"`
	PlanetLayoutId                   uint32                            `json:"planet_layout_id"`
	DebugName                        string                            `json:"debug_name"`
	RegionLowland                    datalib.LevelGenerationRegion     `json:"region_lowland"`
	RegionHighland                   datalib.LevelGenerationRegion     `json:"region_highland"`
	PaletteGroupLowland              SimpleLevelGenerationPaletteGroup `json:"palette_group_lowland"`
	PaletteGroupHighland             SimpleLevelGenerationPaletteGroup `json:"palette_group_highland"`
	ScenarioSettingsLowland          string                            `json:"scenario_settings_lowland"`
	ScenarioSettingsHighland         string                            `json:"scenario_settings_highland"`
	GameplayModifiers                []string                          `json:"gameplay_modifiers"`
	PlanetType                       enum.PlanetType                   `json:"planet_type"`
	Unknown                          uint32                            `json:"unknown"`
	NatureLocationTags               []enum.NatureLocationTag          `json:"nature_location_tags"`
	ScatterSettings                  string                            `json:"scatter_settings"`
	MissionPlanetUnit                string                            `json:"mission_planet_unit"`
	MissionPlanetHologramUnit        string                            `json:"mission_planet_hologram_unit"`
	MissionPlanetUnitPackage         string                            `json:"mission_planet_unit_package"`
	MissionPlanetHologramUnitPackage string                            `json:"mission_planet_hologram_unit_package"`
	SolarSystemSettings              string                            `json:"solar_system_settings"`
	SolarSystemIdSelections          []string                          `json:"solar_system_id_selections"`
	SampleTypes                      []enum.SampleType                 `json:"sample_types"`
	AmbienceSoundIdStart             uint32                            `json:"ambience_sound_id_start"`
	AmbienceSoundIdStop              uint32                            `json:"ambience_sound_id_stop"`
	HologramPlanetMaterial           string                            `json:"hologram_planet_material"`
	PlanetPreviewImage               string                            `json:"planet_preview_image"`
	ResourceRegionOverrides          []SimpleResourceRegionOverride    `json:"resource_region_overrides"`
	ShadingEnvironmentEntity         string                            `json:"shading_environment_entity"`
	WaterEntity                      string                            `json:"water_entity"`
	UnknownHash                      string                            `json:"unknown_hash"`
	PackagePath                      string                            `json:"package_path"`
	UnknownFloat                     float32                           `json:"unknown_float"`
}

func Dump(a components.HashLookup) {
	planetDataArray, err := datalib.LoadPlanetData(a.LookupHash, a.LookupThinHash, a.LookupString)
	if err != nil {
		panic(err)
	}

	simplePlanetDataArray := make([]SimplePlanetData, 0)

	for _, planetData := range planetDataArray {
		gameplayModifiers := make([]string, 0)
		for _, modifier := range planetData.GameplayModifiers {
			gameplayModifiers = append(gameplayModifiers, a.LookupHash(modifier))
		}
		solarSystemIdSelections := make([]string, 0)
		for _, solarSystem := range planetData.SolarSystemIdSelections {
			solarSystemIdSelections = append(solarSystemIdSelections, a.LookupHash(solarSystem))
		}
		resourceRegionOverrides := make([]SimpleResourceRegionOverride, 0)
		for _, override := range planetData.ResourceRegionOverrides {
			resourceRegionOverrides = append(resourceRegionOverrides, SimpleResourceRegionOverride{
				ID:         a.LookupThinHash(override.ID),
				RegionFlag: override.RegionFlag,
			})
		}
		simplePlanetData := SimplePlanetData{
			Inherits:                  planetData.Inherits,
			PlanetNameLoc:             planetData.PlanetNameLoc,
			PlanetDescriptionLoc:      planetData.PlanetDescriptionLoc,
			PlanetDescriptionShortLoc: planetData.PlanetDescriptionShortLoc,
			PlanetSystemNameLoc:       planetData.PlanetSystemNameLoc,
			PlanetLayoutId:            planetData.PlanetLayoutId,
			DebugName:                 planetData.DebugName,
			RegionLowland:             planetData.RegionLowland,
			RegionHighland:            planetData.RegionHighland,

			PaletteGroupLowland: SimpleLevelGenerationPaletteGroup{
				Palette:                 a.LookupHash(planetData.PaletteGroupLowland.Palette),
				AssetGrading:            a.LookupHash(planetData.PaletteGroupLowland.AssetGrading),
				SkySettingsGroup:        a.LookupHash(planetData.PaletteGroupLowland.SkySettingsGroup),
				DayGrading:              a.LookupHash(planetData.PaletteGroupLowland.DayGrading),
				NightGrading:            a.LookupHash(planetData.PaletteGroupLowland.NightGrading),
				SunsetGrading:           a.LookupHash(planetData.PaletteGroupLowland.SunsetGrading),
				WeatherColorSet:         a.LookupThinHash(planetData.PaletteGroupLowland.WeatherColorSet),
				WeatherColorSetInternal: planetData.PaletteGroupLowland.WeatherColorSetInternal,
			},
			PaletteGroupHighland: SimpleLevelGenerationPaletteGroup{
				Palette:                 a.LookupHash(planetData.PaletteGroupHighland.Palette),
				AssetGrading:            a.LookupHash(planetData.PaletteGroupHighland.AssetGrading),
				SkySettingsGroup:        a.LookupHash(planetData.PaletteGroupHighland.SkySettingsGroup),
				DayGrading:              a.LookupHash(planetData.PaletteGroupHighland.DayGrading),
				NightGrading:            a.LookupHash(planetData.PaletteGroupHighland.NightGrading),
				SunsetGrading:           a.LookupHash(planetData.PaletteGroupHighland.SunsetGrading),
				WeatherColorSet:         a.LookupThinHash(planetData.PaletteGroupHighland.WeatherColorSet),
				WeatherColorSetInternal: planetData.PaletteGroupHighland.WeatherColorSetInternal,
			},
			ScenarioSettingsLowland:          a.LookupHash(planetData.ScenarioSettingsLowland),
			ScenarioSettingsHighland:         a.LookupHash(planetData.ScenarioSettingsHighland),
			GameplayModifiers:                gameplayModifiers,
			PlanetType:                       planetData.PlanetType,
			Unknown:                          planetData.Unknown,
			NatureLocationTags:               planetData.NatureLocationTags,
			ScatterSettings:                  a.LookupThinHash(planetData.ScatterSettings),
			MissionPlanetUnit:                a.LookupHash(planetData.MissionPlanetUnit),
			MissionPlanetHologramUnit:        a.LookupHash(planetData.MissionPlanetHologramUnit),
			MissionPlanetUnitPackage:         a.LookupHash(planetData.MissionPlanetUnitPackage),
			MissionPlanetHologramUnitPackage: a.LookupHash(planetData.MissionPlanetHologramUnitPackage),
			SolarSystemSettings:              a.LookupHash(planetData.SolarSystemSettings),
			SolarSystemIdSelections:          solarSystemIdSelections,
			SampleTypes:                      planetData.SampleTypes,
			AmbienceSoundIdStart:             planetData.AmbienceSoundIdStart,
			AmbienceSoundIdStop:              planetData.AmbienceSoundIdStop,
			HologramPlanetMaterial:           a.LookupHash(planetData.HologramPlanetMaterial),
			PlanetPreviewImage:               a.LookupHash(planetData.PlanetPreviewImage),
			ResourceRegionOverrides:          resourceRegionOverrides,
			ShadingEnvironmentEntity:         a.LookupHash(planetData.ShadingEnvironmentEntity),
			WaterEntity:                      a.LookupHash(planetData.WaterEntity),
			UnknownHash:                      a.LookupHash(planetData.UnknownHash),
			PackagePath:                      a.LookupHash(planetData.PackagePath),
			UnknownFloat:                     planetData.UnknownFloat,
		}
		simplePlanetDataArray = append(simplePlanetDataArray, simplePlanetData)
	}

	output, err := json.MarshalIndent(simplePlanetDataArray, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(output))
}
