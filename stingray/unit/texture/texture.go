package texture

import (
	"encoding/binary"
	"io"
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
