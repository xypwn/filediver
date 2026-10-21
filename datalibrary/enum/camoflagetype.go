package enum

type CamouflageType uint32

const (
	CamouflageType_None CamouflageType = iota
	CamouflageType_Default
	CamouflageType_Forest
	CamouflageType_Desert
	CamouflageType_Arctic
	CamouflageType_Crimson
)

func (p CamouflageType) Name() string {
	switch p {
	case CamouflageType_None:
		return "none"
	case CamouflageType_Default:
		return "default"
	case CamouflageType_Forest:
		return "forest"
	case CamouflageType_Desert:
		return "desert"
	case CamouflageType_Arctic:
		return "arctic"
	case CamouflageType_Crimson:
		return "crimson"
	}
	return "unknown"
}

func (p CamouflageType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=CamouflageType
