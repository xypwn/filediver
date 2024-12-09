package unit

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"strconv"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/bones"
	"github.com/xypwn/filediver/stingray/unit"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

type ImageOptions struct {
	Jpeg           bool                 // PNG if false, JPEG if true
	JpegQuality    int                  // Quality if Jpeg == true; interval = [1;100]; 0 for default quality
	PngCompression png.CompressionLevel // Compression if Jpeg == false
	Raw            bool                 // Save raw dds in addition to png/jpg using gltf MSFT DDS extension if true
}

// Adds back in the truncated Z component of a normal map.
func postProcessReconstructNormalZ(img image.Image) error {
	calcZ := func(x, y float64) float64 {
		return math.Sqrt(-x*x - y*y + 1)
	}
	switch img := img.(type) {
	case *image.NRGBA:
		for iY := img.Rect.Min.Y; iY < img.Rect.Max.Y; iY++ {
			for iX := img.Rect.Min.X; iX < img.Rect.Max.X; iX++ {
				idx := img.PixOffset(iX, iY)
				r, g := img.Pix[idx], img.Pix[idx+1]
				x, y := (float64(r)/127.5)-1, (float64(g)/127.5)-1
				z := calcZ(x, y)
				img.Pix[idx+2] = uint8(math.Round((z + 1) * 127.5))
			}
		}
		return nil
	default:
		return errors.New("postProcessReconstructNormalZ: unsupported image type")
	}
}

// Attempts to completely remove the influence of the alpha channel,
// giving the whole image an opacity of 1.
func postProcessToOpaque(img image.Image) error {
	switch img := img.(type) {
	case *image.NRGBA:
		for iY := img.Rect.Min.Y; iY < img.Rect.Max.Y; iY++ {
			for iX := img.Rect.Min.X; iX < img.Rect.Max.X; iX++ {
				idx := img.PixOffset(iX, iY)
				img.Pix[idx+3] = 255
			}
		}
		return nil
	default:
		return errors.New("postProcessToOpaque: unsupported image type")
	}
}

// Adds a texture to doc. Returns new texture ID if err != nil.
// postProcess optionally applies image post-processing.
func writeTexture(ctx extractor.Context, doc *gltf.Document, id stingray.Hash, postProcess func(image.Image) error, imgOpts *ImageOptions) (uint32, error) {
	file, exists := ctx.GetResource(id, stingray.Sum64([]byte("texture")))
	if !exists || !file.Exists(stingray.DataMain) {
		return 0, fmt.Errorf("texture resource %v doesn't exist", id)
	}

	tex, err := texture.Decode(ctx.Ctx(), file, false)
	if err != nil {
		return 0, err
	}

	if len(tex.Images) > 1 {
		tex = dds.StackLayers(tex)
	}

	if postProcess != nil {
		if err := postProcess(tex.Image); err != nil {
			return 0, err
		}
	}
	var encData bytes.Buffer
	var mimeType string
	if imgOpts != nil && imgOpts.Jpeg {
		quality := jpeg.DefaultQuality
		if imgOpts.JpegQuality != 0 {
			quality = imgOpts.JpegQuality
		}
		if err := jpeg.Encode(&encData, tex, &jpeg.Options{Quality: quality}); err != nil {
			return 0, err
		}
		mimeType = "image/jpeg"
	} else {
		compression := png.DefaultCompression
		if imgOpts != nil {
			compression = imgOpts.PngCompression
		}
		if err := (&png.Encoder{
			CompressionLevel: compression,
		}).Encode(&encData, tex); err != nil {
			return 0, err
		}
		mimeType = "image/png"
	}
	imgIdx, err := modeler.WriteImage(doc, id.String(), mimeType, &encData)
	if err != nil {
		return 0, err
	}
	doc.Textures = append(doc.Textures, &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(imgIdx),
	})
	texIdx := uint32(len(doc.Textures) - 1)
	if imgOpts != nil && imgOpts.Raw {
		reader, err := file.OpenMulti(ctx.Ctx(), stingray.DataMain, stingray.DataStream, stingray.DataGPU)
		if err != nil {
			fmt.Printf("writeTexture: Failed to open multireader for raw dds\n")
			return texIdx, nil
		}
		defer reader.Close()
		if _, err := texture.DecodeInfo(reader); err != nil {
			fmt.Printf("writeTexture: Failed to decode stingray texture info\n")
			return texIdx, nil
		}
		mimeType = "image/vnd-ms.dds"
		imgIdx, err = modeler.WriteImage(doc, id.String()+".dds", mimeType, reader)
		if err != nil {
			fmt.Printf("writeTexture: Failed to write dds image to document\n")
			return texIdx, nil
		}
		doc.Textures[texIdx].Extensions = make(gltf.Extensions)
		msftTextureDDS := make(map[string]uint32)
		msftTextureDDS["source"] = imgIdx
		doc.Textures[texIdx].Extensions["MSFT_texture_dds"] = msftTextureDDS
		contained := false
		for _, ext := range doc.ExtensionsUsed {
			if ext == "MSFT_texture_dds" {
				contained = true
				break
			}
		}
		if !contained {
			doc.ExtensionsUsed = append(doc.ExtensionsUsed, "MSFT_texture_dds")
		}
	}
	return texIdx, nil
}

