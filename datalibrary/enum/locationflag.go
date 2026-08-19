package enum

type LocationFlag uint8

const (
	LocationFlag_None LocationFlag = iota
	LocationFlag_Value_1_Len_19
	LocationFlag_Value_2_Len_19
	LocationFlag_Value_3_Len_17
	LocationFlag_Value_4_Len_17
	LocationFlag_Value_5_Len_26
	LocationFlag_Count
)

func (p LocationFlag) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=LocationFlag
