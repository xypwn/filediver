package geometry

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/x448/float16"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
)

type MeshInfo struct {
	Groups          []unit.MeshGroup
	Materials       []stingray.ThinHash
	MeshLayoutIndex uint32
}

type AccessorInfo struct {
	gltf.AccessorType
	gltf.ComponentType
	Size uint32
}

func convertFloat16Slice(gpuR io.ReadSeeker, data []byte, tmpArr []uint16, extra uint32) ([]byte, uint32, error) {
	var err error
	if err = binary.Read(gpuR, binary.LittleEndian, &tmpArr); err != nil {
		return nil, 0, err
	}
	var size uint32 = extra * 4
	for _, tmp := range tmpArr {
		data, err = binary.Append(data, binary.LittleEndian, float16.Frombits(tmp).Float32())
		if err != nil {
			return nil, 0, err
		}
		size += 4
	}
	data = append(data, make([]byte, extra*4)...)
	return data, size, nil
}

func convertVertices(gpuR io.ReadSeeker, layout unit.MeshLayout) ([]byte, []AccessorInfo, error) {
	data := make([]byte, 0)
	dataLen := len(data)
	accessorStructure := make([]AccessorInfo, 0, layout.NumItems)
	if _, err := gpuR.Seek(int64(layout.VertexOffset), io.SeekStart); err != nil {
		return nil, nil, err
	}
	for vertex := 0; vertex < int(layout.NumVertices); vertex += 1 {
		for idx := 0; idx < int(layout.NumItems); idx += 1 {
			item := layout.Items[idx]
			switch item.Format {
			case unit.FormatVec4R10G10B10A2_TYPELESS:
				var tmp uint32
				var val [4]float32
				var err error
				if err = binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
					return nil, nil, err
				}
				val[0] = float32(tmp&0x3ff) / 1023.0
				val[1] = float32((tmp>>10)&0x3ff) / 1023.0
				val[2] = float32((tmp>>20)&0x3ff) / 1023.0
				val[3] = 0.0 // float32((tmp>>30)&0x3) / 3.0 // This causes issues with incorrect bone weights
				data, err = binary.Append(data, binary.LittleEndian, val)
				if err != nil {
					return nil, nil, err
				}
				if vertex == 0 {
					accessorStructure = append(accessorStructure, AccessorInfo{
						AccessorType:  gltf.AccessorVec4,
						ComponentType: gltf.ComponentFloat,
						Size:          16,
					})
				}
			case unit.FormatVec4R10G10B10A2_UNORM:
				var tmp uint32
				var val [3]float32
				var err error
				if err = binary.Read(gpuR, binary.LittleEndian, &tmp); err != nil {
					return nil, nil, err
				}
				val[0] = float32(tmp&0x3ff) / 1023.0
				val[1] = float32((tmp>>10)&0x3ff) / 1023.0
				val[2] = float32((tmp>>20)&0x3ff) / 1023.0
				data, err = binary.Append(data, binary.LittleEndian, val)
				if err != nil {
					return nil, nil, err
				}
				if vertex == 0 {
					accessorStructure = append(accessorStructure, AccessorInfo{
						AccessorType:  gltf.AccessorVec3,
						ComponentType: gltf.ComponentFloat,
						Size:          12,
					})
				}
			case unit.FormatF16:
				fallthrough
			case unit.FormatVec2F16:
				fallthrough
			case unit.FormatVec3F16:
				fallthrough
			case unit.FormatVec4F16:
				tmpArr := make([]uint16, item.Format.Type().Components())
				var err error
				var size, extra uint32
				var accessorType gltf.AccessorType = item.Format.Type()
				if item.Type == unit.ItemBoneWeight && item.Format == unit.FormatVec2F16 {
					accessorType = gltf.AccessorVec4
					extra = 2
				}
				data, size, err = convertFloat16Slice(gpuR, data, tmpArr, extra)
				if err != nil {
					return nil, nil, err
				}
				if vertex == 0 {
					accessorStructure = append(accessorStructure, AccessorInfo{
						AccessorType:  accessorType,
						ComponentType: gltf.ComponentFloat,
						Size:          size,
					})
				}
			case unit.FormatF32:
				fallthrough
			case unit.FormatVec2F:
				fallthrough
			case unit.FormatVec3F:
				fallthrough
			case unit.FormatVec4F:
				fallthrough
			case unit.FormatS32:
				fallthrough
			case unit.FormatS8:
				fallthrough
			case unit.FormatVec2S8:
				fallthrough
			case unit.FormatVec3S8:
				fallthrough
			case unit.FormatVec4S8:
				fallthrough
			case unit.FormatU32:
				fallthrough
			case unit.FormatVec2U32:
				fallthrough
			case unit.FormatVec3U32:
				fallthrough
			case unit.FormatVec4U32:
				data = append(data, make([]byte, item.Format.Size())...)
				if _, err := gpuR.Read(data[dataLen:]); err != nil {
					return nil, nil, err
				}
				if item.Type == unit.ItemBoneWeight && item.Format == unit.FormatF32 {
					var err error
					data, err = binary.Append(data, binary.LittleEndian, [3]float32{})
					if err != nil {
						return nil, nil, err
					}
					item.Format = unit.FormatVec4F
				}
				if vertex == 0 {
					accessorStructure = append(accessorStructure, AccessorInfo{
						AccessorType:  item.Format.Type(),
						ComponentType: item.Format.ComponentType(),
						Size:          uint32(item.Format.Size()),
					})
				}
			default:
				return nil, nil, fmt.Errorf("Unknown format %v for type %v\n", item.Format.String(), item.Type.String())
			}
			dataLen = len(data)
		}
	}
	return data, accessorStructure, nil
}

