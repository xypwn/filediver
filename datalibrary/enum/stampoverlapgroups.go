package enum

type StampOverlapGroups uint8

const (
	StampOverlapGroup_None StampOverlapGroups = iota
	StampOverlapGroups_Value_1_Len_26
	StampOverlapGroup_A
	StampOverlapGroup_B
	StampOverlapGroup_C
	StampOverlapGroup_D
	StampOverlapGroup_E
	StampOverlapGroup_F
	StampOverlapGroup_G
	StampOverlapGroups_Value_9_Len_24
)

func (p StampOverlapGroups) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=StampOverlapGroups
