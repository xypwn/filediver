package texture

import (
	"encoding/binary"
	//"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"github.com/xypwn/filediver/converter"
	"github.com/xypwn/filediver/stingray"
)

type streamableSection struct {
	Offset uint32
	Size   uint32
	Width  uint16
	Height uint16
}
type header struct {
	Unk00    [12]byte
	Sections [15]streamableSection
}

func convertTexture(outPath string, in [3]io.ReadSeeker) error {
	var hdr header
	if err := binary.Read(in[stingray.DataMain], binary.LittleEndian, &hdr); err != nil {
		return err
	}
	var magicNum [4]byte
	if _, err := in[stingray.DataMain].Read(magicNum[:]); err != nil {
		return err
	}
	if _, err := in[stingray.DataMain].Seek(int64(-len(magicNum)), io.SeekCurrent); err != nil {
		return err
	}
	readers := []io.Reader{in[stingray.DataMain]}
	if in[stingray.DataStream] == nil {
		readers = append(readers, in[stingray.DataGPU])
	} else {
		readers = append(readers, in[stingray.DataStream])
	}
	imgIn := io.MultiReader(readers...)
	fileExt := ".texture"
	var img image.Image
	switch string(magicNum[:]) {
	case "DDS ":
		fileExt = ".dds"
	}
	if img != nil {
		fileExt = ".png"
	}
	out, err := os.Create(outPath + fileExt)
	if err != nil {
		return err
	}
	defer out.Close()
	if img == nil {
		if _, err := io.Copy(out, imgIn); err != nil {
			return err
		}
	} else {
		if err := png.Encode(out, img); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	converter.RegisterConverter(
		converter.FlagDataMain|converter.FlagDataStream|converter.FlagDataGPU,
		"texture",
		convertTexture,
	)
}