func getMeshNameFbxConvertAndTransformBone(unitInfo *unit.Info, groupBoneHash stingray.ThinHash) (meshNameBoneIdx int, fbxConvertIdx int, transformBoneIdx int) {
	parentIdx := -1
	fbxConvertIdx = -1
	meshNameBoneIdx = -1
	transformBoneIdx = -1
	gameMeshHash := stingray.Sum64([]byte("game_mesh")).Thin()
	fbxConvertHash := stingray.Sum64([]byte("FbxAxisSystem_ConvertNode")).Thin()
	for boneIdx, bone := range unitInfo.Bones {
		if bone.NameHash == gameMeshHash {
			parentIdx = boneIdx
		}
		if bone.NameHash == fbxConvertHash {
			fbxConvertIdx = boneIdx
		}
		if bone.ParentIndex == uint32(parentIdx) && bone.NameHash == groupBoneHash {
			transformBoneIdx = boneIdx
		}
		if bone.ParentIndex == uint32(fbxConvertIdx) {
			meshNameBoneIdx = boneIdx
		}
	}
	return
}

func remapJoint[E ~[]I, I uint8 | uint32](idxs E, remapList, remappedBoneIndices []uint32) {
	for k := 0; k < 4; k++ {
		if uint32(idxs[k]) >= uint32(len(remapList)) {
			continue
		}
		remapIndex := remapList[idxs[k]]
		idxs[k] = I(remappedBoneIndices[remapIndex])
	}
}

func remapJoints(buffer *gltf.Buffer, stride, bufferOffset, vertexCount uint32, indices []uint32, componentType gltf.ComponentType, remapList, remappedBoneIndices []uint32) error {
	remappedVertices := make(map[uint32]bool)
	for _, vertex := range indices {
		if vertex >= vertexCount {
			continue
		}
		if _, contains := remappedVertices[vertex]; contains {
			continue
		}
		remappedVertices[vertex] = true
		if componentType == gltf.ComponentUbyte {
			boneIndices := make([]uint8, 4)
			if _, err := binary.Decode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, &boneIndices); err != nil {
				return err
			}
			remapJoint(boneIndices, remapList, remappedBoneIndices)
			if _, err := binary.Encode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, boneIndices); err != nil {
				return err
			}
		} else {
			boneIndices := make([]uint32, 4)
			if _, err := binary.Decode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, &boneIndices); err != nil {
				return err
			}
			remapJoint(boneIndices, remapList, remappedBoneIndices)
			if _, err := binary.Encode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, boneIndices); err != nil {
				return err
			}
		}
	}
	return nil
}

