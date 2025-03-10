package wwise

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
	"github.com/xypwn/filediver/wwise"
)

type stingrayBnkHeader struct {
	Unk00 [4]byte
	Size  uint32
	Name  stingray.Hash
}

func OpenRawBnk(in io.ReadSeeker) (io.ReadSeeker, error) {
	fileSize, err := in.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	var hdr stingrayBnkHeader
	if err := binary.Read(in, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if int64(hdr.Size+0x10) != fileSize {
		return nil, fmt.Errorf("size specified in header (%v) does not match actual file size (%v)", hdr.Size+0x10, fileSize)
	}

	return util.NewSectionReadSeeker(
		in,
		0x10,
		fileSize-0x10,
	)
}

func OpenBnk(in io.ReadSeeker) (*wwise.Bnk, error) {
	bnkIn, err := OpenRawBnk(in)
	if err != nil {
		return nil, err
	}

	return wwise.OpenBnk(bnkIn, &wwise.BkhdXorKey{
		/* https://github.com/Xaymar/Hellextractor/issues/25 */
		// "reverse-engineer" the key in code:
		Version: 0x0000008c ^ 0x9211bc20,
		ID:      0x50c63a23 ^ 0xf3d64a1b,
	})
}
