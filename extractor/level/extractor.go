package level

import (
	"encoding/json"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/extractor"
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
	level.Transform
	UnkExtraRotation mgl32.Quat `json:"extra_rotation"`
}

type SimpleUnit struct {
	UnkHash00 string `json:"unk_hash_0"`
	UnkHash01 string `json:"unk_hash_1"`
	Path      string `json:"path"`
	level.Transform
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
