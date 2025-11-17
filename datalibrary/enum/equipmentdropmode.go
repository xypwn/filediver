package enum

type EquipmentDropMode uint32

const (
	EquipmentDropMode_Decay EquipmentDropMode = iota
	EquipmentDropMode_Infinite
	EquipmentDropMode_InfiniteIfNonEmpty
)

func (p EquipmentDropMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EquipmentDropMode
