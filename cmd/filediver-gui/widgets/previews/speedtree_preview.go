package previews

import (
	"bytes"
	"cmp"
	"fmt"
	"image"
	"io"
	"math"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/speedtree"
	"github.com/xypwn/filediver/stingray/unit"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

type speedtreeMaterial struct {
	texAlbedoOpacity uint32
	texNormalRG      uint32

	indexOffset      int32
	indexCount       int32
	opacityThreshold float32
}

type speedtreePreviewObject struct {
	vao          uint32 // vertex array object
	ibo          uint32 // index buffer object
	vbo          uint32 // vertex buffer object
	lutFibonacci uint32

	numVertices int32
	numIndices  int32
	indexType   uint32
	indexStride int32
	materials   []speedtreeMaterial
}

// NOTE(xypwn): We do at most ~10 lookups once per frame,
// so it should be fine to store this in a string map.
type speedtreePreviewUniforms map[string]int32

func (obj *speedtreePreviewObject) genObjects() {
	gl.GenVertexArrays(1, &obj.vao)
	gl.GenBuffers(1, &obj.vbo)
	gl.GenBuffers(1, &obj.ibo)
	gl.GenTextures(1, &obj.lutFibonacci)

	gl.BindVertexArray(obj.vao)
	defer gl.BindVertexArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, obj.vbo)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, obj.ibo)
}

func (obj *speedtreePreviewObject) genMaterials(count int) {
	obj.materials = make([]speedtreeMaterial, count)
	for idx := range obj.materials {
		gl.GenTextures(1, &obj.materials[idx].texAlbedoOpacity)
		gl.GenTextures(1, &obj.materials[idx].texNormalRG)
	}
}

