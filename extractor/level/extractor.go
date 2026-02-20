package level

import (
	"encoding/json"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/extractor/blend_helper"
	extr_material "github.com/xypwn/filediver/extractor/material"
	extr_prefab "github.com/xypwn/filediver/extractor/prefab"
	extr_unit "github.com/xypwn/filediver/extractor/unit"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/level"
)

type SimpleMetadata struct {
	Names []string                `json:"names"`
	Type  level.LevelMetadataType `json:"type"`
	Value any                     `json:"value"`
}

type SimpleMaterial struct {
	Unk00 uint32 `json:"unk_int"`
	Count uint32 `json:"count"`
	Slot  string `json:"slot"`
	Path  string `json:"path"`
}

type SimplePrefab struct {
	UnkHash string `json:"unk_hash"`
	Path    string `json:"path"`
	stingray.Transform
	UnkExtraRotation mgl32.Vec4 `json:"extra_rotation"`
}

type SimpleUnit struct {
	UnkHash00 string `json:"unk_hash_0"`
	UnkHash01 string `json:"unk_hash_1"`
	Path      string `json:"path"`
	stingray.Transform
	UnkFloats [6]float32 `json:"unk_floats"`
}

type SimpleLevel struct {
	Name      string                   `json:"name"`
	Metadata  map[int][]SimpleMetadata `json:"metadata"`
	Prefabs   []SimplePrefab           `json:"prefabs"`
	Materials []SimpleMaterial         `json:"materials"`
	Units     []SimpleUnit             `json:"units"`
}

func ExtractLevelJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	levelData, err := level.LoadLevel(r)
	if err != nil {
		return err
	}
	metadata := make(map[int][]SimpleMetadata)
	for key, entries := range levelData.Metadata {
		metadata[key] = make([]SimpleMetadata, 0)
		for _, entry := range entries {
			varNames := make([]string, 0)
			for _, name := range entry.VariableNames {
				varNames = append(varNames, ctx.LookupThinHash(name))
			}
			simpleEntry := SimpleMetadata{
				Names: varNames,
				Type:  entry.Type,
			}
			switch entry.Type {
			case level.LevelMetadata_uint32:
				simpleEntry.Value = entry.ValueUint
			case level.LevelMetadata_float32:
				simpleEntry.Value = entry.ValueFloat
			case level.LevelMetadata_string:
				simpleEntry.Value = entry.ValueString
			}
			metadata[key] = append(metadata[key], simpleEntry)
		}
	}

	prefabs := make([]SimplePrefab, 0)
	for _, prefab := range levelData.Prefabs {
		prefabs = append(prefabs, SimplePrefab{
			UnkHash:          ctx.LookupHash(prefab.UnkHash00),
			Path:             ctx.LookupHash(prefab.Path),
			Transform:        prefab.Transform,
			UnkExtraRotation: prefab.UnkExtraRotation,
		})
	}

	materials := make([]SimpleMaterial, 0)
	for _, material := range levelData.Materials {
		materials = append(materials, SimpleMaterial{
			Unk00: material.Unk00,
			Count: material.Count,
			Slot:  ctx.LookupThinHash(material.Slot),
			Path:  ctx.LookupHash(material.Path),
		})
	}

	units := make([]SimpleUnit, 0)
	for _, unit := range levelData.Units {
		units = append(units, SimpleUnit{
			UnkHash00: ctx.LookupHash(unit.UnkHash00),
			UnkHash01: ctx.LookupHash(unit.UnkHash01),
			Path:      ctx.LookupHash(unit.Path),
			Transform: unit.Transform,
			UnkFloats: unit.UnkFloats,
		})
	}
	outData := SimpleLevel{
		Name:      ctx.LookupHash(levelData.Name),
		Metadata:  metadata,
		Prefabs:   prefabs,
		Materials: materials,
		Units:     units,
	}

	out, err := ctx.CreateFile(".level.json")
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

