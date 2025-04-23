package geometry_group

import (
	"fmt"
	"io"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/geometry"
	extr_material "github.com/xypwn/filediver/extractor/material"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
	geometrygroup "github.com/xypwn/filediver/stingray/unit/geometry_group"
)

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

	geoGroup, err := geometrygroup.LoadGeometryGroup(fMain)
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

	for unitHash, meshInfo := range geoGroup.MeshInfos {
		unitRes, exists := ctx.GetResource(unitHash, stingray.Sum64([]byte("unit")))
		if !exists {
			return fmt.Errorf("%v.unit does not exist", unitHash.String())
		}
		f, err := unitRes.Open(ctx.Ctx(), stingray.DataMain)
		if err != nil {
			return err
		}
		defer f.Close()

		unitInfo, err := unit.LoadInfo(f)
		if err != nil {
			return err
		}

		// Load materials
		materialIdxs, err := extr_unit.AddMaterials(ctx, doc, imgOpts, unitInfo, nil)
		if err != nil {
			return err
		}

		bonesEnabled := ctx.Config()["no_bones"] != "true"

		var skin *uint32 = nil
		var parent *uint32 = nil
		if bonesEnabled && len(unitInfo.Bones) > 2 {
			skin = gltf.Index(extr_unit.AddSkeleton(ctx, doc, unitInfo, unitHash, nil))
			parent = doc.Skins[*skin].Skeleton
		} else {
			parent = gltf.Index(uint32(len(doc.Nodes)))
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name: unitHash.String(),
			})
			doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, *parent)
		}

		meshInfos := make([]geometry.MeshInfo, 0)
		for _, header := range meshInfo.MeshHeaders {
			meshInfos = append(meshInfos, geometry.MeshInfo{
				Groups:          header.Groups,
				Materials:       header.Materials,
				MeshLayoutIndex: header.MeshLayoutIndex,
			})
		}

		var meshNodes []uint32 = make([]uint32, 0)
		err = geometry.LoadGLTF(ctx, fGPU, doc, unitHash, meshInfos, meshInfo.Bones, geoGroup.MeshLayouts, unitInfo, &meshNodes, materialIdxs, *parent, skin)
		if err != nil {
			return err
		}
		extr_unit.AddPrefabMetadata(ctx, doc, parent, skin, meshNodes, nil)
	}

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

func Convert(currDoc *gltf.Document) func(ctx extractor.Context) error {
	return func(ctx extractor.Context) error {
		opts, err := extr_unit.GetImgOpts(ctx)
		if err != nil {
			return err
		}
		return ConvertOpts(ctx, opts, currDoc)
	}
}
