package widgets

import (
	"errors"
	"log"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
)

type Framebuffer struct {
	width     int32
	height    int32
	fbo       uint32 // frame buffer object
	textureID uint32
	rbo       uint32 // render buffer object
}

func CreateFramebuffer() (*Framebuffer, error) {
	fb := &Framebuffer{}
	gl.GenFramebuffers(1, &fb.fbo)
	gl.GenTextures(1, &fb.textureID)
	gl.GenRenderbuffers(1, &fb.rbo)

	return fb, nil
}

func (fb *Framebuffer) Delete() {
	gl.DeleteFramebuffers(1, &fb.fbo)
	gl.DeleteTextures(1, &fb.textureID)
	gl.DeleteRenderbuffers(1, &fb.rbo)
}

func (fb *Framebuffer) maybeResize(newWidth, newHeight int32) error {
	if newWidth <= 0 || newHeight <= 0 {
		return nil
	}
	if fb.width != newWidth || fb.height != newHeight {
		return fb.resize(newWidth, newHeight)
	} else {
		return nil
	}
}

func (fb *Framebuffer) resize(newWidth, newHeight int32) error {
	fb.width, fb.height = newWidth, newHeight

	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.fbo)
	defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	gl.BindTexture(gl.TEXTURE_2D, fb.textureID)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, fb.width, fb.height, 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
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

func GLView(name string, fb *Framebuffer, draw func()) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	pos := imgui.CursorScreenPos()
	size := imgui.ContentRegionAvail()

	if err := fb.maybeResize(int32(size.X), int32(size.Y)); err != nil {
		log.Println(err)
		return
	}

	if size.X > 0 && size.Y > 0 {
		var oldViewport [4]int32
		gl.GetIntegerv(gl.VIEWPORT, &oldViewport[0])

		gl.BindFramebuffer(gl.FRAMEBUFFER, fb.fbo)
		gl.Viewport(0, 0, int32(size.X), int32(size.Y))
		draw()
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		gl.Viewport(oldViewport[0], oldViewport[1], oldViewport[2], oldViewport[3])

		imgui.WindowDrawList().AddImageV(
			imgui.TextureID(fb.textureID),
			pos,
			pos.Add(size),
			imgui.NewVec2(0, 1),
			imgui.NewVec2(1, 0),
			0xFFFFFFFF,
		)
	}
}
