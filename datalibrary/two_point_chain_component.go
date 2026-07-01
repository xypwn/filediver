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

type TwoPointChainComponent struct {
	ChainUnitPath           stingray.Hash // name length 15
	UnknownBool1            uint8
	_                       [3]uint8
	SourceChainAttachNode   stingray.ThinHash  // name length 24
	UnknownAnimationEvent1  stingray.ThinHash  // name length 26
	UnknownAnimationEvent2  stingray.ThinHash  // name length 32
	UnknownAnimationEvent3  stingray.ThinHash  // name length 34
	UnknownAnimationEvent4  stingray.ThinHash  // name length 33
	UnknownAnimationEvent5  stingray.ThinHash  // name length 28
	UnknownAnimationEvent6  stingray.ThinHash  // name length 34
	UnknownAnimationEvent7  stingray.ThinHash  // name length 35
	UnknownAnimationEvent8  stingray.ThinHash  // name length 23
	UnknownAnimationEvent9  stingray.ThinHash  // name length 29
	UnknownAnimationEvent10 stingray.ThinHash  // name length 20
	UnknownAnimationEvent11 stingray.ThinHash  // name length 26
	UnknownAnimationEvent12 stingray.ThinHash  // name length 21
	UnknownAnimationEvent13 stingray.ThinHash  // name length 27
	UnknownAnimationEvent14 stingray.ThinHash  // name length 27
	UnknownAnimationEvent15 stingray.ThinHash  // name length 33
	InventorySlot           enum.InventorySlot // name length 28
	ChainEndNode            stingray.ThinHash  // name length 14
	WeaponAttachNode        stingray.ThinHash  // name length 18
	Unknown3dVector1        mgl32.Vec3         // name length 34
	UnknownFloat1           float32            // name length 10
	UnknownFloat2           float32            // name length 18
	UnknownFloat3           float32            // name length 24
	UnknownThinHash19       stingray.ThinHash  // name length 30
	Unknown3dVector2        mgl32.Vec3         // name length 46
	UnknownFloat4           float32            // name length 24
}

type SimpleTwoPointChainComponent struct {
	ChainUnitPath           string             `json:"chain_unit_path"`
	UnknownBool1            bool               `json:"unknown_bool_1"`
	SourceChainAttachNode   string             `json:"source_chain_attach_node"`
	UnknownAnimationEvent1  string             `json:"unknown_animation_event_1"`
	UnknownAnimationEvent2  string             `json:"unknown_animation_event_2"`
	UnknownAnimationEvent3  string             `json:"unknown_animation_event_3"`
	UnknownAnimationEvent4  string             `json:"unknown_animation_event_4"`
	UnknownAnimationEvent5  string             `json:"unknown_animation_event_5"`
	UnknownAnimationEvent6  string             `json:"unknown_animation_event_6"`
	UnknownAnimationEvent7  string             `json:"unknown_animation_event_7"`
	UnknownAnimationEvent8  string             `json:"unknown_animation_event_8"`
	UnknownAnimationEvent9  string             `json:"unknown_animation_event_9"`
	UnknownAnimationEvent10 string             `json:"unknown_animation_event_10"`
	UnknownAnimationEvent11 string             `json:"unknown_animation_event_11"`
	UnknownAnimationEvent12 string             `json:"unknown_animation_event_12"`
	UnknownAnimationEvent13 string             `json:"unknown_animation_event_13"`
	UnknownAnimationEvent14 string             `json:"unknown_animation_event_14"`
	UnknownAnimationEvent15 string             `json:"unknown_animation_event_15"`
	InventorySlot           enum.InventorySlot `json:"inventory_slot"`
	ChainEndNode            string             `json:"chain_end_node"`
	WeaponAttachNode        string             `json:"weapon_attach_node"`
	Unknown3dVector1        mgl32.Vec3         `json:"unknown_3d_vector_1"`
	UnknownFloat1           float32            `json:"unknown_float_1"`
	UnknownFloat2           float32            `json:"unknown_float_2"`
	UnknownFloat3           float32            `json:"unknown_float_3"`
	UnknownThinHash19       string             `json:"unknown_thin_hash_19"`
	Unknown3dVector2        mgl32.Vec3         `json:"unknown_3d_vector_2"`
	UnknownFloat4           float32            `json:"unknown_float_4"`
}

