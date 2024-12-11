import struct
import numpy as np
from io import BytesIO
from typing import Any
from enum import IntEnum, auto

class DXGIFormat(IntEnum):
    @staticmethod
    def _generate_next_value_(name: str, start: int, count: int, last_values: list[Any]) -> Any:
        return count
    UNKNOWN = auto()
    R32G32B32A32_TYPELESS = auto()
    R32G32B32A32_FLOAT = auto()
    R32G32B32A32_UINT = auto()
    R32G32B32A32_SINT = auto()
    R32G32B32_TYPELESS = auto()
    R32G32B32_FLOAT = auto()
    R32G32B32_UINT = auto()
    R32G32B32_SINT = auto()
    R16G16B16A16_TYPELESS = auto()
    R16G16B16A16_FLOAT = auto()
    R16G16B16A16_UNORM = auto()
    R16G16B16A16_UINT = auto()
    R16G16B16A16_SNORM = auto()
    R16G16B16A16_SINT = auto()
    R32G32_TYPELESS = auto()
    R32G32_FLOAT = auto()
    R32G32_UINT = auto()
    R32G32_SINT = auto()
    R32G8X24_TYPELESS = auto()
    D32_FLOAT_S8X24_UINT = auto()
    R32_FLOAT_X8X24_TYPELESS = auto()
    X32_TYPELESS_G8X24_UINT = auto()
    R10G10B10A2_TYPELESS = auto()
    R10G10B10A2_UNORM = auto()
    R10G10B10A2_UINT = auto()
    R11G11B10_FLOAT = auto()
    R8G8B8A8_TYPELESS = auto()
    R8G8B8A8_UNORM = auto()
    R8G8B8A8_UNORM_SRGB = auto()
    R8G8B8A8_UINT = auto()
    R8G8B8A8_SNORM = auto()
    R8G8B8A8_SINT = auto()
    R16G16_TYPELESS = auto()
    R16G16_FLOAT = auto()
    R16G16_UNORM = auto()
    R16G16_UINT = auto()
    R16G16_SNORM = auto()
    R16G16_SINT = auto()
    R32_TYPELESS = auto()
    D32_FLOAT = auto()
    R32_FLOAT = auto()
    R32_UINT = auto()
    R32_SINT = auto()
    R24G8_TYPELESS = auto()
    D24_UNORM_S8_UINT = auto()
    R24_UNORM_X8_TYPELESS = auto()
    X24_TYPELESS_G8_UINT = auto()
    R8G8_TYPELESS = auto()
    R8G8_UNORM = auto()
    R8G8_UINT = auto()
    R8G8_SNORM = auto()
    R8G8_SINT = auto()
    R16_TYPELESS = auto()
    R16_FLOAT = auto()
    D16_UNORM = auto()
    R16_UNORM = auto()
    R16_UINT = auto()
    R16_SNORM = auto()
    R16_SINT = auto()
    R8_TYPELESS = auto()
    R8_UNORM = auto()
    R8_UINT = auto()
    R8_SNORM = auto()
    R8_SINT = auto()
    A8_UNORM = auto()
    R1_UNORM = auto()
    R9G9B9E5_SHAREDEXP = auto()
    R8G8_B8G8_UNORM = auto()
    G8R8_G8B8_UNORM = auto()
    BC1_TYPELESS = auto()
    BC1_UNORM = auto()
    BC1_UNORM_SRGB = auto()
    BC2_TYPELESS = auto()
    BC2_UNORM = auto()
    BC2_UNORM_SRGB = auto()
    BC3_TYPELESS = auto()
    BC3_UNORM = auto()
    BC3_UNORM_SRGB = auto()
    BC4_TYPELESS = auto()
    BC4_UNORM = auto()
    BC4_SNORM = auto()
    BC5_TYPELESS = auto()
    BC5_UNORM = auto()
    BC5_SNORM = auto()
    B5G6R5_UNORM = auto()
    B5G5R5A1_UNORM = auto()
    B8G8R8A8_UNORM = auto()
    B8G8R8X8_UNORM = auto()
    R10G10B10_XR_BIAS_A2_UNORM = auto()
    B8G8R8A8_TYPELESS = auto()
    B8G8R8A8_UNORM_SRGB = auto()
    B8G8R8X8_TYPELESS = auto()
    B8G8R8X8_UNORM_SRGB = auto()
    BC6H_TYPELESS = auto()
    BC6H_UF16 = auto()
    BC6H_SF16 = auto()
    BC7_TYPELESS = auto()
    BC7_UNORM = auto()
    BC7_UNORM_SRGB = auto()
    AYUV = auto()
    Y410 = auto()
    Y416 = auto()
    NV12 = auto()
    P010 = auto()
    P016 = auto()
    _420_OPAQUE = auto()
    YUY2 = auto()
    Y210 = auto()
    Y216 = auto()
    NV11 = auto()
    AI44 = auto()
    IA44 = auto()
    P8 = auto()
    A8P8 = auto()
    B4G4R4A4_UNORM = auto()
    P208 = auto()
    V208 = auto()
    V408 = auto()

