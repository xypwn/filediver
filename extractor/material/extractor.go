package material

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"path/filepath"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/blend_helper"
	extr_texture "github.com/xypwn/filediver/extractor/texture"
	"github.com/xypwn/filediver/stingray"
	dlbin "github.com/xypwn/filediver/stingray/dl_bin"
	"github.com/xypwn/filediver/stingray/unit/material"
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

// Returns a function that uses the alpha of an emissive map and an emissive color to create
// a gltf emissive map
func createPostProcessEmissiveColor(color []float32) (func(image.Image) error, error) {
	if len(color) < 3 {
		return nil, fmt.Errorf("createPostProcessEmissiveColor: color %v does not have enough entries", color)
	}
	return func(img image.Image) error {
		switch img := img.(type) {
		case *image.NRGBA:
			for iY := img.Rect.Min.Y; iY < img.Rect.Max.Y; iY++ {
				for iX := img.Rect.Min.X; iX < img.Rect.Max.X; iX++ {
					idx := img.PixOffset(iX, iY)
					alphaPct := float32(img.Pix[idx+3]) / 255.0
					img.Pix[idx] = uint8(color[0] * 255.0 * alphaPct)
					img.Pix[idx+1] = uint8(color[1] * 255.0 * alphaPct)
					img.Pix[idx+2] = uint8(color[2] * 255.0 * alphaPct)
					img.Pix[idx+3] = 255
				}
			}
			return nil
		default:
			return errors.New("postProcessEmissiveColor: unsupported image type")
		}
	}, nil
}

// Moves the clearcoat data to the location expected by the gltf materials
func postProcessIlluminateClearcoat(img image.Image) error {
	/**
	 * illuminate_data:
	 *	R - coat roughness
	 *	G - metallic
	 *	B - coat weight
	 *	A - unknown
	 */
	switch img := img.(type) {
	case *image.NRGBA:
		for iY := img.Rect.Min.Y; iY < img.Rect.Max.Y; iY++ {
			for iX := img.Rect.Min.X; iX < img.Rect.Max.X; iX++ {
				idx := img.PixOffset(iX, iY)
				img.Pix[idx+1] = img.Pix[idx]
				img.Pix[idx] = img.Pix[idx+2]
			}
		}
		return nil
	default:
		return errors.New("postProcessIlluminateClearcoat: unsupported image type")
	}
}

