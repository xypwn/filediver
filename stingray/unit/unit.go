package unit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"

	"github.com/x448/float16"
	"github.com/xypwn/filediver/stingray"
)

type LODEntry struct {
	Detail struct {
		Max float32
		Min float32
	}
	Indices []uint32
}

type LODGroup struct {
	Header struct {
		Unk00     uint32
		UnkHash00 stingray.ThinHash
		Unk01     uint32
		Unk02     uint32
	}
	Entries []LODEntry
	Footer  struct {
		UnkFloats00 [7]float32
		Unk00       uint32
		UnkFloat00  float32
		Unk01       uint32
	}
}

type RemapItem struct {
	IndexDataOffset uint32
	IndexCount      uint32
}

type SkeletonMap struct {
	Count       uint32
	Matrices    [][4][4]float32
	BoneIndices []uint32
	RemapList   [][]uint32
}

type JointTransform struct {
	Rotation    mgl32.Mat3
	Translation mgl32.Vec3
	Scale       mgl32.Vec3
	Skew        float32
}

type JointListHeader struct {
	NumJoints uint32
	Unk00     [12]byte
}

type JointMapEntry struct {
	Increment uint16
	Parent    uint16
}

type Bone struct {
	NameHash    stingray.ThinHash
	Index       uint32
	ParentIndex uint32
	Increment   uint32
	Transform   JointTransform
	Matrix      mgl32.Mat4
	Children    []uint32
}

func (curr *Bone) RecursiveCalcLocalTransforms(bones *[]Bone) {
	for _, i := range curr.Children {
		(*bones)[i].RecursiveCalcLocalTransforms(bones)
	}

	currTransform := curr.Matrix
	if curr.Index != curr.ParentIndex {
		parent := (*bones)[curr.ParentIndex]
		parentTransform := parent.Matrix
		currTransform = parentTransform.Inv().Mul4(currTransform)
	}

	curr.setTransforms(currTransform)
}

func (curr *Bone) setTransforms(matrix mgl32.Mat4) {
	curr.Transform.Translation = matrix.Col(3).Vec3()
	mat3 := matrix.Mat3()
	curr.Transform.Scale = mgl32.Vec3{mat3.Row(0).Len(), mat3.Row(1).Len(), mat3.Row(2).Len()}
	invScale := mgl32.Vec3{1 / curr.Transform.Scale[0], 1 / curr.Transform.Scale[1], 1 / curr.Transform.Scale[1]}
	curr.Transform.Rotation = mat3.Mul3(mgl32.Diag3(invScale))
}

type MeshLayoutItemType uint32

const (
	ItemPosition   MeshLayoutItemType = 0
	ItemNormal     MeshLayoutItemType = 1
	ItemUVCoords   MeshLayoutItemType = 4
	ItemBoneIdx    MeshLayoutItemType = 6
	ItemBoneWeight MeshLayoutItemType = 7
)

func (v MeshLayoutItemType) String() string {
	switch v {
	case ItemPosition:
		return "position"
	case ItemNormal:
		return "normal"
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
	FormatF32                      MeshLayoutItemFormat = 0
	FormatVec2F                    MeshLayoutItemFormat = 1
	FormatVec3F                    MeshLayoutItemFormat = 2
	FormatVec4F                    MeshLayoutItemFormat = 3
	FormatU32                      MeshLayoutItemFormat = 17
	FormatVec2U32                  MeshLayoutItemFormat = 18
	FormatVec3U32                  MeshLayoutItemFormat = 19
	FormatVec4U32                  MeshLayoutItemFormat = 20
	FormatS8                       MeshLayoutItemFormat = 21
	FormatVec2S8                   MeshLayoutItemFormat = 22
	FormatVec3S8                   MeshLayoutItemFormat = 23
	FormatVec4S8                   MeshLayoutItemFormat = 24
	FormatVec4R10G10B10A2_TYPELESS MeshLayoutItemFormat = 25
	FormatVec4R10G10B10A2_UNORM    MeshLayoutItemFormat = 26
	FormatF16                      MeshLayoutItemFormat = 28
	FormatVec2F16                  MeshLayoutItemFormat = 29
	FormatVec3F16                  MeshLayoutItemFormat = 30
	FormatVec4F16                  MeshLayoutItemFormat = 31
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
	case FormatU32:
		return "uint32"
	case FormatVec2U32:
		return "[2]uint32"
	case FormatVec3U32:
		return "[3]uint32"
	case FormatVec4U32:
		return "[4]uint32"
	case FormatS8:
		return "int8"
	case FormatVec2S8:
		return "[2]int8"
	case FormatVec3S8:
		return "[3]int8"
	case FormatVec4S8:
		return "[4]int8"
	case FormatVec4R10G10B10A2_TYPELESS:
		return "packed32"
	case FormatVec4R10G10B10A2_UNORM:
		return "packed32u"
	case FormatF16:
		return "float16"
	case FormatVec2F16:
		return "[2]float16"
	case FormatVec3F16:
		return "[3]float16"
	case FormatVec4F16:
		return "[4]float16"
	default:
		return fmt.Sprint(uint32(v))
	}
}

