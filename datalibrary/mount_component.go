package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type MountInfo struct {
	Path       stingray.Hash     // [adhd]Path to the adhd-entity to mount.
	AttachNode stingray.ThinHash // [string]Node where to attach the adhd-entity.
	MountSide  enum.MountSide    // Is this mounted on the left side of the parent, the right side of the parent, or neither side in particular?
	Name       stingray.ThinHash // [string]The name to use in scripts to refer to this mount.
	_          [4]uint8
}

type MountComponent struct {
	Infos [5]MountInfo // Path and node info for each entity to mount.
}

type SimpleMountInfo struct {
	Path       string         `json:"path"`
	AttachNode string         `json:"attach_node"`
	MountSide  enum.MountSide `json:"mount_side"`
	Name       string         `json:"name"`
}

type SimpleMountComponent struct {
	Infos []SimpleMountInfo `json:"infos"`
}

func (w MountComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	infos := make([]SimpleMountInfo, 0)
	for _, info := range w.Infos {
		if info.Path.Value == 0 && info.Name.Value == 0 {
			break
		}
		infos = append(infos, SimpleMountInfo{
			Path:       lookupHash(info.Path),
			AttachNode: lookupThinHash(info.AttachNode),
			MountSide:  info.MountSide,
			Name:       lookupThinHash(info.Name),
		})
	}
	return SimpleMountComponent{
		Infos: infos,
	}
}

func getMountComponentData() ([]byte, error) {
	mountComponentHash := Sum("MountComponentData")
	mountComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(mountComponentHashData, binary.LittleEndian, mountComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, mountComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getMountComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("MountComponentData")
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
		return nil, fmt.Errorf("MountComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MountComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MountComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MountComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MountComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MountComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("MountComponent") {
		return nil, fmt.Errorf("MountComponentData unexpected format (data type was not MountComponent)")
	}

	mountComponentData, err := getMountComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get mount component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(mountComponentData)

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
		return nil, fmt.Errorf("%v not found in mount component data", hash.String())
	}

	var mountComponentType DLTypeDesc
	mountComponentType, ok = typelib.Types[Sum("MountComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find MountComponent hash in dl_library")
	}

	componentData := make([]byte, mountComponentType.Size)
	if _, err := r.Seek(int64(mountComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseMountComponents() (map[stingray.Hash]MountComponent, error) {
	unitHash := Sum("MountComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var mountType DLTypeDesc
	var ok bool
	mountType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find MountComponentData hash in dl_library")
	}

	if len(mountType.Members) != 2 {
		return nil, fmt.Errorf("MountComponentData unexpected format (there should be 2 members but were actually %v)", len(mountType.Members))
	}

	if mountType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MountComponentData unexpected format (hashmap atom was not inline array)")
	}

	if mountType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("MountComponentData unexpected format (data atom was not inline array)")
	}

	if mountType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MountComponentData unexpected format (hashmap storage was not struct)")
	}

	if mountType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("MountComponentData unexpected format (data storage was not struct)")
	}

	if mountType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("MountComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if mountType.Members[1].TypeID != Sum("MountComponent") {
		return nil, fmt.Errorf("MountComponentData unexpected format (data type was not MountComponent)")
	}

	mountComponentData, err := getMountComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get mount component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(mountComponentData)

	hashmap := make([]ComponentIndexData, mountType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]MountComponent, mountType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]MountComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
