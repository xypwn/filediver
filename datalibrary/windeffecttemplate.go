package datalib

type WindEffectTemplate uint32

const (
	WindEffectTemplate_None WindEffectTemplate = iota
	WindEffectTemplate_SmallExplosion
	WindEffectTemplate_MediumExplosion
	WindEffectTemplate_LargeExplosion
	WindEffectTemplate_HugeExplosion
	WindEffectTemplate_Hellpod
	WindEffectTemplate_BombImpact
	WindEffectTemplate_LargeWeaponFire
	WindEffectTemplate_LargeWeaponFireLoop
	WindEffectTemplate_ImpactMedium
	WindEffectTemplate_ImpactHeavy
)

func (p WindEffectTemplate) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WindEffectTemplate
