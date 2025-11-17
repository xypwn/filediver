package enum

type DepositRefill uint32

const (
	DepositRefill_None DepositRefill = iota
	DepositRefill_Ammo
	DepositRefill_Medic
	DepositRefill_Supplies
	DepositRefill_Count
)

func (p DepositRefill) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DepositRefill
