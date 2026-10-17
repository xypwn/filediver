package materials

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"maps"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/dds"
	extr_entity "github.com/xypwn/filediver/extractor/entity"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/entity"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

var seed = rand.Uint32()

type textureData struct {
	Slot           string
	Name           stingray.Hash
	Targets        []uint32
	Bounds         image.Rectangle
	Depth          int32
	InternalFormat int32
	Format         uint32
	Swizzles       []uint32
	Type           uint32
	Buffer         uint32
	Data           []uint8
}

func getTarget(slot string) uint32 {
	switch slot {
	case "customization_camo_tiler_array", "pattern_masks_array", "id_masks_array", "customization_material_detail_tiler_array", "surface_data_array":
		return gl.TEXTURE_2D_ARRAY
	}
	return gl.TEXTURE_2D
}

type MaterialManager struct {
	textureCache       *glutils.TextureCache
	textureSweepTicker *time.Ticker
	doSweep            bool
	stopTextureSweep   func()

	textureData      map[stingray.Hash]textureData
	textureDataMutex sync.Mutex
	textureWaitGroup sync.WaitGroup

	entityInfo             *entity.Entity
	assetGradingLut        []extr_entity.Vec4F16
	assetGradingLutTexture uint32
	assetGradingLutBuffer  uint32

	iblLutTexture uint32

	detailerLoadState   DetailerLoadState
	detailerTextureData textureData

	lookupThinHash func(stingray.ThinHash) string
	lookupHash     func(stingray.Hash) string
	getResource    GetResourceFunc
	getOverride    func(stingray.FileID) stingray.FileID
}

func (ml *MaterialManager) Sweep() {
	if ml.doSweep {
		ml.textureCache.Sweep()
		ml.doSweep = false
	}
}

func (ml *MaterialManager) Delete() {
	ml.DeleteAssetGradingLut()
	ml.ReleaseMaterialDetailer()
	ml.ReleaseIBLLookup()
	ml.textureCache.DeleteAll()
	ml.stopTextureSweep()
}

func (ml *MaterialManager) DeleteMaterial(material Material) {
	if material != nil {
		material.Delete(ml.textureCache)
	}
}

func (ml *MaterialManager) ReleaseMaterial(material Material) {
	if material != nil {
		material.ReleaseTextures(ml.textureCache)
	}
}

func (ml *MaterialManager) DeleteAssetGradingLut() {
	if ml.assetGradingLut != nil {
		gl.DeleteBuffers(1, &ml.assetGradingLutBuffer)
		ml.assetGradingLutBuffer = 0
		gl.DeleteTextures(1, &ml.assetGradingLutTexture)
		ml.assetGradingLutTexture = 0
		ml.assetGradingLut = nil
	}
}

func (ml *MaterialManager) ReleaseMaterialDetailer() {
	materialDetailerHash := stingray.Sum("content/art_shared/textures/customization/material_library/detail_tilers/customization_detail_tiler_array")
	materialDetailerID := stingray.NewFileID(materialDetailerHash, stingray.Sum("texture"))
	ml.textureCache.Release(materialDetailerID, gl.TEXTURE_2D_ARRAY)
}

func (ml *MaterialManager) SetEntity(entityInfo *entity.Entity) {
	ml.entityInfo = entityInfo
}

func (ml *MaterialManager) ReleaseIBLLookup() {
	iblID := stingray.NewFileID(stingray.Sum("generated_ibl_lut"), stingray.Sum("texture"))
	ml.textureCache.Release(iblID, gl.TEXTURE_2D)
}

