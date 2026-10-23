package materials

import (
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the standard fragment shader
var standardTextureNames = []string{
	"ao_map",
	"color_map",
	"emissive_map",
	"metallic_map",
	"normal_map",
	"roughness_map",
	"ibl_brdf_lut",
}

var standardShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/standard.frag",
}

func isStandardMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return mat.BaseMaterial == stingray.Sum("core/stingray_renderer/shader_import/standard")
}

type StandardMaterial struct {
	*BaseMaterial
}

func NewStandardMaterial(id stingray.FileID, name string, shown bool) *StandardMaterial {
	return &StandardMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *StandardMaterial) Generate() error {
	return mat.generate(
		standardShaderSources,
		0,
		append(baseUniforms, standardTextureNames...),
	)
}

var _ Material = &StandardMaterial{}

func (ml *MaterialManager) updateStandardMaterial(material *StandardMaterial, mat *material.Material) {
	ml.setupMaterialCommon(standardTextureNames, material, mat)
	ml.handleGlobalBlocks(material.BaseMaterial)
}
