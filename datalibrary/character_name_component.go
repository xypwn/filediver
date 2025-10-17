package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type CharacterNameComponent struct {
	Race               enum.RaceType     // Must not be none if the name should be selected from random templates.
	RandomNameTemplate stingray.ThinHash // [string]If not empty, use a random name from the name list and insert it into the NAME localization variable of the template.
	CharacterName      uint32            // [string]The localization string identifier of the character name, as displayed in subtitles.
	UsePersistentSeed  uint8             // [bool]If set, this character uses a predictable seed based on peer id. This is most useful for ship characters.
	_                  [3]uint8
}

type SimpleCharacterNameComponent struct {
	Race               enum.RaceType `json:"race"`
	RandomNameTemplate string        `json:"random_name_template"`
	CharacterName      string        `json:"character_name"`
	UsePersistentSeed  bool          `json:"use_persistent_seed"`
}

func (w CharacterNameComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	return SimpleCharacterNameComponent{
		Race:               w.Race,
		RandomNameTemplate: lookupThinHash(w.RandomNameTemplate),
		CharacterName:      lookupStrings(w.CharacterName),
		UsePersistentSeed:  w.UsePersistentSeed != 0,
	}
}

func getCharacterNameComponentData() ([]byte, error) {
	characterNameComponentHash := Sum("CharacterNameComponentData")
	characterNameComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(characterNameComponentHashData, binary.LittleEndian, characterNameComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, characterNameComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getCharacterNameComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("CharacterNameComponentData")
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
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("CharacterNameComponent") {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (data type was not CharacterNameComponent)")
	}

	characterNameComponentData, err := getCharacterNameComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get character name component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(characterNameComponentData)

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
		return nil, fmt.Errorf("%v not found in character name component data", hash.String())
	}

	var characterNameComponentType DLTypeDesc
	characterNameComponentType, ok = typelib.Types[Sum("CharacterNameComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find CharacterNameComponent hash in dl_library")
	}

	componentData := make([]byte, characterNameComponentType.Size)
	if _, err := r.Seek(int64(characterNameComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseCharacterNameComponents() (map[stingray.Hash]CharacterNameComponent, error) {
	unitHash := Sum("CharacterNameComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var characterNameType DLTypeDesc
	var ok bool
	characterNameType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find CharacterNameComponentData hash in dl_library")
	}

	if len(characterNameType.Members) != 2 {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (there should be 2 members but were actually %v)", len(characterNameType.Members))
	}

	if characterNameType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (hashmap atom was not inline array)")
	}

	if characterNameType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (data atom was not inline array)")
	}

	if characterNameType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (hashmap storage was not struct)")
	}

	if characterNameType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (data storage was not struct)")
	}

	if characterNameType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if characterNameType.Members[1].TypeID != Sum("CharacterNameComponent") {
		return nil, fmt.Errorf("CharacterNameComponentData unexpected format (data type was not CharacterNameComponent)")
	}

	characterNameComponentData, err := getCharacterNameComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get character name component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(characterNameComponentData)

	hashmap := make([]ComponentIndexData, characterNameType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]CharacterNameComponent, characterNameType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]CharacterNameComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
