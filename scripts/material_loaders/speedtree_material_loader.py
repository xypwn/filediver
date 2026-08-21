from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
    ShaderNodeGroup,
)
import bpy

class SpeedtreeMaterialLoader(FilediverMaterialLoaderInterface):
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
        config_nodes: Dict[str, ShaderNodeTexImage|ShaderNodeGroup] = object_mat.node_tree.nodes
        speedtree_color_grading_nodes: Dict[str, ShaderNodeGroup] = config_nodes["Speedtree Color Grading"].node_tree.nodes
        color_grading: Dict[str, ShaderNodeGroup] = speedtree_color_grading_nodes["Group.002"].node_tree.nodes
        asset_color_grading_lut_group: Dict[str, ShaderNodeTexImage] = color_grading["asset_color_grading_lut"].node_tree.nodes

        for usage, image in textures.items():
            match usage:
                case "tex0":
                    config_nodes["Image Texture"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "tex1":
                    config_nodes["Image Texture.002"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "tex2":
                    config_nodes["Image Texture.001"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "asset_color_grading_lut":
                    asset_color_grading_lut_group["Image Texture"].image = image
                    image.colorspace_settings.name = "Non-Color"
                    asset_color_grading_lut_group["Image Texture"].interpolation = "Closest"
                case "emissive_texture":
                    config_nodes["Image Texture.003"].image = image
                    image.alpha_mode = "CHANNEL_PACKED"
        print("    Finalizing material")

        grading_group_id = 0.0
        has_secondworld = "grading_group_id_secondworld" in config["extras"]
        for name, setting in config["extras"].items():
            object_mat[name] = setting
            if name in config_nodes["Speedtree Color Grading"].inputs:
                if name == "grading_group_id":
                    grading_group_id = setting[0]
                config_nodes["Speedtree Color Grading"].inputs[name].default_value = setting[0]
            elif name in config_nodes["Speedtree Emissive"].inputs:
                if name == "emissive_texture":
                    continue
                if name == "emissive_color":
                    config_nodes["Speedtree Emissive"].inputs[name].default_value = setting[:3]
                elif name == "emissive_roughness_range":
                    config_nodes["Speedtree Emissive"].inputs[name].default_value = setting[:2]
                else:
                    config_nodes["Speedtree Emissive"].inputs[name].default_value = setting[0]

        if not has_secondworld:
            config_nodes["Speedtree Color Grading"].inputs["grading_group_id_secondworld"].default_value = grading_group_id
            config_nodes["Speedtree Color Grading"].inputs["grading_group_id_thirdworld"].default_value = grading_group_id

        object_mat["needsBakeUVs"] = False
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "tex0" in config["extras"] and "asset_color_grading_lut" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Speedtree"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None