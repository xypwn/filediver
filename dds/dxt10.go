package dds

import (
	"encoding/binary"
	"io"
)

// https://learn.microsoft.com/en-us/windows/win32/direct3ddds/dds-header-dxt10

type DXGIFormat uint32

const (
	DXGIFormatUnknown DXGIFormat = iota
	DXGIFormatR32G32B32A32Typeless
	DXGIFormatR32G32B32A32Float
	DXGIFormatR32G32B32A32UInt
	DXGIFormatR32G32B32A32SInt
	DXGIFormatR32G32B32Typeless
	DXGIFormatR32G32B32Float
	DXGIFormatR32G32B32UInt
	DXGIFormatR32G32B32SInt
	DXGIFormatR16G16B16A16Typeless
	DXGIFormatR16G16B16A16Float
	DXGIFormatR16G16B16A16Unorm
	DXGIFormatR16G16B16A16UInt
	DXGIFormatR16G16B16A16SNorm
	DXGIFormatR16G16B16A16SInt
	DXGIFormatR32G32Typeless
	DXGIFormatR32G32Float
	DXGIFormatR32G32UInt
	DXGIFormatR32G32SInt
	DXGIFormatR32G8X24Typeless
	DXGIFormatD32FLOATS8X24UInt
	DXGIFormatR32FLOATX8X24Typeless
	DXGIFormatX32TYPELESSG8X24UInt
	DXGIFormatR10G10B10A2Typeless
	DXGIFormatR10G10B10A2UNorm
	DXGIFormatR10G10B10A2UInt
	DXGIFormatR11G11B10Float
	DXGIFormatR8G8B8A8Typeless
	DXGIFormatR8G8B8A8UNorm
	DXGIFormatR8G8B8A8UNormSRGB
	DXGIFormatR8G8B8A8UInt
	DXGIFormatR8G8B8A8SNorm
	DXGIFormatR8G8B8A8SInt
	DXGIFormatR16G16Typeless
	DXGIFormatR16G16Float
	DXGIFormatR16G16UNorm
	DXGIFormatR16G16UInt
	DXGIFormatR16G16SNorm
	DXGIFormatR16G16SInt
	DXGIFormatR32Typeless
	DXGIFormatD32Float
	DXGIFormatR32Float
	DXGIFormatR32UInt
	DXGIFormatR32SInt
	DXGIFormatR24G8Typeless
	DXGIFormatD24UnormS8UInt
	DXGIFormatR24UnormX8Typeless
	DXGIFormatX24TypelessG8UInt
	DXGIFormatR8G8Typeless
	DXGIFormatR8G8UNorm
	DXGIFormatR8G8UInt
	DXGIFormatR8G8SNorm
	DXGIFormatR8G8SInt
	DXGIFormatR16Typeless
	DXGIFormatR16Float
	DXGIFormatD16UNorm
	DXGIFormatR16UNorm
	DXGIFormatR16UInt
	DXGIFormatR16SNorm
	DXGIFormatR16SInt
	DXGIFormatR8Typeless
	DXGIFormatR8UNorm
	DXGIFormatR8UInt
	DXGIFormatR8SNorm
	DXGIFormatR8SInt
	DXGIFormatA8UNorm
	DXGIFormatR1UNorm
	DXGIFormatR9G9B9E5SharedExp
	DXGIFormatR8G8B8G8UNorm
	DXGIFormatG8R8G8B8UNorm
	DXGIFormatBC1Typeless
	DXGIFormatBC1UNorm
	DXGIFormatBC1UNormSRGB
	DXGIFormatBC2Typeless
	DXGIFormatBC2UNorm
	DXGIFormatBC2UNormSRGB
	DXGIFormatBC3Typeless
	DXGIFormatBC3UNorm
	DXGIFormatBC3UNormSRGB
	DXGIFormatBC4Typeless
	DXGIFormatBC4UNorm
	DXGIFormatBC4SNorm
	DXGIFormatBC5Typeless
	DXGIFormatBC5UNorm
	DXGIFormatBC5SNorm
	DXGIFormatB5G6R5UNorm
	DXGIFormatB5G5R5A1UNorm
	DXGIFormatB8G8R8A8UNorm
	DXGIFormatB8G8R8X8UNorm
	DXGIFormatR10G10B10XRBiasA2UNorm
	DXGIFormatB8G8R8A8Typeless
	DXGIFormatB8G8R8A8UNormSRGB
	DXGIFormatB8G8R8X8Typeless
	DXGIFormatB8G8R8X8UNormSRGB
	DXGIFormatBC6HTypeless
	DXGIFormatBC6HUF16
	DXGIFormatBC6HSF16
	DXGIFormatBC7Typeless
	DXGIFormatBC7UNorm
	DXGIFormatBC7UNormSRGB
	DXGIFormatAYUV
	DXGIFormatY410
	DXGIFormatY416
	DXGIFormatNV12
	DXGIFormatP010
	DXGIFormatP016
	DXGIFormat420Opaque
	DXGIFormatYUY2
	DXGIFormatY210
	DXGIFormatY216
	DXGIFormatNV11
	DXGIFormatAI44
	DXGIFormatIA44
	DXGIFormatP8
	DXGIFormatA8P8
	DXGIFormatB4G4R4A4UNorm
	DXGIFormatP208
	DXGIFormatV208
	DXGIFormatV408
	DXGIFormatForceUInt DXGIFormat = 0xffffffff
)

type D3D10ResourceDimension uint32

const (
	D3D10ResourceDimensionUnknown D3D10ResourceDimension = iota
	D3D10ResourceDimensionBuffer
	D3D10ResourceDimensionTexture1D
	D3D10ResourceDimensionTexture2D
	D3D10ResourceDimensionTexture3D
)

type D3D10ResourceMiscFlags uint32

const (
	D3D10ResourceMiscFlagGenerateMisc     D3D10ResourceMiscFlags = 1 << 0
	D3D10ResourceMiscFlagShared           D3D10ResourceMiscFlags = 1 << 1
	D3D10ResourceMiscFlagTextureCube      D3D10ResourceMiscFlags = 1 << 2
	D3D10ResourceMiscFlagSharedKeyedMutex D3D10ResourceMiscFlags = 1 << 4
	D3D10ResourceMiscFlagGDICompatible    D3D10ResourceMiscFlags = 1 << 5
)

type AlphaMode uint32

const (
	AlphaModeUnknown AlphaMode = iota
	AlphaModeStraight
	AlphaModePremultiplied
	AlphaModeOpaque
	AlphaModeCustom
)

type DXT10Header struct {
	DXGIFormat        DXGIFormat
	ResourceDimension D3D10ResourceDimension
	MiscFlag          D3D10ResourceMiscFlags
	ArraySize         uint32
	MiscFlags2        AlphaMode
}

func DecodeDXT10Header(r io.Reader) (DXT10Header, error) {
	var dxt10Hdr DXT10Header
	if err := binary.Read(r, binary.LittleEndian, &dxt10Hdr); err != nil {
		return DXT10Header{}, err
	}
	return dxt10Hdr, nil
}
