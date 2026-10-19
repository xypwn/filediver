package materials

import (
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/xypwn/filediver/stingray"
)

var normalVisUniforms = []string{
	"model",
	"projection",
	"view",
	"len",
	"showTangentBitangent",
	"udimShown",
	"hasVisibilityMasks",
}

var normalVisShaderSources = []string{
	"shaders/object_normal_vis.vert",
	"shaders/object_normal_vis.geom",
	"shaders/object_normal_vis.frag",
}

var speedtreeNormalVisShaderSources = []string{
	"shaders/speedtree_normal_vis.vert",
	"shaders/object_normal_vis.geom",
	"shaders/object_normal_vis.frag",
}

type NormalVisMaterial struct {
	*BaseMaterial
}

func NewNormalVisMaterial() *NormalVisMaterial {
	id := stingray.NewFileID(stingray.Sum("unit_normal_vis"), stingray.Sum("material"))
	return &NormalVisMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     "normal visualization",
			shown:    true,
		},
	}
}

func (mat *NormalVisMaterial) Generate() error {
	return mat.generate(
		normalVisShaderSources,
		0,
		normalVisUniforms,
	)
}

type SpeedtreeNormalVisMaterial struct {
	*BaseMaterial
}

func NewSpeedtreeNormalVisMaterial() *SpeedtreeNormalVisMaterial {
	id := stingray.NewFileID(stingray.Sum("speedtree_normal_vis"), stingray.Sum("material"))
	return &SpeedtreeNormalVisMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     "normal visualization",
			shown:    true,
		},
	}
}

func (mat *SpeedtreeNormalVisMaterial) Generate() error {
	return mat.generate(
		speedtreeNormalVisShaderSources,
		0,
		append(normalVisUniforms, "fibonacci_normal_lut"),
	)
}

func (ml *MaterialManager) UpdateSpeedtreeNormalVisMaterial(material *SpeedtreeNormalVisMaterial) {
	normalLutHash := stingray.Sum("content/art_shared/textures/fibonacci_normal_lut")
	normalLutTexture := NewTexture(normalLutHash, stingray.NewFileID(normalLutHash, stingray.Sum("texture")), gl.TEXTURE_2D, gl.LINEAR, "fibonacci_normal_lut", false, false)

	ml.loadMaterialTextureSync(stingray.Sum("fibonacci_normal_lut").Thin(), normalLutHash)
	material.AddTexture(normalLutTexture)
	ml.FinalizeMaterialTextures(material)
}

var _ Material = &NormalVisMaterial{}
var _ Material = &SpeedtreeNormalVisMaterial{}
