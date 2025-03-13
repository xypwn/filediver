from dlbin import DlBin, HelldiverCustomizationKit
from strings import Strings

from io import BytesIO
from pathlib import Path
import json

from argparse import ArgumentParser

def main():
    parser = ArgumentParser()
    parser.add_argument("-strings", "-s", type=Path, help="Path to .strings file to use for mapping armor set names (0x7c7587b563f10985.strings is en_us)")
    parser.add_argument("armor_sets", type=Path, help="Path to decrypted generated_customization_armor_sets.dl_bin")
    args = parser.parse_args()

    armor_sets: Path = args.armor_sets
    strings_path: Path = args.strings

    with armor_sets.open("rb") as f:
        data = f.read()

    dlbin = DlBin.parse(BytesIO(data))

    json_out = []

    with strings_path.open("rb") as f:
        strings = Strings.parse(f).mapping

    for item in dlbin.items:
        content: HelldiverCustomizationKit = item.content
        json_out.append(content.to_json(strings))

    print(json.dumps(json_out, indent=4))

if __name__ == "__main__":
    main()