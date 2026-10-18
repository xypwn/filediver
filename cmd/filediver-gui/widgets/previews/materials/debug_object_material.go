package materials

import "github.com/xypwn/filediver/stingray"

var debugObjectUniforms = []string{
	"model",
	"projection",
	"view",
	"color",
}

var debugObjectShaderSources = []string{
	"shaders/debug_object.vert",
	"shaders/debug_object.frag",
}

type DebugObjectMaterial struct {
	*BaseMaterial
}

func NewDebugObjectMaterial(name string) *DebugObjectMaterial {
	id := stingray.NewFileID(stingray.Sum("debug_object"), stingray.Sum("material"))
	return &DebugObjectMaterial{
		BaseMaterial: &BaseMaterial{
			id:       id.Name,
			override: id,
			name:     name,
			shown:    true,
		},
	}
}

func (mat *DebugObjectMaterial) Generate() error {
	return mat.generate(
		debugObjectShaderSources,
		0,
		debugObjectUniforms,
	)
}

var _ Material = &DebugObjectMaterial{}
