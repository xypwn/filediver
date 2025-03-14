package widgets

import (
	_ "embed"
	"io"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/stingray/unit"
)

//go:embed shaders/unit_preview.frag
var unitPreviewFragCode string

//go:embed shaders/unit_preview.vert
var unitPreviewVertCode string

// stingray coords to OpenGL coords
var stingrayToGLCoords = mgl32.Mat4{
	1, 0, 0, 0,
	0, 0, 1, 0,
	0, 1, 0, 0,
	0, 0, 0, 1,
}

type UnitPreviewState struct {
	fb      *GLViewState
	program uint32
	vao     uint32 // vertex array object
	ibo     uint32 // index buffer object
	vbo     uint32 // vertex buffer object

	// Uniform locations
	mvpLoc          int32
	modelLoc        int32
	normalLoc       int32
	viewPositionLoc int32

	numIndices   int32
	model        mgl32.Mat4
	projection   mgl32.Mat4
	viewDistance float32
	viewRotation mgl32.Vec2 // {yaw, pitch}

	isDragging bool
	IsUsing    bool // true if window shouldn't handle mouse events
}

func NewUnitPreview() (*UnitPreviewState, error) {
	var err error

	pv := &UnitPreviewState{}

	pv.fb, err = NewGLView()
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

	pv.viewDistance = 25

	return pv, nil
}

func (pv *UnitPreviewState) Delete() {
	pv.fb.Delete()
	gl.DeleteProgram(pv.program)
	gl.DeleteVertexArrays(1, &pv.vao)
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

	positionsSize := len(mesh.Positions) * 3 * 4
	normalsSize := len(mesh.Normals) * 3 * 4
	gl.BufferData(gl.ARRAY_BUFFER, positionsSize+normalsSize, nil, gl.STATIC_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, positionsSize, gl.Ptr(mesh.Positions))
	gl.BufferSubData(gl.ARRAY_BUFFER, positionsSize, normalsSize, gl.Ptr(mesh.Normals))

	pv.numIndices = 0
	for _, indices := range mesh.Indices {
		pv.numIndices += int32(len(indices))
	}
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(pv.numIndices*4), nil, gl.STATIC_DRAW)
	{
		offset := 0
		for _, indices := range mesh.Indices {
			length := len(indices)
			gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, offset*4, length*4, gl.Ptr(indices))
			offset += length
		}
	}

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, true, 3*4, uintptr(positionsSize))
	gl.EnableVertexAttribArray(1)

	pv.mvpLoc = gl.GetUniformLocation(pv.program, gl.Str("mvp\x00"))
	pv.modelLoc = gl.GetUniformLocation(pv.program, gl.Str("model\x00"))
	pv.normalLoc = gl.GetUniformLocation(pv.program, gl.Str("normal\x00"))
	pv.viewPositionLoc = gl.GetUniformLocation(pv.program, gl.Str("viewPosition\x00"))

	pv.model = stingrayToGLCoords

	return nil
}

func UnitPreview(name string, pv *UnitPreviewState) {
	if pv.numIndices == 0 {
		return
	}

	GLView(name, pv.fb,
		func() {
			io := imgui.CurrentIO()

			pv.IsUsing = imgui.IsItemActive() || imgui.IsItemHovered()

			if imgui.IsItemActive() {
				md := io.MouseDelta()
				pv.viewRotation = pv.viewRotation.Add(mgl32.Vec2{md.X, md.Y}.Mul(-0.01))
				pv.viewRotation[1] = mgl32.Clamp(pv.viewRotation[1], -1.55, 1.55)
			}
			if imgui.IsItemHovered() {
				scroll := io.MouseWheel()
				pv.viewDistance = mgl32.Clamp(pv.viewDistance-(0.1*pv.viewDistance*scroll), 0.1, 1000)
			}
		},
		func(pos, size imgui.Vec2) {
			gl.ClearColor(0.2, 0.2, 0.2, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			gl.UseProgram(pv.program)
			defer gl.UseProgram(0)

			gl.BindVertexArray(pv.vao)
			defer gl.BindVertexArray(0)

			var viewPosition mgl32.Vec3
			{
				mat := mgl32.Ident3()
				mat = mat.Mul3(mgl32.Rotate3DY(pv.viewRotation[0]))
				mat = mat.Mul3(mgl32.Rotate3DZ(-pv.viewRotation[1]))
				viewPosition = mat.Mul3x1(mgl32.Vec3{pv.viewDistance, 0, 0})
			}

			view := mgl32.LookAt(
				viewPosition[0], viewPosition[1], viewPosition[2],
				0, 0, 0,
				0, 1, 0,
			)

			pv.projection = mgl32.Perspective(
				mgl32.DegToRad(60),
				size.X/size.Y,
				0.1,
				1000,
			)

			mvp := pv.projection.Mul4(view).Mul4(pv.model)
			gl.UniformMatrix4fv(pv.mvpLoc, 1, false, &mvp[0])
			gl.UniformMatrix4fv(pv.modelLoc, 1, false, &pv.model[0])
			normal := pv.model.Inv().Transpose()
			gl.UniformMatrix4fv(pv.normalLoc, 1, false, &normal[0])
			gl.Uniform3fv(pv.viewPositionLoc, 1, &viewPosition[0])

			gl.DrawElements(gl.TRIANGLES, pv.numIndices, gl.UNSIGNED_INT, nil)
		},
		nil,
	)
}
