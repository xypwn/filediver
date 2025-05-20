from argparse import ArgumentParser
from pathlib import Path
from io import BytesIO
import struct
import sys
import os

from typing import Tuple, List, Union

def read_cstr(data: BytesIO) -> str:
    to_return = ""
    char = data.read(1)
    while char != b"\0" and char != b"":
        try:
            to_return += char.decode()
        except:
            break
        char = data.read(1)
    return to_return

def print_cstrings(data: BytesIO):
    string = read_cstr(data)
    while len(string) != 0:
        print(string)
        string = read_cstr(data)

class Header:
    def __init__(self, magic: str, digest: bytes, version: Tuple[int, int], size: int, parts: int):
        self.magic = magic
        self.digest = digest
        self.version = version
        self.size = size
        self.parts = parts

    @classmethod
    def parse(cls, data: BytesIO) -> 'Header':
        magic = data.read(4).decode()
        digest = data.read(16)
        version = struct.unpack("<HH", data.read(4))
        assert version == (1, 0), f"version was {version}"
        size, parts = struct.unpack("<II", data.read(8))
        return cls(magic, digest, version, size, parts)

class Variable:
    def __init__(self, name: str, cb_offset: int, size: int, flags: int, typename: str, def_val_offset: int, unk: Tuple[int, int, int, int]):
        self.name = name
        self.cb_offset = cb_offset
        self.size = size
        self.flags = flags
        self.type = typename
        self.def_val_offset = def_val_offset
        self.unk = unk

    @classmethod
    def parse(cls, data: BytesIO, chunk_start: int) -> 'Variable':
        name_offset, cb_offset, size, flags, type_offset, def_val_offset = struct.unpack("<IIIIII", data.read(24))
        unkInts = struct.unpack("<IIII", data.read(16))
        pos = data.tell()
        data.seek(chunk_start + 8 + name_offset)
        name = read_cstr(data)
        data.seek(chunk_start + type_offset)
        typename = read_cstr(data)
        data.seek(pos)
        return cls(name, cb_offset, size, flags, typename, def_val_offset, unkInts)

class ConstantBuffer:
    def __init__(self, name: str, variables: List[Variable], size: int, flags: int, typeval: int):
        self.name = name
        self.variables = variables
        self.size = size
        self.flags = flags
        self.type = typeval
    
    @classmethod
    def parse(cls, data: BytesIO, chunk_start: int) -> 'ConstantBuffer':
        name_offset, var_count, var_offset, size, flags, typeval = struct.unpack("<IIIIII", data.read(24))
        pos = data.tell()
        data.seek(chunk_start + 8 + name_offset)
        name = read_cstr(data)
        data.seek(chunk_start + 8 + var_offset)
        variables = [Variable.parse(data, chunk_start) for _ in range(var_count)]
        data.seek(pos)
        return cls(name, variables, size, flags, typeval)

class ResourceBinding:
    def __init__(self, name: str, input_type: int, return_type: int, view_dimension: int, sample_count: int, bind_point: int, bind_count: int, flags: int):
        self.name = name
        self.input_type = input_type
        self.return_type = return_type
        self.view_dimension = view_dimension
        self.sample_count = sample_count
        self.bind_point = bind_point
        self.bind_count = bind_count
        self.flags = flags
    
    @classmethod
    def parse(cls, data: BytesIO, chunk_start: int) -> 'ResourceBinding':
        name_offset, input_type, return_type, view_dimension, sample_count, bind_point, bind_count, flags = struct.unpack("<IIIIIIII", data.read(32))
        pos = data.tell()
        data.seek(chunk_start + 8 + name_offset)
        name = read_cstr(data)
        data.seek(pos)
        return cls(name, input_type, return_type, view_dimension, sample_count, bind_point, bind_count, flags)

