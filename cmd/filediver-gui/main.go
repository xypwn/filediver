package main

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/adrg/xdg"
	"github.com/ebitengine/oto/v3"
	"github.com/ncruces/zenity"
	"github.com/skratchdot/open-golang/open"
	"github.com/xypwn/filediver/app/appconfig"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/getter"
	"github.com/xypwn/filediver/cmd/filediver-gui/imgui_wrapper"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets/previews"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
	"golang.design/x/clipboard"
)

//go:embed LICENSE
var license string

var (
	gameFileTypeDescriptions = map[stingray.Hash]string{
		stingray.Sum("bik"):            "video",
		stingray.Sum("wwise_bank"):     "audio bank",
		stingray.Sum("wwise_stream"):   "loose audio",
		stingray.Sum("texture"):        "image/texture",
		stingray.Sum("unit"):           "3D model",
		stingray.Sum("strings"):        "text table",
		stingray.Sum("package"):        "file bundle",
		stingray.Sum("bones"):          "unit bones",
		stingray.Sum("physics"):        "unit physics",
		stingray.Sum("geometry_group"): "group of 3D models",
		stingray.Sum("material"):       "shader settings",
	}
	gameFileTypeTooltipsExtra = map[stingray.Hash]string{
		stingray.Sum("wwise_stream"): "All wwise_streams are also contained in a wwise_bank.\nYou probably want to use wwise_bank instead.",
		stingray.Sum("package"):      "A package contains references to a bunch of other files.",
	}
)

const ffmpegFeatures = `- Preview video
- Convert audio to OGG/AAC/MP3
- Convert video to MP4`

const scriptsDistFeatures = `- Export models (units/geometry_groups/prefabs) and materials to .blend (Blender)`

type guiApp struct {
	showErrorPopup func(error)

	ctx context.Context

	preferencesPath    string
	preferences        Preferences
	defaultPreferences Preferences

	otoCtx          *oto.Context
	audioSampleRate int

	gameDataLoad   GameDataLoad
	gameDataExport *GameDataExport
	gameData       *GameData

	previewState *previews.AutoPreviewState

	gameFileSearchQuery     string
	filesSelectedForExport  map[stingray.FileID]struct{}
	allSelectedForExport    bool
	gameFileTypeSearchQuery string
	gameFileTypes           []widgets.FilterListSection[stingray.Hash]
	selectedGameFileTypes   map[stingray.Hash]struct{}
	archiveIDSearchQuery    string
	archiveIDs              []widgets.FilterListSection[stingray.Hash]
	selectedArchives        map[stingray.Hash]struct{}
	historyStack            []stingray.FileID
	historyStackIndex       int

	checkUpdatesOnStartupBGDone bool
	checkUpdatesOnStartupFGDone bool
	checkingForUpdates          bool
	checkUpdatesNewVersion      string // empty if unknown
	checkUpdatesDownloadURL     string
	checkUpdatesErr             error
	checkUpdatesLock            sync.Mutex

	exportDir                   string
	exportNotifyWhenDone        bool
	extractorConfig             appconfig.Config
	extractorConfigShowAdvanced bool
	extractorConfigPath         string
	extractorConfigSearchQuery  string

	logger *Logger

	downloadsDir              string
	ffmpegDownloadState       *widgets.DownloaderState
	scriptsDistDownloadState  *widgets.DownloaderState
	prevFfmpegDownloaded      bool
	prevScriptsDistDownloaded bool

	runner *exec.Runner

	popupManager *imutils.PopupManager

	isTypeFilterOpen    bool
	isArchiveFilterOpen bool
	isLogsOpen          bool

	lastBrowserItemCopiedIndex int32
	lastBrowserItemCopiedTime  float64
}

// Call Delete() when done.
// showErrorPopup should show an error popup without
// changing control flow, i.e. without exiting.
func newGUIApp(showErrorPopup func(error)) *guiApp {
	var extractorConfig appconfig.Config
	config.InitDefault(&extractorConfig)

	return &guiApp{
		showErrorPopup:             showErrorPopup,
		ctx:                        context.Background(),
		preferencesPath:            filepath.Join(xdg.DataHome, "filediver", "preferences.json"),
		audioSampleRate:            48000,
		filesSelectedForExport:     map[stingray.FileID]struct{}{},
		selectedGameFileTypes:      map[stingray.Hash]struct{}{},
		selectedArchives:           map[stingray.Hash]struct{}{},
		exportDir:                  filepath.Join(xdg.UserDirs.Download, "filediver_exports"),
		exportNotifyWhenDone:       true,
		extractorConfig:            extractorConfig,
		extractorConfigPath:        filepath.Join(xdg.DataHome, "filediver", "extractor_config.json"),
		logger:                     NewLogger(),
		downloadsDir:               filepath.Join(xdg.DataHome, "filediver"),
		runner:                     exec.NewRunner(),
		popupManager:               imutils.NewPopupManager(),
		lastBrowserItemCopiedIndex: -1,
		lastBrowserItemCopiedTime:  -math.MaxFloat64,
	}
}

func (a *guiApp) Delete() {
	if a.previewState != nil {
		a.previewState.Delete()
	}
}

