package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

type WeatherColorSet struct {
	ID                             uint32     `json:"id"`
	CloudColor                     mgl32.Vec3 `json:"cloud_color"`
	VistaCloudColor                mgl32.Vec3 `json:"vista_cloud_color"`
	FogColor                       mgl32.Vec3 `json:"fog_color"`
	WeatherEffectColor             mgl32.Vec3 `json:"weather_effect_color"`
	WeatherGroundFogVolumeColor    mgl32.Vec3 `json:"weather_ground_fog_volume_color"`
	NearGroundFogColor             mgl32.Vec3 `json:"near_ground_fog_color"`
	ZoneFogTint                    mgl32.Vec3 `json:"zone_fog_tint"`
	WeatherSunColorMult            mgl32.Vec3 `json:"weather_sun_color_mult"`
	WeatherSunIntensityMult        float32    `json:"weather_sun_intensity_mult"`
	ZoneFogMainColor               mgl32.Vec3 `json:"zone_fog_main_color"`
	ZoneFogSecondaryColor          mgl32.Vec3 `json:"zone_fog_secondary_color"`
	ZoneFogBugsColor               mgl32.Vec3 `json:"zone_fog_bugs_color"`
	ZoneFogBotsColor               mgl32.Vec3 `json:"zone_fog_bots_color"`
	ZoneFogIlluminateColor         mgl32.Vec3 `json:"zone_fog_illuminate_color"`
	AmbientDiffuseTint             mgl32.Vec3 `json:"ambient_diffuse_tint"`
	AmbientDiffuseGroundBounceTint mgl32.Vec3 `json:"ambient_diffuse_ground_bounce_tint"`
	FogAmbientTint                 mgl32.Vec3 `json:"fog_ambient_tint"`
	FogAmbientGroundBounceTint     mgl32.Vec3 `json:"fog_ambient_ground_bounce_tint"`
}

type LevelGenerationPaletteGroup struct {
	//Palette                 stingray.Hash
	AssetGrading            stingray.Hash
	SkySettingsGroup        stingray.Hash
	DayGrading              stingray.Hash
	NightGrading            stingray.Hash
	SunsetGrading           stingray.Hash
	WeatherColorSet         stingray.ThinHash
	WeatherColorSetInternal WeatherColorSet
}

type rawLevelGenerationRegion struct {
	NameOffset    int64
	ID            uint32
	Region        enum.LevelGenerationRegionType
	VarListPtr    int64 // Seems to be unused
	UnknownFloat1 float32
	UnknownFloat2 float32
	UnknownInt    uint32
	_             [4]uint8
}

type LevelGenerationRegion struct {
	Name          string                         `json:"name"`
	ID            uint32                         `json:"id"`
	Region        enum.LevelGenerationRegionType `json:"region"`
	VarListPtr    int64                          `json:"var_list_ptr"` // Seems to be unused
	UnknownFloat1 float32                        `json:"unknown_float1"`
	UnknownFloat2 float32                        `json:"unknown_float2"`
	UnknownInt    uint32                         `json:"unknown_int"`
}

type rawPlanetData struct {
	InheritsOffset                   int64
	PlanetNameLoc                    uint32
	PlanetDescriptionLoc             uint32
	PlanetDescriptionShortLoc        uint32
	PlanetSystemNameLoc              uint32
	PlanetLayoutId                   uint32
	_                                [4]uint8
	UnknownEnumOffset                int64
	UnknownEnumCount                 int64
	ResourceOverridesOffset          int64
	ResourceOverridesCount           int64
	DebugNameOffset                  int64
	RegionLowland                    rawLevelGenerationRegion
	RegionHighland                   rawLevelGenerationRegion
	PaletteGroupLowland              LevelGenerationPaletteGroup
	PaletteGroupHighland             LevelGenerationPaletteGroup
	ScenarioSettingsLowland          stingray.Hash
	ScenarioSettingsHighland         stingray.Hash
	GameplayModifiersOffset          int64
	GameplayModifiersCount           uint64
	PlanetType                       enum.PlanetType
	Unknown                          uint32
	NatureLocationTagOffset          uint64
	NatureLocationTagCount           uint64
	ScatterSettings                  stingray.ThinHash
	_                                [4]uint8
	MissionPlanetUnit                stingray.Hash
	MissionPlanetHologramUnit        stingray.Hash
	MissionPlanetUnitPackage         stingray.Hash
	MissionPlanetHologramUnitPackage stingray.Hash
	SolarSystemSettings              stingray.Hash
	SolarSystemIdSelectionOffset     int64
	SolarSystemIdSelectionCount      uint64
	SampleTypesOffset                int64
	SampleTypesCount                 uint64
	AmbienceSoundIdStart             uint32
	AmbienceSoundIdStop              uint32
	HologramPlanetMaterial           stingray.Hash
	PlanetPreviewImage               stingray.Hash
	ResourceRegionOverridesOffset    int64
	ResourceRegionOverridesCount     uint64
	ShadingEnvironmentEntity         stingray.Hash
	WaterEntity                      stingray.Hash
	UnknownHash                      stingray.Hash
	PackagePath                      stingray.Hash
	UnknownFloat                     float32
	_                                [4]uint8
}

