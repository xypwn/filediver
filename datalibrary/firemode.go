package datalib

type FireMode uint32

const (
	FireMode_None FireMode = iota
	FireMode_Automatic
	FireMode_Single
	FireMode_Burst
	FireMode_Safety_Off        // I don't think these are right
	FireMode_Charge_Safety_On  // I don't think these are right
	FireMode_Charge_Safety_Off // I don't think these are right
	FireMode_Safety_On         // I don't think these are right
	FireMode_Count
)

func (p FireMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=FireMode
