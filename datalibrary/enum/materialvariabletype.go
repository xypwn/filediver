package enum

type MaterialVariableType uint32

const (
	MaterialVariableType_Scalar MaterialVariableType = iota
	MaterialVariableType_Vector2
	MaterialVariableType_Vector3
	MaterialVariableType_Vector4
)

func (p MaterialVariableType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=MaterialVariableType
