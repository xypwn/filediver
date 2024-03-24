package extractor

import (
	"io"
	"os"

	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
)

type Config map[string]string

type GetResourceFunc func(name, typ stingray.Hash) *stingray.File
type ExtractFunc func(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config Config, runner *exec.Runner, getResource GetResourceFunc) error

func ExtractFuncRaw(fileExt string) ExtractFunc {
	return func(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config Config, runner *exec.Runner, getResource GetResourceFunc) error {
		f, err := os.Create(outPath + "." + fileExt)
		if err != nil {
			return err
		}
		var readers []io.Reader
		for _, r := range ins[:] {
			if r != nil {
				readers = append(readers, r)
			}
		}
		if _, err := io.Copy(f, io.MultiReader(readers...)); err != nil {
			return err
		}
		return nil
	}
}
