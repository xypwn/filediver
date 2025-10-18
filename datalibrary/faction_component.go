package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type FactionComponent struct {
	Factions              [3]enum.FactionType   // Which factions this entity belongs to.
	TargetingNode         stingray.ThinHash     // [string]Node used for line of sight/aiming at entity.
	UnknownHashes         [4]stingray.Hash      // name length 28
	HalfWidth             float32               // Distance from which an enemy can attack you from the side.
	HalfForwardSize       float32               // Distance from which an enemy can attack you from the front or back.
	Priority              float32               // Used when selecting target
	UnknownFloat1         float32               // name length 22
	UnknownFloat2         float32               // name length 15
	ComplexTargetingNodes [16]stingray.ThinHash // [string]Multiple nodes specified for larger targets, such as vehicles
	UnknownBool           uint8                 // name length 17
	_                     [3]uint8
}

type SimpleFactionComponent struct {
	Factions              []enum.FactionType `json:"factions"`
	TargetingNode         string             `json:"targeting_node"`
	UnknownHashes         []string           `json:"unknown_hashes"`
	HalfWidth             float32            `json:"half_width"`
	HalfForwardSize       float32            `json:"half_forward_size"`
	Priority              float32            `json:"priority"`
	UnknownFloat1         float32            `json:"unknown_float_1"`
	UnknownFloat2         float32            `json:"unknown_float_2"`
	ComplexTargetingNodes []string           `json:"complex_targeting_nodes"`
	UnknownBool           bool               `json:"unknown_bool"`
}

func (w FactionComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	factions := make([]enum.FactionType, 0)
	for _, faction := range w.Factions {
		if faction == enum.FactionType_None {
			break
		}
		factions = append(factions, faction)
	}

	unknownHashes := make([]string, 0)
	for _, hash := range w.UnknownHashes {
		if hash.Value == 0 {
			break
		}
		unknownHashes = append(unknownHashes, lookupHash(hash))
	}

	complexTargetingNodes := make([]string, 0)
	for _, hash := range w.ComplexTargetingNodes {
		if hash.Value == 0 {
			break
		}
		complexTargetingNodes = append(complexTargetingNodes, lookupThinHash(hash))
	}

	return SimpleFactionComponent{
		Factions:              factions,
		TargetingNode:         lookupThinHash(w.TargetingNode),
		UnknownHashes:         unknownHashes,
		HalfWidth:             w.HalfWidth,
		HalfForwardSize:       w.HalfForwardSize,
		Priority:              w.Priority,
		UnknownFloat1:         w.UnknownFloat1,
		UnknownFloat2:         w.UnknownFloat2,
		ComplexTargetingNodes: complexTargetingNodes,
		UnknownBool:           w.UnknownBool != 0,
	}
}

func getFactionComponentData() ([]byte, error) {
	factionComponentHash := Sum("FactionComponentData")
	factionComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(factionComponentHashData, binary.LittleEndian, factionComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, factionComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getFactionComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("FactionComponentData")
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
		return nil, fmt.Errorf("FactionComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("FactionComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("FactionComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("FactionComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("FactionComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("FactionComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("FactionComponent") {
		return nil, fmt.Errorf("FactionComponentData unexpected format (data type was not FactionComponent)")
	}

	factionComponentData, err := getFactionComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get faction component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(factionComponentData)

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
		return nil, fmt.Errorf("%v not found in faction component data", hash.String())
	}

	var factionComponentType DLTypeDesc
	factionComponentType, ok = typelib.Types[Sum("FactionComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find FactionComponent hash in dl_library")
	}

	componentData := make([]byte, factionComponentType.Size)
	if _, err := r.Seek(int64(factionComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseFactionComponents() (map[stingray.Hash]FactionComponent, error) {
	unitHash := Sum("FactionComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var factionType DLTypeDesc
	var ok bool
	factionType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find FactionComponentData hash in dl_library")
	}

	if len(factionType.Members) != 2 {
		return nil, fmt.Errorf("FactionComponentData unexpected format (there should be 2 members but were actually %v)", len(factionType.Members))
	}

	if factionType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("FactionComponentData unexpected format (hashmap atom was not inline array)")
	}

	if factionType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("FactionComponentData unexpected format (data atom was not inline array)")
	}

	if factionType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("FactionComponentData unexpected format (hashmap storage was not struct)")
	}

	if factionType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("FactionComponentData unexpected format (data storage was not struct)")
	}

	if factionType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("FactionComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if factionType.Members[1].TypeID != Sum("FactionComponent") {
		return nil, fmt.Errorf("FactionComponentData unexpected format (data type was not FactionComponent)")
	}

	factionComponentData, err := getFactionComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get faction component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(factionComponentData)

	hashmap := make([]ComponentIndexData, factionType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]FactionComponent, factionType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]FactionComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
