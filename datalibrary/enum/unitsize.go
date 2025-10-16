package enum

type UnitSize uint32

const (
	UnitSize_Small UnitSize = iota
	UnitSize_Medium
	UnitSize_Large
	UnitSize_Massive
	UnitSize_Num
)

func (p UnitSize) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=UnitSize
