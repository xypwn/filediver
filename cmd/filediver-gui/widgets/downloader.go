package widgets

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/mholt/archives"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
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

type downloaderStateInfo struct {
	archiveURL    string
	destBaseDir   string
	stripFirstDir bool
}

func (info *downloaderStateInfo) hashedURL() string {
	h := fnv.New32a()
	h.Write([]byte(info.archiveURL))
	return strconv.Itoa(int(h.Sum32()))
}

func (info *downloaderStateInfo) dirPath() string {
	return filepath.Join(info.destBaseDir, info.hashedURL())
}

func (info *downloaderStateInfo) completeFilePath() string {
	return filepath.Join(info.destBaseDir, info.hashedURL()+".complete")
}

func (info *downloaderStateInfo) tmpArchivePath() string {
	return filepath.Join(info.destBaseDir, info.hashedURL()+".tmp")
}

type downloaderStateState int

const (
	downloaderUnknown downloaderStateState = iota
	downloaderNotDownloaded
	downloaderDownloading
	downloaderDownloaded
)

type downloaderStateProgress struct {
	contentLength int
	currentBytes  int
}

type DownloaderState struct {
	lock      sync.Mutex
	state     downloaderStateState
	cancel    bool
	err       error
	diskUsage int // size of decompressed download on disk
	info      downloaderStateInfo
	progress  downloaderStateProgress
}

// archiveURL can be .zip, .tar, .tar.{gz,xz,bz2,zst}, .rar, .7z
// If stripFirstDir is true, the initial folder of the destination directory
// is removed. This is useful e.g. if all files in the archive are within
// a folder of the same name.
func NewDownloader(archiveURL, destBaseDir string, stripFirstDir bool) *DownloaderState {
	return &DownloaderState{
		info: downloaderStateInfo{
			archiveURL:    archiveURL,
			destBaseDir:   destBaseDir,
			stripFirstDir: stripFirstDir,
		},
	}
}

// Not thread-safe.
func (ds *DownloaderState) goDownload() {
	ds.state = downloaderDownloading
	ds.progress = downloaderStateProgress{}

	info := ds.info // copy for thread safety

	go func() {
		var resErr error

		defer func() {
			ds.lock.Lock()
			ds.cancel = false
			if resErr != nil {
				ds.err = resErr
			}
			ds.state = downloaderUnknown
			ds.check()
			ds.lock.Unlock()
		}()

		if err := os.MkdirAll(info.dirPath(), os.ModePerm); err != nil {
			resErr = err
			return
		}

		resp, err := http.Get(info.archiveURL)
		if err != nil {
			resErr = err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			resErr = fmt.Errorf("get %v: %v", info.archiveURL, resp.Status)
			return
		}

		ds.lock.Lock()
		ds.progress.contentLength = int(resp.ContentLength)
		ds.lock.Unlock()

		f, err := os.Create(info.tmpArchivePath())
		if err != nil {
			resErr = err
			return
		}
		defer func() {
			f.Close()
			_ = os.Remove(info.tmpArchivePath())
		}()

		nBytesCurr := 0
		var buf [65536]byte
		for {
			ds.lock.Lock()
			canceled := ds.cancel
			ds.lock.Unlock()
			if canceled {
				return
			}

			n, err := resp.Body.Read(buf[:])
			if err != nil {
				if err == io.EOF {
					break
				}
				resErr = err
				return
			}

			_, err = f.Write(buf[:n])
			if err != nil {
				resErr = err
				return
			}

			nBytesCurr += n
			ds.lock.Lock()
			ds.progress.currentBytes = nBytesCurr
			ds.lock.Unlock()
		}

		if err := f.Sync(); err != nil {
			resErr = err
			return
		}

		var arEx archives.Extractor
		{
			var ok bool
			arFormat, _, err := archives.Identify(context.Background(), info.archiveURL, nil)
			if err != nil {
				resErr = err
				return
			}
			arEx, ok = arFormat.(archives.Extractor)
			if !ok {
				resErr = errors.New("unable to extract archive")
				return
			}
		}

		arR, err := os.Open(info.tmpArchivePath())
		if err != nil {
			resErr = err
			return
		}
		defer arR.Close()
		if err := arEx.Extract(context.Background(), arR, func(ctx context.Context, fInfo archives.FileInfo) error {
			if !fInfo.Mode().IsRegular() {
				return nil
			}
			f, err := fInfo.Open()
			if err != nil {
				return err
			}
			defer f.Close()
			nameInArchive := filepath.Clean(fInfo.NameInArchive)
			if info.stripFirstDir {
				i := strings.IndexAny(nameInArchive, "/\\")
				if i != -1 {
					nameInArchive = nameInArchive[i+1:]
				}
			}
			dstDir := info.dirPath()
			dst := filepath.Clean(filepath.Join(dstDir, nameInArchive))
			if !strings.HasPrefix(dst, dstDir) {
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
				return err
			}
			out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fInfo.Mode())
			if err != nil {
				return err
			}
			defer out.Close()
			if _, err := io.Copy(out, f); err != nil {
				return err
			}
			return nil
		}); err != nil {
			resErr = err
			return
		}

		if err := os.WriteFile(info.completeFilePath(), []byte{}, 0666); err != nil {
			resErr = err
			return
		}
	}()
}

