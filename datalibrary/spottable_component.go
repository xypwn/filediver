package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type HudMarkerTypeOverride struct {
	Node                stingray.ThinHash  // [string]What node to look for on the unit.
	MaxHitPosDistance   float32            // Max distance between raycast hit and node position.
	MaxHitAngDifference float32            // Max angle difference (2D) between raycast dir and node forward.
	MaxFuzzyDistance    float32            // Max fuzzy distance for the override, a combination between distance and FOV.
	MarkerType          enum.HudMarkerType // Maximum amount rewarded per pick-up. [sic]
}

type SpottableComponent struct {
	VoLine             stingray.ThinHash  // [string]The VO line to play when spotting this entity.
	Radius             mgl32.Vec2         // The radius of the sphere we spot against, at min and max range.
	Range              mgl32.Vec2         // The range in which spotting this thing is possible. Also used to scale radius.
	MarkerType         enum.HudMarkerType // In world hud marker to place when spotting this entity.
	FindableObjective  uint8              // [bool]If set, when spotted this entity will report to the objective component which can be used to complete
	_                  [3]uint8
	SpottingNode       stingray.ThinHash // [string]Node used for positioning the spotting marker and detection bounds
	StartActive        uint8             // [bool]Whether this thing starts spottable
	_                  [3]uint8
	MarkerTypeOverride HudMarkerTypeOverride // An override setting under certain conditions
	MarkerIcon         stingray.Hash         // [material]Icon displayed on HUD when spottable is 'pinged'
	MarkerTextureType  enum.TextureType      // Type of texture used by the Marker Icon
	_                  [4]uint8
}

type SimpleHudMarkerTypeOverride struct {
	Node                string             `json:"node"`
	MaxHitPosDistance   float32            `json:"max_hit_pos_distance"`
	MaxHitAngDifference float32            `json:"max_hit_ang_difference"`
	MaxFuzzyDistance    float32            `json:"max_fuzzy_distance"`
	MarkerType          enum.HudMarkerType `json:"marker_type"`
}

type SimpleSpottableComponent struct {
	VoLine             string                      `json:"vo_line"`
	Radius             mgl32.Vec2                  `json:"radius"`
	Range              mgl32.Vec2                  `json:"range"`
	MarkerType         enum.HudMarkerType          `json:"marker_type"`
	FindableObjective  bool                        `json:"findable_objective"`
	SpottingNode       string                      `json:"spotting_node"`
	StartActive        bool                        `json:"start_active"`
	MarkerTypeOverride SimpleHudMarkerTypeOverride `json:"marker_type_override"`
	MarkerIcon         string                      `json:"marker_icon"`
	MarkerTextureType  enum.TextureType            `json:"marker_texture_type"`
}

func (w SpottableComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleSpottableComponent{
		VoLine:            lookupThinHash(w.VoLine),
		Radius:            w.Radius,
		Range:             w.Range,
		MarkerType:        w.MarkerType,
		FindableObjective: w.FindableObjective != 0,
		SpottingNode:      lookupThinHash(w.SpottingNode),
		StartActive:       w.StartActive != 0,
		MarkerTypeOverride: SimpleHudMarkerTypeOverride{
			Node:                lookupThinHash(w.MarkerTypeOverride.Node),
			MaxHitPosDistance:   w.MarkerTypeOverride.MaxHitPosDistance,
			MaxHitAngDifference: w.MarkerTypeOverride.MaxHitAngDifference,
			MaxFuzzyDistance:    w.MarkerTypeOverride.MaxFuzzyDistance,
			MarkerType:          w.MarkerTypeOverride.MarkerType,
		},
		MarkerIcon:        lookupHash(w.MarkerIcon),
		MarkerTextureType: w.MarkerTextureType,
	}
}

func getSpottableComponentData() ([]byte, error) {
	spottableComponentHash := Sum("SpottableComponentData")
	spottableComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(spottableComponentHashData, binary.LittleEndian, spottableComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, spottableComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getSpottableComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("SpottableComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var unitCmpDataType DLTypeDesc
	var ok bool
	unitCmpDataType, ok = typelib.Types[UnitCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(unitCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("SpottableComponent") {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (data type was not SpottableComponent)")
	}

	spottableComponentData, err := getSpottableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get spottable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(spottableComponentData)

	hashmap := make([]ComponentIndexData, unitCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	var index int32 = -1
	for _, entry := range hashmap {
		if entry.Resource == hash {
			index = int32(entry.Index)
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("%v not found in spottable component data", hash.String())
	}

	var spottableComponentType DLTypeDesc
	spottableComponentType, ok = typelib.Types[Sum("SpottableComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find SpottableComponent hash in dl_library")
	}

	componentData := make([]byte, spottableComponentType.Size)
	if _, err := r.Seek(int64(spottableComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseSpottableComponents() (map[stingray.Hash]SpottableComponent, error) {
	unitHash := Sum("SpottableComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var spottableType DLTypeDesc
	var ok bool
	spottableType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find SpottableComponentData hash in dl_library")
	}

	if len(spottableType.Members) != 2 {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (there should be 2 members but were actually %v)", len(spottableType.Members))
	}

	if spottableType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if spottableType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (data atom was not inline array)")
	}

	if spottableType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (hashmap storage was not struct)")
	}

	if spottableType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (data storage was not struct)")
	}

	if spottableType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if spottableType.Members[1].TypeID != Sum("SpottableComponent") {
		return nil, fmt.Errorf("SpottableComponentData unexpected format (data type was not SpottableComponent)")
	}

	spottableComponentData, err := getSpottableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get spottable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(spottableComponentData)

	hashmap := make([]ComponentIndexData, spottableType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]SpottableComponent, spottableType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]SpottableComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
