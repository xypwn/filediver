package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type WeaponMagazineAnimEvent struct {
	Type                  enum.WeaponReloadEventType // Type of reload.
	AnimationEventWeapon  stingray.ThinHash          // [string]Animatiom event to trigger.
	AnimationEventWielder stingray.ThinHash          // [string]Animatiom event to trigger.
}

type WeaponReloadComponent struct {
	ManualClearing           uint8 // [bool]If this is true, the rules for fast/slow reload change
	ReloadAllowMove          uint8 // [bool]Whether or not the player can move while reloading this weapon.
	_                        [2]uint8
	Ability                  enum.AbilityId             // The ability to play on the wielder when reloading this weapon.
	ReloadAnimEvents         [4]WeaponMagazineAnimEvent // The animation event that will be triggered, depending on the reload type
	Duration                 float32                    // The duration of the reload. We scale the reload ability to match this duration. If 0, use the default ability duration (no scaling).
	HasSharedDeposit         uint8                      // [bool]Should this unit look for a (shared) deposit on the entity it it mounted on when checking for ammo?
	_                        [3]uint8
	ReloadVONormal           stingray.ThinHash // [string]VO event to play when doing a normal reload.
	ReloadVOLastmag          stingray.ThinHash // [string]VO event to play when doing a reload while being on your last mag (1 mag left).
	ReloadVONoMags           stingray.ThinHash // [string]VO event to play when trying to reload but have no mags.
	ReloadVONoMagsNoBackpack stingray.ThinHash // [string]If VO event is present we play this when trying to reload but have no mags and no backpack.
}

type SimpleWeaponMagazineAnimEvent struct {
	Type                  enum.WeaponReloadEventType `json:"type"`                    // Type of reload.
	AnimationEventWeapon  string                     `json:"animation_event_weapon"`  // [string]Animatiom event to trigger.
	AnimationEventWielder string                     `json:"animation_event_wielder"` // [string]Animatiom event to trigger.
}

type SimpleWeaponReloadComponent struct {
	ManualClearing           bool                            `json:"manual_clearing"`               // [bool]If this is true, the rules for fast/slow reload change
	ReloadAllowMove          bool                            `json:"reload_allow_move"`             // [bool]Whether or not the player can move while reloading this weapon.
	Ability                  enum.AbilityId                  `json:"ability"`                       // The ability to play on the wielder when reloading this weapon.
	ReloadAnimEvents         []SimpleWeaponMagazineAnimEvent `json:"reload_anim_events,omitempty"`  // The animation event that will be triggered, depending on the reload type
	Duration                 float32                         `json:"duration"`                      // The duration of the reload. We scale the reload ability to match this duration. If 0, use the default ability duration (no scaling).
	HasSharedDeposit         bool                            `json:"has_shared_deposit"`            // [bool]Should this unit look for a (shared) deposit on the entity it it mounted on when checking for ammo?
	ReloadVONormal           string                          `json:"reload_vo_normal"`              // [string]VO event to play when doing a normal reload.
	ReloadVOLastmag          string                          `json:"reload_vo_lastmag"`             // [string]VO event to play when doing a reload while being on your last mag (1 mag left).
	ReloadVONoMags           string                          `json:"reload_vo_no_mags"`             // [string]VO event to play when trying to reload but have no mags.
	ReloadVONoMagsNoBackpack string                          `json:"reload_vo_no_mags_no_backpack"` // [string]If VO event is present we play this when trying to reload but have no mags and no backpack.
}

func (w WeaponReloadComponent) ToSimple(_ HashLookup, lookupThinHash ThinHashLookup, _ StringsLookup) any {
	reloadAnimEvents := make([]SimpleWeaponMagazineAnimEvent, 0)
	for _, event := range w.ReloadAnimEvents {
		if event.Type == enum.WeaponReloadEventType_None {
			break
		}
		reloadAnimEvents = append(reloadAnimEvents, SimpleWeaponMagazineAnimEvent{
			Type:                  event.Type,
			AnimationEventWeapon:  lookupThinHash(event.AnimationEventWeapon),
			AnimationEventWielder: lookupThinHash(event.AnimationEventWielder),
		})
	}

	return SimpleWeaponReloadComponent{
		ManualClearing:           w.ManualClearing != 0,
		ReloadAllowMove:          w.ReloadAllowMove != 0,
		Ability:                  w.Ability,
		ReloadAnimEvents:         reloadAnimEvents,
		Duration:                 w.Duration,
		HasSharedDeposit:         w.HasSharedDeposit != 0,
		ReloadVONormal:           lookupThinHash(w.ReloadVONormal),
		ReloadVOLastmag:          lookupThinHash(w.ReloadVOLastmag),
		ReloadVONoMags:           lookupThinHash(w.ReloadVONoMags),
		ReloadVONoMagsNoBackpack: lookupThinHash(w.ReloadVONoMagsNoBackpack),
	}
}

func getWeaponReloadComponentData() ([]byte, error) {
	weaponReloadHash := Sum("WeaponReloadComponentData")
	weaponReloadHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponReloadHashData, binary.LittleEndian, weaponReloadHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponReloadHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponReloadComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponReloadCmpDataHash := Sum("WeaponReloadComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponReloadCmpDataType DLTypeDesc
	var ok bool
	weaponReloadCmpDataType, ok = typelib.Types[WeaponReloadCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(weaponReloadCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponReloadCmpDataType.Members))
	}

	if weaponReloadCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponReloadCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (data atom was not inline array)")
	}

	if weaponReloadCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponReloadCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (data storage was not struct)")
	}

	if weaponReloadCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponReloadCmpDataType.Members[1].TypeID != Sum("WeaponReloadComponent") {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (data type was not WeaponReloadComponent)")
	}

	weaponReloadComponentData, err := getWeaponReloadComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon reload component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponReloadComponentData)

	hashmap := make([]ComponentIndexData, weaponReloadCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon reload component data", hash.String())
	}

	var weaponReloadComponentType DLTypeDesc
	weaponReloadComponentType, ok = typelib.Types[Sum("WeaponReloadComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponReloadComponent hash in dl_library")
	}

	componentData := make([]byte, weaponReloadComponentType.Size)
	if _, err := r.Seek(int64(weaponReloadComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponReloadComponents() (map[stingray.Hash]WeaponReloadComponent, error) {
	weaponReloadHash := Sum("WeaponReloadComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponReloadType DLTypeDesc
	var ok bool
	weaponReloadType, ok = typelib.Types[weaponReloadHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponReloadComponentData hash in dl_library")
	}

	if len(weaponReloadType.Members) != 2 {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponReloadType.Members))
	}

	if weaponReloadType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponReloadType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (data atom was not inline array)")
	}

	if weaponReloadType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponReloadType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (data storage was not struct)")
	}

	if weaponReloadType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponReloadType.Members[1].TypeID != Sum("WeaponReloadComponent") {
		return nil, fmt.Errorf("WeaponReloadComponentData unexpected format (data type was not WeaponReloadComponent)")
	}

	weaponReloadComponentData, err := getWeaponReloadComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponReloadComponentData)

	hashmap := make([]ComponentIndexData, weaponReloadType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponReloadComponent, weaponReloadType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponReloadComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
