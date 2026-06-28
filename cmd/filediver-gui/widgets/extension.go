package widgets

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/github"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/tasks"
)

func diskUsage(path string) (int64, error) {
	var res int64
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			res += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return res, nil
}

type Extension struct {
	cancel        func()
	parentDir     string
	name          string
	ghAsset       github.ReleaseAssetInfo
	stripFirstDir bool
	diskUsage     int // size of decompressed download on disk, 0
	ex            *tasks.TaskExecution

	err            error
	checked        atomic.Bool
	presentVersion string // "" if not downloaded
}

func NewExtension(parentDir, name string, ghAsset github.ReleaseAssetInfo, stripFirstDir bool) *Extension {
	return &Extension{
		cancel:        func() {},
		parentDir:     parentDir,
		name:          name,
		ghAsset:       ghAsset,
		stripFirstDir: stripFirstDir,
	}
}

func (es *Extension) checkPresentVersion() (string, error) {
	versionPath := filepath.Join(es.parentDir, es.name+"_version")
	if _, err := os.Stat(versionPath); err == nil {
		verB, err := os.ReadFile(versionPath)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(verB)), nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return "", nil
}

func (es *Extension) goDownload() {
	if !es.checked.Load() {
		es.check()
	}
	ctx, cancel := context.WithCancel(context.Background())
	es.cancel = cancel
	task := tasks.Sequential(
		"Preparing", 0, func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (map[string]any, error) {
			return nil, os.MkdirAll(es.Dir(), os.ModePerm)
		},
		100, tasks.Tasks.Download,
		tasks.Rename("outPath", "path"),
		"Extracting", 10, tasks.Tasks.Unarchive,
		"Installing", 1, func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (map[string]any, error) {
			versionPath := filepath.Join(es.parentDir, es.name+"_version")
			if err := os.WriteFile(versionPath, []byte(es.ghAsset.ResolvedVersion), 0666); err != nil {
				return nil, err
			}
			if err := os.Remove(params["outPath"].(string)); err != nil {
				return nil, err
			}
			es.checked.Store(false)
			return nil, nil
		},
	)
	es.ex = task.Go(ctx, map[string]any{
		"outPath":        filepath.Join(es.Dir(), es.ghAsset.Filename),
		"url":            es.ghAsset.DownloadUrl,
		"progressStatus": true,
		"outDir":         es.Dir(),
		"stripFirstDir":  es.stripFirstDir,
	}, nil)
}

// Checks if the archive is downloaded.
// If it is, check also sets the diskUsage.
func (es *Extension) check() {
	defer func() {
		es.checked.Store(true)
	}()

	presentVersion, err := es.checkPresentVersion()
	if err != nil {
		es.err = fmt.Errorf("checking current version: %w", err)
		return
	}
	es.presentVersion = presentVersion

	if es.presentVersion != "" {
		du, err := diskUsage(es.Dir())
		if err != nil {
			es.err = fmt.Errorf("checking disk usage: %w", err)
			return
		}
		es.diskUsage = int(du)
	}
}

// Returns the output directory.
func (es *Extension) Dir() string {
	return filepath.Join(es.parentDir, es.name)
}

// Returns true if requested version is already downloaded.
func (es *Extension) HaveRequestedVersion() bool {
	return es.presentVersion == es.ghAsset.ResolvedVersion
}

func (es *Extension) Draw(title, description string) {
	imgui.PushIDStr(title)
	defer imgui.PopID()

	imgui.TextUnformatted(title)
	imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v", description)

	const mebi = 1 << 20

	if !es.checked.Load() {
		es.check()
	}
	ex := es.ex.Snap()
	if ex.Done {
		if es.presentVersion == "" {
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Not downloaded")
		} else {
			var prefix string
			if es.HaveRequestedVersion() {
				prefix = "Downloaded"
			} else {
				prefix = "Out of date"
			}
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v (version: %v, size: %3.1f MiB)", prefix, es.presentVersion, float32(es.diskUsage)/mebi)
		}
		if !es.HaveRequestedVersion() {
			label := fnt.I.Download + " Download"
			if es.presentVersion != "" {
				label = fnt.I.Download + " Update to version " + es.ghAsset.ResolvedVersion
			}
			if es.err != nil || ex.Err != nil {
				if es.err != nil {
					imutils.TextError(es.err)
				} else {
					imutils.TextError(ex.Err)
				}
				label = fnt.I.Download + " Retry"
			}
			if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				es.err = nil
				es.goDownload()
			}
		}
		if es.presentVersion != "" {
			if imgui.ButtonV(fnt.I.Delete+" Delete", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				imgui.OpenPopupStr("Confirm delete")
			}
		}
		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		if imgui.BeginPopupModalV("Confirm delete", nil, imgui.WindowFlagsAlwaysAutoResize) {
			imutils.Textf("Delete %v?\nYou can always re-download it.", title)
			if imgui.ButtonV("Delete", imutils.SVec2(80, 0)) {
				if err := os.Remove(filepath.Join(es.parentDir, es.name+"_version")); err == nil {
					_ = os.Remove(filepath.Join(es.parentDir, es.name+".tmp"))
					_ = os.RemoveAll(es.Dir())
				}
				es.checked.Store(false)
				imgui.CloseCurrentPopup()
			}
			imgui.SameLine()
			if imgui.ButtonV("Cancel", imutils.SVec2(80, 0)) {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	} else {
		ex.Draw("##DownloadStatus")
		/*imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Downloading")
		progBarProg := -1 * float32(imgui.Time())
		var progBarText string
		switch es.progress.State {
		case getter.Fetching:
			progBarText = "Fetching"
		case getter.Extracting:
			progBarText = "Extracting"
		case getter.Downloading:
			curr, total := float32(es.progress.ContentCurrentBytes), float32(es.progress.ContentTotalBytes)
			progBarProg = curr / total
			progBarText = fmt.Sprintf("%3.1f/%3.1f MiB", curr/mebi, total/mebi)
		}
		imgui.ProgressBarV(progBarProg, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0), progBarText)*/
		if imgui.ButtonV(fnt.I.Cancel+" Cancel", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			es.cancel()
		}
	}
}
