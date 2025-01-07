import bpy
import json
import os
import sys
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

def decode_glb(f: BytesIO) -> List[GLTFChunk]:
    chunks: List[GLTFChunk] = []
    magic = f.read(4).decode()
    assert magic == "glTF", "Invalid glb file!"
    version, length = struct.unpack("<II", f.read(8))
    assert version == 2
    length -= 12
    while length > 0:
        chunks.append(GLTFChunk.parse(f))
        length -= chunks[-1].length + 8
    return chunks

def load_glb(path: Path, debug: bool) -> dict:
    chunks: List[GLTFChunk] = []
    if str(path) != "-":
        with path.open("rb") as f:
            chunks = decode_glb(f)
    else:
        chunks = decode_glb(sys.stdin.buffer)
    assert chunks[0].type == "JSON"
    gltf = json.loads(chunks[0].data.decode())
    if debug:
        with path.with_suffix(".debug.gltf").open("w") as f:
            json.dump(gltf, f, indent=4)
    gltf["chunks"] = chunks[1:]
    return gltf

def write_glb(path: Path, gltf: dict) -> int:
    chunks: List[GLTFChunk] = gltf["chunks"]
    written = 0
    path.mkdir
    json_data = json.dumps({key: value for key, value in gltf.items() if key != "chunks"}, separators=(',',':')).encode()
    chunks = [GLTFChunk(len(json_data), "JSON", json_data)] + chunks
    with path.open("wb") as f:
        written += f.write(b"glTF")
        written += f.write(struct.pack("<II", 2, 12 + 8 * len(chunks) + sum([chunk.length for chunk in chunks])))
        for chunk in chunks:
            written += f.write(struct.pack("<I4s", chunk.length, chunk.type.encode()))
            written += f.write(chunk.data)
    return written

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
    parser.add_argument("--packall", action="store_true", help="Pack all images")

    script_path = Path(os.path.realpath(__file__)).parent
    resource_path = script_path / "resources"

    args = parser.parse_args()
    input_model: Path = args.input_model
    output: Path = args.output

    print("Beginning import...")

    gltf = load_glb(input_model, args.debug)
    assert gltf["asset"]["generator"] == "https://github.com/xypwn/filediver", f"GLB file was not created by Filediver! (Generator: {gltf['asset']['generator']})"

    print("Deleting Default Cube o7")
    bpy.data.objects["Cube"].select_set(True)
    bpy.ops.object.delete()
    bpy.ops.outliner.orphans_purge()
    print(f"Loading {input_model.name}")
    try:
        path = input_model
        if str(path) == "-":
            path = Path("tmp.glb")
            write_glb(path, gltf)
        bpy.ops.import_scene.gltf(filepath=str(path), import_shading="SMOOTH", bone_heuristic="TEMPERANCE")
    finally:
        if str(input_model) == "-":
            path.unlink()
    print("Loading TheJudSub's HD2 accurate shader")
    shader_script = bpy.data.texts.load(str(resource_path / "Helldivers2 shader script v1.0.6-1.py"))
    shader_script.use_fake_user = True
    shader_module = shader_script.as_module()
    with bpy.data.libraries.load(str(resource_path / "Helldivers2 Shader v1.0.5.blend")) as (shader_blend, our_blend):
        our_blend: BlendData # not actually but they share member names 
        shader_blend: BlendData
        our_blend.materials = shader_blend.materials
    shader_mat = bpy.data.materials["HD2 Shader"]
    shader_mat.use_fake_user = True

    optional_usages = ["decal_sheet", "pattern_masks_array"]
    unused_texture = bpy.data.images.new("unused", 1, 1, alpha=True, float_buffer=True)
    #unused_texture.pixels[3] = 0.0
    exr = make_exr(np.zeros((1, 1, 4), dtype=np.float32))
    unused_texture.use_fake_user = True
    unused_texture.pack(data=exr, data_len=len(exr))
    unused_texture.source = "FILE"
    unused_texture.file_format = "PNG"

    unused_secondary_lut = bpy.data.images.new("unused_secondary_lut", 23, 1, alpha=True)
    exr = make_exr(np.zeros((1, 23, 4), dtype=np.float32))
    unused_secondary_lut.file_format = "OPEN_EXR"
    unused_secondary_lut.use_fake_user = True
    unused_secondary_lut.colorspace_settings.name = "Non-Color"
    unused_secondary_lut.alpha_mode = "CHANNEL_PACKED"
    unused_secondary_lut.pack(data=exr, data_len=len(exr))
    unused_secondary_lut.source = "FILE"

    materialTextures: Dict[str, Dict[str, Image]] = {}

    print("Applying materials to meshes...")
    for node in gltf["nodes"]:
        if "mesh" not in node:
            continue
        mesh = gltf["meshes"][node["mesh"]]
        textures: Dict[str, Image] = {}
        assert len(mesh["primitives"]) == 1
        primitive = mesh["primitives"][0]
        if "material" not in primitive:
            continue
        material = gltf["materials"][primitive["material"]]
        if material["name"] in materialTextures:
            textures = materialTextures[material["name"]]
        else:
            is_pbr = "albedo" in material["extras"] or "albedo_iridescence" in material["extras"] or "normal" in material["extras"]
            if len(material["extras"]) == 0 or ("material_lut" not in material["extras"] and not is_pbr):
                continue
            if not args.packall and is_pbr:
                continue
            print(f"    Packing textures for material {material['name']}")
            try:
                for usage, texIdx in material["extras"].items():
                    textures[usage] = add_texture(gltf, texIdx, usage)
                for usage in optional_usages:
                    if usage not in textures:
                        textures[usage] = unused_texture
                materialTextures[material["name"]] = textures
            except AssertionError as e:
                print(f"Error: {e}")
                continue
            if is_pbr:
                continue

        if node["name"] not in bpy.data.objects:
            for item in bpy.data.objects:
                if item.active_material and item.active_material.name == material["name"]:
                    obj: Object = item
                    break
        else:
            obj: Object = bpy.data.objects[node["name"]]
        if not ("HD2 Mat " + material["name"]) in bpy.data.materials:
            print("    Copying template material")
            object_mat = shader_mat.copy()
            object_mat.name = "HD2 Mat " + material["name"]
            template_group: ShaderNodeGroup = object_mat.node_tree.nodes["HD2 Shader Template"]
            template_group.node_tree = template_group.node_tree.copy()
            template_group.node_tree.name = object_mat.name + shader_module.ScriptVersion
            HD2_Shader = template_group.node_tree

            print("    Applying textures")
            config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
            config_nodes["Secondary Material LUT Texture"].image = unused_secondary_lut
            for usage, image in textures.items():
                match usage:
                    case "id_masks_array":
                        config_nodes["ID Mask Array Texture"].image = image
                        image.colorspace_settings.name = "Non-Color"
                    case "pattern_masks_array":
                        config_nodes["Pattern Mask Array"].image = image
                        image.colorspace_settings.name = "Non-Color"
                    case "decal_sheet":
                        config_nodes["Decal Texture"].image = image
                        image.colorspace_settings.name = "sRGB"
                        image.alpha_mode = "CHANNEL_PACKED"
                    case "material_lut":
                        config_nodes["Primary Material LUT Texture"].image = image
                        image.colorspace_settings.name = "Non-Color"
                    case "pattern_lut":
                        config_nodes["Pattern LUT Texture"].image = image
                        image.colorspace_settings.name = "Non-Color"
                    case "base_data":
                        config_nodes["Normal Map"].image = image
                        image.colorspace_settings.name = "Non-Color"
        
            print("    Finalizing material")
            shader_module.update_images(HD2_Shader, object_mat)
            shader_module.update_slot_defaults(HD2_Shader, object_mat)
            shader_module.connect_input_links(HD2_Shader)
            shader_module.update_array_uvs(object_mat)
        else:
            print(f"    Found existing material 'HD2 Mat {material['name']}'")
            object_mat = bpy.data.materials["HD2 Mat " + material["name"]]

        obj.material_slots[0].material = object_mat
        shader_module.add_bake_uvs(obj)
        obj.select_set(True)
        try:
            bpy.ops.object.shade_smooth()
        except RuntimeError:
            # Context incorrect? We don't actually need to do this so its okay if it fails
            pass
        print(f"Applied material to {node['name']}!")

    bpy.ops.wm.save_mainfile(filepath=str(output))



if __name__ == "__main__":
    main()