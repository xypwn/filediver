package prefab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/blend_helper"
	extr_material "github.com/xypwn/filediver/extractor/material"
	extr_speedtree "github.com/xypwn/filediver/extractor/speedtree"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/prefab"
)

type SimpleHeader struct {
	UnkHash    string `json:"unk_hash"`
	PrefabHash string `json:"name"`
}

type SimpleUnit struct {
	UnkHash0 string `json:"unk_hash_0"`
	Path     string `json:"path"`
	UnkHash1 string `json:"unk_hash_1"`
	UnkHash2 string `json:"unk_hash_2"`
	stingray.Transform
	UnkRotation mgl32.Vec4 `json:"unk_rotation"`
	Index       uint32     `json:"index"`
}

type SimpleNestedPrefab struct {
	UnkInt  uint32 `json:"unk_int"`
	UnkHash string `json:"unk_hash"`
	Path    string `json:"path"`
	stingray.Transform
	UnkFloats mgl32.Vec3 `json:"unk_floats"`
}

type SimplePrefab struct {
	Name    string               `json:"name"`
	Units   []SimpleUnit         `json:"units"`
	Prefabs []SimpleNestedPrefab `json:"prefabs"`
}

func ExtractPrefabJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	prefabData, err := prefab.Load(r)
	if err != nil {
		return err
	}
	units := make([]SimpleUnit, 0)
	for _, unit := range prefabData.Units {
		units = append(units, SimpleUnit{
			UnkHash0:    ctx.LookupHash(stingray.Hash{Value: unit.Unk00}),
			Path:        ctx.LookupHash(unit.Path),
			UnkHash1:    ctx.LookupHash(stingray.Hash{Value: unit.Unk01}),
			UnkHash2:    ctx.LookupHash(stingray.Hash{Value: unit.Unk02}),
			Transform:   unit.Transform,
			UnkRotation: unit.UnkFloats,
			Index:       unit.Index,
		})
	}
	prefabs := make([]SimpleNestedPrefab, 0)
	for _, prefab := range prefabData.NestedPrefabs {
		prefabs = append(prefabs, SimpleNestedPrefab{
			UnkInt:    prefab.UnkInt,
			UnkHash:   ctx.LookupHash(prefab.UnkHash),
			Path:      ctx.LookupHash(prefab.Path),
			Transform: prefab.Transform,
			UnkFloats: prefab.UnkFloats,
		})
	}
	outData := SimplePrefab{
		Name:    ctx.LookupHash(prefabData.NameHash),
		Prefabs: prefabs,
		Units:   units,
	}

	out, err := ctx.CreateFile(".prefab.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(outData); err != nil {
		return err
	}
	return nil
}

