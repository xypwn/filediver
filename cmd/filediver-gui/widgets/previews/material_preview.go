package previews

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/dds"
	extr_material "github.com/xypwn/filediver/extractor/material"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

type MaterialPreviewState struct {
	textures      map[string]*DDSPreviewState
	textureKeys   [][2]string
	activeTexture int
	settings      map[string][]float32
	settingKeys   []string

	offset          imgui.Vec2
	zoom            float32
	linearFiltering bool
	err             error
}

func NewMaterialPreview() *MaterialPreviewState {
	return &MaterialPreviewState{
		textures:      make(map[string]*DDSPreviewState),
		textureKeys:   make([][2]string, 0),
		activeTexture: -1,
		settings:      make(map[string][]float32),
	}
}

func (pv *MaterialPreviewState) Delete() {
	for key := range pv.textures {
		pv.textures[key].Delete()
	}
}

func (pv *MaterialPreviewState) LoadMaterial(mat *material.Material, getResource GetResourceFunc, hashes map[stingray.Hash]string, thinhashes map[stingray.ThinHash]string) error {
	if mat == nil {
		return fmt.Errorf("attempted to load nil material")
	}
	if len(pv.textures) > 0 {
		pv.Delete()
		pv.textures = make(map[string]*DDSPreviewState)
		pv.textureKeys = make([][2]string, 0)
		pv.activeTexture = -1
	}
	if len(pv.settings) > 0 {
		pv.settingKeys = make([]string, 0)
		pv.settings = make(map[string][]float32)
	}
	for key, path := range mat.Textures {
		var imageName, pathName string
		var ok bool
		usage := extr_material.TextureUsage(key.Value)
		imageName = usage.String()
		if pathName, ok = hashes[path]; !ok {
			pathName = path.String()
		}
		dataMain, _, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum64([]byte("texture"))}, stingray.DataMain)
		if err != nil {
			return fmt.Errorf("material texture %v: loading main data: %w", pathName, err)
		}

		dataGPU, _, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum64([]byte("texture"))}, stingray.DataGPU)
		if err != nil {
			return fmt.Errorf("material texture %v: loading gpu data: %w", pathName, err)
		}

		dataStream, _, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum64([]byte("texture"))}, stingray.DataStream)
		if err != nil {
			dataStream = make([]byte, 0)
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

		pv.textureKeys = append(pv.textureKeys, [2]string{imageName, pathName})
		pv.textures[imageName] = NewDDSPreview()
		pv.textures[imageName].LoadImage(img)
		if pv.activeTexture == -1 {
			pv.activeTexture = 0
		}
	}

	unknownUsage := extr_material.SettingsUsage(0x0)
	unknownUsageStr := unknownUsage.String()

	for key, value := range mat.Settings {
		usage := extr_material.SettingsUsage(key.Value)
		keyName := usage.String()
		if keyName == unknownUsageStr {
			keyName += " (" + key.String() + ")"
		}
		pv.settings[keyName] = value
		pv.settingKeys = append(pv.settingKeys, keyName)
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
	if pv.activeTexture >= 0 && pv.activeTexture < len(pv.textureKeys) {
		textureUsage = pv.textureKeys[pv.activeTexture][0]
		texturePath = pv.textureKeys[pv.activeTexture][1]
		ddsPv = pv.textures[textureUsage]
	}
	if ddsPv != nil {
		imutils.Textf("Usage=%v (%v/%v)\nSize=(%v,%v)\nFormat=%v\nPath=%v\nNum Settings=%v", textureUsage, pv.activeTexture+1, len(pv.textures), ddsPv.imageSize.X, ddsPv.imageSize.Y, ddsPv.ddsInfo.DXT10Header.DXGIFormat, texturePath, len(pv.settings))
		ddsPv.offset = pv.offset
		ddsPv.zoom = pv.zoom
	} else {
		imutils.Textf("Usage=N/A (0/0)\nSize=N/A\nFormat=N/A\nPath=N/A\nNum Settings=%v", len(pv.settings))
	}

	{
		size := imgui.ContentRegionAvail()
		size.Y -= imutils.ComboHeight()
		if len(pv.settings) > 0 {
			size.Y -= imgui.TextLineHeightWithSpacing() * float32(min(len(pv.settings)+2, 6))
		}
		imgui.SetNextWindowSize(size)
	}

	if imgui.BeginChildStr("##DDS image preview") {
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
		imgui.SetCursorScreenPos(imgui.NewVec2(pos.X+style.ItemSpacing().X, pos.Y+area.Y/2))
		imgui.BeginDisabledV(len(pv.textureKeys) < 2)
		if imgui.Button(fnt.I("Arrow_left")) {
			pv.activeTexture = pv.activeTexture - 1
			if pv.activeTexture < 0 {
				pv.activeTexture = len(pv.textureKeys) - 1
			}
		}
		imgui.EndDisabled()
		imgui.SetCursorScreenPos(imgui.NewVec2(pos.X+area.X-imgui.FrameHeightWithSpacing()-style.ItemSpacing().X, pos.Y+area.Y/2))
		imgui.BeginDisabledV(len(pv.textureKeys) < 2)
		if imgui.Button(fnt.I("Arrow_right")) {
			pv.activeTexture = pv.activeTexture + 1
			if pv.activeTexture >= len(pv.textureKeys) {
				pv.activeTexture = 0
			}
		}
		imgui.EndDisabled()
	}
	imgui.EndChild()

	if imgui.Button(fnt.I("Home")) {
		pv.offset = imgui.NewVec2(0, 0)
		pv.zoom = 1
	}
	imgui.SetItemTooltip("Reset view")
	imgui.SameLine()
	imgui.BeginDisabledV(ddsPv == nil)
	if imgui.Checkbox("Linear filtering", &pv.linearFiltering) {
		filter := int32(gl.NEAREST)
		if pv.linearFiltering {
			filter = gl.LINEAR
		}
		gl.BindTexture(gl.TEXTURE_2D, ddsPv.textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filter)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	imgui.EndDisabled()
	imgui.SetItemTooltip("Linear filtering \"blurs\" pixels when zooming in. Disable to view individual pixels more clearly.")
	imgui.SameLine()
	imgui.BeginDisabledV(ddsPv == nil || !ddsPv.imageHasAlpha)
	var ignoreAlphaProxy bool
	if ddsPv != nil {
		ignoreAlphaProxy = ddsPv.ignoreAlpha
	}
	if imgui.Checkbox("Ignore alpha", &ignoreAlphaProxy) && ddsPv != nil {
		ddsPv.ignoreAlpha = ignoreAlphaProxy
		swizzleA := int32(gl.ALPHA)
		if ddsPv.ignoreAlpha {
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

	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY | imgui.TableFlagsRowBg
	if len(pv.settings) > 0 && imgui.BeginTableV("##Material Settings", 2, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupColumnV("Value", imgui.TableColumnFlagsWidthStretch, 2, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		clipper := imgui.NewListClipper()
		clipper.Begin(int32(len(pv.settingKeys)))
		for clipper.Step() {
			for row := clipper.DisplayStart(); row < clipper.DisplayEnd(); row++ {
				id := pv.settingKeys[row]
				imgui.PushIDStr(id)

				imgui.TableNextColumn()
				imgui.TextUnformatted(id)

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
				imgui.Text(settingString)
				imgui.PopID()
			}
		}
		imgui.EndTable()
	}
}
