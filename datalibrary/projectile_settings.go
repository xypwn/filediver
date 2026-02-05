package datalib

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type ProjectileInfoBitfield1 uint16

// Name length 16
func (p ProjectileInfoBitfield1) Item0() bool {
	return p&(0x1<<0) != 0
}

// Name length 24
func (p ProjectileInfoBitfield1) Item1() bool {
	return p&(0x1<<1) != 0
}

// Name length 32
func (p ProjectileInfoBitfield1) Item2() bool {
	return p&(0x1<<2) != 0
}

// Name length 29
func (p ProjectileInfoBitfield1) Item3() bool {
	return p&(0x1<<3) != 0
}

// Name length 27
func (p ProjectileInfoBitfield1) Item4() bool {
	return p&(0x1<<4) != 0
}

// Name length 27
func (p ProjectileInfoBitfield1) Item5() bool {
	return p&(0x1<<5) != 0
}

// Name length 13
func (p ProjectileInfoBitfield1) Item6() bool {
	return p&(0x1<<6) != 0
}

// Name length 12
func (p ProjectileInfoBitfield1) Item7() bool {
	return p&(0x1<<7) != 0
}

// Name length 17
func (p ProjectileInfoBitfield1) Item8() bool {
	return p&(0x1<<8) != 0
}

// Name length 15
func (p ProjectileInfoBitfield1) Item9() bool {
	return p&(0x1<<9) != 0
}

// Name length 23
func (p ProjectileInfoBitfield1) Item10() bool {
	return p&(0x1<<10) != 0
}

// Name length 26
func (p ProjectileInfoBitfield1) Item11() bool {
	return p&(0x1<<11) != 0
}

// Name length 28
func (p ProjectileInfoBitfield1) Item12() bool {
	return p&(0x1<<12) != 0
}

type ProjectileInfoBitfieldStruct1 struct {
	Item0  bool `json:"item0"`
	Item1  bool `json:"item1"`
	Item2  bool `json:"item2"`
	Item3  bool `json:"item3"`
	Item4  bool `json:"item4"`
	Item5  bool `json:"item5"`
	Item6  bool `json:"item6"`
	Item7  bool `json:"item7"`
	Item8  bool `json:"item8"`
	Item9  bool `json:"item9"`
	Item10 bool `json:"item10"`
	Item11 bool `json:"item11"`
	Item12 bool `json:"item12"`
}

// Name length 28
func (p ProjectileInfoBitfield1) MarshalJSON() ([]byte, error) {
	return json.Marshal(ProjectileInfoBitfieldStruct1{
		Item0:  p.Item0(),
		Item1:  p.Item1(),
		Item2:  p.Item2(),
		Item3:  p.Item3(),
		Item4:  p.Item4(),
		Item5:  p.Item5(),
		Item6:  p.Item6(),
		Item7:  p.Item7(),
		Item8:  p.Item8(),
		Item9:  p.Item9(),
		Item10: p.Item10(),
		Item11: p.Item11(),
		Item12: p.Item12(),
	})
}

type ProjectileInfoBitfield2 uint8

// Name length 20
func (p ProjectileInfoBitfield2) Item0() bool {
	return p&(0x1<<0) != 0
}

// Name length 22
func (p ProjectileInfoBitfield2) Item1() bool {
	return p&(0x1<<1) != 0
}

type ProjectileInfoBitfieldStruct2 struct {
	Item0 bool `json:"item0"`
	Item1 bool `json:"item1"`
}

// Name length 28
func (p ProjectileInfoBitfield2) MarshalJSON() ([]byte, error) {
	return json.Marshal(ProjectileInfoBitfieldStruct2{
		Item0: p.Item0(),
		Item1: p.Item1(),
	})
}

