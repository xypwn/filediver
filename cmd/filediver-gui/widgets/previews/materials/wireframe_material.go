package materials

import "github.com/xypwn/filediver/stingray"

type WireframeMaterial struct {
	*BaseMaterial
}

var wireframeUniforms = []string{
	"model",
	"projection",
	"view",
	"color",
	"udimShown",
	"hasVisibilityMasks",
}

var wireframeShaderSources = []string{
	"shaders/object_wireframe.vert",
	"shaders/object_wireframe.geom",
	"shaders/object_wireframe.frag",
}

func NewWireframeMaterial() *WireframeMaterial {
	id := stingray.NewFileID(stingray.Sum("wireframe"), stingray.Sum("material"))
	return &WireframeMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     "wireframe",
			shown:    true,
		},
	}
}

func (mat *WireframeMaterial) Generate() error {
	return mat.generate(
		wireframeShaderSources,
		0,
		wireframeUniforms,
	)
}

type SpeedtreeWireframeMaterial struct {
	*BaseMaterial
}

func NewSpeedtreeWireframeMaterial() *SpeedtreeWireframeMaterial {
	id := stingray.NewFileID(stingray.Sum("speedtree_wireframe"), stingray.Sum("material"))
	return &SpeedtreeWireframeMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     "wireframe",
			shown:    true,
		},
	}
}

func (mat *SpeedtreeWireframeMaterial) Generate() error {
	return mat.generate(
		[]string{
			"shaders/speedtree_wireframe.vert",
			"shaders/object_wireframe.geom",
			"shaders/object_wireframe.frag",
		},
		0,
		wireframeUniforms,
	)
}

var _ Material = &WireframeMaterial{}
var _ Material = &SpeedtreeWireframeMaterial{}
