package materials

import (
	"slices"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the rock fragment shader
var vistaTextureNames = []string{
	"base_normal_ao",
	"detail_mask_",
	"emissive_texture",
	"ground_albedo",
	"ground_nar",
	"triplanar_detail_albedo",
	"triplanar_detail_data",
	"vista_detail_albedo",
	"vista_detail_nar",
	"ibl_brdf_lut",
}

var vistaShaderSources = []string{
	"shaders/object.vert",
	"shaders/object.geom",
	"shaders/vista.frag",
}

var vistaMaterials = []stingray.Hash{
	stingray.Sum("content/art_shared/base_shaders/rock_default_vista_new"),
	stingray.Sum("content/art_shared/base_shaders/rock_default_vista_new_emissive"),
}

type VistaType uint32

const (
	VistaDefault VistaType = iota
	VistaEmissive
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=VistaType

func isVistaMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}

	return slices.Contains(vistaMaterials, mat.BaseMaterial)
}

type VistaMaterial struct {
	*BaseMaterial
}

func NewVistaMaterial(id stingray.FileID, name string, shown bool) *VistaMaterial {
	return &VistaMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *VistaMaterial) Generate() error {
	return mat.generate(
		vistaShaderSources,
		0,
		append(baseUniforms, vistaTextureNames...),
	)
}

var _ Material = &VistaMaterial{}

func (ml *MaterialManager) updateVistaMaterial(material *VistaMaterial, mat *material.Material) {
	ml.addAssetGradingLut(material.BaseMaterial)
	vistaType := VistaType(slices.Index(vistaMaterials, mat.BaseMaterial))
	ml.setupMaterialCommon(vistaTextureNames, material, mat)
	ml.handleGlobalBlocks(material.BaseMaterial)
	for _, block := range material.uniformBlocks {
		if block.name != "VistaBlock" {
			continue
		}
		block.set("vista_type", &vistaType)
		block.setDefault("vista_type", &vistaType)
		break
	}
}
