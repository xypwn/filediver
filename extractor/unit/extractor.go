package unit

import (
	"fmt"
	"image/png"
	"io"
	"strconv"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/geometry"
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

// Adds the unit's skeleton to the gltf document
func AddSkeleton(ctx extractor.Context, doc *gltf.Document, unitInfo *unit.Info, armorName *string) uint32 {
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
		// if boneInfo != nil {
		// 	name, exists := boneInfo.NameMap[bone.NameHash]
		// 	if exists {
		// 		boneName = name
		// 	}
		// }
		name, exists := ctx.ThinHashes()[bone.NameHash]
		if exists {
			boneName = name
		}
		var boneIdx uint32
		var contains bool = false
		var parentIndex uint32
		if boneIdx, contains = nodeNames[boneName+fmt.Sprintf("%08x", skeletonId)]; !contains {
			parentBone := unitInfo.Bones[bone.ParentIndex]
			parentName, contains := ctx.ThinHashes()[parentBone.NameHash]
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

func AddMaterials(ctx extractor.Context, doc *gltf.Document, imgOpts *extr_material.ImageOptions, unitInfo *unit.Info, metadata *dlbin.UnitData) (map[stingray.ThinHash]uint32, error) {
	materialIdxs := make(map[stingray.ThinHash]uint32)
	for id, resID := range unitInfo.Materials {
		matRes, exists := ctx.GetResource(resID, stingray.Sum64([]byte("material")))
		if !exists || !matRes.Exists(stingray.DataMain) {
			return nil, fmt.Errorf("referenced material resource %v doesn't exist", resID)
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
			return nil, err
		}

		matIdx, err := extr_material.AddMaterial(ctx, mat, doc, imgOpts, resID.String(), metadata)
		if err != nil {
			return nil, err
		}
		materialIdxs[id] = matIdx
	}
	return materialIdxs, nil
}

func addPrefabMetadata(ctx extractor.Context, doc *gltf.Document, parent *uint32, skin *uint32, meshNodes []uint32, armorSetName *string) {
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
	materialIdxs, err := AddMaterials(ctx, doc, imgOpts, unitInfo, metadata)
	if err != nil {
		return err
	}

	bonesEnabled := ctx.Config()["no_bones"] != "true"

	var skin *uint32 = nil
	var parent *uint32 = nil
	if bonesEnabled && len(unitInfo.Bones) > 2 {
		skin = gltf.Index(AddSkeleton(ctx, doc, unitInfo, armorSetName))
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
		meshInfos := make([]geometry.MeshInfo, 0)
		for _, info := range unitInfo.MeshInfos {
			meshInfos = append(meshInfos, geometry.MeshInfo{
				Groups:          info.Groups,
				Materials:       info.Materials,
				MeshLayoutIndex: uint32(info.Header.LayoutIdx),
			})
		}

		err := geometry.LoadGLTF(ctx, fGPU, doc, meshInfos, unitInfo.GroupBones, unitInfo.MeshLayouts, unitInfo, &meshNodes, materialIdxs, *parent, skin)
		if err != nil {
			return err
		}
	}

	addPrefabMetadata(ctx, doc, parent, skin, meshNodes, armorSetName)

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

	meshInfos := make([]geometry.MeshInfo, 0)
	for _, header := range geoInfo.MeshHeaders {
		meshInfos = append(meshInfos, geometry.MeshInfo{
			Groups:          header.Groups,
			Materials:       header.Materials,
			MeshLayoutIndex: header.MeshLayoutIndex,
		})
	}

	return geometry.LoadGLTF(ctx, gpuR, doc, meshInfos, geoInfo.Bones, geoGroup.MeshLayouts, unitInfo, meshNodes, materialIndices, parent, skin)
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
