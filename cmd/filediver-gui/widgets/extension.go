package widgets

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/github"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/tasks"
)

var extensionDownloadAndInstallTask tasks.TaskFunc = tasks.Pipeline(
	"outDir->prep.dir",
	"Preparing##prep##0.1", func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (map[string]any, error) {
		return nil, os.MkdirAll(params["dir"].(string), os.ModePerm)
	},
	"temp->dl.dest", "url->dl.url", true, "->dl.progressStatus",
	"##dl##100", tasks.Tasks.Download,
	"temp->ex.path", "outDir->ex.dest", "stripFirstDir->ex.stripFirstDir",
	"Extracting##ex##10", tasks.Tasks.Unarchive,
	"temp->inst.temp", "versionFile->inst.versionFile", "resolvedVersion->inst.resolvedVersion",
	"Installing##inst##1", func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (map[string]any, error) {
		temp := params["temp"].(string)
		versionFile := params["versionFile"].(string)
		resolvedVersion := params["resolvedVersion"].(string)
		if err := os.WriteFile(versionFile, []byte(resolvedVersion), 0666); err != nil {
			return nil, err
		}
		if err := os.Remove(temp); err != nil {
			return nil, err
		}
		return nil, nil
	},
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
	ctx           context.Context
	parentDir     string
	name          string
	ghAsset       github.ReleaseAssetInfo
	stripFirstDir bool
	diskUsage     int // size of decompressed download on disk, 0
	ex            *tasks.TaskExecution

	err            error
	checked        bool
	presentVersion string // "" if not downloaded
}

func NewExtension(ctx context.Context, parentDir, name string, ghAsset github.ReleaseAssetInfo, stripFirstDir bool) *Extension {
	return &Extension{
		ctx:           ctx,
		parentDir:     parentDir,
		name:          name,
		ghAsset:       ghAsset,
		stripFirstDir: stripFirstDir,
	}
}

func (ex *Extension) checkPresentVersion() (string, error) {
	versionPath := filepath.Join(ex.parentDir, ex.name+"_version")
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

func (ex *Extension) goDownload() {
	if !ex.checked {
		ex.check()
	}
	ex.ex = extensionDownloadAndInstallTask.Go(ex.ctx, map[string]any{
		"temp":            filepath.Join(ex.Dir(), ex.ghAsset.Filename),
		"url":             ex.ghAsset.DownloadUrl,
		"outDir":          ex.Dir(),
		"stripFirstDir":   ex.stripFirstDir,
		"versionFile":     filepath.Join(ex.parentDir, ex.name+"_version"),
		"resolvedVersion": ex.ghAsset.ResolvedVersion,
	})
}

// Checks if the archive is downloaded.
// If it is, check also sets the diskUsage.
func (ex *Extension) check() {
	defer func() {
		ex.checked = true
	}()

	presentVersion, err := ex.checkPresentVersion()
	if err != nil {
		ex.err = fmt.Errorf("checking current version: %w", err)
		return
	}
	ex.presentVersion = presentVersion

	if ex.presentVersion != "" {
		du, err := diskUsage(ex.Dir())
		if err != nil {
			ex.err = fmt.Errorf("checking disk usage: %w", err)
			return
		}
		ex.diskUsage = int(du)
	}
}

// Returns the output directory.
func (ex *Extension) Dir() string {
	return filepath.Join(ex.parentDir, ex.name)
}

// Returns true if requested version is already downloaded.
func (ex *Extension) HaveRequestedVersion() bool {
	if !ex.checked {
		ex.check()
	}
	return ex.presentVersion == ex.ghAsset.ResolvedVersion
}

func (ex *Extension) DrawAndUpdate(title, description string) {
	imgui.PushIDStr(title)
	defer imgui.PopID()

	imgui.TextUnformatted(title)
	imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v", description)

	const mebi = 1 << 20

	if !ex.checked {
		ex.check()
	}
	ts := ex.ex.Poll()
	if ts.Done {
		if ts.JustFinished {
			ex.checked = false
			ex.ex.Cancel()
		}
		if ex.presentVersion == "" {
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Not downloaded")
		} else {
			var prefix string
			if ex.HaveRequestedVersion() {
				prefix = "Downloaded"
			} else {
				prefix = "Out of date"
			}
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v (version: %v, size: %3.1f MiB)", prefix, ex.presentVersion, float32(ex.diskUsage)/mebi)
		}
		if !ex.HaveRequestedVersion() {
			label := fnt.I.Download + " Download"
			if ex.presentVersion != "" {
				label = fnt.I.Download + " Update to version " + ex.ghAsset.ResolvedVersion
			}
			if ex.err != nil || (ts.Err != nil && !errors.Is(ts.Err, context.Canceled)) {
				if ex.err != nil {
					imutils.TextError(ex.err)
				}
				if ts.Err != nil {
					imutils.TextError(ts.Err)
				}
				label = fnt.I.Download + " Retry"
			}
			if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				ex.err = nil
				ex.goDownload()
			}
		}
		if ex.presentVersion != "" {
			if imgui.ButtonV(fnt.I.Delete+" Delete", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
				imgui.OpenPopupStr("Confirm delete")
			}
		}
		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		if imgui.BeginPopupModalV("Confirm delete", nil, imgui.WindowFlagsAlwaysAutoResize) {
			imutils.Textf("Delete %v?\nYou can always re-download it.", title)
			if imgui.ButtonV("Delete", imutils.SVec2(80, 0)) {
				if err := os.Remove(filepath.Join(ex.parentDir, ex.name+"_version")); err == nil {
					_ = os.Remove(filepath.Join(ex.parentDir, ex.name+".tmp"))
					_ = os.RemoveAll(ex.Dir())
				}
				ex.checked = false
				imgui.CloseCurrentPopup()
			}
			imgui.SameLine()
			if imgui.ButtonV("Cancel", imutils.SVec2(80, 0)) {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	} else {
		DrawTask("##DownloadStatus", ts)
		if imgui.ButtonV(fnt.I.Cancel+" Cancel", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			ex.ex.Cancel()
		}
	}
}
