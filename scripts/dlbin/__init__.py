import struct
from io import BytesIO, SEEK_CUR
from enum import IntEnum

from typing import List, Union, Dict, Tuple
from strings import read_cstr

class Slot(IntEnum):
    HELMET = 0
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
    UnitCustomization = 0xa2ba274a
    WeaponCustomization = 0x1e604234

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
    REINFORCED_EPAULETTES = 10
    FIRE_RESISTANT = 11
    PEAK_PHYSIQUE = 12
    GAS_RESISTANT = 13
    UNFLINCHING = 14
    ACCLIMATED = 15
    SIEGE_READY = 16
    INTEGRATED_EXPLOSIVES = 17
    GUNSLINGER = 18

class MurmurHash:
    def __init__(self, value: int):
        self.value = value

    def __str__(self):
        return f"0x{self.value:016x}"

ARMOR_SET_OFFSET = 0x3f0000

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
        data.seek((offset&0xffffff)-ARMOR_SET_OFFSET)
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
        data.seek((offset&0xffffff)-ARMOR_SET_OFFSET)
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

class UIColors:
    def __init__(self,
            first: Tuple[float, float, float],
            second: Tuple[float, float, float],
            third: Tuple[float, float, float],
            fourth: Tuple[float, float, float]
        ):
        self.first = first
        self.second = second
        self.third = third
        self.fourth = fourth

    def to_json(self):
        return {
            "first": self.first,
            "second": self.second,
            "third": self.third,
            "fourth": self.fourth,
        }

    @classmethod
    def parse(cls, data: BytesIO) -> 'UIColors':
        first = struct.unpack("<fff", data.read(12))
        second = struct.unpack("<fff", data.read(12))
        third = struct.unpack("<fff", data.read(12))
        fourth = struct.unpack("<fff", data.read(12))
        return cls(first, second, third, fourth)

UNIT_CUSTOMIZATION_OFFSET = 0x70000

class UnitCustomizationSkin:
    def __init__(self, 
            debug_name: str,
            unk00: int,
            id: int,
            unk01: int,
            add_path: MurmurHash,
            name_id: int,
            unk02: int,
            thumbnail: MurmurHash,
            colors: UIColors,
            unk03: int,
            unk04: int,
            unk05: int
        ):
        self.debug_name = debug_name
        self.unk00 = unk00
        self.id = id
        self.unk01 = unk01
        self.add_path = add_path
        self.name_id = name_id
        self.unk02 = unk02
        self.thumbnail = thumbnail
        self.colors = colors
        self.unk03 = unk03
        self.unk04 = unk04
        self.unk05 = unk05

    def to_json(self) -> dict:
        return {
            "debug_name": self.debug_name,
            "unk00": self.unk00,
            "id": self.id,
            "unk01": self.unk01,
            "add_path": str(self.add_path),
            "name_id": self.name_id,
            "unk02": self.unk02,
            "thumbnail": str(self.thumbnail),
            "colors": self.colors.to_json(),
            "unk03": self.unk03,
            "unk04": self.unk04,
            "unk05": self.unk05,
        }

    @classmethod
    def parse(cls, data: BytesIO) -> 'UnitCustomizationSkin':
        name_offset, unk00, id, unk01, add_path, name, unk02, thumbnail, colors_offset, unk03, unk04, unk05 = struct.unpack("<IIIIQIIQIIII", data.read(56))
        curr_offset = data.tell()
        name_offset &= 0xFFFFF - UNIT_CUSTOMIZATION_OFFSET
        colors_offset &= 0xFFFFF - UNIT_CUSTOMIZATION_OFFSET
        if name_offset != 0:
            data.seek(name_offset)
            debug_name = read_cstr(data)
        else:
            debug_name = ""
        if colors_offset != 0:
            data.seek(colors_offset)
            colors = UIColors.parse(data)
        else:
            colors = UIColors((0.0, 0.0, 0.0), (0.0, 0.0, 0.0), (0.0, 0.0, 0.0), (0.0, 0.0, 0.0))
        data.seek(curr_offset)
        return cls(debug_name, unk00, id, unk01, MurmurHash(add_path), name, unk02, MurmurHash(thumbnail), colors, unk03, unk04, unk05)

