package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/xypwn/filediver/stingray"
)

type ComponentIndexData struct {
	Resource stingray.Hash
	Index    uint32
	_        [4]uint8
}

type VisibilityMaskInfo struct {
	Name        stingray.ThinHash
	Index       uint16
	StartHidden uint8
	_           uint8
}

type VisibilityRandomization struct {
	Identifier     stingray.ThinHash
	MaskIndexNames [7]stingray.ThinHash
}

type VisibilityMaskComponent struct {
	MaskInfos      [64]VisibilityMaskInfo
	Randomizations [4]VisibilityRandomization
}

func (v *VisibilityMaskComponent) Length() int {
	i := 0
	for _, maskInfo := range v.MaskInfos {
		if maskInfo.Name.Value == 0 {
			return i
		}
		i = i + 1
	}
	return i
}

type DLInstanceHeader struct {
	_       DLHash
	Magic   [4]byte
	Version uint32
	Type    DLHash
	Size    uint32
	Is64Bit uint8
	_       [7]uint8
}

func ParseVisibilityMasks() (map[stingray.Hash]VisibilityMaskComponent, error) {
	visibilityMaskHash := Sum("VisibilityMaskComponentData")
	visibilityMaskHashData := make([]byte, 4)
	if _, err := binary.Encode(visibilityMaskHashData, binary.LittleEndian, visibilityMaskHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, visibilityMaskHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	typelib, err := ParseTypeLib()
	if err != nil {
		return nil, err
	}

	var visibilityMaskType DLTypeDesc
	var ok bool
	visibilityMaskType, ok = typelib.Types[visibilityMaskHash]
	if !ok {
		return nil, fmt.Errorf("could not find VisibilityMaskComponentData hash in dl_library")
	}

	if len(visibilityMaskType.Members) != 2 {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (there should be 2 members but were actually %v)", len(visibilityMaskType.Members))
	}

	if visibilityMaskType.Members[0].Type.Atom != DL_TYPE_ATOM_INLINE_ARRAY {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap atom was not inline array)")
	}

	if visibilityMaskType.Members[1].Type.Atom != DL_TYPE_ATOM_INLINE_ARRAY {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (data atom was not inline array)")
	}

	if visibilityMaskType.Members[0].Type.Storage != DL_TYPE_STORAGE_STRUCT {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap storage was not struct)")
	}

	if visibilityMaskType.Members[1].Type.Storage != DL_TYPE_STORAGE_STRUCT {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (data storage was not struct)")
	}

	if visibilityMaskType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if visibilityMaskType.Members[1].TypeID != Sum("VisibilityMaskComponent") {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (data type was not VisibilityMaskComponent)")
	}

	hashmap := make([]ComponentIndexData, visibilityMaskType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]VisibilityMaskComponent, visibilityMaskType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]VisibilityMaskComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
