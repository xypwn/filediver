package enum

type HellpodVariant uint32

const (
	HellpodVariant_Payload HellpodVariant = iota
	HellpodVariant_Reinforce
	HellpodVariant_InitialSpawn
)

func (p HellpodVariant) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=HellpodVariant
