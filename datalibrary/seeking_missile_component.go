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

type SeekingMissileDeviationSettings struct {
	Enabled             uint8 // [bool] Not 100% sure this is correct, but it matches the name length and would make sense
	_                   [3]uint8
	MinSeekingDeviation float32    // Not sure, name length 21
	MaxSeekingDeviation float32    // Not sure, name length 21
	UnkFloat3           float32    // Not sure, name length 21
	UnkFloat4           float32    // Not sure, name length 21
	UnkVector           mgl32.Vec3 // Name length 27
	UnkVector2          mgl32.Vec3 // Name length 27
	UnkBool             uint8      // Name length 38
	_                   [3]uint8
	UnkFloat5           float32 // Name length 29
	UnkFloat6           float32 // Name length 21
	UnkFloat7           float32 // Name length 18
	UnkFloat8           float32 // Name length 22
	UnkFloat9           float32 // Name length 39
}

type SeekingMissileComponent struct {
	TargetingMode                               enum.SeekingMissileTargetingMode // Determines which targeting mode is used.
	UnknownFloat                                float32                          // name length 19, maybe targeting_mode_time?
	AutoplayStartingEffects                     uint8                            // [bool]If disabled, the missile wont play the audio and trail FX.
	_                                           [3]uint8
	TimeToEnableGuidance                        float32           // Time that it takes to enable guidance, 0 means will be enable from the start, -1 means that we should enable it manually.
	TimeToDisableGuidance                       float32           // Time that it takes to disable guidance, 0 means will be disable from the start, -1 means that we should disable it manually.
	UnknownFloat2                               float32           // name length 27
	UnknownFloat3                               float32           // name length 28
	AngleLostGuidance                           float32           // [0,180] - If the angle between the target and the missile exceeds this, disable guidance
	MovementPredictionAccuracy                  float32           // [0,1] - How much the missile is predicting the point of impact depending on current target velocity (1 is best prediction)
	TimeToEnableMovement                        float32           // Time that it takes to enable movement, 0 means will be enable from the start, -1 means that we should enable it manually.
	TimeToEnableExplosive                       float32           // Time that it takes to enable explosive, 0 means will be enable from the start, -1 means that we should enable it manually.
	MaxLifetime                                 float32           // Time before the missile blows itself up. Set to 0 or lower to disable and have the missile live forever.
	StartingSpeed                               float32           // The speed at which the missile launches at.
	MinimumSpeed                                float32           // The minimum speed at which this missile can travel.
	PreferredSpeed                              float32           // The target speed at which this missile wants to travel.
	Acceleration                                float32           // The missile's speed acceleration.
	MaxAngleToTarget                            float32           // The missile will be at its slowest but turn the fastest at or above this angle to the target.
	MinTurnSpeed                                float32           // How fast this missile can turn while going at its preferred speed.
	MaxTurnSpeed                                float32           // How fast this missile can turn while going at its slowest speed.
	PFactor                                     float32           // How much of the difference in current and target speed to consider when changing speed.
	IFactor                                     float32           // How much of the speed difference over time to consider.
	DFactor                                     float32           // How much of the speed delta to consider.
	TargetUpdateInterval                        float32           // How often do we update our target?
	MissileTrail                                EffectSetting     // Particle effect trail that gets attached to the missile.
	MissileTrailAudioEvent                      stingray.ThinHash // [wwise]What audio event to play when the missile is fired.
	MissileTrailAudioEventStop                  stingray.ThinHash // [wwise]What audio event to play when the missile has exploded.
	TargetDotMinimum                            float32           // What's the minimum dot product between the current forward of this entity and its target position before we can update the position? -1 to disable.
	JavelinFiringMode                           uint8             // [bool]When enabled, the missile that is fired will curve up and then down.
	MovementShouldBeProcessedByProjectileSystem uint8             // [bool]If set, the projectile system will handle collision/explosions/damages/vfx/sfx getting that data from the projectile setting.
	_                                           [2]uint8
	ProjectileTypeToProcess                     enum.ProjectileType             // The projectile type that it should process by the projectile manager, if it set to none it should be passed through the spawn context.
	Deviation                                   SeekingMissileDeviationSettings // A struct containing info about missile deviations
	UnknownFloat4                               float32                         // Name length 31
	UnknownHash                                 stingray.ThinHash               // Name length 28
	UnknownBool                                 uint8                           // Name length 23
	UnknownBool2                                uint8                           // Name length 28
	_                                           [2]uint8
	UnknownFloat5                               float32 // Name length 35
	UnknownFloat6                               float32 // Name length 35
	_                                           [4]uint8
}

