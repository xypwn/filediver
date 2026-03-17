package enum

type RegionFlag uint32

const (
	RegionFlag_None RegionFlag = iota
	RegionFlag_Value_1_Len_17
	RegionFlag_Value_2_Len_17
	RegionFlag_Value_3_Len_15
	RegionFlag_Value_4_Len_15
	RegionFlag_Value_5_Len_24
	RegionFlag_Count
)

func (p RegionFlag) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=RegionFlag
