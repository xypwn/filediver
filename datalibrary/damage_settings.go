package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
)

type DamageStatusEffectInfo struct {
	Type  enum.StatusEffectType `json:"type"`
	Value float32               `json:"value"`
}

type rawDamageInfo struct {
	Type                     enum.DamageInfoType
	Damage                   int32
	DurableDamage            int32
	ArmorPenetrationPerAngle [4]uint32
	DemolitionStrength       uint32
	ForceStrength            uint32
	ForceImpulse             uint32
	ElementType              enum.ElementType
	StatusEffects            [4]DamageStatusEffectInfo
}

type DamageInfo struct {
	Type                     enum.DamageInfoType      `json:"type"`
	Damage                   int32                    `json:"damage"`
	DurableDamage            int32                    `json:"durable_damage"`
	ArmorPenetrationPerAngle []uint32                 `json:"armor_penetration_per_angle"`
	DemolitionStrength       uint32                   `json:"demolition_strength"`
	ForceStrength            uint32                   `json:"force_strength"`
	ForceImpulse             uint32                   `json:"force_impulse"`
	ElementType              enum.ElementType         `json:"element_type"`
	StatusEffects            []DamageStatusEffectInfo `json:"status_effects,omitempty"`
}

type DamageArmorAngles struct {
	Projectile [4]float32 `json:"projectile"`
	DPS        [4]float32 `json:"dps"`
}

type DamageSettings struct {
	Angles DamageArmorAngles `json:"angles"`
	Infos  []DamageInfo      `json:"infos"`
}

func (a rawDamageInfo) Resolve(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) DamageInfo {
	statusEffects := make([]DamageStatusEffectInfo, 0)
	for _, effect := range a.StatusEffects {
		if effect.Type == enum.StatusEffectType_None {
			break
		}
		statusEffects = append(statusEffects, effect)
	}
	return DamageInfo{
		Type:                     a.Type,
		Damage:                   a.Damage,
		DurableDamage:            a.DurableDamage,
		ArmorPenetrationPerAngle: a.ArmorPenetrationPerAngle[:],
		DemolitionStrength:       a.DemolitionStrength,
		ForceStrength:            a.ForceStrength,
		ForceImpulse:             a.ForceImpulse,
		ElementType:              a.ElementType,
		StatusEffects:            statusEffects,
	}
}

func LoadDamageSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) (*DamageSettings, error) {
	r := bytes.NewReader(damageSettings)

	var angles DamageArmorAngles
	infos := make([]DamageInfo, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("DamageSettings") && header.Type != Sum("DamageArmorAngles") {
			return nil, fmt.Errorf("invalid damage settings file")
		}

		base, _ := r.Seek(0, io.SeekCurrent)

		if header.Type == Sum("DamageArmorAngles") {
			if err := binary.Read(r, binary.LittleEndian, &angles); err != nil {
				return nil, fmt.Errorf("reading damage armor angles: %v", err)
			}
			continue
		}

		var rawSettings DLArray
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("reading damage info offset and count: %v", err)
		}

		rawSetting := make([]rawDamageInfo, rawSettings.Count)
		r.Seek(base+rawSettings.Offset, io.SeekStart)
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading damage info array: %v", err)
		}

		for _, info := range rawSetting {
			infos = append(infos, info.Resolve(lookupHash, lookupThinHash, lookupStrings))
		}
	}
	return &DamageSettings{
		Angles: angles,
		Infos:  infos,
	}, nil
}