func ConvertOpts(ctx *extractor.Context, gltfDoc *gltf.Document) error {
	cfg := ctx.Config()
	if cfg.Level.Format == "json" {
		return ExtractLevelJSON(ctx)
	}
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	levelData, err := level.LoadLevel(r)
	if err != nil {
		return err
	}
	imgOpts, err := extr_material.GetImageOpts(ctx)
	if err != nil {
		return err
	}

	var doc *gltf.Document = gltfDoc
	if doc == nil {
		doc = gltf.NewDocument()
		doc.Asset.Generator = "https://github.com/xypwn/filediver"
		if ctx.BuildInfo() != nil {
			doc.Scenes[0].Extras = map[string]any{"Helldivers 2 Version": ctx.BuildInfo().Version}
		}
		doc.Samplers = append(doc.Samplers, &gltf.Sampler{
			MagFilter: gltf.MagLinear,
			MinFilter: gltf.MinLinear,
			WrapS:     gltf.WrapRepeat,
			WrapT:     gltf.WrapRepeat,
		})
	}

	levelIdx := uint32(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name:     ctx.LookupHash(ctx.FileID().Name) + ".level",
		Children: make([]uint32, 0),
	})

	for _, prefab := range levelData.Prefabs {
		ctx.Warnf("level: adding prefab %v", ctx.LookupHash(prefab.Path))
		prefabId := stingray.NewFileID(prefab.Path, stingray.Sum("prefab"))
		node, err := extr_prefab.AddPrefab(ctx.WithFileID(prefabId), doc, imgOpts)
		if err != nil {
			return err
		}
		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return fmt.Errorf("prefab export did not add extras? (should not happen)")
		}
		prefabMetadataIface, contains := extras[extr_prefab.GetPrefabExtrasID(prefabId)]
		if !contains {
			return fmt.Errorf("prefab export did not add metadata? (should not happen)")
		}
		prefabMetadata, ok := prefabMetadataIface.(map[string]any)
		if !ok {
			return fmt.Errorf("prefab metadata could not be converted? (should not happen)")
		}
		parentIface, contains := prefabMetadata["parent"]
		if !contains {
			return fmt.Errorf("prefab parent was not added? (should not happen)")
		}
		if _, ok := parentIface.(uint32); !ok {
			// parent was nil
			ctx.Warnf("Setting prefab parent to level")
			prefabMetadata["parent"] = levelIdx
			extras[extr_prefab.GetPrefabExtrasID(prefabId)] = prefabMetadata
			doc.Extras = extras
		}

		position, rotation, scale := prefab.ToGLTF()
		doc.Nodes[node].Translation = position
		doc.Nodes[node].Rotation = rotation
		doc.Nodes[node].Scale = scale

	}

	for _, unit := range levelData.Units {
		ctx.Warnf("level: adding unit %v", ctx.LookupHash(unit.Path))
		unitId := stingray.NewFileID(unit.Path, stingray.Sum("unit"))
		node, err := extr_prefab.AddOrDuplicateUnit(ctx.WithFileID(unitId), doc, imgOpts)
		if err != nil {
			return err
		}
		extras, ok := doc.Extras.(map[string]any)
		if !ok {
			return fmt.Errorf("unit export did not add extras? (should not happen)")
		}
		unitMetadataIface, contains := extras[extr_unit.GetUnitExtrasID(unitId)]
		if !contains {
			return fmt.Errorf("unit export did not add metadata? (should not happen)")
		}
		unitMetadata, ok := unitMetadataIface.(map[string]any)
		if !ok {
			return fmt.Errorf("unit metadata could not be converted? (should not happen)")
		}
		parentIface, contains := unitMetadata["parent"]
		if !contains {
			return fmt.Errorf("unit parent was not added? (should not happen)")
		}
		if _, ok := parentIface.(uint32); !ok {
			// parent was nil
			ctx.Warnf("Setting unit parent to level")
			unitMetadata["parent"] = levelIdx
			extras[extr_unit.GetUnitExtrasID(unitId)] = unitMetadata
			doc.Extras = extras
		}
		position, rotation, scale := unit.ToGLTF()
		doc.Nodes[node].Translation = position
		doc.Nodes[node].Rotation = rotation
		doc.Nodes[node].Scale = scale
	}

	extractor.ClearChildNodesFromScene(ctx, doc)

	formatIsBlend := cfg.Model.Format == "blend"
	if gltfDoc == nil && !formatIsBlend {
		out, err := ctx.CreateFile(".level.glb")
		if err != nil {
			return err
		}
		enc := gltf.NewEncoder(out)
		if err := enc.Encode(doc); err != nil {
			return err
		}
	} else if gltfDoc == nil && formatIsBlend {
		outPath, err := ctx.AllocateFile(".level.blend")
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
		return ConvertOpts(ctx, currDoc)
	}
}
