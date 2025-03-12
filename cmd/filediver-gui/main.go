package main

import (
	"bytes"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/sdlbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	currentBackend, _ := backend.CreateBackend(sdlbackend.NewSDLBackend())

	currentBackend.SetTargetFPS(60)
	currentBackend.SetWindowFlags(sdlbackend.SDLWindowFlagsResizable, 1)
	currentBackend.CreateWindow("Window Name", 800, 700)

	currentBackend.SetBgColor(imgui.NewVec4(0.2, 0.2, 0.2, 1))
	currentBackend.SetDropCallback(func(paths []string) {
		log.Println("drop:", paths)
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
	gl.Enable(gl.CULL_FACE)

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		var typStr string
		if gltype == gl.DEBUG_TYPE_ERROR {
			typStr = "error: "
		}
		log.Printf("GL: %v%v\n", typStr, message)
	}, nil)

	preview, err := widgets.CreateUnitPreview()
	if err != nil {
		log.Fatal(err)
	}
	defer preview.Delete()
	{
		mainB, err := os.ReadFile("cha_strider.unit.main")
		if err != nil {
			log.Fatal(err)
		}
		gpuB, err := os.ReadFile("cha_strider.unit.gpu")
		if err != nil {
			log.Fatal(err)
		}
		if err := preview.LoadUnit(bytes.NewReader(mainB), bytes.NewReader(gpuB)); err != nil {
			log.Fatal(err)
		}
	}

	currentBackend.Run(func() {
		viewport := imgui.MainViewport()
		imgui.SetNextWindowPos(viewport.Pos())
		imgui.SetNextWindowSize(viewport.Size())
		const mainWindowFlags = imgui.WindowFlagsNoDecoration | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoBringToFrontOnFocus | imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoNavFocus | imgui.WindowFlagsMenuBar
		if imgui.BeginV("Main", nil, mainWindowFlags) {
			if imgui.BeginMenuBar() {
				if imgui.BeginMenu("Help") {
					imgui.MenuItemBool("About")
					imgui.End()
				}

				imgui.EndMenuBar()
			}
		}
		imgui.End()

		{
			offset := imgui.NewVec2(0, 20)
			imgui.SetNextWindowPos(viewport.Pos().Add(offset))
			imgui.SetNextWindowSize(viewport.Size().Sub(offset))
		}
		const dockingSpaceWindowFlags = imgui.WindowFlagsNoDecoration | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoBringToFrontOnFocus | imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoNavFocus
		if imgui.BeginV("MainDockingSpace", nil, dockingSpaceWindowFlags) {
		}
		imgui.End()

		imgui.SetNextWindowSizeV(imgui.NewVec2(400, 400), imgui.CondOnce)
		previewWindowFlags := imgui.WindowFlags(0)
		if preview.IsUsing {
			previewWindowFlags |= imgui.WindowFlagsNoMove
		}
		if imgui.BeginV("Ligma0", nil, previewWindowFlags) {
			widgets.UnitPreview("Unit Preview", preview)
		}
		imgui.End()
	})
}
