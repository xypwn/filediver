package enum

type WeaponFunctionType uint32

const (
	WeaponFunctionType_None WeaponFunctionType = iota
	WeaponFunctionType_Zeroing
	WeaponFunctionType_ROF
	WeaponFunctionType_Firemode
	WeaponFunctionType_Magazine
	WeaponFunctionType_LightMode
	WeaponFunctionType_LaserGuide
	WeaponFunctionType_MuzzleVelocity
	WeaponFunctionType_ProgrammableAmmo
	WeaponFunctionType_LaserPrism
	WeaponFunctionType_Value_10_Len_9
)

func (p WeaponFunctionType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponFunctionType
