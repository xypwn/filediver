package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type WeaponLinkedAmmoComponent struct {
	Tag                             enum.Tag
	AmmoMode                        enum.WeaponLinkedAmmoMode
	InventorySlot                   enum.InventorySlot
	ThisUnitAmmoLevelAnimationVar   stingray.ThinHash
	ThisUnitDisplayMaterial         stingray.ThinHash
	ThisUnitDisplayMaterialVar      stingray.ThinHash
	LinkedUnitAmmoLevelAnimationVar stingray.ThinHash
	UnknownBool1                    uint8
	_                               [3]uint8
	UnknownEnum                     uint32
	UnknownBool2                    uint8
	_                               [3]uint8
	LinkedUnitDisplayMaterial       stingray.ThinHash
	LinkedUnitDisplayMaterialVar    stingray.ThinHash
}

type SimpleWeaponLinkedAmmoComponent struct {
	Tag                             enum.Tag                  `json:"tag"`
	AmmoMode                        enum.WeaponLinkedAmmoMode `json:"ammo_mode"`
	InventorySlot                   enum.InventorySlot        `json:"inventory_slot"`
	ThisUnitAmmoLevelAnimationVar   string                    `json:"this_unit_ammo_level_animation_var,omitempty"`
	ThisUnitDisplayMaterial         string                    `json:"this_unit_display_material,omitempty"`
	ThisUnitDisplayMaterialVar      string                    `json:"this_unit_display_material_var,omitempty"`
	LinkedUnitAmmoLevelAnimationVar string                    `json:"linked_unit_ammo_level_animation_var,omitempty"`
	UnknownBool1                    bool                      `json:"unknown_bool_1"`
	UnknownEnum                     uint32                    `json:"unknown_enum"`
	UnknownBool2                    bool                      `json:"unknown_bool_2"`
	LinkedUnitDisplayMaterial       string                    `json:"linked_unit_display_material,omitempty"`
	LinkedUnitDisplayMaterialVar    string                    `json:"linked_unit_display_material_var,omitempty"`
}

func (m WeaponLinkedAmmoComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	var emptyLookupThinHash ThinHashLookup = func(t stingray.ThinHash) string {
		if t.Value == 0x0 {
			return ""
		}
		return lookupThinHash(t)
	}

	return SimpleWeaponLinkedAmmoComponent{
		Tag:                             m.Tag,
		AmmoMode:                        m.AmmoMode,
		InventorySlot:                   m.InventorySlot,
		ThisUnitAmmoLevelAnimationVar:   emptyLookupThinHash(m.ThisUnitAmmoLevelAnimationVar),
		ThisUnitDisplayMaterial:         emptyLookupThinHash(m.ThisUnitDisplayMaterial),
		ThisUnitDisplayMaterialVar:      emptyLookupThinHash(m.ThisUnitDisplayMaterialVar),
		LinkedUnitAmmoLevelAnimationVar: emptyLookupThinHash(m.LinkedUnitAmmoLevelAnimationVar),
		UnknownBool1:                    m.UnknownBool1 != 0,
		UnknownEnum:                     m.UnknownEnum,
		UnknownBool2:                    m.UnknownBool2 != 0,
		LinkedUnitDisplayMaterial:       emptyLookupThinHash(m.LinkedUnitDisplayMaterial),
		LinkedUnitDisplayMaterialVar:    emptyLookupThinHash(m.LinkedUnitDisplayMaterialVar),
	}
}

func getWeaponLinkedAmmoComponentData() ([]byte, error) {
	weaponLinkedAmmoHash := Sum("WeaponLinkedAmmoComponentData")
	weaponLinkedAmmoHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponLinkedAmmoHashData, binary.LittleEndian, weaponLinkedAmmoHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponLinkedAmmoHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponLinkedAmmoComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponLinkedAmmoCmpDataHash := Sum("WeaponLinkedAmmoComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponLinkedAmmoCmpDataType DLTypeDesc
	var ok bool
	weaponLinkedAmmoCmpDataType, ok = typelib.Types[WeaponLinkedAmmoCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponLinkedAmmoComponentData hash in dl_library")
	}

	if len(weaponLinkedAmmoCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponLinkedAmmoCmpDataType.Members))
	}

	if weaponLinkedAmmoCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponLinkedAmmoCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (data atom was not inline array)")
	}

	if weaponLinkedAmmoCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponLinkedAmmoCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (data storage was not struct)")
	}

	if weaponLinkedAmmoCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponLinkedAmmoCmpDataType.Members[1].TypeID != Sum("WeaponLinkedAmmoComponent") {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (data type was not WeaponLinkedAmmoComponent)")
	}

	weaponLinkedAmmoComponentData, err := getWeaponLinkedAmmoComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon linked ammo component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponLinkedAmmoComponentData)

	hashmap := make([]ComponentIndexData, weaponLinkedAmmoCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon linked ammo component data", hash.String())
	}

	var weaponLinkedAmmoComponentType DLTypeDesc
	weaponLinkedAmmoComponentType, ok = typelib.Types[Sum("WeaponLinkedAmmoComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponLinkedAmmoComponent hash in dl_library")
	}

	componentData := make([]byte, weaponLinkedAmmoComponentType.Size)
	if _, err := r.Seek(int64(weaponLinkedAmmoComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponLinkedAmmoComponents() (map[stingray.Hash]WeaponLinkedAmmoComponent, error) {
	weaponLinkedAmmoHash := Sum("WeaponLinkedAmmoComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponLinkedAmmoType DLTypeDesc
	var ok bool
	weaponLinkedAmmoType, ok = typelib.Types[weaponLinkedAmmoHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponLinkedAmmoComponentData hash in dl_library")
	}

	if len(weaponLinkedAmmoType.Members) != 2 {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponLinkedAmmoType.Members))
	}

	if weaponLinkedAmmoType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponLinkedAmmoType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (data atom was not inline array)")
	}

	if weaponLinkedAmmoType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponLinkedAmmoType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (data storage was not struct)")
	}

	if weaponLinkedAmmoType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponLinkedAmmoType.Members[1].TypeID != Sum("WeaponLinkedAmmoComponent") {
		return nil, fmt.Errorf("WeaponLinkedAmmoComponentData unexpected format (data type was not WeaponLinkedAmmoComponent)")
	}

	weaponLinkedAmmoComponentData, err := getWeaponLinkedAmmoComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponLinkedAmmoComponentData)

	hashmap := make([]ComponentIndexData, weaponLinkedAmmoType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponLinkedAmmoComponent, weaponLinkedAmmoType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponLinkedAmmoComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
