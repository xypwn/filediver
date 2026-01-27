package enum

type ProjectileStatusEffectVolumeTemplate uint32

const (
	ProjectileStatusEffectVolumeTemplate_None ProjectileStatusEffectVolumeTemplate = iota
	ProjectileStatusEffectVolumeTemplate_Flamethrower
	ProjectileStatusEffectVolumeTemplate_Acid_Stream
	ProjectileStatusEffectVolumeTemplate_Count
)

func (p ProjectileStatusEffectVolumeTemplate) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ProjectileStatusEffectVolumeTemplate
