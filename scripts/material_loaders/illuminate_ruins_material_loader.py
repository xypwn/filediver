from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeTexImage,
)
import bpy

class IlluminateRuinsMaterialLoader(FilediverMaterialLoaderInterface):
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
        object_mat.name = f"HD2 IlRuins " + config["name"]

        print("    Applying textures")
        config_nodes: Dict[str, ShaderNodeTexImage] = object_mat.node_tree.nodes
        for usage, image in textures.items():
            match usage:
                case "noise_tiler_mask":
                    config_nodes["Image Texture"].image = image
                    config_nodes["Image Texture.001"].image = image
                    config_nodes["Image Texture.002"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "base_tiler_nan":
                    config_nodes["Image Texture.003"].image = image
                    config_nodes["Image Texture.004"].image = image
                    config_nodes["Image Texture.005"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "detail_trimsheet_metallic_ceramic_masking":
                    config_nodes["Image Texture.006"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "ceramic_detail_tiler_basecolor":
                    config_nodes["Image Texture.007"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "ceramic_detail_tiler_nar":
                    config_nodes["Image Texture.008"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "rock_detail_tiler_basecolor":
                    config_nodes["Image Texture.009"].image = image
                    image.colorspace_settings.name = "sRGB"
                    image.alpha_mode = "CHANNEL_PACKED"
                case "rock_detail_tiler_nar":
                    config_nodes["Image Texture.010"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "detail_trimsheet_nar":
                    config_nodes["Image Texture.011"].image = image
                    image.colorspace_settings.name = "Non-Color"
                case "metallic_lut":
                    config_nodes["Image Texture.012"].image = image
                    image.colorspace_settings.name = "Non-Color"
                

        print("    Applying settings")
        group = config_nodes['Group']
        for name, setting in config["extras"].items():
            if name not in group.inputs:
                match name:
                    case "noise_mask_tiling":
                        config_nodes["Value"].outputs[0].default_value = setting[0]
                    case "base_nar_tiling":
                        config_nodes["Value.001"].outputs[0].default_value = setting[0]
                    case "ceramic_detail_tiling":
                        config_nodes["Value.002"].outputs[0].default_value = setting[0]
                    case "rock_detail_tiling":
                        config_nodes["Value.003"].outputs[0].default_value = setting[0]
                continue
            if len(setting) == 1:
                group.inputs[name].default_value = setting[0]
                continue
            group.inputs[name].default_value = setting[:3]

        print("    Finalizing material")

        object_mat["needsBakeUVs"] = False
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "noise_tiler_mask" in config["extras"] and "noise_mask_tiling" in config["extras"] and "detail_trimsheet_metallic_ceramic_masking" in config["extras"] and "ceramic_detail_tiler_basecolor" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Illuminate Ruins"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 IlRuins " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 IlRuins " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None