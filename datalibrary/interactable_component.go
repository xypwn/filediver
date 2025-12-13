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

type InteractZoneInfo struct {
	Name                      stingray.ThinHash // [string]Name of the interactable zone.
	Radius                    float32           // Radius of the interactable zone.
	ViewDistance              float32           // The distance at which the interactable icon is visible before it fades out.
	Node                      stingray.ThinHash // [string]Name of the attach-node.
	Offset                    mgl32.Vec3        // Offset from 'node'.
	StandLocationOffset       mgl32.Vec3        // Location Offset from 'node', where the player will stand when interacting.
	InteractType              enum.InteractType // The type of interact.
	RequireMigration          uint8             // [bool]Need migration on interact.
	StartEnabled              uint8             // [bool]When spawn, should it be enabled or not?
	_                         [2]uint8
	StartEnabledDelay         float32                     // When spawns, after how many second it should enable the interaction, -1 Indicates that the mechanism is disabled.
	Label                     uint32                      // [string]Localization string describing the interaction in the HUD.
	InactiveLabel             uint32                      // [string]Localization string describing an unavailable interaction in the HUD.
	ApproachDirection         float32                     // 2D angle between approaching unit and interactable, in the interactable's local space, that we center the interaction cone around
	ApproachAngle             float32                     // Half angle of the cone centered around approach_direction that we allow interaction in. If > 180 degreres, full circle is allowed.
	ApproachApexShift         float32                     // Shifts cone towards -direction
	ScreenBorder              int32                       // How far from the screen border interaction point should be. -1 can be interacted even offscreen
	ScriptEvent               stingray.ThinHash           // [string]Script event to be triggered on the interactable upon successful interact. - Require a Behavior Component on the the unit.
	ScriptEventForLinkedUnits stingray.ThinHash           // [string]Script event called on linked units to be triggered on the interactable upon successful interact.
	UnknownScriptEvent        stingray.ThinHash           // Name length 26
	Priority                  uint32                      // This interact zone's priority when within range of other zones. Higher priority values are favored over lower ones.
	BlockingInjuries          enum.InteractableInjuryType // Disable the interaction when the avatar have this injuries.
	ReenableImmediately       uint8                       // [bool]When used, re-enable the zone immediately
	DeleteAfterInteract       uint8                       // [bool]This will delete the entity after a successful interact.
	DisallowProneInteract     uint8                       // [bool]Disallows interacting when interactor is prone.
	UnknownBool               uint8                       // Name length 29
	UnknownBool2              uint8                       // Name length 25
	_                         [3]uint8
	HoldTime                  float32 // Hold time required to trigger this interact.
	HoverIconHeight           float32 // The in-world height offset of the interact type icon from the interact point.
	InteractDiamondScale      float32 // The scale of the interact diamond.
	AutoInteract              uint8   // [bool]If enabled, users will trigger this zone by just getting in proximity
	ShowIfDisabled            uint8   // [bool]If true, then the disabled UI will be shown when this interact is disabled. If false then nothing will be shown.
	_                         [2]uint8
	HintId                    uint32 // [string] Hint to show
	OptHintId                 uint32 // [string] Second hint to show
}

type InteractableComponent struct {
	RequireWaitingForResult            uint8                 // [bool]Requires that the interactor needs to wait for the interaction before making new interacts.
	DisableInteractionWhileInteracting uint8                 // [bool]Indicates if we automatically disable the interactable zones when we are start the interaction
	AllowMove                          uint8                 // [bool]True if this interactable is allowed to move in the world.
	PerformLineOfSightCheck            uint8                 // [bool]True if this interactable should perform a line of sight check towards the interactor.
	Mode                               enum.InteractableMode // Describes who may interact with this interactable.
	Zones                              [8]InteractZoneInfo   // .
	InteractAudioEvent                 stingray.ThinHash     // [string]Audio event to trigger when the item is interacted with
	InteractAudioEventVo               stingray.ThinHash     // [string]Audio VO event to trigger when the item is interacted with
}

