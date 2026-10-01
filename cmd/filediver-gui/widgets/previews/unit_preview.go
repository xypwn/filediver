package previews

import (
	"bytes"
	"cmp"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"maps"
	"math"
	"math/rand/v2"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/x448/float16"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
	geometrygroup "github.com/xypwn/filediver/stingray/unit/geometry_group"
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

// Basic uniforms all shaders use
var baseUniforms = []string{
	"mvp",
	"model",
	"normalMat",
	"viewPosition",
	"hasVisibilityMasks",
	"udimShown",
}

// Textures used by the lut fragment shader
var lutTextureNames = []string{
	"decal_sheet",
	"composite_array",
	"customization_camo_tiler_array",
	"customization_material_detail_tiler_array",
	"pattern_lut",
	"base_data",
	"material_lut",
	"pattern_masks_array",
	"id_masks_array",
	"ibl_brdf_lut",
}

var seed = rand.Uint32()

func setupTexture(textureID, target uint32) {
	gl.BindTexture(target, textureID)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.BindTexture(target, 0)
}

func setupTextureArray(textureID uint32) {
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, textureID)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, 0)
}

type unitPreviewUniformBlock struct {
	name           string
	ubo            uint32
	binding        uint32
	uniformOffsets map[string]int32
	uniformTypes   map[string]glutils.GLType
	defaultValues  map[string][]uint8
	currentValues  map[string][]uint8
}

func (block *unitPreviewUniformBlock) generate(program uint32, name string) {
	block.uniformOffsets = make(map[string]int32)
	block.uniformTypes = make(map[string]glutils.GLType)
	block.defaultValues = make(map[string][]uint8)
	block.currentValues = make(map[string][]uint8)

	cStr, free := gl.Strs(name + "\x00")
	blockIdx := gl.GetProgramResourceIndex(program, gl.UNIFORM_BLOCK, *cStr)
	free()

	if blockIdx == gl.INVALID_INDEX {
		return
	}

	var size int32
	gl.GetActiveUniformBlockiv(program, blockIdx, gl.UNIFORM_BLOCK_DATA_SIZE, &size)

	var numUniforms int32 = 0
	blockProperties := []uint32{gl.NUM_ACTIVE_VARIABLES}
	gl.GetProgramResourceiv(program, gl.UNIFORM_BLOCK, blockIdx, 1, &blockProperties[0], 1, nil, &numUniforms)

	fmt.Printf("Uniform block has %v uniforms\n", numUniforms)
	if numUniforms == 0 {
		return
	}

	activeUniforms := []uint32{gl.ACTIVE_VARIABLES}
	blockUniforms := make([]int32, numUniforms)
	gl.GetProgramResourceiv(program, gl.UNIFORM_BLOCK, blockIdx, 1, &activeUniforms[0], numUniforms, nil, &blockUniforms[0])

	for uniformIdx := range numUniforms {
		uniformInfo := make([]int32, 4)
		uniformProperties := []uint32{gl.NAME_LENGTH, gl.TYPE, gl.ARRAY_SIZE, gl.OFFSET}
		gl.GetProgramResourceiv(program, gl.UNIFORM, uint32(blockUniforms[uniformIdx]), 4, &uniformProperties[0], 4, nil, &uniformInfo[0])

		buf := make([]uint8, uniformInfo[0])
		gl.GetProgramResourceName(program, gl.UNIFORM, uint32(blockUniforms[uniformIdx]), int32(len(buf)), nil, &buf[0])

		uniformName := string(buf[:len(buf)-1])
		uniformType := glutils.GLType(uniformInfo[1])
		fmt.Printf("block uniform %v: offset %v type %v array size %v\n", uniformName, uniformInfo[3], uniformType.String(), uniformInfo[2])
		block.uniformOffsets[uniformName] = uniformInfo[3]
		block.uniformTypes[uniformName] = uniformType
		block.defaultValues[uniformName] = make([]uint8, max(1, uniformInfo[2])*int32(uniformType.Size()))
		block.currentValues[uniformName] = make([]uint8, max(1, uniformInfo[2])*int32(uniformType.Size()))
	}

	gl.GenBuffers(1, &block.ubo)
	gl.BindBuffer(gl.UNIFORM_BUFFER, block.ubo)
	gl.BufferData(gl.UNIFORM_BUFFER, int(size), gl.Ptr(make([]byte, size)), gl.STATIC_DRAW)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	fmt.Printf("Uniform block id: %v\n", block.ubo)
}

// Data must be a slice or ptr
func (block *unitPreviewUniformBlock) set(name string, data any) {
	offset, contains := block.uniformOffsets[name]
	if !contains || binary.Size(data) > len(block.currentValues[name]) {
		return
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, block.ubo)
	gl.BufferSubData(gl.UNIFORM_BUFFER, int(offset), binary.Size(data), gl.Ptr(data))
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	if _, err := binary.Encode(block.currentValues[name], binary.LittleEndian, data); err != nil {
		fmt.Printf("[error] current value for %v could not be encoded: %v\n", name, err)
	}
}

func (block *unitPreviewUniformBlock) reset(name string) {
	data, contains := block.defaultValues[name]
	if !contains {
		return
	}
	block.set(name, data)
}

// Data must be a slice
func (block *unitPreviewUniformBlock) setDefault(name string, data any) {
	_, contains := block.defaultValues[name]
	if !contains {
		return
	}
	if _, err := binary.Encode(block.defaultValues[name], binary.LittleEndian, data); err != nil {
		fmt.Printf("[error] default value for %v could not be encoded: %v\n", name, err)
	}
}

func (block *unitPreviewUniformBlock) get(name string, outData any) {
	data, contains := block.currentValues[name]
	if !contains {
		return
	}
	binary.Decode(data, binary.LittleEndian, outData)
}

type unitPreviewMaterialTexture struct {
	id     uint32
	name   stingray.Hash
	target uint32
}

type unitPreviewMaterial struct {
	name          string
	id            stingray.Hash
	program       uint32
	uniforms      unitPreviewUniforms
	uniformBlocks []unitPreviewUniformBlock
	textures      []unitPreviewMaterialTexture
}

func (mat *unitPreviewMaterial) generate(shaderPaths []string, textures int, uniforms []string) error {
	if mat.program != 0 || len(mat.uniformBlocks) > 0 || len(mat.uniforms) > 0 {
		// already generated this material
		return nil
	}
	var err error
	mat.program, err = glutils.CreateProgramFromSources(
		unitPreviewShaderCode,
		shaderPaths...,
	)
	if err != nil {
		return err
	}
	mat.textures = make([]unitPreviewMaterialTexture, textures)

	mat.uniforms.generate(mat.program, uniforms...)

	var numBlocks int32
	gl.GetProgramInterfaceiv(mat.program, gl.UNIFORM_BLOCK, gl.ACTIVE_RESOURCES, &numBlocks)
	for blockIdx := range numBlocks {
		var length int32
		buf := make([]uint8, 64)
		gl.GetProgramResourceName(mat.program, gl.UNIFORM_BLOCK, uint32(blockIdx), int32(len(buf)), &length, &buf[0])

		query := []uint32{gl.BUFFER_BINDING}
		var bindpoint int32
		gl.GetProgramResourceiv(mat.program, gl.UNIFORM_BLOCK, uint32(blockIdx), 1, &query[0], 1, nil, &bindpoint)

		blockName := string(buf[:length])
		block := unitPreviewUniformBlock{binding: uint32(bindpoint)}
		block.generate(mat.program, blockName)
		mat.uniformBlocks = append(mat.uniformBlocks, block)
	}

	return nil
}

func (mat *unitPreviewMaterial) delete(textureCache *glutils.TextureCache) {
	mat.releaseTextures(textureCache)
	gl.DeleteProgram(mat.program)
}

func (mat *unitPreviewMaterial) releaseTextures(textureCache *glutils.TextureCache) {
	for _, texture := range mat.textures {
		if texture.name.Value == 0x0 {
			gl.DeleteTextures(1, &texture.id)
			continue
		}
		textureCache.Release(texture.name, texture.target)
	}
	mat.textures = make([]unitPreviewMaterialTexture, 0)
}

type unitPreviewObject struct {
	vao       uint32   // vertex array object
	ibos      []uint32 // index buffer objects
	vbo       uint32   // vertex buffer object
	materials []unitPreviewMaterial
	matrix    mgl32.Mat4

	numVertices        int32
	numIndices         []int32
	hasVisibilityMasks int32
}

// NOTE(xypwn): We do at most ~10 lookups once per frame,
// so it should be fine to store this in a string map.
type unitPreviewUniforms map[string]int32

func (obj *unitPreviewObject) genObjects(textures bool, numIbos int32) {
	gl.GenVertexArrays(1, &obj.vao)
	gl.GenBuffers(1, &obj.vbo)
	if numIbos > 0 {
		obj.ibos = make([]uint32, numIbos)
		obj.numIndices = make([]int32, numIbos)
		gl.GenBuffers(numIbos, &obj.ibos[0])
	}

	gl.BindVertexArray(obj.vao)
	defer gl.BindVertexArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, obj.vbo)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

func (uniforms *unitPreviewUniforms) generate(program uint32, names ...string) {
	if *uniforms == nil {
		*uniforms = unitPreviewUniforms{}
	}
	for _, name := range names {
		cStr, free := gl.Strs(name + "\x00")
		loc := gl.GetUniformLocation(program, *cStr)
		free()

		(*uniforms)[name] = loc
	}
}

func (obj unitPreviewObject) deleteObjects(textureCache *glutils.TextureCache) {
	gl.DeleteVertexArrays(1, &obj.vao)
	gl.DeleteBuffers(1, &obj.vbo)
	if len(obj.ibos) > 0 {
		gl.DeleteBuffers(int32(len(obj.ibos)), &obj.ibos[0])
	}
	for _, material := range obj.materials {
		material.delete(textureCache)
	}
}

