package enum

type ArcType uint32

const (
	ArcType_None ArcType = iota
	ArcType_IlluminateTripod
	ArcType_ArcThrower_MK3
	ArcType_ArcShotgun
	ArcType_IlluminateObserverLightning
	ArcType_Value_5_Len_21
	ArcType_Value_6_Len_10
	ArcType_ArcThrower
	ArcType_Value_8_Len_8
	ArcType_IlluminateObelisk
	ArcType_ArcThrower_MK2
	ArcType_IlluminateSummonerSpawnEffect
	ArcType_TeslaTurret
	ArcType_ArcThrower_MK4
	ArcType_Value_14_Len_21
	ArcType_Count
)

func (p ArcType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ArcType
