package extractor

import (
	"io"

	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
)

type Context interface {
	File() *stingray.File
	Runner() *exec.Runner
	Config() map[string]string
	GetResource(name, typ stingray.Hash) (file *stingray.File, exists bool)
	// Call WriteCloser.Close() when done.
	CreateFile(suffix string) (io.WriteCloser, error)
	// Call WriteCloser.Close() when done.
	CreateFileDir(dirSuffix, filename string) (io.WriteCloser, error)
	OutPath() (string, error)
}

type ExtractFunc func(ctx Context) error

func ExtractFuncRaw(suffix string, types ...stingray.DataType) ExtractFunc {
	return func(ctx Context) error {
		r, err := ctx.File().OpenMulti(types...)
		if err != nil {
			return err
		}
		defer r.Close()

		out, err := ctx.CreateFile(suffix)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, r); err != nil {
			return err
		}
		return nil
	}
}
