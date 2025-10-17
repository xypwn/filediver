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

type SimpleHeatsinkOverrides struct {
	HeatCapacity float32               `json:"heat_capacity"` // No clue really
	ProjType     enum.ProjectileType   `json:"projectile_type"`
	Particles    string                `json:"particles"`
	MaybeEvent   string                `json:"maybe_event"`
	StatusEffect enum.StatusEffectType `json:"status_effect"`
}

type SimpleWeaponHeatComponent struct {
	HSOverrides                   []SimpleHeatsinkOverrides `json:"heatsink_overrides,omitempty"`
	UnknownFloat                  float32                   `json:"unknown_float"`
	UnknownFloat2                 float32                   `json:"unknown_float2"`
	UnknownBool                   bool                      `json:"unknown_bool"`
	Magazines                     uint32                    `json:"magazines"`                       // Starting number of magazines.
	MagazinesRefill               uint32                    `json:"magazines_refill"`                // Number of magazines given on refill.
	MagazinesMax                  uint32                    `json:"magazines_max"`                   // Maximum number of magazines.
	OverheatTemperature           float32                   `json:"overheat_temperature"`            // Temperature at which the weapon overheats.
	UnknownFloat3                 float32                   `json:"unknown_float3"`                  // Name length 28 chars
	WarningTemperature            float32                   `json:"warning_temperature"`             // Temperature at which the weapon cues an audible warning.
	EmissionTemperature           float32                   `json:"emission_temperature"`            // Temperature at which the weapon begins emitting smoke/steam particles.
	RTPCTemperature               float32                   `json:"rtpc_temperature"`                // Temperature at which the weapon starts altering the laser RTPC.
	TempGainPerShot               float32                   `json:"temp_gain_per_shot"`              // Amount of temperature added to the weapon per shot when firing.
	TempGainPerSecond             float32                   `json:"temp_gain_per_second"`            // Amount of temperature added to the weapon per second when firing.
	TempGainPerSecondModifier     float32                   `json:"temp_gain_per_second_modifier"`   // Multiplier on .the base temperature gain used by customizations.
	TempLossPerSecond             float32                   `json:"temp_loss_per_second"`            // Amount of temperature removed from the weapon per second when not firing.
	TempLossPerSecondOverheated   float32                   `json:"temp_loss_per_second_overheated"` // Amount of temperature removed from the weapon per second after it's overheated.
	UnknownFloat4                 float32                   `json:"unknown_float4"`                  // idk, name length should be 31 chars
	UnknownFloat5                 float32                   `json:"unknown_float5"`                  // idk, name length should be 31 chars
	NeedsReloadAfterOverheat      bool                      `json:"needs_reload_after_overheat"`     // [bool]Should the weapon require a new magazine after an overheat.
	FiringCharge                  float32                   `json:"firing_charge"`                   // Charge required to fire.
	ChargeGainPerSecond           float32                   `json:"charge_gain_per_second"`          // Amount of charge added to the weapon per second while holding the trigger.
	ChargeLossPerSecond           float32                   `json:"charge_loss_per_second"`          // Amount of charge removed from the weapon per second while not holding the trigger.
	ResetChargeAfterShot          bool                      `json:"reset_charge_after_shot"`         // [bool]Should set the charge to 0 after each shot?.
	SomeOtherBool                 bool                      `json:"some_other_bool"`
	OwnerWarningVOEvent           string                    `json:"owner_warning_vo_event"`           // [string]The VO Event to play on owner when the beam weapon reaches warning level.
	OwnerOverheatVOEvent          string                    `json:"owner_overheat_vo_event"`          // [string]The VO Event to play on owner when the beam weapon gets overheated.
	MinReloadTemperature          float32                   `json:"min_reload_temperature"`           // How much temperature do we need before we can reload
	ChargingStartAudioEvent       string                    `json:"charging_start_audio_event"`       // [wwise]The audio event to play on start charging
	ChargingCompleteAudioEvent    string                    `json:"charging_complete_audio_event"`    // [wwise]The audio event to play when charging complete.
	DischargingStartAudioEvent    string                    `json:"discharging_start_audio_event"`    // [wwise]The audio event to play on start discharging
	DischargingCompleteAudioEvent string                    `json:"discharging_complete_audio_event"` // [wwise]The audio event to play when discharging complete.
	ChargeSourceNode              string                    `json:"charge_source_node"`               // [string]The node to play the charge/discharge audio events at.
	OnOverheatStartSoundEvent     string                    `json:"on_overheat_start_sound_event"`    // [wwise]The wwise sound id to play when overheating starts.
	OnOverheatStopSoundEvent      string                    `json:"on_overheat_stop_sound_event"`     // [wwise]The wwise sound id to play when overheating stops.
	OnTempYellowStartSoundEvent   string                    `json:"on_temp_yellow_start_sound_event"` // [wwise]The wwise sound id to play when yellow temperature starts.
	UnknownSoundEvent             string                    `json:"unknown_sound_event"`              // Its probably an intermediate sound effect. Length 18 chars
	OnTempYellowStopSoundEvent    string                    `json:"on_temp_yellow_stop_sound_event"`  // [wwise]The wwise sound id to play when yellow temperature stops.
	UnknownSoundEvent2            string                    `json:"unknown_sound_event2"`
	UnknownSoundEvent3            string                    `json:"unknown_sound_event3"`
	UnknownSoundEvent4            string                    `json:"unknown_sound_event4"`
	UnknownSoundEvent5            string                    `json:"unknown_sound_event5"`
	UnknownSoundEvent6            string                    `json:"unknown_sound_event6"`
	UnknownSoundEvent7            string                    `json:"unknown_sound_event7"`
	CameraShakeOnChargingStart    string                    `json:"camera_shake_on_charging_start"`     // [camera_shake]The shake effect to play when the laser weapon starts charging. This ends when the charging is complete.
	CameraShakeOnFiringStart      string                    `json:"camera_shake_on_firing_start"`       // [camera_shake]The shake effect to play when the laser weapon starts discharging. This ends when the discharging is complete.
	CameraShakeOnYellowTempStart  string                    `json:"camera_shake_on_yellow_temp_start"`  // [camera_shake]The shake effect to play when the laser weapon enters yellow temperature. This ends when it exits out of yellow temperature.
	CameraShakeOnRedTempStart     string                    `json:"camera_shake_on_red_temp_start"`     // [camera_shake]The shake effect to play when the laser weapon enters red temperature. This ends when it exits out of red temperature.
	CameraShakeOnFireStopConstant string                    `json:"camera_shake_on_fire_stop_constant"` // [camera_shake]The exiting constant effect for when you stop firing the weapon.
	CameraShakeOnFireStopKick     string                    `json:"camera_shake_on_fire_stop_kick"`     // [camera_shake]The exiting kick effect for when you stop firing the weapon.
	OverheatAbility               enum.AbilityId            `json:"overheat_ability"`                   // The ability to play when this weapon overheats.
}

