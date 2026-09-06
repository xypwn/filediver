from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
    ShaderNodeMapping,
)
import bpy

class RiftPlantMaterialLoader(FilediverMaterialLoaderInterface):
    material: Material = None

    def load_material(self, resource_path: str) -> None:
        if f"HD2 {self.key()}" not in bpy.data.materials:
            with bpy.data.libraries.load(str(resource_path / "Helldivers2 Shader v1.0.5.blend")) as (shader_blend, our_blend):
                our_blend: BlendData # not actually but they share member names 
                shader_blend: BlendData
                our_blend.materials = shader_blend.materials
        self.material = bpy.data.materials[f"HD2 {self.key()}"]
        self.material.use_fake_user = True

    def add_material(self, config: dict, textures: Dict[str, bpy.types.Image]) -> bpy.types.Material:
        object_mat = self.material.copy()
        object_mat.name = f"HD2 {self.key()} " + config["name"]

        print("    Applying textures")
        config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
        for usage, image in textures.items():
            match usage:
                case "base_color_metal_map":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = 'CHANNEL_PACKED'
                case "normal_xy_ao_rough_map":
                    config_nodes["Image Texture.001"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "emissive_f_stop_10_intensity_map":
                    config_nodes["Image Texture.002"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = 'CHANNEL_PACKED'
        
        print("    Applying settings")
        emissive_group = object_mat.node_tree.nodes['Group']
        mapping: ShaderNodeMapping = object_mat.node_tree.nodes['Mapping']
        for name, setting in config["extras"].items():
            object_mat[name] = setting
            if name == "tiling":
                mapping.inputs['Scale'].default_value = [setting[0], setting[1], 1.0]
            if name not in emissive_group.inputs:
                continue
            emissive_group.inputs[name].default_value = setting[0]

        print("    Finalizing material")
        return object_mat

    def preprocess_config(self, data, gltf, materialTextures, config):
        _ = data
        _ = gltf
        _ = materialTextures
        return config

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "emissive_f_stop_10_intensity_map" in config["extras"] and "emissive_enabled_mask" in config["extras"] and "pulse_glow_anim_speed" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Rift Plant"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None