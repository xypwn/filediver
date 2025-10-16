package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type LocalUnitComponent struct {
	UnitPath stingray.Hash // [unit]Path to the unit for this entity.
	Scale    float32       // Scale value.
}

type SimpleLocalUnitComponent struct {
	UnitPath string  `json:"unit_path"`
	Scale    float32 `json:"scale"`
}

func (w LocalUnitComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) any {
	return SimpleLocalUnitComponent{
		UnitPath: lookupHash(w.UnitPath),
		Scale:    w.Scale,
	}
}

func getLocalUnitComponentData() ([]byte, error) {
	localUnitComponentHash := Sum("LocalUnitComponentData")
	localUnitComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(localUnitComponentHashData, binary.LittleEndian, localUnitComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, localUnitComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getLocalUnitComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("LocalUnitComponentData")
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
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("LocalUnitComponent") {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (data type was not LocalUnitComponent)")
	}

	localUnitComponentData, err := getLocalUnitComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get local unit component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(localUnitComponentData)

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
		return nil, fmt.Errorf("%v not found in local unit component data", hash.String())
	}

	var localUnitComponentType DLTypeDesc
	localUnitComponentType, ok = typelib.Types[Sum("LocalUnitComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find LocalUnitComponent hash in dl_library")
	}

	componentData := make([]byte, localUnitComponentType.Size)
	if _, err := r.Seek(int64(localUnitComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseLocalUnitComponents() (map[stingray.Hash]LocalUnitComponent, error) {
	unitHash := Sum("LocalUnitComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var localUnitType DLTypeDesc
	var ok bool
	localUnitType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find LocalUnitComponentData hash in dl_library")
	}

	if len(localUnitType.Members) != 2 {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (there should be 2 members but were actually %v)", len(localUnitType.Members))
	}

	if localUnitType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (hashmap atom was not inline array)")
	}

	if localUnitType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (data atom was not inline array)")
	}

	if localUnitType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (hashmap storage was not struct)")
	}

	if localUnitType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (data storage was not struct)")
	}

	if localUnitType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if localUnitType.Members[1].TypeID != Sum("LocalUnitComponent") {
		return nil, fmt.Errorf("LocalUnitComponentData unexpected format (data type was not LocalUnitComponent)")
	}

	localUnitComponentData, err := getLocalUnitComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get local unit component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(localUnitComponentData)

	hashmap := make([]ComponentIndexData, localUnitType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]LocalUnitComponent, localUnitType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]LocalUnitComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
