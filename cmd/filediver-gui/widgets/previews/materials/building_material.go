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
	"decal_sheet",
	"Detail_Data",
	"NAC",
	"surface_data_array",
	"ibl_brdf_lut",
}

var buildingShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/building.frag",
}

var buildingBaseMaterialNames = []stingray.Hash{
	stingray.Sum("content/art_shared/base_shaders/m_colonies_master"),
	stingray.Sum("content/art_shared/base_shaders/m_colonies_master_decal"),
	stingray.Sum("content/art_shared/base_shaders/m_colonies_master_decal_only"),
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
		buildingShaderSources,
		0,
		append(baseUniforms, buildingTextureNames...),
	)
}

var _ Material = &BuildingMaterial{}

func (ml *MaterialManager) updateBuildingMaterial(material *BuildingMaterial, mat *material.Material) {
	mat.Settings[stingray.Sum("fd_decals_enabled").Thin()] = []float32{0.0}
	mat.Settings[stingray.Sum("fd_decal_only").Thin()] = []float32{0.0}
	for slot := range mat.Textures {
		if ml.lookupThinHash(slot) == "decal_sheet" {
			mat.Settings[stingray.Sum("fd_decals_enabled").Thin()] = []float32{1.0}
			break
		}
	}

	if mat.BaseMaterial == stingray.Sum("content/art_shared/base_shaders/m_colonies_master_decal_only") {
		mat.Settings[stingray.Sum("fd_decal_only").Thin()] = []float32{1.0}
	}
	ml.setupMaterialCommon(buildingTextureNames, material, mat)

	for idx, texture := range material.textures {
		if texture.slot != "texture_lut" {
			continue
		}
		material.textures[idx].filter = gl.NEAREST
	}

	ml.handleGlobalBlocks(material.BaseMaterial)
}
