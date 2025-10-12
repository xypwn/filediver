package enum

type NoiseTemplate uint32

const (
	NoiseTemplate_None NoiseTemplate = iota
	NoiseTemplate_Toot
	NoiseTemplate_Engine
	NoiseTemplate_EnemyWarning
	NoiseTemplate_GunshipWarning
	NoiseTemplate_Footsteps
	NoiseTemplate_Weapon_Small
	NoiseTemplate_Weapon_Medium
	NoiseTemplate_Weapon_Large
	NoiseTemplate_Weapon_Huge
	NoiseTemplate_Supressor_Small
	NoiseTemplate_Supressor_Medium
	NoiseTemplate_Supressor_Large
	NoiseTemplate_Supressor_Huge
	NoiseTemplate_Small_Explosion_Frag
	NoiseTemplate_Medium_Explosion_Vehicle
	NoiseTemplate_Large_Explosion_Orbital
	NoiseTemplate_Huge_Explosion_Hellbomb
	NoiseTemplate_BeamHit
	NoiseTemplate_ProjectileHit
	NoiseTemplate_ProjectileHitSuppressed
	NoiseTemplate_Objective_Large_Frequent
	NoiseTemplate_Objective_Large_Seldom
	NoiseTemplate_Objective_Encounter
	NoiseTemplate_Objective_Flag
	NoiseTemplate_StratagemBeacon
	NoiseTemplate_Extraction
	NoiseTemplate_CapitalTower
	NoiseTemplate_Value_28_Len_20
	NoiseTemplate_Max
)

func (p NoiseTemplate) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=NoiseTemplate
