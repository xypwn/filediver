import bpy
import json
import os
import sys
import struct
import numpy as np
from argparse import ArgumentParser
from bpy.types import (
    BlendData,
    Image,
    Object,
    ShaderNodeGroup,
    ShaderNodeTexImage,
    Material,
    Collection,
    Armature,
    ActionKeyframeStrip,
    PoseBone,
    ActionConstraint,
)
from pathlib import Path
from io import BytesIO
from typing import Optional, Dict, List, Tuple, Union
from types import ModuleType
from random import randint
from dataclasses import dataclass
from copy import deepcopy
import math

from dds_float16 import DDS
from openexr.types import OpenEXR

class IDPropertyUIManager:
    def update(subtype=None, min=None, max=None, soft_min=None, soft_max=None, precision=None, step=None, default=None, id_type=None, items=None, description=None):
        """
        Update the RNA information of the IDProperty used for interaction and
        display in the user interface. The required types for many of the keyword
        arguments depend on the type of the property.
        """
    def update_from(ui_manager_source: 'IDPropertyUIManager'):
        """
        Copy UI data from an IDProperty in the source group to a property in this group.
          If the source property has no UI data, the target UI data will be reset if it exists.
          :raises TypeError: If the types of the two properties don't match.
        """
    def as_dict() -> dict:
        """
        Return a dictionary of the property's RNA UI data. The fields in the
        returned dictionary and their types will depend on the property's type.
        """
    def clear():
        """
        Remove the RNA UI data from this IDProperty.
        """

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

def add_texture(gltf, textureIdx, usage: Optional[str] = None) -> Image:
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
    if name in bpy.data.images:
        return bpy.data.images[name]
    blImage = bpy.data.images.new(name, width, height, alpha=True)
    blImage.pack(data=data, data_len=len(data))
    blImage.file_format = fmt
    blImage.source = "FILE"
    blImage.alpha_mode = "CHANNEL_PACKED"
    blImage.use_fake_user = True
    return blImage

def add_accurate_material(shader_mat: Material, material: dict, shader_module, unused_secondary_lut: Image, textures: Dict[str, Image]):
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
    return object_mat

def add_skin_material(skin_mat: Material, material: dict, textures: Dict[str, Image]):
    object_mat = skin_mat.copy()
    object_mat.name = "HD2 Mat " + material["name"]

    print("    Applying textures")
    config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
    for usage, image in textures.items():
        image.colorspace_settings.name = "Non-Color"
        match usage:
            case "color_roughness":
                config_nodes["Image Texture"].image = image
            case "normal_specular_ao":
                config_nodes["Image Texture.001"].image = image
    print("    Finalizing material")
    # Set ethnicity to a random value
    config_nodes["Value"].outputs[0].default_value = float(randint(0, 4))
    return object_mat

def add_lut_skin_material(skin_mat: Material, material: dict, textures: Dict[str, Image]):
    object_mat = skin_mat.copy()
    object_mat.name = "HD2 Mat " + material["name"]

    print("    Applying textures")
    config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
    for usage, image in textures.items():
        image.colorspace_settings.name = "Non-Color"
        match usage:
            case "color_roughness_lut":
                config_nodes["Image Texture"].image = image
                config_nodes["Image Texture"].interpolation = "Closest"
            case "normal_specular_ao":
                config_nodes["Image Texture.001"].image = image
            case "grayscale_skin":
                config_nodes["Image Texture.002"].image = image
                config_nodes["Image Texture.002"].interpolation = "Smart"
    print("    Finalizing material")
    # Set ethnicity to a random value
    config_nodes["Value"].outputs[0].default_value = float(randint(0, 4))
    return object_mat

def load_shaders(resource_path: str) -> Tuple[ModuleType, Material, Material, Material]:
    shader_script = bpy.data.texts.load(str(resource_path / "Helldivers2 shader script v1.0.6-1.py"))
    shader_script.use_fake_user = True
    shader_module = shader_script.as_module()
    with bpy.data.libraries.load(str(resource_path / "Helldivers2 Shader v1.0.5.blend")) as (shader_blend, our_blend):
        our_blend: BlendData # not actually but they share member names 
        shader_blend: BlendData
        our_blend.materials = shader_blend.materials
    shader_mat = bpy.data.materials["HD2 Shader"]
    shader_mat.use_fake_user = True
    skin_mat = bpy.data.materials["HD2 Skin"]
    skin_mat.use_fake_user = True
    lut_skin_mat = bpy.data.materials["HD2 LUT Skin"]
    lut_skin_mat.use_fake_user = True
    return shader_module, shader_mat, skin_mat, lut_skin_mat

