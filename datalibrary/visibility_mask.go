package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

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

type SimpleVisibilityMaskInfo struct {
	Name        string `json:"name"`
	Index       uint16 `json:"index"`
	StartHidden bool   `json:"default_hidden"`
}

type SimpleVisibilityRandomization struct {
	Identifier     string   `json:"id"`
	MaskIndexNames []string `json:"mask_index_names,omitempty"`
}

type SimpleVisibilityMaskComponent struct {
	MaskInfos      []SimpleVisibilityMaskInfo      `json:"mask_infos,omitempty"`
	Randomizations []SimpleVisibilityRandomization `json:"randomizations,omitempty"`
}

func (component VisibilityMaskComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupString StringsLookup) any {
	simpleCmp := SimpleVisibilityMaskComponent{
		MaskInfos:      make([]SimpleVisibilityMaskInfo, 0),
		Randomizations: make([]SimpleVisibilityRandomization, 0),
	}
	for _, info := range component.MaskInfos {
		if info.Name.Value == 0 {
			break
		}
		simpleCmp.MaskInfos = append(simpleCmp.MaskInfos, SimpleVisibilityMaskInfo{
			Name:        lookupThinHash(info.Name),
			Index:       info.Index,
			StartHidden: info.StartHidden != 0,
		})
	}
	for _, rand := range component.Randomizations {
		if rand.Identifier.Value == 0 {
			break
		}
		maskIndexNames := make([]string, 0)
		for _, maskIndexName := range rand.MaskIndexNames {
			if maskIndexName.Value == 0 {
				break
			}
			maskIndexNames = append(maskIndexNames, lookupThinHash(maskIndexName))
		}
		simpleCmp.Randomizations = append(simpleCmp.Randomizations, SimpleVisibilityRandomization{
			Identifier:     lookupThinHash(rand.Identifier),
			MaskIndexNames: maskIndexNames,
		})
	}
	return simpleCmp
}

func getVisibilityMaskComponentData() ([]byte, error) {
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

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getVisibilityMaskComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	visibilityMaskHash := Sum("VisibilityMaskComponentData")
	typelib, err := ParseTypeLib(nil)
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

	if visibilityMaskType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap atom was not inline array)")
	}

	if visibilityMaskType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (data atom was not inline array)")
	}

	if visibilityMaskType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap storage was not struct)")
	}

	if visibilityMaskType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (data storage was not struct)")
	}

	if visibilityMaskType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if visibilityMaskType.Members[1].TypeID != Sum("VisibilityMaskComponent") {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (data type was not VisibilityMaskComponent)")
	}

	visibilityMaskComponentData, err := getVisibilityMaskComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get visibility mask component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(visibilityMaskComponentData)

	hashmap := make([]ComponentIndexData, visibilityMaskType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in visibility mask component data", hash.String())
	}

	var visibilityMaskComponentType DLTypeDesc
	visibilityMaskComponentType, ok = typelib.Types[Sum("VisibilityMaskComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find VisibilityMaskComponent hash in dl_library")
	}

	componentData := make([]byte, visibilityMaskComponentType.Size)
	if _, err := r.Seek(int64(visibilityMaskComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
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

	typelib, err := ParseTypeLib(nil)
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

	if visibilityMaskType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap atom was not inline array)")
	}

	if visibilityMaskType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (data atom was not inline array)")
	}

	if visibilityMaskType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("VisibilityMaskComponentData unexpected format (hashmap storage was not struct)")
	}

	if visibilityMaskType.Members[1].Type.Storage != STRUCT {
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
