from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)
import bpy

class CapeMaterialLoader(FilediverMaterialLoaderInterface):
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
        object_mat.name = f"HD2 Cape " + config["name"]

        print("    Applying textures")
        config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
        for usage, image in textures.items():
            image.colorspace_settings.name = "Non-Color"
            match usage:
                case "cape_tear":
                    config_nodes["Image Texture"].image = image
                case "cape_scalar_fields":
                    config_nodes["Image Texture.001"].image = image
                case "cape_gradient":
                    config_nodes["Image Texture.002"].image = image
                case "base_data":
                    config_nodes["Image Texture.003"].image = image
                case "weathering_dirt":
                    config_nodes["Image Texture.004"].image = image
                case "weathering_special":
                    config_nodes["Image Texture.005"].image = image
                case "blood_splatter_tiler":
                    config_nodes["Image Texture.006"].image = image
                case "bug_splatter_tiler":
                    config_nodes["Image Texture.007"].image = image
                case "decal_sheet":
                    for i in range(8, 12):
                        config_nodes[f"Image Texture.{i:03d}"].image = image
                case "material_lut":
                    for i in range(100, 123):
                        config_nodes[f"Image Texture.{i:03d}"].image = image
                case "cape_lut":
                    for i in range(123, 139):
                        config_nodes[f"Image Texture.{i:03d}"].image = image
                case "palette_lut":
                    config_nodes["Image Texture.012"].image = image
        print("    Applying settings")
        cape_group = object_mat.node_tree.nodes['Group.015']
        weathering_group = object_mat.node_tree.nodes['Group.011']
        for name, setting in config["extras"].items():
            if name == "weathering_tile_factor":
                config_nodes["Value.051"].outputs[0].default_value = setting[0]
                continue
            if name == "blood_scale":
                config_nodes["Value.052"].outputs[0].default_value = setting[0]
                continue
            if name == "gunk_scale":
                config_nodes["Value.053"].outputs[0].default_value = setting[0]
                continue
            if name not in cape_group.inputs and name not in weathering_group.inputs:
                continue
            for group in (weathering_group, cape_group):
                if name not in group.inputs:
                    continue
                if "height_wetness_and_wash" in name:
                    group.inputs[name].default_value = setting[:3]
                    group.inputs[name + " w"].default_value = setting[3]
                    continue
                if len(setting) == 1:
                    group.inputs[name].default_value = setting[0]
                    continue
                group.inputs[name].default_value = setting
        print("    Finalizing material")

        object_mat["needsBakeUVs"] = True
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "material_lut" in config["extras"] and "cape_lut" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Cape"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None