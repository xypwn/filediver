package enum

type RingSortingTypes uint8

const (
	RingSortingTypes_saturn_ring RingSortingTypes = iota
	RingSortingTypes_astroid_belt
)

func (p RingSortingTypes) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=RingSortingTypes
