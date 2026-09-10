package prefab

import (
	"encoding/binary"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Unk00               uint64
	PrefabHash          stingray.Hash
	UnitListOffset      uint32
	UnkListOffset0      uint32
	PrefabListOffset    uint32
	UnkListOffset2      uint32
	UnkListOffset3      uint32
	HierarchyListOffset uint32
	UnkListOffset4      uint32
}

type Unit struct {
	UUID stingray.Hash
	stingray.Hash
	Name  stingray.Hash
	Unk02 uint64
	stingray.Transform
	UnkVec  mgl32.Vec3
	UnkInt  uint32
	Index   uint32
	UnkData [20]uint8
}

func (u *Unit) Path() stingray.Hash {
	return u.Hash
}

type NestedPrefab struct {
	UnkInt uint32
	Name   stingray.Hash
	Path   stingray.Hash
	stingray.Transform
	UnkFloats mgl32.Vec3
}

type Prefab struct {
	NameHash      stingray.Hash
	Units         []Unit
	NestedPrefabs []NestedPrefab
}

func Load(r io.ReadSeeker) (*Prefab, error) {
	base, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	var hdr Header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	var length uint32
	if hdr.UnitListOffset != 0 {
		if _, err := r.Seek(base+int64(hdr.UnitListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return nil, err
		}
	}
	offsets := make([]uint32, length)
	if err := binary.Read(r, binary.LittleEndian, &offsets); err != nil {
		return nil, err
	}
	units := make([]Unit, 0, length)
	for _, offset := range offsets {
		if _, err := r.Seek(base+int64(hdr.UnitListOffset+offset), io.SeekStart); err != nil {
			return nil, err
		}
		var object Unit
		if err := binary.Read(r, binary.LittleEndian, &object); err != nil {
			return nil, err
		}
		units = append(units, object)
	}

	var prefabsLength uint32
	if hdr.PrefabListOffset != 0 {
		if _, err := r.Seek(base+int64(hdr.PrefabListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &prefabsLength); err != nil {
			return nil, err
		}
	}
	prefabs := make([]NestedPrefab, prefabsLength)
	if err := binary.Read(r, binary.LittleEndian, prefabs); err != nil {
		return nil, err
	}
	return &Prefab{
		NameHash:      hdr.PrefabHash,
		Units:         units,
		NestedPrefabs: prefabs,
	}, nil
}
