package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

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
	UnkThinHash1               stingray.ThinHash
	PlanetPackage              stingray.Hash
	ShadingEnvironment         stingray.Hash
	GradingDay                 stingray.Hash
	GradingSunset              stingray.Hash
	GradingNight               stingray.Hash
	DirtColor                  mgl32.Vec3
	LayerId                    stingray.ThinHash
	HologramLightColor         mgl32.Vec3
	HologramOverlayColor       mgl32.Vec3
	DropSelectRouteColor       mgl32.Vec3
	_                          [4]uint8
	HologramMinimapLut         stingray.Hash
	PathSettings               [5]rawLevelGenerationPathSettings
	ResourcePackages           DLArray
	ResourceOverrideTags       uint32
	_                          [4]uint8
	TerrainMaterialPath        stingray.Hash
	DefaultReverbZoneStrOffset int64
	UnkThinHash2               stingray.ThinHash // Name length 13
	DefaultAmbienceSoundId     uint32
	WindWwiseStartEvent        uint32
	WindWwiseStopEvent         uint32
	UnkStr                     int64 // Name length 23
	UnkFloat                   float32
	_                          [4]uint8
}

type EnvironmentSettings struct {
	PlanetType             enum.PlanetType
	DebugName              string
	NameLoc                uint32
	NameLocStr             string
	LoadoutIntelligenceLoc uint32
	UnkThinHash1           stingray.ThinHash
	PlanetPackage          stingray.Hash
	ShadingEnvironment     stingray.Hash
	GradingDay             stingray.Hash
	GradingSunset          stingray.Hash
	GradingNight           stingray.Hash
	DirtColor              mgl32.Vec3
	LayerId                stingray.ThinHash
	HologramLightColor     mgl32.Vec3
	HologramOverlayColor   mgl32.Vec3
	DropSelectRouteColor   mgl32.Vec3
	HologramMinimapLut     stingray.Hash
	PathSettings           [5]rawLevelGenerationPathSettings
	ResourcePackages       []stingray.Hash
	ResourceOverrideTags   uint32
	TerrainMaterialPath    stingray.Hash
	DefaultReverbZoneStr   string
	UnkThinHash2           stingray.ThinHash // name length 13
	DefaultAmbienceSoundId uint32
	WindWwiseStartEvent    uint32
	WindWwiseStopEvent     uint32
	UnkStr                 string  // Name length 23
	UnkFloat               float32 // Name length 29
}

type SimpleEnvironmentSettings struct {
	PlanetType             enum.PlanetType               `json:"planet_type"`
	DebugName              string                        `json:"debug_name"`
	NameLoc                string                        `json:"name_loc"`
	NameLocStr             string                        `json:"name_loc_str"`
	LoadoutIntelligenceLoc string                        `json:"loadout_intelligence_loc"`
	UnkThinHash1           string                        `json:"unk_thinhash_1"`
	PlanetPackage          string                        `json:"planet_package"`
	ShadingEnvironment     string                        `json:"shading_environment"`
	GradingDay             string                        `json:"grading_day"`
	GradingSunset          string                        `json:"grading_sunset"`
	GradingNight           string                        `json:"grading_night"`
	DirtColor              mgl32.Vec3                    `json:"dirt_color"`
	LayerId                string                        `json:"layer_id"`
	HologramLightColor     mgl32.Vec3                    `json:"hologram_light_color"`
	HologramOverlayColor   mgl32.Vec3                    `json:"hologram_overlay_color"`
	DropSelectRouteColor   mgl32.Vec3                    `json:"drop_select_route_color"`
	HologramMinimapLut     string                        `json:"hologram_minimap_lut"`
	PathSettings           []LevelGenerationPathSettings `json:"path_settings"`
	ResourcePackages       []string                      `json:"resource_packages"`
	ResourceOverrideTags   uint32                        `json:"resource_override_tags"`
	TerrainMaterialPath    string                        `json:"terrain_material_path"`
	DefaultReverbZoneStr   string                        `json:"default_reverb_zone"`
	DefaultAmbienceSoundId uint32                        `json:"default_ambience_sound_id"`
	UnkThinHash2           string                        `json:"unk_thinhash_2"`
	WindWwiseStartEvent    uint32                        `json:"wind_wwise_start_event"`
	WindWwiseStopEvent     uint32                        `json:"wind_wwise_stop_event"`
	UnkStr                 string                        `json:"unk_str"`   // Name length 23
	UnkFloat               float32                       `json:"unk_float"` // Name length 29
}

