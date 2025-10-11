package datalib

type UnitEffectOrphanStrategy uint32

const (
	OrphanStrategyNone UnitEffectOrphanStrategy = iota
	OrphanStrategyDestroy
	OrphanStrategyStop
	OrphanStrategyDisconnect
)

func (p UnitEffectOrphanStrategy) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=UnitEffectOrphanStrategy
