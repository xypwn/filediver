package util

import (
	"context"
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
		n, err := io.ReadFull(r.r, p[:rem])
		r.off += int64(n)
		if err != nil {
			return n, err
		}
		return n, io.EOF
	}
	n, err = io.ReadFull(r.r, p)
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

// Cancellable ReadSeekCloser type
type contextReadSeekCloser struct {
	io.ReadSeekCloser
	ctx context.Context
}

func NewContextReadSeekCloser(ctx context.Context, r io.ReadSeekCloser) io.ReadSeekCloser {
	return &contextReadSeekCloser{
		ReadSeekCloser: r,
		ctx:            ctx,
	}
}

func (r *contextReadSeekCloser) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.ReadSeekCloser.Read(p)
}

func (r *contextReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.ReadSeekCloser.Seek(offset, whence)
}
