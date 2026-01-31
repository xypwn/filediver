package previews

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/glutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

type MaterialPreviewState struct {
	textures         map[string]*DDSPreviewState // nil preview state represents null texture
	textureKeys      [][2]string
	activeTexture    int
	settings         map[string][]float32
	settingKeys      []string
	settingsVisible  bool
	baseMaterial     stingray.Hash
	baseMaterialName string

	meshPreviewBuffer *PreviewMeshBuffer
	fb                *widgets.GLViewState
	program           uint32

	offset          imgui.Vec2
	zoom            float32
	linearFiltering bool
	ignoreAlpha     bool
	err             error
}

func NewMaterialPreview() *MaterialPreviewState {
	return &MaterialPreviewState{
		textures:      make(map[string]*DDSPreviewState),
		activeTexture: -1,
		settings:      make(map[string][]float32),
	}
}

func (pv *MaterialPreviewState) Delete() {
	for _, state := range pv.textures {
		if state != nil {
			state.Delete()
		}
	}
	pv.fb.Delete()
	gl.DeleteProgram(pv.program)
	if pv.meshPreviewBuffer != nil {
		pv.meshPreviewBuffer.DeleteObjects()
	}
}

func (pv *MaterialPreviewState) LoadMaterial(mat *material.Material, getResource GetResourceFunc, hashes map[stingray.Hash]string, thinhashes map[stingray.ThinHash]string) error {
	if mat == nil {
		return fmt.Errorf("attempted to load nil material")
	}
	var err error
	pv.Delete()
	clear(pv.textures)
	pv.textureKeys = nil
	pv.activeTexture = -1
	clear(pv.settings)
	pv.settingKeys = nil
	pv.settingsVisible = true
	pv.baseMaterial = stingray.Hash{}
	pv.meshPreviewBuffer, err = getPlane()
	if err != nil {
		return err
	}

	pv.fb, err = widgets.NewGLView()
	if err != nil {
		return err
	}

	pv.program, err = glutils.CreateProgramFromSources(PreviewShaderCode,
		"shaders/3.glsl.vert",
		"shaders/3.glsl.frag",
	)
	if err != nil {
		return err
	}

	sortedTextureUsages := slices.SortedFunc(maps.Keys(mat.Textures), stingray.ThinHash.Cmp)
	for _, key := range sortedTextureUsages {
		path := mat.Textures[key]
		var imageName, pathName string
		var ok bool
		imageName, ok = thinhashes[key]
		if !ok {
			imageName = "Unknown texture usage: " + key.String()
		}
		if pathName, ok = hashes[path]; !ok {
			pathName = path.String()
		}

		textureKey := [2]string{imageName, pathName}

		if path.Value == 0 {
			// Zero texture
			pv.textureKeys = append(pv.textureKeys, textureKey)
			pv.textures[imageName] = nil
			continue
		}

		dataMain, fileExists, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum("texture")}, stingray.DataMain)
		if err != nil {
			return fmt.Errorf("material texture %v: %w", pathName, err)
		}
		if !fileExists {
			return fmt.Errorf("material texture %v: referenced texture does not exist", pathName)
		}

		dataGPU, _, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum("texture")}, stingray.DataGPU)
		if err != nil {
			return fmt.Errorf("material texture %v: %w", pathName, err)
		}

		dataStream, _, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum("texture")}, stingray.DataStream)
		if err != nil && !errors.Is(err, stingray.ErrFileDataTypeNotExist) {
			return fmt.Errorf("material texture %v: %w", pathName, err)
		}

		r := io.MultiReader(
			bytes.NewReader(dataMain),
			bytes.NewReader(dataStream),
			bytes.NewReader(dataGPU),
		)
		if _, err := texture.DecodeInfo(r); err != nil {
			return fmt.Errorf("material texture %v: loading stingray DDS info: %w", pathName, err)
		}
		img, err := dds.Decode(r, false)
		if err != nil {
			return fmt.Errorf("material texture %v: loading DDS image: %w", pathName, err)
		}

		pv.textureKeys = append(pv.textureKeys, textureKey)
		pv.textures[imageName] = NewDDSPreview()
		pv.textures[imageName].LoadImage(img)
	}

	if len(pv.textureKeys) > 0 && pv.activeTexture == -1 {
		pv.activeTexture = 0
	}

	for key, value := range mat.Settings {
		keyName, ok := thinhashes[key]
		if !ok {
			keyName = "unknown setting: " + key.String()
		}
		pv.settings[keyName] = value
		pv.settingKeys = append(pv.settingKeys, keyName)
	}
	slices.Sort(pv.settingKeys)

	pv.baseMaterial = mat.BaseMaterial
	if bmName, ok := hashes[mat.BaseMaterial]; ok {
		pv.baseMaterialName = bmName
	} else {
		pv.baseMaterialName = mat.BaseMaterial.String()
	}

	return nil
}

func MaterialPreview(name string, pv *MaterialPreviewState) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	if pv.err != nil {
		imutils.TextError(pv.err)
		return
	}

	var ddsPv *DDSPreviewState
	var textureUsage, texturePath string
	currTexture := pv.activeTexture
	if pv.activeTexture >= 0 && pv.activeTexture < len(pv.textureKeys) {
		textureUsage = pv.textureKeys[pv.activeTexture][0]
		texturePath = pv.textureKeys[pv.activeTexture][1]
		ddsPv = pv.textures[textureUsage]
	}
	var infoB strings.Builder
	if currTexture != -1 {
		fmt.Fprintf(&infoB, "Usage=%v (%v/%v)\n", textureUsage, pv.activeTexture+1, len(pv.textures))
		if ddsPv != nil {
			fmt.Fprintf(&infoB, "Size=(%v,%v)\nFormat=%v\nPath=%v\n", ddsPv.imageSize.X, ddsPv.imageSize.Y, ddsPv.ddsInfo.DXT10Header.DXGIFormat, texturePath)
			ddsPv.offset = pv.offset
			ddsPv.zoom = pv.zoom
			if ddsPv.linearFiltering != pv.linearFiltering {
				filter := int32(gl.NEAREST)
				if pv.linearFiltering {
					filter = gl.LINEAR
				}
				gl.BindTexture(gl.TEXTURE_2D, ddsPv.textureID)
				gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filter)
				gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filter)
				gl.BindTexture(gl.TEXTURE_2D, 0)
			}
			ddsPv.linearFiltering = pv.linearFiltering
			if ddsPv.imageHasAlpha {
				ddsPv.ignoreAlpha = pv.ignoreAlpha
			}
		} else {
			fmt.Fprintf(&infoB, "Usage=N/A (0/0)\nSize=N/A\nFormat=N/A\nPath=N/A\n")
		}
	} else {
		fmt.Fprintf(&infoB, "Usage=N/A (0/0)\nSize=N/A\nFormat=N/A\nPath=N/A\n")
	}
	fmt.Fprintf(&infoB, "Num Settings=%v\n", len(pv.settings))
	imgui.TextUnformatted(infoB.String())

	{
		size := imgui.ContentRegionAvail()
		size.Y -= imutils.ComboHeight()
		size.Y -= imgui.TextLineHeightWithSpacing() // base material
		imgui.SetNextWindowSize(size)
	}

	cycleTexDelta := 0

	if imgui.Shortcut(imgui.KeyChord(imgui.KeyLeftArrow)) {
		cycleTexDelta--
	}
	if imgui.Shortcut(imgui.KeyChord(imgui.KeyRightArrow)) {
		cycleTexDelta++
	}

	if imgui.BeginChildStrV("##DDS image preview", imgui.NewVec2(0, 0), 0, 0) {
		pos := imgui.CursorScreenPos()
		area := imgui.ContentRegionAvail()
		BuildImagePreviewArea(ddsPv, pos, area)
		// Image Preview Area can modify the image's offset/zoom, but we'd like those to be applied
		// for the entire material, so copy them out
		if ddsPv != nil {
			pv.offset = ddsPv.offset
			pv.zoom = ddsPv.zoom
		}
		style := imgui.CurrentStyle()
		imgui.BeginDisabledV(len(pv.textureKeys) < 2)
		imgui.PushItemFlag(imgui.ItemFlagsNoNav, true)
		imgui.SetCursorScreenPos(imgui.NewVec2(pos.X+style.ItemSpacing().X, pos.Y+area.Y/2))
		if imgui.Button(fnt.I("Arrow_left")) {
			cycleTexDelta--
		}
		imgui.SetCursorScreenPos(imgui.NewVec2(pos.X+area.X-imgui.FrameHeightWithSpacing()-style.ItemSpacing().X, pos.Y+area.Y/2))
		if imgui.Button(fnt.I("Arrow_right")) {
			cycleTexDelta++
		}
		imgui.PopItemFlag()
		imgui.EndDisabled()

		if ddsPv == nil {
			text := "<nil texture>"
			textSize := imgui.CalcTextSize(text)
			textPos := pos.Add(area.Div(2)).Sub(textSize.Div(2))
			imgui.SetCursorScreenPos(textPos)
			imgui.TextUnformatted(text)
		} else if pv.activeTexture != currTexture {
			// Swapping textures, make sure they use the correct ignore alpha & linear filtering
			// values
			textureUsage = pv.textureKeys[pv.activeTexture][0]
			nextDdsPv := pv.textures[textureUsage]
			filter := int32(gl.NEAREST)
			swizzleA := int32(gl.ALPHA)
			if pv.linearFiltering {
				filter = gl.LINEAR
			}
			if pv.ignoreAlpha {
				swizzleA = gl.ONE
			}
			gl.BindTexture(gl.TEXTURE_2D, nextDdsPv.textureID)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filter)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filter)
			if nextDdsPv.imageHasAlpha {
				gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, swizzleA)
			}
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}
	}
	imgui.EndChild()

	if len(pv.textureKeys) > 0 {
		mod := func(a, b int) int { // python-like modulo
			return (a%b + b) % b
		}
		pv.activeTexture = mod(pv.activeTexture+cycleTexDelta, len(pv.textureKeys))
	}

	if imgui.Button(fnt.I("Home")) {
		pv.offset = imgui.NewVec2(0, 0)
		pv.zoom = 1
	}
	imgui.SetItemTooltip("Reset view")
	imgui.SameLine()
	imgui.BeginDisabledV(len(pv.settings) == 0)
	if imgui.Button(fnt.I("Display_settings")) {
		pv.settingsVisible = true
	}
	imgui.EndDisabled()
	imgui.SetItemTooltip("Show material settings window.")
	imgui.SameLine()
	imgui.BeginDisabledV(ddsPv == nil)
	if imgui.Checkbox("Linear filtering", &pv.linearFiltering) {
		filter := int32(gl.NEAREST)
		if pv.linearFiltering {
			filter = gl.LINEAR
		}
		gl.BindTexture(gl.TEXTURE_2D, ddsPv.textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filter)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filter)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	imgui.EndDisabled()
	imgui.SetItemTooltip("Linear filtering \"blurs\" pixels when zooming in. Disable to view individual pixels more clearly.")
	imgui.SameLine()
	imgui.BeginDisabledV(ddsPv == nil || !ddsPv.imageHasAlpha)
	if imgui.Checkbox("Ignore alpha", &pv.ignoreAlpha) && ddsPv != nil {
		swizzleA := int32(gl.ALPHA)
		if pv.ignoreAlpha {
			swizzleA = gl.ONE
		}
		gl.BindTexture(gl.TEXTURE_2D, ddsPv.textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, swizzleA)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	imgui.EndDisabled()
	if ddsPv == nil || !ddsPv.imageHasAlpha {
		imgui.SetItemTooltip("This image doesn't use an alpha component.")
	} else {
		imgui.SetItemTooltip("Ignore alpha component, making RGB components always fully visible.")
	}

	imgui.TextUnformatted("Base material:")
	imgui.SameLine()
	if (pv.baseMaterial == stingray.Hash{}) {
		imgui.TextUnformatted("none")
	} else {
		fileID := stingray.FileID{
			Name: pv.baseMaterial,
			Type: stingray.Sum("material"),
		}
		widgets.GamefileLinkTextF(fileID, "%v", pv.baseMaterialName)
	}

	if len(pv.settings) > 0 && pv.settingsVisible {
		if imgui.BeginV(fnt.I("Display_settings")+" Material Settings", &pv.settingsVisible, imgui.WindowFlagsNone) {
			const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY | imgui.TableFlagsRowBg
			if imgui.BeginTableV("##Material Settings", 2, tableFlags, imgui.NewVec2(0, 0), 0) {
				imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 1, 0)
				imgui.TableSetupColumnV("Value", imgui.TableColumnFlagsWidthStretch, 2, 0)
				imgui.TableSetupScrollFreeze(0, 1)
				imgui.TableHeadersRow()

				for _, id := range pv.settingKeys {
					imgui.PushIDStr(id)

					imgui.TableNextColumn()
					imutils.CopyableTextf("%v", id)

					imgui.TableNextColumn()
					settingValue := pv.settings[id]
					formatted := make([]string, len(settingValue))
					for i := range settingValue {
						formatted[i] = fmt.Sprintf("%.3f", settingValue[i])
					}
					settingString := strings.Join(formatted, ", ")
					if len(settingValue) > 1 {
						settingString = "(" + settingString + ")"
					}
					imgui.TextUnformatted(settingString)

					imgui.PopID()
				}
				imgui.EndTable()
			}
		}
		imgui.End()
	}
}
