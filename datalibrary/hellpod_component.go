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

type HellpodComponent struct {
	HellpodVariant                  enum.HellpodVariant // What kind of hellpod is this? Used for determining special-case settings.
	UnknownBool                     uint8               // Unknown, 15 chars long
	_                               [3]uint8
	SpawnHeight                     float32    // At what Z level should the hellpod spawn?
	UnknownVec                      mgl32.Vec2 // unknown, 22 chars long
	InitialVelocity                 float32    // What's the initial velocity of the hellpod?
	BreakingHeight                  float32    // How far above ground should we start breaking? [sic]
	BreakingVelocity                float32    // What's the breaking velocity of the hellpod? [sic]
	BreakingVelocityTargetTime      float32    // How long until we reach our breaking velocity? [sic]
	UnknownBool2                    uint8      // unknown, 24 chars long
	_                               [3]uint8
	AudioThrustersIgniteLoop        stingray.ThinHash // [string][wwise]Thruster Ignite Sound Loop
	AudioThrustersIgniteEnd         stingray.ThinHash // [string][wwise]Thruster Ignite Sound End
	AudioThrustersIgnition          stingray.ThinHash // [string][wwise]Thruster Ignite Sound
	AudioEntryWooshLoop             stingray.ThinHash // [string][wwise]Woosh Sound Loop
	AudioEntryWooshEnd              stingray.ThinHash // [string][wwise]Woosh Sound End
	AudioChassiShakeLoop            stingray.ThinHash // [string][wwise]Chassi Shake Sound Loop
	AudioChassiShakeEnd             stingray.ThinHash // [string][wwise]Chassi Shake Sound End
	AudioSupersonicBoom             stingray.ThinHash // [string][wwise]Supersonic Boom Sound
	AudioImpact                     stingray.ThinHash // [string][wwise]Impact Sound
	AudioBoltBlowing                stingray.ThinHash // [string][wwise]Bolt Blowing Sound
	AudioLidBlowing                 stingray.ThinHash // [string][wwise]Lid Blowing Sound
	AudioThrusterStart              stingray.ThinHash // [string][wwise]Thruster Start Sound
	AudioHellpodPreImpactRiser      stingray.ThinHash // [string][wwise]Hellpod Preimpact Riser Sound
	AudioMusicHellpodGroundApproach stingray.ThinHash // [string][wwise]Music Hellpod Ground Approach Switch
	AudioMusicHellpodLand           stingray.ThinHash // [string][wwise]Music Hellpod Land Switch
	UnknownBool3                    uint8             // Unknown, name length 12 chars
	_                               [3]uint8
}

type SimpleHellpodComponent struct {
	HellpodVariant                  enum.HellpodVariant `json:"hellpod_variant"`
	UnknownBool                     bool                `json:"unknown_bool"`
	SpawnHeight                     float32             `json:"spawn_height"`
	UnknownVec                      mgl32.Vec2          `json:"unknown_vec"`
	InitialVelocity                 float32             `json:"initial_velocity"`
	BreakingHeight                  float32             `json:"breaking_height"`
	BreakingVelocity                float32             `json:"breaking_velocity"`
	BreakingVelocityTargetTime      float32             `json:"breaking_velocity_target_time"`
	UnknownBool2                    bool                `json:"unknown_bool2"`
	AudioThrustersIgniteLoop        string              `json:"audio_thrusters_ignite_loop"`
	AudioThrustersIgniteEnd         string              `json:"audio_thrusters_ignite_end"`
	AudioThrustersIgnition          string              `json:"audio_thrusters_ignition"`
	AudioEntryWooshLoop             string              `json:"audio_entry_woosh_loop"`
	AudioEntryWooshEnd              string              `json:"audio_entry_woosh_end"`
	AudioChassiShakeLoop            string              `json:"audio_chassi_shake_loop"`
	AudioChassiShakeEnd             string              `json:"audio_chassi_shake_end"`
	AudioSupersonicBoom             string              `json:"audio_supersonic_boom"`
	AudioImpact                     string              `json:"audio_impact"`
	AudioBoltBlowing                string              `json:"audio_bolt_blowing"`
	AudioLidBlowing                 string              `json:"audio_lid_blowing"`
	AudioThrusterStart              string              `json:"audio_thruster_start"`
	AudioHellpodPreImpactRiser      string              `json:"audio_hellpod_pre_impact_riser"`
	AudioMusicHellpodGroundApproach string              `json:"audio_music_hellpod_ground_approach"`
	AudioMusicHellpodLand           string              `json:"audio_music_hellpod_land"`
	UnknownBool3                    bool                `json:"unknown_bool3"`
}

