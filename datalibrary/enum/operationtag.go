package enum

type OperationTag uint32

const (
	OperationTag_None OperationTag = iota
	OperationTag_Value_1_Len_23
	OperationTag_Value_2_Len_34
	OperationTag_Value_3_Len_25
	OperationTag_Value_4_Len_30
	OperationTag_Value_5_Len_27
	OperationTag_Value_6_Len_24
	OperationTag_Value_7_Len_27
	OperationTag_Value_8_Len_27
	OperationTag_Count
)

func (p OperationTag) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=OperationTag
