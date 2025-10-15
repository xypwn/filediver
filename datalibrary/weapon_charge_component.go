package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type ChargeStateSetting struct {
	ChargeTime         float32             `json:"charge_time"`
	ProjType           enum.ProjectileType `json:"projectile_type"`
	ProjectileParticle stingray.Hash       `json:"projectile_particle"`
	UnknownThinHash    stingray.ThinHash   `json:"unknown"` // 17 chars long name
	_                  [4]uint8
}

type ProjectileMultipliers struct {
	SpeedMultiplierMin              float32 `json:"speed_multiplier_min"`              // The value to multiply the projectile speed with when at the smallest charge amount.
	SpeedMultiplierOvercharge       float32 `json:"speed_multiplier_overcharge"`       // The value to multiply the projectile speed with when fully overcharged.
	DamageMultiplierMin             float32 `json:"damage_multiplier_min"`             // The value to multiply the projectile damage with when at the smallest charge amount.
	DamageMultiplierOvercharge      float32 `json:"damage_multiplier_overcharge"`      // The value to multiply the projectile damage with when fully overcharged.
	PenetrationMultiplierMin        float32 `json:"penetration_multiplier_min"`        // The value to multiply the projectile penetration with when at the smallest charge amount.
	PenetrationMultiplierOvercharge float32 `json:"penetration_multiplier_overcharge"` // The value to multiply the projectile penetration with when fully overcharged.
	DistanceMultiplerMin            float32 `json:"distance_multipler_min"`            // The value to multiply the arc distance with when at the smallest charge amount.
	DistanceMultiplerOvercharge     float32 `json:"distance_multipler_overcharge"`     // The value to multiply the arc distance with when fully overcharged.
	ExtraArcSplitsMin               float32 `json:"extra_arc_splits_min"`              // The amount of extra splits this arc can do with minimum charge amount.
	ExtraArcSplitsOvercharge        float32 `json:"extra_arc_splits_overcharge"`       // The amount of extra splits this arc can do when fully overcharged.
	ExtraArcChainsMin               float32 `json:"extra_arc_chains_min"`              // The amount of extra chains this arc can do with minimum charge amount.
	ExtraArcChainsOvercharge        float32 `json:"extra_arc_chains_overcharge"`       // The amount of extra chains this arc can do when fully overcharged.
}

// Has a charge state and a value, name should be 14 chars long
type UnknownChargeStruct struct {
	State enum.ChargeState `json:"charge_state"`
	Value float32          `json:"value"`
}

type WeaponChargeComponent struct {
	ChargeStateSettings     [3]ChargeStateSetting // Min-, Full-, and Over-Charged states
	ProjMultipliers         ProjectileMultipliers // Multipliers of the setting values for the projectile based on the charge amount.
	ChargeStartSoundID      stingray.ThinHash     // [string]Sound to start playing when the chargeup starts.
	ChargeStopSoundID       stingray.ThinHash     // [string]Sound id to play when the chargeup ends.
	ReadyToFireSoundID      stingray.ThinHash     // [string]Sound id to play when the weapon can fire.
	DangerOverchargeSoundID stingray.ThinHash     // [string]Sound id to play when the chargeup enters the danger zone.
	ChargeMesh              stingray.ThinHash     // [string]Mesh to set charge value on.
	ChargeMaterial          stingray.ThinHash     // [string]Material to set charge value on.
	ChargeVariable          stingray.ThinHash     // [string]Material variable to set charge value on.
	_                       [4]uint8
	ChargeUpMuzzleFlash     stingray.Hash     // [particles]Particle effect of the muzzle flash while charging.
	ChargeUpMuzzleFlashLoop stingray.Hash     // [particles]Looping particle effect of the muzzle flash while charging.
	ChargeAnimID            stingray.ThinHash // [string]What the animation variable is for rotating the barrel.
	ChargeEndAnimID         stingray.ThinHash // [string]What the animation variable is for rotating the barrel.
	ChargeRateAnimID        stingray.ThinHash // [string]What the animation variable is for rotating the barrel.
	SpinSpeedAnimID         stingray.ThinHash // [string]What the animation variable is for rotating the barrel.
	AutoFireInSafety        uint8             // [bool]If disabled, will allow the user to keep the charge as long as they are holding the trigger.
	ExplodesOnOvercharged   uint8             // Unknown bool, name length 24 chars
	_                       [2]uint8
	ExplosionAudioEvent     stingray.ThinHash // Unknown, name length 22 chars
	UnknownFloat            float32           // Unknown, probably related to the above
	DryFireAudioEvent       stingray.ThinHash // [string].
	ExplodeType             enum.ExplosionType
	StateValue              UnknownChargeStruct
	_                       [4]uint8
}

