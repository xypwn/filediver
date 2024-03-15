package bik

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/xypwn/filediver/converter"
	"github.com/xypwn/filediver/stingray"
)

type stingrayBikHeader struct {
	Unk00 [16]byte
}

func convertBik(outPath string, in [3]io.ReadSeeker) error {
	const errPfx = "bik: "

	var hdr stingrayBikHeader
	if err := binary.Read(in[stingray.DataMain], binary.LittleEndian, &hdr); err != nil {
		return fmt.Errorf("%v%w", errPfx, err)
	}

	var r io.Reader
	{
		readers := []io.Reader{in[stingray.DataMain]}
		if in[stingray.DataStream] == nil {
			readers = append(readers, in[stingray.DataGPU])
		} else {
			readers = append(readers, in[stingray.DataStream])
		}
		r = io.MultiReader(readers...)
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		ffmpegPath, err = exec.LookPath("./ffmpeg")
	}
	useFfmpeg := err == nil
	if useFfmpeg {
		cmd := exec.Command(ffmpegPath, "-y", "-f", "bink", "-i", "pipe:", outPath+".mp4")
		cmd.Stdin = r
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		out, err := os.Create(outPath + ".bik")
		if err != nil {
			return fmt.Errorf("%v%w", errPfx, err)
		}
		defer out.Close()
		if _, err := io.Copy(out, r); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	converter.RegisterConverter(
		converter.FlagDataMain|converter.FlagDataStream|converter.FlagDataGPU,
		"bik",
		convertBik,
	)
}
