package enum

type SubRegionType uint32

const (
	SubRegionType_None SubRegionType = iota
	SubRegionType_Blocker_Bottom
	SubRegionType_Blocker_Bottom_B
	SubRegionType_Blocker_Bottom_C
	SubRegionType_Blocker_General
	SubRegionType_Blocker_General_B
	SubRegionType_Blocker_General_C
	SubRegionType_A
	SubRegionType_B
	SubRegionType_C
	SubRegionType_D
	SubRegionType_E
	SubRegionType_F
	SubRegionType_G
	SubRegionType_H
	SubRegionType_I
	SubRegionType_J
	SubRegionType_ExtraResources
	SubRegionType_Bugs
	SubRegionType_Value_19_Len_26
	SubRegionType_Value_20_Len_25
	SubRegionType_Cyborgs
	SubRegionType_Value_22_Len_24
	SubRegionType_Illuminate
	SubRegionType_Value_24_Len_20
	SubRegionType_Value_25_Len_25
	SubRegionType_Value_26_Len_27
	SubRegionType_Value_27_Len_27
	SubRegionType_Value_28_Len_27
	SubRegionType_Value_29_Len_27
	SubRegionType_Value_30_Len_27
	SubRegionType_Value_31_Len_27
	SubRegionType_Value_32_Len_27
	SubRegionType_Value_33_Len_27
	SubRegionType_Value_34_Len_27
	SubRegionType_Value_35_Len_18
	SubRegionType_Value_36_Len_35
	SubRegionType_Value_37_Len_35
	SubRegionType_Value_38_Len_35
	SubRegionType_Value_39_Len_35
	SubRegionType_Value_40_Len_19
)

func (p SubRegionType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SubRegionType
