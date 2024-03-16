package extractor

import (
	"io"

	"github.com/xypwn/filediver/stingray"
)

type Config map[string]string

type ExtractFunc func(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config Config) error
