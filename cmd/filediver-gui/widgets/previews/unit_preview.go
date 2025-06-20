package previews

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"slices"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
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
var stingrayToGLCoords = mgl32.Mat4FromRows(
	mgl32.Vec4{1, 0, 0, 0},
	mgl32.Vec4{0, 0, 1, 0},
	mgl32.Vec4{0, -1, 0, 0},
	mgl32.Vec4{0, 0, 0, 1},
)

type unitPreviewObject struct {
	vao       uint32 // vertex array object
	ibo       uint32 // index buffer object
	vbo       uint32 // vertex buffer object
	texAlbedo uint32
	texNormal uint32

	numVertices int32
	numIndices  int32
}

// NOTE(xypwn): We do at most ~10 lookups once per frame,
// so it should be fine to store this in a string map.
type unitPreviewUniforms map[string]int32

func (obj *unitPreviewObject) genObjects(textures bool) {
	gl.GenVertexArrays(1, &obj.vao)
	gl.GenBuffers(1, &obj.vbo)
	gl.GenBuffers(1, &obj.ibo)
	if textures {
		gl.GenTextures(1, &obj.texAlbedo)
		gl.GenTextures(1, &obj.texNormal)
	}

	gl.BindVertexArray(obj.vao)
	defer gl.BindVertexArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, obj.vbo)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, obj.ibo)
}

// Panicks if a name is not a uniform.
func (uniforms *unitPreviewUniforms) generate(program uint32, names ...string) {
	if *uniforms == nil {
		*uniforms = unitPreviewUniforms{}
	}
	for _, name := range names {
		cStr, free := gl.Strs(name + "\x00")
		loc := gl.GetUniformLocation(program, *cStr)
		free()

		if loc == -1 {
			panic(fmt.Sprintf("Invalid uniform name \"%v\" for program %v", name, program))
		}

		(*uniforms)[name] = loc
	}
}

func (obj unitPreviewObject) deleteObjects() {
	gl.DeleteVertexArrays(1, &obj.vao)
	gl.DeleteBuffers(1, &obj.vbo)
	gl.DeleteBuffers(1, &obj.ibo)
	gl.DeleteTextures(1, &obj.texAlbedo)
	gl.DeleteTextures(1, &obj.texNormal)
}

type UnitPreviewState struct {
	fb *widgets.GLViewState

	object                  unitPreviewObject
	objectProgram           uint32
	objectUniforms          unitPreviewUniforms
	objectWireframeProgram  uint32
	objectWireframeUniforms unitPreviewUniforms

	objectNormalVisProgram  uint32
	objectNormalVisUniforms unitPreviewUniforms

	dbgObjProgram  uint32
	dbgObj         unitPreviewObject
	dbgObjUniforms unitPreviewUniforms

	vfov         float32
	model        mgl32.Mat4
	viewDistance float32
	viewRotation mgl32.Vec2 // {yaw, pitch}

	// Axis-aligned bounding box. Don't forget
	// to multiply aabb's vertices with aabbMat first!
	aabb    [2]mgl32.Vec3
	aabbMat mgl32.Mat4

	meshPositions [][3]float32 // for fitting mesh to screen and calculating maxViewDistance

	maxViewDistance float32

	numUdims      uint32
	udimsSelected [64]bool  // udims persistently selected
	udimsShown    [64]int32 // udims visually selected 1 (shown) or 0 (hidden)

	// For dragging selection
	activeUDimListItem  int32
	hoveredUDimListItem int32

	showWireframe          bool
	wireframeColor         [4]float32
	showAABB               bool
	aabbColor              [4]float32
	visualizeNormals       bool
	visualizedNormalsColor [4]float32
	zoomToFitOnLoad        bool
	zoomToFit              bool // set view distance to fit mesh
}

