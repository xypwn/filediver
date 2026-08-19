package enum

type RotationType uint32

const (
	RotationType_None RotationType = iota
	RotationType_FollowWindGround
	RotationType_FollowWindFalling
	RotationType_AlignWithCamera
)

func (p RotationType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=RotationType