type UnitPreviewState struct {
	fb                 *widgets.GLViewState
	textureCache       *glutils.TextureCache
	textureSweepTicker *time.Ticker
	doSweep            bool
	stopTextureSweep   func()

	detailerLoadState   DetailerLoadState
	detailerTextureData TextureData

	loadedUnits             map[stingray.Hash]any
	unitsShown              map[stingray.Hash]bool
	loadUnitGetResourceFunc GetResourceFunc

	unitMatrices         map[stingray.Hash]mgl32.Mat4
	objects              map[stingray.Hash]map[string]unitPreviewObject
	objectsShown         map[stingray.Hash]map[string]bool
	objectsShownDefault  map[stingray.Hash]map[string]bool
	objectsSelected      map[stingray.Hash]map[string]bool
	objectsSettingsShown bool

	materialSettingsShown bool
	materialSettingsDrawn bool

	wireframeMaterial   unitPreviewMaterial
	normalVisMaterial   unitPreviewMaterial
	boundingBoxMaterial unitPreviewMaterial

	skeletons     map[stingray.Hash]unitPreviewObject
	boundingBoxes map[stingray.Hash]map[string]unitPreviewObject

	vfov         float32
	modelPos     mgl32.Vec4
	model        mgl32.Mat4
	viewDistance float32
	viewRotation mgl32.Vec2 // {yaw, pitch}

	// Previous view distance and rotation (for view animation)
	animOrigViewDistance float32
	animOrigViewRotation mgl32.Vec2
	animOrigModelPos     mgl32.Vec4
	animTime             float32 // range [0;1], -1 when not animating

	// Axis-aligned bounding box. Don't forget
	// to multiply aabb's vertices with aabbMat first!
	aabb    map[stingray.Hash]map[string][2]mgl32.Vec3
	aabbMat map[stingray.Hash]map[string]mgl32.Mat4

	// For fitting mesh to screen and debug info
	meshPositions map[stingray.Hash]map[string][][3]float32
	meshNormals   map[stingray.Hash]map[string][][3]float32

	maxViewDistance float32

	numUdims           uint32
	udimsShownDefault  [64]bool
	udimsSelected      [64]bool  // udims persistently selected
	udimsShown         [64]int32 // udims visually selected 1 (shown) or 0 (hidden)
	udimNames          [64]string
	udimsSettingsShown bool
	udimsSettingsDrawn bool

	// For dragging selection
	activeUDimListItem  int32
	hoveredUDimListItem int32

	activeMeshListItem  int32
	hoveredMeshListItem int32

	showWireframe             bool
	wireframeColor            [4]float32
	showAABB                  bool
	aabbColor                 [4]float32
	showSkeleton              bool
	skeletonColor             [4]float32
	visualizeNormals          bool
	visualizeTangentBitangent int32 // 1 or 0
	autoZoomEnabled           bool
	doAutoZoomNextFrame       bool

	armorSets                map[stingray.Hash]datalib.ArmorSet
	getSelectedArchives      func() []stingray.Hash
	previousSelectedArchives []stingray.Hash
	archivesModified         bool
	lookupThinHash           func(stingray.ThinHash) string
	lookupHash               func(stingray.Hash) string
}

func NewUnitPreview(getResource GetResourceFunc, ArmorParams ExtractorArmorParameters, lookupHash func(stingray.Hash) string) (*UnitPreviewState, error) {
	var err error

	pv := &UnitPreviewState{}

	pv.fb, err = widgets.NewGLView()
	if err != nil {
		return nil, err
	}

	pv.lookupHash = lookupHash

	pv.armorSets = ArmorParams.ArmorSets
	pv.getSelectedArchives = ArmorParams.SelectedArchives

	// Keep textures for a minute of disuse
	duration, _ := time.ParseDuration("1m")
	pv.textureCache = glutils.NewTextureCache(duration)

	sweepInterval, _ := time.ParseDuration("30s")
	pv.textureSweepTicker = time.NewTicker(sweepInterval)

	done := make(chan bool)
	pv.stopTextureSweep = func() {
		done <- true
	}
	// schedule a texture sweep every time the ticker ticks
	go func() {
		for {
			select {
			case <-done:
				return
			case <-pv.textureSweepTicker.C:
				pv.doSweep = true
			}
		}
	}()

	pv.detailerLoadState = DetailerNotLoaded
	pv.loadMaterialDetailer(getResource)

	//pv.object.genObjects(true, 0)

	pv.wireframeMaterial.name = "wireframe"
	err = pv.wireframeMaterial.generate(
		[]string{
			"shaders/object_wireframe.vert",
			"shaders/object_wireframe.geom",
			"shaders/object_wireframe.frag",
		},
		0,
		[]string{"mvp", "color", "udimShown", "hasVisibilityMasks"},
	)
	if err != nil {
		return nil, err
	}

	pv.normalVisMaterial.name = "normal visualization"
	err = pv.normalVisMaterial.generate(
		[]string{
			"shaders/object_normal_vis.vert",
			"shaders/object_normal_vis.geom",
			"shaders/object_normal_vis.frag",
		},
		0,
		[]string{"mvp", "len", "showTangentBitangent", "udimShown", "hasVisibilityMasks"},
	)
	if err != nil {
		return nil, err
	}

	pv.boundingBoxMaterial.name = "bounding box"
	err = pv.boundingBoxMaterial.generate(
		[]string{
			"shaders/debug_object.vert",
			"shaders/debug_object.frag",
		},
		0,
		[]string{"mvp", "color"},
	)
	if err != nil {
		return nil, err
	}

	pv.vfov = mgl32.DegToRad(60)
	pv.viewDistance = 25
	pv.modelPos = mgl32.Vec4{0.0, 0.0, 0.0, 1.0}
	pv.viewRotation = mgl32.Vec2{math.Pi, 0.0}

	pv.wireframeColor = [4]float32{1.0, 1.0, 1.0, 0.5}
	pv.skeletonColor = [4]float32{1.0, 0.5, 0.0, 0.5}
	pv.aabbColor = [4]float32{0.3, 0.3, 0.8, 0.2}

	pv.activeUDimListItem = -1
	pv.hoveredUDimListItem = -1

	pv.activeMeshListItem = -1
	pv.hoveredMeshListItem = -1

	return pv, nil
}

func (pv *UnitPreviewState) Delete() {
	pv.fb.Delete()
	for hash := range pv.objects {
		for name := range pv.objects[hash] {
			pv.objects[hash][name].deleteObjects(pv.textureCache)
			pv.boundingBoxes[hash][name].deleteObjects(pv.textureCache)
		}
	}
	pv.wireframeMaterial.delete(pv.textureCache)
	pv.releaseMaterialDetailer()
	pv.textureCache.DeleteAll()
	pv.stopTextureSweep()
}

var objectRegex = regexp.MustCompile("^(g_)?(\\w*?)_?(cull|shadow|rubble|rubble_shadow|shadowmesh|c)?_?(LOD\\d)?$")

func (pv *UnitPreviewState) loadMeshes(lookupThinHash func(stingray.ThinHash) string, meshInfos []unit.MeshInfo, meshLayouts []unit.MeshLayout, gpuData []byte) (map[string]unit.Mesh, map[string]bool, error) {
	minLods := make(map[string]int)
	objects := make(map[string]map[string]int)
	for idx, info := range meshInfos {
		meshName := lookupThinHash(info.Header.MeshName)
		object := objectRegex.FindStringSubmatch(meshName)
		fmt.Printf("Found object:\n    name: %v\n    shadow/rubble: %v\n    LOD: %v\n", object[2], object[3], object[4])
		// rubble/shadow/rubble_shadow
		objectName := object[2]
		if object[3] != "" {
			objectName += "_" + object[3]
		}
		foundObject, ok := objects[objectName]
		if !ok {
			foundObject = make(map[string]int)
		}
		foundObject[object[4]] = idx
		objects[objectName] = foundObject
	}

	for object, lods := range objects {
		sortedLods := slices.Sorted(maps.Keys(lods))
		chosenLod := sortedLods[0]
		if chosenLod == "" {
			chosenLod = "<base-mesh>"
		}
		fmt.Printf("choosing %v as min lod for object %v\n", chosenLod, object)
		minLods[object] = lods[sortedLods[0]]
	}

	meshesToLoad := make([]uint32, 0)
	shownDefault := make(map[string]bool)
	// if we don't have any idea what the lods are, just fall back to original behavior
	if len(minLods) == 0 {
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
			return nil, nil, fmt.Errorf("unable to find mesh to load")
		}
		meshesToLoad = append(meshesToLoad, uint32(highestDetailIdx))
	} else {
		for object, idx := range minLods {
			if !strings.HasPrefix(object, "0x") &&
				(strings.Contains(object, "shadow") ||
					strings.HasPrefix(object, "c_") ||
					strings.HasSuffix(object, "_c") ||
					strings.Contains(object, "cull") ||
					strings.Contains(object, "coll") ||
					strings.HasSuffix(object, "rubble") ||
					object == "ai_blocker") {
				shownDefault[object] = false
			}
			meshesToLoad = append(meshesToLoad, uint32(idx))
		}
	}

	meshes := make(map[string]unit.Mesh)
	{
		var err error
		meshMap, err := unit.LoadMeshes(bytes.NewReader(gpuData), meshInfos, meshLayouts, meshesToLoad)
		if err != nil {
			return nil, nil, err
		}
		if len(minLods) == 0 {
			meshes["highest-detail"] = meshMap[meshesToLoad[0]]
			shownDefault["highest-detail"] = true
		} else {
			for object, idx := range minLods {
				if _, contains := shownDefault[object]; !contains {
					shownDefault[object] = true
				}
				meshes[object] = meshMap[uint32(idx)]
			}
		}
	}
	return meshes, shownDefault, nil
}

func loadDDS(getResource GetResourceFunc, fileName stingray.Hash) (*dds.DDS, error) {
	file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
	var texMain, texStream, texGPU []byte
	var err error
	var exists bool
	if texMain, exists, err = getResource(file, stingray.DataMain); err != nil {
		return nil, fmt.Errorf("load texture %v.texture: %v", fileName.String(), err)
	} else if !exists {
		return nil, fmt.Errorf("%v.texture does not exist", fileName.String())
	}
	texStream, _, _ = getResource(file, stingray.DataStream)
	texGPU, _, _ = getResource(file, stingray.DataGPU)
	dataR := io.MultiReader(
		bytes.NewReader(texMain),
		bytes.NewReader(texStream),
		bytes.NewReader(texGPU),
	)
	if _, err := texture.DecodeInfo(dataR); err != nil {
		return nil, fmt.Errorf("loading stingray DDS info: %w", err)
	}
	dds, err := dds.Decode(dataR, false)
	if err != nil {
		return nil, fmt.Errorf("loading DDS image: %w", err)
	}
	return dds, nil
}

func uploadStingrayTexture(textureID uint32, data TextureData) error {
	gl.BindTexture(data.Target, textureID)
	if data.Target == gl.TEXTURE_2D {
		gl.TexImage2D(data.Target, 0, data.InternalFormat, int32(data.Bounds.Dx()), int32(data.Bounds.Dy()), 0, data.Format, data.Type, gl.Ptr(data.Data))
	} else if data.Target == gl.TEXTURE_2D_ARRAY {
		gl.TexImage3D(data.Target, 0, data.InternalFormat, int32(data.Bounds.Dx()), int32(data.Bounds.Dy()), data.Depth, 0, data.Format, data.Type, gl.Ptr(data.Data))
	}
	gl.BindTexture(data.Target, 0)
	return nil
}

