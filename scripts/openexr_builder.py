import zlib
import numpy as np
import struct

# Very limited EXR writer - supports files of up to 16 scanlines
# which must use RGBA float 32 pixels
#
# Purpose built for converting Helldiver 2 LUTs from DDS R16G16B16A16_FLOAT to EXR RGBA Float32


PIXELTYPE_FLOAT = 2
COMPRESSION_ZIP = 3
LINEORDER_INC_Y = 0

class Box2i:
    def __init__(self, xMin: int, yMin: int, xMax: int, yMax: int) -> None:
        self.xMin = xMin
        self.yMin = yMin
        self.xMax = xMax
        self.yMax = yMax

    def serialize(self) -> bytes:
        return struct.pack("<IIII", self.xMin, self.yMin, self.xMax, self.yMax)

class Box2f:
    def __init__(self, xMin: float, yMin: float, xMax: float, yMax: float) -> None:
        self.xMin = xMin
        self.yMin = yMin
        self.xMax = xMax
        self.yMax = yMax

    def serialize(self) -> bytes:
        return struct.pack("<ffff", self.xMin, self.yMin, self.xMax, self.yMax)

class ChannelList:
    def __init__(self, name: str, pixeltype: int, pLinear: int, xSampling: int, ySampling: int) -> None:
        self.name = name
        self.pixeltype = pixeltype
        self.pLinear = pLinear
        self.xSampling = xSampling
        self.ySampling = ySampling
    
    def serialize(self) -> bytes:
        return struct.pack("<2sIBxxxII", self.name.encode(), self.pixeltype, self.pLinear, self.xSampling, self.ySampling)

def compress_pixels(pixels: np.ndarray) -> bytes:
    r, g, b, a = np.dsplit(pixels, pixels.shape[-1])
    data = b""
    for i in range(pixels.shape[0]):
        data += a[i].tobytes()
        data += b[i].tobytes()
        data += g[i].tobytes()
        data += r[i].tobytes()
    reordered = [0 for _ in range(len(data))]
    x = 0
    for i in range(len(data) // 2):
        reordered[i] = data[x]
        reordered[i + (len(data) + 1) // 2] = data[x+1]
        x += 2
    
    #mangle
    p = reordered[0]
    for i in range(1, len(reordered)):
        d = (int(reordered[i]) - p + 0x180) & 0xff
        p = reordered[i]
        reordered[i] = d
    
    return zlib.compress(bytes(reordered))

def make_exr(pixels: np.ndarray) -> bytes:
    assert pixels.dtype == np.float32
    height, width, channels = pixels.shape
    assert channels == 4, "Only RGBA image types supported"
    assert height <= 16, "Max 16 scanlines supported"
    # magic, version
    data = struct.pack("<IBxxx", 20000630, 2)
    # start attributes
    data += b"channels\0chlist\0"
    chlist_data = b""
    for char in "ABGR":
        chlist = ChannelList(char, PIXELTYPE_FLOAT, 0, 1, 1)
        chlist_data += chlist.serialize()
    chlist_data += b"\0"
    data += struct.pack("<I", len(chlist_data)) + chlist_data
    data += b"compression\0compression\0" + struct.pack("<IB", 1, COMPRESSION_ZIP)
    window = Box2i(0, 0, width - 1, height - 1)
    data += b"dataWindow\0box2i\0" + struct.pack("<I", 16) + window.serialize()
    data += b"displayWindow\0box2i\0" + struct.pack("<I", 16) + window.serialize()
    data += b"lineOrder\0lineOrder\0" + struct.pack("<IB", 1, LINEORDER_INC_Y)
    data += b"pixelAspectRatio\0float\0" + struct.pack("<If", 4, 1.0)
    data += b"screenWindowCenter\0v2f\0" + struct.pack("<Iff", 8, 0.0, 0.0)
    data += b"screenWindowWidth\0float\0" + struct.pack("<If", 4, 1.0)
    data += b"\0"
    # start offset table
    scanlineOffset = len(data) + 8
    data += struct.pack("<QI", scanlineOffset, 0)
    compressed = compress_pixels(pixels)
    data += struct.pack("<I", len(compressed)) + compressed
    return data