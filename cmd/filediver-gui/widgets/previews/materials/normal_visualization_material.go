package materials

import "github.com/xypwn/filediver/stingray"

var normalVisUniforms = []string{
	"model",
	"projection",
	"view",
	"len",
	"showTangentBitangent",
	"udimShown",
	"hasVisibilityMasks",
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
		[]string{
			"shaders/object_normal_vis.vert",
			"shaders/object_normal_vis.geom",
			"shaders/object_normal_vis.frag",
		},
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
		[]string{
			"shaders/speedtree_normal_vis.vert",
			"shaders/object_normal_vis.geom",
			"shaders/object_normal_vis.frag",
		},
		0,
		normalVisUniforms,
	)
}

var _ Material = &NormalVisMaterial{}
var _ Material = &SpeedtreeNormalVisMaterial{}
