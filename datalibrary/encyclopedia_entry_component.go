package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/xypwn/filediver/stingray"
)

type EncyclopediaEntryComponent struct {
	LocName               uint32 // [string]The localization string identifier of the name of this entity (type).
	LocNamePlural         uint32 // [string]Plural Variant of the name.
	LocNameUpper          uint32 // [string]Upper variant of the name.
	LocNameShort          uint32 // [string]Short variant of the name.
	LocNameShortUpper     uint32 // [string]Short and Upper variant of the name.
	Description           uint32 // [string]The localization string identifier the description of this entity.
	DescriptionUpper      uint32 // [string]Upper variant of the description
	DescriptionShort      uint32 // [string]Short variant of the description
	DescriptionShortUpper uint32 // [string]Short and Upper variant of the description
	Prefix                uint32 // [string]Prefix - used when doing fancy name displays.
	Fluff                 uint32 // [string]Similar to description, usually lore text.
	_                     [4]byte
	Icon                  stingray.Hash // [material]The icon to use for this entity (type).
}

type SimpleEncyclopediaEntryComponent struct {
	LocName               string `json:"loc_name,omitempty"`
	LocNamePlural         string `json:"loc_name_plural,omitempty"`
	LocNameUpper          string `json:"loc_name_upper,omitempty"`
	LocNameShort          string `json:"loc_name_short,omitempty"`
	LocNameShortUpper     string `json:"loc_name_short_upper,omitempty"`
	Description           string `json:"description,omitempty"`
	DescriptionUpper      string `json:"description_upper,omitempty"`
	DescriptionShort      string `json:"description_short,omitempty"`
	DescriptionShortUpper string `json:"description_short_upper,omitempty"`
	Prefix                string `json:"prefix,omitempty"`
	Fluff                 string `json:"fluff,omitempty"`
	Icon                  string `json:"icon"`
}

func (w EncyclopediaEntryComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	emptyIfNotFound := func(stringId uint32) string {
		toReturn := lookupStrings(stringId)
		if strings.Contains(toReturn, "String ID not found") {
			return ""
		}
		return toReturn
	}
	return SimpleEncyclopediaEntryComponent{
		LocName:               emptyIfNotFound(w.LocName),
		LocNamePlural:         emptyIfNotFound(w.LocNamePlural),
		LocNameUpper:          emptyIfNotFound(w.LocNameUpper),
		LocNameShort:          emptyIfNotFound(w.LocNameShort),
		LocNameShortUpper:     emptyIfNotFound(w.LocNameShortUpper),
		Description:           emptyIfNotFound(w.Description),
		DescriptionUpper:      emptyIfNotFound(w.DescriptionUpper),
		DescriptionShort:      emptyIfNotFound(w.DescriptionShort),
		DescriptionShortUpper: emptyIfNotFound(w.DescriptionShortUpper),
		Prefix:                emptyIfNotFound(w.Prefix),
		Fluff:                 emptyIfNotFound(w.Fluff),
		Icon:                  lookupHash(w.Icon),
	}
}

func getEncyclopediaEntryComponentData() ([]byte, error) {
	encyclopediaEntryComponentHash := Sum("EncyclopediaEntryComponentData")
	encyclopediaEntryComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(encyclopediaEntryComponentHashData, binary.LittleEndian, encyclopediaEntryComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, encyclopediaEntryComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getEncyclopediaEntryComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	encyclopediaEntryCmpDataHash := Sum("EncyclopediaEntryComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var encyclopediaEntryCmpDataType DLTypeDesc
	var ok bool
	encyclopediaEntryCmpDataType, ok = typelib.Types[encyclopediaEntryCmpDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find ProjectileWeaponComponentData hash in dl_library")
	}

	if len(encyclopediaEntryCmpDataType.Members) != 2 {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (there should be 2 members but were actually %v)", len(encyclopediaEntryCmpDataType.Members))
	}

	if encyclopediaEntryCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (hashmap atom was not inline array)")
	}

	if encyclopediaEntryCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (data atom was not inline array)")
	}

	if encyclopediaEntryCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (hashmap storage was not struct)")
	}

	if encyclopediaEntryCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (data storage was not struct)")
	}

	if encyclopediaEntryCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if encyclopediaEntryCmpDataType.Members[1].TypeID != Sum("EncyclopediaEntryComponent") {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (data type was not EncyclopediaEntryComponent)")
	}

	encyclopediaEntryComponentData, err := getEncyclopediaEntryComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get encyclopedia entry component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(encyclopediaEntryComponentData)

	hashmap := make([]ComponentIndexData, encyclopediaEntryCmpDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
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
		return nil, fmt.Errorf("%v not found in encyclopedia entry component data", hash.String())
	}

	var encyclopediaEntryComponentType DLTypeDesc
	encyclopediaEntryComponentType, ok = typelib.Types[Sum("EncyclopediaEntryComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find EncyclopediaEntryComponent hash in dl_library")
	}

	componentData := make([]byte, encyclopediaEntryComponentType.Size)
	if _, err := r.Seek(int64(encyclopediaEntryComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseEncyclopediaEntryComponents() (map[stingray.Hash]EncyclopediaEntryComponent, error) {
	unitHash := Sum("EncyclopediaEntryComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var encyclopediaEntryType DLTypeDesc
	var ok bool
	encyclopediaEntryType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find EncyclopediaEntryComponentData hash in dl_library")
	}

	if len(encyclopediaEntryType.Members) != 2 {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (there should be 2 members but were actually %v)", len(encyclopediaEntryType.Members))
	}

	if encyclopediaEntryType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (hashmap atom was not inline array)")
	}

	if encyclopediaEntryType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (data atom was not inline array)")
	}

	if encyclopediaEntryType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (hashmap storage was not struct)")
	}

	if encyclopediaEntryType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (data storage was not struct)")
	}

	if encyclopediaEntryType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if encyclopediaEntryType.Members[1].TypeID != Sum("EncyclopediaEntryComponent") {
		return nil, fmt.Errorf("EncyclopediaEntryComponentData unexpected format (data type was not EncyclopediaEntryComponent)")
	}

	encyclopediaEntryComponentData, err := getEncyclopediaEntryComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get encyclopedia entry component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(encyclopediaEntryComponentData)

	hashmap := make([]ComponentIndexData, encyclopediaEntryType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]EncyclopediaEntryComponent, encyclopediaEntryType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]EncyclopediaEntryComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
