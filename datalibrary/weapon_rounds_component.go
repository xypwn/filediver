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

type WeaponRoundsAmmoInfo struct {
	PrimaryRecoilModifiers   RecoilModifiers // Information about the recoil for the primary magazine.
	PrimarySpreadModifiers   SpreadModifiers // Information about the spread for the primary magazine.
	SecondaryRecoilModifiers RecoilModifiers // Information about the recoil for the secondary magazine.
	SecondarySpreadModifiers SpreadModifiers // Information about the spread for the secondary magazine.
}

type WeaponRoundsAmmoType struct {
	PrimaryProjectileType   enum.ProjectileType // Projectile type the first magazine contains.
	AlternateProjectileType enum.ProjectileType // Projectile type the second magazine contains.
}

type WeaponRoundsComponent struct {
	AmmoInfo             WeaponRoundsAmmoInfo // Ammo info for recoil and spread per magazine.
	AmmoType             WeaponRoundsAmmoType // Ammo types per magazine.
	MagazineCapacity     mgl32.Vec2           // Capacity in rounds for each magazine.
	AmmoCapacity         uint32               // Maximum number of rounds.
	AmmoRefill           uint32               // Number of rounds given on refill.
	Ammo                 uint32               // Starting number of rounds.
	ReloadAmount         uint32               // Number of rounds to add to the magazine per reload.
	ReloadThresholds     mgl32.Vec2           // Reload is allowed when less than this amount of rounds are left in the magazine (set per magazine). Defaults to 0 which means 'Same as magazine capacity'.
	Chambered            uint8                // [bool]Can this weapon hold a round in the chamber while reloading. This makes the max amount of bullets capacity + 1 after reload when weapon has rounds remaining
	_                    [3]uint8
	MagazineSwitchAudio  stingray.ThinHash // [wwise]What audio to play when the magazines are switched
	Magazine0WeaponAnim  stingray.ThinHash // [string]What weapon animation to play when switched to magazine 0
	Magazine0WielderAnim stingray.ThinHash // [string]What wielder animation to play when switched to magazine 0
	Magazine1WeaponAnim  stingray.ThinHash // [string]What weapon animation to play when switched to magazine 1
	Magazine1WielderAnim stingray.ThinHash // [string]What wielder animation to play when switched to magazine 1
	MagazineAnimVariable stingray.ThinHash // [string]Do we have an animation variable that we care about?
}

type SimpleWeaponRoundsComponent struct {
	AmmoInfo             WeaponRoundsAmmoInfo `json:"ammo_info"`              // Ammo info for recoil and spread per magazine.
	AmmoType             WeaponRoundsAmmoType `json:"ammo_type"`              // Ammo types per magazine.
	MagazineCapacity     mgl32.Vec2           `json:"magazine_capacity"`      // Capacity in rounds for each magazine.
	AmmoCapacity         uint32               `json:"ammo_capacity"`          // Maximum number of rounds.
	AmmoRefill           uint32               `json:"ammo_refill"`            // Number of rounds given on refill.
	Ammo                 uint32               `json:"ammo"`                   // Starting number of rounds.
	ReloadAmount         uint32               `json:"reload_amount"`          // Number of rounds to add to the magazine per reload.
	ReloadThresholds     mgl32.Vec2           `json:"reload_thresholds"`      // Reload is allowed when less than this amount of rounds are left in the magazine (set per magazine). Defaults to 0 which means 'Same as magazine capacity'.
	Chambered            bool                 `json:"chambered"`              // [bool]Can this weapon hold a round in the chamber while reloading. This makes the max amount of bullets capacity + 1 after reload when weapon has rounds remaining
	MagazineSwitchAudio  string               `json:"magazine_switch_audio"`  // [wwise]What audio to play when the magazines are switched
	Magazine0WeaponAnim  string               `json:"magazine0_weapon_anim"`  // [string]What weapon animation to play when switched to magazine 0
	Magazine0WielderAnim string               `json:"magazine0_wielder_anim"` // [string]What wielder animation to play when switched to magazine 0
	Magazine1WeaponAnim  string               `json:"magazine1_weapon_anim"`  // [string]What weapon animation to play when switched to magazine 1
	Magazine1WielderAnim string               `json:"magazine1_wielder_anim"` // [string]What wielder animation to play when switched to magazine 1
	MagazineAnimVariable string               `json:"magazine_anim_variable"` // [string]Do we have an animation variable that we care about?
}

