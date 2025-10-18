package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type AnimationMergeOptions struct {
	MaxStartTime  float32 `json:"max_start_time"` // Sets the maximum time into an animation that we can start, in seconds. If a value of `0.2` is set, it means that if we can find any existing player playing this animation that has played for 0-0.2 seconds, then we will reuse that player instead of creating a new one. The animation will jump to the position where that player is, instead of starting at t=0. The default is `0`, which allows animation evaluations to be merged only if they began in the same frame.
	MaxDrift      float32 `json:"max_drift"`      // Controls how much the displayed animation time is allowed to drift from the actual animation time. If two animations that are sharing a player are played with different speeds, they will eventually drift apart. At some point the animations will be considered too different and the evaluator will be split in two. This parameter controls when that happens. A value of `0.2` means that if the animations differs more than 0.2 seconds, the evaluator will be split.
	ClockFidelity float32 `json:"clock_fidelity"` // This controls how much the animation will respect the set playing speed. By slightly changing the playing speed of an animation, the animation player can prevent two playing animations from drifting apart and thus forcing a split of the evaluators. A clock fidelity of `0.9` means that the speed will be within 90% of what has been requested: i.e. it is allowed to change by 10 percent. The default value of `1.0` means that the speed will be exactly what has been requested.
}

type AnimationVariable struct {
	Name  stingray.ThinHash // [string]The name of the animation variable to set.
	Value float32           // The value to set on the animation variable.
}

type AnimationComponent struct {
	MergeOptions                 AnimationMergeOptions // Animation merge options (or animation sharing options).
	Variables                    [8]AnimationVariable  // Information about any variables to be set when spawning the animated entity.
	HotjoinSyncAnimationTime     uint8                 // [bool]If set, the time into the current animation is also synced. Useful for long animations, such as the dropship sequence.
	IgnoreInvisibileUnitsForLods uint8                 // [bool]If set, invisible units are ignored for the OOBB used to select LOD steps.
	_                            [2]uint8
}

type SimpleAnimationVariable struct {
	Name  string  `json:"name"`
	Value float32 `json:"value"`
}

type SimpleAnimationComponent struct {
	MergeOptions                 AnimationMergeOptions     `json:"merge_options"`
	Variables                    []SimpleAnimationVariable `json:"variables,omitempty"`
	HotjoinSyncAnimationTime     bool                      `json:"hotjoin_sync_animation_time"`
	IgnoreInvisibileUnitsForLods bool                      `json:"ignore_invisible_units_for_lods"`
}

func (w AnimationComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	variables := make([]SimpleAnimationVariable, 0)
	for _, variable := range w.Variables {
		if variable.Name.Value == 0 {
			break
		}
		variables = append(variables, SimpleAnimationVariable{
			Name:  lookupThinHash(variable.Name),
			Value: variable.Value,
		})
	}
	return SimpleAnimationComponent{
		MergeOptions:                 w.MergeOptions,
		Variables:                    variables,
		HotjoinSyncAnimationTime:     w.HotjoinSyncAnimationTime != 0,
		IgnoreInvisibileUnitsForLods: w.IgnoreInvisibileUnitsForLods != 0,
	}
}

func getAnimationComponentData() ([]byte, error) {
	animationComponentHash := Sum("AnimationComponentData")
	animationComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(animationComponentHashData, binary.LittleEndian, animationComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, animationComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getAnimationComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("AnimationComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var animationCmpDataType DLTypeDesc
	var ok bool
	animationCmpDataType, ok = typelib.Types[UnitCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(animationCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (there should be 2 members but were actually %v)", len(animationCmpDataType.Members))
	}

	if animationCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (hashmap atom was not inline array)")
	}

	if animationCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (data atom was not inline array)")
	}

	if animationCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (hashmap storage was not struct)")
	}

	if animationCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (data storage was not struct)")
	}

	if animationCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if animationCmpDataType.Members[1].TypeID != Sum("AnimationComponent") {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (data type was not AnimationComponent)")
	}

	animationComponentData, err := getAnimationComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get animation component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(animationComponentData)

	hashmap := make([]ComponentIndexData, animationCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in animation component data", hash.String())
	}

	var animationComponentType DLTypeDesc
	animationComponentType, ok = typelib.Types[Sum("AnimationComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find AnimationComponent hash in dl_library")
	}

	componentData := make([]byte, animationComponentType.Size)
	if _, err := r.Seek(int64(animationComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseAnimationComponents() (map[stingray.Hash]AnimationComponent, error) {
	unitHash := Sum("AnimationComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var animationType DLTypeDesc
	var ok bool
	animationType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find AnimationComponentData hash in dl_library")
	}

	if len(animationType.Members) != 2 {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (there should be 2 members but were actually %v)", len(animationType.Members))
	}

	if animationType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (hashmap atom was not inline array)")
	}

	if animationType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (data atom was not inline array)")
	}

	if animationType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (hashmap storage was not struct)")
	}

	if animationType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (data storage was not struct)")
	}

	if animationType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if animationType.Members[1].TypeID != Sum("AnimationComponent") {
		return nil, fmt.Errorf("AnimationComponentData unexpected format (data type was not AnimationComponent)")
	}

	animationComponentData, err := getAnimationComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get animation component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(animationComponentData)

	hashmap := make([]ComponentIndexData, animationType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]AnimationComponent, animationType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]AnimationComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
