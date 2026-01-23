package enum

type RaycastTemplate uint32

const (
	RaycastTemplate_Damage RaycastTemplate = iota
	RaycastTemplate_Value_1_Len_37
	RaycastTemplate_Value_2_Len_31
	RaycastTemplate_DamageLos
	RaycastTemplate_IlluminateDamage
	RaycastTemplate_Aim
	RaycastTemplate_Ground
	RaycastTemplate_Value_7_Len_31
	RaycastTemplate_Climbable
	RaycastTemplate_BidirectionalClimbable
	RaycastTemplate_MultiHitClimbable
	RaycastTemplate_MotionSnag
	RaycastTemplate_Vision
	RaycastTemplate_Player
	RaycastTemplate_BidirectionalGround
	RaycastTemplate_SurfaceEffect
	RaycastTemplate_Flyer
	RaycastTemplate_Danger
	RaycastTemplate_GroundAlign
	RaycastTemplate_GroundAlignFindSurface
	RaycastTemplate_AimBlock
	RaycastTemplate_CameraCollision
	RaycastTemplate_Spotting
	RaycastTemplate_Terrain
	RaycastTemplate_Fire
	RaycastTemplate_GroundAll
	RaycastTemplate_DamageNoDebris
	RaycastTemplate_AimClosest
	RaycastTemplate_StaticObstacle
	RaycastTemplate_AnyStaticOnlyUniqueBodies
	RaycastTemplate_Value_30_Len_25
	RaycastTemplate_Count
)

func (p RaycastTemplate) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=RaycastTemplate
