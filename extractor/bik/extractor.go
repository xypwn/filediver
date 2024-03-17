package bik

import (
	"encoding/binary"
	"io"
	"os"
	"os/exec"

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

func Extract(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
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

func Convert(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		ffmpegPath, err = exec.LookPath("./ffmpeg")
	}
	if err != nil {
		return Extract(outPath, ins, config)
	}

	r, _, err := extract(ins, config)
	if err != nil {
		return err
	}

	cmd := exec.Command(ffmpegPath, "-y", "-f", "bink", "-i", "pipe:", outPath+".mp4")
	cmd.Stdin = r
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
