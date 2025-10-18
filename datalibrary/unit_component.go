package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type UnitComponent struct {
	UnitPath                          stingray.Hash         // [unit]Path to the unit for this entity.
	ScaleMin                          float32               // Min scale value.
	ScaleMax                          float32               // Max scale value.
	Radius                            float32               // The radius of the unit as used by e.g. motion and navigation
	HiddenVisibilityGroups            [20]stingray.ThinHash // [string]What groups to hide when this unit spawns?
	ShowRandomVisibilityGroup         [8]stingray.ThinHash  // [string]Upon spawn, picks a random visibility group and sets it visible.
	OnHotjoinCorpseRepairAbilityPatch enum.AbilityId        // This ability gets ran on hot joining players when the entity is a corpse.
}

type SimpleUnitComponent struct {
	UnitPath                          string         `json:"unit_path"`
	ScaleMin                          float32        `json:"scale_min"`
	ScaleMax                          float32        `json:"scale_max"`
	Radius                            float32        `json:"radius"`
	HiddenVisiblityGroups             []string       `json:"hidden_visibility_groups,omitempty"`
	ShowRandomVisibilityGroup         []string       `json:"show_random_visibility_group,omitempty"`
	OnHotjoinCorpseRepairAbilityPatch enum.AbilityId `json:"on_hotjoin_corpse_repair_ability_patch"`
}

func (w UnitComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	hiddenVisiblityGroups := make([]string, 0)
	for _, hash := range w.HiddenVisibilityGroups {
		if hash.Value == 0 {
			break
		}
		hiddenVisiblityGroups = append(hiddenVisiblityGroups, lookupThinHash(hash))
	}
	showRandomVisibilityGroup := make([]string, 0)
	for _, hash := range w.ShowRandomVisibilityGroup {
		if hash.Value == 0 {
			break
		}
		showRandomVisibilityGroup = append(showRandomVisibilityGroup, lookupThinHash(hash))
	}
	return SimpleUnitComponent{
		UnitPath:                          lookupHash(w.UnitPath),
		ScaleMin:                          w.ScaleMin,
		ScaleMax:                          w.ScaleMax,
		Radius:                            w.Radius,
		HiddenVisiblityGroups:             hiddenVisiblityGroups,
		ShowRandomVisibilityGroup:         showRandomVisibilityGroup,
		OnHotjoinCorpseRepairAbilityPatch: w.OnHotjoinCorpseRepairAbilityPatch,
	}
}

func getUnitComponentData() ([]byte, error) {
	unitComponentHash := Sum("UnitComponentData")
	unitComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(unitComponentHashData, binary.LittleEndian, unitComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, unitComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getUnitComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("UnitComponentData")
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
		return nil, fmt.Errorf("UnitComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("UnitComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("UnitComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("UnitComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("UnitComponent") {
		return nil, fmt.Errorf("UnitComponentData unexpected format (data type was not UnitComponent)")
	}

	unitComponentData, err := getUnitComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get unit component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(unitComponentData)

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
		return nil, fmt.Errorf("%v not found in unit component data", hash.String())
	}

	var unitComponentType DLTypeDesc
	unitComponentType, ok = typelib.Types[Sum("UnitComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find UnitComponent hash in dl_library")
	}

	componentData := make([]byte, unitComponentType.Size)
	if _, err := r.Seek(int64(unitComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseUnitComponents() (map[stingray.Hash]UnitComponent, error) {
	unitHash := Sum("UnitComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var unitType DLTypeDesc
	var ok bool
	unitType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find UnitComponentData hash in dl_library")
	}

	if len(unitType.Members) != 2 {
		return nil, fmt.Errorf("UnitComponentData unexpected format (there should be 2 members but were actually %v)", len(unitType.Members))
	}

	if unitType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitComponentData unexpected format (data atom was not inline array)")
	}

	if unitType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("UnitComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("UnitComponentData unexpected format (data storage was not struct)")
	}

	if unitType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("UnitComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitType.Members[1].TypeID != Sum("UnitComponent") {
		return nil, fmt.Errorf("UnitComponentData unexpected format (data type was not UnitComponent)")
	}

	unitComponentData, err := getUnitComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get unit component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(unitComponentData)

	hashmap := make([]ComponentIndexData, unitType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]UnitComponent, unitType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]UnitComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
