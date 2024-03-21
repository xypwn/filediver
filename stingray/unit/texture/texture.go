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

type Header struct {
	Unk00    [12]byte
	Sections [15]StreamableSection
}

type Texture struct {
	HeaderOffset uint32
	Sections     []StreamableSection
}

func Load(r io.Reader) (*Texture, error) {
	var hdr Header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	numSections := 15
	for i, sec := range hdr.Sections {
		if sec.Size == 0 {
			numSections = i
		}
	}
	return &Texture{
		HeaderOffset: 0xC0,
		Sections:     hdr.Sections[:numSections],
	}, nil
}
