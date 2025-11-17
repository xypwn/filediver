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

type EquipmentComponent struct {
	WieldNode                     stingray.ThinHash // [string]Name of the node on the wielder that this equipment should be attached to when wielded.
	WieldNodeRotationOffset       mgl32.Vec3        // If set, this rotation offset is applied when wielded.
	WieldNodePositionOffset       mgl32.Vec3        // If set, this position offset is applied when wielded.
	HolsterNode                   stingray.ThinHash // [string]Name of the node on the wielder that this equipment should be attached to when holstered.
	HolsterNodeRotationOffset     mgl32.Vec3        // If set, this rotation offset is applied when holstered.
	HolsterNodePositionOffset     mgl32.Vec3        // If set, this position offset is applied when holstered.
	HolsterDuration               float32           // The time it takes to holster this equipment.
	WieldDuration                 float32           // The time it takes to wield this equipment.
	PickupDuration                float32           // The time it takes to pickup this equipment.
	UnknownHash                   stingray.ThinHash // unknown, name length 23
	WieldAnimation                stingray.ThinHash // [string]Name of the animation to play on the wielder when wielding.
	PickupAnimation               stingray.ThinHash // [string]Name of the animation to play on the wielder when picking up otherwise fallback to wield animation.
	HolsterAnimation              stingray.ThinHash // [string]Name of the animation to play on the wielder when holstering.
	DropAnimation                 stingray.ThinHash // [string]Name of the animation to play on the wielder when dropping.
	AmmocheckAnimation            stingray.ThinHash // [string]Name of the animation to play on the wielder when checking ammo.
	UnknownBool                   uint8             // unknown, name length 30
	UnknownBool2                  uint8             // unknown, name length 20
	_                             [2]uint8
	HolsterDurationForAttach      float32                // The time it takes to have the equipment attach itself to the holster node when we start holstering.
	WieldDurationForAttach        float32                // The time it takes to have the equipment attach itself to the wield node when we start wielding.
	AttachLerpTime                float32                // Should we apply lerp to the attachable component when we wield/holster and if so, how long should it lerp for?
	OnHolsterAnimationEvent       stingray.ThinHash      // [string]Animation event to trigger when this entity is holstered.
	WieldAnimationSelf            stingray.ThinHash      // [string]Name of the animation to play on the equipment when wielding.
	PickupAnimationSelf           stingray.ThinHash      // [string]Name of the animation to play on the equipment when picking up otherwise fallback to wield animation.
	HolsterAnimationSelf          stingray.ThinHash      // [string]Name of the animation to play on the equipment when holstering.
	UnknownHash3                  stingray.ThinHash      // Unknown, name length 29
	EquipmentType                 enum.EquipmentType     // The category of this equipment piece, mostly used for AI purposes.
	DropMode                      enum.EquipmentDropMode // How the equipment should be treated when dropped.
	DropAudioEvent                stingray.ThinHash      // [wwise]Audio event to trigger when getting rid of an equipment.
	UnknownAudioEvent             stingray.ThinHash      // Unknown, name length 20, probably an audio event
	DropIconUi                    stingray.Hash          // Unknown, name length 12
	PickupAudioEvent              stingray.ThinHash      // [wwise]Audio event to trigger when putting on a new equipment.
	UnknownEvent2                 stingray.ThinHash      // Unknown, name length 21
	UnknownEvent3                 stingray.ThinHash      // Unknown, name length 33
	UnknownEvent4                 stingray.ThinHash      // Unknown, name length 32
	WeaponSwitchWieldAudioEvent   stingray.ThinHash      // [wwise]Audio event to trigger when wielding the weapon.
	WeaponSwitchHolsterAudioEvent stingray.ThinHash      // [wwise]Audio event to trigger when holstering the weapon.
	PickupAbility                 enum.AbilityId         // Ability to play when picking up from the ground. Currently only used for the avatar.
	DropAbility                   enum.AbilityId         // ability to play when dropping equipment
	GripType                      enum.WeaponGripType    // Grip type for this weapon.
	IsOneHanded                   uint8                  // [bool]Whether the weapon can be wielded in one hand or not.
	AllowWieldOnShip              uint8                  // [bool]If the weapon can be wielded while on the ship, only intended use for debug purposes.
	ShowInPickupWidget            uint8                  // [bool]True the equipment should be shown in the pickup widget.
	_                             uint8
	HolsterUnit                   stingray.Hash            // [unit]Path to the holster unit. This is an additional unit to spawn on the holster node that doesn't move with the equipment. Useful for sheaths etc. Note that since it's an additional unit it's more expensive than building the sheath into the parent unit, so only use when needed.
	HintId                        stingray.ThinHash        // [string]Hint ID to trigger when equipping this equipment
	UnknownHash4                  stingray.ThinHash        // Unknown, name length 14
	EquipmentWwiseIndex           enum.EquipmentWwiseIndex // the wwise item index for this equipment
	BackpackWwiseIndex            enum.BackpackWwiseIndex  // the wwise item index for this backpack if it is a backpack
	UseBackpackFilledSwitch       uint8                    // [bool]If yes set_switch would be called on audio source before pickup drop foley and impact sounds
	_                             [3]uint8
	UnknownHash5                  stingray.ThinHash // Unknown, name length 22
}

