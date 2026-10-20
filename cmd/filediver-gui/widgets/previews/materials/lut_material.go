package materials

import (
	"maps"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the lut fragment shader
var lutTextureNames = []string{
	"decal_sheet",
	"composite_array",
	"customization_camo_tiler_array",
	"customization_material_detail_tiler_array",
	"pattern_lut",
	"base_data",
	"material_lut",
	"pattern_masks_array",
	"id_masks_array",
	"ibl_brdf_lut",
}

var lutShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/lut.frag",
}

type LutMaterial struct {
	*BaseMaterial
}

func isLutMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}
	_, containsIdMasks := mat.Textures[stingray.Sum("id_masks_array").Thin()]
	_, containsMaterialLut := mat.Textures[stingray.Sum("material_lut").Thin()]
	return containsIdMasks && containsMaterialLut
}

func NewLutMaterial(id stingray.FileID, name string, shown bool) *LutMaterial {
	return &LutMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *LutMaterial) Generate() error {
	return mat.generate(
		lutShaderSources,
		0,
		append(baseUniforms, lutTextureNames...),
	)
}

func (ml *MaterialManager) updateLutMaterial(material *LutMaterial, mat *material.Material) {
	if ml.detailerLoadState == DetailerLoaded {
		ml.uploadMaterialDetailer(ml.detailerTextureData)
	}

	originalTextures := maps.Clone(mat.Textures)

	slot := "customization_material_detail_tiler_array"
	materialDetailerHash := stingray.Sum("content/art_shared/textures/customization/material_library/detail_tilers/customization_detail_tiler_array")
	mat.Textures[stingray.Sum(slot).Thin()] = materialDetailerHash

	if _, contains := mat.Textures[stingray.Sum("pattern_lut").Thin()]; !contains {
		mat.Textures[stingray.Sum("pattern_lut").Thin()] = stingray.Hash{Value: 0xcf0cc31b981786c9}
	}

	transparentLut := int32(0)
	if mat.BaseMaterial.Value == 0x8e7efcdf45ac19cf {
		transparentLut = 1
	}

	ml.setupMaterialCommon(lutTextureNames, material, mat)

	for _, block := range material.uniformBlocks {
		if _, contains := block.uniformOffsets["seed"]; contains {
			block.set("seed", &seed)
			block.setDefault("seed", &seed)
		}
		if _, contains := block.uniformOffsets["use_decals"]; contains {
			val, contains := originalTextures[stingray.Sum("decal_sheet").Thin()]
			hasValue := val.Value != 0x0
			useDecals := uint32(0)
			if contains || hasValue {
				useDecals = 1
			}
			block.set("use_decals", &useDecals)
			block.setDefault("use_decals", &useDecals)
		}
		if _, contains := block.uniformOffsets["transparent_lut"]; contains {
			block.set("transparent_lut", &transparentLut)
			block.setDefault("transparent_lut", &transparentLut)
		}
	}

	ml.handleGlobalBlocks(material.BaseMaterial)
}

var _ Material = &LutMaterial{}
