package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type MaterialVariableAnimation struct {
	Name        stingray.ThinHash // [string]The name of the animation. Used to start and stop it from abilities.
	TargetValue mgl32.Vec4        // The target value to animate towards.
	Duration    float32           // The duration in seconds of the animation.
	Loops       uint32            // The amount of time this animation should loop, before ending. Setting it to 0 makes it infinite.
	AnimType    enum.MaterialVariableAnimationType
}

type MaterialVariableValue struct {
	Name  stingray.ThinHash
	Value mgl32.Vec4
}

type MaterialVariables struct {
	VariableName     stingray.ThinHash            // [string]The name of the material variable to set.
	MaterialSlotName stingray.ThinHash            // [string]The material slot to set on. If not set, then applied to all material slots.
	Type             enum.MaterialVariableType    // The type of the material variable to set.
	Value            mgl32.Vec4                   // The value to set on the material variable.
	Animations       [4]MaterialVariableAnimation // The animations that can be used on this variable.
	Values           [4]MaterialVariableValue     // The values that can be set on this variable
	UnknownBool1     uint8                        // unknown, name length 14
	UnknownBool2     uint8                        // unknown, name length 16
	_                [2]uint8
}

type MaterialVariablesComponent struct {
	Variables [8]MaterialVariables // Information about any variables that can be set / modified.
}

type SimpleMaterialVariableAnimation struct {
	Name        string                             `json:"name"`
	TargetValue mgl32.Vec4                         `json:"target_value"`
	Duration    float32                            `json:"duration"`
	Loops       uint32                             `json:"loops"`
	AnimType    enum.MaterialVariableAnimationType `json:"anim_type"`
}

type SimpleMaterialVariableValue struct {
	Name  string     `json:"name"`
	Value mgl32.Vec4 `json:"value"`
}

type SimpleMaterialVariables struct {
	VariableName     string                            `json:"variable_name"`
	MaterialSlotName string                            `json:"material_slot_name"`
	Type             enum.MaterialVariableType         `json:"type"`
	Value            mgl32.Vec4                        `json:"value"`
	Animations       []SimpleMaterialVariableAnimation `json:"animations"`
	Values           []SimpleMaterialVariableValue     `json:"values"`
	UnknownBool1     bool                              `json:"unknown_bool1"`
	UnknownBool2     bool                              `json:"unknown_bool2"`
}

type SimpleMaterialVariablesComponent struct {
	Variables []SimpleMaterialVariables `json:"variables"`
}

func (w MaterialVariablesComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	variables := make([]SimpleMaterialVariables, 0)
	for _, variable := range w.Variables {
		if variable.VariableName.Value == 0 {
			break
		}
		animations := make([]SimpleMaterialVariableAnimation, 0)
		for _, animation := range variable.Animations {
			if animation.Name.Value == 0 {
				break
			}
			animations = append(animations, SimpleMaterialVariableAnimation{
				Name:        lookupThinHash(animation.Name),
				TargetValue: animation.TargetValue,
				Duration:    animation.Duration,
				Loops:       animation.Loops,
				AnimType:    animation.AnimType,
			})
		}
		values := make([]SimpleMaterialVariableValue, 0)
		for _, value := range variable.Values {
			if value.Name.Value == 0 {
				break
			}
			values = append(values, SimpleMaterialVariableValue{
				Name:  lookupThinHash(value.Name),
				Value: value.Value,
			})
		}
		variables = append(variables, SimpleMaterialVariables{
			VariableName:     lookupThinHash(variable.VariableName),
			MaterialSlotName: lookupThinHash(variable.MaterialSlotName),
			Type:             variable.Type,
			Value:            variable.Value,
			Animations:       animations,
			Values:           values,
			UnknownBool1:     variable.UnknownBool1 != 0,
			UnknownBool2:     variable.UnknownBool2 != 0,
		})
	}
	return SimpleMaterialVariablesComponent{
		Variables: variables,
	}
}

func getMaterialVariablesComponentData() ([]byte, error) {
	materialVariablesComponentHash := Sum("MaterialVariablesComponentData")
	materialVariablesComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(materialVariablesComponentHashData, binary.LittleEndian, materialVariablesComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, materialVariablesComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getMaterialVariablesComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("MaterialVariablesComponentData")
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
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("MaterialVariablesComponent") {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (data type was not MaterialVariablesComponent)")
	}

	materialVariablesComponentData, err := getMaterialVariablesComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get material variables component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(materialVariablesComponentData)

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
		return nil, fmt.Errorf("%v not found in material variables component data", hash.String())
	}

	var materialVariablesComponentType DLTypeDesc
	materialVariablesComponentType, ok = typelib.Types[Sum("MaterialVariablesComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find MaterialVariablesComponent hash in dl_library")
	}

	componentData := make([]byte, materialVariablesComponentType.Size)
	if _, err := r.Seek(int64(materialVariablesComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseMaterialVariablesComponents() (map[stingray.Hash]MaterialVariablesComponent, error) {
	unitHash := Sum("MaterialVariablesComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var materialVariablesType DLTypeDesc
	var ok bool
	materialVariablesType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find MaterialVariablesComponentData hash in dl_library")
	}

	if len(materialVariablesType.Members) != 2 {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (there should be 2 members but were actually %v)", len(materialVariablesType.Members))
	}

	if materialVariablesType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (hashmap atom was not inline array)")
	}

	if materialVariablesType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (data atom was not inline array)")
	}

	if materialVariablesType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (hashmap storage was not struct)")
	}

	if materialVariablesType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (data storage was not struct)")
	}

	if materialVariablesType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if materialVariablesType.Members[1].TypeID != Sum("MaterialVariablesComponent") {
		return nil, fmt.Errorf("MaterialVariablesComponentData unexpected format (data type was not MaterialVariablesComponent)")
	}

	materialVariablesComponentData, err := getMaterialVariablesComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get material variables component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(materialVariablesComponentData)

	hashmap := make([]ComponentIndexData, materialVariablesType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]MaterialVariablesComponent, materialVariablesType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]MaterialVariablesComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
