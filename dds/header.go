package dds

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// https://learn.microsoft.com/en-us/windows/win32/direct3ddds/dx-graphics-dds-reference

type CapsFlags uint32

const (
	CapsFlag    CapsFlags = 1 << 3
	CapsTexture CapsFlags = 1 << 12
	CapsMipMap  CapsFlags = 1 << 22
)

type Caps2Flags uint32

const (
	Caps2Cubemap       Caps2Flags = 1 << 9
	Caps2CubemapPlusX  Caps2Flags = 1 << 10
	Caps2CubemapMinusX Caps2Flags = 1 << 11
	Caps2CubemapPlusY  Caps2Flags = 1 << 12
	Caps2CubemapMinusY Caps2Flags = 1 << 13
	Caps2CubemapPlusZ  Caps2Flags = 1 << 14
	Caps2CubemapMinusZ Caps2Flags = 1 << 15
	Caps2Volume        Caps2Flags = 1 << 21
)

type PixelFormatFlags uint32

const (
	PixelFormatFlagAlphaPixels PixelFormatFlags = 1 << 0
	PixelFormatFlagAlpha       PixelFormatFlags = 1 << 1
	PixelFormatFlagFourCC      PixelFormatFlags = 1 << 2
	PixelFormatFlagRGB         PixelFormatFlags = 1 << 6
	PixelFormatFlagYUV         PixelFormatFlags = 1 << 9
	PixelFormatFlagLuminance   PixelFormatFlags = 1 << 17
)

type PixelFormat struct {
	Size        uint32
	Flags       PixelFormatFlags
	FourCC      [4]uint8
	RGBBitCount uint32
	RBitMask    uint32
	GBitMask    uint32
	BBitMask    uint32
	ABitMask    uint32
}

type HeaderFlags uint32

const (
	HeaderFlagCaps        HeaderFlags = 1 << 0
	HeaderFlagHeight      HeaderFlags = 1 << 1
	HeaderFlagWidth       HeaderFlags = 1 << 2
	HeaderFlagPitch       HeaderFlags = 1 << 3
	HeaderFlagPixelFormat HeaderFlags = 1 << 12
	HeaderFlagMipMapCount HeaderFlags = 1 << 17
	HeaderFlagLinearsize  HeaderFlags = 1 << 19
	HeaderFlagDepth       HeaderFlags = 1 << 23
)

type Header struct {
	Size              uint32
	Flags             HeaderFlags
	Height            uint32
	Width             uint32
	PitchOrLinearSize uint32
	Depth             uint32
	MipMapCount       uint32
	Reserved          [11]uint32
	PixelFormat       PixelFormat
	Caps              CapsFlags
	Caps2             Caps2Flags
	Caps3             uint32
	Caps4             uint32
	Reserved2         uint32
}

func DecodeHeader(r io.Reader) (Header, error) {
	var magicNum [4]uint8
	if err := binary.Read(r, binary.LittleEndian, magicNum[:]); err != nil {
		return Header{}, err
	}
	if magicNum != [4]uint8{'D', 'D', 'S', ' '} {
		return Header{}, errors.New("invalid magic number")
	}

	var hdr Header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return Header{}, err
	}

	if hdr.Size != 0x7c {
		return Header{}, fmt.Errorf("invalid header size: %v", hdr.Size)
	}

	if hdr.Flags&HeaderFlagCaps == 0 ||
		hdr.Flags&HeaderFlagWidth == 0 ||
		hdr.Flags&HeaderFlagHeight == 0 ||
		hdr.Flags&HeaderFlagPixelFormat == 0 {
		return Header{}, errors.New("required header flags missing (required: Caps | Width | Height | PixelFormat)")
	}

	if hdr.PixelFormat.Size != 0x20 {
		return Header{}, fmt.Errorf("invalid pixel format header size: %v", hdr.PixelFormat.Size)
	}

	return hdr, nil
}