// Adds a texture to doc. Returns new texture ID if err != nil.
// postProcess optionally applies image post-processing.
func writeTexture(ctx *extractor.Context, doc *gltf.Document, id stingray.Hash, postProcess func(image.Image) error, imgOpts *ImageOptions, suffix string) (uint32, error) {
	// Check if we've already added this texture
	for j, texture := range doc.Textures {
		if doc.Images[*texture.Source].Name == (id.String() + suffix) {
			return uint32(j), nil
		}
	}

	ddsData, err := extr_texture.ExtractDDSData(ctx, stingray.NewFileID(id, stingray.Sum("texture")))
	if err != nil {
		return 0, err
	}
	tex, err := dds.Decode(bytes.NewReader(ddsData), false)
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
	imgIdx, err := modeler.WriteImage(doc, id.String()+suffix, mimeType, &encData)
	if err != nil {
		return 0, err
	}
	doc.Textures = append(doc.Textures, &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(imgIdx),
	})
	texIdx := uint32(len(doc.Textures) - 1)
	if imgOpts != nil && imgOpts.Raw {
		reader := bytes.NewReader(ddsData)
		mimeType = "image/vnd-ms.dds"
		imgIdx, err = modeler.WriteImage(doc, id.String()+suffix+".dds", mimeType, reader)
		if err != nil {
			ctx.Warnf("writeTexture: failed to write dds image to document")
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

func combineIlluminateOcclusionMetallicRoughness(narImg, dataImg image.Image) error {
	narToDataX := float32(dataImg.Bounds().Size().X) / float32(narImg.Bounds().Size().X)
	narToDataY := float32(dataImg.Bounds().Size().Y) / float32(narImg.Bounds().Size().Y)

	narImgNRGBA, ok := narImg.(*image.NRGBA)
	if !ok {
		return fmt.Errorf("combineIlluminateOcclusionMetallicRoughness: unsupported NAR image type")
	}
	dataImgNRGBA, ok := dataImg.(*image.NRGBA)
	if !ok {
		return fmt.Errorf("combineIlluminateOcclusionMetallicRoughness: unsupported illuminate data image type")
	}

	/**
	 * NAR:
	 *	R - normal X
	 *	G - normal Y
	 *	B - ambient occlusion
	 *	A - roughness
	 */
	/**
	 * illuminate_data:
	 *	R - coat roughness
	 *	G - metallic
	 *	B - coat weight
	 *	A - unknown
	 */

	for iY := narImgNRGBA.Rect.Min.Y; iY < narImgNRGBA.Rect.Max.Y; iY++ {
		for iX := narImgNRGBA.Rect.Min.X; iX < narImgNRGBA.Rect.Max.X; iX++ {
			narIdx := narImgNRGBA.PixOffset(iX, iY)
			dataIdx := dataImgNRGBA.PixOffset(min(int(float32(iX)*narToDataX), dataImgNRGBA.Rect.Max.X-1), min(int(float32(iY)*narToDataY), dataImgNRGBA.Rect.Max.Y-1))
			// Move NAR ambient occlusion to red channel
			narImgNRGBA.Pix[narIdx] = narImgNRGBA.Pix[narIdx+2]
			// Move NAR roughness to green channel
			narImgNRGBA.Pix[narIdx+1] = narImgNRGBA.Pix[narIdx+3]
			// Move illuminate data metallic to blue channel
			narImgNRGBA.Pix[narIdx+2] = dataImgNRGBA.Pix[dataIdx+1]
		}
	}
	return nil
}

// Combines illuminate data and NAR into a gltf compliant ao, metallic, roughness map and returns the index
func writeIlluminateOcclusionMetallicRoughnessTexture(ctx *extractor.Context, doc *gltf.Document, narId, ilDataId stingray.Hash, imgOpts *ImageOptions) (uint32, error) {
	// Check if we've already added this texture
	textureName := narId.String() + "_" + ilDataId.String() + "_orm"
	for j, texture := range doc.Textures {
		if doc.Images[*texture.Source].Name == textureName {
			return uint32(j), nil
		}
	}

	narR, err := extr_texture.ExtractDDSData(ctx,
		stingray.NewFileID(narId, stingray.Sum("texture")))
	if err != nil {
		return 0, err
	}
	ilDataR, err := extr_texture.ExtractDDSData(ctx,
		stingray.NewFileID(ilDataId, stingray.Sum("texture")))
	if err != nil {
		return 0, err
	}

	narTex, err := dds.Decode(bytes.NewReader(narR), false)
	if err != nil {
		return 0, err
	}

	ilDataTex, err := dds.Decode(bytes.NewReader(ilDataR), false)
	if err != nil {
		return 0, err
	}

	if len(narTex.Images) > 1 || len(ilDataTex.Images) > 1 {
		return 0, fmt.Errorf("NAR or illuminate data are texture arrays, not sure how to handle")
	}

	if err := combineIlluminateOcclusionMetallicRoughness(narTex.Image, ilDataTex.Image); err != nil {
		return 0, err
	}

	var encData bytes.Buffer
	var mimeType string
	if imgOpts != nil && imgOpts.Jpeg {
		quality := jpeg.DefaultQuality
		if imgOpts.JpegQuality != 0 {
			quality = imgOpts.JpegQuality
		}
		if err := jpeg.Encode(&encData, narTex, &jpeg.Options{Quality: quality}); err != nil {
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
		}).Encode(&encData, narTex); err != nil {
			return 0, err
		}
		mimeType = "image/png"
	}
	imgIdx, err := modeler.WriteImage(doc, textureName, mimeType, &encData)
	if err != nil {
		return 0, err
	}
	doc.Textures = append(doc.Textures, &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(imgIdx),
	})
	texIdx := uint32(len(doc.Textures) - 1)
	return texIdx, nil
}

