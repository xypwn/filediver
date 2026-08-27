package enum

type CityPlacementHeuristic uint32

const (
	CityPlacementHeuristic_Value_0_Len_46 CityPlacementHeuristic = iota
	CityPlacementHeuristic_Value_1_Len_44
	CityPlacementHeuristic_Value_2_Len_39
	CityPlacementHeuristic_Value_3_Len_49
	CityPlacementHeuristic_Value_4_Len_50
	CityPlacementHeuristic_Value_5_Len_41
	CityPlacementHeuristic_Count
)

func (p CityPlacementHeuristic) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=CityPlacementHeuristic