func (a *guiApp) onInitWindow(state *imgui_wrapper.State) error {
	{
		var readyChan chan struct{}
		var err error
		a.otoCtx, readyChan, err = oto.NewContext(&oto.NewContextOptions{
			SampleRate:   a.audioSampleRate,
			ChannelCount: 2,
			Format:       oto.FormatFloat32LE,
		})
		if err != nil {
			return fmt.Errorf("creating audio context: %w", err)
		}
		<-readyChan
	}
	if err := a.otoCtx.Err(); err != nil {
		return fmt.Errorf("audio context: %w", err)
	}

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
			a.showErrorPopup(err)
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
				a.showErrorPopup(err)
			}
		}

		a.ffmpegDownloadState = widgets.NewDownloader(a.downloadsDir, ffmpegInfo)
		a.scriptsDistDownloadState = widgets.NewDownloader(a.downloadsDir, scriptsDistInfo)
	}

	if err := detectAndMaybeDeleteLegacyExtensions(a.downloadsDir); err != nil {
		a.showErrorPopup(err)
	}

	a.prevFfmpegDownloaded = a.ffmpegDownloadState.HaveRequestedVersion()
	a.prevScriptsDistDownloaded = a.scriptsDistDownloadState.HaveRequestedVersion()

	a.redetectRunnerProgs()

	if version != "" {
		a.popupManager.Open["Missing or out of date extensions"] =
			!a.ffmpegDownloadState.HaveRequestedVersion() || !a.scriptsDistDownloadState.HaveRequestedVersion()
		a.popupManager.Open["Welcome to Filediver GUI"] = true
	}

	{
		a.defaultPreferences = Preferences{
			AutoCheckForUpdates:            true,
			PreviewVideoVerticalResolution: 720,
		}

		a.defaultPreferences.GUIScale = state.GUIScale
		a.defaultPreferences.TargetFPS = state.FrameRate

		a.preferences = a.defaultPreferences
		if err := a.preferences.Load(a.preferencesPath); err != nil {
			return fmt.Errorf("loading preferences: %w", err)
		}
		state.GUIScale = a.preferences.GUIScale
		state.FrameRate = a.preferences.TargetFPS
	}

	{
		if err := a.extractorConfig.Load(a.extractorConfigPath); err != nil {
			return fmt.Errorf("loading extractor config: %w", err)
		}
	}

	if a.preferences.AutoCheckForUpdates && version != "" {
		a.goCheckForUpdates(true)
	}

	a.gameDataLoad.GoLoadGameData(a.ctx, "")

	return nil
}

func (a *guiApp) onPreDraw(state *imgui_wrapper.State) error {
	imutils.GlobalScale = state.GUIScale
	if a.gameData != nil && a.previewState == nil {
		var err error
		a.previewState, err = previews.NewAutoPreview(
			a.otoCtx, a.audioSampleRate,
			a.gameData.Hashes,
			a.gameData.ThinHashes,
			func(id stingray.FileID, typ stingray.DataType) (data []byte, exists bool, err error) {
				data, err = a.gameData.DataDir.Read(id, typ)
				if err == stingray.ErrFileDataTypeNotExist {
					return nil, false, nil
				}
				if err != nil {
					return nil, true, err
				}
				return data, true, nil
			},
			a.runner,
		)
		if err != nil {
			return fmt.Errorf("creating preview: %w", err)
		}
	}
	return nil
}

func (a *guiApp) onDraw(state *imgui_wrapper.State) {
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.NewVec2(0, 0))

	menuBarHeight := a.drawMenuBar()

	// Set up dock space and arrange window nodes
	viewport := imgui.MainViewport()
	dockSpacePos := viewport.Pos()
	dockSpaceSize := viewport.Size()
	{
		dockSpacePos.Y += menuBarHeight
		dockSpaceSize.Y -= menuBarHeight
	}
	imgui.SetNextWindowPos(dockSpacePos)
	imgui.SetNextWindowSize(dockSpaceSize)
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

	a.drawBrowserWindow()
	a.drawTypeFilterWindow()
	a.drawArchiveFilterWindow()
	a.drawExportWindow()
	a.drawExtractorConfigWindow()
	a.drawLogWindow()
	a.drawPreviewWindow(state)

	// drawXXXPopup functions use an
	// imutils.PopupManager, meaning
	// the popups will appear in the
	// order they're drawn.
	a.drawCheckForUpdatesPopup()
	a.drawWelcomePopup()
	a.drawExtensionsWarningPopup()
	a.drawExtensionsPopup()
	a.drawPreferencesPopup(state)
	a.drawAboutPopup()
}

func (a *guiApp) goCheckForUpdates(isOnStartup bool) {
	a.checkUpdatesLock.Lock()
	defer a.checkUpdatesLock.Unlock()
	if a.checkingForUpdates {
		return
	}
	a.checkingForUpdates = true
	a.checkUpdatesNewVersion = ""
	a.checkUpdatesDownloadURL = ""
	a.checkUpdatesErr = nil
	go func() {
		ver, url, err := getNewestVersion()
		if err == nil { // check for 404 response -> specific binary not yet available -> build not done yet
			resp, rErr := http.Get(url)
			if rErr != nil {
				err = rErr
			} else {
				resp.Body.Close()
				if resp.StatusCode == 404 {
					ver, url = version, ""
				}
			}
		}

		a.checkUpdatesLock.Lock()
		if isOnStartup {
			a.checkUpdatesOnStartupBGDone = true
		}
		a.checkingForUpdates = false
		a.checkUpdatesNewVersion, a.checkUpdatesDownloadURL, a.checkUpdatesErr = ver, url, err
		a.checkUpdatesLock.Unlock()
	}()
}

func (a *guiApp) redetectRunnerProgs() {
	ffmpegPath := filepath.Join(a.ffmpegDownloadState.Dir(), "bin", "ffmpeg")
	ffmpegArgs := []string{"-y", "-hide_banner", "-loglevel", "error"}
	if !a.runner.Add(ffmpegPath, ffmpegArgs...) {
		// Try to use a local FFmpeg instance if the extension isn't installed
		a.runner.Add("ffmpeg", ffmpegArgs...)
	}
	ffprobePath := filepath.Join(a.ffmpegDownloadState.Dir(), "bin", "ffprobe")
	ffprobeArgs := []string{"-hide_banner", "-loglevel", "error"}
	if !a.runner.Add(ffprobePath, ffprobeArgs...) {
		// Try to use a local FFprobe instance if the extension isn't installed
		a.runner.Add("ffprobe", ffprobeArgs...)
	}
	blenderImporterPath := filepath.Join(a.scriptsDistDownloadState.Dir(), "hd2_accurate_blender_importer", "hd2_accurate_blender_importer")
	a.runner.Add(blenderImporterPath)
}

