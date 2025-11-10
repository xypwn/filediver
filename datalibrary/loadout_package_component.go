package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type LoadoutPackageComponent struct {
	BundleTag                    stingray.ThinHash // [string]The loadout package will be generated with this bundle tag. Has no runtime effect.
	_                            [4]uint8
	Package                      stingray.Hash // [package]The loadout package to laod for this entity.
	AudioResource                stingray.Hash // [wwise_dep]The wwise bank to include in the package for this entity.
	IncludeDefaultCustomizations uint8         // [bool]If set any default customizations are included in the package generated for this .adhd file. Affects the entire package
	UnknownBool                  uint8         // [bool] unknown, 27 chars long
	_                            [6]uint8
}

type SimpleLoadoutPackageComponent struct {
	BundleTag                    string `json:"bundle_tag"`
	Package                      string `json:"package"`
	AudioResource                string `json:"audio_resource"`
	IncludeDefaultCustomizations bool   `json:"include_default_customizations"`
	UnknownBool                  bool   `json:"unknown_bool"`
}

func (w LoadoutPackageComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleLoadoutPackageComponent{
		BundleTag:                    lookupThinHash(w.BundleTag),
		Package:                      lookupHash(w.Package),
		AudioResource:                lookupHash(w.AudioResource),
		IncludeDefaultCustomizations: w.IncludeDefaultCustomizations != 0,
		UnknownBool:                  w.UnknownBool != 0,
	}
}

func getLoadoutPackageComponentData() ([]byte, error) {
	loadoutPackageComponentHash := Sum("LoadoutPackageComponentData")
	loadoutPackageComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(loadoutPackageComponentHashData, binary.LittleEndian, loadoutPackageComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, loadoutPackageComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getLoadoutPackageComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("LoadoutPackageComponentData")
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
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("LoadoutPackageComponent") {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (data type was not LoadoutPackageComponent)")
	}

	loadoutPackageComponentData, err := getLoadoutPackageComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get loadout package component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(loadoutPackageComponentData)

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
		return nil, fmt.Errorf("%v not found in loadout package component data", hash.String())
	}

	var loadoutPackageComponentType DLTypeDesc
	loadoutPackageComponentType, ok = typelib.Types[Sum("LoadoutPackageComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find LoadoutPackageComponent hash in dl_library")
	}

	componentData := make([]byte, loadoutPackageComponentType.Size)
	if _, err := r.Seek(int64(loadoutPackageComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseLoadoutPackageComponents() (map[stingray.Hash]LoadoutPackageComponent, error) {
	unitHash := Sum("LoadoutPackageComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var loadoutPackageType DLTypeDesc
	var ok bool
	loadoutPackageType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find LoadoutPackageComponentData hash in dl_library")
	}

	if len(loadoutPackageType.Members) != 2 {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (there should be 2 members but were actually %v)", len(loadoutPackageType.Members))
	}

	if loadoutPackageType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (hashmap atom was not inline array)")
	}

	if loadoutPackageType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (data atom was not inline array)")
	}

	if loadoutPackageType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (hashmap storage was not struct)")
	}

	if loadoutPackageType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (data storage was not struct)")
	}

	if loadoutPackageType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if loadoutPackageType.Members[1].TypeID != Sum("LoadoutPackageComponent") {
		return nil, fmt.Errorf("LoadoutPackageComponentData unexpected format (data type was not LoadoutPackageComponent)")
	}

	loadoutPackageComponentData, err := getLoadoutPackageComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get loadout package component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(loadoutPackageComponentData)

	hashmap := make([]ComponentIndexData, loadoutPackageType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]LoadoutPackageComponent, loadoutPackageType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]LoadoutPackageComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
