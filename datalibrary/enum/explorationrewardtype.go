package enum

type ExplorationRewardType uint32

const (
	ExplorationRewardType_None ExplorationRewardType = iota
	ExplorationRewardType_exploration_requisition_slips
	ExplorationRewardType_exploration_medals
	ExplorationRewardType_exploration_credit_card
	ExplorationRewardType_Count
)

func (p ExplorationRewardType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ExplorationRewardType
