package bik

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
)

type header struct {
	Unk00 [16]byte
}

func extract(ins [stingray.NumDataType]io.ReadSeeker, _ extractor.Config) (io.Reader, header, error) {
	var hdr header
	if err := binary.Read(ins[stingray.DataMain], binary.LittleEndian, &hdr); err != nil {
		return nil, header{}, err
	}

	readers := []io.Reader{ins[stingray.DataMain]}
	if ins[stingray.DataStream] == nil {
		readers = append(readers, ins[stingray.DataGPU])
	} else {
		readers = append(readers, ins[stingray.DataStream])
	}
	return io.MultiReader(readers...), hdr, nil
}

func Extract(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, _ extractor.GetResourceFunc) error {
	r, _, err := extract(ins, config)
	if err != nil {
		return err
	}

	out, err := os.Create(outPath + ".bik")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func Convert(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, getResource extractor.GetResourceFunc) error {
	if !runner.Has("ffmpeg") {
		return Extract(outPath, ins, config, runner, getResource)
	}

	r, _, err := extract(ins, config)
	if err != nil {
		return err
	}

	return runner.Run(
		"ffmpeg",
		nil,
		r,
		"-f", "bink",
		"-i", "pipe:",
		outPath+".mp4",
	)
}