type TextureUsage uint32

const (
	AlbedoIridescence                     TextureUsage = 0xff2c91cc //stingray.Sum64([]byte("albedo_iridescence")).Thin()
	Albedo                                TextureUsage = 0xac652e43
	Normal                                TextureUsage = 0xcaed6cd6
	MRA                                   TextureUsage = 0x756f6fa6
	EmissiveColor                         TextureUsage = 0xc985395a
	BaseData                              TextureUsage = 0xc2eb8d6e
	MaterialLUT                           TextureUsage = 0x7e662968
	PatternLUT                            TextureUsage = 0x81d4c49d
	CompositeArray                        TextureUsage = 0xa17b45a8
	LensCutoutTexture                     TextureUsage = 0x89bbcec2
	BloodSplatterTiler                    TextureUsage = 0x30e2d136
	WeatheringSpecial                     TextureUsage = 0xd2f99d38
	WeatheringDirt                        TextureUsage = 0x6834aa9b
	BugSplatterTiler                      TextureUsage = 0x37831285
	DecalSheet                            TextureUsage = 0x632a8b80
	CustomizationCamoTilerArray           TextureUsage = 0x0f5ff78d
	PatternMasksArray                     TextureUsage = 0x05a27dd5
	CustomizationMaterialDetailTilerArray TextureUsage = 0xd3a0408e
	IdMasksArray                          TextureUsage = 0xb281e5f2
)

func (usage *TextureUsage) String() string {
	switch *usage {
	case AlbedoIridescence:
		return "albedo_iridescence"
	case Albedo:
		return "albedo"
	case BaseData:
		return "base_data"
	case BloodSplatterTiler:
		return "blood_splatter_tiler"
	case BugSplatterTiler:
		return "bug_splatter_tiler"
	case CompositeArray:
		return "composite_array"
	case CustomizationCamoTilerArray:
		return "customization_camo_tiler_array"
	case CustomizationMaterialDetailTilerArray:
		return "customization_material_detail_tiler_array"
	case DecalSheet:
		return "decal_sheet"
	case IdMasksArray:
		return "id_masks_array"
	case MaterialLUT:
		return "material_lut"
	case Normal:
		return "normal"
	case PatternLUT:
		return "pattern_lut"
	case PatternMasksArray:
		return "pattern_masks_array"
	case WeatheringDirt:
		return "weathering_dirt"
	case WeatheringSpecial:
		return "weathering_special"
	default:
		return "unknown texture usage!"
	}
}

