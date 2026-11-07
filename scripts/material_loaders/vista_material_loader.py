from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
    ShaderNodeGroup,
)
import bpy

class VistaMaterialLoader(FilediverMaterialLoaderInterface):
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
        config_nodes: Dict[str, ShaderNodeTexImage|ShaderNodeGroup] = object_mat.node_tree.nodes
        asset_color_grading_lut_group: Dict[str, ShaderNodeTexImage] = bpy.data.node_groups["asset_color_grading_lut"].nodes

        for usage, image in textures.items():
            match usage:
                case "asset_color_grading_lut":
                    asset_color_grading_lut_group["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                    asset_color_grading_lut_group["Image Texture"].interpolation = "Closest"
                case "detail_mask_":
                    config_nodes["Image Texture.009"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "base_normal_ao":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "triplanar_detail_albedo":
                    config_nodes["Image Texture.004"].image = image
                    config_nodes["Image Texture.005"].image = image
                    config_nodes["Image Texture.006"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "triplanar_detail_data":
                    config_nodes["Image Texture.001"].image = image
                    config_nodes["Image Texture.002"].image = image
                    config_nodes["Image Texture.003"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "ground_albedo":
                    config_nodes["Image Texture.007"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "ground_nar":
                    config_nodes["Image Texture.008"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "vista_detail_albedo":
                    config_nodes["Image Texture.010"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "vista_detail_nar":
                    config_nodes["Image Texture.011"].image = image
                    image.colorspace_settings.name = "Non-Color"
        print("    Finalizing material")

        for name, setting in config["extras"].items():
            object_mat[name] = setting
            if name in config_nodes["Group"].inputs:
                config_nodes["Group"].inputs[name].default_value = setting[0]
            elif name in config_nodes["Group.001"].inputs:
                config_nodes["Group.001"].inputs[name].default_value = setting[0]
            elif name in config_nodes["Group.002"].inputs:
                config_nodes["Group.002"].inputs[name].default_value = setting[0]

        object_mat["needsBakeUVs"] = False
        return object_mat

    def preprocess_config(self, data, gltf, materialTextures, config):
        _ = data
        _ = gltf
        _ = materialTextures
        return config

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "base_normal_ao" in config["extras"] and "triplanar_detail_albedo" in config["extras"] and "triplanar_detail_data" in config["extras"] and "detail_mask_" in config["extras"] and "grading_group_id_vista_detail" in config["extras"] and "asset_color_grading_lut" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Vista"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None
