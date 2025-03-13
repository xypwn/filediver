import numpy as np
import struct
import zlib
from io import BytesIO
from enum import IntEnum
from typing import List, Tuple, Optional
from typing_extensions import TypedDict

from logging import Logger

logger = Logger("OpenEXR.types")

EXR_MAGIC = 20000630

# Reads a null terminated string, returning all bytes except the null terminator
def read_cstr(data: BytesIO) -> str:
    to_return = ""
    char = data.read(1)
    while char != b'\0':
        to_return += char.decode()
        char = data.read(1)
    return to_return

class NotAnEXRFileException(ValueError):
    def __init__(self, *args):
        super().__init__(*args)

class UnsupportedEXRException(ValueError):
    def __init__(self, *args):
        super().__init__(*args)

class PixelType(IntEnum):
    UINT = 0
    HALF = 1
    FLOAT = 2

    def to_numpy(self) -> type:
        match (self):
            case (self.UINT):
                return np.uint32
            case (self.HALF):
                return np.float16
            case (self.FLOAT):
                return np.float32

    @classmethod
    def from_numpy(cls, dtype) -> 'PixelType':
        match(dtype):
            case (np.uint32):
                return cls.UINT
            case (np.float16):
                return cls.HALF
            case (np.float32):
                return cls.FLOAT

class Compression(IntEnum):
    NONE = 0
    RLE = 1
    ZIPS = 2
    ZIP = 3
    PIZ = 4
    PXR24 = 5
    B44 = 6
    B44A = 7
    DWAA = 8
    DWAB = 9

class LineOrder(IntEnum):
    INCREASING_Y = 0
    DECREASING_Y = 1
    RANDOM_Y = 2

class Box2i:
    def __init__(self, xMin: int, yMin: int, xMax: int, yMax: int) -> None:
        self.xMin = xMin
        self.yMin = yMin
        self.xMax = xMax
        self.yMax = yMax

    def serialize(self) -> bytes:
        return struct.pack("<IIIII", 16, self.xMin, self.yMin, self.xMax, self.yMax)

    @classmethod
    def deserialize(cls, data: BytesIO) -> 'Box2i':
        return cls(*(struct.unpack("<IIIII", data.read(20))[1:]))

class Box2f:
    def __init__(self, xMin: float, yMin: float, xMax: float, yMax: float) -> None:
        self.xMin = xMin
        self.yMin = yMin
        self.xMax = xMax
        self.yMax = yMax

    def serialize(self) -> bytes:
        return struct.pack("<Iffff", 16, self.xMin, self.yMin, self.xMax, self.yMax)

    @classmethod
    def deserialize(cls, data: BytesIO) -> 'Box2f':
        return cls(*(struct.unpack("<Iffff", data.read(20))[1:]))

class Channel:
    def __init__(self, name: str, pixeltype: PixelType, pLinear: int, xSampling: int, ySampling: int) -> None:
        self.name = name
        self.pixeltype = pixeltype
        self.pLinear = pLinear
        self.xSampling = xSampling
        self.ySampling = ySampling

    def serialize(self) -> bytes:
        name = self.name.encode()
        length = len(name)
        if not name.endswith(b"\0"):
            length += 1
        return struct.pack(f"<{length}sIBxxxII", name, int(self.pixeltype), self.pLinear, self.xSampling, self.ySampling)

    @classmethod
    def deserialize(cls, data: BytesIO) -> 'Channel':
        name = read_cstr(data)
        pixeltype = PixelType(struct.unpack("<I", data.read(4))[0])
        return cls(name, pixeltype, *struct.unpack("<BxxxII", data.read(12)))

