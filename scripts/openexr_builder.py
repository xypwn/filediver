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
    parser.add_argument("--type", choices=["float", "half"], default="half")
    parser.add_argument("--width", type=int, default=23)
    parser.add_argument("--height", type=int, default=8)
    args = parser.parse_args()

    nsight: bool = args.nsight
    fmt = f"<{'ffff' if (args.type == 'float') else 'eeee'}"
    dir: Path = args.dir
    for file in dir.iterdir():
        if file.suffix == ".dds":
            print("dds")
            try:
                with file.open("rb") as f:
                    dds = DDS.parse(f)
                pixels = dds.pixels().astype(np.float32)
            except (AssertionError, OSError) as e:
                print(f"error: {e}")
                pass
        elif file.suffix == ".csv" and nsight:
            print("csv")
            with file.open("r") as f:
                csv = reader(f)
                _ = next(csv)
                conv = lambda x: int(x, base=16)
                data = [list(map(conv, line)) for line in csv]
                pixels = decode_nsight_data(data, args.width, args.height)
        elif file.suffix == ".bin":
            print("bin")
            with file.open("rb") as f:
                data = f.read()
                line = [list(struct.iter_unpack(fmt, data))]
                pixels = np.array(line, dtype=np.float32)
                print(pixels.shape)
        else:
            continue
        if len(pixels.shape) == 4:
            for i in range(pixels.shape[0]):
                exr = OpenEXR.from_pixels(pixels[i]).serialize()
                exr_path = file.with_suffix(f".{i:03d}.exr")
                with exr_path.open("wb") as f:
                    f.write(exr)
        else:
            exr = OpenEXR.from_pixels(pixels).serialize()
            exr_path = file.with_suffix(".exr")
            with exr_path.open("wb") as f:
                f.write(exr)

def decode_nsight_data(data: List[List[int]], width: int, height: int):
    pixels = [struct.unpack("<eeee", struct.pack("<HHHH", *line)) for line in data]
    rows = [pixels[i*width:(i+1)*width] for i in range(height)]
    img_data = np.array(rows, dtype=np.float16)
    return img_data.astype(np.float32)
    

if __name__ == "__main__":
    main()