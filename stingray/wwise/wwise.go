package wwise

import (
	"bytes"
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

func BnkGetAllReferencedStreamData(
	in io.ReadSeeker,
	tryGetStreamByID func(id uint32) (data []byte, exists bool, err error),
) (map[uint32][]byte, error) {
	streams := make(map[uint32][]byte)

	bnk, err := OpenBnk(in)
	if err != nil {
		return nil, err
	}

	for i := 0; i < bnk.NumFiles(); i++ {
		id := bnk.FileID(i)
		// Stream should either exist as a wwise stream, or be embedded in the wwise bank file
		data, ok, err := tryGetStreamByID(id)
		if err != nil {
			return nil, err
		}
		if !ok {
			rd, err := bnk.OpenFile(i)
			if err != nil {
				return nil, err
			}
			b, err := io.ReadAll(rd)
			if err != nil {
				return nil, err
			}
			data = b
		}
		if len(data) >= 4 && bytes.Equal(data[:4], []byte{0, 4, 2, 0}) {
			// not actually a wwise_stream
		} else {
			streams[id] = data
		}
	}

	for _, obj := range bnk.HircObjects {
		// A source seems to exist when source bits > 0. I'm a bit unsure, though.
		/*if obj.Header.Type == wwise.BnkHircObjectSound {
			ctx.Warnf("%v", obj.Sound.SourceBits)
		}*/
		if obj.Header.Type == wwise.BnkHircObjectSound && obj.Sound.SourceBits > 0 {
			resourceID := obj.Sound.SourceID
			fileID := obj.Header.ObjectID
			data, ok, err := tryGetStreamByID(resourceID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("referenced wwise stream resource with file ID %v does not exist", resourceID)
			}
			streams[fileID] = data
		}
	}

	return streams, nil
}
