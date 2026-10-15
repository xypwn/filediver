package materials

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

// Textures used by the speedtree fragment shader
var speedtreeTextureNames = []string{
	"tex0",
	"tex1",
	"tex2",
	"fibonacci_normal_lut",
	"ibl_brdf_lut",
}

func isSpeedtreeMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}
	_, containsTex0 := mat.Textures[stingray.Sum("tex0").Thin()]
	_, containsTex1 := mat.Textures[stingray.Sum("tex1").Thin()]
	_, containsTex2 := mat.Textures[stingray.Sum("tex2").Thin()]
	return containsTex0 && containsTex1 && containsTex2
}

type SpeedtreeMaterial struct {
	*BaseMaterial
}

func NewSpeedtreeMaterial(id stingray.FileID, name string, shown bool) *SpeedtreeMaterial {
	return &SpeedtreeMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *SpeedtreeMaterial) Generate() error {
	return mat.generate(
		[]string{
			"shaders/speedtree.vert",
			"shaders/speedtree.frag",
		},
		0,
		append(baseUniforms, speedtreeTextureNames...),
	)
}

var _ Material = &SpeedtreeMaterial{}

type SpeedtreeUnitMaterial struct {
	*BaseMaterial
}

func NewSpeedtreeUnitMaterial(id stingray.FileID, name string, shown bool) *SpeedtreeUnitMaterial {
	return &SpeedtreeUnitMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    shown,
		},
	}
}

func (mat *SpeedtreeUnitMaterial) Generate() error {
	return mat.generate(
		[]string{
			"shaders/object.vert",
			"shaders/object.geom",
			"shaders/speedtree.frag",
		},
		0,
		append(baseUniforms, speedtreeTextureNames...),
	)
}

var _ Material = &SpeedtreeUnitMaterial{}

func (ml *MaterialManager) updateSpeedtreeMaterial(material *SpeedtreeMaterial, mat *material.Material) {
	ml.updateAssetGradingLut()

	assetGradingHash := stingray.Sum("asset_grading_lut")
	material.textures = append(material.textures, Texture{
		id:       ml.assetGradingLutTexture,
		target:   gl.TEXTURE_BUFFER,
		name:     assetGradingHash,
		override: stingray.NewFileID(assetGradingHash, stingray.Sum("texture")),
		slot:     "asset_grading_lut",
		created:  false,
		loaded:   true,
	})

	if _, contains := mat.Textures[stingray.Sum("fibonacci_normal_lut").Thin()]; !contains {
		mat.Textures[stingray.Sum("fibonacci_normal_lut").Thin()] = stingray.Sum("content/art_shared/textures/fibonacci_normal_lut")
	}

	ml.setupMaterialCommon(speedtreeTextureNames, material, mat)
}

func (ml *MaterialManager) updateSpeedtreeUnitMaterial(material *SpeedtreeUnitMaterial, mat *material.Material) {
	ml.updateAssetGradingLut()

	assetGradingHash := stingray.Sum("asset_grading_lut")
	material.textures = append(material.textures, Texture{
		id:       ml.assetGradingLutTexture,
		target:   gl.TEXTURE_BUFFER,
		name:     assetGradingHash,
		override: stingray.NewFileID(assetGradingHash, stingray.Sum("texture")),
		slot:     "asset_grading_lut",
		created:  false,
		loaded:   true,
	})

	if _, contains := mat.Textures[stingray.Sum("fibonacci_normal_lut").Thin()]; !contains {
		mat.Textures[stingray.Sum("fibonacci_normal_lut").Thin()] = stingray.Sum("content/art_shared/textures/fibonacci_normal_lut")
	}

	ml.setupMaterialCommon(speedtreeTextureNames, material, mat)
}
