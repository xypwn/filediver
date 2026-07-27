package asset_grading

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

var (
	HueTerrain1                 stingray.ThinHash = stingray.Sum("asset_grade_hue_terrain1").Thin()
	SaturationTerrain1          stingray.ThinHash = stingray.Sum("asset_grade_saturation_terrain1").Thin()
	ValueTerrain1               stingray.ThinHash = stingray.Sum("asset_grade_value_terrain1").Thin()
	ContrastTerrain1            stingray.ThinHash = stingray.Sum("asset_grade_contrast_terrain1").Thin()
	ContrastMidpointTerrain1    stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_terrain1").Thin()
	ColorTerrain1               stingray.ThinHash = stingray.Sum("asset_grade_color_terrain1").Thin()
	HueTerrain1Sec              stingray.ThinHash = stingray.Sum("asset_grade_hue_terrain1_sec").Thin()
	SaturationTerrain1Sec       stingray.ThinHash = stingray.Sum("asset_grade_saturation_terrain1_sec").Thin()
	ValueTerrain1Sec            stingray.ThinHash = stingray.Sum("asset_grade_value_terrain1_sec").Thin()
	ContrastTerrain1Sec         stingray.ThinHash = stingray.Sum("asset_grade_contrast_terrain1_sec").Thin()
	ContrastMidpointTerrain1Sec stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_terrain1_sec").Thin()
	ColorTerrain1Sec            stingray.ThinHash = stingray.Sum("asset_grade_color_terrain1_sec").Thin()
)
var Terrain1 []stingray.ThinHash = []stingray.ThinHash{
	HueTerrain1, SaturationTerrain1, ValueTerrain1, ContrastTerrain1, ContrastMidpointTerrain1, ColorTerrain1,
}
var Terrain1Sec []stingray.ThinHash = []stingray.ThinHash{
	HueTerrain1Sec, SaturationTerrain1Sec, ValueTerrain1Sec, ContrastTerrain1Sec, ContrastMidpointTerrain1Sec, ColorTerrain1Sec,
}
var (
	HueTerrain2                 stingray.ThinHash = stingray.Sum("asset_grade_hue_terrain2").Thin()
	SaturationTerrain2          stingray.ThinHash = stingray.Sum("asset_grade_saturation_terrain2").Thin()
	ValueTerrain2               stingray.ThinHash = stingray.Sum("asset_grade_value_terrain2").Thin()
	ContrastTerrain2            stingray.ThinHash = stingray.Sum("asset_grade_contrast_terrain2").Thin()
	ContrastMidpointTerrain2    stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_terrain2").Thin()
	ColorTerrain2               stingray.ThinHash = stingray.Sum("asset_grade_color_terrain2").Thin()
	HueTerrain2Sec              stingray.ThinHash = stingray.Sum("asset_grade_hue_terrain2_sec").Thin()
	SaturationTerrain2Sec       stingray.ThinHash = stingray.Sum("asset_grade_saturation_terrain2_sec").Thin()
	ValueTerrain2Sec            stingray.ThinHash = stingray.Sum("asset_grade_value_terrain2_sec").Thin()
	ContrastTerrain2Sec         stingray.ThinHash = stingray.Sum("asset_grade_contrast_terrain2_sec").Thin()
	ContrastMidpointTerrain2Sec stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_terrain2_sec").Thin()
	ColorTerrain2Sec            stingray.ThinHash = stingray.Sum("asset_grade_color_terrain2_sec").Thin()
)
var Terrain2 []stingray.ThinHash = []stingray.ThinHash{
	HueTerrain2, SaturationTerrain2, ValueTerrain2, ContrastTerrain2, ContrastMidpointTerrain2, ColorTerrain2,
}
var Terrain2Sec []stingray.ThinHash = []stingray.ThinHash{
	HueTerrain2Sec, SaturationTerrain2Sec, ValueTerrain2Sec, ContrastTerrain2Sec, ContrastMidpointTerrain2Sec, ColorTerrain2Sec,
}
var (
	HueTerrain3                 stingray.ThinHash = stingray.Sum("asset_grade_hue_terrain3").Thin()
	SaturationTerrain3          stingray.ThinHash = stingray.Sum("asset_grade_saturation_terrain3").Thin()
	ValueTerrain3               stingray.ThinHash = stingray.Sum("asset_grade_value_terrain3").Thin()
	ContrastTerrain3            stingray.ThinHash = stingray.Sum("asset_grade_contrast_terrain3").Thin()
	ContrastMidpointTerrain3    stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_terrain3").Thin()
	ColorTerrain3               stingray.ThinHash = stingray.Sum("asset_grade_color_terrain3").Thin()
	HueTerrain3Sec              stingray.ThinHash = stingray.Sum("asset_grade_hue_terrain3_sec").Thin()
	SaturationTerrain3Sec       stingray.ThinHash = stingray.Sum("asset_grade_saturation_terrain3_sec").Thin()
	ValueTerrain3Sec            stingray.ThinHash = stingray.Sum("asset_grade_value_terrain3_sec").Thin()
	ContrastTerrain3Sec         stingray.ThinHash = stingray.Sum("asset_grade_contrast_terrain3_sec").Thin()
	ContrastMidpointTerrain3Sec stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_terrain3_sec").Thin()
	ColorTerrain3Sec            stingray.ThinHash = stingray.Sum("asset_grade_color_terrain3_sec").Thin()
)
var Terrain3 []stingray.ThinHash = []stingray.ThinHash{
	HueTerrain3, SaturationTerrain3, ValueTerrain3, ContrastTerrain3, ContrastMidpointTerrain3, ColorTerrain3,
}
var Terrain3Sec []stingray.ThinHash = []stingray.ThinHash{
	HueTerrain3Sec, SaturationTerrain3Sec, ValueTerrain3Sec, ContrastTerrain3Sec, ContrastMidpointTerrain3Sec, ColorTerrain3Sec,
}
var (
	HueRoads                 stingray.ThinHash = stingray.Sum("asset_grade_hue_roads").Thin()
	SaturationRoads          stingray.ThinHash = stingray.Sum("asset_grade_saturation_roads").Thin()
	ValueRoads               stingray.ThinHash = stingray.Sum("asset_grade_value_roads").Thin()
	ContrastRoads            stingray.ThinHash = stingray.Sum("asset_grade_contrast_roads").Thin()
	ContrastMidpointRoads    stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_roads").Thin()
	ColorRoads               stingray.ThinHash = stingray.Sum("asset_grade_color_roads").Thin()
	HueRoadsSec              stingray.ThinHash = stingray.Sum("asset_grade_hue_roads_sec").Thin()
	SaturationRoadsSec       stingray.ThinHash = stingray.Sum("asset_grade_saturation_roads_sec").Thin()
	ValueRoadsSec            stingray.ThinHash = stingray.Sum("asset_grade_value_roads_sec").Thin()
	ContrastRoadsSec         stingray.ThinHash = stingray.Sum("asset_grade_contrast_roads_sec").Thin()
	ContrastMidpointRoadsSec stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_roads_sec").Thin()
	ColorRoadsSec            stingray.ThinHash = stingray.Sum("asset_grade_color_roads_sec").Thin()
)
var Roads []stingray.ThinHash = []stingray.ThinHash{
	HueRoads, SaturationRoads, ValueRoads, ContrastRoads, ContrastMidpointRoads, ColorRoads,
}
var RoadsSec []stingray.ThinHash = []stingray.ThinHash{
	HueRoadsSec, SaturationRoadsSec, ValueRoadsSec, ContrastRoadsSec, ContrastMidpointRoadsSec, ColorRoadsSec,
}
var (
	HueRocks                 stingray.ThinHash = stingray.Sum("asset_grade_hue_rocks").Thin()
	SaturationRocks          stingray.ThinHash = stingray.Sum("asset_grade_saturation_rocks").Thin()
	ValueRocks               stingray.ThinHash = stingray.Sum("asset_grade_value_rocks").Thin()
	ContrastRocks            stingray.ThinHash = stingray.Sum("asset_grade_contrast_rocks").Thin()
	ContrastMidpointRocks    stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_rocks").Thin()
	ColorRocks               stingray.ThinHash = stingray.Sum("asset_grade_color_rocks").Thin()
	HueRocksSec              stingray.ThinHash = stingray.Sum("asset_grade_hue_rocks_sec").Thin()
	SaturationRocksSec       stingray.ThinHash = stingray.Sum("asset_grade_saturation_rocks_sec").Thin()
	ValueRocksSec            stingray.ThinHash = stingray.Sum("asset_grade_value_rocks_sec").Thin()
	ContrastRocksSec         stingray.ThinHash = stingray.Sum("asset_grade_contrast_rocks_sec").Thin()
	ContrastMidpointRocksSec stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_rocks_sec").Thin()
	ColorRocksSec            stingray.ThinHash = stingray.Sum("asset_grade_color_rocks_sec").Thin()
)
var Rocks []stingray.ThinHash = []stingray.ThinHash{
	HueRocks, SaturationRocks, ValueRocks, ContrastRocks, ContrastMidpointRocks, ColorRocks,
}
var RocksSec []stingray.ThinHash = []stingray.ThinHash{
	HueRocksSec, SaturationRocksSec, ValueRocksSec, ContrastRocksSec, ContrastMidpointRocksSec, ColorRocksSec,
}
var (
	HueLeaves               stingray.ThinHash = stingray.Sum("asset_grade_hue_leaves").Thin()
	SaturationLeaves        stingray.ThinHash = stingray.Sum("asset_grade_saturation_leaves").Thin()
	ValueLeaves             stingray.ThinHash = stingray.Sum("asset_grade_value_leaves").Thin()
	ContrastLeaves          stingray.ThinHash = stingray.Sum("asset_grade_contrast_leaves").Thin()
	ContrastMidpointLeaves  stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_leaves").Thin()
	ColorLeaves             stingray.ThinHash = stingray.Sum("asset_grade_color_leaves").Thin()
	HueLeaves2              stingray.ThinHash = stingray.Sum("asset_grade_hue_leaves2").Thin()
	SaturationLeaves2       stingray.ThinHash = stingray.Sum("asset_grade_saturation_leaves2").Thin()
	ValueLeaves2            stingray.ThinHash = stingray.Sum("asset_grade_value_leaves2").Thin()
	ContrastLeaves2         stingray.ThinHash = stingray.Sum("asset_grade_contrast_leaves2").Thin()
	ContrastMidpointLeaves2 stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_leaves2").Thin()
	ColorLeaves2            stingray.ThinHash = stingray.Sum("asset_grade_color_leaves2").Thin()
	HueLeaves3              stingray.ThinHash = stingray.Sum("asset_grade_hue_leaves3").Thin()
	SaturationLeaves3       stingray.ThinHash = stingray.Sum("asset_grade_saturation_leaves3").Thin()
	ValueLeaves3            stingray.ThinHash = stingray.Sum("asset_grade_value_leaves3").Thin()
	ContrastLeaves3         stingray.ThinHash = stingray.Sum("asset_grade_contrast_leaves3").Thin()
	ContrastMidpointLeaves3 stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_leaves3").Thin()
	ColorLeaves3            stingray.ThinHash = stingray.Sum("asset_grade_color_leaves3").Thin()
)
var Leaves []stingray.ThinHash = []stingray.ThinHash{
	HueLeaves, SaturationLeaves, ValueLeaves, ContrastLeaves, ContrastMidpointLeaves, ColorLeaves,
}
var Leaves2 []stingray.ThinHash = []stingray.ThinHash{
	HueLeaves2, SaturationLeaves2, ValueLeaves2, ContrastLeaves2, ContrastMidpointLeaves2, ColorLeaves2,
}
var Leaves3 []stingray.ThinHash = []stingray.ThinHash{
	HueLeaves3, SaturationLeaves3, ValueLeaves3, ContrastLeaves3, ContrastMidpointLeaves3, ColorLeaves3,
}
var (
	HueTrunks              stingray.ThinHash = stingray.Sum("asset_grade_hue_trunks").Thin()
	SaturationTrunks       stingray.ThinHash = stingray.Sum("asset_grade_saturation_trunks").Thin()
	ValueTrunks            stingray.ThinHash = stingray.Sum("asset_grade_value_trunks").Thin()
	ContrastTrunks         stingray.ThinHash = stingray.Sum("asset_grade_contrast_trunks").Thin()
	ContrastMidpointTrunks stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_trunks").Thin()
	ColorTrunks            stingray.ThinHash = stingray.Sum("asset_grade_color_trunks").Thin()
)
var Trunks []stingray.ThinHash = []stingray.ThinHash{
	HueTrunks, SaturationTrunks, ValueTrunks, ContrastTrunks, ContrastMidpointTrunks, ColorTrunks,
}
var (
	HueGrass               stingray.ThinHash = stingray.Sum("asset_grade_hue_grass").Thin()
	SaturationGrass        stingray.ThinHash = stingray.Sum("asset_grade_saturation_grass").Thin()
	ValueGrass             stingray.ThinHash = stingray.Sum("asset_grade_value_grass").Thin()
	ContrastGrass          stingray.ThinHash = stingray.Sum("asset_grade_contrast_grass").Thin()
	ContrastMidpointGrass  stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_grass").Thin()
	ColorGrass             stingray.ThinHash = stingray.Sum("asset_grade_color_grass").Thin()
	HueGrass2              stingray.ThinHash = stingray.Sum("asset_grade_hue_grass2").Thin()
	SaturationGrass2       stingray.ThinHash = stingray.Sum("asset_grade_saturation_grass2").Thin()
	ValueGrass2            stingray.ThinHash = stingray.Sum("asset_grade_value_grass2").Thin()
	ContrastGrass2         stingray.ThinHash = stingray.Sum("asset_grade_contrast_grass2").Thin()
	ContrastMidpointGrass2 stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_grass2").Thin()
	ColorGrass2            stingray.ThinHash = stingray.Sum("asset_grade_color_grass2").Thin()
	HueGrass3              stingray.ThinHash = stingray.Sum("asset_grade_hue_grass3").Thin()
	SaturationGrass3       stingray.ThinHash = stingray.Sum("asset_grade_saturation_grass3").Thin()
	ValueGrass3            stingray.ThinHash = stingray.Sum("asset_grade_value_grass3").Thin()
	ContrastGrass3         stingray.ThinHash = stingray.Sum("asset_grade_contrast_grass3").Thin()
	ContrastMidpointGrass3 stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_grass3").Thin()
	ColorGrass3            stingray.ThinHash = stingray.Sum("asset_grade_color_grass3").Thin()
)
var Grass []stingray.ThinHash = []stingray.ThinHash{
	HueGrass, SaturationGrass, ValueGrass, ContrastGrass, ContrastMidpointGrass, ColorGrass,
}
var Grass2 []stingray.ThinHash = []stingray.ThinHash{
	HueGrass2, SaturationGrass2, ValueGrass2, ContrastGrass2, ContrastMidpointGrass2, ColorGrass2,
}
var Grass3 []stingray.ThinHash = []stingray.ThinHash{
	HueGrass3, SaturationGrass3, ValueGrass3, ContrastGrass3, ContrastMidpointGrass3, ColorGrass3,
}
var (
	HueBushes               stingray.ThinHash = stingray.Sum("asset_grade_hue_bushes").Thin()
	SaturationBushes        stingray.ThinHash = stingray.Sum("asset_grade_saturation_bushes").Thin()
	ValueBushes             stingray.ThinHash = stingray.Sum("asset_grade_value_bushes").Thin()
	ContrastBushes          stingray.ThinHash = stingray.Sum("asset_grade_contrast_bushes").Thin()
	ContrastMidpointBushes  stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_bushes").Thin()
	ColorBushes             stingray.ThinHash = stingray.Sum("asset_grade_color_bushes").Thin()
	HueBushes2              stingray.ThinHash = stingray.Sum("asset_grade_hue_bushes2").Thin()
	SaturationBushes2       stingray.ThinHash = stingray.Sum("asset_grade_saturation_bushes2").Thin()
	ValueBushes2            stingray.ThinHash = stingray.Sum("asset_grade_value_bushes2").Thin()
	ContrastBushes2         stingray.ThinHash = stingray.Sum("asset_grade_contrast_bushes2").Thin()
	ContrastMidpointBushes2 stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_bushes2").Thin()
	ColorBushes2            stingray.ThinHash = stingray.Sum("asset_grade_color_bushes2").Thin()
)
var Bushes []stingray.ThinHash = []stingray.ThinHash{
	HueBushes, SaturationBushes, ValueBushes, ContrastBushes, ContrastMidpointBushes, ColorBushes,
}
var Bushes2 []stingray.ThinHash = []stingray.ThinHash{
	HueBushes2, SaturationBushes2, ValueBushes2, ContrastBushes2, ContrastMidpointBushes2, ColorBushes2,
}
var (
	HueMoss                 stingray.ThinHash = stingray.Sum("asset_grade_hue_moss").Thin()
	SaturationMoss          stingray.ThinHash = stingray.Sum("asset_grade_saturation_moss").Thin()
	ValueMoss               stingray.ThinHash = stingray.Sum("asset_grade_value_moss").Thin()
	ContrastMoss            stingray.ThinHash = stingray.Sum("asset_grade_contrast_moss").Thin()
	ContrastMidpointMoss    stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_moss").Thin()
	ColorMoss               stingray.ThinHash = stingray.Sum("asset_grade_color_moss").Thin()
	HueMossSec              stingray.ThinHash = stingray.Sum("asset_grade_hue_moss_sec").Thin()
	SaturationMossSec       stingray.ThinHash = stingray.Sum("asset_grade_saturation_moss_sec").Thin()
	ValueMossSec            stingray.ThinHash = stingray.Sum("asset_grade_value_moss_sec").Thin()
	ContrastMossSec         stingray.ThinHash = stingray.Sum("asset_grade_contrast_moss_sec").Thin()
	ContrastMidpointMossSec stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_moss_sec").Thin()
	ColorMossSec            stingray.ThinHash = stingray.Sum("asset_grade_color_moss_sec").Thin()
)
var Moss []stingray.ThinHash = []stingray.ThinHash{
	HueMoss, SaturationMoss, ValueMoss, ContrastMoss, ContrastMidpointMoss, ColorMoss,
}
var MossSec []stingray.ThinHash = []stingray.ThinHash{
	HueMossSec, SaturationMossSec, ValueMossSec, ContrastMossSec, ContrastMidpointMossSec, ColorMossSec,
}
var (
	HuePebbles               stingray.ThinHash = stingray.Sum("asset_grade_hue_pebbles").Thin()
	SaturationPebbles        stingray.ThinHash = stingray.Sum("asset_grade_saturation_pebbles").Thin()
	ValuePebbles             stingray.ThinHash = stingray.Sum("asset_grade_value_pebbles").Thin()
	ContrastPebbles          stingray.ThinHash = stingray.Sum("asset_grade_contrast_pebbles").Thin()
	ContrastMidpointPebbles  stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_pebbles").Thin()
	ColorPebbles             stingray.ThinHash = stingray.Sum("asset_grade_color_pebbles").Thin()
	HuePebbles2              stingray.ThinHash = stingray.Sum("asset_grade_hue_pebbles2").Thin()
	SaturationPebbles2       stingray.ThinHash = stingray.Sum("asset_grade_saturation_pebbles2").Thin()
	ValuePebbles2            stingray.ThinHash = stingray.Sum("asset_grade_value_pebbles2").Thin()
	ContrastPebbles2         stingray.ThinHash = stingray.Sum("asset_grade_contrast_pebbles2").Thin()
	ContrastMidpointPebbles2 stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_pebbles2").Thin()
	ColorPebbles2            stingray.ThinHash = stingray.Sum("asset_grade_color_pebbles2").Thin()
)
var Pebbles []stingray.ThinHash = []stingray.ThinHash{
	HuePebbles, SaturationPebbles, ValuePebbles, ContrastPebbles, ContrastMidpointPebbles, ColorPebbles,
}
var Pebbles2 []stingray.ThinHash = []stingray.ThinHash{
	HuePebbles2, SaturationPebbles2, ValuePebbles2, ContrastPebbles2, ContrastMidpointPebbles2, ColorPebbles2,
}
var (
	HueVista              stingray.ThinHash = stingray.Sum("asset_grade_hue_vista").Thin()
	SaturationVista       stingray.ThinHash = stingray.Sum("asset_grade_saturation_vista").Thin()
	ValueVista            stingray.ThinHash = stingray.Sum("asset_grade_value_vista").Thin()
	ContrastVista         stingray.ThinHash = stingray.Sum("asset_grade_contrast_vista").Thin()
	ContrastMidpointVista stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_vista").Thin()
	ColorVista            stingray.ThinHash = stingray.Sum("asset_grade_color_vista").Thin()
)
var Vista []stingray.ThinHash = []stingray.ThinHash{
	HueVista, SaturationVista, ValueVista, ContrastVista, ContrastMidpointVista, ColorVista,
}
var (
	HueGenericRock              stingray.ThinHash = stingray.Sum("asset_grade_hue_generic_rock").Thin()
	SaturationGenericRock       stingray.ThinHash = stingray.Sum("asset_grade_saturation_generic_rock").Thin()
	ValueGenericRock            stingray.ThinHash = stingray.Sum("asset_grade_value_generic_rock").Thin()
	ContrastGenericRock         stingray.ThinHash = stingray.Sum("asset_grade_contrast_generic_rock").Thin()
	ContrastMidpointGenericRock stingray.ThinHash = stingray.Sum("asset_grade_contrast_midpoint_generic_rock").Thin()
	ColorGenericRock            stingray.ThinHash = stingray.Sum("asset_grade_color_generic_rock").Thin()
)
var GenericRock []stingray.ThinHash = []stingray.ThinHash{
	HueGenericRock, SaturationGenericRock, ValueGenericRock, ContrastGenericRock, ContrastMidpointGenericRock, ColorGenericRock,
}

