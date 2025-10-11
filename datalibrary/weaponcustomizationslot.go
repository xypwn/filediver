package datalib

type WeaponCustomizationSlot uint32

const (
	WeaponCustomizationSlot_None WeaponCustomizationSlot = iota
	WeaponCustomizationSlot_Underbarrel
	WeaponCustomizationSlot_Optics
	WeaponCustomizationSlot_PaintScheme
	WeaponCustomizationSlot_Muzzle
	WeaponCustomizationSlot_Magazine
	WeaponCustomizationSlot_AmmoType
	WeaponCustomizationSlot_AmmoTypeAlternate
	WeaponCustomizationSlot_Internals
	WeaponCustomizationSlot_Triggers
	WeaponCustomizationSlot_Count
)

func (p WeaponCustomizationSlot) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponCustomizationSlot
