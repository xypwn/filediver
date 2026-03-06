package level

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type Unit struct {
	UnkHash00 stingray.Hash
	UnkHash01 stingray.Hash
	Path      stingray.Hash
	_         [8]uint8
	stingray.Transform
	UnkFloats [6]float32 // Maybe a bounding box?
}

func (p *Unit) Unit() stingray.Hash {
	return p.Path
}

type Prefab struct {
	UnkHash00 stingray.Hash
	Path      stingray.Hash
	stingray.Transform
	UnkExtraRotation mgl32.Vec4
}

type MaterialSlotOverrides struct {
	Index     uint32 // Index into unit array?
	Materials []Material
}

type Material struct {
	Slot stingray.ThinHash
	Path stingray.Hash
}

type LevelMetadataType uint32

const (
	LevelMetadata_NotSeen LevelMetadataType = iota
	LevelMetadata_uint32
	LevelMetadata_float32
	LevelMetadata_string
)

func (p LevelMetadataType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=LevelMetadataType

type MetadataEntry struct {
	VariableNames []stingray.ThinHash
	Type          LevelMetadataType
	ValueUint     uint32
	ValueFloat    float32
	ValueString   string
}

func parseMetadataEntry(r io.ReadSeeker) (*MetadataEntry, error) {
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if offset%4 != 0 {
		_, err = r.Seek(4-(offset%4), io.SeekCurrent)
	}
	if err != nil {
		return nil, err
	}

	var hashCount uint32
	if err = binary.Read(r, binary.LittleEndian, &hashCount); err != nil {
		return nil, err
	}

	variableNames := make([]stingray.ThinHash, hashCount)
	if err = binary.Read(r, binary.LittleEndian, variableNames); err != nil {
		return nil, err
	}

	var metadataType LevelMetadataType
	if err = binary.Read(r, binary.LittleEndian, &metadataType); err != nil {
		return nil, err
	}

	var uintVal uint32
	var floatVal float32
	var stringVal string
	switch metadataType {
	case LevelMetadata_uint32:
		if err = binary.Read(r, binary.LittleEndian, &uintVal); err != nil {
			return nil, err
		}
	case LevelMetadata_float32:
		if err = binary.Read(r, binary.LittleEndian, &floatVal); err != nil {
			return nil, err
		}
	case LevelMetadata_string:
		var length uint32
		if err = binary.Read(r, binary.LittleEndian, &length); err != nil {
			return nil, err
		}
		strData := make([]byte, length)
		if _, err = r.Read(strData); err != nil {
			return nil, err
		}
		// Strip null terminator
		stringVal = string(strData[:len(strData)-1])
	}

	return &MetadataEntry{
		VariableNames: variableNames,
		Type:          metadataType,
		ValueUint:     uintVal,
		ValueFloat:    floatVal,
		ValueString:   stringVal,
	}, nil
}

type HashIndexRange struct {
	Hash  stingray.ThinHash
	Start uint32 // Inclusive
	End   uint32 // Exclusive
}

type UnknownTransformedItem struct {
	Hash stingray.Hash
	stingray.Transform
	UnkFloats [6]float32
}

type ExtraUnit struct {
	UnkHash1 stingray.Hash
	Path     stingray.Hash
	UnkHash2 stingray.Hash
	_        [8]uint8
	stingray.Transform
	UnkFloats [3]float32
	UnkInt    uint32
	UnkInt2   uint32
}

type ExtraPrefab struct {
	UnkHash1 stingray.Hash
	Path     stingray.Hash
	stingray.Transform
	UnkFloats [3]float32
	UnkInt    uint32
}

type FloatTwoInts struct {
	UnkFloat float32 `json:"unk_float"`
	UnkInt1  uint32  `json:"unk_int_1"`
	UnkInt2  uint32  `json:"unk_int_2"`
}

type IntsAndFloat struct {
	UnkInts  [14]uint32 `json:"unk_ints"`
	UnkFloat float32    `json:"unk_float"`
}

type rawExtraUnitHeader struct {
	UnkInt                    uint32
	UnkInt2                   uint32
	LevelName                 stingray.Hash
	ExtraUnitsPtrListOffset   uint32 // Relative to this container
	UnkOffset1                uint32
	ExtraPrefabsPtrListOffset uint32
	UnkOffset3                uint32
	UnkIntListOffset          uint32
	UnkFloatTwoIntsListOffset uint32
	UnkIntsAndFloatListOffset uint32
	UnkOffset4                uint32
	UnkOffset5                uint32
	UnkOffset6                uint32
}

type ExtraUnitsContainer struct {
	UnkInt              uint32
	UnkInt2             uint32
	LevelName           stingray.Hash
	ExtraUnits          []ExtraUnit
	ExtraPrefabs        []ExtraPrefab
	UnkIntList          []uint32
	UnkFloatTwoIntsList []FloatTwoInts
	UnkIntsAndFloatList []IntsAndFloat
}

type rawLevel struct {
	Magic                          uint32
	UnitCount                      uint32
	DataCount                      uint32
	UnkOffsets00                   [1]uint32
	MetadataOffset                 uint32
	UnkOffsets01                   [13]uint32
	UnkCount00                     uint32
	UnkOffsets02                   [8]uint32
	PrefabCount                    uint32
	PrefabOffset                   uint32
	UnkHashCount                   uint32
	UnkHashesOffset                uint32
	UnkTransformedItemOffsetOffset uint32 // Double pointers for several items here
	ExtraUnitsInfoOffset           uint32
	UnitHashIndexRangeOffset       uint32
	UnkHashIndexRangeOffset0       uint32
	UnkHashIndexRangeOffset1       uint32
	UnkHashIndexRangeOffset2       uint32
	PrefabHashIndexRangeOffset     uint32
	UnkHashIndexRangeOffset3       uint32
	UnkHashIndexRangeOffset4       uint32
	UnkOffsets03                   [5]uint32
	MaterialOffset                 uint32
	UnkThinHashesOffset            uint32
	UnkOffsets04                   [3]uint32
	UnkIntListOffset               uint32
	UnkOffsets05                   [9]uint32
	Name                           stingray.Hash
	UnkCount01                     uint64
}

type Level struct {
	Name                   stingray.Hash
	Metadata               map[int][]MetadataEntry
	Prefabs                []Prefab
	MaterialOverrides      []MaterialSlotOverrides
	Units                  []Unit
	UnkTransformedItems    []UnknownTransformedItem
	UnkExtraUnitContainers []ExtraUnitsContainer
	UnitHashIndexRange     []HashIndexRange
	UnkHashIndexRange1     []HashIndexRange
	UnkHashIndexRange2     []HashIndexRange
	UnkHashIndexRange3     []HashIndexRange
	PrefabHashIndexRange   []HashIndexRange
	UnkHashIndexRange4     []HashIndexRange
	UnkHashIndexRange5     []HashIndexRange
}

func readExtraUnitsContainer(r io.ReadSeeker, extraUnitsHeaderOffset uint32) (*ExtraUnitsContainer, error) {
	extraUnitsContainer := &ExtraUnitsContainer{}
	var extraUnitsHeader rawExtraUnitHeader
	if err := binary.Read(r, binary.LittleEndian, &extraUnitsHeader); err != nil {
		return nil, err
	}
	extraUnitsContainer.UnkInt = extraUnitsHeader.UnkInt
	extraUnitsContainer.UnkInt2 = extraUnitsHeader.UnkInt2
	extraUnitsContainer.LevelName = extraUnitsHeader.LevelName
	if extraUnitsHeader.ExtraUnitsPtrListOffset != 0 {
		extraUnitsContainer.ExtraUnits = make([]ExtraUnit, 0)
		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.ExtraUnitsPtrListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var extraUnitsOffsetsCount uint32
		if err := binary.Read(r, binary.LittleEndian, &extraUnitsOffsetsCount); err != nil {
			return nil, err
		}
		if extraUnitsOffsetsCount > 512 {
			return nil, fmt.Errorf("extraUnitsOffsetsCount too big: %v", extraUnitsOffsetsCount)
		}
		extraUnitsOffsets := make([]uint32, extraUnitsOffsetsCount)
		if err := binary.Read(r, binary.LittleEndian, extraUnitsOffsets); err != nil {
			return nil, err
		}
		for _, offset := range extraUnitsOffsets {
			if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.ExtraUnitsPtrListOffset+offset), io.SeekStart); err != nil {
				return nil, err
			}
			var unit ExtraUnit
			if err := binary.Read(r, binary.LittleEndian, &unit); err != nil {
				return nil, err
			}
			extraUnitsContainer.ExtraUnits = append(extraUnitsContainer.ExtraUnits, unit)
		}
	}

	if extraUnitsHeader.ExtraPrefabsPtrListOffset != 0 {
		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.ExtraPrefabsPtrListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var extraPrefabsCount uint32
		if err := binary.Read(r, binary.LittleEndian, &extraPrefabsCount); err != nil {
			return nil, err
		}
		var extraPrefabsOffset uint32
		if err := binary.Read(r, binary.LittleEndian, &extraPrefabsOffset); err != nil {
			return nil, fmt.Errorf("reading extraPrefabsOffsets: %v", err)
		}
		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.ExtraPrefabsPtrListOffset+extraPrefabsOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking extra prefab: %v", err)
		}
		extraUnitsContainer.ExtraPrefabs = make([]ExtraPrefab, 0)
		if err := binary.Read(r, binary.LittleEndian, extraUnitsContainer.ExtraPrefabs); err != nil {
			return nil, fmt.Errorf("reading extra prefab: %v", err)
		}
	}

	if extraUnitsHeader.UnkIntListOffset != 0 {
		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.UnkIntListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var unkIntListCount uint32
		if err := binary.Read(r, binary.LittleEndian, &unkIntListCount); err != nil {
			return nil, err
		}
		extraUnitsContainer.UnkIntList = make([]uint32, unkIntListCount)
		if err := binary.Read(r, binary.LittleEndian, extraUnitsContainer.UnkIntList); err != nil {
			return nil, err
		}
	}

	if extraUnitsHeader.UnkFloatTwoIntsListOffset != 0 {
		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.UnkFloatTwoIntsListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var unkFloatTwoIntsListCount, unkFloatTwoIntsListOffset uint32
		if err := binary.Read(r, binary.LittleEndian, &unkFloatTwoIntsListCount); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &unkFloatTwoIntsListOffset); err != nil {
			return nil, err
		}

		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.UnkFloatTwoIntsListOffset+unkFloatTwoIntsListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		extraUnitsContainer.UnkFloatTwoIntsList = make([]FloatTwoInts, unkFloatTwoIntsListCount)
		if err := binary.Read(r, binary.LittleEndian, extraUnitsContainer.UnkFloatTwoIntsList); err != nil {
			return nil, err
		}
	}

	if extraUnitsHeader.UnkIntsAndFloatListOffset != 0 {
		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.UnkIntsAndFloatListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var unkIntsAndFloatListCount, unkIntsAndFloatListOffset uint32
		if err := binary.Read(r, binary.LittleEndian, &unkIntsAndFloatListCount); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &unkIntsAndFloatListOffset); err != nil {
			return nil, err
		}
		if _, err := r.Seek(int64(extraUnitsHeaderOffset+extraUnitsHeader.UnkIntsAndFloatListOffset+unkIntsAndFloatListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		if unkIntsAndFloatListCount > 512 {
			return nil, fmt.Errorf("unkIntsAndFloatListCount too big: %v", unkIntsAndFloatListCount)
		}
		extraUnitsContainer.UnkIntsAndFloatList = make([]IntsAndFloat, unkIntsAndFloatListCount)
		if err := binary.Read(r, binary.LittleEndian, extraUnitsContainer.UnkIntsAndFloatList); err != nil {
			return nil, err
		}
	}

	return extraUnitsContainer, nil
}

