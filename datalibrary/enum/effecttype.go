package enum

type EffectType uint32

const (
	EffectType_None EffectType = iota
	EffectType_Wet
	EffectType_Dry
	EffectType_Hot
	EffectType_Cold
	EffectType_BlownLeaves
)

func (p EffectType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EffectType
