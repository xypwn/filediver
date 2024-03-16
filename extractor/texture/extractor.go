package texture

import (
	"encoding/binary"

	"io"
	"os"

	"github.com/xypwn/filediver/extractor"
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

func extract(ins [stingray.NumDataType]io.ReadSeeker, _ extractor.Config) (io.Reader, string, header, error) {
	var hdr header
	if err := binary.Read(ins[stingray.DataMain], binary.LittleEndian, &hdr); err != nil {
		return nil, "", header{}, err
	}
	var magicNum [4]byte
	if _, err := ins[stingray.DataMain].Read(magicNum[:]); err != nil {
		return nil, "", header{}, err
	}
	if _, err := ins[stingray.DataMain].Seek(int64(-len(magicNum)), io.SeekCurrent); err != nil {
		return nil, "", header{}, err
	}
	readers := []io.Reader{ins[stingray.DataMain]}
	if ins[stingray.DataStream] == nil {
		readers = append(readers, ins[stingray.DataGPU])
	} else {
		readers = append(readers, ins[stingray.DataStream])
	}
	return io.MultiReader(readers...), string(magicNum[:]), hdr, nil
}

func Extract(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
	r, magicNum, _, err := extract(ins, config)
	if err != nil {
		return err
	}

	var fileExt string
	switch magicNum {
	case "DDS ":
		fileExt = ".dds"
	}

	out, err := os.Create(outPath + fileExt)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, r); err != nil {
		return err
	}

	return nil
}