func (v MeshLayoutItemFormat) Size() int {
	switch v {
	case FormatF32:
		return 4
	case FormatVec2F:
		return 8
	case FormatVec3F:
		return 12
	case FormatVec4F:
		return 16
	case FormatU32:
		return 4
	case FormatVec2U32:
		return 8
	case FormatVec3U32:
		return 12
	case FormatVec4U32:
		return 16
	case FormatS8:
		return 1
	case FormatVec2S8:
		return 2
	case FormatVec3S8:
		return 3
	case FormatVec4S8:
		return 4
	case FormatVec4R10G10B10A2_TYPELESS:
		return 4
	case FormatVec4R10G10B10A2_UNORM:
		return 4
	case FormatF16:
		return 2
	case FormatVec2F16:
		return 4
	case FormatVec3F16:
		return 6
	case FormatVec4F16:
		return 8
	default:
		return -1
	}
}

func (v MeshLayoutItemFormat) ComponentType() gltf.ComponentType {
	switch v {
	case FormatF32:
		return gltf.ComponentFloat
	case FormatVec2F:
		return gltf.ComponentFloat
	case FormatVec3F:
		return gltf.ComponentFloat
	case FormatVec4F:
		return gltf.ComponentFloat
	case FormatU32:
		return gltf.ComponentUint
	case FormatVec2U32:
		return gltf.ComponentUint
	case FormatVec3U32:
		return gltf.ComponentUint
	case FormatVec4U32:
		return gltf.ComponentUint
	case FormatS8:
		return gltf.ComponentUbyte
	case FormatVec2S8:
		return gltf.ComponentUbyte
	case FormatVec3S8:
		return gltf.ComponentUbyte
	case FormatVec4S8:
		return gltf.ComponentUbyte
	case FormatVec4R10G10B10A2_TYPELESS:
		return gltf.ComponentUbyte
	case FormatVec4R10G10B10A2_UNORM:
		return gltf.ComponentUbyte
	case FormatF16:
		return gltf.ComponentUshort
	case FormatVec2F16:
		return gltf.ComponentUshort
	case FormatVec3F16:
		return gltf.ComponentUshort
	case FormatVec4F16:
		return gltf.ComponentUshort
	default:
		return gltf.ComponentByte
	}
}