type rawProjectileInfo struct {
	Type                              enum.ProjectileType
	NameUpper                         uint32
	NameCased                         uint32
	ShortName                         uint32
	HudIcon                           stingray.Hash // [material]
	Calibre                           float32
	NumProjectiles                    uint32
	Speed                             float32
	Mass                              float32
	Drag                              float32
	GravityMultiplier                 float32
	SimulationSteps                   uint32
	UnkFloat                          float32 // Name length 16, maybe simulation_delta?
	LifeTime                          float32
	DamageInfoType                    enum.DamageInfoType
	PenetrationSlowdown               float32
	_                                 [4]uint8
	ParticleThrusterEffectPath        stingray.Hash // [particles]
	LowAmmoParticleThrusterEffectPath stingray.Hash // [particles]
	ParticleThrusterEffectLength      float32
	_                                 [4]uint8
	ParticleTrailEffectPath           stingray.Hash // [particles]
	DissipateParticleEffectPath       stingray.Hash // [particles]
	DisintegrationParticleEffectPath  stingray.Hash // [particles]
	HitDirectionHintEffect            stingray.Hash // [particles]
	ProjectileUnitPath                stingray.Hash // [unit]
	ExplosionThresholdAngle           float32
	UnkBool                           uint8 // Name length 23
	_                                 [3]uint8
	ExplosionTypeOnImpact             enum.ExplosionType
	ExplosionProximity                float32
	ExplosionDelay                    float32
	ExplosionTypeExpire               enum.ExplosionType
	ArmingDistance                    float32
	DecalSize                         float32
	SurfaceImpactType                 enum.SurfaceImpactType
	RicochetImpactType                enum.SurfaceImpactType
	MaxRicochets                      uint32
	RicochetThresholdAngle            float32
	RicochetThresholdSpeed            float32
	RicochetAngleLoss                 float32
	RicochetSpreadVertical            float32
	RicochetSpreadHorizontal          float32
	AudioCrackEvent                   uint32
	AudioRicochetEvent                uint32
	AudioTrailEventStart              uint32
	AudioTrailEventStop               uint32
	UnkFloat2                         float32           // name length 31
	UnkAudioEvent                     stingray.ThinHash // name length 28
	UnkBool2                          uint8             // name length 28
	_                                 [3]uint8
	StatusEffectVolumeTemplate        enum.ProjectileStatusEffectVolumeTemplate
	EffectDamageType                  enum.HitEffectDamageType
	FireTemplate                      enum.FireTemplate
	UnkBitfield                       ProjectileInfoBitfield1
	_                                 [2]uint8
	RaycastTemplate                   enum.RaycastTemplate
	UnkAudioEvent2                    stingray.ThinHash // Could also be a thinhash? Name length 36
	UnkBitfield2                      ProjectileInfoBitfield2
	_                                 [3]uint8
	UnkVector                         mgl32.Vec3 // name length 34
	_                                 [4]uint8
}

