package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type InventoryComponent struct {
	DefaultPrimary     stingray.Hash // [adhd]Path to the primary equipment entity.
	DefaultSidearm     stingray.Hash // [adhd]Path to the sidearm equipment entity.
	DefaultSupport     stingray.Hash // [adhd]Path to the support equipment entity.
	DefaultThrowable   stingray.Hash // [adhd]Path to the throwable equipment entity.
	DefaultBackpack    stingray.Hash // [adhd]Path to the backpack entity.
	DefaultMeleeWeapon stingray.Hash // [adhd]Path to the melee entity.
	StartAmountStims   uint32        // max amount of stims. [sic]
	RefillAmountStims  uint32        // max amount of stims. [sic]
	MaxAmountStims     uint32        // max amount of stims.
	CanDropEquipment   uint8         // [bool]Indicates if this unit can drop the equipment or not
	_                  [3]uint8
	UnknownHashArray   [10]stingray.Hash // Unknown, name length 33
}

type SimpleInventoryComponent struct {
	DefaultPrimary     string   `json:"default_primary"`
	DefaultSidearm     string   `json:"default_sidearm"`
	DefaultSupport     string   `json:"default_support"`
	DefaultThrowable   string   `json:"default_throwable"`
	DefaultBackpack    string   `json:"default_backpack"`
	DefaultMeleeWeapon string   `json:"default_melee_weapon"`
	StartAmountStims   uint32   `json:"start_amount_stims"`
	RefillAmountStims  uint32   `json:"refill_amount_stims"`
	MaxAmountStims     uint32   `json:"max_amount_stims"`
	CanDropEquipment   bool     `json:"can_drop_equipment"`
	UnknownHashArray   []string `json:"unknown_hash_array"`
}

func (w InventoryComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	unknownHashArray := make([]string, 0)
	for _, hash := range w.UnknownHashArray {
		if hash.Value == 0 {
			break
		}
		unknownHashArray = append(unknownHashArray, lookupHash(hash))
	}
	return SimpleInventoryComponent{
		DefaultPrimary:     lookupHash(w.DefaultPrimary),
		DefaultSidearm:     lookupHash(w.DefaultSidearm),
		DefaultSupport:     lookupHash(w.DefaultSupport),
		DefaultThrowable:   lookupHash(w.DefaultThrowable),
		DefaultBackpack:    lookupHash(w.DefaultBackpack),
		DefaultMeleeWeapon: lookupHash(w.DefaultMeleeWeapon),
		StartAmountStims:   w.StartAmountStims,
		RefillAmountStims:  w.RefillAmountStims,
		MaxAmountStims:     w.MaxAmountStims,
		CanDropEquipment:   w.CanDropEquipment != 0,
		UnknownHashArray:   unknownHashArray,
	}
}

func getInventoryComponentData() ([]byte, error) {
	inventoryComponentHash := Sum("InventoryComponentData")
	inventoryComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(inventoryComponentHashData, binary.LittleEndian, inventoryComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, inventoryComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getInventoryComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("InventoryComponentData")
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
		return nil, fmt.Errorf("InventoryComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("InventoryComponent") {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (data type was not InventoryComponent)")
	}

	inventoryComponentData, err := getInventoryComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get inventory component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(inventoryComponentData)

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
		return nil, fmt.Errorf("%v not found in inventory component data", hash.String())
	}

	var inventoryComponentType DLTypeDesc
	inventoryComponentType, ok = typelib.Types[Sum("InventoryComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find InventoryComponent hash in dl_library")
	}

	componentData := make([]byte, inventoryComponentType.Size)
	if _, err := r.Seek(int64(inventoryComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseInventoryComponents() (map[stingray.Hash]InventoryComponent, error) {
	unitHash := Sum("InventoryComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var inventoryType DLTypeDesc
	var ok bool
	inventoryType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find InventoryComponentData hash in dl_library")
	}

	if len(inventoryType.Members) != 2 {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (there should be 2 members but were actually %v)", len(inventoryType.Members))
	}

	if inventoryType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (hashmap atom was not inline array)")
	}

	if inventoryType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (data atom was not inline array)")
	}

	if inventoryType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (hashmap storage was not struct)")
	}

	if inventoryType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (data storage was not struct)")
	}

	if inventoryType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if inventoryType.Members[1].TypeID != Sum("InventoryComponent") {
		return nil, fmt.Errorf("InventoryComponentData unexpected format (data type was not InventoryComponent)")
	}

	inventoryComponentData, err := getInventoryComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get inventory component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(inventoryComponentData)

	hashmap := make([]ComponentIndexData, inventoryType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]InventoryComponent, inventoryType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]InventoryComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