type SimpleSeekingMissileDeviationSettings struct {
	Enabled             bool       `json:"enabled"`               // [bool] Not 100% sure this is correct, but it matches the name length and would make sense
	MinSeekingDeviation float32    `json:"min_seeking_deviation"` // Not sure, name length 21
	MaxSeekingDeviation float32    `json:"max_seeking_deviation"` // Not sure, name length 21
	UnkFloat3           float32    `json:"unk_float3"`            // Not sure, name length 21
	UnkFloat4           float32    `json:"unk_float4"`            // Not sure, name length 21
	UnkVector           mgl32.Vec3 `json:"unk_vector"`            // Name length 27
	UnkVector2          mgl32.Vec3 `json:"unk_vector2"`           // Name length 27
	UnkBool             bool       `json:"unk_bool"`              // Name length 38
	UnkFloat5           float32    `json:"unk_float5"`            // Name length 29
	UnkFloat6           float32    `json:"unk_float6"`            // Name length 21
	UnkFloat7           float32    `json:"unk_float7"`            // Name length 18
	UnkFloat8           float32    `json:"unk_float8"`            // Name length 22
	UnkFloat9           float32    `json:"unk_float9"`            // Name length 39
}

type SimpleSeekingMissileComponent struct {
	TargetingMode                               enum.SeekingMissileTargetingMode      `json:"targeting_mode"`                                    // Determines which targeting mode is used.
	UnknownFloat                                float32                               `json:"unknown_float"`                                     // name length 19, maybe targeting_mode_time?
	AutoplayStartingEffects                     bool                                  `json:"autoplay_starting_effects"`                         // [bool]If disabled, the missile wont play the audio and trail FX.
	TimeToEnableGuidance                        float32                               `json:"time_to_enable_guidance"`                           // Time that it takes to enable guidance, 0 means will be enable from the start, -1 means that we should enable it manually.
	TimeToDisableGuidance                       float32                               `json:"time_to_disable_guidance"`                          // Time that it takes to disable guidance, 0 means will be disable from the start, -1 means that we should disable it manually.
	UnknownFloat2                               float32                               `json:"unknown_float2"`                                    // name length 27
	UnknownFloat3                               float32                               `json:"unknown_float3"`                                    // name length 28
	AngleLostGuidance                           float32                               `json:"angle_lost_guidance"`                               // [0,180] - If the angle between the target and the missile exceeds this, disable guidance
	MovementPredictionAccuracy                  float32                               `json:"movement_prediction_accuracy"`                      // [0,1] - How much the missile is predicting the point of impact depending on current target velocity (1 is best prediction)
	TimeToEnableMovement                        float32                               `json:"time_to_enable_movement"`                           // Time that it takes to enable movement, 0 means will be enable from the start, -1 means that we should enable it manually.
	TimeToEnableExplosive                       float32                               `json:"time_to_enable_explosive"`                          // Time that it takes to enable explosive, 0 means will be enable from the start, -1 means that we should enable it manually.
	MaxLifetime                                 float32                               `json:"max_lifetime"`                                      // Time before the missile blows itself up. Set to 0 or lower to disable and have the missile live forever.
	StartingSpeed                               float32                               `json:"starting_speed"`                                    // The speed at which the missile launches at.
	MinimumSpeed                                float32                               `json:"minimum_speed"`                                     // The minimum speed at which this missile can travel.
	PreferredSpeed                              float32                               `json:"preferred_speed"`                                   // The target speed at which this missile wants to travel.
	Acceleration                                float32                               `json:"acceleration"`                                      // The missile's speed acceleration.
	MaxAngleToTarget                            float32                               `json:"max_angle_to_target"`                               // The missile will be at its slowest but turn the fastest at or above this angle to the target.
	MinTurnSpeed                                float32                               `json:"min_turn_speed"`                                    // How fast this missile can turn while going at its preferred speed.
	MaxTurnSpeed                                float32                               `json:"max_turn_speed"`                                    // How fast this missile can turn while going at its slowest speed.
	PFactor                                     float32                               `json:"p_factor"`                                          // How much of the difference in current and target speed to consider when changing speed.
	IFactor                                     float32                               `json:"i_factor"`                                          // How much of the speed difference over time to consider.
	DFactor                                     float32                               `json:"d_factor"`                                          // How much of the speed delta to consider.
	TargetUpdateInterval                        float32                               `json:"target_update_interval"`                            // How often do we update our target?
	MissileTrail                                SimpleEffectSetting                   `json:"missile_trail"`                                     // Particle effect trail that gets attached to the missile.
	MissileTrailAudioEvent                      string                                `json:"missile_trail_audio_event"`                         // [wwise]What audio event to play when the missile is fired.
	MissileTrailAudioEventStop                  string                                `json:"missile_trail_audio_event_stop"`                    // [wwise]What audio event to play when the missile has exploded.
	TargetDotMinimum                            float32                               `json:"target_dot_minimum"`                                // What's the minimum dot product between the current forward of this entity and its target position before we can update the position? -1 to disable.
	JavelinFiringMode                           bool                                  `json:"javelin_firing_mode"`                               // [bool]When enabled, the missile that is fired will curve up and then down.
	MovementShouldBeProcessedByProjectileSystem bool                                  `json:"movement_should_be_processed_by_projectile_system"` // [bool]If set, the projectile system will handle collision/explosions/damages/vfx/sfx getting that data from the projectile setting.
	ProjectileTypeToProcess                     enum.ProjectileType                   `json:"projectile_type_to_process"`                        // The projectile type that it should process by the projectile manager, if it set to none it should be passed through the spawn context.
	Deviation                                   SimpleSeekingMissileDeviationSettings `json:"deviation"`                                         // A struct containing info about missile deviations
	UnknownFloat4                               float32                               `json:"unknown_float4"`                                    // Name length 31
	UnknownHash                                 string                                `json:"unknown_hash"`                                      // Name length 28
	UnknownBool                                 bool                                  `json:"unknown_bool"`                                      // Name length 23
	UnknownBool2                                bool                                  `json:"unknown_bool2"`                                     // Name length 28
	UnknownFloat5                               float32                               `json:"unknown_float5"`                                    // Name length 35
	UnknownFloat6                               float32                               `json:"unknown_float6"`                                    // Name length 35
}

