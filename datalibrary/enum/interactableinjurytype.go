package enum

type InteractableInjuryType uint32

const (
	InteractableInjuryType_None InteractableInjuryType = iota
	InteractableInjuryType_BothArms
	InteractableInjuryType_OneArm
	InteractableInjuryType_BothLegs
	InteractableInjuryType_OneLeg
	InteractableInjuryType_BothArmsAndLegs
)

func (p InteractableInjuryType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=InteractableInjuryType