func addMaterial(ctx extractor.Context, mat *material.Material, doc *gltf.Document, imgOpts *ImageOptions) (uint32, error) {
	usedTextures := make(map[TextureUsage]uint32)
	var baseColorTexture *gltf.TextureInfo
	var metallicRoughnessTexture *gltf.TextureInfo
	var emissiveTexture *gltf.TextureInfo
	var normalTexture *gltf.NormalTexture
	var postProcess func(image.Image) error
	var emissiveFactor [3]float32
	origImgOpts := imgOpts
	lutImgOpts := &ImageOptions{
		Jpeg:           imgOpts.Jpeg,
		JpegQuality:    imgOpts.JpegQuality,
		PngCompression: imgOpts.PngCompression,
		Raw:            true,
	}
	for texUsage := range mat.Textures {
		switch TextureUsage(texUsage.Value) {
		case AlbedoIridescence:
			fallthrough
		case Albedo:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcessToOpaque, imgOpts)
			if err != nil {
				return 0, err
			}
			baseColorTexture = &gltf.TextureInfo{
				Index: index,
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
		case Normal:
			fallthrough
		case BaseData:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcessReconstructNormalZ, imgOpts)
			if err != nil {
				return 0, err
			}
			normalTexture = &gltf.NormalTexture{
				Index: gltf.Index(index),
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
		case MRA:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcess, imgOpts)
			if err != nil {
				return 0, err
			}
			metallicRoughnessTexture = &gltf.TextureInfo{
				Index: index,
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
		case EmissiveColor:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcess, imgOpts)
			if err != nil {
				return 0, err
			}
			emissiveTexture = &gltf.TextureInfo{
				Index: index,
			}
			emissiveFactor[0] = 1.0
			emissiveFactor[1] = 1.0
			emissiveFactor[2] = 1.0
			usedTextures[TextureUsage(texUsage.Value)] = index
		case MaterialLUT:
			fallthrough
		case PatternLUT:
			// Save raw DDS for both LUT types, to later be processed into exr
			imgOpts = lutImgOpts
			fallthrough
		case BloodSplatterTiler:
			fallthrough
		case BugSplatterTiler:
			fallthrough
		case CompositeArray:
			fallthrough
		case CustomizationCamoTilerArray:
			fallthrough
		case CustomizationMaterialDetailTilerArray:
			fallthrough
		case DecalSheet:
			fallthrough
		case IdMasksArray:
			fallthrough
		case LensCutoutTexture:
			fallthrough
		case PatternMasksArray:
			fallthrough
		case WeatheringDirt:
			fallthrough
		case WeatheringSpecial:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcess, imgOpts)
			if err != nil {
				return 0, err
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
			imgOpts = origImgOpts
		default:
			fmt.Printf("addMaterial: Unknown texture usage %v\n", texUsage.String())
		}
	}

	usagesToTextureNames := make(map[string]string)
	for usage, texIdx := range usedTextures {
		usagesToTextureNames[usage.String()] = doc.Images[*doc.Textures[texIdx].Source].Name
	}

	doc.Materials = append(doc.Materials, &gltf.Material{
		Name: mat.BaseMaterial.String(),
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorTexture:         baseColorTexture,
			MetallicRoughnessTexture: metallicRoughnessTexture,
		},
		EmissiveTexture: emissiveTexture,
		EmissiveFactor:  emissiveFactor,
		NormalTexture:   normalTexture,
		Extras:          usagesToTextureNames,
	})
	return uint32(len(doc.Materials) - 1), nil
}

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

	inverseBindMatrices := modeler.WriteAccessor(doc, gltf.TargetNone, matrices)
	jointIndices := make([]uint32, 0)
	boneBaseIndex := uint32(len(doc.Nodes))
	for _, bone := range unitInfo.Bones {
		quat := mgl32.Mat4ToQuat(bone.Transform.Rotation.Mat4())
		boneName := fmt.Sprintf("Bone_%08X", bone.NameHash.Value)
		if boneInfo != nil {
			name, exists := boneInfo.NameMap[bone.NameHash]
			if exists {
				boneName = name
			}
		}
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name:        boneName,
			Rotation:    quat.V.Vec4(quat.W),
			Translation: bone.Transform.Translation,
			Scale:       bone.Transform.Scale,
		})
		boneIdx := uint32(len(doc.Nodes) - 1)
		jointIndices = append(jointIndices, boneIdx)
		parentIndex := bone.ParentIndex + boneBaseIndex
		if parentIndex != boneIdx {
			doc.Nodes[parentIndex].Children = append(doc.Nodes[parentIndex].Children, boneIdx)
		}
	}
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, boneBaseIndex)

	doc.Skins = append(doc.Skins, &gltf.Skin{
		Name:                ctx.File().ID().Name.String(),
		InverseBindMatrices: gltf.Index(inverseBindMatrices),
		Joints:              jointIndices,
	})

	return uint32(len(doc.Skins) - 1)
}

