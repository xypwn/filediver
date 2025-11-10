package enum

type BackpackWwiseIndex int32

const (
	BackpackWwiseIndex_None BackpackWwiseIndex = iota
	BackpackWwiseIndex_Ammo
	BackpackWwiseIndex_AutomaticCannon
	BackpackWwiseIndex_SmallBallisticShield
	BackpackWwiseIndex_BallisticShield
	BackpackWwiseIndex_Displacement
	BackpackWwiseIndex_DroneMg
	BackpackWwiseIndex_EnergyShield
	BackpackWwiseIndex_Jumppack
	BackpackWwiseIndex_Medic
	BackpackWwiseIndex_RecoillessRifle
	BackpackWwiseIndex_Spear
	BackpackWwiseIndex_Support
	BackpackWwiseIndex_Chemicals
	BackpackWwiseIndex_Larva
)

func (p BackpackWwiseIndex) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BackpackWwiseIndex
