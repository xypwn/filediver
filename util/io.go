package util

import (
	"io"
)

type SectionReadSeeker struct {
	r    io.ReadSeeker // const
	base int64         // const
	off  int64
	n    int64 // const
}

func NewSectionReadSeeker(r io.ReadSeeker, off int64, n int64) (*SectionReadSeeker, error) {
	_, err := r.Seek(off, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return &SectionReadSeeker{
		r:    r,
		base: off,
		off:  off,
		n:    n,
	}, nil
}

func (r *SectionReadSeeker) Read(p []byte) (n int, err error) {
	rem := r.base + r.n - r.off
	if rem < int64(len(p)) {
		n, err := r.r.Read(p[:rem])
		r.off += int64(n)
		if err != nil {
			return n, err
		}
		return n, io.EOF
	}
	n, err = r.r.Read(p)
	r.off += int64(n)
	return n, err
}

func (r *SectionReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.off = r.base + offset
	case io.SeekCurrent:
		r.off += offset
	case io.SeekEnd:
		r.off = r.base + r.n + offset
	default:
		panic("unhandled case")
	}
	pos, err := r.r.Seek(r.off, io.SeekStart)
	return pos - r.base, err
}
