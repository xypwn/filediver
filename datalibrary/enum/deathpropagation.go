package enum

type DeathPropagation uint32

const (
	DeathPropagation_None DeathPropagation = iota
	DeathPropagation_ToParent
	DeathPropagation_ToChildren
	DeathPropagation_Both
)

func (p DeathPropagation) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DeathPropagation
