package main

/*
// GLFW
typedef struct GLFWwindow GLFWwindow;
typedef void (*GLFWwindowsizefun)(GLFWwindow* window, int width, int height);
typedef void (*GLFWwindowrefreshfun)(GLFWwindow* window);
int glfwWindowShouldClose(GLFWwindow *window);
void glfwPollEvents();
void glfwGetFramebufferSize(GLFWwindow *window, int *width, int *height);
void glfwSwapInterval(int interval);
GLFWwindow *glfwGetCurrentContext();
void glfwMakeContextCurrent(GLFWwindow *window);
void *glfwGetWindowUserPointer(GLFWwindow *window);
void glfwSetWindowUserPointer(GLFWwindow *window, void *pointer);
void glfwSwapBuffers(GLFWwindow *window);
void glfwDestroyWindow(GLFWwindow *window);
void glfwTerminate();
GLFWwindowsizefun glfwSetWindowSizeCallback(GLFWwindow *window, GLFWwindowsizefun callback);
GLFWwindowrefreshfun glfwSetWindowRefreshCallback(GLFWwindow* window, GLFWwindowrefreshfun callback);

// cimgui
typedef struct ImGuiContext ImGuiContext;
typedef struct ImDrawData ImDrawData;
void igNewFrame();
ImDrawData* igGetDrawData();

// ImGui implementation
void ImGui_ImplOpenGL3_NewFrame();
void ImGui_ImplGlfw_NewFrame();
void ImGui_ImplOpenGL3_RenderDrawData(ImDrawData* draw_data);
void ImGui_ImplOpenGL3_DestroyFontsTexture();
void ImGui_ImplOpenGL3_Shutdown();
void ImGui_ImplGlfw_Shutdown();

// cimgui-go C++ wrapper stuff
typedef void (*VoidCallback)();
void glfw_render(GLFWwindow *window, VoidCallback renderLoop);

// custom functions
void goWindowResizeCallback(GLFWwindow* window, int height, int width);
void goWindowRefreshCallback(GLFWwindow* window);
*/
import "C"

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/ebitengine/oto/v3"
	"github.com/go-gl/gl/v3.2-core/gl"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/stingray"
)

var OnWindowResize func(window *C.GLFWwindow, width int32, height int32)

//export goWindowResizeCallback
func goWindowResizeCallback(window *C.GLFWwindow, width C.int, height C.int) {
	OnWindowResize(window, int32(width), int32(height))
}

var OnWindowRefresh func(window *C.GLFWwindow)

//export goWindowRefreshCallback
func goWindowRefreshCallback(window *C.GLFWwindow) {
	OnWindowRefresh(window)
}

var IconsFontRanges = [3]imgui.Wchar{
	imgui.Wchar(fnt.IconsFontInfo.Min),
	imgui.Wchar(fnt.IconsFontInfo.Max16),
	0,
}

func UpdateGUIScale(guiScale float32) {
	io := imgui.CurrentIO()
	fonts := io.Fonts()
	fonts.Clear()

	style := imgui.NewStyle()

	fontSize := 15 * guiScale
	iconsFontSize := fontSize * 1.2
	{
		cfg := imgui.NewFontConfig()
		cfg.SetFontDataOwnedByAtlas(false)
		fonts.AddFontFromMemoryTTFV(
			uintptr(unsafe.Pointer(&fnt.TextFont[0])),
			int32(len(fnt.TextFont)),
			fontSize,
			cfg,
			nil,
		)
		cfg.Destroy()
	}

	{
		cfg := imgui.NewFontConfig()
		cfg.SetMergeMode(true)
		cfg.SetGlyphOffset(imgui.NewVec2(0, iconsFontSize-fontSize))
		cfg.SetGlyphMinAdvanceX(iconsFontSize)
		cfg.SetFontDataOwnedByAtlas(false)
		fonts.AddFontFromMemoryTTFV(
			uintptr(unsafe.Pointer(&fnt.IconsFont[0])),
			int32(len(fnt.IconsFont)),
			iconsFontSize,
			cfg,
			&IconsFontRanges[0],
		)
		cfg.Destroy()
	}

	C.ImGui_ImplOpenGL3_DestroyFontsTexture()

	io.SetFontGlobalScale(1)
	style.ScaleAllSizes(guiScale)
	io.Ctx().SetStyle(*style)
}

