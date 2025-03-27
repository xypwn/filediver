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
	DXGIFormatR16G16B16A16UNorm
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

func (f DXGIFormat) String() string {
	switch f {
	case DXGIFormatUnknown:
		return "Unknown"
	case DXGIFormatR32G32B32A32Typeless:
		return "R32G32B32A32Typeless"
	case DXGIFormatR32G32B32A32Float:
		return "R32G32B32A32Float"
	case DXGIFormatR32G32B32A32UInt:
		return "R32G32B32A32UInt"
	case DXGIFormatR32G32B32A32SInt:
		return "R32G32B32A32SInt"
	case DXGIFormatR32G32B32Typeless:
		return "R32G32B32Typeless"
	case DXGIFormatR32G32B32Float:
		return "R32G32B32Float"
	case DXGIFormatR32G32B32UInt:
		return "R32G32B32UInt"
	case DXGIFormatR32G32B32SInt:
		return "R32G32B32SInt"
	case DXGIFormatR16G16B16A16Typeless:
		return "R16G16B16A16Typeless"
	case DXGIFormatR16G16B16A16Float:
		return "R16G16B16A16Float"
	case DXGIFormatR16G16B16A16UNorm:
		return "R16G16B16A16UNorm"
	case DXGIFormatR16G16B16A16UInt:
		return "R16G16B16A16UInt"
	case DXGIFormatR16G16B16A16SNorm:
		return "R16G16B16A16SNorm"
	case DXGIFormatR16G16B16A16SInt:
		return "R16G16B16A16SInt"
	case DXGIFormatR32G32Typeless:
		return "R32G32Typeless"
	case DXGIFormatR32G32Float:
		return "R32G32Float"
	case DXGIFormatR32G32UInt:
		return "R32G32UInt"
	case DXGIFormatR32G32SInt:
		return "R32G32SInt"
	case DXGIFormatR32G8X24Typeless:
		return "R32G8X24Typeless"
	case DXGIFormatD32FLOATS8X24UInt:
		return "D32FLOATS8X24UInt"
	case DXGIFormatR32FLOATX8X24Typeless:
		return "R32FLOATX8X24Typeless"
	case DXGIFormatX32TYPELESSG8X24UInt:
		return "X32TYPELESSG8X24UInt"
	case DXGIFormatR10G10B10A2Typeless:
		return "R10G10B10A2Typeless"
	case DXGIFormatR10G10B10A2UNorm:
		return "R10G10B10A2UNorm"
	case DXGIFormatR10G10B10A2UInt:
		return "R10G10B10A2UInt"
	case DXGIFormatR11G11B10Float:
		return "R11G11B10Float"
	case DXGIFormatR8G8B8A8Typeless:
		return "R8G8B8A8Typeless"
	case DXGIFormatR8G8B8A8UNorm:
		return "R8G8B8A8UNorm"
	case DXGIFormatR8G8B8A8UNormSRGB:
		return "R8G8B8A8UNormSRGB"
	case DXGIFormatR8G8B8A8UInt:
		return "R8G8B8A8UInt"
	case DXGIFormatR8G8B8A8SNorm:
		return "R8G8B8A8SNorm"
	case DXGIFormatR8G8B8A8SInt:
		return "R8G8B8A8SInt"
	case DXGIFormatR16G16Typeless:
		return "R16G16Typeless"
	case DXGIFormatR16G16Float:
		return "R16G16Float"
	case DXGIFormatR16G16UNorm:
		return "R16G16UNorm"
	case DXGIFormatR16G16UInt:
		return "R16G16UInt"
	case DXGIFormatR16G16SNorm:
		return "R16G16SNorm"
	case DXGIFormatR16G16SInt:
		return "R16G16SInt"
	case DXGIFormatR32Typeless:
		return "R32Typeless"
	case DXGIFormatD32Float:
		return "D32Float"
	case DXGIFormatR32Float:
		return "R32Float"
	case DXGIFormatR32UInt:
		return "R32UInt"
	case DXGIFormatR32SInt:
		return "R32SInt"
	case DXGIFormatR24G8Typeless:
		return "R24G8Typeless"
	case DXGIFormatD24UnormS8UInt:
		return "D24UnormS8UInt"
	case DXGIFormatR24UnormX8Typeless:
		return "R24UnormX8Typeless"
	case DXGIFormatX24TypelessG8UInt:
		return "X24TypelessG8UInt"
	case DXGIFormatR8G8Typeless:
		return "R8G8Typeless"
	case DXGIFormatR8G8UNorm:
		return "R8G8UNorm"
	case DXGIFormatR8G8UInt:
		return "R8G8UInt"
	case DXGIFormatR8G8SNorm:
		return "R8G8SNorm"
	case DXGIFormatR8G8SInt:
		return "R8G8SInt"
	case DXGIFormatR16Typeless:
		return "R16Typeless"
	case DXGIFormatR16Float:
		return "R16Float"
	case DXGIFormatD16UNorm:
		return "D16UNorm"
	case DXGIFormatR16UNorm:
		return "R16UNorm"
	case DXGIFormatR16UInt:
		return "R16UInt"
	case DXGIFormatR16SNorm:
		return "R16SNorm"
	case DXGIFormatR16SInt:
		return "R16SInt"
	case DXGIFormatR8Typeless:
		return "R8Typeless"
	case DXGIFormatR8UNorm:
		return "R8UNorm"
	case DXGIFormatR8UInt:
		return "R8UInt"
	case DXGIFormatR8SNorm:
		return "R8SNorm"
	case DXGIFormatR8SInt:
		return "R8SInt"
	case DXGIFormatA8UNorm:
		return "A8UNorm"
	case DXGIFormatR1UNorm:
		return "R1UNorm"
	case DXGIFormatR9G9B9E5SharedExp:
		return "R9G9B9E5SharedExp"
	case DXGIFormatR8G8B8G8UNorm:
		return "R8G8B8G8UNorm"
	case DXGIFormatG8R8G8B8UNorm:
		return "G8R8G8B8UNorm"
	case DXGIFormatBC1Typeless:
		return "BC1Typeless"
	case DXGIFormatBC1UNorm:
		return "BC1UNorm"
	case DXGIFormatBC1UNormSRGB:
		return "BC1UNormSRGB"
	case DXGIFormatBC2Typeless:
		return "BC2Typeless"
	case DXGIFormatBC2UNorm:
		return "BC2UNorm"
	case DXGIFormatBC2UNormSRGB:
		return "BC2UNormSRGB"
	case DXGIFormatBC3Typeless:
		return "BC3Typeless"
	case DXGIFormatBC3UNorm:
		return "BC3UNorm"
	case DXGIFormatBC3UNormSRGB:
		return "BC3UNormSRGB"
	case DXGIFormatBC4Typeless:
		return "BC4Typeless"
	case DXGIFormatBC4UNorm:
		return "BC4UNorm"
	case DXGIFormatBC4SNorm:
		return "BC4SNorm"
	case DXGIFormatBC5Typeless:
		return "BC5Typeless"
	case DXGIFormatBC5UNorm:
		return "BC5UNorm"
	case DXGIFormatBC5SNorm:
		return "BC5SNorm"
	case DXGIFormatB5G6R5UNorm:
		return "B5G6R5UNorm"
	case DXGIFormatB5G5R5A1UNorm:
		return "B5G5R5A1UNorm"
	case DXGIFormatB8G8R8A8UNorm:
		return "B8G8R8A8UNorm"
	case DXGIFormatB8G8R8X8UNorm:
		return "B8G8R8X8UNorm"
	case DXGIFormatR10G10B10XRBiasA2UNorm:
		return "R10G10B10XRBiasA2UNorm"
	case DXGIFormatB8G8R8A8Typeless:
		return "B8G8R8A8Typeless"
	case DXGIFormatB8G8R8A8UNormSRGB:
		return "B8G8R8A8UNormSRGB"
	case DXGIFormatB8G8R8X8Typeless:
		return "B8G8R8X8Typeless"
	case DXGIFormatB8G8R8X8UNormSRGB:
		return "B8G8R8X8UNormSRGB"
	case DXGIFormatBC6HTypeless:
		return "BC6HTypeless"
	case DXGIFormatBC6HUF16:
		return "BC6HUF16"
	case DXGIFormatBC6HSF16:
		return "BC6HSF16"
	case DXGIFormatBC7Typeless:
		return "BC7Typeless"
	case DXGIFormatBC7UNorm:
		return "BC7UNorm"
	case DXGIFormatBC7UNormSRGB:
		return "BC7UNormSRGB"
	case DXGIFormatAYUV:
		return "AYUV"
	case DXGIFormatY410:
		return "Y410"
	case DXGIFormatY416:
		return "Y416"
	case DXGIFormatNV12:
		return "NV12"
	case DXGIFormatP010:
		return "P010"
	case DXGIFormatP016:
		return "P016"
	case DXGIFormat420Opaque:
		return "420Opaque"
	case DXGIFormatYUY2:
		return "YUY2"
	case DXGIFormatY210:
		return "Y210"
	case DXGIFormatY216:
		return "Y216"
	case DXGIFormatNV11:
		return "NV11"
	case DXGIFormatAI44:
		return "AI44"
	case DXGIFormatIA44:
		return "IA44"
	case DXGIFormatP8:
		return "P8"
	case DXGIFormatA8P8:
		return "A8P8"
	case DXGIFormatB4G4R4A4UNorm:
		return "B4G4R4A4UNorm"
	case DXGIFormatP208:
		return "P208"
	case DXGIFormatV208:
		return "V208"
	case DXGIFormatV408:
		return "V408"
	case DXGIFormatForceUInt:
		return "ForceUInt"
	default:
		panic("unknown DXGI format")
	}
}

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
