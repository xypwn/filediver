package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type DangerWarningComponent struct {
	WarningType   enum.DangerWarningType // Warning enum
	TriggerRadius float32                // If the player is within this radius that will show the warning. If less than 0 then always active
	StartActive   uint8                  // [bool]If true then this warning is active as soon as this entity exists
	_             [3]uint8
}

type SimpleDangerWarningComponent struct {
	WarningType   enum.DangerWarningType `json:"warning_type"`
	TriggerRadius float32                `json:"trigger_radius"`
	StartActive   bool                   `json:"start_active"`
}

func (w DangerWarningComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleDangerWarningComponent{
		WarningType:   w.WarningType,
		TriggerRadius: w.TriggerRadius,
		StartActive:   w.StartActive != 0,
	}
}

func getDangerWarningComponentData() ([]byte, error) {
	dangerWarningComponentHash := Sum("DangerWarningComponentData")
	dangerWarningComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(dangerWarningComponentHashData, binary.LittleEndian, dangerWarningComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, dangerWarningComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getDangerWarningComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("DangerWarningComponentData")
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
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("DangerWarningComponent") {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (data type was not DangerWarningComponent)")
	}

	dangerWarningComponentData, err := getDangerWarningComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get danger warning component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(dangerWarningComponentData)

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
		return nil, fmt.Errorf("%v not found in danger warning component data", hash.String())
	}

	var dangerWarningComponentType DLTypeDesc
	dangerWarningComponentType, ok = typelib.Types[Sum("DangerWarningComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find DangerWarningComponent hash in dl_library")
	}

	componentData := make([]byte, dangerWarningComponentType.Size)
	if _, err := r.Seek(int64(dangerWarningComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseDangerWarningComponents() (map[stingray.Hash]DangerWarningComponent, error) {
	unitHash := Sum("DangerWarningComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var dangerWarningType DLTypeDesc
	var ok bool
	dangerWarningType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find DangerWarningComponentData hash in dl_library")
	}

	if len(dangerWarningType.Members) != 2 {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (there should be 2 members but were actually %v)", len(dangerWarningType.Members))
	}

	if dangerWarningType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (hashmap atom was not inline array)")
	}

	if dangerWarningType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (data atom was not inline array)")
	}

	if dangerWarningType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (hashmap storage was not struct)")
	}

	if dangerWarningType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (data storage was not struct)")
	}

	if dangerWarningType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if dangerWarningType.Members[1].TypeID != Sum("DangerWarningComponent") {
		return nil, fmt.Errorf("DangerWarningComponentData unexpected format (data type was not DangerWarningComponent)")
	}

	dangerWarningComponentData, err := getDangerWarningComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get danger warning component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(dangerWarningComponentData)

	hashmap := make([]ComponentIndexData, dangerWarningType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]DangerWarningComponent, dangerWarningType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]DangerWarningComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