func ConvertOpts(ctx extractor.Context, imgOpts *ImageOptions) error {
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

	doc := gltf.NewDocument()
	doc.Asset.Generator = "https://github.com/xypwn/filediver"
	doc.Samplers = append(doc.Samplers, &gltf.Sampler{
		MagFilter: gltf.MagLinear,
		MinFilter: gltf.MinLinear,
		WrapS:     gltf.WrapRepeat,
		WrapT:     gltf.WrapRepeat,
	})

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

		matIdx, err := addMaterial(ctx, mat, doc, imgOpts)
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
				meshesToLoad = entries[highestDetailIdx].Indices
			}
		} else {
			for i := uint32(0); i < unitInfo.NumMeshes; i++ {
				meshesToLoad = append(meshesToLoad, i)
			}
		}
	}

	// Load meshes
	meshes, err := unit.LoadMeshes(fGPU, unitInfo, meshesToLoad)
	if err != nil {
		return err
	}
	for meshDisplayNumber, meshID := range meshesToLoad {
		if meshID >= unitInfo.NumMeshes {
			panic("meshID out of bounds")
		}

		mesh := meshes[meshID]
		if len(mesh.UVCoords) == 0 {
			continue
		}

		// Apply vertex transform
		transform := unitInfo.Bones[mesh.Info.Header.TransformIdx].Transform
		transformMatrix := mgl32.Scale3D(transform.Scale.Elem()).Mul4(transform.Rotation.Mat4()).Mul4(mgl32.Translate3D(transform.Translation.Elem()))
		// If translation, rotation, and scale are not identities, apply transform to all vertices
		if !(transformMatrix.ApproxEqual(mgl32.Ident4())) {
			// Apply transformations
			for i := range mesh.Positions {
				p := mgl32.Vec3(mesh.Positions[i])
				p = transformMatrix.Mul4x1(p.Vec4(1)).Vec3()
				mesh.Positions[i] = p
			}
		}

		// Transform coordinates into glTF ones
		for i := range mesh.Positions {
			p := mesh.Positions[i]
			p[1], p[2] = p[2], -p[1]
			mesh.Positions[i] = p
		}

		bonesEnabled := ctx.Config()["no_bones"] != "true"

		var skin *uint32 = nil
		var weights uint32 = 0
		var joints uint32 = 0

		if bonesEnabled {
			if len(unitInfo.SkeletonMaps) > 0 && mesh.Info.Header.SkeletonMapIdx >= 0 {
				if err := remapMeshBones(&mesh, unitInfo.SkeletonMaps[mesh.Info.Header.SkeletonMapIdx]); err != nil {
					return err
				}
			}
			skin = gltf.Index(addSkeleton(ctx, doc, unitInfo, boneInfo))
			weights = modeler.WriteWeights(doc, mesh.BoneWeights)
			joints = modeler.WriteJoints(doc, mesh.BoneIndices[0])
		}

		positions := modeler.WritePosition(doc, mesh.Positions)
		var texCoords []uint32 = make([]uint32, len(mesh.UVCoords))
		for i := range mesh.UVCoords {
			texCoords[i] = modeler.WriteTextureCoord(doc, mesh.UVCoords[i])
		}
		var lodName string
		var lodNode *gltf.Node
		if len(meshesToLoad) > 1 {
			lodName = fmt.Sprintf("LOD %v", meshDisplayNumber)
			lodNode = &gltf.Node{
				Name: lodName,
			}
			doc.Nodes = append(doc.Nodes, lodNode)
			doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)-1))
		}
		for i := range mesh.Indices {
			var componentName string = fmt.Sprintf("Component %v", i)
			if lodName != "" {
				componentName = lodName + " " + componentName
			}

			var material *uint32
			if len(mesh.Info.Materials) > int(mesh.Info.Groups[i].GroupIdx) {
				if idx, ok := materialIdxs[mesh.Info.Materials[int(mesh.Info.Groups[i].GroupIdx)]]; ok {
					material = gltf.Index(idx)
				}
			}

			primitive := &gltf.Primitive{
				Indices: gltf.Index(modeler.WriteIndices(doc, mesh.Indices[i])),
				Attributes: map[string]uint32{
					gltf.POSITION: positions,
				},
				Material: material,
			}

			if bonesEnabled {
				primitive.Attributes[gltf.JOINTS_0] = joints
				primitive.Attributes[gltf.WEIGHTS_0] = weights
			}

			for j := range texCoords {
				primitive.Attributes[fmt.Sprintf("TEXCOORD_%v", j)] = texCoords[j]
			}

			doc.Meshes = append(doc.Meshes, &gltf.Mesh{
				Name: componentName + " Mesh",
				Primitives: []*gltf.Primitive{
					primitive,
				},
			})
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name: componentName,
				Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
				Skin: skin,
			})
			doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)-1))
			if len(meshesToLoad) > 1 {
				lodNode.Children = append(lodNode.Children, uint32(len(doc.Nodes)-1))
			}
		}
	}

	out, err := ctx.CreateFile(".glb")
	if err != nil {
		return err
	}
	enc := gltf.NewEncoder(out)
	if err := enc.Encode(doc); err != nil {
		return err
	}
	return nil
}

func Convert(ctx extractor.Context) error {
	var opts ImageOptions
	if v, ok := ctx.Config()["image_jpeg"]; ok && v == "true" {
		opts.Jpeg = true
	}
	if v, ok := ctx.Config()["jpeg_quality"]; ok {
		quality, err := strconv.Atoi(v)
		if err != nil {
			return err
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
	return ConvertOpts(ctx, &opts)
}