type ResourceRegionOverride struct {
	ID         stingray.ThinHash
	RegionFlag enum.RegionFlag
}

type ResourceOverride struct {
	Type        stingray.Hash
	Replace     stingray.Hash
	ReplaceWith stingray.Hash
}

type PlanetData struct {
	Inherits                         string
	PlanetNameLoc                    string
	PlanetDescriptionLoc             string
	PlanetDescriptionShortLoc        string
	PlanetSystemNameLoc              string
	PlanetLayoutId                   uint32
	UnknownEnumArray                 []uint32
	ResourceOverrides                []ResourceOverride
	DebugName                        string
	RegionLowland                    LevelGenerationRegion
	RegionHighland                   LevelGenerationRegion
	PaletteGroupLowland              LevelGenerationPaletteGroup
	PaletteGroupHighland             LevelGenerationPaletteGroup
	ScenarioSettingsLowland          stingray.Hash
	ScenarioSettingsHighland         stingray.Hash
	GameplayModifiers                []stingray.Hash
	PlanetType                       enum.PlanetType
	Unknown                          uint32
	NatureLocationTags               []enum.NatureLocationTag
	ScatterSettings                  stingray.ThinHash
	MissionPlanetUnit                stingray.Hash
	MissionPlanetHologramUnit        stingray.Hash
	MissionPlanetUnitPackage         stingray.Hash
	MissionPlanetHologramUnitPackage stingray.Hash
	SolarSystemSettings              stingray.Hash
	SolarSystemIdSelections          []stingray.Hash
	SampleTypes                      []enum.SampleType
	AmbienceSoundIdStart             uint32
	AmbienceSoundIdStop              uint32
	HologramPlanetMaterial           stingray.Hash
	PlanetPreviewImage               stingray.Hash
	ResourceRegionOverrides          []ResourceRegionOverride
	ShadingEnvironmentEntity         stingray.Hash
	WaterEntity                      stingray.Hash
	UnknownHash                      stingray.Hash
	PackagePath                      stingray.Hash
	UnknownFloat                     float32
}

func (a rawPlanetData) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) PlanetData {
	return PlanetData{
		PlanetNameLoc:                    lookupStrings(a.PlanetNameLoc),
		PlanetDescriptionLoc:             lookupStrings(a.PlanetDescriptionLoc),
		PlanetDescriptionShortLoc:        lookupStrings(a.PlanetDescriptionShortLoc),
		PlanetSystemNameLoc:              lookupStrings(a.PlanetSystemNameLoc),
		PlanetLayoutId:                   a.PlanetLayoutId,
		PaletteGroupLowland:              a.PaletteGroupLowland,
		PaletteGroupHighland:             a.PaletteGroupHighland,
		ScenarioSettingsLowland:          a.ScenarioSettingsLowland,
		ScenarioSettingsHighland:         a.ScenarioSettingsHighland,
		PlanetType:                       a.PlanetType,
		Unknown:                          a.Unknown,
		ScatterSettings:                  a.ScatterSettings,
		MissionPlanetUnit:                a.MissionPlanetUnit,
		MissionPlanetHologramUnit:        a.MissionPlanetHologramUnit,
		MissionPlanetUnitPackage:         a.MissionPlanetUnitPackage,
		MissionPlanetHologramUnitPackage: a.MissionPlanetHologramUnitPackage,
		SolarSystemSettings:              a.SolarSystemSettings,
		AmbienceSoundIdStart:             a.AmbienceSoundIdStart,
		AmbienceSoundIdStop:              a.AmbienceSoundIdStop,
		HologramPlanetMaterial:           a.HologramPlanetMaterial,
		PlanetPreviewImage:               a.PlanetPreviewImage,
		ShadingEnvironmentEntity:         a.ShadingEnvironmentEntity,
		WaterEntity:                      a.WaterEntity,
		UnknownHash:                      a.UnknownHash,
		PackagePath:                      a.PackagePath,
		UnknownFloat:                     a.UnknownFloat,
	}
}

