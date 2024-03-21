package extractor

import (
	"io"

	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
)

type Config map[string]string

type GetResourceFunc func(name, typ stingray.Hash) *stingray.File
type ExtractFunc func(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config Config, runner *exec.Runner, getResource GetResourceFunc) error