class Attributes(TypedDict):
    channels: List[Channel]
    compression: Compression
    dataWindow: Box2i
    displayWindow: Box2i
    lineOrder: LineOrder
    pixelAspectRatio: float
    screenWindowCenter: Tuple[float, float]
    screenWindowWidth: float

    @classmethod
    def serialize(cls, attrs: 'Attributes') -> bytes:
        data = b""
        for key, value in attrs.items():
            if key in Attributes.__annotations__:
                key_type = Attributes.__annotations__[key]
                data += struct.pack(f"<{len(key.encode())+1}s", key.encode())
                if key_type == float:
                    data += struct.pack("<6sIf", b"float", 4, value)
                elif key_type == Tuple[float, float]:
                    data += struct.pack("<4sIff", b"v2f", 8, *value)
                elif key_type == Compression:
                    data += struct.pack("<12sIB", b"compression", 1, int(value))
                elif key_type == LineOrder:
                    data += struct.pack("<10sIB", b"lineOrder", 1, int(value))
                elif key_type == Box2f:
                    data += struct.pack("<6s", b"box2f") + value.serialize()
                elif key_type == Box2i:
                    data += struct.pack("<6s", b"box2i") + value.serialize()
                elif key_type == List[Channel]:
                    channel_data = b""
                    for channel in value:
                        channel_data += channel.serialize()
                    channel_data += b"\0"
                    data += struct.pack("<7sI", b"chlist", len(channel_data)) + channel_data
            else:
                logger.warning(f"Skipping unknown attribute {key} of type {key_type}")

        data += b"\0"
        return data

    @classmethod
    def deserialize(cls, data: BytesIO) -> 'Attributes':
        attributes = {}

        attr_name = read_cstr(data)
        while attr_name != "":
            attr_type = read_cstr(data)
            attr_size = struct.unpack("<I", data.read(4))[0]
            attr_data = data.read(attr_size)
            attr_reader = BytesIO(attr_data)
            match(attr_name, attr_type):
                case ("channels", "chlist"):
                    attributes["channels"] = []
                    while attr_reader.tell() < attr_size - 1:
                        attributes["channels"].append(Channel.deserialize(attr_reader))
                case ("compression", "compression"):
                    attributes["compression"] = Compression(attr_data[0])
                case ("dataWindow", "box2i"):
                    attr_reader = BytesIO(struct.pack("<I", attr_size) + attr_data)
                    attributes["dataWindow"] = Box2i.deserialize(attr_reader)
                case ("displayWindow", "box2i"):
                    attr_reader = BytesIO(struct.pack("<I", attr_size) + attr_data)
                    attributes["displayWindow"] = Box2i.deserialize(attr_reader)
                case ("lineOrder", "lineOrder"):
                    attributes["lineOrder"] = LineOrder(attr_data[0])
                case ("pixelAspectRatio", "float"):
                    attributes["pixelAspectRatio"] = struct.unpack("<f", attr_data)[0]
                case ("screenWindowCenter", "v2f"):
                    attributes["screenWindowCenter"] = struct.unpack("<2f", attr_data)
                case ("screenWindowWidth", "float"):
                    attributes["screenWindowWidth"] = struct.unpack("<f", attr_data)[0]
                case (_, _):
                    logger.warning(f"Unknown attribute '{attr_name}' of type '{attr_type}'")
            attr_name = read_cstr(data)
        return cls(attributes)