type SimpleEquipmentComponent struct {
	WieldNode                     string                   `json:"wield_node"`
	WieldNodeRotationOffset       mgl32.Vec3               `json:"wield_node_rotation_offset"`
	WieldNodePositionOffset       mgl32.Vec3               `json:"wield_node_position_offset"`
	HolsterNode                   string                   `json:"holster_node"`
	HolsterNodeRotationOffset     mgl32.Vec3               `json:"holster_node_rotation_offset"`
	HolsterNodePositionOffset     mgl32.Vec3               `json:"holster_node_position_offset"`
	HolsterDuration               float32                  `json:"holster_duration"`
	WieldDuration                 float32                  `json:"wield_duration"`
	PickupDuration                float32                  `json:"pickup_duration"`
	UnknownHash                   string                   `json:"unknown_hash"`
	WieldAnimation                string                   `json:"wield_animation"`
	PickupAnimation               string                   `json:"pickup_animation"`
	HolsterAnimation              string                   `json:"holster_animation"`
	DropAnimation                 string                   `json:"drop_animation"`
	AmmocheckAnimation            string                   `json:"ammocheck_animation"`
	UnknownBool                   bool                     `json:"unknown_bool"`
	UnknownBool2                  bool                     `json:"unknown_bool2"`
	HolsterDurationForAttach      float32                  `json:"holster_duration_for_attach"`
	WieldDurationForAttach        float32                  `json:"wield_duration_for_attach"`
	AttachLerpTime                float32                  `json:"attach_lerp_time"`
	OnHolsterAnimationEvent       string                   `json:"on_holster_animation_event"`
	WieldAnimationSelf            string                   `json:"wield_animation_self"`
	PickupAnimationSelf           string                   `json:"pickup_animation_self"`
	HolsterAnimationSelf          string                   `json:"holster_animation_self"`
	UnknownHash3                  string                   `json:"unknown_hash3"`
	EquipmentType                 enum.EquipmentType       `json:"equipment_type"`
	DropMode                      enum.EquipmentDropMode   `json:"drop_mode"`
	DropAudioEvent                string                   `json:"drop_audio_event"`
	UnknownAudioEvent             string                   `json:"unknown_event"`
	DropIconUi                    string                   `json:"drop_icon_ui"`
	PickupAudioEvent              string                   `json:"pickup_audio_event"`
	UnknownEvent2                 string                   `json:"unknown_event2"`
	UnknownEvent3                 string                   `json:"unknown_event3"`
	UnknownEvent4                 string                   `json:"unknown_event4"`
	WeaponSwitchWieldAudioEvent   string                   `json:"weapon_switch_wield_audio_event"`
	WeaponSwitchHolsterAudioEvent string                   `json:"weapon_switch_holster_audio_event"`
	PickupAbility                 enum.AbilityId           `json:"pickup_ability"`
	DropAbility                   enum.AbilityId           `json:"drop_ability"`
	GripType                      enum.WeaponGripType      `json:"grip_type"`
	IsOneHanded                   bool                     `json:"is_one_handed"`
	AllowWieldOnShip              bool                     `json:"allow_wield_on_ship"`
	ShowInPickupWidget            bool                     `json:"show_in_pickup_widget"`
	HolsterUnit                   string                   `json:"holster_unit"`
	HintId                        string                   `json:"hint_id"`
	UnknownHash4                  string                   `json:"unknown_hash4"`
	EquipmentWwiseIndex           enum.EquipmentWwiseIndex `json:"equipment_wwise_index"`
	BackpackWwiseIndex            enum.BackpackWwiseIndex  `json:"backpack_wwise_index"`
	UseBackpackFilledSwitch       bool                     `json:"use_backpack_filled_switch"`
	UnknownHash5                  string                   `json:"unknown_hash5"`
}

