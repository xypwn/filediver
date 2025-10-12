package enum

type CrosshairWeaponType uint32

const (
	CrosshairWeaponType_Default CrosshairWeaponType = iota
	CrosshairWeaponType_CrosshairNever
	CrosshairWeaponType_CrosshairAlways
	CrosshairWeaponType_AssaultRifle
	CrosshairWeaponType_PumpShotgun
	CrosshairWeaponType_AutoShotgun
	CrosshairWeaponType_Pistol
	CrosshairWeaponType_Revolver
	CrosshairWeaponType_Laser
	CrosshairWeaponType_LaserCannon
	CrosshairWeaponType_LaserPulseCannon
	CrosshairWeaponType_Railgun
	CrosshairWeaponType_Javelin
	CrosshairWeaponType_Throwable
	CrosshairWeaponType_Physgun
	CrosshairWeaponType_Value_15
	CrosshairWeaponType_Value_16
	CrosshairWeaponType_Value_17
	CrosshairWeaponType_Value_18
	CrosshairWeaponType_None
	CrosshairWeaponType_Count
)

func (p CrosshairWeaponType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=CrosshairWeaponType
