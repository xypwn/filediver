package materials

import (
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

var emissiveShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/emissive.frag",
}

func isEmissiveMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return mat.BaseMaterial == stingray.Hash{Value: 0xef050126eedc3e05}
}

type EmissiveMaterial struct {
	*BaseMaterial
}

func NewEmissiveMaterial(id stingray.FileID, name string, shown bool) *EmissiveMaterial {
	return &EmissiveMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *EmissiveMaterial) Generate() error {
	return mat.generate(
		emissiveShaderSources,
		0,
		baseUniforms,
	)
}

var _ Material = &EmissiveMaterial{}

func (ml *MaterialManager) updateEmissiveMaterial(material *EmissiveMaterial, mat *material.Material) {
	material.InitSettings(mat.Settings, ml.lookupThinHash)
}