func compareMaterials(doc *gltf.Document, mat *material.Material, matIdx uint32, matName string, unitData *dlbin.UnitData) bool {
	if doc.Materials[matIdx].Name != matName {
		return false
	}
	for texUsage := range mat.Textures {
		usage := TextureUsage(texUsage.Value)
		extras := doc.Materials[matIdx].Extras.(map[string]interface{})
		texIdxInterface, contains := extras[usage.String()]
		if !contains {
			continue
		}
		texIdx, ok := texIdxInterface.(uint32)
		if !ok {
			continue
		}
		texture := doc.Textures[texIdx]
		imgName := doc.Images[*texture.Source].Name
		materialTexName := mat.Textures[texUsage].String()
		if unitData != nil {
			switch usage {
			case MaterialLUT:
				if unitData.MaterialLut.Value == 0 {
					break
				}
				materialTexName = unitData.MaterialLut.String()
			case PatternLUT:
				if unitData.PatternLut.Value == 0 {
					break
				}
				materialTexName = unitData.PatternLut.String()
			case CapeLUT:
				if unitData.CapeLut.Value == 0 {
					break
				}
				materialTexName = unitData.CapeLut.String()
			case BaseData:
				if unitData.BaseData.Value == 0 {
					break
				}
				materialTexName = unitData.BaseData.String()
			case DecalSheet:
				if unitData.DecalSheet.Value == 0 {
					break
				}
				materialTexName = unitData.DecalSheet.String()
			}
		}
		if imgName != materialTexName {
			return false
		}
	}
	return true
}

