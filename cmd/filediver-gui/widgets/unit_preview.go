package widgets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"io"
	"path"
	"slices"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/dds"
	extr_material "github.com/xypwn/filediver/extractor/material"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

//go:embed shaders/*
var unitPreviewShaderCode embed.FS

// stingray coords to OpenGL coords
var stingrayToGLCoords = mgl32.Mat4{
	1, 0, 0, 0,
	0, 0, 1, 0,
	0, 1, 0, 0,
	0, 0, 0, 1,
}

type unitPreviewGLObject struct {
	vao       uint32 // vertex array object
	ibo       uint32 // index buffer object
	vbo       uint32 // vertex buffer object
	texAlbedo uint32

	// Uniform locations
	mvpLoc          int32
	modelLoc        int32
	normalLoc       int32
	viewPositionLoc int32
	colorLoc        int32
	texAlbedoLoc    int32

	numIndices int32
}

func (obj *unitPreviewGLObject) genObjects(textures bool) {
	gl.GenVertexArrays(1, &obj.vao)
	gl.GenBuffers(1, &obj.vbo)
	gl.GenBuffers(1, &obj.ibo)
	if textures {
		gl.GenTextures(1, &obj.texAlbedo)
	}

	gl.BindVertexArray(obj.vao)
	defer gl.BindVertexArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, obj.vbo)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, obj.ibo)
}

func (obj *unitPreviewGLObject) genUniformLocations(program uint32, mvp bool, model bool, normal bool, viewPosition bool, color bool, texAlbedoLoc bool) {
	if mvp {
		obj.mvpLoc = gl.GetUniformLocation(program, gl.Str("mvp\x00"))
	}
	if model {
		obj.modelLoc = gl.GetUniformLocation(program, gl.Str("model\x00"))
	}
	if normal {
		obj.normalLoc = gl.GetUniformLocation(program, gl.Str("normal\x00"))
	}
	if viewPosition {
		obj.viewPositionLoc = gl.GetUniformLocation(program, gl.Str("viewPosition\x00"))
	}
	if color {
		obj.colorLoc = gl.GetUniformLocation(program, gl.Str("color\x00"))
	}
	if texAlbedoLoc {
		obj.texAlbedoLoc = gl.GetUniformLocation(program, gl.Str("texAlbedo\x00"))
	}
}

func (obj unitPreviewGLObject) deleteObjects() {
	gl.DeleteVertexArrays(1, &obj.vao)
	gl.DeleteBuffers(1, &obj.vbo)
	gl.DeleteBuffers(1, &obj.ibo)
	gl.DeleteTextures(1, &obj.texAlbedo)
}

type UnitPreviewState struct {
	fb *GLViewState

	objectProgram uint32
	object        unitPreviewGLObject

	dbgObjProgram uint32
	dbgObj        unitPreviewGLObject

	vfov         float32
	model        mgl32.Mat4
	viewDistance float32
	viewRotation mgl32.Vec2 // {yaw, pitch}

	// Axis-aligned bounding box. Don't forget
	// to multiply aabb's vertices with aabbMat first!
	aabb    [2]mgl32.Vec3
	aabbMat mgl32.Mat4

	showAABB        bool
	zoomToFitOnLoad bool
	zoomToFitAABB   bool // set view distance to fit AABB next frame
}

func NewUnitPreview() (*UnitPreviewState, error) {
	var err error

	pv := &UnitPreviewState{}

	pv.fb, err = NewGLView()
	if err != nil {
		return nil, err
	}

	buildShader := func(name string) (uint32, error) {
		vs, err := unitPreviewShaderCode.ReadFile(path.Join("shaders", name+".vert"))
		if err != nil {
			return 0, err
		}
		fs, err := unitPreviewShaderCode.ReadFile(path.Join("shaders", name+".frag"))
		if err != nil {
			return 0, err
		}
		return glutils.CreateProgramFromSources(string(vs), string(fs))
	}

	pv.objectProgram, err = buildShader("unit_preview")
	if err != nil {
		return nil, err
	}

	pv.dbgObjProgram, err = buildShader("debug_object")
	if err != nil {
		return nil, err
	}

	pv.object.genObjects(true)
	gl.BindTexture(gl.TEXTURE_2D, pv.object.texAlbedo)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	pv.object.genUniformLocations(pv.objectProgram, true, true, true, true, false, true)

	pv.dbgObj.genObjects(false)
	pv.dbgObj.genUniformLocations(pv.dbgObjProgram, true, false, false, false, true, false)

	pv.vfov = mgl32.DegToRad(60)
	pv.viewDistance = 25

	return pv, nil
}

