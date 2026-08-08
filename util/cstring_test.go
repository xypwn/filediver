package util_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xypwn/filediver/util"
)

func TestReadCString(t *testing.T) {
	cases := []struct {
		name    string
		input   []byte
		want    string
		wantErr error
	}{
		{"simple", []byte("hello\x00"), "hello", nil},
		{"empty", []byte("\x00"), "", nil},
		{"empty then garbage", []byte("\x00extra"), "", nil},
		{"no null terminator", []byte("hello"), "hello", io.ErrUnexpectedEOF},
		{"empty no null", []byte{}, "", io.ErrUnexpectedEOF},
		{"leading nulls not skipped", []byte("\x00\x00hello\x00"), "", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := util.ReadCString(bytes.NewReader(c.input))
			require.Equal(t, c.want, got)
			require.ErrorIs(t, err, c.wantErr)
		})
	}
}

func TestReadCStringSkipLeadingZeroes(t *testing.T) {
	cases := []struct {
		name    string
		input   []byte
		want    string
		wantErr error
	}{
		{"simple", []byte("hello\x00"), "hello", nil},
		{"empty", []byte("\x00"), "", io.ErrUnexpectedEOF},
		{"no null terminator", []byte("hello"), "hello", io.ErrUnexpectedEOF},
		{"empty no null", []byte{}, "", io.ErrUnexpectedEOF},
		{"single leading null skipped", []byte("\x00hello\x00"), "hello", nil},
		{"multiple leading nulls skipped", []byte("\x00\x00\x00hello\x00"), "hello", nil},
		{"all nulls", []byte("\x00\x00\x00"), "", io.ErrUnexpectedEOF},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := util.ReadCStringSkipLeadingNulls(bytes.NewReader(c.input))
			require.Equal(t, c.want, got)
			require.ErrorIs(t, err, c.wantErr)
		})
	}
}
