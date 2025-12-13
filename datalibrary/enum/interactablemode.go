package enum

type InteractableMode uint32

const (
	InteractableMode_OwnerOnly InteractableMode = iota
	InteractableMode_Everyone
	InteractableMode_EveryoneButOwner
)

func (p InteractableMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=InteractableMode