func (pv *UnitPreviewState) Delete() {
	pv.fb.Delete()
	gl.DeleteProgram(pv.objectProgram)
	pv.object.deleteObjects()
	pv.dbgObj.deleteObjects()
}

func (pv *UnitPreviewState) LoadUnit(mainData, gpuData []byte, getResource GetResourceFunc) error {
	info, err := unit.LoadInfo(bytes.NewReader(mainData))
	if err != nil {
		return err
	}

	if len(info.MeshInfos) == 0 {
		return fmt.Errorf("unit contains no meshes")
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
		if highestDetailIdx == -1 {
			return fmt.Errorf("unable to find mesh to load")
		}
		meshToLoad = uint32(highestDetailIdx)
	}

	var mesh unit.Mesh
	{
		meshes, err := unit.LoadMeshes(bytes.NewReader(gpuData), info, []uint32{meshToLoad})
		if err != nil {
			return err
		}
		mesh = meshes[meshToLoad]
	}
	{
		pv.aabb = [2]mgl32.Vec3{mesh.Info.Header.AABB.Min, mesh.Info.Header.AABB.Max}
		pv.aabbMat = info.Bones[mesh.Info.Header.AABBTransformIndex].Matrix
	}

	if len(mesh.Positions) == 0 {
		return fmt.Errorf("mesh contains no positions")
	}
	if len(mesh.Normals) == 0 {
		return fmt.Errorf("mesh contains no normals")
	}
	if len(mesh.UVCoords) == 0 || len(mesh.UVCoords[0]) == 0 {
		return fmt.Errorf("mesh contains no UV coordinates")
	}

	// Upload object texture
	albedoTexFileName, albedoRemoveAlpha, err := func() (albedoFileName stingray.Hash, albedoRemoveAlpha bool, err error) {
		for matID, matFileName := range info.Materials {
			if !slices.Contains(mesh.Info.Materials, matID) {
				continue
			}
			matData, ok, err := getResource(stingray.FileID{
				Name: matFileName,
				Type: stingray.Sum64([]byte("material")),
			}, stingray.DataMain)
			if err != nil {
				return stingray.Hash{}, false, fmt.Errorf("load material %v.material: %w", matFileName, err)
			}
			if !ok {
				return stingray.Hash{}, false, fmt.Errorf("load material %v.material does not exist", matFileName)
			}
			mat, err := material.Load(bytes.NewReader(matData))
			// TODO: Use all albedo textures. Currently, simply the first one
			// found is used.
			for texUsage, texFileName := range mat.Textures {
				removeAlpha := true
				switch extr_material.TextureUsage(texUsage.Value) {
				case extr_material.ColorRoughness, extr_material.ColorSpecularB, extr_material.AlbedoIridescence:
					removeAlpha = false
					fallthrough
				case extr_material.CoveringAlbedo, extr_material.InputImage, extr_material.Albedo:
					return texFileName, removeAlpha, nil
				}
			}
		}
		return stingray.Hash{Value: 0}, false, nil
	}()
	if err != nil {
		return err
	}
	if albedoTexFileName.Value != 0 {
		file := stingray.FileID{Name: albedoTexFileName, Type: stingray.Sum64([]byte("texture"))}
		var texMain, texStream, texGPU []byte
		var err error
		texMain, _, err = getResource(file, stingray.DataMain)
		if err == nil {
			texStream, _, err = getResource(file, stingray.DataStream)
		}
		if err == nil {
			texGPU, _, err = getResource(file, stingray.DataGPU)
		}
		if err != nil {
			return fmt.Errorf("load texture %v.texture does not exist", albedoTexFileName)
		}
		dataR := io.MultiReader(
			bytes.NewReader(texMain),
			bytes.NewReader(texStream),
			bytes.NewReader(texGPU),
		)
		if _, err := texture.DecodeInfo(dataR); err != nil {
			return fmt.Errorf("loading stingray DDS info: %w", err)
		}
		dds, err := dds.Decode(dataR, false)
		if err != nil {
			return fmt.Errorf("loading DDS image: %w", err)
		}
		img, ok := dds.Image.(*image.NRGBA)
		if !ok {
			return fmt.Errorf("expected albedo texture to be of type *image.NRGBA")
		}
		gl.BindTexture(gl.TEXTURE_2D, pv.object.texAlbedo)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.Bounds().Dx()), int32(img.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
		if albedoRemoveAlpha {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, gl.ONE)
		}
		gl.BindTexture(gl.TEXTURE_2D, 0)
	} else {
		data := []byte{255, 255, 255, 255}
		gl.BindTexture(gl.TEXTURE_2D, pv.object.texAlbedo)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}

	// Upload object data
	{
		gl.BindVertexArray(pv.object.vao)

		positionsSize := len(mesh.Positions) * 3 * 4
		normalsSize := len(mesh.Normals) * 3 * 4
		uvsSize := len(mesh.UVCoords[0]) * 2 * 4

		gl.BindBuffer(gl.ARRAY_BUFFER, pv.object.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, positionsSize+normalsSize+uvsSize, nil, gl.STATIC_DRAW)
		offset := 0
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, positionsSize, gl.Ptr(mesh.Positions))
		gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, uintptr(offset))
		gl.EnableVertexAttribArray(0)
		offset += positionsSize
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, normalsSize, gl.Ptr(mesh.Normals))
		gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, true, 3*4, uintptr(offset))
		gl.EnableVertexAttribArray(1)
		offset += normalsSize
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, uvsSize, gl.Ptr(mesh.UVCoords[0]))
		gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, 2*4, uintptr(offset))
		gl.EnableVertexAttribArray(2)
		offset += uvsSize

		pv.object.numIndices = 0
		for _, indices := range mesh.Indices {
			pv.object.numIndices += int32(len(indices))
		}
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(pv.object.numIndices*4), nil, gl.STATIC_DRAW)
		{
			offset := 0
			for _, indices := range mesh.Indices {
				length := len(indices)
				gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, offset*4, length*4, gl.Ptr(indices))
				offset += length
			}
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	}

	// Upload debug object data
	{
		gl.BindVertexArray(pv.dbgObj.vao)

		verts := pv.getAABBVertices()
		gl.BindBuffer(gl.ARRAY_BUFFER, pv.dbgObj.vbo)
		defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BufferData(gl.ARRAY_BUFFER, len(verts)*3*4, gl.Ptr(verts[:]), gl.STATIC_DRAW)

		pv.dbgObj.numIndices = int32(len(aabbIndices))
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(pv.dbgObj.numIndices*4), gl.Ptr(aabbIndices[:]), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.BindVertexArray(0)
	}

	pv.model = stingrayToGLCoords

	if pv.zoomToFitOnLoad {
		pv.zoomToFitAABB = true
	}

	return nil
}

