package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type BehaviorComponent struct {
	BehaviorId   enum.BehaviorId `json:"behavior_id"`   // Id of behavior to use.
	TaskCooldown float32         `json:"task_cooldown"` // time before a task can be used again
	TaskUser     enum.TaskUser   `json:"task_user"`     // Which task user I am. Will decide which tasks I will be able to take.
}

func (w BehaviorComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return w
}

func getBehaviorComponentData() ([]byte, error) {
	behaviorComponentHash := Sum("BehaviorComponentData")
	behaviorComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(behaviorComponentHashData, binary.LittleEndian, behaviorComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, behaviorComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getBehaviorComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("BehaviorComponentData")
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
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("BehaviorComponent") {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (data type was not BehaviorComponent)")
	}

	behaviorComponentData, err := getBehaviorComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get behavior component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(behaviorComponentData)

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

	var behaviorComponentType DLTypeDesc
	behaviorComponentType, ok = typelib.Types[Sum("BehaviorComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find BehaviorComponent hash in dl_library")
	}

	componentData := make([]byte, behaviorComponentType.Size)
	if _, err := r.Seek(int64(behaviorComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseBehaviorComponents() (map[stingray.Hash]BehaviorComponent, error) {
	unitHash := Sum("BehaviorComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var behaviorType DLTypeDesc
	var ok bool
	behaviorType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find BehaviorComponentData hash in dl_library")
	}

	if len(behaviorType.Members) != 2 {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (there should be 2 members but were actually %v)", len(behaviorType.Members))
	}

	if behaviorType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (hashmap atom was not inline array)")
	}

	if behaviorType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (data atom was not inline array)")
	}

	if behaviorType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (hashmap storage was not struct)")
	}

	if behaviorType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (data storage was not struct)")
	}

	if behaviorType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if behaviorType.Members[1].TypeID != Sum("BehaviorComponent") {
		return nil, fmt.Errorf("BehaviorComponentData unexpected format (data type was not BehaviorComponent)")
	}

	behaviorComponentData, err := getBehaviorComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get behavior component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(behaviorComponentData)

	hashmap := make([]ComponentIndexData, behaviorType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]BehaviorComponent, behaviorType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]BehaviorComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
