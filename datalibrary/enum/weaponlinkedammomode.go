package enum

type WeaponLinkedAmmoMode uint32

const (
	WeaponLinkedAmmoMode_Value_0_Len_33 WeaponLinkedAmmoMode = iota
	WeaponLinkedAmmoMode_Value_1_Len_34
	WeaponLinkedAmmoMode_Value_2_Len_27
)

func (p WeaponLinkedAmmoMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=WeaponLinkedAmmoMode
