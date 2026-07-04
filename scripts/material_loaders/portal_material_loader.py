from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)
import bpy

class PortalMaterialLoader(FilediverMaterialLoaderInterface):
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
                case "noise_map_01":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "noise_map_02":
                    config_nodes["Image Texture.001"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "edge_noise_map":
                    config_nodes["Image Texture.002"].image = image
                    image.colorspace_settings.name = "Non-Color"
                

        print("    Applying settings")
        for name, setting in config["extras"].items():
            for i in range(4):
                key = f"Group{f'.{i:03d}' if (i > 0) else ''}"
                group = config_nodes.get(key)
                if group is None:
                    continue
                if name not in group.inputs:
                    continue
                if len(setting) == 1:
                    group.inputs[name].default_value = setting[0]
                    continue
                group.inputs[name].default_value = setting[:3]

        print("    Finalizing material")

        object_mat["needsBakeUVs"] = True
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "noise_map_01" in config["extras"] and "noise_map_02" in config["extras"] and "edge_noise_map" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Portal"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None