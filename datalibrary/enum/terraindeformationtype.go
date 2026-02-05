package enum

type TerrainDeformationType uint8

const (
	TerrainDeformationType_None TerrainDeformationType = iota
	TerrainDeformationType_Default
	TerrainDeformationType_Nuke
	TerrainDeformationType_Large
	TerrainDeformationType_Medium
	TerrainDeformationType_Small
	TerrainDeformationType_Scorch
	TerrainDeformationType_FlattenHeightmapAO
	TerrainDeformationType_Value_8_Len_29
	TerrainDeformationType_Value_9_Len_43
	TerrainDeformationType_Value_10_Len_40
	TerrainDeformationType_Value_11_Len_44
	TerrainDeformationType_Value_12_Len_45
	TerrainDeformationType_Value_13_Len_44
	TerrainDeformationType_Count
)

func (p TerrainDeformationType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=TerrainDeformationType
