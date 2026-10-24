package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type WoundableComponent struct {
	WoundLUT  enum.WoundLUT `json:"wount_lut"`  // The object space lookup texture to use for this entity when rendering wounds onto it.
	MinRadius float32       `json:"min_radius"` // The minimum radius of a bullet wound. This is when damage is 0.
	MaxRadius float32       `json:"max_radius"` // The maximum radius of a bullet wound. This is when damage is equal or greater than 'max_damage'.
	MaxDamage float32       `json:"max_damage"` // The damage cap where the radius won't get larger.
}

func (w WoundableComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return w
}

func getWoundableComponentData() ([]byte, error) {
	woundableComponentHash := Sum("WoundableComponentData")
	woundableComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(woundableComponentHashData, binary.LittleEndian, woundableComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, woundableComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWoundableComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	woundableCmpDataHash := Sum("WoundableComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var unitCmpDataType DLTypeDesc
	var ok bool
	unitCmpDataType, ok = typelib.Types[woundableCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find WoundableComponentData hash in dl_library")
	}

	if len(unitCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("WoundableComponent") {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (data type was not WoundableComponent)")
	}

	woundableComponentData, err := getWoundableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get behavior component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(woundableComponentData)

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
		return nil, fmt.Errorf("%v not found in behavior component data", hash.String())
	}

	var woundableComponentType DLTypeDesc
	woundableComponentType, ok = typelib.Types[Sum("WoundableComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WoundableComponent hash in dl_library")
	}

	componentData := make([]byte, woundableComponentType.Size)
	if _, err := r.Seek(int64(woundableComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWoundableComponents() (map[stingray.Hash]WoundableComponent, error) {
	unitHash := Sum("WoundableComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var behaviorType DLTypeDesc
	var ok bool
	behaviorType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find WoundableComponentData hash in dl_library")
	}

	if len(behaviorType.Members) != 2 {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (there should be 2 members but were actually %v)", len(behaviorType.Members))
	}

	if behaviorType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if behaviorType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (data atom was not inline array)")
	}

	if behaviorType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (hashmap storage was not struct)")
	}

	if behaviorType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (data storage was not struct)")
	}

	if behaviorType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if behaviorType.Members[1].TypeID != Sum("WoundableComponent") {
		return nil, fmt.Errorf("WoundableComponentData unexpected format (data type was not WoundableComponent)")
	}

	woundableComponentData, err := getWoundableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get behavior component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(woundableComponentData)

	hashmap := make([]ComponentIndexData, behaviorType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WoundableComponent, behaviorType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WoundableComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
