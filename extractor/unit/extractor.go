package unit

import (
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/blend_helper"
	"github.com/xypwn/filediver/extractor/geometry"
	extr_material "github.com/xypwn/filediver/extractor/material"
	"github.com/xypwn/filediver/extractor/state_machine"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/bones"
	"github.com/xypwn/filediver/stingray/unit"
	geometrygroup "github.com/xypwn/filediver/stingray/unit/geometry_group"
	"github.com/xypwn/filediver/stingray/unit/material"
)

func LoadBoneMap(ctx *extractor.Context, unitInfo *unit.Info) (*bones.Info, error) {
	if unitInfo.BonesHash.Value == 0x0 {
		return nil, nil
	}
	bonesMainR, err := ctx.Open(stingray.NewFileID(unitInfo.BonesHash, stingray.Sum("bones")), stingray.DataMain)
	if err == stingray.ErrFileNotExist {
		return nil, fmt.Errorf("loadBoneMap: bones file does not exist")
	}
	if err != nil {
		return nil, fmt.Errorf("loadBoneMap: %w", err)
	}

	boneInfo, err := bones.LoadBones(bonesMainR)
	return boneInfo, err
}

// Adds the unit's skeleton to the gltf document
func AddSkeleton(ctx *extractor.Context, doc *gltf.Document, unitInfo *unit.Info, skeletonName stingray.Hash, armorName *string) uint32 {
	boneInfo, err := LoadBoneMap(ctx, unitInfo)
	if err != nil {
		ctx.Warnf("addSkeleton: %v", err)
	}

	if boneInfo != nil {
		for bone, name := range boneInfo.NameMap {
			ctx.ThinHashes()[bone] = name
		}
	}

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
		if otherId, ok := otherIdAny.(uint32); doc.Skins[skin].Name == skeletonName.String() || (contains && ok && skeletonId == otherId) {
			skeleton = doc.Skins[skin].Skeleton
			break
		}
	}

	if skeleton == nil {
		unitName := ctx.LookupHash(skeletonName)
		if strings.Contains(unitName, "/") {
			items := strings.Split(unitName, "/")
			unitName = items[len(items)-1]
		}
		idx := len(doc.Nodes)
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: unitName,
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
		Name:                skeletonName.String(),
		InverseBindMatrices: gltf.Index(inverseBindMatrices),
		Joints:              jointIndices,
		Skeleton:            skeleton,
		Extras:              skeletonTag,
	})

	return uint32(len(doc.Skins) - 1)
}

func AddMaterialVariant(ctx *extractor.Context, mat *material.Material, doc *gltf.Document, imgOpts *extr_material.ImageOptions, materialId stingray.ThinHash, skinOverride datalib.UnitSkinOverride, metadata *datalib.UnitData) (*uint32, error) {
	override, ok := skinOverride.Overrides[materialId]
	if !ok {
		return nil, fmt.Errorf("Override for %v not found?", ctx.LookupThinHash(materialId))
	}
	skinMat := material.Material{
		BaseMaterial: mat.BaseMaterial,
		Textures:     maps.Clone(mat.Textures),
		Settings:     maps.Clone(mat.Settings),
	}
	idx := 0
	if ctx.FileID().Name == stingray.Sum("content/fac_helldivers/hellpod/ammo_rack/ammo_rack") ||
		ctx.FileID().Name == stingray.Sum("content/fac_helldivers/hellpod/flag_rack/flag_rack") {
		idx = len(override) - 1
	}
	if override[idx].MaterialLut.Value != 0 {
		skinMat.Textures[stingray.Sum("material_lut").Thin()] = override[idx].MaterialLut
	}
	if override[idx].PatternLut.Value != 0 {
		skinMat.Textures[stingray.Sum("pattern_lut").Thin()] = override[idx].PatternLut
	}
	if override[idx].DecalSheet.Value != 0 {
		skinMat.Textures[stingray.Sum("decal_sheet").Thin()] = override[idx].DecalSheet
	}
	if override[idx].PatternMasksArray.Value != 0 {
		skinMat.Textures[stingray.Sum("pattern_masks_array").Thin()] = override[idx].PatternMasksArray
	}

	skinMatIdx, err := extr_material.AddMaterial(ctx, &skinMat, doc, imgOpts, skinOverride.ID.String()+" "+ctx.LookupThinHash(materialId), metadata)
	if err != nil {
		return nil, err
	}

	return gltf.Index(skinMatIdx), nil
}

