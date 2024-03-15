// The following is mostly manually converted from vgmstream (https://github.com/vgmstream/vgmstream)
package wwise

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/xypwn/filediver/util"
)

type bnkSections struct {
	DidxOffset uint32
	DidxSize   uint32
	DataOffset uint32
	DataSize   uint32
	HircOffset uint32
	HircSize   uint32
}

type bnkHeader struct {
	MagicNum [4]byte
	HdrSize  uint32
	Version  uint32
	ID       uint32
}

type bnkIndex struct {
	ID     uint32
	Offset uint32
	Size   uint32
}

type Bnk struct {
	r        io.ReadSeeker
	sections bnkSections
	files    []bnkIndex
}

type BkhdXorKey struct {
	Version uint32
	ID      uint32
}

func openBnk(r io.ReadSeeker, bkhdKey *BkhdXorKey) (*Bnk, error) {
	// Get file size
	var fileSize int64
	{
		var err error
		fileSize, err = r.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	// Read header
	var hdr bnkHeader
	{
		if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
			return nil, err
		}

		// https://github.com/Xaymar/Hellextractor/issues/25
		if bkhdKey != nil {
			hdr.Version ^= bkhdKey.Version
			hdr.ID ^= bkhdKey.ID
		}

		if string(hdr.MagicNum[:]) != "BKHD" {
			return nil, errors.New("missing BKHD magic number")
		}
		if hdr.Version <= 26 {
			return nil, errors.New("unsupported version")
		}
	}

	// Read sections
	var sections bnkSections
	{
		sc := newChunkScanner(r, 0x08+hdr.HdrSize, uint32(fileSize), binary.LittleEndian)
		for sc.Next() {
			ck := sc.Chunk()
			typeStr := string(ck.Type[:])
			switch typeStr {
			case "BKHD":
				return nil, errors.New("more than 1 header chunk")
			case "DIDX":
				sections.DidxOffset = ck.Offset
				sections.DidxSize = ck.Size
			case "DATA":
				sections.DataOffset = ck.Offset
				sections.DataSize = ck.Size
			case "HIRC":
				// Contains events, tracks, sequences etc.
				// Not handled in this implementation.
				sections.HircOffset = ck.Offset
				sections.HircSize = ck.Size
			case "INIT", "STID", "STMG", "ENVS", "PLAT":
				// Not handled in this implementation.
			default:
				return nil, fmt.Errorf("unsupported chunk type: \"%v\"", typeStr)
			}
			if ck.Offset+ck.Size > uint32(fileSize) {
				return nil, errors.New("broken .bnk")
			}
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
	}

	// Read DIDX (contained file info)
	var files []bnkIndex
	if sections.DidxSize > 0 && sections.DataSize > 0 {
		if _, err := r.Seek(int64(sections.DidxOffset), io.SeekStart); err != nil {
			return nil, err
		}
		for i := uint32(0); i < sections.DidxSize; i += 0x0c {
			var file bnkIndex
			if err := binary.Read(r, binary.LittleEndian, &file); err != nil {
				return nil, err
			}
			files = append(files, file)
		}
	}

	return &Bnk{
		r:        r,
		sections: sections,
		files:    files,
	}, nil
}

// bkhdKey is an optional XOR key to decode bnk version and ID, if required
func OpenBnk(r io.ReadSeeker, bkhdKey *BkhdXorKey) (*Bnk, error) {
	const errPfx = errPfx + "OpenBnk: "

	res, err := openBnk(r, bkhdKey)
	if err != nil {
		return nil, fmt.Errorf("%v%w", errPfx, err)
	}

	return res, nil
}

func (b *Bnk) NumFiles() int {
	return len(b.files)
}

func (b *Bnk) FileID(idx int) uint32 {
	return b.files[idx].ID
}

// Only one file should be read at the same time.
func (b *Bnk) OpenFile(idx int) (io.ReadSeeker, error) {
	const errPfx = errPfx + "Bnk: OpenFile: "

	f := b.files[idx]
	r, err := util.NewSectionReadSeeker(
		b.r,
		int64(b.sections.DataOffset+f.Offset),
		int64(f.Size),
	)
	if err != nil {
		return nil, fmt.Errorf("%v%w", errPfx, err)
	}
	return r, nil
}
