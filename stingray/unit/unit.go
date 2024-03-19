package unit

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type JointTransform struct {
	Rotation    [3][3]float32
	Translation [3]float32
	Scale       [3]float32
	Skew        float32
}

type JointListHeader struct {
	NumJoints uint32
	Unk00     [12]byte
}

type MeshLayoutItemType uint32

const (
	ItemPosition   MeshLayoutItemType = 0
	ItemColor      MeshLayoutItemType = 1
	ItemUVCoords   MeshLayoutItemType = 4
	ItemBoneWeight MeshLayoutItemType = 6
	ItemBoneIdx    MeshLayoutItemType = 7
)

func (v MeshLayoutItemType) String() string {
	switch v {
	case ItemPosition:
		return "position"
	case ItemColor:
		return "color"
	case ItemUVCoords:
		return "UV coords"
	case ItemBoneWeight:
		return "bone weight"
	case ItemBoneIdx:
		return "bone index"
	default:
		return fmt.Sprint(uint32(v))
	}
}

type MeshLayoutItemFormat uint32

const (
	FormatF32     MeshLayoutItemFormat = 0
	FormatVec2F   MeshLayoutItemFormat = 1
	FormatVec3F   MeshLayoutItemFormat = 2
	FormatVec4F   MeshLayoutItemFormat = 3
	FormatVec4U8  MeshLayoutItemFormat = 26
	FormatVec2U16 MeshLayoutItemFormat = 29
	FormatVec4U16 MeshLayoutItemFormat = 31
)

func (v MeshLayoutItemFormat) String() string {
	switch v {
	case FormatF32:
		return "float32"
	case FormatVec2F:
		return "[2]float32"
	case FormatVec3F:
		return "[3]float32"
	case FormatVec4F:
		return "[4]float32"
	case FormatVec4U8:
		return "[4]uint8"
	case FormatVec2U16:
		return "[2]uint16"
	case FormatVec4U16:
		return "[4]uint16"
	default:
		return fmt.Sprint(uint32(v))
	}
}

type MeshLayout struct {
	MagicNum0 [4]byte
	Unk00     [4]byte
	Items     [16]struct {
		Type   MeshLayoutItemType
		Format MeshLayoutItemFormat
		Layer  uint32
		Unk00  [8]byte
	}
	NumItems       uint32
	Unk01          [4]byte
	MagicNum1      [4]byte
	Unk02          [12]byte
	NumPositions   uint32
	PositionStride uint32
	Unk03          [16]byte
	MagicNum2      [4]byte
	Unk04          [12]byte
	NumIndices     uint32
	Unk05          [20]byte
	PositionOffset uint32
	PositionsSize  uint32
	IndexOffset    uint32
	IndicesSize    uint32
	Unk06          [16]byte
}

type MeshHeader struct {
	Unk00          [60]byte
	LayoutIdx      int32
	Unk01          [40]byte
	NumMaterials   uint32
	MaterialOffset uint32
	Unk02          [8]byte
	NumGroups      uint32
	GroupOffset    uint32
}

type MeshGroup struct {
	Unk00          [4]byte
	PositionOffset uint32
	NumPositions   uint32
	IndexOffset    uint32
	NumIndices     uint32
	Unk01          [4]byte
}

type MeshInfo struct {
	Header       MeshHeader
	MaterialIdxs []uint32
	Groups       []MeshGroup
}

type Header struct {
	Unk00                [8]byte
	Bones                stingray.Hash
	Unk01                [8]byte
	UnkHash00            stingray.Hash
	StateMachine         stingray.Hash
	Unk02                [8]byte
	UnkOffset00          uint32
	JointListOffset      uint32
	UnkOffset01          uint32
	UnkOffset02          uint32
	Unk03                [12]byte
	UnkOffset03          uint32
	UnkOffset04          uint32
	Unk04                [4]byte
	UnkOffset05          uint32
	MeshLayoutListOffset uint32
	MeshDataOffset       uint32
	MeshInfoListOffset   uint32
	Unk05                [8]byte
	MaterialListOffset   uint32
}

type Mesh struct {
	Info      MeshInfo
	Positions [][3]float32
	UVCoords  [][2]uint16
	Colors    [][4]uint8
	Indices   []uint32 // triangles
}

type Unit struct {
	JointTransforms        []JointTransform
	JointTransformMatrices [][4][4]float32
	Meshes                 []Mesh
}

