package datalib

type WeaponCustomizationIKAttach uint32

const (
	WeaponCustomizationIKAttach_None WeaponCustomizationIKAttach = iota
	WeaponCustomizationIKAttach_Optics
	WeaponCustomizationIKAttach_Underbarrel
	WeaponCustomizationIKAttach_Muzzle
	WeaponCustomizationIKAttach_Magazine
	WeaponCustomizationIKAttach_Count
)

func (p WeaponCustomizationIKAttach) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponCustomizationIKAttach
