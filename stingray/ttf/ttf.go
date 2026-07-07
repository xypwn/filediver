package ttf

import (
	"encoding/binary"
	"io"
)

type Header struct {
	Size uint32
	_    [12]uint8
}

type TTF struct {
	Header
	Data []byte
}

func (x TTF) MarshalBinary() ([]byte, error) {
	return x.Data, nil
}

func LoadTTF(r io.Reader) (*TTF, error) {
	var hdr Header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	data := make([]byte, hdr.Size)
	if err := binary.Read(r, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return &TTF{
		Header: hdr,
		Data:   data,
	}, nil
}
