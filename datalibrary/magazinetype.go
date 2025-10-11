package datalib

type MagazineType uint32

const (
	MagazineType_Uniform MagazineType = iota
	MagazineType_Pattern
)

func (p MagazineType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=MagazineType
