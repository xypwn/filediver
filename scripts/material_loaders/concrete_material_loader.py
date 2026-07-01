from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
    ShaderNodeUVMap,
)
import bpy

class ConcreteMaterialLoader(FilediverMaterialLoaderInterface):
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
                case "pattern_data":
                    config_nodes["Image Texture"].image = image
                    config_nodes["Image Texture.001"].image = image
                    config_nodes["Image Texture.002"].image = image
                    config_nodes["Image Texture.003"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "surface_data_array":
                    config_nodes["Image Texture.004"].image = image
                    config_nodes["Image Texture.005"].image = image
                    config_nodes["Image Texture.006"].image = image
                    config_nodes["Image Texture.007"].image = image
                    image.colorspace_settings.name = "Non-Color"
        
        print("    Applying settings")
        concrete_group = object_mat.node_tree.nodes['Group']
        pattern_uv_group = object_mat.node_tree.nodes['Group.001']
        surface_uv_group = object_mat.node_tree.nodes['Group.003']
        for name, setting in config["extras"].items():
            if name not in concrete_group.inputs:
                continue
            if name == "disable_triplanar_distance":
                setting[0] *= 100
            if len(setting) == 1:
                concrete_group.inputs[name].default_value = setting[0]
                continue
            if "color" in name and len(setting) == 3:
                setting = setting + [1]
            concrete_group.inputs[name].default_value = setting

        for name, setting in config["extras"].items():
            if name not in pattern_uv_group.inputs:
                continue
            if len(setting) == 1:
                pattern_uv_group.inputs[name].default_value = setting[0]
                continue
            pattern_uv_group.inputs[name].default_value = setting

        for name, setting in config["extras"].items():
            if name not in surface_uv_group.inputs:
                continue
            if len(setting) == 1:
                surface_uv_group.inputs[name].default_value = setting[0]
                continue
            surface_uv_group.inputs[name].default_value = setting

        print("    Finalizing material")
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "pattern_data" in config["extras"] and "material_surface" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Concrete"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None