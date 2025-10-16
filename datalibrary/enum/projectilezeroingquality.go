package enum

type ProjectileZeroingQuality uint32

const (
	ProjectileZeroingQuality_None ProjectileZeroingQuality = iota
	ProjectileZeroingQuality_Cheap
	ProjectileZeroingQuality_Normal
	ProjectileZeroingQuality_High
	ProjectileZeroingQuality_Perfect
	ProjectileZeroingQuality_Count
)

func (p ProjectileZeroingQuality) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ProjectileZeroingQuality
