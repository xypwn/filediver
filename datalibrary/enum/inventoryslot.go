package enum

type InventorySlot uint32

const (
	InventorySlot_Value_0_Len_18 InventorySlot = iota
	InventorySlot_Primary
	InventorySlot_Sidearm
	InventorySlot_Support
	InventorySlot_Grenade
	InventorySlot_Value_5_Len_27
	InventorySlot_Backpack
	InventorySlot_Value_7_Len_18
	InventorySlot_Value_8_Len_19
)

func (p InventorySlot) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=InventorySlot