func AddMaterials(ctx *extractor.Context, doc *gltf.Document, imgOpts *extr_material.ImageOptions, unitInfo *unit.Info, metadata *datalib.UnitData) ([]geometry.MaterialVariantMap, error) {
	materialVariants := make([]geometry.MaterialVariantMap, 0)
	namesToVariantIdx := make(map[string]uint32)
	for id, resID := range unitInfo.Materials {
		matR, err := ctx.Open(stingray.NewFileID(resID, stingray.Sum("material")), stingray.DataMain)
		if err == stingray.ErrFileNotExist {
			return nil, fmt.Errorf("referenced material resource %v doesn't exist", resID)
		}
		if err != nil {
			return nil, err
		}
		mat, err := material.Load(matR)
		if err != nil {
			return nil, err
		}

		resPath := ctx.LookupHash(resID)
		if strings.Contains(resPath, "/") {
			split := strings.Split(resPath, "/")
			resPath = strings.Join(split[len(split)-2:], "/")
		}
		matIdx, err := extr_material.AddMaterial(ctx, mat, doc, imgOpts, ctx.LookupThinHash(id)+" "+resPath, metadata)
		if err != nil {
			return nil, err
		}

		// Using a slice+map combo to maintain variant order
		// Otherwise different pieces would have different variant indices
		// and combining them in a single file would result in random variants
		// selected every time the skin was changed
		if _, ok := namesToVariantIdx["default"]; !ok {
			namesToVariantIdx["default"] = uint32(len(materialVariants))
			materialVariants = append(materialVariants, geometry.MaterialVariantMap{
				Name:                "default",
				MaterialHashToIndex: make(map[stingray.ThinHash]uint32),
			})
		}
		materialVariants[namesToVariantIdx["default"]].MaterialHashToIndex[id] = matIdx

		// Handle variants
		var skinOverrides []datalib.UnitSkinOverride = make([]datalib.UnitSkinOverride, 0)
		for _, skinOverrideGroup := range ctx.SkinOverrideGroups() {
			if !skinOverrideGroup.HasMaterial(id) {
				continue
			}
			skinOverrides = skinOverrideGroup.Skins
		}
		for _, skinOverride := range skinOverrides {
			skinName := cases.Title(language.English).String(skinOverride.Name)

			if _, ok := skinOverride.Overrides[id]; !ok {
				continue
			}

			skinMatIdx, err := AddMaterialVariant(ctx, mat, doc, imgOpts, id, skinOverride, metadata)
			if err != nil {
				// Some materials don't get overriden, that's fine
				continue
			}

			if _, ok := namesToVariantIdx[skinName]; !ok {
				namesToVariantIdx[skinName] = uint32(len(materialVariants))
				materialVariants = append(materialVariants, geometry.MaterialVariantMap{
					Name:                skinName,
					MaterialHashToIndex: make(map[stingray.ThinHash]uint32),
				})
			}
			materialVariants[namesToVariantIdx[skinName]].MaterialHashToIndex[id] = *skinMatIdx
		}
	}
	return materialVariants, nil
}

func AddPrefabMetadata(ctx *extractor.Context, doc *gltf.Document, filename stingray.Hash, parent *uint32, skin *uint32, meshNodes []uint32, armorSetName *string) {
	if armorSetName != nil {
		extras := map[string]any{"armorSet": *armorSetName}
		for _, node := range meshNodes {
			doc.Nodes[node].Extras = extras
		}
	}
	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}
	prefabMetadata := make(map[string]any)
	prefabMetadata["parent"] = *parent
	if skin != nil {
		prefabMetadata["skin"] = *skin
	}
	prefabMetadata["objects"] = meshNodes
	extras[filename.String()] = prefabMetadata
	doc.Extras = extras
}

