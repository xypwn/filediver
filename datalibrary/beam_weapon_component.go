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

type BeamPrisms struct {
	LightBeamType          enum.BeamType `json:"light_beam_type"`
	LightHeatGenerationMul float32       `json:"light_heat_generation_mul"`
	HeavyBeamType          enum.BeamType `json:"heavy_beam_type"`
	HeavyHeatGenerationMul float32       `json:"heavy_heat_generation_mul"`
}

// ScopeResponsiveness           float32       // How quickly the scope/sight follows changes in aim and movement.
type BeamWeaponComponent struct {
	Type                          enum.BeamType // Type of beam it fires.
	Prisms                        BeamPrisms    // Beam Prism Settings for weapons with the weapon function.
	ScopeCrosshair                mgl32.Vec2    // Crosshair position on screen in [-1, 1] range.
	FocalDistances                mgl32.Vec3    // The focal distances of the weapon's lens
	UseFirenodePose               float32
	UseMidiEventSystem            uint8 // [bool]Fire event will be posted using Wwise's MIDI system as a MIDI sequence (cannot be paused/resumed).
	_                             [3]uint8
	MidiTimingRandomization       mgl32.Vec2        // Events posted during the MIDI sequence will have a random time offset, measured in milliseconds.
	MidiStopDelay                 float32           // A delay for when to notify Wwise that the MIDI sequence has stopped, measured in milliseconds.
	FireLoopStartAudioEvent       stingray.ThinHash // [wwise] The looping audio event to start when starting to fire.
	FireLoopStopAudioEvent        stingray.ThinHash // [wwise] The looping audio event to play when stopping fire.
	FireSingleAudioEvent          stingray.ThinHash // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	UnkAudioEvent                 stingray.ThinHash // name length 27
	_                             [4]uint8
	MuzzleFlash                   stingray.Hash     // [particles]Muzzle flash effect
	NoiseTimer                    float32           // How often does the weapon make noise?
	FireSourceNode                stingray.ThinHash // [string]The node to play the firing audio events at.
	DryFireAudioEvent             stingray.ThinHash // [wwise]The wwise sound id to play when dry firing.
	DryFireRepeatAudioEvent       stingray.ThinHash // [wwise]The wwise sound id to play when repeatedly dry firing.
	OnFireStartedWielderAnimEvent stingray.ThinHash // [string]Animation event to trigger on the wielder when we start firing.
	OnFireStoppedWielderAnimEvent stingray.ThinHash // [string]Animation event to trigger on the wielder when we stop firing.
	UnkBool                       uint8             // name length 19
	_                             [3]uint8
	UnkInt                        int32 // name length 16
	UnkInt2                       int32 // name length 19
	_                             [4]uint8
}

type SimpleBeamWeaponComponent struct {
	Type   enum.BeamType `json:"beam_type"`
	Prisms BeamPrisms    `json:"beam_prisms"`
	//ScopeResponsiveness           float32       `json:"scope_responsiveness"`
	ScopeCrosshair                mgl32.Vec2 `json:"scope_crosshair"`
	FocalDistances                mgl32.Vec3 `json:"focal_distances"`
	UseFirenodePose               float32    `json:"use_firenode_pose"`
	UseMidiEventSystem            bool       `json:"use_midi_event_system"`
	MidiTimingRandomization       mgl32.Vec2 `json:"midi_timing_randomization"`
	MidiStopDelay                 float32    `json:"midi_stop_delay"`
	FireLoopStartAudioEvent       string     `json:"fire_loop_start_audio_delay"`
	FireLoopStopAudioEvent        string     `json:"fire_loop_stop_audio_event"`
	FireSingleAudioEvent          string     `json:"fire_single_audio_event"`
	UnkAudioEvent                 string     `json:"unk_audio_event"` // name length 27
	MuzzleFlash                   string     `json:"muzzle_flash"`
	NoiseTimer                    float32    `json:"noise_timer"`
	FireSourceNode                string     `json:"fire_source_node"`
	DryFireAudioEvent             string     `json:"dry_fire_audio_event"`
	DryFireRepeatAudioEvent       string     `json:"dry_fire_repeat_audio_event"`
	OnFireStartedWielderAnimEvent string     `json:"on_fire_started_wielder_anim_event"`
	OnFireStoppedWielderAnimEvent string     `json:"on_fire_stopped_wielder_anim_event"`
	UnkBool                       bool       `json:"unk_bool"`
	UnkInt                        int32      `json:"unk_int"`
	UnkInt2                       int32      `json:"unk_int2"`
}

