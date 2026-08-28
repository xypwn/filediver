package datalib

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
)

type SubRegionFlags struct {
	Bits uint8
}

type rawSubRegionSettings struct {
	Type                  enum.SubRegionType
	Zone                  enum.ZoneId
	UnkFloat1             float32
	Weight                float32
	HeightGroup           enum.SubRegionHeight
	CellMergeChance       float32
	BlockerCrossingChance float32
	UnkFloat2             float32
	AllowsLocation        enum.AllowLocationMode
	Flags                 SubRegionFlags
	_                     [2]uint8
	InvalidZoneProximity  [8]enum.ZoneId
}

type SubRegionSettings struct {
	Type                  enum.SubRegionType     `json:"type"`
	Zone                  enum.ZoneId            `json:"zone"`
	UnkFloat1             float32                `json:"unk_float1"`
	Weight                float32                `json:"weight"`
	HeightGroup           enum.SubRegionHeight   `json:"height_group"`
	CellMergeChance       float32                `json:"cell_merge_chance"`
	BlockerCrossingChance float32                `json:"blocker_crossing_chance"`
	UnkFloat2             float32                `json:"unk_float2"`
	AllowsLocation        enum.AllowLocationMode `json:"allows_location"`
	Flags                 SubRegionFlags         `json:"flags"`
	InvalidZoneProximity  []enum.ZoneId          `json:"invalid_zone_proximity,omitempty"`
}

func (s rawSubRegionSettings) Deserialize() SubRegionSettings {
	invalidZoneProximity := make([]enum.ZoneId, 0)
	for _, zone := range s.InvalidZoneProximity {
		if zone == enum.ZONE_ID_UNKNOWN {
			continue
		}
		invalidZoneProximity = append(invalidZoneProximity, zone)
	}

	return SubRegionSettings{
		Type:                  s.Type,
		Zone:                  s.Zone,
		UnkFloat1:             s.UnkFloat1,
		Weight:                s.Weight,
		HeightGroup:           s.HeightGroup,
		CellMergeChance:       s.CellMergeChance,
		BlockerCrossingChance: s.BlockerCrossingChance,
		UnkFloat2:             s.UnkFloat2,
		AllowsLocation:        s.AllowsLocation,
		Flags:                 s.Flags,
		InvalidZoneProximity:  invalidZoneProximity,
	}
}

type StampRaceAffiliation struct {
	Bits uint8
}

func (s StampRaceAffiliation) Bugs() bool {
	return (s.Bits & (1 << 0)) != 0
}
func (s StampRaceAffiliation) Cyborg() bool {
	return (s.Bits & (1 << 1)) != 0
}
func (s StampRaceAffiliation) Illuminate() bool {
	return (s.Bits & (1 << 2)) != 0
}
func (s StampRaceAffiliation) SuperEarth() bool {
	return (s.Bits & (1 << 3)) != 0
}

func (s StampRaceAffiliation) MarshalJSON() ([]byte, error) {
	result := make([]string, 0)
	if s.Bugs() {
		result = append(result, "bugs")
	}
	if s.Cyborg() {
		result = append(result, "cyborg")
	}
	if s.Illuminate() {
		result = append(result, "illuminate")
	}
	if s.SuperEarth() {
		result = append(result, "super_earth")
	}
	return json.Marshal(result)
}

type SubRegionSettingsOverride struct {
	Type            enum.SubRegionType   `json:"type"`
	Weight          float32              `json:"weight"`
	RegionType      enum.SubRegionType   `json:"region_type"`
	OverrideZoneId  enum.ZoneId          `json:"override_zone_id"`
	RaceAffiliation StampRaceAffiliation `json:"race_affiliation"`
	_               [3]uint8
	OperationType   enum.OperationType `json:"operation_type"`
	UnkInt          uint32             `json:"unk_int"`
}

type rawGenerationRegionSettings struct {
	// These are rather large and mostly pertain to level generation logic
	// currently we only care about these as a means to map planets -> zones
	UnkRegionHash              stingray.Hash
	_                          [560]uint8
	SubregionSettings          [16]rawSubRegionSettings
	_                          [1156]uint8
	SubregionSettingsOverrides [16]SubRegionSettingsOverride
	_                          [36]uint8
}

type GenerationRegionSettings struct {
	UnkRegionHash              stingray.Hash
	SubregionSettings          []SubRegionSettings
	SubregionSettingsOverrides []SubRegionSettingsOverride
}

