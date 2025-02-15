import struct
from io import BytesIO, SEEK_CUR
from enum import IntEnum

from typing import List, Union, Dict

class Slot(IntEnum):
    NONE = 0
    CAPE = 1
    TORSO = 2
    HIPS = 3
    LEFT_LEG = 4
    RIGHT_LEG = 5
    LEFT_ARM = 6
    RIGHT_ARM = 7
    LEFT_SHOULDER = 8
    RIGHT_SHOULDER = 9

class DLItemType(IntEnum):
    ArmorCustomization = 0xd9a55aa0

class Kit(IntEnum):
    ARMOR = 0
    HELMET = 1
    CAPE = 2

class BodyType(IntEnum):
    SLIM = 0
    STOCKY = 1
    UNKNOWN = 2
    ANY = 3

class PieceType(IntEnum):
    ARMOR = 0
    UNDERGARMENT = 1
    ACCESSORY = 2

class Weight(IntEnum):
    LIGHT = 0
    MEDIUM = 1
    HEAVY = 2

class Passive(IntEnum):
    NONE = 0
    PADDING = 1
    TACTICIAN = 2
    FIRE_SUPPORT = 3
    UNK01 = 4
    EXPERIMENTAL = 5
    COMBAT_ENGINEER = 6
    COMBAT_MEDIC = 7
    BATTLE_HARDENED = 8
    HERO = 9
    FIRE_RESISTANT = 10
    PEAK_PHYSIQUE = 11
    GAS_RESISTANT = 12
    UNFLINCHING = 13
    ACCLIMATED = 14
    SIEGE_READY = 15
    INTEGRATED_EXPLOSIVES = 16

class MurmurHash:
    def __init__(self, value: int):
        self.value = value
    
    def __str__(self):
        return f"0x{self.value:016x}"

class Piece:
    def __init__(self,
                 path: MurmurHash,
                 slot: Slot,
                 pieceType: PieceType,
                 weight: Weight,
                 unk: int, 
                 material_lut: MurmurHash,
                 pattern_lut: MurmurHash,
                 cape_lut: MurmurHash,
                 cape_gradient: MurmurHash,
                 cape_nac: MurmurHash,
                 decal_scalar_fields: MurmurHash,
                 base_data: MurmurHash,
                 decal_sheet: MurmurHash,
                 tone_variations: MurmurHash,
    ):
        self.path = path
        self.slot = slot
        self.pieceType = pieceType
        self.weight = weight
        self.unk = unk
        self.material_lut = material_lut
        self.pattern_lut = pattern_lut
        self.cape_lut = cape_lut
        self.cape_gradient = cape_gradient
        self.cape_nac = cape_nac
        self.decal_scalar_fields = decal_scalar_fields
        self.base_data = base_data
        self.decal_sheet = decal_sheet
        self.tone_variations = tone_variations
    
    def __str__(self):
        return f"path: {self.path}\nslot: {self.slot}\npieceType: {self.pieceType}\nweight: {self.weight}\nunk: {self.unk}\nmaterial_lut: {self.material_lut}\npattern_lut: {self.pattern_lut}\ncape_lut: {self.cape_lut}\ncape_gradient: {self.cape_gradient}\ncape_nac: {self.cape_nac}\ndecal_scalar_fields: {self.decal_scalar_fields}\nbase_data: {self.base_data}\ndecal_sheet: {self.decal_sheet}\ntone_variations: {self.tone_variations}\n"

    @classmethod
    def parse(cls, data: BytesIO) -> 'Piece':
        path, slot, pieceType, weight, unk, material_lut, pattern_lut, cape_lut, cape_gradient, cape_nac, decal_scalar_fields, base_data, decal_sheet, tone_variations = struct.unpack("<QIIIIQQQQQQQQQ", data.read(96))
        return cls(MurmurHash(path), Slot(slot), PieceType(pieceType), Weight(weight), unk, *list(map(MurmurHash, (material_lut, pattern_lut, cape_lut, cape_gradient, cape_nac, decal_scalar_fields, base_data, decal_sheet, tone_variations))))

    def to_json(self) -> dict:
        data = {
            "path": str(self.path),
            "slot": self.slot.name,
            "pieceType": self.pieceType.name,
            "weight": self.weight.name,
            "material_lut": str(self.material_lut),
            "pattern_lut": str(self.pattern_lut),
            "cape_lut": str(self.cape_lut),
            "cape_gradient": str(self.cape_gradient),
            "cape_nac": str(self.cape_nac),
            "decal_scalar_fields": str(self.decal_scalar_fields),
            "base_data": str(self.base_data),
            "decal_sheet": str(self.decal_sheet),
            "tone_variations": str(self.tone_variations),
        }
        return data