func (w SeekingMissileComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleSeekingMissileComponent{
		TargetingMode:              w.TargetingMode,
		UnknownFloat:               w.UnknownFloat,
		AutoplayStartingEffects:    w.AutoplayStartingEffects != 0,
		TimeToEnableGuidance:       w.TimeToEnableGuidance,
		TimeToDisableGuidance:      w.TimeToDisableGuidance,
		UnknownFloat2:              w.UnknownFloat2,
		UnknownFloat3:              w.UnknownFloat3,
		AngleLostGuidance:          w.AngleLostGuidance,
		MovementPredictionAccuracy: w.MovementPredictionAccuracy,
		TimeToEnableMovement:       w.TimeToEnableMovement,
		TimeToEnableExplosive:      w.TimeToEnableExplosive,
		MaxLifetime:                w.MaxLifetime,
		StartingSpeed:              w.StartingSpeed,
		MinimumSpeed:               w.MinimumSpeed,
		PreferredSpeed:             w.PreferredSpeed,
		Acceleration:               w.Acceleration,
		MaxAngleToTarget:           w.MaxAngleToTarget,
		MinTurnSpeed:               w.MinTurnSpeed,
		MaxTurnSpeed:               w.MaxTurnSpeed,
		PFactor:                    w.PFactor,
		IFactor:                    w.IFactor,
		DFactor:                    w.DFactor,
		TargetUpdateInterval:       w.TargetUpdateInterval,
		MissileTrail:               w.MissileTrail.ToSimple(lookupHash, lookupThinHash),
		MissileTrailAudioEvent:     lookupThinHash(w.MissileTrailAudioEvent),
		MissileTrailAudioEventStop: lookupThinHash(w.MissileTrailAudioEventStop),
		TargetDotMinimum:           w.TargetDotMinimum,
		JavelinFiringMode:          w.JavelinFiringMode != 0,
		MovementShouldBeProcessedByProjectileSystem: w.MovementShouldBeProcessedByProjectileSystem != 0,
		ProjectileTypeToProcess:                     w.ProjectileTypeToProcess,
		Deviation: SimpleSeekingMissileDeviationSettings{
			Enabled:             w.Deviation.Enabled != 0,
			MinSeekingDeviation: w.Deviation.MinSeekingDeviation,
			MaxSeekingDeviation: w.Deviation.MaxSeekingDeviation,
			UnkFloat3:           w.Deviation.UnkFloat3,
			UnkFloat4:           w.Deviation.UnkFloat4,
			UnkVector:           w.Deviation.UnkVector,
			UnkVector2:          w.Deviation.UnkVector2,
			UnkBool:             w.Deviation.UnkBool != 0,
			UnkFloat5:           w.Deviation.UnkFloat5,
			UnkFloat6:           w.Deviation.UnkFloat6,
			UnkFloat7:           w.Deviation.UnkFloat7,
			UnkFloat8:           w.Deviation.UnkFloat8,
			UnkFloat9:           w.Deviation.UnkFloat9,
		},
		UnknownFloat4: w.UnknownFloat4,
		UnknownHash:   lookupThinHash(w.UnknownHash),
		UnknownBool:   w.UnknownBool != 0,
		UnknownBool2:  w.UnknownBool2 != 0,
		UnknownFloat5: w.UnknownFloat5,
		UnknownFloat6: w.UnknownFloat6,
	}
}

