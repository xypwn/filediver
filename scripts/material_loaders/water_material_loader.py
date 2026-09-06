from pathlib import Path
from typing import Dict, Optional

import bpy
from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)

from .filediver_material_loader_interface import FilediverMaterialLoaderInterface


class WaterMaterialLoader(FilediverMaterialLoaderInterface):
    material: Material

    def load_material(self, resource_path: Path) -> None:
        if f"HD2 {self.key()}" not in bpy.data.materials:
            with bpy.data.libraries.load(
                str(resource_path / "Helldivers2 Shader v1.0.5.blend")
            ) as (shader_blend, our_blend):
                our_blend: BlendData  # not actually but they share member names
                shader_blend: BlendData
                our_blend.materials = shader_blend.materials
        self.material = bpy.data.materials[f"HD2 {self.key()}"]
        self.material.use_fake_user = True

    def add_material(
        self,
        config: dict,
        _: Dict[str, bpy.types.Image],
    ) -> Material:
        object_mat = self.material.copy()
        object_mat.name = f"HD2 {self.key()} " + config["name"]

        print("    Applying settings")
        water_group = object_mat.node_tree.nodes["Group"]
        for name, setting in config["extras"].items():
            object_mat[name] = setting
            if name not in water_group.inputs:
                continue
            water_group.inputs[name].default_value = setting[0]

        print("    Finalizing material")
        return object_mat

    def preprocess_config(self, data, gltf, materialTextures, config):
        _ = data
        _ = gltf
        _ = materialTextures
        return config

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return (
            "water_noise" in config["extras"]
            and "water_normal_map" in config["extras"]
            and "flow_map" in config["extras"]
        )

    @classmethod
    def key(cls) -> str:
        return "Water"

    def get_material(self, config: dict, index: int) -> Optional[Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None
