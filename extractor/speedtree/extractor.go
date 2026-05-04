package speedtree

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/blend_helper"
	"github.com/xypwn/filediver/extractor/geometry"
	extr_material "github.com/xypwn/filediver/extractor/material"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/speedtree"
	"github.com/xypwn/filediver/stingray/unit"
	"github.com/xypwn/filediver/stingray/unit/material"
)

func ExtractSpeedTreeJson(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}

	tree, err := speedtree.LoadSpeedTree(r)
	if err != nil && err != speedtree.SDKParseError {
		return err
	}

	out, err := ctx.CreateFile(".speedtree.json")
	if err != nil {
		return err
	}
	defer out.Close()

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(tree)

	return err
}

var speedtreeVertex16 unit.MeshLayout = unit.MeshLayout{
	NumItems: 4,
	Items: [16]unit.MeshLayoutItem{
		{
			Type:   unit.ItemPosition,
			Format: unit.FormatVec3F16,
			Layer:  0,
		},
		{
			Type:   unit.ItemSpeedTreeNormalX,
			Format: unit.FormatF16,
			Layer:  0,
		},
		{
			Type:   unit.ItemUVCoords,
			Format: unit.FormatVec2F16,
			Layer:  0,
		},
		{
			Type:   unit.ItemSpeedTreeNormalYZ,
			Format: unit.FormatVec2F16,
			Layer:  0,
		},
	},
}

