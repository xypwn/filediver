package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type TwoPointChainAttachTargetComponent struct {
	ChainAttachTarget enum.Tag `json:"chain_attach_target"`
}

func (m TwoPointChainAttachTargetComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return m
}

func getTwoPointChainAttachTargetComponentData() ([]byte, error) {
	twoPointChainAttachTargetHash := Sum("TwoPointChainAttachTargetComponentData")
	twoPointChainAttachTargetHashData := make([]byte, 4)
	if _, err := binary.Encode(twoPointChainAttachTargetHashData, binary.LittleEndian, twoPointChainAttachTargetHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, twoPointChainAttachTargetHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getTwoPointChainAttachTargetComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	TwoPointChainAttachTargetCmpDataHash := Sum("TwoPointChainAttachTargetComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var twoPointChainAttachTargetCmpDataType DLTypeDesc
	var ok bool
	twoPointChainAttachTargetCmpDataType, ok = typelib.Types[TwoPointChainAttachTargetCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find TwoPointChainAttachTargetComponentData hash in dl_library")
	}

	if len(twoPointChainAttachTargetCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (there should be 2 members but were actually %v)", len(twoPointChainAttachTargetCmpDataType.Members))
	}

	if twoPointChainAttachTargetCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (hashmap atom was not inline array)")
	}

	if twoPointChainAttachTargetCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (data atom was not inline array)")
	}

	if twoPointChainAttachTargetCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (hashmap storage was not struct)")
	}

	if twoPointChainAttachTargetCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (data storage was not struct)")
	}

	if twoPointChainAttachTargetCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if twoPointChainAttachTargetCmpDataType.Members[1].TypeID != Sum("TwoPointChainAttachTargetComponent") {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (data type was not TwoPointChainAttachTargetComponent)")
	}

	twoPointChainAttachTargetComponentData, err := getTwoPointChainAttachTargetComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get two point chain attach target component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(twoPointChainAttachTargetComponentData)

	hashmap := make([]ComponentIndexData, twoPointChainAttachTargetCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in two point chain attach target component data", hash.String())
	}

	var twoPointChainAttachTargetComponentType DLTypeDesc
	twoPointChainAttachTargetComponentType, ok = typelib.Types[Sum("TwoPointChainAttachTargetComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find TwoPointChainAttachTargetComponent hash in dl_library")
	}

	componentData := make([]byte, twoPointChainAttachTargetComponentType.Size)
	if _, err := r.Seek(int64(twoPointChainAttachTargetComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseTwoPointChainAttachTargetComponents() (map[stingray.Hash]TwoPointChainAttachTargetComponent, error) {
	twoPointChainAttachTargetHash := Sum("TwoPointChainAttachTargetComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var twoPointChainAttachTargetType DLTypeDesc
	var ok bool
	twoPointChainAttachTargetType, ok = typelib.Types[twoPointChainAttachTargetHash]
	if !ok {
		return nil, fmt.Errorf("could not find TwoPointChainAttachTargetComponentData hash in dl_library")
	}

	if len(twoPointChainAttachTargetType.Members) != 2 {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (there should be 2 members but were actually %v)", len(twoPointChainAttachTargetType.Members))
	}

	if twoPointChainAttachTargetType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (hashmap atom was not inline array)")
	}

	if twoPointChainAttachTargetType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (data atom was not inline array)")
	}

	if twoPointChainAttachTargetType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (hashmap storage was not struct)")
	}

	if twoPointChainAttachTargetType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (data storage was not struct)")
	}

	if twoPointChainAttachTargetType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if twoPointChainAttachTargetType.Members[1].TypeID != Sum("TwoPointChainAttachTargetComponent") {
		return nil, fmt.Errorf("TwoPointChainAttachTargetComponentData unexpected format (data type was not TwoPointChainAttachTargetComponent)")
	}

	twoPointChainAttachTargetComponentData, err := getTwoPointChainAttachTargetComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get two point chain attach targets component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(twoPointChainAttachTargetComponentData)

	hashmap := make([]ComponentIndexData, twoPointChainAttachTargetType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]TwoPointChainAttachTargetComponent, twoPointChainAttachTargetType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]TwoPointChainAttachTargetComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
