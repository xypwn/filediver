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

func (ucct UnitCustomizationCollectionType) Unit() (stingray.Hash, error) {
	switch ucct {
	case CollectionCombatWalker:
		return stingray.Sum("content/fac_helldivers/vehicles/combat_walker/combat_walker"), nil
	case CollectionCombatWalkerEmancipator:
		return stingray.Sum("content/fac_helldivers/vehicles/combat_walker_obsidian/combat_walker_obsidian"), nil
	case CollectionFRV:
		return stingray.Sum("content/fac_helldivers/vehicles/frv/frv"), nil
	case CollectionHellpod:
		return stingray.Sum("content/fac_helldivers/hellpod/hellpod/hellpod"), nil
	case CollectionHellpodRack:
		return stingray.Sum("content/fac_helldivers/hellpod/weapon_rack/weapon_rack"), nil
	case CollectionShuttle:
		return stingray.Sum("content/fac_helldivers/vehicles/shuttle_gunship/shuttle_gunship"), nil
	}
	return stingray.Hash{Value: 0}, fmt.Errorf("Unknown unit for UnitCustomizationCollectionType %v", ucct)
}

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

type UnitCustomizationMaterialOverrides struct {
	MaterialID        stingray.ThinHash `json:"material_id"`
	_                 [4]byte
	MaterialLut       stingray.Hash `json:"material_lut,omitzero"`
	DecalSheet        stingray.Hash `json:"decal_sheet,omitzero"`
	PatternLut        stingray.Hash `json:"pattern_lut,omitzero"`
	PatternMasksArray stingray.Hash `json:"pattern_masks_array,omitzero"`
}

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

// Simpler struct to just contain all the overrides for a skin
type UnitSkinOverride struct {
	Name      string
	ID        stingray.ThinHash
	Overrides map[stingray.ThinHash][]UnitCustomizationMaterialOverrides
}

type UnitSkinOverrideGroup struct {
	CollectionType UnitCustomizationCollectionType
	Skins          []UnitSkinOverride
}