func Load(mainR io.ReadSeeker, gpuR io.ReadSeeker) (*Unit, error) {
	var hdr Header
	if err := binary.Read(mainR, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}

	var jointListHdr JointListHeader
	var jointTransforms []JointTransform
	var jointTransformMatrices [][4][4]float32
	if hdr.JointListOffset != 0 {
		if _, err := mainR.Seek(int64(hdr.JointListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		if err := binary.Read(mainR, binary.LittleEndian, &jointListHdr); err != nil {
			return nil, err
		}
		jointTransforms = make([]JointTransform, jointListHdr.NumJoints)
		for i := range jointTransforms {
			if err := binary.Read(mainR, binary.LittleEndian, &jointTransforms[i]); err != nil {
				return nil, err
			}
		}
		jointTransformMatrices = make([][4][4]float32, jointListHdr.NumJoints)
		for i := range jointTransformMatrices {
			if err := binary.Read(mainR, binary.LittleEndian, &jointTransformMatrices[i]); err != nil {
				return nil, err
			}
		}
	}

	var meshLayouts []MeshLayout
	if hdr.MeshLayoutListOffset != 0 {
		if _, err := mainR.Seek(int64(hdr.MeshLayoutListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		meshLayoutOffsets := make([]uint32, count)
		for i := range meshLayoutOffsets {
			if err := binary.Read(mainR, binary.LittleEndian, &meshLayoutOffsets[i]); err != nil {
				return nil, err
			}
		}
		meshLayouts = make([]MeshLayout, count)
		for i := range meshLayouts {
			offset := hdr.MeshLayoutListOffset + meshLayoutOffsets[i]
			if _, err := mainR.Seek(int64(offset), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &meshLayouts[i]); err != nil {
				return nil, err
			}
		}
	}

	var meshInfos []MeshInfo
	if hdr.MeshInfoListOffset != 0 {
		if _, err := mainR.Seek(int64(hdr.MeshInfoListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		offsets := make([]uint32, count)
		for i := range offsets {
			if err := binary.Read(mainR, binary.LittleEndian, &offsets[i]); err != nil {
				return nil, err
			}
		}
		meshInfos = make([]MeshInfo, 0, count)
		for i := uint32(0); i < count; i++ {
			offset := hdr.MeshInfoListOffset + offsets[i]
			if _, err := mainR.Seek(int64(offset), io.SeekStart); err != nil {
				return nil, err
			}
			var hdr MeshHeader
			if err := binary.Read(mainR, binary.LittleEndian, &hdr); err != nil {
				return nil, err
			}
			if _, err := mainR.Seek(int64(offset+hdr.MaterialOffset), io.SeekStart); err != nil {
				return nil, err
			}
			materialIdxs := make([]uint32, hdr.NumMaterials)
			for j := range materialIdxs {
				if err := binary.Read(mainR, binary.LittleEndian, &materialIdxs[j]); err != nil {
					return nil, err
				}
			}
			if _, err := mainR.Seek(int64(offset+hdr.GroupOffset), io.SeekStart); err != nil {
				return nil, err
			}
			groups := make([]MeshGroup, hdr.NumGroups)
			for j := range groups {
				if err := binary.Read(mainR, binary.LittleEndian, &groups[j]); err != nil {
					return nil, err
				}
			}
			meshInfos = append(meshInfos, MeshInfo{
				Header:       hdr,
				MaterialIdxs: materialIdxs,
				Groups:       groups,
			})
		}
	}

	var meshes []Mesh
	if hdr.MeshDataOffset != 0 {
		if _, err := mainR.Seek(int64(hdr.MeshDataOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		if int(count) != len(meshInfos) {
			return nil, fmt.Errorf("expected mesh data count (%v) to equal mesh info count (%v)", count, len(meshInfos))
		}
		// The following is all in gpu_resources
		meshes = make([]Mesh, 0, count)
		for _, info := range meshInfos {
			if info.Header.LayoutIdx < 0 {
				continue
			}
			var mesh Mesh
			if int(info.Header.LayoutIdx) >= len(meshLayouts) {
				return nil, fmt.Errorf("mesh layout index (%v) is out of bounds of mesh layouts (len=%v)", info.Header.LayoutIdx, len(meshLayouts))
			}
			layout := meshLayouts[info.Header.LayoutIdx]
			for _, group := range info.Groups {
				for i := uint32(0); i < group.NumPositions; i++ {
					offset := layout.PositionOffset +
						group.PositionOffset*layout.PositionStride +
						i*layout.PositionStride
					if _, err := gpuR.Seek(int64(offset), io.SeekStart); err != nil {
						return nil, err
					}
					for _, item := range layout.Items[:layout.NumItems] {
						switch item.Type {
						case ItemPosition:
							if item.Format != FormatVec3F {
								return nil, fmt.Errorf("expected position item to have format [3]float32, but got: %v", item.Format)
							}
							var v [3]float32
							if err := binary.Read(gpuR, binary.LittleEndian, &v); err != nil {
								return nil, err
							}
							mesh.Positions = append(mesh.Positions, v)
						case ItemColor:
							if item.Format != FormatVec4U8 {
								return nil, fmt.Errorf("expected color item to have format [4]uint8, but got: %v", item.Format)
							}
							var v [4]uint8
							if err := binary.Read(gpuR, binary.LittleEndian, &v); err != nil {
								return nil, err
							}
							mesh.Colors = append(mesh.Colors, v)
						case ItemUVCoords:
							switch item.Format {
							case FormatVec2U16:
								var v [2]uint16
								if err := binary.Read(gpuR, binary.LittleEndian, &v); err != nil {
									return nil, err
								}
								mesh.UVCoords = append(mesh.UVCoords, v)
							case FormatVec2F:
								// TODO
							default:
								return nil, fmt.Errorf("expected UV coords item to have format [2]uint16 or [2]float32, but got: %v", item.Format)
							}
						case 5:
							// TODO
						case ItemBoneWeight:
							// TODO
						case ItemBoneIdx:
							// TODO
						default:
							return nil, fmt.Errorf("unknown mesh layout item type: %v", item.Type)
						}
					}
				}
			}
			for _, group := range info.Groups {
				indexStride := layout.IndicesSize / layout.NumIndices
				offset := layout.IndexOffset +
					group.IndexOffset*indexStride
				if _, err := gpuR.Seek(int64(offset), io.SeekStart); err != nil {
					return nil, err
				}
				for i := uint32(0); i < group.NumIndices; i++ {
					var val uint32
					switch indexStride {
					case 4:
						if err := binary.Read(gpuR, binary.LittleEndian, &val); err != nil {
							return nil, err
						}
					case 2:
						var tmp uint16
						if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
							return nil, err
						}
						val = uint32(tmp)
					default:
						return nil, fmt.Errorf("unknown index stride: %v", indexStride)
					}
					mesh.Indices = append(mesh.Indices, val)
				}
			}
			meshes = append(meshes, mesh)
		}
	}

	return &Unit{
		JointTransforms:        jointTransforms,
		JointTransformMatrices: jointTransformMatrices,
		Meshes:                 meshes,
	}, nil
}
