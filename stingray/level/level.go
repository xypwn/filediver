package level

import (
	"encoding/binary"
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

type Material struct {
	Unk00 uint32 // Index into unit array?
	Count uint32 // maybe? Only seen values == 1 so far
	Slot  stingray.ThinHash
	Path  stingray.Hash
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

type rawLevel struct {
	Magic            uint32
	UnitCount        uint32
	DataCount        uint32
	UnkOffsets00     [1]uint32
	MetadataOffset   uint32
	UnkOffsets01     [13]uint32
	UnkCount00       uint32
	UnkOffsets02     [8]uint32
	PrefabCount      uint32
	PrefabOffset     uint32
	UnkOffsets03     [16]uint32
	MaterialOffset   uint32
	UnkOffsets04     [4]uint32
	UnkIntListOffset uint32
	UnkOffsets05     [9]uint32
	Name             stingray.Hash
	UnkCount01       uint64
}

type Level struct {
	Name      stingray.Hash
	Metadata  map[int][]MetadataEntry
	Prefabs   []Prefab
	Materials []Material
	Units     []Unit
}

func LoadLevel(r io.ReadSeeker) (*Level, error) {
	var raw rawLevel
	if err := binary.Read(r, binary.LittleEndian, &raw); err != nil {
		return nil, err
	}
	units := make([]Unit, raw.UnitCount)
	if err := binary.Read(r, binary.LittleEndian, units); err != nil {
		return nil, err
	}

	if _, err := r.Seek(int64(raw.MetadataOffset), io.SeekStart); err != nil {
		return nil, err
	}
	metadataOffsets := make([]uint32, raw.DataCount)
	if err := binary.Read(r, binary.LittleEndian, metadataOffsets); err != nil {
		return nil, err
	}
	metadata := make(map[int][]MetadataEntry)
	for idx, offset := range metadataOffsets {
		if offset == 0 {
			continue
		}
		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		metadata[idx] = make([]MetadataEntry, 0)
		for range count {
			if entry, err := parseMetadataEntry(r); err != nil {
				return nil, err
			} else {
				metadata[idx] = append(metadata[idx], *entry)
			}
		}
	}

	prefabs := make([]Prefab, raw.PrefabCount)
	if _, err := r.Seek(int64(raw.PrefabOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, prefabs); err != nil {
		return nil, err
	}

	if _, err := r.Seek(int64(raw.MaterialOffset), io.SeekStart); err != nil {
		return nil, err
	}
	var materialCount uint32
	if err := binary.Read(r, binary.LittleEndian, &materialCount); err != nil {
		return nil, err
	}
	materials := make([]Material, materialCount)
	if err := binary.Read(r, binary.LittleEndian, materials); err != nil {
		return nil, err
	}

	return &Level{
		Name:      raw.Name,
		Metadata:  metadata,
		Prefabs:   prefabs,
		Materials: materials,
		Units:     units,
	}, nil
}
