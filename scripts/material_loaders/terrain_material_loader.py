from .filediver_material_loader_interface import FilediverMaterialLoaderInterface

from typing import Dict, Optional

from bpy.types import (
    BlendData,
    Material,
    ShaderNodeGroup,
    ShaderNodeTexImage,
    ShaderNodeTree,
)
import bpy

class TerrainMaterialLoader(FilediverMaterialLoaderInterface):
    noise_compare: ShaderNodeTree = None
    noise_generator: ShaderNodeTree = None
    noise_uvs: ShaderNodeTree = None

    def load_material(self, resource_path: str) -> None:
        if "Material Noise Generator" not in bpy.data.node_groups or "Noise Compare" not in bpy.data.node_groups or "Noise UVs" not in bpy.data.node_groups:
            with bpy.data.libraries.load(str(resource_path / "Helldivers2 Shader v1.0.5.blend")) as (shader_blend, our_blend):
                our_blend: BlendData # not actually but they share member names 
                shader_blend: BlendData
                our_blend.node_groups = shader_blend.node_groups
        self.noise_compare = bpy.data.node_groups["Noise Compare"]
        self.noise_generator = bpy.data.node_groups["Material Noise Generator"]
        self.noise_uvs = bpy.data.node_groups["Noise UVs"]

    def add_material(self, config: dict, textures: Dict[str, bpy.types.Image]) -> bpy.types.Material:
        object_mat: Material = bpy.data.materials.new(f"HD2 {self.key()} " + config["name"])
        object_mat.use_nodes = True
        object_mat.node_tree.nodes.remove(object_mat.node_tree.nodes["Principled BSDF"])

        print("    Creating terrain noise")
        tree: ShaderNodeTree = object_mat.node_tree
        origin = (0.0, 0.0)
        for i, terrain_settings in enumerate(config["extras"]["fd_terrain_materials"]):
            terrain_settings: dict
            settings: dict = terrain_settings["settings"]
            location = (origin[0], origin[1] + i * 320)
            uvs: ShaderNodeGroup = tree.nodes.new("ShaderNodeGroup")
            uvs.node_tree = self.noise_uvs
            uvs.name = f"Noise UVs {i}"
            print(uvs.inputs.keys())
            print(settings.keys())
            uvs.inputs["noise_scale"].default_value = settings["noise_scale"][0]
            uvs.location = (location[0] + 160 * 0, location[1])

            noise_map: ShaderNodeTexImage = tree.nodes.new("ShaderNodeTexImage")
            noise_map.image = textures["texture_map_0b1b5dad"]
            noise_map.image.alpha_mode = 'CHANNEL_PACKED'
            noise_map.image.colorspace_settings.name = "Non-Color"
            noise_map.location = (location[0] + 160 * 1, location[1])

            noise_generator: ShaderNodeGroup = tree.nodes.new("ShaderNodeGroup")
            noise_generator.node_tree = self.noise_generator
            noise_generator.name = f"Noise Generator {i}"
            noise_generator.location = (location[0] + 160 * 2.66, location[1])
            has_gradient = False
            for name, value in settings.items():
                if name not in noise_generator.inputs:
                    continue
                if name == "height_limit_start" and not has_gradient:
                    noise_generator.inputs["use_height_limit"].default_value = 1
                if name == "height_gradient_scale" and not has_gradient:
                    has_gradient = True
                    noise_generator.inputs["use_height_limit"].default_value = 0
                    noise_generator.inputs["use_height_limit_with_gradient"].default_value = 1
                noise_generator.inputs[name].default_value = value[0]

            tree.links.new(uvs.outputs[0], noise_map.inputs[0])
            tree.links.new(noise_map.outputs["Color"], noise_generator.inputs["noise xyz"])
            tree.links.new(noise_map.outputs["Alpha"], noise_generator.inputs["noise w"])

        # print("    Applying textures")
        # config_nodes: Dict[str, ShaderNodeTexImage|ShaderNodeGroup] = object_mat.node_tree.nodes
        # asset_color_grading_lut_group: Dict[str, ShaderNodeTexImage] = bpy.data.node_groups["asset_color_grading_lut"].nodes

        # for usage, image in textures.items():
        #     match usage:
        #         case "albedo_blend_tex" | "albedo_tex":
        #             config_nodes["Image Texture"].image = image
        #             image.colorspace_settings.name = "sRGB"
        #             image.alpha_mode = "CHANNEL_PACKED"
        #         case "displacement_tex":
        #             config_nodes["Image Texture.001"].image = image
        #             image.colorspace_settings.name = "Non-Color"
        #         case "nar_tex":
        #             config_nodes["Image Texture.002"].image = image
        #             image.colorspace_settings.name = "Non-Color"
        #         case "asset_color_grading_lut":
        #             asset_color_grading_lut_group["Image Texture"].image = image
        #             image.colorspace_settings.name = "Non-Color"
        #             asset_color_grading_lut_group["Image Texture"].interpolation = "Closest"
        # print("    Finalizing material")

        # for name, setting in config["extras"].items():
        #     object_mat[name] = setting
        #     if name in config_nodes["Terrain Color Grading"].inputs:
        #         config_nodes["Terrain Color Grading"].inputs[name].default_value = setting[0]
        #     if name in config_nodes["Terrain Emissive"].inputs:
        #         if name in ["emissive_base_color_tint", "emissive_color_ao"]:
        #             config_nodes["Terrain Emissive"].inputs[name].default_value = setting[:3]
        #         elif name in ["emissive_base_color_luminance_range", "emissive_ao_range"]:
        #             config_nodes["Terrain Emissive"].inputs[name].default_value = setting[:2]
        #         else:
        #             config_nodes["Terrain Emissive"].inputs[name].default_value = setting[0]

        # object_mat["needsBakeUVs"] = False
        return object_mat

    @classmethod
    def can_configure(cls, config: dict) -> bool:
        return "texture_map_0b1b5dad" in config["extras"] and "fd_terrain_materials" in config["extras"]

    @classmethod
    def key(cls) -> str:
        return "Terrain"

    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        key = f"HD2 {self.key()} " + config["name"]
        i = 1
        while key in bpy.data.materials and bpy.data.materials[key]["gltfId"] != index:
            key = f"HD2 {self.key()} " + config["name"] + f".{i:03d}"
            i += 1
        if key in bpy.data.materials:
            return bpy.data.materials[key]
        return None