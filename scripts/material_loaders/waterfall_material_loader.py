# pyright: basic
from pathlib import Path
from typing import Dict, Optional  # noqa: UP035

import bpy
from bpy.types import (  # pyright: ignore[reportMissingModuleSource]
    BlendData,
    Material,
    ShaderNodeTexImage,
)

from .filediver_material_loader_interface import FilediverMaterialLoaderInterface


class WaterfallMaterialLoader(FilediverMaterialLoaderInterface):
    material: Material

    def load_material(self, resource_path: Path) -> None:
        if f"HD2 {self.key()}" not in bpy.data.materials:
            with bpy.data.libraries.load(  # pyright: ignore[reportGeneralTypeIssues]
                str(resource_path / "Helldivers2 Shader v1.0.5.blend")
            ) as (shader_blend, our_blend):
                our_blend: BlendData  # not actually but they share member names
                shader_blend: BlendData
                our_blend.materials = shader_blend.materials  # pyright: ignore[reportAttributeAccessIssue]
        self.material = bpy.data.materials[f"HD2 {self.key()}"]
        self.material.use_fake_user = True

    def add_material(
        self,
        config: dict,
        textures: Dict[str, bpy.types.Image],  # noqa: UP006
    ) -> Material:
        object_mat = self.material.copy()
        object_mat.name = f"HD2 {self.key()} " + config["name"]

        print("    Applying textures")
        assert object_mat.node_tree is not None
        config_nodes: dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes  # pyright: ignore[reportAssignmentType]
        for usage, image in textures.items():
            assert image.colorspace_settings is not None
            match usage:
                case "distortion_tex":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"  # pyright: ignore[reportAttributeAccessIssue]
                case "water_pack_1":
                    config_nodes["Image Texture.002"].image = image
                    image.colorspace_settings.name = "Non-Color"  # pyright: ignore[reportAttributeAccessIssue]
                case "water_pack_2":
                    config_nodes["Image Texture.001"].image = image
                    image.colorspace_settings.name = "Non-Color"  # pyright: ignore[reportAttributeAccessIssue]

        print("    Applying settings")
        waterfall_group = object_mat.node_tree.nodes["Group.001"]
        uv_group = object_mat.node_tree.nodes["Group"]
        for name, setting in config["extras"].items():
            object_mat[name] = setting
            if name in waterfall_group.inputs:
                if len(setting) == 1:
                    waterfall_group.inputs[name].default_value = setting[0]
                    continue
                if name == "color" and len(setting) == 3:
                    setting += [1]
                waterfall_group.inputs[name].default_value = setting  # pyright: ignore[reportAttributeAccessIssue]
            elif name in uv_group:
                if len(setting) == 1:
                    uv_group.inputs[name].default_value = setting[0]
                    continue
                uv_group.inputs[name].default_value = setting

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
            "distortion_tex" in config["extras"]
            and "water_pack_1" in config["extras"]
            and "water_pack_2" in config["extras"]
        )

    @classmethod
    def key(cls) -> str:
        return "Waterfall"

    def get_material(self, config: dict, index: int) -> Optional[Material]:  # noqa: UP045
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None