func getSeekingMissileComponentData() ([]byte, error) {
	seekingMissileComponentHash := Sum("SeekingMissileComponentData")
	seekingMissileComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(seekingMissileComponentHashData, binary.LittleEndian, seekingMissileComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, seekingMissileComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getSeekingMissileComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("SeekingMissileComponentData")
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
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("SeekingMissileComponent") {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (data type was not SeekingMissileComponent)")
	}

	seekingMissileComponentData, err := getSeekingMissileComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get seeking missile component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(seekingMissileComponentData)

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
		return nil, fmt.Errorf("%v not found in seeking missile component data", hash.String())
	}

	var seekingMissileComponentType DLTypeDesc
	seekingMissileComponentType, ok = typelib.Types[Sum("SeekingMissileComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find SeekingMissileComponent hash in dl_library")
	}

	componentData := make([]byte, seekingMissileComponentType.Size)
	if _, err := r.Seek(int64(seekingMissileComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseSeekingMissileComponents() (map[stingray.Hash]SeekingMissileComponent, error) {
	unitHash := Sum("SeekingMissileComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var seekingMissileType DLTypeDesc
	var ok bool
	seekingMissileType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find SeekingMissileComponentData hash in dl_library")
	}

	if len(seekingMissileType.Members) != 2 {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (there should be 2 members but were actually %v)", len(seekingMissileType.Members))
	}

	if seekingMissileType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (hashmap atom was not inline array)")
	}

	if seekingMissileType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (data atom was not inline array)")
	}

	if seekingMissileType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (hashmap storage was not struct)")
	}

	if seekingMissileType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (data storage was not struct)")
	}

	if seekingMissileType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if seekingMissileType.Members[1].TypeID != Sum("SeekingMissileComponent") {
		return nil, fmt.Errorf("SeekingMissileComponentData unexpected format (data type was not SeekingMissileComponent)")
	}

	seekingMissileComponentData, err := getSeekingMissileComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get seeking missile component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(seekingMissileComponentData)

	hashmap := make([]ComponentIndexData, seekingMissileType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]SeekingMissileComponent, seekingMissileType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]SeekingMissileComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
