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

type CameraShake struct {
	Shake          stingray.Hash     // [camera_shake]
	NodeId         stingray.ThinHash // [string]
	Offset         mgl32.Vec3
	InnerRadius    float32
	InnerOuter     float32
	OrphanedPolicy enum.UnitEffectOrphanStrategy
	Linked         uint8
	_              [3]uint8
}

type SimpleCameraShake struct {
	Shake          string                        `json:"shake"`   // [camera_shake]
	NodeId         string                        `json:"node_id"` // [string]
	Offset         mgl32.Vec3                    `json:"offset"`
	InnerRadius    float32                       `json:"inner_radius"`
	InnerOuter     float32                       `json:"inner_outer"`
	OrphanedPolicy enum.UnitEffectOrphanStrategy `json:"orphaned_policy"`
	Linked         bool                          `json:"linked"`
}

type CameraShakeSettings struct {
	ID        stingray.ThinHash // [string]The id of this effect. Referenced when playing/stopping this particle effect.
	_         [4]uint8
	Settings  CameraShake                   // Camera shake settings.
	OnDestroy enum.UnitEffectOrphanStrategy // The strategy on how to handle the camera shake when this entity is destroyed.
	OnReplace enum.UnitEffectOrphanStrategy // The strategy on how to handle the camera shake when this entity is trying to play an effect that already exists.
	OnDeath   enum.UnitEffectOrphanStrategy // The strategy on how to handle the camera shake when this entity dies.
}

type SimpleCameraShakeSettings struct {
	ID        string                        `json:"id"`         // [string]The id of this effect. Referenced when playing/stopping this particle effect.
	Settings  SimpleCameraShake             `json:"settings"`   // Camera shake settings.
	OnDestroy enum.UnitEffectOrphanStrategy `json:"on_destroy"` // The strategy on how to handle the camera shake when this entity is destroyed.
	OnReplace enum.UnitEffectOrphanStrategy `json:"on_replace"` // The strategy on how to handle the camera shake when this entity is trying to play an effect that already exists.
	OnDeath   enum.UnitEffectOrphanStrategy `json:"on_death"`   // The strategy on how to handle the camera shake when this entity dies.
}

func (w CameraShakeSettings) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleCameraShakeSettings {
	return SimpleCameraShakeSettings{
		ID: lookupThinHash(w.ID),
		Settings: SimpleCameraShake{
			Shake:          lookupHash(w.Settings.Shake),
			NodeId:         lookupThinHash(w.Settings.NodeId),
			Offset:         w.Settings.Offset,
			InnerRadius:    w.Settings.InnerRadius,
			InnerOuter:     w.Settings.InnerOuter,
			OrphanedPolicy: w.Settings.OrphanedPolicy,
			Linked:         w.Settings.Linked != 0,
		},
		OnDestroy: w.OnDestroy,
		OnReplace: w.OnReplace,
		OnDeath:   w.OnDeath,
	}
}

type EffectReferenceComponent struct {
	CallOnDestroyOnEntityRemoval uint8 // [bool]Should this entity being removed count as the entity being destroyed
	_                            [7]uint8
	Effects                      [32]ParticleEffectSetting  // Particle Effect settings.
	Units                        [8]SpawnUnitEffectSettings // The settings for units to spawn.
	CameraShakes                 [4]CameraShakeSettings     // The settings for camera shakes to spawn/not spawn..
}

type SimpleEffectReferenceComponent struct {
	CallOnDestroyOnEntityRemoval bool                            `json:"call_on_destroy_on_entity_removal"`
	Effects                      []SimpleParticleEffectSetting   `json:"effects,omitempty"`
	Units                        []SimpleSpawnUnitEffectSettings `json:"units,omitempty"`
	CameraShakes                 []SimpleCameraShakeSettings     `json:"camera_shakes,omitempty"`
}

