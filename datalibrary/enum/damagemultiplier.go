package enum

type DamageMultiplier uint32

const (
	DamageMultiplier_None DamageMultiplier = iota
	DamageMultiplier_Critical
	DamageMultiplier_Normal
	DamageMultiplier_Reduced
	DamageMultiplier_Symbolic
)

func (p DamageMultiplier) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DamageMultiplier