func AddOrDuplicateModel(ctx *extractor.Context, doc *gltf.Document, imgOpts *extr_material.ImageOptions, obj extractor.Object, parentNode uint32) error {
	if ctxErr := ctx.Ctx().Err(); errors.Is(ctxErr, context.Canceled) {
		return ctxErr
	}
	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}

	var extrasId string
	switch ctx.FileID().Type {
	case stingray.Sum("unit"):
		extrasId = extr_unit.GetUnitExtrasID(ctx.FileID())
	case stingray.Sum("speedtree"):
		extrasId = extr_speedtree.GetSpeedtreeExtrasID(ctx.FileID())
	default:
		return fmt.Errorf("prefab AddOrDuplicateModel: request for extrasId for %v", ctx.LookupHash(ctx.FileID().Type))
	}
	if _, contains := extras[extrasId]; !contains {
		// We have not already loaded this model, load it now
		var err error
		switch ctx.FileID().Type {
		case stingray.Sum("unit"):
			err = extr_unit.ConvertOpts(ctx, imgOpts, doc)
		case stingray.Sum("speedtree"):
			err = extr_speedtree.ConvertOpts(ctx, imgOpts, doc)
		default:
			return fmt.Errorf("prefab AddOrDuplicateModel: trying to load %v filetype", ctx.LookupHash(ctx.FileID().Type))
		}

		if err != nil {
			return err
		}

		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return fmt.Errorf("could not resolve doc extras? (this should not happen)")
		}
		modelMetadataIface, ok := extras[extrasId]
		if !ok {
			return fmt.Errorf("could not resolve %s.%s gltf metadata? (this should not happen)", ctx.FileID().Name.String(), ctx.LookupHash(ctx.FileID().Type))
		}
		modelMetadata, ok := modelMetadataIface.(map[string]any)
		if !ok {
			return fmt.Errorf("could not cast %s.%s gltf metadata? (this should not happen)", ctx.FileID().Name.String(), ctx.LookupHash(ctx.FileID().Type))
		}
		root, ok := modelMetadata["root"].(uint32)
		if !ok {
			return fmt.Errorf("%s.%s did not have a root set? (this should not happen)", ctx.FileID().Name.String(), ctx.LookupHash(ctx.FileID().Type))
		}

		translation, rotation, scale := obj.ToGLTF()
		doc.Nodes[root].Translation = translation
		doc.Nodes[root].Rotation = rotation
		doc.Nodes[root].Scale = scale

		doc.Nodes[parentNode].Children = append(doc.Nodes[parentNode].Children, root)

		modelMetadata["parent"] = parentNode
		extras[extrasId] = modelMetadata
		doc.Extras = extras
	} else {
		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return fmt.Errorf("could not resolve %s.%s gltf metadata? (this should not happen)", ctx.FileID().Name.String(), ctx.LookupHash(ctx.FileID().Type))
		}
		modelMetadataIface, ok := extras[extrasId]
		if !ok {
			return fmt.Errorf("could not resolve %s.%s gltf metadata? (this should not happen)", ctx.FileID().Name.String(), ctx.LookupHash(ctx.FileID().Type))
		}
		modelMetadata, ok := modelMetadataIface.(map[string]any)
		if !ok {
			return fmt.Errorf("could not cast %s.%s gltf metadata? (this should not happen)", ctx.FileID().Name.String(), ctx.LookupHash(ctx.FileID().Type))
		}
		var instanceList []map[string]any
		instanceListIface, contains := modelMetadata["filediver_instances"]
		if !contains {
			instanceList = make([]map[string]any, 0)
		} else if instanceList, ok = instanceListIface.([]map[string]any); !ok {
			return fmt.Errorf("failed to cast instance list")
		}

		translation, rotation, scale := obj.ToGLTF()
		instanceList = append(instanceList, map[string]any{
			"parent":      parentNode,
			"translation": translation,
			"rotation":    rotation,
			"scale":       scale,
		})
		modelMetadata["filediver_instances"] = instanceList
		extras[extrasId] = modelMetadata
		doc.Extras = extras
	}
	return nil
}

func GetPrefabExtrasID(fileId stingray.FileID) string {
	return fileId.Name.String() + ".prefab"
}

