package enum

type DeathDecayMode uint32

const (
	DeathDecayMode_None DeathDecayMode = iota
	DeathDecayMode_Regular
	DeathDecayMode_Long
	DeathDecayMode_Instant
)

func (p DeathDecayMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DeathDecayMode
