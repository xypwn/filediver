package enum

type DangerLevel uint32

const (
	DangerLevel_None DangerLevel = iota
	DangerLevel_Projectile
	DangerLevel_Melee
	DangerLevel_Explosion
	DangerLevel_Damage
	DangerLevel_Beam
	DangerLevel_Grenade
	DangerLevel_Impact
	DangerLevel_Value_8_Len_15
	DangerLevel_Count
)

func (p DangerLevel) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DangerLevel
