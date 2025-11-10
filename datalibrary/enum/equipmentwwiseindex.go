package enum

type EquipmentWwiseIndex int32

const (
	EquipmentWwiseIndex_Carryable_FuelBarrel EquipmentWwiseIndex = iota - 4
	EquipmentWwiseIndex_Carryable_Briefcase
	EquipmentWwiseIndex_Carryable_Blackbox
	EquipmentWwiseIndex_Carryable_Flag
	EquipmentWwiseIndex_None
	EquipmentWwiseIndex_Primary_RifleSm
	EquipmentWwiseIndex_Primary_RifleBig
	EquipmentWwiseIndex_Primary_SMG
	EquipmentWwiseIndex_Primary_Energy
	EquipmentWwiseIndex_Primary_Shotgun
	EquipmentWwiseIndex_Sidearm_Pistol
	EquipmentWwiseIndex_Support_GrenadeLauncher
	EquipmentWwiseIndex_Support_RecoillessRifle
	EquipmentWwiseIndex_Support_MissileLauncher
	EquipmentWwiseIndex_Support_Machinegun
	EquipmentWwiseIndex_Support_Sniper
	EquipmentWwiseIndex_Support_Chemgun
	EquipmentWwiseIndex_Support_Arc_Thrower
	EquipmentWwiseIndex_Melee_Heavy
	EquipmentWwiseIndex_Melee_Light
	EquipmentWwiseIndex_Melee_Spear
	EquipmentWwiseIndex_Melee_Flag
)

func (p EquipmentWwiseIndex) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EquipmentWwiseIndex