func (pv *UnitPreviewState) AcquireNamedTextureOrDefault(name stingray.Hash, defaultColor []byte, getResource GetResourceFunc) (returnPtr *unitPreviewMaterialTexture, err error) {
	toReturn := unitPreviewMaterialTexture{
		name:   stingray.Hash{Value: 0},
		target: gl.TEXTURE_2D,
	}
	defer func() {
		if returnPtr != nil {
			return
		}
		gl.GenTextures(1, &toReturn.id)
		setupTexture(toReturn.id, toReturn.target)
		gl.BindTexture(toReturn.target, toReturn.id)
		gl.TexImage2D(toReturn.target, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(defaultColor))
		gl.BindTexture(toReturn.target, 0)
		returnPtr = &toReturn
	}()
	if name.Value == 0 {
		return
	}
	textureId, created := pv.textureCache.Acquire(name, toReturn.target)
	toReturn.id = textureId
	toReturn.name = name
	if created {
		setupTexture(toReturn.id, toReturn.target)
		var dds *dds.DDS
		dds, err = loadDDS(getResource, name)
		if err != nil {
			pv.textureCache.Delete(name, toReturn.target)
			return
		}
		texData := TextureData{
			Target: gl.TEXTURE_2D,
			Bounds: dds.Bounds(),
			Type:   gl.UNSIGNED_BYTE,
		}
		switch img := dds.Image.(type) {
		case *image.Gray:
			texData.InternalFormat = gl.RED
			texData.Format = gl.RED
			texData.Data = img.Pix
			gl.BindTexture(texData.Target, toReturn.id)
			swizzles := []uint32{gl.RED, gl.RED, gl.RED, gl.RED}
			gl.TextureParameterIuiv(toReturn.id, gl.TEXTURE_SWIZZLE_RGBA, &swizzles[0])
			gl.BindTexture(texData.Target, 0)
		case *image.NRGBA:
			texData.InternalFormat = gl.RGBA
			texData.Format = gl.RGBA
			texData.Data = img.Pix
		default:
			if dds.Info.DXT10Header != nil {
				err = fmt.Errorf("unexpected texture color model: %v", dds.Info.DXT10Header.DXGIFormat.String())
			} else {
				err = fmt.Errorf("unexpected texture color model")
			}
			pv.textureCache.Delete(name, toReturn.target)
			return
		}
		if err = uploadStingrayTexture(toReturn.id, texData); err != nil {
			// Failed to upload data, so delete the entry in the cache
			pv.textureCache.Delete(name, toReturn.target)
			return
		}
		returnPtr = &toReturn
	}
	return
}

func (pv *UnitPreviewState) useBasicMaterial(getResource GetResourceFunc, object *unitPreviewObject, group int, mat *material.Material) error {
	err := object.materials[group].generate(
		[]string{"shaders/object.vert", "shaders/object.geom", "shaders/object.frag"},
		0,
		append(baseUniforms, "texAlbedo", "texNormal", "shouldReconstructNormalZ"),
	)
	if err != nil {
		return err
	}

	gl.UseProgram(object.materials[group].program)
	gl.Uniform1i(object.materials[group].uniforms["texAlbedo"], 0)
	gl.Uniform1i(object.materials[group].uniforms["texNormal"], 1)
	gl.UseProgram(0)

	// Upload object texture
	albedoTexFileName, albedoRemoveAlpha, normalTexFileName, reconstructNormalZ, err := func() (albedoFileName stingray.Hash, albedoRemoveAlpha bool, normalFileName stingray.Hash, reconstructNormalZ bool, err error) {
		// TODO: Use all textures somehow. Currently, simply the first one
		// found is used.
		if mat == nil {
			return
		}
		for texUsage, texFileName := range mat.Textures {
			removeAlpha := true
			switch texUsage {
			case stingray.Sum("color_roughness").Thin(), stingray.Sum("color_specular_b").Thin(), stingray.Sum("albedo_iridescence").Thin():
				removeAlpha = false
				fallthrough
			case stingray.Sum("covering_albedo").Thin(), stingray.Sum("input_image").Thin(), stingray.Sum("albedo").Thin():
				albedoFileName = texFileName
				albedoRemoveAlpha = removeAlpha
			case stingray.Sum("normal_specular_ao").Thin():
				normalFileName = texFileName
				reconstructNormalZ = false
			case stingray.Sum("normal").Thin(), stingray.Sum("normals").Thin(), stingray.Sum("normal_map").Thin(), stingray.Sum("covering_normal").Thin(), stingray.Sum("nac").Thin(), stingray.Sum("base_data").Thin(), stingray.Sum("nar").Thin(), stingray.Sum("normal_ao_roughness").Thin(), stingray.Sum("normal_xy_ao_rough_map").Thin(), stingray.Sum("normal_xy_roughness_opacity").Thin():
				normalFileName = texFileName
				reconstructNormalZ = true
			}
		}
		return
	}()
	if err != nil {
		return err
	}

	albedoTexture, err := pv.AcquireNamedTextureOrDefault(albedoTexFileName, []byte{255, 255, 255, 255}, getResource)
	if err != nil {
		// gui logger warning here
		fmt.Printf("acquiring %v.texture: %v\nfalling back to default value", albedoTexFileName.String(), err)
	}
	if albedoRemoveAlpha {
		gl.BindTexture(gl.TEXTURE_2D, albedoTexture.id)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, gl.ONE)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	object.materials[group].textures = append(object.materials[group].textures, *albedoTexture)

	normalTexture, err := pv.AcquireNamedTextureOrDefault(normalTexFileName, []byte{128, 128, 255, 128}, getResource)
	if err != nil {
		fmt.Printf("acquiring %v.texture: %v\nfalling back to default value", normalTexFileName.String(), err)
	}
	object.materials[group].textures = append(object.materials[group].textures, *normalTexture)

	gl.UseProgram(object.materials[group].program)
	if reconstructNormalZ {
		gl.Uniform1i(object.materials[group].uniforms["shouldReconstructNormalZ"], 1)
	} else {
		gl.Uniform1i(object.materials[group].uniforms["shouldReconstructNormalZ"], 0)
	}
	gl.Uniform1i(object.materials[group].uniforms["hasVisibilityMasks"], 0)
	gl.UseProgram(0)
	return nil
}

func getTarget(slot string) uint32 {
	switch slot {
	case "customization_camo_tiler_array", "pattern_masks_array", "id_masks_array", "customization_material_detail_tiler_array":
		return gl.TEXTURE_2D_ARRAY
	}
	return gl.TEXTURE_2D
}

type TextureData struct {
	Slot           string
	Name           stingray.Hash
	Target         uint32
	Bounds         image.Rectangle
	Depth          int32
	InternalFormat int32
	Format         uint32
	Type           uint32
	Data           []uint8
}

type DetailerLoadState uint8

const (
	DetailerNotLoaded DetailerLoadState = iota
	DetailerLoading
	DetailerLoaded
	DetailerUploaded
)

func (pv *UnitPreviewState) loadMaterialDetailer(getResource GetResourceFunc) {
	slot := "customization_material_detail_tiler_array"
	materialDetailerHash := stingray.Sum("content/art_shared/textures/customization/material_library/detail_tilers/customization_detail_tiler_array")
	var target uint32 = gl.TEXTURE_2D_ARRAY

	if pv.textureCache.Contains(materialDetailerHash, target) {
		// Already loaded, so we don't need to reload the data
		return
	}
	textureData := []TextureData{{
		Slot:   slot,
		Name:   materialDetailerHash,
		Target: target,
	}}

	pv.detailerLoadState = DetailerLoading

	textureLoader := getTextureLoaderFunc(getResource, materialDetailerHash, &textureData, 0)
	go func() {
		textureLoader()
		pv.detailerTextureData = textureData[0]
		pv.detailerLoadState = DetailerLoaded
	}()
}

func (pv *UnitPreviewState) uploadMaterialDetailer(data TextureData) {
	textureId, created := pv.textureCache.Acquire(data.Name, data.Target)
	texture := unitPreviewMaterialTexture{
		id:     textureId,
		name:   data.Name,
		target: data.Target,
	}

	if !created {
		return
	}
	setupTexture(texture.id, data.Target)
	err := uploadStingrayTexture(texture.id, data)
	if err != nil {
		pv.releaseMaterialDetailer()
		pv.detailerLoadState = DetailerNotLoaded
	} else {
		pv.detailerLoadState = DetailerUploaded
	}
	// We can discard the texture data after uploading, so the garbage collector will reclaim the ~110MB of memory
	pv.detailerTextureData = TextureData{}
}

func (pv *UnitPreviewState) releaseMaterialDetailer() {
	materialDetailerHash := stingray.Sum("content/art_shared/textures/customization/material_library/detail_tilers/customization_detail_tiler_array")
	pv.textureCache.Release(materialDetailerHash, gl.TEXTURE_2D_ARRAY)
}

func getTextureLoaderFunc(getResource GetResourceFunc, nameHash stingray.Hash, textureData *[]TextureData, index int) func() {
	return func() {
		(*textureData)[index].Bounds = image.Rect(0, 0, 1, 1)
		(*textureData)[index].Data = make([]uint8, 4)
		(*textureData)[index].Depth = 1

		ddsImage, err := loadDDS(getResource, nameHash)
		if err != nil {
			return
		}
		(*textureData)[index].Bounds = ddsImage.Bounds()
		(*textureData)[index].Depth = int32(len(ddsImage.Images))

		for idx := range ddsImage.Images {
			if ddsImage.Info.DXT10Header != nil && ddsImage.Info.DXT10Header.DXGIFormat == dds.DXGIFormatR16G16B16A16Float {
				(*textureData)[index].Format = gl.RGBA
				(*textureData)[index].InternalFormat = gl.RGBA16F
				(*textureData)[index].Type = gl.HALF_FLOAT
				(*textureData)[index].Data = ddsImage.Images[0].MipMaps[0].Raw
			} else if ddsImage.Info.DXT10Header != nil && ddsImage.Info.DXT10Header.DXGIFormat == dds.DXGIFormatR16G16Float {
				(*textureData)[index].Format = gl.RG
				(*textureData)[index].InternalFormat = gl.RG16F
				(*textureData)[index].Type = gl.HALF_FLOAT
				(*textureData)[index].Data = ddsImage.Images[0].MipMaps[0].Raw
			} else {
				switch img := ddsImage.Images[idx].Image.(type) {
				case *image.NRGBA:
					(*textureData)[index].Format = gl.RGBA
					(*textureData)[index].InternalFormat = gl.RGBA
					(*textureData)[index].Type = gl.UNSIGNED_BYTE
					if idx == 0 {
						(*textureData)[index].Data = make([]uint8, 0, ddsImage.Bounds().Dx()*ddsImage.Bounds().Dy()*int((*textureData)[index].Depth)*4)
					}
					(*textureData)[index].Data = append((*textureData)[index].Data, img.Pix...)
				case *image.Gray:
					(*textureData)[index].Format = gl.RED
					(*textureData)[index].InternalFormat = gl.RED
					(*textureData)[index].Type = gl.UNSIGNED_BYTE
					if idx == 0 {
						(*textureData)[index].Data = make([]uint8, 0, ddsImage.Bounds().Dx()*ddsImage.Bounds().Dy()*int((*textureData)[index].Depth))
					}
					(*textureData)[index].Data = append((*textureData)[index].Data, img.Pix...)
				default:
					fmt.Printf("[error] Failed to convert image %v of %v\n", idx, (*textureData)[index].Slot)
					return
				}
			}
		}
	}
}

