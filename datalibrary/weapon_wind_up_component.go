package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type WeaponWindUpComponent struct {
	WindUpTime    float32           // Time until the weapon fully winds up.
	WindDownTime  float32           // Time until the weapon fully winds down.
	RpmAnimId     stingray.ThinHash // [string]What the animation variable is for rotating the barrel.
	RpmMultiplier float32           // The RPM spin is decided by the weapon RPM, this would act as a multiplier on top.
	WindUpSound   stingray.ThinHash // [wwise]Sound to play when you start winding up.
	WindDownSound stingray.ThinHash // [wwise]Sound to play when you start winding down.
	WindUpAnim    stingray.ThinHash // [string]Animation to play when you start winding up.
	WindDownAnim  stingray.ThinHash // [string]Animation to play when you start winding down.
	UnknownBool   uint8             // Unknown, name length 19
	_             [3]uint8
}

type SimpleWeaponWindUpComponent struct {
	WindUpTime    float32 `json:"wind_up_time"`
	WindDownTime  float32 `json:"wind_down_time"`
	RpmAnimId     string  `json:"rpm_anim_id"`
	RpmMultiplier float32 `json:"rpm_multiplier"`
	WindUpSound   string  `json:"wind_up_sound"`
	WindDownSound string  `json:"wind_down_sound"`
	WindUpAnim    string  `json:"wind_up_anim"`
	WindDownAnim  string  `json:"wind_down_anim"`
	UnknownBool   bool    `json:"unknown_bool"`
}

func (w WeaponWindUpComponent) ToSimple(_ HashLookup, lookupThinHash ThinHashLookup, _ StringsLookup) any {
	return SimpleWeaponWindUpComponent{
		WindUpTime:    w.WindUpTime,
		WindDownTime:  w.WindDownTime,
		RpmAnimId:     lookupThinHash(w.RpmAnimId),
		RpmMultiplier: w.RpmMultiplier,
		WindUpSound:   lookupThinHash(w.WindUpSound),
		WindDownSound: lookupThinHash(w.WindDownSound),
		WindUpAnim:    lookupThinHash(w.WindUpAnim),
		WindDownAnim:  lookupThinHash(w.WindDownAnim),
		UnknownBool:   w.UnknownBool != 0,
	}
}

func getWeaponWindUpComponentData() ([]byte, error) {
	weaponWindUpHash := Sum("WeaponWindUpComponentData")
	weaponWindUpHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponWindUpHashData, binary.LittleEndian, weaponWindUpHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponWindUpHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponWindUpComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponWindUpCmpDataHash := Sum("WeaponWindUpComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponWindUpCmpDataType DLTypeDesc
	var ok bool
	weaponWindUpCmpDataType, ok = typelib.Types[WeaponWindUpCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(weaponWindUpCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponWindUpCmpDataType.Members))
	}

	if weaponWindUpCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponWindUpCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (data atom was not inline array)")
	}

	if weaponWindUpCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponWindUpCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (data storage was not struct)")
	}

	if weaponWindUpCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponWindUpCmpDataType.Members[1].TypeID != Sum("WeaponWindUpComponent") {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (data type was not WeaponWindUpComponent)")
	}

	weaponWindUpComponentData, err := getWeaponWindUpComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon wind up component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponWindUpComponentData)

	hashmap := make([]ComponentIndexData, weaponWindUpCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon wind up component data", hash.String())
	}

	var weaponWindUpComponentType DLTypeDesc
	weaponWindUpComponentType, ok = typelib.Types[Sum("WeaponWindUpComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponWindUpComponent hash in dl_library")
	}

	componentData := make([]byte, weaponWindUpComponentType.Size)
	if _, err := r.Seek(int64(weaponWindUpComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponWindUpComponents() (map[stingray.Hash]WeaponWindUpComponent, error) {
	weaponWindUpHash := Sum("WeaponWindUpComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponWindUpType DLTypeDesc
	var ok bool
	weaponWindUpType, ok = typelib.Types[weaponWindUpHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponWindUpComponentData hash in dl_library")
	}

	if len(weaponWindUpType.Members) != 2 {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponWindUpType.Members))
	}

	if weaponWindUpType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponWindUpType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (data atom was not inline array)")
	}

	if weaponWindUpType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponWindUpType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (data storage was not struct)")
	}

	if weaponWindUpType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponWindUpType.Members[1].TypeID != Sum("WeaponWindUpComponent") {
		return nil, fmt.Errorf("WeaponWindUpComponentData unexpected format (data type was not WeaponWindUpComponent)")
	}

	weaponWindUpComponentData, err := getWeaponWindUpComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponWindUpComponentData)

	hashmap := make([]ComponentIndexData, weaponWindUpType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponWindUpComponent, weaponWindUpType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponWindUpComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
