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
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/level"
)

type SimpleMetadata struct {
	Names []string                `json:"names"`
	Type  level.LevelMetadataType `json:"type"`
	Value any                     `json:"value"`
}

type SimpleMaterialOverride struct {
	Index     uint32           `json:"index"`
	Materials []SimpleMaterial `json:"materials"`
}

type SimpleMaterial struct {
	Slot string `json:"slot"`
	Path string `json:"path"`
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

type SimpleUnknownTransformedItem struct {
	Hash string `json:"hash"`
	stingray.Transform
	UnkFloats [6]float32 `json:"unk_floats"`
}

type SimpleExtraUnit struct {
	UnkHash1 string `json:"unk_hash_1"`
	Path     string `json:"path"`
	UnkHash2 string `json:"unk_hash_2"`
	stingray.Transform
	UnkFloats [3]float32 `json:"unk_floats"`
	UnkInt    uint32     `json:"unk_int"`
	UnkInt2   uint32     `json:"unk_int_2"`
}

type SimpleExtraPrefab struct {
	UnkHash1 string `json:"unk_hash_1"`
	Path     string `json:"path"`
	stingray.Transform
	UnkFloats [3]float32 `json:"unk_floats"`
	UnkInt    uint32     `json:"unk_int"`
}

type SimpleExtraUnitsContainer struct {
	UnkInt              uint32               `json:"unk_int"`
	UnkInt2             uint32               `json:"unk_int_2"`
	LevelName           string               `json:"level_name"`
	ExtraUnits          []SimpleExtraUnit    `json:"extra_units"`
	ExtraPrefabs        []SimpleExtraPrefab  `json:"extra_prefabs"`
	UnkIntList          []uint32             `json:"unk_int_list"`
	UnkFloatTwoIntsList []level.FloatTwoInts `json:"unk_float_two_ints_list"`
	UnkIntsAndFloatList []level.IntsAndFloat `json:"unk_ints_and_float"`
}

type SimpleHashIndexRange struct {
	Hash  string `json:"hash"`
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

type SimpleLevel struct {
	Name                 string                         `json:"name"`
	Metadata             map[int][]SimpleMetadata       `json:"metadata"`
	Prefabs              []SimplePrefab                 `json:"prefabs"`
	MaterialOverrides    []SimpleMaterialOverride       `json:"material_overrides"`
	Units                []SimpleUnit                   `json:"units"`
	UnkTransformedItems  []SimpleUnknownTransformedItem `json:"unk_transformed_item"`
	UnkExtraUnits        []SimpleExtraUnitsContainer    `json:"unk_extra_units"`
	UnitHashIndexRange   []SimpleHashIndexRange         `json:"unit_hash_index_range"`
	UnkHashIndexRange1   []SimpleHashIndexRange         `json:"unk_hash_index_range_1"`
	UnkHashIndexRange2   []SimpleHashIndexRange         `json:"unk_hash_index_range_2"`
	UnkHashIndexRange3   []SimpleHashIndexRange         `json:"unk_hash_index_range_3"`
	PrefabHashIndexRange []SimpleHashIndexRange         `json:"prefab_hash_index_range"`
	UnkHashIndexRange4   []SimpleHashIndexRange         `json:"unk_hash_index_range_4"`
	UnkHashIndexRange5   []SimpleHashIndexRange         `json:"unk_hash_index_range_5"`
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

	materialOverrides := make([]SimpleMaterialOverride, 0)
	for _, materialOverride := range levelData.MaterialOverrides {
		materials := make([]SimpleMaterial, 0)
		for _, material := range materialOverride.Materials {
			materials = append(materials, SimpleMaterial{
				Slot: ctx.LookupThinHash(material.Slot),
				Path: ctx.LookupHash(material.Path),
			})
		}
		materialOverrides = append(materialOverrides, SimpleMaterialOverride{
			Index:     materialOverride.Index,
			Materials: materials,
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
		Name:              ctx.LookupHash(levelData.Name),
		Metadata:          metadata,
		Prefabs:           prefabs,
		MaterialOverrides: materialOverrides,
		Units:             units,
	}

	outData.UnkTransformedItems = make([]SimpleUnknownTransformedItem, 0)
	for _, item := range levelData.UnkTransformedItems {
		outData.UnkTransformedItems = append(outData.UnkTransformedItems, SimpleUnknownTransformedItem{
			Hash:      ctx.LookupHash(item.Hash),
			Transform: item.Transform,
			UnkFloats: item.UnkFloats,
		})
	}

	outData.UnkExtraUnits = make([]SimpleExtraUnitsContainer, 0)
	for _, container := range levelData.UnkExtraUnitContainers {
		extraUnits := make([]SimpleExtraUnit, 0)
		for _, unit := range container.ExtraUnits {
			extraUnits = append(extraUnits, SimpleExtraUnit{
				UnkHash1:  ctx.LookupHash(unit.UnkHash1),
				Path:      ctx.LookupHash(unit.Path),
				UnkHash2:  ctx.LookupHash(unit.UnkHash2),
				Transform: unit.Transform,
				UnkFloats: unit.UnkFloats,
				UnkInt:    unit.UnkInt,
				UnkInt2:   unit.UnkInt2,
			})
		}
		extraPrefabs := make([]SimpleExtraPrefab, 0)
		for _, prefab := range container.ExtraPrefabs {
			extraPrefabs = append(extraPrefabs, SimpleExtraPrefab{
				UnkHash1:  ctx.LookupHash(prefab.UnkHash1),
				Path:      ctx.LookupHash(prefab.Path),
				Transform: prefab.Transform,
				UnkFloats: prefab.UnkFloats,
				UnkInt:    prefab.UnkInt,
			})
		}
		outData.UnkExtraUnits = append(outData.UnkExtraUnits, SimpleExtraUnitsContainer{
			UnkInt:              container.UnkInt,
			UnkInt2:             container.UnkInt2,
			LevelName:           ctx.LookupHash(container.LevelName),
			ExtraUnits:          extraUnits,
			ExtraPrefabs:        extraPrefabs,
			UnkIntList:          container.UnkIntList,
			UnkFloatTwoIntsList: container.UnkFloatTwoIntsList,
			UnkIntsAndFloatList: container.UnkIntsAndFloatList,
		})
	}

	if levelData.UnitHashIndexRange != nil {
		outData.UnitHashIndexRange = make([]SimpleHashIndexRange, 0)
		for _, hashIndexRange := range levelData.UnitHashIndexRange {
			outData.UnitHashIndexRange = append(outData.UnitHashIndexRange, SimpleHashIndexRange{
				Hash:  ctx.LookupThinHash(hashIndexRange.Hash),
				Start: hashIndexRange.Start,
				End:   hashIndexRange.End,
			})
		}
	}

	if levelData.UnkHashIndexRange1 != nil {
		outData.UnkHashIndexRange1 = make([]SimpleHashIndexRange, 0)
		for _, hashIndexRange := range levelData.UnkHashIndexRange1 {
			outData.UnkHashIndexRange1 = append(outData.UnkHashIndexRange1, SimpleHashIndexRange{
				Hash:  ctx.LookupThinHash(hashIndexRange.Hash),
				Start: hashIndexRange.Start,
				End:   hashIndexRange.End,
			})
		}
	}

	if levelData.UnkHashIndexRange2 != nil {
		outData.UnkHashIndexRange2 = make([]SimpleHashIndexRange, 0)
		for _, hashIndexRange := range levelData.UnkHashIndexRange2 {
			outData.UnkHashIndexRange2 = append(outData.UnkHashIndexRange2, SimpleHashIndexRange{
				Hash:  ctx.LookupThinHash(hashIndexRange.Hash),
				Start: hashIndexRange.Start,
				End:   hashIndexRange.End,
			})
		}
	}

	if levelData.UnkHashIndexRange3 != nil {
		outData.UnkHashIndexRange3 = make([]SimpleHashIndexRange, 0)
		for _, hashIndexRange := range levelData.UnkHashIndexRange3 {
			outData.UnkHashIndexRange3 = append(outData.UnkHashIndexRange3, SimpleHashIndexRange{
				Hash:  ctx.LookupThinHash(hashIndexRange.Hash),
				Start: hashIndexRange.Start,
				End:   hashIndexRange.End,
			})
		}
	}

	if levelData.PrefabHashIndexRange != nil {
		outData.PrefabHashIndexRange = make([]SimpleHashIndexRange, 0)
		for _, hashIndexRange := range levelData.PrefabHashIndexRange {
			outData.PrefabHashIndexRange = append(outData.PrefabHashIndexRange, SimpleHashIndexRange{
				Hash:  ctx.LookupThinHash(hashIndexRange.Hash),
				Start: hashIndexRange.Start,
				End:   hashIndexRange.End,
			})
		}
	}

	if levelData.UnkHashIndexRange4 != nil {
		outData.UnkHashIndexRange4 = make([]SimpleHashIndexRange, 0)
		for _, hashIndexRange := range levelData.UnkHashIndexRange4 {
			outData.UnkHashIndexRange4 = append(outData.UnkHashIndexRange4, SimpleHashIndexRange{
				Hash:  ctx.LookupThinHash(hashIndexRange.Hash),
				Start: hashIndexRange.Start,
				End:   hashIndexRange.End,
			})
		}
	}

	if levelData.UnkHashIndexRange5 != nil {
		outData.UnkHashIndexRange5 = make([]SimpleHashIndexRange, 0)
		for _, hashIndexRange := range levelData.UnkHashIndexRange5 {
			outData.UnkHashIndexRange5 = append(outData.UnkHashIndexRange5, SimpleHashIndexRange{
				Hash:  ctx.LookupThinHash(hashIndexRange.Hash),
				Start: hashIndexRange.Start,
				End:   hashIndexRange.End,
			})
		}
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

func GetLevelExtrasID(fileId stingray.FileID) string {
	return fileId.Name.String() + ".level"
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

	extras, ok := doc.Extras.(map[string]any)
	if !ok {
		extras = make(map[string]any)
	}

	levelIdx := uint32(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name:     ctx.LookupHash(ctx.FileID().Name) + ".level",
		Children: make([]uint32, 0),
		Extras: map[string]any{
			"node": levelIdx,
			"hash": GetLevelExtrasID(ctx.FileID()),
		},
	})
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, levelIdx)
	extras[GetLevelExtrasID(ctx.FileID())] = map[string]any{
		"parent": nil,
	}
	doc.Extras = extras

	for idx, prefab := range levelData.Prefabs {
		if ctx.FileID() == ctx.RootFileID() {
			percentComplete := 100 * float32(idx+1) / float32(len(levelData.Units)+len(levelData.Prefabs))
			ctx.Statusf("%.2f%% - %v.prefab", percentComplete, ctx.LookupHash(prefab.Path))
		}
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
			prefabMetadata["parent"] = levelIdx
			extras[extr_prefab.GetPrefabExtrasID(prefabId)] = prefabMetadata
			doc.Extras = extras
		}

		position, rotation, scale := prefab.ToGLTF()
		doc.Nodes[node].Translation = position
		doc.Nodes[node].Rotation = rotation
		doc.Nodes[node].Scale = scale

		doc.Nodes[levelIdx].Children = append(doc.Nodes[levelIdx].Children, node)
	}

	for idx, unit := range levelData.Units {
		if ctx.FileID() == ctx.RootFileID() {
			percentComplete := 100 * float32(idx+1+len(levelData.Prefabs)) / float32(len(levelData.Units)+len(levelData.Prefabs))
			ctx.Statusf("%.2f%% - %v.unit", percentComplete, ctx.LookupHash(unit.Path))
		}
		unitId := stingray.NewFileID(unit.Path, stingray.Sum("unit"))
		err := extr_prefab.AddOrDuplicateUnit(ctx.WithFileID(unitId), doc, imgOpts, &unit, levelIdx)
		if err != nil {
			return err
		}
	}

	extractor.ClearChildNodesFromScene(ctx, doc)

	formatIsBlend := cfg.Model.Format == "blend"
	if gltfDoc == nil && !formatIsBlend {
		ctx.Statusf("Creating glb file...")
		out, err := ctx.CreateFile(".level.glb")
		if err != nil {
			return err
		}
		enc := gltf.NewEncoder(out)
		if err := enc.Encode(doc); err != nil {
			return err
		}
	} else if gltfDoc == nil && formatIsBlend {
		ctx.Statusf("Creating blend file...")
		outPath, err := ctx.AllocateFile(".level.blend")
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