func AddMaterial(ctx *extractor.Context, mat *material.Material, doc *gltf.Document, imgOpts *ImageOptions, matName string, unitData *dlbin.UnitData) (uint32, error) {
	cfg := ctx.Config()

	// Avoid duplicating material if it already is added to document
	for i := range doc.Materials {
		if compareMaterials(doc, mat, uint32(i), matName, unitData) {
			return uint32(i), nil
		}
	}
	usedTextures := make(map[TextureUsage]uint32)
	var baseColorTexture *gltf.TextureInfo
	var metallicRoughnessTexture *gltf.TextureInfo
	var emissiveTexture *gltf.TextureInfo
	var normalTexture *gltf.NormalTexture
	var occlusionTexture *gltf.OcclusionTexture
	var coatTexture *gltf.TextureInfo
	var postProcess func(image.Image) error
	var albedoPostProcess func(image.Image) error = postProcessToOpaque
	var normalPostProcess func(image.Image) error = postProcessReconstructNormalZ
	var emissiveFactor [3]float32
	var emissiveStrength float32 = 1.0
	origImgOpts := imgOpts
	lutImgOpts := &ImageOptions{
		Jpeg:           imgOpts.Jpeg,
		JpegQuality:    imgOpts.JpegQuality,
		PngCompression: imgOpts.PngCompression,
		Raw:            true,
	}
	for texUsage := range mat.Textures {
		switch TextureUsage(texUsage.Value) {
		case ColorRoughness:
			fallthrough
		case ColorSpecularB:
			fallthrough
		case AlbedoIridescence:
			albedoPostProcess = nil
			fallthrough
		case CoveringAlbedo:
			fallthrough
		case InputImage:
			fallthrough
		case ReticleTexture:
			fallthrough
		case Albedo:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], albedoPostProcess, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			baseColorTexture = &gltf.TextureInfo{
				Index: index,
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
			albedoPostProcess = postProcessToOpaque
		case AlbedoEmissive:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcessToOpaque, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			baseColorTexture = &gltf.TextureInfo{
				Index: index,
			}
			emissiveColorSetting, ok := mat.Settings[stingray.ThinHash{Value: uint32(SettingEmissiveColor)}]
			if !ok {
				ctx.Warnf("material %v has AlbedoEmissive texture but no emissive_color", matName)
				continue
			}
			postProcessEmissiveColor, err := createPostProcessEmissiveColor(emissiveColorSetting)
			if err != nil {
				ctx.Warnf("createPostProcessEmissiveColor: %v", err)
				continue
			}
			emissiveIndex, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcessEmissiveColor, imgOpts, "_emissive")
			if err != nil {
				return 0, err
			}
			emissiveTexture = &gltf.TextureInfo{
				Index: emissiveIndex,
			}
			emissiveFactor[0] = 1.0
			emissiveFactor[1] = 1.0
			emissiveFactor[2] = 1.0
			emissiveStrengthSetting, ok := mat.Settings[stingray.ThinHash{Value: uint32(SettingEmissiveIntensity)}]
			if !ok {
				emissiveStrengthSetting, ok = mat.Settings[stingray.ThinHash{Value: uint32(SettingEmissiveMult)}]
			}
			if !ok {
				emissiveStrengthSetting, ok = mat.Settings[stingray.ThinHash{Value: uint32(SettingEmissiveStrength)}]
			}
			if !ok || len(emissiveStrengthSetting) == 0 {
				continue
			}
			emissiveStrength = emissiveStrengthSetting[0]
		case NormalSpecularAO:
			// GLTF normals will look wonky, but our own material will be able to use the specular+ao in this map
			// in blender
			normalPostProcess = nil
			fallthrough
		case Normal:
			fallthrough
		case Normals:
			fallthrough
		case NormalMap:
			fallthrough
		case CoveringNormal:
			fallthrough
		case NAC:
			fallthrough
		case BaseData:
			hash := mat.Textures[texUsage]
			if unitData != nil && TextureUsage(texUsage.Value) == BaseData && unitData.BaseData.Value != 0 {
				hash = unitData.BaseData
			}
			index, err := writeTexture(ctx, doc, hash, normalPostProcess, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			normalTexture = &gltf.NormalTexture{
				Index: gltf.Index(index),
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
			normalPostProcess = postProcessReconstructNormalZ
		case NAR:
			hash := mat.Textures[texUsage]
			index, err := writeTexture(ctx, doc, hash, postProcessReconstructNormalZ, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			normalTexture = &gltf.NormalTexture{
				Index: gltf.Index(index),
			}
			illuminateDataHash, ok := mat.Textures[stingray.Sum("illuminate_data").Thin()]
			if metallicRoughnessTexture == nil && ok {
				metallicRoughnessIndex, err := writeIlluminateOcclusionMetallicRoughnessTexture(ctx, doc, hash, illuminateDataHash, imgOpts)
				if err != nil {
					ctx.Warnf("writeIlluminateOcclusionMetallicRoughnessTexture: %v", err)
					continue
				}
				metallicRoughnessTexture = &gltf.TextureInfo{
					Index: metallicRoughnessIndex,
				}
			}
			if occlusionTexture == nil && ok {
				occlusionIndex, err := writeIlluminateOcclusionMetallicRoughnessTexture(ctx, doc, hash, illuminateDataHash, imgOpts)
				if err != nil {
					ctx.Warnf("writeIlluminateOcclusionMetallicRoughnessTexture: %v", err)
					continue
				}
				occlusionTexture = &gltf.OcclusionTexture{
					Index: gltf.Index(occlusionIndex),
				}
			}
		case IlluminateData:
			hash := mat.Textures[texUsage]
			index, err := writeTexture(ctx, doc, hash, postProcessIlluminateClearcoat, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			coatTexture = &gltf.TextureInfo{
				Index: index,
			}
			narHash, ok := mat.Textures[stingray.Sum("NAR").Thin()]
			if metallicRoughnessTexture == nil && ok {
				metallicRoughnessIndex, err := writeIlluminateOcclusionMetallicRoughnessTexture(ctx, doc, narHash, hash, imgOpts)
				if err != nil {
					ctx.Warnf("writeIlluminateOcclusionMetallicRoughnessTexture: %v", err)
					continue
				}
				metallicRoughnessTexture = &gltf.TextureInfo{
					Index: metallicRoughnessIndex,
				}
			}
			if occlusionTexture == nil && ok {
				occlusionIndex, err := writeIlluminateOcclusionMetallicRoughnessTexture(ctx, doc, narHash, hash, imgOpts)
				if err != nil {
					ctx.Warnf("writeIlluminateOcclusionMetallicRoughnessTexture: %v", err)
					continue
				}
				occlusionTexture = &gltf.OcclusionTexture{
					Index: gltf.Index(occlusionIndex),
				}
			}
		case MRA:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcess, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			metallicRoughnessTexture = &gltf.TextureInfo{
				Index: index,
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
		case EmissiveColor:
			fallthrough
		case LensEmissiveTexture:
			index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcess, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
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
		case TextureLUT:
			fallthrough
		case CapeLUT:
			fallthrough
		case LUTEmissive:
			fallthrough
		case BloodLUT:
			fallthrough
		case BrdfLUT:
			fallthrough
		case ColorLUT:
			fallthrough
		case ColorRoughnessLUT:
			fallthrough
		case ContinentsLUT:
			fallthrough
		case CorporateColorRoughnessLUT:
			fallthrough
		case CosmicDustLUT:
			fallthrough
		case EmissiveNebulaLUT:
			fallthrough
		case EyeLUT:
			fallthrough
		case MinimapLUT:
			fallthrough
		case MoonLUT:
			fallthrough
		case PaletteLUT:
			fallthrough
		case SpaceStarLUT:
			fallthrough
		case SpaceStarLUTTmp:
			fallthrough
		case SpecularBrdfLUT:
			fallthrough
		case SssLUT:
			fallthrough
		case WoundLUTToAdd:
			fallthrough
		case PatternLUT:
			// Save raw DDS for all LUT types, to later be processed into exr
			imgOpts = lutImgOpts
			hash := mat.Textures[texUsage]
			if unitData != nil && TextureUsage(texUsage.Value) == MaterialLUT && unitData.MaterialLut.Value != 0 {
				hash = unitData.MaterialLut
			} else if unitData != nil && TextureUsage(texUsage.Value) == PatternLUT && unitData.PatternLut.Value != 0 {
				hash = unitData.PatternLut
			} else if unitData != nil && TextureUsage(texUsage.Value) == CapeLUT && unitData.CapeLut.Value != 0 {
				hash = unitData.CapeLut
			}
			index, err := writeTexture(ctx, doc, hash, postProcess, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
			imgOpts = origImgOpts
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
		case DetailData:
			fallthrough
		case MetalSurfaceData:
			fallthrough
		case ConcreteSurfaceData:
			fallthrough
		case GrayscaleSkin:
			fallthrough
		case PatternMasksArray:
			hash := mat.Textures[texUsage]
			if unitData != nil && TextureUsage(texUsage.Value) == DecalSheet && unitData.DecalSheet.Value != 0 {
				hash = unitData.DecalSheet
			}
			index, err := writeTexture(ctx, doc, hash, postProcess, imgOpts, "")
			if err != nil {
				ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
				continue
			}
			usedTextures[TextureUsage(texUsage.Value)] = index
		case BloodSplatterTiler:
			fallthrough
		case BugSplatterTiler:
			fallthrough
		case LensCutoutTexture:
			fallthrough
		case ScorchMarks:
			fallthrough
		case SubsurfaceOpacity:
			fallthrough
		case WeatheringDirt:
			fallthrough
		case WeatheringSpecial:
			fallthrough
		case DirtMap:
			fallthrough
		case NoiseArray:
			fallthrough
		case LightBleedMap:
			fallthrough
		case DistortionMap:
			fallthrough
		case WeatheringDataMask:
			fallthrough
		case WoundData:
			fallthrough
		case WoundDerivative:
			if cfg.Unit.AllTextures {
				index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcess, imgOpts, "")
				if err != nil {
					ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
					continue
				}
				usedTextures[TextureUsage(texUsage.Value)] = index
			}
		case WoundNormal:
			if cfg.Unit.AllTextures {
				index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcessReconstructNormalZ, imgOpts, "")
				if err != nil {
					ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
					continue
				}
				usedTextures[TextureUsage(texUsage.Value)] = index
			}
		default:
			if cfg.Unit.AllTextures {
				t := TextureUsage(texUsage.Value)
				ctx.Warnf("addMaterial: unknown/unhandled texture usage %v in material %v", t.String(), matName)
				index, err := writeTexture(ctx, doc, mat.Textures[texUsage], postProcess, imgOpts, "")
				if err != nil {
					ctx.Warnf("writeTexture: %v: %v", TextureUsage(texUsage.Value), err)
					continue
				}
				usedTextures[TextureUsage(texUsage.Value)] = index
			}
		}
	}

	usagesToTextureIndices := make(map[string]interface{})
	for usage, texIdx := range usedTextures {
		usagesToTextureIndices[usage.String()] = texIdx
	}

	for setting, value := range mat.Settings {
		usage := SettingsUsage(setting.Value)
		usagesToTextureIndices[usage.String()] = value
	}

	doc.Materials = append(doc.Materials, &gltf.Material{
		Name: matName,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorTexture:         baseColorTexture,
			MetallicRoughnessTexture: metallicRoughnessTexture,
		},
		EmissiveTexture:  emissiveTexture,
		EmissiveFactor:   emissiveFactor,
		NormalTexture:    normalTexture,
		OcclusionTexture: occlusionTexture,
		Extras:           usagesToTextureIndices,
	})
	if coatTexture != nil {
		clearcoat := make(map[string]interface{})
		clearcoat["clearcoatTexture"] = coatTexture
		clearcoat["clearcoatRoughnessTexture"] = coatTexture
		clearcoat["clearcoatNormalTexture"] = normalTexture
		if doc.Materials[len(doc.Materials)-1].Extensions == nil {
			doc.Materials[len(doc.Materials)-1].Extensions = make(map[string]interface{})
		}
		doc.Materials[len(doc.Materials)-1].Extensions["KHR_materials_clearcoat"] = clearcoat
	}
	if emissiveStrength != 1.0 {
		if doc.Materials[len(doc.Materials)-1].Extensions == nil {
			doc.Materials[len(doc.Materials)-1].Extensions = make(map[string]interface{})
		}
		strength := make(map[string]interface{})
		if emissiveStrength > 1.0 {
			strength["emissiveStrength"] = emissiveStrength
		} else if emissiveStrength != 0.0 {
			strength["emissiveStrength"] = 1.0 / emissiveStrength
		}
		doc.Materials[len(doc.Materials)-1].Extensions["KHR_materials_emissive_strength"] = strength
	}
	return uint32(len(doc.Materials) - 1), nil
}

// Uses ctx.Config().Material.Format as format! Add an extra parameter for
// format if this is made public!
func convertOpts(ctx *extractor.Context, imgOpts *ImageOptions, gltfDoc *gltf.Document) error {
	cfg := ctx.Config()

	fMain, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	mat, err := material.Load(fMain)
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

	positions := [][3]float32{
		{-1.0, 0.0, -1.0},
		{1.0, 0.0, -1.0},
		{1.0, 0.0, 1.0},
		{-1.0, 0.0, 1.0},
	}

	uvCoords := [][2]float32{
		{0.0, 0.0},
		{1.0, 0.0},
		{1.0, 1.0},
		{0.0, 1.0},
	}

	indices := []uint32{
		2, 1, 0,
		0, 3, 2,
	}

	if len(doc.Accessors) < 1 {
		modeler.WriteIndices(doc, indices)
	}

	if len(doc.Accessors) < 2 {
		modeler.WritePosition(doc, positions)
	}

	if len(doc.Accessors) < 3 {
		modeler.WriteTextureCoord(doc, uvCoords)
	}

	if gltfDoc != nil && len(mat.Textures) == 0 {
		return nil
	}

	_, containsIdMasks := mat.Textures[stingray.ThinHash{Value: uint32(IdMasksArray)}]
	if gltfDoc != nil && cfg.Unit.AccurateOnly && !containsIdMasks {
		return nil
	}

	matIdx, err := AddMaterial(ctx, mat, doc, imgOpts, ctx.FileID().Name.String(), nil)
	if err != nil {
		return err
	}

	// If we're writing a combined document and this material has no textures, skip it
	if gltfDoc != nil {
		if extras, ok := doc.Materials[matIdx].Extras.(map[string]uint32); ok {
			if len(extras) == 0 {
				return nil
			}
		} else {
			// Couldn't convert extras to a map, so it doesn't have any entries?
			return nil
		}
	}

	primitive := &gltf.Primitive{
		Indices: gltf.Index(0),
		Attributes: map[string]uint32{
			gltf.POSITION:   1,
			gltf.TEXCOORD_0: 2,
			// gltf.JOINTS_0:   modeler.WriteJoints(doc, boneIndices),
			// gltf.WEIGHTS_0:  modeler.WriteWeights(doc, boneWeights),
		},
		Material: &matIdx,
	}

	doc.Meshes = append(doc.Meshes, &gltf.Mesh{
		Name: ctx.FileID().Name.String(),
		Primitives: []*gltf.Primitive{
			primitive,
		},
	})
	spiral := func(n int) (int, int) {
		// Ulam spiral
		K := math.Ceil(0.5 * (math.Sqrt(float64(n)) - 1))
		d := math.Pow((2*K+1), 2.0) - float64(n)
		if 0 <= d && d <= (2*K+1) {
			return int(-K), int(K + 1 - d)
		} else if d <= (4*K + 1) {
			return int(-3*K - 1 + d), int(-K)
		} else if d <= (6*K + 1) {
			return int(K), int(-5*K - 1 + d)
		} else {
			return int(7*K + 1 - d), int(K)
		}
	}
	y, x := spiral(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name:        ctx.FileID().Name.String() + " Visualizer",
		Mesh:        gltf.Index(uint32(len(doc.Meshes) - 1)),
		Translation: [3]float32{float32(2 * x), 0.0, float32(2 * y)},
	})
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)-1))

	formatIsBlend := cfg.Material.Format == "blend"
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