func (ml *MaterialManager) GenerateIBLLookup(size int32) {
	var fbo, rbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.GenRenderbuffers(1, &rbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, size, size)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)
	gl.Viewport(0, 0, size, size)

	var iblPlaneVAO, iblPlaneVBO, iblPlaneIBO uint32
	iblID := stingray.NewFileID(stingray.Sum("generated_ibl_lut"), stingray.Sum("texture"))
	iblLut, _ := ml.textureCache.Acquire(iblID, gl.TEXTURE_2D)

	gl.BindTexture(gl.TEXTURE_2D, iblLut)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RG16F, size, size, 0, gl.RG, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, iblLut, 0)

	gl.GenVertexArrays(1, &iblPlaneVAO)
	gl.GenBuffers(1, &iblPlaneVBO)
	gl.GenBuffers(1, &iblPlaneIBO)

	vertices := []float32{
		-1.0, -1.0, 0.0, 0.0, 0.0,
		1.0, -1.0, 0.0, 1.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 1.0,
		-1.0, 1.0, 0.0, 0.0, 1.0,
	}
	indices := []uint8{
		0, 1, 2,
		0, 2, 3,
	}
	gl.BindVertexArray(iblPlaneVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, iblPlaneVBO)
	gl.BufferData(gl.ARRAY_BUFFER, binary.Size(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, iblPlaneIBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, binary.Size(indices), gl.Ptr(indices), gl.STATIC_DRAW)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 5*4, 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, 3*4)
	gl.EnableVertexAttribArray(1)

	iblShader, err := glutils.CreateProgramFromSources(
		unitPreviewShaderCode,
		"shaders/integrate_ibl.vert",
		"shaders/integrate_ibl.frag",
	)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(iblShader)

	gl.ClearDepth(1.0)
	gl.Disable(gl.DEPTH_TEST)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_BYTE, nil)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	gl.DeleteProgram(iblShader)
	gl.DeleteBuffers(1, &iblPlaneIBO)
	gl.DeleteBuffers(1, &iblPlaneVBO)
	gl.DeleteVertexArrays(1, &iblPlaneVAO)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)
	gl.DeleteRenderbuffers(1, &rbo)
	gl.DeleteFramebuffers(1, &fbo)
}

func (ml *MaterialManager) ClearTextureData() {
	ml.textureData = make(map[stingray.Hash]textureData)
}

