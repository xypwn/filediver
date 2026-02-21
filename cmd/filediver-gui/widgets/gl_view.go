package widgets

import (
	"errors"
	"log"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
)

type GLViewState struct {
	width      int32
	height     int32
	fbo        uint32 // frame buffer object
	textureID  uint32
	textureRef imgui.TextureRef // must be kept in sync with textureID
	rbo        uint32           // render buffer object
}

func NewGLView() (*GLViewState, error) {
	fb := &GLViewState{}
	gl.GenFramebuffers(1, &fb.fbo)
	gl.GenTextures(1, &fb.textureID)
	fb.textureRef = *imgui.NewTextureRefTextureID(imgui.TextureID(fb.textureID))
	gl.GenRenderbuffers(1, &fb.rbo)

	return fb, nil
}

func (fb *GLViewState) Delete() {
	gl.DeleteFramebuffers(1, &fb.fbo)
	gl.DeleteTextures(1, &fb.textureID)
	gl.DeleteRenderbuffers(1, &fb.rbo)
}

func (fb *GLViewState) maybeResize(newWidth, newHeight int32) error {
	if newWidth <= 0 || newHeight <= 0 {
		return nil
	}
	if fb.width != newWidth || fb.height != newHeight {
		return fb.resize(newWidth, newHeight)
	} else {
		return nil
	}
}

func (fb *GLViewState) resize(newWidth, newHeight int32) error {
	fb.width, fb.height = newWidth, newHeight

	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	gl.BindTexture(gl.TEXTURE_2D, fb.textureID)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, fb.width, fb.height, 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fb.textureID, 0)

	gl.BindRenderbuffer(gl.RENDERBUFFER, fb.rbo)
	defer gl.BindRenderbuffer(gl.RENDERBUFFER, 0)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, fb.width, fb.height)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, fb.rbo)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		return errors.New("framebuffer is incomplete")
	}

	return nil
}

func GLView(name string, fb *GLViewState, size imgui.Vec2, processViewAreaInputIG func(), drawGL func(pos, size imgui.Vec2), drawOverlayIG func(pos, size imgui.Vec2)) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	pos := imgui.CursorScreenPos()
	{
		avail := imgui.ContentRegionAvail()
		if size.X == 0 {
			size.X = avail.X
		}
		if size.Y == 0 {
			size.Y = avail.Y
		}
	}

	if err := fb.maybeResize(int32(size.X), int32(size.Y)); err != nil {
		log.Println(err)
		return
	}

	if size.X > 0 && size.Y > 0 {
		imgui.InvisibleButton("GLViewArea", size)
		if processViewAreaInputIG != nil {
			processViewAreaInputIG()
		}

		var oldViewport [4]int32
		gl.GetIntegerv(gl.VIEWPORT, &oldViewport[0])

		gl.BindFramebuffer(gl.FRAMEBUFFER, fb.fbo)
		gl.Viewport(0, 0, int32(size.X), int32(size.Y))
		drawGL(pos, size)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		gl.Viewport(oldViewport[0], oldViewport[1], oldViewport[2], oldViewport[3])

		imgui.WindowDrawList().AddImageV(
			fb.textureRef,
			pos,
			pos.Add(size),
			imgui.NewVec2(0, 1),
			imgui.NewVec2(1, 0),
			0xFFFFFFFF,
		)

		if drawOverlayIG != nil {
			drawOverlayIG(pos, size)
		}
	}
}