class Scanline:
    def __init__(self, y_coord: int, data_size: int, pixel_data: bytes, pixels: Optional[np.ndarray] = None):
        self.y_coord = y_coord
        self.data_size = data_size
        self.data = pixel_data

        self.__pixels = pixels
        self.__adler32 = zlib.adler32(self.data)

    def pixels(self, attrs: Attributes) -> np.ndarray:
        if self.__pixels is None or self.__adler32 != zlib.crc32(self.data):
            self.decompress(attrs)
        return self.__pixels

    def serialize(self) -> bytes:
        return struct.pack("<II", self.y_coord, self.data_size) + self.data

    @classmethod
    def __reconstruct(cls, data: bytes) -> bytes:
        output = [0] * len(data)
        output[0] = data[0]
        for i in range(1, len(data)):
            output[i] = (output[i - 1] + data[i] - 128) & 0xff
        return bytes(output)

    @classmethod
    def __interleave(cls, data: bytes) -> bytes:
        output = [0] * len(data)
        for i in range(len(data) // 2):
            output[i * 2] = data[i]
            output[i * 2 + 1] = data[i + (len(data) + 1) // 2]
        return bytes(output)

    @classmethod
    def __serialize_pixels(cls, pixels: np.ndarray) -> bytes:
        depth = pixels.shape[-1]
        split = np.dsplit(pixels, pixels.shape[-1])
        if depth == 4:
            r, g, b, a = split
        else:
            r, g, b = split
        data = b""
        for i in range(pixels.shape[0]):
            if depth == 4:
                data += a[i].tobytes()
            data += b[i].tobytes()
            data += g[i].tobytes()
            data += r[i].tobytes()
        return data

    @classmethod
    def __reorder(cls, data: bytes) -> bytes:
        output = [0] * len(data)
        x = 0
        for i in range(len(data) // 2):
            output[i] = data[x]
            output[i + (len(data) + 1) // 2] = data[x+1]
            x += 2
        return bytes(output)
    
    @classmethod
    def __deconstruct(cls, data: bytes) -> bytes:
        output = [0] * len(data)
        output[0] = data[0]
        p = data[0]
        for i in range(1, len(data)):
            d = (int(data[i]) - p + 0x180) & 0xff
            p = data[i]
            output[i] = d
        return bytes(output)

    def decompress(self, attrs: Attributes) -> None:
        match(attrs["compression"]):
            case (Compression.NONE):
                pixel_data = self.data
                max_rows = 1
            case (Compression.ZIP):
                decompressed = zlib.decompress(self.data)
                reconstructed = self.__reconstruct(decompressed)
                pixel_data = self.__interleave(reconstructed)
                max_rows = 16
            case (_):
                raise NotImplementedError(f"{attrs['compression']}")
        dataWindow: Box2i = attrs["dataWindow"]
        width = dataWindow.xMax - dataWindow.xMin + 1
        channelList: List[Channel] = attrs["channels"]
        rows = []
        offset = 0
        last_dtype = None
        warn_mixed_dtype = True
        for _ in range(max_rows):
            channels: List[np.ndarray] = []
            for channel in channelList:
                dtype = channel.pixeltype.to_numpy()
                if last_dtype is not None and last_dtype != dtype and warn_mixed_dtype:
                    logger.warning("Mixed channel data types might lead to oddities!")
                    warn_mixed_dtype = False
                last_dtype = dtype
                channels.append(np.frombuffer(pixel_data, dtype, width, offset))
                offset += channels[-1].nbytes
            rows.append(np.dstack(channels[::-1]))
            if offset >= len(pixel_data):
                break
        self.__pixels = np.vstack(rows)

    def compress(self, attrs: Attributes) -> None:
        if self.__pixels is None:
            logger.error("Attempting to compress data that has not been decompressed!")
            return
        pixel_data = self.__serialize_pixels(self.__pixels)
        match(attrs["compression"]):
            case (Compression.NONE):
                self.data_size = len(pixel_data)
                self.data = pixel_data
            case (Compression.ZIP):
                reordered = self.__reorder(pixel_data)
                deconstructed = self.__deconstruct(reordered)
                compressed = zlib.compress(deconstructed)
                use_compressed = len(compressed) < len(pixel_data)
                self.data_size = len(compressed) if use_compressed else len(pixel_data)
                self.data = compressed if use_compressed else pixel_data
            case (_):
                raise NotImplementedError(str(attrs["compression"]))

    def best_compression(self, compare: List[Compression] = [Compression.NONE, Compression.ZIP]) -> Optional[Compression]:
        if self.__pixels is None:
            logger.error("No pixels to attempt to compress")
            return None
        pixel_data = self.__serialize_pixels(self.__pixels)
        sizes = [0xFFFFFFFF] * len(compare)
        for i, method in enumerate(compare):
            match(method):
                case (Compression.NONE):
                    sizes[i] = len(pixel_data)
                case (Compression.ZIP):
                    reordered = self.__reorder(pixel_data)
                    deconstructed = self.__deconstruct(reordered)
                    compressed = zlib.compress(deconstructed)
                    sizes[i] = len(compressed)
                case (_):
                    continue
        index = sizes.index(min(sizes))
        return compare[index]

    @classmethod
    def deserialize(cls, data: BytesIO) -> 'Scanline':
        y_coord, data_size = struct.unpack("<II", data.read(8))
        pixel_data = data.read(data_size)
        return cls(y_coord, data_size, pixel_data)

    @classmethod
    def from_pixels(cls, y_coord: int, data: np.ndarray) -> 'Scanline':
        pixel_data = cls.__serialize_pixels(data)
        return cls(y_coord, len(pixel_data), pixel_data, pixels=data)


class OpenEXR:
    def __init__(
            self,
            magic: int,
            version: int,
            attributes: Attributes,
            offset_table: List[int],
            scanlines: List[Scanline]
        ):
        self.magic = magic
        self.version = version
        self.attributes = attributes
        self.offset_table = offset_table
        self.scanlines = scanlines

    def serialize(self) -> bytes:
        data = struct.pack("<IBxxx", self.magic, self.version)
        data += Attributes.serialize(self.attributes)

        offset_table = []

        offset = len(data) + len(self.scanlines) * 8
        total_scanline_data = b""

        for scanline in self.scanlines:
            offset_table.append(offset)
            data += struct.pack("<Q", offset)
            scanline_data = scanline.serialize()
            offset += len(scanline_data)
            total_scanline_data += scanline_data

        self.offset_table = offset_table

        data += total_scanline_data
        return data
    
    @classmethod
    def deserialize(cls, data: BytesIO) -> 'OpenEXR':
        magic, version, flags0, flags1, flags2 = struct.unpack("<IBBBB", data.read(8))
        if magic != EXR_MAGIC:
            raise NotAnEXRFileException(f"invalid magic {magic:08x}")
        if version > 2:
            raise UnsupportedEXRException(f"unsupported EXR version {version}")
        if not (flags0 == flags1 == flags2 == 0x0):
            raise UnsupportedEXRException(f"Uusupported flags in EXR: {flags0:02x}{flags1:02x}{flags2:02x}")

        attributes = Attributes.deserialize(data)

        missing_attributes = []
        for key in Attributes.__annotations__:
            if key not in attributes:
                missing_attributes.append(key)
        if missing_attributes:
            join_exp = "\", \""
            raise UnsupportedEXRException(f"required attributes \"{join_exp.join(missing_attributes)}\" missing")


        scanline_count = 0
        match(attributes["compression"]):
            case (Compression.NONE):
                scanline_count = attributes["dataWindow"].yMax - attributes["dataWindow"].yMin + 1
            case (Compression.ZIP):
                height = attributes["dataWindow"].yMax - attributes["dataWindow"].yMin + 1
                scanline_count = height // 16 + (1 if height % 16 != 0 else 0)
            case (_):
                raise UnsupportedEXRException(f"compression type {attributes['compression'].name} not supported")

        offset_table = struct.unpack(f"<{scanline_count}Q", data.read(8 * scanline_count))
        scanlines: List[Scanline] = [Scanline.deserialize(data) for _ in range(scanline_count)]

        return cls(magic, version, attributes, offset_table, scanlines)

    def pixels(self) -> np.ndarray:
        pixels = []
        for scanline in self.scanlines:
            pixels.append(scanline.pixels(self.attributes))
        return np.vstack(pixels)

    @classmethod
    def from_pixels(cls, pixels: np.ndarray) -> 'OpenEXR':
        height, width, channels = pixels.shape
        if height < 1:
            raise ValueError("empty pixel array")
        if channels not in [3, 4]:
            raise ValueError("only RGB or RGBA channel layouts are supported")
        channel_names = "ABGR"
        if channels == 3:
            channel_names = "BGR"
        attributes: Attributes = {}
        attributes["channels"] = [Channel(name, PixelType.from_numpy(pixels.dtype), 0, 1, 1) for name in channel_names]
        attributes["dataWindow"] = Box2i(0, 0, width - 1, height - 1)
        attributes["displayWindow"] = Box2i(0, 0, width - 1, height - 1)
        attributes["lineOrder"] = LineOrder.INCREASING_Y
        attributes["pixelAspectRatio"] = 1.0
        attributes["screenWindowCenter"] = (0.0, 0.0)
        attributes["screenWindowWidth"] = 1.0

        scanline_pixels = np.vsplit(pixels, [16])
        scanlines: List[Scanline] = []
        y_coord = 0
        for line in scanline_pixels:
            if line.shape[0] == 0:
                break
            scanlines.append(Scanline.from_pixels(y_coord, line))
            y_coord += line.shape[0]

        if height > 1:
            attributes["compression"] = Compression.ZIP
        else:
            compression = scanlines[0].best_compression()
            attributes["compression"] = compression
        
        for scanline in scanlines:
            scanline.compress(attributes)

        exr = cls(EXR_MAGIC, 2, attributes, [], scanlines)
        exr.serialize()

        return exr
