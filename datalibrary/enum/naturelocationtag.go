package enum

type NatureLocationTag uint8

const (
	NatureLocationTag_none NatureLocationTag = iota
	NatureLocationTag_crater
	NatureLocationTag_lake
	NatureLocationTag_canyon
)

func (p NatureLocationTag) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=NatureLocationTag
