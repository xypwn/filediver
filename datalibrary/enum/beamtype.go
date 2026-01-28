package enum

type BeamType uint32

const (
	BeamType_None BeamType = iota
	BeamType_LaserCannon
	BeamType_BotTorch
	BeamType_ShoulderMountedCamera
	BeamType_TripodEye_Wide
	BeamType_Medium_Wide
	BeamType_Medium
	BeamType_ObserverEye
	BeamType_LaserCannon_Turret
	BeamType_Small
	BeamType_TripodEye
	BeamType_PhysGun
	BeamType_Turret
	BeamType_Large_Heavy
	BeamType_Large_Light
	BeamType_Value_15_Len_33
	BeamType_Value_16_Len_22
	BeamType_Wm_Cannon
	BeamType_TripodEye_Heavy
	BeamType_Large
	BeamType_Drone_Laser
	BeamType_Value_21_Len_24
	BeamType_OutcastFlashbeamChargeup
	BeamType_Turret_Tactical
	BeamType_Medium_Focused
	BeamType_Count
)

func (p BeamType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BeamType
