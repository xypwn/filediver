package enum

type ElementType uint32

const (
	ElementType_None ElementType = iota
	ElementType_Fire
	ElementType_Electricity
	ElementType_Acid
	ElementType_Bleed
	ElementType_Gas
	ElementType_Value_6_Len_14
	ElementType_Count
)

func (p ElementType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ElementType
