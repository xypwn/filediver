package geometrygroup

import (
	"encoding/binary"
	"io"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
)

type MeshInfoItem struct {
	MeshLayoutIndex uint32
	Unk00           [20]byte
	HashCount       uint32
	HashOffset      uint32
	Unk01           [8]byte
	GroupCount      uint32
	GroupOffset     uint32
}

type MeshHeader struct {
	MeshLayoutIndex uint32
	Materials       []stingray.ThinHash
	Groups          []unit.MeshGroup
}

type MeshInfo struct {
	Bones       []stingray.ThinHash
	MeshHeaders []MeshHeader
}

type UnitEntry struct {
	Unk  stingray.Hash
	Unit stingray.Hash
}

type GeometryGroupHeader struct {
	UnkHash              stingray.Hash
	ModelCount           uint32
	MeshLayoutListOffset uint32
}

type GeometryGroup struct {
	UnkHash     stingray.Hash
	MeshLayouts []unit.MeshLayout
	// Unit hashes to GG MeshInfo
	MeshInfos map[stingray.Hash]MeshInfo
}

func LoadGeometryGroup(mainR io.ReadSeeker) (*GeometryGroup, error) {
	var hdr GeometryGroupHeader
	if err := binary.Read(mainR, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}

	unitEntries := make([]UnitEntry, hdr.ModelCount)
	if err := binary.Read(mainR, binary.LittleEndian, &unitEntries); err != nil {
		return nil, err
	}

	meshInfoOffsets := make([]uint32, hdr.ModelCount)
	if err := binary.Read(mainR, binary.LittleEndian, &meshInfoOffsets); err != nil {
		return nil, err
	}

	meshInfoMap := make(map[stingray.Hash]MeshInfo)
	for i, offset := range meshInfoOffsets {
		if _, err := mainR.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		boneHashes := make([]stingray.ThinHash, count)
		if err := binary.Read(mainR, binary.LittleEndian, &boneHashes); err != nil {
			return nil, err
		}
		meshInfoItemOffsets := make([]uint32, count)
		if err := binary.Read(mainR, binary.LittleEndian, &meshInfoItemOffsets); err != nil {
			return nil, err
		}
		meshHeaders := make([]MeshHeader, 0, count)
		for _, itemOffset := range meshInfoItemOffsets {
			if _, err := mainR.Seek(int64(itemOffset+offset), io.SeekStart); err != nil {
				return nil, err
			}
			var meshInfoItem MeshInfoItem
			if err := binary.Read(mainR, binary.LittleEndian, &meshInfoItem); err != nil {
				return nil, err
			}

			meshHeader := MeshHeader{
				MeshLayoutIndex: meshInfoItem.MeshLayoutIndex,
				Materials:       make([]stingray.ThinHash, meshInfoItem.HashCount),
				Groups:          make([]unit.MeshGroup, meshInfoItem.GroupCount),
			}

			if _, err := mainR.Seek(int64(meshInfoItem.HashOffset+itemOffset+offset), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &meshHeader.Materials); err != nil {
				return nil, err
			}

			if _, err := mainR.Seek(int64(meshInfoItem.GroupOffset+itemOffset+offset), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &meshHeader.Groups); err != nil {
				return nil, err
			}

			meshHeaders = append(meshHeaders, meshHeader)
		}
		meshInfoMap[unitEntries[i].Unit] = MeshInfo{
			Bones:       boneHashes,
			MeshHeaders: meshHeaders,
		}
	}

	if _, err := mainR.Seek(int64(hdr.MeshLayoutListOffset), io.SeekStart); err != nil {
		return nil, err
	}

	var layoutCount uint32
	if err := binary.Read(mainR, binary.LittleEndian, &layoutCount); err != nil {
		return nil, err
	}

	layoutOffsets := make([]uint32, layoutCount)
	if err := binary.Read(mainR, binary.LittleEndian, &layoutOffsets); err != nil {
		return nil, err
	}

	meshLayouts := make([]unit.MeshLayout, 0, layoutCount)
	for _, offset := range layoutOffsets {
		if _, err := mainR.Seek(int64(offset+hdr.MeshLayoutListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var layout unit.MeshLayout
		if err := binary.Read(mainR, binary.LittleEndian, &layout); err != nil {
			return nil, err
		}
		meshLayouts = append(meshLayouts, layout)
	}

	return &GeometryGroup{
		UnkHash:     hdr.UnkHash,
		MeshLayouts: meshLayouts,
		MeshInfos:   meshInfoMap,
	}, nil
}
