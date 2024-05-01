package unit

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

// Adds back in the truncated Z component of a normal map.
func reconstructNormalZ(c color.Color) color.Color {
	var x, y float64
	if nc, ok := c.(color.NRGBA); ok {
		x, y = (float64(nc.R)/127.5)-1, (float64(nc.G)/127.5)-1
	} else {
		iX, iY, _, _ := c.RGBA()
		x, y = (float64(iX)/32767.5)-1, (float64(iY)/32767.5)-1
	}
	z := math.Sqrt(-x*x - y*y + 1)
	return color.RGBA64{
		R: uint16(math.Max(math.Min(math.Round((x+1)*32767.5), 65535), 0)),
		G: uint16(math.Max(math.Min(math.Round((y+1)*32767.5), 65535), 0)),
		B: uint16(math.Max(math.Min(math.Round((z+1)*32767.5), 65535), 0)),
		A: uint16(65535),
	}
}

// Attempts to completely remove the influence of the alpha channel,
// giving the whole image an opacity of 1.
// ONLY works with non-premultiplied formats due to a Go PNG bug
// (see https://github.com/golang/go/issues/26001)
func tryToOpaque(c color.Color) color.Color {
	if nc, ok := c.(color.NRGBA); ok {
		return color.NRGBA{
			R: nc.R,
			G: nc.G,
			B: nc.B,
			A: 255,
		}
	} else if nc, ok := c.(color.NRGBA64); ok {
		return color.NRGBA64{
			R: nc.R,
			G: nc.G,
			B: nc.B,
			A: 65535,
		}
	} else if nc, ok := c.(color.NYCbCrA); ok {
		return color.NYCbCrA{
			YCbCr: color.YCbCr{
				Y:  nc.Y,
				Cb: nc.Cb,
				Cr: nc.Cr,
			},
			A: 255,
		}
	}
	return c
}

type textureType int

const (
	textureTypeBaseColor textureType = iota
	textureTypeNormal
)

func tryWriteTexture(ctx extractor.Context, mat *material.Material, texType textureType, doc *gltf.Document) (uint32, bool, error) {
	var id stingray.Hash
	var pixelConv func(color.Color) color.Color
	switch texType {
	case textureTypeBaseColor:
		var ok bool
		id, ok = mat.Textures[stingray.Sum64([]byte("albedo_iridescence")).Thin()]
		if ok {
			pixelConv = tryToOpaque
			break
		}
		id, ok = mat.Textures[stingray.Sum64([]byte("albedo")).Thin()]
		if ok {
			break
		}
		return 0, false, nil
	case textureTypeNormal:
		var ok bool
		id, ok = mat.Textures[stingray.Sum64([]byte("normal")).Thin()]
		if ok {
			pixelConv = reconstructNormalZ
			break
		}
		return 0, false, nil
	default:
		panic("unhandled case")
	}
	res, err := writeTexture(ctx, doc, id, pixelConv)
	if err != nil {
		return 0, false, err
	}
	return res, true, nil
}

// Adds a texture to doc. Returns new texture ID if err != nil.
// pixelConv optionally converts individual pixel colors.
func writeTexture(ctx extractor.Context, doc *gltf.Document, id stingray.Hash, pixelConv func(color.Color) color.Color) (uint32, error) {
	file, exists := ctx.GetResource(id, stingray.Sum64([]byte("texture")))
	if !exists || !file.Exists(stingray.DataMain) {
		return 0, fmt.Errorf("texture resource %v doesn't exist", id)
	}

	tex, err := texture.Decode(ctx.Ctx(), file, false)
	if err != nil {
		return 0, err
	}

	if pixelConv != nil {
		if img, ok := tex.Image.(interface {
			image.Image
			Set(int, int, color.Color)
		}); ok {
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
					img.Set(x, y, pixelConv(img.At(x, y)))
				}
			}
		} else {
			return 0, errors.New("DDS image does not support Set()")
		}
	}
	// Keeping this around since PNG is lossless, but JPEG is significantly faster
	// TODO: Allow specifying PNG/JPEG as argument
	/*var pngData bytes.Buffer
	if err := (&png.Encoder{
		CompressionLevel: png.DefaultCompression,
	}).Encode(&pngData, tex); err != nil {
		return 0, err
	}
	imgIdx, err := modeler.WriteImage(doc, id.String(), "image/png", &pngData)
	if err != nil {
		return 0, err
	}*/
	var jpgData bytes.Buffer
	if err := jpeg.Encode(&jpgData, tex, &jpeg.Options{Quality: 95}); err != nil {
		return 0, err
	}
	imgIdx, err := modeler.WriteImage(doc, id.String(), "image/jpeg", &jpgData)
	if err != nil {
		return 0, err
	}
	doc.Textures = append(doc.Textures, &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(imgIdx),
	})
	return uint32(len(doc.Textures) - 1), nil
}

