package enum

type PatrolSetting uint32

const (
	PatrolSetting_Value_0_Len_21 PatrolSetting = iota
	PatrolSetting_Value_1_Len_20
	PatrolSetting_Value_2_Len_18
	PatrolSetting_None
	PatrolSetting_Count
)

func (p PatrolSetting) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=PatrolSetting