// We use state tracking for allSelectedForExport
// since recalculating each frame would eat up
// ~5ms (on a good PC!) for nothing.
// This function should be used whenever
// the status of all items being selected may have
// changed.
func (a *guiApp) calcAllSelectedForExport() bool {
	if a.gameData == nil || len(a.gameData.SortedSearchResultFileIDs) == 0 {
		return false
	}
	for _, id := range a.gameData.SortedSearchResultFileIDs {
		_, sel := a.filesSelectedForExport[id]
		if !sel {
			return false
		}
	}
	return true
}

func (a *guiApp) historyPush(prevItem, newItem stingray.FileID) {
	a.historyStack = a.historyStack[:a.historyStackIndex]
	if (prevItem != stingray.FileID{}) {
		a.historyStack = append(a.historyStack, prevItem)
	}
	a.historyStack = append(a.historyStack, newItem)
	a.historyStackIndex = len(a.historyStack) - 1

	const limit = 128
	if len(a.historyStack) > limit {
		cut := len(a.historyStack) - limit
		a.historyStack = a.historyStack[cut:]
		a.historyStackIndex -= cut
	}
}

// newItem is only valid if changed is true.
func (a *guiApp) drawHistoryButtons() (newItem stingray.FileID, changed bool) {
	imgui.BeginDisabledV(len(a.historyStack) == 0)
	imgui.BeginDisabledV(a.historyStackIndex == 0)
	imgui.SetNextItemShortcut(imgui.KeyChord(imgui.ModAlt | imgui.KeyLeftArrow))
	if imgui.Button(fnt.I("Arrow_back")) {
		a.historyStackIndex--
		newItem = a.historyStack[a.historyStackIndex]
		changed = true
	}
	imgui.SetItemTooltip("Jump to previous item in history (Alt+Left-arrow)")
	imgui.EndDisabled()
	imgui.SameLineV(0, imutils.S(4))
	imgui.BeginDisabledV(a.historyStackIndex == len(a.historyStack)-1)
	imgui.SetNextItemShortcut(imgui.KeyChord(imgui.ModAlt | imgui.KeyRightArrow))
	if imgui.Button(fnt.I("Arrow_forward")) {
		a.historyStackIndex++
		newItem = a.historyStack[a.historyStackIndex]
		changed = true
	}
	imgui.SetItemTooltip("Jump to next item in history (Alt+Right-arrow)")
	imgui.EndDisabled()
	imgui.EndDisabled()
	return
}

func (a *guiApp) drawMenuBar() (menuBarHeight float32) {
	viewport := imgui.MainViewport()

	imgui.SetNextWindowPos(viewport.Pos())
	imgui.SetNextWindowSize(imgui.NewVec2(viewport.Size().X, 0))
	const mainWindowFlags = imgui.WindowFlagsNoDecoration | imgui.WindowFlagsNoResize | imgui.WindowFlagsNoBringToFrontOnFocus | imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoNavFocus | imgui.WindowFlagsMenuBar | imgui.WindowFlagsNoDocking
	if imgui.BeginV("##Main", nil, mainWindowFlags) {
		if imgui.BeginMenuBar() {
			imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imutils.SVec2(3.5, 3.5))
			menuBarHeight = imgui.FrameHeight()
			if imgui.BeginMenu("Help") {
				imgui.Separator()
				if imgui.MenuItemBool(fnt.I("Info") + " About") {
					a.popupManager.Open["About"] = true
				}
				imgui.Separator()
				imgui.EndMenu()
			}
			if imgui.BeginMenu("Settings") {
				imgui.Separator()
				if imgui.MenuItemBool(fnt.I("Extension") + " Extensions") {
					a.popupManager.Open["Extensions"] = true
				}
				imgui.Separator()
				if imgui.MenuItemBool(fnt.I("Sync") + " Check for updates") {
					a.goCheckForUpdates(false)
					a.popupManager.Open["Check for updates"] = true
				}
				imgui.Separator()
				if imgui.MenuItemBool(fnt.I("Settings") + " Preferences") {
					a.popupManager.Open["Preferences"] = true
				}
				imgui.Separator()
				imgui.EndMenu()
			}
			imgui.EndMenuBar()
			imgui.PopStyleVar()
		}
	}
	imgui.End()

	return
}

