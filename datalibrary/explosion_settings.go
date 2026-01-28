package datalib

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type ExplosionInfoBitfield1 uint8

func (e ExplosionInfoBitfield1) DoKnockback() bool {
	return (e & 0x1 << 0) != 0
}

func (e ExplosionInfoBitfield1) DisableDefaultSurfaceImpact() bool {
	return (e & 0x1 << 1) != 0
}

func (e ExplosionInfoBitfield1) IsHugeExplosion() bool {
	return (e & 0x1 << 2) != 0
}

func (e ExplosionInfoBitfield1) LookForAudioSourceHijack() bool {
	return (e & 0x1 << 3) != 0
}

// Name length 29
func (e ExplosionInfoBitfield1) Item4() bool {
	return (e & 0x1 << 4) != 0
}

type ExplosionInfoBitfieldStruct1 struct {
	DoKnockback                 bool `json:"do_knockback"`
	DisableDefaultSurfaceImpact bool `json:"disable_default_surface_impact"`
	IsHugeExplosion             bool `json:"is_huge_explosion"`
	LookForAudioSourceHijack    bool `json:"look_for_audio_source_hijack"`
	Item4                       bool `json:"item4"`
}

func (p ExplosionInfoBitfield1) MarshalJSON() ([]byte, error) {
	return json.Marshal(ExplosionInfoBitfieldStruct1{
		DoKnockback:                 p.DoKnockback(),
		DisableDefaultSurfaceImpact: p.DisableDefaultSurfaceImpact(),
		IsHugeExplosion:             p.IsHugeExplosion(),
		LookForAudioSourceHijack:    p.LookForAudioSourceHijack(),
		Item4:                       p.Item4(),
	})
}

type ExplosionInfoBitfield2 uint8

// Name length 22
func (e ExplosionInfoBitfield2) Item0() bool {
	return (e & 0x1 << 0) != 0
}

// Name length 35
func (e ExplosionInfoBitfield2) Item1() bool {
	return (e & 0x1 << 1) != 0
}

type ExplosionInfoBitfieldStruct2 struct {
	Item0 bool `json:"item0"`
	Item1 bool `json:"item1"`
}

func (p ExplosionInfoBitfield2) MarshalJSON() ([]byte, error) {
	return json.Marshal(ExplosionInfoBitfieldStruct2{
		Item0: p.Item0(),
		Item1: p.Item1(),
	})
}

type rawExplosionInfo struct {
	Type                        enum.ExplosionType
	DamageType                  enum.DamageInfoType
	UnkHash                     stingray.ThinHash // name length 24
	UnkBool                     uint8             // name length 12
	_                           [3]uint8
	InnerRadius                 float32
	OuterRadius                 float32
	StaggerRadius               float32
	ConeAngle                   float32
	CameraShakeRadius           float32
	_                           [4]uint8
	ShakeAssets                 DLArray       // [camera_shake]
	ParticleEffectPath          stingray.Hash // [particles]
	AudioEvent                  uint32
	CraterType                  enum.TerrainDeformationType
	_                           [3]uint8
	SurfaceImpactType           enum.SurfaceImpactType
	HitEffectDamageType         enum.HitEffectDamageType
	NumShrapnelProjectiles      uint32
	ShrapnelProjectileType      enum.ProjectileType
	NoiseTemplate               enum.NoiseTemplate
	WindEffectTemplate          enum.WindEffectTemplate
	FireTemplate                enum.FireTemplate
	PersistentStatusVolume      enum.StatusEffectTemplateType
	StatusVolumeEffectTime      float32
	StatusVolumeEffectAudioPlay uint32
	StatusVolumeEffectAudioStop uint32
	ExplosionInfoBitfield1
	_       [3]uint8
	ArcType enum.ArcType
	ExplosionInfoBitfield2
	_        [3]uint8
	UnkFloat float32 // name length 21
	_        [4]uint8
}

