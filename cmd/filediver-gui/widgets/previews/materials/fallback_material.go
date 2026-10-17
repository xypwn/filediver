package materials

import (
	"slices"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

type FallbackMaterial struct {
	*BaseMaterial
}

func NewFallbackMaterial(id stingray.FileID, name string, shown bool) *FallbackMaterial {
	return &FallbackMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *FallbackMaterial) Generate() error {
	return mat.generate(
		[]string{
			"shaders/object.vert",
			"shaders/object.geom",
			"shaders/object.frag",
		},
		0,
		append(baseUniforms, "texAlbedo", "texNormal", "shouldReconstructNormalZ"),
	)
}

var _ Material = &FallbackMaterial{}

func (ml *MaterialManager) updateFallbackMaterial(material *FallbackMaterial, mat *material.Material) {
	albedoRemoveAlphaTextures := []string{
		"covering_albedo", "input_image", "albedo",
	}
	albedoTextures := []string{
		"color_roughness", "color_specular_b", "albedo_iridescence",
		"covering_albedo", "input_image", "albedo",
	}
	normalReconstructTextures := []string{
		"normal", "normals", "normal_map", "covering_normal", "nac", "base_data", "nar", "normal_ao_roughness", "normal_xy_ao_rough_map", "normal_xy_roughness_opacity",
	}
	normalTextures := []string{
		"normal_specular_ao",
		"normal", "normals", "normal_map", "covering_normal", "nac", "base_data", "nar", "normal_ao_roughness", "normal_xy_ao_rough_map", "normal_xy_roughness_opacity",
	}
	allTextures := make([]string, 0)
	allTextures = append(allTextures, albedoTextures...)
	allTextures = append(allTextures, normalTextures...)
	ml.loadMaterialTexturesAsync(mat, allTextures)
	albedoRemoveAlpha := false
	albedoName := stingray.Hash{}
	for _, texture := range albedoTextures {
		if mat == nil {
			break
		}
		if name, contains := mat.Textures[stingray.Sum(texture).Thin()]; contains {
			albedoRemoveAlpha = slices.Contains(albedoRemoveAlphaTextures, texture)
			albedoName = name
			break
		}
	}
	albedoTexture := Texture{
		id:           0,
		target:       gl.TEXTURE_2D,
		name:         albedoName,
		override:     stingray.NewFileID(albedoName, stingray.Sum("texture")),
		slot:         "texAlbedo",
		created:      false,
		loaded:       false,
		defaultColor: []byte{255, 255, 255, 255},
		filter:       gl.LINEAR,
	}
	if albedoRemoveAlpha {
		data := ml.textureData[albedoName]
		if data.Swizzles == nil {
			data.Swizzles = []uint32{gl.RED, gl.GREEN, gl.BLUE, gl.ALPHA}
		}
		data.Swizzles[3] = gl.ONE
		ml.textureData[albedoName] = data
	}
	material.textures = append(material.textures, albedoTexture)

	reconstructNormalZ := false
	normalName := stingray.Hash{}
	for _, texture := range normalTextures {
		if mat == nil {
			break
		}
		if name, contains := mat.Textures[stingray.Sum(texture).Thin()]; contains {
			reconstructNormalZ = slices.Contains(normalReconstructTextures, texture)
			normalName = name
			break
		}
	}
	normalTexture := Texture{
		id:           0,
		target:       gl.TEXTURE_2D,
		name:         normalName,
		override:     stingray.NewFileID(normalName, stingray.Sum("texture")),
		slot:         "texNormal",
		created:      false,
		loaded:       false,
		defaultColor: []byte{128, 128, 255, 128},
		filter:       gl.LINEAR,
	}
	material.textures = append(material.textures, normalTexture)

	if reconstructNormalZ {
		gl.ProgramUniform1i(material.program, material.uniforms.locations["shouldReconstructNormalZ"], 1)
	} else {
		gl.ProgramUniform1i(material.program, material.uniforms.locations["shouldReconstructNormalZ"], 0)
	}
	gl.ProgramUniform1i(material.program, material.uniforms.locations["hasVisibilityMasks"], 0)
}