func (a *guiApp) drawBrowserWindow() {
	if imgui.Begin(fnt.I("View_list") + " Browser") {
		if a.gameData == nil {
			a.gameDataLoad.Lock()
			if a.gameDataLoad.Done {
				if a.gameDataLoad.Err == nil {
					if a.gameData == nil {
						a.gameData = a.gameDataLoad.Result
						types := make(map[stingray.Hash]struct{})
						for id := range a.gameData.DataDir.Files {
							types[id.Type] = struct{}{}
						}
						a.gameFileTypes = []widgets.FilterListSection[stingray.Hash]{
							{Title: "Previewable and exportable"},
							{Title: "Just exportable"},
							{Title: "Not exportable"},
						}
						for _, typ := range slices.SortedFunc(maps.Keys(types), func(h1, h2 stingray.Hash) int {
							return strings.Compare(a.gameData.LookupHash(h1), a.gameData.LookupHash(h2))
						}) {
							var sectionIdx int
							switch typ {
							case // previewable and exportable
								stingray.Sum("bik"),
								stingray.Sum("texture"),
								stingray.Sum("material"),
								stingray.Sum("wwise_bank"),
								stingray.Sum("wwise_stream"),
								stingray.Sum("unit"),
								stingray.Sum("strings"):
								sectionIdx = 0
							default:
								typName := a.gameData.LookupHash(typ)
								if appconfig.Extractable[typName] {
									// just exportable
									sectionIdx = 1
								} else {
									// not exportable
									sectionIdx = 2
								}
							}
							a.gameFileTypes[sectionIdx].Items = append(a.gameFileTypes[sectionIdx].Items, typ)
						}
						a.archiveIDs = []widgets.FilterListSection[stingray.Hash]{
							{
								Items: slices.SortedFunc(maps.Keys(a.gameData.DataDir.Archives), func(h1, h2 stingray.Hash) int {
									return strings.Compare(a.gameData.LookupHash(h1), a.gameData.LookupHash(h2))
								}),
							},
						}
					}
				} else {
					imutils.TextError(a.gameDataLoad.Err)
				}
			} else {
				progress := a.gameDataLoad.Progress
				if progress != 1 {
					imutils.Textf(fnt.I("Hourglass_top") + " Loading game data...")
				} else {
					imutils.Textf(fnt.I("Hourglass_top") + " Processing game data...")
					progress = -float32(imgui.Time())
				}
				imgui.ProgressBar(progress)
			}
			a.gameDataLoad.Unlock()
		} else {
			var newActiveFileID stingray.FileID
			if a.previewState != nil {
				newActiveFileID = a.previewState.ActiveID()
			}

			// Set this to true if the new activeFileID
			// may not be visible in the current clipping
			// region.
			forceUpdateSelected := false
			// true if the item was selected through a history button.
			noPushToHistory := false

			imgui.SetNextItemWidth(imgui.ContentRegionAvail().X - imutils.CheckboxHeight())
			if imgui.Shortcut(imgui.KeyChord(imgui.ModCtrl | imgui.KeyF)) {
				imgui.SetKeyboardFocusHere()
			}
			if imgui.InputTextWithHint("##SearchName", fnt.I("Search")+" Filter by file name...", &a.gameFileSearchQuery, 0, nil) {
				a.gameData.UpdateSearchQuery(a.gameFileSearchQuery, a.selectedGameFileTypes, a.selectedArchives)
				a.allSelectedForExport = a.calcAllSelectedForExport()
			}
			searchInputTextData := imgui.CurrentContext().LastItemData()
			imgui.SetItemTooltip("Filter by file name (Ctrl+F)")
			imgui.SameLine()
			if imgui.Button(fnt.I("Help")) {
				imgui.OpenPopupStr("##AdvancedSearchHelp")
			}
			imgui.SetItemTooltip("Advanced search help")
			if imgui.BeginPopup("##AdvancedSearchHelp") {
				DrawAdvancedSearchHelp()
				imgui.EndPopup()
			}

			if a.gameData.FilterExprErr != nil {
				itm := searchInputTextData
				bottomLeft := imgui.NewVec2(itm.Rect().Min.X, itm.Rect().Max.Y)
				width := itm.Rect().Max.X - itm.Rect().Min.X
				flags := imgui.ChildFlagsFrameStyle | imgui.ChildFlagsAutoResizeY | imgui.ChildFlagsAlwaysAutoResize
				imgui.SetNextWindowPos(bottomLeft)
				if imgui.BeginChildStrV("FilterExprErr", imgui.NewVec2(width, 0), flags, 0) {
					imutils.TextError(a.gameData.FilterExprErr)
				}
				imgui.EndChild()
			}

			if newActiveID, ok := a.drawHistoryButtons(); ok {
				newActiveFileID = newActiveID
				forceUpdateSelected = true
				noPushToHistory = true
			}
			imgui.SameLine()
			widgets.FilterListButton("Types", &a.isTypeFilterOpen, a.selectedGameFileTypes)
			imgui.SameLine()
			widgets.FilterListButton("Archives", &a.isArchiveFilterOpen, a.selectedArchives)

			imgui.BeginDisabledV(len(a.gameData.SortedSearchResultFileIDs) == 0)
			imgui.PushItemFlag(imgui.ItemFlags(imgui.ItemFlagsMixedValue),
				!a.allSelectedForExport && len(a.filesSelectedForExport) > 0)
			if imgui.Checkbox("Select all for export", &a.allSelectedForExport) {
				if a.allSelectedForExport {
					for _, id := range a.gameData.SortedSearchResultFileIDs {
						a.filesSelectedForExport[id] = struct{}{}
					}
				} else {
					for _, id := range a.gameData.SortedSearchResultFileIDs {
						delete(a.filesSelectedForExport, id)
					}
				}
			}
			imgui.PopItemFlag()
			imgui.EndDisabled()
			if len(a.gameData.SortedSearchResultFileIDs) != 0 {
				if a.allSelectedForExport {
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

				if jumpToFile, ok := widgets.PopGamefileLinkFile(); ok {
					if i := slices.Index(a.gameData.SortedSearchResultFileIDs, jumpToFile); i != -1 {
						newActiveFileID = jumpToFile
						forceUpdateSelected = true
					}
				}

				clipper := imgui.NewListClipper()
				clipper.Begin(int32(len(a.gameData.SortedSearchResultFileIDs)))
				if forceUpdateSelected {
					if i := slices.Index(a.gameData.SortedSearchResultFileIDs, newActiveFileID); i != -1 {
						clipper.IncludeItemByIndex(int32(i))
					}
				}
				for clipper.Step() {
					for row := clipper.DisplayStart(); row < clipper.DisplayEnd(); row++ {
						id := a.gameData.SortedSearchResultFileIDs[row]
						imgui.PushIDStr(id.Name.String() + id.Type.String())

						// Kind of hacky way to not add to checkbox's tooltip
						ttExport := false

						toggleExport := func(id stingray.FileID) {
							_, export := a.filesSelectedForExport[id]
							if export {
								delete(a.filesSelectedForExport, id)
							} else {
								a.filesSelectedForExport[id] = struct{}{}
							}
							a.allSelectedForExport = a.calcAllSelectedForExport()
						}

						imgui.TableNextColumn()
						imgui.PushItemFlag(imgui.ItemFlagsNoNav, true)
						_, export := a.filesSelectedForExport[id]
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
							a.gameData.LookupHash(id.Name),
							id == newActiveFileID,
							imgui.SelectableFlagsSpanAllColumns|imgui.SelectableFlags(imgui.SelectableFlagsSelectOnNav),
							imgui.NewVec2(0, 0),
						) || (forceUpdateSelected && id == newActiveFileID)
						copied := false
						if id == newActiveFileID && imgui.Shortcut(imgui.KeyChord(imgui.ModCtrl|imgui.KeyC)) {
							copied = true
						}
						if id == newActiveFileID && imgui.Shortcut(imgui.KeyChord(imgui.KeySpace)) {
							toggleExport(id)
						}
						if imgui.IsItemClickedV(imgui.MouseButtonRight) {
							copied = true
						}
						if copied {
							a.lastBrowserItemCopiedIndex = row
							a.lastBrowserItemCopiedTime = imgui.Time()
							imgui.SetClipboardText(id.Name.String())
						}
						if !ttExport && imgui.IsItemHovered() {
							var text strings.Builder
							if a.lastBrowserItemCopiedIndex == row &&
								imgui.Time()-a.lastBrowserItemCopiedTime < 1 {
								fmt.Fprintf(&text, "%v Copied '%v'\n", fnt.I("Check"), id.Name.String())
							} else {
								if name, ok := a.gameData.Hashes[id.Name]; ok {
									fmt.Fprintf(&text, "Name: %v, hash=%v\n", name, id.Name)
								} else {
									fmt.Fprintf(&text, "Name: hash=%v\n", id.Name)
								}
								fmt.Fprintf(&text, "Type: %v, hash=%v\n", a.gameData.LookupHash(id.Type), id.Type)
								fmt.Fprintf(&text, "%v Left-click to preview\n", fnt.I("Preview"))
								fmt.Fprintf(&text, "%v Right-click or Ctrl+C to copy name hash to clipboard\n", fnt.I("Content_copy"))
								fmt.Fprintf(&text, "%v Down/up arrow keys to select next/previous\n", fnt.I("Unfold_more"))
								fmt.Fprintf(&text, "%v Use the checkbox or space key to select/deselect for export\n", fnt.I("File_export"))
							}
							imgui.SetTooltip(text.String())
						}
						imgui.TableNextColumn()

						imgui.TextUnformatted(a.gameData.LookupHash(id.Type))
						imgui.PopID()

						if selected {
							newActiveFileID = id
							if forceUpdateSelected {
								imgui.SetScrollHereY()
							}
						}
					}
				}

				imgui.EndTable()
			}

			if newActiveFileID != a.previewState.ActiveID() {
				if !noPushToHistory {
					a.historyPush(a.previewState.ActiveID(), newActiveFileID)
				}
				a.previewState.LoadFile(a.ctx, newActiveFileID, a.preferences.PreviewVideoVerticalResolution)
			}

			imutils.Textf("Showing %v/%v files (%v selected for export)", len(a.gameData.SortedSearchResultFileIDs), len(a.gameData.DataDir.Files), len(a.filesSelectedForExport))
		}
	}
	imgui.End()
}

