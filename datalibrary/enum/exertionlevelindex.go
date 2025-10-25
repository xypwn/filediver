package enum

type ExertionLevelIndex uint32

const (
	ExertionLevelIndex_None ExertionLevelIndex = iota
	ExertionLevelIndex_Inconsiderable
	ExertionLevelIndex_Low
	ExertionLevelIndex_Medium
	ExertionLevelIndex_High
	ExertionLevelIndex_Injured
	ExertionLevelIndex_Hemmorage
	ExertionLevelIndex_HemmorageCritical
)

func (p ExertionLevelIndex) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ExertionLevelIndex