// Checks if the archive is downloaded.
// If it is, check also sets the diskUsage.
// Guarantees the resulting state to be
// downloaderDownloaded or downloaderNotDownloaded.
// Not thread-safe.
func (ds *DownloaderState) check() {
	info, err := os.Stat(ds.info.completeFilePath())
	if err == nil && info.Mode().IsRegular() {
		ds.state = downloaderDownloaded
		du, err := diskUsage(ds.info.dirPath())
		if err != nil {
			ds.err = fmt.Errorf("checking disk usage: %w", err)
		}
		ds.diskUsage = int(du)
	} else {
		if !os.IsNotExist(err) {
			ds.err = err
		}
		ds.state = downloaderNotDownloaded
	}
}

// Returns the output directory.
func (ds *DownloaderState) Dir() string {
	ds.lock.Lock()
	res := ds.info.dirPath()
	ds.lock.Unlock()
	return res
}

// Returns true if already downloaded.
func (ds *DownloaderState) Have() bool {
	ds.lock.Lock()
	if ds.state == downloaderUnknown {
		ds.check()
	}
	ok := ds.state == downloaderDownloaded
	ds.lock.Unlock()
	return ok
}

func Downloader(title, description string, ds *DownloaderState) {
	imgui.PushIDStr(title)
	defer imgui.PopID()

	imgui.TextUnformatted(title)
	imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v", description)

	ds.lock.Lock()
	if ds.state == downloaderUnknown {
		ds.check()
	}
	switch ds.state {
	case downloaderNotDownloaded:
		imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Not downloaded")
		label := fnt.I("Download") + " Download"
		if ds.err != nil {
			imutils.TextError(ds.err)
			label = fnt.I("Download") + " Retry"
		}
		if imgui.ButtonV(label, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			ds.goDownload()
		}
	case downloaderDownloading:
		imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Downloading")
		progBarProg := -1 * float32(imgui.Time())
		var progBarText string
		switch ds.progress.currentBytes {
		case 0:
			progBarText = "Fetching"
		case ds.progress.contentLength:
			progBarText = "Extracting"
		default:
			progBarProg = float32(ds.progress.currentBytes) / float32(ds.progress.contentLength)
			progBarText = fmt.Sprintf("%3.1f/%3.1f MB", float32(ds.progress.currentBytes)/1e6, float32(ds.progress.contentLength)/1e6)
		}
		imgui.ProgressBarV(progBarProg, imgui.NewVec2(-math.SmallestNonzeroFloat32, 0), progBarText)
		if imgui.ButtonV(fnt.I("Cancel")+" Cancel", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			ds.cancel = true
		}
	case downloaderDownloaded:
		imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Downloaded (%3.1f MB)", float32(ds.diskUsage)/1e6)
		if imgui.ButtonV(fnt.I("Delete")+" Delete", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
			imgui.OpenPopupStr("Confirm delete")
		}

		imgui.SetNextWindowPosV(imgui.MainViewport().Center(), imgui.CondAlways, imgui.NewVec2(0.5, 0.5))
		if imgui.BeginPopupModalV("Confirm delete", nil, imgui.WindowFlagsAlwaysAutoResize) {
			imutils.Textf("Delete %v?\nYou can always re-download it.", title)
			if imgui.ButtonV("Delete", imgui.NewVec2(120, 0)) {
				if err := os.Remove(ds.info.completeFilePath()); err == nil {
					_ = os.RemoveAll(ds.info.dirPath())
					ds.state = downloaderNotDownloaded
				}
				imgui.CloseCurrentPopup()
			}
			imgui.SameLine()
			if imgui.ButtonV("Cancel", imgui.NewVec2(120, 0)) {
				imgui.CloseCurrentPopup()
			}
			imgui.EndPopup()
		}
	default:
		imutils.TextError(errors.New("unexpected state"))
	}
	ds.lock.Unlock()
}
