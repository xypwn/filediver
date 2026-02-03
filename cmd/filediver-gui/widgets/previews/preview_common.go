package previews

import (
	"embed"
	"encoding/binary"
	"fmt"
	"image"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit"
)

//go:embed shaders/*
var PreviewShaderCode embed.FS

type GetResourceFunc func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error)

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
		imgui.WindowDrawList().AddImage(imgui.TextureID(pv.textureID), imgPos, imgPos.Add(scaledImageSize))
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

type PreviewMeshBuffer struct {
	vao uint32 // vertex array object
	ibo uint32 // index buffer object
	vbo uint32 // vertex buffer object

	numVertices int32
	numIndices  int32
	idxStride   int32
}

func (buf *PreviewMeshBuffer) GenObjects() {
	gl.GenVertexArrays(1, &buf.vao)
	gl.GenBuffers(1, &buf.vbo)
	gl.GenBuffers(1, &buf.ibo)

	gl.BindVertexArray(buf.vao)
	defer gl.BindVertexArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, buf.vbo)
	defer gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, buf.ibo)
}

func (buf PreviewMeshBuffer) DeleteObjects() {
	gl.DeleteVertexArrays(1, &buf.vao)
	gl.DeleteBuffers(1, &buf.vbo)
	gl.DeleteBuffers(1, &buf.ibo)
}

func getLayoutIdx(item unit.MeshLayoutItem) uint32 {
	switch item.Type {
	case unit.ItemPosition:
		return 0
	case unit.ItemNormal:
		return 1
	case unit.ItemTangent:
		return 2
	case unit.ItemBinormal:
		return 3
	case unit.ItemUVCoords:
		return 4 + (item.Layer << 4)
	case unit.ItemBoneIdx:
		return 5 + (item.Layer << 4)
	case unit.ItemBoneWeight:
		return 6 + (item.Layer << 4)
	}
	return 0xffffffff
}

func getLayoutType(item unit.MeshLayoutItem) uint32 {
	switch item.Format {
	case unit.FormatF32, unit.FormatVec2F, unit.FormatVec3F, unit.FormatVec4F:
		return gl.FLOAT
	case unit.FormatF16, unit.FormatVec2F16, unit.FormatVec3F16, unit.FormatVec4F16:
		return gl.HALF_FLOAT
	case unit.FormatU32, unit.FormatVec2U32, unit.FormatVec3U32, unit.FormatVec4U32:
		return gl.UNSIGNED_INT
	case unit.FormatS32:
		return gl.INT
	case unit.FormatS8, unit.FormatVec2S8, unit.FormatVec3S8, unit.FormatVec4S8:
		return gl.BYTE
	case unit.FormatVec4R10G10B10A2_TYPELESS, unit.FormatVec4R10G10B10A2_UNORM:
		return gl.UNSIGNED_INT
	}
	return 0
}

func (buf *PreviewMeshBuffer) LoadLayout(layout unit.MeshLayout, gpuData []byte) {
	buf.numVertices = int32(layout.NumVertices)
	buf.numIndices = int32(layout.NumIndices)
	buf.idxStride = int32(layout.IndicesSize / layout.NumIndices)

	gl.BindVertexArray(buf.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, int(layout.PositionsSize), gl.Ptr(&gpuData[layout.VertexOffset]), gl.STATIC_DRAW)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(layout.IndicesSize), gl.Ptr(&gpuData[layout.IndexOffset]), gl.STATIC_DRAW)

	var offset uintptr = 0
	for idx := range layout.NumItems {
		layoutIdx := getLayoutIdx(layout.Items[idx])
		gl.VertexAttribPointerWithOffset(
			layoutIdx,
			int32(layout.Items[idx].Format.Type().Components()),
			getLayoutType(layout.Items[idx]),
			false,
			int32(layout.VertexStride),
			offset,
		)
		gl.EnableVertexAttribArray(layoutIdx)
		offset += uintptr(layout.Items[idx].Format.Size())
	}
}

var planeVertices []float32 = []float32{
	//  Position,     Normal,  UV0,  UV1,  UV2, Idx, Weight,
	-1, -1, 0, 1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0,
	1, -1, 0, 1, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0,
	-1, 1, 0, 1, 0, 0, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0,
}

var planeIndices []uint32 = []uint32{
	0, 1, 3,
	1, 2, 3,
}

func getPlane() (*PreviewMeshBuffer, error) {
	var meshPreviewBuffer PreviewMeshBuffer
	meshPreviewBuffer.GenObjects()
	gpuData, err := binary.Append(nil, binary.LittleEndian, planeVertices)
	if err != nil {
		return nil, err
	}
	gpuData, err = binary.Append(gpuData, binary.LittleEndian, planeIndices)
	if err != nil {
		return nil, err
	}
	meshPreviewBuffer.LoadLayout(unit.MeshLayout{
		NumItems: 7,
		Items: [16]unit.MeshLayoutItem{
			{
				Type:   unit.ItemPosition,
				Format: unit.FormatVec4F,
				Layer:  0,
			},
			{
				Type:   unit.ItemNormal,
				Format: unit.FormatVec4F,
				Layer:  0,
			},
			{
				Type:   unit.ItemUVCoords,
				Format: unit.FormatVec2F,
				Layer:  0,
			},
			{
				Type:   unit.ItemUVCoords,
				Format: unit.FormatVec2F,
				Layer:  1,
			},
			{
				Type:   unit.ItemUVCoords,
				Format: unit.FormatVec2F,
				Layer:  2,
			},
			{
				Type:   unit.ItemBoneIdx,
				Format: unit.FormatU32,
				Layer:  0,
			},
			{
				Type:   unit.ItemBoneWeight,
				Format: unit.FormatF32,
				Layer:  0,
			},
		},
	}, gpuData)

	return &meshPreviewBuffer, nil
}

func ComputeMVP(model mgl32.Mat4, viewRotation mgl32.Vec2, viewDistance, vfov, aspectRatio float32) (
	normal mgl32.Mat3,
	viewPosition mgl32.Vec3,
	view mgl32.Mat4,
	projection mgl32.Mat4,
) {
	normal = model.Inv().Transpose().Mat3()
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
		vfov,
		aspectRatio,
		0.001,
		32768,
	)
	return
}
