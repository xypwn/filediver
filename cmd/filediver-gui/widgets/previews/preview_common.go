package previews

import (
	"fmt"
	"image"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
)

type GetResourceFunc func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error)

type DDSPreviewState struct {
	textureID       uint32
	textureRef      imgui.TextureRef // must be kept in sync with textureID
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
	pv.textureRef = *imgui.NewTextureRefTextureID(imgui.TextureID(pv.textureID))
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

func BuildImagePreviewArea(pv *DDSPreviewState, pos, area imgui.Vec2) {
	scaledImageSize := imgui.NewVec2(0, 0)
	if pv != nil {
		pv.offset.X = min(max(-1, pv.offset.X), 1)
		pv.offset.Y = min(max(-1, pv.offset.Y), 1)

		scale := pv.zoom
		{
			fitXScale, fitYScale := area.X/pv.imageSize.X, area.Y/pv.imageSize.Y
			scale *= min(fitXScale, fitYScale)
		}

		scaledImageSize = pv.imageSize.Mul(scale)
		offsetPx := imgui.NewVec2(pv.offset.X*scaledImageSize.X/2, pv.offset.Y*scaledImageSize.Y/2)
		imgPos := pos.Sub(scaledImageSize.Div(2)).Add(area.Div(2)).Add(offsetPx)
		imgui.WindowDrawList().AddImage(pv.textureRef, imgPos, imgPos.Add(scaledImageSize))
	}
	imgui.SetNextItemAllowOverlap()
	imgui.InvisibleButton("##overlay", area)
	io := imgui.CurrentIO()
	if imgui.IsItemActive() && pv != nil {
		md := io.MouseDelta()
		md.X /= scaledImageSize.X / 2
		md.Y /= scaledImageSize.Y / 2
		pv.offset = pv.offset.Add(md)
	}
	if imgui.IsItemHovered() && pv != nil {
		scroll := io.MouseWheel()
		pv.zoom = min(max(0.9, pv.zoom+(0.1*pv.zoom*scroll)), 1000)
	}
}
