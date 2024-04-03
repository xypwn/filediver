package unit

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

// Adds back in the truncated Z component of a normal map.
func reconstructNormalZ(c color.Color) color.Color {
	iX, iY, _, _ := c.RGBA()
	x, y := (float64(iX)/32767.5)-1, (float64(iY)/32767.5)-1
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

// Adds a texture to doc. Returns new texture ID if err != nil.
// pixelConv optionally converts individual pixel colors.
func writeTexture(doc *gltf.Document, getResource extractor.GetResourceFunc, id stingray.Hash, pixelConv func(color.Color) color.Color) (uint32, error) {
	texRes := getResource(id, stingray.Sum64([]byte("texture")))
	if texRes == nil || !texRes.Exists(stingray.DataMain) || !texRes.Exists(stingray.DataStream) {
		return 0, fmt.Errorf("texture resource %v doesn't exist", id)
	}
	fMain, err := texRes.Open(stingray.DataMain)
	if err != nil {
		return 0, err
	}
	defer fMain.Close()
	fStream, err := texRes.Open(stingray.DataStream)
	if err != nil {
		return 0, err
	}
	defer fStream.Close()

	tex, err := texture.Load(fMain)
	if err != nil {
		return 0, err
	}
	if _, err := fMain.Seek(int64(tex.HeaderOffset), io.SeekStart); err != nil {
		return 0, err
	}
	dds, err := dds.Decode(io.MultiReader(fMain, fStream), false)
	if err != nil {
		return 0, err
	}
	if pixelConv != nil {
		if img, ok := dds.Image.(interface {
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
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, dds); err != nil {
		return 0, err
	}
	imgIdx, err := modeler.WriteImage(doc, id.String(), "image/png", &pngData)
	if err != nil {
		return 0, err
	}
	doc.Textures = append(doc.Textures, &gltf.Texture{
		Sampler: gltf.Index(0),
		Source:  gltf.Index(imgIdx),
	})
	return uint32(len(doc.Textures) - 1), nil
}

func Convert(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, _ *exec.Runner, getResource extractor.GetResourceFunc) error {
	u, err := unit.Load(ins[stingray.DataMain], ins[stingray.DataGPU])
	if err != nil {
		return err
	}

	// Transform coordinates into glTF ones
	for _, mesh := range u.Meshes {
		for i := range mesh.Positions {
			p := mesh.Positions[i]
			p[0], p[1], p[2] = p[1], p[2], p[0]
			mesh.Positions[i] = p
		}
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
		matRes := getResource(resID, stingray.Sum64([]byte("material")))
		if matRes == nil || !matRes.Exists(stingray.DataMain) {
			return fmt.Errorf("referenced material resource %v doesn't exist", resID)
		}
		mat, err := func() (*material.Material, error) {
			f, err := matRes.Open(stingray.DataMain)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			return material.Load(f)
		}()
		if err != nil {
			return err
		}
		texIDBaseColor, ok := mat.Textures[stingray.Sum64([]byte("albedo_iridescence")).Thin()]
		if !ok {
			continue
		}
		texIdxBaseColor, err := writeTexture(doc, getResource, texIDBaseColor, tryToOpaque)
		if err != nil {
			return err
		}
		texIDNormal, ok := mat.Textures[stingray.Sum64([]byte("normal")).Thin()]
		if !ok {
			continue
		}
		texIdxNormal, err := writeTexture(doc, getResource, texIDNormal, reconstructNormalZ)
		if err != nil {
			return err
		}
		/*texIDEmissive, ok := mat.Textures[stingray.Sum64([]byte("subsurface_opacity")).Thin()]
		if !ok {
			continue
		}
		texIdxEmissive, err := writeTexture(texIDEmissive, nil)
		if err != nil {
			return err
		}*/
		doc.Materials = append(doc.Materials, &gltf.Material{
			Name: resID.String(),
			PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
				BaseColorTexture: &gltf.TextureInfo{
					Index: texIdxBaseColor,
				},
			},
			NormalTexture: &gltf.NormalTexture{
				Index: gltf.Index(texIdxNormal),
			},
			/*EmissiveTexture: &gltf.TextureInfo{
				Index: texIdxEmissive,
			},
			EmissiveFactor: [3]float32{0.5, 0.5, 0.5},*/
			AlphaMode: gltf.AlphaOpaque,
		})
		materialIdxs[id] = uint32(len(doc.Materials) - 1)
	}

	// Load meshes
	for _, mesh := range u.Meshes {
		if len(mesh.UVCoords) == 0 {
			continue
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
	}

	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, 0)
	if err := gltf.SaveBinary(doc, outPath+".glb"); err != nil {
		return err
	}

	return nil
}