func (a EnvironmentSettings) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleEnvironmentSettings {
	pathSettings := make([]LevelGenerationPathSettings, 0)
	for _, path := range a.PathSettings {
		pathSettings = append(pathSettings, path.Resolve(lookupHash, lookupThinHash, lookupStrings))
	}

	resourcePackages := make([]string, 0)
	for _, resourcePackage := range a.ResourcePackages {
		resourcePackages = append(resourcePackages, lookupHash(resourcePackage))
	}
	return SimpleEnvironmentSettings{
		PlanetType:             a.PlanetType,
		DebugName:              a.DebugName,
		NameLoc:                lookupStrings(a.NameLoc),
		NameLocStr:             a.NameLocStr,
		LoadoutIntelligenceLoc: lookupStrings(a.LoadoutIntelligenceLoc),
		UnkThinHash1:           lookupThinHash(a.UnkThinHash1),
		PlanetPackage:          lookupHash(a.PlanetPackage),
		ShadingEnvironment:     lookupHash(a.ShadingEnvironment),
		GradingDay:             lookupHash(a.GradingDay),
		GradingSunset:          lookupHash(a.GradingSunset),
		GradingNight:           lookupHash(a.GradingNight),
		DirtColor:              a.DirtColor,
		LayerId:                lookupThinHash(a.LayerId),
		HologramLightColor:     a.HologramLightColor,
		HologramOverlayColor:   a.HologramOverlayColor,
		DropSelectRouteColor:   a.DropSelectRouteColor,
		HologramMinimapLut:     lookupHash(a.HologramMinimapLut),
		PathSettings:           pathSettings,
		ResourcePackages:       resourcePackages,
		ResourceOverrideTags:   a.ResourceOverrideTags,
		TerrainMaterialPath:    lookupHash(a.TerrainMaterialPath),
		DefaultReverbZoneStr:   a.DefaultReverbZoneStr,
		UnkThinHash2:           lookupThinHash(a.UnkThinHash2),
		DefaultAmbienceSoundId: a.DefaultAmbienceSoundId,
		WindWwiseStartEvent:    a.WindWwiseStartEvent,
		WindWwiseStopEvent:     a.WindWwiseStopEvent,
		UnkStr:                 a.UnkStr,
		UnkFloat:               a.UnkFloat,
	}
}

func LoadEnvironmentSettings() ([]EnvironmentSettings, error) {
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

		setting := EnvironmentSettings{
			PlanetType:             rawSettings.PlanetType,
			NameLoc:                rawSettings.NameLoc,
			LoadoutIntelligenceLoc: rawSettings.LoadoutIntelligenceLoc,
			UnkThinHash1:           rawSettings.UnkThinHash1,
			PlanetPackage:          rawSettings.PlanetPackage,
			ShadingEnvironment:     rawSettings.ShadingEnvironment,
			GradingDay:             rawSettings.GradingDay,
			GradingSunset:          rawSettings.GradingSunset,
			GradingNight:           rawSettings.GradingNight,
			DirtColor:              rawSettings.DirtColor,
			LayerId:                rawSettings.LayerId,
			HologramLightColor:     rawSettings.HologramLightColor,
			HologramOverlayColor:   rawSettings.HologramOverlayColor,
			DropSelectRouteColor:   rawSettings.DropSelectRouteColor,
			HologramMinimapLut:     rawSettings.HologramMinimapLut,
			PathSettings:           rawSettings.PathSettings,
			ResourceOverrideTags:   rawSettings.ResourceOverrideTags,
			TerrainMaterialPath:    rawSettings.TerrainMaterialPath,
			DefaultAmbienceSoundId: rawSettings.DefaultAmbienceSoundId,
			UnkThinHash2:           rawSettings.UnkThinHash2,
			WindWwiseStartEvent:    rawSettings.WindWwiseStartEvent,
			WindWwiseStopEvent:     rawSettings.WindWwiseStopEvent,
			UnkFloat:               rawSettings.UnkFloat,
		}

		setting.ResourcePackages = make([]stingray.Hash, rawSettings.ResourcePackages.Count)
		r.Seek(base+rawSettings.ResourcePackages.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &setting.ResourcePackages); err != nil {
			return nil, fmt.Errorf("reading resource packages array: %v", err)
		}

		if rawSettings.DebugNameOffset > 0 {
			r.Seek(base+rawSettings.DebugNameOffset, io.SeekStart)
			debugName, err := util.ReadCString(r)
			if err != nil {
				return nil, fmt.Errorf("reading debug name: %v", err)
			}
			setting.DebugName = debugName
		}

		if rawSettings.NameLocStr > 0 {
			r.Seek(base+rawSettings.NameLocStr, io.SeekStart)
			nameLocStr, err := util.ReadCString(r)
			if err != nil {
				return nil, fmt.Errorf("reading localized name: %v", err)
			}
			setting.NameLocStr = nameLocStr
		}

		if rawSettings.DefaultReverbZoneStrOffset > 0 {
			r.Seek(base+rawSettings.DefaultReverbZoneStrOffset, io.SeekStart)
			defaultReverbZoneStr, err := util.ReadCString(r)
			if err != nil {
				//return nil, fmt.Errorf("reading default reverb zone string: %v", err)
				setting.DefaultReverbZoneStr = strconv.FormatInt(rawSettings.DefaultReverbZoneStrOffset, 16)
			} else {
				setting.DefaultReverbZoneStr = defaultReverbZoneStr
			}
		}

		if rawSettings.UnkStr > 0 {
			r.Seek(base+rawSettings.UnkStr, io.SeekStart)
			unkStr, err := util.ReadCString(r)
			if err != nil {
				//return nil, fmt.Errorf("reading unknown string: %v", err)
				setting.UnkStr = strconv.FormatInt(rawSettings.UnkStr, 16)
			} else {
				setting.UnkStr = unkStr
			}
		}

		infos = append(infos, setting)
		r.Seek(base+int64(header.Size), io.SeekStart)
	}
	return infos, nil
}
