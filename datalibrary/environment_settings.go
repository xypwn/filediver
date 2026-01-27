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

type rawLevelGenerationPathSettings struct {
	MeshMaterials         [4]stingray.Hash
	MeshMaterialOpacities [4]float32
	SplineType            enum.StampSplineType
	CollisionWidth        float32
	MeshWidth             float32
	MeshTilingLength      float32
	CreatesPathCollision  uint8
	_                     [7]uint8
}

type LevelGenerationPathSettings struct {
	MeshMaterials         []string             `json:"mesh_materials"`
	MeshMaterialOpacities []float32            `json:"mesh_material_opacities"`
	SplineType            enum.StampSplineType `json:"spline_type"`
	CollisionWidth        float32              `json:"collision_width"`
	MeshWidth             float32              `json:"mesh_width"`
	MeshTilingLength      float32              `json:"mesh_tiling_length"`
	CreatesPathCollision  bool                 `json:"creates_path_collision"`
}

func (l rawLevelGenerationPathSettings) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) LevelGenerationPathSettings {
	meshMaterials := make([]string, 0)
	meshMaterialOpacities := make([]float32, 0)
	for i, material := range l.MeshMaterials {
		if material.Value == 0 && l.MeshMaterialOpacities[i] == 0 {
			break
		}
		meshMaterials = append(meshMaterials, lookupHash(material))
		meshMaterialOpacities = append(meshMaterialOpacities, l.MeshMaterialOpacities[i])
	}

	return LevelGenerationPathSettings{
		MeshMaterials:         meshMaterials,
		MeshMaterialOpacities: meshMaterialOpacities,
		SplineType:            l.SplineType,
		CollisionWidth:        l.CollisionWidth,
		MeshWidth:             l.MeshWidth,
		MeshTilingLength:      l.MeshTilingLength,
		CreatesPathCollision:  l.CreatesPathCollision != 0,
	}
}

type rawEnvironmentSettings struct {
	PlanetType                 enum.PlanetType
	_                          [4]uint8
	DebugNameOffset            int64
	NameLoc                    uint32
	_                          [4]uint8
	NameLocStr                 int64
	LoadoutIntelligenceLoc     uint32
	_                          [4]uint8
	WwiseStateStrOffset        int64
	PlanetMaterialId           stingray.Hash
	EnvSharedPackage           stingray.Hash
	TerrainProjectorPackage    stingray.Hash
	ShadingEnvironment         stingray.Hash
	GradingDay                 stingray.Hash
	GradingSunset              stingray.Hash
	GradingNight               stingray.Hash
	MinimapPackage             stingray.Hash
	MinimapUtilityLevel        stingray.Hash
	DirtColor                  mgl32.Vec3
	LayerId                    stingray.ThinHash
	HologramLightColor         mgl32.Vec3
	HologramOverlayColor       mgl32.Vec3
	DropSelectRouteColor       mgl32.Vec3
	_                          [4]uint8
	HologramMinimapLut         stingray.Hash
	PathSettings               [5]rawLevelGenerationPathSettings
	UtilityLevel               stingray.Hash
	ResourcePackages           DLArray
	ResourceOverrideTags       uint32
	_                          [4]uint8
	TerrainMaterialPath        stingray.Hash
	DefaultReverbZoneStrOffset int64
	DefaultAmbienceSoundId     uint32
	WindWwiseStartEvent        uint32
	WindWwiseStopEvent         uint32
	UnkEvent                   uint32 // Name length 32
	UnkEvent2                  uint32 // Name length 31
	_                          [4]uint8
	UnkStr                     int64 // Name length 23
	UnkFloat                   float32
	_                          [4]uint8
}

type itmEnvironmentSettings struct {
	PlanetType              enum.PlanetType
	DebugName               string
	NameLoc                 uint32
	NameLocStr              string
	LoadoutIntelligenceLoc  uint32
	WwiseStateStr           string
	PlanetMaterialId        stingray.Hash
	EnvSharedPackage        stingray.Hash
	TerrainProjectorPackage stingray.Hash
	ShadingEnvironment      stingray.Hash
	GradingDay              stingray.Hash
	GradingSunset           stingray.Hash
	GradingNight            stingray.Hash
	MinimapPackage          stingray.Hash
	MinimapUtilityLevel     stingray.Hash
	DirtColor               mgl32.Vec3
	LayerId                 stingray.ThinHash
	HologramLightColor      mgl32.Vec3
	HologramOverlayColor    mgl32.Vec3
	DropSelectRouteColor    mgl32.Vec3
	HologramMinimapLut      stingray.Hash
	PathSettings            [5]rawLevelGenerationPathSettings
	UtilityLevel            stingray.Hash
	ResourcePackages        []stingray.Hash
	ResourceOverrideTags    uint32
	TerrainMaterialPath     stingray.Hash
	DefaultReverbZoneStr    string
	DefaultAmbienceSoundId  uint32
	WindWwiseStartEvent     uint32
	WindWwiseStopEvent      uint32
	UnkEvent                uint32  // Name length 32
	UnkEvent2               uint32  // Name length 31
	UnkStr                  string  // Name length 23
	UnkFloat                float32 // Name length 29
}

