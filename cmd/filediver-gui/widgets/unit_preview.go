package widgets

import (
	_ "embed"
	"io"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/stingray/unit"
)

//go:embed shaders/unit_preview.frag
var unitPreviewFragCode string

//go:embed shaders/unit_preview.vert
var unitPreviewVertCode string

type UnitPreviewState struct {
	fb         *Framebuffer
	program    uint32
	vao        uint32 // vertex array object
	ibo        uint32 // index buffer object
	vbo        uint32 // vertex buffer object
	mvpLoc     int32  // MVP matrix uniform location
	numIndices int32
}

func CreateUnitPreview() (*UnitPreviewState, error) {
	var err error

	pv := &UnitPreviewState{}

	pv.fb, err = CreateFramebuffer()
	if err != nil {
		return nil, err
	}

	pv.program, err = glutils.CreateProgramFromSources(unitPreviewVertCode, unitPreviewFragCode)
	if err != nil {
		return nil, err
	}

	gl.GenVertexArrays(1, &pv.vao)
	gl.GenBuffers(1, &pv.vbo)
	gl.GenBuffers(1, &pv.ibo)

	gl.BindVertexArray(pv.vao)
	defer gl.BindVertexArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, pv.vbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, pv.ibo)

	return pv, nil
}

func (pv *UnitPreviewState) Delete() {
	pv.fb.Delete()
	gl.DeleteProgram(pv.program)
	gl.DeleteBuffers(1, &pv.vao)
	gl.DeleteBuffers(1, &pv.vbo)
	gl.DeleteBuffers(1, &pv.ibo)
}

func (pv *UnitPreviewState) LoadUnit(mainR, gpuR io.ReadSeeker) error {
	info, err := unit.LoadInfo(mainR)
	if err != nil {
		return err
	}

	var meshToLoad uint32
	{
		highestDetailIdx := -1
		highestDetailCount := -1
		for i, info := range info.MeshInfos {
			for _, group := range info.Groups {
				if int(group.NumIndices) > highestDetailCount && info.Header.MeshType != unit.MeshTypeUnknown00 {
					highestDetailIdx = i
					highestDetailCount = int(group.NumIndices)
				}
			}
		}
		if highestDetailIdx != -1 {
			meshToLoad = uint32(highestDetailIdx)
		}
	}

	var mesh unit.Mesh
	{
		meshes, err := unit.LoadMeshes(gpuR, info, []uint32{meshToLoad})
		if err != nil {
			return err
		}
		mesh = meshes[meshToLoad]
	}

	gl.BindVertexArray(pv.vao)
	defer gl.BindVertexArray(0)

	gl.BufferData(gl.ARRAY_BUFFER, len(mesh.Positions)*3*4, gl.Ptr(mesh.Positions), gl.STATIC_DRAW)

	indices := mesh.Indices[0]
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, int32(3), gl.FLOAT, false, int32(3*4), nil)
	gl.EnableVertexAttribArray(0)

	pv.numIndices = int32(len(indices))

	pv.mvpLoc = gl.GetUniformLocation(pv.program, gl.Str("mvp\x00"))

	return nil
}

func UnitPreview(name string, pv *UnitPreviewState) {
	GLView(name, pv.fb, func(width, height int32) {
		gl.ClearColor(0.2, 0.2, 0.2, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.UseProgram(pv.program)
		defer gl.UseProgram(0)

		gl.BindVertexArray(pv.vao)
		defer gl.BindVertexArray(0)

		model := mgl32.Ident4()
		view := mgl32.LookAt(
			15, 15, 15,
			0, 0, 0,
			0, 1, 0,
		)
		proj := mgl32.Perspective(
			mgl32.DegToRad(60),
			float32(width)/float32(height),
			0.1,
			1000,
		)
		mvp := proj.Mul4(view).Mul4(model)
		gl.UniformMatrix4fv(pv.mvpLoc, 1, false, &mvp[0])

		gl.DrawElements(gl.TRIANGLES, pv.numIndices, gl.UNSIGNED_INT, nil)
	})
}
