package main

/*
void ImGui_ImplOpenGL3_DestroyFontsTexture();

typedef struct GLFWwindow GLFWwindow;
GLFWwindow *glfwGetCurrentContext(void);
void glfwMakeContextCurrent(GLFWwindow *window);
*/
import "C"

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v3.2-core/gl"
	icon_fonts "github.com/juliettef/IconFontCppHeaders"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/stingray"
)

//go:embed fonts/Roboto-Regular.ttf
var TextFont []byte

//go:embed fonts/MaterialSymbolsOutlined.ttf
var IconsFont []byte

var IconsFontInfo = icon_fonts.IconsMaterialSymbols
var Icons = IconsFontInfo.Icons
var IconsFontRanges = [3]imgui.Wchar{
	imgui.Wchar(IconsFontInfo.Min),
	imgui.Wchar(IconsFontInfo.Max16),
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
			uintptr(unsafe.Pointer(&TextFont[0])),
			int32(len(TextFont)),
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
			uintptr(unsafe.Pointer(&IconsFont[0])),
			int32(len(IconsFont)),
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

	currentBackend.SetWindowFlags(glfwbackend.GLFWWindowFlagsResizable, 1)
	currentBackend.CreateWindow("Filediver GUI", 800, 700)

	// HACK: Window creation resets the GLFW context for some reason, so we restore it here
	C.glfwMakeContextCurrent(glfwWindow)

	currentBackend.SetBgColor(imgui.NewVec4(0.2, 0.2, 0.2, 1))
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

	var targetFPS uint = 60
	shouldUpdateTargetFPS := true

	currentBackend.SetBeforeRenderHook(func() {
		if shouldUpdateGUIScale {
			UpdateGUIScale(guiScale)
			shouldUpdateGUIScale = false
		}
		if shouldUpdateTargetFPS {
			currentBackend.SetTargetFPS(targetFPS)
			shouldUpdateTargetFPS = false
		}
	})

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

	ctx := context.Background()

	var gameDataLoad GameDataLoad
	var gameData *GameData

	gameDataLoad.GoLoadGameData(ctx)

	unitPreviewState, err := widgets.CreateUnitPreview()
	if err != nil {
		log.Fatal("Error creating unit preview:", err)
	}
	defer func() {
		if unitPreviewState != nil {
			unitPreviewState.Delete()
		}
	}()
	loadUnit := func(fileID stingray.FileID) {
		if gameData == nil {
			return
		}
		file, ok := gameData.DataDir.Files[fileID]
		if !ok {
			log.Println("Unit file does not exist")
			return
		}
		fMain, err := file.Open(ctx, stingray.DataMain)
		if err != nil {
			log.Println("Error opening unit file:", err)
			return
		}
		fGPU, err := file.Open(ctx, stingray.DataGPU)
		if err != nil {
			log.Println("Error opening unit file:", err)
			return
		}
		if err := unitPreviewState.LoadUnit(fMain, fGPU); err != nil {
			log.Println("Error loading unit:", err)
			return
		}
	}

	var gameFileSearchQuery string

	isPreferencesOpen := false

	currentBackend.Run(func() {
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
					imgui.MenuItemBool(Icons["Help"] + " Tutorial")
					imgui.MenuItemBool(Icons["Info"] + " About")
					imgui.EndMenu()
				}
				if imgui.BeginMenu("Settings") {
					if imgui.MenuItemBool(Icons["Settings"] + " Preferences") {
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
					imgui.TextUnformatted(Icons["Hourglass_top"] + " Loading game data...")
					imgui.ProgressBar(gameDataLoad.Progress)
				}
				gameDataLoad.Unlock()
			} else {
				if imgui.InputTextWithHint("##Search", Icons["Search"]+" Search...", &gameFileSearchQuery, 0, nil) {
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
								if id.Type == stingray.Sum64([]byte("unit")) {
									loadUnit(id)
								}
							}
						}
					}

					imgui.EndTable()
				}
			}
		}
		imgui.End()

		previewWindowFlags := imgui.WindowFlags(0)
		if unitPreviewState.IsUsing {
			previewWindowFlags |= imgui.WindowFlagsNoMove
		}
		if imgui.BeginV("Preview", nil, previewWindowFlags) {
			widgets.UnitPreview("Unit Preview", unitPreviewState)
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
					targetFPS = uint(v)
					shouldUpdateTargetFPS = true
				},
			)

			if imgui.ButtonV("Close", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				imgui.CloseCurrentPopup()
				isPreferencesOpen = false
			}
			imgui.EndPopup()
		}
	})
}