func (b BeamWeaponComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleBeamWeaponComponent{
		Type:   b.Type,
		Prisms: b.Prisms,
		//ScopeResponsiveness:           b.ScopeResponsiveness,
		ScopeCrosshair:                b.ScopeCrosshair,
		FocalDistances:                b.FocalDistances,
		UseFirenodePose:               b.UseFirenodePose,
		UseMidiEventSystem:            b.UseMidiEventSystem != 0,
		MidiTimingRandomization:       b.MidiTimingRandomization,
		MidiStopDelay:                 b.MidiStopDelay,
		FireLoopStartAudioEvent:       lookupThinHash(b.FireLoopStartAudioEvent),
		FireLoopStopAudioEvent:        lookupThinHash(b.FireLoopStopAudioEvent),
		FireSingleAudioEvent:          lookupThinHash(b.FireSingleAudioEvent),
		UnkAudioEvent:                 lookupThinHash(b.UnkAudioEvent),
		MuzzleFlash:                   lookupHash(b.MuzzleFlash),
		NoiseTimer:                    b.NoiseTimer,
		FireSourceNode:                lookupThinHash(b.FireSourceNode),
		DryFireAudioEvent:             lookupThinHash(b.DryFireAudioEvent),
		DryFireRepeatAudioEvent:       lookupThinHash(b.DryFireRepeatAudioEvent),
		OnFireStartedWielderAnimEvent: lookupThinHash(b.OnFireStartedWielderAnimEvent),
		OnFireStoppedWielderAnimEvent: lookupThinHash(b.OnFireStoppedWielderAnimEvent),
		UnkBool:                       b.UnkBool != 0,
		UnkInt:                        b.UnkInt,
		UnkInt2:                       b.UnkInt2,
	}
}

func getBeamWeaponComponentData() ([]byte, error) {
	beamWeaponHash := Sum("BeamWeaponComponentData")
	beamWeaponHashData := make([]byte, 4)
	if _, err := binary.Encode(beamWeaponHashData, binary.LittleEndian, beamWeaponHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, beamWeaponHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getBeamWeaponComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	BeamWeaponCmpDataHash := Sum("BeamWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var beamWeaponCmpDataType DLTypeDesc
	var ok bool
	beamWeaponCmpDataType, ok = typelib.Types[BeamWeaponCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find BeamWeaponComponentData hash in dl_library")
	}

	if len(beamWeaponCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(beamWeaponCmpDataType.Members))
	}

	if beamWeaponCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if beamWeaponCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if beamWeaponCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if beamWeaponCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (data storage was not struct)")
	}

	if beamWeaponCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if beamWeaponCmpDataType.Members[1].TypeID != Sum("BeamWeaponComponent") {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (data type was not BeamWeaponComponent)")
	}

	beamWeaponComponentData, err := getBeamWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get beam weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(beamWeaponComponentData)

	hashmap := make([]ComponentIndexData, beamWeaponCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in beam weapon component data", hash.String())
	}

	var beamWeaponComponentType DLTypeDesc
	beamWeaponComponentType, ok = typelib.Types[Sum("BeamWeaponComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find BeamWeaponComponent hash in dl_library")
	}

	componentData := make([]byte, beamWeaponComponentType.Size)
	if _, err := r.Seek(int64(beamWeaponComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseBeamWeaponComponents() (map[stingray.Hash]BeamWeaponComponent, error) {
	beamWeaponHash := Sum("BeamWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var beamWeaponType DLTypeDesc
	var ok bool
	beamWeaponType, ok = typelib.Types[beamWeaponHash]
	if !ok {
		return nil, fmt.Errorf("could not find BeamWeaponComponentData hash in dl_library")
	}

	if len(beamWeaponType.Members) != 2 {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(beamWeaponType.Members))
	}

	if beamWeaponType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if beamWeaponType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if beamWeaponType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if beamWeaponType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (data storage was not struct)")
	}

	if beamWeaponType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if beamWeaponType.Members[1].TypeID != Sum("BeamWeaponComponent") {
		return nil, fmt.Errorf("BeamWeaponComponentData unexpected format (data type was not BeamWeaponComponent)")
	}

	beamWeaponComponentData, err := getBeamWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(beamWeaponComponentData)

	hashmap := make([]ComponentIndexData, beamWeaponType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]BeamWeaponComponent, beamWeaponType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]BeamWeaponComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
