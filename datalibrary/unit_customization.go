package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type GetResourceFunc func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error)

type UnitCustomizationCollectionType uint32

const (
	CollectionShuttle UnitCustomizationCollectionType = iota
	CollectionHellpod
	CollectionHellpodRack
	CollectionCombatWalker
	CollectionCombatWalkerEmancipator
	CollectionFRV
	CollectionCount
)

func (ucct UnitCustomizationCollectionType) MarshalText() ([]byte, error) {
	return []byte(ucct.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=UnitCustomizationCollectionType

type UnitCustomizationCollectionCategoryType uint32

const (
	CategoryHangar UnitCustomizationCollectionCategoryType = iota
	CategoryDeliverySystem
	CategoryVehicleBay
	CategoryCount
)

func (uccct UnitCustomizationCollectionCategoryType) MarshalText() ([]byte, error) {
	return []byte(uccct.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=UnitCustomizationCollectionCategoryType

type rawUnitCustomizationSettings struct {
	ParentCollectionType UnitCustomizationCollectionType
	CollectionType       UnitCustomizationCollectionType
	ObjectName           uint32
	SkinName             uint32
	CategoryType         UnitCustomizationCollectionCategoryType
	_                    [4]uint8
	SkinsOffset          uint64
	SkinsCount           uint64
	ShowroomOffset       mgl32.Vec3
	ShowroomRotation     mgl32.Vec3
}

type UnitCustomizationSettings struct {
	ParentCollectionType UnitCustomizationCollectionType         `json:"parent_collection_type"`
	CollectionType       UnitCustomizationCollectionType         `json:"collection_type"`
	ObjectName           string                                  `json:"object_name"`
	SkinName             string                                  `json:"skin_name"`
	CategoryType         UnitCustomizationCollectionCategoryType `json:"category_type"`
	Skins                []UnitCustomizationSetting              `json:"skins,omitempty"`
	ShowroomOffset       mgl32.Vec3                              `json:"showroom_offset"`
	ShowroomRotation     mgl32.Vec3                              `json:"showroom_rotation"`
}

type rawUnitCustomizationSetting struct {
	DebugNameOffset      uint64
	ID                   stingray.ThinHash
	_                    [4]byte
	AddPath              uint64
	Name                 uint32
	_                    [4]byte
	Thumbnail            stingray.Hash
	UIWidgetColorsOffset uint64
	UIWidgetColorsCount  uint64
}

type UnitCustomizationSetting struct {
	DebugName      string            `json:"debug_name"`
	ID             stingray.ThinHash `json:"id"`
	Archive        stingray.Hash     `json:"archive"`
	Name           string            `json:"name"`
	Thumbnail      stingray.Hash     `json:"thumbnail"`
	UIWidgetColors []mgl32.Vec3      `json:"ui_widget_colors"`
}

type UnitCustomizationMaterialOverrides struct {
	MaterialID        stingray.ThinHash
	MaterialLut       stingray.Hash
	DecalSheet        stingray.Hash
	PatternLut        stingray.Hash
	PatternMasksArray stingray.Hash
}

type UnitCustomizationComponent struct {
	MaterialsTexturesOverrides    []UnitCustomizationMaterialOverrides
	MountedWeaponTextureOverrides []UnitCustomizationMaterialOverrides
}

func ParseUnitCustomizationSettings(getResource GetResourceFunc, stringmap map[uint32]string) ([]UnitCustomizationSettings, error) {
	// I guess this hash_lookup file must be the material add path lookup file or something
	// Each hash lookup seems to have a slightly different format, so I don't have a general parser
	// for them
	hashLookupData, ok, err := getResource(stingray.FileID{
		Name: stingray.Hash{Value: 0x7056bc19c69f0f07},
		Type: stingray.Sum("hash_lookup"),
	}, stingray.DataMain)

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("could not find add path hash lookup main data")
	}

	addPathMap, err := parseHashLookup(bytes.NewReader(hashLookupData))
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(unitCustomizationSettings)

	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("read count: %v", err)
	}

	toReturn := make([]UnitCustomizationSettings, 0)
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("read header: %v", err)
		}

		if header.Type != Sum("UnitCustomizationSettings") {
			return nil, fmt.Errorf("invalid unit customization settings!")
		}

		base, _ := r.Seek(0, io.SeekCurrent)
		var rawSettings rawUnitCustomizationSettings
		if err := binary.Read(r, binary.LittleEndian, &rawSettings); err != nil {
			return nil, fmt.Errorf("read rawSettings: %v", err)
		}

		rawSettingsSlice := make([]rawUnitCustomizationSetting, rawSettings.SkinsCount)
		_, err := r.Seek(base+int64(rawSettings.SkinsOffset), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("read seek skins: %v", err)
		}

		if err := binary.Read(r, binary.LittleEndian, &rawSettingsSlice); err != nil {
			return nil, fmt.Errorf("read rawSettingsSlice: %v", err)
		}

		skins := make([]UnitCustomizationSetting, 0)
		for _, rawSkin := range rawSettingsSlice {
			var skin UnitCustomizationSetting
			debugNameBytes := unitCustomizationSettings[base+int64(rawSkin.DebugNameOffset):]
			terminator := bytes.IndexByte(debugNameBytes, 0)
			if terminator != -1 {
				skin.DebugName = string(debugNameBytes[:terminator])
			}

			skin.UIWidgetColors = make([]mgl32.Vec3, rawSkin.UIWidgetColorsCount)
			_, err := r.Seek(base+int64(rawSkin.UIWidgetColorsOffset), io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("seek uiwidgetoffset: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, &skin.UIWidgetColors); err != nil {
				return nil, fmt.Errorf("read uiwidgetcolors: %v", err)
			}

			skin.ID = rawSkin.ID
			skin.Thumbnail = rawSkin.Thumbnail
			if rawSkin.AddPath != 0x0 {
				var ok bool
				skin.Archive, ok = addPathMap[rawSkin.AddPath]
				if !ok {
					return nil, fmt.Errorf("could not find %x in hash lookup table", rawSkin.AddPath)
				}
			}

			name, ok := stringmap[rawSkin.Name]
			if !ok {
				skin.Name = strconv.FormatUint(uint64(rawSkin.Name), 16)
			} else {
				skin.Name = name
			}

			skins = append(skins, skin)
		}

		objectName, ok := stringmap[rawSettings.ObjectName]
		if !ok {
			objectName = strconv.FormatUint(uint64(rawSettings.ObjectName), 16)
		}
		skinName, ok := stringmap[rawSettings.SkinName]
		if !ok {
			skinName = strconv.FormatUint(uint64(rawSettings.SkinName), 16)
		}
		toReturn = append(toReturn, UnitCustomizationSettings{
			ParentCollectionType: rawSettings.ParentCollectionType,
			CollectionType:       rawSettings.CollectionType,
			ObjectName:           objectName,
			SkinName:             skinName,
			CategoryType:         rawSettings.CategoryType,
			Skins:                skins,
			ShowroomOffset:       rawSettings.ShowroomOffset,
			ShowroomRotation:     rawSettings.ShowroomRotation,
		})

		r.Seek(base+int64(header.Size), io.SeekStart)
	}

	return toReturn, nil
}

func ParseUnitCustomizationComponents() (map[stingray.Hash]UnitCustomizationComponent, error) {
	unitCustomizationComponentDataHash := Sum("UnitCustomizationComponentData")
	unitCustomizationComponentDataHashData := make([]byte, 4)
	if _, err := binary.Encode(unitCustomizationComponentDataHashData, binary.LittleEndian, unitCustomizationComponentDataHash); err != nil {
		return nil, err
	}
	r := bytes.NewReader(entities[bytes.Index(entities, unitCustomizationComponentDataHashData):])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	typelib, err := ParseTypeLib(nil)
	if err != nil {
		return nil, err
	}

	var unitCustomizationComponentDataType DLTypeDesc
	var ok bool
	unitCustomizationComponentDataType, ok = typelib.Types[unitCustomizationComponentDataHash]
	if !ok {
		return nil, fmt.Errorf("could not find UnitCustomizationComponentData hash in dl_library")
	}

	if len(unitCustomizationComponentDataType.Members) != 2 {
		return nil, fmt.Errorf("UnitCustomizationComponentData unexpected format (there should be 2 members but were actually %v)", len(unitCustomizationComponentDataType.Members))
	}

	if unitCustomizationComponentDataType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitCustomizationComponentData unexpected format (hashmap atom was not inline array)")
	}

	if unitCustomizationComponentDataType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitCustomizationComponentData unexpected format (data atom was not inline array)")
	}

	if unitCustomizationComponentDataType.Members[0].Type.Storage != STRUCT {
		return nil, fmt.Errorf("UnitCustomizationComponentData unexpected format (hashmap storage was not struct)")
	}

	if unitCustomizationComponentDataType.Members[1].Type.Storage != STRUCT {
		return nil, fmt.Errorf("UnitCustomizationComponentData unexpected format (data storage was not struct)")
	}

	if unitCustomizationComponentDataType.Members[0].TypeID != Sum("ComponentIndexData") {
		return nil, fmt.Errorf("UnitCustomizationComponentData unexpected format (hashmap type was not ComponentIndexData)")
	}

	if unitCustomizationComponentDataType.Members[1].TypeID != Sum("UnitCustomizationComponent") {
		return nil, fmt.Errorf("UnitCustomizationComponentData unexpected format (data type was not UnitCustomizationComponent)")
	}

	hashmap := make([]ComponentIndexData, unitCustomizationComponentDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	var unitCustomizationComponentType DLTypeDesc
	unitCustomizationComponentType, ok = typelib.Types[Sum("UnitCustomizationComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find UnitCustomizationComponent hash in dl_library")
	}

	if len(unitCustomizationComponentType.Members) != 2 {
		return nil, fmt.Errorf("UnitCustomizationComponent unexpected format (there should be 2 members but were actually %v)", len(unitCustomizationComponentType.Members))
	}

	if unitCustomizationComponentType.Members[0].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitCustomizationComponent unexpected format (materials_textures_overrides was not inline array)")
	}

	if unitCustomizationComponentType.Members[1].Type.Atom != INLINE_ARRAY {
		return nil, fmt.Errorf("UnitCustomizationComponent unexpected format (mounted_weapon_texture_overrides was not inline array)")
	}

	if unitCustomizationComponentType.Members[0].TypeID != Sum("UnitCustomizationMaterialOverrides") {
		return nil, fmt.Errorf("UnitCustomizationComponent unexpected format (materials_textures_overrides type was not UnitCustomizationMaterialOverrides)")
	}

	if unitCustomizationComponentType.Members[1].TypeID != Sum("UnitCustomizationMaterialOverrides") {
		return nil, fmt.Errorf("UnitCustomizationComponent unexpected format (mounted_weapon_texture_overrides type was not UnitCustomizationMaterialOverrides)")
	}

	matTextOverridesLen := unitCustomizationComponentType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen()
	mountedWeaponOverridesLen := unitCustomizationComponentType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen()
	data := make([]UnitCustomizationComponent, 0)
	for i := uint16(0); i < unitCustomizationComponentDataType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen(); i++ {
		materialsTexturesOverrides := make([]UnitCustomizationMaterialOverrides, matTextOverridesLen)
		if err := binary.Read(r, binary.LittleEndian, &materialsTexturesOverrides); err != nil {
			return nil, err
		}
		mountedWeaponTextureOverrides := make([]UnitCustomizationMaterialOverrides, mountedWeaponOverridesLen)
		if err := binary.Read(r, binary.LittleEndian, &mountedWeaponTextureOverrides); err != nil {
			return nil, err
		}
		data = append(data, UnitCustomizationComponent{
			MaterialsTexturesOverrides:    materialsTexturesOverrides,
			MountedWeaponTextureOverrides: mountedWeaponTextureOverrides,
		})
	}

	result := make(map[stingray.Hash]UnitCustomizationComponent)
	for _, component := range hashmap {
		if component.Resource.Value == 0x0 {
			continue
		}
		result[component.Resource] = data[component.Index]
	}

	return result, nil
}