type itmExplosionInfo struct {
	Type                        enum.ExplosionType
	DamageType                  enum.DamageInfoType
	UnkHash                     stingray.ThinHash // name length 24
	UnkBool                     uint8             // name length 12
	InnerRadius                 float32
	OuterRadius                 float32
	StaggerRadius               float32
	ConeAngle                   float32
	CameraShakeRadius           float32
	ShakeAssets                 []stingray.Hash // [camera_shake]
	ParticleEffectPath          stingray.Hash   // [particles]
	AudioEvent                  uint32
	CraterType                  enum.TerrainDeformationType
	SurfaceImpactType           enum.SurfaceImpactType
	HitEffectDamageType         enum.HitEffectDamageType
	NumShrapnelProjectiles      uint32
	ShrapnelProjectileType      enum.ProjectileType
	NoiseTemplate               enum.NoiseTemplate
	WindEffectTemplate          enum.WindEffectTemplate
	FireTemplate                enum.FireTemplate
	PersistentStatusVolume      enum.StatusEffectTemplateType
	StatusVolumeEffectTime      float32
	StatusVolumeEffectAudioPlay uint32
	StatusVolumeEffectAudioStop uint32
	ExplosionInfoBitfield1
	ArcType enum.ArcType
	ExplosionInfoBitfield2
	UnkFloat float32 // name length 21
}

type ExplosionInfo struct {
	Type                        enum.ExplosionType            `json:"type"`
	DamageType                  enum.DamageInfoType           `json:"damage_type"`
	UnkHash                     string                        `json:"unk_hash"`
	UnkBool                     uint8                         `json:"unk_bool"`
	InnerRadius                 float32                       `json:"inner_radius"`
	OuterRadius                 float32                       `json:"outer_radius"`
	StaggerRadius               float32                       `json:"stagger_radius"`
	ConeAngle                   float32                       `json:"cone_angle"`
	CameraShakeRadius           float32                       `json:"camera_shake_radius"`
	ShakeAssets                 []string                      `json:"shake_assets"`
	ParticleEffectPath          string                        `json:"particle_effect_path"`
	AudioEvent                  uint32                        `json:"audio_event"`
	CraterType                  enum.TerrainDeformationType   `json:"crater_type"`
	SurfaceImpactType           enum.SurfaceImpactType        `json:"surface_impact_type"`
	HitEffectDamageType         enum.HitEffectDamageType      `json:"hit_effect_damage_type"`
	NumShrapnelProjectiles      uint32                        `json:"num_shrapnel_projectiles"`
	ShrapnelProjectileType      enum.ProjectileType           `json:"shrapnel_projectile_type"`
	NoiseTemplate               enum.NoiseTemplate            `json:"noise_template"`
	WindEffectTemplate          enum.WindEffectTemplate       `json:"wind_effect_template"`
	FireTemplate                enum.FireTemplate             `json:"fire_template"`
	PersistentStatusVolume      enum.StatusEffectTemplateType `json:"persistent_status_volume"`
	StatusVolumeEffectTime      float32                       `json:"status_volume_effect_time"`
	StatusVolumeEffectAudioPlay uint32                        `json:"status_volume_effect_audio_play"`
	StatusVolumeEffectAudioStop uint32                        `json:"status_volume_effect_audio_stop"`
	ExplosionInfoBitfield1      `json:"explosion_info_bitfield1"`
	ArcType                     enum.ArcType `json:"arc_type"`
	ExplosionInfoBitfield2      `json:"explosion_info_bitfield2"`
	UnkFloat                    float32 `json:"unk_float"`
}

