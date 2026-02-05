package enum

type FireTemplate uint32

const (
	FireTemplate_None FireTemplate = iota
	FireTemplate_Small
	FireTemplate_Medium
	FireTemplate_Large
	FireTemplate_Massive
	FireTemplate_Count
)

func (p FireTemplate) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=FireTemplate