func AddPrefab(ctx *extractor.Context, doc *gltf.Document, imgOpts *extr_material.ImageOptions) (uint32, error) {
	if ctxErr := ctx.Ctx().Err(); errors.Is(ctxErr, context.Canceled) {
		return 0, ctxErr
	}
	fMain, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return 0, err
	}
	prefabData, err := prefab.Load(fMain)
	if err != nil {
		return 0, err
	}

	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}

	prefabRoot := uint32(len(doc.Nodes))
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, prefabRoot)
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name:     ctx.LookupHash(ctx.FileID().Name) + ".prefab",
		Children: make([]uint32, 0),
		Extras: map[string]any{
			"hash": GetPrefabExtrasID(ctx.FileID()),
			"node": prefabRoot,
		},
	})

	extras[GetPrefabExtrasID(ctx.FileID())] = map[string]any{
		"node":   prefabRoot,
		"parent": nil,
	}
	doc.Extras = extras

	for idx, object := range prefabData.Units {
		if ctx.FileID() == ctx.RootFileID() {
			percentComplete := 100 * float32(idx+1) / float32(len(prefabData.Units)+len(prefabData.NestedPrefabs))
			ctx.Statusf("%.2f%% - %v.unit", percentComplete, ctx.LookupHash(object.Path))
		}
		unitId := stingray.NewFileID(object.Unit(), stingray.Sum("unit"))
		err := AddOrDuplicateModel(ctx.WithFileID(unitId), doc, imgOpts, &object, prefabRoot)
		if err != nil {
			return 0, err
		}
	}

	for idx, nested := range prefabData.NestedPrefabs {
		if ctx.FileID() == ctx.RootFileID() {
			percentComplete := 100 * float32(idx+1+len(prefabData.Units)) / float32(len(prefabData.Units)+len(prefabData.NestedPrefabs))
			ctx.Statusf("%.2f%% - %v.prefab", percentComplete, ctx.LookupHash(nested.Path))
		}
		nestedId := stingray.NewFileID(nested.Path, stingray.Sum("prefab"))
		root, err := AddPrefab(ctx.WithFileID(nestedId), doc, imgOpts)
		if err != nil {
			return 0, fmt.Errorf("extracting prefab %v: %v", ctx.LookupHash(nested.Path), err)
		}
		doc.Nodes[prefabRoot].Children = append(doc.Nodes[prefabRoot].Children, root)
		translation, rotation, scale := nested.ToGLTF()
		doc.Nodes[root].Translation = translation
		doc.Nodes[root].Rotation = rotation
		doc.Nodes[root].Scale = scale

		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.prefab gltf metadata? (this should not happen)", nestedId.Name.String())
		}
		prefabMetadataIface, ok := extras[GetPrefabExtrasID(nestedId)]
		if !ok {
			return 0, fmt.Errorf("could not resolve %s.prefab gltf metadata? (this should not happen)", nestedId.Name.String())
		}
		prefabMetadata, ok := prefabMetadataIface.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("could not cast %s.prefab gltf metadata? (this should not happen)", nestedId.Name.String())
		}
		prefabMetadata["parent"] = prefabRoot
		extras[GetPrefabExtrasID(nestedId)] = prefabMetadata
		doc.Extras = extras
	}

	return prefabRoot, nil
}

func ConvertOpts(ctx *extractor.Context, gltfDoc *gltf.Document) error {
	cfg := ctx.Config()
	if cfg.Prefab.Format == "json" {
		return ExtractPrefabJSON(ctx)
	}
	imgOpts, err := extr_material.GetImageOpts(ctx)
	if err != nil {
		return err
	}

	doc := extractor.GetDocument(ctx, gltfDoc)

	if _, err := AddPrefab(ctx, doc, imgOpts); err != nil {
		return err
	}

	extractor.ClearChildNodesFromScene(ctx, doc)

	formatIsBlend := cfg.Model.Format == "blend"
	if gltfDoc == nil && !formatIsBlend {
		ctx.Statusf("Creating glb file...")
		out, err := ctx.CreateFile(".prefab.glb")
		if err != nil {
			return err
		}
		enc := gltf.NewEncoder(out)
		if err := enc.Encode(doc); err != nil {
			return err
		}
	} else if gltfDoc == nil && formatIsBlend {
		ctx.Statusf("Creating blend file...")
		outPath, err := ctx.AllocateFile(".prefab.blend")
		if err != nil {
			return err
		}
		err = blend_helper.ExportBlend(doc, outPath, ctx.Runner())
		if err != nil {
			return err
		}
	}
	ctx.Statusf("Done")
	return nil
}

func Convert(currDoc *gltf.Document) func(ctx *extractor.Context) error {
	return func(ctx *extractor.Context) error {
		return ConvertOpts(ctx, currDoc)
	}
}
