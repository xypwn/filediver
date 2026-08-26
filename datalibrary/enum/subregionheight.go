package enum

type SubRegionHeight uint32

const (
	SubRegionHeight_None SubRegionHeight = iota
	SubRegionHeight_Bottom_Flat
	SubRegionHeight_Low
	SubRegionHeight_Middle
	SubRegionHeight_High
	SubRegionHeight_Value_5_Len_32
	SubRegionHeight_All SubRegionHeight = 0xffffffff
)

func (p SubRegionHeight) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SubRegionHeight