type SimpleInteractZoneInfo struct {
	Name                      string                      `json:"name"`                          // [string]Name of the interactable zone.
	Radius                    float32                     `json:"radius"`                        // Radius of the interactable zone.
	ViewDistance              float32                     `json:"view_distance"`                 // The distance at which the interactable icon is visible before it fades out.
	Node                      string                      `json:"node"`                          // [string]Name of the attach-node.
	Offset                    mgl32.Vec3                  `json:"offset"`                        // Offset from 'node'.
	StandLocationOffset       mgl32.Vec3                  `json:"stand_location_offset"`         // Location Offset from 'node', where the player will stand when interacting.
	InteractType              enum.InteractType           `json:"interact_type"`                 // The type of interact.
	RequireMigration          bool                        `json:"require_migration"`             // [bool]Need migration on interact.
	StartEnabled              bool                        `json:"start_enabled"`                 // [bool]When spawn, should it be enabled or not?
	StartEnabledDelay         float32                     `json:"start_enabled_delay"`           // When spawns, after how many second it should enable the interaction, -1 Indicates that the mechanism is disabled.
	Label                     string                      `json:"label"`                         // [string]Localization string describing the interaction in the HUD.
	InactiveLabel             string                      `json:"inactive_label"`                // [string]Localization string describing an unavailable interaction in the HUD.
	ApproachDirection         float32                     `json:"approach_direction"`            // 2D angle between approaching unit and interactable, in the interactable's local space, that we center the interaction cone around
	ApproachAngle             float32                     `json:"approach_angle"`                // Half angle of the cone centered around approach_direction that we allow interaction in. If > 180 degreres, full circle is allowed.
	ApproachApexShift         float32                     `json:"approach_apex_shift"`           // Shifts cone towards -direction
	ScreenBorder              int32                       `json:"screen_border"`                 // How far from the screen border interaction point should be. -1 can be interacted even offscreen
	ScriptEvent               string                      `json:"script_event"`                  // [string]Script event to be triggered on the interactable upon successful interact. - Require a Behavior Component on the the unit.
	ScriptEventForLinkedUnits string                      `json:"script_event_for_linked_units"` // [string]Script event called on linked units to be triggered on the interactable upon successful interact.
	UnknownScriptEvent        string                      `json:"unknown_script_event"`          // Name length 26
	Priority                  uint32                      `json:"priority"`                      // This interact zone's priority when within range of other zones. Higher priority values are favored over lower ones.
	BlockingInjuries          enum.InteractableInjuryType `json:"blocking_injuries"`             // Disable the interaction when the avatar have this injuries.
	ReenableImmediately       bool                        `json:"reenable_immediately"`          // [bool]When used, re-enable the zone immediately
	DeleteAfterInteract       bool                        `json:"delete_after_interact"`         // [bool]This will delete the entity after a successful interact.
	DisallowProneInteract     bool                        `json:"disallow_prone_interact"`       // [bool]Disallows interacting when interactor is prone.
	UnknownBool               bool                        `json:"unknown_bool"`                  // Name length 29
	UnknownBool2              bool                        `json:"unknown_bool2"`                 // Name length 25
	HoldTime                  float32                     `json:"hold_time"`                     // Hold time required to trigger this interact.
	HoverIconHeight           float32                     `json:"hover_icon_height"`             // The in-world height offset of the interact type icon from the interact point.
	InteractDiamondScale      float32                     `json:"interact_diamond_scale"`        // The scale of the interact diamond.
	AutoInteract              bool                        `json:"auto_interact"`                 // [bool]If enabled, users will trigger this zone by just getting in proximity
	ShowIfDisabled            bool                        `json:"show_if_disabled"`              // [bool]If true, then the disabled UI will be shown when this interact is disabled. If false then nothing will be shown.
	HintId                    string                      `json:"hint_id"`                       // [string] Hint to show
	OptHintId                 string                      `json:"opt_hint_id"`                   // [string] Second hint to show
}

type SimpleInteractableComponent struct {
	RequireWaitingForResult            bool                     `json:"require_waiting_for_result"`            // [bool]Requires that the interactor needs to wait for the interaction before making new interacts.
	DisableInteractionWhileInteracting bool                     `json:"disable_interaction_while_interacting"` // [bool]Indicates if we automatically disable the interactable zones when we are start the interaction
	AllowMove                          bool                     `json:"allow_move"`                            // [bool]True if this interactable is allowed to move in the world.
	PerformLineOfSightCheck            bool                     `json:"perform_line_of_sight_check"`           // [bool]True if this interactable should perform a line of sight check towards the interactor.
	Mode                               enum.InteractableMode    `json:"mode"`                                  // Describes who may interact with this interactable.
	Zones                              []SimpleInteractZoneInfo `json:"zones"`                                 // .
	InteractAudioEvent                 string                   `json:"interact_audio_event"`                  // [string]Audio event to trigger when the item is interacted with
	InteractAudioEventVo               string                   `json:"interact_audio_event_vo"`               // [string]Audio VO event to trigger when the item is interacted with
}