func isLUTMaterial(mat *material.Material) bool {
	if mat == nil {
		return false
	}
	_, containsIdMasks := mat.Textures[stingray.Sum("id_masks_array").Thin()]
	_, containsMaterialLut := mat.Textures[stingray.Sum("material_lut").Thin()]
	return containsIdMasks && containsMaterialLut
}

func overrideLUTMaterial(mat *material.Material, armorInfo *datalib.UnitData) {
	if mat == nil || mat.Textures == nil || armorInfo == nil {
		return
	}
	if armorInfo.MaterialLut.Value != 0 {
		mat.Textures[stingray.Sum("material_lut").Thin()] = armorInfo.MaterialLut
	}
	if armorInfo.PatternLut.Value != 0 {
		mat.Textures[stingray.Sum("pattern_lut").Thin()] = armorInfo.PatternLut
	}
	if armorInfo.CapeLut.Value != 0 {
		mat.Textures[stingray.Sum("cape_lut").Thin()] = armorInfo.CapeLut
	}
	if armorInfo.CapeGradient.Value != 0 {
		mat.Textures[stingray.Sum("cape_gradient").Thin()] = armorInfo.CapeGradient
	}
	if armorInfo.CapeNac.Value != 0 {
		mat.Textures[stingray.Sum("cape_nac").Thin()] = armorInfo.CapeNac
	}
	if armorInfo.DecalScalarFields.Value != 0 {
		if _, contains := mat.Textures[stingray.Sum("id_masks_array").Thin()]; contains {
			mat.Textures[stingray.Sum("id_masks_array").Thin()] = armorInfo.DecalScalarFields
		}
		if _, contains := mat.Textures[stingray.Sum("decal_scalar_fields").Thin()]; contains {
			mat.Textures[stingray.Sum("decal_scalar_fields").Thin()] = armorInfo.DecalScalarFields
		}
	}
	if armorInfo.BaseData.Value != 0 {
		mat.Textures[stingray.Sum("base_data").Thin()] = armorInfo.BaseData
	}
	if armorInfo.DecalSheet.Value != 0 {
		mat.Textures[stingray.Sum("decal_sheet").Thin()] = armorInfo.DecalSheet
	}
}

func (pv *UnitPreviewState) useLUTMaterial(getResource GetResourceFunc, object *unitPreviewObject, group int, mat *material.Material, lookupThinHash func(stingray.ThinHash) string) error {
	err := object.materials[group].generate(
		[]string{"shaders/object.vert", "shaders/object.geom", "shaders/lut.frag"},
		0,
		append(baseUniforms, lutTextureNames...),
	)
	if err != nil {
		return err
	}

	if pv.detailerLoadState == DetailerLoaded {
		pv.uploadMaterialDetailer(pv.detailerTextureData)
	}

	slot := "customization_material_detail_tiler_array"
	materialDetailerHash := stingray.Sum("content/art_shared/textures/customization/material_library/detail_tilers/customization_detail_tiler_array")
	mat.Textures[stingray.Sum(slot).Thin()] = materialDetailerHash

	slot = "ibl_brdf_lut"
	iblBRDFHash := stingray.Sum("core/stingray_renderer/lookup_tables/ibl_specular_brdf_lut")
	mat.Textures[stingray.Sum(slot).Thin()] = iblBRDFHash

	if _, contains := mat.Textures[stingray.Sum("pattern_masks_array").Thin()]; !contains {
		mat.Textures[stingray.Sum("pattern_masks_array").Thin()] = stingray.Sum("content/art_shared/textures/black_all_channels_dummy")
	}
	if _, contains := mat.Textures[stingray.Sum("composite_array").Thin()]; !contains {
		mat.Textures[stingray.Sum("composite_array").Thin()] = stingray.Sum("content/art_shared/textures/black_all_channels_dummy")
	}
	if _, contains := mat.Textures[stingray.Sum("pattern_lut").Thin()]; !contains {
		mat.Textures[stingray.Sum("pattern_lut").Thin()] = stingray.Hash{Value: 0xcf0cc31b981786c9}
	}

	var textureWaitGroup sync.WaitGroup
	textureData := make([]TextureData, 0)
	slotHashes := slices.SortedFunc(maps.Keys(mat.Textures), stingray.ThinHash.Cmp)
	for _, slotHash := range slotHashes {
		slot := lookupThinHash(slotHash)
		if !slices.Contains(lutTextureNames, slot) {
			continue
		}
		nameHash := mat.Textures[slotHash]
		target := getTarget(slot)

		index := len(textureData)
		textureData = append(textureData, TextureData{
			Slot:   slot,
			Name:   nameHash,
			Target: target,
		})

		if pv.textureCache.Contains(nameHash, target) {
			// Already loaded, so we don't need to reload the data
			continue
		}

		textureWaitGroup.Go(getTextureLoaderFunc(getResource, nameHash, &textureData, index))
	}

	textureWaitGroup.Wait()

	gl.UseProgram(object.materials[group].program)
	for _, data := range textureData {
		textureId, created := pv.textureCache.Acquire(data.Name, data.Target)
		texture := unitPreviewMaterialTexture{
			id:     textureId,
			name:   data.Name,
			target: data.Target,
		}

		var err error
		if created {
			setupTexture(texture.id, data.Target)
			err = uploadStingrayTexture(texture.id, data)
		}
		if err != nil {
			pv.textureCache.Delete(data.Name, texture.target)
			return err
		}

		gl.Uniform1i(object.materials[group].uniforms[data.Slot], int32(len(object.materials[group].textures)))
		object.materials[group].textures = append(object.materials[group].textures, texture)
	}

	for setting, value := range mat.Settings {
		settingName := lookupThinHash(setting)
		for _, block := range object.materials[group].uniformBlocks {
			if _, contains := block.uniformOffsets[settingName]; !contains {
				continue
			}
			block.set(settingName, value)
			block.setDefault(settingName, value)
		}
	}
	for _, block := range object.materials[group].uniformBlocks {
		if _, contains := block.uniformOffsets["seed"]; contains {
			block.set("seed", &seed)
			block.setDefault("seed", &seed)
		}
		if _, contains := block.uniformOffsets["use_decals"]; contains {
			val, contains := mat.Textures[stingray.Sum("decal_sheet").Thin()]
			hasValue := val.Value != 0x0
			useDecals := uint32(0)
			if contains || hasValue {
				useDecals = 1
			}
			block.set("use_decals", &useDecals)
			block.setDefault("use_decals", &useDecals)
		}
	}

	gl.UseProgram(0)
	return nil
}

func loadMaterial(getResource GetResourceFunc, matFileName stingray.Hash) (*material.Material, error) {
	if matFileName.Value == 0x0 {
		return nil, fmt.Errorf("nil material")
	}
	matData, ok, err := getResource(stingray.FileID{
		Name: matFileName,
		Type: stingray.Sum("material"),
	}, stingray.DataMain)
	if err != nil {
		return nil, fmt.Errorf("load material %v.material: %w", matFileName, err)
	}
	if !ok {
		return nil, fmt.Errorf("load material %v.material does not exist", matFileName)
	}
	return material.LoadMain(bytes.NewReader(matData))
}

func (pv *UnitPreviewState) loadMaterials(getResource GetResourceFunc, object *unitPreviewObject, armorInfo *datalib.UnitData, lookupThinHash func(stingray.ThinHash) string) error {
	for group := range object.materials {
		object.materials[group].releaseTextures(pv.textureCache)

		mat, err := loadMaterial(getResource, object.materials[group].id)
		if err == nil && isLUTMaterial(mat) {
			overrideLUTMaterial(mat, armorInfo)
			if err := pv.useLUTMaterial(getResource, object, group, mat, lookupThinHash); err == nil {
				continue
			} else {
				fmt.Printf("got error when enabling lut material: %v\n", err)
			}
			// fall back to basic material if the lut material fails to load
		}
		if err := pv.useBasicMaterial(getResource, object, group, mat); err != nil {
			return err
		}
	}
	return nil
}

func (pv *UnitPreviewState) RemoveUnit(hash stingray.Hash) {
	if _, contains := pv.loadedUnits[hash]; !contains {
		return
	}
	for name := range pv.objects[hash] {
		pv.objects[hash][name].deleteObjects(pv.textureCache)
		pv.boundingBoxes[hash][name].deleteObjects(pv.textureCache)
	}
	delete(pv.loadedUnits, hash)
	delete(pv.unitMatrices, hash)
	delete(pv.objects, hash)
	delete(pv.boundingBoxes, hash)
	delete(pv.unitsShown, hash)
	delete(pv.objectsShown, hash)
	delete(pv.objectsSelected, hash)
	delete(pv.objectsShownDefault, hash)
	delete(pv.aabb, hash)
	delete(pv.aabbMat, hash)
	delete(pv.meshPositions, hash)
	delete(pv.meshNormals, hash)
}

func (pv *UnitPreviewState) Clear() {
	for hash := range pv.objects {
		pv.RemoveUnit(hash)
	}
}