func (a *guiApp) drawTypeFilterWindow() {
	gameFileTypeLabel := func(x stingray.Hash) string {
		items := make([]string, 0, 3)
		if name, ok := a.gameData.Hashes[x]; ok {
			items = append(items, name)
		}
		if desc, ok := gameFileTypeDescriptions[x]; ok {
			items = append(items, "("+desc+")")
		}
		if _, ok := gameFileTypeTooltipsExtra[x]; ok {
			items = append(items, fnt.I("Info"))
		}
		return strings.Join(items, " ")
	}
	imgui.SetNextWindowSizeV(imutils.SVec2(250, 500), imgui.CondOnce)
	if widgets.FilterListWindow("Type Filter",
		&a.isTypeFilterOpen,
		"Search Types",
		&a.gameFileTypeSearchQuery,
		a.gameFileTypes,
		&a.selectedGameFileTypes,
		gameFileTypeLabel,
		func(x stingray.Hash, checked *bool) {
			label := gameFileTypeLabel(x)
			imgui.Checkbox(label, checked)
			tt := fmt.Sprintf("hash=%v", x)
			if ttExtra, ok := gameFileTypeTooltipsExtra[x]; ok {
				tt += "\n" + ttExtra
			}
			imgui.SetItemTooltip(tt)
		},
	) {
		a.gameData.UpdateSearchQuery(a.gameFileSearchQuery, a.selectedGameFileTypes, a.selectedArchives)
		a.allSelectedForExport = a.calcAllSelectedForExport()
	}
}

