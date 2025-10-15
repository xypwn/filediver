package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type ArcWeaponComponent struct {
	Type                        enum.ArcType // Type of arc it fires
	RoundsPerMinute             float32      // Rounds per minute depending on weapon setting.
	InfiniteAmmo                uint8        // [bool]True if this projectile weapon can never run out of ammo.
	RPCSyncedFireEvents         uint8        // [bool]ONLY USE FOR SINGLE-FIRE/SLOW FIRING WEAPONS. Primarily useful for sniper rifles, explosive one-shots etc. that need the firing event to be highly accurately synced!
	_                           [2]uint8
	FireSingleAudioEvent        stingray.ThinHash     // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	FireFailAudioEvent          stingray.ThinHash     // [wwise]The audio event to trigger when the arc fails to hit anything on the first shot.
	HapticsFireSingleAudioEvent stingray.ThinHash     // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	FireSourceNode              stingray.ThinHash     // [string]The node to play the firing audio events at.
	OnRoundFiredShakes          WeaponCameraShakeInfo // Settings for local and in-world camera shakes to play on every round fired.
	MuzzleFlash                 stingray.Hash         // [particles]Particle effect of the muzzle flash, played on attach_muzzle.
	MuzzleFlashFail             stingray.Hash         // [particles]Particle effect of the muzzle flash, when the arc weapon fails to hit, played on attach_muzzle.
}

type SimpleArcWeaponComponent struct {
	Type                        enum.ArcType                `json:"arc_type"`                        // Type of arc it fires
	RoundsPerMinute             float32                     `json:"rounds_per_minute"`               // Rounds per minute depending on weapon setting.
	InfiniteAmmo                bool                        `json:"infinite_ammo"`                   // [bool]True if this projectile weapon can never run out of ammo.
	RPCSyncedFireEvents         bool                        `json:"rpc_synced_fire_events"`          // [bool]ONLY USE FOR SINGLE-FIRE/SLOW FIRING WEAPONS. Primarily useful for sniper rifles, explosive one-shots etc. that need the firing event to be highly accurately synced!
	FireSingleAudioEvent        string                      `json:"fire_single_audio_event"`         // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	FireFailAudioEvent          string                      `json:"fire_fail_audio_event"`           // [wwise]The audio event to trigger when the arc fails to hit anything on the first shot.
	HapticsFireSingleAudioEvent string                      `json:"haptics_fire_single_audio_event"` // [wwise]The audio event to trigger when doing single-fire (if we don't have looping sounds).
	FireSourceNode              string                      `json:"fire_source_node"`                // [string]The node to play the firing audio events at.
	OnRoundFiredShakes          SimpleWeaponCameraShakeInfo `json:"on_round_fired_shakes"`           // Settings for local and in-world camera shakes to play on every round fired.
	MuzzleFlash                 string                      `json:"muzzle_flash"`                    // [particles]Particle effect of the muzzle flash, played on attach_muzzle.
	MuzzleFlashFail             string                      `json:"muzzle_flash_fail"`               // [particles]Particle effect of the muzzle flash, when the arc weapon fails to hit, played on attach_muzzle.
}

func (a ArcWeaponComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup) any {
	return SimpleArcWeaponComponent{
		Type:                        a.Type,
		RoundsPerMinute:             a.RoundsPerMinute,
		InfiniteAmmo:                a.InfiniteAmmo != 0,
		RPCSyncedFireEvents:         a.RPCSyncedFireEvents != 0,
		FireSingleAudioEvent:        lookupThinHash(a.FireSingleAudioEvent),
		FireFailAudioEvent:          lookupThinHash(a.FireFailAudioEvent),
		HapticsFireSingleAudioEvent: lookupThinHash(a.HapticsFireSingleAudioEvent),
		FireSourceNode:              lookupThinHash(a.FireSourceNode),
		OnRoundFiredShakes: SimpleWeaponCameraShakeInfo{
			WorldShakeEffect: lookupHash(a.OnRoundFiredShakes.WorldShakeEffect),
			LocalShakeEffect: lookupHash(a.OnRoundFiredShakes.LocalShakeEffect),
			FPVShakeEffect:   lookupHash(a.OnRoundFiredShakes.FPVShakeEffect),
			InnerRadius:      a.OnRoundFiredShakes.InnerRadius,
			OuterRadius:      a.OnRoundFiredShakes.OuterRadius,
		},
		MuzzleFlash:     lookupHash(a.MuzzleFlash),
		MuzzleFlashFail: lookupHash(a.MuzzleFlashFail),
	}
}

func getArcWeaponComponentData() ([]byte, error) {
	arcWeaponHash := Sum("ArcWeaponComponentData")
	arcWeaponHashData := make([]byte, 4)
	if _, err := binary.Encode(arcWeaponHashData, binary.LittleEndian, arcWeaponHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, arcWeaponHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getArcWeaponComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	ArcWeaponCmpDataHash := Sum("ArcWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var arcWeaponCmpDataType DLTypeDesc
	var ok bool
	arcWeaponCmpDataType, ok = typelib.Types[ArcWeaponCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ArcWeaponComponentData hash in dl_library")
	}

	if len(arcWeaponCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(arcWeaponCmpDataType.Members))
	}

	if arcWeaponCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if arcWeaponCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if arcWeaponCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if arcWeaponCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (data storage was not struct)")
	}

	if arcWeaponCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if arcWeaponCmpDataType.Members[1].TypeID != Sum("ArcWeaponComponent") {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (data type was not ArcWeaponComponent)")
	}

	arcWeaponComponentData, err := getArcWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get arc weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(arcWeaponComponentData)

	hashmap := make([]ComponentIndexData, arcWeaponCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in arc weapon component data", hash.String())
	}

	var arcWeaponComponentType DLTypeDesc
	arcWeaponComponentType, ok = typelib.Types[Sum("ArcWeaponComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find ArcWeaponComponent hash in dl_library")
	}

	componentData := make([]byte, arcWeaponComponentType.Size)
	if _, err := r.Seek(int64(arcWeaponComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseArcWeaponComponents() (map[stingray.Hash]ArcWeaponComponent, error) {
	arcWeaponHash := Sum("ArcWeaponComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var arcWeaponType DLTypeDesc
	var ok bool
	arcWeaponType, ok = typelib.Types[arcWeaponHash]
	if !ok {
		return nil, fmt.Errorf("could not find ArcWeaponComponentData hash in dl_library")
	}

	if len(arcWeaponType.Members) != 2 {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (there should be 2 members but were actually %v)", len(arcWeaponType.Members))
	}

	if arcWeaponType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (hashmap atom was not inline array)")
	}

	if arcWeaponType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (data atom was not inline array)")
	}

	if arcWeaponType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (hashmap storage was not struct)")
	}

	if arcWeaponType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (data storage was not struct)")
	}

	if arcWeaponType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if arcWeaponType.Members[1].TypeID != Sum("ArcWeaponComponent") {
		return nil, fmt.Errorf("ArcWeaponComponentData unexpected format (data type was not ArcWeaponComponent)")
	}

	arcWeaponComponentData, err := getArcWeaponComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(arcWeaponComponentData)

	hashmap := make([]ComponentIndexData, arcWeaponType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]ArcWeaponComponent, arcWeaponType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]ArcWeaponComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
