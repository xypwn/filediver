package stingray_package

import (
	"encoding/binary"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type Header struct {
	Magic     [4]uint8
	Unk00     uint32
	FileCount uint32
	Unk01     uint32
}

type Item struct {
	Type stingray.Hash
	Name stingray.Hash
}

type Package struct {
	Header
	Items []Item
}

func LoadPackage(r io.Reader) (*Package, error) {
	header := Header{}
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	items := make([]Item, header.FileCount)
	if err := binary.Read(r, binary.LittleEndian, items); err != nil {
		return nil, err
	}
	return &Package{
		Header: header,
		Items:  items,
	}, nil
}
