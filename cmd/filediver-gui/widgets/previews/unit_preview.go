package previews

import (
	"bytes"
	"cmp"
	"embed"
	"errors"
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

var lutTextureNames = []string{
	"decal_sheet",
	"customization_camo_tiler_array",
	"customization_material_detail_tiler_array",
	"pattern_lut",
	"base_data",
	"material_lut",
	"pattern_masks_array",
	"id_masks_array",
}

func setupTexture(textureID uint32) {
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
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

type unitPreviewMaterial struct {
	program  uint32
	uniforms unitPreviewUniforms
	textures []uint32
	targets  []uint32
}

func (mat *unitPreviewMaterial) generate(shaderPaths []string, textures int, uniforms []string) error {
	var err error
	mat.program, err = glutils.CreateProgramFromSources(
		unitPreviewShaderCode,
		shaderPaths...,
	)
	if err != nil {
		return err
	}
	mat.textures = make([]uint32, textures)
	mat.targets = make([]uint32, textures)
	if len(mat.textures) > 0 {
		gl.GenTextures(int32(textures), &mat.textures[0])
	}

	mat.uniforms.generate(mat.program, uniforms...)

	return nil
}

func (mat *unitPreviewMaterial) delete() {
	if len(mat.textures) > 0 {
		gl.DeleteTextures(int32(len(mat.textures)), &mat.textures[0])
	}
	gl.DeleteProgram(mat.program)

}

type unitPreviewObject struct {
	vao       uint32   // vertex array object
	ibos      []uint32 // index buffer objects
	vbo       uint32   // vertex buffer object
	materials []unitPreviewMaterial

	numVertices int32
	numIndices  []int32
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

// Panicks if a name is not a uniform.
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

func (obj unitPreviewObject) deleteObjects() {
	gl.DeleteVertexArrays(1, &obj.vao)
	gl.DeleteBuffers(1, &obj.vbo)
	if len(obj.ibos) > 0 {
		gl.DeleteBuffers(int32(len(obj.ibos)), &obj.ibos[0])
	}
	for _, material := range obj.materials {
		material.delete()
	}
}

type UnitPreviewState struct {
	fb *widgets.GLViewState

	object            unitPreviewObject
	wireframeMaterial unitPreviewMaterial

	normalVisMaterial unitPreviewMaterial

	dbgObjProgram  uint32
	dbgObj         unitPreviewObject
	dbgObjUniforms unitPreviewUniforms

	vfov         float32
	modelPos     mgl32.Vec4
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

	// For fitting mesh to screen and debug info
	meshPositions [][3]float32
	meshNormals   [][3]float32

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

func NewUnitPreview() (*UnitPreviewState, error) {
	var err error

	pv := &UnitPreviewState{}

	pv.fb, err = widgets.NewGLView()
	if err != nil {
		return nil, err
	}

	pv.object.genObjects(true, 0)

	err = pv.wireframeMaterial.generate(
		[]string{
			"shaders/object_wireframe.vert",
			"shaders/object_wireframe.geom",
			"shaders/object_wireframe.frag",
		},
		0,
		[]string{"mvp", "color", "udimShown"},
	)
	if err != nil {
		return nil, err
	}

	err = pv.normalVisMaterial.generate(
		[]string{
			"shaders/object_normal_vis.vert",
			"shaders/object_normal_vis.geom",
			"shaders/object_normal_vis.frag",
		},
		0,
		[]string{"mvp", "len", "showTangentBitangent", "udimShown"},
	)
	if err != nil {
		return nil, err
	}

	pv.dbgObj.genObjects(false, 1)
	pv.dbgObjProgram, err = glutils.CreateProgramFromSources(unitPreviewShaderCode,
		"shaders/debug_object.vert",
		"shaders/debug_object.frag",
	)
	if err != nil {
		return nil, err
	}
	pv.dbgObjUniforms.generate(pv.dbgObjProgram, "mvp", "color")

	pv.vfov = mgl32.DegToRad(60)
	pv.viewDistance = 25

	pv.wireframeColor = [4]float32{1.0, 1.0, 1.0, 0.5}
	pv.aabbColor = [4]float32{0.3, 0.3, 0.8, 0.2}

	return pv, nil
}

func (pv *UnitPreviewState) Delete() {
	pv.fb.Delete()
	pv.object.deleteObjects()
	pv.wireframeMaterial.delete()
	pv.dbgObj.deleteObjects()
}

func (pv *UnitPreviewState) loadMesh(meshInfos []unit.MeshInfo, meshLayouts []unit.MeshLayout, gpuData []byte) (unit.Mesh, error) {
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

func uploadStingrayTexture(getResource GetResourceFunc, textureID uint32, fileName stingray.Hash) error {
	file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
	var texMain, texStream, texGPU []byte
	var err error
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

func uploadStingrayTextureArray(getResource GetResourceFunc, textureID uint32, fileName stingray.Hash) error {
	file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
	var texMain, texStream, texGPU []byte
	var err error
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
	completeRaw := make([]uint8, 0, len(dds.Images[0].MipMaps[0].Raw)*len(dds.Images))
	for _, img := range dds.Images {
		completeRaw = append(completeRaw, img.MipMaps[0].Raw...)
	}
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, textureID)
	gl.TexImage3D(gl.TEXTURE_2D_ARRAY, 0, gl.RGBA, int32(dds.Bounds().Dx()), int32(dds.Bounds().Dy()), int32(len(dds.Images)), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(completeRaw))
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, 0)
	return nil
}

func uploadStingrayLUT(getResource GetResourceFunc, textureID uint32, fileName stingray.Hash) error {
	file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
	var texMain, texStream, texGPU []byte
	var err error
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
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, int32(dds.Image.Bounds().Dx()), int32(dds.Image.Bounds().Dy()), 0, gl.RGBA, gl.HALF_FLOAT, gl.Ptr(dds.Images[0].MipMaps[0].Raw))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	return nil
}

var baseUniforms = []string{"mvp", "model", "normalMat", "viewPosition", "udimShown"}

func (pv *UnitPreviewState) useBasicMaterial(getResource GetResourceFunc, info *unit.Info, mesh unit.Mesh, group int, mat *material.Material) error {
	err := pv.object.materials[group].generate(
		[]string{"shaders/object.vert", "shaders/object.frag"},
		2,
		append(baseUniforms, "texAlbedo", "texNormal", "shouldReconstructNormalZ"),
	)
	if err != nil {
		return err
	}

	for _, texture := range pv.object.materials[group].textures {
		setupTexture(texture)
	}

	gl.UseProgram(pv.object.materials[group].program)
	gl.Uniform1i(pv.object.materials[group].uniforms["texAlbedo"], 0)
	gl.Uniform1i(pv.object.materials[group].uniforms["texNormal"], 1)
	gl.UseProgram(0)

	// Upload object texture
	albedoTexFileName, albedoRemoveAlpha, normalTexFileName, reconstructNormalZ, err := func() (albedoFileName stingray.Hash, albedoRemoveAlpha bool, normalFileName stingray.Hash, reconstructNormalZ bool, err error) {
		// TODO: Use all textures somehow. Currently, simply the first one
		// found is used.
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

	if albedoTexFileName.Value != 0 {
		if err := uploadStingrayTexture(getResource, pv.object.materials[group].textures[0], albedoTexFileName); err != nil {
			return err
		}
		if albedoRemoveAlpha {
			gl.BindTexture(gl.TEXTURE_2D, pv.object.materials[group].textures[0])
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, gl.ONE)
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}
	} else {
		data := []byte{255, 255, 255, 255}
		gl.BindTexture(gl.TEXTURE_2D, pv.object.materials[group].textures[0])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	if normalTexFileName.Value != 0 {
		if err := uploadStingrayTexture(getResource, pv.object.materials[group].textures[1], normalTexFileName); err != nil {
			return err
		}
	} else {
		data := []byte{128, 128, 255, 128}
		gl.BindTexture(gl.TEXTURE_2D, pv.object.materials[group].textures[1])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
		gl.BindTexture(gl.TEXTURE_2D, 0)
		reconstructNormalZ = false
	}

	pv.object.materials[group].targets[0] = gl.TEXTURE_2D
	pv.object.materials[group].targets[1] = gl.TEXTURE_2D

	gl.UseProgram(pv.object.materials[group].program)
	if reconstructNormalZ {
		gl.Uniform1i(pv.object.materials[group].uniforms["shouldReconstructNormalZ"], 1)
	} else {
		gl.Uniform1i(pv.object.materials[group].uniforms["shouldReconstructNormalZ"], 0)
	}
	gl.UseProgram(0)
	return nil
}

func isLUTMaterial(mat *material.Material) bool {
	_, containsIdMasks := mat.Textures[stingray.Sum("id_masks_array").Thin()]
	_, containsMaterialLut := mat.Textures[stingray.Sum("material_lut").Thin()]
	return containsIdMasks && containsMaterialLut
}

func (pv *UnitPreviewState) useLUTMaterial(getResource GetResourceFunc, info *unit.Info, mesh unit.Mesh, group int, mat *material.Material) error {
	err := pv.object.materials[group].generate(
		[]string{"shaders/object.vert", "shaders/lut.frag"},
		len(lutTextureNames),
		append(baseUniforms, lutTextureNames...),
	)
	if err != nil {
		return err
	}

	for idx, texture := range pv.object.materials[group].textures {
		switch lutTextureNames[idx] {
		case "customization_camo_tiler_array", "customization_material_detail_tiler_array", "pattern_masks_array", "id_masks_array":
			setupTextureArray(texture)
			pv.object.materials[group].targets[idx] = gl.TEXTURE_2D_ARRAY
		default:
			setupTexture(texture)
			pv.object.materials[group].targets[idx] = gl.TEXTURE_2D
		}
		textureHash, ok := mat.Textures[stingray.Sum(lutTextureNames[idx]).Thin()]
		if !ok {
			continue
		}
		switch lutTextureNames[idx] {
		case "customization_camo_tiler_array", "customization_material_detail_tiler_array", "pattern_masks_array", "id_masks_array":
			uploadStingrayTextureArray(getResource, texture, textureHash)
		case "pattern_lut", "material_lut":
			uploadStingrayLUT(getResource, texture, textureHash)
		default:
			uploadStingrayTexture(getResource, texture, textureHash)
		}
	}

	gl.UseProgram(pv.object.materials[group].program)
	for index, textureName := range lutTextureNames {
		gl.Uniform1i(pv.object.materials[group].uniforms[textureName], int32(index))
	}

	gl.UseProgram(0)
	return nil
}

func loadMaterial(getResource GetResourceFunc, info *unit.Info, mesh unit.Mesh, group int) (*material.Material, error) {
	idx := mesh.Info.Groups[group].MaterialIdx
	matID := mesh.Info.Materials[idx]
	matFileName, ok := info.Materials[matID]
	if !ok {
		return nil, fmt.Errorf("load material: id %v not found", matID.String())
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

func (pv *UnitPreviewState) LoadUnit(fileID stingray.Hash, mainData, gpuData []byte, getResource GetResourceFunc, thinhashes map[stingray.ThinHash]string) error {
	info, err := unit.LoadInfo(bytes.NewReader(mainData))
	if err != nil {
		return err
	}

	if len(info.MeshInfos) == 0 && len(info.TerrainInfos) == 0 && info.GeometryGroup.Value == 0x0 {
		return fmt.Errorf("unit contains no meshes")
	}

	var mesh unit.Mesh
	if len(info.MeshInfos) > 0 {
		mesh, err = pv.loadMesh(info.MeshInfos, info.MeshLayouts, gpuData)
		if err != nil {
			return err
		}
	} else if len(info.TerrainInfos) > 0 {
		mesh, err = unit.LoadTerrain(info.TerrainInfos[0])
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
		mesh, err = pv.loadMesh(meshInfos, geoGroup.MeshLayouts, geoGPU)
		if err != nil {
			return err
		}
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

	// Create index buffers
	{
		if len(pv.object.ibos) != 0 {
			gl.DeleteBuffers(int32(len(pv.object.ibos)), &pv.object.ibos[0])
		}
		pv.object.ibos = make([]uint32, len(mesh.Indices))
		gl.GenBuffers(int32(len(pv.object.ibos)), &pv.object.ibos[0])

		pv.object.numIndices = make([]int32, len(mesh.Indices))
		pv.object.materials = make([]unitPreviewMaterial, len(mesh.Indices))
		for group := range pv.object.materials {
			mat, err := loadMaterial(getResource, info, mesh, group)
			if err != nil {
				return err
			}
			if isLUTMaterial(mat) {
				if err := pv.useLUTMaterial(getResource, info, mesh, group, mat); err == nil {
					continue
				} else {
					fmt.Printf("got error when enabling lut material: %v\n", err)
				}
				// fall back to basic material if the lut material fails to load
			}
			if err := pv.useBasicMaterial(getResource, info, mesh, group, mat); err != nil {
				return err
			}
		}
	}

	if len(mesh.Positions) != len(mesh.UVCoords[0]) {
		return errors.New("expected positions and UVs to have the same length")
	}

	pv.object.numVertices = int32(len(mesh.Positions))

	pv.numUdims = 0
	for _, uv := range mesh.UVCoords[0] {
		udim := uint32(uv[0]) | uint32(1-uv[1])<<5
		pv.numUdims = max(pv.numUdims, udim+1)
	}
	if pv.numUdims >= 64 {
		pv.numUdims = 1
	}

	// Upload object data
	{
		gl.BindVertexArray(pv.object.vao)

		positionsSize := len(mesh.Positions) * 3 * 4
		normalsSize := len(mesh.Normals) * 3 * 4
		uvsSize := len(mesh.UVCoords[0]) * 2 * 4
		tangentsSize := len(mesh.Tangents) * 4 * 4
		bitangentsSize := len(mesh.Bitangents) * 3 * 4

		gl.BindBuffer(gl.ARRAY_BUFFER, pv.object.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, positionsSize+normalsSize+uvsSize+tangentsSize+bitangentsSize, nil, gl.STATIC_DRAW)
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

		for group, indices := range mesh.Indices {
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, pv.object.ibos[group])
			gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
			pv.object.numIndices[group] = int32(len(indices))
		}

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	}

	pv.meshPositions = mesh.Positions
	pv.meshNormals = mesh.Normals

	// Upload debug object data
	{
		gl.BindVertexArray(pv.dbgObj.vao)

		verts := pv.getAABBVertices()
		gl.BindBuffer(gl.ARRAY_BUFFER, pv.dbgObj.vbo)
		defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BufferData(gl.ARRAY_BUFFER, len(verts)*3*4, gl.Ptr(verts[:]), gl.STATIC_DRAW)

		pv.dbgObj.numIndices[0] = int32(len(aabbIndices))
		pv.dbgObj.numVertices = int32(len(verts))
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(pv.dbgObj.numIndices[0]*4), gl.Ptr(aabbIndices[:]), gl.STATIC_DRAW)

		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
		gl.EnableVertexAttribArray(0)

		gl.BindVertexArray(0)
	}

	pv.model = stingrayToGLCoords
	pv.modelPos = mgl32.Vec4{0, 0, 0, 1}

	if pv.autoZoomEnabled {
		pv.doAutoZoomNextFrame = true
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

func (pv *UnitPreviewState) computeMVP(aspectRatio float32, animate bool) (
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

func sum(s []int32) (result int32) {
	result = 0
	for _, val := range s {
		result += val
	}
	return
}

func UnitPreview(name string, pv *UnitPreviewState) {
	if len(pv.object.ibos) == 0 {
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
				md4 := mgl32.Vec4{md.X, -md.Y, 0.0, 1.0}
				if io.KeyShift() && md4.Vec2().LenSqr() > 0 {
					_, _, view, projection := pv.computeMVP(viewSize.X/viewSize.Y, false)
					modelViewProj := projection.Mul4(view).Mul4(pv.model)
					invModelViewProjection := modelViewProj.Inv()

					projected := modelViewProj.Mul4x1(pv.modelPos.Vec3().Vec4(1.0))
					// Set depth to current model position
					md4[2] = projected.Z() / projected.W()

					positionDelta := invModelViewProjection.Mul4x1(md4)
					positionDelta = positionDelta.Mul(1 / positionDelta.W())
					pv.modelPos = pv.modelPos.Add(positionDelta.Mul(io.DeltaTime() / 2).Vec3().Vec4(0.0))
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

			normal, viewPosition, view, projection := pv.computeMVP(size.X/size.Y, true)
			translation := mgl32.Translate3D(pv.modelPos.Vec3().Elem())
			mvp := projection.Mul4(view).Mul4(pv.model.Mul4(translation))

			// Draw object
			gl.Enable(gl.DEPTH_TEST)
			gl.BindVertexArray(pv.object.vao)
			if pv.showWireframe {
				gl.UseProgram(pv.wireframeMaterial.program)
				gl.UniformMatrix4fv(pv.wireframeMaterial.uniforms["mvp"], 1, false, &mvp[0])
				gl.Uniform4fv(pv.wireframeMaterial.uniforms["color"], 1, &pv.wireframeColor[0])
				gl.Uniform1iv(pv.wireframeMaterial.uniforms["udimShown"], 64, &pv.udimsShown[0])
			}
			for group, ibo := range pv.object.ibos {
				if !pv.showWireframe {
					gl.UseProgram(pv.object.materials[group].program)
					gl.UniformMatrix4fv(pv.object.materials[group].uniforms["mvp"], 1, false, &mvp[0])
					gl.UniformMatrix4fv(pv.object.materials[group].uniforms["model"], 1, false, &pv.model[0])
					gl.UniformMatrix3fv(pv.object.materials[group].uniforms["normalMat"], 1, false, &normal[0])
					gl.Uniform3fv(pv.object.materials[group].uniforms["viewPosition"], 1, &viewPosition[0])
					gl.Uniform1iv(pv.object.materials[group].uniforms["udimShown"], 64, &pv.udimsShown[0])
					for idx, texture := range pv.object.materials[group].textures {
						target := pv.object.materials[group].targets[idx]
						gl.ActiveTexture(gl.TEXTURE0 + uint32(idx))
						gl.BindTexture(target, texture)
					}
				}
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
				gl.DrawElements(gl.TRIANGLES, pv.object.numIndices[group], gl.UNSIGNED_INT, nil)
			}
			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, 0)
			gl.BindVertexArray(0)
			gl.UseProgram(0)
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

			// Draw normal visualization
			if pv.visualizeNormals {
				gl.UseProgram(pv.normalVisMaterial.program)
				gl.BindVertexArray(pv.object.vao)
				gl.UniformMatrix4fv(pv.normalVisMaterial.uniforms["mvp"], 1, false, &mvp[0])
				gl.Uniform1f(pv.normalVisMaterial.uniforms["len"], pv.viewDistance*0.02)
				gl.Uniform1iv(pv.normalVisMaterial.uniforms["showTangentBitangent"], 1, &pv.visualizeTangentBitangent)
				gl.Uniform1iv(pv.normalVisMaterial.uniforms["udimShown"], 64, &pv.udimsShown[0])
				for group, ibo := range pv.object.ibos {
					gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
					gl.DrawElements(gl.POINTS, pv.object.numIndices[group], gl.UNSIGNED_INT, nil)
				}
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0) // TODO: Make this not draw duplicate vertices
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
				gl.DrawElements(gl.TRIANGLES, pv.dbgObj.numIndices[0], gl.UNSIGNED_INT, nil)
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

			_, _, view, projection := pv.computeMVP(size.X/size.Y, false)
			mvp := projection.Mul4(view).Mul4(pv.model)

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
				for i, vtx := range pv.meshPositions {
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
					fmt.Sprintf("Pos: %v\nNormal: %v", pv.meshPositions[closestIdx], pv.meshNormals[closestIdx]),
				)
			}
		},
	)

	if imgui.Button(fnt.I.Home) {
		pv.viewRotation = mgl32.Vec2{}
		pv.modelPos = mgl32.Vec4{0, 0, 0, 1}
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
		indexCount := sum(pv.object.numIndices)
		imutils.Textf("Indices: %v", indexCount)
		imutils.Textf("Vertices: %v", pv.object.numVertices)
		imutils.Textf("Triangles: %v", indexCount/3)
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
			imgui.WindowDrawList().AddRectV(draggingMinPos, draggingMaxPos, imgui.ColorU32Col(imgui.ColButtonActive), 0, 0, imgui.DrawFlagsNone)
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
