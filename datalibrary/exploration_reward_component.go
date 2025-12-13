package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type ExplorationRewardComponent struct {
	AbilityIdAtInteraction           enum.AbilityId    // Indicates the ability that the reward entity will play once the avatar interact with it.
	AudioEventAtInteraction          stingray.ThinHash // [wwise]Indicates the audio event that the reward entity will fire once the avatar interact with it.
	AnimationEventAtInteraction      stingray.ThinHash // [string]Indicates the animation event that the reward entity will fire once the avatar interact with it.
	AbilityIdAtInteractionInteractor enum.AbilityId    // Indicates the ability that the avatar will play once interacts with the reward.
	ShouldDestroyAfterInteraction    uint8             // [bool]Indicates if we should destroy the entity after interaction.
	_                                [3]uint8
	DefaultExplorationRewardType     enum.ExplorationRewardType // Indicates the reward type that we should give by default if we don't specify any script data in the marker.
}

type SimpleExplorationRewardComponent struct {
	AbilityIdAtInteraction           enum.AbilityId             `json:"ability_id_at_interaction"`
	AudioEventAtInteraction          string                     `json:"audio_event_at_interaction"`
	AnimationEventAtInteraction      string                     `json:"animation_event_at_interaction"`
	AbilityIdAtInteractionInteractor enum.AbilityId             `json:"ability_id_at_interaction_interactor"`
	ShouldDestroyAfterInteraction    bool                       `json:"should_destroy_after_interaction"`
	DefaultExplorationRewardType     enum.ExplorationRewardType `json:"default_exploration_reward_type"`
}

func (w ExplorationRewardComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleExplorationRewardComponent{
		AbilityIdAtInteraction:           w.AbilityIdAtInteraction,
		AudioEventAtInteraction:          lookupThinHash(w.AudioEventAtInteraction),
		AnimationEventAtInteraction:      lookupThinHash(w.AnimationEventAtInteraction),
		AbilityIdAtInteractionInteractor: w.AbilityIdAtInteractionInteractor,
		ShouldDestroyAfterInteraction:    w.ShouldDestroyAfterInteraction != 0,
		DefaultExplorationRewardType:     w.DefaultExplorationRewardType,
	}
}

func getExplorationRewardComponentData() ([]byte, error) {
	explorationRewardComponentHash := Sum("ExplorationRewardComponentData")
	explorationRewardComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(explorationRewardComponentHashData, binary.LittleEndian, explorationRewardComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, explorationRewardComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getExplorationRewardComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("ExplorationRewardComponentData")
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
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("ExplorationRewardComponent") {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (data type was not ExplorationRewardComponent)")
	}

	explorationRewardComponentData, err := getExplorationRewardComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get exploration reward component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(explorationRewardComponentData)

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
		return nil, fmt.Errorf("%v not found in exploration reward component data", hash.String())
	}

	var explorationRewardComponentType DLTypeDesc
	explorationRewardComponentType, ok = typelib.Types[Sum("ExplorationRewardComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find ExplorationRewardComponent hash in dl_library")
	}

	componentData := make([]byte, explorationRewardComponentType.Size)
	if _, err := r.Seek(int64(explorationRewardComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseExplorationRewardComponents() (map[stingray.Hash]ExplorationRewardComponent, error) {
	unitHash := Sum("ExplorationRewardComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var explorationRewardType DLTypeDesc
	var ok bool
	explorationRewardType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find ExplorationRewardComponentData hash in dl_library")
	}

	if len(explorationRewardType.Members) != 2 {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (there should be 2 members but were actually %v)", len(explorationRewardType.Members))
	}

	if explorationRewardType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (hashmap atom was not inline array)")
	}

	if explorationRewardType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (data atom was not inline array)")
	}

	if explorationRewardType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (hashmap storage was not struct)")
	}

	if explorationRewardType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (data storage was not struct)")
	}

	if explorationRewardType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if explorationRewardType.Members[1].TypeID != Sum("ExplorationRewardComponent") {
		return nil, fmt.Errorf("ExplorationRewardComponentData unexpected format (data type was not ExplorationRewardComponent)")
	}

	explorationRewardComponentData, err := getExplorationRewardComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get exploration reward component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(explorationRewardComponentData)

	hashmap := make([]ComponentIndexData, explorationRewardType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]ExplorationRewardComponent, explorationRewardType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]ExplorationRewardComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
