package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type EnemyPackageComponent struct {
	BundleTag        stingray.ThinHash // [string]The loadout package will be generated with this bundle tag. Has no runtime effect.
	_                [4]uint8
	Package          stingray.Hash // [package]The loadout package to laod for this entity.
	SubentityPackage stingray.Hash // Name length 17 chars.
	AudioResource    stingray.Hash // [wwise_dep]The wwise bank to include in the package for this entity.
}

type SimpleEnemyPackageComponent struct {
	BundleTag        string `json:"bundle_tag"`
	Package          string `json:"package"`
	SubentityPackage string `json:"subentity_package"`
	AudioResource    string `json:"audio_resource"`
}

func (w EnemyPackageComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleEnemyPackageComponent{
		BundleTag:        lookupThinHash(w.BundleTag),
		Package:          lookupHash(w.Package),
		SubentityPackage: lookupHash(w.SubentityPackage),
		AudioResource:    lookupHash(w.AudioResource),
	}
}

func getEnemyPackageComponentData() ([]byte, error) {
	enemyPackageComponentHash := Sum("EnemyPackageComponentData")
	enemyPackageComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(enemyPackageComponentHashData, binary.LittleEndian, enemyPackageComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, enemyPackageComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getEnemyPackageComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	enemyPackageCmpDataHash := Sum("EnemyPackageComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var enemyPackageCmpDataType DLTypeDesc
	var ok bool
	enemyPackageCmpDataType, ok = typelib.Types[enemyPackageCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(enemyPackageCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (there should be 2 members but were actually %v)", len(enemyPackageCmpDataType.Members))
	}

	if enemyPackageCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (hashmap atom was not inline array)")
	}

	if enemyPackageCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (data atom was not inline array)")
	}

	if enemyPackageCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (hashmap storage was not struct)")
	}

	if enemyPackageCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (data storage was not struct)")
	}

	if enemyPackageCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if enemyPackageCmpDataType.Members[1].TypeID != Sum("EnemyPackageComponent") {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (data type was not EnemyPackageComponent)")
	}

	enemyPackageComponentData, err := getEnemyPackageComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get enemy package component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(enemyPackageComponentData)

	hashmap := make([]ComponentIndexData, enemyPackageCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in enemy package component data", hash.String())
	}

	var enemyPackageComponentType DLTypeDesc
	enemyPackageComponentType, ok = typelib.Types[Sum("EnemyPackageComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find EnemyPackageComponent hash in dl_library")
	}

	componentData := make([]byte, enemyPackageComponentType.Size)
	if _, err := r.Seek(int64(enemyPackageComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseEnemyPackageComponents() (map[stingray.Hash]EnemyPackageComponent, error) {
	unitHash := Sum("EnemyPackageComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var enemyPackageType DLTypeDesc
	var ok bool
	enemyPackageType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find EnemyPackageComponentData hash in dl_library")
	}

	if len(enemyPackageType.Members) != 2 {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (there should be 2 members but were actually %v)", len(enemyPackageType.Members))
	}

	if enemyPackageType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (hashmap atom was not inline array)")
	}

	if enemyPackageType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (data atom was not inline array)")
	}

	if enemyPackageType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (hashmap storage was not struct)")
	}

	if enemyPackageType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (data storage was not struct)")
	}

	if enemyPackageType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if enemyPackageType.Members[1].TypeID != Sum("EnemyPackageComponent") {
		return nil, fmt.Errorf("EnemyPackageComponentData unexpected format (data type was not EnemyPackageComponent)")
	}

	enemyPackageComponentData, err := getEnemyPackageComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get enemy package component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(enemyPackageComponentData)

	hashmap := make([]ComponentIndexData, enemyPackageType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]EnemyPackageComponent, enemyPackageType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]EnemyPackageComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
