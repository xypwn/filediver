from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
    ShaderNodeGroup,
)
import bpy

class TerrainMaterialLoader(FilediverMaterialLoaderInterface):
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
                case "albedo_blend_tex" | "albedo_tex":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "displacement_tex":
                    config_nodes["Image Texture.001"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "nar_tex":
                    config_nodes["Image Texture.002"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "asset_color_grading_lut":
                    asset_color_grading_lut_group["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                    asset_color_grading_lut_group["Image Texture"].interpolation = "Closest"
        print("    Finalizing material")

        for name, setting in config["extras"].items():
            object_mat[name] = setting
            if name in config_nodes["Terrain Color Grading"].inputs:
                config_nodes["Terrain Color Grading"].inputs[name].default_value = setting[0]
            if name in config_nodes["Terrain Emissive"].inputs:
                if name in ["emissive_base_color_tint", "emissive_color_ao"]:
                    config_nodes["Terrain Emissive"].inputs[name].default_value = setting[:3]
                elif name in ["emissive_base_color_luminance_range", "emissive_ao_range"]:
                    config_nodes["Terrain Emissive"].inputs[name].default_value = setting[:2]
                else:
                    config_nodes["Terrain Emissive"].inputs[name].default_value = setting[0]

        object_mat["needsBakeUVs"] = False
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return ("albedo_blend_tex" in config["extras"] or "albedo_tex") and "nar_tex" in config["extras"] and "asset_color_grading_lut" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Terrain"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None