func addMeshLayoutVertexBuffer(doc *gltf.Document, data []byte, accessorInfo []AccessorInfo) (uint32, error) {
	ensurePadding(doc)
	buffer := lastBuffer(doc)
	offset := uint32(len(buffer.Data))

	buffer.Data = append(buffer.Data, data...)
	buffer.ByteLength += uint32(len(data))
	convertedVertexStride := uint32(0)
	for _, info := range accessorInfo {
		convertedVertexStride += info.Size
	}

	doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
		Buffer:     uint32(len(doc.Buffers)) - 1,
		ByteLength: uint32(len(data)),
		ByteOffset: offset,
		ByteStride: convertedVertexStride,
		Target:     gltf.TargetArrayBuffer,
	})

	return uint32(len(doc.BufferViews) - 1), nil
}

func createAttributes(doc *gltf.Document, layout unit.MeshLayout, accessorInfo []AccessorInfo) map[string]*gltf.Accessor {
	attributes := make(map[string]*gltf.Accessor)

	var byteOffset uint32 = 0
	for j := 0; j < int(layout.NumItems); j++ {
		item := accessorInfo[j]
		accessor := &gltf.Accessor{
			BufferView:    gltf.Index(uint32(len(doc.BufferViews)) - 1),
			ByteOffset:    uint32(byteOffset),
			ComponentType: item.ComponentType,
			Type:          item.AccessorType,
			Count:         layout.NumVertices,
		}
		byteOffset += item.Size
		if layout.Items[j].Type == unit.ItemBoneIdx && layout.Items[j].Layer != 0 {
			continue
		}
		switch layout.Items[j].Type {
		case unit.ItemPosition:
			attributes[gltf.POSITION] = accessor
		case unit.ItemNormal:
			attributes["COLOR_1"] = accessor
		case unit.ItemUVCoords:
			attributes[fmt.Sprintf("TEXCOORD_%v", layout.Items[j].Layer)] = accessor
		case unit.ItemBoneIdx:
			attributes[fmt.Sprintf("JOINTS_%v", layout.Items[j].Layer)] = accessor
		case unit.ItemBoneWeight:
			attributes[fmt.Sprintf("WEIGHTS_%v", layout.Items[j].Layer)] = accessor
		}
	}
	return attributes
}

func loadMeshLayoutIndices(gpuR io.ReadSeeker, doc *gltf.Document, layout unit.MeshLayout) (*gltf.Accessor, error) {
	ensurePadding(doc)
	buffer := lastBuffer(doc)
	dataOffset := uint32(len(buffer.Data))
	buffer.ByteLength += layout.IndicesSize
	buffer.Data = append(buffer.Data, make([]byte, layout.IndicesSize)...)
	if _, err := gpuR.Seek(int64(layout.IndexOffset), io.SeekStart); err != nil {
		return nil, err
	}
	var read int
	var err error
	if read, err = gpuR.Read(buffer.Data[dataOffset:]); err != nil {
		return nil, err
	}
	if read != int(layout.IndicesSize) {
		return nil, fmt.Errorf("Read an unexpected amount of data when copying geometry group indices")
	}

	indexStride := layout.IndicesSize / layout.NumIndices
	indexType := gltf.ComponentUshort
	if indexStride == 4 {
		indexType = gltf.ComponentUint
	}

	doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
		Buffer:     uint32(len(doc.Buffers)) - 1,
		ByteLength: layout.IndicesSize,
		ByteOffset: dataOffset,
		Target:     gltf.TargetElementArrayBuffer,
	})

	return &gltf.Accessor{
		BufferView:    gltf.Index(uint32(len(doc.BufferViews)) - 1),
		ByteOffset:    0,
		ComponentType: indexType,
		Type:          gltf.AccessorScalar,
		Count:         layout.NumIndices,
	}, nil
}

func getMaxIndex(buffer *gltf.Buffer, offset, indexCount uint32, componentType gltf.ComponentType) (uint32, error) {
	max := uint32(0)
	if componentType == gltf.ComponentUshort {
		var indexSlice []uint16 = make([]uint16, indexCount)
		_, err := binary.Decode(buffer.Data[offset:], binary.LittleEndian, &indexSlice)
		if err != nil {
			return 0, err
		}
		for _, idx := range indexSlice {
			if uint32(idx) > max {
				max = uint32(idx)
			}
		}
	} else {
		var indexSlice []uint32 = make([]uint32, indexCount)
		_, err := binary.Decode(buffer.Data[offset:], binary.LittleEndian, &indexSlice)
		if err != nil {
			return 0, err
		}
		for _, idx := range indexSlice {
			if idx > max {
				max = idx
			}
		}
	}
	return max, nil
}