// Panicks if a name is not a uniform.
func (uniforms *speedtreePreviewUniforms) generate(program uint32, names ...string) {
	if *uniforms == nil {
		*uniforms = speedtreePreviewUniforms{}
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

func (obj speedtreePreviewObject) deleteObjects() {
	gl.DeleteVertexArrays(1, &obj.vao)
	gl.DeleteBuffers(1, &obj.vbo)
	gl.DeleteBuffers(1, &obj.ibo)
	gl.DeleteTextures(1, &obj.lutFibonacci)
	for _, material := range obj.materials {
		gl.DeleteTextures(1, &material.texAlbedoOpacity)
		gl.DeleteTextures(1, &material.texNormalRG)
	}
}

type SpeedtreePreviewState struct {
	fb *widgets.GLViewState

	object                  speedtreePreviewObject
	treeProgram             uint32
	treeUniforms            speedtreePreviewUniforms
	objectWireframeProgram  uint32
	objectWireframeUniforms speedtreePreviewUniforms

	objectNormalVisProgram  uint32
	objectNormalVisUniforms speedtreePreviewUniforms

	dbgObjProgram  uint32
	dbgObj         speedtreePreviewObject
	dbgObjUniforms speedtreePreviewUniforms

	vfov         float32
	model        mgl32.Mat4
	viewDistance float32
	viewRotation mgl32.Vec2 // {yaw, pitch}

	// Previous view distance and rotation (for view animation)
	animOrigViewDistance float32
	animOrigViewRotation mgl32.Vec2
	animTime             float32 // range [0;1], -1 when not animating

	// Axis-aligned bounding box. Don't forget
	// to multiply aabb's vertices with aabbMat first!
	aabb    [2]mgl32.Vec3
	aabbMat mgl32.Mat4

	// // For fitting mesh to screen and debug info
	// meshPositions [][3]float32
	// meshNormals   [][3]float32

	maxViewDistance float32

	numUdims          uint32
	udimsShownDefault [64]bool
	udimsSelected     [64]bool  // udims persistently selected
	udimsShown        [64]int32 // udims visually selected 1 (shown) or 0 (hidden)
	udimNames         [64]string

	// For dragging selection
	activeUDimListItem  int32
	hoveredUDimListItem int32

	showWireframe             bool
	wireframeColor            [4]float32
	showAABB                  bool
	aabbColor                 [4]float32
	visualizeNormals          bool
	visualizeTangentBitangent int32 // 1 or 0
	autoZoomEnabled           bool
	doAutoZoomNextFrame       bool
}

func NewSpeedtreePreview() (*SpeedtreePreviewState, error) {
	var err error

	pv := &SpeedtreePreviewState{}

	pv.fb, err = widgets.NewGLView()
	if err != nil {
		return nil, err
	}

	pv.object.genObjects()
	pv.treeProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/speedtree.vert",
		"shaders/speedtree.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.treeUniforms.generate(pv.treeProgram, "mvp", "model", "normalMat", "viewPosition", "texAlbedo", "texNormal", "opacityThreshold", "fibonacci_normal_lut")

	pv.objectWireframeProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/speedtree_wireframe.vert",
		"shaders/object_wireframe.geom",
		"shaders/object_wireframe.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.objectWireframeUniforms.generate(pv.objectWireframeProgram, "mvp", "color")

	pv.objectNormalVisProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/speedtree_normal_vis.vert",
		"shaders/object_normal_vis.geom",
		"shaders/object_normal_vis.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.objectNormalVisUniforms.generate(pv.objectNormalVisProgram, "mvp", "len", "showTangentBitangent", "fibonacci_normal_lut")

	pv.dbgObj.genObjects()
	pv.dbgObjProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/debug_object.vert",
		"shaders/debug_object.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.dbgObjUniforms.generate(pv.dbgObjProgram, "mvp")

	setupLUT := func(textureID uint32) {
		gl.BindTexture(gl.TEXTURE_2D, textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	setupLUT(pv.object.lutFibonacci)

	gl.UseProgram(pv.treeProgram)
	gl.Uniform1i(pv.treeUniforms["texAlbedo"], 0)
	gl.Uniform1i(pv.treeUniforms["texNormal"], 1)
	gl.Uniform1i(pv.treeUniforms["fibonacci_normal_lut"], 2)
	gl.UseProgram(0)

	gl.UseProgram(pv.objectNormalVisProgram)
	gl.Uniform1i(pv.objectNormalVisUniforms["fibonacci_normal_lut"], 2)
	gl.UseProgram(0)

	pv.vfov = mgl32.DegToRad(60)
	pv.viewDistance = 25

	pv.wireframeColor = [4]float32{1.0, 1.0, 1.0, 0.5}
	pv.aabbColor = [4]float32{0.3, 0.3, 0.8, 0.2}

	return pv, nil
}

func (pv *SpeedtreePreviewState) Delete() {
	pv.fb.Delete()
	gl.DeleteProgram(pv.treeProgram)
	pv.object.deleteObjects()
	pv.dbgObj.deleteObjects()
}

func (pv *SpeedtreePreviewState) loadMesh(meshInfos []unit.MeshInfo, meshLayouts []unit.MeshLayout, gpuData []byte) (unit.Mesh, error) {
	var meshToLoad uint32
	{
		highestDetailIdx := -1
		highestDetailCount := -1
		for i, info := range meshInfos {
			for _, group := range info.Groups {
				if int(group.NumIndices) > highestDetailCount && info.Header.MeshType != unit.MeshTypeUnknown00 {
					highestDetailIdx = i
					highestDetailCount = int(group.NumIndices)
				}
			}
		}
		if highestDetailIdx == -1 {
			return unit.Mesh{}, fmt.Errorf("unable to find mesh to load")
		}
		meshToLoad = uint32(highestDetailIdx)
	}

	var mesh unit.Mesh
	{
		meshes, err := unit.LoadMeshes(bytes.NewReader(gpuData), meshInfos, meshLayouts, []uint32{meshToLoad})
		if err != nil {
			return unit.Mesh{}, err
		}
		mesh = meshes[meshToLoad]
	}
	return mesh, nil
}

func (pv *SpeedtreePreviewState) LoadSpeedtree(fileID stingray.Hash, mainData, gpuData []byte, getResource GetResourceFunc, thinhashes map[stingray.ThinHash]string) error {
	info, err := speedtree.LoadSpeedTree(bytes.NewReader(mainData))
	if err != nil {
		return err
	}

	pv.aabb = [2]mgl32.Vec3{
		info.Extents[0],
		info.Extents[1],
	}
	pv.aabbMat = mgl32.Ident4()
	lod0 := info.IndexDefinitions[0]

	// Upload object data
	{
		gl.BindVertexArray(pv.object.vao)

		vertexDef := info.VertexDefinitions[lod0.VertexDef]

		gl.BindBuffer(gl.ARRAY_BUFFER, pv.object.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, int(vertexDef.Count*vertexDef.Stride), gl.Ptr(&gpuData[vertexDef.Offset]), gl.STATIC_DRAW)

		offset := uintptr(0)
		for idx, attr := range info.VertexXML.Stream.Attribute {
			var gltype uint32
			var size uint32
			switch attr.Type {
			case "byte":
				gltype = gl.BYTE
				size = 1
			case "ubyte":
				gltype = gl.UNSIGNED_BYTE
				size = 1
			case "short":
				gltype = gl.SHORT
				size = 2
			case "ushort":
				gltype = gl.UNSIGNED_SHORT
				size = 2
			case "int":
				gltype = gl.INT
				size = 4
			case "uint":
				gltype = gl.UNSIGNED_INT
				size = 4
			case "half":
				gltype = gl.HALF_FLOAT
				size = 2
			case "float":
				gltype = gl.FLOAT
				size = 4
			case "double":
				gltype = gl.DOUBLE
				size = 8
			default:
				gltype = gl.FLOAT
				size = 4
			}

			normalize := false
			if attr.Normalize != nil {
				normalize = *attr.Normalize
			}
			switch gltype {
			case gl.BYTE, gl.UNSIGNED_BYTE, gl.SHORT, gl.UNSIGNED_SHORT, gl.INT, gl.UNSIGNED_INT:
				if !normalize {
					gl.VertexAttribIPointerWithOffset(uint32(idx), int32(attr.Count), gltype, int32(vertexDef.Stride), offset)
				} else {
					gl.VertexAttribPointerWithOffset(uint32(idx), int32(attr.Count), gltype, normalize, int32(vertexDef.Stride), offset)
				}
			default:
				gl.VertexAttribPointerWithOffset(uint32(idx), int32(attr.Count), gltype, normalize, int32(vertexDef.Stride), offset)
			}
			gl.EnableVertexAttribArray(uint32(idx))
			offset += uintptr(size * uint32(attr.Count))
		}
		pv.object.numVertices = int32(vertexDef.Count)

		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(lod0.Count*lod0.Stride), gl.Ptr(&gpuData[lod0.Offset]), gl.STATIC_DRAW)
		pv.object.numIndices = int32(lod0.Count)
		pv.object.indexStride = int32(lod0.Stride)
		switch lod0.Stride {
		case 2:
			pv.object.indexType = gl.UNSIGNED_SHORT
		case 4:
			pv.object.indexType = gl.UNSIGNED_INT
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	}

	// upload debug object data
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

	uploadStingrayTexture := func(textureID uint32, fileName stingray.Hash) error {
		file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
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

	uploadStingrayLUT := func(textureID uint32, fileName stingray.Hash) error {
		file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
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
		gl.BindTexture(gl.TEXTURE_2D, textureID)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, int32(dds.Image.Bounds().Dx()), int32(dds.Image.Bounds().Dy()), 0, gl.RGBA, gl.FLOAT, gl.Ptr(dds.Images[0].MipMaps[0].Raw))
		gl.BindTexture(gl.TEXTURE_2D, 0)
		return nil
	}

	setupTexture := func(textureID uint32) {
		gl.BindTexture(gl.TEXTURE_2D, textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}

	getTextureSlotPath := func(mat *material.Material, targetSlot stingray.ThinHash) stingray.Hash {
		for texSlot, path := range mat.Textures {
			if texSlot == targetSlot {
				return path
			}
		}
		return stingray.Hash{Value: 0x0}
	}

	uploadMissingTexture := func(textureID uint32, color []byte) error {
		gl.BindTexture(gl.TEXTURE_2D, textureID)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(color))
		gl.BindTexture(gl.TEXTURE_2D, 0)
		return nil
	}

	// Load materials
	{
		pv.object.genMaterials(int(lod0.MeshCount))
		materialIndex := 0
		for _, meshDef := range info.MeshDefinitions[lod0.MeshOffset : lod0.MeshOffset+lod0.MeshCount] {
			var treeMaterial speedtree.MaterialDefinition
			for _, matDef := range info.StingrayMaterials {
				if matDef.Index == uint64(meshDef.Material) {
					treeMaterial = matDef
					break
				}
			}
			matData, ok, err := getResource(stingray.FileID{
				Name: treeMaterial.Path,
				Type: stingray.Sum("material"),
			}, stingray.DataMain)
			if err != nil {
				return fmt.Errorf("load material %v.material: %w", treeMaterial.Path, err)
			}
			if !ok {
				return fmt.Errorf("load material %v.material does not exist", treeMaterial.Path)
			}

			mat, err := material.LoadMain(bytes.NewReader(matData))
			if err != nil {
				return fmt.Errorf("load material %v.material: %w", treeMaterial.Path, err)
			}
			opacityThreshold, ok := mat.Settings[stingray.Sum("opacity_threshold").Thin()]
			if !ok {
				opacityThreshold = []float32{0.5}
			}
			pv.object.materials[materialIndex].opacityThreshold = opacityThreshold[0]

			setupTexture(pv.object.materials[materialIndex].texAlbedoOpacity)
			setupTexture(pv.object.materials[materialIndex].texNormalRG)

			albedoOpacityHash := getTextureSlotPath(mat, stingray.Sum("tex0").Thin())
			if albedoOpacityHash.Value == 0x0 {
				err = uploadMissingTexture(pv.object.materials[materialIndex].texAlbedoOpacity, []byte{255, 255, 255, 255})
			} else {
				err = uploadStingrayTexture(pv.object.materials[materialIndex].texAlbedoOpacity, albedoOpacityHash)
			}
			if err != nil {
				return fmt.Errorf("load tex0 %v.texture: %w", albedoOpacityHash, err)
			}

			normalHash := getTextureSlotPath(mat, stingray.Sum("tex1").Thin())
			if normalHash.Value == 0x0 {
				err = uploadMissingTexture(pv.object.materials[materialIndex].texNormalRG, []byte{128, 128, 255, 255})
			} else {
				err = uploadStingrayTexture(pv.object.materials[materialIndex].texNormalRG, normalHash)
			}
			if err != nil {
				return fmt.Errorf("load tex1 %v.texture: %w", normalHash, err)
			}

			pv.object.materials[materialIndex].indexCount = int32(meshDef.IndexCount)
			pv.object.materials[materialIndex].indexOffset = int32(meshDef.IndexOffset)

			materialIndex++
		}
	}

	// Load fibonacci lut
	{
		err = uploadStingrayLUT(pv.object.lutFibonacci, stingray.Sum("content/art_shared/textures/fibonacci_normal_lut"))
		if err != nil {
			return fmt.Errorf("upload fibonacci normal lut: %w", err)
		}
	}

	pv.model = stingrayToGLCoords

	if pv.autoZoomEnabled {
		pv.doAutoZoomNextFrame = true
	}

	// Calculate max zoom out distance
	{
		// Get origin sphere around mesh
		var maxDistSqrFromOrigin float32
		for _, p := range pv.getAABBVertices() {
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

	for i := range pv.udimsShownDefault {
		pv.udimsShownDefault[i] = true
		pv.udimNames[i] = ""
	}
	visibilityMasks, err := datalib.ParseVisibilityMasks()
	if err != nil {
		return err
	}
	if visibilityMask, ok := visibilityMasks[fileID]; ok {
		for _, info := range visibilityMask.MaskInfos {
			if int(info.Index) >= len(pv.udimsShownDefault) {
				// No support for udims with index > 64 at the moment
				continue
			}
			pv.udimsShownDefault[info.Index] = info.StartHidden == 0
			name, ok := thinhashes[info.Name]
			if !ok {
				name = info.Name.String()
			}
			pv.udimNames[info.Index] = name
		}
	}
	pv.udimsSelected = pv.udimsShownDefault

	return nil
}

func (pv *SpeedtreePreviewState) computeMVP(aspectRatio float32, animate bool) (
	normal mgl32.Mat3,
	viewPosition mgl32.Vec3,
	view mgl32.Mat4,
	projection mgl32.Mat4,
) {
	var viewDistance float32
	var viewRotation mgl32.Vec2

	if animate && pv.animTime >= 0 && pv.animTime <= 1 {
		// Animate -> lerp original to current by animTime
		viewDistance = pv.animOrigViewDistance*(1-pv.animTime) + pv.viewDistance*pv.animTime
		viewRotation = pv.animOrigViewRotation.Mul(1 - pv.animTime).Add(pv.viewRotation.Mul(pv.animTime))
	} else {
		viewDistance = pv.viewDistance
		viewRotation = pv.viewRotation
	}

	normal = pv.model.Inv().Transpose().Mat3()
	{
		mat := mgl32.Ident3()
		mat = mat.Mul3(mgl32.Rotate3DY(viewRotation[0]))
		mat = mat.Mul3(mgl32.Rotate3DX(viewRotation[1]))
		viewPosition = mat.Mul3x1(mgl32.Vec3{0, 0, viewDistance})
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

func (pv *SpeedtreePreviewState) getAABBVertices() [8]mgl32.Vec3 {
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

func SpeedtreePreview(name string, pv *SpeedtreePreviewState) {
	if pv.object.numIndices == 0 {
		return
	}

	imgui.PushIDStr(name)
	defer imgui.PopID()

	viewPos := imgui.CursorScreenPos()
	viewSize := imgui.ContentRegionAvail()
	viewSize.Y -= imutils.CheckboxHeight()

	if pv.animTime == -1 || pv.animTime >= 1 {
		pv.animOrigViewDistance = pv.viewDistance
		pv.animOrigViewRotation = pv.viewRotation
		pv.animTime = -1
	}

	widgets.GLView(name, pv.fb, viewSize,
		func() {
			io := imgui.CurrentIO()

			if imgui.IsItemActive() {
				md := io.MouseDelta()
				pv.viewRotation = pv.viewRotation.Add(mgl32.Vec2{md.X, md.Y}.Mul(-0.01))
				pv.viewRotation[1] = mgl32.Clamp(pv.viewRotation[1], -1.55, 1.55)
			}
			if imgui.IsItemDeactivated() && pv.autoZoomEnabled {
				pv.doAutoZoomNextFrame = true
			}
			if imgui.IsItemHovered() {
				scroll := io.MouseWheel()
				pv.viewDistance -= 0.1 * pv.viewDistance * scroll
				if scroll != 0 {
					pv.autoZoomEnabled = false
				}
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

			normal, viewPosition, view, projection := pv.computeMVP(size.X/size.Y, true)
			mvp := projection.Mul4(view).Mul4(pv.model)

			// Draw object
			gl.Enable(gl.DEPTH_TEST)
			if !pv.showWireframe {
				gl.UseProgram(pv.treeProgram)
			} else {
				gl.UseProgram(pv.objectWireframeProgram)
			}
			gl.BindVertexArray(pv.object.vao)
			if !pv.showWireframe {
				gl.UniformMatrix4fv(pv.treeUniforms["mvp"], 1, false, &mvp[0])
				gl.UniformMatrix4fv(pv.treeUniforms["model"], 1, false, &pv.model[0])
				gl.UniformMatrix3fv(pv.treeUniforms["normalMat"], 1, false, &normal[0])
				gl.Uniform3fv(pv.treeUniforms["viewPosition"], 1, &viewPosition[0])
			} else {
				gl.UniformMatrix4fv(pv.objectWireframeUniforms["mvp"], 1, false, &mvp[0])
				gl.Uniform4fv(pv.objectWireframeUniforms["color"], 1, &pv.wireframeColor[0])
			}
			for idx := range pv.object.materials {
				if !pv.showWireframe {
					gl.Uniform1f(pv.treeUniforms["opacityThreshold"], pv.object.materials[idx].opacityThreshold)
					gl.ActiveTexture(gl.TEXTURE0)
					gl.BindTexture(gl.TEXTURE_2D, pv.object.materials[idx].texAlbedoOpacity)
					gl.ActiveTexture(gl.TEXTURE1)
					gl.BindTexture(gl.TEXTURE_2D, pv.object.materials[idx].texNormalRG)
					gl.ActiveTexture(gl.TEXTURE2)
					gl.BindTexture(gl.TEXTURE_2D, pv.object.lutFibonacci)
				}
				gl.DrawElements(gl.TRIANGLES, pv.object.materials[idx].indexCount, pv.object.indexType, gl.Ptr(uintptr(pv.object.indexStride*pv.object.materials[idx].indexOffset)))
			}
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, 0)
			gl.BindVertexArray(0)
			gl.UseProgram(0)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

			// Draw normal visualization
			if pv.visualizeNormals {
				gl.UseProgram(pv.objectNormalVisProgram)
				gl.BindVertexArray(pv.object.vao)
				gl.ActiveTexture(gl.TEXTURE2)
				gl.BindTexture(gl.TEXTURE_2D, pv.object.lutFibonacci)
				gl.UniformMatrix4fv(pv.objectNormalVisUniforms["mvp"], 1, false, &mvp[0])
				gl.Uniform1f(pv.objectNormalVisUniforms["len"], pv.viewDistance*0.02)
				gl.Uniform1iv(pv.objectNormalVisUniforms["showTangentBitangent"], 1, &pv.visualizeTangentBitangent)
				gl.DrawElements(gl.POINTS, pv.object.numIndices, pv.object.indexType, nil) // TODO: Make this not draw duplicate vertices
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

			if pv.doAutoZoomNextFrame {
				pv.viewDistance = pv.maxViewDistance

				_, viewPosition, view, projection := pv.computeMVP(size.X/size.Y, false)

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

				maxCamDistDelta := float32(-math.MaxFloat32)
				for _, vert := range pv.getAABBVertices() {
					maxCamDistDelta = max(maxCamDistDelta,
						fitVertexCamDistDelta(vert))
				}
				pv.viewDistance += maxCamDistDelta
				pv.viewDistance *= 1.02

				pv.doAutoZoomNextFrame = false

				pv.animTime = 0
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
				indicatorPos := pos.Add(imutils.SVec2(10, 10))
				dl.AddRectFilled(
					indicatorPos.Add(imutils.SVec2(0, 0)),
					indicatorPos.Add(imutils.SVec2(2, 10)),
					indicatorColor,
				)
				dl.AddRectFilled(
					indicatorPos.Add(imutils.SVec2(0, 4)),
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 0).Add(imutils.SVec2(0, 6))),
					indicatorColor,
				)
				dl.AddRectFilled(
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 0).Add(imutils.SVec2(-2, 0))),
					indicatorPos.Add(imgui.NewVec2(indicatorWidth, 0).Add(imutils.SVec2(0, 10))),
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
				textPos := indicatorPos.Add(imgui.NewVec2(indicatorWidth/2-textSize.X/2, 0)).Add(imutils.SVec2(0, 12))
				dl.AddRectFilled(
					textPos.Add(imutils.SVec2(-4, 0)),
					textPos.Add(textSize).Add(imutils.SVec2(4, 0)),
					imgui.ColorU32Vec4(imgui.NewVec4(0, 0, 0, 0.5)),
				)
				dl.AddTextVec2(
					textPos,
					indicatorColor,
					text,
				)

			}

			// _, _, view, projection := pv.computeMVP(size.X/size.Y, false)
			// mvp := projection.Mul4(view).Mul4(pv.model)

			// // Show hovered vertex info
			// if pv.visualizeNormals {
			// 	igRelMousePos := imgui.MousePos().Sub(pos)
			// 	mousePos := mgl32.Vec2{
			// 		igRelMousePos.X,
			// 		igRelMousePos.Y,
			// 	}
			// 	var closestPos mgl32.Vec2
			// 	closestDist := float32(math.MaxFloat32)
			// 	var closestIdx int
			// 	for i, vtx := range pv.meshPositions {
			// 		v := mvp.Mul4x1(mgl32.Vec3(vtx).Vec4(1.0))
			// 		v = v.Mul(1 / v.W())
			// 		v[0] = (v[0] + 1) * size.X * 0.5
			// 		v[1] = (-v[1] + 1) * size.Y * 0.5
			// 		dist := mousePos.Sub(v.Vec2()).LenSqr()
			// 		if dist < closestDist {
			// 			closestPos = v.Vec2()
			// 			closestDist = dist
			// 			closestIdx = i
			// 		}
			// 	}
			// 	markerPos := pos.Add(imgui.NewVec2(closestPos.X(), closestPos.Y()))
			// 	dl.AddCircleFilled(
			// 		markerPos,
			// 		imutils.S(2),
			// 		imgui.ColorU32Vec4(imgui.NewVec4(1, 0, 0, 1)),
			// 	)
			// 	dl.AddTextVec2(
			// 		markerPos,
			// 		imgui.ColorU32Vec4(imgui.NewVec4(1, 1, 0, 1)),
			// 		fmt.Sprintf("Pos: %v\nNormal: %v", pv.meshPositions[closestIdx], pv.meshNormals[closestIdx]),
			// 	)
			// }
		},
	)

	if imgui.Button(fnt.I.Home) {
		pv.viewRotation = mgl32.Vec2{}
		pv.doAutoZoomNextFrame = true
		pv.animTime = 0
	}
	imgui.SetItemTooltip("Reset view")
	imgui.SameLine()
	if imgui.Button(fnt.I.DataObject) {
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
		imgui.SameLineV(imutils.S(170), -1)
		imgui.ColorEdit4V("Wireframe color", &pv.wireframeColor, colorPickerFlags)

		imgui.Checkbox("Bounding box", &pv.showAABB)
		imgui.SameLine()
		imgui.TextUnformatted(fnt.I.Warning)
		imgui.SetItemTooltip("Bounding boxes are known to sometimes be wrong")
		imgui.SameLineV(imutils.S(170), -1)
		imgui.ColorEdit4V("Bounding box color", &pv.aabbColor, colorPickerFlags)

		imgui.Checkbox("Vertex normals", &pv.visualizeNormals)
		imgui.SetItemTooltip("Normal is blue")
		{
			imgui.BeginDisabledV(!pv.visualizeNormals)
			check := pv.visualizeTangentBitangent != 0
			imgui.Checkbox("Vertex tangent and bitangent", &check)
			if pv.visualizeNormals {
				imgui.SetItemTooltip("Tangent is red, bitangent is green")
			} else {
				imgui.SetItemTooltip("Requires normals to be shown")
			}
			pv.visualizeTangentBitangent = 0
			if check {
				pv.visualizeTangentBitangent = 1
			}
			imgui.EndDisabled()
		}
		imgui.Unindent()

		imgui.EndPopup()
	}
	imgui.SameLine()
	if imgui.Checkbox(fnt.I.AllOut+" Auto-zoom", &pv.autoZoomEnabled) && pv.autoZoomEnabled {
		pv.doAutoZoomNextFrame = true
	}
	imgui.SameLine()
	// UDim selection
	nextActiveUDimListItem := int32(-1)
	nextHoveredUDimListItem := int32(-1)
	imgui.BeginDisabledV(pv.numUdims <= 1)
	if imgui.Button("UDims Selection") {
		imgui.OpenPopupStr("UDims")
		imgui.SetNextWindowPos(viewPos.Sub(imutils.SVec2(240, 0)))
		imgui.SetNextWindowSize(imgui.NewVec2(imutils.S(240), viewSize.Y))
	}
	if pv.numUdims <= 1 {
		imgui.SetItemTooltip("Mesh has no UDims")
	}
	imgui.EndDisabled()
	if imgui.InternalBeginPopupEx(imgui.IDStr("UDims"), imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoSavedSettings) {
		if imgui.Button("Reset") {
			pv.udimsSelected = pv.udimsShownDefault
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
				icon = fnt.I.Visibility
			} else {
				icon = fnt.I.VisibilityOff
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
			imutils.Textf(fmt.Sprintf("%s %02d: %s", icon, i, cmp.Or(pv.udimNames[i], "unknown")))
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

	if pv.animTime != -1 {
		pv.animTime += 5 * imgui.CurrentIO().DeltaTime()
	}
}