type ProjectileInfo struct {
	Type                              enum.ProjectileType                       `json:"type"`
	NameUpper                         string                                    `json:"name_upper"`
	NameCased                         string                                    `json:"name_cased"`
	ShortName                         string                                    `json:"short_name"`
	HudIcon                           string                                    `json:"hud_icon"` // [material]
	Calibre                           float32                                   `json:"calibre"`
	NumProjectiles                    uint32                                    `json:"num_projectiles"`
	Speed                             float32                                   `json:"speed"`
	Mass                              float32                                   `json:"mass"`
	Drag                              float32                                   `json:"drag"`
	GravityMultiplier                 float32                                   `json:"gravity_multiplier"`
	SimulationSteps                   uint32                                    `json:"simulation_steps"`
	UnkFloat                          float32                                   `json:"unk_float"` // Name length 16, maybe simulation_delta?
	LifeTime                          float32                                   `json:"life_time"`
	DamageInfoType                    enum.DamageInfoType                       `json:"damage_info_type"`
	PenetrationSlowdown               float32                                   `json:"penetration_slowdown"`
	ParticleThrusterEffectPath        string                                    `json:"particle_thruster_effect_path"`          // [particles]
	LowAmmoParticleThrusterEffectPath string                                    `json:"low_ammo_particle_thruster_effect_path"` // [particles]
	ParticleThrusterEffectLength      float32                                   `json:"particle_thruster_effect_length"`
	ParticleTrailEffectPath           string                                    `json:"particle_trail_effect_path"`          // [particles]
	DissipateParticleEffectPath       string                                    `json:"dissipate_particle_effect_path"`      // [particles]
	DisintegrationParticleEffectPath  string                                    `json:"disintegration_particle_effect_path"` // [particles]
	HitDirectionHintEffect            string                                    `json:"hit_direction_hint_effect"`           // [particles]
	ProjectileUnitPath                string                                    `json:"projectile_unit_path"`                // [unit]
	ExplosionThresholdAngle           float32                                   `json:"explosion_threshold_angle"`
	UnkBool                           bool                                      `json:"unk_bool"` // Name length 23
	ExplosionTypeOnImpact             enum.ExplosionType                        `json:"explosion_type_on_impact"`
	ExplosionProximity                float32                                   `json:"explosion_proximity"`
	ExplosionDelay                    float32                                   `json:"explosion_delay"`
	ExplosionTypeExpire               enum.ExplosionType                        `json:"explosion_type_expire"`
	ArmingDistance                    float32                                   `json:"arming_distance"`
	DecalSize                         float32                                   `json:"decal_size"`
	SurfaceImpactType                 enum.SurfaceImpactType                    `json:"surface_impact_type"`
	RicochetImpactType                enum.SurfaceImpactType                    `json:"ricochet_impact_type"`
	MaxRicochets                      uint32                                    `json:"max_ricochets"`
	RicochetThresholdAngle            float32                                   `json:"ricochet_threshold_angle"`
	RicochetThresholdSpeed            float32                                   `json:"ricochet_threshold_speed"`
	RicochetAngleLoss                 float32                                   `json:"ricochet_angle_loss"`
	RicochetSpreadVertical            float32                                   `json:"ricochet_spread_vertical"`
	RicochetSpreadHorizontal          float32                                   `json:"ricochet_spread_horizontal"`
	AudioCrackEvent                   uint32                                    `json:"audio_crack_event"`
	AudioRicochetEvent                uint32                                    `json:"audio_ricochet_event"`
	AudioTrailEventStart              uint32                                    `json:"audio_trail_event_start"`
	AudioTrailEventStop               uint32                                    `json:"audio_trail_event_stop"`
	UnkFloat2                         float32                                   `json:"unk_float2"`      // name length 31
	UnkAudioEvent                     string                                    `json:"unk_audio_event"` // name length 28
	UnkBool2                          uint8                                     `json:"unk_bool2"`       // name length 28
	StatusEffectVolumeTemplate        enum.ProjectileStatusEffectVolumeTemplate `json:"status_effect_volume_template"`
	EffectDamageType                  enum.HitEffectDamageType                  `json:"effect_damage_type"`
	FireTemplate                      enum.FireTemplate                         `json:"fire_template"`
	UnkBitfield                       ProjectileInfoBitfield1                   `json:"unk_bitfield"`
	RaycastTemplate                   enum.RaycastTemplate                      `json:"raycast_template"`
	UnkAudioEvent2                    string                                    `json:"unk_audio_event2"` // Could also be a thinhash? Name length 36
	UnkBitfield2                      ProjectileInfoBitfield2                   `json:"unk_bitfield2"`
	UnkVector                         mgl32.Vec3                                `json:"unk_vector"` // name length 34
}

