from bpy.types import BlendData, Image
from typing import Dict, Tuple, Optional
import numpy as np
from io import BytesIO
from pathlib import Path
import struct

from openexr.types import OpenEXR
from dds_float16 import DDS

def get_data(gltf: dict, bufferViewIdx: int) -> bytes:
    bufferView: Dict[str, int] = gltf["bufferViews"][bufferViewIdx]
    bufferIdx = bufferView["buffer"]
    startOffset = bufferView.get("byteOffset", 0)
    endOffset = startOffset + bufferView["byteLength"]
    return gltf["chunks"][bufferIdx].data[startOffset:endOffset]

def get_texture_data(gltf: dict, textureIdx: int, dds_if_avail: bool = True) -> bytes:
    texture = gltf["textures"][textureIdx]
    if "extensions" in texture and dds_if_avail:
        sourceIdx = texture["extensions"]["MSFT_texture_dds"]["source"]
    else:
        sourceIdx = texture["source"]
    image = gltf["images"][sourceIdx]
    return get_data(gltf, image["bufferView"])

def get_texture_image(gltf: dict, textureIdx: int, dds_if_avail: bool = True) -> dict:
    texture = gltf["textures"][textureIdx]
    if "extensions" in texture and dds_if_avail:
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

def add_texture(bldata: BlendData, gltf, textureIdx, usage: Optional[str] = None) -> Image:
    image = get_texture_image(gltf, textureIdx)
    data = get_texture_data(gltf, textureIdx)
    name = image['name']
    if usage is not None:
        name = f"{usage} {image['name']}"
    name = str(Path(name).with_suffix(".png"))
    if image["mimeType"] == "image/vnd-ms.dds":
        try:
            dds = DDS.parse(BytesIO(data))
            data = OpenEXR.from_pixels(dds.pixels().astype(np.float32)).serialize()
            name = str(Path(name).with_suffix(".exr"))
            height, width = dds.header.height, dds.header.width
            fmt = "OPEN_EXR"
        except AssertionError:
            image = get_texture_image(gltf, textureIdx, False)
            data = get_texture_data(gltf, textureIdx, False)
            height, width = get_png_dimensions(data)
            fmt = "PNG"
    else:
        height, width = get_png_dimensions(data)
        fmt = "PNG"
    if name in bldata.images:
        return bldata.images[name]
    blImage = bldata.images.new(name, width, height, alpha=True)
    blImage.pack(data=data, data_len=len(data))
    blImage.file_format = fmt
    blImage.source = "FILE"
    blImage.alpha_mode = "CHANNEL_PACKED"
    blImage.use_fake_user = True
    return blImage

def create_empty_texture(data: BlendData, name: str, size: Tuple[int, int], fmt: str = 'PNG', colorspace: str = 'sRGB', alpha_mode: str = 'CHANNEL_PACKED') -> Image:
    unused_texture = data.images.new(name, size[0], size[1], alpha=True, float_buffer=True)
    #unused_texture.pixels[3] = 0.0
    exr = OpenEXR.from_pixels(np.zeros((size[1], size[0], 4), dtype=np.float32)).serialize()
    unused_texture.use_fake_user = True
    unused_texture.pack(data=exr, data_len=len(exr))
    unused_texture.source = "FILE"
    unused_texture.file_format = fmt
    unused_texture.colorspace_settings.name = colorspace
    unused_texture.alpha_mode = alpha_mode
    return unused_texture

def get_textures(data: BlendData, gltf: dict, materialTextures: Dict[int, Dict[str, Image]], materialIndex: int, material: dict, needs_further_processing: bool):
    textures = {}
    optional_usages = ["decal_sheet", "pattern_masks_array"]
    if materialIndex in materialTextures:
        textures = materialTextures[materialIndex]
    else:
        if len(material["extras"]) == 0 or not needs_further_processing:
            return None
        print(f"    Packing textures for material {material['name']}")
        try:
            if "unused" not in data.images:
                unused_texture = create_empty_texture(data, "unused", (1, 1))
            else:
                unused_texture = data.images["unused"]
            textures["filediver_unused"] = unused_texture
            for usage, texIdx in material["extras"].items():
                if type(texIdx) != int:
                    continue
                textures[usage] = add_texture(data, gltf, texIdx, usage)
            for usage in optional_usages:
                if usage not in textures:
                    textures[usage] = unused_texture
            materialTextures[materialIndex] = textures
        except AssertionError as e:
            print(f"Error: {e}")
            return None
    return textures