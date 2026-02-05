package enum

type StampSplineType uint32

const (
	StampSpline_Path StampSplineType = iota
	StampSpline_SuperEarthRoad
	StampSpline_BugRoad
	StampSpline_BotRoad
	StampSpline_IlluminateRoad
	StampSpline_Count
	StampSpline_Any
)

func (p StampSplineType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=StampSplineType
