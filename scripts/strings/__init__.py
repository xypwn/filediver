import struct
from io import BytesIO
from typing import List, Dict
from enum import IntEnum

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

class Language(IntEnum):
    DE      = 0xba39c3ec
    ES      = 0x31806842
    ES_MX   = 0xe5c65a36
    EN_GB   = 0x6f4515cb
    EN_US   = 0x03f97b57
    FR      = 0xfea0f61f
    IT      = 0xe2fb1acd
    JA      = 0x90b6af29
    KO      = 0xbbd7b5d1
    NL      = 0x11592f05
    PL      = 0x0f8857aa
    PT      = 0x4a2ca9c9
    PT_BR   = 0x6ef58def
    RU      = 0xc5bb18ed
    ZH_HANS = 0x82874cc2
    ZH_HANT = 0x9eba952a

    @classmethod
    def from_string(cls, name: str):
        if name.lower() == "de" or name.lower() == "german":
            return cls.DE
        if name.lower() == "es" or name.lower() == "spanish":
            return cls.ES
        if name.lower() == "es-mx" or name.lower() == "spanish (mexico)":
            return cls.ES_MX
        if name.lower() == "en-gb" or name.lower() == "english (uk)":
            return cls.EN_GB
        if name.lower() == "en-us" or name.lower() == "english (us)":
            return cls.EN_US
        if name.lower() == "fr" or name.lower() == "french":
            return cls.FR
        if name.lower() == "it" or name.lower() == "italian":
            return cls.IT
        if name.lower() == "ja" or name.lower() == "japanese":
            return cls.JA
        if name.lower() == "ko" or name.lower() == "korean":
            return cls.KO
        if name.lower() == "nl" or name.lower() == "dutch":
            return cls.NL
        if name.lower() == "pl" or name.lower() == "polish":
            return cls.PL
        if name.lower() == "pt" or name.lower() == "portuguese":
            return cls.PT
        if name.lower() == "pt-br" or name.lower() == "portuguese (brazil)":
            return cls.PT_BR
        if name.lower() == "ru" or name.lower() == "russian":
            return cls.RU
        if name.lower() == "zh-hans" or name.lower() == "simplified chinese":
            return cls.ZH_HANS
        if name.lower() == "zh-hant" or name.lower() == "traditional chinese":
            return cls.ZH_HANT
        raise ValueError(f"{name} is not a supported language")

SUPPORTED_LANGUAGE_NAMES = [
    "de",
    "es",
    "es-mx",
    "en-gb",
    "en-us",
    "fr",
    "it",
    "ja",
    "ko",
    "nl",
    "pl",
    "pt",
    "pt-br",
    "ru",
    "zh-hans",
    "zh-hant",
    "German",
    "Spanish",
    "Spanish (Mexico)",
    "English (UK)",
    "English (US)",
    "French",
    "Italian",
    "Japanese",
    "Korean",
    "Dutch",
    "Polish",
    "Portuguese",
    "Portuguese (Brazil)",
    "Russian",
    "Simplified Chinese",
    "Traditional Chinese"
]

class Strings:
    def __init__(self, magic: int, version: int, language: int, string_ids: List[int], strings: List[str]):
        self.magic = magic
        self.version = version
        self.language = language
        self.string_ids = string_ids
        self.strings = strings
        self.mapping: Dict[int, str] = {key: val for key, val in zip(string_ids, strings)}

    @classmethod
    def parse(cls, data: BytesIO):
        magic, version, count = struct.unpack("<III", data.read(12))
        if count == 0:
            return cls(magic, version, 0, [], [])
        language = struct.unpack("<I", data.read(4))[0]
        string_ids = [val[0] for val in struct.iter_unpack("<I", data.read(4 * count))]
        string_offsets = [val[0] for val in struct.iter_unpack("<I", data.read(4 * count))]
        strings = []
        for offset in string_offsets:
            data.seek(offset)
            strings.append(read_cstr(data))
        return cls(magic, version, language, string_ids, strings)

    def __getitem__(self, key: int):
        return self.mapping[key]
    
    def __contains__(self, key: int):
        return key in self.mapping