var speedtreeVertex20 unit.MeshLayout = unit.MeshLayout{
	NumItems: 8,
	Items: [16]unit.MeshLayoutItem{
		{
			Type:   unit.ItemPosition,
			Format: unit.FormatVec3F16,
			Layer:  0,
		},
		{
			Type:   unit.ItemSpeedTreeU,
			Format: unit.FormatF16,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatVec3F16,
			Layer:  0,
		},
		{
			Type:   unit.ItemSpeedTreeV,
			Format: unit.FormatF16,
			Layer:  0,
		},
		{
			Type:   unit.ItemNormal,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemTangent,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
	},
}

var speedtreeVertex24 unit.MeshLayout = unit.MeshLayout{
	NumItems: 12,
	Items: [16]unit.MeshLayoutItem{
		{
			Type:   unit.ItemPosition,
			Format: unit.FormatVec3F16,
			Layer:  0,
		},
		{
			Type:   unit.ItemSpeedTreeU,
			Format: unit.FormatF16,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatVec3F16,
			Layer:  0,
		},
		{
			Type:   unit.ItemSpeedTreeV,
			Format: unit.FormatF16,
			Layer:  0,
		},
		{
			Type:   unit.ItemNormal,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemTangent,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
		{
			Type:   unit.ItemIgnore,
			Format: unit.FormatS8,
			Layer:  0,
		},
	},
}

func ConvertOpts(ctx *extractor.Context, imgOpts *extr_material.ImageOptions, gltfDoc *gltf.Document) error {
	cfg := ctx.Config()
	if cfg.SpeedTree.Format == "json" {
		return ExtractSpeedTreeJson(ctx)
	}
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

	treeInfo, err := speedtree.LoadSpeedTree(fMain)
	if err != nil {
		return err
	}

	doc := extractor.GetDocument(ctx, gltfDoc)

	if err := geometry.InitFibonacciLut(ctx); err != nil {
		return err
	}

	materialMap := make(map[uint64]uint32)
	for _, mat := range treeInfo.StingrayMaterials {
		matR, err := ctx.Open(stingray.NewFileID(mat.Path, stingray.Sum("material")), stingray.DataMain)
		if err == stingray.ErrFileNotExist {
			return fmt.Errorf("referenced material resource %v doesn't exist", mat.Path)
		}
		if err != nil {
			return err
		}
		matInfo, err := material.LoadMain(matR)
		if err != nil {
			return err
		}

		matIdx, err := extr_material.AddMaterial(ctx, matInfo, doc, imgOpts, treeInfo.SDKMaterials[mat.Index].Name+fmt.Sprintf(" %v", ctx.LookupHash(mat.Path)), nil)
		if err != nil {
			return err
		}
		materialMap[mat.Index] = matIdx
	}

	//layoutAttributes := make(map[int]map[string]*gltf.Accessor)
	groupAttributes := make(map[int]gltf.Attribute)
	//vertexDefToBuffer := make(map[int]uint32)
	for idx, vertexDef := range treeInfo.VertexDefinitions {
		var layout unit.MeshLayout
		switch vertexDef.Stride {
		case 24:
			layout = speedtreeVertex24
		case 20:
			layout = speedtreeVertex20
		case 16:
			layout = speedtreeVertex16
		default:
			return fmt.Errorf("Unknown speed tree vertex stride: %v", vertexDef.Stride)
		}
		layout.VertexOffset = vertexDef.Offset
		layout.NumVertices = vertexDef.Count
		data, accessorInfo, err := geometry.ConvertVertices(fGPU, layout)
		if err != nil {
			return err
		}
		vertexBuffer, err := geometry.AddMeshLayoutVertexBuffer(doc, data, accessorInfo)
		if err != nil {
			return err
		}
		//vertexDefToBuffer[idx] = vertexBuffer
		layoutAttributes := geometry.CreateAttributes(doc, layout, accessorInfo)

		attr, err := geometry.AddGroupAttributes(doc, unit.MeshGroup{NumVertices: vertexDef.Count}, layoutAttributes, vertexBuffer, vertexDef.Count-1)
		if err != nil {
			return err
		}
		gltfMin := stingray.ToGLTFMatrix.Mul4x1(treeInfo.Extents[0].Vec4(1.0)).Vec3()
		gltfMax := stingray.ToGLTFMatrix.Mul4x1(treeInfo.Extents[1].Vec4(1.0)).Vec3()
		doc.Accessors[attr[gltf.POSITION]].Min = gltfMin[:]
		doc.Accessors[attr[gltf.POSITION]].Max = gltfMax[:]

		offset := doc.Accessors[attr[gltf.POSITION]].ByteOffset + doc.BufferViews[vertexBuffer].ByteOffset
		stride := doc.BufferViews[vertexBuffer].ByteStride
		buffer := doc.Buffers[doc.BufferViews[vertexBuffer].Buffer]
		if err := geometry.TransformVertices(buffer, offset, stride, 0, doc.Accessors[attr[gltf.POSITION]].Count, stingray.ToGLTFMatrix); err != nil {
			return err
		}

		normalsAttr, ok := attr[gltf.NORMAL]
		if ok {
			offset = doc.Accessors[normalsAttr].ByteOffset + doc.BufferViews[vertexBuffer].ByteOffset
			if err := geometry.TransformVertices(buffer, offset, stride, 0, doc.Accessors[normalsAttr].Count, stingray.ToGLTFMatrix); err != nil {
				return err
			}
		}

		tangentsAttr, ok := attr[gltf.TANGENT]
		if ok {
			offset = doc.Accessors[tangentsAttr].ByteOffset + doc.BufferViews[vertexBuffer].ByteOffset
			if err := geometry.TransformVertices(buffer, offset, stride, 0, doc.Accessors[tangentsAttr].Count, stingray.ToGLTFMatrix); err != nil {
				return err
			}
		}
		groupAttributes[idx] = attr
	}

	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: ctx.LookupHash(ctx.FileID().Name) + ".speedtree",
	})
	parent := uint32(len(doc.Nodes) - 1)
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, parent)
	indexDefToAccessor := make(map[int]*gltf.Accessor)
	for idx, indexDef := range treeInfo.IndexDefinitions {
		layout := unit.MeshLayout{
			IndicesSize: indexDef.Count * indexDef.Stride,
			IndexOffset: indexDef.Offset,
			NumIndices:  indexDef.Count,
		}
		indexAccessor, err := geometry.LoadMeshLayoutIndices(fGPU, doc, layout)
		if err != nil {
			return err
		}
		indexDefToAccessor[idx] = indexAccessor

		primitives := make([]*gltf.Primitive, 0)
		for i := range indexDef.MeshCount {
			indexOffset := treeInfo.MeshDefinitions[indexDef.MeshOffset+i].IndexOffset
			indexCount := treeInfo.MeshDefinitions[indexDef.MeshOffset+i].IndexCount
			materialId := treeInfo.MeshDefinitions[indexDef.MeshOffset+i].Material
			doc.Accessors = append(doc.Accessors, &gltf.Accessor{
				BufferView:    indexAccessor.BufferView,
				ByteOffset:    indexOffset * indexAccessor.ComponentType.ByteSize(),
				ComponentType: indexAccessor.ComponentType,
				Type:          gltf.AccessorScalar,
				Count:         indexCount,
			})
			meshIndices := gltf.Index(uint32(len(doc.Accessors)) - 1)

			primitives = append(primitives, &gltf.Primitive{
				Attributes: groupAttributes[int(indexDef.VertexDef)],
				Indices:    meshIndices,
				Material:   gltf.Index(materialMap[uint64(materialId)]),
			})
		}
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{
			Primitives: primitives,
		})
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: fmt.Sprintf("Speedtree LOD%v", idx),
			Mesh: gltf.Index(uint32(len(doc.Meshes) - 1)),
		})
		doc.Nodes[parent].Children = append(doc.Nodes[parent].Children, uint32(len(doc.Nodes)-1))
	}

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

func Convert(currDoc *gltf.Document) func(ctx *extractor.Context) error {
	return func(ctx *extractor.Context) error {
		opts, err := extr_material.GetImageOpts(ctx)
		if err != nil {
			return err
		}
		return ConvertOpts(ctx, opts, currDoc)
	}
}
