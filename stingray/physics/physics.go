package physics

import (
	"encoding/binary"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Type                stingray.Hash
	Filehash            stingray.Hash
	Size                uint32
	_                   [12]byte
	NameEnd             [24]byte
	Unk00               [5]uint32
	PointerListOffset   uint32
	FirstSectionOffset  uint32
	SecondSectionOffset uint32
}

type Physics struct {
	Header
	// TODO: Add data chunks
}

func LoadPhysics(mainR io.ReadSeeker) (*Physics, error) {
	var hdr Header
	if err := binary.Read(mainR, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	return &Physics{
		Header: hdr,
	}, nil
}
