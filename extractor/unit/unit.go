package unit

import (
	"fmt"
	"io"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
)

func Convert(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
	u, err := unit.Load(ins[stingray.DataMain], ins[stingray.DataGPU])
	if err != nil {
		return err
	}

	for _, mesh := range u.Meshes {
		for i := range mesh.Positions {
			p := mesh.Positions[i]
			p[0], p[1], p[2] = p[1], p[2], p[0]
			mesh.Positions[i] = p
		}
	}

	doc := gltf.NewDocument()
	doc.Asset.Generator = "https://github.com/xypwn/filediver"
	for i, mesh := range u.Meshes {
		name := fmt.Sprintf("Mesh %v", i)
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{
			Name: name,
			Primitives: []*gltf.Primitive{{
				Indices: gltf.Index(modeler.WriteIndices(doc, mesh.Indices)),
				Attributes: map[string]uint32{
					gltf.POSITION: modeler.WritePosition(doc, mesh.Positions),
					gltf.COLOR_0:  modeler.WriteColor(doc, mesh.Colors),
				},
			}},
		})
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: name,
			Mesh: gltf.Index(uint32(i)),
		})
	}
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, 0)
	gltf.SaveBinary(doc, outPath+".glb")

	return nil
}
