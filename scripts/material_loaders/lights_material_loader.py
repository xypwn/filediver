from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)
import bpy

class LightsMaterialLoader(FilediverMaterialLoaderInterface):
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
                case "base_data":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
        
        print("    Applying settings")
        group = object_mat.node_tree.nodes['Group']
        for name, setting in config["extras"].items():
            if name == "emissive_intensity":
                object_mat.node_tree.nodes['Principled BSDF'].inputs['Emission Strength'].default_value = setting[0]
                continue
            if name not in group.inputs:
                continue
            if len(setting) == 1:
                group.inputs[name].default_value = setting[0]
                continue
            group.inputs[name].default_value = setting

        print("    Finalizing material")
        object_mat["needsBakeUVs"] = False
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "albedo_color" in config["extras"] and "albedo_color_curavture" in config["extras"] and "use_uv_for_coloring" in config["extras"] and "emissive_intensity" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Lights"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None