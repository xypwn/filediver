package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type SkySetting struct {
	RayleighBeta                              mgl32.Vec3 `json:"rayleigh_beta"`
	RayleighTintHaxDay                        mgl32.Vec3 `json:"rayleigh_tint_hax_day"`
	RayleighTintHaxSunset                     mgl32.Vec3 `json:"rayleigh_tint_hax_sunset"`
	RayleighTintHaxNight                      mgl32.Vec3 `json:"rayleigh_tint_hax_night"`
	MieBeta                                   float32    `json:"mie_beta"`
	MieTintHaxDay                             mgl32.Vec3 `json:"mie_tint_hax_day"`
	MieTintHaxSunset                          mgl32.Vec3 `json:"mie_tint_hax_sunset"`
	MieTintHaxNight                           mgl32.Vec3 `json:"mie_tint_hax_night"`
	MieHeight                                 float32    `json:"mie_height"`
	AtmospherePhase                           float32    `json:"atmosphere_phase"`
	AtmosphereSaturationDay                   float32    `json:"atmosphere_saturation_day"`
	AtmosphereSaturationSunset                float32    `json:"atmosphere_saturation_sunset"`
	AtmosphereSaturationNight                 float32    `json:"atmosphere_saturation_night"`
	AtmosphereMinimumLight                    float32    `json:"atmosphere_minimum_light"`
	FogForwardscatterPhase                    float32    `json:"fog_forwardscatter_phase"`
	FogBackscatterPhase                       float32    `json:"fog_backscatter_phase"`
	FogBackscatterLerp                        float32    `json:"fog_backscatter_lerp"`
	CloudColorHaxDay                          mgl32.Vec3 `json:"cloud_color_hax_day"`
	CloudColorHaxSunset                       mgl32.Vec3 `json:"cloud_color_hax_sunset"`
	CloudColorHaxNight                        mgl32.Vec3 `json:"cloud_color_hax_night"`
	MaterialWetnessDay                        float32    `json:"material_wetness_day"`
	MaterialWetnessSunset                     float32    `json:"material_wetness_sunset"`
	MaterialWetnessNight                      float32    `json:"material_wetness_night"`
	StarIntensity                             float32    `json:"star_intensity"`
	InSpaceStarIntensity                      float32    `json:"in_space_star_intensity"`
	AmbientDiffuseIntensityDay                float32    `json:"ambient_diffuse_intensity_day"`
	AmbientDiffuseIntensitySunset             float32    `json:"ambient_diffuse_intensity_sunset"`
	AmbientDiffuseIntensityNight              float32    `json:"ambient_diffuse_intensity_night"`
	AmbientDiffuseGroundBounceIntensityDay    float32    `json:"ambient_diffuse_ground_bounce_intensity_day"`
	AmbientDiffuseGroundBounceIntensitySunset float32    `json:"ambient_diffuse_ground_bounce_intensity_sunset"`
	AmbientDiffuseGroundBounceIntensityNight  float32    `json:"ambient_diffuse_ground_bounce_intensity_night"`
	FogAmbientIntensityDay                    float32    `json:"fog_ambient_intensity_day"`
	FogAmbientIntensitySunset                 float32    `json:"fog_ambient_intensity_sunset"`
	FogAmbientIntensityNight                  float32    `json:"fog_ambient_intensity_night"`
	FogAmbientGroundBounceIntensityDay        float32    `json:"fog_ambient_ground_bounce_intensity_day"`
	FogAmbientGroundBounceIntensitySunset     float32    `json:"fog_ambient_ground_bounce_intensity_sunset"`
	FogAmbientGroundBounceIntensityNight      float32    `json:"fog_ambient_ground_bounce_intensity_night"`
	UnknownFloat_1                            float32    `json:"unknown_float_1"`
	UnknownFloat_2                            float32    `json:"unknown_float_2"`
	UnknownFloat_3                            float32    `json:"unknown_float_3"`
	UnknownVector_1                           mgl32.Vec3 `json:"unknown_vector_1"`
	UnknownVector_2                           mgl32.Vec3 `json:"unknown_vector_2"`
}

type SkySettings struct {
	ID       stingray.Hash
	Settings []SkySetting
}

type rawSkySettings struct {
	ID stingray.Hash
	DLArray
}

func LoadSkySettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]SkySettings, error) {
	r := bytes.NewReader(skySettings)

	settings := make([]SkySettings, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("SkySettings") {
			return nil, fmt.Errorf("invalid sky settings file")
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding sky settings base: %v", err)
		}

		var rawSetting rawSkySettings
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading sky settings: %v", err)
		}

		settingArray := make([]SkySetting, rawSetting.Count)
		if _, err := r.Seek(base+rawSetting.Offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking sky setting array: %v", err)
		}
		if err := binary.Read(r, binary.LittleEndian, settingArray); err != nil {
			return nil, fmt.Errorf("reading sky setting array: %v", err)
		}

		settings = append(settings, SkySettings{
			ID:       rawSetting.ID,
			Settings: settingArray,
		})

		if _, err := r.Seek(base+int64(header.Size), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking next dl entry: %v", err)
		}
	}
	return settings, nil
}
