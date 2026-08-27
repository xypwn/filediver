package datalib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/datalibrary/enum"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

type rawPlanetRegionSettings struct {
	TemplateOffset           int64
	UnkThinHash0             stingray.ThinHash
	UnkThinHash1             stingray.ThinHash
	RegionPreviewTexturePath stingray.Hash
	RegionRaceType           enum.RaceType
	_                        [4]uint8
	OperationTagsOffset      int64
	OperationTagsCount       uint64
	UnkInt0                  uint32
	UnkInt1                  uint32
	RegionOperationType      enum.OperationType
	_                        [4]uint8
	UnkThinHashes0Offset     int64
	UnkThinHashes0Count      uint64
	UnkEnumType              uint32
	_                        [4]uint8
	UnkThinHashes1Offset     int64
	UnkThinHashes1Count      uint64
	UnkEnumItemsOffset       int64
	UnkEnumItemsCount        uint64
	UnkInt2                  uint32
}

type PlanetRegionSettings struct {
	Template                 string
	UnkThinHash0             stingray.ThinHash
	UnkThinHash1             stingray.ThinHash
	RegionPreviewTexturePath stingray.Hash
	RegionRaceType           enum.RaceType
	OperationTags            []enum.OperationTag
	UnkInt0                  uint32
	UnkInt1                  uint32
	RegionOperationType      enum.OperationType
	UnkThinHashes0           []stingray.ThinHash
	UnkEnumType              uint32
	UnkThinHashes1           []stingray.ThinHash
	UnkEnumItems             []uint32
	UnkInt2                  uint32
}

type PlanetRegionSettingsMap map[stingray.Hash]PlanetRegionSettings

func LoadPlanetRegionSettings(lookupHash HashLookup, lookupThinHash ThinHashLookup, lookupStrings StringsLookup) (PlanetRegionSettingsMap, error) {
	r := bytes.NewReader(planetRegionSettings)

	infos := make([]PlanetRegionSettings, 0)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading count: %v", err)
	}
	for i := uint32(0); i < count; i++ {
		var header DLSubdataHeader
		if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
			return nil, fmt.Errorf("reading item %v: %v", i, err)
		}

		if header.Type != Sum("PlanetRegionSettings") {
			return nil, fmt.Errorf("invalid planet region settings file: type is %v at index %v", header.Type.String(), i)
		}

		base, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("finding planet region settings base: %v", err)
		}

		var rawSetting rawPlanetRegionSettings
		if err := binary.Read(r, binary.LittleEndian, &rawSetting); err != nil {
			return nil, fmt.Errorf("reading planet region settings: %v", err)
		}

		_, err = r.Seek(base+int64(rawSetting.TemplateOffset), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking planet region template: %v", err)
		}

		template, err := util.ReadCString(r)
		if err != nil {
			return nil, fmt.Errorf("reading planet region template: %v", err)
		}

		operationTags := make([]enum.OperationTag, rawSetting.OperationTagsCount)
		if rawSetting.OperationTagsOffset > 0 {
			_, err = r.Seek(base+int64(rawSetting.OperationTagsOffset), io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("seeking planet region operation tags: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, operationTags); err != nil {
				return nil, fmt.Errorf("reading planet region operation tags: %v", err)
			}
		}

		unkThinHashes0 := make([]stingray.ThinHash, rawSetting.UnkThinHashes0Count)
		if rawSetting.UnkThinHashes0Offset > 0 {
			_, err = r.Seek(base+int64(rawSetting.UnkThinHashes0Offset), io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("seeking planet region unk thinhashes 0: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, unkThinHashes0); err != nil {
				return nil, fmt.Errorf("reading planet region unk thinhashes 0: %v", err)
			}
		}

		unkThinHashes1 := make([]stingray.ThinHash, rawSetting.UnkThinHashes1Count)
		if rawSetting.UnkThinHashes1Offset > 0 {
			_, err = r.Seek(base+int64(rawSetting.UnkThinHashes1Offset), io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("seeking planet region unk thinhashes 1: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, unkThinHashes1); err != nil {
				return nil, fmt.Errorf("reading planet region unk thinhashes 1: %v", err)
			}
		}

		unkEnumItems := make([]uint32, rawSetting.UnkEnumItemsCount)
		if rawSetting.UnkEnumItemsOffset > 0 {
			_, err = r.Seek(base+int64(rawSetting.UnkEnumItemsOffset), io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("seeking planet region unk enum items: %v", err)
			}
			if err := binary.Read(r, binary.LittleEndian, unkEnumItems); err != nil {
				return nil, fmt.Errorf("reading planet region unk enum items: %v", err)
			}
		}

		_, err = r.Seek(base+int64(header.Size), io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking next planet region settings: %v", err)
		}

		infos = append(infos, PlanetRegionSettings{
			Template:                 template,
			UnkThinHash0:             rawSetting.UnkThinHash0,
			UnkThinHash1:             rawSetting.UnkThinHash1,
			RegionPreviewTexturePath: rawSetting.RegionPreviewTexturePath,
			RegionRaceType:           rawSetting.RegionRaceType,
			OperationTags:            operationTags,
			UnkInt0:                  rawSetting.UnkInt0,
			UnkInt1:                  rawSetting.UnkInt1,
			RegionOperationType:      rawSetting.RegionOperationType,
			UnkThinHashes0:           unkThinHashes0,
			UnkEnumType:              rawSetting.UnkEnumType,
			UnkThinHashes1:           unkThinHashes1,
			UnkEnumItems:             unkEnumItems,
			UnkInt2:                  rawSetting.UnkInt2,
		})
	}
	hashes := make([]stingray.Hash, count)
	if err := binary.Read(r, binary.LittleEndian, hashes); err != nil {
		return nil, fmt.Errorf("reading hashes: %v", err)
	}

	result := make(PlanetRegionSettingsMap)
	for i := range count {
		result[hashes[i]] = infos[i]
	}
	return result, nil
}
