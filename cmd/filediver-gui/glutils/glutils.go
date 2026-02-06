package glutils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"

	"github.com/go-gl/gl/v4.3-core/gl"
)

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
