package enum

type HitEffectDamageType uint32

const (
	HitEffectDamageType_None HitEffectDamageType = iota
	HitEffectDamageType_PiercingSmall
	HitEffectDamageType_PiercingMedium
	HitEffectDamageType_PiercingLarge
	HitEffectDamageType_PiercingLargeHEAT
	HitEffectDamageType_IncendiarySmall
	HitEffectDamageType_IncendiaryMedium
	HitEffectDamageType_BeamSmall
	HitEffectDamageType_BeamLarge
	HitEffectDamageType_Blunt
	HitEffectDamageType_SlashingSmall
	HitEffectDamageType_SlashingLarge
	HitEffectDamageType_Value_12_Len_31
	HitEffectDamageType_Value_13_Len_37
	HitEffectDamageType_Value_14_Len_30
	HitEffectDamageType_Value_15_Len_29
	HitEffectDamageType_Value_16_Len_28
	HitEffectDamageType_Value_17_Len_26
	HitEffectDamageType_Count
)

func (p HitEffectDamageType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=HitEffectDamageType
