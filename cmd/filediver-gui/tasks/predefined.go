package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

// Collection of some useful predefined tasks.
var Tasks = struct {
	// Params:
	//   dest (string): destination path (only present if successful)
	//   url (string): download URL
	//   perms (int, optional): unix perms for downloaded file
	//   progressStatus (bool, optional): write progress in mebibytes to status
	// Results: none
	Download TaskFunc
	// Params:
	//   path (string): path to archive
	//   dest (string): destination folder
	//   stripFirstDir (bool, optional): strip first directory from contained file paths
	// Results:
	//   files ([]string): extracted file names
	Unarchive TaskFunc
}{
	Download: func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (_ map[string]any, err error) {
		dest := params["dest"].(string)
		url := params["url"].(string)
		perms, ok := params["perms"].(int)
		if !ok {
			perms = 0666
		}
		progressStatus := params["progressStatus"] == true

		onStatus("Fetching")
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(perms))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		defer func() {
			if err != nil {
				os.Remove(dest)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("get %v: %v", url, resp.Status)
		}

		onStatus("Downloading")
		currentBytes := 0
		var buf [65536]byte
		for {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			n, err := resp.Body.Read(buf[:])
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			_, err = f.Write(buf[:n])
			if err != nil {
				return nil, err
			}

			if resp.ContentLength > 0 {
				if progressStatus {
					const mebi = 1 << 20
					onStatus(fmt.Sprintf("Downloading (%.1f/%.1f MiB)", float64(currentBytes)/mebi, float64(resp.ContentLength)/mebi))
				}
				onProgress(float64(currentBytes) / float64(resp.ContentLength))
			}
			currentBytes += n
		}
		if err := f.Sync(); err != nil {
			return nil, err
		}
		return nil, nil
	},
	Unarchive: func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (_ map[string]any, err error) {
		arPath := params["path"].(string)
		dest := params["dest"].(string)
		stripFirstDir := params["stripFirstDir"] == true

		var arEx archives.Extractor
		{
			var ok bool
			arFormat, _, err := archives.Identify(ctx, arPath, nil)
			if err != nil {
				return nil, err
			}
			arEx, ok = arFormat.(archives.Extractor)
			if !ok {
				return nil, errors.New("unable to extract archive")
			}
		}

		// Limited exponential growth function, since we have to
		// guess the number of files in the archive.
		// 0 at x = 0, 1 at x = ∞, 0.5 at x = lambda.
		limitedExponential := func(x, lambda float64) float64 {
			return 1 - math.Pow(2, -x/lambda)
		}

		var extractedFileNames []string
		arR, err := os.Open(arPath)
		if err != nil {
			return nil, err
		}
		defer arR.Close()
		if err := arEx.Extract(ctx, arR, func(ctx context.Context, fInfo archives.FileInfo) error {
			if err := ctx.Err(); err != nil {
				return err
			}

			if !fInfo.Mode().IsRegular() {
				return nil
			}
			f, err := fInfo.Open()
			if err != nil {
				return err
			}
			defer f.Close()
			nameInArchive := filepath.Clean(fInfo.NameInArchive)
			if stripFirstDir {
				i := strings.IndexAny(nameInArchive, "/\\")
				if i != -1 {
					nameInArchive = nameInArchive[i+1:]
				}
			}
			path := filepath.Clean(filepath.Join(dest, nameInArchive))
			if !strings.HasPrefix(path, dest) {
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
			extractedFileNames = append(extractedFileNames, nameInArchive)
			onProgress(limitedExponential(float64(len(extractedFileNames)), 5))
			return nil
		}); err != nil {
			return nil, err
		}
		return map[string]any{"files": extractedFileNames}, nil
	},
}
