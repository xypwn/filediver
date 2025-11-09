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

type RackAttach struct {
	Item                  stingray.Hash
	Node                  stingray.ThinHash // [string]Node where to attach the hellpod rack payload.
	Offset                mgl32.Vec3        // Offset from 'node'.
	RotationOffset        mgl32.Vec3        // Rotation offset from 'node' in degrees.
	DeployAnimationEvent  stingray.ThinHash // [string]Animation to play when deploying this attach.
	RetractAnimationEvent stingray.ThinHash // [string]Animation to play when retracting this attach.
	RetractAudioEvent     stingray.ThinHash // [wwise]
	ApplyDeltas           uint8             // [bool]If true, apply any stratagem customization deltas to this entity when spawning.
	_                     [3]uint8
	RackSide              enum.RackSide // Is this item on the left or right side of the rack? Cannot be left on None.
	UnknownBool           uint8         // [bool]Unknown, name length 25
	_                     [7]uint8
}

type HellpodRackComponent struct {
	Payloads                  [8]RackAttach // Payloads to attach (what & where).
	LastRetractDelay          float32       // Extra time deloy to play the retract functionality on the last retracting attachment.
	DisableRetract            uint8         // [bool]Special case to disable the retract animations from playing on both sides of the rack.
	_                         [3]uint8
	MapIcon                   stingray.Hash     // [material]Map Icon for this hellpod rack
	MapName                   uint32            // [string]Localization tag for the map name of this hellpod rack
	RackDeployAudioEvent      stingray.ThinHash // [wwise]
	RackRetractAudioEvent     stingray.ThinHash // [wwise]
	DoorRetractAudioEvent     stingray.ThinHash // [wwise]
	DeployAbility             enum.AbilityId    // Ability to play when deploying.
	LastRetractAbility        enum.AbilityId    // Ability to play when we are out of items.
	RandomPayloadSize         uint32            // If we have a random payload size above 0, then we will attach random items from the payload settings. Override spawn_payload_size
	SpawnPayloadSize          uint32            // Indicates how many payloads should spawn from the settings list.
	DisableInitialInteraction uint8             // [bool]Should we prevent the initial interaction for the payload, meaning you have to destroy the rack in order to get them?
	_                         [7]uint8
}

type SimpleRackAttach struct {
	Item                  string        `json:"item"`
	Node                  string        `json:"node"`
	Offset                mgl32.Vec3    `json:"offset"`
	RotationOffset        mgl32.Vec3    `json:"rotation_offset"`
	DeployAnimationEvent  string        `json:"deploy_animation_event"`
	RetractAnimationEvent string        `json:"retract_animation_event"`
	RetractAudioEvent     string        `json:"retract_audio_event"`
	ApplyDeltas           bool          `json:"apply_deltas"`
	RackSide              enum.RackSide `json:"rack_side"`
	UnknownBool           bool          `json:"unknown_bool"`
}

type SimpleHellpodRackComponent struct {
	Payloads                  []SimpleRackAttach `json:"payloads"`
	LastRetractDelay          float32            `json:"last_retract_delay"`
	DisableRetract            bool               `json:"disable_retract"`
	MapIcon                   string             `json:"map_icon"`
	MapName                   string             `json:"map_name"`
	RackDeployAudioEvent      string             `json:"rack_deploy_audio_event"`
	RackRetractAudioEvent     string             `json:"rack_retract_audio_event"`
	DoorRetractAudioEvent     string             `json:"door_retract_audio_event"`
	DeployAbility             enum.AbilityId     `json:"deploy_ability"`
	LastRetractAbility        enum.AbilityId     `json:"last_retract_ability"`
	RandomPayloadSize         uint32             `json:"random_payload_size"`
	SpawnPayloadSize          uint32             `json:"spawn_payload_size"`
	DisableInitialInteraction bool               `json:"disable_initial_interaction"`
}

func (w HellpodRackComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	payloads := make([]SimpleRackAttach, 0)
	for _, payload := range w.Payloads {
		if payload.Item.Value == 0 {
			break
		}
		payloads = append(payloads, SimpleRackAttach{
			Item:                  lookupHash(payload.Item),
			Node:                  lookupThinHash(payload.Node),
			Offset:                payload.Offset,
			RotationOffset:        payload.RotationOffset,
			DeployAnimationEvent:  lookupThinHash(payload.DeployAnimationEvent),
			RetractAnimationEvent: lookupThinHash(payload.RetractAnimationEvent),
			RetractAudioEvent:     lookupThinHash(payload.RetractAudioEvent),
			ApplyDeltas:           payload.ApplyDeltas != 0,
			RackSide:              payload.RackSide,
			UnknownBool:           payload.UnknownBool != 0,
		})
	}
	return SimpleHellpodRackComponent{
		Payloads:                  payloads,
		LastRetractDelay:          w.LastRetractDelay,
		DisableRetract:            w.DisableRetract != 0,
		MapIcon:                   lookupHash(w.MapIcon),
		MapName:                   lookupStrings(w.MapName),
		RackDeployAudioEvent:      lookupThinHash(w.RackDeployAudioEvent),
		RackRetractAudioEvent:     lookupThinHash(w.RackRetractAudioEvent),
		DoorRetractAudioEvent:     lookupThinHash(w.DoorRetractAudioEvent),
		DeployAbility:             w.DeployAbility,
		LastRetractAbility:        w.LastRetractAbility,
		RandomPayloadSize:         w.RandomPayloadSize,
		SpawnPayloadSize:          w.SpawnPayloadSize,
		DisableInitialInteraction: w.DisableInitialInteraction != 0,
	}
}

func getHellpodRackComponentData() ([]byte, error) {
	hellpodRackComponentHash := Sum("HellpodRackComponentData")
	hellpodRackComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(hellpodRackComponentHashData, binary.LittleEndian, hellpodRackComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, hellpodRackComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getHellpodRackComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("HellpodRackComponentData")
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
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("HellpodRackComponent") {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (data type was not HellpodRackComponent)")
	}

	hellpodRackComponentData, err := getHellpodRackComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get hellpod rack component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(hellpodRackComponentData)

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
		return nil, fmt.Errorf("%v not found in hellpod rack component data", hash.String())
	}

	var hellpodRackComponentType DLTypeDesc
	hellpodRackComponentType, ok = typelib.Types[Sum("HellpodRackComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find HellpodRackComponent hash in dl_library")
	}

	componentData := make([]byte, hellpodRackComponentType.Size)
	if _, err := r.Seek(int64(hellpodRackComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseHellpodRackComponents() (map[stingray.Hash]HellpodRackComponent, error) {
	unitHash := Sum("HellpodRackComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var hellpodRackType DLTypeDesc
	var ok bool
	hellpodRackType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find HellpodRackComponentData hash in dl_library")
	}

	if len(hellpodRackType.Members) != 2 {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (there should be 2 members but were actually %v)", len(hellpodRackType.Members))
	}

	if hellpodRackType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (hashmap atom was not inline array)")
	}

	if hellpodRackType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (data atom was not inline array)")
	}

	if hellpodRackType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (hashmap storage was not struct)")
	}

	if hellpodRackType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (data storage was not struct)")
	}

	if hellpodRackType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if hellpodRackType.Members[1].TypeID != Sum("HellpodRackComponent") {
		return nil, fmt.Errorf("HellpodRackComponentData unexpected format (data type was not HellpodRackComponent)")
	}

	hellpodRackComponentData, err := getHellpodRackComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get hellpod rack component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(hellpodRackComponentData)

	hashmap := make([]ComponentIndexData, hellpodRackType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]HellpodRackComponent, hellpodRackType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]HellpodRackComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