func Convert(ctx extractor.Context) error {
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

	u, err := unit.Load(fMain, fGPU)
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
	for id, resID := range u.Materials {
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

		/*materialNames := make(map[stingray.ThinHash]string)
		f, err := os.Open("material_textures.txt")
		if err != nil {
			return err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			s := sc.Text()
			materialNames[stingray.Sum64([]byte(s)).Thin()] = s
		}
		fmt.Println()
		for k, v := range mat.Textures {
			name := k.String()
			if s, ok := materialNames[k]; ok {
				name = s
			}
			fmt.Println(name, v)
		}*/

		/*for k, v := range mat.Textures {
			texRes, exists := ctx.GetResource(v, stingray.Sum64([]byte("texture")))
			if !exists || !texRes.Exists(stingray.DataMain) {
				return fmt.Errorf("texture resource %v doesn't exist", id)
			}

			tex, err := texture.Decode(texRes, false)
			if err != nil {
				return err
			}

			if err := func() error {
				out, err := ctx.CreateFileDir(".unit.textures", k.String()+"_"+v.String()+".png")
				if err != nil {
					return err
				}
				defer out.Close()
				if err := png.Encode(out, tex); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return err
			}
		}*/

		texIdxBaseColor, ok, err := tryWriteTexture(ctx, mat, textureTypeBaseColor, doc)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		texIdxNormal, ok, err := tryWriteTexture(ctx, mat, textureTypeNormal, doc)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		doc.Materials = append(doc.Materials, &gltf.Material{
			Name: resID.String(),
			PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
				BaseColorTexture: &gltf.TextureInfo{
					Index: texIdxBaseColor,
				},
				MetallicFactor:  gltf.Float(0.5),
				RoughnessFactor: gltf.Float(1),
			},
			NormalTexture: &gltf.NormalTexture{
				Index: gltf.Index(texIdxNormal),
			},
		})
		materialIdxs[id] = uint32(len(doc.Materials) - 1)
	}

	// Determine which meshes to convert
	var meshIndices []uint32
	switch ctx.Config()["meshes"] {
	case "all":
		for i := range u.Meshes {
			meshIndices = append(meshIndices, uint32(i))
		}
	default: // "highest_detail"
		if len(u.LODGroups) > 0 {
			entries := u.LODGroups[0].Entries
			highestDetailIdx := -1
			for i := range entries {
				if highestDetailIdx == -1 || entries[i].Detail.Max > entries[highestDetailIdx].Detail.Max {
					highestDetailIdx = i
				}
			}
			if highestDetailIdx != -1 {
				meshIndices = entries[highestDetailIdx].Indices
			}
		}
	}

	// Load meshes
	for _, meshIdx := range meshIndices {
		if int(meshIdx) >= len(u.Meshes) {
			// TODO: Figure out WTF is going on here
			continue
		}

		mesh := u.Meshes[meshIdx]
		if len(mesh.UVCoords) == 0 {
			continue
		}

		// Transform coordinates into glTF ones
		for i := range mesh.Positions {
			p := mesh.Positions[i]
			p[0], p[1], p[2] = p[1], p[2], p[0]
			mesh.Positions[i] = p
		}

		name := fmt.Sprintf("Mesh %v", len(doc.Meshes))
		var material *uint32
		if len(mesh.Info.Materials) > 0 {
			if idx, ok := materialIdxs[mesh.Info.Materials[0]]; ok {
				material = gltf.Index(idx)
			}
		}
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{
			Name: name,
			Primitives: []*gltf.Primitive{
				{
					Indices: gltf.Index(modeler.WriteIndices(doc, mesh.Indices)),
					Attributes: map[string]uint32{
						gltf.POSITION:   modeler.WritePosition(doc, mesh.Positions),
						gltf.TEXCOORD_0: modeler.WriteTextureCoord(doc, mesh.UVCoords),
						//gltf.COLOR_0:    modeler.WriteColor(doc, mesh.Colors),
					},
					Material: material,
				},
			},
		})
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: name,
			Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
		})
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)-1))
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