def create_empty_texture(name: str, size: Tuple[int, int], fmt: str = 'PNG', colorspace: str = 'sRGB', alpha_mode: str = 'CHANNEL_PACKED') -> Image:
    unused_texture = bpy.data.images.new(name, size[0], size[1], alpha=True, float_buffer=True)
    #unused_texture.pixels[3] = 0.0
    exr = OpenEXR.from_pixels(np.zeros((size[1], size[0], 4), dtype=np.float32)).serialize()
    unused_texture.use_fake_user = True
    unused_texture.pack(data=exr, data_len=len(exr))
    unused_texture.source = "FILE"
    unused_texture.file_format = fmt
    unused_texture.colorspace_settings.name = colorspace
    unused_texture.alpha_mode = alpha_mode
    return unused_texture

def add_to_armor_set(node: Dict):
    if node["extras"]["armorSet"] not in bpy.data.collections:
        bpy.data.collections.new(node["extras"]["armorSet"])
        bpy.data.scenes[0].collection.children.link(bpy.data.collections[node["extras"]["armorSet"]])
    collection: Collection = bpy.data.collections[node["extras"]["armorSet"]]
    obj: Object = bpy.data.objects[node["name"]]
    for other in obj.users_collection:
        other.objects.unlink(obj)
    collection.objects.link(obj)

def convert_materials(gltf: Dict, node: Dict, variants: List[Dict], hasVariants: bool, materialTextures: Dict[int, Dict[str, Image]], packall: bool, shader_module: ModuleType, shader_mat: Material, skin_mat: Material, lut_skin_mat: Material, unused_texture: Image, unused_secondary_lut: Image):
    optional_usages = ["decal_sheet", "pattern_masks_array"]

    mesh = gltf["meshes"][node["mesh"]]
    textures: Dict[str, Image] = {}
    if node["name"] in bpy.data.objects:
        obj: Object = bpy.data.objects[node["name"]]
    else:
        obj = None
    for primIdx, primitive in enumerate(mesh["primitives"]):
        for varIdx in range(len(variants)):
            if "material" not in primitive:
                continue
            materialIndex = primitive["material"]
            mappings = primitive.get("extensions", {}).get("KHR_materials_variants", {}).get("mappings")
            if hasVariants and mappings is not None:
                for mapping in mappings:
                    if varIdx in mapping["variants"]:
                        materialIndex = mapping["material"]
                        break
            material = gltf["materials"][materialIndex]
            if node["name"] not in bpy.data.objects and obj is None:
                for item in bpy.data.objects:
                    if item.active_material and item.active_material.name == material["name"]:
                        obj: Object = item
                        break
            is_pbr = "albedo" in material["extras"] or "albedo_iridescence" in material["extras"] or "normal" in material["extras"]
            is_tex_array_skin = "color_roughness" in material["extras"] and "normal_specular_ao" in material["extras"] and len(material["extras"]) == 2
            is_lut_skin = "grayscale_skin" in material["extras"] and "color_roughness_lut" in material["extras"]
            is_lut = "material_lut" in material["extras"]
            if materialIndex in materialTextures:
                textures = materialTextures[materialIndex]
            else:
                if len(material["extras"]) == 0 or not any((is_pbr, is_tex_array_skin, is_lut_skin, is_lut)):
                    continue
                if not packall and is_pbr:
                    continue
                print(f"    Packing textures for material {material['name']}")
                try:
                    for usage, texIdx in material["extras"].items():
                        if type(texIdx) != int:
                            continue
                        textures[usage] = add_texture(gltf, texIdx, usage)
                    for usage in optional_usages:
                        if usage not in textures:
                            textures[usage] = unused_texture
                    materialTextures[materialIndex] = textures
                except AssertionError as e:
                    print(f"Error: {e}")
                    continue
                if is_pbr:
                    continue

            key = "HD2 Mat " + material["name"]
            i = 1
            while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != materialIndex:
                key = "HD2 Mat " + material["name"] + f".{i:03d}"
                i += 1
            if not key in bpy.data.materials:
                print("    Copying template material")
                if is_lut:
                    object_mat = add_accurate_material(shader_mat, material, shader_module, unused_secondary_lut, textures)
                elif is_tex_array_skin:
                    object_mat = add_skin_material(skin_mat, material, textures)
                elif is_lut_skin:
                    object_mat = add_lut_skin_material(lut_skin_mat, material, textures)
                object_mat["gltfId"] = materialIndex
            else:
                print(f"    Found existing material '{key}'")
                object_mat = bpy.data.materials[key]

            obj.material_slots[primIdx].material = object_mat
            if hasVariants:
                # Taken from gltf2 extension
                mesh = obj.data
                found = False
                variant_found = False
                for i in mesh.gltf2_variant_mesh_data:
                    if i.material_slot_index == primIdx and i.material == object_mat:
                        found = True
                        variant_primitive = i
                    elif i.material_slot_index == primIdx and varIdx == i.variants[0].variant.variant_idx:
                        found = True
                        variant_found = True
                        variant_primitive = i
                        i.material = object_mat
                if not found:
                    variant_primitive = obj.data.gltf2_variant_mesh_data.add()
                    variant_primitive.material_slot_index = primIdx
                    variant_primitive.material = object_mat
                if not variant_found:
                    vari = variant_primitive.variants.add()
                    vari.variant.variant_idx = varIdx

    shader_module.add_bake_uvs(obj)
    obj.select_set(True)
    print(f"Applied material to {node['name']}!")

