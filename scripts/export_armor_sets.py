from dlbin import DlBin, HelldiverCustomizationKit
from strings import Strings, SUPPORTED_LANGUAGE_NAMES, Language

from io import BytesIO
from pathlib import Path
from typing import Dict
import json

from argparse import ArgumentParser

def main():
    parser = ArgumentParser()
    parser.add_argument("-strings-dir", "-s", type=Path, help="Path to directory of strings files to use for mapping armor set names")
    parser.add_argument("-language", "-l", default="en-us", choices=SUPPORTED_LANGUAGE_NAMES)
    parser.add_argument("armor_sets", type=Path, help="Path to decrypted generated_customization_armor_sets.dl_bin")
    args = parser.parse_args()

    armor_sets: Path = args.armor_sets
    strings_dir: Path = args.strings_dir
    language: Language = Language.from_string(args.language)

    with armor_sets.open("rb") as f:
        data = f.read()

    dlbin = DlBin.parse(BytesIO(data))

    json_out = []

    mapping: Dict[int, str] = {}
    for strings_file in strings_dir.iterdir():
        if not strings_file.is_file():
            continue
        if strings_file.suffix not in [".strings", ".main"]:
            continue
        with strings_file.open("rb") as f:
            try:
                strings = Strings.parse(f)
                if strings.language != language:
                    continue
                mapping.update(strings.mapping)
            except:
                continue

    for item in dlbin.items:
        content: HelldiverCustomizationKit = item.content
        json_out.append(content.to_json(mapping))

    print(json.dumps(json_out, indent=4))

if __name__ == "__main__":
    main()