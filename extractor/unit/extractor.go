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

func convertFloat16Slice(gpuR io.ReadSeeker, data []byte, tmpArr []uint16) ([]byte, error) {
	var err error
	if err = binary.Read(gpuR, binary.LittleEndian, &tmpArr); err != nil {
		return nil, err
	}
	for _, tmp := range tmpArr {
		data, err = binary.Append(data, binary.LittleEndian, float16.Frombits(tmp).Float32())
		if err != nil {
			return nil, err
		}
	}
	return data, nil
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
				tmpArr := make([]uint16, 1)
				var err error
				data, err = convertFloat16Slice(gpuR, data, tmpArr)
				if err != nil {
					return nil, nil, err
				}
				if vertex == 0 {
					accessorStructure = append(accessorStructure, AccessorInfo{
						AccessorType:  gltf.AccessorScalar,
						ComponentType: gltf.ComponentFloat,
						Size:          4,
					})
				}
			case unit.FormatVec2F16:
				tmpArr := make([]uint16, 2)
				var err error
				data, err = convertFloat16Slice(gpuR, data, tmpArr)
				if err != nil {
					return nil, nil, err
				}
				if vertex == 0 {
					accessorStructure = append(accessorStructure, AccessorInfo{
						AccessorType:  gltf.AccessorVec2,
						ComponentType: gltf.ComponentFloat,
						Size:          8,
					})
				}
			case unit.FormatVec3F16:
				tmpArr := make([]uint16, 3)
				var err error
				data, err = convertFloat16Slice(gpuR, data, tmpArr)
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
			case unit.FormatVec4F16:
				tmpArr := make([]uint16, 4)
				var err error
				data, err = convertFloat16Slice(gpuR, data, tmpArr)
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
			default:
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

func remapJoint[E ~[]I, I uint8 | uint32](mapping unit.SkeletonMap, idxs E, component int) error {
	for k := 0; k < 4; k++ {
		if component >= len(mapping.RemapList) {
			return fmt.Errorf("%v out of range of components", component)
		}
		remapList := mapping.RemapList[component]
		if uint32(idxs[k]) >= uint32(len(remapList)) {
			continue
		}
		remapIndex := remapList[idxs[k]]
		idxs[k] = I(mapping.BoneIndices[remapIndex])
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

		gameMeshHash := stingray.Sum64([]byte("game_mesh")).Thin()
		parentIdx := unitInfo.Bones[groupNameBoneIdx].ParentIndex

		if ctx.Config()["include_lods"] != "true" &&
			unitInfo.Bones[parentIdx].NameHash != gameMeshHash ||
			strings.Contains(groupName, "_LOD") {
			continue
		}

		var fbxConvertIdx, transformBoneIdxGeo int = -1, -1
		var transformMatrix mgl32.Mat4 = mgl32.Ident4()
		_, fbxConvertIdx, transformBoneIdxGeo = getMeshNameFbxConvertAndTransformBone(unitInfo, bones[i])
		vertexBuffer, contains := layoutToVertexBufferView[header.MeshLayoutIndex]
		layout := meshLayouts[header.MeshLayoutIndex]
		if !contains {
			ensurePadding(doc)
			buffer := lastBuffer(doc)
			offset := uint32(len(buffer.Data))

			data, accessorInfo, err := convertVertices(gpuR, layout)
			if err != nil {
				return err
			}
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

			vertexBuffer = uint32(len(doc.BufferViews) - 1)
			layoutToVertexBufferView[header.MeshLayoutIndex] = vertexBuffer

			layoutAttributes[header.MeshLayoutIndex] = make(map[string]*gltf.Accessor)

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
				switch layout.Items[j].Type {
				case unit.ItemPosition:
					layoutAttributes[header.MeshLayoutIndex][gltf.POSITION] = accessor
				case unit.ItemNormal:
					layoutAttributes[header.MeshLayoutIndex][gltf.COLOR_0] = accessor
				case unit.ItemUVCoords:
					layoutAttributes[header.MeshLayoutIndex][fmt.Sprintf("TEXCOORD_%v", layout.Items[j].Layer)] = accessor
				case unit.ItemBoneIdx:
					layoutAttributes[header.MeshLayoutIndex][fmt.Sprintf("JOINTS_%v", layout.Items[j].Layer)] = accessor
				case unit.ItemBoneWeight:
					layoutAttributes[header.MeshLayoutIndex][fmt.Sprintf("WEIGHTS_%v", layout.Items[j].Layer)] = accessor
				}
			}
		}

		indexAccessor, contains := layoutToIndexAccessor[header.MeshLayoutIndex]

		if !contains {
			ensurePadding(doc)
			buffer := lastBuffer(doc)
			dataOffset := uint32(len(buffer.Data))
			buffer.ByteLength += layout.IndicesSize
			buffer.Data = append(buffer.Data, make([]byte, layout.IndicesSize)...)
			if _, err := gpuR.Seek(int64(layout.IndexOffset), io.SeekStart); err != nil {
				return err
			}
			var read int
			var err error
			if read, err = gpuR.Read(buffer.Data[dataOffset:]); err != nil {
				return err
			}
			if read != int(layout.IndicesSize) {
				return fmt.Errorf("Read an unexpected amount of data when copying geometry group indices")
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

			indexAccessor = &gltf.Accessor{
				BufferView:    gltf.Index(uint32(len(doc.BufferViews)) - 1),
				ByteOffset:    0,
				ComponentType: indexType,
				Type:          gltf.AccessorScalar,
				Count:         layout.NumIndices,
			}
			layoutToIndexAccessor[header.MeshLayoutIndex] = indexAccessor
		}

		primitives := make([]*gltf.Primitive, 0, 1)
		nodeName := fmt.Sprintf("%v %v", unitName, groupName)
		for j, group := range header.Groups {
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

			vertexBuffer := layoutToVertexBufferView[header.MeshLayoutIndex]
			groupAttr := make(gltf.Attribute)
			for key, accessor := range layoutAttributes[header.MeshLayoutIndex] {
				doc.Accessors = append(doc.Accessors, &gltf.Accessor{
					BufferView:    gltf.Index(vertexBuffer),
					ByteOffset:    accessor.ByteOffset + doc.BufferViews[vertexBuffer].ByteStride*group.VertexOffset,
					ComponentType: accessor.ComponentType,
					Type:          accessor.Type,
					Count:         group.NumVertices,
				})
				groupAttr[key] = uint32(len(doc.Accessors)) - 1
				if j != 0 {
					// Only transform the positions and remap the joints once
					continue
				}
				if key == gltf.POSITION {
					minTransformed := transformMatrix.Mul4x1(mgl32.Vec3(meshHeader.AABB.Min).Vec4(1)).Vec3()
					maxTransformed := transformMatrix.Mul4x1(mgl32.Vec3(meshHeader.AABB.Max).Vec4(1)).Vec3()
					doc.Accessors[groupAttr[key]].Min = minTransformed[:]
					doc.Accessors[groupAttr[key]].Max = maxTransformed[:]
					for k := 0; k < 3; k++ {
						if doc.Accessors[groupAttr[key]].Min[k] > doc.Accessors[groupAttr[key]].Max[k] {
							temp := doc.Accessors[groupAttr[key]].Max[k]
							doc.Accessors[groupAttr[key]].Max[k] = doc.Accessors[groupAttr[key]].Min[k]
							doc.Accessors[groupAttr[key]].Min[k] = temp
						}
					}
					if !transformMatrix.ApproxEqual(mgl32.Ident4()) {
						bufferOffset := doc.Accessors[groupAttr[key]].ByteOffset + doc.BufferViews[vertexBuffer].ByteOffset
						stride := doc.BufferViews[vertexBuffer].ByteStride
						buffer := doc.Buffers[doc.BufferViews[vertexBuffer].Buffer]
						for vertex := uint32(0); vertex < group.NumVertices; vertex += 1 {
							var position mgl32.Vec3
							if _, err := binary.Decode(buffer.Data[vertex*stride+bufferOffset:], binary.LittleEndian, &position); err != nil {
								return err
							}
							position = transformMatrix.Mul4x1(position.Vec4(1)).Vec3()
							if _, err := binary.Encode(buffer.Data[vertex*stride+bufferOffset:], binary.LittleEndian, position); err != nil {
								return err
							}
						}
					}
				} else if strings.Contains(key, "JOINTS") {
					bufferOffset := doc.Accessors[groupAttr[key]].ByteOffset + doc.BufferViews[vertexBuffer].ByteOffset
					stride := doc.BufferViews[vertexBuffer].ByteStride
					buffer := doc.Buffers[doc.BufferViews[vertexBuffer].Buffer]
					skeletonMap := unitInfo.SkeletonMaps[meshHeader.SkeletonMapIdx]
					for vertex := uint32(0); vertex < group.NumVertices; vertex += 1 {
						compType := doc.Accessors[groupAttr[key]].ComponentType
						if compType == gltf.ComponentUbyte {
							boneIndices := make([]uint8, 4)
							if _, err := binary.Decode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, &boneIndices); err != nil {
								return err
							}
							if err := remapJoint(skeletonMap, boneIndices, j); err != nil {
								continue
							}
							if _, err := binary.Encode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, boneIndices); err != nil {
								return err
							}
						} else {
							boneIndices := make([]uint32, 4)
							if _, err := binary.Decode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, &boneIndices); err != nil {
								return err
							}
							if err := remapJoint(skeletonMap, boneIndices, j); err != nil {
								continue
							}
							if _, err := binary.Encode(buffer.Data[stride*vertex+bufferOffset:], binary.LittleEndian, boneIndices); err != nil {
								return err
							}
						}
					}
				}
			}
			doc.Accessors = append(doc.Accessors, &gltf.Accessor{
				BufferView:    indexAccessor.BufferView,
				ByteOffset:    group.IndexOffset * indexAccessor.ComponentType.ByteSize(),
				ComponentType: indexAccessor.ComponentType,
				Type:          gltf.AccessorScalar,
				Count:         group.NumIndices,
			})

			if transformMatrix.Det() < 0 {
				bufferOffset := indexAccessor.ByteOffset + doc.BufferViews[*indexAccessor.BufferView].ByteOffset
				buffer := doc.Buffers[doc.BufferViews[*indexAccessor.BufferView].Buffer]
				var indexSlice interface{}
				if indexAccessor.ComponentType == gltf.ComponentUshort {
					indexSlice = make([]uint16, group.NumIndices)
				} else {
					indexSlice = make([]uint32, group.NumIndices)
				}
				if _, err := binary.Decode(buffer.Data[bufferOffset:], binary.LittleEndian, &indexSlice); err != nil {
					return err
				}
				if indexAccessor.ComponentType == gltf.ComponentUshort {
					slices.Reverse(indexSlice.([]uint16))
				} else {
					slices.Reverse(indexSlice.([]uint32))
				}
				if _, err := binary.Encode(buffer.Data[bufferOffset:], binary.LittleEndian, indexSlice); err != nil {
					return err
				}
			}

			var material *uint32
			materialVal, ok := materialIndices[header.Materials[j]]
			if ok {
				material = &materialVal
			}

			primitives = append(primitives, &gltf.Primitive{
				Attributes: groupAttr,
				Indices:    gltf.Index(uint32(len(doc.Accessors)) - 1),
				Material:   material,
			})
		}
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{
			Primitives: primitives,
		})

		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: nodeName,
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
