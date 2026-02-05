package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type WieldableComponent struct {
	AllowAds                       uint8 // [bool]If set the wielding player can enter aim-down-sights mode for this weapon.
	_                              [3]uint8
	WielderEnterAnimationEvent     stingray.ThinHash // [string]Animation event to send when a wielder is set.
	WielderExitAnimationEvent      stingray.ThinHash // [string]Animation event to send when the wielder is removed.
	WieldStateAudioEvent           stingray.ThinHash // [wwise]Wwise event to set the currently-equipped weapon state
	SelfWielderEnterAnimationEvent stingray.ThinHash // Maybe the animation event sent to the wielder on entry?
	SelfWielderExitAnimationEvent  stingray.ThinHash // Maybe the animation event sent to the wielder on exit?
	UnkBool                        uint8             // name length 42
	_                              [3]uint8
}

type SimpleWieldableComponent struct {
	AllowAds                       bool   `json:"allow_ads"`
	WielderEnterAnimationEvent     string `json:"wielder_enter_animation_event"`
	WielderExitAnimationEvent      string `json:"wielder_exit_animation_event"`
	WieldStateAudioEvent           string `json:"wield_state_audio_event"`
	SelfWielderEnterAnimationEvent string `json:"self_wielder_enter_animation_event"`
	SelfWielderExitAnimationEvent  string `json:"self_wielder_exit_animation_event"`
	UnkBool                        bool   `json:"unk_bool"`
}

func (w WieldableComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleWieldableComponent{
		AllowAds:                       w.AllowAds != 0,
		WielderEnterAnimationEvent:     lookupThinHash(w.WielderEnterAnimationEvent),
		WielderExitAnimationEvent:      lookupThinHash(w.WielderExitAnimationEvent),
		WieldStateAudioEvent:           lookupThinHash(w.WieldStateAudioEvent),
		SelfWielderEnterAnimationEvent: lookupThinHash(w.SelfWielderEnterAnimationEvent),
		SelfWielderExitAnimationEvent:  lookupThinHash(w.SelfWielderExitAnimationEvent),
		UnkBool:                        w.UnkBool != 0,
	}
}

func getWieldableComponentData() ([]byte, error) {
	wieldableComponentHash := Sum("WieldableComponentData")
	wieldableComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(wieldableComponentHashData, binary.LittleEndian, wieldableComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, wieldableComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWieldableComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("WieldableComponentData")
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
		return nil, fmt.Errorf("WieldableComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("WieldableComponent") {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (data type was not WieldableComponent)")
	}

	wieldableComponentData, err := getWieldableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get wieldable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(wieldableComponentData)

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
		return nil, fmt.Errorf("%v not found in wieldable component data", hash.String())
	}

	var wieldableComponentType DLTypeDesc
	wieldableComponentType, ok = typelib.Types[Sum("WieldableComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WieldableComponent hash in dl_library")
	}

	componentData := make([]byte, wieldableComponentType.Size)
	if _, err := r.Seek(int64(wieldableComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWieldableComponents() (map[stingray.Hash]WieldableComponent, error) {
	unitHash := Sum("WieldableComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var wieldableType DLTypeDesc
	var ok bool
	wieldableType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find WieldableComponentData hash in dl_library")
	}

	if len(wieldableType.Members) != 2 {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (there should be 2 members but were actually %v)", len(wieldableType.Members))
	}

	if wieldableType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if wieldableType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (data atom was not inline array)")
	}

	if wieldableType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (hashmap storage was not struct)")
	}

	if wieldableType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (data storage was not struct)")
	}

	if wieldableType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if wieldableType.Members[1].TypeID != Sum("WieldableComponent") {
		return nil, fmt.Errorf("WieldableComponentData unexpected format (data type was not WieldableComponent)")
	}

	wieldableComponentData, err := getWieldableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get wieldable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(wieldableComponentData)

	hashmap := make([]ComponentIndexData, wieldableType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WieldableComponent, wieldableType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WieldableComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
