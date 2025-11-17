package enum

type EquipmentType uint32

const (
	EquipmentType_All EquipmentType = iota
	EquipmentType_AssaultRifle
	EquipmentType_MarksmanRifle
	EquipmentType_SniperRifle
	EquipmentType_SMG
	EquipmentType_MachineGun
	EquipmentType_Shotgun
	EquipmentType_Explosive
	EquipmentType_Energy
	EquipmentType_Incendiary
	EquipmentType_Pistol
	EquipmentType_Revolver
	EquipmentType_EnergySidearm
	EquipmentType_LauncherSidearm
	EquipmentType_StandardGrenade
	EquipmentType_SpecialGrenade
	EquipmentType_Backpack
	EquipmentType_BackpackSupport
	EquipmentType_BackpackJumppack
	EquipmentType_BackpackMedic
	EquipmentType_BackpackDrone
	EquipmentType_BackpackEnergyShield
	EquipmentType_BackpackSupply
	EquipmentType_BackpackDisplacement
	EquipmentType_BackpackShield
	EquipmentType_Value_25_Len_31
	EquipmentType_Value_26_Len_30
	EquipmentType_Objective
	EquipmentType_Value_28_Len_29
	EquipmentType_Melee
	EquipmentType_Value_30_Len_21
	EquipmentType_Hidden
	EquipmentType_Count
)

func (p EquipmentType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EquipmentType
