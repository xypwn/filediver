package main

/*
// GLFW
typedef struct GLFWwindow GLFWwindow;
typedef struct GLFWmonitor GLFWmonitor;
typedef struct GLFWvidmode {
	int width;
	int height;
	int redBits;
	int greenBits;
	int blueBits;
	int refreshRate;
} GLFWvidmode;
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
GLFWmonitor *glfwGetPrimaryMonitor();
GLFWvidmode *glfwGetVideoMode(GLFWmonitor *monitor);
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
	"maps"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/adrg/xdg"
	"github.com/ebitengine/oto/v3"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/ncruces/zenity"
	"github.com/skratchdot/open-golang/open"
	"github.com/xypwn/filediver/app/appconfig"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/getter"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
	"golang.design/x/clipboard"
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

//go:embed LICENSE
var license string

var baseGlyphRanges = [...]imgui.Wchar{
	0x0020, 0x00ff, // basic latin + supplement
	0x0100, 0x017f, // latin extended-A
	0x0400, 0x052f, // cyrillic + cyrillic supplement
	0,
}

var iconGlyphRanges = [...]imgui.Wchar{
	imgui.Wchar(fnt.IconFontInfo.Min),
	imgui.Wchar(fnt.IconFontInfo.Max16),
	0,
}

func UpdateGUIScale(guiScale float32, needCJKFonts bool) {
	io := imgui.CurrentIO()
	fonts := io.Fonts()
	fonts.Clear()

	style := imgui.NewStyle()

	fontSize := 16 * guiScale
	type fontSpec struct {
		scale       float32
		glyphRange  *imgui.Wchar
		ttfData     []byte
		extraConfig func(*imgui.FontConfig)
	}
	var fontSpecs []fontSpec
	// Base font
	fontSpecs = append(fontSpecs, fontSpec{
		scale:      1,
		glyphRange: &baseGlyphRanges[0],
		ttfData:    fnt.TextFont,
	})
	if needCJKFonts {
		fontSpecs = append(fontSpecs,
			// Japanese
			fontSpec{
				scale:      1.2,
				glyphRange: (&imgui.FontAtlas{}).GlyphRangesJapanese(),
				ttfData:    fnt.TextFontJP,
			},
			// Korean
			fontSpec{
				scale:      1.2,
				glyphRange: (&imgui.FontAtlas{}).GlyphRangesKorean(),
				ttfData:    fnt.TextFontKR,
			},
			// Chinese
			fontSpec{
				scale:      1.2,
				glyphRange: (&imgui.FontAtlas{}).GlyphRangesChineseFull(),
				ttfData:    fnt.TextFontCN,
			},
		)
	}
	// Icons
	fontSpecs = append(fontSpecs, fontSpec{
		scale:      1,
		glyphRange: &iconGlyphRanges[0],
		ttfData:    fnt.IconFont,
		extraConfig: func(fc *imgui.FontConfig) {
			fc.SetGlyphOffset(imgui.NewVec2(0, (0.2)*fontSize))
			fc.SetGlyphMinAdvanceX(1 * fontSize)
		},
	})
	for i, spec := range fontSpecs {
		fc := imgui.NewFontConfig()
		if i != 0 {
			fc.SetMergeMode(true)
		}
		fc.SetFontDataOwnedByAtlas(false)
		if spec.extraConfig != nil {
			spec.extraConfig(fc)
		}
		fonts.AddFontFromMemoryTTFV(
			uintptr(unsafe.Pointer(&spec.ttfData[0])),
			int32(len(spec.ttfData)),
			fontSize*spec.scale,
			fc,
			spec.glyphRange,
		)
		fc.Destroy()
	}

	C.ImGui_ImplOpenGL3_DestroyFontsTexture()

	io.SetFontGlobalScale(1)
	style.ScaleAllSizes(guiScale)
	io.Ctx().SetStyle(*style)
}

func run(onError func(error)) error {
	runtime.LockOSThread()

	var glfwWindow *C.GLFWwindow

	currentBackend, err := backend.CreateBackend(glfwbackend.NewGLFWBackend())
	if err != nil {
		return fmt.Errorf("creating backend: %w", err)
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
	currentBackend.SetWindowSizeLimits(250, 150, -1, -1)

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

	defaultPreferences := Preferences{
		AutoCheckForUpdates:            true,
		PreviewVideoVerticalResolution: 720,
	}
	{
		_, yScale := currentBackend.ContentScale()
		defaultPreferences.GUIScale = yScale

		monitor := C.glfwGetPrimaryMonitor()
		videoMode := C.glfwGetVideoMode(monitor)
		defaultPreferences.TargetFPS = float64(videoMode.refreshRate)
	}
	preferences := defaultPreferences
	preferencesPath := filepath.Join(xdg.DataHome, "filediver", "preferences.json")
	if err := preferences.Load(preferencesPath); err != nil {
		return fmt.Errorf("loading preferences: %w", err)
	}

	shouldUpdateGUIScale := true
	loadCJKFonts := false

	if err := gl.Init(); err != nil {
		return fmt.Errorf("initializing OpenGL: %w", err)
	}
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
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
		if err != nil {
			return fmt.Errorf("creating audio context: %w", err)
		}
		<-readyChan
	}
	if err := otoCtx.Err(); err != nil {
		return fmt.Errorf("audio context: %w", err)
	}

	ctx := context.Background()

	var gameDataLoad GameDataLoad
	var gameDataExport *GameDataExport
	var gameData *GameData

	gameDataLoad.GoLoadGameData(ctx, "")

	var previewState *widgets.FileAutoPreviewState
	defer func() {
		if previewState != nil {
			previewState.Delete()
		}
	}()
	var gameFileSearchQuery string
	filesSelectedForExport := map[stingray.FileID]struct{}{}
	allSelectedForExport := false
	var gameFileTypeSearchQuery string
	var gameFileTypes []widgets.FilterListSection[stingray.Hash]
	selectedGameFileTypes := map[stingray.Hash]struct{}{}
	var archiveIDSearchQuery string
	var archiveIDs []widgets.FilterListSection[stingray.Hash]
	selectedArchives := map[stingray.Hash]struct{}{}

	gameFileTypeDescriptions := map[stingray.Hash]string{
		stingray.Sum64([]byte("bik")):            "video",
		stingray.Sum64([]byte("wwise_bank")):     "audio bank",
		stingray.Sum64([]byte("wwise_stream")):   "loose audio",
		stingray.Sum64([]byte("texture")):        "image/texture",
		stingray.Sum64([]byte("unit")):           "3D model",
		stingray.Sum64([]byte("strings")):        "text table",
		stingray.Sum64([]byte("package")):        "file bundle",
		stingray.Sum64([]byte("bones")):          "unit bones",
		stingray.Sum64([]byte("physics")):        "unit physics",
		stingray.Sum64([]byte("geometry_group")): "group of 3D models",
	}
	gameFileTypeTooltipsExtra := map[stingray.Hash]string{
		stingray.Sum64([]byte("wwise_stream")): "All wwise_streams are also contained in a wwise_bank.\nYou probably want to use wwise_bank instead.",
		stingray.Sum64([]byte("package")):      "A package contains references to a bunch of other files.",
	}

	var checkUpdatesOnStartupBGDone bool
	var checkUpdatesOnStartupFGDone bool
	var checkingForUpdates bool
	var checkUpdatesNewVersion string // empty if unknown
	var checkUpdatesDownloadURL string
	var checkUpdatesErr error
	var checkUpdatesLock sync.Mutex

	goCheckForUpdates := func(setOnStartupBGDoneWhenDone bool) {
		checkUpdatesLock.Lock()
		defer checkUpdatesLock.Unlock()
		if checkingForUpdates {
			return
		}
		checkingForUpdates = true
		checkUpdatesNewVersion = ""
		checkUpdatesDownloadURL = ""
		checkUpdatesErr = nil
		go func() {
			ver, url, err := getNewestVersion()
			checkUpdatesLock.Lock()
			defer checkUpdatesLock.Unlock()
			checkUpdatesNewVersion, checkUpdatesDownloadURL, checkUpdatesErr = ver, url, err
			if setOnStartupBGDoneWhenDone {
				checkUpdatesOnStartupBGDone = true
			}
			checkingForUpdates = false
		}()
	}
	if preferences.AutoCheckForUpdates && version != "" {
		goCheckForUpdates(true)
	}

	exportDir := filepath.Join(xdg.UserDirs.Download, "filediver_exports")
	exportNotifyWhenDone := true
	var extractorConfig appconfig.Config
	config.InitDefault(&extractorConfig)
	prevExtractorConfig := extractorConfig
	var extractorConfigShowAdvanced bool

	logger := NewLogger()

	const ffmpegFeatures = `- Preview video
- Convert audio to OGG/AAC/MP3
- Convert video to MP4`

	const scriptsDistFeatures = `- Export models (units)/materials/geometry groups to .blend (Blender)`

	downloadsDir := filepath.Join(xdg.DataHome, "filediver")
	var ffmpegDownloadState *widgets.DownloaderState
	var scriptsDistDownloadState *widgets.DownloaderState
	{
		ffmpegTarget := getter.Target{
			SubdirName:        "ffmpeg",
			GHUser:            "BtbN",
			GHRepo:            "FFmpeg-Builds",
			PinnedVersion:     "latest",
			GHFilenameWindows: "ffmpeg-master-latest-win64-gpl.zip",
			GHFilenameLinux:   "ffmpeg-master-latest-linux64-gpl.tar.xz",
			StripFirstDir:     true,
		}
		ffmpegInfo, err := ffmpegTarget.GetInfo(false)
		if err != nil {
			onError(err)
		}
		scriptsDistTarget := getter.Target{
			SubdirName:        "filediver-scripts",
			GHUser:            "xypwn",
			GHRepo:            "filediver",
			PinnedVersion:     version,
			GHFilenameWindows: "scripts-dist-windows.zip",
			GHFilenameLinux:   "scripts-dist-linux.tar.xz",
			StripFirstDir:     true,
		}
		var scriptsDistInfo getter.Info
		if version != "" {
			var err error
			scriptsDistInfo, err = scriptsDistTarget.GetInfo(false)
			if err != nil {
				onError(err)
			}
		}

		ffmpegDownloadState = widgets.NewDownloader(downloadsDir, ffmpegInfo)
		scriptsDistDownloadState = widgets.NewDownloader(downloadsDir, scriptsDistInfo)
	}
	if err := detectAndMaybeDeleteLegacyExtensions(downloadsDir); err != nil {
		onError(err)
	}

	prevFfmpegDownloaded := ffmpegDownloadState.HaveRequestedVersion()
	prevScriptsDistDownloaded := scriptsDistDownloadState.HaveRequestedVersion()

	runner := exec.NewRunner()
	redetectRunnerProgs := func() {
		ffmpegPath := filepath.Join(ffmpegDownloadState.Dir(), "bin", "ffmpeg")
		ffmpegArgs := []string{"-y", "-hide_banner", "-loglevel", "error"}
		if !runner.Add(ffmpegPath, ffmpegArgs...) {
			// Try to use a local FFmpeg instance if the extension isn't installed
			runner.Add("ffmpeg", ffmpegArgs...)
		}
		ffprobePath := filepath.Join(ffmpegDownloadState.Dir(), "bin", "ffprobe")
		ffprobeArgs := []string{"-hide_banner", "-loglevel", "error"}
		if !runner.Add(ffprobePath, ffprobeArgs...) {
			// Try to use a local FFprobe instance if the extension isn't installed
			runner.Add("ffprobe", ffprobeArgs...)
		}
		blenderImporterPath := filepath.Join(scriptsDistDownloadState.Dir(), "hd2_accurate_blender_importer", "hd2_accurate_blender_importer")
		runner.Add(blenderImporterPath)
	}
	redetectRunnerProgs()

	popupManager := imutils.NewPopupManager()
	if version != "" {
		popupManager.Open["Some optional extensions missing or out of date"] =
			!ffmpegDownloadState.HaveRequestedVersion() || !scriptsDistDownloadState.HaveRequestedVersion()
		popupManager.Open["Welcome to Filediver GUI"] = true
	}

	isTypeFilterOpen := false
	isArchiveFilterOpen := false
	isLogsOpen := false

	preDraw := func() error {
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
				runner,
			)
			if err != nil {
				return fmt.Errorf("creating preview: %w", err)
			}
		}

		if shouldUpdateGUIScale {
			UpdateGUIScale(preferences.GUIScale, loadCJKFonts)
			shouldUpdateGUIScale = false
		}

		return nil
	}

	lastBrowserItemCopiedIndex := int32(-1)
	lastBrowserItemCopiedTime := -math.MaxFloat64

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
				imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.NewVec2(5, 5))
				if imgui.BeginMenu("Help") {
					imgui.Separator()
					if imgui.MenuItemBool(fnt.I("Info") + " About") {
						popupManager.Open["About"] = true
					}
					imgui.Separator()
					imgui.EndMenu()
				}
				if imgui.BeginMenu("Settings") {
					imgui.Separator()
					if imgui.MenuItemBool(fnt.I("Extension") + " Extensions") {
						popupManager.Open["Extensions"] = true
					}
					imgui.Separator()
					if imgui.MenuItemBool(fnt.I("Sync") + " Check for updates") {
						goCheckForUpdates(false)
						popupManager.Open["Check for updates"] = true
					}
					imgui.Separator()
					if imgui.MenuItemBool(fnt.I("Settings") + " Preferences") {
						popupManager.Open["Preferences"] = true
					}
					imgui.Separator()
					imgui.EndMenu()
				}
				imgui.EndMenuBar()
				imgui.PopStyleVar()
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
				var leftID, topLeftID, bottomLeftID, rightID imgui.ID
				imgui.InternalDockBuilderSplitNode(id, imgui.DirLeft, 0.5, &leftID, &rightID)
				imgui.InternalDockBuilderSplitNode(leftID, imgui.DirDown, 0.4, &bottomLeftID, &topLeftID)
				imgui.InternalDockBuilderDockWindow(fnt.I("View_list")+" Browser", topLeftID)
				imgui.InternalDockBuilderDockWindow(fnt.I("File_export")+" Export", bottomLeftID)
				imgui.InternalDockBuilderDockWindow(fnt.I("Settings_applications")+" Extractor config", bottomLeftID)
				imgui.InternalDockBuilderDockWindow(fnt.I("Preview")+" Preview", rightID)
				imgui.InternalDockBuilderFinish(id)
			}
			imgui.DockSpaceV(id, imgui.NewVec2(0, 0), 0, winClass)
		}
		imgui.End()
		imgui.PopStyleVar()

		// We use state tracking for allSelectedForExport
		// since recalculating each frame would eat up
		// ~5ms (on a good PC!) for nothing.
		// This function should be used whenever
		// the status of all items being selected may have
		// changed.
		calcAllSelectedForExport := func() bool {
			if gameData == nil || len(gameData.SortedSearchResultFileIDs) == 0 {
				return false
			}
			for _, id := range gameData.SortedSearchResultFileIDs {
				_, sel := filesSelectedForExport[id]
				if !sel {
					return false
				}
			}
			return true
		}

		if imgui.Begin(fnt.I("View_list") + " Browser") {
			if gameData == nil {
				gameDataLoad.Lock()
				if gameDataLoad.Done {
					if gameDataLoad.Err == nil {
						if gameData == nil {
							gameData = gameDataLoad.Result
							types := make(map[stingray.Hash]struct{})
							for _, f := range gameData.DataDir.Files {
								types[f.ID().Type] = struct{}{}
							}
							gameFileTypes = []widgets.FilterListSection[stingray.Hash]{
								{Title: "Previewable and exportable"},
								{Title: "Just exportable"},
								{Title: "Not exportable"},
							}
							for _, typ := range slices.SortedFunc(maps.Keys(types), func(h1, h2 stingray.Hash) int {
								return strings.Compare(gameData.LookupHash(h1), gameData.LookupHash(h2))
							}) {
								var sectionIdx int
								switch typ {
								case // previewable and exportable
									stingray.Sum64([]byte("bik")),
									stingray.Sum64([]byte("texture")),
									stingray.Sum64([]byte("wwise_bank")),
									stingray.Sum64([]byte("wwise_stream")),
									stingray.Sum64([]byte("unit")),
									stingray.Sum64([]byte("strings")):
									sectionIdx = 0
								default:
									typName := gameData.LookupHash(typ)
									if appconfig.Extractable[typName] {
										// just exportable
										sectionIdx = 1
									} else {
										// not exportable
										sectionIdx = 2
									}
								}
								gameFileTypes[sectionIdx].Items = append(gameFileTypes[sectionIdx].Items, typ)
							}
							archiveIDs = []widgets.FilterListSection[stingray.Hash]{
								{
									Items: slices.SortedFunc(maps.Keys(gameData.DataDir.FilesByTriad), func(h1, h2 stingray.Hash) int {
										return strings.Compare(gameData.LookupHash(h1), gameData.LookupHash(h2))
									}),
								},
							}
						}
					} else {
						imutils.TextError(gameDataLoad.Err)
					}
				} else {
					progress := gameDataLoad.Progress
					if progress != 1 {
						imutils.Textf(fnt.I("Hourglass_top") + " Loading game data...")
					} else {
						imutils.Textf(fnt.I("Hourglass_top") + " Processing game data...")
						progress = -float32(imgui.Time())
					}
					imgui.ProgressBar(progress)
				}
				gameDataLoad.Unlock()
			} else {
				var activeFileID stingray.FileID
				if previewState != nil {
					activeFileID = previewState.ActiveID()
				}

				imgui.SetNextItemWidth(-math.SmallestNonzeroFloat32)
				if imgui.Shortcut(imgui.KeyChord(imgui.ModCtrl | imgui.KeyF)) {
					imgui.SetKeyboardFocusHere()
				}
				if imgui.InputTextWithHint("##SearchName", fnt.I("Search")+" Filter by file name...", &gameFileSearchQuery, 0, nil) {
					gameData.UpdateSearchQuery(gameFileSearchQuery, selectedGameFileTypes, selectedArchives)
					allSelectedForExport = calcAllSelectedForExport()
				}
				imgui.SetItemTooltip("Filter by file name (Ctrl+F)")

				widgets.FilterListButton("Types", &isTypeFilterOpen, selectedGameFileTypes)
				imgui.SameLine()
				widgets.FilterListButton("Archives", &isArchiveFilterOpen, selectedArchives)

				imgui.BeginDisabledV(len(gameData.SortedSearchResultFileIDs) == 0)
				imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsMixedValue),
					!allSelectedForExport && len(filesSelectedForExport) > 0)
				if imgui.Checkbox("Select all for export", &allSelectedForExport) {
					if allSelectedForExport {
						for _, id := range gameData.SortedSearchResultFileIDs {
							filesSelectedForExport[id] = struct{}{}
						}
					} else {
						for _, id := range gameData.SortedSearchResultFileIDs {
							delete(filesSelectedForExport, id)
						}
					}
				}
				imgui.PopItemFlag()
				imgui.EndDisabled()
				if len(gameData.SortedSearchResultFileIDs) != 0 {
					if allSelectedForExport {
						imgui.SetItemTooltip("Deselect all currently visible files for export")
					} else {
						imgui.SetItemTooltip("Select all currently visible files for export")
					}
				}
				{
					size := imgui.ContentRegionAvail()
					size.Y -= imgui.TextLineHeightWithSpacing()
					imgui.SetNextWindowSize(size)
				}
				const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY | imgui.TableFlagsRowBg
				if imgui.BeginTableV("##Game Files", 3, tableFlags, imgui.NewVec2(0, 0), 0) {
					imgui.TableSetupColumnV(fnt.I("File_export"), imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoResize, imgui.FrameHeight(), 0)
					imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 3, 0)
					imgui.TableSetupColumnV("Type", imgui.TableColumnFlagsWidthStretch, 1, 0)
					imgui.TableSetupScrollFreeze(0, 1)
					imgui.TableHeadersRow()

					clipper := imgui.NewListClipper()
					clipper.Begin(int32(len(gameData.SortedSearchResultFileIDs)))
					for clipper.Step() {
						for row := clipper.DisplayStart(); row < clipper.DisplayEnd(); row++ {
							id := gameData.SortedSearchResultFileIDs[row]
							imgui.PushIDStr(id.Name.String() + id.Type.String())

							// Kind of hacky way to not add to checkbox's tooltip
							ttExport := false

							toggleExport := func(id stingray.FileID) {
								_, export := filesSelectedForExport[id]
								if export {
									delete(filesSelectedForExport, id)
								} else {
									filesSelectedForExport[id] = struct{}{}
								}
								allSelectedForExport = calcAllSelectedForExport()
							}

							imgui.TableNextColumn()
							imgui.PushItemFlag(imgui.ItemFlagsNoNav, true)
							_, export := filesSelectedForExport[id]
							if imgui.Checkbox("", &export) {
								toggleExport(id)
							}
							imgui.PopItemFlag()
							if imgui.IsItemHovered() {
								ttExport = true
								if export {
									imgui.SetTooltip("Deselect for export")
								} else {
									imgui.SetTooltip("Select for export")
								}
							}

							imgui.TableNextColumn()
							selected := imgui.SelectableBoolV(
								gameData.LookupHash(id.Name),
								id == activeFileID,
								imgui.SelectableFlagsSpanAllColumns|imgui.SelectableFlags(imgui.SelectableFlagsSelectOnNav),
								imgui.NewVec2(0, 0),
							)
							copied := false
							if id == activeFileID && imgui.Shortcut(imgui.KeyChord(imgui.ModCtrl|imgui.KeyC)) {
								copied = true
							}
							if id == activeFileID && imgui.Shortcut(imgui.KeyChord(imgui.KeySpace)) {
								toggleExport(id)
							}
							if imgui.IsItemClickedV(imgui.MouseButtonRight) {
								copied = true
							}
							if copied {
								lastBrowserItemCopiedIndex = row
								lastBrowserItemCopiedTime = imgui.Time()
								imgui.SetClipboardText(id.Name.String())
							}
							if !ttExport && imgui.IsItemHovered() {
								var text strings.Builder
								if lastBrowserItemCopiedIndex == row &&
									imgui.Time()-lastBrowserItemCopiedTime < 1 {
									fmt.Fprintf(&text, "%v Copied '%v'\n", fnt.I("Check"), id.Name.String())
								} else {
									if name, ok := gameData.Hashes[id.Name]; ok {
										fmt.Fprintf(&text, "Name: %v, hash=%v\n", name, id.Name)
									} else {
										fmt.Fprintf(&text, "Name: hash=%v\n", id.Name)
									}
									fmt.Fprintf(&text, "Type: %v, hash=%v\n", gameData.LookupHash(id.Type), id.Type)
									fmt.Fprintf(&text, "%v Left-click to preview\n", fnt.I("Preview"))
									fmt.Fprintf(&text, "%v Right-click or Ctrl+C to copy name hash to clipboard\n", fnt.I("Content_copy"))
									fmt.Fprintf(&text, "%v Down/up arrow keys to select next/previous\n", fnt.I("Unfold_more"))
									fmt.Fprintf(&text, "%v Use the checkbox or space key to select/deselect for export\n", fnt.I("File_export"))
								}
								imgui.SetTooltip(text.String())
							}
							imgui.TableNextColumn()

							imgui.TextUnformatted(gameData.LookupHash(id.Type))
							imgui.PopID()

							if selected {
								previewState.LoadFile(ctx, gameData.DataDir.Files[id], preferences.PreviewVideoVerticalResolution)
							}
						}
					}

					imgui.EndTable()
				}
				imutils.Textf("Showing %v/%v files (%v selected for export)", len(gameData.SortedSearchResultFileIDs), len(gameData.DataDir.Files), len(filesSelectedForExport))
			}
		}
		imgui.End()

		imgui.SetNextWindowSizeV(imgui.NewVec2(300, 600), imgui.CondOnce)
		if widgets.FilterListWindow("Type Filter",
			&isTypeFilterOpen,
			"Search Types",
			&gameFileTypeSearchQuery,
			gameFileTypes,
			&selectedGameFileTypes,
			func(x stingray.Hash) string {
				s := gameData.LookupHash(x)
				if desc, ok := gameFileTypeDescriptions[x]; ok {
					s += " (" + desc + ")"
				}
				if _, ok := gameFileTypeTooltipsExtra[x]; ok {
					s += " " + fnt.I("Info")
				}
				return s
			},
			func(x stingray.Hash) string {
				s := fmt.Sprintf("hash=%v", x)
				if ttExtra, ok := gameFileTypeTooltipsExtra[x]; ok {
					s += "\n" + ttExtra
				}
				return s
			},
		) {
			gameData.UpdateSearchQuery(gameFileSearchQuery, selectedGameFileTypes, selectedArchives)
			allSelectedForExport = calcAllSelectedForExport()
		}

		imgui.SetNextWindowSizeV(imgui.NewVec2(300, 600), imgui.CondOnce)
		if widgets.FilterListWindow("Archive Filter",
			&isArchiveFilterOpen,
			"Search Archives",
			&archiveIDSearchQuery,
			archiveIDs,
			&selectedArchives,
			func(x stingray.Hash) string {
				return gameData.LookupHash(x)
			},
			func(x stingray.Hash) string {
				return fmt.Sprintf("hash=%v", x)
			},
		) {
			gameData.UpdateSearchQuery(gameFileSearchQuery, selectedGameFileTypes, selectedArchives)
			allSelectedForExport = calcAllSelectedForExport()
		}

		if imgui.Begin(fnt.I("File_export") + " Export") {
			imgui.BeginDisabledV(gameDataExport != nil)
			imutils.FilePicker("Output directory", &exportDir, true)
			imgui.EndDisabled()

			imgui.Checkbox(fnt.I("Notifications")+" Notify when done", &exportNotifyWhenDone)

			imgui.Separator()

			{
				imgui.PushIDStr("Open/close extractor logs")
				var numErrsWarnsStr string
				if logger.NumErrs() > 0 || logger.NumWarns() > 0 {
					var items []string
					if logger.NumErrs() > 0 {
						items = append(items, fmt.Sprintf("%v %v", logger.NumErrs(), fnt.I("Error")))
					}
					if logger.NumWarns() > 0 {
						items = append(items, fmt.Sprintf("%v %v", logger.NumWarns(), fnt.I("Warning")))
					}
					numErrsWarnsStr = fmt.Sprintf(" (%v)", strings.Join(items, ","))
				}
				var label string
				if isLogsOpen {
					label = fmt.Sprintf("%v Close extractor logs%v %v", fnt.I("List"), numErrsWarnsStr, fnt.I("Close"))
				} else {
					label = fmt.Sprintf("%v Open extractor logs%v %v", fnt.I("List"), numErrsWarnsStr, fnt.I("Open_in_new"))
				}
				if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
					isLogsOpen = !isLogsOpen
				}
				imgui.PopID()
			}

			if imgui.ButtonV(fnt.I("Folder_open")+" Open output folder", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				_ = os.MkdirAll(exportDir, os.ModePerm)
				open.Start(exportDir)
			}

			imgui.Separator()

			if gameDataExport == nil {
				imgui.PushIDStr("Begin export button")
				imgui.BeginDisabledV(len(filesSelectedForExport) == 0 || gameData == nil)
				label := fmt.Sprintf("%v Begin export (%v)", fnt.I("File_export"), len(filesSelectedForExport))
				if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) && gameData != nil {
					logger.Reset()
					gameDataExport = gameData.GoExport(
						ctx,
						slices.Collect(maps.Keys(filesSelectedForExport)),
						exportDir,
						extractorConfig,
						runner,
						logger,
					)
				}
				if gameData == nil {
					imgui.SetItemTooltip("Game data not loaded")
				} else if len(filesSelectedForExport) == 0 {
					imgui.SetItemTooltip("Nothing selected for export")
				}
				imgui.EndDisabled()
				imgui.PopID()
			} else {
				if gameDataExport.Done {
					if !gameDataExport.Canceled && exportNotifyWhenDone {
						pluralS := ""
						if gameDataExport.NumFiles != 1 {
							pluralS = "s"
						}
						var text string
						if !logger.HaveFatalErr() {
							text = fmt.Sprintf("Filediver has finished exporting %v file%v", gameDataExport.NumFiles, pluralS)
						} else {
							text = "An internal error occurred during exporting. Please create an issue on GitHub."
						}
						if logger.NumErrs() > 0 || logger.NumWarns() > 0 {
							text += "\n"
							text += fmt.Sprintf("Errors: %v, Warnings: %v", logger.NumErrs(), logger.NumWarns())
							text += "\nSee logs."
						}
						if logger.NumErrs() > 0 {
							isLogsOpen = true
						}
						zenity.Notify(text,
							zenity.Title("Finished exporting"),
							zenity.InfoIcon,
						)
					}
					gameDataExport = nil
				} else {
					imgui.ProgressBarV(
						float32(gameDataExport.CurrentFileIndex)/float32(gameDataExport.NumFiles),
						imgui.NewVec2(-math.SmallestNonzeroFloat32, 0),
						fmt.Sprintf("%v/%v", gameDataExport.CurrentFileIndex+1, gameDataExport.NumFiles),
					)
					imutils.Textf("%v", gameDataExport.CurrentFileName)
					if imgui.ButtonV(fnt.I("Cancel")+" Cancel export", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
						gameDataExport.Cancel()
					}
				}
			}
		}
		imgui.End()

		if imgui.Begin(fnt.I("Settings_applications") + " Extractor config") {
			if widgets.ConfigEditor(&extractorConfig, &extractorConfigShowAdvanced) {
				if extractorConfig.Gamedir != prevExtractorConfig.Gamedir {
					gameData = nil
					gameDataLoad.GoLoadGameData(ctx, extractorConfig.Gamedir)
				}
				prevExtractorConfig = extractorConfig
			}
		}
		imgui.End()

		if isLogsOpen {
			imgui.SetNextWindowPosV(viewport.Center(), imgui.CondOnce, imgui.NewVec2(0.5, 0.5))
			imgui.SetNextWindowSizeV(imgui.NewVec2(400, 400), imgui.CondOnce)
			if imgui.BeginV("Extractor logs", &isLogsOpen, 0) {
				avail := imgui.ContentRegionAvail()
				avail.Y -= imgui.FrameHeightWithSpacing()
				imgui.SetNextWindowSize(avail)
				if imgui.BeginChildStr("Log view") {
					LogView(logger)
				}
				imgui.EndChild()

				ctx := imgui.CurrentContext()
				btnLabel := fnt.I("Content_copy") + " Copy all to clipboard"
				btnID := imgui.IDStr(btnLabel)
				if imgui.ButtonV(btnLabel, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
					imgui.SetClipboardText(logger.String())
				}
				if ctx.LastActiveId() == btnID && ctx.LastActiveIdTimer() < 1 {
					imgui.SetItemTooltip(fnt.I("Check") + " Copied")
				}
			}
			imgui.End()
		}

		if imgui.Begin(fnt.I("Preview") + " Preview") {
			if previewState != nil {
				if !widgets.FileAutoPreview("Preview", previewState) {
					active := previewState.ActiveID()
					imgui.PushTextWrapPos()
					if active.Name.Value == 0 {
						imgui.TextUnformatted("Nothing selected to preview. Select an item from the Browser.")
					} else if gameData != nil {
						imutils.Textf("Cannot preview type %v.", gameData.LookupHash(active.Type))
					}
				}
				if previewState.NeedCJKFont() && !loadCJKFonts {
					loadCJKFonts = true
					shouldUpdateGUIScale = true
				}
			} else {
				imgui.PushTextWrapPos()
				imgui.TextUnformatted("Loading game data...")
			}
		}
		imgui.End()

		checkUpdatesLock.Lock()
		if checkUpdatesOnStartupBGDone && !checkUpdatesOnStartupFGDone {
			if checkUpdatesErr == nil && checkUpdatesNewVersion != version {
				popupManager.Open["Check for updates"] = true
			}
			checkUpdatesOnStartupFGDone = true
		}
		checkUpdatesLock.Unlock()

		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		popupManager.Popup("Check for updates", func(close func()) {
			checkUpdatesLock.Lock()
			if checkUpdatesErr == nil {
				if checkingForUpdates {
					imgui.ProgressBarV(-1*float32(imgui.Time()), imgui.NewVec2(250, 0), "Checking for updates")
				} else if checkUpdatesNewVersion == version {
					imutils.Textcf(imgui.NewVec4(0, 0.7, 0, 1), fnt.I("Check")+"Up-to-date (version %v)", version)
				} else {
					imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0, 1), fnt.I("Exclamation")+"New version available: %v", checkUpdatesNewVersion)
					if imgui.ButtonV(fnt.I("Download")+" Download "+checkUpdatesNewVersion, imgui.NewVec2(250, 0)) {
						open.Start(checkUpdatesDownloadURL)
					}
					imgui.SetItemTooltip("Open '" + checkUpdatesDownloadURL + "'")
				}
			} else {
				imutils.TextError(checkUpdatesErr)
			}
			checkUpdatesLock.Unlock()
			imgui.Separator()
			if imgui.ButtonV("Close", imgui.NewVec2(250, 0)) {
				close()
			}
		}, imgui.WindowFlagsAlwaysAutoResize, true)

		imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		popupManager.Popup("Welcome to Filediver GUI", func(close func()) {
			imgui.PushTextWrapPos()
			imgui.TextUnformatted("Filediver GUI is still work-in-progress.")
			imgui.TextUnformatted("If you encounter any bugs, please create an issue on the project's GitHub page.")
			imgui.Separator()
			imgui.TextLinkOpenURLV("Filediver Issues", "https://github.com/xypwn/filediver/issues")
			imgui.Separator()
			if imgui.ButtonV("OK", imgui.NewVec2(400, 0)) {
				close()
			}
		}, imgui.WindowFlagsAlwaysAutoResize, true)

		imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		popupManager.Popup("Some optional extensions missing or out of date", func(close func()) {
			imgui.PushTextWrapPos()
			imgui.TextUnformatted("The following recommended extensions are missing or out of date:")
			if !ffmpegDownloadState.HaveRequestedVersion() {
				imgui.TextWrapped("FFmpeg")
				imgui.Indent()
				imgui.TextWrapped(ffmpegFeatures)
				imgui.Unindent()
			}
			if !scriptsDistDownloadState.HaveRequestedVersion() {
				imgui.TextWrapped("ScriptsDist")
				imgui.Indent()
				imgui.TextWrapped(scriptsDistFeatures)
				imgui.Unindent()
			}
			imgui.TextUnformatted("You can download these extensions by clicking \"Manage extensions\", or by going to Settings->Extensions.")
			imgui.Separator()
			if imgui.ButtonV("Close", imgui.NewVec2(120, 0)) {
				close()
			}
			imgui.SameLine()
			if imgui.ButtonV("Manage extensions", imgui.NewVec2(160, 0)) {
				close()
				popupManager.Open["Extensions"] = true
			}
		}, imgui.WindowFlagsAlwaysAutoResize, true)

		if have := ffmpegDownloadState.HaveRequestedVersion(); prevFfmpegDownloaded != have {
			redetectRunnerProgs()
			prevFfmpegDownloaded = have
		}
		if have := scriptsDistDownloadState.HaveRequestedVersion(); prevScriptsDistDownloaded != have {
			redetectRunnerProgs()
			prevScriptsDistDownloaded = have
		}

		imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		popupManager.Popup("Extensions", func(close func()) {
			widgets.Downloader("FFmpeg", ffmpegFeatures, ffmpegDownloadState)
			imgui.Separator()
			widgets.Downloader("ScriptsDist", scriptsDistFeatures, scriptsDistDownloadState)
			imgui.Separator()
			if imgui.ButtonV(fnt.I("Folder_open")+" Open extensions folder", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				_ = os.MkdirAll(downloadsDir, os.ModePerm)
				open.Start(downloadsDir)
			}
			imgui.Separator()
			if imgui.ButtonV("Close", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				close()
			}
		}, imgui.WindowFlagsAlwaysAutoResize, true)

		imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		popupManager.Popup("Preferences", func(close func()) {
			prevPrefs := preferences
			if imutils.ComboChoice(
				"GUI Scale",
				&preferences.GUIScale,
				[]float32{0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2, 2.25, 2.5, 2.75, 3},
			) {
				shouldUpdateGUIScale = true
			}
			imutils.ComboChoice(
				"Target FPS",
				&preferences.TargetFPS,
				[]float64{15, 30, 60, 75, 90, 120, 144, 165, 244, 300},
			)
			imutils.ComboChoiceAny(
				"Video Preview Resolution",
				&preferences.PreviewVideoVerticalResolution,
				[]int{160, 240, 360, 480, 720, 1080, 1440, 2160},
				func(a, b int) bool { return a == b },
				func(x int) string { return fmt.Sprintf("%vp", x) },
			)
			imgui.SetItemTooltip(fnt.I("Info") + ` The preview player isn't very well optimized, so higher
resolutions may cause low frame rates and poor responsiveness.`)
			imgui.Checkbox("Check for updates on start", &preferences.AutoCheckForUpdates)

			imgui.BeginDisabledV(preferences == defaultPreferences)
			if imgui.ButtonV(fnt.I("Undo")+" Reset all", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				preferences = defaultPreferences
			}
			imgui.EndDisabled()

			if preferences != prevPrefs {
				shouldUpdateGUIScale = preferences.GUIScale != prevPrefs.GUIScale
				if err := preferences.Save(preferencesPath); err != nil {
					onError(err)
				}
			}

			imgui.Separator()
			if imgui.ButtonV(fnt.I("Folder_open")+" Open preferences folder", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				dir := filepath.Dir(preferencesPath)
				_ = os.MkdirAll(dir, os.ModePerm)
				open.Start(dir)
			}
			imgui.Separator()
			if imgui.ButtonV("Close", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				close()
			}
		}, imgui.WindowFlagsAlwaysAutoResize, true)

		imgui.SetNextWindowSizeV(imgui.NewVec2(500, 400), imgui.CondOnce)
		imgui.SetNextWindowSizeConstraints(imgui.NewVec2(300, 200), viewport.Size())
		imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		popupManager.Popup("About", func(close func()) {
			imgui.TextUnformatted("Filediver GUI")
			imgui.SameLine()
			imgui.TextLinkOpenURLV("(GitHub)", "https://github.com/xypwn/filediver")
			if version != "" {
				imutils.Textf("version: %v", version)
			} else {
				imgui.TextUnformatted("development version")
			}
			imgui.Separator()
			if imgui.CollapsingHeaderBoolPtr("License", nil) {
				imgui.PushTextWrapPos()
				imgui.TextUnformatted(license)
			}
			imgui.Separator()
			if imgui.CollapsingHeaderBoolPtr("Font Licenses", nil) {
				imgui.Indent()
				if imgui.CollapsingHeaderBoolPtr("Noto", nil) {
					imgui.PushTextWrapPos()
					imgui.TextUnformatted(fnt.TextFontLicense)
				}
				if imgui.CollapsingHeaderBoolPtr("Material Symbols", nil) {
					imgui.PushTextWrapPos()
					imgui.TextUnformatted(fnt.IconFontLicense)
				}
				imgui.Unindent()
			}
			imgui.Separator()
			if imgui.ButtonV("Close", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
				close()
			}
		}, imgui.WindowFlagsNoMove, true)
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

		targetFrameTime := time.Duration(float64(time.Second) / preferences.TargetFPS)
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
		if err := preDraw(); err != nil {
			return err
		}

		timeToDraw := time.Now().Sub(lastDrawTimestamp)
		numFramesToDraw := timeToDraw.Seconds() * preferences.TargetFPS
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

	return nil
}

func main() {
	if version == "" || version == "v0.0.0" {
		fmt.Println(`Development version detected.
To use blender exporter, please pass a valid version to the build (this is because filediver has to know which version of Blender exporter it wants).
You can do this via 'go run -ldflags "-X main.version=v0.0.0" ./cmd/filediver-gui' (replace v0.0.0 with a real version).`)
	}

	clipboardOk := clipboard.Init() == nil

	onError := func(err error) {
		text := "Error: " + err.Error()
		if clipboardOk {
			clipboard.Write(clipboard.FmtText, []byte(text))
			text += "\n\n(error copied to clipboard)"
		}
		zenity.Info(text,
			zenity.Title("Internal Filediver error"),
			zenity.ErrorIcon,
		)
	}

	defer func() {
		if err := recover(); err != nil {
			onError(fmt.Errorf("panic: %v\nstack trace: %s", err, debug.Stack()))
		}
	}()

	if err := run(onError); err != nil {
		onError(err)
	}
}
