package enum

type WeaponCustomizationUnknownEnum uint32

const (
	WeaponCustomizationUnknownEnum_None WeaponCustomizationUnknownEnum = iota
	WeaponCustomizationUnknownEnum_Value_1_Len_9
	WeaponCustomizationUnknownEnum_Value_2_Len_8
	WeaponCustomizationUnknownEnum_Value_3_Len_5
	WeaponCustomizationUnknownEnum_Count
)

func (p WeaponCustomizationUnknownEnum) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponCustomizationUnknownEnum
