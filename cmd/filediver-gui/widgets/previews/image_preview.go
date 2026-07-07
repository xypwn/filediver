package previews

import (
	"fmt"
	"image"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
)

var placeholderImagePreviewImage = &ImagePreviewImage{}

type ImagePreviewFlags int

const (
	LinearFilteringButton ImagePreviewFlags = 1 << iota
	IgnoreAlphaButton
	MultipleImages
)

type ImagePreviewImage struct {
	DrawInfo func()
	Alt      string // text to display if the image is nil, defaults to "<no image>"

	textureId  uint32
	textureRef imgui.TextureRef // must be kept in sync with textureID
	hasAlpha   bool
	size       imgui.Vec2
}

type ImagePreview struct {
	Flags    ImagePreviewFlags
	DrawInfo func()
	Alt      string // text to display when there are no images, defaults to "<no images>"

	Images          []*ImagePreviewImage
	imageIdx        int
	offset          imgui.Vec2 // -1 < x,y < 1
	zoom            float32
	linearFiltering bool
	ignoreAlpha     bool
	err             error
}

func NewImagePreview() *ImagePreview {
	pv := &ImagePreview{Alt: "<no images>", zoom: 1}
	return pv
}

func (pv *ImagePreview) deleteImagesTextures() {
	for _, img := range pv.Images {
		if img.size != (imgui.Vec2{}) {
			gl.DeleteTextures(1, &img.textureId)
		}
	}
}

func (pv *ImagePreview) Delete() {
	pv.deleteImagesTextures()
}

func (pv *ImagePreview) addImage(img image.Image) {
	pvImg := &ImagePreviewImage{Alt: "<no image>"}
	defer func() {
		pv.Images = append(pv.Images, pvImg)
	}()

	if img == nil {
		return
	}

	gl.GenTextures(1, &pvImg.textureId)
	pvImg.textureRef = *imgui.NewTextureRefTextureID(imgui.TextureID(pvImg.textureId))
	gl.BindTexture(gl.TEXTURE_2D, pvImg.textureId)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	pv.linearFiltering = true
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)

	gl.BindTexture(gl.TEXTURE_2D, pvImg.textureId)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)

	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	data := make([]uint8, 4*width*height)
	switch img := img.(type) {
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

	pvImg.hasAlpha = false
	for i := range width * height {
		a := data[4*i+3]
		if a != 255 {
			pvImg.hasAlpha = true
			break
		}
	}

	pvImg.size = imgui.NewVec2(float32(width), float32(height))
	gl.BindTexture(gl.TEXTURE_2D, pvImg.textureId)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(pvImg.size.X), int32(pvImg.size.Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// Each image must be of type any of [*image.Gray], [*image.Gray16], [*image.NRGBA], [*image.NRGBA64]
// for performance reasons.
func (pv *ImagePreview) LoadImages(imgs []image.Image) {
	pv.deleteImagesTextures()
	pv.err = nil
	pv.Images = nil
	pv.imageIdx = 0
	pv.zoom = 1
	pv.offset = imgui.NewVec2(0, 0)

	for _, img := range imgs {
		pv.addImage(img)
	}
}

func (pv *ImagePreview) drawImage(pvImg *ImagePreviewImage, pos, area imgui.Vec2) {
	if pvImg.size == (imgui.Vec2{}) {
		var text string
		if pvImg != placeholderImagePreviewImage {
			text = pvImg.Alt
		} else {
			text = pv.Alt
		}
		textSize := imgui.CalcTextSize(text)
		textPos := pos.Add(area.Div(2)).Sub(textSize.Div(2))
		imgui.SetCursorScreenPos(textPos)
		imgui.TextUnformatted(text)
		return
	}

	scaledImageSize := imgui.NewVec2(0, 0)
	if pv != nil {
		pv.offset.X = min(max(-1, pv.offset.X), 1)
		pv.offset.Y = min(max(-1, pv.offset.Y), 1)

		scale := pv.zoom
		{
			fitXScale, fitYScale := area.X/pvImg.size.X, area.Y/pvImg.size.Y
			scale *= min(fitXScale, fitYScale)
		}

		scaledImageSize = pvImg.size.Mul(scale)
		offsetPx := imgui.NewVec2(pv.offset.X*scaledImageSize.X/2, pv.offset.Y*scaledImageSize.Y/2)
		imgPos := pos.Sub(scaledImageSize.Div(2)).Add(area.Div(2)).Add(offsetPx)
		imgui.WindowDrawList().AddImage(pvImg.textureRef, imgPos, imgPos.Add(scaledImageSize))
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

func (pv *ImagePreview) Draw(name string) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	if pv.err != nil {
		imutils.TextError(pv.err)
		return
	}

	pvImg := placeholderImagePreviewImage
	if len(pv.Images) != 0 {
		pvImg = pv.Images[pv.imageIdx]
	}
	if pv.DrawInfo != nil {
		pv.DrawInfo()
	}
	if pvImg.DrawInfo != nil {
		pvImg.DrawInfo()
	}
	{
		size := imgui.ContentRegionAvail()
		size.Y -= imutils.ComboHeight()
		imgui.SetNextWindowSize(size)
	}

	if imgui.BeginChildStrV("##image preview", imgui.NewVec2(0, 0), 0, 0) {
		pos := imgui.CursorScreenPos()
		area := imgui.ContentRegionAvail()
		pv.drawImage(pvImg, pos, area)

		if pv.Flags&MultipleImages != 0 {
			cycleTexDelta := 0
			if imgui.Shortcut(imgui.KeyChord(imgui.KeyLeftArrow)) {
				cycleTexDelta--
			}
			if imgui.Shortcut(imgui.KeyChord(imgui.KeyRightArrow)) {
				cycleTexDelta++
			}
			style := imgui.CurrentStyle()
			imgui.BeginDisabledV(len(pv.Images) < 2)
			imgui.PushItemFlag(imgui.ItemFlagsNoNav, true)
			imgui.SetCursorScreenPos(imgui.NewVec2(pos.X+style.ItemSpacing().X, pos.Y+area.Y/2))
			if imgui.Button(fnt.I.ArrowLeft) {
				cycleTexDelta--
			}
			imgui.SetCursorScreenPos(imgui.NewVec2(pos.X+area.X-imgui.FrameHeightWithSpacing()-style.ItemSpacing().X, pos.Y+area.Y/2))
			if imgui.Button(fnt.I.ArrowRight) {
				cycleTexDelta++
			}
			imgui.PopItemFlag()
			imgui.EndDisabled()

			if len(pv.Images) > 0 {
				mod := func(a, b int) int { // python-like modulo
					return (a%b + b) % b
				}
				pv.imageIdx = mod(pv.imageIdx+cycleTexDelta, len(pv.Images))
			}
		}
	}
	imgui.EndChild()

	if imgui.Button(fnt.I.Home) {
		pv.offset = imgui.NewVec2(0, 0)
		pv.zoom = 1
	}
	imgui.SetItemTooltip("Reset view")
	if pv.Flags&LinearFilteringButton != 0 {
		imgui.SameLine()
		if imgui.Checkbox("Linear filtering", &pv.linearFiltering) {
			filter := int32(gl.NEAREST)
			if pv.linearFiltering {
				filter = gl.LINEAR
			}
			gl.BindTexture(gl.TEXTURE_2D, pvImg.textureId)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filter)
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}
		imgui.SetItemTooltip("Linear filtering \"blurs\" pixels when zooming in. Disable to view individual pixels more clearly.")
	}
	if pv.Flags&IgnoreAlphaButton != 0 {
		imgui.SameLine()
		imgui.BeginDisabledV(!pvImg.hasAlpha)
		if imgui.Checkbox("Ignore alpha", &pv.ignoreAlpha) {
			swizzleA := int32(gl.ALPHA)
			if pv.ignoreAlpha {
				swizzleA = gl.ONE
			}
			gl.BindTexture(gl.TEXTURE_2D, pvImg.textureId)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_SWIZZLE_A, swizzleA)
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}
		imgui.EndDisabled()
		if !pvImg.hasAlpha {
			imgui.SetItemTooltip("This image doesn't use an alpha component.")
		} else {
			imgui.SetItemTooltip("Ignore alpha component, making RGB components always fully visible.")
		}
	}
}
