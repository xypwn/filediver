package datalib

type WeaponCustomizationSortGroups uint32

const (
	WeaponCustomizationSortGroups_None WeaponCustomizationSortGroups = iota
	WeaponCustomizationSortGroups_Paint_Default
	WeaponCustomizationSortGroups_Paint_Solid
	WeaponCustomizationSortGroups_Paint_Camo
	WeaponCustomizationSortGroups_Count
)

func (p WeaponCustomizationSortGroups) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponCustomizationSortGroups
