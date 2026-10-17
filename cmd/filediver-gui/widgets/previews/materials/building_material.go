package materials

import (
	"slices"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the standard fragment shader
var buildingTextureNames = []string{
	"texture_lut",
	"Detail_Data",
	"NAC",
	"surface_data_array",
	"ibl_brdf_lut",
}

var buildingBaseMaterialNames = []stingray.Hash{
	stingray.Sum("content/art_shared/base_shaders/m_colonies_master"),
	stingray.Sum("content/art_shared/base_shaders/m_colonies_medium_small"),
}

func isBuildingMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return slices.Contains(buildingBaseMaterialNames, mat.BaseMaterial)
}

type BuildingMaterial struct {
	*BaseMaterial
}

func NewBuildingMaterial(id stingray.FileID, name string, shown bool) *BuildingMaterial {
	return &BuildingMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *BuildingMaterial) Generate() error {
	return mat.generate(
		[]string{
			"shaders/object.vert",
			"shaders/object.geom",
			"shaders/building.frag",
		},
		0,
		append(baseUniforms, buildingTextureNames...),
	)
}

var _ Material = &BuildingMaterial{}

func (ml *MaterialManager) updateBuildingMaterial(material *BuildingMaterial, mat *material.Material) {
	ml.setupMaterialCommon(buildingTextureNames, material, mat)

	for idx, texture := range material.textures {
		if texture.slot != "texture_lut" {
			continue
		}
		material.textures[idx].filter = gl.NEAREST
	}
}