func (a *guiApp) drawArchiveFilterWindow() {
	imgui.SetNextWindowSizeV(imutils.SVec2(250, 500), imgui.CondOnce)
	if widgets.FilterListWindow("Archive Filter",
		&a.isArchiveFilterOpen,
		"Search Archives",
		&a.archiveIDSearchQuery,
		a.archiveIDs,
		&a.selectedArchives,
		func(x stingray.Hash) string {
			items := make([]string, 0, 3)
			if as, ok := a.gameData.ArmorSets[x]; ok {
				items = append(items, as.Name)
			}
			if name, ok := a.gameData.Hashes[x]; ok {
				items = append(items, name)
			}
			items = append(items, x.String())
			return strings.Join(items, " ")
		},
		func(x stingray.Hash, checked *bool) {
			var label string
			if armorSet, ok := a.gameData.ArmorSets[x]; ok {
				label = x.String() + " (" + armorSet.Name + ")"
			} else {
				label, ok = a.gameData.Hashes[x]
				if !ok {
					label = x.String()
				}
			}
			imgui.Checkbox(label, checked)
			copied := imgui.IsItemClickedV(imgui.MouseButtonRight)
			if imgui.IsItemFocused() && imgui.Shortcut(imgui.KeyChord(imgui.ModCtrl|imgui.KeyC)) {
				copied = true
			}
			if imutils.StickyActivate(copied, 1) {
				imgui.SetItemTooltip(fmt.Sprintf("%v Copied '%v'", fnt.I("Check"), x))
			} else {
				var tt strings.Builder
				if name, ok := a.gameData.Hashes[x]; ok {
					fmt.Fprintf(&tt, "Archive: %v, hash=%v\n", name, x)
				} else {
					fmt.Fprintf(&tt, "Archive: hash=%v\n", x)
				}
				if armorSet, ok := a.gameData.ArmorSets[x]; ok {
					fmt.Fprintf(&tt, "Armor set: %v\n", armorSet.Name)
				}
				fmt.Fprintf(&tt, "Contains %v files\n", len(a.gameData.DataDir.Archives[x]))
				fmt.Fprintf(&tt, "%v Left-click or Space to select/deselect\n", fnt.I("Left_click"))
				fmt.Fprintf(&tt, "%v Right-click or Ctrl+C to copy name hash to clipboard\n", fnt.I("Content_copy"))
				imgui.SetItemTooltip(tt.String())
			}
			if copied {
				imgui.SetClipboardText(x.String())
			}
		},
	) {
		a.gameData.UpdateSearchQuery(a.gameFileSearchQuery, a.selectedGameFileTypes, a.selectedArchives)
		a.allSelectedForExport = a.calcAllSelectedForExport()
	}
}

func (a *guiApp) drawExportWindow() {
	if imgui.Begin(fnt.I("File_export") + " Export") {
		imgui.BeginDisabledV(a.gameDataExport != nil)
		imutils.FilePicker("Output directory", &a.exportDir, true)
		imgui.EndDisabled()

		imgui.Checkbox(fnt.I("Notifications")+" Notify when done", &a.exportNotifyWhenDone)

		imgui.Separator()

		{
			imgui.PushIDStr("Open/close extractor logs")
			var numErrsWarnsStr string
			if a.logger.NumErrs() > 0 || a.logger.NumWarns() > 0 {
				var items []string
				if a.logger.NumErrs() > 0 {
					items = append(items, fmt.Sprintf("%v %v", a.logger.NumErrs(), fnt.I("Error")))
				}
				if a.logger.NumWarns() > 0 {
					items = append(items, fmt.Sprintf("%v %v", a.logger.NumWarns(), fnt.I("Warning")))
				}
				numErrsWarnsStr = fmt.Sprintf(" (%v)", strings.Join(items, ","))
			}
			var label string
			if a.isLogsOpen {
				label = fmt.Sprintf("%v Close extractor logs%v %v", fnt.I("List"), numErrsWarnsStr, fnt.I("Close"))
			} else {
				label = fmt.Sprintf("%v Open extractor logs%v %v", fnt.I("List"), numErrsWarnsStr, fnt.I("Open_in_new"))
			}
			if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				a.isLogsOpen = !a.isLogsOpen
			}
			imgui.PopID()
		}

		if imgui.ButtonV(fnt.I("Folder_open")+" Open output folder", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			_ = os.MkdirAll(a.exportDir, os.ModePerm)
			open.Start(a.exportDir)
		}

		imgui.Separator()

		if a.gameDataExport == nil {
			imgui.PushIDStr("Begin export button")
			imgui.BeginDisabledV(len(a.filesSelectedForExport) == 0 || a.gameData == nil)
			label := fmt.Sprintf("%v Begin export (%v)", fnt.I("File_export"), len(a.filesSelectedForExport))
			if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) && a.gameData != nil {
				a.redetectRunnerProgs()
				a.logger.Reset()
				a.gameDataExport = a.gameData.GoExport(
					a.ctx,
					slices.SortedFunc(maps.Keys(a.filesSelectedForExport), (stingray.FileID).Cmp),
					a.exportDir,
					a.extractorConfig,
					a.runner,
					slices.SortedFunc(maps.Keys(a.selectedArchives), (stingray.Hash).Cmp),
					a.logger,
				)
			}
			if a.gameData == nil {
				imgui.SetItemTooltip("Game data not loaded")
			} else if len(a.filesSelectedForExport) == 0 {
				imgui.SetItemTooltip("Nothing selected for export")
			}
			imgui.EndDisabled()
			imgui.PopID()
		} else {
			if a.gameDataExport.Done {
				if !a.gameDataExport.Canceled && a.exportNotifyWhenDone {
					pluralS := ""
					if a.gameDataExport.NumFiles != 1 {
						pluralS = "s"
					}
					var text string
					if !a.logger.HaveFatalErr() {
						text = fmt.Sprintf("Filediver has finished exporting %v file%v", a.gameDataExport.NumFiles, pluralS)
					} else {
						text = "An internal error occurred during exporting. Please create an issue on GitHub."
					}
					if a.logger.NumErrs() > 0 || a.logger.NumWarns() > 0 {
						text += "\n"
						text += fmt.Sprintf("Errors: %v, Warnings: %v", a.logger.NumErrs(), a.logger.NumWarns())
						text += "\nSee logs."
					}
					if a.logger.NumErrs() > 0 {
						a.isLogsOpen = true
					}
					zenity.Notify(text,
						zenity.Title("Finished exporting"),
						zenity.InfoIcon,
					)
				}
				a.gameDataExport = nil
			} else {
				imgui.ProgressBarV(
					float32(a.gameDataExport.CurrentFileIndex)/float32(a.gameDataExport.NumFiles),
					imgui.NewVec2(-math.SmallestNonzeroFloat32, 0),
					fmt.Sprintf("%v/%v", a.gameDataExport.CurrentFileIndex+1, a.gameDataExport.NumFiles),
				)
				imutils.Textf("%v", a.gameDataExport.CurrentFileName)
				if imgui.ButtonV(fnt.I("Cancel")+" Cancel export", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
					a.gameDataExport.Cancel()
				}
			}
		}
	}
	imgui.End()
}

