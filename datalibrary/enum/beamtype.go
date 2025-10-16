package enum

type BeamType uint32

const (
	BeamType_None BeamType = iota
	BeamType_LaserCannon
	BeamType_BotTorch
	BeamType_Medium_Focused
	BeamType_Medium_Wide
	BeamType_Medium
	BeamType_ObserverEye
	BeamType_LaserCannon_Turret
	BeamType_Small
	BeamType_TripodEye
	BeamType_PhysGun
	BeamType_Unknown_Len_6
	BeamType_Large_Heavy
	BeamType_Large_Light
	BeamType_Large_Focused
	BeamType_Large_Unfocused
	BeamType_Large
	BeamType_Unknown_Len_11
	BeamType_Unknown_Len_15
	BeamType_OutcastFlashbeamChargeup
	BeamType_Unknown_Len_14
	BeamType_Count
)

func (p BeamType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BeamType
