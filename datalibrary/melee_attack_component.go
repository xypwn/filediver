package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type MeleeAttackComponent struct {
	OverrideMeleeAbility         enum.AbilityId `json:"override_melee_ability"`           // If not invalid, it'll override the melee ability that gets executed while helding this entity, this could be a weapon, an attachment or a carryable.
	Weight                       uint32         `json:"weight"`                           // Used to indicate how much weight this override has when trying to pick the final override_melee_ability
	OverrideMeleeAbilityForProne enum.AbilityId `json:"override_melee_ability_for_prone"` // If not invalid, it'll override the melee ability that gets executed while being prone and helding this entity, this could be a weapon, an attachment or a carryable.
	WeightForProne               uint32         `json:"weight_for_prone"`                 // Used to indicate how much weight this override has when trying to pick the final override_melee_ability_for_prone
}

func (m MeleeAttackComponent) ToSimple(_ HashLookup, _ ThinHashLookup, _ StringsLookup) any {
	return m
}

func getMeleeAttackComponentData() ([]byte, error) {
	meleeAttackHash := Sum("MeleeAttackComponentData")
	meleeAttackHashData := make([]byte, 4)
	if _, err := binary.Encode(meleeAttackHashData, binary.LittleEndian, meleeAttackHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, meleeAttackHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getMeleeAttackComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	MeleeAttackCmpDataHash := Sum("MeleeAttackComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var meleeAttackCmpDataType DLTypeDesc
	var ok bool
	meleeAttackCmpDataType, ok = typelib.Types[MeleeAttackCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(meleeAttackCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (there should be 2 members but were actually %v)", len(meleeAttackCmpDataType.Members))
	}

	if meleeAttackCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (hashmap atom was not inline array)")
	}

	if meleeAttackCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (data atom was not inline array)")
	}

	if meleeAttackCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (hashmap storage was not struct)")
	}

	if meleeAttackCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (data storage was not struct)")
	}

	if meleeAttackCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if meleeAttackCmpDataType.Members[1].TypeID != Sum("MeleeAttackComponent") {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (data type was not MeleeAttackComponent)")
	}

	meleeAttackComponentData, err := getMeleeAttackComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get melee attack component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(meleeAttackComponentData)

	hashmap := make([]ComponentIndexData, meleeAttackCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in melee attack component data", hash.String())
	}

	var meleeAttackComponentType DLTypeDesc
	meleeAttackComponentType, ok = typelib.Types[Sum("MeleeAttackComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find MeleeAttackComponent hash in dl_library")
	}

	componentData := make([]byte, meleeAttackComponentType.Size)
	if _, err := r.Seek(int64(meleeAttackComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseMeleeAttackComponents() (map[stingray.Hash]MeleeAttackComponent, error) {
	meleeAttackHash := Sum("MeleeAttackComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var meleeAttackType DLTypeDesc
	var ok bool
	meleeAttackType, ok = typelib.Types[meleeAttackHash]
	if !ok {
		return nil, fmt.Errorf("could not find MeleeAttackComponentData hash in dl_library")
	}

	if len(meleeAttackType.Members) != 2 {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (there should be 2 members but were actually %v)", len(meleeAttackType.Members))
	}

	if meleeAttackType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (hashmap atom was not inline array)")
	}

	if meleeAttackType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (data atom was not inline array)")
	}

	if meleeAttackType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (hashmap storage was not struct)")
	}

	if meleeAttackType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (data storage was not struct)")
	}

	if meleeAttackType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if meleeAttackType.Members[1].TypeID != Sum("MeleeAttackComponent") {
		return nil, fmt.Errorf("MeleeAttackComponentData unexpected format (data type was not MeleeAttackComponent)")
	}

	meleeAttackComponentData, err := getMeleeAttackComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(meleeAttackComponentData)

	hashmap := make([]ComponentIndexData, meleeAttackType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]MeleeAttackComponent, meleeAttackType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]MeleeAttackComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
