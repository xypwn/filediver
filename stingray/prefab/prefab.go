package prefab

import (
	"encoding/binary"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Unk00               uint64
	PrefabHash          stingray.Hash
	ObjectListOffset    uint32
	UnkListOffset0      uint32
	UnkListOffset1      uint32
	UnkListOffset2      uint32
	UnkListOffset3      uint32
	HierarchyListOffset uint32
	UnkListOffset4      uint32
}

type PrefabObject struct {
	Unk00     uint64
	UnitHash  stingray.Hash
	Unk01     uint64
	Unk02     uint64
	Position  [3]float32
	Rotation  [4]float32
	Scale     [3]float32
	UnkFloats [4]float32
	Index     uint32
	UnkData   [20]uint8
}

type Prefab struct {
	NameHash stingray.Hash
	Objects  []PrefabObject
}

func Load(r io.ReadSeeker) (*Prefab, error) {
	var hdr Header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if _, err := r.Seek(int64(hdr.ObjectListOffset), io.SeekStart); err != nil {
		return nil, err
	}
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	offsets := make([]uint32, length)
	if err := binary.Read(r, binary.LittleEndian, &offsets); err != nil {
		return nil, err
	}
	objects := make([]PrefabObject, 0, length)
	for _, offset := range offsets {
		if _, err := r.Seek(int64(hdr.ObjectListOffset+offset), io.SeekStart); err != nil {
			return nil, err
		}
		var object PrefabObject
		if err := binary.Read(r, binary.LittleEndian, &object); err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return &Prefab{
		NameHash: hdr.PrefabHash,
		Objects:  objects,
	}, nil
}
