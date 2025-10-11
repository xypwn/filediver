package datalib

type ChargeState uint32

const (
	ChargeState_MinCharge ChargeState = iota
	ChargeState_FullCharge
	ChargeState_OverCharge
	ChargeState_Count
)

func (p ChargeState) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ChargeState