var Groups [][]stingray.ThinHash = [][]stingray.ThinHash{
	{}, Terrain1, Terrain1Sec, Terrain2, Terrain2Sec, Terrain3, Terrain3Sec, Roads, RoadsSec, Rocks, RocksSec, Leaves, Leaves2, Leaves3, Trunks, Grass, Grass2, Grass3, Bushes, Bushes2, Moss, MossSec, Pebbles, Pebbles2, Vista, GenericRock,
}

type GradingType uint32

const (
	None GradingType = iota
	Hue
	Saturation
	Value
	Contrast
	ContrastMidpoint
	Color
)

var Order []GradingType = []GradingType{
	Color,
	Hue,
	Saturation,
	Value,
	Contrast,
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=GradingType

func (g GradingType) MarshalText() ([]byte, error) {
	return []byte(g.String()), nil
}

func (g *GradingType) UnmarshalText(data []byte) error {
	switch string(data) {
	case None.String():
		*g = None
	case Hue.String():
		*g = Hue
	case Saturation.String():
		*g = Saturation
	case Value.String():
		*g = Value
	case Contrast.String():
		*g = Contrast
	case ContrastMidpoint.String():
		*g = ContrastMidpoint
	case Color.String():
		*g = Color
	default:
		return errors.ErrUnsupported
	}
	return nil
}

const (
	sqrt3over3 float32 = 0.577350269
	// from luminance formula from wikipedia
	lumR float32 = 0.3086
	lumG float32 = 0.6094
	lumB float32 = 0.0820
)

func getF32(data any) (float32, error) {
	s, ok := data.(float32)
	if !ok {
		s64, ok := data.(float64)
		if !ok {
			return 0, fmt.Errorf("could not cast to float")
		}
		s = float32(s64)
	}
	return s, nil
}

func getVec3(data any) (mgl32.Vec3, error) {
	s, ok := data.(mgl32.Vec3)
	if !ok {
		s64, ok := data.([3]float64)
		if !ok {
			return mgl32.Vec3{}, fmt.Errorf("could not cast to float")
		}
		s = mgl32.Vec3{float32(s64[0]), float32(s64[1]), float32(s64[2])}
	}
	return s, nil
}

// These matrices are from https://lisyarus.github.io/blog/posts/transforming-colors-with-matrices.html
// Also helpful: https://docs.rainmeter.net/tips/colormatrix-guide/ (just chop off the 5th row and col)
func (g GradingType) GetMatrix(data any) mgl32.Mat4 {
	switch g {
	case Hue:
		angle, err := getF32(data)
		if err != nil {
			return mgl32.Ident4()
		}
		return mgl32.HomogRotate3D(angle*2*math.Pi, mgl32.Vec3{sqrt3over3, sqrt3over3, sqrt3over3}).Transpose()
	case Saturation:
		s, err := getF32(data)
		if err != nil {
			return mgl32.Ident4()
		}
		return mgl32.Mat4FromCols(
			mgl32.Vec4{s + (1-s)*lumR, (1 - s) * lumG, (1 - s) * lumB, 0.0},
			mgl32.Vec4{(1 - s) * lumR, s + (1-s)*lumG, (1 - s) * lumB, 0.0},
			mgl32.Vec4{(1 - s) * lumR, (1 - s) * lumG, s + (1-s)*lumB, 0.0},
			mgl32.Vec4{0.0, 0.0, 0.0, 1.0},
		)
	case Value:
		val, err := getF32(data)
		if err != nil {
			return mgl32.Ident4()
		}
		return mgl32.Diag4(mgl32.Vec4{val, val, val, 1})
	case Contrast:
		val, err := getF32(data)
		if err != nil {
			return mgl32.Ident4()
		}
		return mgl32.Mat4FromCols(
			mgl32.Vec4{val, 0.0, 0.0, (1 - val) / 2},
			mgl32.Vec4{0.0, val, 0.0, (1 - val) / 2},
			mgl32.Vec4{0.0, 0.0, val, (1 - val) / 2},
			mgl32.Vec4{0.0, 0.0, 0.0, 1.0},
		)
	case ContrastMidpoint:
		val, err := getF32(data)
		if err != nil {
			return mgl32.Ident4()
		}
		return mgl32.Mat4FromCols(
			mgl32.Vec4{0.0, 0.0, 0.0, val},
			mgl32.Vec4{0.0, 0.0, 0.0, val},
			mgl32.Vec4{0.0, 0.0, 0.0, val},
			mgl32.Vec4{0.0, 0.0, 0.0, 1.0},
		)
	case Color:
		color, err := getVec3(data)
		if err != nil {
			return mgl32.Ident4()
		}
		return mgl32.Mat4FromCols(
			mgl32.Vec4{0.0, 0.0, 0.0, color[0]},
			mgl32.Vec4{0.0, 0.0, 0.0, color[1]},
			mgl32.Vec4{0.0, 0.0, 0.0, color[2]},
			mgl32.Vec4{0.0, 0.0, 0.0, 1.0},
		)
	}
	return mgl32.Ident4()
}

func GetGradingType(t stingray.ThinHash) (GradingType, error) {
	switch t {
	case HueBushes, HueBushes2, HueGenericRock, HueGrass, HueGrass2, HueGrass3, HueLeaves, HueLeaves2, HueLeaves3, HueMoss, HueMossSec, HuePebbles, HuePebbles2, HueRoads, HueRoadsSec, HueRocks, HueRocksSec, HueTerrain1, HueTerrain1Sec, HueTerrain2, HueTerrain2Sec, HueTerrain3, HueTerrain3Sec, HueTrunks, HueVista:
		return Hue, nil
	case SaturationBushes, SaturationBushes2, SaturationGenericRock, SaturationGrass, SaturationGrass2, SaturationGrass3, SaturationLeaves, SaturationLeaves2, SaturationLeaves3, SaturationMoss, SaturationMossSec, SaturationPebbles, SaturationPebbles2, SaturationRoads, SaturationRoadsSec, SaturationRocks, SaturationRocksSec, SaturationTerrain1, SaturationTerrain1Sec, SaturationTerrain2, SaturationTerrain2Sec, SaturationTerrain3, SaturationTerrain3Sec, SaturationTrunks, SaturationVista:
		return Saturation, nil
	case ValueBushes, ValueBushes2, ValueGenericRock, ValueGrass, ValueGrass2, ValueGrass3, ValueLeaves, ValueLeaves2, ValueLeaves3, ValueMoss, ValueMossSec, ValuePebbles, ValuePebbles2, ValueRoads, ValueRoadsSec, ValueRocks, ValueRocksSec, ValueTerrain1, ValueTerrain1Sec, ValueTerrain2, ValueTerrain2Sec, ValueTerrain3, ValueTerrain3Sec, ValueTrunks, ValueVista:
		return Value, nil
	case ContrastBushes, ContrastBushes2, ContrastGenericRock, ContrastGrass, ContrastGrass2, ContrastGrass3, ContrastLeaves, ContrastLeaves2, ContrastLeaves3, ContrastMoss, ContrastMossSec, ContrastPebbles, ContrastPebbles2, ContrastRoads, ContrastRoadsSec, ContrastRocks, ContrastRocksSec, ContrastTerrain1, ContrastTerrain1Sec, ContrastTerrain2, ContrastTerrain2Sec, ContrastTerrain3, ContrastTerrain3Sec, ContrastTrunks, ContrastVista:
		return Contrast, nil
	case ContrastMidpointBushes, ContrastMidpointBushes2, ContrastMidpointGenericRock, ContrastMidpointGrass, ContrastMidpointGrass2, ContrastMidpointGrass3, ContrastMidpointLeaves, ContrastMidpointLeaves2, ContrastMidpointLeaves3, ContrastMidpointMoss, ContrastMidpointMossSec, ContrastMidpointPebbles, ContrastMidpointPebbles2, ContrastMidpointRoads, ContrastMidpointRoadsSec, ContrastMidpointRocks, ContrastMidpointRocksSec, ContrastMidpointTerrain1, ContrastMidpointTerrain1Sec, ContrastMidpointTerrain2, ContrastMidpointTerrain2Sec, ContrastMidpointTerrain3, ContrastMidpointTerrain3Sec, ContrastMidpointTrunks, ContrastMidpointVista:
		return ContrastMidpoint, nil
	case ColorBushes, ColorBushes2, ColorGenericRock, ColorGrass, ColorGrass2, ColorGrass3, ColorLeaves, ColorLeaves2, ColorLeaves3, ColorMoss, ColorMossSec, ColorPebbles, ColorPebbles2, ColorRoads, ColorRoadsSec, ColorRocks, ColorRocksSec, ColorTerrain1, ColorTerrain1Sec, ColorTerrain2, ColorTerrain2Sec, ColorTerrain3, ColorTerrain3Sec, ColorTrunks, ColorVista:
		return Color, nil
	default:
		return None, fmt.Errorf("not a grading type: %v", t.String())
	}
}

func GetGradingGroupId(t stingray.ThinHash) (uint32, error) {
	for i, group := range Groups {
		if slices.Contains(group, t) {
			return uint32(i), nil
		}
	}
	return 0, fmt.Errorf("not a grading type")
}
