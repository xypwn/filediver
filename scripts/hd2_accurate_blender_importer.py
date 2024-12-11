import bpy
import json
import os
import struct
import numpy as np
from argparse import ArgumentParser
from bpy.types import BlendData, Image, Object, ShaderNodeGroup, ShaderNodeTexImage
from pathlib import Path
from io import BytesIO
from typing import Optional, Dict, List, Tuple

from dds_float16 import DDS
from openexr_builder import make_exr

class GLTFChunk:
    def __init__(self, length: int, type: str, data: bytes) -> None:
        self.length = length
        self.type = type
        self.data = data
    
    @classmethod
    def parse(cls, data: BytesIO) -> 'GLTFChunk':
        type: bytes
        length, type = struct.unpack("<I4s", data.read(8))
        return cls(length, type.decode().strip("\0"), data.read(length))

def load_glb(path: Path, debug: bool) -> dict:
    chunks: List[GLTFChunk] = []
    with path.open("rb") as f:
        magic = f.read(4).decode()
        assert magic == "glTF", "Invalid glb file!"
        version, length = struct.unpack("<II", f.read(8))
        assert version == 2
        while f.tell() != length:
            chunks.append(GLTFChunk.parse(f))
    assert chunks[0].type == "JSON"
    gltf = json.loads(chunks[0].data.decode())
    if debug:
        with path.with_suffix(".debug.gltf").open("w") as f:
            json.dump(gltf, f, indent=4)
    gltf["chunks"] = chunks[1:]
    return gltf

def get_data(gltf: dict, bufferViewIdx: int) -> bytes:
    bufferView: Dict[str, int] = gltf["bufferViews"][bufferViewIdx]
    bufferIdx = bufferView["buffer"]
    startOffset = bufferView.get("byteOffset", 0)
    endOffset = startOffset + bufferView["byteLength"]
    return gltf["chunks"][bufferIdx].data[startOffset:endOffset]

def get_texture_data(gltf: dict, textureIdx: int) -> bytes:
    texture = gltf["textures"][textureIdx]
    if "extensions" in texture:
        sourceIdx = texture["extensions"]["MSFT_texture_dds"]["source"]
    else:
        sourceIdx = texture["source"]
    image = gltf["images"][sourceIdx]
    return get_data(gltf, image["bufferView"])

def get_texture_image(gltf: dict, textureIdx: int) -> dict:
    texture = gltf["textures"][textureIdx]
    if "extensions" in texture:
        sourceIdx = texture["extensions"]["MSFT_texture_dds"]["source"]
    else:
        sourceIdx = texture["source"]
    image = gltf["images"][sourceIdx]
    return image

# Returns the PNGs dimensions as a height, width tuple
def get_png_dimensions(png: bytes) -> Tuple[int, int]:
    data = BytesIO(png)
    magic0, magic1, magic2 = struct.unpack("<B3sI", data.read(8))
    assert magic0 == 137 and magic1.decode() == "PNG" and magic2 == 0x0A1A0A0D, "Invalid PNG file"
    length, chunkType = struct.unpack(">I4s", data.read(8))
    assert chunkType.decode() == "IHDR" and length == 13
    width, height = struct.unpack(">II", data.read(8))
    data.close()
    return height, width

def add_texture(gltf, textureIdx, usage: Optional[str] = None) -> Image:
    image = get_texture_image(gltf, textureIdx)
    data = get_texture_data(gltf, textureIdx)
    name = image['name']
    if usage is not None:
        name = f"{usage} {image['name']}"
    name = str(Path(name).with_suffix(".png"))
    if image["mimeType"] == "image/vnd-ms.dds":
        dds = DDS.parse(BytesIO(data))
        data = make_exr(dds.pixels().astype(np.float32))
        name = str(Path(name).with_suffix(".exr"))
        height, width = dds.header.height, dds.header.width
        fmt = "OPEN_EXR"
    else:
        height, width = get_png_dimensions(data)
        fmt = "PNG"
    if name in bpy.data.images:
        return bpy.data.images[name]
    blImage = bpy.data.images.new(name, width, height, alpha=True)
    blImage.pack(data=data, data_len=len(data))
    blImage.file_format = fmt
    blImage.source = "FILE"
    blImage.alpha_mode = "CHANNEL_PACKED"
    blImage.use_fake_user = True
    return blImage

def main():
    parser = ArgumentParser("hd2_accurate_blender_importer")
    parser.add_argument("input_model", type=Path, help="Path to filediver-exported .glb to import into a .blend file")
    parser.add_argument("output", type=Path, help="Location to save .blend file")
    parser.add_argument("--debug", action="store_true", help="Export debug data")

    script_path = Path(os.path.realpath(__file__)).parent
    resource_path = script_path / "resources"

    args = parser.parse_args()
    input_model: Path = args.input_model
    output: Path = args.output

    gltf = load_glb(input_model, args.debug)
    assert gltf["asset"]["generator"] == "https://github.com/xypwn/filediver", f"GLB file was not created by Filediver! (Generator: {gltf['asset']['generator']})"

    for i in range(len(gltf["textures"])):
        add_texture(gltf, i)

    bpy.ops.wm.save_mainfile(filepath=str(output))



if __name__ == "__main__":
    main()