package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type HellpodPayloadComponent struct {
	RemoveTime       float32        `json:"remove_time"`        // The time between playing the retraction animation event and removing the entity.
	LifeTime         float32        `json:"life_time"`          // The life time of this payload. 0 is infinite.
	OnRetractAbility enum.AbilityId `json:"on_retract_ability"` // Ability to play on this unit when it's retracted.
}

func (w HellpodPayloadComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return w
}

func getHellpodPayloadComponentData() ([]byte, error) {
	hellpodPayloadComponentHash := Sum("HellpodPayloadComponentData")
	hellpodPayloadComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(hellpodPayloadComponentHashData, binary.LittleEndian, hellpodPayloadComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, hellpodPayloadComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getHellpodPayloadComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("HellpodPayloadComponentData")
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
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("HellpodPayloadComponent") {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (data type was not HellpodPayloadComponent)")
	}

	hellpodPayloadComponentData, err := getHellpodPayloadComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get hellpod payload component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(hellpodPayloadComponentData)

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
		return nil, fmt.Errorf("%v not found in hellpod payload component data", hash.String())
	}

	var hellpodPayloadComponentType DLTypeDesc
	hellpodPayloadComponentType, ok = typelib.Types[Sum("HellpodPayloadComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find HellpodPayloadComponent hash in dl_library")
	}

	componentData := make([]byte, hellpodPayloadComponentType.Size)
	if _, err := r.Seek(int64(hellpodPayloadComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseHellpodPayloadComponents() (map[stingray.Hash]HellpodPayloadComponent, error) {
	unitHash := Sum("HellpodPayloadComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var hellpodPayloadType DLTypeDesc
	var ok bool
	hellpodPayloadType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find HellpodPayloadComponentData hash in dl_library")
	}

	if len(hellpodPayloadType.Members) != 2 {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (there should be 2 members but were actually %v)", len(hellpodPayloadType.Members))
	}

	if hellpodPayloadType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (hashmap atom was not inline array)")
	}

	if hellpodPayloadType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (data atom was not inline array)")
	}

	if hellpodPayloadType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (hashmap storage was not struct)")
	}

	if hellpodPayloadType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (data storage was not struct)")
	}

	if hellpodPayloadType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if hellpodPayloadType.Members[1].TypeID != Sum("HellpodPayloadComponent") {
		return nil, fmt.Errorf("HellpodPayloadComponentData unexpected format (data type was not HellpodPayloadComponent)")
	}

	hellpodPayloadComponentData, err := getHellpodPayloadComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get hellpod payload component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(hellpodPayloadComponentData)

	hashmap := make([]ComponentIndexData, hellpodPayloadType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]HellpodPayloadComponent, hellpodPayloadType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]HellpodPayloadComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
