package texture

import (
	"fmt"
	"os"

	"io"

	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

func extract(outPath string, ins [stingray.NumDataType]io.ReadSeeker, convert func(w io.Writer, r io.Reader) error) error {
	tex, err := texture.Load(ins[stingray.DataMain])
	if err != nil {
		return err
	}

	if _, err := ins[stingray.DataMain].Seek(int64(tex.HeaderOffset), io.SeekStart); err != nil {
		return err
	}
	var magicNum [4]byte
	if _, err := ins[stingray.DataMain].Read(magicNum[:]); err != nil {
		return err
	}

	if _, err := ins[stingray.DataMain].Seek(int64(tex.HeaderOffset), io.SeekStart); err != nil {
		return err
	}

	switch string(magicNum[:]) {
	case "DDS ":
		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()
		readers := []io.Reader{ins[stingray.DataMain]}
		if ins[stingray.DataStream] != nil {
			readers = append(readers, ins[stingray.DataStream])
		}
		if ins[stingray.DataGPU] != nil {
			readers = append(readers, ins[stingray.DataGPU])
		}
		if err := convert(out, io.MultiReader(readers...)); err != nil {
			out.Close()
			if err := os.Remove(outPath); err != nil {
				return nil
			}
			return err
		}
	default:
		return fmt.Errorf("unrecognized texture format (magic number: \"%v\")", magicNum)
	}

	return nil
}

func Extract(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, _ extractor.GetResourceFunc) error {
	return extract(outPath+".dds", ins, func(w io.Writer, r io.Reader) error {
		if _, err := io.Copy(w, r); err != nil {
			return err
		}
		return nil
	})
}

func Convert(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config, runner *exec.Runner, _ extractor.GetResourceFunc) error {
	return extract(outPath+".png", ins, func(w io.Writer, r io.Reader) error {
		return runner.Run("magick", w, r, "dds:-", "png:-")
	})
}