def hide_visibility_group(node: Dict):
    obj: Object = bpy.data.objects[node["name"]]
    obj.hide_render = True
    obj.hide_set(True)

@dataclass
class GLTFAnimation:
    name: str
    index: int

@dataclass
class GLTFDriverInformation:
    expression: str
    variables: List[str]
    limits: List[Tuple[float, float, float]]
    type: str

@dataclass
class GLTFState:
    name: str
    type: str
    blend_mask: int
    animations: List[GLTFAnimation] = None
    loop: bool = True
    additive: bool = True
    blend_variable: str = None
    custom_blend_functions: List[GLTFDriverInformation] = None
    ragdoll_name: str = None
    
    @classmethod
    def from_json(cls, data: dict) -> 'GLTFState':
        animations = None
        custom_blend_functions = None
        if "animations" in data:
            animations = [GLTFAnimation(**animation) for animation in data["animations"]]
            del data["animations"]
        if "custom_blend_functions" in data:
            custom_blend_functions = [GLTFDriverInformation(**blend_function) for blend_function in data["custom_blend_functions"]]
            del data["custom_blend_functions"]
        return cls(animations=animations, custom_blend_functions=custom_blend_functions, **data)

@dataclass
class GLTFLayer:
    default_state: int
    states: List[GLTFState]
    
    @classmethod
    def from_json(cls, data: dict) -> 'GLTFLayer':
        states = [GLTFState.from_json(state) for state in data["states"]]
        del data["states"]
        return cls(states=states, **data)

@dataclass
class GLTFVariable:
    name: str
    default: float

@dataclass
class GLTFStateMachine:
    name: str
    layers: List[GLTFLayer]
    animation_events: List[str]
    animation_variables: List[GLTFVariable]
    blend_masks: List[Dict[str, float]]
    all_bones: List[str]
    
    @classmethod
    def from_json(cls, data: dict) -> 'GLTFStateMachine':
        # don't be destructive for no reason
        data = deepcopy(data)
        layers = [GLTFLayer.from_json(layer) for layer in data["layers"]]
        animation_variables = [GLTFVariable(**variable) for variable in data["animation_variables"]]
        del data["animation_variables"]
        del data["layers"]
        return cls(layers=layers, animation_variables=animation_variables, **data)

def no_set(self, val):
    return