func (w InteractZoneInfo) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleInteractZoneInfo {
	return SimpleInteractZoneInfo{
		Name:                      lookupThinHash(w.Name),
		Radius:                    w.Radius,
		ViewDistance:              w.ViewDistance,
		Node:                      lookupThinHash(w.Node),
		Offset:                    w.Offset,
		StandLocationOffset:       w.StandLocationOffset,
		InteractType:              w.InteractType,
		RequireMigration:          w.RequireMigration != 0,
		StartEnabled:              w.StartEnabled != 0,
		StartEnabledDelay:         w.StartEnabledDelay,
		Label:                     lookupStrings(w.Label),
		InactiveLabel:             lookupStrings(w.InactiveLabel),
		ApproachDirection:         w.ApproachDirection,
		ApproachAngle:             w.ApproachAngle,
		ApproachApexShift:         w.ApproachApexShift,
		ScreenBorder:              w.ScreenBorder,
		ScriptEvent:               lookupThinHash(w.ScriptEvent),
		ScriptEventForLinkedUnits: lookupThinHash(w.ScriptEventForLinkedUnits),
		UnknownScriptEvent:        lookupThinHash(w.UnknownScriptEvent),
		Priority:                  w.Priority,
		BlockingInjuries:          w.BlockingInjuries,
		ReenableImmediately:       w.ReenableImmediately != 0,
		DeleteAfterInteract:       w.DeleteAfterInteract != 0,
		DisallowProneInteract:     w.DisallowProneInteract != 0,
		UnknownBool:               w.UnknownBool != 0,
		UnknownBool2:              w.UnknownBool2 != 0,
		HoldTime:                  w.HoldTime,
		HoverIconHeight:           w.HoverIconHeight,
		InteractDiamondScale:      w.InteractDiamondScale,
		AutoInteract:              w.AutoInteract != 0,
		ShowIfDisabled:            w.ShowIfDisabled != 0,
		HintId:                    lookupStrings(w.HintId),
		OptHintId:                 lookupStrings(w.OptHintId),
	}
}

func (w InteractableComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	zones := make([]SimpleInteractZoneInfo, 0)
	for _, zone := range w.Zones {
		if zone.Name.Value == 0 || zone.InteractType == enum.InteractType_None {
			break
		}
		zones = append(zones, zone.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleInteractableComponent{
		RequireWaitingForResult:            w.RequireWaitingForResult != 0,
		DisableInteractionWhileInteracting: w.DisableInteractionWhileInteracting != 0,
		AllowMove:                          w.AllowMove != 0,
		PerformLineOfSightCheck:            w.PerformLineOfSightCheck != 0,
		Mode:                               w.Mode,
		Zones:                              zones,
		InteractAudioEvent:                 lookupThinHash(w.InteractAudioEvent),
		InteractAudioEventVo:               lookupThinHash(w.InteractAudioEventVo),
	}
}

func getInteractableComponentData() ([]byte, error) {
	interactableComponentHash := Sum("InteractableComponentData")
	interactableComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(interactableComponentHashData, binary.LittleEndian, interactableComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, interactableComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getInteractableComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("InteractableComponentData")
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
		return nil, fmt.Errorf("InteractableComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("InteractableComponent") {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (data type was not InteractableComponent)")
	}

	interactableComponentData, err := getInteractableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get interactable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(interactableComponentData)

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
		return nil, fmt.Errorf("%v not found in interactable component data", hash.String())
	}

	var interactableComponentType DLTypeDesc
	interactableComponentType, ok = typelib.Types[Sum("InteractableComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find InteractableComponent hash in dl_library")
	}

	componentData := make([]byte, interactableComponentType.Size)
	if _, err := r.Seek(int64(interactableComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseInteractableComponents() (map[stingray.Hash]InteractableComponent, error) {
	unitHash := Sum("InteractableComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var interactableType DLTypeDesc
	var ok bool
	interactableType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find InteractableComponentData hash in dl_library")
	}

	if len(interactableType.Members) != 2 {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (there should be 2 members but were actually %v)", len(interactableType.Members))
	}

	if interactableType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if interactableType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (data atom was not inline array)")
	}

	if interactableType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (hashmap storage was not struct)")
	}

	if interactableType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (data storage was not struct)")
	}

	if interactableType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if interactableType.Members[1].TypeID != Sum("InteractableComponent") {
		return nil, fmt.Errorf("InteractableComponentData unexpected format (data type was not InteractableComponent)")
	}

	interactableComponentData, err := getInteractableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get interactable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(interactableComponentData)

	hashmap := make([]ComponentIndexData, interactableType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]InteractableComponent, interactableType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]InteractableComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
