# pyright: basic
from pathlib import Path
from typing import Dict, Optional

import bpy
from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)

from .filediver_material_loader_interface import FilediverMaterialLoaderInterface


class ConcreteAltMaterialLoader(FilediverMaterialLoaderInterface):
    material: Material = None

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
        self, config: dict, textures: Dict[str, bpy.types.Image]
    ) -> Material:
        object_mat = self.material.copy()
        object_mat.name = f"HD2 {self.key()} " + config["name"]

        print("    Applying textures")
        assert object_mat.node_tree is not None
        config_nodes: dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
        for usage, image in textures.items():
            assert image.colorspace_settings is not None
            match usage:
                case "concrete_sampler":
                    config_nodes["Image Texture.001"].image = image
                    config_nodes["Image Texture.002"].image = image
                    config_nodes["Image Texture.003"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "texture_map_319d3bb5":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "trim_decal":
                    config_nodes["Image Texture.004"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"

        print("    Applying settings")
        concrete_group = object_mat.node_tree.nodes["Group"]
        uv_group = object_mat.node_tree.nodes["Triplanar UVs"]
        for name, setting in config["extras"].items():
            object_mat[name] = setting
            if name == "concrete_tiling":
                uv_group.inputs["tiling_factor"].default_value = setting[0]
                continue
            if name not in concrete_group.inputs:
                continue
            if len(setting) == 1:
                concrete_group.inputs[name].default_value = setting[0]
                continue
            concrete_group.inputs[name].default_value = setting

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
            "concrete_sampler" in config["extras"]
            and "texture_map_319d3bb5" in config["extras"]
        )

    @classmethod
    def key(cls) -> str:
        return "Concrete Alt"

    def get_material(self, config: dict, index: int) -> Optional[Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None
