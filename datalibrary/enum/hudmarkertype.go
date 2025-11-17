package enum

type HudMarkerType uint32

const (
	HudMarkerType_Generic HudMarkerType = iota
	HudMarkerType_Enemy
	HudMarkerType_EnemyHeavy
	HudMarkerType_EnemyMassive
	HudMarkerType_EnemyFlyer
	HudMarkerType_Squad
	HudMarkerType_SquadHeavy
	HudMarkerType_SquadMassive
	HudMarkerType_SquadFlyer
	HudMarkerType_SquadMember
	HudMarkerType_Item
	HudMarkerType_Sample
	HudMarkerType_Intel
	HudMarkerType_Vehicle
	HudMarkerType_Turret
	HudMarkerType_DefensiveEmplacement
	HudMarkerType_Mines
	HudMarkerType_Value_17_Len_8
	HudMarkerType_Objective
	HudMarkerType_ObjectiveCode
	HudMarkerType_Stratagem
	HudMarkerType_Map
	HudMarkerType_AvatarReloadRequest
	HudMarkerType_AvatarDowned
	HudMarkerType_AvatarComms
	HudMarkerType_Count
)

func (p HudMarkerType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=HudMarkerType
