from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)
import bpy

class IlluminateBuildingMonoplanarMaterialLoader(FilediverMaterialLoaderInterface):
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
        object_mat.name = f"HD2 IllBldg " + config["name"]

        print("    Applying textures")
        config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
        for usage, image in textures.items():
            match usage:
                case "mask":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "albedo_array":
                    config_nodes["Image Texture.001"].image = image
                    config_nodes["Image Texture.002"].image = image
                    config_nodes["Image Texture.003"].image = image
                    config_nodes["Image Texture.004"].image = image
                    config_nodes["Image Texture.005"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "normal_array":
                    config_nodes["Image Texture.006"].image = image
                    config_nodes["Image Texture.007"].image = image
                    config_nodes["Image Texture.008"].image = image
                    config_nodes["Image Texture.009"].image = image
                    config_nodes["Image Texture.010"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "normal_map":
                    config_nodes["Image Texture.011"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "emissive":
                    config_nodes["Image Texture.012"].image = image
                    image.colorspace_settings.name = "Non-Color"

        print("    Applying settings")
        for name, setting in config["extras"].items():
            match name:
                case "surface_tiling":
                    object_mat.node_tree.nodes['Value.003'].outputs[0].default_value = setting[0]
                case "emissive_color":
                    object_mat.node_tree.nodes['RGB'].outputs[0].default_value = setting[0:3] + [1]
                case "emissive_power":
                    object_mat.node_tree.nodes['Principled BSDF.001'].inputs['Emission Strength'].default_value = setting[0]

        print("    Finalizing material")

        object_mat["needsBakeUVs"] = True
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "albedo_array" in config["extras"] and "surface_tiling" in config["extras"] and "bcm_tex_a" not in config["extras"] and "noise_power" not in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Illuminate Building Single Plane"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 IllBldg " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 IllBldg " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None