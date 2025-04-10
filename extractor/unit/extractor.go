package unit

import (
	"fmt"
	"image/png"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

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
		err := loadGeometryGroupMeshes(ctx, gltfDoc, unitInfo, &meshNodes)
		if err != nil {
			return err
		}
		// FIXME
		return nil
	} else {
		var meshes map[uint32]unit.Mesh
		allMeshes := make([]uint32, unitInfo.NumMeshes)
		for i := range allMeshes {
			allMeshes[i] = uint32(i)
		}
		meshes, err = unit.LoadMeshes(fGPU, unitInfo, allMeshes)
		if err != nil {
			return err
		}

		if ctx.Config()["bounding_boxes"] == "true" {
			for i, mesh := range meshes {
				addBoundingBox(doc, fmt.Sprintf("Mesh %d Bounding Box", i), mesh, unitInfo, armorSetName)
			}
		}

		for meshDisplayNumber, meshID := range meshesToLoad {
			if meshID >= unitInfo.NumMeshes {
				panic("meshID out of bounds")
			}

			mesh := meshes[meshID]
			if len(mesh.UVCoords) == 0 {
				continue
			}

			var groupBoneIdx *uint32 = nil
			var meshNameBoneIdx *uint32 = nil
			// Extract some bone indices used even if skeletons are disabled
			transformBoneIdx := mesh.Info.Header.TransformIdx
			parentIdx := -1
			fbxConvertIdx := -1
			gameMeshHash := stingray.Sum64([]byte("game_mesh")).Thin()
			fbxConvertHash := stingray.Sum64([]byte("FbxAxisSystem_ConvertNode")).Thin()
			for boneIdx, bone := range unitInfo.Bones {
				if bone.NameHash == gameMeshHash {
					parentIdx = int(boneIdx)
				}
				if bone.NameHash == fbxConvertHash {
					fbxConvertIdx = int(boneIdx)
				}
				if bone.ParentIndex == uint32(parentIdx) && bone.NameHash == mesh.Info.Header.GroupBoneHash {
					transformBoneIdx = uint32(boneIdx)
				}
				if bone.ParentIndex == uint32(fbxConvertIdx) {
					meshNameBoneIdx = gltf.Index(uint32(boneIdx))
				}
			}

			// Apply vertex transform
			if bonesEnabled {
				groupBoneIdx = &unitInfo.Bones[transformBoneIdx].ParentIndex
				transformMatrix := unitInfo.Bones[transformBoneIdx].Matrix
				// If translation, rotation, and scale are identities, use the TransformIndex instead
				if transformMatrix.ApproxEqual(mgl32.Ident4()) {
					transformMatrix = unitInfo.Bones[mesh.Info.Header.TransformIdx].Matrix
				}
				if !transformMatrix.ApproxEqual(mgl32.Ident4()) {
					// Apply transformations
					for i := range mesh.Positions {
						p := mgl32.Vec3(mesh.Positions[i])
						p = transformMatrix.Mul4x1(p.Vec4(1)).Vec3()
						mesh.Positions[i] = p
					}
					if transformMatrix.Det() < 0 {
						// If the matrix flips the vertices, we need to flip the normals
						// Reversing the indices accomplishes this task
						for i := range mesh.Indices {
							slices.Reverse(mesh.Indices[i])
						}
					}
				}
			}

			// Transform coordinates into glTF ones
			fbxTransformMatrix := mgl32.Ident4()
			if fbxConvertIdx != -1 {
				fbxTransformMatrix = unitInfo.Bones[fbxConvertIdx].Matrix
			}
			if fbxTransformMatrix.ApproxEqual(mgl32.Ident4()) {
				// tbh should probably invert the fbx transform and combine with this, but... this should be fine
				for i := range mesh.Positions {
					p := mesh.Positions[i]
					p[1], p[2] = p[2], -p[1]
					mesh.Positions[i] = p

					n := mesh.Normals[i]
					n[1], n[2] = n[2], -n[1]
					mesh.Normals[i] = n
				}
			}

			var weights *uint32 = nil
			var joints *uint32 = nil

			if bonesEnabled && len(mesh.BoneWeights) > 0 {
				if len(unitInfo.SkeletonMaps) > 0 && mesh.Info.Header.SkeletonMapIdx >= 0 {
					if err := remapMeshBones(&mesh, unitInfo.SkeletonMaps[mesh.Info.Header.SkeletonMapIdx]); err != nil {
						return err
					}
				}
				weights = gltf.Index(modeler.WriteWeights(doc, mesh.BoneWeights))
				joints = gltf.Index(modeler.WriteJoints(doc, mesh.BoneIndices[0]))
			}

			positions := modeler.WritePosition(doc, mesh.Positions)
			normals := modeler.WriteAccessor(doc, gltf.TargetArrayBuffer, mesh.Normals)
			var texCoords []uint32 = make([]uint32, len(mesh.UVCoords))
			for i := range mesh.UVCoords {
				texCoords[i] = modeler.WriteTextureCoord(doc, mesh.UVCoords[i])
			}
			var lodName string
			if len(meshesToLoad) > 1 {
				lodName = fmt.Sprintf("LOD %v", meshDisplayNumber)
			}

			var groupName string
			var meshName string
			if groupBoneIdx != nil && skin != nil && meshNameBoneIdx != nil {
				meshName = strings.TrimPrefix(doc.Nodes[doc.Skins[*skin].Joints[*meshNameBoneIdx]].Name, "Bone_")
				groupName = strings.TrimPrefix(doc.Nodes[doc.Skins[*skin].Joints[*groupBoneIdx]].Name, "grp_") + " "
			}

			for i := range mesh.Indices {
				var componentName string = fmt.Sprintf("%smesh %v", groupName, i)
				if metadata != nil {
					componentName = fmt.Sprintf("%s_%s_%s mesh %v", metadata.Slot.String(), metadata.Type.String(), metadata.BodyType.String(), i)
				} else {
					if lodName != "" {
						componentName = lodName + " " + componentName
					}
					if meshName != "" {
						componentName = meshName + " " + componentName
					}
				}

				var material *uint32
				if len(mesh.Info.Materials) > int(mesh.Info.Groups[i].GroupIdx) {
					if idx, ok := materialIdxs[mesh.Info.Materials[int(mesh.Info.Groups[i].GroupIdx)]]; ok {
						material = gltf.Index(idx)
					}
				}

				if ctx.Config()["join_components"] == "true" {
					indices := gltf.Index(modeler.WriteIndices(doc, mesh.Indices[i]))
					nodeIdx := writeMesh(ctx, doc, componentName, indices, positions, normals, texCoords, material, joints, weights, skin, armorSetName)

					doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, nodeIdx)
					if skin == nil {
						doc.Nodes[*parent].Children = append(doc.Nodes[*parent].Children, nodeIdx)
					}
					meshNodes = append(meshNodes, nodeIdx)
				} else {
					var components map[uint32][]uint32 = make(map[uint32][]uint32)
					make_key := func(uv [2]float32) uint32 {
						key := (uint32(uv[0]) & 0x1f) | (uint32(uv[1]) << 5)
						if uv[1] < 0 {
							key = (uint32(uv[0]) & 0x1f) | (uint32(-uv[1]+1) << 5)
						}
						return key
					}
					for j := range mesh.Indices[i] {
						if j%3 != 0 || mesh.Indices[i][j] >= uint32(len(mesh.UVCoords[0])) || mesh.Indices[i][j+1] >= uint32(len(mesh.UVCoords[0])) || mesh.Indices[i][j+2] >= uint32(len(mesh.UVCoords[0])) {
							continue
						}
						// Need to use all three sets of uvs since there are cases where
						// 2 of the vertices will be at X.00 and the third could be at (X-1).99
						// or at X.01, and we need to use the minimum of the three to avoid
						// stray triangles
						key0 := make_key(mesh.UVCoords[0][mesh.Indices[i][j]])
						key1 := make_key(mesh.UVCoords[0][mesh.Indices[i][j+1]])
						key2 := make_key(mesh.UVCoords[0][mesh.Indices[i][j+2]])
						key := key0
						if key1 < key {
							key = key1
						}
						if key2 < key {
							key = key2
						}
						components[key] = append(components[key], mesh.Indices[i][j], mesh.Indices[i][j+1], mesh.Indices[i][j+2])
					}

					componentNum := 0
					for _, componentIndices := range components {
						indices := gltf.Index(modeler.WriteIndices(doc, componentIndices))
						nodeIdx := writeMesh(ctx, doc, fmt.Sprintf("%s cmp %v", componentName, componentNum), indices, positions, normals, texCoords, material, joints, weights, skin, armorSetName)

						doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, nodeIdx)
						if skin == nil {
							doc.Nodes[*parent].Children = append(doc.Nodes[*parent].Children, nodeIdx)
						}
						meshNodes = append(meshNodes, nodeIdx)
						componentNum += 1
					}
				}
			}
		}
	}

	// Add some metadata so prefab loading can find our parent node easily
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