func findBoneRecursive(doc *gltf.Document, currentNode uint32, boneName string) *uint32 {
	if doc.Nodes[currentNode].Name == boneName {
		return gltf.Index(currentNode)
	}
	for _, child := range doc.Nodes[currentNode].Children {
		res := findBoneRecursive(doc, child, boneName)
		if res != nil {
			return res
		}
	}
	return nil
}

func findBone(ctx *extractor.Context, doc *gltf.Document, parent *uint32, namehash stingray.ThinHash) *uint32 {
	boneName, ok := ctx.ThinHashes()[namehash]
	if !ok {
		boneName = fmt.Sprintf("Bone_%08x", namehash.Value)
	}

	return findBoneRecursive(doc, *parent, boneName)
}

func AddLights(ctx *extractor.Context, doc *gltf.Document, unitInfo *unit.Info, parent *uint32) {
	if len(unitInfo.Lights) == 0 {
		return
	}

	if !slices.Contains(doc.ExtensionsUsed, "KHR_lights_punctual") {
		doc.ExtensionsUsed = append(doc.ExtensionsUsed, "KHR_lights_punctual")
	}

	gltfLights := make([]map[string]any, 0)
	for _, light := range unitInfo.Lights {
		if light.BoneIndex >= uint32(len(unitInfo.Bones)) {
			ctx.Warnf("light %v has bone index exceeding length of unit bones list", ctx.LookupThinHash(light.NameHash))
			return
		}

		boneGltfIdx := findBone(ctx, doc, parent, unitInfo.Bones[light.BoneIndex].NameHash)
		if boneGltfIdx == nil {
			ctx.Warnf("could not find bone %v to attach light %v", ctx.LookupThinHash(unitInfo.Bones[light.BoneIndex].NameHash), ctx.LookupThinHash(light.NameHash))
			return
		}

		lightGltfIdx := len(gltfLights)
		lightNodeIdx := len(doc.Nodes)
		quat := mgl32.QuatRotate(mgl32.DegToRad(90), mgl32.Vec3{1.0, 0.0, 0.0})
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name:     ctx.LookupThinHash(light.NameHash),
			Rotation: quat.V.Vec4(quat.W),
			Extensions: map[string]any{
				"KHR_lights_punctual": map[string]any{
					"light": lightGltfIdx,
				},
			},
		})

		doc.Nodes[*boneGltfIdx].Children = append(doc.Nodes[*boneGltfIdx].Children, uint32(lightNodeIdx))

		maxColor := float32(math.Max(math.Max(float64(light.Color[0]), float64(light.Color[1])), float64(light.Color[2])))
		intensity := light.Intensity
		if maxColor > 1.0 {
			light.Color[0] /= maxColor
			light.Color[1] /= maxColor
			light.Color[2] /= maxColor
			intensity *= maxColor
		}
		gltfLight := map[string]any{
			"name":      ctx.LookupThinHash(light.NameHash),
			"type":      light.Type.ToGLTF(),
			"color":     light.Color,
			"intensity": intensity * 100.0,
		}
		if light.Type != unit.LightDirectional {
			gltfLight["range"] = light.FalloffEnd
		}
		if light.Type == unit.LightSpot {
			spot := map[string]any{
				"innerConeAngle": light.SpotInnerAngle,
				"outerConeAngle": light.SpotOuterAngle,
			}
			gltfLight["spot"] = spot
		}
		gltfLights = append(gltfLights, gltfLight)
	}
	if doc.Extensions == nil {
		doc.Extensions = make(gltf.Extensions)
	}
	doc.Extensions["KHR_lights_punctual"] = map[string]any{
		"lights": gltfLights,
	}
}

func ConvertOpts(ctx *extractor.Context, imgOpts *extr_material.ImageOptions, gltfDoc *gltf.Document) error {
	fMain, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	var fGPU io.ReadSeeker
	if ctx.Exists(ctx.FileID(), stingray.DataGPU) {
		fGPU, err = ctx.Open(ctx.FileID(), stingray.DataGPU)
		if err != nil {
			return err
		}
	}

	return ConvertBuffer(fMain, fGPU, ctx.FileID().Name, ctx, imgOpts, gltfDoc)
}