type SimpleGenerationRegionSettings struct {
	UnkRegionHash              string                      `json:"unk_region_hash"`
	SubregionSettings          []SubRegionSettings         `json:"subregion_settings,omitempty"`
	SubregionSettingsOverrides []SubRegionSettingsOverride `json:"subregion_settings_overrides,omitempty"`
}

func (s GenerationRegionSettings) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleGenerationRegionSettings {
	subregionSettings := make([]SubRegionSettings, 0)
	for _, subregion := range s.SubregionSettings {
		if subregion.Weight == 0 {
			continue
		}
		subregionSettings = append(subregionSettings, subregion)
	}
	subregionSettingsOverrides := make([]SubRegionSettingsOverride, 0)
	for _, subregion := range s.SubregionSettingsOverrides {
		if subregion.Type == enum.SubRegionType_None || subregion.RegionType == enum.SubRegionType_Value_24_Len_20 {
			continue
		}
		subregionSettingsOverrides = append(subregionSettingsOverrides, subregion)
	}
	return SimpleGenerationRegionSettings{
		UnkRegionHash:              lookupHash(s.UnkRegionHash),
		SubregionSettings:          subregionSettings,
		SubregionSettingsOverrides: subregionSettingsOverrides,
	}
}

func (s rawGenerationRegionSettings) Deserialize(r io.ReadSeeker, base int64) (*GenerationRegionSettings, error) {
	subregionSettings := make([]SubRegionSettings, 0)
	for _, subregion := range s.SubregionSettings {
		if subregion.Type == enum.SubRegionType_None {
			continue
		}
		subregionSettings = append(subregionSettings, subregion.Deserialize())
	}
	subregionSettingsOverrides := make([]SubRegionSettingsOverride, 0)
	for _, subregion := range s.SubregionSettingsOverrides {
		if subregion.Type == enum.SubRegionType_None {
			continue
		}
		subregionSettingsOverrides = append(subregionSettingsOverrides, subregion)
	}

	return &GenerationRegionSettings{
		UnkRegionHash:              s.UnkRegionHash,
		SubregionSettings:          subregionSettings,
		SubregionSettingsOverrides: subregionSettingsOverrides,
	}, nil
}

type rawGenerationRegionVariantSettings struct {
	Name                                        DLString
	Id                                          stingray.ThinHash
	Type                                        enum.LevelGenerationRegionVariantType
	Weight                                      float32
	_                                           [4]uint8
	VistaPath                                   stingray.Hash
	ReplaceWaterWithMaterial                    uint8
	_                                           [3]uint8
	ReplaceWaterWithMaterialHeightOffset        float32
	ReplaceWaterWithMaterialWaterSubmergeOffset float32
	_                                           [4]uint8
	Settings                                    DLPtr
}

type GenerationRegionVariantSettings struct {
	Name                                        *string
	Id                                          stingray.ThinHash
	Type                                        enum.LevelGenerationRegionVariantType
	Weight                                      float32
	VistaPath                                   stingray.Hash
	ReplaceWaterWithMaterial                    bool
	ReplaceWaterWithMaterialHeightOffset        float32
	ReplaceWaterWithMaterialWaterSubmergeOffset float32
	Settings                                    *GenerationRegionSettings
}

type SimpleGenerationRegionVariantSettings struct {
	Name                                        *string                               `json:"name"`
	Id                                          string                                `json:"id"`
	Type                                        enum.LevelGenerationRegionVariantType `json:"type"`
	Weight                                      float32                               `json:"weight"`
	VistaPath                                   string                                `json:"vista_path"`
	ReplaceWaterWithMaterial                    bool                                  `json:"replace_water_with_material"`
	ReplaceWaterWithMaterialHeightOffset        float32                               `json:"replace_water_with_material_height_offset"`
	ReplaceWaterWithMaterialWaterSubmergeOffset float32                               `json:"replace_water_with_material_water_submerge_offset"`
	Settings                                    *SimpleGenerationRegionSettings       `json:"settings"`
}

