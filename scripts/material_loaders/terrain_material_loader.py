from .filediver_material_loader_interface import FilediverMaterialLoaderInterface
from .terrain_projector_material_loader import TerrainProjectorMaterialLoader
from .image_utils import get_textures

from typing import Dict, Optional, TypedDict, Tuple

from bpy.types import (
    BlendData,
    Image,
    Material,
    ShaderNodeGroup,
    ShaderNodeOutputMaterial,
    ShaderNodeTexImage,
    ShaderNodeTree,
    NodeGroupOutput,
)
import bpy

class Vector2(TypedDict):
    x: float
    y: float

class TerrainMaterialLoader(FilediverMaterialLoaderInterface):
    noise_compare: ShaderNodeTree = None
    noise_generator: ShaderNodeTree = None
    noise_uvs: ShaderNodeTree = None
    terrain_loader: TerrainProjectorMaterialLoader = TerrainProjectorMaterialLoader()

    def __configure_terrain_template(self):
        terrain_material = bpy.data.materials[f"HD2 {self.key()}"]
        terrain_node_group = bpy.data.node_groups.new("Base Terrain Template", "ShaderNodeTree")
        material_output: ShaderNodeOutputMaterial = terrain_material.node_tree.get_output_node("ALL")
        output_location = material_output.location
        terrain_node_group.interface.new_socket(name = "Output", in_out='OUTPUT', socket_type = 'NodeSocketShader')

        group_output: NodeGroupOutput = terrain_node_group.nodes.new("NodeGroupOutput")
        group_output.location = output_location

        for node in terrain_material.node_tree.nodes:
            if node.bl_idname == "ShaderNodeOutputMaterial":
                continue
            copied_node = terrain_node_group.nodes.new(node.bl_idname)
            copied_node.location = node.location
            if node.bl_idname == "ShaderNodeGroup":
                node: ShaderNodeGroup
                copied_node: ShaderNodeGroup
                copied_node.node_tree = node.node_tree
            copied_node.name = node.name
            copied_node.label = node.label
            copied_node.hide = node.hide

        for node in terrain_material.node_tree.nodes:
            if node.bl_idname == "ShaderNodeOutputMaterial":
                continue
            for socket in node.inputs:
                for link in socket.links:
                    socket_a = terrain_node_group.nodes[link.from_node.name].outputs[link.from_socket.identifier]
                    socket_b = terrain_node_group.nodes[link.to_node.name].inputs[link.to_socket.identifier]
                    terrain_node_group.links.new(socket_a, socket_b)

        terrain_node_group.links.new(terrain_node_group.nodes["Mix Shader"].outputs["Shader"], group_output.inputs["Output"])

    def load_material(self, resource_path: str) -> None:
        if "Material Noise Generator" not in bpy.data.node_groups or "Noise Compare" not in bpy.data.node_groups or "Noise UVs" not in bpy.data.node_groups:
            with bpy.data.libraries.load(str(resource_path / "Helldivers2 Shader v1.0.5.blend")) as (shader_blend, our_blend):
                our_blend: BlendData # not actually but they share member names 
                shader_blend: BlendData
                our_blend.node_groups = shader_blend.node_groups
                if f"HD2 {self.key()}" not in bpy.data.materials:
                    our_blend.materials = shader_blend.materials
        self.noise_compare = bpy.data.node_groups["Noise Compare"]
        self.noise_generator = bpy.data.node_groups["Material Noise Generator"]
        self.noise_uvs = bpy.data.node_groups["Noise UVs"]
        if "Base Terrain Template" not in bpy.data.node_groups:
            self.__configure_terrain_template()


    def __add_terrain_shader(self, tree: ShaderNodeTree, terrain_settings: dict, location: Vector2) -> ShaderNodeGroup:
        terrain_group: ShaderNodeGroup = tree.nodes.new("ShaderNodeGroup")
        terrain_group.node_tree = terrain_settings["node_group"]
        terrain_group.location = (location["x"], location["y"])
        return terrain_group

    def __add_noise(self, settings: dict, textures: Dict[str, bpy.types.Image], tree: ShaderNodeTree, location: Vector2) -> Tuple[ShaderNodeGroup, ShaderNodeTexImage, ShaderNodeGroup]:
        uvs: ShaderNodeGroup = tree.nodes.new("ShaderNodeGroup")
        uvs.node_tree = self.noise_uvs
        uvs.inputs["noise_scale"].default_value = settings["noise_scale"][0]
        uvs.location = (location["x"], location["y"])

        noise_map: ShaderNodeTexImage = tree.nodes.new("ShaderNodeTexImage")
        noise_map.image = textures["texture_map_0b1b5dad"]
        noise_map.image.alpha_mode = 'CHANNEL_PACKED'
        noise_map.image.colorspace_settings.name = "Non-Color"
        location["x"] = location["x"] + uvs.width + 20
        noise_map.location = (location["x"], location["y"])

        noise_generator: ShaderNodeGroup = tree.nodes.new("ShaderNodeGroup")
        noise_generator.node_tree = self.noise_generator
        location["x"] = location["x"] + noise_map.width + 20
        noise_generator.location = (location["x"], location["y"])

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

        return (uvs, noise_map, noise_generator)

    def add_material(self, config: dict, textures: Dict[str, bpy.types.Image]) -> bpy.types.Material:
        object_mat: Material = bpy.data.materials.new(f"HD2 {self.key()} " + config["name"])
        object_mat.use_nodes = True
        object_mat.node_tree.nodes.remove(object_mat.node_tree.nodes["Principled BSDF"])

        print("    Creating terrain noise")
        tree: ShaderNodeTree = object_mat.node_tree
        origin: Vector2 = Vector2(x = 0.0, y = 0.0)
        vertical_offset = 640.0
        for i, terrain_settings in enumerate(config["extras"]["fd_terrain_materials"]):
            terrain_settings: dict
            settings: dict = terrain_settings["settings"]
            location: Vector2 = Vector2(x = origin["x"], y = origin["y"] + vertical_offset * i)

            uvs, noise_map, noise_generator = self.__add_noise(settings, textures, tree, location)
            uvs.name = f"Noise UVs {i}"
            noise_generator.name = f"Noise Generator {i}"

            shader_location = Vector2(x = location["x"], y = location["y"] + noise_generator.height)
            terrain_shader = self.__add_terrain_shader(tree, terrain_settings, shader_location)

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

    def preprocess_config(self, data: BlendData, gltf: dict, materialTextures: Dict[int, Dict[str, Image]], config: dict):
        print("    Preprocessing terrain materials")

        for i in range(len(config["extras"]["fd_terrain_materials"])):
            idx = config["extras"]["fd_terrain_materials"][i]["index"]
            material = gltf["materials"][idx]
            textures = get_textures(data, gltf, materialTextures, idx, material, True)
            terrain_name = f"Terrain {config['extras']['fd_terrain_materials'][i]['settings']['material'][0]}"
            if terrain_name not in bpy.data.node_groups:
                terrain_node_group: ShaderNodeTree = bpy.data.node_groups["Base Terrain Template"].copy()
                terrain_node_group.name = terrain_name
                asset_color_grading_lut_group: Dict[str, ShaderNodeTexImage] = bpy.data.node_groups["asset_color_grading_lut"].nodes
                config_nodes: Dict[str, ShaderNodeTexImage|ShaderNodeGroup] = terrain_node_group.nodes
                self.terrain_loader.configure_terrain_group(config, textures, config_nodes, asset_color_grading_lut_group)
            else:
                terrain_node_group = bpy.data.node_groups[terrain_name]
            
            config["extras"]["fd_terrain_materials"][i]["node_group"] = terrain_node_group

        return config

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