func (pv *UnitPreviewState) computeMVP(aspectRatio float32) (
	normal mgl32.Mat4,
	viewPosition mgl32.Vec3,
	view mgl32.Mat4,
	projection mgl32.Mat4,
	mvp mgl32.Mat4,
) {
	normal = pv.model.Inv().Transpose()
	{
		mat := mgl32.Ident3()
		mat = mat.Mul3(mgl32.Rotate3DY(pv.viewRotation[0]))
		mat = mat.Mul3(mgl32.Rotate3DZ(-pv.viewRotation[1]))
		viewPosition = mat.Mul3x1(mgl32.Vec3{pv.viewDistance, 0, 0})
	}
	view = mgl32.LookAt(
		viewPosition[0], viewPosition[1], viewPosition[2],
		0, 0, 0,
		0, 1, 0,
	)
	projection = mgl32.Perspective(
		pv.vfov,
		aspectRatio,
		0.001,
		32768,
	)
	mvp = projection.Mul4(view).Mul4(pv.model)
	return
}

var aabbIndices = [12 * 3]uint32{
	1, 2, 0,
	1, 3, 2,
	0, 6, 4,
	0, 2, 6,
	4, 7, 5,
	4, 6, 7,
	5, 3, 1,
	5, 7, 3,
	2, 3, 7,
	2, 7, 6,
	0, 4, 5,
	0, 5, 1,
}

func (pv *UnitPreviewState) getAABBVertices() [8]mgl32.Vec3 {
	return [8]mgl32.Vec3{
		{pv.aabb[0][0], pv.aabb[0][1], pv.aabb[0][2]},
		{pv.aabb[0][0], pv.aabb[0][1], pv.aabb[1][2]},
		{pv.aabb[0][0], pv.aabb[1][1], pv.aabb[0][2]},
		{pv.aabb[0][0], pv.aabb[1][1], pv.aabb[1][2]},
		{pv.aabb[1][0], pv.aabb[0][1], pv.aabb[0][2]},
		{pv.aabb[1][0], pv.aabb[0][1], pv.aabb[1][2]},
		{pv.aabb[1][0], pv.aabb[1][1], pv.aabb[0][2]},
		{pv.aabb[1][0], pv.aabb[1][1], pv.aabb[1][2]},
	}
}