func LoadLevel(r io.ReadSeeker) (*Level, error) {
	var raw rawLevel
	if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
		return nil, fmt.Errorf("read raw level: %v", err)
	}
	units := make([]Unit, raw.UnitCount)
	if err := binary.Read(r, binary.LittleEndian, units); err != nil {
		return nil, fmt.Errorf("read units: %v", err)
	}

	if _, err := r.Seek(int64(raw.MetadataOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek metadata offset values: %v", err)
	}
	metadataOffsets := make([]uint32, raw.DataCount)
	if err := binary.Read(r, binary.LittleEndian, metadataOffsets); err != nil {
		return nil, fmt.Errorf("read metadata offsets: %v", err)
	}
	metadata := make(map[int][]MetadataEntry)
	for idx, offset := range metadataOffsets {
		if offset == 0 {
			continue
		}
		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek metadata offset: %v", err)
		}
		var count uint32
		if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
			return nil, fmt.Errorf("read metadata entry count: %v", err)
		}
		metadata[idx] = make([]MetadataEntry, 0)
		for range count {
			if entry, err := parseMetadataEntry(r); err != nil {
				return nil, fmt.Errorf("parse metadata entry: %v", err)
			} else {
				metadata[idx] = append(metadata[idx], *entry)
			}
		}
	}

	prefabs := make([]Prefab, raw.PrefabCount)
	if _, err := r.Seek(int64(raw.PrefabOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek prefab offset: %v", err)
	}
	if err := binary.Read(r, binary.LittleEndian, prefabs); err != nil {
		return nil, fmt.Errorf("read prefabs: %v", err)
	}

	materialOverrides := make([]MaterialSlotOverrides, 0)
	if raw.MaterialOffset != 0 {
		if _, err := r.Seek(int64(raw.MaterialOffset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek material offset: %v", err)
		}
		var materialOverrideCount uint32
		if err := binary.Read(r, binary.LittleEndian, &materialOverrideCount); err != nil {
			return nil, fmt.Errorf("read material override count: %v", err)
		}
		for range materialOverrideCount {
			var index, count uint32
			if err := binary.Read(r, binary.LittleEndian, &index); err != nil {
				return nil, fmt.Errorf("read material override index: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
				return nil, fmt.Errorf("read material override count: %v", err)
			}
			materials := make([]Material, count)
			if err := binary.Read(r, binary.LittleEndian, materials); err != nil {
				return nil, fmt.Errorf("read material overrides: %v", err)
			}
			materialOverrides = append(materialOverrides, MaterialSlotOverrides{
				Index:     index,
				Materials: materials,
			})
		}
	}

	unkTransformedItems := make([]UnknownTransformedItem, 0)
	if raw.UnkTransformedItemOffsetOffset != 0 {
		if _, err := r.Seek(int64(raw.UnkTransformedItemOffsetOffset), io.SeekStart); err != nil {
			return nil, err
		}
		unkTransformedItemOffsets := make([]uint32, raw.UnkHashCount)
		if err := binary.Read(r, binary.LittleEndian, unkTransformedItemOffsets); err != nil {
			return nil, err
		}
		for _, offset := range unkTransformedItemOffsets {
			if offset == 0 {
				continue
			}
			if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
				return nil, err
			}
			unkTransformedItem := UnknownTransformedItem{}
			if err := binary.Read(r, binary.LittleEndian, &unkTransformedItem); err != nil {
				return nil, err
			}
			unkTransformedItems = append(unkTransformedItems, unkTransformedItem)
		}
	}

	extraUnitsContainers := make([]ExtraUnitsContainer, 0)
	if raw.ExtraUnitsInfoOffset != 0 {
		if _, err := r.Seek(int64(raw.ExtraUnitsInfoOffset), io.SeekStart); err != nil {
			return nil, err
		}
		containerInfos := make([]struct {
			Offset uint32
			Size   uint32
		}, raw.UnkHashCount)

		if err := binary.Read(r, binary.LittleEndian, containerInfos); err != nil {
			return nil, err
		}

		for _, containerInfo := range containerInfos {
			if containerInfo.Offset == 0 || containerInfo.Size == 0 {
				continue
			}
			if _, err := r.Seek(int64(containerInfo.Offset), io.SeekStart); err != nil {
				return nil, err
			}
			extraUnitContainer, err := readExtraUnitsContainer(r, containerInfo.Offset)
			if err != nil {
				return nil, err
			}
			if extraUnitContainer != nil {
				extraUnitsContainers = append(extraUnitsContainers, *extraUnitContainer)
			}
		}
	}

	readHashIndexRangeList := func(r io.ReadSeeker, offset uint32) ([]HashIndexRange, error) {
		if offset == 0 {
			return nil, nil
		}
		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, err
		}
		var hashIndexRangeListCount uint32
		if err := binary.Read(r, binary.LittleEndian, &hashIndexRangeListCount); err != nil {
			return nil, err
		}
		if hashIndexRangeListCount > 512 {
			return nil, fmt.Errorf("hashIndexRangeListCount too big: %v", hashIndexRangeListCount)
		}
		hashIndexRangeList := make([]HashIndexRange, hashIndexRangeListCount)
		if err := binary.Read(r, binary.LittleEndian, hashIndexRangeList); err != nil {
			return nil, err
		}
		return hashIndexRangeList, nil
	}

	unitHashIndexRangeList, err := readHashIndexRangeList(r, raw.UnitHashIndexRangeOffset)
	if err != nil {
		return nil, err
	}

	unkHashIndexRangeList0, err := readHashIndexRangeList(r, raw.UnkHashIndexRangeOffset0)
	if err != nil {
		return nil, err
	}

	unkHashIndexRangeList1, err := readHashIndexRangeList(r, raw.UnkHashIndexRangeOffset1)
	if err != nil {
		return nil, err
	}

	unkHashIndexRangeList2, err := readHashIndexRangeList(r, raw.UnkHashIndexRangeOffset2)
	if err != nil {
		return nil, err
	}

	prefabHashIndexRangeList, err := readHashIndexRangeList(r, raw.PrefabHashIndexRangeOffset)
	if err != nil {
		return nil, err
	}

	unkHashIndexRangeList3, err := readHashIndexRangeList(r, raw.UnkHashIndexRangeOffset3)
	if err != nil {
		return nil, err
	}

	unkHashIndexRangeList4, err := readHashIndexRangeList(r, raw.UnkHashIndexRangeOffset4)
	if err != nil {
		return nil, err
	}

	return &Level{
		Name:                   raw.Name,
		Metadata:               metadata,
		Prefabs:                prefabs,
		MaterialOverrides:      materialOverrides,
		Units:                  units,
		UnkTransformedItems:    unkTransformedItems,
		UnkExtraUnitContainers: extraUnitsContainers,
		UnitHashIndexRange:     unitHashIndexRangeList,
		UnkHashIndexRange1:     unkHashIndexRangeList0,
		UnkHashIndexRange2:     unkHashIndexRangeList1,
		UnkHashIndexRange3:     unkHashIndexRangeList2,
		PrefabHashIndexRange:   prefabHashIndexRangeList,
		UnkHashIndexRange4:     unkHashIndexRangeList3,
		UnkHashIndexRange5:     unkHashIndexRangeList4,
	}, nil
}
