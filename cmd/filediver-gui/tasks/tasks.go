package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/mholt/archives"
)

type TaskFunc func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (result map[string]any, err error)

// args:
// - Task: will be executed in order
// - string: name of next task
// - float32/float64/int: weight of next task (default 1)
func Sequential(args ...any) TaskFunc {
	type subtask struct {
		fn     TaskFunc
		name   string
		weight float64
	}

	var subtasks []subtask
	weight := float64(1)
	var name string
	var weightSum float64
	addSubtask := func(fn TaskFunc) {
		st := subtask{
			fn:     fn,
			name:   name,
			weight: weight,
		}
		subtasks = append(subtasks, st)
		weightSum += weight
		weight = 1
		name = ""
	}
	for _, arg := range args {
		switch arg := arg.(type) {
		case float64:
			weight = arg
		case float32:
			weight = float64(arg)
		case int:
			weight = float64(arg)
		case string:
			name = arg
		case TaskFunc:
			addSubtask(arg)
		case func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (result map[string]any, err error):
			addSubtask(arg)
		default:
			panic("unknown arg type: " + reflect.TypeOf(arg).String())
		}
	}
	for i := range subtasks {
		subtasks[i].weight /= weightSum
	}

	return func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (_ map[string]any, err error) {
		var totalProg float64
		runSub := func(sub subtask, params map[string]any) (map[string]any, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			onStatus(sub.name)
			res, err := sub.fn(
				ctx,
				params,
				func(prog float64) {
					onProgress(totalProg + prog*sub.weight)
				},
				func(s string) {
					n := sub.name
					if n != "" {
						n += ": "
					}
					n += s
					onStatus(n)
				},
			)
			if err != nil {
				return nil, err
			}
			totalProg += sub.weight
			onProgress(totalProg)
			return res, err
		}
		result := make(map[string]any)
		maps.Copy(result, params)
		for _, sub := range subtasks {
			res, err := runSub(sub, result)
			if err != nil {
				return nil, err
			}
			maps.Copy(result, res)
		}
		return result, nil
	}
}

func Rename(oldnew ...string) TaskFunc {
	if len(oldnew)%2 == 1 {
		panic("expected oldnew to consist of pairs (number of items must be even)")
	}
	return func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (result map[string]any, err error) {
		result = make(map[string]any)
		for i := 0; i < len(oldnew); i += 2 {
			o := oldnew[i]
			n := oldnew[i+1]
			result[n] = params[o]
		}
		return result, nil
	}
}

type TaskExecution struct {
	Lock   sync.Mutex
	Fn     TaskFunc
	Status string
	Res    any
	Err    error
	Prog   float64
	Done   bool
}

func NewErroredTaskExecution(err error) *TaskExecution {
	return &TaskExecution{
		Err:  err,
		Done: true,
	}
}

// Returns an independent snapshot of the current execution status, which
// can be safely used.
// If ex is nil, returns a TaskExecution with Done set to true and all other fields zeroed.
func (ex *TaskExecution) Snap() TaskExecution {
	if ex == nil {
		return TaskExecution{Done: true}
	}

	ex.Lock.Lock()
	defer ex.Lock.Unlock()

	return TaskExecution{
		Fn:     ex.Fn,
		Status: ex.Status,
		Res:    ex.Res,
		Err:    ex.Err,
		Prog:   ex.Prog,
		Done:   ex.Done,
	}
}

func (t TaskFunc) Go(ctx context.Context, params map[string]any, cleanup func(params map[string]any)) *TaskExecution {
	ex := &TaskExecution{
		Fn: t,
	}
	go func() {
		res, err := t(ctx, params, func(prog float64) {
			ex.Lock.Lock()
			ex.Prog = min(max(prog, 0), 1)
			ex.Lock.Unlock()
		}, func(s string) {
			ex.Lock.Lock()
			ex.Status = s
			ex.Lock.Unlock()
		})
		ex.Lock.Lock()
		ex.Err = err
		if err == nil {
			ex.Res = res
		}
		ex.Prog = 1
		ex.Done = true
		ex.Lock.Unlock()
		if cleanup != nil {
			cleanup(params)
		}
	}()
	return ex
}

func (taskExecution *TaskExecution) Draw(name string) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	ex := taskExecution.Snap()

	text := name
	if i := strings.Index(text, "##"); i != -1 {
		text = text[:i]
	}
	if ex.Status != "" {
		if text != "" {
			text += ": "
		}
		text += ex.Status
	}
	imgui.TextUnformatted(text)
	imgui.ProgressBar(float32(ex.Prog))
}

// Collection of predefined tasks
var Tasks = struct {
	// Params:
	//   outPath: destination path (only present if successful)
	//   url: download URL
	//   progressStatus (optional): write progress in mebibytes to status
	// Results: none
	Download TaskFunc
	// Params:
	//   path: path to archive
	//   outDir: destination folder
	//   stripFirstDir (optional): strip first directory from contained file paths
	// Results:
	//   extractedFiles: slice of extracted file names
	Unarchive TaskFunc
}{
	Download: func(ctx context.Context, params map[string]any, onProgress func(prog float64), onStatus func(string)) (_ map[string]any, err error) {
		outPath := params["outPath"].(string)
		url := params["url"].(string)
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

		f, err := os.Create(outPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		defer func() {
			if err != nil {
				os.Remove(outPath)
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
					onStatus(fmt.Sprintf("Downloading (%3.1f/%3.1f MiB)", float64(currentBytes)/mebi, float64(resp.ContentLength)/mebi))
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
		outDir := params["outDir"].(string)
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
			path := filepath.Clean(filepath.Join(outDir, nameInArchive))
			if !strings.HasPrefix(path, outDir) {
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
		return map[string]any{"extractedFileNames": extractedFileNames}, nil
	},
}
