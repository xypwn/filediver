package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type rawArcInfo struct {
	Type                              enum.ArcType
	Speed                             float32
	Distance                          float32
	DistanceAtMaxAngleSpread          float32
	DistanceAtMaxAngleSpreadFirstShot float32
	MaxAngleSpread                    float32
	MaxAngleSpreadFirstShot           float32
	MaxChainLength                    uint32
	MaxChainSplit                     uint32
	DamageInfoType                    enum.DamageInfoType
	CameraShakePath                   stingray.Hash // [camera_shake]
	ParticleEffectPath                stingray.Hash // [particles]
	ParticleLingerTime                float32
	_                                 [4]uint8
	ExplosionEffectPath               stingray.Hash // [particles]
	HitEffectDamageType               enum.HitEffectDamageType
	HitAudioEvent                     uint32
	BeamPath                          stingray.Hash // [unit]
	UnkInt1                           uint32
	UnkBool                           uint8
	_                                 [3]uint8
	UnkInt2                           uint32
	_                                 [4]uint8
}

type ArcInfo struct {
	Type                              enum.ArcType             `json:"type"`
	Speed                             float32                  `json:"speed"`
	Distance                          float32                  `json:"distance"`
	DistanceAtMaxAngleSpread          float32                  `json:"distance_at_max_angle_spread"`
	DistanceAtMaxAngleSpreadFirstShot float32                  `json:"distance_at_max_angle_spread_first_shot"`
	MaxAngleSpread                    float32                  `json:"max_angle_spread"`
	MaxAngleSpreadFirstShot           float32                  `json:"max_angle_spread_first_shot"`
	MaxChainLength                    uint32                   `json:"max_chain_length"`
	MaxChainSplit                     uint32                   `json:"max_chain_split"`
	DamageInfoType                    enum.DamageInfoType      `json:"damage_info_type"`
	CameraShakePath                   string                   `json:"camera_shake_path"`    // [camera_shake]
	ParticleEffectPath                string                   `json:"particle_effect_path"` // [particles]
	ParticleLingerTime                float32                  `json:"particle_linger_time"`
	ExplosionEffectPath               string                   `json:"explosion_effect_path"` // [particles]
	HitEffectDamageType               enum.HitEffectDamageType `json:"hit_effect_damage_type"`
	HitAudioEvent                     uint32                   `json:"hit_audio_event"`
	BeamPath                          string                   `json:"beam_path"` // [unit]
	UnkInt1                           uint32                   `json:"unk_int1"`
	UnkBool                           bool                     `json:"unk_bool"`
	UnkInt2                           uint32                   `json:"unk_int2"`
}

func (a rawArcInfo) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ArcInfo {
	return ArcInfo{
		Type:                              a.Type,
		Speed:                             a.Speed,
		Distance:                          a.Distance,
		DistanceAtMaxAngleSpread:          a.DistanceAtMaxAngleSpread,
		DistanceAtMaxAngleSpreadFirstShot: a.DistanceAtMaxAngleSpreadFirstShot,
		MaxAngleSpread:                    a.MaxAngleSpread,
		MaxAngleSpreadFirstShot:           a.MaxAngleSpreadFirstShot,
		MaxChainLength:                    a.MaxChainLength,
		MaxChainSplit:                     a.MaxChainSplit,
		DamageInfoType:                    a.DamageInfoType,
		CameraShakePath:                   lookupHash(a.CameraShakePath),
		ParticleEffectPath:                lookupHash(a.ParticleEffectPath),
		ParticleLingerTime:                a.ParticleLingerTime,
		ExplosionEffectPath:               lookupHash(a.ExplosionEffectPath),
		HitEffectDamageType:               a.HitEffectDamageType,
		HitAudioEvent:                     a.HitAudioEvent,
		BeamPath:                          lookupHash(a.BeamPath),
		UnkInt1:                           a.UnkInt1,
		UnkBool:                           a.UnkBool != 0,
		UnkInt2:                           a.UnkInt2,
	}
}

func LoadArcSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([][]ArcInfo, error) {
	r := bytes.NewReader(arcSettings)

	settings := make([][]ArcInfo, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("ArcSettings") {
			return nil, fmt.Errorf("invalid arc settings file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var rawSettings DLArray
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		rawSetting := make([]rawArcInfo, rawSettings.Count)
		r.Seek(base+rawSettings.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading arc info array: %v", err)
		}

		setting := make([]ArcInfo, 0)
		for _, info := range rawSetting {
			setting = append(setting, info.Resolve(lookupHash, lookupThinHash, lookupStrings))
		}
		settings = append(settings, setting)
	}
	return settings, nil
}
