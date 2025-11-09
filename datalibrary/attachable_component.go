package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type AttachableComponent struct {
	DynamicActor                stingray.ThinHash // [string]Name of the dynamic actor that should be disabled when attached and enabled when dropped.
	AttachedDynamicActor        stingray.ThinHash // [string]Name of the dynamic actor that should be enabled when attached and disabled when dropped.
	UseAllDynamicActors         uint8             // [bool]True if all dynamic actors should be disabled when attached and enabled when dropped.
	_                           [3]uint8
	Offset                      mgl32.Vec3        // If set, this offset is applied on top of any other offsets when attaching the entity.
	RotationOffset              mgl32.Vec3        // If set, this rotation (in angles) is applied on top of any other offsets when attaching the entity.
	OffsetNode                  stingray.ThinHash // [string]If specified, the local space position and rotation of the node relative to the unit root will be added to the offset when attaching.
	AttachAnim                  stingray.ThinHash // [string]Animation to play when attaching.
	ParentAttachAnim            stingray.ThinHash // [string]Animation to play on the parent when attaching.
	DetachAnim                  stingray.ThinHash // [string]Animation to play when detaching.
	ParentDetachAnim            stingray.ThinHash // [string]Animation to play on the parent when detaching.
	AffectScale                 uint8             // [bool]This should only be applied to main entities that will attach to something in the world, like hellpods.
	IgnoreWeaponDamageForParent uint8             // [bool]This variable will decide if we should be able to damage our parent with our weapons.
	InheritAngularMomentum      uint8             // [bool]Should this attachable inherit angular momentum?
	UnknownBool1                uint8             // [bool]Unknown, name length 21
	UnknownBool2                uint8             // [bool]Unknown, name length 24
	UnknownBool3                uint8             // [bool]Unknown, name length 24
}

type SimpleAttachableComponent struct {
	DynamicActor                string     `json:"dynamic_actor"`
	AttachedDynamicActor        string     `json:"attached_dynamic_actor"`
	UseAllDynamicActors         bool       `json:"use_all_dynamic_actors"`
	Offset                      mgl32.Vec3 `json:"offset"`
	RotationOffset              mgl32.Vec3 `json:"rotation_offset"`
	OffsetNode                  string     `json:"offset_node"`
	AttachAnim                  string     `json:"attach_anim"`
	ParentAttachAnim            string     `json:"parent_attach_anim"`
	DetachAnim                  string     `json:"detach_anim"`
	ParentDetachAnim            string     `json:"parent_detach_anim"`
	AffectScale                 bool       `json:"affect_scale"`
	IgnoreWeaponDamageForParent bool       `json:"ignore_weapon_damage_for_parent"`
	InheritAngularMomentum      bool       `json:"inherit_angular_momentum"`
	UnknownBool1                bool       `json:"unknown_bool1"`
	UnknownBool2                bool       `json:"unknown_bool2"`
	UnknownBool3                bool       `json:"unknown_bool3"`
}

func (w AttachableComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleAttachableComponent{
		DynamicActor:                lookupThinHash(w.DynamicActor),
		AttachedDynamicActor:        lookupThinHash(w.AttachedDynamicActor),
		UseAllDynamicActors:         w.UseAllDynamicActors != 0,
		Offset:                      w.Offset,
		RotationOffset:              w.RotationOffset,
		OffsetNode:                  lookupThinHash(w.OffsetNode),
		AttachAnim:                  lookupThinHash(w.AttachAnim),
		ParentAttachAnim:            lookupThinHash(w.ParentAttachAnim),
		DetachAnim:                  lookupThinHash(w.DetachAnim),
		ParentDetachAnim:            lookupThinHash(w.ParentDetachAnim),
		AffectScale:                 w.AffectScale != 0,
		IgnoreWeaponDamageForParent: w.IgnoreWeaponDamageForParent != 0,
		InheritAngularMomentum:      w.InheritAngularMomentum != 0,
		UnknownBool1:                w.UnknownBool1 != 0,
		UnknownBool2:                w.UnknownBool2 != 0,
		UnknownBool3:                w.UnknownBool3 != 0,
	}
}

func getAttachableComponentData() ([]byte, error) {
	attachableComponentHash := Sum("AttachableComponentData")
	attachableComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(attachableComponentHashData, binary.LittleEndian, attachableComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, attachableComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getAttachableComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("AttachableComponentData")
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
		return nil, fmt.Errorf("AttachableComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("AttachableComponent") {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (data type was not AttachableComponent)")
	}

	attachableComponentData, err := getAttachableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get attachable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(attachableComponentData)

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
		return nil, fmt.Errorf("%v not found in attachable component data", hash.String())
	}

	var attachableComponentType DLTypeDesc
	attachableComponentType, ok = typelib.Types[Sum("AttachableComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find AttachableComponent hash in dl_library")
	}

	componentData := make([]byte, attachableComponentType.Size)
	if _, err := r.Seek(int64(attachableComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseAttachableComponents() (map[stingray.Hash]AttachableComponent, error) {
	unitHash := Sum("AttachableComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var attachableType DLTypeDesc
	var ok bool
	attachableType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find AttachableComponentData hash in dl_library")
	}

	if len(attachableType.Members) != 2 {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (there should be 2 members but were actually %v)", len(attachableType.Members))
	}

	if attachableType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (hashmap atom was not inline array)")
	}

	if attachableType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (data atom was not inline array)")
	}

	if attachableType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (hashmap storage was not struct)")
	}

	if attachableType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (data storage was not struct)")
	}

	if attachableType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if attachableType.Members[1].TypeID != Sum("AttachableComponent") {
		return nil, fmt.Errorf("AttachableComponentData unexpected format (data type was not AttachableComponent)")
	}

	attachableComponentData, err := getAttachableComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get attachable component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(attachableComponentData)

	hashmap := make([]ComponentIndexData, attachableType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]AttachableComponent, attachableType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]AttachableComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
