package enum

type ExplosionVerificationMode uint32

const (
	ExplosionVerificationMode_None ExplosionVerificationMode = iota
	ExplosionVerificationMode_OuterRadius
	ExplosionVerificationMode_All
)

func (p ExplosionVerificationMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ExplosionVerificationMode
