// The following is manually converted from vgmstream (https://github.com/vgmstream/vgmstream)
package vorbis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/jfreymuth/vorbis"
)

type Config struct {
	Channels      uint16
	SampleRate    uint32
	Blocksize0Exp uint8
	Blocksize1Exp uint8

	SetupID   uint32
	Endian    binary.ByteOrder
	StreamEnd uint32
}

type wPacket struct {
	HeaderSize uint16
	PacketSize uint16
	HasNext    bool
	NextB      uint8 // first byte of next packet
}

func readPacket(d *Decoder, r io.ReadSeeker, isSetup bool) (wPacket, []byte, error) {
	var wp wPacket
	wp.HeaderSize = 0x02
	if err := binary.Read(r, d.cfg.Endian, &wp.PacketSize); err != nil {
		return wPacket{}, nil, err
	}
	if wp.PacketSize == 0 {
		return wPacket{}, nil, errors.New("invalid packet size: 0")
	}
	readSize := wp.PacketSize
	if !isSetup {
		// peek into next mod packet'd first byte
		readSize += wp.HeaderSize + 0x01
	}

	data := make([]byte, readSize)
	read, err := io.ReadFull(r, data)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return wPacket{}, nil, err
	}

	if !isSetup && read == int(readSize) {
		wp.HasNext = true
		wp.NextB = data[wp.PacketSize+wp.HeaderSize]
	} else {
		wp.HasNext = false
	}

	if !isSetup {
		// un-peek into next mod packet's first byte
		if _, err := r.Seek(-int64(wp.HeaderSize+0x01), io.SeekCurrent); err != nil {
			return wPacket{}, nil, err
		}
	}

	return wp, data, nil
}

// Convert wwise setup to vorbis setup
func convertSetup(d *Decoder, w io.Writer, r io.ReadSeeker) (wPacket, error) {
	wp, data, err := readPacket(d, r, true)
	if err != nil {
		return wPacket{}, err
	}

	if err := ww2oggConvertSetup(d, w, bytes.NewReader(data)); err != nil {
		return wPacket{}, err
	}

	return wp, nil
}

// Convert wwise data packet into vorbis data packet
func convertPacket(d *Decoder, w io.Writer, r io.ReadSeeker) (wPacket, error) {
	wp, data, err := readPacket(d, r, false)
	if err != nil {
		return wPacket{}, err
	}

	if err := ww2oggConvertPacket(d, w, wp, bytes.NewReader(data)); err != nil {
		return wPacket{}, err
	}

	return wp, nil
}

type Decoder struct {
	r             io.ReadSeeker
	vorbDec       *vorbis.Decoder
	buf           *bytes.Buffer
	cfg           Config
	sampleBuf     []float32
	modeBlockFlag [64 + 1]bool
	modeBits      uint8
	prevBlockFlag bool
}

func NewDecoder(r io.ReadSeeker, cfg Config) (*Decoder, error) {
	vorbDec := &vorbis.Decoder{}

	packetID := [6]byte{'v', 'o', 'r', 'b', 'i', 's'}

	buf := &bytes.Buffer{}

	// Identification packet
	{
		buf.Reset()
		pkt := struct {
			Type           uint8
			ID             [6]byte
			Version        uint32
			Channels       uint8
			SampleRate     uint32
			MaxBitrate     uint32
			NominalBitrate uint32
			MinimumBitrate uint32
			Blocksize      uint8
			FramingFlag    uint8
		}{
			Type:           0x01,     // packet type: id
			ID:             packetID, // id: always "vorbis"
			Version:        0x00,
			Channels:       uint8(cfg.Channels),
			SampleRate:     cfg.SampleRate,
			MaxBitrate:     0, // optional hint
			NominalBitrate: 0, // optional hint
			MinimumBitrate: 0, // optional hint
			Blocksize:      (cfg.Blocksize0Exp << 4) | cfg.Blocksize1Exp,
			FramingFlag:    0x01, // fixed
		}

		if err := binary.Write(buf, binary.LittleEndian, pkt); err != nil {
			return nil, err
		}

		if err := vorbDec.ReadHeader(buf.Bytes()); err != nil {
			return nil, err
		}
	}

	// Comment packet
	{
		buf.Reset()
		const vendor = "filediver wwise vorbis decoder"
		pkt := struct {
			Type               uint8
			ID                 [6]byte
			VendorLength       uint32
			VendorString       [len(vendor)]byte
			UserCommentListLen uint32
			FramingFlag        uint8
		}{
			Type:               0x03,     // packet type: comment
			ID:                 packetID, // id: always "vorbis"
			VendorLength:       uint32(len(vendor)),
			UserCommentListLen: 0,    // no user comments
			FramingFlag:        0x01, // fixed
		}
		copy(pkt.VendorString[:], vendor)

		if err := binary.Write(buf, binary.LittleEndian, pkt); err != nil {
			return nil, err
		}

		if err := vorbDec.ReadHeader(buf.Bytes()); err != nil {
			return nil, err
		}
	}

	// Setup packet
	d := &Decoder{
		r:         r,
		vorbDec:   vorbDec,
		buf:       buf,
		cfg:       cfg,
		sampleBuf: make([]float32, vorbDec.BufferSize()),
	}
	{
		buf.Reset()
		_, err := convertSetup(d, buf, r)
		if err != nil {
			return nil, err
		}
		if err := vorbDec.ReadHeader(buf.Bytes()); err != nil {
			return nil, err
		}
	}

	if !vorbDec.HeadersRead() {
		return nil, errors.New("headers not read")
	}

	return d, nil
}

// Decodes the next packet.
// Returned slice is only valid until next call of Decode.
func (d *Decoder) Decode() ([]float32, error) {
	d.buf.Reset()
	wp, err := convertPacket(d, d.buf, d.r)
	if err != nil {
		return nil, err
	}

	if !wp.HasNext {
		return nil, io.EOF
	}

	return d.vorbDec.DecodeInto(d.buf.Bytes(), d.sampleBuf)
}

// Maximum amount of samples that can be decoded from a single packet.
func (d *Decoder) BufferSize() int {
	return d.vorbDec.BufferSize()
}