func (s GenerationRegionVariantSettings) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleGenerationRegionVariantSettings {
	var settingsPtr *SimpleGenerationRegionSettings
	if s.Settings != nil {
		settings := s.Settings.ToSimple(lookupHash, lookupThinHash, lookupStrings)
		settingsPtr = &settings
	}

	return SimpleGenerationRegionVariantSettings{
		Name:                                 s.Name,
		Id:                                   lookupThinHash(s.Id),
		Type:                                 s.Type,
		Weight:                               s.Weight,
		VistaPath:                            lookupHash(s.VistaPath),
		ReplaceWaterWithMaterial:             s.ReplaceWaterWithMaterial,
		ReplaceWaterWithMaterialHeightOffset: s.ReplaceWaterWithMaterialHeightOffset,
		ReplaceWaterWithMaterialWaterSubmergeOffset: s.ReplaceWaterWithMaterialWaterSubmergeOffset,
		Settings: settingsPtr,
	}
}

func (s rawGenerationRegionVariantSettings) Deserialize(r io.ReadSeeker, base int64) (*GenerationRegionVariantSettings, error) {
	name, err := s.Name.Resolve(r, base)
	if err != nil {
		return nil, err
	}

	rawSettings, err := ResolveDLPtr[rawGenerationRegionSettings](s.Settings, r, base)
	if err != nil {
		return nil, err
	}

	var settings *GenerationRegionSettings
	if rawSettings != nil {
		settings, err = rawSettings.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
	}

	return &GenerationRegionVariantSettings{
		Name:                                 name,
		Id:                                   s.Id,
		Type:                                 s.Type,
		Weight:                               s.Weight,
		VistaPath:                            s.VistaPath,
		ReplaceWaterWithMaterial:             s.ReplaceWaterWithMaterial != 0,
		ReplaceWaterWithMaterialHeightOffset: s.ReplaceWaterWithMaterialHeightOffset,
		ReplaceWaterWithMaterialWaterSubmergeOffset: s.ReplaceWaterWithMaterialWaterSubmergeOffset,
		Settings: settings,
	}, nil
}

type rawGenerationRegionVariantList struct {
	Name            DLString
	Id              stingray.ThinHash
	_               [4]uint8
	ResourcePackage stingray.Hash
	Group           DLPtr
	Variants        DLArray
}

type GenerationRegionVariantList struct {
	Name            *string
	Id              stingray.ThinHash
	ResourcePackage stingray.Hash
	Group           *GenerationRegionGroup
	Variants        []GenerationRegionVariantSettings
}

type SimpleGenerationRegionVariantList struct {
	Name            *string                                 `json:"name"`
	Id              string                                  `json:"id"`
	ResourcePackage string                                  `json:"resource_package"`
	Group           *SimpleGenerationRegionGroup            `json:"group"`
	Variants        []SimpleGenerationRegionVariantSettings `json:"variants"`
}

