// Implements LSb-First reader and writer for reading/writing arbitrary bit width
// integers from/to normal go byte streams.
// An example use case is Vorbis coding.
// Note that this implementation is completely unoptimized.
// The Writer needs to cache the byte it is about write, so FlushByte() must be
// called to ensure the stream is re-aligned to 8bits and all data is written.
package bitio

import (
	"io"
)

// Writes individual bits.
// Bit order: LSb-First
type Writer struct {
	io.Writer

	totalWritten int
	bitBuf       uint8
	bitBufLen    uint8
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

// n may be inaccurate if writing fails within byte
// boundaries.
func (w *Writer) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if _, err := w.WriteBits(uint64(b), 8); err != nil {
			return n, err
		}
		n++
	}
	return
}

func (w *Writer) WriteBit(b bool) error {
	if b {
		w.bitBuf |= 1 << w.bitBufLen
	}
	w.bitBufLen++
	w.totalWritten++

	if w.bitBufLen == 8 {
		if err := w.FlushByte(); err != nil {
			return err
		}
	}
	return nil
}

// Writes nb bits of b in LSb-First order.
// Returns number of bits successfully written.
func (w *Writer) WriteBits(b uint64, nb uint8) (n int, err error) {
	if nb >= 64 {
		panic("attempt to write more than 64 bits of uint64")
	}
	for i := uint8(0); i < nb; i++ {
		if err := w.WriteBit((b & (1 << i)) != 0); err != nil {
			return int(i), err
		}
	}
	return n, nil
}

// Arg format is: value, nBits, value, nBits...
// Returns number of bits successfully written.
func (w *Writer) WriteBitsMany(args ...uint64) (n int, err error) {
	if len(args)&1 != 0 {
		panic("expected even number of arguments")
	}
	for i := 0; i < len(args)/2; i++ {
		nb := args[2*i+1]
		if nb >= 64 {
			panic("attempt to write more than 64 bits of uint64")
		}
		nw, err := w.WriteBits(args[2*i], uint8(nb))
		if err != nil {
			return n, err
		}
		n += nw
	}
	return
}

func (w *Writer) FlushByte() error {
	if w.IsAligned() {
		return nil
	}
	b := [1]byte{w.bitBuf}
	if _, err := w.Writer.Write(b[:]); err != nil {
		return err
	}
	w.bitBuf = 0
	w.bitBufLen = 0
	return nil
}

// Returns true if the writer is currently 8-bit aligned
// (and thus flushing will do nothing).
func (w *Writer) IsAligned() bool {
	return w.bitBufLen == 0
}

func (w *Writer) BitsWritten() int {
	return w.totalWritten
}

// Reads individual bits.
// Bit order: LSb-First
type Reader struct {
	io.Reader

	totalRead int
	bitBuf    uint8
	bitBufLen uint8
}

func NewReader(r io.Reader) *Reader {
	return &Reader{Reader: r}
}

// n may be inaccurate if reading fails within byte
// boundaries.
func (r *Reader) Read(p []byte) (n int, err error) {
	for n < len(p) {
		v, _, err := r.ReadBits(8)
		if err != nil {
			return n, err
		}
		p[n] = uint8(v)
		n++
	}
	return
}

func (r *Reader) ReadBit() (bool, error) {
	if r.bitBufLen == 0 {
		var buf [1]byte
		if _, err := r.Reader.Read(buf[:]); err != nil {
			return false, err
		}
		r.bitBuf = buf[0]
		r.bitBufLen = 8
	}

	r.totalRead++
	r.bitBufLen--
	return (r.bitBuf & (0x80 >> r.bitBufLen)) != 0, nil
}

func (r *Reader) ReadBits(nb uint8) (val uint64, n int, err error) {
	if nb >= 64 {
		panic("attempt to read more than 64 bits into uint64")
	}
	var res uint64
	for i := uint8(0); i < nb; i++ {
		b, err := r.ReadBit()
		if err != nil {
			return 0, int(i), err
		}
		if b {
			res |= 1 << i
		}
	}
	return res, int(nb), nil
}

// Returns true if the reader is currently 8-bit aligned
// (and thus is in sync with the underlying reader)
func (r *Reader) IsAligned() bool {
	return r.bitBufLen == 0
}

func (r *Reader) BitsRead() int {
	return r.totalRead
}
