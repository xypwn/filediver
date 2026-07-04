from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
)
import bpy

class FenceMaterialLoader(FilediverMaterialLoaderInterface):
    material: Material = None

    def load_material(self, resource_path: str) -> None:
        if f"HD2 {self.key()}" not in bpy.data.materials:
            with bpy.data.libraries.load(str(resource_path / "Helldivers2 Shader v1.0.5.blend")) as (shader_blend, our_blend):
                our_blend: BlendData # not actually but they share member names 
                shader_blend: BlendData
                our_blend.materials = shader_blend.materials
        self.material = bpy.data.materials[f"HD2 {self.key()}"]
        self.material.use_fake_user = True

    def add_material(self, config: dict, _: Dict[str, bpy.types.Image]) -> bpy.types.Material:
        object_mat = self.material.copy()
        object_mat.name = f"HD2 {self.key()} " + config["name"]

        print("    Applying settings")
        fence_group = object_mat.node_tree.nodes['Group']
        for name, setting in config["extras"].items():
            if name not in fence_group.inputs:
                continue
            if len(setting) == 1:
                fence_group.inputs[name].default_value = setting[0]
                continue
            fence_group.inputs[name].default_value = setting[:3]

        print("    Finalizing material")

        object_mat["needsBakeUVs"] = False
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "texture_map_319d3bb5" in config["extras"] and "fence_offset" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Fence"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None