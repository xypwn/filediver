package enum

type MountSide uint32

const (
	MountSide_Common MountSide = iota
	MountSide_Right
	MountSide_Left
)

func (p MountSide) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=MountSide
