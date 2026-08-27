package enum

type BanterEventType uint8

const (
	BanterEventType_None BanterEventType = iota
	BanterEventType_IdleOnMission
	BanterEventType_IdleOnShip
	BanterEventType_IdleYawn
	BanterEventType_MissionStartReady
	BanterEventType_ExtractionWaiting
	BanterEventType_MissionNearbyHelldiver
	BanterEventType_ShipNearbyHelldiver
	BanterEventType_InLoadoutHellpod
	BanterEventType_Customization
	BanterEventType_MeetNewHelldiver
	BanterEventType_ApproachExtraction
	BanterEventType_WaitingForObjective
	BanterEventType_ObjectiveComplete
	BanterEventType_ApproachObjective
	BanterEventType_LocationCarnage
	BanterEventType_Count
)

func (p BanterEventType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=BanterEventType
