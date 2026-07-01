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
                case "texture_lut":
                    config_nodes["Image Texture.001"].image = image
                    config_nodes["Image Texture.001"].interpolation = "Closest"
                    image.colorspace_settings.name = "sRGB"
                case "NAC":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "decal_sheet":
                    config_nodes["Image Texture.002"].image = image
                    image.colorspace_settings.name = "sRGB"
        
        print("    Applying settings")
        building_group = object_mat.node_tree.nodes['Group.002']
        for name, setting in config["extras"].items():
            if name not in building_group.inputs or name == "NAC":
                continue
            if "roughness_build_up" in name:
                building_group.inputs[name].default_value = setting[:3]
                building_group.inputs[name + " w"].default_value = setting[3]
                continue
            if len(setting) == 1:
                building_group.inputs[name].default_value = setting[0]
                continue
            building_group.inputs[name].default_value = setting
        if "decal_wear" not in config["extras"]:
            config_nodes["Image Texture.002"].image = textures["filediver_unused"]
        decal_uv_node: ShaderNodeUVMap = object_mat.node_tree.nodes['UV Map']
        suffix = f'.{config["extras"].get("filediver_decal_uvmap", 0):03d}' if config["extras"].get("filediver_decal_uvmap", 0) > 0 else ""
        decal_uv_node.uv_map = f"UVMap{suffix}"

        print("    Finalizing material")

        object_mat["needsBakeUVs"] = False
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