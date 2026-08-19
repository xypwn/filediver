package enum

type ZoneMusicType uint8

const (
	ZoneMusicType_None ZoneMusicType = iota
	ZoneMusicType_Bug
)

func (p ZoneMusicType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ZoneMusicType
