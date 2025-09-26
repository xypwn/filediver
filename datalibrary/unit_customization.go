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

type hashLookup0x7056 struct {
	ParentCount            uint32
	Parents                []hashLookupParent
	HashCount1             uint32
	Hashes1                []stingray.Hash
	HashMap1EntryCount     uint32
	HashMap1               []hashLookupMapEntry
	HashCount2             uint32
	Hashes2                []stingray.Hash
	UnknownTypeIndicator   uint32
	Hashes2MappingCount    uint32
	Hashes2Mapping         []hashLookupHashMapping
	ThinHashMap1EntryCount uint32
	ThinHashMap1           []hashLookupThinMapEntry
	HashCount3             uint32
	Hashes3                []stingray.Hash
	HashMap2EntryCount     uint32
	HashMap2               []hashLookupMapEntry
	LookupTreeCount1       uint32
	LookupTrees1           []hashLookupTree
	HashMap3EntryCount     uint32
	HashMap3               []hashLookupMapEntry
	LookupTreeCount2       uint32
	LookupTrees2           []hashLookupTree
	HashMap4EntryCount     uint32
	HashMap4               []hashLookupMapEntry
	LookupTreeCount3       uint32
	LookupTrees3           []hashLookupTree
	HashMap5EntryCount     uint32
	HashMap5               []hashLookupMapEntry
	LookupTreeCount4       uint32
	LookupTrees4           []hashLookupTree
	DEADBEE7               uint32
}

type hashLookupParent struct {
	ItemCount uint32
	Items     []stingray.Hash
}

type hashLookupMapEntry struct {
	Key   uint64
	Value uint64
}

type hashLookupThinMapEntry struct {
	Hash  uint32
	Index uint32
}

type hashLookupHashMapping struct {
	Type  uint32
	Index uint32
	Count uint32
}

type hashLookupTree struct {
	Type       stingray.ThinHash
	UnkInt     uint32
	EntryCount uint32
	Entries    []hashLookupMapEntry
}

func parseHashLookup(r io.Reader) (map[uint64]stingray.Hash, error) {
	var hashLookup hashLookup0x7056
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.ParentCount); err != nil {
		return nil, err
	}

	hashLookup.Parents = make([]hashLookupParent, 0)
	for i := uint32(0); i < hashLookup.ParentCount; i++ {
		var count uint32 = 0
		for count == 0 {
			if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
				return nil, err
			}
		}
		items := make([]stingray.Hash, count)
		if err := binary.Read(r, binary.LittleEndian, &items); err != nil {
			return nil, err
		}
		hashLookup.Parents = append(hashLookup.Parents, hashLookupParent{
			ItemCount: count,
			Items:     items,
		})
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashCount1); err != nil {
		return nil, err
	}
	hashLookup.Hashes1 = make([]stingray.Hash, hashLookup.HashCount1)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes1); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap1EntryCount); err != nil {
		return nil, err
	}
	hashLookup.HashMap1 = make([]hashLookupMapEntry, hashLookup.HashMap1EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap1); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashCount2); err != nil {
		return nil, err
	}
	hashLookup.Hashes2 = make([]stingray.Hash, hashLookup.HashCount2)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes2); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.UnknownTypeIndicator); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes2MappingCount); err != nil {
		return nil, err
	}
	hashLookup.Hashes2Mapping = make([]hashLookupHashMapping, hashLookup.Hashes2MappingCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes2Mapping); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.ThinHashMap1EntryCount); err != nil {
		return nil, err
	}
	hashLookup.ThinHashMap1 = make([]hashLookupThinMapEntry, hashLookup.ThinHashMap1EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.ThinHashMap1); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashCount3); err != nil {
		return nil, err
	}
	hashLookup.Hashes3 = make([]stingray.Hash, hashLookup.HashCount3)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.Hashes3); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap2EntryCount); err != nil {
		return nil, err
	}
	hashLookup.HashMap2 = make([]hashLookupMapEntry, hashLookup.HashMap2EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap2); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount1); err != nil {
		return nil, err
	}
	hashLookup.LookupTrees1 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount1; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, err
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, err
		}
		hashLookup.LookupTrees1 = append(hashLookup.LookupTrees1, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap3EntryCount); err != nil {
		return nil, err
	}
	hashLookup.HashMap3 = make([]hashLookupMapEntry, hashLookup.HashMap3EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap3); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount2); err != nil {
		return nil, err
	}
	hashLookup.LookupTrees2 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount2; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, err
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, err
		}
		hashLookup.LookupTrees2 = append(hashLookup.LookupTrees2, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap4EntryCount); err != nil {
		return nil, err
	}
	hashLookup.HashMap4 = make([]hashLookupMapEntry, hashLookup.HashMap4EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap4); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount3); err != nil {
		return nil, err
	}
	hashLookup.LookupTrees3 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount3; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, err
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, err
		}
		hashLookup.LookupTrees3 = append(hashLookup.LookupTrees3, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap5EntryCount); err != nil {
		return nil, err
	}
	hashLookup.HashMap5 = make([]hashLookupMapEntry, hashLookup.HashMap5EntryCount)
	if err := binary.Read(r, binary.LittleEndian, &hashLookup.HashMap5); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.LookupTreeCount4); err != nil {
		return nil, err
	}
	hashLookup.LookupTrees4 = make([]hashLookupTree, 0)
	for i := uint32(0); i < hashLookup.LookupTreeCount3; i++ {
		var tree hashLookupTree
		if err := binary.Read(r, binary.LittleEndian, &tree.Type); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.UnkInt); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &tree.EntryCount); err != nil {
			return nil, err
		}
		tree.Entries = make([]hashLookupMapEntry, tree.EntryCount)
		if err := binary.Read(r, binary.LittleEndian, &tree.Entries); err != nil {
			return nil, err
		}
		hashLookup.LookupTrees4 = append(hashLookup.LookupTrees4, tree)
	}

	if err := binary.Read(r, binary.LittleEndian, &hashLookup.DEADBEE7); err != nil {
		return nil, err
	}

	if hashLookup.DEADBEE7 != 0xDEADBEE7 {
		return nil, fmt.Errorf("invalid format for 0x7056bc19c69f0f07.hash_lookup, expected final bytes read to be 0xDEADBEE7 but were %#08x", hashLookup.DEADBEE7)
	}

	toReturn := make(map[uint64]stingray.Hash)
	for _, entry := range hashLookup.HashMap5 {
		if entry.Key == 0x0 {
			continue
		}
		toReturn[entry.Key] = stingray.Hash{Value: entry.Value}
	}
	return toReturn, nil
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
