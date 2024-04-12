package texture

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
)

type StreamableSection struct {
	Offset uint32
	Size   uint32
	Width  uint16
	Height uint16
}

type Info struct {
	HeaderOffset uint32
	Sections     []StreamableSection
}

// Decode Stingray's texture info header wrapped around the DDS file.
// r must be from a file of type stingray.DataMain.
func DecodeInfo(r io.Reader) (*Info, error) {
	var hdr struct {
		Unk00    [12]byte
		Sections [15]StreamableSection
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	numSections := 15
	for i, sec := range hdr.Sections {
		if sec.Size == 0 {
			numSections = i
		}
	}
	return &Info{
		HeaderOffset: 0xC0, // size of hdr
		Sections:     hdr.Sections[:numSections],
	}, nil
}

func decode(f *stingray.File, readMipMaps bool) (*dds.DDS, error) {
	if !f.Exists(stingray.DataMain) {
		return nil, errors.New("no main data")
	}
	r, err := f.OpenMulti(stingray.DataMain, stingray.DataStream, stingray.DataGPU)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if _, err := DecodeInfo(r); err != nil {
		return nil, err
	}

	tex, err := dds.Decode(r, readMipMaps)
	if err != nil {
		return nil, fmt.Errorf("dds: %w", err)
	}
	return tex, nil
}

// Decode DDS texture with Stingray wrapper.
func Decode(f *stingray.File, readMipMaps bool) (*dds.DDS, error) {
	tex, err := decode(f, readMipMaps)
	if err != nil {
		return nil, fmt.Errorf("stingray texture: %w", err)
	}
	return tex, nil
}