class DX10ResourceDimension(IntEnum):
    @staticmethod
    def _generate_next_value_(name: str, start: int, count: int, last_values: list[Any]) -> Any:
        return count
    Unknown = auto()
    Buffer = auto()
    Texture1D = auto()
    Texture2D = auto()
    Texture3D = auto()

class DX10Header:
    def __init__(self, dxgifmt: DXGIFormat, resdim: DX10ResourceDimension, miscflags: int, arraysize: int, miscflags2: int) -> None:
        self.dxgifmt = dxgifmt
        self.resdim = resdim
        self.miscflags = miscflags
        self.arraysize = arraysize
        self.miscflags2 = miscflags2

    @classmethod
    def parse(cls, data: BytesIO) -> 'DX10Header':
        dxgifmt_int, resdim, miscflags, arraysize, miscflags2 = struct.unpack("<5I", data.read(20))
        return cls(DXGIFormat(dxgifmt_int), DX10ResourceDimension(resdim), miscflags, arraysize, miscflags2)

class DDSPixelFormat:
    def __init__(self, size: int, flags: int, fourcc: bytes, bitcount: int, redmask: int, grnmask: int, blumask: int, alpmask: int) -> None:
        self.size = size
        self.flags = flags
        self.fourcc = fourcc
        self.bitcount = bitcount
        self.redmask = redmask
        self.grnmask = grnmask
        self.blumask = blumask
        self.alpmask = alpmask
    
    @classmethod
    def parse(cls, data: BytesIO) -> 'DDSPixelFormat':
        size, flags, fourcc, bitcount, redmask, grnmask, blumask, alpmask = struct.unpack("<II4sIIIII", data.read(32))
        return cls(size, flags, fourcc, bitcount, redmask, grnmask, blumask, alpmask)

class DDSHeader:
    def __init__(self, magic: bytes, size: int, flags: int, height: int, width: int, pitch: int, depth: int, mipmaps: int, dds_pix_fmt: DDSPixelFormat, caps: int, caps2: int, caps3: int, caps4: int, dx10header: DX10Header):
        self.magic = magic
        self.size = size
        self.flags = flags
        self.height = height
        self.width = width
        self.pitch = pitch
        self.depth = depth
        self.mipmaps = mipmaps
        self.dds_pix_fmt = dds_pix_fmt
        self.caps = caps
        self.caps2 = caps2
        self.caps3 = caps3
        self.caps4 = caps4
        self.dx10header = dx10header
    
    @classmethod
    def parse(cls, data: BytesIO) -> 'DDSHeader':
        magic, size, flags, height, width, pitch, depth, mipmaps = struct.unpack("<4sIIIIIII", data.read(32))
        _ = struct.unpack("<11I", data.read(44))
        dds_pix_fmt = DDSPixelFormat.parse(data)
        caps, caps2, caps3, caps4, _ = struct.unpack("<5I", data.read(20))
        if dds_pix_fmt.fourcc == b"DX10":
            dx10header = DX10Header.parse(data)
        else:
            dx10header = None
        return cls(magic, size, flags, height, width, pitch, depth, mipmaps, dds_pix_fmt, caps, caps2, caps3, caps4, dx10header)


class DDS:
    def __init__(self, header: DDSHeader, data: bytes):
        self.header = header
        self.data = data

    @classmethod
    def parse(cls, data: BytesIO) -> 'DDS':
        return cls(DDSHeader.parse(data), data.read())

    # Returns the largest mipmap's pixels
    def pixels(self) -> np.ndarray:
        assert self.header.dx10header.dxgifmt == DXGIFormat.R16G16B16A16_FLOAT, "This script only reads the float16 RGBA format"
        stride = 8
        offset = 0
        pixels = []
        for _ in range(self.header.height):
            pixels.append([pixel for pixel in struct.iter_unpack("<eeee", self.data[offset:offset + stride * self.header.width])])
            offset += stride * self.header.width
        return np.array(pixels, dtype=np.float16)

