package widgets

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/getter"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
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

type DownloaderState struct {
	cancel    func()
	parentDir string
	info      getter.Info
	diskUsage int // size of decompressed download on disk, 0

	lock           sync.Mutex // locks all fields below lock
	checked        bool
	err            error
	presentVersion string // "" if not downloaded
	progress       getter.Progress
}

func NewDownloader(parentDir string, info getter.Info) *DownloaderState {
	return &DownloaderState{
		cancel:    func() {},
		parentDir: parentDir,
		info:      info,
	}
}

// Not thread-safe.
func (ds *DownloaderState) goDownload() {
	if !ds.checked {
		ds.check()
	}
	ctx, cancel := context.WithCancel(context.Background())
	ds.cancel = cancel
	ds.info.GoDownload(ctx, ds.parentDir, func(p getter.Progress, err error) {
		ds.lock.Lock()
		ds.progress = p
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				ds.err = err
			}
			ds.progress = getter.Progress{}
		}
		if ds.progress.State == getter.Done {
			ds.check()
		}
		ds.lock.Unlock()
	})
}

// Checks if the archive is downloaded.
// If it is, check also sets the diskUsage.
// Not thread-safe.
func (ds *DownloaderState) check() {
	defer func() {
		ds.checked = true
	}()

	dir, presentVersion, err := ds.info.Target.Check(ds.parentDir)
	if err != nil {
		ds.err = fmt.Errorf("checking current version: %w", err)
		return
	}
	ds.presentVersion = presentVersion

	if ds.presentVersion != "" {
		du, err := diskUsage(dir)
		if err != nil {
			ds.err = fmt.Errorf("checking disk usage: %w", err)
			return
		}
		ds.diskUsage = int(du)
		ds.progress.State = getter.Done
	}
}

// Returns the output directory.
func (ds *DownloaderState) Dir() string {
	ds.lock.Lock()
	res := ds.info.Target.Dir(ds.parentDir)
	ds.lock.Unlock()
	return res
}

// Returns true if requested version is already downloaded.
func (ds *DownloaderState) HaveRequestedVersion() bool {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if !ds.checked {
		ds.check()
	}
	return ds.presentVersion == ds.info.ResolvedVersion
}

func Downloader(title, description string, ds *DownloaderState) {
	imgui.PushIDStr(title)
	defer imgui.PopID()

	imgui.TextUnformatted(title)
	imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v", description)

	const mebi = 1048576

	ds.lock.Lock()
	if !ds.checked {
		ds.check()
	}
	switch ds.progress.State {
	case getter.Unknown, getter.Done:
		if ds.presentVersion == "" {
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Not downloaded")
		} else {
			var prefix string
			if ds.presentVersion == ds.info.ResolvedVersion {
				prefix = "Downloaded"
			} else {
				prefix = "Out of date"
			}
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v (version: %v, size: %3.1f MiB)", prefix, ds.presentVersion, float32(ds.diskUsage)/mebi)
		}
		if ds.presentVersion != ds.info.ResolvedVersion {
			label := fnt.I.Download + " Download"
			if ds.presentVersion != "" {
				label = fnt.I.Download + " Update to version " + ds.info.ResolvedVersion
			}
			if ds.err != nil {
				imutils.TextError(ds.err)
				label = fnt.I.Download + " Retry"
			}
			if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				ds.goDownload()
			}
		}
		if ds.presentVersion != "" {
			if imgui.ButtonV(fnt.I.Delete+" Delete", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				imgui.OpenPopupStr("Confirm delete")
			}
		}
		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		if imgui.BeginPopupModalV("Confirm delete", nil, imgui.WindowFlagsAlwaysAutoResize) {
			imutils.Textf("Delete %v?\nYou can always re-download it.", title)
			if imgui.ButtonV("Delete", imutils.SVec2(80, 0)) {
				contentsPath, tmpPath, versionPath := ds.info.Target.Paths(ds.parentDir)
				if err := os.Remove(versionPath); err == nil {
					_ = os.Remove(tmpPath)
					_ = os.RemoveAll(contentsPath)
					ds.presentVersion = ""
					ds.progress = getter.Progress{}
				}
				imgui.CloseCurrentPopup()
			}
			imgui.SameLine()
			if imgui.ButtonV("Cancel", imutils.SVec2(80, 0)) {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	case getter.Fetching, getter.Downloading, getter.Extracting:
		imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Downloading")
		progBarProg := -1 * float32(imgui.Time())
		var progBarText string
		switch ds.progress.State {
		case getter.Fetching:
			progBarText = "Fetching"
		case getter.Extracting:
			progBarText = "Extracting"
		case getter.Downloading:
			curr, total := float32(ds.progress.ContentCurrentBytes), float32(ds.progress.ContentTotalBytes)
			progBarProg = curr / total
			progBarText = fmt.Sprintf("%3.1f/%3.1f MiB", curr/mebi, total/mebi)
		}
		imgui.ProgressBarV(progBarProg, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0), progBarText)
		if imgui.ButtonV(fnt.I.Cancel+" Cancel", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			ds.cancel()
		}
	default:
		imutils.TextError(errors.New("unexpected state"))
	}
	ds.lock.Unlock()
}
