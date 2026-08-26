package enum

type AllowLocationMode uint8

const (
	AllowLocationMode_All AllowLocationMode = iota
	AllowLocationMode_SmallOnly
	AllowLocationMode_None
)

func (p AllowLocationMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=AllowLocationMode
