package glutils

import (
	"errors"
	"fmt"

	"github.com/go-gl/gl/v3.2-core/gl"
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
func CreateProgram(vertShader, fragShader uint32) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)
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

func CreateProgramFromSources(vertSource string, fragSource string) (uint32, error) {
	vert, err := CreateShader(vertSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, fmt.Errorf("vertex shader: %w", err)
	}
	defer gl.DeleteShader(vert)

	frag, err := CreateShader(fragSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, fmt.Errorf("fragment shader: %w", err)
	}
	defer gl.DeleteShader(frag)

	program, err := CreateProgram(vert, frag)
	if err != nil {
		return 0, fmt.Errorf("link shader program: %w", err)
	}

	return program, nil
}