func (u UnitSkinOverrideGroup) HasMaterial(matId stingray.ThinHash) bool {
	for i := range u.Skins {
		if _, ok := u.Skins[i].Overrides[matId]; ok {
			return true
		}
	}
	return false
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

func (u *UnitCustomizationSettings) GetSkinOverrideGroup() UnitSkinOverrideGroup {
	toReturn := UnitSkinOverrideGroup{
		Skins:          make([]UnitSkinOverride, 0),
		CollectionType: u.CollectionType,
	}
	for _, skin := range u.Skins {
		skinOverride := UnitSkinOverride{
			Name:      skin.Name,
			ID:        skin.ID,
			Overrides: make(map[stingray.ThinHash][]UnitCustomizationMaterialOverrides),
		}
		for _, override := range skin.Customization.MaterialsTexturesOverrides {
			if _, ok := skinOverride.Overrides[override.MaterialID]; !ok {
				skinOverride.Overrides[override.MaterialID] = make([]UnitCustomizationMaterialOverrides, 0)
			}
			skinOverride.Overrides[override.MaterialID] = append(skinOverride.Overrides[override.MaterialID], override)
		}
		for _, override := range skin.Customization.MountedWeaponTextureOverrides {
			if _, ok := skinOverride.Overrides[override.MaterialID]; !ok {
				skinOverride.Overrides[override.MaterialID] = make([]UnitCustomizationMaterialOverrides, 0)
			}
			skinOverride.Overrides[override.MaterialID] = append(skinOverride.Overrides[override.MaterialID], override)
		}
		toReturn.Skins = append(toReturn.Skins, skinOverride)
	}
	return toReturn
}

type rawUnitCustomizationSetting struct {
	DebugNameOffset      uint64
	ID                   stingray.ThinHash
	_                    [4]byte
	AddPath              stingray.Hash
	Name                 uint32
	_                    [4]byte
	Thumbnail            stingray.Hash
	UIWidgetColorsOffset uint64
	UIWidgetColorsCount  uint64
}

type UnitCustomizationSetting struct {
	Name           string                     `json:"name"`
	DebugName      string                     `json:"debug_name"`
	ID             stingray.ThinHash          `json:"id"`
	Archive        stingray.Hash              `json:"archive"`
	Customization  UnitCustomizationComponent `json:"customization"`
	Thumbnail      stingray.Hash              `json:"thumbnail"`
	UIWidgetColors []mgl32.Vec3               `json:"ui_widget_colors"`
}

type UnitCustomizationComponent struct {
	MaterialsTexturesOverrides    []UnitCustomizationMaterialOverrides `json:"materials_textures_overrides"`
	MountedWeaponTextureOverrides []UnitCustomizationMaterialOverrides `json:"mounted_weapon_texture_overrides"`
}

type SimpleUnitCustomizationMaterialOverrides struct {
	MaterialID        string `json:"material"`
	MaterialLut       string `json:"material_lut"`
	DecalSheet        string `json:"decal_sheet"`
	PatternLut        string `json:"pattern_lut"`
	PatternMasksArray string `json:"pattern_masks_array"`
}

type SimpleUnitCustomizationComponent struct {
	MaterialsTexturesOverrides    []SimpleUnitCustomizationMaterialOverrides `json:"materials_textures_overrides"`
	MountedWeaponTextureOverrides []SimpleUnitCustomizationMaterialOverrides `json:"mounted_weapon_texture_overrides"`
}

type SimpleUnitCustomizationSetting struct {
	Name           string                           `json:"name"`
	DebugName      string                           `json:"debug_name"`
	ID             string                           `json:"id"`
	Archive        string                           `json:"archive"`
	Customization  SimpleUnitCustomizationComponent `json:"customization"`
	Thumbnail      string                           `json:"thumbnail"`
	UIWidgetColors []mgl32.Vec3                     `json:"ui_widget_colors"`
}

type SimpleUnitCustomizationSettings struct {
	ParentCollectionType UnitCustomizationCollectionType         `json:"parent_collection_type"`
	CollectionType       UnitCustomizationCollectionType         `json:"collection_type"`
	ObjectName           string                                  `json:"object_name"`
	SkinName             string                                  `json:"skin_name"`
	CategoryType         UnitCustomizationCollectionCategoryType `json:"category_type"`
	Skins                []SimpleUnitCustomizationSetting        `json:"skins,omitempty"`
	ShowroomOffset       mgl32.Vec3                              `json:"showroom_offset"`
	ShowroomRotation     mgl32.Vec3                              `json:"showroom_rotation"`
}

func (customization UnitCustomizationSettings) ToSimple(lookupHash func(stingray.Hash) string, lookupThinHash func(stingray.ThinHash) string) any {
	simpleSettings := SimpleUnitCustomizationSettings{
		ParentCollectionType: customization.ParentCollectionType,
		CollectionType:       customization.CollectionType,
		ObjectName:           customization.ObjectName,
		SkinName:             customization.SkinName,
		CategoryType:         customization.CategoryType,
		ShowroomOffset:       customization.ShowroomOffset,
		ShowroomRotation:     customization.ShowroomRotation,
		Skins:                make([]SimpleUnitCustomizationSetting, 0),
	}
	for _, skin := range customization.Skins {
		simpleSetting := SimpleUnitCustomizationSetting{
			Name:           skin.Name,
			DebugName:      skin.DebugName,
			ID:             lookupThinHash(skin.ID),
			Archive:        lookupHash(skin.Archive),
			Thumbnail:      lookupHash(skin.Thumbnail),
			UIWidgetColors: skin.UIWidgetColors,
			Customization: SimpleUnitCustomizationComponent{
				MaterialsTexturesOverrides:    make([]SimpleUnitCustomizationMaterialOverrides, 0),
				MountedWeaponTextureOverrides: make([]SimpleUnitCustomizationMaterialOverrides, 0),
			},
		}
		for _, mto := range skin.Customization.MaterialsTexturesOverrides {
			simpleSetting.Customization.MaterialsTexturesOverrides = append(simpleSetting.Customization.MaterialsTexturesOverrides, SimpleUnitCustomizationMaterialOverrides{
				MaterialID:        lookupThinHash(mto.MaterialID),
				MaterialLut:       lookupHash(mto.MaterialLut),
				DecalSheet:        lookupHash(mto.DecalSheet),
				PatternLut:        lookupHash(mto.PatternLut),
				PatternMasksArray: lookupHash(mto.PatternMasksArray),
			})
		}
		for _, mwto := range skin.Customization.MountedWeaponTextureOverrides {
			simpleSetting.Customization.MountedWeaponTextureOverrides = append(simpleSetting.Customization.MountedWeaponTextureOverrides, SimpleUnitCustomizationMaterialOverrides{
				MaterialID:        lookupThinHash(mwto.MaterialID),
				MaterialLut:       lookupHash(mwto.MaterialLut),
				DecalSheet:        lookupHash(mwto.DecalSheet),
				PatternLut:        lookupHash(mwto.PatternLut),
				PatternMasksArray: lookupHash(mwto.PatternMasksArray),
			})
		}
		simpleSettings.Skins = append(simpleSettings.Skins, simpleSetting)
	}
	return simpleSettings
}

func (customization UnitCustomizationComponent) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) any {
	matOverrides := make([]SimpleUnitCustomizationMaterialOverrides, 0)
	for _, override := range customization.MaterialsTexturesOverrides {
		if override.MaterialID.Value == 0 {
			break
		}
		matOverrides = append(matOverrides, SimpleUnitCustomizationMaterialOverrides{
			MaterialID:        lookupThinHash(override.MaterialID),
			MaterialLut:       lookupHash(override.MaterialLut),
			DecalSheet:        lookupHash(override.DecalSheet),
			PatternLut:        lookupHash(override.PatternLut),
			PatternMasksArray: lookupHash(override.PatternMasksArray),
		})
	}

	mountOverrides := make([]SimpleUnitCustomizationMaterialOverrides, 0)
	for _, override := range customization.MountedWeaponTextureOverrides {
		if override.MaterialID.Value == 0 {
			break
		}
		mountOverrides = append(mountOverrides, SimpleUnitCustomizationMaterialOverrides{
			MaterialID:        lookupThinHash(override.MaterialID),
			MaterialLut:       lookupHash(override.MaterialLut),
			DecalSheet:        lookupHash(override.DecalSheet),
			PatternLut:        lookupHash(override.PatternLut),
			PatternMasksArray: lookupHash(override.PatternMasksArray),
		})
	}

	return SimpleUnitCustomizationComponent{
		MaterialsTexturesOverrides:    matOverrides,
		MountedWeaponTextureOverrides: mountOverrides,
	}
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
		return nil, fmt.Errorf("error getting hash lookup resource: %v", err)
	}

	if !ok {
		return nil, fmt.Errorf("could not find add path hash lookup main data")
	}

	addPathMap, err := parseHashLookup(bytes.NewReader(hashLookupData))
	if err != nil {
		return nil, fmt.Errorf("error parsing hash lookup: %v", err)
	}

	deltas, err := ParseEntityDeltas()
	if err != nil {
		return nil, fmt.Errorf("error parsing entity deltas: %v", err)
	}

	matTextOverridesLen, mountedWeaponOverridesLen, err := getOverrideArrayLengths(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting override array lengths: %v", err)
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
			return nil, fmt.Errorf("seek skins: %v", err)
		}

		if err := binary.Read(r, binary.LittleEndian, &rawSettingsSlice); err != nil {
			return nil, fmt.Errorf("read rawSettingsSlice: %v", err)
		}

		unitHash, err := rawSettings.CollectionType.Unit()
		if err != nil {
			return nil, fmt.Errorf("error getting unit for collection type: %v", err)
		}
		componentData, err := getUnitCustomizationComponentDataForHash(unitHash)
		if err != nil {
			return nil, fmt.Errorf("error getting unit customization component data for hash %v: %v", unitHash.String(), err)
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
			var overrideComponentData []byte
			if rawSkin.AddPath.Value == 0x0 {
				continue
			}

			skin.Archive, ok = addPathMap[rawSkin.AddPath.Value]
			if !ok {
				return nil, fmt.Errorf("could not find %x in hash lookup table", rawSkin.AddPath.Value)
			}

			delta, ok := deltas[rawSkin.AddPath]
			if !ok {
				return nil, fmt.Errorf("could not find %x in entity deltas", rawSkin.AddPath.Value)
			}

			overrideComponentData, err = PatchComponent(Sum("UnitCustomizationComponentData"), componentData, delta)
			if err != nil {
				return nil, fmt.Errorf("error patching component data: %v", err)
			}

			var matOverrides UnitCustomizationComponent
			materialsTexturesOverrides := make([]UnitCustomizationMaterialOverrides, matTextOverridesLen)
			mountedWeaponTextureOverrides := make([]UnitCustomizationMaterialOverrides, mountedWeaponOverridesLen)
			length, err := binary.Decode(overrideComponentData, binary.LittleEndian, &materialsTexturesOverrides)
			if err != nil {
				return nil, fmt.Errorf("error decoding material texture overrides: %v", err)
			}
			_, err = binary.Decode(overrideComponentData[length:], binary.LittleEndian, &mountedWeaponTextureOverrides)
			if err != nil {
				return nil, fmt.Errorf("error decoding mounted weapon texture overrides: %v", err)
			}
			matOverrides.MaterialsTexturesOverrides = make([]UnitCustomizationMaterialOverrides, 0)
			matOverrides.MountedWeaponTextureOverrides = make([]UnitCustomizationMaterialOverrides, 0)
			for _, override := range materialsTexturesOverrides {
				if override.MaterialID.Value == 0 {
					break
				}
				matOverrides.MaterialsTexturesOverrides = append(matOverrides.MaterialsTexturesOverrides, override)
			}
			for _, override := range mountedWeaponTextureOverrides {
				if override.MaterialID.Value == 0 {
					break
				}
				matOverrides.MountedWeaponTextureOverrides = append(matOverrides.MountedWeaponTextureOverrides, override)
			}
			skin.Customization = matOverrides

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

	componentMap := make(map[stingray.ThinHash]*UnitCustomizationSetting)
	for _, customization := range toReturn {
		if customization.CollectionType == CollectionHellpod {
			for i := range customization.Skins {
				if _, ok := componentMap[customization.Skins[i].ID]; ok {
					continue
				}
				componentMap[customization.Skins[i].ID] = &customization.Skins[i]
			}
		}
	}

	for i, customization := range toReturn {
		if customization.CollectionType != CollectionHellpodRack {
			continue
		}
		for j := range customization.Skins {
			component, ok := componentMap[customization.Skins[j].ID]
			if !ok {
				continue
			}
			toReturn[i].Skins[j].Name = component.Name
		}
	}

	return toReturn, nil
}