func (ml *MaterialManager) loadDDS(fileName stingray.Hash) (*dds.DDS, error) {
	file := stingray.FileID{Name: fileName, Type: stingray.Sum("texture")}
	var texMain, texStream, texGPU []byte
	var err error
	var exists bool
	if texMain, exists, err = ml.getResource(file, stingray.DataMain); err != nil {
		return nil, fmt.Errorf("load texture %v.texture: %v", fileName.String(), err)
	} else if !exists {
		return nil, fmt.Errorf("%v.texture does not exist", fileName.String())
	}
	texStream, _, _ = ml.getResource(file, stingray.DataStream)
	texGPU, _, _ = ml.getResource(file, stingray.DataGPU)
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

func (ml *MaterialManager) updateAssetGradingLut() error {
	if ml.assetGradingLut == nil {
		ml.assetGradingLut = extr_entity.CreateIdentityColorGradingLut()
		gl.GenTextures(1, &ml.assetGradingLutTexture)
		gl.GenBuffers(1, &ml.assetGradingLutBuffer)
	}
	if ml.entityInfo != nil {
		entityAssetGradingLut, err := extr_entity.CreateColorGradingLut(ml.entityInfo)
		if err == nil {
			ml.assetGradingLut = entityAssetGradingLut
		}
	}

	assetGradingData, err := binary.Append(nil, binary.LittleEndian, ml.assetGradingLut)
	if err != nil {
		return err
	}

	gl.BindBuffer(gl.TEXTURE_BUFFER, ml.assetGradingLutBuffer)
	gl.BufferData(gl.TEXTURE_BUFFER, len(assetGradingData), gl.Ptr(assetGradingData), gl.STATIC_DRAW)
	gl.BindBuffer(gl.TEXTURE_BUFFER, 0)

	gl.BindTexture(gl.TEXTURE_BUFFER, ml.assetGradingLutTexture)
	gl.TexBuffer(gl.TEXTURE_BUFFER, gl.RGBA16F, ml.assetGradingLutBuffer)
	gl.BindTexture(gl.TEXTURE_BUFFER, 0)

	return nil
}

type DetailerLoadState uint8

const (
	DetailerNotLoaded DetailerLoadState = iota
	DetailerLoading
	DetailerLoaded
	DetailerUploaded
)

func (ml *MaterialManager) loadMaterialDetailer() {
	slot := "customization_material_detail_tiler_array"
	materialDetailerHash := stingray.Sum("content/art_shared/textures/customization/material_library/detail_tilers/customization_detail_tiler_array")
	materialDetailerID := stingray.NewFileID(materialDetailerHash, stingray.Sum("texture"))
	var target uint32 = gl.TEXTURE_2D_ARRAY

	if ml.textureCache.Contains(materialDetailerID, target) {
		// Already loaded, so we don't need to reload the data
		return
	}

	if ml.textureData == nil {
		ml.textureData = make(map[stingray.Hash]textureData)
	}
	ml.textureData[materialDetailerHash] = textureData{
		Slot:    slot,
		Name:    materialDetailerHash,
		Targets: []uint32{target},
	}

	ml.detailerLoadState = DetailerLoading

	textureLoader := ml.getTextureLoaderFunc(materialDetailerHash)
	go func() {
		textureLoader()
		ml.detailerTextureData = ml.textureData[materialDetailerHash]
		ml.detailerLoadState = DetailerLoaded
	}()
}

func (ml *MaterialManager) uploadMaterialDetailer(data textureData) {
	detailerId := stingray.NewFileID(data.Name, stingray.Sum("texture"))
	textureId, created := ml.textureCache.Acquire(detailerId, data.Targets[0])
	texture := Texture{
		id:       textureId,
		override: detailerId,
		name:     data.Name,
		target:   data.Targets[0],
		loaded:   false,
		filter:   gl.LINEAR,
	}

	if !created {
		return
	}
	setupTexture(texture.id, texture.target, int32(texture.filter))
	err := uploadStingrayTexture(texture.id, texture.target, data)
	if err != nil {
		ml.releaseMaterialDetailer()
		ml.detailerLoadState = DetailerNotLoaded
	} else {
		ml.detailerLoadState = DetailerUploaded
	}
	// We can discard the texture data after uploading, so the garbage collector will reclaim the ~110MB of memory
	ml.detailerTextureData = textureData{}
}

func (ml *MaterialManager) releaseMaterialDetailer() {
	materialDetailerHash := stingray.Sum("content/art_shared/textures/customization/material_library/detail_tilers/customization_detail_tiler_array")
	materialDetailerID := stingray.NewFileID(materialDetailerHash, stingray.Sum("texture"))
	ml.textureCache.Release(materialDetailerID, gl.TEXTURE_2D_ARRAY)
}

func (ml *MaterialManager) getTextureLoaderFunc(nameHash stingray.Hash) func() {
	return func() {
		ml.textureDataMutex.Lock()
		data := ml.textureData[nameHash]
		ml.textureDataMutex.Unlock()
		data.Bounds = image.Rect(0, 0, 1, 1)
		data.Data = make([]uint8, 4)
		data.Depth = 1
		ml.textureDataMutex.Lock()
		ml.textureData[nameHash] = data
		ml.textureDataMutex.Unlock()

		ddsImage, err := ml.loadDDS(nameHash)
		if err != nil {
			return
		}
		data.Bounds = ddsImage.Bounds()
		data.Depth = int32(len(ddsImage.Images))

		for idx := range ddsImage.Images {
			if ddsImage.Info.DXT10Header != nil && ddsImage.Info.DXT10Header.DXGIFormat == dds.DXGIFormatR16G16B16A16Float {
				data.Format = gl.RGBA
				data.InternalFormat = gl.RGBA16F
				data.Type = gl.HALF_FLOAT
				data.Data = ddsImage.Images[0].MipMaps[0].Raw
			} else if ddsImage.Info.DXT10Header != nil && ddsImage.Info.DXT10Header.DXGIFormat == dds.DXGIFormatR32G32B32A32Float {
				data.Format = gl.RGBA
				data.InternalFormat = gl.RGBA32F
				data.Type = gl.FLOAT
				data.Data = ddsImage.Images[0].MipMaps[0].Raw
			} else if ddsImage.Info.DXT10Header != nil && ddsImage.Info.DXT10Header.DXGIFormat == dds.DXGIFormatR16G16Float {
				data.Format = gl.RGBA
				data.InternalFormat = gl.RG16F
				data.Type = gl.HALF_FLOAT
				data.Data = ddsImage.Images[0].MipMaps[0].Raw
			} else {
				switch img := ddsImage.Images[idx].Image.(type) {
				case *image.NRGBA:
					data.Format = gl.RGBA
					data.InternalFormat = gl.RGBA
					data.Type = gl.UNSIGNED_BYTE
					if idx == 0 {
						data.Data = make([]uint8, 0, ddsImage.Bounds().Dx()*ddsImage.Bounds().Dy()*int(data.Depth)*4)
					}
					data.Data = append(data.Data, img.Pix...)
				case *image.Gray:
					data.Format = gl.RED
					data.InternalFormat = gl.RED
					data.Type = gl.UNSIGNED_BYTE
					data.Swizzles = []uint32{gl.RED, gl.RED, gl.RED, gl.RED}
					if idx == 0 {
						data.Data = make([]uint8, 0, ddsImage.Bounds().Dx()*ddsImage.Bounds().Dy()*int(data.Depth))
					}
					data.Data = append(data.Data, img.Pix...)
				default:
					fmt.Printf("[error] Failed to convert image %v of %v\n", idx, data.Slot)
					return
				}
			}
		}
		ml.textureDataMutex.Lock()
		ml.textureData[nameHash] = data
		ml.textureDataMutex.Unlock()
	}
}

func (ml *MaterialManager) loadMaterialTextureAsync(slotHash stingray.ThinHash, nameHash stingray.Hash) {
	slot := ml.lookupThinHash(slotHash)
	target := getTarget(slot)

	ml.textureDataMutex.Lock()
	existing, contains := ml.textureData[nameHash]
	ml.textureDataMutex.Unlock()

	overrideID := ml.getOverride(stingray.NewFileID(nameHash, stingray.Sum("texture")))

	if contains && slices.Contains(existing.Targets, target) {
		// already loading, so we don't need to create another load request
		fmt.Printf("texture data already loading for %v.texture with override %v.texture\n", ml.lookupHash(nameHash), ml.lookupHash(overrideID.Name))
		return
	}
	targets := make([]uint32, 0)
	data := textureData{}
	if contains {
		targets = existing.Targets
		data = existing
	}
	targets = append(targets, target)
	data.Slot = slot
	data.Name = nameHash
	data.Targets = targets

	ml.textureDataMutex.Lock()
	ml.textureData[nameHash] = data
	ml.textureDataMutex.Unlock()

	if ml.textureCache.Contains(overrideID, target) {
		// Already loaded, so we don't need to reload the data
		fmt.Printf("texture data already cached for %v.texture\n", ml.lookupHash(overrideID.Name))
		return
	}

	ml.textureWaitGroup.Go(ml.getTextureLoaderFunc(nameHash))
}

func (ml *MaterialManager) loadMaterialTextureSync(slotHash stingray.ThinHash, nameHash stingray.Hash) {
	slot := ml.lookupThinHash(slotHash)
	target := getTarget(slot)

	ml.textureDataMutex.Lock()
	existing, contains := ml.textureData[nameHash]
	ml.textureDataMutex.Unlock()

	overrideID := ml.getOverride(stingray.NewFileID(nameHash, stingray.Sum("texture")))

	if contains && slices.Contains(existing.Targets, target) {
		// already loading, so we don't need to create another load request
		fmt.Printf("texture data already loading for %v.texture with override %v.texture\n", ml.lookupHash(nameHash), ml.lookupHash(overrideID.Name))
		return
	}
	targets := make([]uint32, 0)
	data := textureData{}
	if contains {
		targets = existing.Targets
		data = existing
	}
	targets = append(targets, target)
	data.Slot = slot
	data.Name = nameHash
	data.Targets = targets

	ml.textureDataMutex.Lock()
	ml.textureData[nameHash] = data
	ml.textureDataMutex.Unlock()

	if ml.textureCache.Contains(overrideID, target) {
		// Already loaded, so we don't need to reload the data
		fmt.Printf("texture data already cached for %v.texture\n", ml.lookupHash(overrideID.Name))
		return
	}

	ml.getTextureLoaderFunc(nameHash)()
}

func (ml *MaterialManager) loadMaterialTexturesAsync(mat *material.Material, textureNames []string) {
	if mat == nil || textureNames == nil {
		return
	}
	if ml.textureData == nil {
		ml.textureData = make(map[stingray.Hash]textureData)
	}
	slotHashes := slices.SortedFunc(maps.Keys(mat.Textures), stingray.ThinHash.Cmp)
	for _, slotHash := range slotHashes {
		slot := ml.lookupThinHash(slotHash)
		if !slices.Contains(textureNames, slot) {
			continue
		}
		nameHash := mat.Textures[slotHash]
		ml.loadMaterialTextureAsync(slotHash, nameHash)
	}
}

func (ml *MaterialManager) FinalizeMaterialTextures(material Material) error {
	for idx, texture := range material.Textures() {
		if texture.override == texture.previousOverride {
			continue
		}
		var id uint32
		uniform := material.GetUniform(texture.slot)
		if texture.name.Value == 0 {
			if texture.defaultColor == nil {
				texture.defaultColor = make([]byte, 4)
			}
			gl.GenTextures(1, &material.Textures()[idx].id)
			id = material.Textures()[idx].id
			setupTexture(id, texture.target, int32(texture.filter))
			gl.BindTexture(texture.target, id)
			gl.TexImage2D(texture.target, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(texture.defaultColor))
			gl.BindTexture(texture.target, 0)
			gl.ProgramUniform1i(material.Program(), uniform, int32(idx))
			continue
		} else if !texture.loaded {
			id, created := ml.textureCache.Acquire(texture.override, texture.target)
			material.Textures()[idx].id = id
			material.Textures()[idx].created = created
			material.Textures()[idx].previousOverride = texture.override
			var err error
			if material.Textures()[idx].created {
				setupTexture(id, texture.target, int32(texture.filter))
				err = uploadStingrayTexture(id, texture.target, ml.textureData[texture.name])
			}
			if err != nil {
				ml.textureCache.Delete(texture.override, texture.target)
				return err
			}
		}
		gl.ProgramUniform1i(material.Program(), uniform, int32(idx))
	}
	return nil
}

func (ml *MaterialManager) Wait() {
	ml.textureWaitGroup.Wait()
}

func (ml *MaterialManager) setupMaterialCommon(textureNames []string, previewMaterial Material, mat *material.Material) {
	slot := "ibl_brdf_lut"
	iblBRDFHash := stingray.Sum("generated_ibl_lut")
	mat.Textures[stingray.Sum(slot).Thin()] = iblBRDFHash

	for _, slot := range textureNames {
		slotHash := stingray.Sum(slot).Thin()
		if _, contains := mat.Textures[slotHash]; !contains {
			mat.Textures[slotHash] = stingray.Sum("content/art_shared/textures/black_all_channels_dummy")
		}
	}

	ml.loadMaterialTexturesAsync(mat, textureNames)

	sortedSlots := slices.SortedFunc(maps.Keys(mat.Textures), stingray.ThinHash.Cmp)
	for _, slotHash := range sortedSlots {
		slot := ml.lookupThinHash(slotHash)
		if !slices.Contains(textureNames, slot) {
			continue
		}
		textureHash := mat.Textures[slotHash]
		target := getTarget(slot)
		texture := NewTexture(
			textureHash,
			ml.getOverride(stingray.NewFileID(textureHash, stingray.Sum("texture"))),
			target,
			gl.LINEAR,
			slot,
			false,
			false,
		)

		previewMaterial.AddTexture(texture)
	}

	previewMaterial.InitSettings(mat.Settings, ml.lookupThinHash)
}

func (ml *MaterialManager) ReloadTextures(material Material, mat *material.Material) {
	for texIdx := range material.Textures() {
		texture := material.Textures()[texIdx]
		textureID := stingray.NewFileID(texture.name, stingray.Sum("texture"))
		textureOverrideID := ml.getOverride(textureID)
		if texture.override == textureOverrideID {
			continue
		}

		material.Textures()[texIdx].release(ml.textureCache)
		material.Textures()[texIdx].override = textureOverrideID
		nameHash := mat.Textures[stingray.Sum(texture.slot).Thin()]

		ml.loadMaterialTextureAsync(stingray.Sum(texture.slot).Thin(), nameHash)
	}
}

func (ml *MaterialManager) LoadMaterial(id stingray.FileID, name string, shown bool, mat *material.Material, existing Material, objectId stingray.FileID) Material {
	if existing != nil {
		existing.ReleaseTextures(ml.textureCache)
	}

	var err error
	defer func() {
		if err != nil {
			fmt.Printf("[error] material generation: %v\n", err)
		}
	}()

	if isLutMaterial(mat) {
		toReturn, ok := existing.(*LutMaterial)
		if existing == nil || !ok {
			toReturn = NewLutMaterial(id, name, shown)
			err = toReturn.Generate()
		}

		ml.updateLutMaterial(toReturn, mat)
		return toReturn
	}

	if isStandardMaterial(mat) {
		toReturn, ok := existing.(*StandardMaterial)
		if existing == nil || !ok {
			toReturn = NewStandardMaterial(id, name, shown)
			err = toReturn.Generate()
		}

		ml.updateStandardMaterial(toReturn, mat)
		return toReturn
	}

	if isBuildingMaterial(mat) {
		toReturn, ok := existing.(*BuildingMaterial)
		if existing == nil || !ok {
			toReturn = NewBuildingMaterial(id, name, shown)
			err = toReturn.Generate()
		}

		ml.updateBuildingMaterial(toReturn, mat)
		return toReturn
	}

	if isEmissiveMaterial(mat) {
		toReturn, ok := existing.(*EmissiveMaterial)
		if existing == nil || !ok {
			toReturn = NewEmissiveMaterial(id, name, shown)
			err = toReturn.Generate()
		}

		ml.updateEmissiveMaterial(toReturn, mat)
		return toReturn
	}

	if isSignMaterial(mat) {
		toReturn, ok := existing.(*SignMaterial)
		if existing == nil || !ok {
			toReturn = NewSignMaterial(id, name, shown)
			err = toReturn.Generate()
		}

		ml.updateSignMaterial(toReturn, mat)
		return toReturn
	}

	if isSpeedtreeMaterial(mat) {
		if objectId.Type == stingray.Sum("speedtree") {
			toReturn, ok := existing.(*SpeedtreeMaterial)
			if existing == nil || !ok {
				toReturn = NewSpeedtreeMaterial(id, name, shown)
				err = toReturn.Generate()
			}

			ml.updateSpeedtreeMaterial(toReturn, mat)
			return toReturn
		} else {
			// Apparently speedtree materials aren't unique to speedtree file types
			toReturn, ok := existing.(*SpeedtreeUnitMaterial)
			if existing == nil || !ok {
				toReturn = NewSpeedtreeUnitMaterial(id, name, shown)
				err = toReturn.Generate()
			}

			ml.updateSpeedtreeUnitMaterial(toReturn, mat)
			return toReturn
		}
	}

	toReturn, ok := existing.(*FallbackMaterial)
	if existing == nil || !ok {
		toReturn = NewFallbackMaterial(id, name, shown)
		err = toReturn.Generate()
	}

	ml.updateFallbackMaterial(toReturn, mat)
	return toReturn
}

func NewMaterialManager(getResource GetResourceFunc, getOverride func(stingray.FileID) stingray.FileID, lookupThinHash func(stingray.ThinHash) string, lookupHash func(stingray.Hash) string) *MaterialManager {
	ml := &MaterialManager{}

	ml.getResource = getResource
	ml.getOverride = getOverride
	ml.lookupHash = lookupHash
	ml.lookupThinHash = lookupThinHash

	// Keep textures for a minute of disuse
	duration, _ := time.ParseDuration("1m")
	ml.textureCache = glutils.NewTextureCache(duration)

	sweepInterval, _ := time.ParseDuration("30s")
	ml.textureSweepTicker = time.NewTicker(sweepInterval)

	done := make(chan bool)
	ml.stopTextureSweep = func() {
		done <- true
	}
	// schedule a texture sweep every time the ticker ticks
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ml.textureSweepTicker.C:
				ml.doSweep = true
			}
		}
	}()

	ml.detailerLoadState = DetailerNotLoaded
	ml.loadMaterialDetailer()

	return ml
}
