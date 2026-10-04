package glutils

import (
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"path"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type GLError uint32

const (
	GL_NO_ERROR                      GLError = gl.NO_ERROR
	GL_INVALID_ENUM                  GLError = gl.INVALID_ENUM
	GL_INVALID_VALUE                 GLError = gl.INVALID_VALUE
	GL_INVALID_OPERATION             GLError = gl.INVALID_OPERATION
	GL_INVALID_FRAMEBUFFER_OPERATION GLError = gl.INVALID_FRAMEBUFFER_OPERATION
	GL_OUT_OF_MEMORY                 GLError = gl.OUT_OF_MEMORY
	GL_STACK_UNDERFLOW               GLError = gl.STACK_UNDERFLOW
	GL_STACK_OVERFLOW                GLError = gl.STACK_OVERFLOW
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=GLError

type GLTarget uint32

const (
	GL_TEXTURE_1D             GLTarget = gl.TEXTURE_1D
	GL_TEXTURE_2D             GLTarget = gl.TEXTURE_2D
	GL_TEXTURE_3D             GLTarget = gl.TEXTURE_3D
	GL_TEXTURE_1D_ARRAY       GLTarget = gl.TEXTURE_1D_ARRAY
	GL_TEXTURE_2D_ARRAY       GLTarget = gl.TEXTURE_2D_ARRAY
	GL_TEXTURE_CUBE_MAP       GLTarget = gl.TEXTURE_CUBE_MAP
	GL_TEXTURE_CUBE_MAP_ARRAY GLTarget = gl.TEXTURE_CUBE_MAP_ARRAY
	GL_TEXTURE_BUFFER         GLTarget = gl.TEXTURE_BUFFER
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=GLTarget

type GLType uint32

const (
	GL_BOOL           GLType = gl.BOOL
	GL_BYTE           GLType = gl.BYTE
	GL_UNSIGNED_BYTE  GLType = gl.UNSIGNED_BYTE
	GL_SHORT          GLType = gl.SHORT
	GL_UNSIGNED_SHORT GLType = gl.UNSIGNED_SHORT
	GL_INT            GLType = gl.INT
	GL_UNSIGNED_INT   GLType = gl.UNSIGNED_INT
	GL_FIXED          GLType = gl.FIXED
	GL_HALF_FLOAT     GLType = gl.HALF_FLOAT
	GL_FLOAT          GLType = gl.FLOAT
	GL_FLOAT_VEC2     GLType = gl.FLOAT_VEC2
	GL_FLOAT_VEC3     GLType = gl.FLOAT_VEC3
	GL_FLOAT_VEC4     GLType = gl.FLOAT_VEC4
	GL_DOUBLE         GLType = gl.DOUBLE
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=GLType

func (t GLType) Size() uint32 {
	switch t {
	case GL_BYTE, GL_UNSIGNED_BYTE:
		return 1
	case GL_SHORT, GL_UNSIGNED_SHORT, GL_HALF_FLOAT:
		return 2
	case GL_INT, GL_UNSIGNED_INT, GL_FIXED, GL_FLOAT, GL_BOOL:
		return 4
	case GL_DOUBLE, GL_FLOAT_VEC2:
		return 8
	case GL_FLOAT_VEC3:
		return 12
	case GL_FLOAT_VEC4:
		return 16
	}
	return 0
}

func (t GLType) Count() int32 {
	switch t {
	case GL_BYTE, GL_UNSIGNED_BYTE, GL_SHORT, GL_UNSIGNED_SHORT, GL_HALF_FLOAT, GL_INT, GL_UNSIGNED_INT, GL_FIXED, GL_FLOAT, GL_BOOL, GL_DOUBLE:
		return 1
	case GL_FLOAT_VEC2:
		return 2
	case GL_FLOAT_VEC3:
		return 3
	case GL_FLOAT_VEC4:
		return 4
	}
	return 0
}

func CreateShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	{
		length := int32(len(source))
		cSrc, free := gl.Strs(source)
		gl.ShaderSource(shader, 1, cSrc, &length)
		free()
	}
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := string(make([]uint8, logLength))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		gl.DeleteShader(shader)

		return 0, errors.New(log[:len(log)-1])
	}

	return shader, nil
}

// Doesn't delete the shaders (you may delete them after calling)
func CreateProgram(shaders ...uint32) (uint32, error) {
	program := gl.CreateProgram()
	for _, shader := range shaders {
		gl.AttachShader(program, shader)
	}
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := string(make([]uint8, logLength))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		gl.DeleteProgram(program)

		return 0, errors.New(log[:len(log)-1])
	}

	return program, nil
}

// Recognized extensions: .frag, .vert, .geom
func CreateProgramFromSources(fs fs.FS, paths ...string) (uint32, error) {
	var shaders []uint32
	for _, p := range paths {
		var shaderType uint32
		var shaderTypeName string
		ext := path.Ext(p)
		switch ext {
		case ".frag":
			shaderType = gl.FRAGMENT_SHADER
			shaderTypeName = "fragment"
		case ".vert":
			shaderType = gl.VERTEX_SHADER
			shaderTypeName = "vertex"
		case ".geom":
			shaderType = gl.GEOMETRY_SHADER
			shaderTypeName = "geometry"
		default:
			return 0, fmt.Errorf("\"%v\": unknown shader extension \"%v\"", p, ext)
		}
		r, err := fs.Open(p)
		if err != nil {
			return 0, fmt.Errorf("opening \"%v\": %w", p, err)
		}
		defer r.Close()
		data, err := io.ReadAll(r)
		if err != nil {
			return 0, fmt.Errorf("reading \"%v\": %w", p, err)
		}
		shader, err := CreateShader(string(data), shaderType)
		if err != nil {
			return 0, fmt.Errorf("%v shader: %w", shaderTypeName, err)
		}
		defer gl.DeleteShader(shader)
		shaders = append(shaders, shader)
	}

	program, err := CreateProgram(shaders...)
	if err != nil {
		return 0, fmt.Errorf("link shader program: %w", err)
	}

	return program, nil
}

// Writes the given Go image to the given OpenGL texture.
func ImageToTexture(texId uint32, img image.Image) error {
	gl.BindTexture(gl.TEXTURE_2D, texId)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)

	// NOTE(xypwn): This is almost duplicated with preview_common.go for performance
	// reasons (different assumptions).
	// Here, we can't assume the stride being to equal the width and the bounds starting
	// at 0 (which we do assume in preview_common.go).
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	data := make([]uint8, 4*width*height)
	switch img := img.(type) {
	case *image.Gray:
		for y := range height {
			for x := range width {
				i := y*width + x
				j := y*img.Stride + x*1
				y := img.Pix[j]
				data[4*i+0] = y
				data[4*i+1] = y
				data[4*i+2] = y
				data[4*i+3] = 255
			}
		}
	case *image.Gray16:
		for y := range height {
			for x := range width {
				i := y*width + x
				j := y*img.Stride + x*2
				y := img.Pix[j]
				data[4*i+0] = y
				data[4*i+1] = y
				data[4*i+2] = y
				data[4*i+3] = 255
			}
		}
	case *image.NRGBA:
		for y := range height {
			for x := range width {
				i := y*width + x
				j := y*img.Stride + x*4
				data[4*i+0] = img.Pix[j+0]
				data[4*i+1] = img.Pix[j+1]
				data[4*i+2] = img.Pix[j+2]
				data[4*i+3] = img.Pix[j+3]
			}
		}
	case *image.NRGBA64:
		for y := range height {
			for x := range width {
				i := y*width + x
				j := y*img.Stride + x*4
				data[4*i+0] = img.Pix[j+0]
				data[4*i+1] = img.Pix[j+2]
				data[4*i+2] = img.Pix[j+4]
				data[4*i+3] = img.Pix[j+6]
			}
		}
	default:
		return fmt.Errorf("unhandled image type %T", img)
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(width), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	return nil
}