func (a rawProjectileInfo) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ProjectileInfo {
	return ProjectileInfo{
		Type:                              a.Type,
		NameUpper:                         lookupStrings(a.NameUpper),
		NameCased:                         lookupStrings(a.NameCased),
		ShortName:                         lookupStrings(a.ShortName),
		HudIcon:                           lookupHash(a.HudIcon),
		Calibre:                           a.Calibre,
		NumProjectiles:                    a.NumProjectiles,
		Speed:                             a.Speed,
		Mass:                              a.Mass,
		Drag:                              a.Drag,
		GravityMultiplier:                 a.GravityMultiplier,
		SimulationSteps:                   a.SimulationSteps,
		UnkFloat:                          a.UnkFloat,
		LifeTime:                          a.LifeTime,
		DamageInfoType:                    a.DamageInfoType,
		PenetrationSlowdown:               a.PenetrationSlowdown,
		ParticleThrusterEffectPath:        lookupHash(a.ParticleThrusterEffectPath),
		LowAmmoParticleThrusterEffectPath: lookupHash(a.LowAmmoParticleThrusterEffectPath),
		ParticleThrusterEffectLength:      a.ParticleThrusterEffectLength,
		ParticleTrailEffectPath:           lookupHash(a.ParticleTrailEffectPath),
		DissipateParticleEffectPath:       lookupHash(a.DissipateParticleEffectPath),
		DisintegrationParticleEffectPath:  lookupHash(a.DisintegrationParticleEffectPath),
		HitDirectionHintEffect:            lookupHash(a.HitDirectionHintEffect),
		ProjectileUnitPath:                lookupHash(a.ProjectileUnitPath),
		ExplosionThresholdAngle:           a.ExplosionThresholdAngle,
		UnkBool:                           a.UnkBool != 0,
		ExplosionTypeOnImpact:             a.ExplosionTypeOnImpact,
		ExplosionProximity:                a.ExplosionProximity,
		ExplosionDelay:                    a.ExplosionDelay,
		ExplosionTypeExpire:               a.ExplosionTypeExpire,
		ArmingDistance:                    a.ArmingDistance,
		DecalSize:                         a.DecalSize,
		SurfaceImpactType:                 a.SurfaceImpactType,
		RicochetImpactType:                a.RicochetImpactType,
		MaxRicochets:                      a.MaxRicochets,
		RicochetThresholdAngle:            a.RicochetThresholdAngle,
		RicochetThresholdSpeed:            a.RicochetThresholdSpeed,
		RicochetAngleLoss:                 a.RicochetAngleLoss,
		RicochetSpreadVertical:            a.RicochetSpreadVertical,
		RicochetSpreadHorizontal:          a.RicochetSpreadHorizontal,
		AudioCrackEvent:                   a.AudioCrackEvent,
		AudioRicochetEvent:                a.AudioRicochetEvent,
		AudioTrailEventStart:              a.AudioTrailEventStart,
		AudioTrailEventStop:               a.AudioTrailEventStop,
		UnkFloat2:                         a.UnkFloat2,
		UnkAudioEvent:                     lookupThinHash(a.UnkAudioEvent),
		UnkBool2:                          a.UnkBool2,
		StatusEffectVolumeTemplate:        a.StatusEffectVolumeTemplate,
		EffectDamageType:                  a.EffectDamageType,
		FireTemplate:                      a.FireTemplate,
		UnkBitfield:                       a.UnkBitfield,
		RaycastTemplate:                   a.RaycastTemplate,
		UnkAudioEvent2:                    lookupThinHash(a.UnkAudioEvent2),
		UnkBitfield2:                      a.UnkBitfield2,
		UnkVector:                         a.UnkVector,
	}
}

func LoadProjectileSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]ProjectileInfo, error) {
	r := bytes.NewReader(projectileSettings)

	infos := make([]ProjectileInfo, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("ProjectileSettings") {
			return nil, fmt.Errorf("invalid projectile settings file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		var rawSettings DLArray
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("reading projectile info offset and count: %v", err)
		}

		rawSetting := make([]rawProjectileInfo, rawSettings.Count)
		r.Seek(base+rawSettings.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading projectile info array: %v", err)
		}

		for _, info := range rawSetting {
			infos = append(infos, info.Resolve(lookupHash, lookupThinHash, lookupStrings))
		}
	}
	return infos, nil
}
