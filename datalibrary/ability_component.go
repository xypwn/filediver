package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type SpawnAbility struct {
	Job     enum.AiJob     `json:"job"`     // If spawned with this job, the corresponding ability will trigger. AiJob_Any will match against all jobs.
	Ability enum.AbilityId `json:"ability"` // The ability to play if spawned with this job.
}

type AbilityComponent struct {
	SpawnAbilities [4]SpawnAbility // Ability to play if spawned with the corresponding job (AiJob_Count)
}

type SimpleAbilityComponent struct {
	SpawnAbilities []SpawnAbility `json:"spawn_abilities"`
}

func (a AbilityComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	abilities := make([]SpawnAbility, 0)
	for _, ability := range a.SpawnAbilities {
		if ability.Job == enum.AiJob_Any && ability.Ability == enum.AbilityId_Invalid && len(abilities) > 0 {
			break
		}
		abilities = append(abilities, ability)
	}

	return SimpleAbilityComponent{
		SpawnAbilities: abilities,
	}
}

func getAbilityComponentData() ([]byte, error) {
	abilityHash := Sum("AbilityComponentData")
	abilityHashData := make([]byte, 4)
	if _, err := binary.Encode(abilityHashData, binary.LittleEndian, abilityHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, abilityHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getAbilityComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	AbilityCmpDataHash := Sum("AbilityComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var abilityCmpDataType DLTypeDesc
	var ok bool
	abilityCmpDataType, ok = typelib.Types[AbilityCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find AbilityComponentData hash in dl_library")
	}

	if len(abilityCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (there should be 2 members but were actually %v)", len(abilityCmpDataType.Members))
	}

	if abilityCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (hashmap atom was not inline array)")
	}

	if abilityCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (data atom was not inline array)")
	}

	if abilityCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (hashmap storage was not struct)")
	}

	if abilityCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (data storage was not struct)")
	}

	if abilityCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if abilityCmpDataType.Members[1].TypeID != Sum("AbilityComponent") {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (data type was not AbilityComponent)")
	}

	abilityComponentData, err := getAbilityComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get ability component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(abilityComponentData)

	hashmap := make([]ComponentIndexData, abilityCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in ability component data", hash.String())
	}

	var abilityComponentType DLTypeDesc
	abilityComponentType, ok = typelib.Types[Sum("AbilityComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find AbilityComponent hash in dl_library")
	}

	componentData := make([]byte, abilityComponentType.Size)
	if _, err := r.Seek(int64(abilityComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseAbilityComponents() (map[stingray.Hash]AbilityComponent, error) {
	abilityHash := Sum("AbilityComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var abilityType DLTypeDesc
	var ok bool
	abilityType, ok = typelib.Types[abilityHash]
	if !ok {
		return nil, fmt.Errorf("could not find AbilityComponentData hash in dl_library")
	}

	if len(abilityType.Members) != 2 {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (there should be 2 members but were actually %v)", len(abilityType.Members))
	}

	if abilityType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (hashmap atom was not inline array)")
	}

	if abilityType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (data atom was not inline array)")
	}

	if abilityType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (hashmap storage was not struct)")
	}

	if abilityType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (data storage was not struct)")
	}

	if abilityType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if abilityType.Members[1].TypeID != Sum("AbilityComponent") {
		return nil, fmt.Errorf("AbilityComponentData unexpected format (data type was not AbilityComponent)")
	}

	abilityComponentData, err := getAbilityComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(abilityComponentData)

	hashmap := make([]ComponentIndexData, abilityType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]AbilityComponent, abilityType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]AbilityComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