class RDEF:
    def __init__(self, cbufs: List[ConstantBuffer], rbinds: List[ResourceBinding], version: Tuple[int, int], program_type: int, flags: int, creator_offset: int):
        self.cbufs = cbufs
        self.rbinds = rbinds
        self.version = version
        self.program_type = program_type
        self.flags = flags
        self.creator_offset = creator_offset

    @classmethod
    def parse(cls, data: BytesIO, chunk_start: int) -> 'RDEF':
        cbuf_count, cbuf_offset, rbind_count, rbind_offset, minor, major, program_type, flags, creator_offset = struct.unpack("<IIIIBBHII", data.read(28))
        pos = data.tell()
        data.seek(chunk_start + 8 + cbuf_offset)
        cbufs = [ConstantBuffer.parse(data, chunk_start) for _ in range(cbuf_count)]
        data.seek(chunk_start + 8 + rbind_offset)
        rbinds = [ResourceBinding.parse(data, chunk_start) for _ in range(rbind_count)]
        data.seek(pos)
        return cls(cbufs, rbinds, (major, minor), program_type, flags, creator_offset)

class PartHeader:
    def __init__(self, name: str, size: int):
        self.name = name
        self.size = size

    @classmethod
    def parse(cls, data: BytesIO) -> 'PartHeader':
        name, size = struct.unpack("<4sI", data.read(8))
        return cls(name.decode(), size)

class Part:
    def __init__(self, header: PartHeader, content: Union[bytes, RDEF]):
        self.header = header
        self.content = content
    
    @classmethod
    def parse(cls, data: BytesIO) -> 'Part':
        chunk_start = data.tell()
        header = PartHeader.parse(data)
        if header.name == "RDEF":
            content = RDEF.parse(data, chunk_start)
            data.seek(chunk_start+8+header.size)
        else:
            content = data.read(header.size)
        return cls(header, content)

class DXBC:
    def __init__(self, header: Header, parts: List[Part]):
        self.header = header
        self.parts = parts

    @classmethod
    def parse(cls, data: BytesIO) -> 'DXBC':
        base = data.tell()
        header = Header.parse(data)
        offsets = struct.unpack(f"<{header.parts}I", data.read(header.parts * 4))
        parts: List[Part] = []
        for offset in offsets:
            data.seek(base + offset)
            parts.append(Part.parse(data))
        return cls(header, parts)

def dump_params(f: BytesIO):
    offset = 0
    prev_offset = offset
    f.seek(0, os.SEEK_END)
    end = f.tell()
    f.seek(0, os.SEEK_SET)
    data = b""
    while offset < end:
        data = f.read(min(end - offset, 2**16))
        if b"DXBC" in data:
            index = data.index(b"DXBC")
            f.seek(offset+index, os.SEEK_SET)
            try:
                dxbc = DXBC.parse(f)
                for part in dxbc.parts:
                    if part.header.name != "RDEF":
                        continue
                    rdef: RDEF = part.content
                    for cb in rdef.cbufs:
                        #if cb.name == "c_per_object" or cb.name == "c_per_instance":
                        for var in cb.variables:
                            print(var.name)
                    break
            except AssertionError as e:
                print(e.args, file=sys.stderr)
                f.seek(offset + index + 4)
        offset = f.tell()
        if offset == prev_offset:
            f.seek(offset + 4)
            offset = f.tell()
        prev_offset = offset
        print(f"{offset:x}/{end:x}", end="\r", flush=True, file=sys.stderr)
    print("", flush=True, file=sys.stderr)

def main():
    parser = ArgumentParser()
    parser.add_argument("core_file", type=Path)
    args = parser.parse_args()
    core_file: Path = args.core_file
    print(core_file.name, flush=True, file=sys.stderr)

    if core_file.is_file():
        with core_file.open("rb") as f:
            dump_params(f)
    else:
        for filename in core_file.iterdir():
            if filename.is_file():
                with filename.open("rb") as f:
                    dump_params(f)

if __name__ == "__main__":
    main()