func (m TwoPointChainComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleTwoPointChainComponent{
		ChainUnitPath:           lookupHash(m.ChainUnitPath),
		UnknownBool1:            m.UnknownBool1 != 0,
		SourceChainAttachNode:   lookupThinHash(m.SourceChainAttachNode),
		UnknownAnimationEvent1:  lookupThinHash(m.UnknownAnimationEvent1),
		UnknownAnimationEvent2:  lookupThinHash(m.UnknownAnimationEvent2),
		UnknownAnimationEvent3:  lookupThinHash(m.UnknownAnimationEvent3),
		UnknownAnimationEvent4:  lookupThinHash(m.UnknownAnimationEvent4),
		UnknownAnimationEvent5:  lookupThinHash(m.UnknownAnimationEvent5),
		UnknownAnimationEvent6:  lookupThinHash(m.UnknownAnimationEvent6),
		UnknownAnimationEvent7:  lookupThinHash(m.UnknownAnimationEvent7),
		UnknownAnimationEvent8:  lookupThinHash(m.UnknownAnimationEvent8),
		UnknownAnimationEvent9:  lookupThinHash(m.UnknownAnimationEvent9),
		UnknownAnimationEvent10: lookupThinHash(m.UnknownAnimationEvent10),
		UnknownAnimationEvent11: lookupThinHash(m.UnknownAnimationEvent11),
		UnknownAnimationEvent12: lookupThinHash(m.UnknownAnimationEvent12),
		UnknownAnimationEvent13: lookupThinHash(m.UnknownAnimationEvent13),
		UnknownAnimationEvent14: lookupThinHash(m.UnknownAnimationEvent14),
		UnknownAnimationEvent15: lookupThinHash(m.UnknownAnimationEvent15),
		InventorySlot:           m.InventorySlot,
		ChainEndNode:            lookupThinHash(m.ChainEndNode),
		WeaponAttachNode:        lookupThinHash(m.WeaponAttachNode),
		Unknown3dVector1:        m.Unknown3dVector1,
		UnknownFloat1:           m.UnknownFloat1,
		UnknownFloat2:           m.UnknownFloat2,
		UnknownFloat3:           m.UnknownFloat3,
		UnknownThinHash19:       lookupThinHash(m.UnknownThinHash19),
		Unknown3dVector2:        m.Unknown3dVector2,
		UnknownFloat4:           m.UnknownFloat4,
	}
}

func getTwoPointChainComponentData() ([]byte, error) {
	twoPointChainHash := Sum("TwoPointChainComponentData")
	twoPointChainHashData := make([]byte, 4)
	if _, err := binary.Encode(twoPointChainHashData, binary.LittleEndian, twoPointChainHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, twoPointChainHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getTwoPointChainComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	TwoPointChainCmpDataHash := Sum("TwoPointChainComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var twoPointChainCmpDataType DLTypeDesc
	var ok bool
	twoPointChainCmpDataType, ok = typelib.Types[TwoPointChainCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find TwoPointChainComponentData hash in dl_library")
	}

	if len(twoPointChainCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (there should be 2 members but were actually %v)", len(twoPointChainCmpDataType.Members))
	}

	if twoPointChainCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (hashmap atom was not inline array)")
	}

	if twoPointChainCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (data atom was not inline array)")
	}

	if twoPointChainCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (hashmap storage was not struct)")
	}

	if twoPointChainCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (data storage was not struct)")
	}

	if twoPointChainCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if twoPointChainCmpDataType.Members[1].TypeID != Sum("TwoPointChainComponent") {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (data type was not TwoPointChainComponent)")
	}

	twoPointChainComponentData, err := getTwoPointChainComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get two point chain component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(twoPointChainComponentData)

	hashmap := make([]ComponentIndexData, twoPointChainCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in two point chain component data", hash.String())
	}

	var twoPointChainComponentType DLTypeDesc
	twoPointChainComponentType, ok = typelib.Types[Sum("TwoPointChainComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find TwoPointChainComponent hash in dl_library")
	}

	componentData := make([]byte, twoPointChainComponentType.Size)
	if _, err := r.Seek(int64(twoPointChainComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseTwoPointChainComponents() (map[stingray.Hash]TwoPointChainComponent, error) {
	twoPointChainHash := Sum("TwoPointChainComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var twoPointChainType DLTypeDesc
	var ok bool
	twoPointChainType, ok = typelib.Types[twoPointChainHash]
	if !ok {
		return nil, fmt.Errorf("could not find TwoPointChainComponentData hash in dl_library")
	}

	if len(twoPointChainType.Members) != 2 {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (there should be 2 members but were actually %v)", len(twoPointChainType.Members))
	}

	if twoPointChainType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (hashmap atom was not inline array)")
	}

	if twoPointChainType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (data atom was not inline array)")
	}

	if twoPointChainType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (hashmap storage was not struct)")
	}

	if twoPointChainType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (data storage was not struct)")
	}

	if twoPointChainType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if twoPointChainType.Members[1].TypeID != Sum("TwoPointChainComponent") {
		return nil, fmt.Errorf("TwoPointChainComponentData unexpected format (data type was not TwoPointChainComponent)")
	}

	twoPointChainComponentData, err := getTwoPointChainComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get two point chains component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(twoPointChainComponentData)

	hashmap := make([]ComponentIndexData, twoPointChainType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]TwoPointChainComponent, twoPointChainType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]TwoPointChainComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
