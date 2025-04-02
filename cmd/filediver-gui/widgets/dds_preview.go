package widgets

import (
	"fmt"
	"image"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/dds"
)

type DDSPreviewState struct {
	textureID       uint32
	imageHasAlpha   bool
	imageSize       imgui.Vec2
	ddsInfo         dds.Info
	offset          imgui.Vec2 // -1 < x,y < 1
	zoom            float32
	linearFiltering bool
	ignoreAlpha     bool
	err             error
}

func NewDDSPreview() *DDSPreviewState {
	pv := &DDSPreviewState{
		zoom: 1,
	}

	gl.GenTextures(1, &pv.textureID)
	gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	pv.linearFiltering = true
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)

	return pv
}

func (pv *DDSPreviewState) Delete() {
	gl.DeleteTextures(1, &pv.textureID)
}

func (pv *DDSPreviewState) LoadImage(img *dds.DDS) {
	pv.err = nil

	gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	data := make([]uint8, 4*width*height)
	switch img := img.Image.(type) {
	case *image.Gray:
		for i := range width * height {
			y := img.Pix[i]
			data[4*i+0] = y
			data[4*i+1] = y
			data[4*i+2] = y
			data[4*i+3] = 255
		}
	case *image.Gray16:
		for i := range width * height {
			y := img.Pix[2*i]
			data[4*i+0] = y
			data[4*i+1] = y
			data[4*i+2] = y
			data[4*i+3] = 255
		}
	case *image.NRGBA:
		copy(data, img.Pix)
	case *image.NRGBA64:
		for i := range width * height {
			data[4*i+0] = img.Pix[8*i+0]
			data[4*i+1] = img.Pix[8*i+2]
			data[4*i+2] = img.Pix[8*i+4]
			data[4*i+3] = img.Pix[8*i+6]
		}
	default:
		pv.err = fmt.Errorf("unhandled image type %T", img)
		return
	}

	pv.imageHasAlpha = false
	for i := range width * height {
		a := data[4*i+3]
		if a != 255 {
			pv.imageHasAlpha = true
			break
		}
	}

	pv.imageSize = imgui.NewVec2(float32(width), float32(height))
	pv.ddsInfo = img.Info
	gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(pv.imageSize.X), int32(pv.imageSize.Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

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

		pv.offset.X = min(max(-1, pv.offset.X), 1)
		pv.offset.Y = min(max(-1, pv.offset.Y), 1)

		scale := pv.zoom
		{
			fitXScale, fitYScale := area.X/pv.imageSize.X, area.Y/pv.imageSize.Y
			scale *= min(fitXScale, fitYScale)
		}

		scaledImageSize := pv.imageSize.Mul(scale)
		offsetPx := imgui.NewVec2(pv.offset.X*scaledImageSize.X/2, pv.offset.Y*scaledImageSize.Y/2)
		imgPos := pos.Sub(scaledImageSize.Div(2)).Add(area.Div(2)).Add(offsetPx)
		imgui.WindowDrawList().AddImage(imgui.TextureID(pv.textureID), imgPos, imgPos.Add(scaledImageSize))
		imgui.InvisibleButton("##overlay", area)
		io := imgui.CurrentIO()
		if imgui.IsItemActive() {
			md := io.MouseDelta()
			md.X /= scaledImageSize.X / 2
			md.Y /= scaledImageSize.Y / 2
			pv.offset = pv.offset.Add(md)
		}
		if imgui.IsItemHovered() {
			scroll := io.MouseWheel()
			pv.zoom = min(max(0.9, pv.zoom+(0.1*pv.zoom*scroll)), 1000)
		}
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
	if !pv.imageHasAlpha {
		imgui.BeginDisabled()
	}
	if imgui.Checkbox("Ignore alpha", &pv.ignoreAlpha) {
		swizzleA := int32(gl.ALPHA)
		if pv.ignoreAlpha {
			swizzleA = gl.ONE
		}
		gl.BindTexture(gl.TEXTURE_2D, pv.textureID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, swizzleA)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}
	if !pv.imageHasAlpha {
		imgui.EndDisabled()
		imgui.SetItemTooltip("This image doesn't use an alpha component.")
	} else {
		imgui.SetItemTooltip("Ignore alpha component, making RGB components always fully visible.")
	}
}