func loadGeometryGroupMeshes(ctx extractor.Context, gltfDoc *gltf.Document, unitInfo *unit.Info, meshNodes *[]uint32) error {
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
	if !ok {
		name, ok := ctx.Hashes()[ctx.File().ID().Name]
		if !ok {
			name = ctx.File().ID().Name.String()
		}
		return fmt.Errorf("%v.geometry_group does not contain %v.unit", unitInfo.GeometryGroup.String(), name)
	}

	fmt.Printf("\nUnit materials:\n")
	for thin, material := range unitInfo.Materials {
		thinName, ok := ctx.ThinHashes()[thin]
		if !ok {
			thinName = thin.String()
		}
		fullName, ok := ctx.Hashes()[material]
		if !ok {
			fullName = material.String()
		}
		fmt.Printf("    %v: %v\n", thinName, fullName)
	}

	fmt.Printf("\nUnit bones:\n")
	for i, bone := range unitInfo.Bones {
		name, ok := ctx.ThinHashes()[bone.NameHash]
		if !ok {
			name = bone.NameHash.String()
		}
		fmt.Printf("    %v: %v\n", i, name)
	}

	fmt.Printf("\nGeometryGroup bones:\n")
	for _, thin := range geoInfo.Bones {
		name, ok := ctx.ThinHashes()[thin]
		if !ok {
			name = thin.String()
		}
		fmt.Printf("    %v\n", name)
	}

	fmt.Printf("\nGeometryGroup materials:\n")
	for i, header := range geoInfo.MeshHeaders {
		fmt.Printf("    Mesh %v:\n", i)
		for _, group := range header.Groups {
			fmt.Printf("      %v vertices\n", group.NumVertices)
			fmt.Printf("      %v indices\n", group.NumIndices)
		}
		for _, thin := range header.Materials {
			name, ok := ctx.ThinHashes()[thin]
			if !ok {
				name = thin.String()
			}
			fmt.Printf("      %v\n", name)
		}
	}

	return nil
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
