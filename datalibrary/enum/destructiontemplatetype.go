package enum

type DestructionTemplateType uint32

const (
	DestructionTemplateType_none DestructionTemplateType = iota
	DestructionTemplateType_SmallArms
	DestructionTemplateType_FragGrenade_HeavyMachinegun_Sniper_Car_Mech
	DestructionTemplateType_Autocannon_Charger_Recoilless_LAV
	DestructionTemplateType_Strider_EagleRockets
	DestructionTemplateType_OrbitalBombardment
	DestructionTemplateType_Hellbomb
	DestructionTemplateType_ForceDestroyTrees
	DestructionTemplateType_ExplodingBarrel
	DestructionTemplateType_Script
	DestructionTemplateType_Count
)

func (p DestructionTemplateType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=DestructionTemplateType