func NewUnitPreview() (*UnitPreviewState, error) {
	var err error

	pv := &UnitPreviewState{}

	pv.fb, err = widgets.NewGLView()
	if err != nil {
		return nil, err
	}

	pv.object.genObjects(true)
	pv.objectProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/object.vert",
		"shaders/object.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.objectUniforms.generate(pv.objectProgram, "mvp", "model", "normalMat", "viewPosition", "texAlbedo", "texNormal", "shouldReconstructNormalZ", "udimShown")

	pv.objectWireframeProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/object_wireframe.vert",
		"shaders/object_wireframe.geom",
		"shaders/object_wireframe.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.objectWireframeUniforms.generate(pv.objectWireframeProgram, "mvp", "color", "udimShown")

	pv.objectNormalVisProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/object_normal_vis.vert",
		"shaders/object_normal_vis.geom",
		"shaders/object_normal_vis.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.objectNormalVisUniforms.generate(pv.objectNormalVisProgram, "mvp", "len", "color", "udimShown")

	pv.dbgObj.genObjects(false)
	pv.dbgObjProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/debug_object.vert",
		"shaders/debug_object.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.dbgObjUniforms.generate(pv.dbgObjProgram, "mvp", "color")

	setupTexture := func(textureID uint32) {
		gl.BindTexture(gl.TEXTURE_2D, textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	setupTexture(pv.object.texAlbedo)
	setupTexture(pv.object.texNormal)

	gl.UseProgram(pv.objectProgram)
	gl.Uniform1i(pv.objectUniforms["texAlbedo"], 0)
	gl.Uniform1i(pv.objectUniforms["texNormal"], 1)
	gl.UseProgram(0)

	pv.vfov = mgl32.DegToRad(60)
	pv.viewDistance = 25

	pv.wireframeColor = [4]float32{1.0, 1.0, 1.0, 0.5}
	pv.aabbColor = [4]float32{0.3, 0.3, 0.8, 0.2}
	pv.visualizedNormalsColor = [4]float32{1.0, 1.0, 0.0, 1.0}

	return pv, nil
}

func (pv *UnitPreviewState) Delete() {
	pv.fb.Delete()
	gl.DeleteProgram(pv.objectProgram)
	pv.object.deleteObjects()
	pv.dbgObj.deleteObjects()
}

func (pv *UnitPreviewState) LoadUnit(mainData, gpuData []byte, getResource GetResourceFunc) error {
	for i := range pv.udimsSelected {
		pv.udimsSelected[i] = true
	}

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
	if len(mesh.UVCoords) == 0 || len(mesh.UVCoords[0]) == 0 {
		return fmt.Errorf("mesh contains no UV coordinates")
	}

	// Upload object texture
	albedoTexFileName, albedoRemoveAlpha, normalTexFileName, reconstructNormalZ, err := func() (albedoFileName stingray.Hash, albedoRemoveAlpha bool, normalFileName stingray.Hash, reconstructNormalZ bool, err error) {
		for matID, matFileName := range info.Materials {
			if !slices.Contains(mesh.Info.Materials, matID) {
				continue
			}
			matData, ok, err := getResource(stingray.FileID{
				Name: matFileName,
				Type: stingray.Sum64([]byte("material")),
			}, stingray.DataMain)
			if err != nil {
				return stingray.Hash{}, false, stingray.Hash{}, false, fmt.Errorf("load material %v.material: %w", matFileName, err)
			}
			if !ok {
				return stingray.Hash{}, false, stingray.Hash{}, false, fmt.Errorf("load material %v.material does not exist", matFileName)
			}
			mat, err := material.Load(bytes.NewReader(matData))
			if err != nil {
				return stingray.Hash{}, false, stingray.Hash{}, false, fmt.Errorf("load material %v.material: %w", matFileName, err)
			}
			// TODO: Use all textures somehow. Currently, simply the first one
			// found is used.
			for texUsage, texFileName := range mat.Textures {
				removeAlpha := true
				switch extr_material.TextureUsage(texUsage.Value) {
				case extr_material.ColorRoughness, extr_material.ColorSpecularB, extr_material.AlbedoIridescence:
					removeAlpha = false
					fallthrough
				case extr_material.CoveringAlbedo, extr_material.InputImage, extr_material.Albedo:
					albedoFileName = texFileName
					albedoRemoveAlpha = removeAlpha
				case extr_material.NormalSpecularAO:
					normalFileName = texFileName
					reconstructNormalZ = false
				case extr_material.Normal, extr_material.Normals, extr_material.NormalMap, extr_material.CoveringNormal, extr_material.NAC, extr_material.BaseData, extr_material.NAR:
					normalFileName = texFileName
					reconstructNormalZ = true
				}
			}
		}
		return
	}()
	if err != nil {
		return err
	}
	uploadStingrayTexture := func(textureID uint32, fileName stingray.Hash) error {
		file := stingray.FileID{Name: fileName, Type: stingray.Sum64([]byte("texture"))}
		var texMain, texStream, texGPU []byte
		if texMain, _, err = getResource(file, stingray.DataMain); err != nil {
			return fmt.Errorf("load texture %v.texture: %w", fileName, err)
		}
		texStream, _, _ = getResource(file, stingray.DataStream)
		texGPU, _, _ = getResource(file, stingray.DataGPU)
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
			return fmt.Errorf("expected texture to be of type *image.NRGBA")
		}
		gl.BindTexture(gl.TEXTURE_2D, textureID)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.Bounds().Dx()), int32(img.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
		gl.BindTexture(gl.TEXTURE_2D, 0)
		return nil
	}

	if albedoTexFileName.Value != 0 {
		if err := uploadStingrayTexture(pv.object.texAlbedo, albedoTexFileName); err != nil {
			return err
		}
		if albedoRemoveAlpha {
			gl.BindTexture(gl.TEXTURE_2D, pv.object.texAlbedo)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, gl.ONE)
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}
	} else {
		data := []byte{255, 255, 255, 255}
		gl.BindTexture(gl.TEXTURE_2D, pv.object.texAlbedo)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	if normalTexFileName.Value != 0 {
		if err := uploadStingrayTexture(pv.object.texNormal, normalTexFileName); err != nil {
			return err
		}
	} else {
		data := []byte{128, 128, 255, 128}
		gl.BindTexture(gl.TEXTURE_2D, pv.object.texNormal)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
		gl.BindTexture(gl.TEXTURE_2D, 0)
		reconstructNormalZ = false
	}

	gl.UseProgram(pv.objectProgram)
	if reconstructNormalZ {
		gl.Uniform1i(pv.objectUniforms["shouldReconstructNormalZ"], 1)
	} else {
		gl.Uniform1i(pv.objectUniforms["shouldReconstructNormalZ"], 0)
	}
	gl.UseProgram(0)

	// Flatten indices
	var indices []uint32
	{
		n := 0
		for _, idxs := range mesh.Indices {
			n += len(idxs)
		}
		indices = make([]uint32, 0, n)
	}
	for _, idxs := range mesh.Indices {
		indices = append(indices, idxs...)
	}

	if len(mesh.Positions) != len(mesh.UVCoords[0]) {
		return errors.New("expected positions and UVs to have the same length")
	}

	pv.object.numIndices = int32(len(indices))
	pv.object.numVertices = int32(len(mesh.Positions))

	// Calculate tangents and bitangents
	normals := make([][3]float32, len(mesh.Positions))
	tangents := make([][3]float32, len(mesh.Positions))
	bitangents := make([][3]float32, len(mesh.Positions))
	nIndicesPerPos := make([]int, len(mesh.Positions))
	for i := 0; i < len(indices); i += 3 {
		i1 := indices[i+0]
		i2 := indices[i+1]
		i3 := indices[i+2]

		p1 := mgl32.Vec3(mesh.Positions[i1])
		p2 := mgl32.Vec3(mesh.Positions[i2])
		p3 := mgl32.Vec3(mesh.Positions[i3])
		uv1 := mgl32.Vec2(mesh.UVCoords[0][i1])
		uv2 := mgl32.Vec2(mesh.UVCoords[0][i2])
		uv3 := mgl32.Vec2(mesh.UVCoords[0][i3])

		edge1 := p2.Sub(p1)
		edge2 := p3.Sub(p1)
		deltaUV1 := uv2.Sub(uv1)
		deltaUV2 := uv3.Sub(uv1)

		tb := mgl32.Mat2FromRows(
			mgl32.Vec2{deltaUV2.Y(), -deltaUV1.Y()},
			mgl32.Vec2{-deltaUV2.X(), deltaUV1.X()},
		).Mul2x3(mgl32.Mat2x3FromRows(
			edge1,
			edge2,
		)).Mul(
			1.0 / (deltaUV1.X()*deltaUV2.Y() - deltaUV2.X()*deltaUV1.Y()),
		)

		normal := edge1.Cross(edge2).Normalize()
		tangent, bitangent := tb.Rows()
		tangent = tangent.Normalize()
		bitangent = bitangent.Normalize()

		// Accumulate
		normals[i1] = mgl32.Vec3(normals[i1]).Add(normal)
		normals[i2] = mgl32.Vec3(normals[i2]).Add(normal)
		normals[i3] = mgl32.Vec3(normals[i3]).Add(normal)
		tangents[i1] = mgl32.Vec3(tangents[i1]).Add(tangent)
		tangents[i2] = mgl32.Vec3(tangents[i2]).Add(tangent)
		tangents[i3] = mgl32.Vec3(tangents[i3]).Add(tangent)
		bitangents[i1] = mgl32.Vec3(bitangents[i1]).Add(bitangent)
		bitangents[i2] = mgl32.Vec3(bitangents[i2]).Add(bitangent)
		bitangents[i3] = mgl32.Vec3(bitangents[i3]).Add(bitangent)
		nIndicesPerPos[i1]++
		nIndicesPerPos[i2]++
		nIndicesPerPos[i3]++
	}
	for _, i := range indices {
		// Average
		normals[i] = mgl32.Vec3(normals[i]).Mul(1 / float32(nIndicesPerPos[i])).Normalize()
		tangents[i] = mgl32.Vec3(tangents[i]).Mul(1 / float32(nIndicesPerPos[i])).Normalize()
		bitangents[i] = mgl32.Vec3(bitangents[i]).Mul(1 / float32(nIndicesPerPos[i])).Normalize()
	}

	// NOTE(xypwn): These mesh.Normals don't seem quite right
	// (hence we calculated them manually in the previous block).
	// Adding 0.5 seems to help marginally, but many normals are
	// still messed up.
	// See "Mesh info"->"Vertex normals" in the in-app unit preview
	// after uncommenting this.
	/*normals = mesh.Normals
	for i := range normals {
		normals[i] = mgl32.Vec3(normals[i]).Sub(mgl32.Vec3{0.5, 0.5, 0.5})
	}*/

	pv.numUdims = 0
	for _, uv := range mesh.UVCoords[0] {
		udim := uint32(uv[0]) | uint32(1-uv[1])<<5
		pv.numUdims = max(pv.numUdims, udim+1)
	}
	if pv.numUdims >= 64 {
		return fmt.Errorf("expected at most 64 udims, but got %v", pv.numUdims)
	}

	// Upload object data
	{
		gl.BindVertexArray(pv.object.vao)

		positionsSize := len(mesh.Positions) * 3 * 4
		normalsSize := len(normals) * 3 * 4
		uvsSize := len(mesh.UVCoords[0]) * 2 * 4
		tangentsSize := len(tangents) * 3 * 4
		bitangentsSize := len(bitangents) * 3 * 4

		gl.BindBuffer(gl.ARRAY_BUFFER, pv.object.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, positionsSize+normalsSize+uvsSize+tangentsSize+bitangentsSize, nil, gl.STATIC_DRAW)
		offset := 0
		//
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, positionsSize, gl.Ptr(mesh.Positions))
		gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, uintptr(offset))
		gl.EnableVertexAttribArray(0)
		offset += positionsSize
		//
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, normalsSize, gl.Ptr(normals))
		gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, true, 3*4, uintptr(offset))
		gl.EnableVertexAttribArray(1)
		offset += normalsSize
		//
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, uvsSize, gl.Ptr(mesh.UVCoords[0]))
		gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, 2*4, uintptr(offset))
		gl.EnableVertexAttribArray(2)
		offset += uvsSize
		//
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, tangentsSize, gl.Ptr(tangents))
		gl.VertexAttribPointerWithOffset(3, 3, gl.FLOAT, true, 3*4, uintptr(offset))
		gl.EnableVertexAttribArray(3)
		offset += tangentsSize
		//
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, bitangentsSize, gl.Ptr(bitangents))
		gl.VertexAttribPointerWithOffset(4, 3, gl.FLOAT, true, 3*4, uintptr(offset))
		gl.EnableVertexAttribArray(4)
		offset += bitangentsSize

		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	}

	pv.meshPositions = mesh.Positions

	// Upload debug object data
	{
		gl.BindVertexArray(pv.dbgObj.vao)

		verts := pv.getAABBVertices()
		gl.BindBuffer(gl.ARRAY_BUFFER, pv.dbgObj.vbo)
		defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BufferData(gl.ARRAY_BUFFER, len(verts)*3*4, gl.Ptr(verts[:]), gl.STATIC_DRAW)

		pv.dbgObj.numIndices = int32(len(aabbIndices))
		pv.dbgObj.numVertices = int32(len(verts))
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(pv.dbgObj.numIndices*4), gl.Ptr(aabbIndices[:]), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.BindVertexArray(0)
	}

	pv.model = stingrayToGLCoords

	if pv.zoomToFitOnLoad {
		pv.zoomToFit = true
	}

	// Calculate max zoom out distance
	{
		// Get origin sphere around mesh
		var maxDistSqrFromOrigin float32
		for _, p := range pv.meshPositions {
			maxDistSqrFromOrigin = max(maxDistSqrFromOrigin,
				mgl32.Vec3(p).LenSqr())
		}
		maxDistFromOrigin := float32(math.Sqrt(float64(maxDistSqrFromOrigin)))

		// Calculate camera distance to fit vertical frustum into disk orthogonal
		// to view direction with radius of the sphere. Ideally, we'd want
		// to fit the sphere into the frustum, but the disk should be close
		// enough.
		// tan(vfov/2) = maxDistFromOrigin/viewDistance
		pv.maxViewDistance = float32(float64(maxDistFromOrigin) / math.Tan(float64(pv.vfov/2)))

		// We want to be able to zoom out a bit further.
		pv.maxViewDistance *= 2
	}

	return nil
}