func (w EffectReferenceComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	simpleEffects := make([]SimpleParticleEffectSetting, 0)
	for _, effect := range w.Effects {
		if effect.ID.Value == 0 {
			break
		}
		simpleEffects = append(simpleEffects, effect.ToSimple(lookupHash, lookupThinHash))
	}

	simpleUnits := make([]SimpleSpawnUnitEffectSettings, 0)
	for _, unit := range w.Units {
		if unit.ID.Value == 0 || unit.ID.Value == 0x4ccb322c {
			break
		}
		simpleUnits = append(simpleUnits, unit.ToSimple(lookupHash, lookupThinHash))
	}

	simpleCameraShakes := make([]SimpleCameraShakeSettings, 0)
	for _, setting := range w.CameraShakes {
		if setting.ID.Value == 0 || setting.ID.Value == 0x4ccb322c {
			break
		}
		simpleCameraShakes = append(simpleCameraShakes, setting.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	return SimpleEffectReferenceComponent{
		CallOnDestroyOnEntityRemoval: w.CallOnDestroyOnEntityRemoval != 0,
		Effects:                      simpleEffects,
		Units:                        simpleUnits,
		CameraShakes:                 simpleCameraShakes,
	}
}

func getEffectReferenceComponentData() ([]byte, error) {
	effectReferenceComponentHash := Sum("EffectReferenceComponentData")
	effectReferenceComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(effectReferenceComponentHashData, binary.LittleEndian, effectReferenceComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, effectReferenceComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getEffectReferenceComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	effectReferenceCmpDataHash := Sum("EffectReferenceComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var effectReferenceCmpDataType DLTypeDesc
	var ok bool
	effectReferenceCmpDataType, ok = typelib.Types[effectReferenceCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(effectReferenceCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (there should be 2 members but were actually %v)", len(effectReferenceCmpDataType.Members))
	}

	if effectReferenceCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (hashmap atom was not inline array)")
	}

	if effectReferenceCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (data atom was not inline array)")
	}

	if effectReferenceCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (hashmap storage was not struct)")
	}

	if effectReferenceCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (data storage was not struct)")
	}

	if effectReferenceCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if effectReferenceCmpDataType.Members[1].TypeID != Sum("EffectReferenceComponent") {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (data type was not EffectReferenceComponent)")
	}

	effectReferenceComponentData, err := getEffectReferenceComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get effect reference component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(effectReferenceComponentData)

	hashmap := make([]ComponentIndexData, effectReferenceCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in effect reference component data", hash.String())
	}

	var effectReferenceComponentType DLTypeDesc
	effectReferenceComponentType, ok = typelib.Types[Sum("EffectReferenceComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find EffectReferenceComponent hash in dl_library")
	}

	componentData := make([]byte, effectReferenceComponentType.Size)
	if _, err := r.Seek(int64(effectReferenceComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseEffectReferenceComponents() (map[stingray.Hash]EffectReferenceComponent, error) {
	unitHash := Sum("EffectReferenceComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var effectReferenceType DLTypeDesc
	var ok bool
	effectReferenceType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find EffectReferenceComponentData hash in dl_library")
	}

	if len(effectReferenceType.Members) != 2 {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (there should be 2 members but were actually %v)", len(effectReferenceType.Members))
	}

	if effectReferenceType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (hashmap atom was not inline array)")
	}

	if effectReferenceType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (data atom was not inline array)")
	}

	if effectReferenceType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (hashmap storage was not struct)")
	}

	if effectReferenceType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (data storage was not struct)")
	}

	if effectReferenceType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if effectReferenceType.Members[1].TypeID != Sum("EffectReferenceComponent") {
		return nil, fmt.Errorf("EffectReferenceComponentData unexpected format (data type was not EffectReferenceComponent)")
	}

	effectReferenceComponentData, err := getEffectReferenceComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get effect reference component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(effectReferenceComponentData)

	hashmap := make([]ComponentIndexData, effectReferenceType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]EffectReferenceComponent, effectReferenceType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]EffectReferenceComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
