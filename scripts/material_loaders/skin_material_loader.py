from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)
import bpy

from random import randint

class SkinMaterialLoader(FilediverMaterialLoaderInterface):
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
            image.colorspace_settings.name = "Non-Color"
            match usage:
                case "color_roughness":
                    config_nodes["Image Texture"].image = image
                case "normal_specular_ao":
                    config_nodes["Image Texture.001"].image = image
        print("    Finalizing material")
        # Set ethnicity to a random value
        config_nodes["Value"].outputs[0].default_value = float(randint(0, 4))

        object_mat["needsBakeUVs"] = True
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "color_roughness" in config["extras"] and "normal_specular_ao" in config["extras"] and len(config["extras"]) == 2

    @classmethod
    def key(cls) -> str:
        return "Skin"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None