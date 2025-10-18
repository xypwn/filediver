package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type UnitMaterialTuple struct {
	UnitPath                 stingray.Hash     // [unit]Any units with this path will use the appropriate mesh name
	BugScreenMaterial        stingray.Hash     // [material]The material that gets set if this is a bug faction
	BotScreenMaterial        stingray.Hash     // [material]The material that gets set if this is a bot faction
	IlluminateScreenMaterial stingray.Hash     // [material]The material that gets set if this is a illuminate faction
	SuperearthScreenMaterial stingray.Hash     // [material]The material that gets set if this is a super earth faction
	SlotName                 stingray.ThinHash // [string]The slot that gets the appropriate faction material
	_                        [4]uint8
}

func (w UnitMaterialTuple) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleUnitMaterialTuple {
	return SimpleUnitMaterialTuple{
		UnitPath:                 lookupHash(w.UnitPath),
		BugScreenMaterial:        lookupHash(w.BugScreenMaterial),
		BotScreenMaterial:        lookupHash(w.BotScreenMaterial),
		IlluminateScreenMaterial: lookupHash(w.IlluminateScreenMaterial),
		SuperearthScreenMaterial: lookupHash(w.SuperearthScreenMaterial),
		SlotName:                 lookupThinHash(w.SlotName),
	}
}

type TruthTransmitterFactionComponent struct {
	UnitMaterials [8]UnitMaterialTuple // The mapping between which unit's material name is set.
}

type SimpleUnitMaterialTuple struct {
	UnitPath                 string `json:"unit_path"`
	BugScreenMaterial        string `json:"bug_screen_material"`
	BotScreenMaterial        string `json:"bot_screen_material"`
	IlluminateScreenMaterial string `json:"illuminate_screen_material"`
	SuperearthScreenMaterial string `json:"superearth_screen_material"`
	SlotName                 string `json:"slot_name"`
}

type SimpleTruthTransmitterFactionComponent struct {
	UnitMaterials []SimpleUnitMaterialTuple `json:"unit_materials,omitempty"`
}

func (w TruthTransmitterFactionComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	unitMaterials := make([]SimpleUnitMaterialTuple, 0)
	for _, mat := range w.UnitMaterials {
		if mat.UnitPath.Value == 0 {
			break
		}
		unitMaterials = append(unitMaterials, mat.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}
	return SimpleTruthTransmitterFactionComponent{
		UnitMaterials: unitMaterials,
	}
}

func getTruthTransmitterFactionComponentData() ([]byte, error) {
	truthTransmitterFactionComponentHash := Sum("TruthTransmitterFactionComponentData")
	truthTransmitterFactionComponentHashData := make([]byte, 4)
	if _, err := binary.Encode(truthTransmitterFactionComponentHashData, binary.LittleEndian, truthTransmitterFactionComponentHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, truthTransmitterFactionComponentHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)
	return data, err
}

func getTruthTransmitterFactionComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	UnitCmpDataHash := Sum("TruthTransmitterFactionComponentData")
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
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCmpDataType.Members))
	}

	if unitCmpDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCmpDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (data atom was not inline array)")
	}

	if unitCmpDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCmpDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (data storage was not struct)")
	}

	if unitCmpDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCmpDataType.Members[1].TypeID != Sum("TruthTransmitterFactionComponent") {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (data type was not TruthTransmitterFactionComponent)")
	}

	truthTransmitterFactionComponentData, err := getTruthTransmitterFactionComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get truth transmitter faction component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(truthTransmitterFactionComponentData)

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
		return nil, fmt.Errorf("%v not found in truth transmitter faction component data", hash.String())
	}

	var truthTransmitterFactionComponentType DLTypeDesc
	truthTransmitterFactionComponentType, ok = typelib.Types[Sum("TruthTransmitterFactionComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find TruthTransmitterFactionComponent hash in dl_library")
	}

	componentData := make([]byte, truthTransmitterFactionComponentType.Size)
	if _, err := r.Seek(int64(truthTransmitterFactionComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func ParseTruthTransmitterFactionComponents() (map[stingray.Hash]TruthTransmitterFactionComponent, error) {
	unitHash := Sum("TruthTransmitterFactionComponentData")
	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var factionType DLTypeDesc
	var ok bool
	factionType, ok = typelib.Types[unitHash]
	if !ok {
		return nil, fmt.Errorf("could not find TruthTransmitterFactionComponentData hash in dl_library")
	}

	if len(factionType.Members) != 2 {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (there should be 2 members but were actually %v)", len(factionType.Members))
	}

	if factionType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (hashmap atom was not inline array)")
	}

	if factionType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (data atom was not inline array)")
	}

	if factionType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (hashmap storage was not struct)")
	}

	if factionType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (data storage was not struct)")
	}

	if factionType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if factionType.Members[1].TypeID != Sum("TruthTransmitterFactionComponent") {
		return nil, fmt.Errorf("TruthTransmitterFactionComponentData unexpected format (data type was not TruthTransmitterFactionComponent)")
	}

	truthTransmitterFactionComponentData, err := getTruthTransmitterFactionComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get truth transmitter faction component data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(truthTransmitterFactionComponentData)

	hashmap := make([]ComponentIndexData, factionType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	data := make([]TruthTransmitterFactionComponent, factionType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	result := make(map[stingray.Hash]TruthTransmitterFactionComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
