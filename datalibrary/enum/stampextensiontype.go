package enum

type StampExtensionType uint8

const (
	StampExtension_Expand StampExtensionType = iota
	StampExtension_Location
)

func (p StampExtensionType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=StampExtensionType