class UnitCustomizationSetting:
    def __init__(self, 
                 parent_type: int,
                 typ: int,
                 object_name_id: int,
                 skin_name_id: int,
                 category_type: int,
                 unk00: int,
                 unk01: int,
                 unk02: int,
                 showroom_offset: List[float],
                 showroom_rotation: List[float],
                 skins: List[UnitCustomizationSkin]
        ):
        self.parent_type = parent_type
        self.type = typ
        self.object_name_id = object_name_id
        self.skin_name_id = skin_name_id
        self.category_type = category_type
        self.unk00 = unk00
        self.unk01 = unk01
        self.unk02 = unk02
        self.showroom_offset = showroom_offset
        self.showroom_rotation = showroom_rotation
        self.skins = skins

    def to_json(self) -> dict:
        return {
            "parent_type": self.parent_type,
            "type": self.type,
            "object_name_id": self.object_name_id,
            "skin_name_id": self.skin_name_id,
            "category_type": self.category_type,
            "unk00": self.unk00,
            "unk01": self.unk01,
            "unk02": self.unk02,
            "showroom_offset": self.showroom_offset,
            "showroom_rotation": self.showroom_rotation,
            "skins": [skin.to_json() for skin in self.skins],
        }

    @classmethod
    def parse(cls, data: BytesIO) -> 'UnitCustomizationSetting':
        parent_type, typ, object_name, skin_name, category_type, unk00, skins_offset, unk01, count, unk02 = struct.unpack("<IIIIIIIIII", data.read(40))
        showroom_offset = struct.unpack("<fff", data.read(12))
        showroom_rotation = struct.unpack("<fff", data.read(12))
        skins_offset &= 0xFFFFF - UNIT_CUSTOMIZATION_OFFSET
        data.seek(skins_offset)
        skins = [UnitCustomizationSkin.parse(data) for _ in range(count)]
        return cls(parent_type, typ, object_name, skin_name, category_type, unk00, unk01, unk02, showroom_offset, showroom_rotation, skins)


WEAPON_CUSTOMIZATION_OFFSET = 0x70000

class WeaponCustomizationSlot(IntEnum):
    NONE = 0
    UNDERBARREL = 1
    OPTICS = 2
    PAINTSCHEME = 3
    MUZZLE = 4
    MAGAZINE = 5
    AMMOTYPE = 6
    AMMOTYPE_ALT = 7
    INTERNALS = 8
    TRIGGER = 9

class WeaponCustomizationSetting:
    def __init__(self, debug_name: str, id: int, name_upper: int, name_cased: int, description: int, fluff: int, add_path: MurmurHash, slots: List[WeaponCustomizationSlot], unk00, unk01, unk02, unk03):
        self.debug_name = debug_name
        self.id = id
        self.name_upper = name_upper
        self.name_cased = name_cased
        self.description = description
        self.fluff = fluff
        self.add_path = add_path
        self.slots = slots
        self.unk00 = unk00
        self.unk01 = unk01
        self.unk02 = unk02
        self.unk03 = unk03

    def to_json(self, string_mapping: Dict[int, str] = {}) -> dict:
        def named(name_id: int) -> str|int:
            if name_id in string_mapping:
                return string_mapping[name_id]
            return name_id
        
        data = {
            "id": self.id,
            "name_upper": named(self.name_upper),
            "name_cased": named(self.name_cased),
            "debug_name": self.debug_name,
            "description": named(self.description),
            "fluff": named(self.fluff),
            "add_path": str(self.add_path),
            "slots": [slot.name for slot in self.slots]
        }
        return data

    @classmethod
    def parse(cls, data: BytesIO) -> 'WeaponCustomizationSetting':
        nameOffset, unk00, id, name_upper, name_cased, description, fluff, add_path = struct.unpack("<IIIIIIQQ", data.read(40))
        unk01, slotOffset, unk02, count, unk03 = struct.unpack("<QIIII", data.read(24))
        prev = data.tell()
        data.seek((nameOffset & 0xFFFFF) - WEAPON_CUSTOMIZATION_OFFSET)
        debug_name = read_cstr(data)
        data.seek((slotOffset & 0xFFFFF) - WEAPON_CUSTOMIZATION_OFFSET)
        slots = [WeaponCustomizationSlot(val[0]) for val in struct.iter_unpack("<I", data.read(count * 4))]
        data.seek(prev)
        return cls(debug_name, id, name_upper, name_cased, description, fluff, MurmurHash(add_path), slots, unk00, unk01, unk02, unk03)

class DlBinItem:
    def __init__(self, magic: str, unk00: int, kind: DLItemType, size: int, unk02: int, unk03: int, content: Union[bytes, HelldiverCustomizationKit, UnitCustomizationSetting]):
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
        elif DLItemType(kind) == DLItemType.UnitCustomization:
            data.seek(contentStart)
            content = UnitCustomizationSetting.parse(data)
            data.seek(contentEnd)
        elif DLItemType(kind) == DLItemType.WeaponCustomization:
            data.seek(contentStart)
            offset, unk00, count, unk01 = struct.unpack("<IIII", data.read(16))
            data.seek((offset & 0xFFFFF) - WEAPON_CUSTOMIZATION_OFFSET)
            content = [WeaponCustomizationSetting.parse(data) for _ in range(count)]
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