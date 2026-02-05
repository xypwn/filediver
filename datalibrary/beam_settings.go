package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type rawBeamInfo struct {
	Type               enum.BeamType
	Radius             float32
	Length             float32
	DamageInfoType     enum.DamageInfoType
	FocalShoulderWidth float32
	DecalSize          float32
	SurfaceImpactType  enum.SurfaceImpactType
	_                  [4]uint8
	BeamEffectPath     stingray.Hash // [particles]
	FalloffEffectPath  stingray.Hash // [particles]
	BeamUnitPath       stingray.Hash // [unit]
	AudioTrail         uint32
	AudioTrailStop     uint32
	RaycastTemplate    enum.RaycastTemplate
	OverlapBeamQuery   stingray.ThinHash
	EffectDamageType   enum.HitEffectDamageType
	UnkFloat           float32 // name length 18
	ExplosionType      enum.ExplosionType
	UnkBool            uint8 // name length 27
	_                  [3]uint8
	DangerLevel        enum.DangerLevel
	_                  [4]uint8
}

type BeamInfo struct {
	Type               enum.BeamType            `json:"type"`
	Radius             float32                  `json:"radius"`
	Length             float32                  `json:"length"`
	DamageInfoType     enum.DamageInfoType      `json:"damage_info_type"`
	FocalShoulderWidth float32                  `json:"focal_shoulder_width"`
	DecalSize          float32                  `json:"decal_size"`
	SurfaceImpactType  enum.SurfaceImpactType   `json:"surface_impact_type"`
	BeamEffectPath     string                   `json:"beam_effect_path"`    // [particles]
	FalloffEffectPath  string                   `json:"falloff_effect_path"` // [particles]
	BeamUnitPath       string                   `json:"beam_unit_path"`      // [unit]
	AudioTrail         uint32                   `json:"audio_trail"`
	AudioTrailStop     uint32                   `json:"audio_trail_stop"`
	RaycastTemplate    enum.RaycastTemplate     `json:"raycast_template"`
	OverlapBeamQuery   string                   `json:"overlap_beam_query"`
	EffectDamageType   enum.HitEffectDamageType `json:"effect_damage_type"`
	UnkFloat           float32                  `json:"unk_float"` // name length 18
	ExplosionType      enum.ExplosionType       `json:"explosion_type"`
	UnkBool            bool                     `json:"unk_bool"` // name length 27
	DangerLevel        enum.DangerLevel         `json:"danger_level"`
}

func (a rawBeamInfo) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) BeamInfo {
	return BeamInfo{
		Type:               a.Type,
		Radius:             a.Radius,
		Length:             a.Length,
		DamageInfoType:     a.DamageInfoType,
		FocalShoulderWidth: a.FocalShoulderWidth,
		DecalSize:          a.DecalSize,
		SurfaceImpactType:  a.SurfaceImpactType,
		BeamEffectPath:     lookupHash(a.BeamEffectPath),
		FalloffEffectPath:  lookupHash(a.FalloffEffectPath),
		BeamUnitPath:       lookupHash(a.BeamUnitPath),
		AudioTrail:         a.AudioTrail,
		AudioTrailStop:     a.AudioTrailStop,
		RaycastTemplate:    a.RaycastTemplate,
		OverlapBeamQuery:   lookupThinHash(a.OverlapBeamQuery),
		EffectDamageType:   a.EffectDamageType,
		UnkFloat:           a.UnkFloat,
		ExplosionType:      a.ExplosionType,
		UnkBool:            a.UnkBool != 0,
		DangerLevel:        a.DangerLevel,
	}
}

func LoadBeamSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([][]BeamInfo, error) {
	r := bytes.NewReader(beamSettings)

	settings := make([][]BeamInfo, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("BeamSettings") {
			return nil, fmt.Errorf("invalid beam settings file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var rawSettings DLArray
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		rawSetting := make([]rawBeamInfo, rawSettings.Count)
		r.Seek(base+rawSettings.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading beam info array: %v", err)
		}

		setting := make([]BeamInfo, 0)
		for _, info := range rawSetting {
			setting = append(setting, info.Resolve(lookupHash, lookupThinHash, lookupStrings))
		}
		settings = append(settings, setting)
	}
	return settings, nil
}
