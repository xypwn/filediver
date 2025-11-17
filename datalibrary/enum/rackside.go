package enum

type RackSide uint32

const (
	RackSide_None RackSide = iota
	RackSide_Right
	RackSide_Left
)

func (p RackSide) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=RackSide