func (w HellpodComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleHellpodComponent{
		HellpodVariant:                  w.HellpodVariant,
		UnknownBool:                     w.UnknownBool != 0,
		SpawnHeight:                     w.SpawnHeight,
		UnknownVec:                      w.UnknownVec,
		InitialVelocity:                 w.InitialVelocity,
		BreakingHeight:                  w.BreakingHeight,
		BreakingVelocity:                w.BreakingVelocity,
		BreakingVelocityTargetTime:      w.BreakingVelocityTargetTime,
		UnknownBool2:                    w.UnknownBool2 != 0,
		AudioThrustersIgniteLoop:        lookupThinHash(w.AudioThrustersIgniteLoop),
		AudioThrustersIgniteEnd:         lookupThinHash(w.AudioThrustersIgniteEnd),
		AudioThrustersIgnition:          lookupThinHash(w.AudioThrustersIgnition),
		AudioEntryWooshLoop:             lookupThinHash(w.AudioEntryWooshLoop),
		AudioEntryWooshEnd:              lookupThinHash(w.AudioEntryWooshEnd),
		AudioChassiShakeLoop:            lookupThinHash(w.AudioChassiShakeLoop),
		AudioChassiShakeEnd:             lookupThinHash(w.AudioChassiShakeEnd),
		AudioSupersonicBoom:             lookupThinHash(w.AudioSupersonicBoom),
		AudioImpact:                     lookupThinHash(w.AudioImpact),
		AudioBoltBlowing:                lookupThinHash(w.AudioBoltBlowing),
		AudioLidBlowing:                 lookupThinHash(w.AudioLidBlowing),
		AudioThrusterStart:              lookupThinHash(w.AudioThrusterStart),
		AudioHellpodPreImpactRiser:      lookupThinHash(w.AudioHellpodPreImpactRiser),
		AudioMusicHellpodGroundApproach: lookupThinHash(w.AudioMusicHellpodGroundApproach),
		AudioMusicHellpodLand:           lookupThinHash(w.AudioMusicHellpodLand),
		UnknownBool3:                    w.UnknownBool3 != 0,
	}
}

func getHellpodComponentData() ([]byte, error) {
	hellpodComponentHash := Sum("HellpodComponentData")
	hellpodComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(hellpodComponentHashData, binary.LittleEndian, hellpodComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, hellpodComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getHellpodComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("HellpodComponentData")
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
		return nil, fmt.Errorf("HellpodComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("HellpodComponent") {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (data type was not HellpodComponent)")
	}

	hellpodComponentData, err := getHellpodComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get hellpod component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(hellpodComponentData)

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
		return nil, fmt.Errorf("%v not found in hellpod component data", hash.String())
	}

	var hellpodComponentType DLTypeDesc
	hellpodComponentType, ok = typelib.Types[Sum("HellpodComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find HellpodComponent hash in dl_library")
	}

	componentData := make([]byte, hellpodComponentType.Size)
	if _, err := r.Seek(int64(hellpodComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseHellpodComponents() (map[stingray.Hash]HellpodComponent, error) {
	unitHash := Sum("HellpodComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var hellpodType DLTypeDesc
	var ok bool
	hellpodType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find HellpodComponentData hash in dl_library")
	}

	if len(hellpodType.Members) != 2 {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (there should be 2 members but were actually %v)", len(hellpodType.Members))
	}

	if hellpodType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (hashmap atom was not inline array)")
	}

	if hellpodType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (data atom was not inline array)")
	}

	if hellpodType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (hashmap storage was not struct)")
	}

	if hellpodType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (data storage was not struct)")
	}

	if hellpodType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if hellpodType.Members[1].TypeID != Sum("HellpodComponent") {
		return nil, fmt.Errorf("HellpodComponentData unexpected format (data type was not HellpodComponent)")
	}

	hellpodComponentData, err := getHellpodComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get hellpod component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(hellpodComponentData)

	hashmap := make([]ComponentIndexData, hellpodType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]HellpodComponent, hellpodType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]HellpodComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
