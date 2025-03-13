import struct
from io import BytesIO
from typing import List, Dict

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

class Strings:
    def __init__(self, magic: int, version: int, hash_maybe: int, string_ids: List[int], strings: List[str]):
        self.magic = magic
        self.version = version
        self.hash_maybe = hash_maybe
        self.string_ids = string_ids
        self.strings = strings
        self.mapping: Dict[int, str] = {key: val for key, val in zip(string_ids, strings)}

    @classmethod
    def parse(cls, data: BytesIO):
        magic, version, count = struct.unpack("<III", data.read(12))
        if count == 0:
            return cls(magic, version, 0, [], [])
        unk = struct.unpack("<I", data.read(4))[0]
        string_ids = [val[0] for val in struct.iter_unpack("<I", data.read(4 * count))]
        string_offsets = [val[0] for val in struct.iter_unpack("<I", data.read(4 * count))]
        strings = []
        for offset in string_offsets:
            data.seek(offset)
            strings.append(read_cstr(data))
        return cls(magic, version, unk, string_ids, strings)

    def __getitem__(self, key: int):
        return self.mapping[key]
    
    def __contains__(self, key: int):
        return key in self.mapping