type SimpleWeaponChargeComponent struct {
	ChargeStateSettings     [3]ChargeStateSetting `json:"charge_state_settings"`       // Min-, Full-, and Over-Charged states
	ProjMultipliers         ProjectileMultipliers `json:"proj_multipliers"`            // Multipliers of the setting values for the projectile based on the charge amount.
	ChargeStartSoundID      string                `json:"charge_start_sound_id"`       // [string]Sound to start playing when the chargeup starts.
	ChargeStopSoundID       string                `json:"charge_stop_sound_id"`        // [string]Sound id to play when the chargeup ends.
	ReadyToFireSoundID      string                `json:"ready_to_fire_sound_id"`      // [string]Sound id to play when the weapon can fire.
	DangerOverchargeSoundID string                `json:"danger_overcharge_sound_id"`  // [string]Sound id to play when the chargeup enters the danger zone.
	ChargeMesh              string                `json:"charge_mesh"`                 // [string]Mesh to set charge value on.
	ChargeMaterial          string                `json:"charge_material"`             // [string]Material to set charge value on.
	ChargeVariable          string                `json:"charge_variable"`             // [string]Material variable to set charge value on.
	ChargeUpMuzzleFlash     string                `json:"charge_up_muzzle_flash"`      // [particles]Particle effect of the muzzle flash while charging.
	ChargeUpMuzzleFlashLoop string                `json:"charge_up_muzzle_flash_loop"` // [particles]Looping particle effect of the muzzle flash while charging.
	ChargeAnimID            string                `json:"charge_anim_id"`              // [string]What the animation variable is for rotating the barrel.
	ChargeEndAnimID         string                `json:"charge_end_anim_id"`          // [string]What the animation variable is for rotating the barrel.
	ChargeRateAnimID        string                `json:"charge_rate_anim_id"`         // [string]What the animation variable is for rotating the barrel.
	SpinSpeedAnimID         string                `json:"spin_speed_anim_id"`          // [string]What the animation variable is for rotating the barrel.
	AutoFireInSafety        bool                  `json:"auto_fire_in_safety"`         // [bool]If disabled, will allow the user to keep the charge as long as they are holding the trigger.
	ExplodesOnOvercharged   bool                  `json:"explodes_on_overcharged"`     // Unknown bool, name length 24 chars
	ExplosionAudioEvent     string                `json:"explosion_audio_event"`       // Unknown, name length 22 chars
	UnknownFloat            float32               `json:"unknown_float"`               // Unknown, probably related to the above
	DryFireAudioEvent       string                `json:"dry_fire_audio_event"`        // [string].
	ExplodeType             enum.ExplosionType    `json:"explosion_type"`
	StateValue              UnknownChargeStruct   `json:"state_value"`
}

