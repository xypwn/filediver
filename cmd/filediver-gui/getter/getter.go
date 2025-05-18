package getter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/mholt/archives"
)

var reGHVersionedReleaseDLURL = regexp.MustCompile(`^https:\/\/github\.com\/[^\/]+\/[^\/]+\/releases\/download\/([^\/]+)\/([^\/]+)$`)

type Target struct {
	SubdirName        string
	GHUser            string
	GHRepo            string
	PinnedVersion     string // empty for latest, or a specific version
	GHFilenameWindows string
	GHFilenameLinux   string
	StripFirstDir     bool // remove first top-level folder from destination directory
}

// GetInfo will only use the network if allowNetworkVersionResolution is set to true.
func (t Target) GetInfo(allowNetworkVersionResolution bool) (Info, error) {
	if runtime.GOARCH != "amd64" {
		return Info{}, fmt.Errorf("unsupported CPU architecture: %v", runtime.GOARCH)
	}

	var ghFilename string
	switch runtime.GOOS {
	case "windows":
		ghFilename = t.GHFilenameWindows
	case "linux":
		ghFilename = t.GHFilenameLinux
	default:
		return Info{}, fmt.Errorf("unsupported OS: %v", runtime.GOOS)
	}
	var url string
	if t.PinnedVersion == "" {
		url = fmt.Sprintf("https://github.com/%v/%v/releases/latest/download/%v", t.GHUser, t.GHRepo, ghFilename)
	} else {
		url = fmt.Sprintf("https://github.com/%v/%v/releases/download/%v/%v", t.GHUser, t.GHRepo, t.PinnedVersion, ghFilename)
	}

	versionFromURL := func(url string) (version string, ok bool) {
		m := reGHVersionedReleaseDLURL.FindStringSubmatch(url)
		if len(m) <= 1 {
			return "", false
		}
		return m[1], true
	}
	if ver, ok := versionFromURL(url); ok {
		return Info{Target: t, ResolvedVersion: ver, DownloadURL: url}, nil
	}

	if !allowNetworkVersionResolution {
		return Info{}, errors.New("network version resolution not allowed")
	}

	resp, err := (&http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}).Get(url)
	if err != nil {
		return Info{}, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		return Info{}, fmt.Errorf("invalid GitHub release asset URL \"%v\": expected status 302, but got %v", url, resp.Status)
	}

	loc, err := resp.Location()
	if err != nil {
		return Info{}, err
	}
	if ver, ok := versionFromURL(loc.String()); ok {
		return Info{Target: t, ResolvedVersion: ver, DownloadURL: loc.String()}, nil
	} else {
		return Info{}, fmt.Errorf("invalid GitHub release asset URL \"%v\": contains no version information", url)
	}
}

func (t Target) Dir(parentDir string) string {
	return filepath.Join(parentDir, t.SubdirName)
}

func (t Target) Paths(parentDir string) (contentsPath, tmpPath, versionPath string) {
	dir := t.Dir(parentDir)
	contentsPath = dir
	tmpPath = dir + "_tmp"
	versionPath = dir + "_version"
	return
}

// presentVersion may be "" if no version is present.
func (t Target) Check(parentDir string) (dir, presentVersion string, err error) {
	dir, _, versionPath := t.Paths(parentDir)

	if _, err := os.Stat(versionPath); err == nil {
		verB, err := os.ReadFile(versionPath)
		if err != nil {
			return "", "", err
		}
		presentVersion = strings.TrimSpace(string(verB))
	} else if !os.IsNotExist(err) {
		return "", "", err
	}

	return
}

type Info struct {
	Target          Target
	ResolvedVersion string
	DownloadURL     string
}

func (info Info) Download(ctx context.Context, parentDir string, onProgress func(Progress)) error {
	progress := Progress{}
	progress.State = Fetching
	onProgress(progress)

	dir, tmpPath, versionPath := info.Target.Paths(parentDir)

	req, err := http.NewRequestWithContext(ctx, "GET", info.DownloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get %v: %v", info.DownloadURL, resp.Status)
	}

	progress.State = Downloading
	progress.ContentTotalBytes = int(resp.ContentLength)
	onProgress(progress)

	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		_ = os.Remove(tmpPath)
	}()

	var buf [65536]byte
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		n, err := resp.Body.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		_, err = f.Write(buf[:n])
		if err != nil {
			return err
		}

		progress.ContentCurrentBytes += n
		onProgress(progress)
	}

	if err := f.Sync(); err != nil {
		return err
	}

	progress.State = Fetching
	onProgress(progress)

	var arEx archives.Extractor
	{
		var ok bool
		arFormat, _, err := archives.Identify(ctx, info.DownloadURL, nil)
		if err != nil {
			return err
		}
		arEx, ok = arFormat.(archives.Extractor)
		if !ok {
			return errors.New("unable to extract archive")
		}
	}

	if err := os.Remove(versionPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return err
	}

	arR, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer arR.Close()
	if err := arEx.Extract(ctx, arR, func(ctx context.Context, fInfo archives.FileInfo) error {
		if !fInfo.Mode().IsRegular() {
			return nil
		}
		f, err := fInfo.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		nameInArchive := filepath.Clean(fInfo.NameInArchive)
		if info.Target.StripFirstDir {
			i := strings.IndexAny(nameInArchive, "/\\")
			if i != -1 {
				nameInArchive = nameInArchive[i+1:]
			}
		}
		path := filepath.Clean(filepath.Join(dir, nameInArchive))
		if !strings.HasPrefix(path, dir) {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fInfo.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, f); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := os.WriteFile(versionPath, []byte(info.ResolvedVersion), 0666); err != nil {
		return err
	}

	progress.State = Done
	onProgress(progress)
	return nil
}

func (info Info) GoDownload(ctx context.Context, parentDir string, onProgress func(Progress, error)) {
	go func() {
		var lastP Progress
		err := info.Download(ctx, parentDir, func(p Progress) {
			onProgress(p, nil)
			lastP = p
		})
		if err != nil {
			onProgress(lastP, err)
		}
	}()
}
