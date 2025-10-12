package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type MagazinePattern struct {
	Projectiles     [32]enum.ProjectileType // Pattern of projectiles to fire. None denotes end if the full size is not used. Pattern is repeated and aligned so the last projectile in the pattern is always the last in the magazine (should it not divide evenly).
	FirstProjectile enum.ProjectileType     // to trigger start/stop event
}

type WeaponMagazineComponent struct {
	Type            enum.MagazineType // Type of magazine.
	Pattern         MagazinePattern   // Only used if magazine type is Pattern.
	Capacity        uint32            // Number of rounds in one magazine.
	Magazines       uint32            // Starting number of magazines
	MagazinesRefill uint32            // Number of magazines given on refill.
	MagazinesMax    uint32            // Maximum number of magazines.
	ReloadThreshold uint32            // Reload is allowed when less than this amount of rounds are left in the clip. Defaults to 0 which means 'Same as clip capacity'.
	Chambered       uint8             // [bool]Can this weapon hold a round in the chamber while reloading. This makes the max amount of bullets capacity + 1 after reload when weapon has rounds remaining
	UnknownBool     uint8
	_               [2]uint8
}

func getWeaponMagazineComponentData() ([]byte, error) {
	weaponMagazineHash := Sum("WeaponMagazineComponentData")
	weaponMagazineHashData := make([]byte, 4)
	if _, err := binary.Encode(weaponMagazineHashData, binary.LittleEndian, weaponMagazineHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, weaponMagazineHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getWeaponMagazineComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	WeaponMagazineCmpDataHash := Sum("WeaponMagazineComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponMagazineCmpDataType DLTypeDesc
	var ok bool
	weaponMagazineCmpDataType, ok = typelib.Types[WeaponMagazineCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(weaponMagazineCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponMagazineCmpDataType.Members))
	}

	if weaponMagazineCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponMagazineCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (data atom was not inline array)")
	}

	if weaponMagazineCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponMagazineCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (data storage was not struct)")
	}

	if weaponMagazineCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponMagazineCmpDataType.Members[1].TypeID != Sum("WeaponMagazineComponent") {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (data type was not WeaponMagazineComponent)")
	}

	weaponMagazineComponentData, err := getWeaponMagazineComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get weapon magazine component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponMagazineComponentData)

	hashmap := make([]ComponentIndexData, weaponMagazineCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in weapon magazine component data", hash.String())
	}

	var weaponMagazineComponentType DLTypeDesc
	weaponMagazineComponentType, ok = typelib.Types[Sum("WeaponMagazineComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponMagazineComponent hash in dl_library")
	}

	componentData := make([]byte, weaponMagazineComponentType.Size)
	if _, err := r.Seek(int64(weaponMagazineComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseWeaponMagazineComponents() (map[stingray.Hash]WeaponMagazineComponent, error) {
	weaponMagazineHash := Sum("WeaponMagazineComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var weaponMagazineType DLTypeDesc
	var ok bool
	weaponMagazineType, ok = typelib.Types[weaponMagazineHash]
	if !ok {
		return nil, fmt.Errorf("could not find WeaponMagazineComponentData hash in dl_library")
	}

	if len(weaponMagazineType.Members) != 2 {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (there should be 2 members but were actually %v)", len(weaponMagazineType.Members))
	}

	if weaponMagazineType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (hashmap atom was not inline array)")
	}

	if weaponMagazineType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (data atom was not inline array)")
	}

	if weaponMagazineType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (hashmap storage was not struct)")
	}

	if weaponMagazineType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (data storage was not struct)")
	}

	if weaponMagazineType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if weaponMagazineType.Members[1].TypeID != Sum("WeaponMagazineComponent") {
		return nil, fmt.Errorf("WeaponMagazineComponentData unexpected format (data type was not WeaponMagazineComponent)")
	}

	weaponMagazineComponentData, err := getWeaponMagazineComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get projectile weapon component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(weaponMagazineComponentData)

	hashmap := make([]ComponentIndexData, weaponMagazineType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]WeaponMagazineComponent, weaponMagazineType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]WeaponMagazineComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
