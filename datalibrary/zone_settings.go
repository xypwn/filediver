package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

type WindNoiseSettings struct {
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

type QuakeEffectSettings struct {
	Interval               uint8
	_                      [7]uint8
	ShakeEffectId          stingray.Hash
	AudioEffectId          stingray.ThinHash
	ShakeEffectInnerRadius uint8
	ShakeEffectOuterRadius uint8
	_                      [2]uint8
}

type SimpleQuakeEffectSettings struct {
	Interval               uint8  `json:"interval"`
	ShakeEffectId          string `json:"shake_effect_id"`
	AudioEffectId          string `json:"audio_effect_id"`
	ShakeEffectInnerRadius uint8  `json:"shake_effect_inner_radius"`
	ShakeEffectOuterRadius uint8  `json:"shake_effect_outer_radius"`
}

func (s QuakeEffectSettings) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleQuakeEffectSettings {
	return SimpleQuakeEffectSettings{
		Interval:               s.Interval,
		ShakeEffectId:          lookupHash(s.ShakeEffectId),
		AudioEffectId:          lookupThinHash(s.AudioEffectId),
		ShakeEffectInnerRadius: s.ShakeEffectInnerRadius,
		ShakeEffectOuterRadius: s.ShakeEffectOuterRadius,
	}
}

type rawBoundaryWallGroup struct {
	BoundaryWallPieces1 DLArray
	BoundaryWallPieces2 DLArray
	BoundaryWallPieces3 DLArray
	BoundaryWallPieces4 DLArray
	BoundaryWallPieces5 DLArray
	BoundaryWallPieces6 DLArray
	EntryPoints         DLArray
}

type BoundaryWallPiece struct {
	PiecePath stingray.Hash
	UnkFloat  float32
	_         [4]uint8
}

type SimpleBoundaryWallPiece struct {
	PiecePath string  `json:"piece_path"`
	UnkFloat  float32 `json:"unk_float"`
}

func (s BoundaryWallPiece) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleBoundaryWallPiece {
	return SimpleBoundaryWallPiece{
		PiecePath: lookupHash(s.PiecePath),
		UnkFloat:  s.UnkFloat,
	}
}

type BoundaryWallEntryPoint struct {
	PiecePath stingray.Hash
	UnkFloat1 float32
	UnkFloat2 float32
}

type SimpleBoundaryWallEntryPoint struct {
	PiecePath string  `json:"piece_path"`
	UnkFloat1 float32 `json:"unk_float1"`
	UnkFloat2 float32 `json:"unk_float2"`
}

func (s BoundaryWallEntryPoint) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleBoundaryWallEntryPoint {
	return SimpleBoundaryWallEntryPoint{
		PiecePath: lookupHash(s.PiecePath),
		UnkFloat1: s.UnkFloat1,
		UnkFloat2: s.UnkFloat2,
	}
}

type BoundaryWallGroup struct {
	BoundaryWallPieces1 []BoundaryWallPiece
	BoundaryWallPieces2 []BoundaryWallPiece
	BoundaryWallPieces3 []BoundaryWallPiece
	BoundaryWallPieces4 []BoundaryWallPiece
	BoundaryWallPieces5 []BoundaryWallPiece
	BoundaryWallPieces6 []BoundaryWallPiece
	EntryPoints         []BoundaryWallEntryPoint
}

type SimpleBoundaryWallGroup struct {
	BoundaryWallPieces1 []SimpleBoundaryWallPiece      `json:"boundary_wall_pieces1"`
	BoundaryWallPieces2 []SimpleBoundaryWallPiece      `json:"boundary_wall_pieces2"`
	BoundaryWallPieces3 []SimpleBoundaryWallPiece      `json:"boundary_wall_pieces3"`
	BoundaryWallPieces4 []SimpleBoundaryWallPiece      `json:"boundary_wall_pieces4"`
	BoundaryWallPieces5 []SimpleBoundaryWallPiece      `json:"boundary_wall_pieces5"`
	BoundaryWallPieces6 []SimpleBoundaryWallPiece      `json:"boundary_wall_pieces6"`
	EntryPoints         []SimpleBoundaryWallEntryPoint `json:"entry_points"`
}

func (s BoundaryWallGroup) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleBoundaryWallGroup {
	bwp1 := make([]SimpleBoundaryWallPiece, 0)
	for _, piece := range s.BoundaryWallPieces1 {
		bwp1 = append(bwp1, piece.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	bwp2 := make([]SimpleBoundaryWallPiece, 0)
	for _, piece := range s.BoundaryWallPieces2 {
		bwp2 = append(bwp2, piece.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	bwp3 := make([]SimpleBoundaryWallPiece, 0)
	for _, piece := range s.BoundaryWallPieces3 {
		bwp3 = append(bwp3, piece.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	bwp4 := make([]SimpleBoundaryWallPiece, 0)
	for _, piece := range s.BoundaryWallPieces4 {
		bwp4 = append(bwp4, piece.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	bwp5 := make([]SimpleBoundaryWallPiece, 0)
	for _, piece := range s.BoundaryWallPieces5 {
		bwp5 = append(bwp5, piece.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	bwp6 := make([]SimpleBoundaryWallPiece, 0)
	for _, piece := range s.BoundaryWallPieces6 {
		bwp6 = append(bwp6, piece.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	entrypoints := make([]SimpleBoundaryWallEntryPoint, 0)
	for _, piece := range s.EntryPoints {
		entrypoints = append(entrypoints, piece.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleBoundaryWallGroup{
		BoundaryWallPieces1: bwp1,
		BoundaryWallPieces2: bwp2,
		BoundaryWallPieces3: bwp3,
		BoundaryWallPieces4: bwp4,
		BoundaryWallPieces5: bwp5,
		BoundaryWallPieces6: bwp6,
		EntryPoints:         entrypoints,
	}
}

func (g rawBoundaryWallGroup) Deserialize(r io.ReadSeeker, base int64) (*BoundaryWallGroup, error) {
	bwp1, err := ResolveDLArray[BoundaryWallPiece](g.BoundaryWallPieces1, r, base)
	if err != nil {
		return nil, err
	}
	bwp2, err := ResolveDLArray[BoundaryWallPiece](g.BoundaryWallPieces2, r, base)
	if err != nil {
		return nil, err
	}
	bwp3, err := ResolveDLArray[BoundaryWallPiece](g.BoundaryWallPieces3, r, base)
	if err != nil {
		return nil, err
	}
	bwp4, err := ResolveDLArray[BoundaryWallPiece](g.BoundaryWallPieces4, r, base)
	if err != nil {
		return nil, err
	}
	bwp5, err := ResolveDLArray[BoundaryWallPiece](g.BoundaryWallPieces5, r, base)
	if err != nil {
		return nil, err
	}
	bwp6, err := ResolveDLArray[BoundaryWallPiece](g.BoundaryWallPieces6, r, base)
	if err != nil {
		return nil, err
	}
	entrypoints, err := ResolveDLArray[BoundaryWallEntryPoint](g.EntryPoints, r, base)
	if err != nil {
		return nil, err
	}
	return &BoundaryWallGroup{
		BoundaryWallPieces1: bwp1,
		BoundaryWallPieces2: bwp2,
		BoundaryWallPieces3: bwp3,
		BoundaryWallPieces4: bwp4,
		BoundaryWallPieces5: bwp5,
		BoundaryWallPieces6: bwp6,
		EntryPoints:         entrypoints,
	}, nil
}

type BoundaryWallUnknownStruct1 struct {
	UnkFloat1 float32 `json:"unk_float1"`
	UnkFloat2 float32 `json:"unk_float2"`
	UnkFloat3 float32 `json:"unk_float3"`
	UnkFloat4 float32 `json:"unk_float4"`
	UnkFloat5 float32 `json:"unk_float5"`
}

type BoundaryWallUnknownStruct2 struct {
	UnkFloat1 float32 `json:"unk_float1"`
	UnkFloat2 float32 `json:"unk_float2"`
	UnkFloat3 float32 `json:"unk_float3"`
	UnkFloat4 float32 `json:"unk_float4"`
	UnkInt1   uint32  `json:"unk_int1"`
	UnkFloat5 float32 `json:"unk_float5"`
}

type BoundaryWallUnknownStruct3 struct {
	UnkFloat1 float32 `json:"unk_float1"`
	UnkFloat2 float32 `json:"unk_float2"`
}

type rawBoundaryWallUnknownStruct4 struct {
	UnkStruct1     BoundaryWallUnknownStruct1
	UnkStruct2     BoundaryWallUnknownStruct2
	_              [4]uint8
	UnkStructArray DLArray // BoundaryWallUnknownStruct3
}

type rawBoundaryWallSettings struct {
	UnkFloat1  float32
	UnkFloat2  float32
	Group      rawBoundaryWallGroup
	UnkStruct1 rawBoundaryWallUnknownStruct4
}

type BoundaryWallUnknownStruct4 struct {
	UnkStruct1     BoundaryWallUnknownStruct1   `json:"boundary_wall_info1"`
	UnkStruct2     BoundaryWallUnknownStruct2   `json:"boundary_wall_info2"`
	UnkStructArray []BoundaryWallUnknownStruct3 `json:"boundary_wall_info_array"`
}

func (s rawBoundaryWallUnknownStruct4) Deserialize(r io.ReadSeeker, base int64) (*BoundaryWallUnknownStruct4, error) {
	unkStructArray, err := ResolveDLArray[BoundaryWallUnknownStruct3](s.UnkStructArray, r, base)
	if err != nil {
		return nil, err
	}
	return &BoundaryWallUnknownStruct4{
		UnkStruct1:     s.UnkStruct1,
		UnkStruct2:     s.UnkStruct2,
		UnkStructArray: unkStructArray,
	}, nil
}

type BoundaryWallSettings struct {
	UnkFloat1 float32
	UnkFloat2 float32
	Group     BoundaryWallGroup
	UnkStruct BoundaryWallUnknownStruct4
}

type SimpleBoundaryWallSettings struct {
	UnkFloat1 float32                    `json:"unk_float1"`
	UnkFloat2 float32                    `json:"unk_float2"`
	Group     SimpleBoundaryWallGroup    `json:"group"`
	UnkStruct BoundaryWallUnknownStruct4 `json:"unk_struct"`
}

func (s BoundaryWallSettings) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleBoundaryWallSettings {
	return SimpleBoundaryWallSettings{
		UnkFloat1: s.UnkFloat1,
		UnkFloat2: s.UnkFloat2,
		Group:     s.Group.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		UnkStruct: s.UnkStruct,
	}
}

func (s rawBoundaryWallSettings) Deserialize(r io.ReadSeeker, base int64) (*BoundaryWallSettings, error) {
	unkStruct, err := s.UnkStruct1.Deserialize(r, base)
	if err != nil {
		return nil, err
	}
	group, err := s.Group.Deserialize(r, base)
	if err != nil {
		return nil, err
	}
	return &BoundaryWallSettings{
		UnkFloat1: s.UnkFloat1,
		UnkFloat2: s.UnkFloat2,
		Group:     *group,
		UnkStruct: *unkStruct,
	}, nil
}

type rawMinimapScatter struct {
	UnitResource  stingray.Hash
	Weight        float32
	SizeVariation float32
	Color         mgl32.Vec3
	Position      mgl32.Vec3
	Scale         mgl32.Vec3
	Rotation      float32
	UnkFloat1     float32
	UnkBitfield   uint8
	_             [3]uint8
	VectorArray   DLArray
	UnkFloat2     float32
	_             [4]uint8
}

type rawMinimapScatterSet struct {
	UnkFloat              float32
	Density               float32
	WaterLowerSpawnCutoff float32
	WaterUpperSpawnCutoff float32
	UnkBitfield           uint8
	_                     [7]uint8
	MinimapScatterArray   DLArray
}

type MinimapScatter struct {
	UnitResource  stingray.Hash
	Weight        float32
	SizeVariation float32
	Color         mgl32.Vec3
	Position      mgl32.Vec3
	Scale         mgl32.Vec3
	Rotation      float32
	UnkFloat1     float32
	UnkBitfield   uint8
	VectorArray   []mgl32.Vec2
	UnkFloat2     float32
}

type SimpleMinimapScatter struct {
	UnitResource  string       `json:"unit_resource"`
	Weight        float32      `json:"weight"`
	SizeVariation float32      `json:"size_variation"`
	Color         mgl32.Vec3   `json:"color"`
	Position      mgl32.Vec3   `json:"position"`
	Scale         mgl32.Vec3   `json:"scale"`
	Rotation      float32      `json:"rotation"`
	UnkFloat1     float32      `json:"unk_float1"`
	UnkBitfield   uint8        `json:"unk_bitfield"`
	VectorArray   []mgl32.Vec2 `json:"vector_array"`
	UnkFloat2     float32      `json:"unk_float2"`
}

func (s MinimapScatter) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleMinimapScatter {
	return SimpleMinimapScatter{
		UnitResource:  lookupHash(s.UnitResource),
		Weight:        s.Weight,
		SizeVariation: s.SizeVariation,
		Color:         s.Color,
		Position:      s.Position,
		Scale:         s.Scale,
		Rotation:      s.Rotation,
		UnkFloat1:     s.UnkFloat1,
		UnkBitfield:   s.UnkBitfield,
		VectorArray:   s.VectorArray,
		UnkFloat2:     s.UnkFloat2,
	}
}

func (s rawMinimapScatter) Deserialize(r io.ReadSeeker, base int64) (*MinimapScatter, error) {
	vectorArray, err := ResolveDLArray[mgl32.Vec2](s.VectorArray, r, base)
	if err != nil {
		return nil, err
	}
	return &MinimapScatter{
		UnitResource:  s.UnitResource,
		Weight:        s.Weight,
		SizeVariation: s.SizeVariation,
		Color:         s.Color,
		Position:      s.Position,
		Scale:         s.Scale,
		Rotation:      s.Rotation,
		UnkFloat1:     s.UnkFloat1,
		UnkBitfield:   s.UnkBitfield,
		VectorArray:   vectorArray,
		UnkFloat2:     s.UnkFloat2,
	}, nil
}

type MinimapScatterSet struct {
	UnkFloat              float32
	Density               float32
	WaterLowerSpawnCutoff float32
	WaterUpperSpawnCutoff float32
	UnkBitfield           uint8
	MinimapScatterArray   []MinimapScatter
}

type SimpleMinimapScatterSet struct {
	UnkFloat              float32                `json:"unk_float"`
	Density               float32                `json:"density"`
	WaterLowerSpawnCutoff float32                `json:"water_lower_spawn_cutoff"`
	WaterUpperSpawnCutoff float32                `json:"water_upper_spawn_cutoff"`
	UnkBitfield           uint8                  `json:"unk_bitfield"`
	MinimapScatterArray   []SimpleMinimapScatter `json:"minimap_scatter_array"`
}

func (s MinimapScatterSet) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleMinimapScatterSet {
	scatterArray := make([]SimpleMinimapScatter, 0)
	for _, scatter := range s.MinimapScatterArray {
		scatterArray = append(scatterArray, scatter.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleMinimapScatterSet{
		UnkFloat:              s.UnkFloat,
		Density:               s.Density,
		WaterLowerSpawnCutoff: s.WaterLowerSpawnCutoff,
		WaterUpperSpawnCutoff: s.WaterUpperSpawnCutoff,
		UnkBitfield:           s.UnkBitfield,
		MinimapScatterArray:   scatterArray,
	}
}

func (s rawMinimapScatterSet) Deserialize(r io.ReadSeeker, base int64) (*MinimapScatterSet, error) {
	rawScatterArray, err := ResolveDLArray[rawMinimapScatter](s.MinimapScatterArray, r, base)
	if err != nil {
		return nil, err
	}
	scatterArray := make([]MinimapScatter, len(rawScatterArray))
	for _, rawScatter := range rawScatterArray {
		scatter, err := rawScatter.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		scatterArray = append(scatterArray, *scatter)
	}
	return &MinimapScatterSet{
		UnkFloat:              s.UnkFloat,
		Density:               s.Density,
		WaterLowerSpawnCutoff: s.WaterLowerSpawnCutoff,
		WaterUpperSpawnCutoff: s.WaterUpperSpawnCutoff,
		UnkBitfield:           s.UnkBitfield,
		MinimapScatterArray:   scatterArray,
	}, nil
}

type rawZoneSettings struct {
	StampGroups                    DLArray
	MaterialLookupUnit             stingray.Hash
	ColorVariationGenerators       DLArray
	ShaderProperties1              DLArray
	FogVolumeGenerators            DLArray
	MinimapVisualizationGenerators DLArray
	MaterialGenerators             DLArray
	ShaderProperties2              DLArray
	HeightGenerators               DLArray
	DefaultMaterial                int32
	MinimapColor                   mgl32.Vec3
	MinimapTerrainType             enum.MinimapTerrainType
	_                              [7]uint8
	ScatterSetting                 DLString
	CameraEnvironmentEffects       DLArray
	WindNoiseIds                   WindNoiseSettings
	WaterMaterial                  stingray.Hash
	WaterLevelOffset               float32
	WaterHeight                    float32
	WaterZoneOpacityCutoff         float32
	MaxWaterDepth                  float32
	HeightModificationCurve        DLArray
	ReverbZone                     DLString
	AmbienceSoundId                stingray.ThinHash
	_                              [4]uint8
	QuakeEffect                    QuakeEffectSettings
	ZoneMusicType                  enum.ZoneMusicType
	BanterEventType                enum.BanterEventType
	_                              [2]uint8
	PlayerKillHeight               float32
	MinimapScatter                 rawMinimapScatterSet
	LocationHeightOffset           float32
	_                              [4]uint8
	UnkString                      DLString
	UnkThinHash1                   stingray.ThinHash
	UnkThinHash2                   stingray.ThinHash
	UnkBitfield1                   uint8
	_                              [3]uint8
	PatrolSetting                  enum.PatrolSetting
	PatrolFloat1                   float32
	PatrolFloat2                   float32
	PatrolFloat3                   float32
	PatrolFloat4                   float32
	PatrolFloat5                   float32
	PatrolFloat6                   float32
	PatrolFloat7                   float32
	PatrolFloat8                   float32
	PatrolFloat9                   float32
	PatrolFloat10                  float32
	PatrolFloat11                  float32
	UnkBool1                       uint8
	UnkBitfield2                   uint8
	_                              [2]uint8
	CityPlacementHeuristic         enum.CityPlacementHeuristic
	UnknownEnum                    uint32
	UnkInt1                        int32
	UnkInt2                        uint32
	Vectors1                       [3]mgl32.Vec2
	Vectors2                       [3]mgl32.Vec2
	Vectors3                       [3]mgl32.Vec2
	UnkFloat1                      float32
	UnkFloat2                      float32
	UnkFloat3                      float32
	UnkFloat4                      float32
	LocationFlags                  [4]enum.LocationFlag
	_                              [4]uint8
	ColoniesCurbPath               stingray.Hash
	CurvedCurbPath                 stingray.Hash
	UnkFLoat5                      float32
	UnkFLoat6                      float32
	UnkInt3                        uint32
	_                              [4]uint8
	DetailUnit1                    stingray.Hash
	DetailUnit2                    stingray.Hash
	DetailUnit3                    stingray.Hash
	TrafficlightPrefab             stingray.Hash
	UnkFloat7                      float32
	UnkFloat8                      float32
	UnkFloat9                      float32
	UnkFloat10                     float32
	UnkFloat11                     float32
	_                              [4]uint8
	CityStreetlights               stingray.Hash
	UnkFloat12                     float32
	UnkFloat13                     float32
	UnkFloat14                     float32
	_                              [4]uint8
	DetailUnit4                    stingray.Hash
	UnkFloat15                     float32
	UnkFloat16                     float32
	UnkFloat17                     float32
	UnkFloat18                     float32
	UnkFloat19                     float32
	UnkFloat20                     float32
	UnkInt4                        uint64
	UnkString2                     DLString
	UnkFloat21                     float32
	UnkFloat22                     float32
	BoundaryWallSettings           rawBoundaryWallSettings
	RoadEmbankmentUnits            DLArray
	UnkInt5                        uint32
	UnkBitfield3                   uint8
	_                              [3]uint8
}

type StampGroupFlags struct {
	Bits uint16
}

type DifficultyBitfield struct {
	Bits uint16
}

type StampBitfield struct {
	Bits uint16
}

type rawStampGroup struct {
	Stamps                      DLArray
	GridCellSize                float32
	UnkFloat1                   float32
	MaxRandomOffsetInCell       float32
	FillRate                    float32
	MaxPossibleCellsPerMap      int32
	Flags                       StampGroupFlags
	_                           [2]uint8
	StampMaxOverlap             float32
	StampMinOverlap             float32
	StampOverlapGroup           enum.StampOverlapGroups
	_                           [7]uint8
	StampOverlapSettings        DLArray
	RouteStampOffset            float32
	MaxDistanceFromRoute        float32
	MinDistanceFromRoute        float32
	MaxDistanceFromPlayableArea float32
	DifficultyBitfield
	_         [2]uint8
	UnkFloat2 float32
}

type StampWeights struct {
	SlopeWeight          mgl32.Vec4 `json:"slope_weight,omitzero"`
	HeightWeight         mgl32.Vec4 `json:"height_weight,omitzero"`
	OutsideLevelInterval mgl32.Vec4 `json:"outside_level_interval,omitzero"`
	RegionWeight         mgl32.Vec4 `json:"region_weight,omitzero"`
	Flags                uint8      `json:"flags,omitzero"`
	_                    [3]uint8
}

type rawStampInfo struct {
	Stamp               DLPtr
	Path                stingray.Hash
	Weights             StampWeights
	Weight              float32
	FlatteningIntensity float32
	UnkFloat1           float32
	StampBitfield
	_ [6]uint8
}

type rawStamp struct {
	CollisionCircles    DLArray
	CollisionRectangles DLArray
	Entrypoints         DLArray
	Splines             DLArray
	Extensions          DLArray
	Proxies             DLArray
	RotationType        enum.RotationType
}

type StampCircle struct {
	Position            mgl32.Vec2 `json:"position"`
	Radius              float32    `json:"radius"`
	FlatteningIntensity float32    `json:"flattening_intensity"`
	FlattenTerrain      bool       `json:"flatten_terrain"`
	_                   [3]uint8
}

type StampRectangle struct {
	Position            mgl32.Vec2 `json:"position"`
	HalfExtents         mgl32.Vec2 `json:"half_extents"`
	Rotation            float32    `json:"rotation"`
	FlatteningIntensity float32    `json:"flattening_intensity"`
	FlattenTerrain      bool       `json:"flatten_terrain"`
	_                   [3]uint8
}

type StampSplineConnection struct {
	Index   uint8 `json:"index"`
	Segment uint8 `json:"segment"`
}

type StampEntry struct {
	Type             enum.StampEntryType `json:"type"`
	_                [3]uint8
	Direction        float32                      `json:"direction"`
	SatelliteType    enum.StampEntrySatelliteType `json:"satellite_type"`
	SplineConnection StampSplineConnection        `json:"spline_connection"`
	_                [1]uint8
}

type rawStampSpline struct {
	SplineType  enum.StampSplineType
	_           [4]uint8
	Segments    DLArray
	Entrypoints [2]int8
	_           [6]uint8
}

type StampSpline struct {
	SplineType  enum.StampSplineType `json:"spline_type"`
	Segments    []mgl32.Vec2         `json:"segments"`
	Entrypoints []int8               `json:"entrypoints"`
}

func (s rawStampSpline) Deserialize(r io.ReadSeeker, base int64) (*StampSpline, error) {
	segments, err := ResolveDLArray[mgl32.Vec2](s.Segments, r, base)
	if err != nil {
		return nil, err
	}
	return &StampSpline{
		SplineType:  s.SplineType,
		Segments:    segments,
		Entrypoints: s.Entrypoints[:],
	}, nil
}

type rawStampExtension struct {
	CollisionCircles    DLArray
	CollisionRectangles DLArray
	LayerName           stingray.ThinHash
	Type                enum.StampExtensionType
	_                   [3]uint8
	WeightNoSpawn       float32
	WeightSpawnLayer    float32
	WeightSpawnNature   float32
}

type StampExtension struct {
	CollisionCircles    []StampCircle
	CollisionRectangles []StampRectangle
	LayerName           stingray.ThinHash
	Type                enum.StampExtensionType
	WeightNoSpawn       float32
	WeightSpawnLayer    float32
	WeightSpawnNature   float32
}

type SimpleStampExtension struct {
	CollisionCircles    []StampCircle           `json:"collision_circles"`
	CollisionRectangles []StampRectangle        `json:"collision_rectangles"`
	LayerName           string                  `json:"layer_name"`
	Type                enum.StampExtensionType `json:"type"`
	WeightNoSpawn       float32                 `json:"weight_no_spawn"`
	WeightSpawnLayer    float32                 `json:"weight_spawn_layer"`
	WeightSpawnNature   float32                 `json:"weaight_spawn_nature"`
}

func (s StampExtension) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleStampExtension {
	return SimpleStampExtension{
		CollisionCircles:    s.CollisionCircles,
		CollisionRectangles: s.CollisionRectangles,
		LayerName:           lookupThinHash(s.LayerName),
		Type:                s.Type,
		WeightNoSpawn:       s.WeightNoSpawn,
		WeightSpawnLayer:    s.WeightSpawnLayer,
		WeightSpawnNature:   s.WeightSpawnNature,
	}
}

func (s rawStampExtension) Deserialize(r io.ReadSeeker, base int64) (*StampExtension, error) {
	circles, err := ResolveDLArray[StampCircle](s.CollisionCircles, r, base)
	if err != nil {
		return nil, err
	}
	rects, err := ResolveDLArray[StampRectangle](s.CollisionRectangles, r, base)
	if err != nil {
		return nil, err
	}
	return &StampExtension{
		CollisionCircles:    circles,
		CollisionRectangles: rects,
		LayerName:           s.LayerName,
		Type:                s.Type,
		WeightNoSpawn:       s.WeightNoSpawn,
		WeightSpawnLayer:    s.WeightSpawnLayer,
		WeightSpawnNature:   s.WeightSpawnNature,
	}, nil
}

type StampProxy struct {
	UnitResource stingray.Hash
	Rotation     mgl32.Vec4
	Position     mgl32.Vec3
	Scale        mgl32.Vec3
}

type SimpleStampProxy struct {
	UnitResource string     `json:"unit_resource"`
	Rotation     mgl32.Vec4 `json:"rotation"`
	Position     mgl32.Vec3 `json:"position"`
	Scale        mgl32.Vec3 `json:"scale"`
}

func (s StampProxy) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleStampProxy {
	return SimpleStampProxy{
		UnitResource: lookupHash(s.UnitResource),
		Rotation:     s.Rotation,
		Position:     s.Position,
		Scale:        s.Scale,
	}
}

type Stamp struct {
	CollisionCircles    []StampCircle
	CollisionRectangles []StampRectangle
	Entrypoints         []StampEntry
	Splines             []StampSpline
	Extensions          []StampExtension
	Proxies             []StampProxy
	RotationType        enum.RotationType
}

type SimpleStamp struct {
	CollisionCircles    []StampCircle          `json:"collision_circles"`
	CollisionRectangles []StampRectangle       `json:"collision_rectangles"`
	Entrypoints         []StampEntry           `json:"entrypoints"`
	Splines             []StampSpline          `json:"splines"`
	Extensions          []SimpleStampExtension `json:"extensions"`
	Proxies             []SimpleStampProxy     `json:"proxies"`
	RotationType        enum.RotationType      `json:"rotation_type"`
}

func (s Stamp) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleStamp {
	extensions := make([]SimpleStampExtension, 0)
	for _, extension := range s.Extensions {
		extensions = append(extensions, extension.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	proxies := make([]SimpleStampProxy, 0)
	for _, proxy := range s.Proxies {
		proxies = append(proxies, proxy.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleStamp{
		CollisionCircles:    s.CollisionCircles,
		CollisionRectangles: s.CollisionRectangles,
		Entrypoints:         s.Entrypoints,
		Splines:             s.Splines,
		Extensions:          extensions,
		Proxies:             proxies,
		RotationType:        s.RotationType,
	}
}

func (s rawStamp) Deserialize(r io.ReadSeeker, base int64) (*Stamp, error) {
	collisionCircles, err := ResolveDLArray[StampCircle](s.CollisionCircles, r, base)
	if err != nil {
		return nil, err
	}
	collisionRectangles, err := ResolveDLArray[StampRectangle](s.CollisionRectangles, r, base)
	if err != nil {
		return nil, err
	}
	entrypoints, err := ResolveDLArray[StampEntry](s.Entrypoints, r, base)
	if err != nil {
		return nil, err
	}
	splines, err := ResolveDLArray[StampSpline](s.Splines, r, base)
	if err != nil {
		return nil, err
	}
	extensions, err := ResolveDLArray[StampExtension](s.Extensions, r, base)
	if err != nil {
		return nil, err
	}
	proxies, err := ResolveDLArray[StampProxy](s.Proxies, r, base)
	if err != nil {
		return nil, err
	}

	return &Stamp{
		CollisionCircles:    collisionCircles,
		CollisionRectangles: collisionRectangles,
		Entrypoints:         entrypoints,
		Splines:             splines,
		Extensions:          extensions,
		Proxies:             proxies,
		RotationType:        s.RotationType,
	}, nil
}

type StampInfo struct {
	*Stamp
	Path                stingray.Hash
	Weights             StampWeights
	Weight              float32
	FlatteningIntensity float32
	UnkFloat1           float32
	StampBitfield
}

type SimpleStampInfo struct {
	Stamp               *SimpleStamp `json:"stamp,omitempty"`
	Path                string       `json:"path"`
	Weights             StampWeights `json:"weights,omitzero"`
	Weight              float32      `json:"weight"`
	FlatteningIntensity float32      `json:"flattening_intensity"`
	UnkFloat1           float32      `json:"unk_float"`
	StampBitfield       `json:"unk_bitfield"`
}

func (v StampInfo) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleStampInfo {
	var simpleStamp *SimpleStamp
	if v.Stamp != nil {
		value := v.Stamp.ToSimple(lookupHash, lookupThinHash, lookupStrings)
		simpleStamp = &value
	}

	return SimpleStampInfo{
		Stamp:               simpleStamp,
		Path:                lookupHash(v.Path),
		Weights:             v.Weights,
		Weight:              v.Weight,
		FlatteningIntensity: v.FlatteningIntensity,
		UnkFloat1:           v.UnkFloat1,
		StampBitfield:       v.StampBitfield,
	}
}

func (s rawStampInfo) Deserialize(r io.ReadSeeker, base int64) (*StampInfo, error) {
	raw, err := ResolveDLPtr[rawStamp](s.Stamp, r, base)
	if err != nil {
		return nil, err
	}
	var stamp *Stamp
	if raw != nil {
		stamp, err = raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
	}
	return &StampInfo{
		Stamp:               stamp,
		Path:                s.Path,
		Weights:             s.Weights,
		Weight:              s.Weight,
		FlatteningIntensity: s.FlatteningIntensity,
		UnkFloat1:           s.UnkFloat1,
		StampBitfield:       s.StampBitfield,
	}, nil
}

type StampOverlapSettings struct {
	StampOverlapGroup enum.StampOverlapGroups `json:"stamp_overlap_group"`
	_                 [3]uint8
	StampMaxOverlap   float32 `json:"stamp_max_overlap"`
	StampMinOverlap   float32 `json:"stamp_min_overlap"`
}

type StampGroup struct {
	Stamps                      []StampInfo
	GridCellSize                float32
	UnkFloat1                   float32
	MaxRandomOffsetInCell       float32
	FillRate                    float32
	MaxPossibleCellsPerMap      int32
	Flags                       StampGroupFlags
	StampMaxOverlap             float32
	StampMinOverlap             float32
	StampOverlapGroup           enum.StampOverlapGroups
	StampOverlapSettings        []StampOverlapSettings
	RouteStampOffset            float32
	MaxDistanceFromRoute        float32
	MinDistanceFromRoute        float32
	MaxDistanceFromPlayableArea float32
	DifficultyBitfield
	UnkFloat2 float32
}

type SimpleStampGroup struct {
	Stamps                      []SimpleStampInfo       `json:"stamps"`
	GridCellSize                float32                 `json:"grid_cell_size"`
	UnkFloat1                   float32                 `json:"unk_float1"`
	MaxRandomOffsetInCell       float32                 `json:"max_random_offset_in_cell"`
	FillRate                    float32                 `json:"fill_rate"`
	MaxPossibleCellsPerMap      int32                   `json:"max_possible_cells_per_map"`
	Flags                       StampGroupFlags         `json:"flags"`
	StampMaxOverlap             float32                 `json:"stamp_max_overlap"`
	StampMinOverlap             float32                 `json:"stamp_min_overlap"`
	StampOverlapGroup           enum.StampOverlapGroups `json:"stamp_overlap_group"`
	StampOverlapSettings        []StampOverlapSettings  `json:"stamp_overlap_settings"`
	RouteStampOffset            float32                 `json:"route_stamp_offset"`
	MaxDistanceFromRoute        float32                 `json:"max_distance_from_route"`
	MinDistanceFromRoute        float32                 `json:"min_distance_from_route"`
	MaxDistanceFromPlayableArea float32                 `json:"max_distance_from_playable_area"`
	DifficultyBitfield          `json:"difficulty_bitfield"`
	UnkFloat2                   float32 `json:"unk_float2"`
}

func (v StampGroup) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleStampGroup {
	stamps := make([]SimpleStampInfo, 0)
	for _, stamp := range v.Stamps {
		stamps = append(stamps, stamp.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleStampGroup{
		Stamps:                      stamps,
		GridCellSize:                v.GridCellSize,
		UnkFloat1:                   v.UnkFloat1,
		MaxRandomOffsetInCell:       v.MaxRandomOffsetInCell,
		FillRate:                    v.FillRate,
		MaxPossibleCellsPerMap:      v.MaxPossibleCellsPerMap,
		Flags:                       v.Flags,
		StampMaxOverlap:             v.StampMaxOverlap,
		StampMinOverlap:             v.StampMinOverlap,
		StampOverlapGroup:           v.StampOverlapGroup,
		StampOverlapSettings:        v.StampOverlapSettings,
		RouteStampOffset:            v.RouteStampOffset,
		MaxDistanceFromRoute:        v.MaxDistanceFromRoute,
		MinDistanceFromRoute:        v.MinDistanceFromRoute,
		MaxDistanceFromPlayableArea: v.MaxDistanceFromPlayableArea,
		DifficultyBitfield:          v.DifficultyBitfield,
		UnkFloat2:                   v.UnkFloat2,
	}
}

func (g rawStampGroup) Deserialize(r io.ReadSeeker, base int64) (*StampGroup, error) {
	rawStamps, err := ResolveDLArray[rawStampInfo](g.Stamps, r, base)
	if err != nil {
		return nil, err
	}

	stampOverlapSettings, err := ResolveDLArray[StampOverlapSettings](g.StampOverlapSettings, r, base)
	if err != nil {
		return nil, err
	}

	stamps := make([]StampInfo, 0)
	for _, raw := range rawStamps {
		stamp, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		stamps = append(stamps, *stamp)
	}

	return &StampGroup{
		Stamps:                      stamps,
		GridCellSize:                g.GridCellSize,
		UnkFloat1:                   g.UnkFloat1,
		MaxRandomOffsetInCell:       g.MaxRandomOffsetInCell,
		FillRate:                    g.FillRate,
		MaxPossibleCellsPerMap:      g.MaxPossibleCellsPerMap,
		Flags:                       g.Flags,
		StampMaxOverlap:             g.StampMaxOverlap,
		StampMinOverlap:             g.StampMinOverlap,
		StampOverlapGroup:           g.StampOverlapGroup,
		StampOverlapSettings:        stampOverlapSettings,
		RouteStampOffset:            g.RouteStampOffset,
		MaxDistanceFromRoute:        g.MaxDistanceFromRoute,
		MinDistanceFromRoute:        g.MinDistanceFromRoute,
		MaxDistanceFromPlayableArea: g.MaxDistanceFromPlayableArea,
		DifficultyBitfield:          g.DifficultyBitfield,
		UnkFloat2:                   g.UnkFloat2,
	}, nil
}

type ShaderValue struct {
	mgl32.Vec4
	Count uint32
}

type ShaderSetting struct {
	Name   stingray.ThinHash
	Value  float32
	Vector ShaderValue
}

type SimpleShaderValue struct {
	X *float32 `json:"x,omitempty"`
	Y *float32 `json:"y,omitempty"`
	Z *float32 `json:"z,omitempty"`
	W *float32 `json:"w,omitempty"`
}

func (v ShaderValue) ToSimple(_ HashLookup, _ ThinHashLookup, _ StringsLookup) *SimpleShaderValue {
	var vector *SimpleShaderValue
	if v.Count > 0 {
		vector = &SimpleShaderValue{}
		if v.Count >= 1 {
			x := v.X()
			vector.X = &x
		}
		if v.Count >= 2 {
			y := v.Y()
			vector.Y = &y
		}
		if v.Count >= 3 {
			z := v.Z()
			vector.Z = &z
		}
		if v.Count >= 4 {
			w := v.W()
			vector.W = &w
		}
	}
	return vector
}

type SimpleShaderSetting struct {
	Name   string             `json:"name"`
	Value  *float32           `json:"value,omitempty"`
	Vector *SimpleShaderValue `json:"vector,omitempty"`
}

func (setting ShaderSetting) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleShaderSetting {
	vector := setting.Vector.ToSimple(lookupHash, lookupThinHash, lookupStrings)
	var val *float32
	if vector == nil {
		val = &setting.Value
	}

	return SimpleShaderSetting{
		Name:   lookupThinHash(setting.Name),
		Value:  val,
		Vector: vector,
	}
}

type rawShaderProperties struct {
	Shader   stingray.Hash
	Settings DLArray
}

type ShaderProperties struct {
	Shader   stingray.Hash
	Settings []ShaderSetting
}

type SimpleShaderProperties struct {
	Shader   string                `json:"shader"`
	Settings []SimpleShaderSetting `json:"settings,omitempty"`
}

func (p ShaderProperties) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleShaderProperties {
	settings := make([]SimpleShaderSetting, 0)
	for _, setting := range p.Settings {
		settings = append(settings, setting.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleShaderProperties{
		Shader:   lookupHash(p.Shader),
		Settings: settings,
	}
}

func (p ShaderProperties) Map() map[stingray.ThinHash]ShaderSetting {
	result := make(map[stingray.ThinHash]ShaderSetting)
	for _, setting := range p.Settings {
		result[setting.Name] = setting
	}
	return result
}

func (g rawShaderProperties) Deserialize(r io.ReadSeeker, base int64) (*ShaderProperties, error) {
	settings, err := ResolveDLArray[ShaderSetting](g.Settings, r, base)
	if err != nil {
		return nil, err
	}
	return &ShaderProperties{
		Shader:   g.Shader,
		Settings: settings,
	}, nil
}

type FogVolumeShaderSetting struct {
	Name       stingray.ThinHash
	Value      float32
	Vector     ShaderValue
	ColorGroup enum.ZoneFogColorGroups
	_          [3]uint8
}

type SimpleFogVolumeShaderSetting struct {
	Name       string                  `json:"name"`
	Value      *float32                `json:"value,omitempty"`
	Vector     *SimpleShaderValue      `json:"vector,omitempty"`
	ColorGroup enum.ZoneFogColorGroups `json:"color_group"`
}

func (setting FogVolumeShaderSetting) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleFogVolumeShaderSetting {
	vector := setting.Vector.ToSimple(lookupHash, lookupThinHash, lookupStrings)
	var val *float32
	if vector == nil {
		val = &setting.Value
	}

	return SimpleFogVolumeShaderSetting{
		Name:       lookupThinHash(setting.Name),
		Value:      val,
		Vector:     vector,
		ColorGroup: setting.ColorGroup,
	}
}

type rawFogVolumeShaderProperties struct {
	Shader   stingray.Hash
	Settings DLArray
}

type FogVolumeShaderProperties struct {
	Shader   stingray.Hash
	Settings []FogVolumeShaderSetting
}

type SimpleFogVolumeShaderProperties struct {
	Shader   string                         `json:"shader"`
	Settings []SimpleFogVolumeShaderSetting `json:"settings,omitempty"`
}

func (p FogVolumeShaderProperties) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleFogVolumeShaderProperties {
	settings := make([]SimpleFogVolumeShaderSetting, 0)
	for _, setting := range p.Settings {
		settings = append(settings, setting.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleFogVolumeShaderProperties{
		Shader:   lookupHash(p.Shader),
		Settings: settings,
	}
}

func (g rawFogVolumeShaderProperties) Deserialize(r io.ReadSeeker, base int64) (*FogVolumeShaderProperties, error) {
	settings, err := ResolveDLArray[FogVolumeShaderSetting](g.Settings, r, base)
	if err != nil {
		return nil, err
	}
	return &FogVolumeShaderProperties{
		Shader:   g.Shader,
		Settings: settings,
	}, nil
}

type rawCameraEffect struct {
	RotationType               enum.RotationType
	EffectType                 enum.EffectType
	OverridesEffectTypes       DLArray
	StartAtWindSpeed           float32
	StopAtWindSpeed            float32
	StartAtNightAmount         float32
	StopAtNightAmount          float32
	StartAtSunsetAmount        float32
	StopAtSunsetAmount         float32
	CameraEffectId             stingray.Hash
	OverriddenResponseEffectId stingray.Hash
	WeatherTypes               DLArray
}

type CameraEffect struct {
	RotationType               enum.RotationType
	EffectType                 enum.EffectType
	OverridesEffectTypes       []enum.EffectType
	StartAtWindSpeed           float32
	StopAtWindSpeed            float32
	StartAtNightAmount         float32
	StopAtNightAmount          float32
	StartAtSunsetAmount        float32
	StopAtSunsetAmount         float32
	CameraEffectId             stingray.Hash
	OverriddenResponseEffectId stingray.Hash
	WeatherTypes               []enum.WeatherType
}

type SimpleCameraEffect struct {
	RotationType               enum.RotationType  `json:"rotation_type"`
	EffectType                 enum.EffectType    `json:"effect_type"`
	OverridesEffectTypes       []enum.EffectType  `json:"overrides_effect_types"`
	StartAtWindSpeed           float32            `json:"start_at_wind_speed"`
	StopAtWindSpeed            float32            `json:"stop_at_wind_speed"`
	StartAtNightAmount         float32            `json:"start_at_night_amount"`
	StopAtNightAmount          float32            `json:"stop_at_night_amount"`
	StartAtSunsetAmount        float32            `json:"start_at_sunset_amount"`
	StopAtSunsetAmount         float32            `json:"stop_at_sunset_amount"`
	CameraEffectId             string             `json:"camera_effect_id"`
	OverriddenResponseEffectId string             `json:"overridden_response_effect_id"`
	WeatherTypes               []enum.WeatherType `json:"weather_types"`
}

func (v CameraEffect) ToSimple(lookupHash HashLookup, _ ThinHashLookup, _ StringsLookup) SimpleCameraEffect {
	return SimpleCameraEffect{
		RotationType:               v.RotationType,
		EffectType:                 v.EffectType,
		OverridesEffectTypes:       v.OverridesEffectTypes,
		StartAtWindSpeed:           v.StartAtWindSpeed,
		StopAtWindSpeed:            v.StopAtWindSpeed,
		StartAtNightAmount:         v.StartAtNightAmount,
		StopAtNightAmount:          v.StopAtNightAmount,
		StartAtSunsetAmount:        v.StartAtSunsetAmount,
		StopAtSunsetAmount:         v.StopAtSunsetAmount,
		CameraEffectId:             lookupHash(v.CameraEffectId),
		OverriddenResponseEffectId: lookupHash(v.OverriddenResponseEffectId),
		WeatherTypes:               v.WeatherTypes,
	}
}

func (g rawCameraEffect) Deserialize(r io.ReadSeeker, base int64) (*CameraEffect, error) {
	overridesEffectTypes, err := ResolveDLArray[enum.EffectType](g.OverridesEffectTypes, r, base)
	if err != nil {
		return nil, err
	}
	weatherTypes, err := ResolveDLArray[enum.WeatherType](g.WeatherTypes, r, base)
	if err != nil {
		return nil, err
	}
	return &CameraEffect{
		RotationType:               g.RotationType,
		EffectType:                 g.EffectType,
		OverridesEffectTypes:       overridesEffectTypes,
		StartAtWindSpeed:           g.StartAtWindSpeed,
		StopAtWindSpeed:            g.StopAtWindSpeed,
		StartAtNightAmount:         g.StartAtNightAmount,
		StopAtNightAmount:          g.StopAtNightAmount,
		StartAtSunsetAmount:        g.StartAtSunsetAmount,
		StopAtSunsetAmount:         g.StopAtSunsetAmount,
		CameraEffectId:             g.CameraEffectId,
		OverriddenResponseEffectId: g.OverriddenResponseEffectId,
		WeatherTypes:               weatherTypes,
	}, nil
}

type ZoneSettings struct {
	StampGroups                    []StampGroup
	MaterialLookupUnit             stingray.Hash
	ColorVariationGenerators       []ShaderProperties
	ShaderProperties1              []ShaderProperties
	FogVolumeGenerators            []FogVolumeShaderProperties
	MinimapVisualizationGenerators []ShaderProperties
	MaterialGenerators             []ShaderProperties
	ShaderProperties2              []ShaderProperties
	HeightGenerators               []ShaderProperties
	DefaultMaterial                int32
	MinimapColor                   mgl32.Vec3
	MinimapTerrainType             enum.MinimapTerrainType
	ScatterSetting                 *string
	CameraEnvironmentEffects       []CameraEffect
	WindNoiseIds                   WindNoiseSettings
	WaterMaterial                  stingray.Hash
	WaterLevelOffset               float32
	WaterHeight                    float32
	WaterZoneOpacityCutoff         float32
	MaxWaterDepth                  float32
	HeightModificationCurve        []float32
	ReverbZone                     *string
	AmbienceSoundId                stingray.ThinHash
	QuakeEffect                    QuakeEffectSettings
	ZoneMusicType                  enum.ZoneMusicType
	BanterEventType                enum.BanterEventType
	PlayerKillHeight               float32
	MinimapScatter                 MinimapScatterSet
	LocationHeightOffset           float32
	UnkString                      *string
	UnkThinHash1                   stingray.ThinHash
	UnkThinHash2                   stingray.ThinHash
	UnkBitfield1                   uint8
	PatrolSetting                  enum.PatrolSetting
	PatrolFloat1                   float32
	PatrolFloat2                   float32
	PatrolFloat3                   float32
	PatrolFloat4                   float32
	PatrolFloat5                   float32
	PatrolFloat6                   float32
	PatrolFloat7                   float32
	PatrolFloat8                   float32
	PatrolFloat9                   float32
	PatrolFloat10                  float32
	PatrolFloat11                  float32
	UnkBool1                       bool
	UnkBitfield2                   uint8
	CityPlacementHeuristic         enum.CityPlacementHeuristic
	UnknownEnum                    uint32
	UnkInt1                        int32
	UnkInt2                        uint32
	Vectors1                       [3]mgl32.Vec2
	Vectors2                       [3]mgl32.Vec2
	Vectors3                       [3]mgl32.Vec2
	UnkFloat1                      float32
	UnkFloat2                      float32
	UnkFloat3                      float32
	UnkFloat4                      float32
	LocationFlags                  []enum.LocationFlag
	ColoniesCurbPath               stingray.Hash
	CurvedCurbPath                 stingray.Hash
	UnkFLoat5                      float32
	UnkFLoat6                      float32
	UnkInt3                        uint32
	DetailUnit1                    stingray.Hash
	DetailUnit2                    stingray.Hash
	DetailUnit3                    stingray.Hash
	TrafficlightPrefab             stingray.Hash
	UnkFloat7                      float32
	UnkFloat8                      float32
	UnkFloat9                      float32
	UnkFloat10                     float32
	UnkFloat11                     float32
	CityStreetlights               stingray.Hash
	UnkFloat12                     float32
	UnkFloat13                     float32
	UnkFloat14                     float32
	DetailUnit4                    stingray.Hash
	UnkFloat15                     float32
	UnkFloat16                     float32
	UnkFloat17                     float32
	UnkFloat18                     float32
	UnkFloat19                     float32
	UnkFloat20                     float32
	UnkInt4                        uint64
	UnkString2                     *string
	UnkFloat21                     float32
	UnkFloat22                     float32
	BoundaryWallSettings           BoundaryWallSettings
	RoadEmbankmentUnits            []stingray.Hash
	UnkInt5                        uint32
	UnkBitfield3                   uint8
}

type SimpleZoneSettings struct {
	StampGroups                    []SimpleStampGroup                `json:"stamp_groups"`
	MaterialLookupUnit             string                            `json:"material_lookup_unit"`
	ColorVariationGenerators       []SimpleShaderProperties          `json:"color_variation_generators"`
	ShaderProperties1              []SimpleShaderProperties          `json:"shader_properties1"`
	FogVolumeGenerators            []SimpleFogVolumeShaderProperties `json:"fog_volume_generators"`
	MinimapVisualizationGenerators []SimpleShaderProperties          `json:"minimap_visualization_generators"`
	MaterialGenerators             []SimpleShaderProperties          `json:"material_generators"`
	ShaderProperties2              []SimpleShaderProperties          `json:"shader_properties2"`
	HeightGenerators               []SimpleShaderProperties          `json:"height_generators"`
	DefaultMaterial                int32                             `json:"default_material"`
	MinimapColor                   mgl32.Vec3                        `json:"minimap_color"`
	MinimapTerrainType             enum.MinimapTerrainType           `json:"minimap_terrain_type"`
	ScatterSetting                 *string                           `json:"scatter_setting"`
	CameraEnvironmentEffects       []SimpleCameraEffect              `json:"camera_environment_effects"`
	WindNoiseIds                   WindNoiseSettings                 `json:"wind_noise_ids"`
	WaterMaterial                  string                            `json:"water_material"`
	WaterLevelOffset               float32                           `json:"water_level_offset"`
	WaterHeight                    float32                           `json:"water_height"`
	WaterZoneOpacityCutoff         float32                           `json:"water_zone_opacity_cutoff"`
	MaxWaterDepth                  float32                           `json:"max_water_depth"`
	HeightModificationCurve        []float32                         `json:"height_modification_curve"`
	ReverbZone                     *string                           `json:"reverb_zone"`
	AmbienceSoundId                string                            `json:"ambience_sound_id"`
	QuakeEffect                    SimpleQuakeEffectSettings         `json:"quake_effect"`
	ZoneMusicType                  enum.ZoneMusicType                `json:"zone_music_type"`
	BanterEventType                enum.BanterEventType              `json:"banter_event_type"`
	PlayerKillHeight               float32                           `json:"player_kill_height"`
	MinimapScatter                 SimpleMinimapScatterSet           `json:"minimap_scatter"`
	LocationHeightOffset           float32                           `json:"location_height_offset"`
	UnkString                      *string                           `json:"unk_string"`
	UnkThinHash1                   string                            `json:"unk_thin_hash1"`
	UnkThinHash2                   string                            `json:"unk_thin_hash2"`
	UnkBitfield1                   uint8                             `json:"unk_bitfield1"`
	PatrolSetting                  enum.PatrolSetting                `json:"patrol_setting"`
	PatrolFloat1                   float32                           `json:"patrol_float1"`
	PatrolFloat2                   float32                           `json:"patrol_float2"`
	PatrolFloat3                   float32                           `json:"patrol_float3"`
	PatrolFloat4                   float32                           `json:"patrol_float4"`
	PatrolFloat5                   float32                           `json:"patrol_float5"`
	PatrolFloat6                   float32                           `json:"patrol_float6"`
	PatrolFloat7                   float32                           `json:"patrol_float7"`
	PatrolFloat8                   float32                           `json:"patrol_float8"`
	PatrolFloat9                   float32                           `json:"patrol_float9"`
	PatrolFloat10                  float32                           `json:"patrol_float10"`
	PatrolFloat11                  float32                           `json:"patrol_float11"`
	UnkBool1                       bool                              `json:"unk_bool1"`
	UnkBitfield2                   uint8                             `json:"unk_bitfield2"`
	CityPlacementHeuristic         enum.CityPlacementHeuristic       `json:"city_placement_heuristic"`
	UnknownEnum                    uint32                            `json:"unknown_enum"`
	UnkInt1                        int32                             `json:"unk_int1"`
	UnkInt2                        uint32                            `json:"unk_int2"`
	Vectors1                       [3]mgl32.Vec2                     `json:"vectors1"`
	Vectors2                       [3]mgl32.Vec2                     `json:"vectors2"`
	Vectors3                       [3]mgl32.Vec2                     `json:"vectors3"`
	UnkFloat1                      float32                           `json:"unk_float1"`
	UnkFloat2                      float32                           `json:"unk_float2"`
	UnkFloat3                      float32                           `json:"unk_float3"`
	UnkFloat4                      float32                           `json:"unk_float4"`
	LocationFlags                  []enum.LocationFlag               `json:"location_flags"`
	ColoniesCurbPath               string                            `json:"colonies_curb_path"`
	CurvedCurbPath                 string                            `json:"curved_curb_path"`
	UnkFloat5                      float32                           `json:"unk_float5"`
	UnkFloat6                      float32                           `json:"unk_float6"`
	UnkInt3                        uint32                            `json:"unk_int3"`
	DetailUnit1                    string                            `json:"detail_unit1"`
	DetailUnit2                    string                            `json:"detail_unit2"`
	DetailUnit3                    string                            `json:"detail_unit3"`
	TrafficlightPrefab             string                            `json:"trafficlight_prefab"`
	UnkFloat7                      float32                           `json:"unk_float7"`
	UnkFloat8                      float32                           `json:"unk_float8"`
	UnkFloat9                      float32                           `json:"unk_float9"`
	UnkFloat10                     float32                           `json:"unk_float10"`
	UnkFloat11                     float32                           `json:"unk_float11"`
	CityStreetlights               string                            `json:"city_streetlights"`
	UnkFloat12                     float32                           `json:"unk_float12"`
	UnkFloat13                     float32                           `json:"unk_float13"`
	UnkFloat14                     float32                           `json:"unk_float14"`
	DetailUnit4                    string                            `json:"detail_unit4"`
	UnkFloat15                     float32                           `json:"unk_float15"`
	UnkFloat16                     float32                           `json:"unk_float16"`
	UnkFloat17                     float32                           `json:"unk_float17"`
	UnkFloat18                     float32                           `json:"unk_float18"`
	UnkFloat19                     float32                           `json:"unk_float19"`
	UnkFloat20                     float32                           `json:"unk_float20"`
	UnkInt4                        uint64                            `json:"unk_int4"`
	UnkString2                     *string                           `json:"unk_string2"`
	UnkFloat21                     float32                           `json:"unk_float21"`
	UnkFloat22                     float32                           `json:"unk_float22"`
	BoundaryWallSettings           SimpleBoundaryWallSettings        `json:"boundary_wall_settings"`
	RoadEmbankmentUnits            []string                          `json:"road_embankment_units"`
	UnkInt5                        uint32                            `json:"unk_int5"`
	UnkBitfield3                   uint8                             `json:"unk_bitfield3"`
}

func (z ZoneSettings) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleZoneSettings {
	stampGroups := make([]SimpleStampGroup, 0)
	for _, group := range z.StampGroups {
		stampGroups = append(stampGroups, group.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	colorVariationGenerators := make([]SimpleShaderProperties, 0)
	for _, prop := range z.ColorVariationGenerators {
		colorVariationGenerators = append(colorVariationGenerators, prop.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	shaderProperties1 := make([]SimpleShaderProperties, 0)
	for _, prop := range z.ShaderProperties1 {
		shaderProperties1 = append(shaderProperties1, prop.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	fogVolumeGenerators := make([]SimpleFogVolumeShaderProperties, 0)
	for _, prop := range z.FogVolumeGenerators {
		fogVolumeGenerators = append(fogVolumeGenerators, prop.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	minimapVisualizationGenerators := make([]SimpleShaderProperties, 0)
	for _, prop := range z.MinimapVisualizationGenerators {
		minimapVisualizationGenerators = append(minimapVisualizationGenerators, prop.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	materialGenerators := make([]SimpleShaderProperties, 0)
	for _, prop := range z.MaterialGenerators {
		materialGenerators = append(materialGenerators, prop.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	shaderProperties2 := make([]SimpleShaderProperties, 0)
	for _, prop := range z.ShaderProperties2 {
		shaderProperties2 = append(shaderProperties2, prop.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	heightGenerators := make([]SimpleShaderProperties, 0)
	for _, prop := range z.HeightGenerators {
		heightGenerators = append(heightGenerators, prop.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	roadEmbankmentUnits := make([]string, 0)
	for _, unit := range z.RoadEmbankmentUnits {
		roadEmbankmentUnits = append(roadEmbankmentUnits, lookupHash(unit))
	}

	cameraEffects := make([]SimpleCameraEffect, 0)
	for _, effect := range z.CameraEnvironmentEffects {
		cameraEffects = append(cameraEffects, effect.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleZoneSettings{
		StampGroups:                    stampGroups,
		MaterialLookupUnit:             lookupHash(z.MaterialLookupUnit),
		ColorVariationGenerators:       colorVariationGenerators,
		ShaderProperties1:              shaderProperties1,
		FogVolumeGenerators:            fogVolumeGenerators,
		MinimapVisualizationGenerators: minimapVisualizationGenerators,
		MaterialGenerators:             materialGenerators,
		ShaderProperties2:              shaderProperties2,
		HeightGenerators:               heightGenerators,
		DefaultMaterial:                z.DefaultMaterial,
		MinimapColor:                   z.MinimapColor,
		MinimapTerrainType:             z.MinimapTerrainType,
		ScatterSetting:                 z.ScatterSetting,
		CameraEnvironmentEffects:       cameraEffects,
		WindNoiseIds:                   z.WindNoiseIds,
		WaterMaterial:                  lookupHash(z.WaterMaterial),
		WaterLevelOffset:               z.WaterLevelOffset,
		WaterHeight:                    z.WaterHeight,
		WaterZoneOpacityCutoff:         z.WaterZoneOpacityCutoff,
		MaxWaterDepth:                  z.MaxWaterDepth,
		HeightModificationCurve:        z.HeightModificationCurve,
		ReverbZone:                     z.ReverbZone,
		AmbienceSoundId:                lookupThinHash(z.AmbienceSoundId),
		QuakeEffect:                    z.QuakeEffect.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		ZoneMusicType:                  z.ZoneMusicType,
		BanterEventType:                z.BanterEventType,
		PlayerKillHeight:               z.PlayerKillHeight,
		MinimapScatter:                 z.MinimapScatter.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		LocationHeightOffset:           z.LocationHeightOffset,
		UnkString:                      z.UnkString,
		UnkThinHash1:                   lookupThinHash(z.UnkThinHash1),
		UnkThinHash2:                   lookupThinHash(z.UnkThinHash2),
		UnkBitfield1:                   z.UnkBitfield1,
		PatrolSetting:                  z.PatrolSetting,
		PatrolFloat1:                   z.PatrolFloat1,
		PatrolFloat2:                   z.PatrolFloat2,
		PatrolFloat3:                   z.PatrolFloat3,
		PatrolFloat4:                   z.PatrolFloat4,
		PatrolFloat5:                   z.PatrolFloat5,
		PatrolFloat6:                   z.PatrolFloat6,
		PatrolFloat7:                   z.PatrolFloat7,
		PatrolFloat8:                   z.PatrolFloat8,
		PatrolFloat9:                   z.PatrolFloat9,
		PatrolFloat10:                  z.PatrolFloat10,
		PatrolFloat11:                  z.PatrolFloat11,
		UnkBool1:                       z.UnkBool1,
		UnkBitfield2:                   z.UnkBitfield2,
		CityPlacementHeuristic:         z.CityPlacementHeuristic,
		UnknownEnum:                    z.UnknownEnum,
		UnkInt1:                        z.UnkInt1,
		UnkInt2:                        z.UnkInt2,
		Vectors1:                       z.Vectors1,
		Vectors2:                       z.Vectors2,
		Vectors3:                       z.Vectors3,
		UnkFloat1:                      z.UnkFloat1,
		UnkFloat2:                      z.UnkFloat2,
		UnkFloat3:                      z.UnkFloat3,
		UnkFloat4:                      z.UnkFloat4,
		LocationFlags:                  z.LocationFlags,
		ColoniesCurbPath:               lookupHash(z.ColoniesCurbPath),
		CurvedCurbPath:                 lookupHash(z.CurvedCurbPath),
		UnkFloat5:                      z.UnkFLoat5,
		UnkFloat6:                      z.UnkFLoat6,
		UnkInt3:                        z.UnkInt3,
		DetailUnit1:                    lookupHash(z.DetailUnit1),
		DetailUnit2:                    lookupHash(z.DetailUnit2),
		DetailUnit3:                    lookupHash(z.DetailUnit3),
		TrafficlightPrefab:             lookupHash(z.TrafficlightPrefab),
		UnkFloat7:                      z.UnkFloat7,
		UnkFloat8:                      z.UnkFloat8,
		UnkFloat9:                      z.UnkFloat9,
		UnkFloat10:                     z.UnkFloat10,
		UnkFloat11:                     z.UnkFloat11,
		CityStreetlights:               lookupHash(z.CityStreetlights),
		UnkFloat12:                     z.UnkFloat12,
		UnkFloat13:                     z.UnkFloat13,
		UnkFloat14:                     z.UnkFloat14,
		DetailUnit4:                    lookupHash(z.DetailUnit4),
		UnkFloat15:                     z.UnkFloat15,
		UnkFloat16:                     z.UnkFloat16,
		UnkFloat17:                     z.UnkFloat17,
		UnkFloat18:                     z.UnkFloat18,
		UnkFloat19:                     z.UnkFloat19,
		UnkFloat20:                     z.UnkFloat20,
		UnkInt4:                        z.UnkInt4,
		UnkString2:                     z.UnkString2,
		UnkFloat21:                     z.UnkFloat21,
		UnkFloat22:                     z.UnkFloat22,
		BoundaryWallSettings:           z.BoundaryWallSettings.ToSimple(lookupHash, lookupThinHash, lookupStrings),
		RoadEmbankmentUnits:            roadEmbankmentUnits,
		UnkInt5:                        z.UnkInt5,
		UnkBitfield3:                   z.UnkBitfield3,
	}
}

func (z rawZoneSettings) Deserialize(r io.ReadSeeker, base int64) (*ZoneSettings, error) {
	rawStampGroups, err := ResolveDLArray[rawStampGroup](z.StampGroups, r, base)
	if err != nil {
		return nil, err
	}
	rawColorVariationGenerators, err := ResolveDLArray[rawShaderProperties](z.ColorVariationGenerators, r, base)
	if err != nil {
		return nil, err
	}
	rawShaderProperties1, err := ResolveDLArray[rawShaderProperties](z.ShaderProperties1, r, base)
	if err != nil {
		return nil, err
	}
	rawFogVolumeGenerators, err := ResolveDLArray[rawFogVolumeShaderProperties](z.FogVolumeGenerators, r, base)
	if err != nil {
		return nil, err
	}
	rawMinimapVisualizationGenerators, err := ResolveDLArray[rawShaderProperties](z.MinimapVisualizationGenerators, r, base)
	if err != nil {
		return nil, err
	}
	rawMaterialGenerators, err := ResolveDLArray[rawShaderProperties](z.MaterialGenerators, r, base)
	if err != nil {
		return nil, err
	}
	rawShaderProperties2, err := ResolveDLArray[rawShaderProperties](z.ShaderProperties2, r, base)
	if err != nil {
		return nil, err
	}
	rawHeightGenerators, err := ResolveDLArray[rawShaderProperties](z.HeightGenerators, r, base)
	if err != nil {
		return nil, err
	}

	stampGroups := make([]StampGroup, 0)
	for _, raw := range rawStampGroups {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		stampGroups = append(stampGroups, *props)
	}
	colorVariationGenerators := make([]ShaderProperties, 0)
	for _, raw := range rawColorVariationGenerators {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		colorVariationGenerators = append(colorVariationGenerators, *props)
	}
	shaderProperties1 := make([]ShaderProperties, 0)
	for _, raw := range rawShaderProperties1 {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		shaderProperties1 = append(shaderProperties1, *props)
	}
	fogVolumeGenerators := make([]FogVolumeShaderProperties, 0)
	for _, raw := range rawFogVolumeGenerators {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		fogVolumeGenerators = append(fogVolumeGenerators, *props)
	}
	minimapVisualizationGenerators := make([]ShaderProperties, 0)
	for _, raw := range rawMinimapVisualizationGenerators {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		minimapVisualizationGenerators = append(minimapVisualizationGenerators, *props)
	}
	materialGenerators := make([]ShaderProperties, 0)
	for _, raw := range rawMaterialGenerators {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		materialGenerators = append(materialGenerators, *props)
	}
	shaderProperties2 := make([]ShaderProperties, 0)
	for _, raw := range rawShaderProperties2 {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		shaderProperties2 = append(shaderProperties2, *props)
	}
	heightGenerators := make([]ShaderProperties, 0)
	for _, raw := range rawHeightGenerators {
		props, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		heightGenerators = append(heightGenerators, *props)
	}

	var scatterSetting *string
	if z.ScatterSetting.Offset > 0 {
		if _, err := r.Seek(base+z.ScatterSetting.Offset, io.SeekStart); err != nil {
			return nil, err
		}
		scatterSettingVal, err := util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		scatterSetting = &scatterSettingVal
	}

	rawCameraEnvironmentEffects, err := ResolveDLArray[rawCameraEffect](z.CameraEnvironmentEffects, r, base)
	if err != nil {
		return nil, err
	}
	cameraEnvironmentEffects := make([]CameraEffect, 0)
	for _, raw := range rawCameraEnvironmentEffects {
		effect, err := raw.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		cameraEnvironmentEffects = append(cameraEnvironmentEffects, *effect)
	}

	heightModificationCurve, err := ResolveDLArray[float32](z.HeightModificationCurve, r, base)
	if err != nil {
		return nil, err
	}

	var reverbZone *string
	if z.ReverbZone.Offset > 0 {
		if _, err := r.Seek(base+z.ReverbZone.Offset, io.SeekStart); err != nil {
			return nil, err
		}
		reverbZoneVal, err := util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		reverbZone = &reverbZoneVal
	}

	minimapScatter, err := z.MinimapScatter.Deserialize(r, base)
	if err != nil {
		return nil, err
	}

	var unkString *string
	if z.UnkString.Offset > 0 {
		if _, err := r.Seek(base+z.UnkString.Offset, io.SeekStart); err != nil {
			return nil, err
		}
		unkStringVal, err := util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		unkString = &unkStringVal
	}

	var unkString2 *string
	if z.UnkString2.Offset > 0 {
		if _, err := r.Seek(base+z.UnkString2.Offset, io.SeekStart); err != nil {
			return nil, err
		}
		unkString2Val, err := util.ReadCString(r)
		if err != nil {
			return nil, err
		}
		unkString2 = &unkString2Val
	}

	boundaryWallSettings, err := z.BoundaryWallSettings.Deserialize(r, base)
	if err != nil {
		return nil, err
	}

	roadEmbankmentUnits, err := ResolveDLArray[stingray.Hash](z.RoadEmbankmentUnits, r, base)
	if err != nil {
		return nil, err
	}

	return &ZoneSettings{
		StampGroups:                    stampGroups,
		MaterialLookupUnit:             z.MaterialLookupUnit,
		ColorVariationGenerators:       colorVariationGenerators,
		ShaderProperties1:              shaderProperties1,
		FogVolumeGenerators:            fogVolumeGenerators,
		MinimapVisualizationGenerators: minimapVisualizationGenerators,
		MaterialGenerators:             materialGenerators,
		ShaderProperties2:              shaderProperties2,
		HeightGenerators:               heightGenerators,
		DefaultMaterial:                z.DefaultMaterial,
		MinimapColor:                   z.MinimapColor,
		MinimapTerrainType:             z.MinimapTerrainType,
		ScatterSetting:                 scatterSetting,
		CameraEnvironmentEffects:       cameraEnvironmentEffects,
		WindNoiseIds:                   z.WindNoiseIds,
		WaterMaterial:                  z.WaterMaterial,
		WaterLevelOffset:               z.WaterLevelOffset,
		WaterHeight:                    z.WaterHeight,
		WaterZoneOpacityCutoff:         z.WaterZoneOpacityCutoff,
		MaxWaterDepth:                  z.MaxWaterDepth,
		HeightModificationCurve:        heightModificationCurve,
		ReverbZone:                     reverbZone,
		AmbienceSoundId:                z.AmbienceSoundId,
		QuakeEffect:                    z.QuakeEffect,
		ZoneMusicType:                  z.ZoneMusicType,
		BanterEventType:                z.BanterEventType,
		PlayerKillHeight:               z.PlayerKillHeight,
		MinimapScatter:                 *minimapScatter,
		LocationHeightOffset:           z.LocationHeightOffset,
		UnkString:                      unkString,
		UnkThinHash1:                   z.UnkThinHash1,
		UnkThinHash2:                   z.UnkThinHash2,
		UnkBitfield1:                   z.UnkBitfield1,
		PatrolSetting:                  z.PatrolSetting,
		PatrolFloat1:                   z.PatrolFloat1,
		PatrolFloat2:                   z.PatrolFloat2,
		PatrolFloat3:                   z.PatrolFloat3,
		PatrolFloat4:                   z.PatrolFloat4,
		PatrolFloat5:                   z.PatrolFloat5,
		PatrolFloat6:                   z.PatrolFloat6,
		PatrolFloat7:                   z.PatrolFloat7,
		PatrolFloat8:                   z.PatrolFloat8,
		PatrolFloat9:                   z.PatrolFloat9,
		PatrolFloat10:                  z.PatrolFloat10,
		PatrolFloat11:                  z.PatrolFloat11,
		UnkBool1:                       z.UnkBool1 != 0,
		UnkBitfield2:                   z.UnkBitfield2,
		CityPlacementHeuristic:         z.CityPlacementHeuristic,
		UnknownEnum:                    z.UnknownEnum,
		UnkInt1:                        z.UnkInt1,
		UnkInt2:                        z.UnkInt2,
		Vectors1:                       z.Vectors1,
		Vectors2:                       z.Vectors2,
		Vectors3:                       z.Vectors3,
		UnkFloat1:                      z.UnkFloat1,
		UnkFloat2:                      z.UnkFloat2,
		UnkFloat3:                      z.UnkFloat3,
		UnkFloat4:                      z.UnkFloat4,
		LocationFlags:                  z.LocationFlags[:],
		ColoniesCurbPath:               z.ColoniesCurbPath,
		CurvedCurbPath:                 z.CurvedCurbPath,
		UnkFLoat5:                      z.UnkFLoat5,
		UnkFLoat6:                      z.UnkFLoat6,
		UnkInt3:                        z.UnkInt3,
		DetailUnit1:                    z.DetailUnit1,
		DetailUnit2:                    z.DetailUnit2,
		DetailUnit3:                    z.DetailUnit3,
		TrafficlightPrefab:             z.TrafficlightPrefab,
		UnkFloat7:                      z.UnkFloat7,
		UnkFloat8:                      z.UnkFloat8,
		UnkFloat9:                      z.UnkFloat9,
		UnkFloat10:                     z.UnkFloat10,
		UnkFloat11:                     z.UnkFloat11,
		CityStreetlights:               z.CityStreetlights,
		UnkFloat12:                     z.UnkFloat12,
		UnkFloat13:                     z.UnkFloat13,
		UnkFloat14:                     z.UnkFloat14,
		DetailUnit4:                    z.DetailUnit4,
		UnkFloat15:                     z.UnkFloat15,
		UnkFloat16:                     z.UnkFloat16,
		UnkFloat17:                     z.UnkFloat17,
		UnkFloat18:                     z.UnkFloat18,
		UnkFloat19:                     z.UnkFloat19,
		UnkFloat20:                     z.UnkFloat20,
		UnkInt4:                        z.UnkInt4,
		UnkString2:                     unkString2,
		UnkFloat21:                     z.UnkFloat21,
		UnkFloat22:                     z.UnkFloat22,
		BoundaryWallSettings:           *boundaryWallSettings,
		RoadEmbankmentUnits:            roadEmbankmentUnits,
		UnkInt5:                        z.UnkInt5,
		UnkBitfield3:                   z.UnkBitfield3,
	}, nil
}

var zones []ZoneSettings

func LoadZoneSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]ZoneSettings, error) {
	if zones != nil {
		// memoize this since we might look this up many times when exporting terrain
		return zones, nil
	}
	r := bytes.NewReader(zoneSettings)

	infos := make([]ZoneSettings, 0)
	padding := make([]byte, 4)
	if err := binary.Read(r, binary.LittleEndian, padding); err != nil {
		return nil, fmt.Errorf("reading padding: %v", err)
	}
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("ZoneSettings") {
			return nil, fmt.Errorf("invalid zone settings file: type is %v at index %v", header.Type.String(), i)
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding zone settings base: %v", err)
		}

		var rawSetting rawZoneSettings
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading zone settings: %v", err)
		}

		setting, err := rawSetting.Deserialize(r, base)
		if err != nil {
			return nil, fmt.Errorf("dereferencing zone setting arrays/ptrs: %v", err)
		}

		_, err = r.Seek(base+int64(header.Size), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking next zone settings: %v", err)
		}

		infos = append(infos, *setting)
	}
	zones = infos

	return infos, nil
}
