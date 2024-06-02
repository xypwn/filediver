package unit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/go-gl/mathgl/mgl32"

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

type SkeletonRemap struct {
	Unk00   uint32
	Unk01   uint32
	Count   uint32
	Indices []uint32
}

type SkeletonMap struct {
	Count uint32
	// This appears to always be 16?
	Unk00       uint32
	BoneIndices []uint32
	RemapData   SkeletonRemap
	// Maybe inverse bind matrices?
	Matrices [][4][4]float32
}

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
	Matrix      [4][4]float32
	Children    []uint32
}

func (curr *Bone) remap(bones *[]Bone) {
	for _, i := range curr.Children {
		(*bones)[i].remap(bones)
	}
	if curr.Index == curr.ParentIndex {
		return
	}
	parent := (*bones)[curr.ParentIndex]
	currTranslation := mgl32.Vec3(curr.Transform.Translation)
	parentTranslation := mgl32.Vec3(parent.Transform.Translation)
	currTranslation = currTranslation.Sub(parentTranslation)

	currRotation := mgl32.Mat3FromRows(curr.Transform.Rotation[0], curr.Transform.Rotation[1], curr.Transform.Rotation[2])
	parentRotation := mgl32.Mat3FromRows(parent.Transform.Rotation[0], parent.Transform.Rotation[1], parent.Transform.Rotation[2])
	currRotation = currRotation.Mul3(parentRotation.Inv())

	curr.Transform.Translation = currTranslation
	curr.Transform.Rotation = [3][3]float32{currRotation.Row(0), currRotation.Row(1), currRotation.Row(2)}
}

type MeshLayoutItemType uint32

const (
	ItemPosition   MeshLayoutItemType = 0
	ItemColor      MeshLayoutItemType = 1
	ItemUVCoords   MeshLayoutItemType = 4
	ItemBoneIdx    MeshLayoutItemType = 6
	ItemBoneWeight MeshLayoutItemType = 7
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
	FormatF32       MeshLayoutItemFormat = 0
	FormatVec2F     MeshLayoutItemFormat = 1
	FormatVec3F     MeshLayoutItemFormat = 2
	FormatVec4F     MeshLayoutItemFormat = 3
	FormatVec4S8    MeshLayoutItemFormat = 24
	FormatVec4Norm8 MeshLayoutItemFormat = 25
	FormatVec4U8    MeshLayoutItemFormat = 26
	FormatF16       MeshLayoutItemFormat = 28
	FormatVec2F16   MeshLayoutItemFormat = 29
	FormatVec3F16   MeshLayoutItemFormat = 30
	FormatVec4F16   MeshLayoutItemFormat = 31
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
	case FormatVec4S8:
		return "[4]int8"
	case FormatVec4Norm8:
		return "[4]uint8"
	case FormatVec4U8:
		return "[4]uint8"
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
	Unk00        [4]byte
	VertexOffset uint32
	NumVertices  uint32
	IndexOffset  uint32
	NumIndices   uint32
	Unk01        [4]byte
}

type MeshInfo struct {
	Header    MeshHeader
	Materials []stingray.ThinHash
	Groups    []MeshGroup
}

type Header struct {
	Unk00                 [8]byte
	Bones                 stingray.Hash
	Unk01                 [8]byte
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
	UVCoords    [][2]float32
	Colors      [][4]float32
	BoneIndices [][4]uint8
	BoneWeights [][4]float32
	Indices     []uint32
}

type Info struct {
	LODGroups              []LODGroup
	SkeletonMaps           []SkeletonMap
	Bones                  []Bone
	JointTransformMatrices [][4][4]float32
	Materials              map[stingray.ThinHash]stingray.Hash
	NumMeshes              uint32
	meshInfos              []MeshInfo
	meshLayouts            []MeshLayout
}