func UnitPreview(name string, pv *UnitPreviewState) {
	if pv.object.numIndices == 0 {
		return
	}

	imgui.PushIDStr(name)
	defer imgui.PopID()

	viewSize := imgui.ContentRegionAvail()
	viewSize.Y -= imutils.CheckboxHeight()

	GLView(name, pv.fb, viewSize,
		func() {
			io := imgui.CurrentIO()

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

			normal, viewPosition, _, _, mvp := pv.computeMVP(size.X / size.Y)

			// Draw object
			gl.Enable(gl.DEPTH_TEST)
			gl.UseProgram(pv.objectProgram)
			gl.BindVertexArray(pv.object.vao)
			gl.UniformMatrix4fv(pv.object.mvpLoc, 1, false, &mvp[0])
			gl.UniformMatrix4fv(pv.object.modelLoc, 1, false, &pv.model[0])
			gl.UniformMatrix4fv(pv.object.normalLoc, 1, false, &normal[0])
			gl.Uniform3fv(pv.object.viewPositionLoc, 1, &viewPosition[0])
			gl.BindTexture(gl.TEXTURE_2D, pv.object.texAlbedo)
			gl.DrawElements(gl.TRIANGLES, pv.object.numIndices, gl.UNSIGNED_INT, nil)
			gl.BindTexture(gl.TEXTURE_2D, 0)
			gl.BindVertexArray(0)
			gl.UseProgram(0)

			// Draw debug object
			if pv.showAABB {
				gl.Disable(gl.DEPTH_TEST)
				gl.UseProgram(pv.dbgObjProgram)
				gl.BindVertexArray(pv.dbgObj.vao)
				{
					aabbMVP := mvp.Mul4(pv.aabbMat)
					gl.UniformMatrix4fv(pv.dbgObj.mvpLoc, 1, false, &aabbMVP[0])
				}
				{
					col := [4]float32{0.3, 0.3, 0.8, 0.2}
					gl.Uniform4fv(pv.dbgObj.colorLoc, 1, &col[0])
				}
				gl.DrawElements(gl.TRIANGLES, pv.dbgObj.numIndices, gl.UNSIGNED_INT, nil)
				gl.BindVertexArray(0)
				gl.UseProgram(0)
			}

			if pv.zoomToFitAABB {
				pv.viewDistance = 32768

				_, viewPosition, view, projection, _ := pv.computeMVP(size.X / size.Y)

				aabbVerts3 := pv.getAABBVertices()

				var aabbVerts [8]mgl32.Vec4
				for i := range aabbVerts {
					p := aabbVerts3[i].Vec4(1.0)
					p = pv.model.Mul4x1(pv.aabbMat.Mul4x1(p))
					aabbVerts[i] = p
				}

				var projAABBVerts [8]mgl32.Vec2
				for i, p := range aabbVerts {
					p = projection.Mul4x1(view.Mul4x1(p))
					p = p.Mul(1 / p.W())
					projAABBVerts[i] = p.Vec2()
				}

				// Select max projected coordinate (max distance from screen center)
				var maxDist float32
				var maxDistVert mgl32.Vec4
				for i, p := range projAABBVerts {
					for _, v := range p {
						v := mgl32.Abs(v)
						if v > maxDist {
							maxDist = v
							maxDistVert = aabbVerts[i]
						}
					}
				}

				// Calculate orthogonal distance from camera to selected vertex
				var od float32
				{
					viewDir := mgl32.Vec3{}.Sub(viewPosition).Normalize()
					vert := maxDistVert
					vert = vert.Mul(1 / vert.W())
					camToVert := vert.Vec3().Sub(viewPosition)
					od = viewDir.Dot(camToVert)
				}

				// Fit model to screen
				pv.viewDistance -= od * (1 - maxDist)

				pv.zoomToFitAABB = false
			}
		},
		nil,
	)

	if imgui.Button(fnt.I("Home")) {
		pv.viewRotation = mgl32.Vec2{}
		pv.viewDistance = 25
	}
	imgui.SetItemTooltip("Reset view")
	imgui.SameLine()
	imgui.Checkbox("Auto-zoom on load", &pv.zoomToFitOnLoad)
	imgui.SameLine()
	imgui.Checkbox("Show bounding box", &pv.showAABB)
}
