package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

// DLHash 8f124598 name length 16
type HeatsinkOverrides struct {
	HeatCapacity float32 // No clue really
	ProjType     enum.ProjectileType
	Particles    stingray.Hash
	MaybeEvent   stingray.ThinHash
	StatusEffect enum.StatusEffectType
}

type WeaponHeatComponent struct {
	HSOverrides                   [3]HeatsinkOverrides
	UnknownFloat                  float32
	UnknownFloat2                 float32
	UnknownBool                   uint8
	_                             [3]uint8
	Magazines                     uint32  // Starting number of magazines.
	MagazinesRefill               uint32  // Number of magazines given on refill.
	MagazinesMax                  uint32  // Maximum number of magazines.
	OverheatTemperature           float32 // Temperature at which the weapon overheats.
	UnknownFloat3                 float32 // Name length 28 chars
	WarningTemperature            float32 // Temperature at which the weapon cues an audible warning.
	EmissionTemperature           float32 // Temperature at which the weapon begins emitting smoke/steam particles.
	RTPCTemperature               float32 // Temperature at which the weapon starts altering the laser RTPC.
	TempGainPerShot               float32 // Amount of temperature added to the weapon per shot when firing.
	TempGainPerSecond             float32 // Amount of temperature added to the weapon per second when firing.
	TempGainPerSecondModifier     float32 // Multiplier on .the base temperature gain used by customizations.
	TempLossPerSecond             float32 // Amount of temperature removed from the weapon per second when not firing.
	TempLossPerSecondOverheated   float32 // Amount of temperature removed from the weapon per second after it's overheated.
	UnknownFloat4                 float32 // idk, name length should be 31 chars
	UnknownFloat5                 float32 // idk, name length should be 31 chars
	NeedsReloadAfterOverheat      uint8   // [bool]Should the weapon require a new magazine after an overheat.
	_                             [3]uint8
	FiringCharge                  float32 // Charge required to fire.
	ChargeGainPerSecond           float32 // Amount of charge added to the weapon per second while holding the trigger.
	ChargeLossPerSecond           float32 // Amount of charge removed from the weapon per second while not holding the trigger.
	ResetChargeAfterShot          uint8   // [bool]Should set the charge to 0 after each shot?.
	SomeOtherBool                 uint8
	_                             [2]uint8
	OwnerWarningVOEvent           stingray.ThinHash // [string]The VO Event to play on owner when the beam weapon reaches warning level.
	OwnerOverheatVOEvent          stingray.ThinHash // [string]The VO Event to play on owner when the beam weapon gets overheated.
	MinReloadTemperature          float32           // How much temperature do we need before we can reload
	ChargingStartAudioEvent       stingray.ThinHash // [wwise]The audio event to play on start charging
	ChargingCompleteAudioEvent    stingray.ThinHash // [wwise]The audio event to play when charging complete.
	DischargingStartAudioEvent    stingray.ThinHash // [wwise]The audio event to play on start discharging
	DischargingCompleteAudioEvent stingray.ThinHash // [wwise]The audio event to play when discharging complete.
	ChargeSourceNode              stingray.ThinHash // [string]The node to play the charge/discharge audio events at.
	OnOverheatStartSoundEvent     stingray.ThinHash // [wwise]The wwise sound id to play when overheating starts.
	OnOverheatStopSoundEvent      stingray.ThinHash // [wwise]The wwise sound id to play when overheating stops.
	OnTempYellowStartSoundEvent   stingray.ThinHash // [wwise]The wwise sound id to play when yellow temperature starts.
	UnknownSoundEvent             stingray.ThinHash // Its probably an intermediate sound effect. Length 18 chars
	OnTempYellowStopSoundEvent    stingray.ThinHash // [wwise]The wwise sound id to play when yellow temperature stops.
	UnknownSoundEvent2            stingray.ThinHash
	UnknownSoundEvent3            stingray.ThinHash
	UnknownSoundEvent4            stingray.ThinHash
	UnknownSoundEvent5            stingray.ThinHash
	UnknownSoundEvent6            stingray.ThinHash
	UnknownSoundEvent7            stingray.ThinHash
	CameraShakeOnChargingStart    stingray.Hash  // [camera_shake]The shake effect to play when the laser weapon starts charging. This ends when the charging is complete.
	CameraShakeOnFiringStart      stingray.Hash  // [camera_shake]The shake effect to play when the laser weapon starts discharging. This ends when the discharging is complete.
	CameraShakeOnYellowTempStart  stingray.Hash  // [camera_shake]The shake effect to play when the laser weapon enters yellow temperature. This ends when it exits out of yellow temperature.
	CameraShakeOnRedTempStart     stingray.Hash  // [camera_shake]The shake effect to play when the laser weapon enters red temperature. This ends when it exits out of red temperature.
	CameraShakeOnFireStopConstant stingray.Hash  // [camera_shake]The exiting constant effect for when you stop firing the weapon.
	CameraShakeOnFireStopKick     stingray.Hash  // [camera_shake]The exiting kick effect for when you stop firing the weapon.
	OverheatAbility               enum.AbilityId // The ability to play when this weapon overheats.
}

func getWeaponHeatComponentData() ([]byte, error) {
	weaponHeatHash := Sum("WeaponHeatComponentData")
	weaponHeatHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponHeatHashData, binary.LittleEndian, weaponHeatHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponHeatHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponHeatComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponHeatCmpDataHash := Sum("WeaponHeatComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponHeatCmpDataType DLTypeDesc
	var ok bool
	weaponHeatCmpDataType, ok = typelib.Types[WeaponHeatCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(weaponHeatCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponHeatCmpDataType.Members))
	}

	if weaponHeatCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponHeatCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (data atom was not inline array)")
	}

	if weaponHeatCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponHeatCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (data storage was not struct)")
	}

	if weaponHeatCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponHeatCmpDataType.Members[1].TypeID != Sum("WeaponHeatComponent") {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (data type was not WeaponHeatComponent)")
	}

	weaponHeatComponentData, err := getWeaponHeatComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon heat component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponHeatComponentData)

	hashmap := make([]ComponentIndexData, weaponHeatCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon heat component data", hash.String())
	}

	var weaponHeatComponentType DLTypeDesc
	weaponHeatComponentType, ok = typelib.Types[Sum("WeaponHeatComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponHeatComponent hash in dl_library")
	}

	componentData := make([]byte, weaponHeatComponentType.Size)
	if _, err := r.Seek(int64(weaponHeatComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponHeatComponents() (map[stingray.Hash]WeaponHeatComponent, error) {
	weaponHeatHash := Sum("WeaponHeatComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponHeatType DLTypeDesc
	var ok bool
	weaponHeatType, ok = typelib.Types[weaponHeatHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponHeatComponentData hash in dl_library")
	}

	if len(weaponHeatType.Members) != 2 {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponHeatType.Members))
	}

	if weaponHeatType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponHeatType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (data atom was not inline array)")
	}

	if weaponHeatType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponHeatType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (data storage was not struct)")
	}

	if weaponHeatType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponHeatType.Members[1].TypeID != Sum("WeaponHeatComponent") {
		return nil, fmt.Errorf("WeaponHeatComponentData unexpected format (data type was not WeaponHeatComponent)")
	}

	weaponHeatComponentData, err := getWeaponHeatComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponHeatComponentData)

	hashmap := make([]ComponentIndexData, weaponHeatType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponHeatComponent, weaponHeatType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponHeatComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