class Body:
    def __init__(self, bodyType: BodyType, unk00: int, pieces: List[Piece], unk01: int, count: int, unk02: int):
        self.bodyType = bodyType
        self.unk00 = unk00
        self.pieces = pieces
        self.unk01 = unk01
        self.count = count
        self.unk02 = unk02

    @classmethod
    def parse(cls, data: BytesIO) -> 'Body':
        bodyType, unk00, offset, unk01, count, unk02 = struct.unpack("<IIIIII", data.read(24))
        prev = data.tell()
        data.seek((offset&0xfffff)-0xa0000)
        pieces = [Piece.parse(data) for _ in range(count)]
        data.seek(prev)
        return cls(BodyType(bodyType), unk00, pieces, unk01, count, unk02)
    
    def to_json(self) -> dict:
        data = {
            "bodyType": self.bodyType.name,
            "pieces": [piece.to_json() for piece in self.pieces],
        }
        return data
    
class HelldiverCustomizationKit:
    def __init__(self, _id: int, dlc_id: int, set_id: int, name_upper: int, name_cased: int, description: int, rarity: int, passive: Passive, triad: MurmurHash, kit_type: Kit, unk00: int, bodyTypes: List[Body], unk01: int, count: int, unk02: int):
        self._id = _id
        self.dlc_id = dlc_id
        self.set_id = set_id
        self.name_upper = name_upper
        self.name_cased = name_cased
        self.description = description
        self.rarity = rarity
        self.passive = passive
        self.triad = triad
        self.kit_type = kit_type
        self.unk00 = unk00
        self.body_types = bodyTypes
        self.unk01 = unk01
        self.count = count
        self.unk02 = unk02
    
    @classmethod
    def parse(cls, data: BytesIO) -> 'HelldiverCustomizationKit':
        _id, dlc_id, set_id, name_upper, name_cased, description, rarity, passive = struct.unpack("<IIIIIIII", data.read(32))
        triad, kit_type, unk01, offset, unk02, count, unk03 = struct.unpack("<QIIIIII", data.read(32))
        prev = data.tell()
        data.seek((offset&0xfffff)-0xa0000)
        body_types = [Body.parse(data) for _ in range(count)]
        data.seek(prev)
        return cls(_id, dlc_id, set_id, name_upper, name_cased, description, rarity, Passive(passive), MurmurHash(triad), Kit(kit_type), unk01, body_types, unk02, count, unk03)
    
    def to_json(self, string_mapping: Dict[int, str] = {}) -> dict:
        def named(name_id: int) -> str|int:
            if name_id in string_mapping:
                return string_mapping[name_id]
            return name_id
        
        data = {
            "id": self._id,
            "dlc_id": self.dlc_id,
            "set_id": self.set_id,
            "name_upper": named(self.name_upper),
            "name_cased": named(self.name_cased),
            "description": named(self.description),
            "rarity": self.rarity,
            "passive": self.passive.name,
            "archive": f"{self.triad.value:x}",
            "kit_type": self.kit_type.name,
            "body_types": [body.to_json() for body in self.body_types],
        }
        return data

class DlBinItem:
    def __init__(self, magic: str, unk00: int, kind: DLItemType, size: int, unk02: int, unk03: int, content: Union[bytes, HelldiverCustomizationKit]):
        self.magic = magic
        self.unk00 = unk00
        self.kind = kind
        self.size = size
        self.unk02 = unk02
        self.unk03 = unk03
        self.content = content

    @classmethod
    def parse(cls, data: BytesIO) -> 'DlBinItem':
        magic, unk00, kind, size, unk02, unk03 = struct.unpack("<4sIIIII", data.read(24))
        assert magic == b"LDLD"
        contentStart = data.tell()
        content = data.read(size)
        contentEnd = data.tell()
        if DLItemType(kind) == DLItemType.ArmorCustomization:
            data.seek(contentStart)
            content = HelldiverCustomizationKit.parse(data)
            data.seek(contentEnd)
        return cls(magic, unk00, DLItemType(kind), size, unk02, unk03, content)
    
    def serialize(self) -> bytes:
        return struct.pack("<4sIIIII", self.magic, self.unk00, self.unk01, self.size, self.unk02, self.unk03) + self.content


class DlBin:
    def __init__(self, count: int, items: List[DlBinItem]):
        self.count = count
        self.items = items

    @classmethod
    def parse(cls, data: BytesIO) -> 'DlBin':
        count = struct.unpack("<I", data.read(4))[0]
        items = [DlBinItem.parse(data) for _ in range(count)]
        while data.read(4) == b"LDLD":
            data.seek(-4, SEEK_CUR)
            items.append(DlBinItem.parse(data))
        else:
            data.seek(-4, SEEK_CUR)
        return cls(count, items)
    
    def serialize(self) -> bytes:
        return struct.pack("<I", self.count) + b''.join([item.serialize() for item in self.items])