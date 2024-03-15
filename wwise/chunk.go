package wwise

import (
	"encoding/binary"
	"io"
)

type chunk struct {
	Type   [4]byte
	Size   uint32
	Offset uint32
}

type chunkScanner struct {
	r      io.ReadSeeker
	endian binary.ByteOrder
	pos    uint32
	end    uint32
	err    error
	chunk  chunk
}

func newChunkScanner(r io.ReadSeeker, pos, end uint32, endian binary.ByteOrder) *chunkScanner {
	return &chunkScanner{
		r:      r,
		endian: endian,
		pos:    pos,
		end:    end,
	}
}

func (c *chunkScanner) Next() bool {
	if c.pos >= c.end {
		return false
	}
	if _, err := c.r.Seek(int64(c.pos), io.SeekStart); err != nil {
		c.err = err
		return false
	}
	if err := binary.Read(c.r, c.endian, &c.chunk.Type); err != nil {
		c.err = err
		return false
	}
	if err := binary.Read(c.r, c.endian, &c.chunk.Size); err != nil {
		c.err = err
		return false
	}
	c.chunk.Offset = c.pos + 0x08
	c.pos += 0x08 + c.chunk.Size
	return true
}

func (c *chunkScanner) Chunk() chunk {
	return c.chunk
}

func (c *chunkScanner) Err() error {
	return c.err
}