func getUnitCustomizationComponentData() ([]byte, error) {
	unitCustomizationComponentDataHash := Sum("UnitCustomizationComponentData")
	unitCustomizationComponentDataHashData := make([]byte, 4)
	if _, err := binary.Encode(unitCustomizationComponentDataHashData, binary.LittleEndian, unitCustomizationComponentDataHash); err != nil {
		return nil, err
	}
	hashIndex := bytes.Index(entities, unitCustomizationComponentDataHashData)
	r := bytes.NewReader(entities[hashIndex:])
	var header DLInstanceHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	data := make([]byte, header.Size)
	_, err := r.Read(data)

	return data, err
}

func getUnitCustomizationComponentDataForHash(hash stingray.Hash) ([]byte, error) {
	unitCustomizationComponentDataHash := Sum("UnitCustomizationComponentData")
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

	customizationData, err := getUnitCustomizationComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get unit customization data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(customizationData)

	hashmapEntries := make([]ComponentIndexData, unitCustomizationComponentDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmapEntries); err != nil {
		return nil, err
	}

	var index int32 = -1
	for _, entry := range hashmapEntries {
		if entry.Resource == hash {
			index = int32(entry.Index)
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("%v not found in unit customization component data", hash.String())
	}

	var unitCustomizationComponentType DLTypeDesc
	unitCustomizationComponentType, ok = typelib.Types[Sum("UnitCustomizationComponent")]
	if !ok {
		return nil, fmt.Errorf("could not find UnitCustomizationComponent hash in dl_library")
	}

	componentData := make([]byte, unitCustomizationComponentType.Size)
	if _, err := r.Seek(int64(unitCustomizationComponentType.Size*uint32(index)), io.SeekCurrent); err != nil {
		return nil, err
	}
	_, err = r.Read(componentData)
	return componentData, err
}

func getOverrideArrayLengths(typelib *DLTypeLib) (int, int, error) {
	if typelib == nil {
		var err error
		typelib, err = ParseTypeLib(nil)
		if err != nil {
			return -1, -1, err
		}
	}
	var unitCustomizationComponentType DLTypeDesc
	unitCustomizationComponentType, ok := typelib.Types[Sum("UnitCustomizationComponent")]
	if !ok {
		return -1, -1, fmt.Errorf("could not find UnitCustomizationComponent hash in dl_library")
	}

	if len(unitCustomizationComponentType.Members) != 2 {
		return -1, -1, fmt.Errorf("UnitCustomizationComponent unexpected format (there should be 2 members but were actually %v)", len(unitCustomizationComponentType.Members))
	}

	if unitCustomizationComponentType.Members[0].Type.Atom != INLINE_ARRAY {
		return -1, -1, fmt.Errorf("UnitCustomizationComponent unexpected format (materials_textures_overrides was not inline array)")
	}

	if unitCustomizationComponentType.Members[1].Type.Atom != INLINE_ARRAY {
		return -1, -1, fmt.Errorf("UnitCustomizationComponent unexpected format (mounted_weapon_texture_overrides was not inline array)")
	}

	if unitCustomizationComponentType.Members[0].TypeID != Sum("UnitCustomizationMaterialOverrides") {
		return -1, -1, fmt.Errorf("UnitCustomizationComponent unexpected format (materials_textures_overrides type was not UnitCustomizationMaterialOverrides)")
	}

	if unitCustomizationComponentType.Members[1].TypeID != Sum("UnitCustomizationMaterialOverrides") {
		return -1, -1, fmt.Errorf("UnitCustomizationComponent unexpected format (mounted_weapon_texture_overrides type was not UnitCustomizationMaterialOverrides)")
	}

	matTextOverridesLen := unitCustomizationComponentType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen()
	mountedWeaponOverridesLen := unitCustomizationComponentType.Members[1].Type.BitfieldInfoOrArrayLen.GetArrayLen()
	return int(matTextOverridesLen), int(mountedWeaponOverridesLen), nil
}

func ParseUnitCustomizationComponents() (map[stingray.Hash]UnitCustomizationComponent, error) {
	unitCustomizationComponentDataHash := Sum("UnitCustomizationComponentData")
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

	customizationData, err := getUnitCustomizationComponentData()
	if err != nil {
		return nil, fmt.Errorf("Could not get unit customization data from generated_entities.dl_bin: %v", err)
	}
	r := bytes.NewReader(customizationData)

	hashmap := make([]ComponentIndexData, unitCustomizationComponentDataType.Members[0].Type.BitfieldInfoOrArrayLen.GetArrayLen())
	if err := binary.Read(r, binary.LittleEndian, &hashmap); err != nil {
		return nil, err
	}

	matTextOverridesLen, mountedWeaponOverridesLen, err := getOverrideArrayLengths(typelib)
	if err != nil {
		return nil, err
	}

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
