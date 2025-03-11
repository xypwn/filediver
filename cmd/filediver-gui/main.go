package main

import (
	"log"
	"runtime"
	"strings"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/sdlbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
)

var bg = [4]float32{0, 0, 0, 1}
var color = [4]float32{0, 1, 0, 0.5}

func init() {
	runtime.LockOSThread()
}

func main() {
	currentBackend, _ := backend.CreateBackend(sdlbackend.NewSDLBackend())

	currentBackend.SetTargetFPS(60)
	currentBackend.SetWindowFlags(sdlbackend.SDLWindowFlagsResizable, 1)
	currentBackend.CreateWindow("Window Name", 800, 700)

	currentBackend.SetBgColor(imgui.NewVec4(bg[0], bg[1], bg[2], bg[3]))
	currentBackend.SetDropCallback(func(paths []string) {
		println(paths)
	})

	flags := imgui.CurrentIO().ConfigFlags()
	flags &= ^imgui.ConfigFlagsViewportsEnable
	flags |= imgui.ConfigFlagsDockingEnable | imgui.ConfigFlagsViewportsEnable
	imgui.CurrentIO().SetConfigFlags(flags)
	imgui.CurrentIO().SetIniFilename("")

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	program := createShaderProgram()

	vertices := []float32{
		0.0, 0.5, 0.0,
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)

	colorLoc := gl.GetUniformLocation(program, gl.Str("color\x00"))
	gl.UseProgram(program)
	gl.Uniform4fv(colorLoc, 1, &color[0])

	framebuffer, err := widgets.CreateFramebuffer()
	if err != nil {
		log.Fatal(err)
	}
	defer framebuffer.Delete()

	currentBackend.Run(func() {
		viewport := imgui.MainViewport()
		imgui.SetNextWindowPos(viewport.Pos())
		imgui.SetNextWindowSize(viewport.Size())
		if imgui.BeginV("Main", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoBringToFrontOnFocus|imgui.WindowFlagsNoSavedSettings|imgui.WindowFlagsNoDocking|imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoNavFocus|imgui.WindowFlagsMenuBar) {
			if imgui.BeginMenuBar() {
				if imgui.BeginMenu("Help") {
					imgui.MenuItemBool("About")
					imgui.End()
				}

				imgui.EndMenuBar()
			}
		}
		imgui.End()

		if imgui.Begin("Ligma0") {
			if imgui.ColorEdit4V("Background", &bg, imgui.ColorEditFlagsAlphaBar) {
				currentBackend.SetBgColor(imgui.NewVec4(bg[0], bg[1], bg[2], bg[3]))
			}

			if imgui.ColorEdit4V("Color", &color, imgui.ColorEditFlagsAlphaBar) {
				gl.Uniform4fv(colorLoc, 1, &color[0])
			}

			widgets.GLView("3D View", framebuffer, func() {
				gl.ClearColor(0.5, 0.5, 0.5, 1)
				gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

				gl.UseProgram(program)
				gl.BindVertexArray(vao)
				gl.DrawArrays(gl.TRIANGLES, 0, 3)
				gl.BindVertexArray(0)
				gl.UseProgram(0)
			})
		}

		imgui.End()

	})
}

const vertexShaderSource = `#version 320 es
layout(location = 0) in vec3 position;

void main() {
    gl_Position = vec4(position, 1.0);
}
`

const fragmentShaderSource = `#version 320 es
precision mediump float;

layout(location = 0) out vec4 fragColor;
uniform vec4 color;

void main() {
    fragColor = color;
}
`

func compileShader(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)
	cSources, free := gl.Strs(source)
	defer free()
	gl.ShaderSource(shader, 1, cSources, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var length int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &length)
		log.Println("Shader compilation failed:", source)
		infoLog := strings.Repeat("\x00", int(length+1))
		gl.GetShaderInfoLog(shader, length, nil, gl.Str(infoLog))
		log.Println(infoLog)
	}

	return shader
}

func createShaderProgram() uint32 {
	vertexShader := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	fragmentShader := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}
