from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from types import ModuleType
from typing import Dict, Tuple, Optional

from bpy.types import (
    BlendData,
    Material,
    Image,
    ShaderNodeGroup,
    ShaderNodeTexImage,
)
import bpy

import numpy as np

from openexr.types import OpenEXR

def create_empty_texture(name: str, size: Tuple[int, int], fmt: str = 'PNG', colorspace: str = 'sRGB', alpha_mode: str = 'CHANNEL_PACKED') -> Image:
    unused_texture = bpy.data.images.new(name, size[0], size[1], alpha=True, float_buffer=True)
    exr = OpenEXR.from_pixels(np.zeros((size[1], size[0], 4), dtype=np.float32)).serialize()
    unused_texture.use_fake_user = True
    unused_texture.pack(data=exr, data_len=len(exr))
    unused_texture.source = "FILE"
    unused_texture.file_format = fmt
    unused_texture.colorspace_settings.name = colorspace
    unused_texture.alpha_mode = alpha_mode
    return unused_texture

class ArmorMaterialLoader(FilediverMaterialLoaderInterface):
    shader_module: ModuleType = None
    material: Material = None
    unused_secondary_lut: Image = None

    def load_material(self, resource_path: str) -> None:
        shader_script = bpy.data.texts.load(str(resource_path / "Helldivers2 shader script v1.0.6-1.py"))
        shader_script.use_fake_user = True
        self.shader_module = shader_script.as_module()
        if "HD2 Shader" not in bpy.data.materials:
            with bpy.data.libraries.load(str(resource_path / "Helldivers2 Shader v1.0.5.blend")) as (shader_blend, our_blend):
                our_blend: BlendData # not actually but they share member names 
                shader_blend: BlendData
                our_blend.materials = shader_blend.materials
        self.material = bpy.data.materials["HD2 Shader"]
        self.material.use_fake_user = True

        if self.unused_secondary_lut is None:
            self.unused_secondary_lut = create_empty_texture("unused_secondary_lut", (23, 1), fmt='OPEN_EXR', colorspace='Non-Color')

    def add_material(self, config: dict, textures: Dict[str, bpy.types.Image]) -> bpy.types.Material:
        object_mat = self.material.copy()
        object_mat.name = f"HD2 {self.key()} " + config["name"]
        template_group: ShaderNodeGroup = object_mat.node_tree.nodes["HD2 Shader Template"]
        template_group.node_tree = template_group.node_tree.copy()
        template_group.node_tree.name = object_mat.name + self.shader_module.ScriptVersion
        HD2_Shader = template_group.node_tree

        print("    Applying textures")
        config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
        config_nodes["Secondary Material LUT Texture"].image = self.unused_secondary_lut
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
        
        detail_tile_factor_mult = config.get("extras", {}).get("detail_tile_factor_mult")
        if detail_tile_factor_mult is not None:
            config_nodes["HD2 Shader Template"].inputs['detail_tile_factor_mult'].default_value = detail_tile_factor_mult[0]

        print("    Finalizing material")
        self.shader_module.update_images(HD2_Shader, object_mat)
        self.shader_module.update_slot_defaults(HD2_Shader, object_mat)
        self.shader_module.connect_input_links(HD2_Shader)
        self.shader_module.update_array_uvs(object_mat)

        object_mat["needsBakeUVs"] = True
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "material_lut" in config["extras"] and not "cape_lut" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Mat"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None