func addPositionMinMax(doc *gltf.Document, transformMatrix mgl32.Mat4, min, max mgl32.Vec3, accessor uint32) {
	minTransformed := transformMatrix.Mul4x1(min.Vec4(1)).Vec3()
	maxTransformed := transformMatrix.Mul4x1(max.Vec4(1)).Vec3()
	doc.Accessors[accessor].Min = minTransformed[:]
	doc.Accessors[accessor].Max = maxTransformed[:]
	for k := 0; k < 3; k++ {
		if doc.Accessors[accessor].Min[k] > doc.Accessors[accessor].Max[k] {
			temp := doc.Accessors[accessor].Max[k]
			doc.Accessors[accessor].Max[k] = doc.Accessors[accessor].Min[k]
			doc.Accessors[accessor].Min[k] = temp
		}
	}
}

func transformVertices(buffer *gltf.Buffer, bufferOffset, stride, vertexOffset, vertexCount uint32, transformMatrix mgl32.Mat4) error {
	for vertex := vertexOffset; vertex < vertexCount; vertex += 1 {
		var position mgl32.Vec3
		if _, err := binary.Decode(buffer.Data[vertex*stride+bufferOffset:], binary.LittleEndian, &position); err != nil {
			return err
		}
		position = transformMatrix.Mul4x1(position.Vec4(1)).Vec3()
		if _, err := binary.Encode(buffer.Data[vertex*stride+bufferOffset:], binary.LittleEndian, position); err != nil {
			return err
		}
	}
	return nil
}

func addGroupAttributes(doc *gltf.Document, group unit.MeshGroup, groupLayoutAttributes map[string]*gltf.Accessor, vertexBuffer, maxIndex uint32) (gltf.Attribute, error) {
	groupAttr := make(gltf.Attribute)
	for key, layoutAttrAccessor := range groupLayoutAttributes {
		doc.Accessors = append(doc.Accessors, &gltf.Accessor{
			BufferView:    gltf.Index(vertexBuffer),
			ByteOffset:    layoutAttrAccessor.ByteOffset + doc.BufferViews[vertexBuffer].ByteStride*group.VertexOffset,
			ComponentType: layoutAttrAccessor.ComponentType,
			Type:          layoutAttrAccessor.Type,
			Count:         group.NumVertices,
		})
		accessor := doc.Accessors[len(doc.Accessors)-1]
		if (maxIndex + 1) > accessor.Count {
			accessor.Count = uint32(maxIndex + 1)
		}
		groupAttr[key] = uint32(len(doc.Accessors)) - 1
	}
	return groupAttr, nil
}

// Flips normals by reversing the winding order of vertices
func flipNormals(buffer *gltf.Buffer, componentType gltf.ComponentType, indexCount, bufferOffset uint32) error {
	var indexSlice interface{}
	if componentType == gltf.ComponentUshort {
		indexSlice = make([]uint16, indexCount)
	} else {
		indexSlice = make([]uint32, indexCount)
	}
	if _, err := binary.Decode(buffer.Data[bufferOffset:], binary.LittleEndian, &indexSlice); err != nil {
		return err
	}
	if componentType == gltf.ComponentUshort {
		slices.Reverse(indexSlice.([]uint16))
	} else {
		slices.Reverse(indexSlice.([]uint32))
	}
	if _, err := binary.Encode(buffer.Data[bufferOffset:], binary.LittleEndian, indexSlice); err != nil {
		return err
	}
	return nil
}

