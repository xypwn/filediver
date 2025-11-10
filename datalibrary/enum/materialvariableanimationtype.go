package enum

type MaterialVariableAnimationType uint32

const (
	MaterialVariableAnimationType_Linear MaterialVariableAnimationType = iota
	MaterialVariableAnimationType_PingPong
)

func (p MaterialVariableAnimationType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=MaterialVariableAnimationType
