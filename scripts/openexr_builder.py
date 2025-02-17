import zlib
import numpy as np
import struct
from typing import List

from openexr.types import OpenEXR

# Very limited EXR writer - supports files of up to 16 scanlines
# which must use RGBA float 32 pixels
#
# Purpose built for converting Helldiver 2 LUTs from DDS R16G16B16A16_FLOAT to EXR RGBA Float32


PIXELTYPE_FLOAT = 2
COMPRESSION_ZIP = 3
LINEORDER_INC_Y = 0

def main():
    from argparse import ArgumentParser
    from pathlib import Path
    from dds_float16 import DDS
    from csv import reader
    parser = ArgumentParser("openexr_builder")
    parser.add_argument("dir", type=Path)
    parser.add_argument("--nsight", action="store_true")
    parser.add_argument("--width", type=int, default=23)
    parser.add_argument("--height", type=int, default=8)
    args = parser.parse_args()

    nsight: bool = args.nsight
    dir: Path = args.dir
    if not nsight:
        for file in dir.iterdir():
            if file.suffix != ".dds":
                continue
            try:
                with file.open("rb") as f:
                    dds = DDS.parse(f)
                exr = OpenEXR.from_pixels(dds.pixels().astype(np.float32)).serialize()
                exr_path = file.with_suffix(".exr")
                with exr_path.open("wb") as f:
                    f.write(exr)
            except (AssertionError, OSError):
                pass
    else:
        assert dir.is_file() and dir.suffix == ".csv"
        with dir.open("r") as f:
            csv = reader(f)
            _ = next(csv)
            conv = lambda x: int(x, base=16)
            data = [list(map(conv, line)) for line in csv]
            exr = decode_nsight_data(data, args.width, args.height)
        exr_path = dir.with_suffix(".exr")
        with exr_path.open("wb") as f:
            f.write(exr)

def decode_nsight_data(data: List[List[int]], width: int, height: int):
    pixels = [struct.unpack("<eeee", struct.pack("<HHHH", *line)) for line in data]
    rows = [pixels[i*width:(i+1)*width] for i in range(height)]
    img_data = np.array(rows, dtype=np.float16)
    return make_exr(img_data.astype(np.float32))
    

if __name__ == "__main__":
    main()