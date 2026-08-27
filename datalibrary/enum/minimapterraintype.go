package enum

type MinimapTerrainType uint8

const (
	MinimapTerrainType_Open MinimapTerrainType = iota
	MinimapTerrainType_Difficult
	MinimapTerrainType_Rocky
	MinimapTerrainType_Vegetation
	MinimapTerrainType_Impassable
	MinimapTerrainType_Bug
)

func (p MinimapTerrainType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=MinimapTerrainType
