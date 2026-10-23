package materials

import (
	"slices"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the standard fragment shader
var bugsTextureNames = []string{
	"albedo_iridescence",
	"clearcoat_roughness_ao_spec",
	"normal",
	"subsurface_opacity",
	"detail_normal_1",
	"texture_map_02",
	"emissive_color",
	"sclar_field_opacity",
	"ibl_brdf_lut",
}

var bugsShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/bugs.frag",
}

var bugsMaterials = []stingray.Hash{
	stingray.Sum("content/art_shared/base_shaders/m_bug_master_emissive"),
	stingray.Sum("content/art_shared/base_shaders/m_bug_master"),
	stingray.Sum("content/art_shared/base_shaders/m_bug_master_massive"),
	stingray.Hash{Value: 0x8f669f365f24594e},
}

func isBugsMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return slices.Contains(bugsMaterials, mat.BaseMaterial)
}

type BugsMaterial struct {
	*BaseMaterial
}

func NewBugsMaterial(id stingray.FileID, name string, shown bool) *BugsMaterial {
	return &BugsMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *BugsMaterial) Generate() error {
	return mat.generate(
		bugsShaderSources,
		0,
		append(baseUniforms, bugsTextureNames...),
	)
}

var _ Material = &BugsMaterial{}

func (ml *MaterialManager) updateBugsMaterial(material *BugsMaterial, mat *material.Material) {
	ml.setupMaterialCommon(bugsTextureNames, material, mat)
	ml.handleGlobalBlocks(material.BaseMaterial)
}
