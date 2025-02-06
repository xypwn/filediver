# Extracts dl_bin files from a helldivers memory dump
# Obtaining a memory dump is an exercise left to the reader

import struct
from typing import List, Tuple
from argparse import ArgumentParser
from pathlib import Path
from pprint import pprint
import os

from dlbin import DlBin

search_bytes = b'\x4c\x44\x4c\x44\x01\x00\x00\x00'

def main():
    parser = ArgumentParser()
    parser.add_argument("core_file", type=Path)
    parser.add_argument("--game-dir", type=Path, default=Path("C:/Program Files (x86)/Steam/steamapps/common/Helldivers 2/data/game"))
    parser.add_argument("--offset", type=int, default=0x13d200000)
    args = parser.parse_args()

    offset: int = args.offset
    core_file: Path = args.core_file
    game_dir: Path = args.game_dir

    dlbin_sizes: List[Tuple[str, int]] = []
    assert game_dir.is_dir()
    for file in game_dir.iterdir():
        if not file.is_file():
            continue
        if not file.suffix == ".dl_bin":
            continue
        dlbin_sizes.append((file.name, file.stat().st_size))

    dlbin_sizes = sorted(dlbin_sizes, key=lambda x: x[1])
    
    pprint(dlbin_sizes)

    with core_file.open("rb") as f:
        f.seek(0, os.SEEK_END)
        end = f.tell()
        if end <= offset:
            print(f"Offset too large! (0x{offset:x} >= 0x{end:x})")
            return
        while offset < end:
            f.seek(offset, os.SEEK_SET)
            print(f"{offset:x}/{end:x}", end="\r", flush=True)
            chunk = f.read(0x1000)
            try:
                if search_bytes in chunk and (idx := chunk.index(search_bytes)) >= 4:
                    f.seek(offset + idx - 4)
                    start = f.tell()
                    count = struct.unpack("<I", f.peek(4)[:4])[0]
                    assert count != 0
                    if count > 1024:
                        count = 1
                    dlbin = DlBin.parse(f)
                    size = f.tell() - start
                    for name, encsize in dlbin_sizes:
                        if (encsize - 48) >= size:
                            decrypted_path = Path(name).with_suffix(".dec.dl_bin")
                            if decrypted_path.exists():
                                exist_diff = (encsize - 48) - decrypted_path.stat().st_size
                                new_diff = (encsize - 48) - size
                                if not (new_diff < exist_diff and new_diff > 0):
                                    #print(f"{decrypted_path} already exists, not overwriting")
                                    break
                            #print(f"Writing file as {decrypted_path}")
                            print(f"Found a file of size {size} at {start:x}!")
                            with decrypted_path.open("wb") as o:
                                o.write(dlbin.serialize())
                            break
            except AssertionError:
                f.seek(offset + 0x1000, os.SEEK_SET)

            offset = f.tell()
            if offset % 0x1000 != 0:
                offset += 0x1000 - (offset % 0x1000)

if __name__ == "__main__":
    main()