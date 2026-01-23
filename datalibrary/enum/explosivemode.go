package enum

type ExplosiveMode uint32

const (
	ExplosiveMode_Timed ExplosiveMode = iota
	ExplosiveMode_Impact
	ExplosiveMode_External
	ExplosiveMode_TimedStatusEffect
	ExplosiveMode_Unknown_32_Chars
)

func (p ExplosiveMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ExplosiveMode
