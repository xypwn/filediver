package unit

import (
	"encoding/binary"
	"fmt"
	"image/png"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/x448/float16"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/xypwn/filediver/extractor"
	extr_material "github.com/xypwn/filediver/extractor/material"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/bones"
	dlbin "github.com/xypwn/filediver/stingray/dl_bin"
	"github.com/xypwn/filediver/stingray/unit"
	geometrygroup "github.com/xypwn/filediver/stingray/unit/geometry_group"
	"github.com/xypwn/filediver/stingray/unit/material"
)

func loadBoneMap(ctx extractor.Context) (*bones.BoneInfo, error) {
	bonesId := ctx.File().ID()
	bonesId.Type = stingray.Sum64([]byte("bones"))
	bonesFile, exists := ctx.GetResource(bonesId.Name, bonesId.Type)
	if !exists {
		return nil, fmt.Errorf("loadBoneMap: bones file does not exist")
	}
	bonesMain, err := bonesFile.Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return nil, fmt.Errorf("loadBoneMap: bones file does not have a main component")
	}

	boneInfo, err := bones.LoadBones(bonesMain)
	return boneInfo, err
}

func remapMeshBones(mesh *unit.Mesh, mapping unit.SkeletonMap) error {
	var remappedIndices map[uint32]bool = make(map[uint32]bool)
	for component := range mesh.Indices {
		for _, index := range mesh.Indices[component] {
			if _, contains := remappedIndices[index]; contains {
				continue
			}
			layer := 0
			if len(mesh.BoneIndices) > 0 {
				for j := range mesh.BoneIndices[layer][index] {
					boneIndex := mesh.BoneIndices[layer][index][j]
					remapList := mapping.RemapList[component]
					if int(boneIndex) >= len(remapList) {
						if layer > 0 {
							break
						}
						return fmt.Errorf("vertex %v has boneIndex exceeding remapList length", index)
					}
					remapIndex := remapList[boneIndex]
					mesh.BoneIndices[layer][index][j] = uint8(mapping.BoneIndices[remapIndex])
					remappedIndices[index] = true
				}
			}
		}
	}
	return nil
}

// Adds the unit's skeleton to the gltf document
func addSkeleton(ctx extractor.Context, doc *gltf.Document, unitInfo *unit.Info, boneInfo *bones.BoneInfo, armorName *string) uint32 {
	var matrices [][4][4]float32 = make([][4][4]float32, len(unitInfo.JointTransformMatrices))
	gltfConversionMatrix := mgl32.HomogRotate3DX(mgl32.DegToRad(-90.0)).Mul4(mgl32.HomogRotate3DZ(mgl32.DegToRad(0)))
	for i := range matrices {
		jtm := unitInfo.JointTransformMatrices[i]
		bindMatrix := mgl32.Mat4FromRows(jtm[0], jtm[1], jtm[2], jtm[3]).Transpose()
		bindMatrix = gltfConversionMatrix.Mul4(bindMatrix)
		row0, row1, row2, row3 := bindMatrix.Inv().Rows()
		matrices[i] = [4][4]float32{row0, row1, row2, row3}
		unitInfo.Bones[i].Matrix = bindMatrix
	}

	unitInfo.Bones[0].RecursiveCalcLocalTransforms(&unitInfo.Bones)

	var nodeNames map[string]uint32 = make(map[string]uint32)
	for i, node := range doc.Nodes {
		if node.Extras == nil {
			continue
		}
		extras, ok := node.Extras.(map[string]any)
		if !ok {
			continue
		}
		skeletonIdAny, ok := extras["skeletonId"]
		if !ok {
			continue
		}
		skeletonId, ok := skeletonIdAny.(uint32)
		if !ok {
			continue
		}
		nodeNames[node.Name+fmt.Sprintf("%08x", skeletonId)] = uint32(i)
	}

	skeletonId := unitInfo.Bones[2].NameHash.Value
	var skeletonTag map[string]any = make(map[string]any)
	skeletonTag["skeletonId"] = skeletonId
	if armorName != nil {
		skeletonTag["armorSet"] = *armorName
	}

	inverseBindMatrices := modeler.WriteAccessor(doc, gltf.TargetNone, matrices)
	jointIndices := make([]uint32, 0)
	boneBaseIndex := uint32(len(doc.Nodes))
	rootNodeIndex := boneBaseIndex
	for i, bone := range unitInfo.Bones {
		quat := mgl32.Mat4ToQuat(bone.Transform.Rotation.Mat4())
		boneName := fmt.Sprintf("Bone_%08x", bone.NameHash.Value)
		if boneInfo != nil {
			name, exists := boneInfo.NameMap[bone.NameHash]
			if exists {
				boneName = name
			}
		}
		var boneIdx uint32
		var contains bool = false
		var parentIndex uint32
		if boneIdx, contains = nodeNames[boneName+fmt.Sprintf("%08x", skeletonId)]; !contains {
			parentBone := unitInfo.Bones[bone.ParentIndex]
			parentName, contains := boneInfo.NameMap[parentBone.NameHash]
			if !contains {
				parentName = fmt.Sprintf("Bone_%08x", parentBone.NameHash.Value)
			}
			parentIndex, contains = nodeNames[parentName+fmt.Sprintf("%08x", skeletonId)]
			if !contains {
				parentIndex = bone.ParentIndex + boneBaseIndex
			}

			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name:        boneName,
				Rotation:    quat.V.Vec4(quat.W),
				Translation: bone.Transform.Translation,
				Scale:       bone.Transform.Scale,
				Extras:      skeletonTag,
			})
			boneIdx = uint32(len(doc.Nodes) - 1)

			if parentIndex != boneIdx && parentIndex < boneIdx {
				doc.Nodes[parentIndex].Children = append(doc.Nodes[parentIndex].Children, boneIdx)
			}
		} else {
			if i == 0 {
				rootNodeIndex = boneIdx
			}
			boneBaseIndex -= 1
		}
		jointIndices = append(jointIndices, boneIdx)
	}

	var skeleton *uint32 = nil
	for skin := range doc.Skins {
		extras, ok := doc.Skins[skin].Extras.(map[string]any)
		if !ok {
			extras = make(map[string]any)
		}
		otherIdAny, contains := extras["skeletonId"]
		if otherId, ok := otherIdAny.(uint32); doc.Skins[skin].Name == ctx.File().ID().Name.String() || (contains && ok && skeletonId == otherId) {
			skeleton = doc.Skins[skin].Skeleton
			break
		}
	}

	if skeleton == nil {
		idx := len(doc.Nodes)
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: ctx.File().ID().Name.String(),
			Children: []uint32{
				rootNodeIndex,
			},
		})
		if armorName != nil {
			extras := map[string]any{"armorSet": *armorName}
			doc.Nodes[idx].Extras = extras
		}
		skeleton = gltf.Index(uint32(idx))
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, *skeleton)
	}

	doc.Skins = append(doc.Skins, &gltf.Skin{
		Name:                ctx.File().ID().Name.String(),
		InverseBindMatrices: gltf.Index(inverseBindMatrices),
		Joints:              jointIndices,
		Skeleton:            skeleton,
		Extras:              skeletonTag,
	})

	return uint32(len(doc.Skins) - 1)
}

func writeMesh(ctx extractor.Context, doc *gltf.Document, componentName string, indices *uint32, positions uint32, normals uint32, texCoords []uint32, material *uint32, joints *uint32, weights *uint32, skin *uint32, armorSet *string) uint32 {
	primitive := &gltf.Primitive{
		Indices: indices,
		Attributes: map[string]uint32{
			gltf.POSITION: positions,
			gltf.NORMAL:   normals,
		},
		Material: material,
	}

	if joints != nil {
		primitive.Attributes[gltf.JOINTS_0] = *joints
		primitive.Attributes[gltf.WEIGHTS_0] = *weights
	}

	for j := range texCoords {
		primitive.Attributes[fmt.Sprintf("TEXCOORD_%v", j)] = texCoords[j]
	}

	doc.Meshes = append(doc.Meshes, &gltf.Mesh{
		Name: ctx.File().ID().Name.String(),
		Primitives: []*gltf.Primitive{
			primitive,
		},
	})
	idx := len(doc.Nodes)
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: componentName,
		Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
		Skin: skin,
	})
	if armorSet != nil {
		extras := map[string]any{"armorSet": *armorSet}
		doc.Nodes[idx].Extras = extras
	}
	return uint32(idx)
}

func addBoundingBox(doc *gltf.Document, name string, mesh unit.Mesh, info *unit.Info, armorSet *string) {
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

	vMin := mesh.Info.Header.AABB.Min
	vMax := mesh.Info.Header.AABB.Max

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

	boundingBoxTransformIdx := mesh.Info.Header.AABBTransformIndex
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
	idx := len(doc.Nodes)
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: name,
		Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
	})
	if armorSet != nil {
		extras := map[string]any{"armorSet": *armorSet}
		doc.Nodes[idx].Extras = extras
	}
}

func ConvertOpts(ctx extractor.Context, imgOpts *extr_material.ImageOptions, gltfDoc *gltf.Document) error {
	fMain, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer fMain.Close()
	var fGPU io.ReadSeekCloser
	if ctx.File().Exists(stingray.DataGPU) {
		fGPU, err = ctx.File().Open(ctx.Ctx(), stingray.DataGPU)
		if err != nil {
			return err
		}
		defer fGPU.Close()
	}

	boneInfo, _ := loadBoneMap(ctx)
	if boneInfo == nil {
		boneInfo, _ = bones.PlayerBones()
	} else {
		playerBones, _ := bones.PlayerBones()
		for k, v := range playerBones.NameMap {
			if _, contains := boneInfo.NameMap[k]; !contains {
				boneInfo.NameMap[k] = v
			}
		}
	}
	for k, v := range ctx.ThinHashes() {
		if _, contains := boneInfo.NameMap[k]; !contains {
			boneInfo.NameMap[k] = v
		}
	}

	unitInfo, err := unit.LoadInfo(fMain)
	if err != nil {
		return err
	}

	var doc *gltf.Document = gltfDoc
	if doc == nil {
		doc = gltf.NewDocument()
		doc.Asset.Generator = "https://github.com/xypwn/filediver"
		doc.Samplers = append(doc.Samplers, &gltf.Sampler{
			MagFilter: gltf.MagLinear,
			MinFilter: gltf.MinLinear,
			WrapS:     gltf.WrapRepeat,
			WrapT:     gltf.WrapRepeat,
		})
	}

	// Get metadata
	var metadata *dlbin.UnitData = nil
	var armorSetName *string = nil
	var triadID *stingray.Hash = nil
	for _, triad := range ctx.TriadIDs() {
		for _, fileTriad := range ctx.File().TriadIDs() {
			if fileTriad == triad {
				triadID = &triad
			}
		}
	}
	if triadID != nil {
		armorSet, ok := ctx.ArmorSets()[*triadID]
		if ok {
			armorSetName = &armorSet.Name
			if _, contains := armorSet.UnitMetadata[ctx.File().ID().Name]; contains {
				value := armorSet.UnitMetadata[ctx.File().ID().Name]
				metadata = &value
			}
		}
	}

	// Load materials
	materialIdxs := make(map[stingray.ThinHash]uint32)
	for id, resID := range unitInfo.Materials {
		matRes, exists := ctx.GetResource(resID, stingray.Sum64([]byte("material")))
		if !exists || !matRes.Exists(stingray.DataMain) {
			return fmt.Errorf("referenced material resource %v doesn't exist", resID)
		}
		mat, err := func() (*material.Material, error) {
			f, err := matRes.Open(ctx.Ctx(), stingray.DataMain)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			return material.Load(f)
		}()
		if err != nil {
			return err
		}

		matIdx, err := extr_material.AddMaterial(ctx, mat, doc, imgOpts, resID.String(), metadata)
		if err != nil {
			return err
		}
		materialIdxs[id] = matIdx
	}

	// Determine which meshes to convert
	var meshesToLoad []uint32
	if ctx.Config()["include_lods"] == "true" {
		for i := uint32(0); i < unitInfo.NumMeshes; i++ {
			meshesToLoad = append(meshesToLoad, i)
		}
	} else {
		if len(unitInfo.MeshInfos) > 0 {
			highestDetailIdx := -1
			highestDetailCount := -1
			for i, info := range unitInfo.MeshInfos {
				for _, group := range info.Groups {
					if int(group.NumIndices) > highestDetailCount && info.Header.MeshType != unit.MeshTypeUnknown00 {
						highestDetailIdx = i
						highestDetailCount = int(group.NumIndices)
					}
				}
			}
			if highestDetailIdx != -1 {
				meshesToLoad = []uint32{uint32(highestDetailIdx)}
			}
		} else {
			ctx.Warnf("Adding LODs anyway since no lodgroups in unitInfo?")
			for i := uint32(0); i < unitInfo.NumMeshes; i++ {
				meshesToLoad = append(meshesToLoad, i)
			}
		}
	}

	bonesEnabled := ctx.Config()["no_bones"] != "true"

	var skin *uint32 = nil
	var parent *uint32 = nil
	if bonesEnabled && len(unitInfo.Bones) > 2 {
		skin = gltf.Index(addSkeleton(ctx, doc, unitInfo, boneInfo, armorSetName))
		parent = doc.Skins[*skin].Skeleton
	} else {
		parent = gltf.Index(uint32(len(doc.Nodes)))
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: ctx.File().ID().Name.String(),
		})
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, *parent)
		if armorSetName != nil {
			extras := map[string]any{"armorSet": *armorSetName}
			doc.Nodes[*parent].Extras = extras
		}
	}

	var meshNodes []uint32 = make([]uint32, 0)

	if unitInfo.GeometryGroup.Value != 0x0 {
		err := loadGeometryGroupMeshes(ctx, doc, unitInfo, &meshNodes, materialIdxs, *parent, skin)
		if err != nil {
			return err
		}
	} else {
		meshInfos := make([]MeshInfo, 0)
		for _, info := range unitInfo.MeshInfos {
			meshInfos = append(meshInfos, MeshInfo{
				Groups:          info.Groups,
				Materials:       info.Materials,
				MeshLayoutIndex: uint32(info.Header.LayoutIdx),
			})
		}

		err := loadMeshLayouts(ctx, fGPU, doc, meshInfos, unitInfo.GroupBones, unitInfo.MeshLayouts, unitInfo, &meshNodes, materialIdxs, *parent, skin)
		if err != nil {
			return err
		}
	}

	// Add some metadata so prefab loading can find our parent node easily
	if armorSetName != nil {
		extras := map[string]any{"armorSet": *armorSetName}
		for _, node := range meshNodes {
			doc.Nodes[node].Extras = extras
		}
	}
	extras, ok := doc.Extras.(map[string]map[string]interface{})
	if !ok {
		extras = make(map[string]map[string]interface{})
	}
	extras[ctx.File().ID().Name.String()] = make(map[string]interface{})
	extras[ctx.File().ID().Name.String()]["parent"] = *parent
	if skin != nil {
		extras[ctx.File().ID().Name.String()]["skin"] = *skin
	}
	extras[ctx.File().ID().Name.String()]["objects"] = meshNodes
	doc.Extras = extras

	formatIsBlend := ctx.Config()["format"] == "blend" && ctx.Runner().Has("hd2_accurate_blender_importer")
	if gltfDoc == nil && !formatIsBlend {
		out, err := ctx.CreateFile(".glb")
		if err != nil {
			return err
		}
		enc := gltf.NewEncoder(out)
		if err := enc.Encode(doc); err != nil {
			return err
		}
	} else if gltfDoc == nil && formatIsBlend {
		path, err := ctx.(interface{ OutPath() (string, error) }).OutPath()
		if err != nil {
			return err
		}
		err = extractor.ExportBlend(doc, path, ctx.Runner())
		if err != nil {
			return err
		}
	}
	return nil
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

type AccessorInfo struct {
	gltf.AccessorType
	gltf.ComponentType
	Size uint32
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
	for _, vertex := range indices {
		if vertex >= vertexCount {
			continue
		}
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

func loadGeometryGroupMeshes(ctx extractor.Context, doc *gltf.Document, unitInfo *unit.Info, meshNodes *[]uint32, materialIndices map[stingray.ThinHash]uint32, parent uint32, skin *uint32) error {
	geoRes, exists := ctx.GetResource(unitInfo.GeometryGroup, stingray.Sum64([]byte("geometry_group")))
	if !exists {
		return fmt.Errorf("%v.geometry_group does not exist", unitInfo.GeometryGroup.String())
	}
	f, err := geoRes.Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer f.Close()

	geoGroup, err := geometrygroup.LoadGeometryGroup(f)
	if err != nil {
		return err
	}

	geoInfo, ok := geoGroup.MeshInfos[ctx.File().ID().Name]
	unitName, contains := ctx.Hashes()[ctx.File().ID().Name]
	if !contains {
		unitName = ctx.File().ID().Name.String()
	}
	if !ok {
		return fmt.Errorf("%v.geometry_group does not contain %v.unit", unitInfo.GeometryGroup.String(), unitName)
	}

	gpuR, err := geoRes.Open(ctx.Ctx(), stingray.DataGPU)
	if err != nil {
		return err
	}
	defer gpuR.Close()

	meshInfos := make([]MeshInfo, 0)
	for _, header := range geoInfo.MeshHeaders {
		meshInfos = append(meshInfos, MeshInfo{
			header.Groups,
			header.Materials,
			header.MeshLayoutIndex,
		})
	}

	return loadMeshLayouts(ctx, gpuR, doc, meshInfos, geoInfo.Bones, geoGroup.MeshLayouts, unitInfo, meshNodes, materialIndices, parent, skin)
}

type MeshInfo struct {
	Groups          []unit.MeshGroup
	Materials       []stingray.ThinHash
	MeshLayoutIndex uint32
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
	indexSlice := make([]uint32, indexAccessor.Count)
	indexOffset := indexAccessor.ByteOffset + doc.BufferViews[*indexAccessor.BufferView].ByteOffset
	buffer := doc.Buffers[doc.BufferViews[*indexAccessor.BufferView].Buffer]
	if indexAccessor.ComponentType == gltf.ComponentUshort {
		slice := make([]uint16, indexAccessor.Count)
		if _, err := binary.Decode(buffer.Data[indexOffset:], binary.LittleEndian, &slice); err != nil {
			return nil, err
		}
		for i, item := range slice {
			indexSlice[i] = uint32(item)
		}
	} else {
		if _, err := binary.Decode(buffer.Data[indexOffset:], binary.LittleEndian, &indexSlice); err != nil {
			return nil, err
		}
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

func loadMeshLayouts(ctx extractor.Context, gpuR io.ReadSeeker, doc *gltf.Document, meshInfos []MeshInfo, bones []stingray.ThinHash, meshLayouts []unit.MeshLayout, unitInfo *unit.Info, meshNodes *[]uint32, materialIndices map[stingray.ThinHash]uint32, parent uint32, skin *uint32) error {
	unitName, contains := ctx.Hashes()[ctx.File().ID().Name]
	if !contains {
		unitName = ctx.File().ID().Name.String()
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
				strings.Contains(groupName, "culling")) {
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
			// Check if this group is a gib, if it is skip it unless include_lods is set
			var materialName string
			if _, contains := ctx.ThinHashes()[header.Materials[j]]; contains {
				materialName = ctx.ThinHashes()[header.Materials[j]]
			} else {
				materialName = header.Materials[j].String()
			}

			if strings.Contains(materialName, "gibs") && ctx.Config()["include_lods"] != "true" {
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

			var material *uint32
			materialVal, ok := materialIndices[header.Materials[j]]
			if ok {
				material = &materialVal
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

func getImgOpts(ctx extractor.Context) (*extr_material.ImageOptions, error) {
	var opts extr_material.ImageOptions
	if v, ok := ctx.Config()["image_jpeg"]; ok && v == "true" {
		opts.Jpeg = true
	}
	if v, ok := ctx.Config()["jpeg_quality"]; ok {
		quality, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		opts.JpegQuality = quality
	}
	if v, ok := ctx.Config()["png_compression"]; ok {
		switch v {
		case "default":
			opts.PngCompression = png.DefaultCompression
		case "none":
			opts.PngCompression = png.NoCompression
		case "fast":
			opts.PngCompression = png.BestSpeed
		case "best":
			opts.PngCompression = png.BestCompression
		}
	}
	return &opts, nil
}

func Convert(currDoc *gltf.Document) func(ctx extractor.Context) error {
	return func(ctx extractor.Context) error {
		opts, err := getImgOpts(ctx)
		if err != nil {
			return err
		}
		return ConvertOpts(ctx, opts, currDoc)
	}
}