func main() {
	runtime.LockOSThread()

	var glfwWindow *C.GLFWwindow

	currentBackend, err := backend.CreateBackend(glfwbackend.NewGLFWBackend())
	if err != nil {
		log.Fatalf("Error creating backend: %v", err)
	}
	currentBackend.SetAfterCreateContextHook(func() {
		glfwWindow = C.glfwGetCurrentContext()
	})

	{
		const GLFW_CONTEXT_DEBUG = 0x00022007
		currentBackend.SetWindowFlags(GLFW_CONTEXT_DEBUG, 1)
	}
	currentBackend.SetWindowFlags(glfwbackend.GLFWWindowFlagsResizable, 1)
	currentBackend.CreateWindow("Filediver GUI", 800, 700)

	C.glfwMakeContextCurrent(glfwWindow)
	C.glfwSwapInterval(0)

	currentBackend.SetDropCallback(func(paths []string) {
		log.Println("drop:", paths)
	})

	io := imgui.CurrentIO()
	flags := io.ConfigFlags()
	flags |= imgui.ConfigFlagsDockingEnable | imgui.ConfigFlagsViewportsEnable
	io.SetConfigFlags(flags)
	io.SetIniFilename("")

	var guiScale float32
	{
		_, yScale := currentBackend.ContentScale()
		guiScale = yScale
	}
	shouldUpdateGUIScale := true

	var targetFPS float64 = 60

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

	const audioSampleRate = 48000
	var otoCtx *oto.Context
	{
		var readyChan chan struct{}
		otoCtx, readyChan, err = oto.NewContext(&oto.NewContextOptions{
			SampleRate:   audioSampleRate,
			ChannelCount: 2,
			Format:       oto.FormatFloat32LE,
		})
		<-readyChan
	}

	ctx := context.Background()

	var gameDataLoad GameDataLoad
	var gameData *GameData

	gameDataLoad.GoLoadGameData(ctx)

	var previewState *widgets.FileAutoPreviewState
	defer func() {
		if previewState != nil {
			previewState.Delete()
		}
	}()

	var gameFileSearchQuery string

	isPreferencesOpen := false

	preDraw := func() {
		if gameData != nil && previewState == nil {
			var err error
			previewState, err = widgets.NewFileAutoPreview(
				otoCtx, audioSampleRate,
				gameData.Hashes,
				func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error) {
					file, ok := gameData.DataDir.Files[id]
					if !ok {
						return nil, false, nil
					}
					data, err = file.Read(typ)
					if err != nil {
						return nil, true, err
					}
					return data, true, nil
				},
			)
			if err != nil {
				log.Fatal("Error creating unit preview:", err)
			}
		}

		if shouldUpdateGUIScale {
			UpdateGUIScale(guiScale)
			shouldUpdateGUIScale = false
		}
	}

	draw := func() {
		viewport := imgui.MainViewport()
		imgui.SetNextWindowPos(viewport.Pos())
		imgui.SetNextWindowSize(imgui.NewVec2(viewport.Size().X, 0))
		imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.NewVec2(0, 0))
		dockSpacePos := viewport.Pos()
		dockSpaceSize := viewport.Size()
		const mainWindowFlags = imgui.WindowFlagsNoDecoration | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoBringToFrontOnFocus | imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoNavFocus | imgui.WindowFlagsMenuBar | imgui.WindowFlagsNoDocking
		if imgui.BeginV("##Main", nil, mainWindowFlags) {
			if imgui.BeginMenuBar() {
				{
					menuHeight := imgui.FrameHeight()
					dockSpacePos.Y += menuHeight
					dockSpaceSize.Y -= menuHeight
				}
				if imgui.BeginMenu("Help") {
					imgui.MenuItemBool(fnt.I("Help") + " Tutorial")
					imgui.MenuItemBool(fnt.I("Info") + " About")
					imgui.EndMenu()
				}
				if imgui.BeginMenu("Settings") {
					if imgui.MenuItemBool(fnt.I("Settings") + " Preferences") {
						isPreferencesOpen = true
					}
					imgui.EndMenu()
				}
				imgui.EndMenuBar()
			}
		}
		imgui.End()
		imgui.PopStyleVar()

		imgui.SetNextWindowPos(dockSpacePos)
		imgui.SetNextWindowSize(dockSpaceSize)
		imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.NewVec2(0, 0))
		const dockSpaceWindowFlags = imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoDocking | imgui.WindowFlagsNoBringToFrontOnFocus | imgui.WindowFlagsNoNavFocus //| imgui.WindowFlagsNoBackground
		if imgui.BeginV("##MainDockSpace", nil, dockSpaceWindowFlags) {
			winClass := imgui.NewWindowClass() // passing nil as window class causes a nil pointer dereference (probably an error in the binding generation)
			id := imgui.IDStr("MainDockSpace")
			if imgui.InternalDockBuilderGetNode(id).CData == nil {
				imgui.InternalDockBuilderAddNodeV(id, imgui.DockNodeFlags(imgui.DockNodeFlagsDockSpace))
				imgui.InternalDockBuilderSetNodeSize(id, dockSpaceSize)
				var leftID, rightID imgui.ID
				imgui.InternalDockBuilderSplitNode(id, imgui.DirLeft, 0.5, &leftID, &rightID)
				imgui.InternalDockBuilderDockWindow("Browser", leftID)
				imgui.InternalDockBuilderDockWindow("Preview", rightID)
				imgui.InternalDockBuilderFinish(id)
			}
			imgui.DockSpaceV(id, imgui.NewVec2(0, 0), 0, winClass)
		}
		imgui.End()
		imgui.PopStyleVar()

		if imgui.Begin("Browser") {
			if gameData == nil {
				gameDataLoad.Lock()
				if gameDataLoad.Done {
					if gameDataLoad.Err == nil {
						if gameData == nil {
							gameData = gameDataLoad.Result
						}
					} else {
						imgui.TextUnformatted(fmt.Sprintf("Error: %v", gameDataLoad.Err))
					}
				} else {
					imgui.TextUnformatted(fnt.I("Hourglass_top") + " Loading game data...")
					imgui.ProgressBar(gameDataLoad.Progress)
				}
				gameDataLoad.Unlock()
			} else {
				if imgui.InputTextWithHint("##Search", fnt.I("Search")+" Search...", &gameFileSearchQuery, 0, nil) {
					gameData.UpdateSearchQuery(gameFileSearchQuery)
				}
				const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY
				if imgui.BeginTableV("##Game Files", 2, tableFlags, imgui.NewVec2(0, 0), 0) {
					imgui.TableSetupColumn("Name")
					imgui.TableSetupColumn("Type")
					imgui.TableSetupScrollFreeze(0, 1)
					imgui.TableHeadersRow()

					clipper := imgui.NewListClipper()
					clipper.Begin(int32(len(gameData.SortedSearchResultFileIDs)))
					for clipper.Step() {
						for row := clipper.DisplayStart(); row < clipper.DisplayEnd(); row++ {
							id := gameData.SortedSearchResultFileIDs[row]
							imgui.PushIDStr(id.Name.String() + id.Type.String()) // might be a bit slow
							imgui.TableNextColumn()
							selected := imgui.SelectableBoolV(gameData.LookupHash(id.Name), false, imgui.SelectableFlagsSpanAllColumns, imgui.NewVec2(0, 0))
							imgui.TableNextColumn()
							imgui.TextUnformatted(gameData.LookupHash(id.Type))
							imgui.PopID()
							if selected {
								previewState.LoadFile(ctx, gameData.DataDir.Files[id])
							}
						}
					}

					imgui.EndTable()
				}
			}
		}
		imgui.End()

		if imgui.Begin("Preview") {
			if previewState == nil || !widgets.FileAutoPreview("Preview", previewState) {
				imgui.TextUnformatted("Nothing to preview")
			}
		}
		imgui.End()

		if isPreferencesOpen {
			imgui.OpenPopupStr("Preferences")
		}
		imgui.SetNextWindowSize(imgui.NewVec2(0, 0))
		imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		if imgui.BeginPopupModalV("Preferences", &isPreferencesOpen, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize) {
			valueCombo := func(title string, selected float32, values []float32, onChanged func(v float32)) {
				if imgui.BeginCombo(title, fmt.Sprint(selected)) {
					for _, value := range values {
						if value == selected {
							imgui.SetItemDefaultFocus()
						}
						label := fmt.Sprint(value)
						if value == selected {
							label += " (selected)"
						}
						if imgui.SelectableBool(label) {
							onChanged(value)
						}
					}
					imgui.EndCombo()
				}
			}
			valueCombo(
				"GUI Scale",
				guiScale,
				[]float32{0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2, 2.25, 2.5, 2.75, 3},
				func(v float32) {
					guiScale = v
					shouldUpdateGUIScale = true
				},
			)
			valueCombo(
				"Target FPS",
				float32(targetFPS),
				[]float32{15, 30, 60, 75, 90, 120, 144, 165, 244, 300},
				func(v float32) {
					targetFPS = float64(v)
				},
			)

			if imgui.ButtonV("Close", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				imgui.CloseCurrentPopup()
				isPreferencesOpen = false
			}
			imgui.EndPopup()
		}
	}

	lastDrawTimestamp := time.Now()
	drawAndPresentFrame := func() {
		C.ImGui_ImplGlfw_NewFrame()
		C.ImGui_ImplOpenGL3_NewFrame()
		C.igNewFrame()
		gl.ClearColor(0.2, 0.2, 0.2, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		draw()

		imgui.Render()
		C.ImGui_ImplOpenGL3_RenderDrawData(C.igGetDrawData())

		if imgui.CurrentIO().ConfigFlags()&imgui.ConfigFlagsViewportsEnable != 0 {
			prevContext := C.glfwGetCurrentContext()
			imgui.UpdatePlatformWindows()
			imgui.RenderPlatformWindowsDefault()
			C.glfwMakeContextCurrent(prevContext)
		}

		C.glfwSwapBuffers(glfwWindow)

		targetFrameTime := time.Duration(float64(time.Second) / targetFPS)
		lastDrawTimestamp = lastDrawTimestamp.Add(targetFrameTime)
	}
	C.glfwSetWindowSizeCallback(glfwWindow, (*[0]byte)(C.goWindowResizeCallback))
	OnWindowResize = func(window *C.GLFWwindow, width, height int32) {
		gl.Viewport(0, 0, width, height)
		//drawAndPresentFrame()
	}
	C.glfwSetWindowRefreshCallback(glfwWindow, (*[0]byte)(C.goWindowRefreshCallback))
	OnWindowRefresh = func(window *C.GLFWwindow) {
		drawAndPresentFrame()
	}
	for C.glfwWindowShouldClose(glfwWindow) == 0 {
		preDraw()

		timeToDraw := time.Now().Sub(lastDrawTimestamp)
		numFramesToDraw := timeToDraw.Seconds() * targetFPS
		if timeToDraw < 0 {
			// Frame over-draw
			//log.Printf("Skipped %.5f frames due to over-draw", -numFramesToDraw)
			lastDrawTimestamp = time.Now()
		} else if timeToDraw >= time.Second {
			// Frame under-draw
			//log.Printf("Dropped %.5f frames due to lag", numFramesToDraw)
			lastDrawTimestamp = time.Now()
		} else if numFramesToDraw >= 1 {
			C.glfwPollEvents()
			drawAndPresentFrame()
		} else {
			time.Sleep(timeToDraw)
		}
	}
	C.ImGui_ImplOpenGL3_Shutdown()
	C.ImGui_ImplGlfw_Shutdown()
	imgui.DestroyContext()
	C.glfwDestroyWindow(glfwWindow)
	C.glfwTerminate()
}
