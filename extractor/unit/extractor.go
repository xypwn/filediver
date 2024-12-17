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
	"github.com/xypwn/filediver/stingray/unit"
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
func addSkeleton(ctx extractor.Context, doc *gltf.Document, unitInfo *unit.Info, boneInfo *bones.BoneInfo) uint32 {
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
		skeletonId := node.Extras.(map[string]uint32)["skeletonId"]
		nodeNames[node.Name+fmt.Sprintf("%08x", skeletonId)] = uint32(i)
	}

	skeletonId := unitInfo.Bones[2].NameHash.Value
	var skeletonTag map[string]uint32 = make(map[string]uint32)
	skeletonTag["skeletonId"] = skeletonId

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
	if !slices.Contains(doc.Scenes[0].Nodes, rootNodeIndex) {
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, rootNodeIndex)
	}

	doc.Skins = append(doc.Skins, &gltf.Skin{
		Name:                ctx.File().ID().Name.String(),
		InverseBindMatrices: gltf.Index(inverseBindMatrices),
		Joints:              jointIndices,
	})

	return uint32(len(doc.Skins) - 1)
}

func writeMesh(ctx extractor.Context, doc *gltf.Document, componentName string, indices *uint32, positions uint32, texCoords []uint32, material *uint32, joints *uint32, weights *uint32, skin *uint32) uint32 {
	primitive := &gltf.Primitive{
		Indices: indices,
		Attributes: map[string]uint32{
			gltf.POSITION: positions,
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
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: componentName,
		Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
		Skin: skin,
	})
	return uint32(len(doc.Nodes) - 1)
}

func addBoundingBox(doc *gltf.Document, name string, mesh unit.Mesh, info *unit.Info) {
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
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: name,
		Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
	})
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

		matIdx, err := extr_material.AddMaterial(ctx, mat, doc, imgOpts, resID.String())
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
		if len(unitInfo.LODGroups) > 0 {
			entries := unitInfo.LODGroups[0].Entries
			highestDetailIdx := -1
			for i := range entries {
				if highestDetailIdx == -1 || entries[i].Detail.Max > entries[highestDetailIdx].Detail.Max {
					highestDetailIdx = i
				}
			}
			if highestDetailIdx != -1 {
				meshesToLoad = entries[highestDetailIdx].Indices[:1]
			}
		} else {
			fmt.Println("\nAdding LODs anyway since no lodgroups in unitInfo?")
			for i := uint32(0); i < unitInfo.NumMeshes; i++ {
				meshesToLoad = append(meshesToLoad, i)
			}
		}
	}

	bonesEnabled := ctx.Config()["no_bones"] != "true"

	var meshes map[uint32]unit.Mesh

	if ctx.Config()["bounding_boxes"] == "true" {
		allMeshes := make([]uint32, unitInfo.NumMeshes)
		for i := range allMeshes {
			allMeshes[i] = uint32(i)
		}
		meshes, err = unit.LoadMeshes(fGPU, unitInfo, allMeshes)
		if err != nil {
			return err
		}

		for i, mesh := range meshes {
			addBoundingBox(doc, fmt.Sprintf("Mesh %d AABB", i), mesh, unitInfo)
		}
	} else {
		// Load meshes
		meshes, err = unit.LoadMeshes(fGPU, unitInfo, meshesToLoad)
		if err != nil {
			return err
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
		if bonesEnabled {
			// Apply vertex transform
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
				if bone.ParentIndex == uint32(parentIdx) {
					transformBoneIdx = uint32(boneIdx)
				}
				if bone.ParentIndex == uint32(fbxConvertIdx) {
					meshNameBoneIdx = gltf.Index(uint32(boneIdx))
				}
			}
			groupBoneIdx = &transformBoneIdx
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
		for i := range mesh.Positions {
			p := mesh.Positions[i]
			p[1], p[2] = p[2], -p[1]
			mesh.Positions[i] = p
		}

		var skin *uint32 = nil
		var weights *uint32 = nil
		var joints *uint32 = nil

		if bonesEnabled && len(mesh.BoneWeights) > 0 {
			if len(unitInfo.SkeletonMaps) > 0 && mesh.Info.Header.SkeletonMapIdx >= 0 {
				if err := remapMeshBones(&mesh, unitInfo.SkeletonMaps[mesh.Info.Header.SkeletonMapIdx]); err != nil {
					return err
				}
			}
			skin = gltf.Index(addSkeleton(ctx, doc, unitInfo, boneInfo))
			weights = gltf.Index(modeler.WriteWeights(doc, mesh.BoneWeights))
			joints = gltf.Index(modeler.WriteJoints(doc, mesh.BoneIndices[0]))
		}

		positions := modeler.WritePosition(doc, mesh.Positions)
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
			if lodName != "" {
				componentName = lodName + " " + componentName
			}
			if meshName != "" {
				componentName = meshName + " " + componentName
			}

			var material *uint32
			if len(mesh.Info.Materials) > int(mesh.Info.Groups[i].GroupIdx) {
				if idx, ok := materialIdxs[mesh.Info.Materials[int(mesh.Info.Groups[i].GroupIdx)]]; ok {
					material = gltf.Index(idx)
				}
			}

			if ctx.Config()["join_components"] == "true" {
				indices := gltf.Index(modeler.WriteIndices(doc, mesh.Indices[i]))
				nodeIdx := writeMesh(ctx, doc, componentName, indices, positions, texCoords, material, joints, weights, skin)

				doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, nodeIdx)
			} else {
				var components map[uint32][]uint32 = make(map[uint32][]uint32)
				for j := range mesh.Indices[i] {
					uv := mesh.UVCoords[0][mesh.Indices[i][j]]
					key := (uint32(uv[0]) & 0x1f) | (uint32(uv[1]) << 5)
					if uv[1] < 0 {
						key = (uint32(uv[0]) & 0x1f) | (uint32(-uv[1]+1) << 5)
					}
					components[key] = append(components[key], mesh.Indices[i][j])
				}

				componentNum := 0
				for _, componentIndices := range components {
					indices := gltf.Index(modeler.WriteIndices(doc, componentIndices))
					nodeIdx := writeMesh(ctx, doc, fmt.Sprintf("%s cmp %v", componentName, componentNum), indices, positions, texCoords, material, joints, weights, skin)

					doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, nodeIdx)
					componentNum += 1
				}
			}
		}
	}

	if gltfDoc == nil && (ctx.Config()["format"] == "glb" || ctx.Config()["format"] == "blend_glb") {
		out, err := ctx.CreateFile(".glb")
		if err != nil {
			return err
		}
		enc := gltf.NewEncoder(out)
		if err := enc.Encode(doc); err != nil {
			return err
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
