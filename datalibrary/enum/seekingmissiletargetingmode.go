package enum

type SeekingMissileTargetingMode uint32

const (
	SeekingMissileTargetingMode_Behavior SeekingMissileTargetingMode = iota
	SeekingMissileTargetingMode_Guidance
	SeekingMissileTargetingMode_TargetLock
	SeekingMissileTargetingMode_External
	SeekingMissileTargetingMode_Count
)

func (p SeekingMissileTargetingMode) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SeekingMissileTargetingMode
