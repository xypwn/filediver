package bitio_test

import (
	"bytes"
	"testing"

	"github.com/xypwn/filediver/bitio"
)

func TestBitio(t *testing.T) {
	b := &bytes.Buffer{}

	// Writing
	w := bitio.NewWriter(b)
	if _, err := w.WriteBits(0b1100, 4); err != nil {
		t.Error(err)
	}
	if _, err := w.WriteBits(0b111, 3); err != nil {
		t.Error(err)
	}
	if _, err := w.WriteBits(0b0010001, 7); err != nil {
		t.Error(err)
	}
	if _, err := w.WriteBitsMany(
		0b1101100111001, 13,
		0b00000, 5,
	); err != nil {
		t.Error(err)
	}
	if !w.IsAligned() {
		t.Error("expected w to be byte-aligned")
	}

	// Reading
	r := bitio.NewReader(b)
	if v, _, err := r.ReadBits(8); err != nil {
		t.Error(err)
	} else if v != 0b11111100 {
		t.Errorf("unexpected value: %08b", v)
	}
	if v, _, err := r.ReadBits(7); err != nil {
		t.Error(err)
	} else if v != 0b1001000 {
		t.Errorf("unexpected value: %07b", v)
	}
	if v, _, err := r.ReadBits(9); err != nil {
		t.Error(err)
	} else if v != 0b110011100 {
		t.Errorf("unexpected value: %09b", v)
	}
	if v, _, err := r.ReadBits(8); err != nil {
		t.Error(err)
	} else if v != 0b00000110 {
		t.Errorf("unexpected value: %08b", v)
	}
	if !r.IsAligned() {
		t.Error("expected r to be byte-aligned")
	}

	// Alignment and flushing
	if _, err := w.WriteBits(0b1111, 4); err != nil {
		t.Error(err)
	}
	if w.IsAligned() {
		t.Error("expected w to not be byte-aligned")
	}
	if err := w.FlushByte(); err != nil {
		t.Error(err)
	}
	if !w.IsAligned() {
		t.Error("expected w to be byte-aligned")
	}
	if v, _, err := r.ReadBits(6); err != nil {
		t.Error(err)
	} else if v != 0b001111 {
		t.Errorf("unexpected value: %06b", v)
	}
	if r.IsAligned() {
		t.Error("expected r to not be byte-aligned")
	}
}
