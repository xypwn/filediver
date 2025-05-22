from io import BytesIO, SEEK_END, SEEK_SET
from typing import Tuple, List, Union
from enum import IntEnum

import struct
import math

ROOT_2 = math.sqrt(2)

class PackedQuaternion:
    def __init__(self, first: int, second: int, third: int, largest: int):
        self.first = first
        self.second = second
        self.third = third
        self.largest = largest

    def quaternion(self) -> Tuple[float, float, float, float]:
        first_float = (float(self.first - 512) / 1024.0) * ROOT_2
        second_float = (float(self.second - 512) / 1024.0) * ROOT_2
        third_float = (float(self.third - 512) / 1024.0) * ROOT_2
        largest_float = 1.0 - math.sqrt(first_float ** 2 + second_float ** 2 + third_float ** 2)
        match self.largest:
            case 0:
                return (largest_float, first_float, second_float, third_float)
            case 1:
                return (first_float, largest_float, second_float, third_float)
            case 2:
                return (first_float, second_float, largest_float, third_float)
            case 3:
                return (first_float, second_float, third_float, largest_float)

    @classmethod
    def parse(cls, data: BytesIO) -> "PackedQuaternion":
        value = struct.unpack(">I", data.read(4))[0]
        largest = value >> 30
        third = (value >> 20) & 0x3ff
        second = (value >> 10) & 0x3ff
        first = value & 0x3ff
        return cls(first, second, third, largest)

class BoneInitialState:
    def __init__(self, position: Tuple[int], rotation: PackedQuaternion, scale: Tuple[float]):
        self.position = position
        self.rotation = rotation
        self.scale = scale

    @classmethod
    def parse(cls, data: BytesIO) -> "BoneInitialState":
        position = struct.unpack("<HHH", data.read(6))
        rotation = PackedQuaternion.parse(data)
        scale = struct.unpack("<eee", data.read(6))
        return cls(position, rotation, scale)

class Header:
    def __init__(self, unk00: int, boneCount: int, animationLength: float, size: int, unk01: Tuple[int, int], unk02: int, variableBits: List[int], boneStates: List[BoneInitialState]):
        self.unk00 = unk00
        self.boneCount = boneCount
        self.animationLength = animationLength
        self.size = size
        self.unk01 = unk01
        self.unk02 = unk02
        self.variableBits = variableBits
        self.boneStates = boneStates

    @classmethod
    def parse(cls, data: BytesIO) -> "Header":
        unk00, boneCount, animationLength, size = struct.unpack("<IIfI", data.read(16))
        unk01 = struct.unpack("<II", data.read(8))
        unk02 = struct.unpack("<H", data.read(2))[0]
        variableBits = []
        val = data.read(1)[0]
        while val & 0x80:
            variableBits.append(val)
            val = data.read(1)[0]
        variableBits.append(val)
        boneStates = [BoneInitialState.parse(data) for _ in range(boneCount)]

        return cls(unk00, boneCount, animationLength, size, unk01, unk02, variableBits, boneStates)

class EntryKind(IntEnum):
    UNKNOWN0 = 0
    UNKNOWN1 = 1
    POSITION = 2
    ROTATION = 3

class EntryHeader:
    def __init__(self, kind: EntryKind, bone: int, time: int):
        self.kind = kind
        self.bone = bone
        self.time = time

    @classmethod
    def parse(cls, data: BytesIO) -> "EntryHeader":
        first = struct.unpack("<H", data.read(2))[0]
        if first == 0x0003:
            raise StopIteration()
        second = struct.unpack("<H", data.read(2))[0]
        value = (first, second)
        kind = EntryKind(value[0] >> 14)
        bone = (value[0] >> 4) & 0x3ff
        time = ((value[0] & 0xf) << 16) | value[1]
        return cls(kind, bone, time)

class Entry:
    def __init__(self, header: EntryHeader, data: Union[PackedQuaternion, Tuple[int, int, int]]):
        self.header = header
        self.data = data

    @classmethod
    def parse(cls, data: BytesIO) -> "Entry":
        header = EntryHeader.parse(data)
        match header.kind:
            case EntryKind.ROTATION:
                entryData = PackedQuaternion.parse(data)
            case EntryKind.POSITION:
                entryData = struct.unpack("<HHH", data.read(6))
            case _:
                raise NotImplementedError(f"Unknown animation entry kind {int(header.kind)}")
        return cls(header, entryData)

class Animation:
    def __init__(self, header: Header, entries: List[Entry], size: int):
        self.header = header
        self.entries = entries
        self.size = size

    @classmethod
    def parse(cls, data: BytesIO) -> "Animation":
        header = Header.parse(data)
        entries = []
        while True:
            try:
                entries.append(Entry.parse(data))
            except StopIteration:
                break
        size = struct.unpack("<I", data.read(4))[0]

        return cls(header, entries, size)