func (a *guiApp) drawExtractorConfigWindow() {
	if imgui.Begin(fnt.I("Settings_applications") + " Extractor config") {
		prevExtrCfg := a.extractorConfig
		if widgets.ConfigEditor(&a.extractorConfig, &a.extractorConfigShowAdvanced, &a.extractorConfigSearchQuery) {
			if a.extractorConfig.Gamedir != prevExtrCfg.Gamedir {
				a.gameData = nil
				gameDir := ""
				if a.extractorConfig.Gamedir != "<auto-detect>" {
					gameDir = a.extractorConfig.Gamedir
				}
				a.gameDataLoad.GoLoadGameData(a.ctx, gameDir)
			}
			if a.extractorConfig != prevExtrCfg {
				if err := a.extractorConfig.Save(a.extractorConfigPath); err != nil {
					a.showErrorPopup(err)
				}
			}
		}
	}
	imgui.End()
}

func (a *guiApp) drawLogWindow() {
	if !a.isLogsOpen {
		return
	}

	viewport := imgui.MainViewport()
	imgui.SetNextWindowPosV(viewport.Center(), imgui.CondOnce, imgui.NewVec2(0.5, 0.5))
	imgui.SetNextWindowSizeV(imutils.SVec2(350, 350), imgui.CondOnce)
	if imgui.BeginV("Extractor logs", &a.isLogsOpen, 0) {
		avail := imgui.ContentRegionAvail()
		avail.Y -= imgui.FrameHeightWithSpacing()
		imgui.SetNextWindowSize(avail)
		if imgui.BeginChildStr("Log view") {
			LogView(a.logger)
		}
		imgui.EndChild()

		ctx := imgui.CurrentContext()
		btnLabel := fnt.I("Content_copy") + " Copy all to clipboard"
		btnID := imgui.IDStr(btnLabel)
		if imgui.ButtonV(btnLabel, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			imgui.SetClipboardText(a.logger.String())
		}
		if ctx.LastActiveId() == btnID && ctx.LastActiveIdTimer() < 1 {
			imgui.SetItemTooltip(fnt.I("Check") + " Copied")
		}
	}
	imgui.End()
}

func (a *guiApp) drawPreviewWindow(state *imgui_wrapper.State) {
	if imgui.Begin(fnt.I("Preview") + " Preview") {
		if a.previewState != nil {
			if !previews.AutoPreview("Preview", a.previewState) {
				active := a.previewState.ActiveID()
				imgui.PushTextWrapPos()
				if active.Name.Value == 0 {
					imgui.TextUnformatted("Nothing selected to preview. Select an item from the Browser.")
				} else if a.gameData != nil {
					imutils.Textf("Cannot preview type %v.", a.gameData.LookupHash(active.Type))
				}
			}
			if a.previewState.NeedCJKFont() && !state.LoadCJKFonts {
				state.LoadCJKFonts = true
			}
		} else {
			imgui.PushTextWrapPos()
			imgui.TextUnformatted("Loading game data...")
		}
	}
	imgui.End()
}

func (a *guiApp) drawCheckForUpdatesPopup() {
	a.checkUpdatesLock.Lock()
	if a.checkUpdatesOnStartupBGDone && !a.checkUpdatesOnStartupFGDone {
		if a.checkUpdatesErr == nil && a.checkUpdatesNewVersion != version {
			a.popupManager.Open["Check for updates"] = true
		}
		a.checkUpdatesOnStartupFGDone = true
	}
	a.checkUpdatesLock.Unlock()

	imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
	a.popupManager.Popup("Check for updates", func(close func()) {
		a.checkUpdatesLock.Lock()
		if a.checkUpdatesErr == nil {
			if a.checkingForUpdates {
				imgui.ProgressBarV(-1*float32(imgui.Time()), imutils.SVec2(200, 0), "Checking for updates")
			} else if a.checkUpdatesNewVersion == version {
				imutils.Textcf(imgui.NewVec4(0, 0.7, 0, 1), fnt.I("Check")+"Up-to-date (version %v)", version)
			} else {
				imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0, 1), fnt.I("Exclamation")+"New version available: %v", a.checkUpdatesNewVersion)
				if imgui.ButtonV(fnt.I("Download")+" Download "+a.checkUpdatesNewVersion, imutils.SVec2(200, 0)) {
					open.Start(a.checkUpdatesDownloadURL)
				}
				imgui.SetItemTooltip("Open '" + a.checkUpdatesDownloadURL + "'")
			}
		} else {
			imutils.TextError(a.checkUpdatesErr)
		}
		a.checkUpdatesLock.Unlock()
		imgui.Separator()
		if imgui.ButtonV("Close", imutils.SVec2(200, 0)) {
			close()
		}
	}, imgui.WindowFlagsAlwaysAutoResize, true)
}

func (a *guiApp) drawWelcomePopup() {
	viewport := imgui.MainViewport()
	imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
	a.popupManager.Popup("Welcome to Filediver GUI", func(close func()) {
		imgui.PushTextWrapPos()
		imgui.TextUnformatted("Filediver GUI is still work-in-progress.")
		imgui.TextUnformatted("If you encounter any bugs, please create an issue on the project's GitHub page.")
		imgui.Separator()
		imgui.TextLinkOpenURLV("Filediver Issues", "https://github.com/xypwn/filediver/issues")
		imgui.Separator()
		if imgui.ButtonV("OK", imutils.SVec2(300, 0)) {
			close()
		}
	}, imgui.WindowFlagsAlwaysAutoResize, true)
}