func (pv *UnitPreviewState) LoadUnit(fileID stingray.Hash, mainData, gpuData []byte, getResource GetResourceFunc, thinhashes map[stingray.ThinHash]string) error {
	info, err := unit.LoadInfo(bytes.NewReader(mainData))
	if err != nil {
		return err
	}

	if pv.loadedUnits == nil {
		pv.loadedUnits = make(map[stingray.Hash]any)
	}
	pv.Clear()

	pv.lookupThinHash = func(hash stingray.ThinHash) string {
		if name, ok := thinhashes[hash]; ok {
			return name
		}
		return hash.String()
	}

	var armorInfo *datalib.UnitData
	selectedArchives := pv.getSelectedArchives()
	for idx := range selectedArchives {
		var set datalib.ArmorSet
		var contains bool
		if set, contains = pv.armorSets[selectedArchives[idx]]; !contains {
			continue
		}

		value, contains := set.UnitMetadata[fileID]
		if !contains {
			continue
		}

		armorInfo = &value
	}
	pv.previousSelectedArchives = selectedArchives

	if len(info.MeshInfos) == 0 && len(info.TerrainInfos) == 0 && info.GeometryGroup.Value == 0x0 {
		return fmt.Errorf("unit contains no meshes")
	}

	setModelPos := len(pv.objects) == 0

	if pv.objects == nil {
		pv.unitMatrices = make(map[stingray.Hash]mgl32.Mat4)
		pv.objects = make(map[stingray.Hash]map[string]unitPreviewObject)
		pv.boundingBoxes = make(map[stingray.Hash]map[string]unitPreviewObject)
		pv.unitsShown = make(map[stingray.Hash]bool)
		pv.objectsShown = make(map[stingray.Hash]map[string]bool)
		pv.objectsSelected = make(map[stingray.Hash]map[string]bool)
		pv.objectsShownDefault = make(map[stingray.Hash]map[string]bool)
		pv.aabb = make(map[stingray.Hash]map[string][2]mgl32.Vec3)
		pv.aabbMat = make(map[stingray.Hash]map[string]mgl32.Mat4, 0)
		pv.meshPositions = make(map[stingray.Hash]map[string][][3]float32)
		pv.meshNormals = make(map[stingray.Hash]map[string][][3]float32)
	}

	var meshes map[string]unit.Mesh
	if len(info.MeshInfos) > 0 {
		var defaultShown map[string]bool
		meshes, defaultShown, err = pv.loadMeshes(pv.lookupThinHash, info.MeshInfos, info.MeshLayouts, gpuData)
		if err != nil {
			return err
		}
		pv.objectsShownDefault[fileID] = defaultShown
	} else if len(info.TerrainInfos) > 0 {
		mesh, err := unit.LoadTerrain(info.TerrainInfos[0])
		if err != nil {
			return err
		}
		terrainConversionMatrix := mgl32.Rotate3DX(math.Pi).Mat4().Mul4(stingray.ToGLTFMatrix)
		for i := range mesh.Positions {
			mesh.Positions[i] = terrainConversionMatrix.Mul4x1(mgl32.Vec3(mesh.Positions[i]).Vec4(1)).Vec3()
			mesh.Normals[i] = terrainConversionMatrix.Mul4x1(mgl32.Vec3(mesh.Normals[i]).Vec4(1)).Vec3()
			mesh.Tangents[i] = terrainConversionMatrix.Mul4x1(mgl32.Vec4(mesh.Tangents[i]))
			mesh.Bitangents[i] = terrainConversionMatrix.Mul4x1(mgl32.Vec3(mesh.Bitangents[i]).Vec4(1)).Vec3()
		}
		meshes = make(map[string]unit.Mesh)
		meshes["terrain"] = mesh
		pv.objectsShownDefault[fileID] = map[string]bool{"terrain": true}
	}
	if info.GeometryGroup.Value != 0x0 {
		geoID := stingray.NewFileID(info.GeometryGroup, stingray.Sum("geometry_group"))
		geoMain, exists, err := getResource(geoID, stingray.DataMain)
		if !exists {
			return fmt.Errorf("%v.geometry_group does not exist", info.GeometryGroup.String())
		} else if err != nil {
			return fmt.Errorf("failed to load %v.geometry_group: %v", info.GeometryGroup.String(), err)
		}
		geoGroup, err := geometrygroup.LoadGeometryGroup(bytes.NewReader(geoMain))
		if err != nil {
			return fmt.Errorf("failed to parse %v.geometry_group: %v", info.GeometryGroup.String(), err)
		}
		geoInfo, ok := geoGroup.MeshInfos[fileID]
		if !ok {
			return fmt.Errorf("%v.geometry_group does not contain %v.unit", info.GeometryGroup.String(), fileID)
		}
		meshInfos := make([]unit.MeshInfo, 0)
		for _, header := range geoInfo.MeshHeaders {
			meshInfos = append(meshInfos, unit.MeshInfo{
				Groups: header.Groups,
				Header: unit.MeshHeader{
					MeshType:  unit.MeshTypeUnknown01,
					LayoutIdx: int32(header.MeshLayoutIndex),
				},
				Materials: header.Materials,
			})
		}
		geoGPU, exists, err := getResource(geoID, stingray.DataGPU)
		if !exists {
			return fmt.Errorf("%v.geometry_group missing gpu data", info.GeometryGroup.String())
		} else if err != nil {
			return fmt.Errorf("failed to load %v.geometry_group gpu data: %v", info.GeometryGroup.String(), err)
		}
		var defaultShown map[string]bool
		meshes, defaultShown, err = pv.loadMeshes(pv.lookupThinHash, meshInfos, geoGroup.MeshLayouts, geoGPU)
		if err != nil {
			return err
		}
		pv.objectsShownDefault[fileID] = defaultShown
	}
	pv.objectsShown[fileID] = maps.Clone(pv.objectsShownDefault[fileID])
	pv.objectsSelected[fileID] = maps.Clone(pv.objectsShownDefault[fileID])
	pv.objects[fileID] = make(map[string]unitPreviewObject)
	pv.meshPositions[fileID] = make(map[string][][3]float32)
	pv.meshNormals[fileID] = make(map[string][][3]float32)
	pv.aabb[fileID] = make(map[string][2]mgl32.Vec3)
	pv.aabbMat[fileID] = make(map[string]mgl32.Mat4, 0)
	pv.boundingBoxes[fileID] = make(map[string]unitPreviewObject)
	pv.unitsShown[fileID] = true
	pv.unitMatrices[fileID] = mgl32.Ident4()
	for name, mesh := range meshes {
		pv.aabb[fileID][name] = [2]mgl32.Vec3{mesh.Info.Header.AABB.Min, mesh.Info.Header.AABB.Max}
		pv.aabbMat[fileID][name] = info.Bones[mesh.Info.Header.AABBTransformIndex].Matrix

		if len(mesh.Positions) == 0 {
			// gui logger warning probably
			fmt.Printf("mesh %v contains no positions\n", name)
			continue
		}
		if len(mesh.Normals) == 0 {
			// gui logger warning probably
			fmt.Printf("mesh %v contains no normals\n", name)
			continue
		}
		if len(mesh.UVCoords) == 0 || len(mesh.UVCoords[0]) == 0 {
			// gui logger warning probably
			fmt.Printf("mesh %v contains no UV coordinates\n", name)
			continue
		}

		object := unitPreviewObject{}
		object.matrix = info.Bones[mesh.Info.Header.TransformIdx].Matrix

		// Create index buffers
		{
			object.genObjects(true, 0)
			object.ibos = make([]uint32, len(mesh.Indices))
			gl.GenBuffers(int32(len(object.ibos)), &object.ibos[0])

			for idx := range object.materials {
				// release textures and delete shaders from old materials
				object.materials[idx].delete(pv.textureCache)
			}
			object.numIndices = make([]int32, len(mesh.Indices))
			object.materials = make([]unitPreviewMaterial, len(mesh.Indices))
			for group := range object.materials {
				if group >= len(mesh.Info.Groups) {
					continue
					//return fmt.Errorf("group %v not found", group)
				}
				idx := mesh.Info.Groups[group].MaterialIdx
				matID := mesh.Info.Materials[idx]
				matFileName, ok := info.Materials[matID]
				if !ok {
					continue
					//return fmt.Errorf("load material: id %v not found", matID.String())
				}
				object.materials[group].id = matFileName
				object.materials[group].name = pv.lookupThinHash(matID)
			}
			if err := pv.loadMaterials(getResource, &object, armorInfo, pv.lookupThinHash); err != nil {
				return err
			}
		}

		if len(mesh.Positions) != len(mesh.UVCoords[0]) {
			return errors.New("expected positions and UVs to have the same length")
		}

		object.numVertices = int32(len(mesh.Positions))

		// Upload object data
		{
			gl.BindVertexArray(object.vao)

			positionsSize := len(mesh.Positions) * 3 * 4
			normalsSize := len(mesh.Normals) * 3 * 4
			uvsSize := len(mesh.UVCoords[0]) * 2 * 4
			tangentsSize := len(mesh.Tangents) * 4 * 4
			bitangentsSize := len(mesh.Bitangents) * 3 * 4

			gl.BindBuffer(gl.ARRAY_BUFFER, object.vbo)
			gl.BufferData(gl.ARRAY_BUFFER, positionsSize+normalsSize+uvsSize*min(len(mesh.UVCoords), 3)+tangentsSize+bitangentsSize, nil, gl.STATIC_DRAW)
			offset := 0
			//
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, positionsSize, gl.Ptr(mesh.Positions))
			gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 3*4, uintptr(offset))
			gl.EnableVertexAttribArray(0)
			offset += positionsSize
			//
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, normalsSize, gl.Ptr(mesh.Normals))
			gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, true, 3*4, uintptr(offset))
			gl.EnableVertexAttribArray(1)
			offset += normalsSize
			//
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, uvsSize, gl.Ptr(mesh.UVCoords[0]))
			gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, 2*4, uintptr(offset))
			gl.EnableVertexAttribArray(2)
			offset += uvsSize
			//
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, tangentsSize, gl.Ptr(mesh.Tangents))
			gl.VertexAttribPointerWithOffset(3, 3, gl.FLOAT, true, 4*4, uintptr(offset))
			gl.EnableVertexAttribArray(3)
			offset += tangentsSize
			//
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, bitangentsSize, gl.Ptr(mesh.Bitangents))
			gl.VertexAttribPointerWithOffset(4, 3, gl.FLOAT, true, 3*4, uintptr(offset))
			gl.EnableVertexAttribArray(4)
			offset += bitangentsSize
			//
			if len(mesh.UVCoords) >= 3 {
				for layer, uvcoords := range mesh.UVCoords[1:3] {
					index := uint32(5 + layer)
					uvsSize := len(uvcoords) * 2 * 4
					fmt.Printf("size %v offset %v index %v\n", uvsSize, offset, index)
					gl.BufferSubData(gl.ARRAY_BUFFER, offset, uvsSize, gl.Ptr(uvcoords))
					gl.VertexAttribPointerWithOffset(index, 2, gl.FLOAT, false, 2*4, uintptr(offset))
					gl.EnableVertexAttribArray(index)
					offset += uvsSize
				}
			}

			object.numIndices = make([]int32, len(mesh.Indices))
			for group, indices := range mesh.Indices {
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, object.ibos[group])
				gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
				object.numIndices[group] = int32(len(indices))
			}

			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		}
		pv.meshPositions[fileID][name] = mesh.Positions
		pv.meshNormals[fileID][name] = mesh.Normals
		pv.objects[fileID][name] = object

		// Upload bounding box data
		{
			boundingBox := unitPreviewObject{}
			boundingBox.genObjects(false, 1)
			gl.BindVertexArray(boundingBox.vao)

			verts := pv.getAABBVertices(fileID, name)
			gl.BindBuffer(gl.ARRAY_BUFFER, boundingBox.vbo)
			defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.BufferData(gl.ARRAY_BUFFER, len(verts)*3*4, gl.Ptr(verts[:]), gl.STATIC_DRAW)

			boundingBox.numIndices[0] = int32(len(aabbIndices))
			boundingBox.numVertices = int32(len(verts))
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, boundingBox.ibos[0])
			defer gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
			gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(boundingBox.numIndices[0]*4), gl.Ptr(aabbIndices[:]), gl.STATIC_DRAW)

			gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
			gl.EnableVertexAttribArray(0)
			pv.boundingBoxes[fileID][name] = boundingBox
		}
	}

	skeletonVertices := make([]mgl32.Vec3, 0)
	skeletonIndices := make([]uint32, 0)
	root := info.Bones[0]
	var recurseSkeleton func(unit.Bone) uint32
	recurseSkeleton = func(curr unit.Bone) uint32 {
		currentVertex := uint32(len(skeletonVertices))
		skeletonVertices = append(skeletonVertices, curr.Matrix.Mul4x1(mgl32.Vec4{0.0, 0.0, 0.0, 1.0}).Vec3())
		for _, idx := range curr.Children {
			childVertexIdx := recurseSkeleton(info.Bones[idx])
			skeletonIndices = append(skeletonIndices, currentVertex, childVertexIdx)
		}
		return currentVertex
	}
	recurseSkeleton(root)

	if len(skeletonIndices) > 0 {
		if pv.skeletons == nil {
			pv.skeletons = make(map[stingray.Hash]unitPreviewObject)
		}
		skeleton := unitPreviewObject{}
		skeleton.genObjects(false, 1)
		gl.BindVertexArray(skeleton.vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, skeleton.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(skeletonVertices)*3*4, gl.Ptr(skeletonVertices), gl.STATIC_DRAW)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, skeleton.ibos[0])
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(skeletonIndices)*4, gl.Ptr(skeletonIndices), gl.STATIC_DRAW)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
		gl.EnableVertexAttribArray(0)
		skeleton.numVertices = int32(len(skeletonVertices))
		skeleton.numIndices[0] = int32(len(skeletonIndices))
		pv.skeletons[fileID] = skeleton
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	pv.model = stingrayToGLCoords

	if pv.autoZoomEnabled {
		pv.doAutoZoomNextFrame = true
	}

	// Calculate max zoom out distance
	{
		// Get origin sphere around mesh
		var maxDistSqrFromOrigin float32
		for name := range pv.objects[fileID] {
			for _, p := range pv.getAABBVertices(fileID, name) {
				maxDistSqrFromOrigin = max(maxDistSqrFromOrigin,
					pv.aabbMat[fileID][name].Mul4x1(p.Vec4(1.0)).Vec3().LenSqr())
			}
			maxDistFromOrigin := float32(math.Sqrt(float64(maxDistSqrFromOrigin)))

			// Calculate camera distance to fit vertical frustum into disk orthogonal
			// to view direction with radius of the sphere. Ideally, we'd want
			// to fit the sphere into the frustum, but the disk should be close
			// enough.
			// tan(vfov/2) = maxDistFromOrigin/viewDistance
			pv.maxViewDistance = float32(float64(maxDistFromOrigin) / math.Tan(float64(pv.vfov/2)))

		}
		// We want to be able to zoom out a bit further.
		pv.maxViewDistance *= 2
	}

	for i := range pv.udimsShownDefault {
		pv.udimsShownDefault[i] = true
		pv.udimNames[i] = ""
	}
	pv.numUdims = 1
	visibilityMasks, err := datalib.ParseVisibilityMasks()
	if err != nil {
		return err
	}
	visibilityMask, ok := visibilityMasks[fileID]
	if !ok {
		entityHash := datalib.UnitsToEntities(fileID)
		visibilityMask, ok = visibilityMasks[entityHash]
	}
	if ok {
		for name := range pv.objects[fileID] {
			object := pv.objects[fileID][name]
			object.hasVisibilityMasks = 1
			pv.objects[fileID][name] = object
			for _, material := range pv.objects[fileID][name].materials {
				gl.UseProgram(material.program)
				gl.Uniform1ui(material.uniforms["hasVisibilityMasks"], 1)
				gl.UseProgram(0)
			}
		}
		for _, info := range visibilityMask.MaskInfos {
			if int(info.Index) >= len(pv.udimsShownDefault) {
				// No support for udims with index > 64 at the moment
				continue
			}
			if info.Name.Value == 0 {
				continue
			}
			pv.numUdims = max(uint32(info.Index)+1, pv.numUdims)

			pv.udimsShownDefault[info.Index] = info.StartHidden == 0
			name, ok := thinhashes[info.Name]
			if !ok {
				name = info.Name.String()
			}
			pv.udimNames[info.Index] = name
		}
	}
	pv.udimsSelected = pv.udimsShownDefault

	if setModelPos {
		for name := range pv.objects[fileID] {
			if shown, contains := pv.objectsShown[fileID][name]; contains && !shown {
				continue
			}
			pv.modelPos = pv.objects[fileID][name].matrix.Inv().Mul4x1(mgl32.Vec4{0, 0, 0, 1})
			break
		}
	}

	pv.loadUnitGetResourceFunc = getResource
	pv.loadedUnits[fileID] = true

	return nil
}