func (w WeaponChargeComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) any {
	return SimpleWeaponChargeComponent{
		ChargeStateSettings:     w.ChargeStateSettings,
		ProjMultipliers:         w.ProjMultipliers,
		ChargeStartSoundID:      lookupThinHash(w.ChargeStartSoundID),
		ChargeStopSoundID:       lookupThinHash(w.ChargeStopSoundID),
		ReadyToFireSoundID:      lookupThinHash(w.ReadyToFireSoundID),
		DangerOverchargeSoundID: lookupThinHash(w.DangerOverchargeSoundID),
		ChargeMesh:              lookupThinHash(w.ChargeMesh),
		ChargeMaterial:          lookupThinHash(w.ChargeMaterial),
		ChargeVariable:          lookupThinHash(w.ChargeVariable),
		ChargeUpMuzzleFlash:     lookupHash(w.ChargeUpMuzzleFlash),
		ChargeUpMuzzleFlashLoop: lookupHash(w.ChargeUpMuzzleFlashLoop),
		ChargeAnimID:            lookupThinHash(w.ChargeAnimID),
		ChargeEndAnimID:         lookupThinHash(w.ChargeEndAnimID),
		ChargeRateAnimID:        lookupThinHash(w.ChargeRateAnimID),
		SpinSpeedAnimID:         lookupThinHash(w.SpinSpeedAnimID),
		AutoFireInSafety:        w.AutoFireInSafety != 0,
		ExplodesOnOvercharged:   w.ExplodesOnOvercharged != 0,
		ExplosionAudioEvent:     lookupThinHash(w.ExplosionAudioEvent),
		UnknownFloat:            w.UnknownFloat,
		DryFireAudioEvent:       lookupThinHash(w.DryFireAudioEvent),
		ExplodeType:             w.ExplodeType,
		StateValue:              w.StateValue,
	}
}

func getWeaponChargeComponentData() ([]byte, error) {
	weaponChargeHash := Sum("WeaponChargeComponentData")
	weaponChargeHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponChargeHashData, binary.LittleEndian, weaponChargeHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponChargeHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponChargeComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponChargeCmpDataHash := Sum("WeaponChargeComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponChargeCmpDataType DLTypeDesc
	var ok bool
	weaponChargeCmpDataType, ok = typelib.Types[WeaponChargeCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(weaponChargeCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponChargeCmpDataType.Members))
	}

	if weaponChargeCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponChargeCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (data atom was not inline array)")
	}

	if weaponChargeCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponChargeCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (data storage was not struct)")
	}

	if weaponChargeCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponChargeCmpDataType.Members[1].TypeID != Sum("WeaponChargeComponent") {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (data type was not WeaponChargeComponent)")
	}

	weaponChargeComponentData, err := getWeaponChargeComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon charge component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponChargeComponentData)

	hashmap := make([]ComponentIndexData, weaponChargeCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon charge component data", hash.String())
	}

	var weaponChargeComponentType DLTypeDesc
	weaponChargeComponentType, ok = typelib.Types[Sum("WeaponChargeComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponChargeComponent hash in dl_library")
	}

	componentData := make([]byte, weaponChargeComponentType.Size)
	if _, err := r.Seek(int64(weaponChargeComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponChargeComponents() (map[stingray.Hash]WeaponChargeComponent, error) {
	weaponChargeHash := Sum("WeaponChargeComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponChargeType DLTypeDesc
	var ok bool
	weaponChargeType, ok = typelib.Types[weaponChargeHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponChargeComponentData hash in dl_library")
	}

	if len(weaponChargeType.Members) != 2 {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponChargeType.Members))
	}

	if weaponChargeType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponChargeType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (data atom was not inline array)")
	}

	if weaponChargeType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponChargeType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (data storage was not struct)")
	}

	if weaponChargeType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponChargeType.Members[1].TypeID != Sum("WeaponChargeComponent") {
		return nil, fmt.Errorf("WeaponChargeComponentData unexpected format (data type was not WeaponChargeComponent)")
	}

	weaponChargeComponentData, err := getWeaponChargeComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponChargeComponentData)

	hashmap := make([]ComponentIndexData, weaponChargeType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponChargeComponent, weaponChargeType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponChargeComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
