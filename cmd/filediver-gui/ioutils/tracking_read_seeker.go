package ioutils

import (
	"io"
	"sync/atomic"
)

type trackingReadSeeker struct {
	r   io.ReadSeeker
	pos *atomic.Int64
}

// A tracking reader allows you to thread-safely track the current read position
// using an atomic int64.
func NewTrackingReadSeeker(r io.ReadSeeker, posVar *atomic.Int64) *trackingReadSeeker {
	return &trackingReadSeeker{
		r,
		posVar,
	}
}

func (r *trackingReadSeeker) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.pos.Add(int64(n))
	return n, err
}

func (r *trackingReadSeeker) Seek(offset int64, whence int) (int64, error) {
	i, err := r.r.Seek(offset, whence)
	r.pos.Store(i)
	return i, err
}
