package materials

import (
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the standard fragment shader
var signTextureNames = []string{
	"distortion_map",
	"dirt_map",
	"light_bleed_map",
	"input_image",
	"normal_map",
	"roughness_map",
	"ibl_brdf_lut",
}

func isSignMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return mat.BaseMaterial == stingray.Hash{Value: 0x2f8279894fea04c5}
}

type SignMaterial struct {
	*BaseMaterial
}

func NewSignMaterial(id stingray.FileID, name string, shown bool) *SignMaterial {
	return &SignMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *SignMaterial) Generate() error {
	return mat.generate(
		[]string{
			"shaders/object.vert",
			"shaders/object.geom",
			"shaders/sign.frag",
		},
		0,
		append(baseUniforms, standardTextureNames...),
	)
}

var _ Material = &SignMaterial{}

func (ml *MaterialManager) updateSignMaterial(material *SignMaterial, mat *material.Material) {
	ml.setupMaterialCommon(signTextureNames, material, mat)
}