type EnvironmentSettings struct {
	PlanetType              enum.PlanetType               `json:"planet_type"`
	DebugName               string                        `json:"debug_name"`
	NameLoc                 string                        `json:"name_loc"`
	NameLocStr              string                        `json:"name_loc_str"`
	LoadoutIntelligenceLoc  string                        `json:"loadout_intelligence_loc"`
	WwiseStateStr           string                        `json:"wwise_state"`
	PlanetMaterialId        string                        `json:"planet_material_id"`
	EnvSharedPackage        string                        `json:"env_shared_package"`
	TerrainProjectorPackage string                        `json:"terrain_projector_package"`
	ShadingEnvironment      string                        `json:"shading_environment"`
	GradingDay              string                        `json:"grading_day"`
	GradingSunset           string                        `json:"grading_sunset"`
	GradingNight            string                        `json:"grading_night"`
	MinimapPackage          string                        `json:"minimap_package"`
	MinimapUtilityLevel     string                        `json:"minimap_utility_level"`
	DirtColor               mgl32.Vec3                    `json:"dirt_color"`
	LayerId                 string                        `json:"layer_id"`
	HologramLightColor      mgl32.Vec3                    `json:"hologram_light_color"`
	HologramOverlayColor    mgl32.Vec3                    `json:"hologram_overlay_color"`
	DropSelectRouteColor    mgl32.Vec3                    `json:"drop_select_route_color"`
	HologramMinimapLut      string                        `json:"hologram_minimap_lut"`
	PathSettings            []LevelGenerationPathSettings `json:"path_settings"`
	UtilityLevel            string                        `json:"utility_level"`
	ResourcePackages        []string                      `json:"resource_packages"`
	ResourceOverrideTags    uint32                        `json:"resource_override_tags"`
	TerrainMaterialPath     string                        `json:"terrain_material_path"`
	DefaultReverbZoneStr    string                        `json:"default_reverb_zone"`
	DefaultAmbienceSoundId  uint32                        `json:"default_ambience_sound_id"`
	WindWwiseStartEvent     uint32                        `json:"wind_wwise_start_event"`
	WindWwiseStopEvent      uint32                        `json:"wind_wwise_stop_event"`
	UnkEvent                uint32                        `json:"unk_event"`  // Name length 32
	UnkEvent2               uint32                        `json:"unk_event2"` // Name length 31
	UnkStr                  string                        `json:"unk_str"`    // Name length 23
	UnkFloat                float32                       `json:"unk_float"`  // Name length 29
}

func (a itmEnvironmentSettings) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) EnvironmentSettings {
	pathSettings := make([]LevelGenerationPathSettings, 0)
	for _, path := range a.PathSettings {
		pathSettings = append(pathSettings, path.Resolve(lookupHash, lookupThinHash, lookupStrings))
	}

	resourcePackages := make([]string, 0)
	for _, resourcePackage := range a.ResourcePackages {
		resourcePackages = append(resourcePackages, lookupHash(resourcePackage))
	}
	return EnvironmentSettings{
		PlanetType:              a.PlanetType,
		DebugName:               a.DebugName,
		NameLoc:                 lookupStrings(a.NameLoc),
		NameLocStr:              a.NameLocStr,
		LoadoutIntelligenceLoc:  lookupStrings(a.LoadoutIntelligenceLoc),
		WwiseStateStr:           a.WwiseStateStr,
		PlanetMaterialId:        lookupHash(a.PlanetMaterialId),
		EnvSharedPackage:        lookupHash(a.EnvSharedPackage),
		TerrainProjectorPackage: lookupHash(a.TerrainProjectorPackage),
		ShadingEnvironment:      lookupHash(a.ShadingEnvironment),
		GradingDay:              lookupHash(a.GradingDay),
		GradingSunset:           lookupHash(a.GradingSunset),
		GradingNight:            lookupHash(a.GradingNight),
		MinimapPackage:          lookupHash(a.MinimapPackage),
		MinimapUtilityLevel:     lookupHash(a.MinimapUtilityLevel),
		DirtColor:               a.DirtColor,
		LayerId:                 lookupThinHash(a.LayerId),
		HologramLightColor:      a.HologramLightColor,
		HologramOverlayColor:    a.HologramOverlayColor,
		DropSelectRouteColor:    a.DropSelectRouteColor,
		HologramMinimapLut:      lookupHash(a.HologramMinimapLut),
		PathSettings:            pathSettings,
		UtilityLevel:            lookupHash(a.UtilityLevel),
		ResourcePackages:        resourcePackages,
		ResourceOverrideTags:    a.ResourceOverrideTags,
		TerrainMaterialPath:     lookupHash(a.TerrainMaterialPath),
		DefaultReverbZoneStr:    a.DefaultReverbZoneStr,
		DefaultAmbienceSoundId:  a.DefaultAmbienceSoundId,
		WindWwiseStartEvent:     a.WindWwiseStartEvent,
		WindWwiseStopEvent:      a.WindWwiseStopEvent,
		UnkEvent:                a.UnkEvent,
		UnkEvent2:               a.UnkEvent2,
		UnkStr:                  a.UnkStr,
		UnkFloat:                a.UnkFloat,
	}
}

func LoadEnvironmentSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]EnvironmentSettings, error) {
	r := bytes.NewReader(environmentSettings)

	infos := make([]EnvironmentSettings, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("EnvironmentSettings") {
			return nil, fmt.Errorf("invalid projectile settings file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var rawSettings rawEnvironmentSettings
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("reading environment settings: %v", err)
		}

		intermediate := itmEnvironmentSettings{
			PlanetType:              rawSettings.PlanetType,
			NameLoc:                 rawSettings.NameLoc,
			LoadoutIntelligenceLoc:  rawSettings.LoadoutIntelligenceLoc,
			PlanetMaterialId:        rawSettings.PlanetMaterialId,
			EnvSharedPackage:        rawSettings.EnvSharedPackage,
			TerrainProjectorPackage: rawSettings.TerrainProjectorPackage,
			ShadingEnvironment:      rawSettings.ShadingEnvironment,
			GradingDay:              rawSettings.GradingDay,
			GradingSunset:           rawSettings.GradingSunset,
			GradingNight:            rawSettings.GradingNight,
			MinimapPackage:          rawSettings.MinimapPackage,
			MinimapUtilityLevel:     rawSettings.MinimapUtilityLevel,
			DirtColor:               rawSettings.DirtColor,
			LayerId:                 rawSettings.LayerId,
			HologramLightColor:      rawSettings.HologramLightColor,
			HologramOverlayColor:    rawSettings.HologramOverlayColor,
			DropSelectRouteColor:    rawSettings.DropSelectRouteColor,
			HologramMinimapLut:      rawSettings.HologramMinimapLut,
			PathSettings:            rawSettings.PathSettings,
			UtilityLevel:            rawSettings.UtilityLevel,
			ResourceOverrideTags:    rawSettings.ResourceOverrideTags,
			TerrainMaterialPath:     rawSettings.TerrainMaterialPath,
			DefaultAmbienceSoundId:  rawSettings.DefaultAmbienceSoundId,
			WindWwiseStartEvent:     rawSettings.WindWwiseStartEvent,
			WindWwiseStopEvent:      rawSettings.WindWwiseStopEvent,
			UnkEvent:                rawSettings.UnkEvent,
			UnkEvent2:               rawSettings.UnkEvent2,
			UnkFloat:                rawSettings.UnkFloat,
		}

		intermediate.ResourcePackages = make([]stingray.Hash, rawSettings.ResourcePackages.Count)
		r.Seek(base+rawSettings.ResourcePackages.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &intermediate.ResourcePackages); err != nil {
			return nil, fmt.Errorf("reading resource packages array: %v", err)
		}

		if rawSettings.DebugNameOffset > 0 {
			r.Seek(base+rawSettings.DebugNameOffset, io.SeekStart)
			debugName, err := util.ReadCString(r)
			if err != nil {
				return nil, err
			}
			if debugName != nil {
				intermediate.DebugName = *debugName
			}
		}

		if rawSettings.NameLocStr > 0 {
			r.Seek(base+rawSettings.NameLocStr, io.SeekStart)
			nameLocStr, err := util.ReadCString(r)
			if err != nil {
				return nil, err
			}
			if nameLocStr != nil {
				intermediate.NameLocStr = *nameLocStr
			}
		}

		if rawSettings.WwiseStateStrOffset > 0 {
			r.Seek(base+rawSettings.WwiseStateStrOffset, io.SeekStart)
			wwiseStateStr, err := util.ReadCString(r)
			if err != nil {
				return nil, err
			}
			if wwiseStateStr != nil {
				intermediate.WwiseStateStr = *wwiseStateStr
			}
		}

		if rawSettings.DefaultReverbZoneStrOffset > 0 {
			r.Seek(base+rawSettings.DefaultReverbZoneStrOffset, io.SeekStart)
			defaultReverbZoneStr, err := util.ReadCString(r)
			if err != nil {
				return nil, err
			}
			if defaultReverbZoneStr != nil {
				intermediate.DefaultReverbZoneStr = *defaultReverbZoneStr
			}
		}

		if rawSettings.UnkStr > 0 {
			r.Seek(base+rawSettings.UnkStr, io.SeekStart)
			unkStr, err := util.ReadCString(r)
			if err != nil {
				return nil, err
			}
			if unkStr != nil {
				intermediate.UnkStr = *unkStr
			}
		}

		infos = append(infos, intermediate.Resolve(lookupHash, lookupThinHash, lookupStrings))
		r.Seek(base+int64(header.Size), io.SeekStart)
	}
	return infos, nil
}
