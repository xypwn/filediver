package enum

type ZoneTag uint8

const (
	ZoneTag_None ZoneTag = iota
	ZoneTag_Value_1_Len_14
	ZoneTag_Value_2_Len_14
	ZoneTag_Value_3_Len_12
	ZoneTag_Value_4_Len_15
	ZoneTag_Value_5_Len_12
	ZoneTag_Value_6_Len_11
	ZoneTag_Value_7_Len_11
	ZoneTag_Value_8_Len_18
	ZoneTag_Value_9_Len_13
	ZoneTag_Value_10_Len_16
	ZoneTag_Value_11_Len_14
	ZoneTag_Value_12_Len_14
	ZoneTag_Value_13_Len_13
	ZoneTag_Value_14_Len_21
	ZoneTag_Value_15_Len_18
	ZoneTag_Value_16_Len_15
	ZoneTag_Value_17_Len_34
	ZoneTag_Count
)

func (p ZoneTag) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ZoneTag