func ConvertBuffer(fMain, fGPU io.ReadSeeker, filename stingray.Hash, ctx *extractor.Context, imgOpts *extr_material.ImageOptions, gltfDoc *gltf.Document) error {
	cfg := ctx.Config()

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
	var metadata *datalib.UnitData = nil
	var armorSetName *string = nil
	if armorSet, ok := ctx.GuessFileArmorSet(ctx.FileID()); ok {
		armorSetName = &armorSet.Name
		if _, contains := armorSet.UnitMetadata[filename]; contains {
			value := armorSet.UnitMetadata[filename]
			metadata = &value
		}
	}

	// Load materials
	materialIdxs, err := AddMaterials(ctx, doc, imgOpts, unitInfo, metadata)
	if err != nil {
		return err
	}

	bonesEnabled := !cfg.Model.NoBones
	animationsEnabled := cfg.Model.EnableAnimations

	var skin *uint32 = nil
	var parent *uint32 = nil
	if bonesEnabled && len(unitInfo.Bones) > 2 {
		skin = gltf.Index(AddSkeleton(ctx, doc, unitInfo, filename, armorSetName))
		parent = doc.Skins[*skin].Skeleton
		if animationsEnabled {
			state_machine.AddAnimationSet(ctx, doc, unitInfo)
		}
		AddLights(ctx, doc, unitInfo, parent)
	} else {
		parent = gltf.Index(uint32(len(doc.Nodes)))
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: filename.String(),
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

		err := geometry.LoadGLTF(ctx, fGPU, doc, filename, meshInfos, unitInfo.GroupBones, unitInfo.MeshLayouts, unitInfo, &meshNodes, materialIdxs, *parent, skin)
		if err != nil {
			return err
		}
	}

	AddPrefabMetadata(ctx, doc, filename, parent, skin, meshNodes, armorSetName)

	formatIsBlend := cfg.Model.Format == "blend"
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
		outPath, err := ctx.AllocateFile(".blend")
		if err != nil {
			return err
		}
		err = blend_helper.ExportBlend(doc, outPath, ctx.Runner())
		if err != nil {
			return err
		}
	}
	return nil
}

func loadGeometryGroupMeshes(ctx *extractor.Context, doc *gltf.Document, unitInfo *unit.Info, meshNodes *[]uint32, materialIndices []geometry.MaterialVariantMap, parent uint32, skin *uint32) error {
	geoID := stingray.NewFileID(unitInfo.GeometryGroup, stingray.Sum("geometry_group"))
	f, err := ctx.Open(geoID, stingray.DataMain)
	if err == stingray.ErrFileNotExist {
		return fmt.Errorf("%v.geometry_group does not exist", unitInfo.GeometryGroup.String())
	}
	if err != nil {
		return err
	}

	geoGroup, err := geometrygroup.LoadGeometryGroup(f)
	if err != nil {
		return err
	}

	geoInfo, ok := geoGroup.MeshInfos[ctx.FileID().Name]
	unitName, contains := ctx.Hashes()[ctx.FileID().Name]
	if !contains {
		unitName = ctx.FileID().Name.String()
	}
	if !ok {
		return fmt.Errorf("%v.geometry_group does not contain %v.unit", unitInfo.GeometryGroup.String(), unitName)
	}

	gpuR, err := ctx.Open(geoID, stingray.DataGPU)
	if err != nil {
		return err
	}

	meshInfos := make([]geometry.MeshInfo, 0)
	for _, header := range geoInfo.MeshHeaders {
		meshInfos = append(meshInfos, geometry.MeshInfo{
			Groups:          header.Groups,
			Materials:       header.Materials,
			MeshLayoutIndex: header.MeshLayoutIndex,
		})
	}

	return geometry.LoadGLTF(ctx, gpuR, doc, ctx.FileID().Name, meshInfos, geoInfo.Bones, geoGroup.MeshLayouts, unitInfo, meshNodes, materialIndices, parent, skin)
}

func Convert(currDoc *gltf.Document) func(ctx *extractor.Context) error {
	return func(ctx *extractor.Context) error {
		opts, err := extr_material.GetImageOpts(ctx)
		if err != nil {
			return err
		}
		return ConvertOpts(ctx, opts, currDoc)
	}
}
