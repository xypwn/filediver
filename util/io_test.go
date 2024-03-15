package util_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/xypwn/filediver/util"
)

func TestSectionReadSeeker(t *testing.T) {
	r, err := util.NewSectionReadSeeker(
		bytes.NewReader([]byte("Hello, world!")),
		2,
		10,
	)
	if err != nil {
		t.Error(err)
	}
	// Basic read
	{
		var buf [5]byte
		if n, err := r.Read(buf[:]); err != nil {
			t.Error(err)
		} else if n != 5 {
			t.Errorf("wrong num bytes read: %v", n)
		} else if string(buf[:]) != "llo, " {
			t.Errorf("unexpected value: \"%v\"", string(buf[:]))
		}
	}
	// Read beyond set limit
	{
		var buf [6]byte
		if n, err := r.Read(buf[:]); err != io.EOF {
			t.Errorf("expected io.EOF, but got: %v", err)
		} else if n != 5 {
			t.Errorf("wrong num bytes read: %v", n)
		} else if string(buf[:n]) != "world" {
			t.Errorf("unexpected value: \"%v\"", string(buf[:n]))
		}
	}
	// Seek
	{
		if n, err := r.Seek(5, io.SeekStart); err != nil {
			t.Error(err)
		} else if n != 5 {
			t.Errorf("unexpected pos after seek: %v", n)
		}
		var buf [2]byte
		if n, err := r.Read(buf[:]); err != nil {
			t.Error(err)
		} else if n != 2 {
			t.Errorf("wrong num bytes read: %v", n)
		} else if string(buf[:]) != "wo" {
			t.Errorf("unexpected value: \"%v\"", string(buf[:]))
		}
	}
}
