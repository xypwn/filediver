package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type TagComponent struct {
	Tags            [64]enum.Tag
	OnDestroyedTags [4]enum.Tag
}

type SimpleTagComponent struct {
	Tags            []enum.Tag `json:"tags,omitempty"`
	OnDestroyedTags []enum.Tag `json:"on_destroyed_tags,omitempty"`
}

func (w TagComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	tags := make([]enum.Tag, 0)
	for _, tag := range w.Tags {
		if tag == enum.Tag_None {
			break
		}
		tags = append(tags, tag)
	}
	onDestroyedTags := make([]enum.Tag, 0)
	for _, tag := range w.OnDestroyedTags {
		if tag == enum.Tag_None {
			break
		}
		onDestroyedTags = append(onDestroyedTags, tag)
	}
	return SimpleTagComponent{
		Tags:            tags,
		OnDestroyedTags: onDestroyedTags,
	}
}

func getTagComponentData() ([]byte, error) {
	tagComponentHash := Sum("TagComponentData")
	tagComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(tagComponentHashData, binary.LittleEndian, tagComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, tagComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getTagComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("TagComponentData")
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
		return nil, fmt.Errorf("TagComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TagComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TagComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TagComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TagComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TagComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("TagComponent") {
		return nil, fmt.Errorf("TagComponentData unexpected format (data type was not TagComponent)")
	}

	tagComponentData, err := getTagComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get tag component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(tagComponentData)

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
		return nil, fmt.Errorf("%v not found in tag component data", hash.String())
	}

	var tagComponentType DLTypeDesc
	tagComponentType, ok = typelib.Types[Sum("TagComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find TagComponent hash in dl_library")
	}

	componentData := make([]byte, tagComponentType.Size)
	if _, err := r.Seek(int64(tagComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseTagComponents() (map[stingray.Hash]TagComponent, error) {
	unitHash := Sum("TagComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var tagType DLTypeDesc
	var ok bool
	tagType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find TagComponentData hash in dl_library")
	}

	if len(tagType.Members) != 2 {
		return nil, fmt.Errorf("TagComponentData unexpected format (there should be 2 members but were actually %v)", len(tagType.Members))
	}

	if tagType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TagComponentData unexpected format (hashmap atom was not inline array)")
	}

	if tagType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TagComponentData unexpected format (data atom was not inline array)")
	}

	if tagType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TagComponentData unexpected format (hashmap storage was not struct)")
	}

	if tagType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TagComponentData unexpected format (data storage was not struct)")
	}

	if tagType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TagComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if tagType.Members[1].TypeID != Sum("TagComponent") {
		return nil, fmt.Errorf("TagComponentData unexpected format (data type was not TagComponent)")
	}

	tagComponentData, err := getTagComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get tag component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(tagComponentData)

	hashmap := make([]ComponentIndexData, tagType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]TagComponent, tagType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]TagComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