func (w EquipmentComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleEquipmentComponent{
		WieldNode:                     lookupThinHash(w.WieldNode),
		WieldNodeRotationOffset:       w.WieldNodeRotationOffset,
		WieldNodePositionOffset:       w.WieldNodePositionOffset,
		HolsterNode:                   lookupThinHash(w.HolsterNode),
		HolsterNodeRotationOffset:     w.HolsterNodeRotationOffset,
		HolsterNodePositionOffset:     w.HolsterNodePositionOffset,
		HolsterDuration:               w.HolsterDuration,
		WieldDuration:                 w.WieldDuration,
		PickupDuration:                w.PickupDuration,
		UnknownHash:                   lookupThinHash(w.UnknownHash),
		WieldAnimation:                lookupThinHash(w.WieldAnimation),
		PickupAnimation:               lookupThinHash(w.PickupAnimation),
		HolsterAnimation:              lookupThinHash(w.HolsterAnimation),
		DropAnimation:                 lookupThinHash(w.DropAnimation),
		AmmocheckAnimation:            lookupThinHash(w.AmmocheckAnimation),
		UnknownBool:                   w.UnknownBool != 0,
		UnknownBool2:                  w.UnknownBool2 != 0,
		HolsterDurationForAttach:      w.HolsterDurationForAttach,
		WieldDurationForAttach:        w.WieldDurationForAttach,
		AttachLerpTime:                w.AttachLerpTime,
		OnHolsterAnimationEvent:       lookupThinHash(w.OnHolsterAnimationEvent),
		WieldAnimationSelf:            lookupThinHash(w.WieldAnimationSelf),
		PickupAnimationSelf:           lookupThinHash(w.PickupAnimationSelf),
		HolsterAnimationSelf:          lookupThinHash(w.HolsterAnimationSelf),
		UnknownHash3:                  lookupThinHash(w.UnknownHash3),
		EquipmentType:                 w.EquipmentType,
		DropMode:                      w.DropMode,
		DropAudioEvent:                lookupThinHash(w.DropAudioEvent),
		UnknownAudioEvent:             lookupThinHash(w.UnknownAudioEvent),
		DropIconUi:                    lookupHash(w.DropIconUi),
		PickupAudioEvent:              lookupThinHash(w.PickupAudioEvent),
		UnknownEvent2:                 lookupThinHash(w.UnknownEvent2),
		UnknownEvent3:                 lookupThinHash(w.UnknownEvent3),
		UnknownEvent4:                 lookupThinHash(w.UnknownEvent4),
		WeaponSwitchWieldAudioEvent:   lookupThinHash(w.WeaponSwitchWieldAudioEvent),
		WeaponSwitchHolsterAudioEvent: lookupThinHash(w.WeaponSwitchHolsterAudioEvent),
		PickupAbility:                 w.PickupAbility,
		DropAbility:                   w.DropAbility,
		GripType:                      w.GripType,
		IsOneHanded:                   w.IsOneHanded != 0,
		AllowWieldOnShip:              w.AllowWieldOnShip != 0,
		ShowInPickupWidget:            w.ShowInPickupWidget != 0,
		HolsterUnit:                   lookupHash(w.HolsterUnit),
		HintId:                        lookupThinHash(w.HintId),
		UnknownHash4:                  lookupThinHash(w.UnknownHash4),
		EquipmentWwiseIndex:           w.EquipmentWwiseIndex,
		BackpackWwiseIndex:            w.BackpackWwiseIndex,
		UseBackpackFilledSwitch:       w.UseBackpackFilledSwitch != 0,
		UnknownHash5:                  lookupThinHash(w.UnknownHash5),
	}
}

func getEquipmentComponentData() ([]byte, error) {
	equipmentComponentHash := Sum("EquipmentComponentData")
	equipmentComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(equipmentComponentHashData, binary.LittleEndian, equipmentComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, equipmentComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getEquipmentComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("EquipmentComponentData")
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
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("EquipmentComponent") {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (data type was not EquipmentComponent)")
	}

	equipmentComponentData, err := getEquipmentComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get equipment component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(equipmentComponentData)

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
		return nil, fmt.Errorf("%v not found in equipment component data", hash.String())
	}

	var equipmentComponentType DLTypeDesc
	equipmentComponentType, ok = typelib.Types[Sum("EquipmentComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find EquipmentComponent hash in dl_library")
	}

	componentData := make([]byte, equipmentComponentType.Size)
	if _, err := r.Seek(int64(equipmentComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseEquipmentComponents() (map[stingray.Hash]EquipmentComponent, error) {
	unitHash := Sum("EquipmentComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var equipmentType DLTypeDesc
	var ok bool
	equipmentType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find EquipmentComponentData hash in dl_library")
	}

	if len(equipmentType.Members) != 2 {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (there should be 2 members but were actually %v)", len(equipmentType.Members))
	}

	if equipmentType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (hashmap atom was not inline array)")
	}

	if equipmentType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (data atom was not inline array)")
	}

	if equipmentType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (hashmap storage was not struct)")
	}

	if equipmentType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (data storage was not struct)")
	}

	if equipmentType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if equipmentType.Members[1].TypeID != Sum("EquipmentComponent") {
		return nil, fmt.Errorf("EquipmentComponentData unexpected format (data type was not EquipmentComponent)")
	}

	equipmentComponentData, err := getEquipmentComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get equipment component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(equipmentComponentData)

	hashmap := make([]ComponentIndexData, equipmentType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]EquipmentComponent, equipmentType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]EquipmentComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