func loadMesh(gpuR io.ReadSeeker, info MeshInfo, layout MeshLayout) (Mesh, error) {
	var mesh Mesh
	mesh.Info = info
	mesh.Positions = make([][3]float32, 0, layout.NumVertices)
	mesh.UVCoords = make([][2]float32, 0, layout.NumVertices)
	mesh.Colors = make([][4]float32, 0, layout.NumVertices)
	mesh.BoneIndices = make([][4]uint8, 0, layout.NumVertices)
	mesh.BoneWeights = make([][4]float32, 0, layout.NumVertices)
	for _, group := range info.Groups {
		for i := uint32(0); i < group.NumVertices; i++ {
			offset := layout.VertexOffset +
				group.VertexOffset*layout.VertexStride +
				i*layout.VertexStride
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
				case ItemColor:
					var val [4]float32
					switch item.Format {
					case FormatVec4U8:
						var tmp [4]uint8
						if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
							return Mesh{}, err
						}
						for i := range tmp {
							val[i] = float32(tmp[i]) / 255
						}
					case FormatVec4F16:
						var tmp [4]uint16
						if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
							return Mesh{}, err
						}
						for i := range tmp {
							val[i] = float16.Frombits(tmp[i]).Float32()
						}
					default:
						return Mesh{}, fmt.Errorf("expected color item to have format [4]uint8 or [4]float16, but got: %v", item.Format)
					}
					mesh.Colors = append(mesh.Colors, val)
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
					if item.Layer == 0 {
						mesh.UVCoords = append(mesh.UVCoords, val)
					}
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
					case FormatVec4Norm8:
						var tmp [4]uint8
						if err := binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
							return Mesh{}, err
						}
						for i := range tmp {
							val[i] = float32(tmp[i]) / 255.0
						}
					default:
						return Mesh{}, fmt.Errorf("expected bone index item to have format [4]float16 or [4]uint8, but got: %v", item.Format.String())
					}
					mesh.BoneWeights = append(mesh.BoneWeights, val)
				case ItemBoneIdx:
					if item.Format != FormatVec4S8 {
						return Mesh{}, fmt.Errorf("expected bone index item to have format [4]uint8, but got: %v", item.Format.String())
					}
					var val [4]uint8
					if err := binary.Read(gpuR, binary.LittleEndian, &val); err != nil {
						return Mesh{}, err
					}
					mesh.BoneIndices = append(mesh.BoneIndices, val)
				default:
					return Mesh{}, fmt.Errorf("unknown mesh layout item type: %v", item.Type)
				}
			}
		}
	}
	mesh.Indices = make([]uint32, 0, layout.NumIndices)
	for _, group := range info.Groups {
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
			mesh.Indices = append(mesh.Indices, val)
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
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].Unk00); err != nil {
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
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].RemapData.Unk00); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].RemapData.Unk01); err != nil {
				return nil, err
			}
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].RemapData.Count); err != nil {
				return nil, err
			}
			skeletonMapList[i].RemapData.Indices = make([]uint32, skeletonMapList[i].RemapData.Count)
			if err := binary.Read(mainR, binary.LittleEndian, &skeletonMapList[i].RemapData.Indices); err != nil {
				return nil, err
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
			bones[i].Matrix = jointTransformMatrices[i]
			bones[i].NameHash = nameHashes[i]
			bones[i].Transform = jointTransforms[i]
			if bones[i].ParentIndex != uint32(i) {
				bones[jointMap[i].Parent].Children = append(bones[jointMap[i].Parent].Children, uint32(i))
			}
		}

		if jointListHdr.NumJoints > 0 {
			bones[0].remap(&bones)
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
		LODGroups:              lodGroups,
		SkeletonMaps:           skeletonMapList,
		Bones:                  bones,
		JointTransformMatrices: jointTransformMatrices,
		Materials:              materialMap,
		NumMeshes:              uint32(len(meshInfos)),
		meshInfos:              meshInfos,
		meshLayouts:            meshLayouts,
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
		if int(id) > len(info.meshInfos) {
			return nil, fmt.Errorf("mesh ID (%v) is out of bounds of meshes (len=%v)", id, len(info.meshInfos))
		}
		meshInfo := info.meshInfos[id]
		if meshInfo.Header.LayoutIdx < 0 {
			continue
		}
		layout := info.meshLayouts[meshInfo.Header.LayoutIdx]
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
