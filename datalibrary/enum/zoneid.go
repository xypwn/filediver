package enum

type ZoneId uint32

const (
	ZONE_ID_UNKNOWN ZoneId = iota
	ZoneId_Value_1_Len_25
	ZoneId_Value_2_Len_33
	ZoneId_Value_3_Len_23
	ZoneId_Value_4_Len_31
	ZoneId_Value_5_Len_35
	ZoneId_Value_6_Len_38
	ZoneId_Value_7_Len_41
	ZoneId_Value_8_Len_45
	ZoneId_Value_9_Len_48
	ZoneId_Value_10_Len_52
	ZoneId_Value_11_Len_42
	ZoneId_Value_12_Len_29
	ZoneId_Value_13_Len_40
	ZoneId_Value_14_Len_39
	ZoneId_Value_15_Len_39
	ZoneId_Value_16_Len_37
	ZoneId_Value_17_Len_39
	ZoneId_Value_18_Len_38
	ZoneId_Value_19_Len_37
	ZoneId_Value_20_Len_29
	ZoneId_Value_21_Len_39
	ZoneId_Value_22_Len_36
	ZoneId_Value_23_Len_39
	ZoneId_Value_24_Len_43
	ZoneId_Value_25_Len_29
	ZoneId_Value_26_Len_40
	ZoneId_Value_27_Len_27
	ZoneId_Value_28_Len_29
	ZoneId_Value_29_Len_30
	ZoneId_Value_30_Len_39
	ZoneId_Value_31_Len_36
	ZoneId_Value_32_Len_40
	ZoneId_Value_33_Len_37
	ZoneId_Value_34_Len_45
	ZoneId_Value_35_Len_37
	ZoneId_Value_36_Len_33
	ZoneId_Value_37_Len_27
	ZoneId_Value_38_Len_28
	ZoneId_Value_39_Len_31
	ZoneId_Value_40_Len_36
	ZoneId_Value_41_Len_36
	ZoneId_Value_42_Len_34
	ZoneId_Value_43_Len_32
	ZoneId_Value_44_Len_24
	ZoneId_Value_45_Len_23
	ZoneId_Value_46_Len_28
	ZoneId_Value_47_Len_24
	ZoneId_Value_48_Len_31
	ZoneId_Value_49_Len_24
	ZoneId_Value_50_Len_39
	ZoneId_Value_51_Len_24
	ZoneId_Value_52_Len_27
	ZoneId_Value_53_Len_47
	ZoneId_Value_54_Len_44
	ZoneId_Value_55_Len_36
	ZoneId_Value_56_Len_36
	ZoneId_Value_57_Len_42
	ZoneId_Value_58_Len_47
	ZoneId_Value_59_Len_36
	ZoneId_Value_60_Len_38
	ZoneId_Value_61_Len_35
	ZoneId_Value_62_Len_35
	ZoneId_Value_63_Len_28
	ZoneId_Value_64_Len_39
	ZoneId_Value_65_Len_30
	ZoneId_Value_66_Len_33
	ZoneId_Value_67_Len_41
	ZoneId_Value_68_Len_32
	ZoneId_Value_69_Len_30
	ZoneId_Value_70_Len_38
	ZoneId_Value_71_Len_29
	ZoneId_Value_72_Len_28
	ZoneId_Value_73_Len_30
	ZoneId_Value_74_Len_29
	ZoneId_Value_75_Len_29
	ZoneId_Value_76_Len_37
	ZoneId_Value_77_Len_31
	ZoneId_Value_78_Len_33
	ZoneId_Value_79_Len_32
	ZoneId_Value_80_Len_30
	ZoneId_Value_81_Len_35
	ZoneId_Value_82_Len_24
	ZoneId_Value_83_Len_26
	ZoneId_Value_84_Len_31
	ZoneId_Value_85_Len_27
	ZoneId_Value_86_Len_25
	ZoneId_Value_87_Len_23
	ZONE_ID_wall_stamps
	ZONE_ID_NUM_ZONES
)

func (p ZoneId) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ZoneId