func (h WeaponHeatComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	heatsinkOverrides := make([]SimpleHeatsinkOverrides, 0)
	for _, override := range h.HSOverrides {
		if override.Particles.Value == 0 {
			break
		}
		heatsinkOverrides = append(heatsinkOverrides, SimpleHeatsinkOverrides{
			HeatCapacity: override.HeatCapacity,
			ProjType:     override.ProjType,
			Particles:    lookupHash(override.Particles),
			MaybeEvent:   lookupThinHash(override.MaybeEvent),
			StatusEffect: override.StatusEffect,
		})
	}

	return SimpleWeaponHeatComponent{
		HSOverrides:                   heatsinkOverrides,
		UnknownFloat:                  h.UnknownFloat,
		UnknownFloat2:                 h.UnknownFloat2,
		UnknownBool:                   h.UnknownBool != 0,
		Magazines:                     h.Magazines,
		MagazinesRefill:               h.MagazinesRefill,
		MagazinesMax:                  h.MagazinesMax,
		OverheatTemperature:           h.OverheatTemperature,
		UnknownFloat3:                 h.UnknownFloat3,
		WarningTemperature:            h.WarningTemperature,
		EmissionTemperature:           h.EmissionTemperature,
		RTPCTemperature:               h.RTPCTemperature,
		TempGainPerShot:               h.TempGainPerShot,
		TempGainPerSecond:             h.TempGainPerSecond,
		TempGainPerSecondModifier:     h.TempGainPerSecondModifier,
		TempLossPerSecond:             h.TempLossPerSecond,
		TempLossPerSecondOverheated:   h.TempLossPerSecondOverheated,
		UnknownFloat4:                 h.UnknownFloat4,
		UnknownFloat5:                 h.UnknownFloat5,
		NeedsReloadAfterOverheat:      h.NeedsReloadAfterOverheat != 0,
		FiringCharge:                  h.FiringCharge,
		ChargeGainPerSecond:           h.ChargeGainPerSecond,
		ChargeLossPerSecond:           h.ChargeLossPerSecond,
		ResetChargeAfterShot:          h.ResetChargeAfterShot != 0,
		SomeOtherBool:                 h.SomeOtherBool != 0,
		OwnerWarningVOEvent:           lookupThinHash(h.OwnerWarningVOEvent),
		OwnerOverheatVOEvent:          lookupThinHash(h.OwnerOverheatVOEvent),
		MinReloadTemperature:          h.MinReloadTemperature,
		ChargingStartAudioEvent:       lookupThinHash(h.ChargingStartAudioEvent),
		ChargingCompleteAudioEvent:    lookupThinHash(h.ChargingCompleteAudioEvent),
		DischargingStartAudioEvent:    lookupThinHash(h.DischargingStartAudioEvent),
		DischargingCompleteAudioEvent: lookupThinHash(h.DischargingCompleteAudioEvent),
		ChargeSourceNode:              lookupThinHash(h.ChargeSourceNode),
		OnOverheatStartSoundEvent:     lookupThinHash(h.OnOverheatStartSoundEvent),
		OnOverheatStopSoundEvent:      lookupThinHash(h.OnOverheatStopSoundEvent),
		OnTempYellowStartSoundEvent:   lookupThinHash(h.OnTempYellowStartSoundEvent),
		UnknownSoundEvent:             lookupThinHash(h.UnknownSoundEvent),
		OnTempYellowStopSoundEvent:    lookupThinHash(h.OnTempYellowStopSoundEvent),
		UnknownSoundEvent2:            lookupThinHash(h.UnknownSoundEvent2),
		UnknownSoundEvent3:            lookupThinHash(h.UnknownSoundEvent3),
		UnknownSoundEvent4:            lookupThinHash(h.UnknownSoundEvent4),
		UnknownSoundEvent5:            lookupThinHash(h.UnknownSoundEvent5),
		UnknownSoundEvent6:            lookupThinHash(h.UnknownSoundEvent6),
		UnknownSoundEvent7:            lookupThinHash(h.UnknownSoundEvent7),
		CameraShakeOnChargingStart:    lookupHash(h.CameraShakeOnChargingStart),
		CameraShakeOnFiringStart:      lookupHash(h.CameraShakeOnFiringStart),
		CameraShakeOnYellowTempStart:  lookupHash(h.CameraShakeOnYellowTempStart),
		CameraShakeOnRedTempStart:     lookupHash(h.CameraShakeOnRedTempStart),
		CameraShakeOnFireStopConstant: lookupHash(h.CameraShakeOnFireStopConstant),
		CameraShakeOnFireStopKick:     lookupHash(h.CameraShakeOnFireStopKick),
		OverheatAbility:               h.OverheatAbility,
	}
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
