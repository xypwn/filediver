package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type MaterialSwap struct {
	Name     stingray.ThinHash // [string]The name of the swap.
	_        [4]uint8
	Material stingray.Hash // [material]The material to set.
}

type EnemyTagSwap struct {
	Tag  enum.EnemyTag
	Name stingray.ThinHash
}

type MaterialSwapSlot struct {
	MaterialSlotName stingray.ThinHash // [string]The material slot to set on. If not set, then applied to all material slots.
	_                [4]uint8
	SwapSettings     [8]MaterialSwap
}

type MaterialSwapComponent struct {
	MaterialSlots [8]MaterialSwapSlot  // Material slots that can be set
	EnemyNames    [8]stingray.ThinHash // name is guessed - name length 11
	EnemyTagSwaps [8]EnemyTagSwap
}

type SimpleMaterialSwap struct {
	Name     string `json:"name"`
	Material string `json:"material"`
}

type SimpleEnemyTagSwap struct {
	Tag  enum.EnemyTag `json:"tag"`
	Name string        `json:"name"`
}

type SimpleMaterialSwapSlot struct {
	MaterialSlotName string               `json:"material_slot_name"`
	SwapSettings     []SimpleMaterialSwap `json:"swap_settings"`
}

type SimpleMaterialSwapComponent struct {
	MaterialSlots []SimpleMaterialSwapSlot `json:"material_slots"`
	EnemyNames    []string                 `json:"enemy_names"`
	EnemyTagSwaps []SimpleEnemyTagSwap     `json:"enemy_tag_swaps"`
}

func (w MaterialSwapComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	materialSlots := make([]SimpleMaterialSwapSlot, 0)
	for _, slot := range w.MaterialSlots {
		swapSettings := make([]SimpleMaterialSwap, 0)
		for _, swap := range slot.SwapSettings {
			if swap.Name.Value == 0 {
				break
			}
			swapSettings = append(swapSettings, SimpleMaterialSwap{
				Name:     lookupThinHash(swap.Name),
				Material: lookupHash(swap.Material),
			})
		}
		if len(swapSettings) == 0 && slot.MaterialSlotName.Value == 0 {
			break
		}
		materialSlots = append(materialSlots, SimpleMaterialSwapSlot{
			MaterialSlotName: lookupThinHash(slot.MaterialSlotName),
			SwapSettings:     swapSettings,
		})
	}

	enemyNames := make([]string, 0)
	for _, name := range w.EnemyNames {
		if name.Value == 0 {
			break
		}
		enemyNames = append(enemyNames, lookupThinHash(name))
	}

	enemyTagSwaps := make([]SimpleEnemyTagSwap, 0)
	for _, tagSwap := range w.EnemyTagSwaps {
		if tagSwap.Tag == enum.EnemyTag_None {
			break
		}
		enemyTagSwaps = append(enemyTagSwaps, SimpleEnemyTagSwap{
			Tag:  tagSwap.Tag,
			Name: lookupThinHash(tagSwap.Name),
		})
	}

	return SimpleMaterialSwapComponent{
		MaterialSlots: materialSlots,
		EnemyNames:    enemyNames,
		EnemyTagSwaps: enemyTagSwaps,
	}
}

func getMaterialSwapComponentData() ([]byte, error) {
	materialSwapComponentHash := Sum("MaterialSwapComponentData")
	materialSwapComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(materialSwapComponentHashData, binary.LittleEndian, materialSwapComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, materialSwapComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getMaterialSwapComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("MaterialSwapComponentData")
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
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("MaterialSwapComponent") {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (data type was not MaterialSwapComponent)")
	}

	materialSwapComponentData, err := getMaterialSwapComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get material swap component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(materialSwapComponentData)

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
		return nil, fmt.Errorf("%v not found in material swap component data", hash.String())
	}

	var materialSwapComponentType DLTypeDesc
	materialSwapComponentType, ok = typelib.Types[Sum("MaterialSwapComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find MaterialSwapComponent hash in dl_library")
	}

	componentData := make([]byte, materialSwapComponentType.Size)
	if _, err := r.Seek(int64(materialSwapComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseMaterialSwapComponents() (map[stingray.Hash]MaterialSwapComponent, error) {
	unitHash := Sum("MaterialSwapComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var materialSwapType DLTypeDesc
	var ok bool
	materialSwapType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find MaterialSwapComponentData hash in dl_library")
	}

	if len(materialSwapType.Members) != 2 {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (there should be 2 members but were actually %v)", len(materialSwapType.Members))
	}

	if materialSwapType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (hashmap atom was not inline array)")
	}

	if materialSwapType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (data atom was not inline array)")
	}

	if materialSwapType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (hashmap storage was not struct)")
	}

	if materialSwapType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (data storage was not struct)")
	}

	if materialSwapType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if materialSwapType.Members[1].TypeID != Sum("MaterialSwapComponent") {
		return nil, fmt.Errorf("MaterialSwapComponentData unexpected format (data type was not MaterialSwapComponent)")
	}

	materialSwapComponentData, err := getMaterialSwapComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get material swap component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(materialSwapComponentData)

	hashmap := make([]ComponentIndexData, materialSwapType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]MaterialSwapComponent, materialSwapType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]MaterialSwapComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
