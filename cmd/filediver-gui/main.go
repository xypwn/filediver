package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/sdlbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

func init() {
	runtime.LockOSThread()
}

type LoadGameDataState struct {
	sync.Mutex
	Progress float32
	Result   struct {
		*app.App
		SortedFileIDs []stingray.FileID
	}
	Err         error
	Done        bool
	JustGotDone bool
}

func loadGameData(state *LoadGameDataState) {
	ctx := context.Background()
	gameDir, err := app.DetectGameDir()
	if err != nil {
		state.Lock()
		state.Err = fmt.Errorf("Helldivers 2 Steam installation path not found: %w", err)
		state.Unlock()
		return
	}

	defer func() {
		state.Lock()
		state.Done = true
		state.JustGotDone = true
		if state.Result.App != nil {
			files := state.Result.DataDir.Files
			for id := range files {
				state.Result.SortedFileIDs = append(state.Result.SortedFileIDs, id)
			}
			fileName := func(id stingray.FileID) string {
				return state.Result.LookupHash(id.Name) + "." + state.Result.LookupHash(id.Type)
			}
			slices.SortFunc(state.Result.SortedFileIDs, func(aID, bID stingray.FileID) int {
				return strings.Compare(fileName(aID), fileName(bID))
			})
		}
		state.Unlock()
	}()

	a, err := app.OpenGameDir(ctx, gameDir, app.ParseHashes(hashes.Hashes), app.ParseHashes(hashes.ThinHashes), nil, stingray.Hash{}, func(curr, total int) {
		state.Lock()
		state.Progress = float32(curr) / float32(total)
		state.Unlock()
	})
	if err != nil {
		state.Lock()
		state.Err = err
		state.Unlock()
		return
	}

	state.Lock()
	state.Result.App = a
	state.Unlock()
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
	gl.FrontFace(gl.CW)

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		var typStr string
		if gltype == gl.DEBUG_TYPE_ERROR {
			typStr = "error: "
		}
		log.Printf("GL: %v%v\n", typStr, message)
	}, nil)

	var loadGameDataState LoadGameDataState

	go loadGameData(&loadGameDataState)

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

	var gameFileSearchQuery string
	var matchingFiles map[stingray.FileID]*stingray.File

	currentBackend.Run(func() {
		viewport := imgui.MainViewport()
		imgui.SetNextWindowPos(viewport.Pos())
		imgui.SetNextWindowSize(viewport.Size())
		const mainWindowFlags = imgui.WindowFlagsNoDecoration | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoBringToFrontOnFocus | imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoNavFocus | imgui.WindowFlagsMenuBar
		if imgui.BeginV("##Main", nil, mainWindowFlags) {
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
		if imgui.BeginV("##MainDockingSpace", nil, dockingSpaceWindowFlags) {
		}
		imgui.End()

		imgui.SetNextWindowPosV(viewport.Pos().Add(imgui.NewVec2(400, 400)), imgui.CondOnce, imgui.Vec2{})
		imgui.SetNextWindowSizeV(imgui.NewVec2(400, 400), imgui.CondOnce)
		if imgui.Begin("Browser") {
			loadGameDataState.Lock()
			if loadGameDataState.Done {
				if loadGameDataState.Err == nil {
					a := loadGameDataState.Result
					if loadGameDataState.JustGotDone {
						matchingFiles = a.DataDir.Files
					}
					if imgui.InputTextWithHint("##Search", "Search", &gameFileSearchQuery, 0, nil) {
						cfg, err := app.ParseExtractorConfig(app.ConfigFormat, "")
						if err != nil {
							log.Printf("app.ParseExtractorConfig: %v\n", err)
						}
						matchingFiles, err = a.MatchingFiles(gameFileSearchQuery, "", nil, app.ConfigFormat, cfg)
						if err != nil {
							log.Printf("app.MatchingFiles: %v\n", err)
						}
					}
					imgui.BeginListBoxV("##Game Files", imgui.ContentRegionAvail())
					for _, id := range loadGameDataState.Result.SortedFileIDs {
						if _, ok := matchingFiles[id]; ok {
							imgui.TextUnformatted(fmt.Sprintf("%v.%v", a.LookupHash(id.Name), a.LookupHash(id.Type)))
						}
					}
					imgui.EndListBox()
				} else {
					imgui.TextUnformatted(fmt.Sprintf("Error: %v", loadGameDataState.Err))
				}
				loadGameDataState.JustGotDone = false
			} else {
				imgui.TextUnformatted("Loading game data...")
				imgui.ProgressBar(loadGameDataState.Progress)
			}
			loadGameDataState.Unlock()
		}
		imgui.End()

		imgui.SetNextWindowSizeV(imgui.NewVec2(400, 400), imgui.CondOnce)
		previewWindowFlags := imgui.WindowFlags(0)
		if preview.IsUsing {
			previewWindowFlags |= imgui.WindowFlagsNoMove
		}
		if imgui.BeginV("Preview", nil, previewWindowFlags) {
			widgets.UnitPreview("Unit Preview", preview)
		}
		imgui.End()
	})
}