func (a *guiApp) drawExtensionsWarningPopup() {
	viewport := imgui.MainViewport()
	imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
	a.popupManager.Popup("Missing or out of date extensions", func(close func()) {
		imgui.PushTextWrapPos()
		imgui.TextUnformatted("The following recommended extensions are missing or out of date:")
		if !a.ffmpegDownloadState.HaveRequestedVersion() {
			imgui.TextWrapped("FFmpeg")
			imgui.Indent()
			imgui.TextWrapped(ffmpegFeatures)
			imgui.Unindent()
		}
		if !a.scriptsDistDownloadState.HaveRequestedVersion() {
			imgui.TextWrapped("ScriptsDist")
			imgui.Indent()
			imgui.TextWrapped(scriptsDistFeatures)
			imgui.Unindent()
		}
		imgui.TextUnformatted("You can download these extensions by clicking \"Manage extensions\", or by going to Settings->Extensions.")
		imgui.Separator()
		if imgui.ButtonV("Close", imutils.SVec2(100, 0)) {
			close()
		}
		imgui.SameLine()
		if imgui.ButtonV("Manage extensions", imutils.SVec2(120, 0)) {
			close()
			a.popupManager.Open["Extensions"] = true
		}
	}, imgui.WindowFlagsAlwaysAutoResize, true)
}

func (a *guiApp) drawExtensionsPopup() {
	viewport := imgui.MainViewport()
	imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
	a.popupManager.Popup("Extensions", func(close func()) {
		widgets.Downloader("FFmpeg", ffmpegFeatures, a.ffmpegDownloadState)
		imgui.Separator()
		widgets.Downloader("ScriptsDist", scriptsDistFeatures, a.scriptsDistDownloadState)
		imgui.Separator()
		if imgui.ButtonV(fnt.I("Folder_open")+" Open extensions folder", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
			_ = os.MkdirAll(a.downloadsDir, os.ModePerm)
			if err := open.Start(a.downloadsDir); err != nil {
				a.showErrorPopup(err)
			}
		}
		imgui.Separator()
		if imgui.ButtonV("Close", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
			close()
		}
	}, imgui.WindowFlagsAlwaysAutoResize, true)
}

func (a *guiApp) drawPreferencesPopup(state *imgui_wrapper.State) {
	viewport := imgui.MainViewport()
	imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
	a.popupManager.Popup("Preferences", func(close func()) {
		prevPrefs := a.preferences
		imutils.ComboChoice(
			"GUI Scale",
			&a.preferences.GUIScale,
			[]float32{0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2, 2.25, 2.5, 2.75, 3},
		)
		imutils.ComboChoice(
			"Target FPS",
			&a.preferences.TargetFPS,
			[]float64{15, 30, 60, 75, 90, 120, 144, 165, 244, 300},
		)
		imutils.ComboChoiceAny(
			"Video Preview Resolution",
			&a.preferences.PreviewVideoVerticalResolution,
			[]int{160, 240, 360, 480, 720, 1080, 1440, 2160},
			func(a, b int) bool { return a == b },
			func(x int) string { return fmt.Sprintf("%vp", x) },
		)
		imgui.SetItemTooltip(fnt.I("Info") + ` The preview player isn't very well optimized, so higher
resolutions may cause low frame rates and poor responsiveness.`)
		imgui.Checkbox("Check for updates on start", &a.preferences.AutoCheckForUpdates)

		imgui.BeginDisabledV(a.preferences == a.defaultPreferences)
		if imgui.ButtonV(fnt.I("Undo")+" Reset all", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
			a.preferences = a.defaultPreferences
		}
		imgui.EndDisabled()

		if a.preferences != prevPrefs {
			state.GUIScale = a.preferences.GUIScale
			state.FrameRate = a.preferences.TargetFPS
			if err := a.preferences.Save(a.preferencesPath); err != nil {
				a.showErrorPopup(err)
			}
		}

		imgui.Separator()
		if imgui.ButtonV(fnt.I("Folder_open")+" Open preferences folder", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
			dir := filepath.Dir(a.preferencesPath)
			_ = os.MkdirAll(dir, os.ModePerm)
			if err := open.Start(dir); err != nil {
				a.showErrorPopup(err)
			}
		}
		imgui.Separator()
		if imgui.ButtonV("Close", imgui.NewVec2(imgui.ContentRegionAvail().X, 0)) {
			close()
		}
	}, imgui.WindowFlagsAlwaysAutoResize, true)
}

func (a *guiApp) drawAboutPopup() {
	viewport := imgui.MainViewport()

	imgui.SetNextWindowSizeV(imutils.SVec2(400, 350), imgui.CondOnce)
	imgui.SetNextWindowSizeConstraints(imutils.SVec2(250, 170), viewport.Size())
	imgui.SetNextWindowPosV(viewport.Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
	a.popupManager.Popup("About", func(close func()) {
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

func main() {
	if version == "" || version == "v0.0.0" {
		fmt.Println(`Development version detected.
To use blender exporter, please pass a valid version to the build (this is because filediver has to know which version of Blender exporter it wants).
You can do this via 'go run -ldflags "-X main.version=v0.0.0" ./cmd/filediver-gui' (replace v0.0.0 with a real version).`)
	}

	clipboardOk := clipboard.Init() == nil

	onError := func(err error) {
		fmt.Println("Error:", err)

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

	app := newGUIApp(onError)
	defer app.Delete()
	if err := imgui_wrapper.Main("Filediver GUI", imgui_wrapper.Options{
		WindowSize:     imgui.NewVec2(800, 700),
		WindowMinSize:  imgui.NewVec2(250, 150),
		OnInitWindow:   app.onInitWindow,
		OnPreDraw:      app.onPreDraw,
		OnDraw:         app.onDraw,
		GLDebugContext: true,
	}); err != nil {
		onError(err)
		return
	}
}