func (a itmExplosionInfo) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ExplosionInfo {
	shakeAssets := make([]string, 0)
	for _, asset := range a.ShakeAssets {
		if asset.Value == 0 {
			break
		}
		shakeAssets = append(shakeAssets, lookupHash(asset))
	}
	return ExplosionInfo{
		Type:                        a.Type,
		DamageType:                  a.DamageType,
		UnkHash:                     lookupThinHash(a.UnkHash),
		UnkBool:                     a.UnkBool,
		InnerRadius:                 a.InnerRadius,
		OuterRadius:                 a.OuterRadius,
		StaggerRadius:               a.StaggerRadius,
		ConeAngle:                   a.ConeAngle,
		CameraShakeRadius:           a.CameraShakeRadius,
		ShakeAssets:                 shakeAssets,
		ParticleEffectPath:          lookupHash(a.ParticleEffectPath),
		AudioEvent:                  a.AudioEvent,
		CraterType:                  a.CraterType,
		SurfaceImpactType:           a.SurfaceImpactType,
		HitEffectDamageType:         a.HitEffectDamageType,
		NumShrapnelProjectiles:      a.NumShrapnelProjectiles,
		ShrapnelProjectileType:      a.ShrapnelProjectileType,
		NoiseTemplate:               a.NoiseTemplate,
		WindEffectTemplate:          a.WindEffectTemplate,
		FireTemplate:                a.FireTemplate,
		PersistentStatusVolume:      a.PersistentStatusVolume,
		StatusVolumeEffectTime:      a.StatusVolumeEffectTime,
		StatusVolumeEffectAudioPlay: a.StatusVolumeEffectAudioPlay,
		StatusVolumeEffectAudioStop: a.StatusVolumeEffectAudioStop,
		ExplosionInfoBitfield1:      a.ExplosionInfoBitfield1,
		ArcType:                     a.ArcType,
		ExplosionInfoBitfield2:      a.ExplosionInfoBitfield2,
		UnkFloat:                    a.UnkFloat,
	}
}

func LoadExplosionSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([][]ExplosionInfo, error) {
	r := bytes.NewReader(explosionSettings)

	settings := make([][]ExplosionInfo, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("ExplosionSettings") {
			return nil, fmt.Errorf("invalid explosion settings file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var rawSettings DLArray
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		rawSetting := make([]rawExplosionInfo, rawSettings.Count)
		r.Seek(base+rawSettings.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading explosion info array: %v", err)
		}

		setting := make([]ExplosionInfo, 0)
		for _, info := range rawSetting {
			shakeAssets := make([]stingray.Hash, info.ShakeAssets.Count)
			r.Seek(base+info.ShakeAssets.Offset, io.SeekStart)
			if err := binary.Read(r, binary.LittleEndian, &shakeAssets); err != nil {
				return nil, fmt.Errorf("reading shake assets array: %v", err)
			}
			itm := itmExplosionInfo{
				Type:                        info.Type,
				DamageType:                  info.DamageType,
				UnkHash:                     info.UnkHash,
				UnkBool:                     info.UnkBool,
				InnerRadius:                 info.InnerRadius,
				OuterRadius:                 info.OuterRadius,
				StaggerRadius:               info.StaggerRadius,
				ConeAngle:                   info.ConeAngle,
				CameraShakeRadius:           info.CameraShakeRadius,
				ShakeAssets:                 shakeAssets,
				ParticleEffectPath:          info.ParticleEffectPath,
				AudioEvent:                  info.AudioEvent,
				CraterType:                  info.CraterType,
				SurfaceImpactType:           info.SurfaceImpactType,
				HitEffectDamageType:         info.HitEffectDamageType,
				NumShrapnelProjectiles:      info.NumShrapnelProjectiles,
				ShrapnelProjectileType:      info.ShrapnelProjectileType,
				NoiseTemplate:               info.NoiseTemplate,
				WindEffectTemplate:          info.WindEffectTemplate,
				FireTemplate:                info.FireTemplate,
				PersistentStatusVolume:      info.PersistentStatusVolume,
				StatusVolumeEffectTime:      info.StatusVolumeEffectTime,
				StatusVolumeEffectAudioPlay: info.StatusVolumeEffectAudioPlay,
				StatusVolumeEffectAudioStop: info.StatusVolumeEffectAudioStop,
				ExplosionInfoBitfield1:      info.ExplosionInfoBitfield1,
				ArcType:                     info.ArcType,
				ExplosionInfoBitfield2:      info.ExplosionInfoBitfield2,
				UnkFloat:                    info.UnkFloat,
			}
			setting = append(setting, itm.Resolve(lookupHash, lookupThinHash, lookupStrings))
		}
		settings = append(settings, setting)
	}
	return settings, nil
}