func (w WeaponRoundsComponent) ToSimple(_ HashLookup, lookupThinHash ThinHashLookup) any {
	return SimpleWeaponRoundsComponent{
		AmmoInfo:             w.AmmoInfo,
		AmmoType:             w.AmmoType,
		MagazineCapacity:     w.MagazineCapacity,
		AmmoCapacity:         w.AmmoCapacity,
		AmmoRefill:           w.AmmoRefill,
		Ammo:                 w.Ammo,
		ReloadAmount:         w.ReloadAmount,
		ReloadThresholds:     w.ReloadThresholds,
		Chambered:            w.Chambered != 0,
		MagazineSwitchAudio:  lookupThinHash(w.MagazineSwitchAudio),
		Magazine0WeaponAnim:  lookupThinHash(w.Magazine0WeaponAnim),
		Magazine0WielderAnim: lookupThinHash(w.Magazine0WielderAnim),
		Magazine1WeaponAnim:  lookupThinHash(w.Magazine1WeaponAnim),
		Magazine1WielderAnim: lookupThinHash(w.Magazine1WielderAnim),
		MagazineAnimVariable: lookupThinHash(w.MagazineAnimVariable),
	}
}

func getWeaponRoundsComponentData() ([]byte, error) {
	weaponRoundsHash := Sum("WeaponRoundsComponentData")
	weaponRoundsHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponRoundsHashData, binary.LittleEndian, weaponRoundsHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponRoundsHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponRoundsComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponRoundsCmpDataHash := Sum("WeaponRoundsComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponRoundsCmpDataType DLTypeDesc
	var ok bool
	weaponRoundsCmpDataType, ok = typelib.Types[WeaponRoundsCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(weaponRoundsCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponRoundsCmpDataType.Members))
	}

	if weaponRoundsCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponRoundsCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (data atom was not inline array)")
	}

	if weaponRoundsCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponRoundsCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (data storage was not struct)")
	}

	if weaponRoundsCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponRoundsCmpDataType.Members[1].TypeID != Sum("WeaponRoundsComponent") {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (data type was not WeaponRoundsComponent)")
	}

	weaponRoundsComponentData, err := getWeaponRoundsComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon rounds component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponRoundsComponentData)

	hashmap := make([]ComponentIndexData, weaponRoundsCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon rounds component data", hash.String())
	}

	var weaponRoundsComponentType DLTypeDesc
	weaponRoundsComponentType, ok = typelib.Types[Sum("WeaponRoundsComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponRoundsComponent hash in dl_library")
	}

	componentData := make([]byte, weaponRoundsComponentType.Size)
	if _, err := r.Seek(int64(weaponRoundsComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponRoundsComponents() (map[stingray.Hash]WeaponRoundsComponent, error) {
	weaponRoundsHash := Sum("WeaponRoundsComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponRoundsType DLTypeDesc
	var ok bool
	weaponRoundsType, ok = typelib.Types[weaponRoundsHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponRoundsComponentData hash in dl_library")
	}

	if len(weaponRoundsType.Members) != 2 {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponRoundsType.Members))
	}

	if weaponRoundsType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponRoundsType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (data atom was not inline array)")
	}

	if weaponRoundsType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponRoundsType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (data storage was not struct)")
	}

	if weaponRoundsType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponRoundsType.Members[1].TypeID != Sum("WeaponRoundsComponent") {
		return nil, fmt.Errorf("WeaponRoundsComponentData unexpected format (data type was not WeaponRoundsComponent)")
	}

	weaponRoundsComponentData, err := getWeaponRoundsComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponRoundsComponentData)

	hashmap := make([]ComponentIndexData, weaponRoundsType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponRoundsComponent, weaponRoundsType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponRoundsComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
