package vorbis

import (
	"github.com/xypwn/filediver/bitio"
)

// Returns number of bytes needed to store v.
func ilog(v uint64) uint8 {
	var n uint8
	for v != 0 {
		v >>= 1
		n++
	}
	return n
}

func copyBits(bw *bitio.Writer, br *bitio.Reader, nBits uint8) (uint64, error) {
	val, _, err := br.ReadBits(nBits)
	if err != nil {
		return 0, err
	}
	if _, err := bw.WriteBits(val, nBits); err != nil {
		return 0, err
	}
	return val, nil
}