func (s GenerationRegionVariantList) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleGenerationRegionVariantList {
	var groupPtr *SimpleGenerationRegionGroup
	if s.Group != nil {
		group := s.Group.ToSimple(lookupHash, lookupThinHash, lookupStrings)
		groupPtr = &group
	}

	variants := make([]SimpleGenerationRegionVariantSettings, 0)
	for _, variant := range s.Variants {
		if variant.Weight == 0 {
			continue
		}
		variants = append(variants, variant.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleGenerationRegionVariantList{
		Name:            s.Name,
		Id:              lookupThinHash(s.Id),
		ResourcePackage: lookupHash(s.ResourcePackage),
		Group:           groupPtr,
		Variants:        variants,
	}
}

func (s rawGenerationRegionVariantList) Deserialize(r io.ReadSeeker, base int64) (*GenerationRegionVariantList, error) {
	name, err := s.Name.Resolve(r, base)
	if err != nil {
		return nil, err
	}

	rawGroup, err := ResolveDLPtr[rawGenerationRegionGroup](s.Group, r, base)
	if err != nil {
		return nil, err
	}

	var group *GenerationRegionGroup
	if rawGroup != nil {
		group, err = rawGroup.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
	}

	rawVariants, err := ResolveDLArray[rawGenerationRegionVariantSettings](s.Variants, r, base)
	if err != nil {
		return nil, err
	}
	variants := make([]GenerationRegionVariantSettings, 0)
	for _, variant := range rawVariants {
		resolved, err := variant.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		if resolved == nil {
			continue
		}
		variants = append(variants, *resolved)
	}

	return &GenerationRegionVariantList{
		Name:            name,
		Id:              s.Id,
		ResourcePackage: s.ResourcePackage,
		Group:           group,
		Variants:        variants,
	}, nil
}

type rawGenerationRegionGroup struct {
	Name          DLString
	Id            stingray.ThinHash
	_             [4]uint8
	Path          stingray.Hash
	SharedPackage stingray.Hash
	LocName       uint32
	GwColor       mgl32.Vec3
	Regions       DLArray
}

type GenerationRegionGroup struct {
	Name          *string
	Id            stingray.ThinHash
	Path          stingray.Hash
	SharedPackage stingray.Hash
	LocName       uint32
	GwColor       mgl32.Vec3
	Regions       []GenerationRegionVariantList
}

type SimpleGenerationRegionGroup struct {
	Name          *string                             `json:"name"`
	Id            string                              `json:"id"`
	Path          string                              `json:"path"`
	SharedPackage string                              `json:"shared_package"`
	LocName       string                              `json:"loc_name"`
	GwColor       mgl32.Vec3                          `json:"gw_color"`
	Regions       []SimpleGenerationRegionVariantList `json:"regions"`
}

func (s GenerationRegionGroup) ToSimple(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) SimpleGenerationRegionGroup {
	regions := make([]SimpleGenerationRegionVariantList, 0)
	for _, region := range s.Regions {
		regions = append(regions, region.ToSimple(lookupHash, lookupThinHash, lookupStrings))
	}

	return SimpleGenerationRegionGroup{
		Name:          s.Name,
		Id:            lookupThinHash(s.Id),
		Path:          lookupHash(s.Path),
		SharedPackage: lookupHash(s.SharedPackage),
		LocName:       lookupStrings(s.LocName),
		GwColor:       s.GwColor,
		Regions:       regions,
	}
}

func (s GenerationRegionGroup) RegionsMap() map[stingray.ThinHash]GenerationRegionVariantList {
	result := make(map[stingray.ThinHash]GenerationRegionVariantList)
	for _, region := range s.Regions {
		result[region.Id] = region
	}
	return result
}

func (s rawGenerationRegionGroup) Deserialize(r io.ReadSeeker, base int64) (*GenerationRegionGroup, error) {
	name, err := s.Name.Resolve(r, base)
	if err != nil {
		return nil, err
	}

	rawRegions, err := ResolveDLArray[rawGenerationRegionVariantList](s.Regions, r, base)
	if err != nil {
		return nil, err
	}

	regions := make([]GenerationRegionVariantList, 0)
	for _, region := range rawRegions {
		resolved, err := region.Deserialize(r, base)
		if err != nil {
			return nil, err
		}
		if resolved == nil {
			continue
		}
		regions = append(regions, *resolved)
	}

	return &GenerationRegionGroup{
		Name:          name,
		Id:            s.Id,
		Path:          s.Path,
		SharedPackage: s.SharedPackage,
		LocName:       s.LocName,
		GwColor:       s.GwColor,
		Regions:       regions,
	}, nil
}

func LoadRegionSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]GenerationRegionSettings, error) {
	r := bytes.NewReader(regionSettings)

	infos := make([]GenerationRegionSettings, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding region settings base: %v", err)
		}

		if header.Type != Sum("GenerationRegionSettings") {
			return nil, fmt.Errorf("invalid region settings file: got %v", header.Type.String())
		}

		var rawSetting rawGenerationRegionSettings
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading region settings: %v", err)
		}

		setting, err := rawSetting.Deserialize(r, base)
		if err != nil {
			return nil, fmt.Errorf("dereferencing region setting arrays/ptrs: %v", err)
		}

		_, err = r.Seek(base+int64(header.Size), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking next region settings: %v", err)
		}

		infos = append(infos, *setting)
	}

	return infos, nil
}

func LoadRegionGroups(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) ([]GenerationRegionGroup, error) {
	r := bytes.NewReader(regionSettings)

	infos := make([]GenerationRegionGroup, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		// Skip region settings at beginning of file
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		_, err := r.Seek(int64(header.Size), io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("skipping region settings: %v", err)
		}
	}

	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading region group count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding region groups base: %v", err)
		}

		if header.Type != Sum("GenerationRegionGroup") {
			return nil, fmt.Errorf("invalid region settings file")
		}

		var rawGroup rawGenerationRegionGroup
		if err := binary.Read(r, binary.LittleEndian, &rawGroup); err != nil {
			return nil, fmt.Errorf("reading region groups: %v", err)
		}

		group, err := rawGroup.Deserialize(r, base)
		if err != nil {
			return nil, fmt.Errorf("dereferencing region groups arrays/ptrs: %v", err)
		}

		_, err = r.Seek(base+int64(header.Size), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking next region groups: %v", err)
		}

		infos = append(infos, *group)
	}

	return infos, nil
}