func (v MeshLayoutItemFormat) Type() gltf.AccessorType {
	switch v {
	case FormatF32:
		return gltf.AccessorScalar
	case FormatVec2F:
		return gltf.AccessorVec2
	case FormatVec3F:
		return gltf.AccessorVec3
	case FormatVec4F:
		return gltf.AccessorVec4
	case FormatU32:
		return gltf.AccessorScalar
	case FormatVec2U32:
		return gltf.AccessorVec2
	case FormatVec3U32:
		return gltf.AccessorVec3
	case FormatVec4U32:
		return gltf.AccessorVec4
	case FormatS8:
		return gltf.AccessorScalar
	case FormatVec2S8:
		return gltf.AccessorVec2
	case FormatVec3S8:
		return gltf.AccessorVec3
	case FormatVec4S8:
		return gltf.AccessorVec4
	case FormatVec4R10G10B10A2_TYPELESS:
		return gltf.AccessorVec4
	case FormatVec4R10G10B10A2_UNORM:
		return gltf.AccessorVec4
	case FormatF16:
		return gltf.AccessorScalar
	case FormatVec2F16:
		return gltf.AccessorVec2
	case FormatVec3F16:
		return gltf.AccessorVec3
	case FormatVec4F16:
		return gltf.AccessorVec4
	default:
		return gltf.AccessorScalar
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
	NumItems      uint32
	Unk01         [4]byte
	MagicNum1     [4]byte
	Unk02         [12]byte
	NumVertices   uint32
	VertexStride  uint32
	Unk03         [16]byte
	MagicNum2     [4]byte
	Unk04         [12]byte
	NumIndices    uint32
	Unk05         [20]byte
	VertexOffset  uint32
	PositionsSize uint32
	IndexOffset   uint32
	IndicesSize   uint32
	Unk06         [16]byte
}

type MeshHeaderType uint32

const (
	MeshTypeUnknown00 MeshHeaderType = 0x000
	MeshTypeUnknown01 MeshHeaderType = 0x10a
	MeshTypeUnknown02 MeshHeaderType = 0x189
)

func (m *MeshHeaderType) String() string {
	switch *m {
	case MeshTypeUnknown00:
		return "Bounding Box, maybe?"
	case MeshTypeUnknown01:
		return "Low detail?"
	case MeshTypeUnknown02:
		return "Full mesh?"
	default:
		return fmt.Sprintf("Very unknown value for mesh type: %08x", *m)
	}
}

type MeshHeader struct {
	Unk00 [8]byte
	AABB  struct {
		Min [3]float32
		Max [3]float32
	}
	UnkFloat00         float32
	MeshType           MeshHeaderType
	GroupBoneHash      stingray.ThinHash
	AABBTransformIndex uint32
	TransformIdx       uint32
	UnkInt03           uint32
	SkeletonMapIdx     int32
	LayoutIdx          int32
	Unk01              [40]byte
	NumMaterials       uint32
	MaterialOffset     uint32
	Unk02              [8]byte
	NumGroups          uint32
	GroupOffset        uint32
}

type MeshGroup struct {
	GroupIdx       uint32
	VertexOffset   uint32
	NumVertices    uint32
	IndexOffset    uint32
	NumIndices     uint32
	RepeatGroupIdx uint32
}

type MeshInfo struct {
	Header    MeshHeader
	Materials []stingray.ThinHash
	Groups    []MeshGroup
}

type Header struct {
	Unk00                 [8]byte
	Bones                 stingray.Hash
	GeometryGroup         stingray.Hash
	UnkHash00             stingray.Hash
	StateMachine          stingray.Hash
	Unk02                 [8]byte
	LODGroupListOffset    uint32
	JointListOffset       uint32
	UnkOffset01           uint32
	UnkOffset02           uint32
	Unk03                 [12]byte
	UnkOffset03           uint32
	UnkOffset04           uint32
	Unk04                 [4]byte
	SkeletonMapListOffset uint32
	MeshLayoutListOffset  uint32
	MeshDataOffset        uint32
	MeshInfoListOffset    uint32
	Unk05                 [8]byte
	MaterialListOffset    uint32
}

type Mesh struct {
	Info        MeshInfo
	Positions   [][3]float32
	UVCoords    [][][2]float32
	Normals     [][3]float32
	BoneIndices [][][4]uint8
	BoneWeights [][4]float32
	Indices     [][]uint32
}

type Info struct {
	GeometryGroup          stingray.Hash
	LODGroups              []LODGroup
	SkeletonMaps           []SkeletonMap
	Bones                  []Bone
	JointTransformMatrices [][4][4]float32
	Materials              map[stingray.ThinHash]stingray.Hash
	NumMeshes              uint32
	MeshInfos              []MeshInfo
	MeshLayouts            []MeshLayout
}

func loadMesh(gpuR io.ReadSeeker, info MeshInfo, layout MeshLayout) (Mesh, error) {
	var mesh Mesh
	var uvCoordLayers uint32 = 1
	var boneIdxLayers uint32 = 1
	for i := 0; i < int(layout.NumItems); i += 1 {
		switch layout.Items[i].Type {
		case ItemBoneIdx:
			if layout.Items[i].Layer >= boneIdxLayers {
				boneIdxLayers = layout.Items[i].Layer + 1
			}
		case ItemUVCoords:
			if layout.Items[i].Layer >= uvCoordLayers {
				uvCoordLayers = layout.Items[i].Layer + 1
			}
		}
	}
	mesh.Info = info
	mesh.Positions = make([][3]float32, 0, layout.NumVertices)
	mesh.UVCoords = make([][][2]float32, uvCoordLayers)
	for layer := 0; layer < int(uvCoordLayers); layer++ {
		mesh.UVCoords[layer] = make([][2]float32, 0, layout.NumVertices)
	}
	mesh.Normals = make([][3]float32, 0, layout.NumVertices)
	mesh.BoneIndices = make([][][4]uint8, boneIdxLayers)
	for layer := 0; layer < int(boneIdxLayers); layer++ {
		mesh.BoneIndices[layer] = make([][4]uint8, 0, layout.NumVertices)
	}
	mesh.BoneWeights = make([][4]float32, 0, layout.NumVertices)
	for i := uint32(0); i < layout.NumVertices; i++ {
		offset := layout.VertexOffset + i*layout.VertexStride
		if _, err := gpuR.Seek(int64(offset), io.SeekStart); err != nil {
			return Mesh{}, err
		}
		for _, item := range layout.Items[:layout.NumItems] {
			switch item.Type {
			case ItemPosition:
				if item.Format != FormatVec3F {
					return Mesh{}, fmt.Errorf("expected position item to have format [3]float32, but got: %v", item.Format)
				}
				var v [3]float32
				if err := binary.Read(gpuR, binary.LittleEndian, &v); err != nil {
					return Mesh{}, err
				}
				mesh.Positions = append(mesh.Positions, v)
			case ItemNormal:
				var val [3]float32
				switch item.Format {
				case FormatVec4R10G10B10A2_UNORM:
					var tmp uint32
					if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
						return Mesh{}, err
					}
					val[0] = float32(tmp&0x3ff) / 1023.0
					val[1] = float32((tmp>>10)&0x3ff) / 1023.0
					val[2] = float32((tmp>>20)&0x3ff) / 1023.0
				case FormatVec4F16:
					var tmp [4]uint16
					if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
						return Mesh{}, err
					}
					for i := range val {
						val[i] = float16.Frombits(tmp[i]).Float32()
					}
				default:
					return Mesh{}, fmt.Errorf("expected normal item to have format packed32u or [4]float16, but got: %v", item.Format)
				}
				mesh.Normals = append(mesh.Normals, val)
			case 2:
				if item.Format != FormatVec4F16 {
					return Mesh{}, fmt.Errorf("expected type 2 item to have format [4]float16, but got: %v", item.Format)
				}
				//fmt.Println(item.Format)
				// TODO
			case 3:
				if item.Format != FormatVec4F16 {
					return Mesh{}, fmt.Errorf("expected type 3 item to have format [4]float16, but got: %v", item.Format)
				}
				//fmt.Println(item.Format)
				// TODO
			case ItemUVCoords:
				var val [2]float32
				switch item.Format {
				case FormatVec2F16:
					var tmp [2]uint16
					if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
						return Mesh{}, err
					}
					for i := range tmp {
						val[i] = float16.Frombits(tmp[i]).Float32()
					}
				case FormatVec2F:
					if err := binary.Read(gpuR, binary.LittleEndian, &val); err != nil {
						return Mesh{}, err
					}
				default:
					return Mesh{}, fmt.Errorf("expected UV coords item to have format [2]float16 or [2]float32, but got: %v", item.Format)
				}
				mesh.UVCoords[item.Layer] = append(mesh.UVCoords[item.Layer], val)
			case 5:
				if item.Format != 4 {
					return Mesh{}, fmt.Errorf("expected type 5 item to have format [4]uint8, but got: %v", item.Format)
				}
				var v [4]uint8
				if err := binary.Read(gpuR, binary.LittleEndian, &v); err != nil {
					return Mesh{}, err
				}
				//fmt.Println(v)
				_ = v
				// TODO
			case ItemBoneWeight:
				var val [4]float32
				switch item.Format {
				case FormatVec4F16:
					var tmp [4]uint16
					if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
						return Mesh{}, err
					}
					for i := range tmp {
						val[i] = float16.Frombits(tmp[i]).Float32()
					}
				case FormatVec4R10G10B10A2_TYPELESS:
					var tmp uint32
					if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
						return Mesh{}, err
					}
					val[0] = float32(tmp&0x3ff) / 1023.0
					val[1] = float32((tmp>>10)&0x3ff) / 1023.0
					val[2] = float32((tmp>>20)&0x3ff) / 1023.0
					val[3] = 0.0 // float32((tmp>>30)&0x3) / 3.0 // This causes issues with incorrect bone weights
				case FormatVec2F16:
					var tmp [2]uint16
					if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
						return Mesh{}, err
					}
					for i := range tmp {
						val[i] = float16.Frombits(tmp[i]).Float32()
					}
					val[2] = 0
					val[3] = 0
				case FormatF32:
					binary.Read(gpuR, binary.LittleEndian, &val[0])
				default:
					return Mesh{}, fmt.Errorf("expected bone weight item to have format float32, [4]float16, [2]float16, or packed32, but got: %v", item.Format.String())
				}
				mesh.BoneWeights = append(mesh.BoneWeights, val)
			case ItemBoneIdx:
				var val [4]uint8
				switch item.Format {
				case FormatVec4S8:
					if err := binary.Read(gpuR, binary.LittleEndian, &val); err != nil {
						return Mesh{}, err
					}
				case FormatVec4U32:
					var tmp [4]uint32
					if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
						return Mesh{}, err
					}
					for i := range tmp {
						if tmp[i] > 0xff {
							return Mesh{}, fmt.Errorf("unexpected bone index value - %v exceeds max u8 value", tmp[i])
						}
						val[i] = uint8(tmp[i])
					}
				default:
					return Mesh{}, fmt.Errorf("expected bone index item to have format [4]uint8 or [4]uint32, but got: %v", item.Format.String())
				}
				mesh.BoneIndices[item.Layer] = append(mesh.BoneIndices[item.Layer], val)
			default:
				return Mesh{}, fmt.Errorf("unknown mesh layout item type: %v", item.Type)
			}
		}
	}
	mesh.Indices = make([][]uint32, len(info.Groups))
	for grp, group := range info.Groups {
		mesh.Indices[grp] = make([]uint32, 0, group.NumIndices)
		indexStride := layout.IndicesSize / layout.NumIndices
		offset := layout.IndexOffset +
			group.IndexOffset*indexStride
		if _, err := gpuR.Seek(int64(offset), io.SeekStart); err != nil {
			return Mesh{}, err
		}
		for i := uint32(0); i < group.NumIndices; i++ {
			var val uint32
			switch indexStride {
			case 4:
				if err := binary.Read(gpuR, binary.LittleEndian, &val); err != nil {
					return Mesh{}, err
				}
			case 2:
				var tmp uint16
				if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
					return Mesh{}, err
				}
				val = uint32(tmp)
			default:
				return Mesh{}, fmt.Errorf("unknown index stride: %v", indexStride)
			}
			mesh.Indices[grp] = append(mesh.Indices[grp], val+group.VertexOffset)
		}
	}
	return mesh, nil
}

