package previews

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
)

func DDSPreview(name string, pv *DDSPreviewState) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	if pv.err != nil {
		imutils.TextError(pv.err)
		return
	}

	imutils.Textf("Size=(%v,%v)\nFormat=%v", pv.imageSize.X, pv.imageSize.Y, pv.ddsInfo.DXT10Header.DXGIFormat)

	{
		size := imgui.ContentRegionAvail()
		size.Y -= imutils.ComboHeight()
		imgui.SetNextWindowSize(size)
	}
	if imgui.BeginChildStr("##DDS image preview") {
		pos := imgui.CursorScreenPos()
		area := imgui.ContentRegionAvail()
		BuildImagePreviewArea(pv, pos, area)
	}
	imgui.EndChild()

	if imgui.Button(fnt.I("Home")) {
		pv.offset = imgui.NewVec2(0, 0)
		pv.zoom = 1
	}
	imgui.SetItemTooltip("Reset view")
	imgui.SameLine()
	if imgui.Checkbox("Linear filtering", &pv.linearFiltering) {
		filter := int32(gl.NEAREST)
		if pv.linearFiltering {
			filter = gl.LINEAR
		}
		gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filter)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	imgui.SetItemTooltip("Linear filtering \"blurs\" pixels when zooming in. Disable to view individual pixels more clearly.")
	imgui.SameLine()
	imgui.BeginDisabledV(!pv.imageHasAlpha)
	if imgui.Checkbox("Ignore alpha", &pv.ignoreAlpha) {
		swizzleA := int32(gl.ALPHA)
		if pv.ignoreAlpha {
			swizzleA = gl.ONE
		}
		gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, swizzleA)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	imgui.EndDisabled()
	if !pv.imageHasAlpha {
		imgui.SetItemTooltip("This image doesn't use an alpha component.")
	} else {
		imgui.SetItemTooltip("Ignore alpha component, making RGB components always fully visible.")
	}
}