func separateUDims(doc *gltf.Document, indexAccessor, texcoordAccessor *gltf.Accessor) (map[uint32][]uint32, error) {
	indexBufferView := doc.BufferViews[*indexAccessor.BufferView]
	buffer := doc.Buffers[indexBufferView.Buffer]
	indexSlice, err := getIndices(buffer, indexBufferView, indexAccessor)
	if err != nil {
		return nil, err
	}

	texcoordOffset := texcoordAccessor.ByteOffset + doc.BufferViews[*texcoordAccessor.BufferView].ByteOffset
	vertexStride := doc.BufferViews[*texcoordAccessor.BufferView].ByteStride
	buffer = doc.Buffers[doc.BufferViews[*texcoordAccessor.BufferView].Buffer]

	UDIMs := make(map[uint32][]uint32)
	for i := uint32(0); i+2 < uint32(len(indexSlice)); i += 3 {
		var uv [2]float32
		vertex := indexSlice[i]
		if _, err := binary.Decode(buffer.Data[vertex*vertexStride+texcoordOffset:], binary.LittleEndian, &uv); err != nil {
			return nil, err
		}

		udim := make([]uint32, 3)
		udim[0] = uint32(uv[0]) | uint32(1-uv[1])<<5
		for j := i + 1; j < i+3; j += 1 {
			vertex := indexSlice[j]
			if _, err := binary.Decode(buffer.Data[vertex*vertexStride+texcoordOffset:], binary.LittleEndian, &uv); err != nil {
				return nil, err
			}
			udim[j-i] = uint32(uv[0]) | uint32(1-uv[1])<<5
		}
		var minUdim uint32
		if udim[0] < udim[1] && udim[0] < udim[2] {
			minUdim = udim[0]
		} else if udim[1] < udim[2] {
			minUdim = udim[1]
		} else {
			minUdim = udim[2]
		}
		UDIMs[minUdim] = append(UDIMs[minUdim], indexSlice[i], indexSlice[i+1], indexSlice[i+2])
	}

	return UDIMs, nil
}

func getIndices(buffer *gltf.Buffer, bufferView *gltf.BufferView, idxAccessor *gltf.Accessor) ([]uint32, error) {
	idxBufferOffset := idxAccessor.ByteOffset + bufferView.ByteOffset
	indices := make([]uint32, idxAccessor.Count)
	if idxAccessor.ComponentType == gltf.ComponentUshort {
		temp := make([]uint16, idxAccessor.Count)
		if _, err := binary.Decode(buffer.Data[idxBufferOffset:], binary.LittleEndian, &temp); err != nil {
			return nil, err
		}
		for i, item := range temp {
			indices[i] = uint32(item)
		}
	} else {
		if _, err := binary.Decode(buffer.Data[idxBufferOffset:], binary.LittleEndian, &indices); err != nil {
			return nil, err
		}
	}
	return indices, nil
}

func ensurePadding(doc *gltf.Document) {
	buffer := lastBuffer(doc)
	padding := getPadding(uint32(len(buffer.Data)))
	buffer.Data = append(buffer.Data, make([]byte, padding)...)
	buffer.ByteLength += padding
}

func lastBuffer(doc *gltf.Document) *gltf.Buffer {
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, new(gltf.Buffer))
	}
	return doc.Buffers[len(doc.Buffers)-1]
}

func getPadding(offset uint32) uint32 {
	padAlign := offset % 4
	if padAlign == 0 {
		return 0
	}
	return 4 - padAlign
}

func addBoundingBox(doc *gltf.Document, name string, meshHeader unit.MeshHeader, info *unit.Info, meshNodes *[]uint32) {
	var indices []uint32 = []uint32{
		0, 1,
		0, 5,
		0, 3,
		1, 4,
		1, 2,
		5, 4,
		5, 6,
		4, 7,
		3, 2,
		3, 6,
		6, 7,
		2, 7,
	}

	vMin := meshHeader.AABB.Min
	vMax := meshHeader.AABB.Max

	var vertices [][3]float32 = [][3]float32{
		vMin,
		{vMax[0], vMin[1], vMin[2]},
		{vMax[0], vMin[1], vMax[2]},
		{vMin[0], vMin[1], vMax[2]},
		{vMax[0], vMax[1], vMin[2]},
		{vMin[0], vMax[1], vMin[2]},
		{vMin[0], vMax[1], vMax[2]},
		vMax,
	}

	boundingBoxTransformIdx := meshHeader.AABBTransformIndex
	for i := range vertices {
		vertices[i] = info.Bones[boundingBoxTransformIdx].Matrix.Mul4x1(mgl32.Vec3(vertices[i]).Vec4(1.0)).Vec3()
		vertices[i][1], vertices[i][2] = vertices[i][2], -vertices[i][1]
	}

	positions := modeler.WritePosition(doc, vertices)
	index := gltf.Index(modeler.WriteIndices(doc, indices))

	primitive := &gltf.Primitive{
		Indices: index,
		Attributes: map[string]uint32{
			gltf.POSITION: positions,
		},
		Mode: gltf.PrimitiveLines,
	}

	doc.Meshes = append(doc.Meshes, &gltf.Mesh{
		Name: name,
		Primitives: []*gltf.Primitive{
			primitive,
		},
	})
	idx := uint32(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: name,
		Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
	})
	*meshNodes = append(*meshNodes, idx)
}