func LoadInfo(mainR io.ReadSeeker) (*Info, error) {
	var hdr Header
	if err := binary.Read(mainR, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}

	var lodGroups []LODGroup
	if hdr.LODGroupListOffset != 0 {
		if _, err := mainR.Seek(int64(hdr.LODGroupListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		lodGroupOffsets := make([]uint32, count)
		for i := range lodGroupOffsets {
			if err := binary.Read(mainR, binary.LittleEndian, &lodGroupOffsets[i]); err != nil {
				return nil, err
			}
		}
		lodGroups = make([]LODGroup, count)
		for i, lodGroupOffset := range lodGroupOffsets {
			if _, err := mainR.Seek(int64(hdr.LODGroupListOffset+lodGroupOffset), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &lodGroups[i].Header); err != nil {
				return nil, err
			}
			var count uint32
			if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
				return nil, err
			}
			entryOffsets := make([]uint32, count)
			for j := range entryOffsets {
				if err := binary.Read(mainR, binary.LittleEndian, &entryOffsets[j]); err != nil {
					return nil, err
				}
			}
			var footerOffset uint32
			if err := binary.Read(mainR, binary.LittleEndian, &footerOffset); err != nil {
				return nil, err
			}
			lodGroups[i].Entries = make([]LODEntry, count)
			for j, entryOffset := range entryOffsets {
				if _, err := mainR.Seek(int64(hdr.LODGroupListOffset+lodGroupOffset+entryOffset), io.SeekStart); err != nil {
					return nil, err
				}
				if err := binary.Read(mainR, binary.LittleEndian, &lodGroups[i].Entries[j].Detail); err != nil {
					return nil, err
				}
				var count uint32
				if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
					return nil, err
				}
				data := make([]uint32, count)
				for k := range data {
					if err := binary.Read(mainR, binary.LittleEndian, &data[k]); err != nil {
						return nil, err
					}
				}
				lodGroups[i].Entries[j].Indices = data
			}
			if _, err := mainR.Seek(int64(hdr.LODGroupListOffset+lodGroupOffset+footerOffset), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &lodGroups[i].Footer); err != nil {
				return nil, err
			}
		}
	}

	var skeletonMapList []SkeletonMap
	if hdr.SkeletonMapListOffset != 0 {
		if _, err := mainR.Seek(int64(hdr.SkeletonMapListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		skeletonMapOffsets := make([]uint32, count)
		if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapOffsets); err != nil {
			return nil, err
		}
		skeletonMapList = make([]SkeletonMap, count)
		for i, skeletonMapOffset := range skeletonMapOffsets {
			if _, err := mainR.Seek(int64(hdr.SkeletonMapListOffset+skeletonMapOffset), io.SeekStart); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].Count); err != nil {
				return nil, err
			}
			var matricesOffset uint32
			if err := binary.Read(mainR, binary.LittleEndian, &matricesOffset); err != nil {
				return nil, err
			}
			var indicesOffset uint32
			if err := binary.Read(mainR, binary.LittleEndian, &indicesOffset); err != nil {
				return nil, err
			}
			var remapOffset uint32
			if err := binary.Read(mainR, binary.LittleEndian, &remapOffset); err != nil {
				return nil, err
			}
			if _, err := mainR.Seek(int64(hdr.SkeletonMapListOffset+skeletonMapOffset+matricesOffset), io.SeekStart); err != nil {
				return nil, err
			}
			skeletonMapList[i].Matrices = make([][4][4]float32, skeletonMapList[i].Count)
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].Matrices); err != nil {
				return nil, err
			}
			if _, err := mainR.Seek(int64(hdr.SkeletonMapListOffset+skeletonMapOffset+indicesOffset), io.SeekStart); err != nil {
				return nil, err
			}
			skeletonMapList[i].BoneIndices = make([]uint32, skeletonMapList[i].Count)
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].BoneIndices); err != nil {
				return nil, err
			}
			if _, err := mainR.Seek(int64(hdr.SkeletonMapListOffset+skeletonMapOffset+remapOffset), io.SeekStart); err != nil {
				return nil, err
			}
			var remapListCount uint32
			if err := binary.Read(mainR, binary.LittleEndian, &remapListCount); err != nil {
				return nil, err
			}
			var remapListItems []RemapItem = make([]RemapItem, remapListCount)
			if err := binary.Read(mainR, binary.LittleEndian, &remapListItems); err != nil {
				return nil, err
			}
			skeletonMapList[i].RemapList = make([][]uint32, remapListCount)
			for j := range remapListItems {
				skeletonMapList[i].RemapList[j] = make([]uint32, remapListItems[j].IndexCount)
				if _, err := mainR.Seek(int64(hdr.SkeletonMapListOffset+skeletonMapOffset+remapOffset+remapListItems[j].IndexDataOffset), io.SeekStart); err != nil {
					return nil, err
				}
				if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].RemapList[j]); err != nil {
					return nil, err
				}
			}
		}
	}

	var jointListHdr JointListHeader
	var jointTransforms []JointTransform
	var jointTransformMatrices [][4][4]float32
	var jointMap []JointMapEntry
	var nameHashes []stingray.ThinHash
	var bones []Bone
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
		jointMap = make([]JointMapEntry, jointListHdr.NumJoints)
		for i := range jointMap {
			if err := binary.Read(mainR, binary.LittleEndian, &jointMap[i]); err != nil {
				return nil, err
			}
		}
		nameHashes = make([]stingray.ThinHash, jointListHdr.NumJoints)
		for i := range nameHashes {
			if err := binary.Read(mainR, binary.LittleEndian, &nameHashes[i]); err != nil {
				return nil, err
			}
		}
		bones = make([]Bone, jointListHdr.NumJoints)
		for i := range bones {
			bones[i].Index = uint32(i)
			bones[i].ParentIndex = uint32(jointMap[i].Parent)
			bones[i].Increment = uint32(jointMap[i].Increment)
			jtm := jointTransformMatrices[i]
			bones[i].Matrix = mgl32.Mat4FromRows(jtm[0], jtm[1], jtm[2], jtm[3]).Transpose()
			bones[i].NameHash = nameHashes[i]
			bones[i].Transform = jointTransforms[i]
			if bones[i].ParentIndex != uint32(i) {
				bones[jointMap[i].Parent].Children = append(bones[jointMap[i].Parent].Children, uint32(i))
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
			materials := make([]stingray.ThinHash, hdr.NumMaterials)
			for j := range materials {
				if err := binary.Read(mainR, binary.LittleEndian, &materials[j]); err != nil {
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
				Header:    hdr,
				Materials: materials,
				Groups:    groups,
			})
		}
	}

	materialMap := make(map[stingray.ThinHash]stingray.Hash)
	if hdr.MaterialListOffset > 0 {
		if _, err := mainR.Seek(int64(hdr.MaterialListOffset), io.SeekStart); err != nil {
			return nil, err
		}
		var count uint32
		if err := binary.Read(mainR, binary.LittleEndian, &count); err != nil {
			return nil, err
		}
		keys := make([]stingray.ThinHash, count)
		for i := range keys {
			if err := binary.Read(mainR, binary.LittleEndian, &keys[i]); err != nil {
				return nil, err
			}
		}
		values := make([]stingray.Hash, count)
		for i := range values {
			if err := binary.Read(mainR, binary.LittleEndian, &values[i]); err != nil {
				return nil, err
			}
		}
		for i := range keys {
			materialMap[keys[i]] = values[i]
		}
	}

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
		for _, meshInfo := range meshInfos {
			if meshInfo.Header.LayoutIdx < 0 {
				continue
			}
			if int(meshInfo.Header.LayoutIdx) >= len(meshLayouts) {
				return nil, fmt.Errorf("mesh layout index (%v) is out of bounds of mesh layouts (len=%v)", meshInfo.Header.LayoutIdx, len(meshLayouts))
			}
		}
	}

	return &Info{
		GeometryGroup:          hdr.GeometryGroup,
		LODGroups:              lodGroups,
		SkeletonMaps:           skeletonMapList,
		Bones:                  bones,
		JointTransformMatrices: jointTransformMatrices,
		Materials:              materialMap,
		NumMeshes:              uint32(len(meshInfos)),
		MeshInfos:              meshInfos,
		MeshLayouts:            meshLayouts,
	}, nil
}

// idsToLoad contains the mesh IDs (=indices) of the meshes which should be loaded.
// To load all meshes, pass a slice with value {0,1,...,info.NumMeshes-1}.
func LoadMeshes(gpuR io.ReadSeeker, info *Info, idsToLoad []uint32) (map[uint32]Mesh, error) {
	meshes := make(map[uint32]Mesh)
	for _, id := range idsToLoad {
		if _, ok := meshes[id]; ok {
			continue
		}
		if int(id) > len(info.MeshInfos) {
			return nil, fmt.Errorf("mesh ID (%v) is out of bounds of meshes (len=%v)", id, len(info.MeshInfos))
		}
		meshInfo := info.MeshInfos[id]
		if meshInfo.Header.LayoutIdx < 0 {
			continue
		}
		layout := info.MeshLayouts[meshInfo.Header.LayoutIdx]
		if len(meshInfo.Groups) > 0 && gpuR == nil {
			return nil, errors.New("mesh group exists, but GPU resource data is nil")
		}
		mesh, err := loadMesh(gpuR, meshInfo, layout)
		if err != nil {
			return nil, err
		}
		meshes[id] = mesh
	}
	return meshes, nil
}