func (pv *UnitPreviewState) computeMVP(aspectRatio float32, animate bool) (
	modelPos mgl32.Vec4,
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
		modelPos = pv.animOrigModelPos.Mul(1 - pv.animTime).Add(pv.modelPos.Mul(pv.animTime))
	} else {
		viewDistance = pv.viewDistance
		viewRotation = pv.viewRotation
		modelPos = pv.modelPos
	}

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
		max(0.001, 0.001*viewDistance),
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

func (pv *UnitPreviewState) getAABBVertices(unit stingray.Hash, name string) [8]mgl32.Vec3 {
	return [8]mgl32.Vec3{
		{pv.aabb[unit][name][0][0], pv.aabb[unit][name][0][1], pv.aabb[unit][name][0][2]},
		{pv.aabb[unit][name][0][0], pv.aabb[unit][name][0][1], pv.aabb[unit][name][1][2]},
		{pv.aabb[unit][name][0][0], pv.aabb[unit][name][1][1], pv.aabb[unit][name][0][2]},
		{pv.aabb[unit][name][0][0], pv.aabb[unit][name][1][1], pv.aabb[unit][name][1][2]},
		{pv.aabb[unit][name][1][0], pv.aabb[unit][name][0][1], pv.aabb[unit][name][0][2]},
		{pv.aabb[unit][name][1][0], pv.aabb[unit][name][0][1], pv.aabb[unit][name][1][2]},
		{pv.aabb[unit][name][1][0], pv.aabb[unit][name][1][1], pv.aabb[unit][name][0][2]},
		{pv.aabb[unit][name][1][0], pv.aabb[unit][name][1][1], pv.aabb[unit][name][1][2]},
	}
}

func sum(s []int32) (result int32) {
	result = 0
	for _, val := range s {
		result += val
	}
	return
}