func LoadPlanetData(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]PlanetData, error) {
	r := bytes.NewReader(planetData)

	infos := make([]PlanetData, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("LevelGenerationPlanetData") {
			return nil, fmt.Errorf("invalid planet data file")
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding planet data base: %v", err)
		}

		var rawSetting rawPlanetData
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading planet data: %v", err)
		}

		setting := rawSetting.Resolve(lookupHash, lookupThinHash, lookupStrings)
		if _, err := r.Seek(int64(base+rawSetting.InheritsOffset), io.SeekStart); err != nil {
			return nil, err
		}
		inherits, err := util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		setting.Inherits = *inherits

		resourceOverrides := make([]ResourceOverride, rawSetting.ResourceOverridesCount)
		if rawSetting.ResourceOverridesOffset > 0 {
			if _, err := r.Seek(int64(base+rawSetting.ResourceOverridesOffset), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(r, binary.LittleEndian, &resourceOverrides); err != nil {
				return nil, fmt.Errorf("reading resource overrides: %v", err)
			}
		}
		setting.ResourceOverrides = resourceOverrides

		if _, err := r.Seek(int64(base+rawSetting.DebugNameOffset), io.SeekStart); err != nil {
			return nil, err
		}
		debugName, err := util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		setting.DebugName = *debugName

		if _, err := r.Seek(int64(base+rawSetting.RegionLowland.NameOffset), io.SeekStart); err != nil {
			return nil, err
		}
		name, err := util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		setting.RegionLowland = LevelGenerationRegion{
			Name:          *name,
			ID:            rawSetting.RegionLowland.ID,
			Region:        rawSetting.RegionLowland.Region,
			VarListPtr:    rawSetting.RegionLowland.VarListPtr,
			UnknownFloat1: rawSetting.RegionLowland.UnknownFloat1,
			UnknownFloat2: rawSetting.RegionLowland.UnknownFloat2,
			UnknownInt:    rawSetting.RegionLowland.UnknownInt,
		}
		if _, err := r.Seek(int64(base+rawSetting.RegionHighland.NameOffset), io.SeekStart); err != nil {
			return nil, err
		}
		name, err = util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		setting.RegionHighland = LevelGenerationRegion{
			Name:          *name,
			ID:            rawSetting.RegionHighland.ID,
			Region:        rawSetting.RegionHighland.Region,
			VarListPtr:    rawSetting.RegionHighland.VarListPtr,
			UnknownFloat1: rawSetting.RegionHighland.UnknownFloat1,
			UnknownFloat2: rawSetting.RegionHighland.UnknownFloat2,
			UnknownInt:    rawSetting.RegionHighland.UnknownInt,
		}

		if _, err := r.Seek(int64(base+rawSetting.GameplayModifiersOffset), io.SeekStart); err != nil {
			return nil, err
		}
		gameplayModifiers := make([]stingray.Hash, rawSetting.GameplayModifiersCount)
		if err := binary.Read(r, binary.LittleEndian, gameplayModifiers); err != nil {
			return nil, err
		}

		if _, err := r.Seek(int64(base+int64(rawSetting.NatureLocationTagOffset)), io.SeekStart); err != nil {
			return nil, err
		}
		natureLocationTags := make([]enum.NatureLocationTag, rawSetting.NatureLocationTagCount)
		if err := binary.Read(r, binary.LittleEndian, natureLocationTags); err != nil {
			return nil, err
		}

		if _, err := r.Seek(int64(base+rawSetting.SolarSystemIdSelectionOffset), io.SeekStart); err != nil {
			return nil, err
		}
		solarSystemIdSelections := make([]stingray.Hash, rawSetting.SolarSystemIdSelectionCount)
		if err := binary.Read(r, binary.LittleEndian, solarSystemIdSelections); err != nil {
			return nil, err
		}

		if _, err := r.Seek(int64(base+rawSetting.SampleTypesOffset), io.SeekStart); err != nil {
			return nil, err
		}
		sampleTypes := make([]enum.SampleType, rawSetting.SampleTypesCount)
		if err := binary.Read(r, binary.LittleEndian, sampleTypes); err != nil {
			return nil, err
		}

		if _, err := r.Seek(int64(base+rawSetting.ResourceRegionOverridesOffset), io.SeekStart); err != nil {
			return nil, err
		}
		resourceRegionOverrides := make([]ResourceRegionOverride, rawSetting.ResourceRegionOverridesCount)
		if err := binary.Read(r, binary.LittleEndian, resourceRegionOverrides); err != nil {
			return nil, err
		}
		setting.GameplayModifiers = gameplayModifiers
		setting.NatureLocationTags = natureLocationTags
		setting.SolarSystemIdSelections = solarSystemIdSelections
		setting.SampleTypes = sampleTypes
		setting.ResourceRegionOverrides = resourceRegionOverrides
		infos = append(infos, setting)
	}
	return infos, nil
}
