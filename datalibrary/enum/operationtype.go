package enum

type OperationType uint32

const (
	OperationType_Value_0_Len_21 OperationType = iota
	OperationType_Value_1_Len_31
	OperationType_Value_2_Len_34
	OperationType_Value_3_Len_24
	OperationType_Value_4_Len_25
	OperationType_Value_5_Len_24
	OperationType_Value_6_Len_24
	OperationType_Value_7_Len_27
	OperationType_Value_8_Len_25
	OperationType_Value_9_Len_26
	OperationType_Value_10_Len_34
	OperationType_Value_11_Len_34
	OperationType_Value_12_Len_31
	OperationType_Value_13_Len_33
	OperationType_None
)

func (p OperationType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=OperationType
