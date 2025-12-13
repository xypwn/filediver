package enum

type AiJob uint32

const (
	AiJob_Guard AiJob = iota
	AiJob_Patrol
	AiJob_GuardForce
	AiJob_Encounter
	AiJob_ProducedFighter
	AiJob_Convoy
	AiJob_Convoy_Defending
	AiJob_Border_Travelling
	AiJob_Value_8_Len_15
	AiJob_Value_9_Len_16
	AiJob_Count
	AiJob_Any
)

func (p AiJob) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=AiJob