func GetImageOpts(ctx *extractor.Context) (*ImageOptions, error) {
	cfg := ctx.Config()

	var opts ImageOptions
	opts.Jpeg = cfg.Unit.ImageFormat == "jpeg"
	opts.JpegQuality = cfg.Unit.JpegQuality
	switch cfg.Unit.PngCompression {
	case "default":
		opts.PngCompression = png.DefaultCompression
	case "none":
		opts.PngCompression = png.NoCompression
	case "fast":
		opts.PngCompression = png.BestSpeed
	case "best":
		opts.PngCompression = png.BestCompression
	}
	return &opts, nil
}

func Convert(currDoc *gltf.Document) func(ctx *extractor.Context) error {
	return func(ctx *extractor.Context) error {
		opts, err := GetImageOpts(ctx)
		if err != nil {
			return err
		}
		return convertOpts(ctx, opts, currDoc)
	}
}

// Uses ctx.Config().Material.TexturesFormat as format for individual textures!
// Add an extra parameter for format when this is used by another extractor.
func ConvertTextures(ctx *extractor.Context) error {
	cfg := ctx.Config()

	fMain, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	mat, err := material.Load(fMain)
	if err != nil {
		return err
	}

	for _, texture := range mat.Textures {
		id := stingray.NewFileID(texture, stingray.Sum("texture"))
		var data []byte
		var err error
		if cfg.Material.TexturesFormat == "dds" {
			data, err = extr_texture.ExtractDDSData(ctx, id)
		} else {
			data, err = extr_texture.ConvertToPNGData(ctx, id)
		}
		if err != nil {
			ctx.Warnf("read %v.texture: %w", ctx.LookupHash(texture), err)
			continue
		}

		texName, ok := ctx.Hashes()[texture]
		if ok {
			// textures are usually in the format
			// [...]/textures/[texName]
			if idx := strings.Index(texName, "/textures/"); idx != -1 {
				texName = texName[idx+len("/textures/"):]
			}
		} else {
			texName = texture.String()
		}

		out, err := ctx.CreateFile(filepath.Join(".dir", texName+"."+cfg.Material.TexturesFormat))
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = out.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}