func (pv *UnitPreviewState) Draw(previewId string) {
	if len(pv.objects) == 0 {
		return
	}
	// for idx := range pv.objects {
	// 	if len(pv.objects[idx].ibos) == 0 {
	// 		return
	// 	}
	// }

	imgui.PushIDStr(previewId)
	defer imgui.PopID()

	viewSize := imgui.ContentRegionAvail()
	viewSize.Y -= imutils.CheckboxHeight()

	if pv.animTime == -1 || pv.animTime >= 1 {
		pv.animOrigViewDistance = pv.viewDistance
		pv.animOrigViewRotation = pv.viewRotation
		pv.animOrigModelPos = pv.modelPos
		pv.animTime = -1
	}

	widgets.GLView(previewId, pv.fb, viewSize,
		func() {
			io := imgui.CurrentIO()

			if imgui.IsItemActive() {
				md := io.MouseDelta()
				md4 := mgl32.Vec4{md.X, -md.Y, 0.0, 1.0}
				if io.KeyShift() && md4.Vec2().LenSqr() > 0 {
					modelPos, _, view, projection := pv.computeMVP(viewSize.X/viewSize.Y, false)
					modelViewProj := projection.Mul4(view).Mul4(pv.model)
					invModelViewProjection := modelViewProj.Inv()

					projected := modelViewProj.Mul4x1(modelPos.Vec3().Vec4(1.0))
					// Set depth to current model position
					md4[2] = projected.Z() / projected.W()

					positionDelta := invModelViewProjection.Mul4x1(md4)
					positionDelta = positionDelta.Mul(1 / positionDelta.W())
					pv.modelPos = modelPos.Add(positionDelta.Mul(io.DeltaTime() / 2).Vec3().Vec4(0.0))
				} else {
					pv.viewRotation = pv.viewRotation.Add(mgl32.Vec2{md.X, md.Y}.Mul(-0.01))
					pv.viewRotation[1] = mgl32.Clamp(pv.viewRotation[1], -1.55, 1.55)
				}
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

			modelPos, viewPosition, view, projection := pv.computeMVP(size.X/size.Y, true)
			translation := mgl32.Translate3D(modelPos.Vec3().Elem())

			// Draw object
			gl.Enable(gl.DEPTH_TEST)
			for hash := range pv.objects {
				for name := range pv.objects[hash] {
					if shown, contains := pv.objectsShown[hash][name]; contains && !shown {
						continue
					}
					model := pv.model.Mul4(translation.Mul4(pv.unitMatrices[hash].Mul4(pv.objects[hash][name].matrix)))
					mvp := projection.Mul4(view).Mul4(model)
					normal := model.Inv().Transpose().Mat3()
					gl.BindVertexArray(pv.objects[hash][name].vao)
					if pv.showWireframe {
						gl.UseProgram(pv.wireframeMaterial.program)
						gl.UniformMatrix4fv(pv.wireframeMaterial.uniforms["mvp"], 1, false, &mvp[0])
						gl.Uniform4fv(pv.wireframeMaterial.uniforms["color"], 1, &pv.wireframeColor[0])
						gl.Uniform1i(pv.wireframeMaterial.uniforms["hasVisibilityMasks"], pv.objects[hash][name].hasVisibilityMasks)
						gl.Uniform1iv(pv.wireframeMaterial.uniforms["udimShown"], 64, &pv.udimsShown[0])
					}
					for group, ibo := range pv.objects[hash][name].ibos {
						if !pv.showWireframe {
							gl.UseProgram(pv.objects[hash][name].materials[group].program)
							gl.UniformMatrix4fv(pv.objects[hash][name].materials[group].uniforms["mvp"], 1, false, &mvp[0])
							gl.UniformMatrix4fv(pv.objects[hash][name].materials[group].uniforms["model"], 1, false, &model[0])
							gl.UniformMatrix3fv(pv.objects[hash][name].materials[group].uniforms["normalMat"], 1, false, &normal[0])
							gl.Uniform3fv(pv.objects[hash][name].materials[group].uniforms["viewPosition"], 1, &viewPosition[0])
							gl.Uniform1iv(pv.objects[hash][name].materials[group].uniforms["udimShown"], 64, &pv.udimsShown[0])
							for _, uniformBlock := range pv.objects[hash][name].materials[group].uniformBlocks {
								gl.BindBufferBase(gl.UNIFORM_BUFFER, uniformBlock.binding, uniformBlock.ubo)
							}
							for idx, texture := range pv.objects[hash][name].materials[group].textures {
								gl.ActiveTexture(gl.TEXTURE0 + uint32(idx))
								gl.BindTexture(texture.target, texture.id)
								glError := gl.GetError()
								if glError != 0 {
									fmt.Printf("[error] binding texture %v (%v) in group %v as target %v generated error %v\n", texture.name.String(), texture.id, group, glutils.GLTarget(texture.target).String(), glutils.GLError(glError).String())
								}
							}
						}
						gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
						gl.DrawElements(gl.TRIANGLES, pv.objects[hash][name].numIndices[group], gl.UNSIGNED_INT, nil)
						if !pv.showWireframe {
							for idx, texture := range pv.objects[hash][name].materials[group].textures {
								gl.ActiveTexture(gl.TEXTURE0 + uint32(idx))
								gl.BindTexture(texture.target, 0)
								glError := gl.GetError()
								if glError != 0 {
									fmt.Printf("[error] unbinding texture %v (%v) in group %v as target %v generated error %v\n", texture.name.String(), texture.id, group, glutils.GLTarget(texture.target).String(), glutils.GLError(glError).String())
								}
							}
						}
					}
				}
			}
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
			gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, 0)
			gl.BindTexture(gl.TEXTURE_2D_ARRAY, 0)
			gl.BindVertexArray(0)
			gl.UseProgram(0)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

			// Draw normal visualization
			if pv.visualizeNormals {
				for hash := range pv.objects {
					for name := range pv.objects[hash] {
						if shown, contains := pv.objectsShown[hash][name]; contains && !shown {
							continue
						}
						model := pv.model.Mul4(translation.Mul4(pv.unitMatrices[hash].Mul4(pv.objects[hash][name].matrix)))
						mvp := projection.Mul4(view).Mul4(model)
						gl.UseProgram(pv.normalVisMaterial.program)
						gl.BindVertexArray(pv.objects[hash][name].vao)
						gl.UniformMatrix4fv(pv.normalVisMaterial.uniforms["mvp"], 1, false, &mvp[0])
						gl.Uniform1f(pv.normalVisMaterial.uniforms["len"], pv.viewDistance*0.02)
						gl.Uniform1iv(pv.normalVisMaterial.uniforms["showTangentBitangent"], 1, &pv.visualizeTangentBitangent)
						gl.Uniform1i(pv.normalVisMaterial.uniforms["hasVisibilityMasks"], pv.objects[hash][name].hasVisibilityMasks)
						gl.Uniform1iv(pv.normalVisMaterial.uniforms["udimShown"], 64, &pv.udimsShown[0])
						for group, ibo := range pv.objects[hash][name].ibos {
							gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
							gl.DrawElements(gl.POINTS, pv.objects[hash][name].numIndices[group], gl.UNSIGNED_INT, nil)
						}
					}
				}
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0) // TODO: Make this not draw duplicate vertices
			}

			// Draw debug object
			if pv.showAABB {
				gl.Disable(gl.DEPTH_TEST)
				gl.UseProgram(pv.boundingBoxMaterial.program)
				gl.Uniform4fv(pv.boundingBoxMaterial.uniforms["color"], 1, &pv.aabbColor[0])
				for hash := range pv.boundingBoxes {
					for name := range pv.boundingBoxes[hash] {
						if shown, contains := pv.objectsShown[hash][name]; contains && !shown {
							continue
						}
						gl.BindVertexArray(pv.boundingBoxes[hash][name].vao)
						{
							model := pv.model.Mul4(translation.Mul4(pv.unitMatrices[hash].Mul4(pv.aabbMat[hash][name])))
							mvp := projection.Mul4(view).Mul4(model)
							gl.UniformMatrix4fv(pv.boundingBoxMaterial.uniforms["mvp"], 1, false, &mvp[0])
						}
						gl.DrawElements(gl.TRIANGLES, pv.boundingBoxes[hash][name].numIndices[0], gl.UNSIGNED_INT, nil)
					}
				}
			}

			if pv.showSkeleton {
				gl.Disable(gl.DEPTH_TEST)
				gl.UseProgram(pv.boundingBoxMaterial.program)
				gl.Uniform4fv(pv.boundingBoxMaterial.uniforms["color"], 1, &pv.skeletonColor[0])

				for hash := range pv.skeletons {
					model := pv.model.Mul4(translation.Mul4(pv.unitMatrices[hash]))
					mvp := projection.Mul4(view).Mul4(model)

					gl.UniformMatrix4fv(pv.boundingBoxMaterial.uniforms["mvp"], 1, false, &mvp[0])
					gl.BindVertexArray(pv.skeletons[hash].vao)
					gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, pv.skeletons[hash].ibos[0])
					gl.DrawElements(gl.LINES, pv.skeletons[hash].numIndices[0], gl.UNSIGNED_INT, nil)
				}
			}

			gl.BindVertexArray(0)
			gl.UseProgram(0)

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
				for hash := range pv.objects {
					for name := range pv.objects[hash] {
						positions := pv.getAABBVertices(hash, name)
						for _, vert := range positions {
							maxCamDistDelta = max(maxCamDistDelta,
								fitVertexCamDistDelta(pv.aabbMat[hash][name].Mul4x1(vert.Vec4(1.0)).Vec3()))
						}
					}
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

			modelPos, _, view, projection := pv.computeMVP(size.X/size.Y, false)
			translation := mgl32.Translate3D(modelPos.Vec3().Elem())

			// Show hovered vertex info
			if pv.visualizeNormals {
				igRelMousePos := imgui.MousePos().Sub(pos)
				mousePos := mgl32.Vec2{
					igRelMousePos.X,
					igRelMousePos.Y,
				}
				var closestPos mgl32.Vec2
				closestDist := float32(math.MaxFloat32)
				var closestIdx int
				for hash := range pv.meshPositions {
					for name := range pv.meshPositions[hash] {
						if shown, contains := pv.objectsShown[hash][name]; contains && !shown {
							continue
						}
						mvp := projection.Mul4(view).Mul4(pv.model.Mul4(translation.Mul4(pv.objects[hash][name].matrix)))
						for i, vtx := range pv.meshPositions[hash][name] {
							v := mvp.Mul4x1(mgl32.Vec3(vtx).Vec4(1.0))
							v = v.Mul(1 / v.W())
							v[0] = (v[0] + 1) * size.X * 0.5
							v[1] = (-v[1] + 1) * size.Y * 0.5
							dist := mousePos.Sub(v.Vec2()).LenSqr()
							if dist < closestDist {
								closestPos = v.Vec2()
								closestDist = dist
								closestIdx = i
							}
						}
						markerPos := pos.Add(imgui.NewVec2(closestPos.X(), closestPos.Y()))
						dl.AddCircleFilled(
							markerPos,
							imutils.S(2),
							imgui.ColorU32Vec4(imgui.NewVec4(1, 0, 0, 1)),
						)
						dl.AddTextVec2(
							markerPos,
							imgui.ColorU32Vec4(imgui.NewVec4(1, 1, 0, 1)),
							fmt.Sprintf("Pos: %v\nNormal: %v", pv.meshPositions[hash][name][closestIdx], pv.meshNormals[hash][name][closestIdx]),
						)
					}
				}
			}
		},
	)

	if imgui.Button(fnt.I.Home) {
		pv.viewRotation = mgl32.Vec2{math.Pi, 0.0}
		pv.modelPos = mgl32.Vec4{0, 0, 0, 1}
		for hash := range pv.objects {
			for name := range pv.objects[hash] {
				if shown, contains := pv.objectsShown[hash][name]; contains && !shown {
					continue
				}
				pv.modelPos = pv.objects[hash][name].matrix.Inv().Mul4x1(mgl32.Vec4{0, 0, 0, 1})
				break
			}
		}
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
		var numVertices, numIndices int = 0, 0
		for hash := range pv.objects {
			for name := range pv.objects[hash] {
				numVertices += int(pv.objects[hash][name].numVertices)
				indexCount := sum(pv.objects[hash][name].numIndices)
				numIndices += int(indexCount)
			}
		}
		imutils.Textf("Indices: %v", numIndices)
		imutils.Textf("Vertices: %v", numVertices)
		imutils.Textf("Triangles: %v", numIndices/3)
		imgui.Unindent()

		imgui.Separator()

		const colorPickerFlags = imgui.ColorEditFlagsNoInputs | imgui.ColorEditFlagsAlphaBar | imgui.ColorEditFlagsNoLabel
		imgui.TextUnformatted("Display")
		imgui.Indent()
		imgui.Checkbox("Wireframe mode", &pv.showWireframe)
		imgui.SameLineV(imutils.S(170), -1)
		imgui.ColorEdit4V("Wireframe color", &pv.wireframeColor, colorPickerFlags)

		imgui.Checkbox("Show Skeleton", &pv.showSkeleton)
		imgui.SameLineV(imutils.S(170), -1)
		imgui.ColorEdit4V("Skeleton color", &pv.skeletonColor, colorPickerFlags)

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
	imgui.BeginDisabledV(pv.numUdims <= 1)
	label := "Visibility Masks Selection"
	if pv.numUdims <= 1 || !pv.udimsSettingsShown {
		label = "Show " + label
	} else {
		label = "Hide " + label
	}
	if imgui.Button(label) {
		pv.udimsSettingsShown = !pv.udimsSettingsShown
	}
	if pv.numUdims <= 1 {
		imgui.SetItemTooltip("Unit has no visibility masks")
	}
	imgui.EndDisabled()
	if !pv.udimsSettingsShown || !pv.udimsSettingsDrawn {
		for i := range pv.udimsShown {
			if pv.udimsSelected[i] {
				pv.udimsShown[i] = 1
			} else {
				pv.udimsShown[i] = 0
			}
		}
	}

	imgui.SameLine()
	label = "Mesh Selection"
	if !pv.objectsSettingsShown {
		label = "Show " + label
	} else {
		label = "Hide " + label
	}
	if imgui.Button(label) {
		pv.objectsSettingsShown = !pv.objectsSettingsShown
	}

	imgui.SameLine()
	label = "Material Settings Editor"
	if !pv.materialSettingsShown {
		label = "Show " + label
	} else {
		label = "Hide " + label
	}
	if imgui.Button(label) {
		pv.materialSettingsShown = !pv.materialSettingsShown
	}

	if pv.animTime != -1 {
		pv.animTime += 5 * imgui.CurrentIO().DeltaTime()
	}

	if pv.doSweep {
		// We sweep the textures here so we aren't modifying opengl state during a draw call
		pv.textureCache.Sweep()
		pv.doSweep = false
	}

	currentArchives := pv.getSelectedArchives()
	if len(pv.previousSelectedArchives) != len(currentArchives) {
		for hash := range pv.objects {
			var armorInfo *datalib.UnitData
			for idx := range currentArchives {
				var set datalib.ArmorSet
				var contains bool
				if set, contains = pv.armorSets[currentArchives[idx]]; !contains {
					continue
				}

				value, contains := set.UnitMetadata[hash]
				if !contains {
					continue
				}

				armorInfo = &value
			}

			for name := range pv.objects[hash] {
				object := pv.objects[hash][name]
				pv.loadMaterials(pv.loadUnitGetResourceFunc, &object, armorInfo, pv.lookupThinHash)
				pv.objects[hash][name] = object
			}
		}

		pv.previousSelectedArchives = currentArchives
	}
}

func (pv *UnitPreviewState) DrawSettings() {
	if pv.udimsSettingsShown && pv.numUdims > 1 {
		pv.drawVisibilityMaskSelector()
	}
	if pv.objectsSettingsShown {
		pv.drawMeshSelector()
	}
	if pv.materialSettingsShown {
		pv.drawMaterialSettingsEditor()
	}
}