func LoadGLTF(ctx extractor.Context, gpuR io.ReadSeeker, doc *gltf.Document, name stingray.Hash, meshInfos []MeshInfo, bones []stingray.ThinHash, meshLayouts []unit.MeshLayout, unitInfo *unit.Info, meshNodes *[]uint32, materialIndices map[stingray.ThinHash]uint32, parent uint32, skin *uint32) error {
	unitName, contains := ctx.Hashes()[name]
	if !contains {
		unitName = name.String()
	} else {
		items := strings.Split(unitName, "/")
		unitName = items[len(items)-1]
	}

	layoutToVertexBufferView := make(map[uint32]uint32)
	layoutToIndexAccessor := make(map[uint32]*gltf.Accessor)
	layoutAttributes := make(map[uint32]map[string]*gltf.Accessor)

	for i, header := range meshInfos {
		if header.MeshLayoutIndex >= uint32(len(meshLayouts)) {
			return fmt.Errorf("MeshLayoutIndex out of bounds")
		}

		groupNameBoneIdx := -1
		for k, bone := range unitInfo.Bones {
			if bone.NameHash == bones[i] {
				groupNameBoneIdx = k
				break
			}
		}

		var groupName string
		if _, contains := ctx.ThinHashes()[bones[i]]; contains {
			groupName = ctx.ThinHashes()[bones[i]]
		} else {
			groupName = bones[i].String()
		}

		if ctx.Config()["include_lods"] != "true" &&
			(strings.Contains(groupName, "shadow") ||
				strings.Contains(groupName, "_LOD") ||
				strings.Contains(groupName, "cull") ||
				strings.Contains(groupName, "collision")) {
			continue
		}

		var fbxConvertIdx, transformBoneIdxGeo int = -1, -1
		var transformMatrix mgl32.Mat4 = mgl32.Ident4()
		var err error
		_, fbxConvertIdx, transformBoneIdxGeo = getMeshNameFbxConvertAndTransformBone(unitInfo, bones[i])
		vertexBuffer, contains := layoutToVertexBufferView[header.MeshLayoutIndex]
		layout := meshLayouts[header.MeshLayoutIndex]
		if !contains {
			data, accessorInfo, err := convertVertices(gpuR, layout)
			if err != nil {
				return err
			}
			vertexBuffer, err = addMeshLayoutVertexBuffer(doc, data, accessorInfo)
			if err != nil {
				return err
			}
			layoutToVertexBufferView[header.MeshLayoutIndex] = vertexBuffer
			layoutAttributes[header.MeshLayoutIndex] = createAttributes(doc, layout, accessorInfo)
		}

		indexAccessor, contains := layoutToIndexAccessor[header.MeshLayoutIndex]
		if !contains {
			indexAccessor, err = loadMeshLayoutIndices(gpuR, doc, layout)
			if err != nil {
				return err
			}
			layoutToIndexAccessor[header.MeshLayoutIndex] = indexAccessor
		}

		udimPrimitives := make(map[uint32][]*gltf.Primitive)
		nodeName := fmt.Sprintf("%v %v", unitName, groupName)
		var transformed bool = false
		remapped := make(map[uint32]bool)
		var previousPositionAccessor *gltf.Accessor
		for j, group := range header.Groups {
			// Check if this group is a gib or collision mesh, if it is skip it unless include_lods is set
			var materialName string
			if j < len(header.Materials) {
				if _, contains := ctx.ThinHashes()[header.Materials[j]]; contains {
					materialName = ctx.ThinHashes()[header.Materials[j]]
				} else {
					materialName = header.Materials[j].String()
				}
			} else {
				materialName = "unknown"
			}

			if (strings.Contains(materialName, "gibs") || strings.Contains(materialName, "collis")) && ctx.Config()["include_lods"] != "true" {
				continue
			}

			// Add geometry data accessors
			doc.Accessors = append(doc.Accessors, &gltf.Accessor{
				BufferView:    indexAccessor.BufferView,
				ByteOffset:    group.IndexOffset * indexAccessor.ComponentType.ByteSize(),
				ComponentType: indexAccessor.ComponentType,
				Type:          gltf.AccessorScalar,
				Count:         group.NumIndices,
			})
			groupIndices := gltf.Index(uint32(len(doc.Accessors)) - 1)

			offset := doc.BufferViews[*indexAccessor.BufferView].ByteOffset + doc.Accessors[*groupIndices].ByteOffset
			buffer := doc.Buffers[doc.BufferViews[*indexAccessor.BufferView].Buffer]
			maxIndex, err := getMaxIndex(buffer, offset, group.NumIndices, indexAccessor.ComponentType)
			if err != nil {
				return err
			}

			groupAttr, err := addGroupAttributes(doc, group, layoutAttributes[header.MeshLayoutIndex], vertexBuffer, maxIndex)
			if err != nil {
				return err
			}

			// Post process data:
			//   * Reorient in gltf space and align position with group matrix
			//   * Remap raw joints using skeleton maps
			//   * Flip normals if reorientation changed winding order of vertices
			//   * Separate UDIMs
			var transformBoneIdxMesh int32 = -1
			var meshHeader unit.MeshHeader
			for _, meshInfo := range unitInfo.MeshInfos {
				if meshInfo.Header.GroupBoneHash == bones[i] {
					transformBoneIdxMesh = int32(meshInfo.Header.TransformIdx)
					meshHeader = meshInfo.Header
					break
				}
			}
			if transformBoneIdxGeo != -1 {
				transformMatrix = unitInfo.Bones[transformBoneIdxGeo].Matrix
			}
			// If translation, rotation, and scale are identities, use the TransformIndex instead
			if transformMatrix.ApproxEqual(mgl32.Ident4()) && transformBoneIdxMesh != -1 {
				transformMatrix = unitInfo.Bones[transformBoneIdxMesh].Matrix
			}

			if transformBoneIdxGeo == -1 && transformBoneIdxMesh == -1 && groupNameBoneIdx != -1 {
				transformMatrix = unitInfo.Bones[groupNameBoneIdx].Matrix
			}

			// Transform coordinates into glTF ones
			fbxTransformMatrix := mgl32.Mat4([16]float32{
				1, 0, 0, 0,
				0, 0, 1, 0,
				0, -1, 0, 0,
				0, 0, 0, 1,
			})
			if fbxConvertIdx == -1 {
				transformMatrix = fbxTransformMatrix.Mul4(transformMatrix)
			}

			if positionAccessor, contains := groupAttr[gltf.POSITION]; contains {
				addPositionMinMax(doc, transformMatrix, mgl32.Vec3(meshHeader.AABB.Min), mgl32.Vec3(meshHeader.AABB.Max), positionAccessor)

				var vertexOffset uint32 = 0
				if previousPositionAccessor != nil && previousPositionAccessor.Count < doc.Accessors[positionAccessor].Count {
					// Check if there are vertices that still need to be transformed
					vertexOffset = previousPositionAccessor.Count
				}
				if !((transformed && vertexOffset == 0) || transformMatrix.ApproxEqual(mgl32.Ident4())) {
					// Only transform vertices once, and only perform the multiplications if the transform does something
					bufferOffset := doc.Accessors[positionAccessor].ByteOffset + doc.BufferViews[vertexBuffer].ByteOffset
					stride := doc.BufferViews[vertexBuffer].ByteStride
					buffer := doc.Buffers[doc.BufferViews[vertexBuffer].Buffer]
					err := transformVertices(buffer, bufferOffset, stride, vertexOffset, doc.Accessors[positionAccessor].Count, transformMatrix)
					if err != nil {
						return err
					}
					transformed = true
				}
				previousPositionAccessor = doc.Accessors[positionAccessor]
			}

			_, beenRemapped := remapped[*groupIndices]
			if jointsAccessor, contains := groupAttr[gltf.JOINTS_0]; contains && !beenRemapped {
				bufferOffset := doc.Accessors[jointsAccessor].ByteOffset + doc.BufferViews[vertexBuffer].ByteOffset
				stride := doc.BufferViews[vertexBuffer].ByteStride
				buffer := doc.Buffers[doc.BufferViews[vertexBuffer].Buffer]
				skeletonMap := unitInfo.SkeletonMaps[meshHeader.SkeletonMapIdx]
				if j >= len(skeletonMap.RemapList) {
					return fmt.Errorf("%v out of range of components", j)
				}
				remapList := skeletonMap.RemapList[j]
				idxAccessor := doc.Accessors[*groupIndices]
				idxBufferView := doc.BufferViews[*doc.Accessors[*groupIndices].BufferView]
				idxBuffer := doc.Buffers[idxBufferView.Buffer]
				indices, err := getIndices(idxBuffer, idxBufferView, idxAccessor)
				if err != nil {
					return err
				}
				err = remapJoints(buffer, stride, bufferOffset, group.NumVertices, indices, doc.Accessors[jointsAccessor].ComponentType, remapList, skeletonMap.BoneIndices)
				if err != nil {
					return err
				}
				remapped[*groupIndices] = true
			}

			if transformMatrix.Det() < 0 {
				// The transform flipped the winding order of our vertices, so we need to flip the index order to compensate
				bufferOffset := indexAccessor.ByteOffset + doc.BufferViews[*indexAccessor.BufferView].ByteOffset
				buffer := doc.Buffers[doc.BufferViews[*indexAccessor.BufferView].Buffer]
				flipNormals(buffer, indexAccessor.ComponentType, group.NumIndices, bufferOffset)
			}

			udimIndexAccessors := make(map[uint32]uint32)
			if ctx.Config()["join_components"] != "true" {
				texcoordIndex, ok := groupAttr[gltf.TEXCOORD_0]
				// Don't separate udims of LODs or shadow meshes
				if ok && !strings.Contains(groupName, "LOD") && !strings.Contains(groupName, "shadow") {
					texcoordAccessor := doc.Accessors[texcoordIndex]
					groupIndexAccessor := doc.Accessors[*groupIndices]
					var UDIMs map[uint32][]uint32
					if UDIMs, err = separateUDims(doc, groupIndexAccessor, texcoordAccessor); err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						for udim, indices := range UDIMs {
							udimIndexAccessors[udim] = modeler.WriteIndices(doc, indices)
						}
					}
				}
			}
			if ctx.Config()["join_components"] == "true" || (ctx.Config()["include_lods"] == "true" && (strings.Contains(groupName, "LOD") || strings.Contains(groupName, "shadow"))) {
				udimIndexAccessors[0] = *groupIndices
			}

			if ctx.Config()["bounding_boxes"] == "true" {
				addBoundingBox(doc, nodeName+" Bounding Box", meshHeader, unitInfo, meshNodes)
			}

			var material *uint32
			// There are a couple of models where there are fewer materials than meshes, so this is here
			// to prevent us from panicking if we're exporting one of those
			if j < len(header.Materials) {
				materialVal, ok := materialIndices[header.Materials[j]]
				if ok {
					material = &materialVal
				}
			}

			for udim, indexAccessor := range udimIndexAccessors {
				udimPrimitives[udim] = append(udimPrimitives[udim], &gltf.Primitive{
					Attributes: groupAttr,
					Indices:    gltf.Index(indexAccessor),
					Material:   material,
				})
			}
		}
		for udim, primitives := range udimPrimitives {
			doc.Meshes = append(doc.Meshes, &gltf.Mesh{
				Primitives: primitives,
			})

			udimNodeName := nodeName
			if len(udimPrimitives) > 1 {
				udimNodeName = fmt.Sprintf("%v udim %v", nodeName, udim)
			}
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name: udimNodeName,
				Mesh: gltf.Index(uint32(len(doc.Meshes)) - 1),
			})
			node := uint32(len(doc.Nodes)) - 1
			if _, contains := primitives[0].Attributes[gltf.JOINTS_0]; contains {
				doc.Nodes[node].Skin = skin
			} else {
				doc.Nodes[parent].Children = append(doc.Nodes[parent].Children, node)
			}
			*meshNodes = append(*meshNodes, node)
		}
	}

	return nil
}