func (pv *UnitPreviewState) computeMVP(aspectRatio float32) (
	normal mgl32.Mat3,
	viewPosition mgl32.Vec3,
	view mgl32.Mat4,
	projection mgl32.Mat4,
) {
	normal = pv.model.Inv().Transpose().Mat3()
	{
		mat := mgl32.Ident3()
		mat = mat.Mul3(mgl32.Rotate3DY(pv.viewRotation[0]))
		mat = mat.Mul3(mgl32.Rotate3DX(pv.viewRotation[1]))
		viewPosition = mat.Mul3x1(mgl32.Vec3{0, 0, pv.viewDistance})
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

	viewPos := imgui.CursorScreenPos()
	viewSize := imgui.ContentRegionAvail()
	viewSize.Y -= imutils.CheckboxHeight()

	widgets.GLView(name, pv.fb, viewSize,
		func() {
			io := imgui.CurrentIO()

			if imgui.IsItemActive() {
				md := io.MouseDelta()
				pv.viewRotation = pv.viewRotation.Add(mgl32.Vec2{md.X, md.Y}.Mul(-0.01))
				pv.viewRotation[1] = mgl32.Clamp(pv.viewRotation[1], -1.55, 1.55)
			}
			if imgui.IsItemHovered() {
				scroll := io.MouseWheel()
				pv.viewDistance -= 0.1 * pv.viewDistance * scroll
			}
			pv.viewDistance = mgl32.Clamp(
				pv.viewDistance,
				0.001,
				pv.maxViewDistance,
			)
		},
		func(pos, size imgui.Vec2) {
			gl.ClearColor(0.2, 0.2, 0.2, 1)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			normal, viewPosition, view, projection := pv.computeMVP(size.X / size.Y)
			mvp := projection.Mul4(view).Mul4(pv.model)

			// Draw object
			gl.Enable(gl.DEPTH_TEST)
			if !pv.showWireframe {
				gl.UseProgram(pv.objectProgram)
			} else {
				gl.UseProgram(pv.objectWireframeProgram)
			}
			gl.BindVertexArray(pv.object.vao)
			if !pv.showWireframe {
				gl.UniformMatrix4fv(pv.objectUniforms["mvp"], 1, false, &mvp[0])
				gl.UniformMatrix4fv(pv.objectUniforms["model"], 1, false, &pv.model[0])
				gl.UniformMatrix3fv(pv.objectUniforms["normalMat"], 1, false, &normal[0])
				gl.Uniform3fv(pv.objectUniforms["viewPosition"], 1, &viewPosition[0])
				gl.Uniform1iv(pv.objectUniforms["udimShown"], 64, &pv.udimsShown[0])
				gl.ActiveTexture(gl.TEXTURE0)
				gl.BindTexture(gl.TEXTURE_2D, pv.object.texAlbedo)
				gl.ActiveTexture(gl.TEXTURE1)
				gl.BindTexture(gl.TEXTURE_2D, pv.object.texNormal)
			} else {
				gl.UniformMatrix4fv(pv.objectWireframeUniforms["mvp"], 1, false, &mvp[0])
				gl.Uniform4fv(pv.objectWireframeUniforms["color"], 1, &pv.wireframeColor[0])
				gl.Uniform1iv(pv.objectWireframeUniforms["udimShown"], 64, &pv.udimsShown[0])
			}
			gl.DrawElements(gl.TRIANGLES, pv.object.numIndices, gl.UNSIGNED_INT, nil)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, 0)
			gl.BindVertexArray(0)
			gl.UseProgram(0)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

			// Draw normal visualization
			if pv.visualizeNormals {
				gl.UseProgram(pv.objectNormalVisProgram)
				gl.BindVertexArray(pv.object.vao)
				gl.UniformMatrix4fv(pv.objectNormalVisUniforms["mvp"], 1, false, &mvp[0])
				gl.Uniform1f(pv.objectNormalVisUniforms["len"], pv.viewDistance*0.02)
				gl.Uniform4fv(pv.objectNormalVisUniforms["color"], 1, &pv.visualizedNormalsColor[0])
				gl.Uniform1iv(pv.objectNormalVisUniforms["udimShown"], 64, &pv.udimsShown[0])
				gl.DrawElements(gl.TRIANGLES, pv.object.numIndices, gl.UNSIGNED_INT, nil)
				gl.BindVertexArray(0)
				gl.UseProgram(0)
			}

			// Draw debug object
			if pv.showAABB {
				gl.Disable(gl.DEPTH_TEST)
				gl.UseProgram(pv.dbgObjProgram)
				gl.BindVertexArray(pv.dbgObj.vao)
				{
					aabbMVP := mvp.Mul4(pv.aabbMat)
					gl.UniformMatrix4fv(pv.dbgObjUniforms["mvp"], 1, false, &aabbMVP[0])
				}
				gl.Uniform4fv(pv.dbgObjUniforms["color"], 1, &pv.aabbColor[0])
				gl.DrawElements(gl.TRIANGLES, pv.dbgObj.numIndices, gl.UNSIGNED_INT, nil)
				gl.BindVertexArray(0)
				gl.UseProgram(0)
			}

			if pv.zoomToFit {
				pv.viewDistance = pv.maxViewDistance

				_, viewPosition, view, projection := pv.computeMVP(size.X / size.Y)

				fitVertexCamDistDelta := func(vertex mgl32.Vec3) float32 {
					v := vertex.Vec4(1.0)
					v = pv.model.Mul4x1(v)

					// NOTE(xypwn): I think the projections are still off, but
					// this whole code seems to at least do what I wanted it
					// to now.
					projV := projection.Mul4x1(view.Mul4x1(v))
					projV = projV.Mul(1 / projV.W())

					// Component with maximum distance from screen center
					maxDist := max(mgl32.Abs(projV.X()), mgl32.Abs(projV.Y()))

					// Calculate orthogonal distance from camera to selected vertex
					var od float32
					{
						viewDir := mgl32.Vec3{}.Sub(viewPosition).Normalize()
						vert := v
						vert = vert.Mul(1 / vert.W())
						camToVert := vert.Vec3().Sub(viewPosition)
						od = viewDir.Dot(camToVert)
					}

					// Fit model to screen
					return od * (maxDist - 1)
				}

				// NOTE(xypwn): I used to use the AABB vertices for this, but they would often be
				// wrong. Using all of the mesh positions instead takes no more than ~10ms on
				// all of the models I've tried.
				maxCamDistDelta := float32(-math.MaxFloat32)
				for _, vert := range pv.meshPositions {
					maxCamDistDelta = max(maxCamDistDelta,
						fitVertexCamDistDelta(vert))
				}
				pv.viewDistance += maxCamDistDelta
				pv.viewDistance *= 1.02

				pv.zoomToFit = false
			}
		},
		func(pos, size imgui.Vec2) {
			dl := imgui.WindowDrawList()

			// Scale indicator
			{
				// Screen size in world here refers to how large the screen
				// content rectangle would be if it intersected
				// the origin.
				// tan(vfov/2) = screenHeightInWorld/camDist
				screenHeightInWorld := float32(math.Tan(float64(pv.vfov/2)) * float64(pv.viewDistance))
				screenWidthInWorld := screenHeightInWorld / size.Y * size.X
				indicatorWidthInWorld := screenWidthInWorld / 2
				{
					order := float32(
						math.Pow(
							10,
							math.Floor(math.Log10(float64(indicatorWidthInWorld)))-1,
						),
					)
					indicatorWidthInWorld = order * float32(math.Floor(float64(indicatorWidthInWorld/order)))
				}
				indicatorColor := imgui.ColorU32Col(imgui.ColText)

				indicatorWidth := size.X * indicatorWidthInWorld / screenWidthInWorld
				indicatorPos := pos.Add(imgui.NewVec2(10, 10))
				dl.AddRectFilled(
					indicatorPos.Add(imgui.NewVec2(0, 0)),
					indicatorPos.Add(imgui.NewVec2(2, 10)),
					indicatorColor,
				)
				dl.AddRectFilled(
					indicatorPos.Add(imgui.NewVec2(0, 4)),
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 6)),
					indicatorColor,
				)
				dl.AddRectFilled(
					indicatorPos.Add(imgui.NewVec2(indicatorWidth-2, 0)),
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 10)),
					indicatorColor,
				)

				var dimPrefix string
				var dim float32
				if indicatorWidthInWorld >= 1e3 {
					dimPrefix = "k"
					dim = 1e3
				} else if indicatorWidthInWorld >= 1 {
					dimPrefix = ""
					dim = 1
				} else if indicatorWidthInWorld >= 1e-2 {
					dimPrefix = "c"
					dim = 1e-2
				} else if indicatorWidthInWorld >= 1e-3 {
					dimPrefix = "m"
					dim = 1e-3
				} else {
					dimPrefix = "µ"
					dim = 1e-6
				}
				text := fmt.Sprintf(
					"%v%vm",
					strings.TrimRight(strings.TrimRight(
						fmt.Sprintf("%.3f", indicatorWidthInWorld/dim),
						"0"), "."),
					dimPrefix,
				)
				textSize := imgui.CalcTextSize(text)
				textPos := indicatorPos.Add(imgui.NewVec2(indicatorWidth/2-textSize.X/2, 12))
				dl.AddRectFilled(
					textPos.Add(imgui.NewVec2(-4, 0)),
					textPos.Add(textSize).Add(imgui.NewVec2(4, 0)),
					imgui.ColorU32Vec4(imgui.NewVec4(0, 0, 0, 0.5)),
				)
				dl.AddTextVec2(
					textPos,
					indicatorColor,
					text,
				)
			}
		},
	)

	if imgui.Button(fnt.I("Home")) {
		pv.viewRotation = mgl32.Vec2{}
		pv.zoomToFit = true
	}
	imgui.SetItemTooltip("Reset view")
	imgui.SameLine()
	if imgui.Button(fnt.I("Data_object")) {
		imgui.OpenPopupStr("Debug info")
	}
	imgui.SetItemTooltip("Mesh debug info...")
	if imgui.BeginPopup("Debug info") {
		imgui.TextUnformatted("Mesh info")
		imgui.Indent()
		imutils.Textf("Indices: %v", pv.object.numIndices)
		imutils.Textf("Vertices: %v", pv.object.numVertices)
		imutils.Textf("Triangles: %v", pv.object.numIndices/3)
		imgui.Unindent()

		imgui.Separator()

		const colorPickerFlags = imgui.ColorEditFlagsNoInputs | imgui.ColorEditFlagsAlphaBar | imgui.ColorEditFlagsNoLabel
		imgui.TextUnformatted("Display")
		imgui.Indent()
		imgui.Checkbox("Wireframe mode", &pv.showWireframe)
		imgui.SameLineV(200, -1)
		imgui.ColorEdit4V("Wireframe color", &pv.wireframeColor, colorPickerFlags)

		imgui.Checkbox("Bounding box", &pv.showAABB)
		imgui.SameLine()
		imgui.TextUnformatted(fnt.I("Warning"))
		imgui.SetItemTooltip("Bounding boxes are known to sometimes be wrong")
		imgui.SameLineV(200, -1)
		imgui.ColorEdit4V("Bounding box color", &pv.aabbColor, colorPickerFlags)

		imgui.Checkbox("Vertex normals", &pv.visualizeNormals)
		imgui.SameLineV(200, -1)
		imgui.ColorEdit4V("Vertex normals color", &pv.visualizedNormalsColor, colorPickerFlags)
		imgui.Unindent()

		imgui.EndPopup()
	}
	imgui.SameLine()
	if imgui.Checkbox("Auto-zoom on load", &pv.zoomToFitOnLoad) && pv.zoomToFitOnLoad {
		pv.zoomToFit = true
	}
	imgui.SameLine()
	// UDim selection
	nextActiveUDimListItem := int32(-1)
	nextHoveredUDimListItem := int32(-1)
	imgui.BeginDisabledV(pv.numUdims <= 1)
	if imgui.Button("UDims Selection") {
		imgui.OpenPopupStr("UDims")
		imgui.SetNextWindowPos(viewPos.Sub(imgui.NewVec2(150, 0)))
	}
	if pv.numUdims <= 1 {
		imgui.SetItemTooltip("Mesh has no UDims")
	}
	imgui.EndDisabled()
	imgui.SetNextWindowSize(imgui.NewVec2(150, viewSize.Y))
	if imgui.BeginPopup("UDims") {
		if imgui.Button("Reset") {
			for i := range pv.udimsSelected {
				pv.udimsSelected[i] = true
			}
		}
		imgui.Separator()
		imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing,
			imgui.NewVec2(imgui.CurrentStyle().ItemSpacing().X, 0))
		dragging := pv.activeUDimListItem != -1 && pv.hoveredUDimListItem != -1
		var draggingMin, draggingMax int32
		if dragging {
			draggingMin = min(pv.activeUDimListItem, pv.hoveredUDimListItem)
			draggingMax = max(pv.activeUDimListItem, pv.hoveredUDimListItem)
		}
		var draggingMinPos, draggingMaxPos imgui.Vec2
		for i := range int32(pv.numUdims) {
			selected := pv.udimsSelected[i]
			if dragging {
				if i >= draggingMin && i <= draggingMax {
					selected = !selected
				}
				if imgui.IsMouseClickedBool(imgui.MouseButtonRight) {
					imgui.CurrentContext().SetActiveId(0)
				}
			}
			if selected {
				pv.udimsShown[i] = 1
			} else {
				pv.udimsShown[i] = 0
			}
			if imgui.IsMouseReleased(imgui.MouseButtonLeft) {
				pv.udimsSelected[i] = selected
			}
			var icon string
			if selected {
				icon = fnt.I("Visibility")
			} else {
				icon = fnt.I("Visibility_off")
			}
			imgui.PushIDInt(i)
			pos := imgui.CursorScreenPos()
			size := imgui.NewVec2(imgui.ContentRegionAvail().X, imgui.FontSize())
			if dragging {
				if i == draggingMin {
					draggingMinPos = pos
				}
				if i == draggingMax {
					draggingMaxPos = pos.Add(size)
				}
			}
			if selected {
				imgui.WindowDrawList().AddRectFilled(pos, pos.Add(size), imgui.ColorU32Col(imgui.ColButton))
			}
			imutils.Textf(fmt.Sprintf("%v %v", icon, i))
			imgui.SetCursorScreenPos(pos)
			imgui.SetNextItemAllowOverlap()
			imgui.InvisibleButton("btn", size)
			if imgui.IsItemActive() {
				nextActiveUDimListItem = i
			}
			hovered := imgui.ItemStatusFlags(imgui.CurrentContext().LastItemData().CData.StatusFlags)&imgui.ItemStatusFlagsHoveredRect != 0
			if hovered {
				nextHoveredUDimListItem = i
			}
			imgui.SetItemTooltip(`Click to toggle item visibility
Drag to toggle multiple items (right-click to cancel)`)
			imgui.PopID()
		}
		imgui.PopStyleVar()
		if dragging {
			imgui.WindowDrawList().AddRectV(draggingMinPos, draggingMaxPos, imgui.ColorU32Col(imgui.ColButtonActive), 0, 0, 2)
		}
		imgui.EndPopup()
	} else {
		for i := range pv.udimsShown {
			if pv.udimsSelected[i] {
				pv.udimsShown[i] = 1
			} else {
				pv.udimsShown[i] = 0
			}
		}
	}
	pv.activeUDimListItem = nextActiveUDimListItem
	pv.hoveredUDimListItem = nextHoveredUDimListItem

}