func (pv *UnitPreviewState) drawMaterialSettingsEditor() {
	defer imgui.End()
	if pv.materialSettingsDrawn = imgui.BeginV(fnt.I.Settings+" Material Settings Editor", &pv.materialSettingsShown, imgui.WindowFlagsNoFocusOnAppearing); !pv.materialSettingsDrawn {
		return
	}
	sortedUnitKeys := slices.SortedFunc(maps.Keys(pv.objects), stingray.Hash.Cmp)
	for _, hash := range sortedUnitKeys {
		root := imgui.CursorScreenPos()
		size := imgui.NewVec2(imgui.ContentRegionAvail().X, imgui.FontSize())
		imgui.PushIDStr(hash.String())
		defer imgui.PopID()
		shown := pv.unitsShown[hash]
		var icon string
		if shown {
			icon = fnt.I.Visibility
		} else {
			icon = fnt.I.VisibilityOff
		}
		imutils.Textf(fmt.Sprintf("%s %s", icon, filepath.Base(pv.lookupHash(hash))))
		imgui.SetCursorScreenPos(root)
		imgui.SetNextItemAllowOverlap()
		if imgui.InvisibleButton("btnUnit", size) {
			pv.unitsShown[hash] = !shown
			//pv.unitsSelected[hash] = !shown
		}
		if !shown {
			continue
		}
		sortedObjectKeys := slices.Sorted(maps.Keys(pv.objects[hash]))
		for _, name := range sortedObjectKeys {
			imgui.Indent()
			shown := pv.objectsShown[hash][name]
			var icon string
			if shown {
				icon = fnt.I.Visibility
			} else {
				icon = fnt.I.VisibilityOff
			}
			imgui.PushIDStr(hash.String() + name)
			defer imgui.PopID()
			pos := imgui.CursorScreenPos()
			imutils.Textf(fmt.Sprintf("%s %s", icon, name))
			imgui.SetCursorScreenPos(pos)
			imgui.Unindent()
			imgui.SetNextItemAllowOverlap()
			if imgui.InvisibleButton("btn", size) {
				pv.objectsShown[hash][name] = !shown
				pv.objectsSelected[hash][name] = !shown
			}
			// pos = imgui.CursorScreenPos()
			// pos.X = pos.X + imgui.CurrentStyle().IndentSpacing()
			// imgui.SetCursorScreenPos(pos)
			if !shown {
				continue
			}
			imgui.Indent()
			for idx := range pv.objects[hash][name].materials {
				imgui.Indent()
				imutils.Textf(fmt.Sprintf("%s %s", fnt.I.Texture, pv.objects[hash][name].materials[idx].name))
				for block := range pv.objects[hash][name].materials[idx].uniformBlocks {
					imgui.Indent()
					uniformBlock := pv.objects[hash][name].materials[idx].uniformBlocks[block]
					sortedSettingsKeys := slices.Sorted(maps.Keys(uniformBlock.currentValues))
					longest := slices.MaxFunc(sortedSettingsKeys, func(a, b string) int { return cmp.Compare(len(a), len(b)) })
					for _, key := range sortedSettingsKeys {
						imgui.PushIDStr(fmt.Sprintf("%v %v", block, key))
						defer imgui.PopID()
						style := imgui.CurrentStyle()
						imgui.SetNextItemWidth(size.X - style.IndentSpacing()*3 - (style.ItemInnerSpacing().X + imgui.CalcTextSize(longest).X))
						var n int32 = 4
						switch uniformBlock.uniformTypes[key] {
						case glutils.GL_BOOL:
							var val uint32
							uniformBlock.get(key, &val)
							result := val != 0
							if imgui.Checkbox(key, &result) {
								uniformBlock.set(key, []uint32{(val + 1) % 2})
							}
						case glutils.GL_BYTE:
							var val int8
							uniformBlock.get(key, &val)
							oldval := val
							if imgui.InputScalar(key, imgui.DataTypeS8, uintptr(unsafe.Pointer(&val))) && oldval != val {
								uniformBlock.set(key, &val)
							}
						case glutils.GL_UNSIGNED_BYTE:
							var val uint8
							uniformBlock.get(key, &val)
							oldval := val
							if imgui.InputScalar(key, imgui.DataTypeU8, uintptr(unsafe.Pointer(&val))) && oldval != val {
								uniformBlock.set(key, &val)
							}
						case glutils.GL_SHORT:
							var val int16
							uniformBlock.get(key, &val)
							oldval := val
							if imgui.InputScalar(key, imgui.DataTypeS16, uintptr(unsafe.Pointer(&val))) && oldval != val {
								uniformBlock.set(key, &val)
							}
						case glutils.GL_UNSIGNED_SHORT:
							var val uint16
							uniformBlock.get(key, &val)
							oldval := val
							if imgui.InputScalar(key, imgui.DataTypeU16, uintptr(unsafe.Pointer(&val))) && oldval != val {
								uniformBlock.set(key, &val)
							}
						case glutils.GL_INT:
							var val int32
							uniformBlock.get(key, &val)
							oldval := val
							if imgui.InputScalar(key, imgui.DataTypeS32, uintptr(unsafe.Pointer(&val))) && oldval != val {
								uniformBlock.set(key, &val)
							}
						case glutils.GL_UNSIGNED_INT:
							var val uint32
							uniformBlock.get(key, &val)
							oldval := val
							if imgui.InputScalar(key, imgui.DataTypeU32, uintptr(unsafe.Pointer(&val))) && oldval != val {
								uniformBlock.set(key, &val)
							}
						case glutils.GL_HALF_FLOAT:
							var val float16.Float16
							uniformBlock.get(key, &val)
							floatVal := val.Float32()
							oldval := floatVal
							if imgui.InputScalar(key, imgui.DataTypeFloat, uintptr(unsafe.Pointer(&floatVal))) && oldval != floatVal {
								val = float16.Fromfloat32(floatVal)
								uniformBlock.set(key, &val)
							}
						case glutils.GL_FLOAT:
							var val float32
							uniformBlock.get(key, &val)
							oldval := val
							if imgui.InputScalar(key, imgui.DataTypeFloat, uintptr(unsafe.Pointer(&val))) && oldval != val {
								uniformBlock.set(key, &val)
							}
						case glutils.GL_FLOAT_VEC2:
							n = min(2, n)
							fallthrough
						case glutils.GL_FLOAT_VEC3:
							n = min(3, n)
							fallthrough
						case glutils.GL_FLOAT_VEC4:
							n = min(4, n)
							vals := make([]float32, n)
							uniformBlock.get(key, vals)
							if imgui.InputScalarN(key, imgui.DataTypeFloat, uintptr(unsafe.Pointer(&vals[0])), n) {
								uniformBlock.set(key, vals)
							}
						}
					}
					imgui.Unindent()
				}
				imgui.Unindent()
			}
			imgui.Unindent()
		}
	}
}

func (pv *UnitPreviewState) drawVisibilityMaskSelector() {
	if pv.udimsSettingsDrawn = imgui.BeginV(fnt.I.FolderEye+" Visibility Mask Selection", &pv.udimsSettingsShown, imgui.WindowFlagsNoFocusOnAppearing); pv.udimsSettingsDrawn {
		nextActiveUDimListItem := int32(-1)
		nextHoveredUDimListItem := int32(-1)
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
			imgui.WindowDrawList().AddRectV(draggingMinPos, draggingMaxPos, imgui.ColorU32Col(imgui.ColButtonActive), 0, 0, imgui.DrawFlagsNone)
		}
		pv.activeUDimListItem = nextActiveUDimListItem
		pv.hoveredUDimListItem = nextHoveredUDimListItem
	}
	imgui.End()
}

func (pv *UnitPreviewState) drawMeshSelector() {
	if pv.udimsSettingsDrawn = imgui.BeginV(fnt.I.FolderEye+" Mesh Selection", &pv.udimsSettingsShown, imgui.WindowFlagsNoFocusOnAppearing); pv.udimsSettingsDrawn {
		nextActiveMeshListItem := int32(-1)
		nextHoveredMeshListItem := int32(-1)
		if imgui.Button("Reset") {
			pv.objectsSelected = pv.objectsShownDefault
		}
		imgui.Separator()
		imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing,
			imgui.NewVec2(imgui.CurrentStyle().ItemSpacing().X, 0))
		dragging := pv.activeMeshListItem != -1 && pv.hoveredMeshListItem != -1
		var draggingMin, draggingMax int32
		if dragging {
			draggingMin = min(pv.activeMeshListItem, pv.hoveredMeshListItem)
			draggingMax = max(pv.activeMeshListItem, pv.hoveredMeshListItem)
		}
		var draggingMinPos, draggingMaxPos imgui.Vec2
		sortedUnitKeys := slices.SortedFunc(maps.Keys(pv.objects), stingray.Hash.Cmp)
		for _, hash := range sortedUnitKeys {
			sortedObjectKeys := slices.Sorted(maps.Keys(pv.objects[hash]))
			for i := range int32(len(pv.objects[hash])) {
				selected := pv.objectsSelected[hash][sortedObjectKeys[i]]
				if dragging {
					if i >= draggingMin && i <= draggingMax {
						selected = !selected
					}
					if imgui.IsMouseClickedBool(imgui.MouseButtonRight) {
						imgui.CurrentContext().SetActiveId(0)
					}
				}
				pv.objectsShown[hash][sortedObjectKeys[i]] = selected
				if imgui.IsMouseReleased(imgui.MouseButtonLeft) {
					pv.objectsSelected[hash][sortedObjectKeys[i]] = selected
				}
				var icon string
				if selected {
					icon = fnt.I.Visibility
				} else {
					icon = fnt.I.VisibilityOff
				}
				imgui.PushIDStr(hash.String() + fmt.Sprint(i))
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
				imutils.Textf(fmt.Sprintf("%s %s", icon, sortedObjectKeys[i]))
				imgui.SetCursorScreenPos(pos)
				imgui.SetNextItemAllowOverlap()
				imgui.InvisibleButton("btn", size)
				if imgui.IsItemActive() {
					nextActiveMeshListItem = i
				}
				hovered := imgui.ItemStatusFlags(imgui.CurrentContext().LastItemData().CData.StatusFlags)&imgui.ItemStatusFlagsHoveredRect != 0
				if hovered {
					nextHoveredMeshListItem = i
				}
				imgui.SetItemTooltip(`Click to toggle item visibility
Drag to toggle multiple items (right-click to cancel)`)
				imgui.PopID()
			}
		}
		imgui.PopStyleVar()
		if dragging {
			imgui.WindowDrawList().AddRectV(draggingMinPos, draggingMaxPos, imgui.ColorU32Col(imgui.ColButtonActive), 0, 0, imgui.DrawFlagsNone)
		}
		pv.activeMeshListItem = nextActiveMeshListItem
		pv.hoveredMeshListItem = nextHoveredMeshListItem
	}
	imgui.End()
}