def add_state_machine(gltf: Dict, node: Dict):
    obj: Object = bpy.data.objects[node["name"]]
    assert type(obj.data) is Armature
    controller = GLTFStateMachine.from_json(gltf["extras"]["state_machines"][0])
    print(f"    Adding state machine {controller.name} to {obj.name}")
    print(f"        Adding {len(controller.animation_variables)} variables")

    for variable in controller.animation_variables:
        obj[variable.name] = float(variable.default)
        obj.id_properties_ui(variable.name).update(min=variable.default, max=variable.default, default=variable.default)

    state_machine_empty = bpy.data.objects.new(obj.name+".state_machine", None)
    state_machine_empty.parent = obj
    state_machine_empty.empty_display_type = 'SINGLE_ARROW'
    state_machine_empty.hide_select = True
    state_machine_empty.hide_viewport = True
    for collection in obj.users_collection:
        collection.objects.link(state_machine_empty)

    for layerIdx, layer in enumerate(controller.layers):
        layer_empty = bpy.data.objects.new(f"{obj.name} layer {layerIdx}", None)
        layer_empty.parent = state_machine_empty
        layer_empty.empty_display_size = 0.1
        layer_empty.empty_display_type = 'CUBE'
        layer_empty["state"] = layer.default_state
        layer_empty.id_properties_ui("state").update(min=0, max=len(layer.states)-1)
        for collection in obj.users_collection:
            collection.objects.link(layer_empty)

        layer_action = bpy.data.actions.new(f"{obj.name} layer {layerIdx}")
        slot = layer_action.slots.new('OBJECT', obj.name)
        anim = layer_empty.animation_data_create()
        anim.action = layer_action
        anim.action_slot = slot
        layer_empty.keyframe_insert(data_path='["state"]', frame=1.0)
        for action_layer in layer_action.layers:
            for strip in action_layer.strips:
                if strip.type != 'KEYFRAME':
                    continue
                strip: ActionKeyframeStrip
                bag = strip.channelbag(slot)
                for curve in bag.fcurves:
                    for frame in curve.keyframe_points:
                        frame.interpolation = 'CONSTANT'

        track = obj.animation_data.nla_tracks.new()
        track.name = layer_action.name
        nla_strip = track.strips.new(layer_action.name, start=1, action=layer_action)
        nla_strip.action_frame_end = 250.0
        obj.animation_data.action = layer_action
        for stateIdx, state in enumerate(layer.states):
            objects_to_constrain: List[Tuple[PoseBone, float]] = []
            if state.blend_mask != -1:
                for key, value in controller.blend_masks[state.blend_mask].items():
                    objects_to_constrain.append((obj.pose.bones.get(key), float(value)))
            else:
                for name in controller.all_bones:
                    objects_to_constrain.append((obj.pose.bones.get(name), 1.0))
            if state.animations is None:
                continue
            time_blend_function = None
            if state.type == "StateType_CustomBlend" and len(state.animations) != len(state.custom_blend_functions) or state.type == "StateType_Clip" and state.custom_blend_functions is not None:
                for i, time_blend_fn in enumerate(state.custom_blend_functions):
                    if time_blend_fn.type == "DriverType_EvalTime":
                        time_blend_function = time_blend_fn
                        state.custom_blend_functions.pop(i)
                        break
                if len(state.custom_blend_functions) == 0:
                    state.custom_blend_functions = None
            for animIdx, animation in enumerate(state.animations):
                animation_action = bpy.data.actions[animation.name]
                for cObj, mask_influence in objects_to_constrain:
                    constraint: ActionConstraint = cObj.constraints.new('ACTION')
                    constraint.name = f"{cObj.name} layer {layerIdx} state {stateIdx} animation {animIdx}"
                    constraint.action = animation_action
                    constraint.use_eval_time = True
                    if mask_influence == 1:
                        constraint.mix_mode = 'REPLACE'
                    else:
                        constraint.mix_mode = 'AFTER_SPLIT'
                    influence_driver_expression = f"({mask_influence}) * float(state == {stateIdx})"
                    variables = [(layer_empty, "state")]
                    if state.type == "StateType_CustomBlend" and state.custom_blend_functions is not None and animIdx < len(state.custom_blend_functions):
                        influence_driver_expression += f" * {state.custom_blend_functions[animIdx].expression}"
                        variables.extend(zip([obj] * len(state.custom_blend_functions[animIdx].variables), state.custom_blend_functions[animIdx].variables))
                    influence = constraint.driver_add("influence")
                    for vObj, variableName in variables:
                        variable = influence.driver.variables.new()
                        variable.targets[0].id = vObj
                        variable.targets[0].data_path = f'["{variableName}"]'
                        variable.name = variableName
                        manager: IDPropertyUIManager = vObj.id_properties_ui(variableName)
                        manager_dict = manager.as_dict()
                        if state.custom_blend_functions is None or variableName not in state.custom_blend_functions[animIdx].variables:
                            continue
                        varIdx = state.custom_blend_functions[animIdx].variables.index(variableName)
                        limits = state.custom_blend_functions[animIdx].limits[varIdx]
                        if limits[1] < manager_dict["min"]:
                            manager.update(min=limits[1], soft_min=limits[1])
                        if limits[1] > manager_dict["max"]:
                            manager.update(max=limits[1], soft_max=limits[1])
                        manager_dict = manager.as_dict()
                        if manager_dict["max"] - manager_dict["min"] > 10:
                            manager.update(step=1.0)

                    influence.driver.expression = influence_driver_expression

                    animation_track = obj.animation_data.nla_tracks.get(animation.name)
                    animation_strip = animation_track.strips[0]
                    time = constraint.driver_add("eval_time")
                    if state.type == "StateType_Blend1D":
                        variable = time.driver.variables.new()
                        variable.targets[0].id = obj
                        variable.targets[0].data_path = f'["{state.blend_variable}"]'
                        variable.name = state.blend_variable
                        manager: IDPropertyUIManager = obj.id_properties_ui(state.blend_variable)
                        manager.update(min=0.0, max=1.0, soft_min=0.0, soft_max=1.0)
                        time.driver.expression = f"clamp({state.blend_variable})"
                    elif not state.loop:
                        layer_empty["start_frame"] = float(1.0)
                        manager: IDPropertyUIManager = layer_empty.id_properties_ui("start_frame")
                        manager.update(min=0.0, max=math.inf)
                        variable = time.driver.variables.new()
                        variable.targets[0].id = layer_empty
                        variable.targets[0].data_path = '["start_frame"]'
                        variable.name = "start_frame"
                        time.driver.expression = f"clamp((frame - start_frame) / {animation_strip.frame_end - animation_strip.frame_start})"
                    elif time_blend_function is not None:
                        if time_blend_function.expression != "0.0":
                            variable = time.driver.variables.new()
                            variable.targets[0].id = obj
                            variable.targets[0].data_path = f'["{time_blend_function.variables[0]}"]'
                            variable.name = time_blend_function.variables[0]
                            manager: IDPropertyUIManager = obj.id_properties_ui(time_blend_function.variables[0])
                            if manager.as_dict()["max"] < time_blend_function.limits[0][1]:
                                manager.update(max=math.inf, soft_max=math.inf)
                            variableFps = time.driver.variables.new()
                            variableFps.targets[0].id_type = 'SCENE'
                            variableFps.targets[0].id = bpy.data.scenes[0]
                            variableFps.targets[0].data_path = "render.fps"
                            variableFps.name = "fps"
                        time.driver.expression = time_blend_function.expression
                    else:
                        time.driver.expression = f"fmod(frame, max(1.0, {animation_strip.frame_end - animation_strip.frame_start})) / max(1.0, {animation_strip.frame_end - animation_strip.frame_start})"
                    constraint.frame_start = int(animation_strip.frame_start)-1
                    constraint.frame_end = int(animation_strip.frame_end)

    obj.animation_data.action = None

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
        if "extras" in gltf and "frameRate" in gltf["extras"]:
            print(f'Setting FPS to {gltf["extras"]["frameRate"]}')
            bpy.context.scene.render.fps = gltf["extras"]["frameRate"]
        bpy.ops.import_scene.gltf(filepath=str(path), bone_heuristic="TEMPERANCE")
    finally:
        if str(input_model) == "-":
            path.unlink()
    print("Loading TheJudSub's HD2 accurate shader")
    shader_module, shader_mat, skin_mat, lut_skin_mat = load_shaders(resource_path)

    unused_texture = create_empty_texture("unused", (1, 1))
    unused_secondary_lut = create_empty_texture("unused_secondary_lut", (23, 1), fmt='OPEN_EXR', colorspace='Non-Color')

    materialTextures: Dict[int, Dict[str, Image]] = {}

    variants = gltf.get("extensions", {}).get("KHR_materials_variants", {}).get("variants")
    hasVariants = variants is not None

    if hasVariants and not bpy.context.preferences.addons['io_scene_gltf2'].preferences.KHR_materials_variants_ui:
        print("Enabling variants UI")
        bpy.context.preferences.addons['io_scene_gltf2'].preferences.KHR_materials_variants_ui = True

    if variants is None:
        variants = [{
            "name": "default"
        }]

    print("Applying helldivers customizations...")
    for node in gltf["nodes"]:
        if node.get("extras", {}).get("armorSet") is not None and node["name"] in bpy.data.objects:
            add_to_armor_set(node)
        if node.get("extras", {}).get("default_hidden") == 1 and node["name"] in bpy.data.objects:
            hide_visibility_group(node)
        if "mesh" in node:
            convert_materials(gltf, node, variants, hasVariants, materialTextures, args.packall, shader_module, shader_mat, skin_mat, lut_skin_mat, unused_texture, unused_secondary_lut)
        if "state_machine" in node.get("extras", {}):
            add_state_machine(gltf, node)
        children = node.get("children")
        if node["name"] in bpy.data.objects and children is not None and gltf["nodes"][children[0]]["name"] == "StingrayEntityRoot":
            object: Object = bpy.data.objects[node["name"]]
            object.data.display_type = "WIRE"
        

    if hasVariants:
        # Reset to default variant
        bpy.data.scenes[0].gltf2_active_variant = 0
        for obj in bpy.data.objects:
            if obj.type != "MESH":
                continue
            mesh = obj.data
            for i in mesh.gltf2_variant_mesh_data:
                if i.variants[0].variant.variant_idx == 0:
                    obj.material_slots[i.material_slot_index].material = i.material

    bpy.ops.wm.save_mainfile(filepath=str(output))



if __name__ == "__main__":
    main()