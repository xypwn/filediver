# Original shader created by @thejudsub on The Helldivers Archive Discord Server
# Shader post can be found here: https://discord.com/channels/1210541115829260328/1222290154409033889

# Modified slightly for use with filediver
# Modified quite a bit to add cape support

# Shader bundled with filediver with permission from @thejudsub

import bpy
from bpy.types import (
    NodeTree,
    Material,
    Object,
    ShaderNodeGroup,
    ShaderNodeMix,
    ShaderNodeMath,
    ShaderNodeTexImage,
    ShaderNodeVectorMath,
    ShaderNodeUVMap,
    ShaderNodeSeparateColor,
    ShaderNodeSeparateXYZ,
    ShaderNodeCombineXYZ,
    ShaderNodeValue,
    ShaderNodeNewGeometry,
    ShaderNodeClamp,
    ShaderNodeBsdfTransparent,
    NodeGroupInput,
    NodeGroupOutput,
    NodeSocketFloat,
)
from mathutils import Vector
from typing import Optional, List
import os
import math

ScriptVersion = "_1.0.6c"

class NODE_PT_MAINPANEL(bpy.types.Panel):
    bl_label = "Helldivers 2 Cape Shader"
    bl_idname = "NODE_PT_capecontrols"
    bl_space_type = 'NODE_EDITOR'
    bl_region_type = 'UI'
    bl_category = 'HD2 Shader'

    def draw(self, context):
        layout = self.layout

        col = layout.column()
        col.operator('node.create_operator')
        col.operator('node.update_operator')

def get_current_material_name():
    obj = bpy.context.active_object
    if obj is not None:
        if obj.material_slots:
            return obj.material_slots[0].name
    return None

def create_HD2_Shader(context, operator, group_name, material: Optional[Material] = None, obj: Optional[Object] = None):
    bpy.context.scene.use_nodes = True
    
    if material is None:
        material = bpy.context.active_object.active_material
    if obj is None:
        obj = bpy.context.active_object

    HD2_Shader: NodeTree = bpy.data.node_groups.new(group_name, 'ShaderNodeTree')
    
    current_material_name = get_current_material_name()
    
    #automatically set the material to use the newly generated shader preset, reposition it, and remove copies
    hd2_shader_group_node: ShaderNodeGroup = material.node_tree.nodes["HD2 Shader Template"]
    hd2_shader_group_node.node_tree = HD2_Shader

    HD2_Shader.name = current_material_name + ScriptVersion

    output_socket = HD2_Shader.interface.new_socket(name = "Output", in_out='OUTPUT', socket_type = 'NodeSocketShader')
    #ID Mask Array UV Fix
    id_mask_array_uv_fix = HD2_Shader.interface.new_socket(name = "ID Mask Array UV Fix", in_out='INPUT', socket_type = 'NodeSocketColor')
    id_mask_array_uv_fix.hide_value = True
    #Pattern Mask Array UV Fix
    pattern_mask_array_uv_fix = HD2_Shader.interface.new_socket(name = "Pattern Mask Array UV Fix", in_out='INPUT', socket_type = 'NodeSocketColor')
    pattern_mask_array_uv_fix.hide_value = True
    #Socket ID Mask Array
    id_mask_array_socket = HD2_Shader.interface.new_socket(name = "ID Mask Array", in_out='INPUT', socket_type = 'NodeSocketColor')
    id_mask_array_socket.hide_value = True
    #Socket ID Mask Array(alpha)
    id_mask_array_alpha__socket = HD2_Shader.interface.new_socket(name = "ID Mask Array(alpha)", in_out='INPUT', socket_type = 'NodeSocketFloat')
    id_mask_array_alpha__socket.default_value = 0.0
    id_mask_array_alpha__socket.hide_value = True
    #Socket Pattern Mask Array
    pattern_mask_array_socket = HD2_Shader.interface.new_socket(name = "Pattern Mask Array", in_out='INPUT', socket_type = 'NodeSocketColor')
    pattern_mask_array_socket.default_value = (0.0, 0.0, 0.0, 0.0)
    pattern_mask_array_socket.hide_value = True
    #Socket Decal Texture
    decal_texture_socket = HD2_Shader.interface.new_socket(name = "Decal Texture", in_out='INPUT', socket_type = 'NodeSocketColor')
    decal_texture_socket.hide_value = True
    #Socket Decal Texture(alpha)
    decal_texture_alpha__socket = HD2_Shader.interface.new_socket(name = "Decal Texture(alpha)", in_out='INPUT', socket_type = 'NodeSocketFloat')
    decal_texture_alpha__socket.default_value = 0.0
    decal_texture_alpha__socket.hide_value = True
    #Socket Primary Material LUT
    primary_material_lut_socket = HD2_Shader.interface.new_socket(name = "Primary Material LUT", in_out='INPUT', socket_type = 'NodeSocketColor')
    primary_material_lut_socket.hide_value = True
    #Socket Cape LUT
    cape_lut_socket = HD2_Shader.interface.new_socket(name = "Cape LUT", in_out='INPUT', socket_type = 'NodeSocketColor')
    cape_lut_socket.hide_value = True
    #Socket Pattern LUT
    pattern_lut_socket = HD2_Shader.interface.new_socket(name = "Pattern LUT", in_out='INPUT', socket_type = 'NodeSocketColor')
    pattern_lut_socket.hide_value = True
    #Socket Normal Map
    normal_map_socket = HD2_Shader.interface.new_socket(name = "Normal Map", in_out='INPUT', socket_type = 'NodeSocketVector')
    normal_map_socket.default_value = (0.5, 0.5, 1.0)
    normal_map_socket.hide_value = True
    #Socket Normal Map_Alpha
    normal_map_alpha_socket = HD2_Shader.interface.new_socket(name = "Normal Map_Alpha", in_out='INPUT', socket_type = 'NodeSocketFloat')
    normal_map_alpha_socket.default_value = 0.0
    normal_map_alpha_socket.hide_value = True
    #Socket detail_tile_factor_mult
    detail_tile_factor_mult_socket = HD2_Shader.interface.new_socket(name = "detail_tile_factor_mult", in_out='INPUT', socket_type = 'NodeSocketFloat')
    detail_tile_factor_mult_socket.default_value = 1.000
    #Panel Bake_Outputs
    bake_outputs_panel = HD2_Shader.interface.new_panel("Bake_Outputs")
    #Socket Color
    color_socket = HD2_Shader.interface.new_socket(name = "Bake_Color", in_out='OUTPUT', socket_type = 'NodeSocketColor', parent = bake_outputs_panel)
    #Socket Ambient Occlusion
    ambient_occlusion_socket = HD2_Shader.interface.new_socket(name = "Bake_Ambient Occlusion", in_out='OUTPUT', socket_type = 'NodeSocketFloat', parent = bake_outputs_panel)
    #Socket Metallic
    metallic_socket = HD2_Shader.interface.new_socket(name = "Bake_Metallic", in_out='OUTPUT', socket_type = 'NodeSocketFloat', parent = bake_outputs_panel)
    #Socket Roughness
    roughness_socket = HD2_Shader.interface.new_socket(name = "Bake_Roughness", in_out='OUTPUT', socket_type = 'NodeSocketFloat', parent = bake_outputs_panel)
    #Socket Normal
    normal_socket = HD2_Shader.interface.new_socket(name = "Bake_Normal", in_out='OUTPUT', socket_type = 'NodeSocketVector', parent = bake_outputs_panel)
    #Socket Clearcoat Normal
    clearcoat_normal_socket = HD2_Shader.interface.new_socket(name = "Bake_Clearcoat Normal", in_out='OUTPUT', socket_type = 'NodeSocketVector', parent = bake_outputs_panel)
    #Socket Alpha
    alpha_socket = HD2_Shader.interface.new_socket(name = "Bake_Alpha", in_out='OUTPUT', socket_type = 'NodeSocketFloat', parent = bake_outputs_panel)

    cape_decal_tree = create_cape_decal_template()
    
#sets the detail texture tile size to 1.000 by default
    hd2_shader_group_node.inputs[12].default_value = (1.000)

    print("################################"+current_material_name+ScriptVersion+"################################")
    
    try:
        add_bake_uvs(obj)
    except:
        pass

#delete warning frames        
    try:       
        for DeleteFrame01 in material.node_tree.nodes:
            if DeleteFrame01.label == "WARNING!Current texture must be a PNG file. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['ID Mask Array Filetype Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame02 in material.node_tree.nodes:
            if DeleteFrame02.label == "WARNING!Current texture must be a PNG file. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Pattern Mask Array Filetype Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame03 in material.node_tree.nodes:
            if DeleteFrame03.label == "WARNING!Current texture must be a PNG file. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Decal Texture Filetype Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame04 in material.node_tree.nodes:
            if DeleteFrame04.label == "WARNING!Current texture isnt a Material LUT. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Primary Material LUT Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame05 in material.node_tree.nodes:
            if DeleteFrame05.label == "WARNING!Current texture must be an EXR file. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Primary Material LUT Filetype Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame06 in material.node_tree.nodes:
            if DeleteFrame06.label == "WARNING!Current texture isnt a Cape LUT. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Cape LUT Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame07 in material.node_tree.nodes:
            if DeleteFrame07.label == "WARNING!Current texture must be an EXR file. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Cape LUT Filetype Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame08 in material.node_tree.nodes:
            if DeleteFrame08.label == "WARNING!Current texture isnt a Pattern LUT. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Pattern LUT Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame09 in material.node_tree.nodes:
            if DeleteFrame09.label == "WARNING!Current texture must be an EXR file. Errors may occur->":
                material.node_tree.nodes.remove(material.node_tree.nodes['Pattern LUT Filetype Warning'])
    except:
        pass
        
    try:       
        for DeleteFrame10 in material.node_tree.nodes:
            if DeleteFrame10.label == "WARNING!Update your nodes to 1.0.5. Errors may occur otherwise":
                material.node_tree.nodes.remove(material.node_tree.nodes['Shader Version Warning1'])
    except:
        pass
        
    try:       
        for DeleteFrame11 in material.node_tree.nodes:
            if DeleteFrame11.label == "Download the new version from the link below":
                material.node_tree.nodes.remove(material.node_tree.nodes['Shader Version Warning2'])
    except:
        pass
        
    try:       
        for DeleteFrame12 in material.node_tree.nodes:
            if DeleteFrame12.label == "https://tinyurl.com/Helldivers2Shader":
                material.node_tree.nodes.remove(material.node_tree.nodes['Shader Version Warning3'])
    except:
        pass
    
    try:       
        for DeleteFrame13 in material.node_tree.nodes:
            if DeleteFrame13.label == "WARNING!The shader only works for Blender 4.0+":
                material.node_tree.nodes.remove(material.node_tree.nodes['Shader Version Warning0'])
    except:
        pass
    
    update_array_uvs(material)
#reconnect nodes if they get undone
    try:
        for IDMaskNoncolor in material.node_tree.nodes:
            if IDMaskNoncolor.name == "Reroute.001":
                material.node_tree.links.new(IDMaskNoncolor.outputs[0], hd2_shader_group_node.inputs[2])
                material.node_tree.links.new(material.node_tree.nodes['Reroute.003'].outputs[0], hd2_shader_group_node.inputs[3])
                if material.node_tree.nodes['ID Mask Array Texture'].type == 'TEX_IMAGE' and material.node_tree.nodes['ID Mask Array Texture'].image:
                    material.node_tree.nodes['ID Mask Array Texture'].image.colorspace_settings.name = "Non-Color"
    except:
        pass
        
    try:
        for ArrayUVOverride in material.node_tree.nodes:
            if ArrayUVOverride.name == "Reroute.034":
                material.node_tree.links.new(ArrayUVOverride.outputs[0], hd2_shader_group_node.inputs[0])
                material.node_tree.links.new(material.node_tree.nodes['Reroute.035'].outputs[0], hd2_shader_group_node.inputs[1])
    except:
        pass

    try:
        for PatternMaskNoncolor in material.node_tree.nodes:
            if PatternMaskNoncolor.name == "Reroute.005":
                material.node_tree.links.new(PatternMaskNoncolor.outputs[0], hd2_shader_group_node.inputs[4])
                if material.node_tree.nodes['Pattern Mask Array'].type == 'TEX_IMAGE' and material.node_tree.nodes['Pattern Mask Array'].image:
                    material.node_tree.nodes['Pattern Mask Array'].image.colorspace_settings.name = "Non-Color"
    except:
        pass
        
    try:
        for DecalTexsRGB in material.node_tree.nodes:
            if DecalTexsRGB.name == "Reroute.007":
                material.node_tree.links.new(DecalTexsRGB.outputs[0], hd2_shader_group_node.inputs[5])
                material.node_tree.links.new(material.node_tree.nodes['Reroute.009'].outputs[0], hd2_shader_group_node.inputs[6])
                if material.node_tree.nodes['Decal Texture'].type == 'TEX_IMAGE' and material.node_tree.nodes['Decal Texture'].image:
                    material.node_tree.nodes['Decal Texture'].image.colorspace_settings.name = "sRGB"
                    material.node_tree.nodes['Decal Texture'].image.alpha_mode = "CHANNEL_PACKED"
    except:
        pass
        
    try:
        for PrimaryMaterialLUTLinear in material.node_tree.nodes:
            if PrimaryMaterialLUTLinear.name == "Reroute.013":
                material.node_tree.links.new(PrimaryMaterialLUTLinear.outputs[0], hd2_shader_group_node.inputs[7])
                if material.node_tree.nodes['Primary Material LUT Texture'].type == 'TEX_IMAGE' and material.node_tree.nodes['Primary Material LUT Texture'].image:
                    material.node_tree.nodes['Primary Material LUT Texture'].image.colorspace_settings.name = "Non-Color"
                    material.node_tree.nodes['Primary Material LUT Texture'].image.alpha_mode = "CHANNEL_PACKED"
    except:
        pass
        
    try:
        CapeLUTLinear = material.node_tree.nodes["Reroute.015"]
        material.node_tree.links.new(CapeLUTLinear.outputs[0], hd2_shader_group_node.inputs[8])
        if material.node_tree.nodes['Cape LUT Texture'].type == 'TEX_IMAGE' and material.node_tree.nodes['Cape LUT Texture'].image:
            material.node_tree.nodes['Cape LUT Texture'].image.colorspace_settings.name = "Non-Color"
            material.node_tree.nodes['Cape LUT Texture'].image.alpha_mode = "CHANNEL_PACKED"
    except:
        pass
    
    try:
        for PatternLUTLinear in material.node_tree.nodes:
            if PatternLUTLinear.name == "Reroute.019":
                material.node_tree.links.new(PatternLUTLinear.outputs[0], hd2_shader_group_node.inputs[9])
                if material.node_tree.nodes['Pattern LUT Texture'].type == 'TEX_IMAGE' and material.node_tree.nodes['Pattern LUT Texture'].image:
                    material.node_tree.nodes['Pattern LUT Texture'].image.colorspace_settings.name = "Non-Color"
                    material.node_tree.nodes['Pattern LUT Texture'].alpha_mode = "CHANNEL_PACKED"
    except:
        pass
        
    try:
        for NormalMapNoncolor in material.node_tree.nodes:
            if NormalMapNoncolor.name == "Reroute.020":
                material.node_tree.links.new(NormalMapNoncolor.outputs[0], hd2_shader_group_node.inputs[10])
                material.node_tree.links.new(material.node_tree.nodes['Reroute.021'].outputs[0], hd2_shader_group_node.inputs[11])
                if material.node_tree.nodes['Normal Map'].type == 'TEX_IMAGE' and material.node_tree.nodes['Normal Map'].image:
                    material.node_tree.nodes['Normal Map'].image.colorspace_settings.name = "Non-Color"
    except:
        pass
        
    try:
        for ReconnectToMaterialOutput in material.node_tree.nodes:
            if ReconnectToMaterialOutput.name == "HD2 Shader Template":
                material.node_tree.links.new(ReconnectToMaterialOutput.outputs[0], material.node_tree.nodes['Material Output'].inputs[0])
    except:
        pass

#create warnings for user
    #get filetype of node
    try:
        IDMaskArrayFiletype = material.node_tree.nodes['ID Mask Array Texture'].image.file_format
    except:
        pass
        
    IDMaskArrayLocation = material.node_tree.nodes['ID Mask Array Texture'].location
    try:
        if not IDMaskArrayFiletype == "PNG":
            #Tell user that the texture isnt a proper Material LUT
            IDMaskArrayFiletypeWarning = material.node_tree.nodes.new("NodeFrame")
            IDMaskArrayFiletypeWarning.label = "WARNING!Current texture must be a PNG file. Errors may occur->"
            IDMaskArrayFiletypeWarning.name = "ID Mask Array Filetype Warning"
            IDMaskArrayFiletypeWarning.use_custom_color = True
            IDMaskArrayFiletypeWarning.color = (1,0,0)
            IDMaskArrayFiletypeWarning.label_size = 20
            IDMaskArrayFiletypeWarning.shrink = True
            IDMaskArrayFiletypeWarning.location = (IDMaskArrayLocation.x - 660, IDMaskArrayLocation.y - 100)
            IDMaskArrayFiletypeWarning.height, IDMaskArrayFiletypeWarning.width = 50,650
    except:
        pass
    
    try:
        IDMaskArraySizeX = material.node_tree.nodes['ID Mask Array Texture'].inputs[0].node.image.size[0]
        IDMaskArraySizeY = material.node_tree.nodes['ID Mask Array Texture'].inputs[0].node.image.size[1]
        if (IDMaskArraySizeY/IDMaskArraySizeX) >= 2.0:
            material.node_tree.nodes['ID Mask UV'].inputs[0].default_value = (1.000)
    except:
        pass

    try:
        material.node_tree.nodes['ID Mask Array Texture'].label = ("ID Mask Array Texture "+"("+str(IDMaskArraySizeX)+"x"+str(IDMaskArraySizeY)+")"+" "+IDMaskArrayFiletype)
    except:
        material.node_tree.nodes['ID Mask Array Texture'].label = ("ID Mask Array Texture")
        
    try:
        PatternMaskArrayFiletype = material.node_tree.nodes['Pattern Mask Array'].image.file_format
    except:
        pass
        
    PatternMaskArrayLocation = material.node_tree.nodes['Pattern Mask Array'].location
    try:
        if not PatternMaskArrayFiletype == "PNG":
            #Tell user that the texture isnt a proper Pattern Mask
            PatternMaskArrayFiletypeWarning = material.node_tree.nodes.new("NodeFrame")
            PatternMaskArrayFiletypeWarning.label = "WARNING!Current texture must be a PNG file. Errors may occur->"
            PatternMaskArrayFiletypeWarning.name = "Pattern Mask Array Filetype Warning"
            PatternMaskArrayFiletypeWarning.use_custom_color = True
            PatternMaskArrayFiletypeWarning.color = (1,0,0)
            PatternMaskArrayFiletypeWarning.label_size = 20
            PatternMaskArrayFiletypeWarning.shrink = True
            PatternMaskArrayFiletypeWarning.location = (PatternMaskArrayLocation.x - 660, PatternMaskArrayLocation.y - 100)
            PatternMaskArrayFiletypeWarning.height, PatternMaskArrayFiletypeWarning.width = 50,650
    except:
        pass
        
    try:
        PatternMaskSizeX = material.node_tree.nodes['Pattern Mask Array'].inputs[0].node.image.size[0]
        PatternMaskSizeY = material.node_tree.nodes['Pattern Mask Array'].inputs[0].node.image.size[1]
        material.node_tree.nodes['Pattern Mask Array'].label = ("Pattern Mask Array "+"("+str(PatternMaskSizeX)+"x"+str(PatternMaskSizeY)+")"+" "+PatternMaskArrayFiletype)
    except:
        material.node_tree.nodes['Pattern Mask Array'].label = ("Pattern Mask Array")
        
    try:
        print("Pattern Mask = "+PatternMaskArrayFiletype,(PatternMaskSizeX,PatternMaskSizeY))
    except:
        print("No Pattern Mask Texture detected")
        
    try:
        DecalTexSizeX = material.node_tree.nodes['Decal Texture'].inputs[0].node.image.size[0]
        DecalTexSizeY = material.node_tree.nodes['Decal Texture'].inputs[0].node.image.size[1]
    except:
        pass
        
    try:
        DecalTextureFiletype = material.node_tree.nodes['Decal Texture'].image.file_format
        print("Decal Texture = "+DecalTextureFiletype,(DecalTexSizeX,DecalTexSizeY))
    except:
        print("No Decal Texture detected")
        
    try:
        material.node_tree.nodes['Decal Texture'].label = ("Decal Texture "+"("+str(DecalTexSizeX)+"x"+str(DecalTexSizeY)+")"+" "+DecalTextureFiletype)
    except:
        material.node_tree.nodes['Decal Texture'].label = ("Decal Texture")
        
    DecalTextureLocation = material.node_tree.nodes['Decal Texture'].location
    try:
        if DecalTextureFiletype not in ["PNG", "OPEN_EXR"]:
            #Tell user that the texture isnt a proper Decal texture
            DecalTextureFiletypeWarning = material.node_tree.nodes.new("NodeFrame")
            DecalTextureFiletypeWarning.label = "WARNING!Current texture must be a PNG file. Errors may occur->"
            DecalTextureFiletypeWarning.name = "Decal Texture Filetype Warning"
            DecalTextureFiletypeWarning.use_custom_color = True
            DecalTextureFiletypeWarning.color = (1,0,0)
            DecalTextureFiletypeWarning.label_size = 20
            DecalTextureFiletypeWarning.shrink = True
            DecalTextureFiletypeWarning.location = (DecalTextureLocation.x - 660, DecalTextureLocation.y - 100)
            DecalTextureFiletypeWarning.height, DecalTextureFiletypeWarning.width = 50,650
    except:
        pass
        
    #get pixel height of Primary Material LUT
    try:
        PrimaryMaterialLUTSizeX = material.node_tree.nodes['Primary Material LUT Texture'].inputs[0].node.image.size[0]
        PrimaryMaterialLUTSizeY = material.node_tree.nodes['Primary Material LUT Texture'].inputs[0].node.image.size[1]
    except:
        pass

    #get location of node        
    PrimaryMatLUTLocation = material.node_tree.nodes['Primary Material LUT Texture'].location
    try:
        if not PrimaryMaterialLUTSizeY >= 2 and PrimaryMaterialLUTSizeX == 23:
            #Tell user that the texture isnt a proper Material LUT
            PrimaryMaterialLUTWarning = material.node_tree.nodes.new("NodeFrame")
            PrimaryMaterialLUTWarning.label = "WARNING!Current texture isnt a Material LUT. Errors may occur->"
            PrimaryMaterialLUTWarning.name = "Primary Material LUT Warning"
            PrimaryMaterialLUTWarning.use_custom_color = True
            PrimaryMaterialLUTWarning.color = (1,0,0)
            PrimaryMaterialLUTWarning.label_size = 20
            PrimaryMaterialLUTWarning.shrink = True
            PrimaryMaterialLUTWarning.location = (PrimaryMatLUTLocation.x - 660, PrimaryMatLUTLocation.y - 50)
            PrimaryMaterialLUTWarning.height, PrimaryMaterialLUTWarning.width = 50,650
    except:
        pass
    
    try:
        PrimaryMaterialLUTFiletype = material.node_tree.nodes['Primary Material LUT Texture'].image.file_format
        print("Primary Material LUT = "+PrimaryMaterialLUTFiletype,(PrimaryMaterialLUTSizeX,PrimaryMaterialLUTSizeY))
    except:
        print("No Primary Material LUT detected")
        
    try:
        if not PrimaryMaterialLUTFiletype == "OPEN_EXR":
            #Tell user that the texture isnt a proper Material LUT
            PrimaryMaterialLUTFiletypeWarning = material.node_tree.nodes.new("NodeFrame")
            PrimaryMaterialLUTFiletypeWarning.label = "WARNING!Current texture must be an EXR file. Errors may occur->"
            PrimaryMaterialLUTFiletypeWarning.name = "Primary Material LUT Filetype Warning"
            PrimaryMaterialLUTFiletypeWarning.use_custom_color = True
            PrimaryMaterialLUTFiletypeWarning.color = (1,0,0)
            PrimaryMaterialLUTFiletypeWarning.label_size = 20
            PrimaryMaterialLUTFiletypeWarning.shrink = True
            PrimaryMaterialLUTFiletypeWarning.location = (PrimaryMatLUTLocation.x - 660, PrimaryMatLUTLocation.y - 100)
            PrimaryMaterialLUTFiletypeWarning.height, PrimaryMaterialLUTFiletypeWarning.width = (50,650)
    except:
        pass
        
    try:
        material.node_tree.nodes['Primary Material LUT Texture'].label = ("Primary Material LUT Texture "+"("+str(PrimaryMaterialLUTSizeX)+"x"+str(PrimaryMaterialLUTSizeY)+")"+" "+PrimaryMaterialLUTFiletype)
    except:
        material.node_tree.nodes['Primary Material LUT Texture'].label = ("Primary Material LUT Texture")
        
    #get pixel height of Pattern LUT
    try:
        PatternLUTSizeX = material.node_tree.nodes['Pattern LUT Texture'].inputs[0].node.image.size[0]
        PatternLUTSizeY = material.node_tree.nodes['Pattern LUT Texture'].inputs[0].node.image.size[1]
    except:
        pass
        
    #get location of primary material lut node    
    PatternLUTLocation = material.node_tree.nodes['Pattern LUT Texture'].location
    try:
        if not PatternLUTSizeX == 3 and PatternLUTSizeY == 1:
                #Tell user that the texture isnt a proper Material LUT
                PatternLUTWarning = material.node_tree.nodes.new("NodeFrame")
                PatternLUTWarning.label = "WARNING!Current texture isnt a Pattern LUT. Errors may occur->"
                PatternLUTWarning.name = "Pattern LUT Warning"
                PatternLUTWarning.use_custom_color = True
                PatternLUTWarning.color = (1,0,0)
                PatternLUTWarning.label_size = 20
                PatternLUTWarning.shrink = True
                PatternLUTWarning.location = (PatternLUTLocation.x - 660, PatternLUTLocation.y - 100)
                PatternLUTWarning.height, PatternLUTWarning.width = (50,650)
    except:
        pass

    try:
        PatternLUTFiletype = material.node_tree.nodes['Pattern LUT Texture'].image.file_format
    except:
        pass
        
    try:
        if not PatternLUTFiletype == "OPEN_EXR":
            PatternLUTFiletypeWarning = material.node_tree.nodes.new("NodeFrame")
            PatternLUTFiletypeWarning.label = "WARNING!Current texture must be an EXR file. Errors may occur->"
            PatternLUTFiletypeWarning.name = "Pattern LUT Filetype Warning"
            PatternLUTFiletypeWarning.use_custom_color = True
            PatternLUTFiletypeWarning.color = (1,0,0)
            PatternLUTFiletypeWarning.label_size = 20
            PatternLUTFiletypeWarning.shrink = True
            PatternLUTFiletypeWarning.location = (PatternLUTLocation.x - 660, PatternLUTLocation.y - 50)
            PatternLUTFiletypeWarning.height, PatternLUTFiletypeWarning.width = (50,650)
    except:
        pass
        
    try:
        material.node_tree.nodes['Pattern LUT Texture'].label = ("Pattern LUT Texture "+"("+str(PatternLUTSizeX)+"x"+str(PatternLUTSizeY)+")"+" "+PatternLUTFiletype)
    except:
        material.node_tree.nodes['Pattern LUT Texture'].label = ("Pattern LUT Texture")
        
    try:
        CapeLUTSizeX = material.node_tree.nodes['Cape LUT Texture'].inputs[0].node.image.size[0]
        CapeLUTSizeY = material.node_tree.nodes['Cape LUT Texture'].inputs[0].node.image.size[1]
    except:
        pass
        
    #get location of primary material lut node    
    CapeLUTLocation = material.node_tree.nodes['Cape LUT Texture'].location
    try:
        if not CapeLUTSizeX == 16 and CapeLUTSizeY == 5:
                #Tell user that the texture isnt a proper Material LUT
                CapeLUTWarning = material.node_tree.nodes.new("NodeFrame")
                CapeLUTWarning.label = "WARNING!Current texture isnt a Cape LUT. Errors may occur->"
                CapeLUTWarning.name = "Cape LUT Warning"
                CapeLUTWarning.use_custom_color = True
                CapeLUTWarning.color = (1,0,0)
                CapeLUTWarning.label_size = 20
                CapeLUTWarning.shrink = True
                CapeLUTWarning.location = (CapeLUTLocation.x - 660, CapeLUTLocation.y - 100)
                CapeLUTWarning.height, CapeLUTWarning.width = (50,650)
    except:
        pass

    try:
        CapeLUTFiletype = material.node_tree.nodes['Cape LUT Texture'].image.file_format
        print("Cape LUT = "+CapeLUTFiletype,(CapeLUTSizeX,CapeLUTSizeY))
    except:
        print("No Cape LUT detected")
        
    try:
        if not CapeLUTFiletype == "OPEN_EXR":
            CapeLUTFiletypeWarning = material.node_tree.nodes.new("NodeFrame")
            CapeLUTFiletypeWarning.label = "WARNING!Current texture must be an EXR file. Errors may occur->"
            CapeLUTFiletypeWarning.name = "Cape LUT Filetype Warning"
            CapeLUTFiletypeWarning.use_custom_color = True
            CapeLUTFiletypeWarning.color = (1,0,0)
            CapeLUTFiletypeWarning.label_size = 20
            CapeLUTFiletypeWarning.shrink = True
            CapeLUTFiletypeWarning.location = (CapeLUTLocation.x - 660, CapeLUTLocation.y - 50)
            CapeLUTFiletypeWarning.height, CapeLUTFiletypeWarning.width = (50,650)
    except:
        pass
    
    try:
        print("Pattern LUT = "+PatternLUTFiletype,(PatternLUTSizeX,PatternLUTSizeY))
    except:
        print("No Pattern LUT detected")
            
    try:
        material.node_tree.nodes['Cape LUT Texture'].label = ("Cape LUT Texture "+"("+str(CapeLUTSizeX)+"x"+str(CapeLUTSizeY)+")"+" "+CapeLUTFiletype)
    except:
        material.node_tree.nodes['Cape LUT Texture'].label = ("Cape LUT Texture")
    
    try:
        NormalMapFiletype = material.node_tree.nodes['Normal Map'].image.file_format
    except:
        pass
    
    try:
        NormalSizeX = material.node_tree.nodes['Normal Map'].inputs[0].node.image.size[0]
        NormalSizeY = material.node_tree.nodes['Normal Map'].inputs[0].node.image.size[1]
        material.node_tree.nodes['Normal Map'].label = ("Normal Map "+"("+str(NormalSizeX)+"x"+str(NormalSizeY)+")"+" "+NormalMapFiletype)
    except:
        material.node_tree.nodes['Normal Map'].label = ("Normal Map")
        
    try:
        print("Normal Map = "+NormalMapFiletype,(NormalSizeX,NormalSizeY))
    except:
        print("No Normal Map Texture detected")
        
    ShaderLocation = hd2_shader_group_node.location
    
    try:
        if bpy.app.version < (4, 0, 0) :
            ShaderUpdateWarning0 = material.node_tree.nodes.new("NodeFrame")
            ShaderUpdateWarning0.label = "WARNING!The shader only works for Blender 4.0+"
            ShaderUpdateWarning0.name = "Shader Version Warning1"
            ShaderUpdateWarning0.use_custom_color = True
            ShaderUpdateWarning0.color = (1,0,0)
            ShaderUpdateWarning0.label_size = 20
            ShaderUpdateWarning0.shrink = True
            ShaderUpdateWarning0.location = (ShaderLocation.x - 200, ShaderLocation.y + 90)
            ShaderUpdateWarning0.height, ShaderUpdateWarning0.width = (50,650)
    except:
        pass
    
    try:
        if not hd2_shader_group_node.label == "HD2 Shader Template v1.0.5" and bpy.app.version > (4, 0, 0):
            ShaderUpdateWarning1 = material.node_tree.nodes.new("NodeFrame")
            ShaderUpdateWarning1.label = "WARNING!Update your nodes to 1.0.5. Errors may occur otherwise"
            ShaderUpdateWarning1.name = "Shader Version Warning1"
            ShaderUpdateWarning1.use_custom_color = True
            ShaderUpdateWarning1.color = (1,0,0)
            ShaderUpdateWarning1.label_size = 20
            ShaderUpdateWarning1.shrink = True
            ShaderUpdateWarning1.location = (ShaderLocation.x - 200, ShaderLocation.y + 200)
            ShaderUpdateWarning1.height, ShaderUpdateWarning1.width = (50,650)

            ShaderUpdateWarning2 = material.node_tree.nodes.new("NodeFrame")
            ShaderUpdateWarning2.label = "Download the new version from the link below"
            ShaderUpdateWarning2.name = "Shader Version Warning2"
            ShaderUpdateWarning2.use_custom_color = True
            ShaderUpdateWarning2.color = (1,0,0)
            ShaderUpdateWarning2.label_size = 20
            ShaderUpdateWarning2.shrink = True
            ShaderUpdateWarning2.location = (ShaderLocation.x - 200, ShaderLocation.y + 150)
            ShaderUpdateWarning2.height, ShaderUpdateWarning2.width = (50,650)

            ShaderUpdateWarning3 = material.node_tree.nodes.new("NodeFrame")
            ShaderUpdateWarning3.label = "https://tinyurl.com/Helldivers2Shader"
            ShaderUpdateWarning3.name = "Shader Version Warning3"
            ShaderUpdateWarning3.use_custom_color = True
            ShaderUpdateWarning3.color = (1,0,0)
            ShaderUpdateWarning3.label_size = 20
            ShaderUpdateWarning3.shrink = True
            ShaderUpdateWarning3.location = (ShaderLocation.x - 200, ShaderLocation.y + 90)
            ShaderUpdateWarning3.height, ShaderUpdateWarning3.width = (50,650)
    except:
        pass
    
    
#initialize hd2_shader nodes       
    #node Frame.024
    frame_024 = HD2_Shader.nodes.new("NodeFrame")
    frame_024.label = "ID Mask Processing"
    frame_024.name = "Frame.024"
    frame_024.use_custom_color = True
    frame_024.color = (0.07028805464506149, 0.6079999804496765, 0.36107540130615234)
    frame_024.label_size = 20
    frame_024.shrink = True
    
    #node Frame.047
    frame_047 = HD2_Shader.nodes.new("NodeFrame")
    frame_047.label = "Material LUT 22"
    frame_047.name = "Frame.047"
    frame_047.use_custom_color = True
    frame_047.color = (0.0, 0.14901961386203766, 0.0)
    frame_047.label_size = 20
    frame_047.shrink = True
    
    #node Frame.046
    frame_046 = HD2_Shader.nodes.new("NodeFrame")
    frame_046.label = "Material LUT 21"
    frame_046.name = "Frame.046"
    frame_046.use_custom_color = True
    frame_046.color = (0.0, 0.14901961386203766, 0.0)
    frame_046.label_size = 20
    frame_046.shrink = True
    
    #node Frame.045
    frame_045 = HD2_Shader.nodes.new("NodeFrame")
    frame_045.label = "Material LUT 20"
    frame_045.name = "Frame.045"
    frame_045.use_custom_color = True
    frame_045.color = (0.0, 0.14901961386203766, 0.0)
    frame_045.label_size = 20
    frame_045.shrink = True
    
    #node Frame.044
    frame_044 = HD2_Shader.nodes.new("NodeFrame")
    frame_044.label = "Material LUT 19"
    frame_044.name = "Frame.044"
    frame_044.use_custom_color = True
    frame_044.color = (0.0, 0.14901961386203766, 0.0)
    frame_044.label_size = 20
    frame_044.shrink = True
    
    #node Frame.043
    frame_043 = HD2_Shader.nodes.new("NodeFrame")
    frame_043.label = "Material LUT 18"
    frame_043.name = "Frame.043"
    frame_043.use_custom_color = True
    frame_043.color = (0.0, 0.14901961386203766, 0.0)
    frame_043.label_size = 20
    frame_043.shrink = True
    
    #node Frame.042
    frame_042 = HD2_Shader.nodes.new("NodeFrame")
    frame_042.label = "Material LUT 17"
    frame_042.name = "Frame.042"
    frame_042.use_custom_color = True
    frame_042.color = (0.0, 0.14901961386203766, 0.0)
    frame_042.label_size = 20
    frame_042.shrink = True
    
    #node Frame.041
    frame_041 = HD2_Shader.nodes.new("NodeFrame")
    frame_041.label = "Material LUT 16"
    frame_041.name = "Frame.041"
    frame_041.use_custom_color = True
    frame_041.color = (0.0, 0.14901961386203766, 0.0)
    frame_041.label_size = 20
    frame_041.shrink = True
    
    #node Frame.040
    frame_040 = HD2_Shader.nodes.new("NodeFrame")
    frame_040.label = "Material LUT 15"
    frame_040.name = "Frame.040"
    frame_040.use_custom_color = True
    frame_040.color = (0.0, 0.14901961386203766, 0.0)
    frame_040.label_size = 20
    frame_040.shrink = True
    
    #node Frame.039
    frame_039 = HD2_Shader.nodes.new("NodeFrame")
    frame_039.label = "Material LUT 14"
    frame_039.name = "Frame.039"
    frame_039.use_custom_color = True
    frame_039.color = (0.0, 0.14901961386203766, 0.0)
    frame_039.label_size = 20
    frame_039.shrink = True
    
    #node Frame.038
    frame_038 = HD2_Shader.nodes.new("NodeFrame")
    frame_038.label = "Material LUT 13"
    frame_038.name = "Frame.038"
    frame_038.use_custom_color = True
    frame_038.color = (0.0, 0.14901961386203766, 0.0)
    frame_038.label_size = 20
    frame_038.shrink = True
    
    #node Frame.037
    frame_037 = HD2_Shader.nodes.new("NodeFrame")
    frame_037.label = "Material LUT 12"
    frame_037.name = "Frame.037"
    frame_037.use_custom_color = True
    frame_037.color = (0.0, 0.14901961386203766, 0.0)
    frame_037.label_size = 20
    frame_037.shrink = True
    
    #node Frame.036
    frame_036 = HD2_Shader.nodes.new("NodeFrame")
    frame_036.label = "Material LUT 11"
    frame_036.name = "Frame.036"
    frame_036.use_custom_color = True
    frame_036.color = (0.0, 0.14901961386203766, 0.0)
    frame_036.label_size = 20
    frame_036.shrink = True
    
    #node Frame.035
    frame_035 = HD2_Shader.nodes.new("NodeFrame")
    frame_035.label = "Material LUT 10"
    frame_035.name = "Frame.035"
    frame_035.use_custom_color = True
    frame_035.color = (0.0, 0.14901961386203766, 0.0)
    frame_035.label_size = 20
    frame_035.shrink = True
    
    #node Frame.034
    frame_034 = HD2_Shader.nodes.new("NodeFrame")
    frame_034.label = "Material LUT 09"
    frame_034.name = "Frame.034"
    frame_034.use_custom_color = True
    frame_034.color = (0.0, 0.14901961386203766, 0.0)
    frame_034.label_size = 20
    frame_034.shrink = True
    
    #node Frame.033
    frame_033 = HD2_Shader.nodes.new("NodeFrame")
    frame_033.label = "Material LUT 08"
    frame_033.name = "Frame.033"
    frame_033.use_custom_color = True
    frame_033.color = (0.0, 0.14901961386203766, 0.0)
    frame_033.label_size = 20
    frame_033.shrink = True
    
    #node Frame.032
    frame_032 = HD2_Shader.nodes.new("NodeFrame")
    frame_032.label = "Material LUT 07"
    frame_032.name = "Frame.032"
    frame_032.use_custom_color = True
    frame_032.color = (0.0, 0.14901961386203766, 0.0)
    frame_032.label_size = 20
    frame_032.shrink = True
    
    #node Frame.031
    frame_031 = HD2_Shader.nodes.new("NodeFrame")
    frame_031.label = "Material LUT 06"
    frame_031.name = "Frame.031"
    frame_031.use_custom_color = True
    frame_031.color = (0.0, 0.14901961386203766, 0.0)
    frame_031.label_size = 20
    frame_031.shrink = True
    
    #node Frame.030
    frame_030 = HD2_Shader.nodes.new("NodeFrame")
    frame_030.label = "Material LUT 05"
    frame_030.name = "Frame.030"
    frame_030.use_custom_color = True
    frame_030.color = (0.0, 0.14901961386203766, 0.0)
    frame_030.label_size = 20
    frame_030.shrink = True
    
    #node Frame.028
    frame_028 = HD2_Shader.nodes.new("NodeFrame")
    frame_028.label = "Material LUT 03"
    frame_028.name = "Frame.028"
    frame_028.use_custom_color = True
    frame_028.color = (0.0, 0.14901961386203766, 0.0)
    frame_028.label_size = 20
    frame_028.shrink = True
    
    #node Frame.027
    frame_027 = HD2_Shader.nodes.new("NodeFrame")
    frame_027.label = "Material LUT 02"
    frame_027.name = "Frame.027"
    frame_027.use_custom_color = True
    frame_027.color = (0.0, 0.14901961386203766, 0.0)
    frame_027.label_size = 20
    frame_027.shrink = True
    
    #node Frame.026
    frame_026 = HD2_Shader.nodes.new("NodeFrame")
    frame_026.label = "Material LUT 01"
    frame_026.name = "Frame.026"
    frame_026.use_custom_color = True
    frame_026.color = (0.0, 0.14901961386203766, 0.0)
    frame_026.label_size = 20
    frame_026.shrink = True
    
    #node Frame.025
    frame_025 = HD2_Shader.nodes.new("NodeFrame")
    frame_025.label = "Material LUT 00"
    frame_025.name = "Frame.025"
    frame_025.use_custom_color = True
    frame_025.color = (0.0, 0.14901961386203766, 0.0)
    frame_025.label_size = 20
    frame_025.shrink = True
    
    #node Frame.016
    frame_016 = HD2_Shader.nodes.new("NodeFrame")
    frame_016.label = "Not enough data"
    frame_016.name = "Frame.016"
    frame_016.use_custom_color = True
    frame_016.color = (0.6079999804496765, 0.03723505139350891, 0.0)
    frame_016.label_size = 20
    frame_016.shrink = True
    
    #node Frame.004
    frame_004 = HD2_Shader.nodes.new("NodeFrame")
    frame_004.label = "if (r5.x != 0) (#250)"
    frame_004.name = "Frame.004"
    frame_004.label_size = 30
    frame_004.shrink = True
    
    #node Frame.005
    frame_005 = HD2_Shader.nodes.new("NodeFrame")
    frame_005.label = "ELSE"
    frame_005.name = "Frame.005"
    frame_005.use_custom_color = True
    frame_005.color = (0.5340853333473206, 0.5340853333473206, 0.5340853333473206)
    frame_005.label_size = 20
    frame_005.shrink = True
    
    #node Frame.013
    frame_013 = HD2_Shader.nodes.new("NodeFrame")
    frame_013.label = "if (r7.y != 0) (#268)"
    frame_013.name = "Frame.013"
    frame_013.label_size = 30
    frame_013.shrink = True
    
    #node Frame.014
    frame_014 = HD2_Shader.nodes.new("NodeFrame")
    frame_014.label = "Not enough data"
    frame_014.name = "Frame.014"
    frame_014.use_custom_color = True
    frame_014.color = (0.6079999804496765, 0.0, 0.018280426040291786)
    frame_014.label_size = 20
    frame_014.shrink = True
    
    #node Frame.015
    frame_015 = HD2_Shader.nodes.new("NodeFrame")
    frame_015.label = "ELSE"
    frame_015.name = "Frame.015"
    frame_015.use_custom_color = True
    frame_015.color = (0.6079999804496765, 0.6079999804496765, 0.6079999804496765)
    frame_015.label_size = 20
    frame_015.shrink = True
    
    #node Frame.001
    frame_001 = HD2_Shader.nodes.new("NodeFrame")
    frame_001.label = "if (r15.y != 0) (#301)"
    frame_001.name = "Frame.001"
    frame_001.label_size = 30
    frame_001.shrink = True
    
    #node Frame.020
    frame_020 = HD2_Shader.nodes.new("NodeFrame")
    frame_020.label = "ELSE"
    frame_020.name = "Frame.020"
    frame_020.use_custom_color = True
    frame_020.color = (0.6079999804496765, 0.6079999804496765, 0.6079999804496765)
    frame_020.label_size = 20
    frame_020.shrink = True
    
    #node Frame
    frame = HD2_Shader.nodes.new("NodeFrame")
    frame.label = "if (r5.x != 0) (#350)"
    frame.name = "Frame"
    frame.label_size = 30
    frame.shrink = True
    
    #node Frame.006
    frame_006 = HD2_Shader.nodes.new("NodeFrame")
    frame_006.label = "Not enough data"
    frame_006.name = "Frame.006"
    frame_006.use_custom_color = True
    frame_006.color = (0.6079999804496765, 0.0, 0.004571700934320688)
    frame_006.label_size = 20
    frame_006.shrink = True
    
    #node Frame.017
    frame_017 = HD2_Shader.nodes.new("NodeFrame")
    frame_017.label = "if (r15.y != 0) (#368)"
    frame_017.name = "Frame.017"
    frame_017.label_size = 30
    frame_017.shrink = True
    
    #node Frame.018
    frame_018 = HD2_Shader.nodes.new("NodeFrame")
    frame_018.label = "Impossible End"
    frame_018.name = "Frame.018"
    frame_018.use_custom_color = True
    frame_018.color = (0.6079999804496765, 0.02018662914633751, 0.0)
    frame_018.label_size = 20
    frame_018.shrink = True
    
    #node Frame.019
    frame_019 = HD2_Shader.nodes.new("NodeFrame")
    frame_019.label = "vec4 Dot Product (#327)"
    frame_019.name = "Frame.019"
    frame_019.use_custom_color = True
    frame_019.color = (0.1667698621749878, 0.16676993668079376, 0.3352196514606476)
    frame_019.label_size = 20
    frame_019.shrink = True
    
    #node Frame.048
    frame_048 = HD2_Shader.nodes.new("NodeFrame")
    frame_048.label = "vec4 Dot Product r6.z (#415)"
    frame_048.name = "Frame.048"
    frame_048.use_custom_color = True
    frame_048.color = (0.1667698621749878, 0.16676993668079376, 0.3352196514606476)
    frame_048.label_size = 20
    frame_048.shrink = True
    
    #node Frame.007
    frame_007 = HD2_Shader.nodes.new("NodeFrame")
    frame_007.label = "if (r6.w != 0) (#443)"
    frame_007.name = "Frame.007"
    frame_007.label_size = 30
    frame_007.shrink = True
    
    #node Frame.008
    frame_008 = HD2_Shader.nodes.new("NodeFrame")
    frame_008.label = "if (r10.x != 0) (#468)"
    frame_008.name = "Frame.008"
    frame_008.use_custom_color = True
    frame_008.color = (0.17613081634044647, 0.17613081634044647, 0.17613081634044647)
    frame_008.label_size = 30
    frame_008.shrink = True
    
    #node Frame.009
    frame_009 = HD2_Shader.nodes.new("NodeFrame")
    frame_009.label = "ELSE"
    frame_009.name = "Frame.009"
    frame_009.use_custom_color = True
    frame_009.color = (0.6079999804496765, 0.6079999804496765, 0.6079999804496765)
    frame_009.label_size = 20
    frame_009.shrink = True
    
    #node Frame.010
    frame_010 = HD2_Shader.nodes.new("NodeFrame")
    frame_010.label = "ELSE"
    frame_010.name = "Frame.010"
    frame_010.use_custom_color = True
    frame_010.color = (0.6079999804496765, 0.6079999804496765, 0.6079999804496765)
    frame_010.label_size = 20
    frame_010.shrink = True
    
    #node Frame.002
    frame_002 = HD2_Shader.nodes.new("NodeFrame")
    frame_002.label = "if (r7.y != 0) (#497)"
    frame_002.name = "Frame.002"
    frame_002.label_size = 30
    frame_002.shrink = True
    
    #node Frame.003
    frame_003 = HD2_Shader.nodes.new("NodeFrame")
    frame_003.label = "if (r7.y != 0) (#500)"
    frame_003.name = "Frame.003"
    frame_003.use_custom_color = True
    frame_003.color = (0.15000000596046448, 0.15000000596046448, 0.15000000596046448)
    frame_003.label_size = 30
    frame_003.shrink = True
    
    #node Frame.011
    frame_011 = HD2_Shader.nodes.new("NodeFrame")
    frame_011.label = "if (r10.x != 0) (#521)"
    frame_011.name = "Frame.011"
    frame_011.use_custom_color = True
    frame_011.color = (0.25, 0.25, 0.25)
    frame_011.label_size = 30
    frame_011.shrink = True
    
    #node Frame.012
    frame_012 = HD2_Shader.nodes.new("NodeFrame")
    frame_012.label = "if (r10.x != 0) (#525)"
    frame_012.name = "Frame.012"
    frame_012.use_custom_color = True
    frame_012.color = (0.3499999940395355, 0.3499999940395355, 0.3499999940395355)
    frame_012.label_size = 30
    frame_012.shrink = True
    
    #node Frame.021
    frame_021 = HD2_Shader.nodes.new("NodeFrame")
    frame_021.label = "if (r7.x != 0)"
    frame_021.name = "Frame.021"
    frame_021.label_size = 30
    frame_021.shrink = True
    
    #node Frame.022
    frame_022 = HD2_Shader.nodes.new("NodeFrame")
    frame_022.name = "Frame.022"
    frame_022.label_size = 30
    frame_022.shrink = True
    
    #node Frame.023
    frame_023 = HD2_Shader.nodes.new("NodeFrame")
    frame_023.label = "ELSE"
    frame_023.name = "Frame.023"
    frame_023.use_custom_color = True
    frame_023.color = (0.5965854525566101, 0.5965854525566101, 0.5965854525566101)
    frame_023.label_size = 20
    frame_023.shrink = True
    
    #node Frame.049
    frame_049 = HD2_Shader.nodes.new("NodeFrame")
    frame_049.label = "Metal/Specular"
    frame_049.name = "Frame.049"
    frame_049.use_custom_color = True
    frame_049.color = (0.42044681310653687, 0.42044681310653687, 0.42044681310653687)
    frame_049.label_size = 20
    frame_049.shrink = True
    
    #node Frame.050
    frame_050 = HD2_Shader.nodes.new("NodeFrame")
    frame_050.label = "Normals"
    frame_050.name = "Frame.050"
    frame_050.use_custom_color = True
    frame_050.color = (0.5, 0.5, 1.0)
    frame_050.label_size = 20
    frame_050.shrink = True
    
    #node Frame.051
    frame_051 = HD2_Shader.nodes.new("NodeFrame")
    frame_051.label = "Clearcoat Normals"
    frame_051.name = "Frame.051"
    frame_051.use_custom_color = True
    frame_051.color = (0.3499999940395355, 0.3499999940395355, 1.0)
    frame_051.label_size = 20
    frame_051.shrink = True
    
    #node Frame.053
    frame_053 = HD2_Shader.nodes.new("NodeFrame")
    frame_053.label = "Ambient Occlusion"
    frame_053.name = "Frame.053"
    frame_053.use_custom_color = True
    frame_053.color = (0.30681028962135315, 0.20898514986038208, 0.1259160190820694)
    frame_053.label_size = 20
    frame_053.shrink = True
    
    #node Frame.052
    frame_052 = HD2_Shader.nodes.new("NodeFrame")
    frame_052.label = "Map Assembly"
    frame_052.name = "Frame.052"
    frame_052.use_custom_color = True
    frame_052.color = (0.3170004189014435, 0.0, 0.3636285364627838)
    frame_052.label_size = 20
    frame_052.shrink = True
    
    #node Frame.029
    frame_029 = HD2_Shader.nodes.new("NodeFrame")
    frame_029.label = "Material LUT 04"
    frame_029.name = "Frame.029"
    frame_029.use_custom_color = True
    frame_029.color = (0.0, 0.14901961386203766, 0.0)
    frame_029.label_size = 20
    frame_029.shrink = True
    
    #node Combine XYZ.076
    combine_xyz_076 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_076.name = "Combine XYZ.076"
    #X
    combine_xyz_076.inputs[0].default_value = 0.5454545617103577
    #Z
    combine_xyz_076.inputs[2].default_value = 0.0
    
    #node Combine XYZ.077
    combine_xyz_077 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_077.name = "Combine XYZ.077"
    #X
    combine_xyz_077.inputs[0].default_value = 0.5454545617103577
    #Z
    combine_xyz_077.inputs[2].default_value = 0.0
    
    #node Mix.055
    mix_055 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_055.name = "Mix.055"
    mix_055.blend_type = 'MIX'
    mix_055.clamp_factor = True
    mix_055.clamp_result = False
    mix_055.data_type = 'RGBA'
    mix_055.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_055.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_055.inputs[2].default_value = 0.0
    #B_Float
    mix_055.inputs[3].default_value = 0.0
    #A_Vector
    mix_055.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_055.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_055.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_055.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.056
    mix_056 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_056.name = "Mix.056"
    mix_056.blend_type = 'MIX'
    mix_056.clamp_factor = True
    mix_056.clamp_result = False
    mix_056.data_type = 'FLOAT'
    mix_056.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_056.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_056.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_056.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_056.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_056.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_056.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_056.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Primary Material LUT_12
    primary_material_lut_12 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_12.label = "Primary Material LUT_12"
    primary_material_lut_12.name = "Primary Material LUT_12"
    primary_material_lut_12.use_custom_color = True
    primary_material_lut_12.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_12.extension = 'EXTEND'
    primary_material_lut_12.image_user.frame_current = 1
    primary_material_lut_12.image_user.frame_duration = 1
    primary_material_lut_12.image_user.frame_offset = 429
    primary_material_lut_12.image_user.frame_start = 1
    primary_material_lut_12.image_user.tile = 0
    primary_material_lut_12.image_user.use_auto_refresh = False
    primary_material_lut_12.image_user.use_cyclic = False
    primary_material_lut_12.interpolation = 'Closest'
    primary_material_lut_12.projection = 'FLAT'
    primary_material_lut_12.projection_blend = 0.0

    #node Primary Material LUT_13
    primary_material_lut_13 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_13.label = "Primary Material LUT_13"
    primary_material_lut_13.name = "Primary Material LUT_13"
    primary_material_lut_13.use_custom_color = True
    primary_material_lut_13.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_13.extension = 'EXTEND'
    primary_material_lut_13.image_user.frame_current = 1
    primary_material_lut_13.image_user.frame_duration = 1
    primary_material_lut_13.image_user.frame_offset = 429
    primary_material_lut_13.image_user.frame_start = 1
    primary_material_lut_13.image_user.tile = 0
    primary_material_lut_13.image_user.use_auto_refresh = False
    primary_material_lut_13.image_user.use_cyclic = False
    primary_material_lut_13.interpolation = 'Closest'
    primary_material_lut_13.projection = 'FLAT'
    primary_material_lut_13.projection_blend = 0.0

    #node Combine XYZ.078
    combine_xyz_078 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_078.name = "Combine XYZ.078"
    #X
    combine_xyz_078.inputs[0].default_value = 0.5909090638160706
    #Z
    combine_xyz_078.inputs[2].default_value = 0.0
    
    #node Combine XYZ.084
    combine_xyz_084 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_084.name = "Combine XYZ.084"
    #X
    combine_xyz_084.inputs[0].default_value = 0.7272727489471436
    #Z
    combine_xyz_084.inputs[2].default_value = 0.0
    
    #node Combine XYZ.085
    combine_xyz_085 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_085.name = "Combine XYZ.085"
    #X
    combine_xyz_085.inputs[0].default_value = 0.7272727489471436
    #Z
    combine_xyz_085.inputs[2].default_value = 0.0
    
    #node Primary Material LUT_16
    primary_material_lut_16 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_16.label = "Primary Material LUT_16"
    primary_material_lut_16.name = "Primary Material LUT_16"
    primary_material_lut_16.use_custom_color = True
    primary_material_lut_16.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_16.extension = 'EXTEND'
    primary_material_lut_16.image_user.frame_current = 1
    primary_material_lut_16.image_user.frame_duration = 1
    primary_material_lut_16.image_user.frame_offset = 429
    primary_material_lut_16.image_user.frame_start = 1
    primary_material_lut_16.image_user.tile = 0
    primary_material_lut_16.image_user.use_auto_refresh = False
    primary_material_lut_16.image_user.use_cyclic = False
    primary_material_lut_16.interpolation = 'Closest'
    primary_material_lut_16.projection = 'FLAT'
    primary_material_lut_16.projection_blend = 0.0

    #node Mix.071
    mix_071 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_071.name = "Mix.071"
    mix_071.blend_type = 'MIX'
    mix_071.clamp_factor = True
    mix_071.clamp_result = False
    mix_071.data_type = 'RGBA'
    mix_071.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_071.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_071.inputs[2].default_value = 0.0
    #B_Float
    mix_071.inputs[3].default_value = 0.0
    #A_Vector
    mix_071.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_071.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_071.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_071.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.057
    mix_057 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_057.name = "Mix.057"
    mix_057.blend_type = 'MIX'
    mix_057.clamp_factor = True
    mix_057.clamp_result = False
    mix_057.data_type = 'RGBA'
    mix_057.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_057.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_057.inputs[2].default_value = 0.0
    #B_Float
    mix_057.inputs[3].default_value = 0.0
    #A_Vector
    mix_057.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_057.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_057.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_057.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.058
    mix_058 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_058.name = "Mix.058"
    mix_058.blend_type = 'MIX'
    mix_058.clamp_factor = True
    mix_058.clamp_result = False
    mix_058.data_type = 'FLOAT'
    mix_058.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_058.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_058.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_058.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_058.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_058.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_058.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_058.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.080
    combine_xyz_080 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_080.name = "Combine XYZ.080"
    #X
    combine_xyz_080.inputs[0].default_value = 0.6363636255264282
    #Z
    combine_xyz_080.inputs[2].default_value = 0.0
    
    #node Combine XYZ.081
    combine_xyz_081 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_081.name = "Combine XYZ.081"
    #X
    combine_xyz_081.inputs[0].default_value = 0.6363636255264282
    #Z
    combine_xyz_081.inputs[2].default_value = 0.0
    
    #node Primary Material LUT_14
    primary_material_lut_14 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_14.label = "Primary Material LUT_14"
    primary_material_lut_14.name = "Primary Material LUT_14"
    primary_material_lut_14.use_custom_color = True
    primary_material_lut_14.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_14.extension = 'EXTEND'
    primary_material_lut_14.image_user.frame_current = 1
    primary_material_lut_14.image_user.frame_duration = 1
    primary_material_lut_14.image_user.frame_offset = 429
    primary_material_lut_14.image_user.frame_start = 1
    primary_material_lut_14.image_user.tile = 0
    primary_material_lut_14.image_user.use_auto_refresh = False
    primary_material_lut_14.image_user.use_cyclic = False
    primary_material_lut_14.interpolation = 'Closest'
    primary_material_lut_14.projection = 'FLAT'
    primary_material_lut_14.projection_blend = 0.0

    #node Mix.059
    mix_059 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_059.name = "Mix.059"
    mix_059.blend_type = 'MIX'
    mix_059.clamp_factor = True
    mix_059.clamp_result = False
    mix_059.data_type = 'RGBA'
    mix_059.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_059.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_059.inputs[2].default_value = 0.0
    #B_Float
    mix_059.inputs[3].default_value = 0.0
    #A_Vector
    mix_059.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_059.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_059.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_059.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.060
    mix_060 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_060.name = "Mix.060"
    mix_060.blend_type = 'MIX'
    mix_060.clamp_factor = True
    mix_060.clamp_result = False
    mix_060.data_type = 'FLOAT'
    mix_060.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_060.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_060.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_060.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_060.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_060.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_060.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_060.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.061
    mix_061 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_061.name = "Mix.061"
    mix_061.blend_type = 'MIX'
    mix_061.clamp_factor = True
    mix_061.clamp_result = False
    mix_061.data_type = 'RGBA'
    mix_061.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_061.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_061.inputs[2].default_value = 0.0
    #B_Float
    mix_061.inputs[3].default_value = 0.0
    #A_Vector
    mix_061.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_061.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_061.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_061.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.062
    mix_062 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_062.name = "Mix.062"
    mix_062.blend_type = 'MIX'
    mix_062.clamp_factor = True
    mix_062.clamp_result = False
    mix_062.data_type = 'FLOAT'
    mix_062.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_062.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_062.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_062.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_062.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_062.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_062.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_062.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.083
    combine_xyz_083 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_083.name = "Combine XYZ.083"
    #X
    combine_xyz_083.inputs[0].default_value = 0.6818181872367859
    #Z
    combine_xyz_083.inputs[2].default_value = 0.0
    
    #node Primary Material LUT_15
    primary_material_lut_15 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_15.label = "Primary Material LUT_15"
    primary_material_lut_15.name = "Primary Material LUT_15"
    primary_material_lut_15.use_custom_color = True
    primary_material_lut_15.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_15.extension = 'EXTEND'
    primary_material_lut_15.image_user.frame_current = 1
    primary_material_lut_15.image_user.frame_duration = 1
    primary_material_lut_15.image_user.frame_offset = 429
    primary_material_lut_15.image_user.frame_start = 1
    primary_material_lut_15.image_user.tile = 0
    primary_material_lut_15.image_user.use_auto_refresh = False
    primary_material_lut_15.image_user.use_cyclic = False
    primary_material_lut_15.interpolation = 'Closest'
    primary_material_lut_15.projection = 'FLAT'
    primary_material_lut_15.projection_blend = 0.0
    
    #node Mix.063
    mix_063 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_063.name = "Mix.063"
    mix_063.blend_type = 'MIX'
    mix_063.clamp_factor = True
    mix_063.clamp_result = False
    mix_063.data_type = 'RGBA'
    mix_063.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_063.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_063.inputs[2].default_value = 0.0
    #B_Float
    mix_063.inputs[3].default_value = 0.0
    #A_Vector
    mix_063.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_063.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_063.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_063.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.064
    mix_064 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_064.name = "Mix.064"
    mix_064.blend_type = 'MIX'
    mix_064.clamp_factor = True
    mix_064.clamp_result = False
    mix_064.data_type = 'FLOAT'
    mix_064.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_064.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_064.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_064.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_064.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_064.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_064.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_064.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.070
    mix_070 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_070.name = "Mix.070"
    mix_070.blend_type = 'MIX'
    mix_070.clamp_factor = True
    mix_070.clamp_result = False
    mix_070.data_type = 'FLOAT'
    mix_070.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_070.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_070.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_070.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_070.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_070.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_070.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_070.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.073
    mix_073 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_073.name = "Mix.073"
    mix_073.blend_type = 'MIX'
    mix_073.clamp_factor = True
    mix_073.clamp_result = False
    mix_073.data_type = 'RGBA'
    mix_073.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_073.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_073.inputs[2].default_value = 0.0
    #B_Float
    mix_073.inputs[3].default_value = 0.0
    #A_Vector
    mix_073.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_073.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_073.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_073.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.076
    mix_076 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_076.name = "Mix.076"
    mix_076.blend_type = 'MIX'
    mix_076.clamp_factor = True
    mix_076.clamp_result = False
    mix_076.data_type = 'FLOAT'
    mix_076.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_076.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_076.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_076.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_076.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_076.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_076.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_076.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Separate XYZ.052
    separate_xyz_052 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_052.name = "Separate XYZ.052"
    
    #node Mix.032
    mix_032 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_032.name = "Mix.032"
    mix_032.blend_type = 'MIX'
    mix_032.clamp_factor = True
    mix_032.clamp_result = False
    mix_032.data_type = 'RGBA'
    mix_032.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_032.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_032.inputs[2].default_value = 0.0
    #B_Float
    mix_032.inputs[3].default_value = 0.0
    #A_Vector
    mix_032.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_032.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_032.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_032.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.056
    combine_xyz_056 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_056.name = "Combine XYZ.056"
    #X
    combine_xyz_056.inputs[0].default_value = 0.09090909361839294
    #Z
    combine_xyz_056.inputs[2].default_value = 0.0
    
    #node Mix.035
    mix_035 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_035.name = "Mix.035"
    mix_035.blend_type = 'MIX'
    mix_035.clamp_factor = True
    mix_035.clamp_result = False
    mix_035.data_type = 'RGBA'
    mix_035.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_035.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_035.inputs[2].default_value = 0.0
    #B_Float
    mix_035.inputs[3].default_value = 0.0
    #A_Vector
    mix_035.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_035.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_035.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_035.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.057
    combine_xyz_057 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_057.name = "Combine XYZ.057"
    #X
    combine_xyz_057.inputs[0].default_value = 0.09090909361839294
    #Z
    combine_xyz_057.inputs[2].default_value = 0.0
    
    #node Mix.036
    mix_036 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_036.name = "Mix.036"
    mix_036.blend_type = 'MIX'
    mix_036.clamp_factor = True
    mix_036.clamp_result = False
    mix_036.data_type = 'FLOAT'
    mix_036.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_036.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_036.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_036.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_036.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_036.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_036.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_036.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Primary Material LUT_02
    primary_material_lut_02 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_02.label = "Primary Material LUT_02"
    primary_material_lut_02.name = "Primary Material LUT_02"
    primary_material_lut_02.use_custom_color = True
    primary_material_lut_02.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_02.extension = 'EXTEND'
    primary_material_lut_02.image_user.frame_current = 1
    primary_material_lut_02.image_user.frame_duration = 1
    primary_material_lut_02.image_user.frame_offset = 429
    primary_material_lut_02.image_user.frame_start = 1
    primary_material_lut_02.image_user.tile = 0
    primary_material_lut_02.image_user.use_auto_refresh = False
    primary_material_lut_02.image_user.use_cyclic = False
    primary_material_lut_02.interpolation = 'Closest'
    primary_material_lut_02.projection = 'FLAT'
    primary_material_lut_02.projection_blend = 0.0
    
    #node Combine XYZ.058
    combine_xyz_058 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_058.name = "Combine XYZ.058"
    #X
    combine_xyz_058.inputs[0].default_value = 0.13636364042758942
    #Z
    combine_xyz_058.inputs[2].default_value = 0.0
    
    #node Primary Material LUT_03
    primary_material_lut_03 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_03.label = "Primary Material LUT_03"
    primary_material_lut_03.name = "Primary Material LUT_03"
    primary_material_lut_03.use_custom_color = True
    primary_material_lut_03.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_03.extension = 'EXTEND'
    primary_material_lut_03.image_user.frame_current = 1
    primary_material_lut_03.image_user.frame_duration = 1
    primary_material_lut_03.image_user.frame_offset = 429
    primary_material_lut_03.image_user.frame_start = 1
    primary_material_lut_03.image_user.tile = 0
    primary_material_lut_03.image_user.use_auto_refresh = False
    primary_material_lut_03.image_user.use_cyclic = False
    primary_material_lut_03.interpolation = 'Closest'
    primary_material_lut_03.projection = 'FLAT'
    primary_material_lut_03.projection_blend = 0.0
    
    #node Combine XYZ.059
    combine_xyz_059 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_059.name = "Combine XYZ.059"
    #X
    combine_xyz_059.inputs[0].default_value = 0.13636364042758942
    #Z
    combine_xyz_059.inputs[2].default_value = 0.0
    
    #node Combine XYZ.062
    combine_xyz_062 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_062.name = "Combine XYZ.062"
    #X
    combine_xyz_062.inputs[0].default_value = 0.22727273404598236
    #Z
    combine_xyz_062.inputs[2].default_value = 0.0
    
    #node Mix.041
    mix_041 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_041.name = "Mix.041"
    mix_041.blend_type = 'MIX'
    mix_041.clamp_factor = True
    mix_041.clamp_result = False
    mix_041.data_type = 'RGBA'
    mix_041.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_041.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_041.inputs[2].default_value = 0.0
    #B_Float
    mix_041.inputs[3].default_value = 0.0
    #A_Vector
    mix_041.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_041.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_041.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_041.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.042
    mix_042 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_042.name = "Mix.042"
    mix_042.blend_type = 'MIX'
    mix_042.clamp_factor = True
    mix_042.clamp_result = False
    mix_042.data_type = 'FLOAT'
    mix_042.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_042.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_042.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_042.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_042.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_042.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_042.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_042.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.043
    mix_043 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_043.name = "Mix.043"
    mix_043.blend_type = 'MIX'
    mix_043.clamp_factor = True
    mix_043.clamp_result = False
    mix_043.data_type = 'RGBA'
    mix_043.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_043.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_043.inputs[2].default_value = 0.0
    #B_Float
    mix_043.inputs[3].default_value = 0.0
    #A_Vector
    mix_043.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_043.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_043.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_043.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.068
    combine_xyz_068 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_068.name = "Combine XYZ.068"
    #X
    combine_xyz_068.inputs[0].default_value = 0.3636363744735718
    #Z
    combine_xyz_068.inputs[2].default_value = 0.0
    
    #node Mix.047
    mix_047 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_047.name = "Mix.047"
    mix_047.blend_type = 'MIX'
    mix_047.clamp_factor = True
    mix_047.clamp_result = False
    mix_047.data_type = 'RGBA'
    mix_047.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_047.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_047.inputs[2].default_value = 0.0
    #B_Float
    mix_047.inputs[3].default_value = 0.0
    #A_Vector
    mix_047.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_047.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_047.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_047.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.048
    mix_048 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_048.name = "Mix.048"
    mix_048.blend_type = 'MIX'
    mix_048.clamp_factor = True
    mix_048.clamp_result = False
    mix_048.data_type = 'FLOAT'
    mix_048.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_048.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_048.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_048.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_048.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_048.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_048.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_048.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Primary Material LUT_08
    primary_material_lut_08 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_08.label = "Primary Material LUT_08"
    primary_material_lut_08.name = "Primary Material LUT_08"
    primary_material_lut_08.use_custom_color = True
    primary_material_lut_08.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_08.extension = 'EXTEND'
    primary_material_lut_08.image_user.frame_current = 1
    primary_material_lut_08.image_user.frame_duration = 1
    primary_material_lut_08.image_user.frame_offset = 429
    primary_material_lut_08.image_user.frame_start = 1
    primary_material_lut_08.image_user.tile = 0
    primary_material_lut_08.image_user.use_auto_refresh = False
    primary_material_lut_08.image_user.use_cyclic = False
    primary_material_lut_08.interpolation = 'Closest'
    primary_material_lut_08.projection = 'FLAT'
    primary_material_lut_08.projection_blend = 0.0
    
    #node Combine XYZ.070
    combine_xyz_070 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_070.name = "Combine XYZ.070"
    #X
    combine_xyz_070.inputs[0].default_value = 0.40909090638160706
    #Z
    combine_xyz_070.inputs[2].default_value = 0.0
    
    #node Combine XYZ.071
    combine_xyz_071 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_071.name = "Combine XYZ.071"
    #X
    combine_xyz_071.inputs[0].default_value = 0.40909090638160706
    #Z
    combine_xyz_071.inputs[2].default_value = 0.0
    
    #node Mix.050
    mix_050 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_050.name = "Mix.050"
    mix_050.blend_type = 'MIX'
    mix_050.clamp_factor = True
    mix_050.clamp_result = False
    mix_050.data_type = 'FLOAT'
    mix_050.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_050.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_050.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_050.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_050.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_050.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_050.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_050.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Primary Material LUT_09
    primary_material_lut_09 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_09.label = "Primary Material LUT_09"
    primary_material_lut_09.name = "Primary Material LUT_09"
    primary_material_lut_09.use_custom_color = True
    primary_material_lut_09.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_09.extension = 'EXTEND'
    primary_material_lut_09.image_user.frame_current = 1
    primary_material_lut_09.image_user.frame_duration = 1
    primary_material_lut_09.image_user.frame_offset = 429
    primary_material_lut_09.image_user.frame_start = 1
    primary_material_lut_09.image_user.tile = 0
    primary_material_lut_09.image_user.use_auto_refresh = False
    primary_material_lut_09.image_user.use_cyclic = False
    primary_material_lut_09.interpolation = 'Closest'
    primary_material_lut_09.projection = 'FLAT'
    primary_material_lut_09.projection_blend = 0.0
    
    #node Combine XYZ.072
    combine_xyz_072 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_072.name = "Combine XYZ.072"
    #X
    combine_xyz_072.inputs[0].default_value = 0.4545454680919647
    #Z
    combine_xyz_072.inputs[2].default_value = 0.0
    
    #node Combine XYZ.073
    combine_xyz_073 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_073.name = "Combine XYZ.073"
    #X
    combine_xyz_073.inputs[0].default_value = 0.4545454680919647
    #Z
    combine_xyz_073.inputs[2].default_value = 0.0
    
    #node Mix.051
    mix_051 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_051.name = "Mix.051"
    mix_051.blend_type = 'MIX'
    mix_051.clamp_factor = True
    mix_051.clamp_result = False
    mix_051.data_type = 'RGBA'
    mix_051.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_051.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_051.inputs[2].default_value = 0.0
    #B_Float
    mix_051.inputs[3].default_value = 0.0
    #A_Vector
    mix_051.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_051.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_051.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_051.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Primary Material LUT_10
    primary_material_lut_10 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_10.label = "Primary Material LUT_10"
    primary_material_lut_10.name = "Primary Material LUT_10"
    primary_material_lut_10.use_custom_color = True
    primary_material_lut_10.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_10.extension = 'EXTEND'
    primary_material_lut_10.image_user.frame_current = 1
    primary_material_lut_10.image_user.frame_duration = 1
    primary_material_lut_10.image_user.frame_offset = 429
    primary_material_lut_10.image_user.frame_start = 1
    primary_material_lut_10.image_user.tile = 0
    primary_material_lut_10.image_user.use_auto_refresh = False
    primary_material_lut_10.image_user.use_cyclic = False
    primary_material_lut_10.interpolation = 'Closest'
    primary_material_lut_10.projection = 'FLAT'
    primary_material_lut_10.projection_blend = 0.0
    
    #node Mix.053
    mix_053 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_053.name = "Mix.053"
    mix_053.blend_type = 'MIX'
    mix_053.clamp_factor = True
    mix_053.clamp_result = False
    mix_053.data_type = 'RGBA'
    mix_053.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_053.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_053.inputs[2].default_value = 0.0
    #B_Float
    mix_053.inputs[3].default_value = 0.0
    #A_Vector
    mix_053.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_053.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_053.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_053.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.054
    mix_054 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_054.name = "Mix.054"
    mix_054.blend_type = 'MIX'
    mix_054.clamp_factor = True
    mix_054.clamp_result = False
    mix_054.data_type = 'FLOAT'
    mix_054.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_054.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_054.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_054.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_054.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_054.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_054.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_054.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Primary Material LUT_11
    primary_material_lut_11 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_11.label = "Primary Material LUT_11"
    primary_material_lut_11.name = "Primary Material LUT_11"
    primary_material_lut_11.use_custom_color = True
    primary_material_lut_11.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_11.extension = 'EXTEND'
    primary_material_lut_11.image_user.frame_current = 1
    primary_material_lut_11.image_user.frame_duration = 1
    primary_material_lut_11.image_user.frame_offset = 429
    primary_material_lut_11.image_user.frame_start = 1
    primary_material_lut_11.image_user.tile = 0
    primary_material_lut_11.image_user.use_auto_refresh = False
    primary_material_lut_11.image_user.use_cyclic = False
    primary_material_lut_11.interpolation = 'Closest'
    primary_material_lut_11.projection = 'FLAT'
    primary_material_lut_11.projection_blend = 0.0
    
    #node Combine XYZ.063
    combine_xyz_063 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_063.name = "Combine XYZ.063"
    #X
    combine_xyz_063.inputs[0].default_value = 0.22727273404598236
    #Z
    combine_xyz_063.inputs[2].default_value = 0.0
    
    #node Combine XYZ.069
    combine_xyz_069 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_069.name = "Combine XYZ.069"
    #X
    combine_xyz_069.inputs[0].default_value = 0.3636363744735718
    #Z
    combine_xyz_069.inputs[2].default_value = 0.0
    
    #node Combine XYZ.075
    combine_xyz_075 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_075.name = "Combine XYZ.075"
    #X
    combine_xyz_075.inputs[0].default_value = 0.5
    #Z
    combine_xyz_075.inputs[2].default_value = 0.0
    
    #node Combine XYZ.074
    combine_xyz_074 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_074.name = "Combine XYZ.074"
    #X
    combine_xyz_074.inputs[0].default_value = 0.5
    #Z
    combine_xyz_074.inputs[2].default_value = 0.0
    
    #node Mix.049
    mix_049 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_049.name = "Mix.049"
    mix_049.blend_type = 'MIX'
    mix_049.clamp_factor = True
    mix_049.clamp_result = False
    mix_049.data_type = 'RGBA'
    mix_049.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_049.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_049.inputs[2].default_value = 0.0
    #B_Float
    mix_049.inputs[3].default_value = 0.0
    #A_Vector
    mix_049.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_049.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_049.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_049.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.052
    mix_052 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_052.name = "Mix.052"
    mix_052.blend_type = 'MIX'
    mix_052.clamp_factor = True
    mix_052.clamp_result = False
    mix_052.data_type = 'FLOAT'
    mix_052.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_052.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_052.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_052.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_052.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_052.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_052.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_052.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Gamma.005
    gamma_005 = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma_005.name = "Gamma.005"
    #Gamma
    gamma_005.inputs[1].default_value = 2.200000047683716

    #node Gamma.006
    gamma_006 = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma_006.name = "Gamma.006"
    #Gamma
    gamma_006.inputs[1].default_value = 2.200000047683716

    #node Gamma.004
    gamma_004 = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma_004.name = "Gamma.004"
    #Gamma
    gamma_004.inputs[1].default_value = 2.200000047683716
    
    #node Separate XYZ.038
    separate_xyz_038 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_038.name = "Separate XYZ.038"
    
    #node Separate XYZ.057
    separate_xyz_057 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_057.name = "Separate XYZ.057"
    
    #node Separate XYZ.058
    separate_xyz_058 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_058.name = "Separate XYZ.058"
    
    #node Mix.065
    mix_065 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_065.name = "Mix.065"
    mix_065.blend_type = 'MIX'
    mix_065.clamp_factor = True
    mix_065.clamp_result = False
    mix_065.data_type = 'RGBA'
    mix_065.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_065.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_065.inputs[2].default_value = 0.0
    #B_Float
    mix_065.inputs[3].default_value = 0.0
    #A_Vector
    mix_065.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_065.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_065.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_065.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.067
    mix_067 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_067.name = "Mix.067"
    mix_067.blend_type = 'MIX'
    mix_067.clamp_factor = True
    mix_067.clamp_result = False
    mix_067.data_type = 'RGBA'
    mix_067.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_067.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_067.inputs[2].default_value = 0.0
    #B_Float
    mix_067.inputs[3].default_value = 0.0
    #A_Vector
    mix_067.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_067.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_067.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_067.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.191
    math_191 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_191.label = "-r2.x"
    math_191.name = "Math.191"
    math_191.operation = 'MULTIPLY'
    math_191.use_clamp = False
    #Value_001
    math_191.inputs[1].default_value = -1.0
    #Value_002
    math_191.inputs[2].default_value = 0.5
    
    #node Clamp.004
    clamp_004 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_004.label = "r10.w (#224)"
    clamp_004.name = "Clamp.004"
    clamp_004.clamp_type = 'MINMAX'
    #Min
    clamp_004.inputs[1].default_value = 0.0
    #Max
    clamp_004.inputs[2].default_value = 1.0
    
    #node Separate XYZ.028
    separate_xyz_028 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_028.name = "Separate XYZ.028"
    
    #node Combine XYZ.028
    combine_xyz_028 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_028.label = "r7.zxy"
    combine_xyz_028.name = "Combine XYZ.028"
    
    #node Vector Math.023
    vector_math_023 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_023.label = "r11.xyz (#225)"
    vector_math_023.name = "Vector Math.023"
    vector_math_023.operation = 'MULTIPLY_ADD'
    #Vector_001
    vector_math_023.inputs[1].default_value = (0.4000000059604645, -1.0, 1.0)
    #Vector_002
    vector_math_023.inputs[2].default_value = (0.6000000238418579, 1.0, 0.0)
    #Scale
    vector_math_023.inputs[3].default_value = 1.0
    
    #node Math.075
    math_075 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_075.name = "Math.075"
    math_075.operation = 'MULTIPLY'
    math_075.use_clamp = False
    #Value_001
    math_075.inputs[1].default_value = -1.0
    #Value_002
    math_075.inputs[2].default_value = 0.5
    
    #node Separate XYZ.045
    separate_xyz_045 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_045.name = "Separate XYZ.045"
    
    #node Math.074
    math_074 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_074.label = "r3.y (#237)"
    math_074.name = "Math.074"
    math_074.operation = 'ADD'
    math_074.use_clamp = False
    #Value
    math_074.inputs[0].default_value = 1.0
    #Value_002
    math_074.inputs[2].default_value = 0.5
    
    #node Vector Math.024
    vector_math_024 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_024.label = "r6.zw (#226)"
    vector_math_024.name = "Vector Math.024"
    vector_math_024.operation = 'MULTIPLY_ADD'
    #Vector_001
    vector_math_024.inputs[1].default_value = (2.0, 2.0, 0.0)
    #Vector_002
    vector_math_024.inputs[2].default_value = (-1.0, -1.0, 0.0)
    #Scale
    vector_math_024.inputs[3].default_value = 1.0
    
    #node Clamp.003
    clamp_003 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_003.label = "r10.z (#223)"
    clamp_003.name = "Clamp.003"
    clamp_003.clamp_type = 'MINMAX'
    #Min
    clamp_003.inputs[1].default_value = 0.0
    #Max
    clamp_003.inputs[2].default_value = 1.0
    
    #node Separate XYZ.029
    separate_xyz_029 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_029.name = "Separate XYZ.029"
    
    #node Vector Math.025
    vector_math_025 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_025.label = "r7.xy (#227)"
    vector_math_025.name = "Vector Math.025"
    vector_math_025.operation = 'MULTIPLY'
    #Vector_002
    vector_math_025.inputs[2].default_value = (-1.0, -1.0, 0.0)
    #Scale
    vector_math_025.inputs[3].default_value = 1.0
    
    #node Separate XYZ.030
    separate_xyz_030 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_030.name = "Separate XYZ.030"
    
    #node Math.051
    math_051 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_051.label = "-r6.z"
    math_051.name = "Math.051"
    math_051.operation = 'MULTIPLY'
    math_051.use_clamp = False
    #Value_001
    math_051.inputs[1].default_value = -1.0
    #Value_002
    math_051.inputs[2].default_value = 0.5
    
    #node Math.054
    math_054 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_054.label = "-r6.w"
    math_054.name = "Math.054"
    math_054.operation = 'MULTIPLY'
    math_054.use_clamp = False
    #Value_001
    math_054.inputs[1].default_value = -1.0
    #Value_002
    math_054.inputs[2].default_value = 0.5
    
    #node Math.052
    math_052 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_052.label = "r3.y (#228)"
    math_052.name = "Math.052"
    math_052.operation = 'MULTIPLY_ADD'
    math_052.use_clamp = False
    #Value_002
    math_052.inputs[2].default_value = 1.0
    
    #node Math.055
    math_055 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_055.label = "r3.y (#229)"
    math_055.name = "Math.055"
    math_055.operation = 'MULTIPLY_ADD'
    math_055.use_clamp = False
    
    #node Math.056
    math_056 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_056.label = "r4.w (#230)"
    math_056.name = "Math.056"
    math_056.operation = 'MAXIMUM'
    math_056.use_clamp = False
    #Value_002
    math_056.inputs[2].default_value = 1.0
    
    #node Math.053
    math_053 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_053.label = "r4.w (#231)"
    math_053.name = "Math.053"
    math_053.operation = 'MULTIPLY'
    math_053.use_clamp = False
    #Value
    math_053.inputs[0].default_value = 6.103515625e-05
    #Value_002
    math_053.inputs[2].default_value = 0.5
    
    #node Math.057
    math_057 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_057.label = "r3.y (#232)"
    math_057.name = "Math.057"
    math_057.operation = 'MAXIMUM'
    math_057.use_clamp = False
    #Value_002
    math_057.inputs[2].default_value = 0.5
    
    #node Math.058
    math_058 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_058.label = "r3.y (#233)"
    math_058.name = "Math.058"
    math_058.operation = 'SQRT'
    math_058.use_clamp = False
    #Value_001
    math_058.inputs[1].default_value = -1.0
    #Value_002
    math_058.inputs[2].default_value = 1.0
    
    #node Vector Math.026
    vector_math_026 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_026.label = "r6.zw (#234)"
    vector_math_026.name = "Vector Math.026"
    vector_math_026.operation = 'MULTIPLY'
    #Vector_002
    vector_math_026.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_026.inputs[3].default_value = 1.0
    
    #node Separate XYZ.031
    separate_xyz_031 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_031.name = "Separate XYZ.031"
    
    #node Vector Math.032
    vector_math_032 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_032.label = "r2.yzw (#235)"
    vector_math_032.name = "Vector Math.032"
    vector_math_032.operation = 'ADD'
    #Vector_001
    vector_math_032.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_032.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_032.inputs[3].default_value = 1.0
    
    #node Mix.013
    mix_013 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_013.label = "r6.zw if (r5.x != 0)"
    mix_013.name = "Mix.013"
    mix_013.blend_type = 'MIX'
    mix_013.clamp_factor = True
    mix_013.clamp_result = False
    mix_013.data_type = 'VECTOR'
    mix_013.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_013.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_013.inputs[2].default_value = 0.0
    #B_Float
    mix_013.inputs[3].default_value = 0.0
    #A_Color
    mix_013.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_013.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_013.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_013.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.059
    math_059 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_059.name = "Math.059"
    math_059.operation = 'MULTIPLY'
    math_059.use_clamp = False
    #Value_001
    math_059.inputs[1].default_value = -1.0
    #Value_002
    math_059.inputs[2].default_value = 0.5
    
    #node Math.147
    math_147 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_147.name = "Math.147"
    math_147.operation = 'LESS_THAN'
    math_147.use_clamp = False
    #Value_001
    math_147.inputs[1].default_value = 2.0
    #Value_002
    math_147.inputs[2].default_value = 0.5
    
    #node Math.148
    math_148 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_148.label = "r7.x (#265)"
    math_148.name = "Math.148"
    math_148.operation = 'SUBTRACT'
    math_148.use_clamp = False
    #Value
    math_148.inputs[0].default_value = 1.0
    #Value_002
    math_148.inputs[2].default_value = 0.5
    
    #node Mix.004
    mix_004 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_004.label = "r7.y (#267)"
    mix_004.name = "Mix.004"
    mix_004.blend_type = 'MIX'
    mix_004.clamp_factor = True
    mix_004.clamp_result = False
    mix_004.data_type = 'FLOAT'
    mix_004.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_004.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_004.inputs[2].default_value = 0.0
    #A_Vector
    mix_004.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_004.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_004.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_004.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_004.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_004.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.149
    math_149 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_149.label = "r7.y (#266)"
    math_149.name = "Math.149"
    math_149.operation = 'LESS_THAN'
    math_149.use_clamp = False
    #Value
    math_149.inputs[0].default_value = 0.0
    #Value_002
    math_149.inputs[2].default_value = 0.5
    
    #node Combine XYZ.043
    combine_xyz_043 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_043.label = "r14.zw"
    combine_xyz_043.name = "Combine XYZ.043"
    #Z
    combine_xyz_043.inputs[2].default_value = 0.0
    
    #node Vector Math.035
    vector_math_035 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_035.label = "r6.zw (#258)"
    vector_math_035.name = "Vector Math.035"
    vector_math_035.operation = 'MULTIPLY'
    #Vector
    vector_math_035.inputs[0].default_value = (-1.0, 1.0, 0.0)
    #Vector_002
    vector_math_035.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_035.inputs[3].default_value = 1.0
    
    #node Combine XYZ.044
    combine_xyz_044 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_044.label = "r6.zw (#260)"
    combine_xyz_044.name = "Combine XYZ.044"
    combine_xyz_044.inputs[2].hide = True
    #X
    combine_xyz_044.inputs[0].default_value = 0.0
    #Y
    combine_xyz_044.inputs[1].default_value = 0.0
    #Z
    combine_xyz_044.inputs[2].default_value = 0.0
    
    #node Math.230
    math_230 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_230.name = "Math.230"
    math_230.operation = 'ADD'
    math_230.use_clamp = False
    #Value
    math_230.inputs[0].default_value = -0.5
    #Value_002
    math_230.inputs[2].default_value = 0.5
    
    #node Math.150
    math_150 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_150.name = "Math.150"
    math_150.operation = 'COMPARE'
    math_150.use_clamp = False
    #Value_001
    math_150.inputs[1].default_value = 0.0
    #Value_002
    math_150.inputs[2].default_value = 0.0
    
    #node Math.152
    math_152 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_152.label = "r7.y (#269)"
    math_152.name = "Math.152"
    math_152.operation = 'MULTIPLY'
    math_152.use_clamp = False
    #Value_002
    math_152.inputs[2].default_value = 0.0
    
    #node Math.154
    math_154 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_154.label = "r6.x (#272)"
    math_154.name = "Math.154"
    math_154.operation = 'FLOORED_MODULO'
    math_154.use_clamp = False
    #Value_001
    math_154.inputs[1].default_value = 26.0
    #Value_002
    math_154.inputs[2].default_value = 0.5
    
    #node Math.155
    math_155 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_155.label = "r16.z (#273)"
    math_155.name = "Math.155"
    math_155.operation = 'FLOOR'
    math_155.use_clamp = False
    #Value_001
    math_155.inputs[1].default_value = 0.5
    #Value_002
    math_155.inputs[2].default_value = 0.5
    
    #node Vector Math.086
    vector_math_086 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_086.name = "Vector Math.086"
    vector_math_086.operation = 'FRACTION'
    #Vector_001
    vector_math_086.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_086.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_086.inputs[3].default_value = 1.0
    
    #node Vector Math.087
    vector_math_087 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_087.name = "Vector Math.087"
    vector_math_087.operation = 'ADD'
    #Vector_002
    vector_math_087.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_087.inputs[3].default_value = 1.0
    
    #node Vector Math.088
    vector_math_088 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_088.name = "Vector Math.088"
    vector_math_088.operation = 'MULTIPLY'
    #Vector_001
    vector_math_088.inputs[1].default_value = (1.0, 0.03846200183033943, 0.0)
    #Vector_002
    vector_math_088.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_088.inputs[3].default_value = 1.0
    
    #node Math.235
    math_235 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_235.name = "Math.235"
    math_235.operation = 'ADD'
    math_235.use_clamp = False
    #Value
    math_235.inputs[0].default_value = -0.5
    #Value_002
    math_235.inputs[2].default_value = 0.5
    
    #node Vector Math.085
    vector_math_085 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_085.name = "Vector Math.085"
    vector_math_085.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_085.inputs[3].default_value = 1.0
    
    #node Math.231
    math_231 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_231.name = "Math.231"
    math_231.operation = 'FLOOR'
    math_231.use_clamp = False
    #Value_001
    math_231.inputs[1].default_value = 0.5
    #Value_002
    math_231.inputs[2].default_value = 0.5
    
    #node Math.232
    math_232 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_232.name = "Math.232"
    math_232.operation = 'SUBTRACT'
    math_232.use_clamp = False
    #Value
    math_232.inputs[0].default_value = 25.0
    #Value_002
    math_232.inputs[2].default_value = 0.5
    
    #node Combine XYZ.040
    combine_xyz_040 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_040.name = "Combine XYZ.040"
    #X
    combine_xyz_040.inputs[0].default_value = 0.0
    #Z
    combine_xyz_040.inputs[2].default_value = 0.0
    
    #node Vector Math.034
    vector_math_034 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_034.label = "r11.yz (#277)"
    vector_math_034.name = "Vector Math.034"
    vector_math_034.operation = 'MULTIPLY'
    #Vector_002
    vector_math_034.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_034.inputs[3].default_value = 1.0
    
    #node Separate XYZ.061
    separate_xyz_061 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_061.name = "Separate XYZ.061"
    
    #node Separate XYZ.047
    separate_xyz_047 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_047.name = "Separate XYZ.047"
    
    #node Math.087
    math_087 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_087.label = "-r6.x"
    math_087.name = "Math.087"
    math_087.operation = 'MULTIPLY'
    math_087.use_clamp = False
    #Value_001
    math_087.inputs[1].default_value = -1.0
    #Value_002
    math_087.inputs[2].default_value = 0.5
    
    #node Math.156
    math_156 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_156.label = "r7.y (#278)"
    math_156.name = "Math.156"
    math_156.operation = 'MULTIPLY_ADD'
    math_156.use_clamp = False
    #Value_002
    math_156.inputs[2].default_value = 1.0
    
    #node Math.159
    math_159 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_159.label = "r11.y (#280)"
    math_159.name = "Math.159"
    math_159.operation = 'MAXIMUM'
    math_159.use_clamp = False
    #Value_002
    math_159.inputs[2].default_value = 0.5
    
    #node Math.157
    math_157 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_157.label = "r7.y (#279)"
    math_157.name = "Math.157"
    math_157.operation = 'MULTIPLY_ADD'
    math_157.use_clamp = False
    
    #node Math.160
    math_160 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_160.label = "r11.y (#281)"
    math_160.name = "Math.160"
    math_160.operation = 'MULTIPLY'
    math_160.use_clamp = False
    #Value
    math_160.inputs[0].default_value = 6.0999998822808266e-05
    #Value_002
    math_160.inputs[2].default_value = 0.5
    
    #node Math.161
    math_161 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_161.label = "r7.y (#282)"
    math_161.name = "Math.161"
    math_161.operation = 'MAXIMUM'
    math_161.use_clamp = False
    #Value_002
    math_161.inputs[2].default_value = 0.5
    
    #node Math.162
    math_162 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_162.label = "r7.y (#283)"
    math_162.name = "Math.162"
    math_162.operation = 'SQRT'
    math_162.use_clamp = False
    #Value_001
    math_162.inputs[1].default_value = 0.5
    #Value_002
    math_162.inputs[2].default_value = 0.5
    
    #node Math.163
    math_163 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_163.label = "-r7.yy"
    math_163.name = "Math.163"
    math_163.operation = 'MULTIPLY'
    math_163.use_clamp = False
    #Value_001
    math_163.inputs[1].default_value = -1.0
    #Value_002
    math_163.inputs[2].default_value = 0.5
    
    #node Vector Math.055
    vector_math_055 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_055.label = "r6.xy (#284)"
    vector_math_055.name = "Vector Math.055"
    vector_math_055.operation = 'MULTIPLY'
    #Vector_002
    vector_math_055.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_055.inputs[3].default_value = 1.0
    
    #node Vector Math.056
    vector_math_056 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_056.label = "r6.xy (#285)"
    vector_math_056.name = "Vector Math.056"
    vector_math_056.operation = 'MULTIPLY'
    #Vector_002
    vector_math_056.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_056.inputs[3].default_value = 1.0
    
    #node Separate XYZ.062
    separate_xyz_062 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_062.name = "Separate XYZ.062"
    
    #node Vector Math.057
    vector_math_057 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_057.label = "r11.yzw (#286)"
    vector_math_057.name = "Vector Math.057"
    vector_math_057.operation = 'MULTIPLY'
    #Vector_001
    vector_math_057.inputs[1].default_value = (1.0, 1.0, 1.0)
    #Vector_002
    vector_math_057.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_057.inputs[3].default_value = 1.0
    
    #node Math.158
    math_158 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_158.label = "-r6.x"
    math_158.name = "Math.158"
    math_158.operation = 'MULTIPLY'
    math_158.use_clamp = False
    #Value_001
    math_158.inputs[1].default_value = -1.0
    #Value_002
    math_158.inputs[2].default_value = 0.5
    
    #node Math.077
    math_077 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_077.name = "Math.077"
    math_077.operation = 'COMPARE'
    math_077.use_clamp = False
    #Value_001
    math_077.inputs[1].default_value = 0.0
    #Value_002
    math_077.inputs[2].default_value = 0.0
    
    #node Math.078
    math_078 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_078.label = "if (r15.y(#300) != 0) (#301)"
    math_078.name = "Math.078"
    math_078.operation = 'SUBTRACT'
    math_078.use_clamp = False
    #Value
    math_078.inputs[0].default_value = 1.0
    #Value_002
    math_078.inputs[2].default_value = 0.0
    
    #node Math.244
    math_244 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_244.label = "r15.y (#300)"
    math_244.name = "Math.244"
    math_244.operation = 'COMPARE'
    math_244.use_clamp = False
    #Value_001
    math_244.inputs[1].default_value = 4.0
    #Value_002
    math_244.inputs[2].default_value = 0.0
    
    #node Math.046
    math_046 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_046.label = "r7.y (#304)"
    math_046.name = "Math.046"
    math_046.operation = 'FLOOR'
    math_046.use_clamp = False
    #Value_001
    math_046.inputs[1].default_value = 0.5
    #Value_002
    math_046.inputs[2].default_value = 0.5
    
    #node Math.047
    math_047 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_047.label = "r4.w (#305)"
    math_047.name = "Math.047"
    math_047.operation = 'FLOORED_MODULO'
    math_047.use_clamp = False
    #Value_002
    math_047.inputs[2].default_value = 0.5
    
    #node Math.048
    math_048 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_048.label = "r17.z (#307)"
    math_048.name = "Math.048"
    math_048.operation = 'FLOOR'
    math_048.use_clamp = False
    #Value_001
    math_048.inputs[1].default_value = 0.5
    #Value_002
    math_048.inputs[2].default_value = 0.5
    
    #node Math.245
    math_245 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_245.label = "r15.z (#300)"
    math_245.name = "Math.245"
    math_245.operation = 'COMPARE'
    math_245.use_clamp = False
    #Value_001
    math_245.inputs[1].default_value = 1.0
    #Value_002
    math_245.inputs[2].default_value = 0.0
    
    #node Value.002
    value_002 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_002.label = "r4.w (#244) (amount of detail texs)"
    value_002.name = "Value.002"
    
    value_002.outputs[0].default_value = 2.0
    #node Math.246
    math_246 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_246.name = "Math.246"
    math_246.operation = 'FLOOR'
    math_246.use_clamp = False
    #Value_001
    math_246.inputs[1].default_value = 0.5
    #Value_002
    math_246.inputs[2].default_value = 0.5
    
    #node Math.247
    math_247 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_247.name = "Math.247"
    math_247.operation = 'SUBTRACT'
    math_247.use_clamp = False
    #Value
    math_247.inputs[0].default_value = 1.0
    #Value_002
    math_247.inputs[2].default_value = 0.5
    
    #node Combine XYZ.034
    combine_xyz_034 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_034.name = "Combine XYZ.034"
    #X
    combine_xyz_034.inputs[0].default_value = 0.0
    #Z
    combine_xyz_034.inputs[2].default_value = 0.0
    
    #node Separate XYZ.027
    separate_xyz_027 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_027.name = "Separate XYZ.027"
    
    #node Math.049
    math_049 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_049.label = "r14.w (#310)"
    math_049.name = "Math.049"
    math_049.operation = 'MULTIPLY'
    math_049.use_clamp = False
    #Value_001
    math_049.inputs[1].default_value = -1.0
    #Value_002
    math_049.inputs[2].default_value = 0.5
    
    #node Combine XYZ.027
    combine_xyz_027 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_027.label = "r17.xy (#311)"
    combine_xyz_027.name = "Combine XYZ.027"
    #Z
    combine_xyz_027.inputs[2].default_value = 0.0
    
    #node Math.044
    math_044 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_044.label = "r6.y (#302)"
    math_044.name = "Math.044"
    math_044.operation = 'MULTIPLY'
    math_044.use_clamp = False
    #Value_002
    math_044.inputs[2].default_value = 0.5
    
    #node Math.045
    math_045 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_045.label = "r6.y (#303)"
    math_045.name = "Math.045"
    math_045.operation = 'MULTIPLY'
    math_045.use_clamp = False
    #Value
    math_045.inputs[0].default_value = 4.0
    #Value_002
    math_045.inputs[2].default_value = 0.5
    
    #node Vector Math.090
    vector_math_090 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_090.name = "Vector Math.090"
    vector_math_090.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_090.inputs[3].default_value = 1.0
    
    #node Vector Math.095
    vector_math_095 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_095.name = "Vector Math.095"
    vector_math_095.operation = 'FRACTION'
    #Vector_001
    vector_math_095.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_095.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_095.inputs[3].default_value = 1.0
    
    #node Vector Math.100
    vector_math_100 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_100.name = "Vector Math.100"
    vector_math_100.operation = 'ADD'
    #Vector_002
    vector_math_100.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_100.inputs[3].default_value = 1.0
    
    #node Vector Math.101
    vector_math_101 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_101.name = "Vector Math.101"
    vector_math_101.operation = 'MULTIPLY'
    #Vector_001
    vector_math_101.inputs[1].default_value = (1.0, 0.5, 0.0)
    #Vector_002
    vector_math_101.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_101.inputs[3].default_value = 1.0
    
    #node Mix.012
    mix_012 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_012.name = "Mix.012"
    mix_012.blend_type = 'MIX'
    mix_012.clamp_factor = True
    mix_012.clamp_result = False
    mix_012.data_type = 'FLOAT'
    mix_012.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_012.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_012.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_012.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_012.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_012.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_012.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_012.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Value.003
    value_003 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_003.label = "r14.y (#312)"
    value_003.name = "Value.003"
    
    value_003.outputs[0].default_value = 0.0
    #node Combine XYZ.038
    combine_xyz_038 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_038.label = "r18.xy"
    combine_xyz_038.name = "Combine XYZ.038"
    #Z
    combine_xyz_038.inputs[2].default_value = 0.0
    
    #node Math.068
    math_068 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_068.label = "r6.y (#329)"
    math_068.name = "Math.068"
    math_068.operation = 'MULTIPLY'
    math_068.use_clamp = False
    #Value_002
    math_068.inputs[2].default_value = 0.5
    
    #node Math.069
    math_069 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_069.label = "r6.y (#330)"
    math_069.name = "Math.069"
    math_069.operation = 'MAXIMUM'
    math_069.use_clamp = False
    #Value
    math_069.inputs[0].default_value = 0.0
    #Value_002
    math_069.inputs[2].default_value = 0.5
    
    #node Math.070
    math_070 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_070.label = "r6.y (#331)"
    math_070.name = "Math.070"
    math_070.operation = 'MULTIPLY'
    math_070.use_clamp = False
    #Value
    math_070.inputs[0].default_value = 4.0
    #Value_002
    math_070.inputs[2].default_value = 0.5
    
    #node Vector Math.031
    vector_math_031 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_031.label = "r19.xy (#332)"
    vector_math_031.name = "Vector Math.031"
    vector_math_031.operation = 'MULTIPLY'
    #Vector_002
    vector_math_031.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_031.inputs[3].default_value = 1.0
    
    #node Mix.037
    mix_037 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_037.name = "Mix.037"
    mix_037.blend_type = 'MIX'
    mix_037.clamp_factor = True
    mix_037.clamp_result = False
    mix_037.data_type = 'RGBA'
    mix_037.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_037.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_037.inputs[2].default_value = 0.0
    #B_Float
    mix_037.inputs[3].default_value = 0.0
    #A_Vector
    mix_037.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_037.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_037.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_037.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.038
    mix_038 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_038.name = "Mix.038"
    mix_038.blend_type = 'MIX'
    mix_038.clamp_factor = True
    mix_038.clamp_result = False
    mix_038.data_type = 'FLOAT'
    mix_038.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_038.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_038.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_038.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_038.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_038.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_038.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_038.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Separate XYZ.044
    separate_xyz_044 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_044.name = "Separate XYZ.044"
    
    #node Combine XYZ.049
    combine_xyz_049 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_049.name = "Combine XYZ.049"
    #Z
    combine_xyz_049.inputs[2].default_value = 0.0
    
    #node Vector Math.103
    vector_math_103 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_103.label = "r18.xyz (#334)"
    vector_math_103.name = "Vector Math.103"
    vector_math_103.operation = 'ADD'
    #Vector_002
    vector_math_103.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_103.inputs[3].default_value = 1.0
    
    #node Math.251
    math_251 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_251.name = "Math.251"
    math_251.operation = 'ADD'
    math_251.use_clamp = False
    #Value
    math_251.inputs[0].default_value = 0.0
    #Value_002
    math_251.inputs[2].default_value = 0.5
    
    #node Math.252
    math_252 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_252.name = "Math.252"
    math_252.operation = 'MULTIPLY_ADD'
    math_252.use_clamp = False
    
    #node Math.253
    math_253 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_253.name = "Math.253"
    math_253.operation = 'MULTIPLY_ADD'
    math_253.use_clamp = False
    
    #node Math.254
    math_254 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_254.name = "Math.254"
    math_254.operation = 'MULTIPLY'
    math_254.use_clamp = False
    #Value_002
    math_254.inputs[2].default_value = 0.5
    
    #node Math.255
    math_255 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_255.name = "Math.255"
    math_255.operation = 'MULTIPLY_ADD'
    math_255.use_clamp = False
    
    #node Separate XYZ.042
    separate_xyz_042 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_042.name = "Separate XYZ.042"
    
    #node Math.076
    math_076 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_076.label = "r19.w (#338)"
    math_076.name = "Math.076"
    math_076.operation = 'MULTIPLY'
    math_076.use_clamp = False
    #Value_002
    math_076.inputs[2].default_value = 0.5
    
    #node Math.072
    math_072 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_072.label = "r7.y (#336)"
    math_072.name = "Math.072"
    math_072.operation = 'ADD'
    math_072.use_clamp = False
    #Value
    math_072.inputs[0].default_value = 1.0
    #Value_002
    math_072.inputs[2].default_value = 0.5
    
    #node Math.073
    math_073 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_073.label = "r6.y (#337)"
    math_073.name = "Math.073"
    math_073.operation = 'MULTIPLY_ADD'
    math_073.use_clamp = False
    
    #node Math.071
    math_071 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_071.name = "Math.071"
    math_071.operation = 'ADD'
    math_071.use_clamp = False
    #Value
    math_071.inputs[0].default_value = 0.5
    #Value_002
    math_071.inputs[2].default_value = 0.5
    
    #node Clamp.006
    clamp_006 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_006.label = "r6.y (#335)"
    clamp_006.name = "Clamp.006"
    clamp_006.clamp_type = 'MINMAX'
    #Min
    clamp_006.inputs[1].default_value = 0.0
    #Max
    clamp_006.inputs[2].default_value = 1.0
    
    #node Math.002
    math_002 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_002.label = "r7.y (#346)"
    math_002.name = "Math.002"
    math_002.operation = 'ADD'
    math_002.use_clamp = False
    #Value
    math_002.inputs[0].default_value = -1.0
    #Value_002
    math_002.inputs[2].default_value = 0.5
    
    #node Math.006
    math_006 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_006.name = "Math.006"
    math_006.operation = 'MULTIPLY'
    math_006.use_clamp = False
    #Value_001
    math_006.inputs[1].default_value = -1.0
    #Value_002
    math_006.inputs[2].default_value = 1.0
    
    #node Math.005
    math_005 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_005.label = "r7.z (#348)"
    math_005.name = "Math.005"
    math_005.operation = 'ADD'
    math_005.use_clamp = False
    #Value_002
    math_005.inputs[2].default_value = 1.0
    
    #node Math.004
    math_004 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_004.label = "r7.y (#347)"
    math_004.name = "Math.004"
    math_004.operation = 'MULTIPLY_ADD'
    math_004.use_clamp = False
    #Value_002
    math_004.inputs[2].default_value = 1.0
    
    #node Math.007
    math_007 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_007.label = "r7.y (#349)"
    math_007.name = "Math.007"
    math_007.operation = 'MULTIPLY_ADD'
    math_007.use_clamp = False
    
    #node Clamp.007
    clamp_007 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_007.label = "r6.y (#341)"
    clamp_007.name = "Clamp.007"
    clamp_007.clamp_type = 'MINMAX'
    #Min
    clamp_007.inputs[1].default_value = 0.0
    #Max
    clamp_007.inputs[2].default_value = 1.0
    
    #node Math.088
    math_088 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_088.name = "Math.088"
    math_088.operation = 'ADD'
    math_088.use_clamp = False
    #Value_002
    math_088.inputs[2].default_value = 0.5
    
    #node Separate XYZ.046
    separate_xyz_046 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_046.name = "Separate XYZ.046"
    
    #node Math.008
    math_008 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_008.name = "Math.008"
    math_008.operation = 'COMPARE'
    math_008.use_clamp = False
    #Value_001
    math_008.inputs[1].default_value = 0.0
    #Value_002
    math_008.inputs[2].default_value = 0.0
    
    #node Vector Math.004
    vector_math_004 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_004.label = "r7.zw (#352)"
    vector_math_004.name = "Vector Math.004"
    vector_math_004.operation = 'MULTIPLY'
    #Vector_002
    vector_math_004.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_004.inputs[3].default_value = 1.0
    
    #node Separate XYZ.004
    separate_xyz_004 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_004.name = "Separate XYZ.004"
    
    #node Separate XYZ.001
    separate_xyz_001 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_001.name = "Separate XYZ.001"
    
    #node Math.011
    math_011 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_011.name = "Math.011"
    math_011.operation = 'MULTIPLY'
    math_011.use_clamp = False
    #Value_001
    math_011.inputs[1].default_value = -1.0
    #Value_002
    math_011.inputs[2].default_value = 0.5
    
    #node Math.010
    math_010 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_010.label = "r5.x (#353)"
    math_010.name = "Math.010"
    math_010.operation = 'MULTIPLY_ADD'
    math_010.use_clamp = False
    #Value_002
    math_010.inputs[2].default_value = 1.0
    
    #node Math.013
    math_013 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_013.name = "Math.013"
    math_013.operation = 'MULTIPLY'
    math_013.use_clamp = False
    #Value_001
    math_013.inputs[1].default_value = -1.0
    #Value_002
    math_013.inputs[2].default_value = 0.5
    
    #node Math.012
    math_012 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_012.label = "r5.x (#354)"
    math_012.name = "Math.012"
    math_012.operation = 'MULTIPLY_ADD'
    math_012.use_clamp = False
    
    #node Math.016
    math_016 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_016.label = "r5.x (#357)"
    math_016.name = "Math.016"
    math_016.operation = 'MAXIMUM'
    math_016.use_clamp = False
    #Value_002
    math_016.inputs[2].default_value = 0.5
    
    #node Math.017
    math_017 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_017.label = "r5.x (#358)"
    math_017.name = "Math.017"
    math_017.operation = 'SQRT'
    math_017.use_clamp = False
    #Value_001
    math_017.inputs[1].default_value = 0.5
    #Value_002
    math_017.inputs[2].default_value = 0.5
    
    #node Math.018
    math_018 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_018.label = "-r5.xx"
    math_018.name = "Math.018"
    math_018.operation = 'MULTIPLY'
    math_018.use_clamp = False
    #Value_001
    math_018.inputs[1].default_value = -1.0
    #Value_002
    math_018.inputs[2].default_value = 0.5
    
    #node Math.015
    math_015 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_015.label = "r7.z (#356)"
    math_015.name = "Math.015"
    math_015.operation = 'MULTIPLY'
    math_015.use_clamp = False
    #Value
    math_015.inputs[0].default_value = 6.0999998822808266e-05
    #Value_002
    math_015.inputs[2].default_value = 0.5
    
    #node Math.014
    math_014 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_014.label = "r7.z (#355)"
    math_014.name = "Math.014"
    math_014.operation = 'MAXIMUM'
    math_014.use_clamp = False
    #Value_002
    math_014.inputs[2].default_value = 0.5
    
    #node Vector Math.006
    vector_math_006 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_006.label = "r6.zw (#360)"
    vector_math_006.name = "Vector Math.006"
    vector_math_006.operation = 'MULTIPLY'
    #Vector_002
    vector_math_006.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_006.inputs[3].default_value = 1.0
    
    #node Separate XYZ.006
    separate_xyz_006 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_006.name = "Separate XYZ.006"
    
    #node Vector Math.005
    vector_math_005 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_005.label = "r6.zw (#359)"
    vector_math_005.name = "Vector Math.005"
    vector_math_005.operation = 'MULTIPLY'
    #Vector_002
    vector_math_005.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_005.inputs[3].default_value = 1.0
    
    #node Math.009
    math_009 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_009.label = "if (r5.x(#349) != 0) (#350)"
    math_009.name = "Math.009"
    math_009.operation = 'SUBTRACT'
    math_009.use_clamp = False
    #Value
    math_009.inputs[0].default_value = 1.0
    #Value_002
    math_009.inputs[2].default_value = 0.5
    
    #node Vector Math.008
    vector_math_008 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_008.label = "r12.xzw (#362)"
    vector_math_008.name = "Vector Math.008"
    vector_math_008.operation = 'MULTIPLY'
    #Vector_001
    vector_math_008.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_008.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_008.inputs[3].default_value = 1.0
    
    #node Math.164
    math_164 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_164.name = "Math.164"
    math_164.operation = 'COMPARE'
    math_164.use_clamp = False
    #Value_001
    math_164.inputs[1].default_value = 0.0
    #Value_002
    math_164.inputs[2].default_value = 0.0
    
    #node Math.165
    math_165 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_165.label = "if (r15.y != 0) (#368)"
    math_165.name = "Math.165"
    math_165.operation = 'SUBTRACT'
    math_165.use_clamp = False
    #Value
    math_165.inputs[0].default_value = 1.0
    #Value_002
    math_165.inputs[2].default_value = 0.5
    
    #node Math.169
    math_169 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_169.label = "r5.x (#372)"
    math_169.name = "Math.169"
    math_169.operation = 'MULTIPLY_ADD'
    math_169.use_clamp = False
    
    #node Math.170
    math_170 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_170.label = "r7.z (#373)"
    math_170.name = "Math.170"
    math_170.operation = 'MAXIMUM'
    math_170.use_clamp = False
    #Value_002
    math_170.inputs[2].default_value = 0.5
    
    #node Math.171
    math_171 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_171.label = "r7.z (#374)"
    math_171.name = "Math.171"
    math_171.operation = 'MULTIPLY'
    math_171.use_clamp = False
    #Value
    math_171.inputs[0].default_value = 6.0999998822808266e-05
    #Value_002
    math_171.inputs[2].default_value = 0.5
    
    #node Math.172
    math_172 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_172.label = "r5.x (#375)"
    math_172.name = "Math.172"
    math_172.operation = 'MAXIMUM'
    math_172.use_clamp = False
    #Value_002
    math_172.inputs[2].default_value = 0.5
    
    #node Separate XYZ.065
    separate_xyz_065 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_065.name = "Separate XYZ.065"
    
    #node Math.168
    math_168 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_168.name = "Math.168"
    math_168.operation = 'MULTIPLY'
    math_168.use_clamp = False
    #Value_001
    math_168.inputs[1].default_value = -1.0
    #Value_002
    math_168.inputs[2].default_value = 0.5
    
    #node Vector Math.059
    vector_math_059 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_059.label = "r7.zw (#370)"
    vector_math_059.name = "Vector Math.059"
    vector_math_059.operation = 'MULTIPLY'
    #Vector_002
    vector_math_059.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_059.inputs[3].default_value = 1.0
    
    #node Vector Math.062
    vector_math_062 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_062.label = "r6.zw (#379)"
    vector_math_062.name = "Vector Math.062"
    vector_math_062.operation = 'MULTIPLY'
    #Vector_002
    vector_math_062.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_062.inputs[3].default_value = 1.0
    
    #node Vector Math.029
    vector_math_029 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_029.label = "r7.yzw (#403)"
    vector_math_029.name = "Vector Math.029"
    vector_math_029.operation = 'ADD'
    #Vector_002
    vector_math_029.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_029.inputs[3].default_value = 1.0
    
    #node Vector Math.030
    vector_math_030 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_030.label = "r5.x (#404)"
    vector_math_030.name = "Vector Math.030"
    vector_math_030.operation = 'DOT_PRODUCT'
    #Vector_002
    vector_math_030.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_030.inputs[3].default_value = 1.0
    
    #node Math.094
    math_094 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_094.label = "r5.x (#405)"
    math_094.name = "Math.094"
    math_094.operation = 'ADD'
    math_094.use_clamp = False
    #Value_002
    math_094.inputs[2].default_value = 0.5
    
    #node Separate XYZ.039
    separate_xyz_039 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_039.name = "Separate XYZ.039"
    
    #node Math.064
    math_064 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_064.label = "r5.x (#398)"
    math_064.name = "Math.064"
    math_064.operation = 'MULTIPLY'
    math_064.use_clamp = False
    #Value_002
    math_064.inputs[2].default_value = 0.5
    
    #node Math.065
    math_065 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_065.label = "r5.x (#399)"
    math_065.name = "Math.065"
    math_065.operation = 'MAXIMUM'
    math_065.use_clamp = False
    #Value
    math_065.inputs[0].default_value = 0.0
    #Value_002
    math_065.inputs[2].default_value = 0.5
    
    #node Math.066
    math_066 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_066.label = "r5.x (#400)"
    math_066.name = "Math.066"
    math_066.operation = 'MULTIPLY'
    math_066.use_clamp = False
    #Value
    math_066.inputs[0].default_value = 4.0
    #Value_002
    math_066.inputs[2].default_value = 0.5
    
    #node Combine XYZ.035
    combine_xyz_035 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_035.label = "r14.xy"
    combine_xyz_035.name = "Combine XYZ.035"
    #Z
    combine_xyz_035.inputs[2].default_value = 0.0
    
    #node Combine XYZ.036
    combine_xyz_036 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_036.label = "r5.xx"
    combine_xyz_036.name = "Combine XYZ.036"
    #Z
    combine_xyz_036.inputs[2].default_value = 0.0
    
    #node Vector Math.028
    vector_math_028 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_028.label = "r21.xy (#401)"
    vector_math_028.name = "Vector Math.028"
    vector_math_028.operation = 'MULTIPLY'
    #Vector_002
    vector_math_028.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_028.inputs[3].default_value = 1.0
    
    #node Value.005
    value_005 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_005.label = "r21.z (#402)"
    value_005.name = "Value.005"
    
    value_005.outputs[0].default_value = 0.0
    #node Combine XYZ.037
    combine_xyz_037 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_037.name = "Combine XYZ.037"
    
    #node Math.095
    math_095 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_095.label = "r6.z (#408)"
    math_095.name = "Math.095"
    math_095.operation = 'MULTIPLY'
    math_095.use_clamp = False
    #Value_002
    math_095.inputs[2].default_value = 0.5
    
    #node Math.096
    math_096 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_096.label = "r6.z (#409)"
    math_096.name = "Math.096"
    math_096.operation = 'MAXIMUM'
    math_096.use_clamp = False
    #Value
    math_096.inputs[0].default_value = 0.0
    #Value_002
    math_096.inputs[2].default_value = 0.5
    
    #node Math.097
    math_097 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_097.label = "r6.z (#410)"
    math_097.name = "Math.097"
    math_097.operation = 'MULTIPLY'
    math_097.use_clamp = False
    #Value
    math_097.inputs[0].default_value = 4.0
    #Value_002
    math_097.inputs[2].default_value = 0.5
    
    #node Combine XYZ.007
    combine_xyz_007 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_007.name = "Combine XYZ.007"
    #Z
    combine_xyz_007.inputs[2].default_value = 0.0
    
    #node Vector Math.045
    vector_math_045 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_045.label = "r21.xy (#411)"
    vector_math_045.name = "Vector Math.045"
    vector_math_045.operation = 'MULTIPLY'
    #Vector_002
    vector_math_045.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_045.inputs[3].default_value = 1.0
    
    #node Separate XYZ.010
    separate_xyz_010 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_010.name = "Separate XYZ.010"
    
    #node Combine XYZ.050
    combine_xyz_050 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_050.name = "Combine XYZ.050"
    #Z
    combine_xyz_050.inputs[2].default_value = 0.0
    
    #node Math.261
    math_261 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_261.label = "r14.w (#413)"
    math_261.name = "Math.261"
    math_261.operation = 'ADD'
    math_261.use_clamp = False
    #Value
    math_261.inputs[0].default_value = 0.0
    #Value_002
    math_261.inputs[2].default_value = 0.5
    
    #node Vector Math.105
    vector_math_105 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_105.label = "r14.xyz (#413)"
    vector_math_105.name = "Vector Math.105"
    vector_math_105.operation = 'ADD'
    #Vector_002
    vector_math_105.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_105.inputs[3].default_value = 1.0
    
    #node Separate XYZ.011
    separate_xyz_011 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_011.name = "Separate XYZ.011"
    
    #node Separate XYZ.055
    separate_xyz_055 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_055.name = "Separate XYZ.055"
    
    #node Combine XYZ.009
    combine_xyz_009 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_009.label = "r10.xy (#414)"
    combine_xyz_009.name = "Combine XYZ.009"
    
    #node Separate XYZ.012
    separate_xyz_012 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_012.name = "Separate XYZ.012"
    
    #node Math.264
    math_264 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_264.name = "Math.264"
    math_264.operation = 'MULTIPLY'
    math_264.use_clamp = False
    #Value_002
    math_264.inputs[2].default_value = 0.5
    
    #node Math.263
    math_263 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_263.name = "Math.263"
    math_263.operation = 'MULTIPLY_ADD'
    math_263.use_clamp = False
    
    #node Math.262
    math_262 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_262.name = "Math.262"
    math_262.operation = 'MULTIPLY_ADD'
    math_262.use_clamp = False
    
    #node Math.265
    math_265 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_265.name = "Math.265"
    math_265.operation = 'MULTIPLY_ADD'
    math_265.use_clamp = False
    
    #node Math.098
    math_098 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_098.name = "Math.098"
    math_098.operation = 'ADD'
    math_098.use_clamp = False
    #Value_002
    math_098.inputs[2].default_value = 0.5
    
    #node Clamp.009
    clamp_009 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_009.label = "r6.z (#416)"
    clamp_009.name = "Clamp.009"
    clamp_009.clamp_type = 'MINMAX'
    #Min
    clamp_009.inputs[1].default_value = 0.0
    #Max
    clamp_009.inputs[2].default_value = 1.0
    
    #node Clamp.010
    clamp_010 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_010.label = "r16.y (#426)"
    clamp_010.name = "Clamp.010"
    clamp_010.clamp_type = 'MINMAX'
    #Min
    clamp_010.inputs[1].default_value = 0.0
    #Max
    clamp_010.inputs[2].default_value = 1.0
    
    #node Vector Math.048
    vector_math_048 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_048.label = "r6.w (#425)"
    vector_math_048.name = "Vector Math.048"
    vector_math_048.operation = 'DOT_PRODUCT'
    #Vector_002
    vector_math_048.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_048.inputs[3].default_value = 1.0
    
    #node Math.102
    math_102 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_102.name = "Math.102"
    math_102.operation = 'ADD'
    math_102.use_clamp = False
    #Value_002
    math_102.inputs[2].default_value = 0.5
    
    #node Combine XYZ.012
    combine_xyz_012 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_012.name = "Combine XYZ.012"
    
    #node Vector Math.047
    vector_math_047 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_047.label = "r7.yzw (#424)"
    vector_math_047.name = "Vector Math.047"
    vector_math_047.operation = 'ADD'
    #Vector_002
    vector_math_047.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_047.inputs[3].default_value = 1.0
    
    #node Math.100
    math_100 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_100.label = "r6.w (#420)"
    math_100.name = "Math.100"
    math_100.operation = 'MAXIMUM'
    math_100.use_clamp = False
    #Value
    math_100.inputs[0].default_value = 0.0
    #Value_002
    math_100.inputs[2].default_value = 0.5
    
    #node Math.099
    math_099 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_099.label = "r6.w (#419)"
    math_099.name = "Math.099"
    math_099.operation = 'MULTIPLY'
    math_099.use_clamp = False
    #Value_002
    math_099.inputs[2].default_value = 0.5
    
    #node Math.101
    math_101 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_101.label = "r6.w (#421)"
    math_101.name = "Math.101"
    math_101.operation = 'MULTIPLY'
    math_101.use_clamp = False
    #Value
    math_101.inputs[0].default_value = 4.0
    #Value_002
    math_101.inputs[2].default_value = 0.5
    
    #node Combine XYZ.011
    combine_xyz_011 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_011.label = "r14.xy"
    combine_xyz_011.name = "Combine XYZ.011"
    #Z
    combine_xyz_011.inputs[2].default_value = 0.0
    
    #node Vector Math.046
    vector_math_046 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_046.label = "r21.xy (#422)"
    vector_math_046.name = "Vector Math.046"
    vector_math_046.operation = 'MULTIPLY'
    #Vector_002
    vector_math_046.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_046.inputs[3].default_value = 1.0
    
    #node Separate XYZ.013
    separate_xyz_013 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_013.name = "Separate XYZ.013"
    
    #node Separate XYZ.050
    separate_xyz_050 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_050.name = "Separate XYZ.050"
    
    #node Vector Math.051
    vector_math_051 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_051.label = "r6.w (#435)"
    vector_math_051.name = "Vector Math.051"
    vector_math_051.operation = 'DOT_PRODUCT'
    #Vector_002
    vector_math_051.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_051.inputs[3].default_value = 1.0
    
    #node Clamp.011
    clamp_011 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_011.label = "r14.w (#436)"
    clamp_011.name = "Clamp.011"
    clamp_011.clamp_type = 'MINMAX'
    #Min
    clamp_011.inputs[1].default_value = 0.0
    #Max
    clamp_011.inputs[2].default_value = 1.0
    
    #node Math.106
    math_106 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_106.name = "Math.106"
    math_106.operation = 'ADD'
    math_106.use_clamp = False
    #Value_002
    math_106.inputs[2].default_value = 0.5
    
    #node Combine XYZ.014
    combine_xyz_014 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_014.name = "Combine XYZ.014"
    
    #node Vector Math.050
    vector_math_050 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_050.label = "r7.yzw (#434)"
    vector_math_050.name = "Vector Math.050"
    vector_math_050.operation = 'ADD'
    #Vector_002
    vector_math_050.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_050.inputs[3].default_value = 1.0
    
    #node Math.103
    math_103 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_103.label = "r6.w (#429)"
    math_103.name = "Math.103"
    math_103.operation = 'MULTIPLY'
    math_103.use_clamp = False
    #Value_002
    math_103.inputs[2].default_value = 0.5
    
    #node Math.104
    math_104 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_104.label = "r6.w (#430)"
    math_104.name = "Math.104"
    math_104.operation = 'MAXIMUM'
    math_104.use_clamp = False
    #Value
    math_104.inputs[0].default_value = 0.0
    #Value_002
    math_104.inputs[2].default_value = 0.5
    
    #node Combine XYZ.013
    combine_xyz_013 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_013.name = "Combine XYZ.013"
    #Z
    combine_xyz_013.inputs[2].default_value = 0.0
    
    #node Math.105
    math_105 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_105.label = "r6.w (#431)"
    math_105.name = "Math.105"
    math_105.operation = 'MULTIPLY'
    math_105.use_clamp = False
    #Value
    math_105.inputs[0].default_value = 4.0
    #Value_002
    math_105.inputs[2].default_value = 0.5
    
    #node Vector Math.049
    vector_math_049 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_049.label = "r21.xy (#432)"
    vector_math_049.name = "Vector Math.049"
    vector_math_049.operation = 'MULTIPLY'
    #Vector_002
    vector_math_049.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_049.inputs[3].default_value = 1.0
    
    #node Separate XYZ.014
    separate_xyz_014 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_014.name = "Separate XYZ.014"
    
    #node Value.009
    value_009 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_009.label = "r21.z (#423)"
    value_009.name = "Value.009"
    
    value_009.outputs[0].default_value = 0.0
    #node Combine XYZ.015
    combine_xyz_015 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_015.label = "r14.xy (#439)"
    combine_xyz_015.name = "Combine XYZ.015"
    #Z
    combine_xyz_015.inputs[2].default_value = 0.0
    
    #node Separate XYZ.053
    separate_xyz_053 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_053.name = "Separate XYZ.053"
    
    #node Separate XYZ.015
    separate_xyz_015 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_015.name = "Separate XYZ.015"
    
    #node Clamp.013
    clamp_013 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_013.name = "Clamp.013"
    clamp_013.clamp_type = 'MINMAX'
    #Min
    clamp_013.inputs[1].default_value = 0.0
    #Max
    clamp_013.inputs[2].default_value = 1.0
    
    #node Clamp.012
    clamp_012 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_012.name = "Clamp.012"
    clamp_012.clamp_type = 'MINMAX'
    #Min
    clamp_012.inputs[1].default_value = 0.0
    #Max
    clamp_012.inputs[2].default_value = 1.0
    
    #node Value.010
    value_010 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_010.label = "r21.z (#433)"
    value_010.name = "Value.010"
    
    value_010.outputs[0].default_value = 0.0
    #node Vector Math.106
    vector_math_106 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_106.name = "Vector Math.106"
    vector_math_106.operation = 'ADD'
    #Vector_002
    vector_math_106.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_106.inputs[3].default_value = 1.0
    
    #node Vector Math.107
    vector_math_107 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_107.name = "Vector Math.107"
    vector_math_107.operation = 'MULTIPLY'
    #Vector_001
    vector_math_107.inputs[1].default_value = (1.0, 0.20000000298023224, 0.0)
    #Vector_002
    vector_math_107.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_107.inputs[3].default_value = 1.0
    
    #node Combine XYZ.010
    combine_xyz_010 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_010.name = "Combine XYZ.010"
    #Z
    combine_xyz_010.inputs[2].default_value = 0.0
    
    #node Clamp.020
    clamp_020 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_020.name = "Clamp.020"
    clamp_020.clamp_type = 'MINMAX'
    #Min
    clamp_020.inputs[1].default_value = 0.004999999888241291
    #Max
    clamp_020.inputs[2].default_value = 0.9950000047683716
    
    #node Separate XYZ.002
    separate_xyz_002 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_002.name = "Separate XYZ.002"
    
    #node Vector Math.108
    vector_math_108 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_108.name = "Vector Math.108"
    vector_math_108.operation = 'FRACTION'
    #Vector_001
    vector_math_108.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_108.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_108.inputs[3].default_value = 1.0
    
    #node Vector Math.109
    vector_math_109 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_109.name = "Vector Math.109"
    vector_math_109.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_109.inputs[3].default_value = 1.0
    
    #node Math.019
    math_019 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_019.label = "r7.y (#455)"
    math_019.name = "Math.019"
    math_019.operation = 'MULTIPLY'
    math_019.use_clamp = False
    #Value_002
    math_019.inputs[2].default_value = 0.5
    
    #node Math.266
    math_266 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_266.name = "Math.266"
    math_266.operation = 'FLOOR'
    math_266.use_clamp = False
    #Value_001
    math_266.inputs[1].default_value = 0.5
    #Value_002
    math_266.inputs[2].default_value = 0.5
    
    #node Math.267
    math_267 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_267.name = "Math.267"
    math_267.operation = 'SUBTRACT'
    math_267.use_clamp = False
    #Value
    math_267.inputs[0].default_value = 4.0
    #Value_002
    math_267.inputs[2].default_value = 0.5
    
    #node Vector Math.009
    vector_math_009 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_009.label = "r7.yzw (#459)"
    vector_math_009.name = "Vector Math.009"
    vector_math_009.operation = 'ADD'
    #Vector
    vector_math_009.inputs[0].default_value = (-0.5, -0.5, -0.5)
    #Vector_002
    vector_math_009.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_009.inputs[3].default_value = 1.0
    
    #node Vector Math.010
    vector_math_010 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_010.name = "Vector Math.010"
    vector_math_010.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_010.inputs[3].default_value = 1.0
    
    #node Separate XYZ.007
    separate_xyz_007 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_007.name = "Separate XYZ.007"
    
    #node Clamp.001
    clamp_001 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_001.name = "Clamp.001"
    clamp_001.clamp_type = 'MINMAX'
    #Min
    clamp_001.inputs[1].default_value = 0.0
    #Max
    clamp_001.inputs[2].default_value = 1.0
    
    #node Clamp.002
    clamp_002 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_002.name = "Clamp.002"
    clamp_002.clamp_type = 'MINMAX'
    #Min
    clamp_002.inputs[1].default_value = 0.0
    #Max
    clamp_002.inputs[2].default_value = 1.0
    
    #node Clamp
    clamp = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp.name = "Clamp"
    clamp.clamp_type = 'MINMAX'
    #Min
    clamp.inputs[1].default_value = 0.0
    #Max
    clamp.inputs[2].default_value = 1.0
    
    #node Combine XYZ.001
    combine_xyz_001 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_001.label = "r7.yzw (#460)"
    combine_xyz_001.name = "Combine XYZ.001"
    
    #node Separate XYZ.008
    separate_xyz_008 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_008.name = "Separate XYZ.008"
    
    #node Vector Math.012
    vector_math_012 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_012.name = "Vector Math.012"
    vector_math_012.operation = 'MULTIPLY'
    #Vector_001
    vector_math_012.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_012.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_012.inputs[3].default_value = 1.0
    
    #node Vector Math.011
    vector_math_011 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_011.label = "r10.xyw (#461)"
    vector_math_011.name = "Vector Math.011"
    vector_math_011.operation = 'ADD'
    #Vector_002
    vector_math_011.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_011.inputs[3].default_value = 1.0
    
    #node Vector Math.013
    vector_math_013 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_013.label = "r10.xyw (#462)"
    vector_math_013.name = "Vector Math.013"
    vector_math_013.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_013.inputs[3].default_value = 1.0
    
    #node Vector Math.014
    vector_math_014 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_014.label = "-r10.xyw"
    vector_math_014.name = "Vector Math.014"
    vector_math_014.operation = 'MULTIPLY'
    #Vector_001
    vector_math_014.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_014.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_014.inputs[3].default_value = 1.0
    
    #node Vector Math.037
    vector_math_037 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_037.label = "r22.xyz (#465)"
    vector_math_037.name = "Vector Math.037"
    vector_math_037.operation = 'ADD'
    #Vector_002
    vector_math_037.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_037.inputs[3].default_value = 1.0
    
    #node Vector Math.038
    vector_math_038 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_038.label = "r10.xyw (#464)"
    vector_math_038.name = "Vector Math.038"
    vector_math_038.operation = 'MULTIPLY'
    #Vector_001
    vector_math_038.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_038.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_038.inputs[3].default_value = 1.0
    
    #node Vector Math.039
    vector_math_039 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_039.label = "r7.yzw (#466)"
    vector_math_039.name = "Vector Math.039"
    vector_math_039.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_039.inputs[3].default_value = 1.0
    
    #node Vector Math.015
    vector_math_015 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_015.label = "r22.xyz (#463)"
    vector_math_015.name = "Vector Math.015"
    vector_math_015.operation = 'ADD'
    #Vector_002
    vector_math_015.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_015.inputs[3].default_value = 1.0
    
    #node Vector Math.036
    vector_math_036 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_036.label = "r10.xyw (#464)"
    vector_math_036.name = "Vector Math.036"
    vector_math_036.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_036.inputs[3].default_value = 1.0
    
    #node Combine XYZ.051
    combine_xyz_051 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_051.name = "Combine XYZ.051"
    #X
    combine_xyz_051.inputs[0].default_value = 0.0
    #Z
    combine_xyz_051.inputs[2].default_value = 0.0
    
    #node Value.007
    value_007 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_007.label = "r6.w (#491)"
    value_007.name = "Value.007"
    
    value_007.outputs[0].default_value = 1.0
    #node Mix.003
    mix_003 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_003.label = "r6.w"
    mix_003.name = "Mix.003"
    mix_003.blend_type = 'MIX'
    mix_003.clamp_factor = True
    mix_003.clamp_result = False
    mix_003.data_type = 'FLOAT'
    mix_003.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_003.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_003.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_003.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_003.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_003.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_003.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_003.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.107
    math_107 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_107.name = "Math.107"
    math_107.operation = 'MULTIPLY'
    math_107.use_clamp = False
    #Value_001
    math_107.inputs[1].default_value = -1.0
    #Value_002
    math_107.inputs[2].default_value = 0.5
    
    #node Math.092
    math_092 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_092.label = "r10.x (#485)"
    math_092.name = "Math.092"
    math_092.operation = 'MULTIPLY'
    math_092.use_clamp = False
    #Value_002
    math_092.inputs[2].default_value = 0.5
    
    #node Math.093
    math_093 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_093.label = "r10.y (#486)"
    math_093.name = "Math.093"
    math_093.operation = 'ADD'
    math_093.use_clamp = False
    #Value_002
    math_093.inputs[2].default_value = 0.5
    
    #node Vector Math.043
    vector_math_043 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_043.label = "-r7.yzw"
    vector_math_043.name = "Vector Math.043"
    vector_math_043.operation = 'MULTIPLY'
    #Vector_001
    vector_math_043.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_043.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_043.inputs[3].default_value = 1.0
    
    #node Vector Math.044
    vector_math_044 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_044.label = "r8.xyz (#481)"
    vector_math_044.name = "Vector Math.044"
    vector_math_044.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_044.inputs[3].default_value = 1.0
    
    #node Vector Math.042
    vector_math_042 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_042.label = "r10.xyw (#480)"
    vector_math_042.name = "Vector Math.042"
    vector_math_042.operation = 'ADD'
    #Vector_002
    vector_math_042.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_042.inputs[3].default_value = 1.0
    
    #node Math.108
    math_108 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_108.label = "r6.z (#487)"
    math_108.name = "Math.108"
    math_108.operation = 'MULTIPLY_ADD'
    math_108.use_clamp = False
    
    #node Mix.007
    mix_007 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_007.label = "r8.xyz (#490)"
    mix_007.name = "Mix.007"
    mix_007.blend_type = 'MIX'
    mix_007.clamp_factor = True
    mix_007.clamp_result = False
    mix_007.data_type = 'VECTOR'
    mix_007.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_007.inputs[1].default_value = (0.0, 0.0, 0.0)
    #A_Float
    mix_007.inputs[2].default_value = 0.0
    #B_Float
    mix_007.inputs[3].default_value = 0.0
    #A_Color
    mix_007.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_007.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_007.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_007.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.109
    math_109 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_109.label = "r16.y (#488)"
    math_109.name = "Math.109"
    math_109.operation = 'MULTIPLY'
    math_109.use_clamp = False
    #Value_002
    math_109.inputs[2].default_value = 0.5
    
    #node Mix.002
    mix_002 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_002.name = "Mix.002"
    mix_002.blend_type = 'MIX'
    mix_002.clamp_factor = True
    mix_002.clamp_result = False
    mix_002.data_type = 'FLOAT'
    mix_002.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_002.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_002.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_002.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_002.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_002.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_002.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_002.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Value.011
    value_011 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_011.label = "r6.w (#494)"
    value_011.name = "Value.011"
    
    value_011.outputs[0].default_value = 1.0
    #node Mix.009
    mix_009 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_009.label = "r6.z if (r10.x != 0)"
    mix_009.name = "Mix.009"
    mix_009.blend_type = 'MIX'
    mix_009.clamp_factor = True
    mix_009.clamp_result = False
    mix_009.data_type = 'FLOAT'
    mix_009.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_009.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_009.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_009.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_009.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_009.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_009.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_009.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.010
    mix_010 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_010.label = "r16.y if (r10.x != 0)"
    mix_010.name = "Mix.010"
    mix_010.blend_type = 'MIX'
    mix_010.clamp_factor = True
    mix_010.clamp_result = False
    mix_010.data_type = 'FLOAT'
    mix_010.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_010.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_010.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_010.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_010.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_010.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_010.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_010.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.001
    mix_001 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_001.name = "Mix.001"
    mix_001.blend_type = 'MIX'
    mix_001.clamp_factor = True
    mix_001.clamp_result = False
    mix_001.data_type = 'VECTOR'
    mix_001.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_001.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_001.inputs[2].default_value = 0.0
    #B_Float
    mix_001.inputs[3].default_value = 0.0
    #A_Color
    mix_001.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_001.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_001.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_001.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.114
    math_114 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_114.label = "r19.z (#514)"
    math_114.name = "Math.114"
    math_114.operation = 'FLOOR'
    math_114.use_clamp = False
    #Value_001
    math_114.inputs[1].default_value = 0.5
    #Value_002
    math_114.inputs[2].default_value = 0.5
    
    #node Math.111
    math_111 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_111.label = "r7.w (#512)"
    math_111.name = "Math.111"
    math_111.operation = 'FLOOR'
    math_111.use_clamp = False
    #Value_001
    math_111.inputs[1].default_value = 0.5
    #Value_002
    math_111.inputs[2].default_value = 0.5
    
    #node Math.112
    math_112 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_112.label = "r7.y (#513)"
    math_112.name = "Math.112"
    math_112.operation = 'FLOORED_MODULO'
    math_112.use_clamp = False
    #Value_002
    math_112.inputs[2].default_value = 0.5
    
    #node Math.113
    math_113 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_113.name = "Math.113"
    math_113.operation = 'FLOOR'
    math_113.use_clamp = False
    #Value_001
    math_113.inputs[1].default_value = 0.5
    #Value_002
    math_113.inputs[2].default_value = 0.5
    
    #node Vector Math.053
    vector_math_053 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_053.label = "-r8.xyz"
    vector_math_053.name = "Vector Math.053"
    vector_math_053.operation = 'MULTIPLY'
    #Vector_001
    vector_math_053.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_053.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_053.inputs[3].default_value = 1.0
    
    #node Vector Math.052
    vector_math_052 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_052.label = "r10.xyw (#522)"
    vector_math_052.name = "Vector Math.052"
    vector_math_052.operation = 'ADD'
    #Vector_002
    vector_math_052.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_052.inputs[3].default_value = 1.0
    
    #node Vector Math.054
    vector_math_054 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_054.label = "r8.xyz (#523)"
    vector_math_054.name = "Vector Math.054"
    vector_math_054.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_054.inputs[3].default_value = 1.0
    
    #node Combine XYZ.016
    combine_xyz_016 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_016.name = "Combine XYZ.016"
    #X
    combine_xyz_016.inputs[0].default_value = 0.8333333134651184
    #Y
    combine_xyz_016.inputs[1].default_value = 0.0
    #Z
    combine_xyz_016.inputs[2].default_value = 0.0
    
    #node Math.130
    math_130 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_130.label = "r10.y (#531)"
    math_130.name = "Math.130"
    math_130.operation = 'MULTIPLY_ADD'
    math_130.use_clamp = False
    
    #node Math.131
    math_131 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_131.name = "Math.131"
    math_131.operation = 'MULTIPLY'
    math_131.use_clamp = False
    #Value_001
    math_131.inputs[1].default_value = -1.0
    #Value_002
    math_131.inputs[2].default_value = 0.5
    
    #node Math.132
    math_132 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_132.label = "r10.y (#532)"
    math_132.name = "Math.132"
    math_132.operation = 'MAXIMUM'
    math_132.use_clamp = False
    #Value
    math_132.inputs[0].default_value = 0.0
    #Value_002
    math_132.inputs[2].default_value = 0.5
    
    #node Math.134
    math_134 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_134.label = "r10.w (#533)"
    math_134.name = "Math.134"
    math_134.operation = 'ADD'
    math_134.use_clamp = False
    #Value
    math_134.inputs[0].default_value = 2.0
    #Value_002
    math_134.inputs[2].default_value = 0.5
    
    #node Math.136
    math_136 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_136.label = "r10.y (#535)"
    math_136.name = "Math.136"
    math_136.operation = 'MULTIPLY'
    math_136.use_clamp = False
    #Value_002
    math_136.inputs[2].default_value = 0.5
    
    #node Math.138
    math_138 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_138.label = "r6.z (#537)"
    math_138.name = "Math.138"
    math_138.operation = 'MULTIPLY_ADD'
    math_138.use_clamp = False
    
    #node Math.133
    math_133 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_133.label = "-r6.z"
    math_133.name = "Math.133"
    math_133.operation = 'MULTIPLY'
    math_133.use_clamp = False
    #Value_001
    math_133.inputs[1].default_value = -1.0
    #Value_002
    math_133.inputs[2].default_value = 0.5
    
    #node Math.135
    math_135 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_135.label = "r10.w (#534)"
    math_135.name = "Math.135"
    math_135.operation = 'DIVIDE'
    math_135.use_clamp = False
    #Value
    math_135.inputs[0].default_value = 1.0
    #Value_002
    math_135.inputs[2].default_value = 0.5
    
    #node Math.137
    math_137 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_137.label = "r10.x (#536)"
    math_137.name = "Math.137"
    math_137.operation = 'ADD'
    math_137.use_clamp = False
    #Value_002
    math_137.inputs[2].default_value = 0.5
    
    #node Math.140
    math_140 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_140.label = "-r7.z"
    math_140.name = "Math.140"
    math_140.operation = 'MULTIPLY'
    math_140.use_clamp = False
    #Value_001
    math_140.inputs[1].default_value = -1.0
    #Value_002
    math_140.inputs[2].default_value = 0.5
    
    #node Math.139
    math_139 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_139.label = "r10.x (#538)"
    math_139.name = "Math.139"
    math_139.operation = 'MULTIPLY_ADD'
    math_139.use_clamp = False
    #Value_002
    math_139.inputs[2].default_value = 1.0
    
    #node Math.141
    math_141 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_141.label = "r16.y (#539)"
    math_141.name = "Math.141"
    math_141.operation = 'MULTIPLY'
    math_141.use_clamp = False
    #Value_002
    math_141.inputs[2].default_value = 1.0
    
    #node Math.142
    math_142 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_142.label = "-r5.x"
    math_142.name = "Math.142"
    math_142.operation = 'MULTIPLY'
    math_142.use_clamp = False
    #Value_001
    math_142.inputs[1].default_value = -1.0
    #Value_002
    math_142.inputs[2].default_value = 0.5
    
    #node Math.144
    math_144 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_144.label = "r5.x (#541)"
    math_144.name = "Math.144"
    math_144.operation = 'MULTIPLY_ADD'
    math_144.use_clamp = False
    
    #node Math.143
    math_143 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_143.label = "r10.x (#540)"
    math_143.name = "Math.143"
    math_143.operation = 'ADD'
    math_143.use_clamp = False
    #Value
    math_143.inputs[0].default_value = 0.5
    #Value_002
    math_143.inputs[2].default_value = 0.5
    
    #node Separate XYZ
    separate_xyz = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz.name = "Separate XYZ"
    
    #node Separate XYZ.003
    separate_xyz_003 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_003.name = "Separate XYZ.003"
    
    #node Combine XYZ
    combine_xyz = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz.label = "r7.x"
    combine_xyz.name = "Combine XYZ"
    #Z
    combine_xyz.inputs[2].default_value = 0.0
    
    #node Vector Math.073
    vector_math_073 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_073.label = "r7.xy"
    vector_math_073.name = "Vector Math.073"
    vector_math_073.operation = 'ADD'
    #Vector_001
    vector_math_073.inputs[1].default_value = (-0.5, -0.5, 0.0)
    #Vector_002
    vector_math_073.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_073.inputs[3].default_value = 1.0
    
    #node UV Map.002
    uv_map_002 = HD2_Shader.nodes.new("ShaderNodeUVMap")
    uv_map_002.label = "v4.xy"
    uv_map_002.name = "UV Map.002"
    uv_map_002.from_instancer = False
    uv_map_002.uv_map = "UVMap.002"
    
    #node Math
    math_1 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_1.name = "Math"
    math_1.operation = 'LESS_THAN'
    math_1.use_clamp = False
    #Value_001
    math_1.inputs[1].default_value = 0.5
    #Value_002
    math_1.inputs[2].default_value = 0.5
    
    #node Math.001
    math_001 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_001.name = "Math.001"
    math_001.operation = 'LESS_THAN'
    math_001.use_clamp = False
    #Value_001
    math_001.inputs[1].default_value = 0.5
    #Value_002
    math_001.inputs[2].default_value = 0.5
    
    #node Mix.014
    mix_014 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_014.label = "r6.z"
    mix_014.name = "Mix.014"
    mix_014.blend_type = 'MIX'
    mix_014.clamp_factor = True
    mix_014.clamp_result = False
    mix_014.data_type = 'FLOAT'
    mix_014.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_014.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_014.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_014.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_014.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_014.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_014.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_014.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.016
    mix_016 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_016.label = "r8.xyz"
    mix_016.name = "Mix.016"
    mix_016.blend_type = 'MIX'
    mix_016.clamp_factor = True
    mix_016.clamp_result = False
    mix_016.data_type = 'VECTOR'
    mix_016.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_016.inputs[0].default_value = 0
    
    #node Math.196
    math_196 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_196.label = "-r7.w"
    math_196.name = "Math.196"
    math_196.operation = 'MULTIPLY'
    math_196.use_clamp = False
    #Value_001
    math_196.inputs[1].default_value = -1.0
    #Value_002
    math_196.inputs[2].default_value = 0.5
    
    #node Clamp.015
    clamp_015 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_015.label = "r7.w"
    clamp_015.name = "Clamp.015"
    clamp_015.clamp_type = 'MINMAX'
    #Min
    clamp_015.inputs[1].default_value = 0.0
    #Max
    clamp_015.inputs[2].default_value = 1.0
    
    #node Math.197
    math_197 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_197.label = "r7.x"
    math_197.name = "Math.197"
    math_197.operation = 'ADD'
    math_197.use_clamp = False
    #Value_001
    math_197.inputs[1].default_value = 1.0
    #Value_002
    math_197.inputs[2].default_value = 0.5
    
    #node Math.198
    math_198 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_198.label = "r14.x"
    math_198.name = "Math.198"
    math_198.operation = 'MULTIPLY'
    math_198.use_clamp = False
    #Value_002
    math_198.inputs[2].default_value = 0.5
    
    #node Math.195
    math_195 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_195.name = "Math.195"
    math_195.operation = 'ADD'
    math_195.use_clamp = False
    #Value_002
    math_195.inputs[1].default_value = 0.0
    math_195.inputs[2].default_value = 0.0
    
    #node Math.193
    math_193 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_193.name = "Math.193"
    math_193.operation = 'MULTIPLY'
    math_193.use_clamp = False
    #Value_001
    math_193.inputs[1].default_value = -1.0
    #Value_002
    math_193.inputs[2].default_value = 0.5
    
    #node Vector Math.074
    vector_math_074 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_074.label = "-r8.xyz"
    vector_math_074.name = "Vector Math.074"
    vector_math_074.operation = 'MULTIPLY'
    #Vector_001
    vector_math_074.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_074.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_074.inputs[3].default_value = 1.0
    
    #node Vector Math.075
    vector_math_075 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_075.label = "r7.xyz"
    vector_math_075.name = "Vector Math.075"
    vector_math_075.operation = 'ADD'
    #Vector_002
    vector_math_075.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_075.inputs[3].default_value = 1.0
    
    #node Vector Math.076
    vector_math_076 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_076.label = "r8.xyz"
    vector_math_076.name = "Vector Math.076"
    vector_math_076.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_076.inputs[3].default_value = 1.0
    
    #node Math.200
    math_200 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_200.label = "-r14.y"
    math_200.name = "Math.200"
    math_200.operation = 'MULTIPLY'
    math_200.use_clamp = False
    #Value_001
    math_200.inputs[1].default_value = -1.0
    #Value_002
    math_200.inputs[2].default_value = 0.5
    
    #node Math.199
    math_199 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_199.label = "r14.y"
    math_199.name = "Math.199"
    math_199.operation = 'MULTIPLY_ADD'
    math_199.use_clamp = False
    
    #node Math.202
    math_202 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_202.label = "-r6.y"
    math_202.name = "Math.202"
    math_202.operation = 'MULTIPLY'
    math_202.use_clamp = False
    #Value_001
    math_202.inputs[1].default_value = -1.0
    #Value_002
    math_202.inputs[2].default_value = 0.5
    
    #node Math.206
    math_206 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_206.label = "-r6.z"
    math_206.name = "Math.206"
    math_206.operation = 'MULTIPLY'
    math_206.use_clamp = False
    #Value_001
    math_206.inputs[1].default_value = -1.0
    #Value_002
    math_206.inputs[2].default_value = 0.5
    
    #node Math.207
    math_207 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_207.label = "r7.y"
    math_207.name = "Math.207"
    math_207.operation = 'ADD'
    math_207.use_clamp = False
    #Value_001
    math_207.inputs[1].default_value = 0.4000000059604645
    #Value_002
    math_207.inputs[2].default_value = 0.5
    
    #node Math.201
    math_201 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_201.label = "r7.x"
    math_201.name = "Math.201"
    math_201.operation = 'ADD'
    math_201.use_clamp = False
    #Value_002
    math_201.inputs[2].default_value = 0.5
    
    #node Math.203
    math_203 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_203.label = "r7.x"
    math_203.name = "Math.203"
    math_203.operation = 'MAXIMUM'
    math_203.use_clamp = False
    #Value
    math_203.inputs[0].default_value = 0.0
    #Value_002
    math_203.inputs[2].default_value = 0.5
    
    #node Math.208
    math_208 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_208.label = "r6.z"
    math_208.name = "Math.208"
    math_208.operation = 'MULTIPLY_ADD'
    math_208.use_clamp = False
    
    #node Math.192
    math_192 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_192.label = "if (r7.x != 0)"
    math_192.name = "Math.192"
    math_192.operation = 'SUBTRACT'
    math_192.use_clamp = False
    #Value
    math_192.inputs[0].default_value = 1.0
    #Value_002
    math_192.inputs[2].default_value = 0.5
    
    #node Math.022
    math_022 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_022.name = "Math.022"
    math_022.operation = 'COMPARE'
    math_022.use_clamp = False
    #Value_001
    math_022.inputs[1].default_value = 0.0
    #Value_002
    math_022.inputs[2].default_value = 0.0
    
    #node Mix
    mix_1 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_1.label = "r8.xyz if (r7.x != 0)"
    mix_1.name = "Mix"
    mix_1.blend_type = 'MIX'
    mix_1.clamp_factor = True
    mix_1.clamp_result = False
    mix_1.data_type = 'VECTOR'
    mix_1.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_1.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_1.inputs[2].default_value = 0.0
    #B_Float
    mix_1.inputs[3].default_value = 0.0
    #A_Color
    mix_1.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_1.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_1.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_1.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.017
    mix_017 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_017.name = "Mix.017"
    mix_017.blend_type = 'MIX'
    mix_017.clamp_factor = True
    mix_017.clamp_result = False
    mix_017.data_type = 'FLOAT'
    mix_017.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_017.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_017.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_017.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_017.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_017.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_017.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_017.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.211
    math_211 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_211.name = "Math.211"
    math_211.operation = 'SUBTRACT'
    math_211.use_clamp = False
    #Value
    math_211.inputs[0].default_value = 1.0
    #Value_002
    math_211.inputs[2].default_value = 0.5
    
    #node Separate XYZ.019
    separate_xyz_019 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_019.name = "Separate XYZ.019"
    
    #node Clamp.016
    clamp_016 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_016.name = "Clamp.016"
    clamp_016.clamp_type = 'MINMAX'
    #Min
    clamp_016.inputs[1].default_value = 0.0
    #Max
    clamp_016.inputs[2].default_value = 1.0
    
    #node Clamp.017
    clamp_017 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_017.name = "Clamp.017"
    clamp_017.clamp_type = 'MINMAX'
    #Min
    clamp_017.inputs[1].default_value = 0.0
    #Max
    clamp_017.inputs[2].default_value = 1.0
    
    #node Math.212
    math_212 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_212.label = "-r6.y"
    math_212.name = "Math.212"
    math_212.operation = 'MULTIPLY'
    math_212.use_clamp = False
    #Value_001
    math_212.inputs[1].default_value = -1.0
    #Value_002
    math_212.inputs[2].default_value = 0.5
    
    #node Combine XYZ.008
    combine_xyz_008 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_008.label = "r7.xy (#553)"
    combine_xyz_008.name = "Combine XYZ.008"
    #Z
    combine_xyz_008.inputs[2].default_value = 0.0
    
    #node Math.213
    math_213 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_213.label = "r7.z (#554)"
    math_213.name = "Math.213"
    math_213.operation = 'ADD'
    math_213.use_clamp = False
    #Value
    math_213.inputs[0].default_value = 1.0
    #Value_002
    math_213.inputs[2].default_value = 0.5
    
    #node Separate XYZ.021
    separate_xyz_021 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_021.name = "Separate XYZ.021"
    
    #node Math.214
    math_214 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_214.label = "r7.x (#555)"
    math_214.name = "Math.214"
    math_214.operation = 'MULTIPLY'
    math_214.use_clamp = False
    #Value_002
    math_214.inputs[2].default_value = 0.5
    
    #node Math.218
    math_218 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_218.name = "Math.218"
    math_218.operation = 'MULTIPLY'
    math_218.use_clamp = False
    #Value_001
    math_218.inputs[1].default_value = -1.0
    #Value_002
    math_218.inputs[2].default_value = 0.5
    
    #node Math.219
    math_219 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_219.label = "r7.y (#561)"
    math_219.name = "Math.219"
    math_219.operation = 'MULTIPLY_ADD'
    math_219.use_clamp = False
    #Value_002
    math_219.inputs[2].default_value = 1.0
    
    #node Math.221
    math_221 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_221.name = "Math.221"
    math_221.operation = 'MULTIPLY'
    math_221.use_clamp = False
    #Value_001
    math_221.inputs[1].default_value = -1.0
    #Value_002
    math_221.inputs[2].default_value = 0.5
    
    #node Math.220
    math_220 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_220.label = "r7.w (#563)"
    math_220.name = "Math.220"
    math_220.operation = 'ADD'
    math_220.use_clamp = False
    #Value
    math_220.inputs[0].default_value = 1.0
    #Value_002
    math_220.inputs[2].default_value = 0.5
    
    #node Separate XYZ.026
    separate_xyz_026 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_026.name = "Separate XYZ.026"
    
    #node Math.215
    math_215 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_215.name = "Math.215"
    math_215.operation = 'SQRT'
    math_215.use_clamp = False
    #Value_001
    math_215.inputs[1].default_value = 0.5
    #Value_002
    math_215.inputs[2].default_value = 0.5
    
    #node Math.216
    math_216 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_216.name = "Math.216"
    math_216.operation = 'SQRT'
    math_216.use_clamp = False
    #Value_001
    math_216.inputs[1].default_value = 0.5
    #Value_002
    math_216.inputs[2].default_value = 0.5
    
    #node Math.217
    math_217 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_217.name = "Math.217"
    math_217.operation = 'SQRT'
    math_217.use_clamp = False
    #Value_001
    math_217.inputs[1].default_value = 0.5
    #Value_002
    math_217.inputs[2].default_value = 0.5
    
    #node Combine XYZ.021
    combine_xyz_021 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_021.label = "r10.xyw (#558)"
    combine_xyz_021.name = "Combine XYZ.021"
    
    #node Vector Math.080
    vector_math_080 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_080.label = "r10.xyw (#559)"
    vector_math_080.name = "Vector Math.080"
    vector_math_080.operation = 'ADD'
    #Vector_002
    vector_math_080.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_080.inputs[3].default_value = 1.0
    
    #node Vector Math.079
    vector_math_079 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_079.name = "Vector Math.079"
    vector_math_079.operation = 'MULTIPLY'
    #Vector_001
    vector_math_079.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_079.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_079.inputs[3].default_value = 1.0
    
    #node Vector Math.082
    vector_math_082 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_082.label = "r7.z (#562)"
    vector_math_082.name = "Vector Math.082"
    vector_math_082.operation = 'DOT_PRODUCT'
    #Vector_001
    vector_math_082.inputs[1].default_value = (0.30000001192092896, 0.5899999737739563, 0.10999999940395355)
    #Vector_002
    vector_math_082.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_082.inputs[3].default_value = 1.0
    
    #node Vector Math.081
    vector_math_081 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_081.label = "r9.xyz (#560)"
    vector_math_081.name = "Vector Math.081"
    vector_math_081.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_081.inputs[3].default_value = 1.0
    
    #node Math.222
    math_222 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_222.label = "r7.y (#564)"
    math_222.name = "Math.222"
    math_222.operation = 'MULTIPLY_ADD'
    math_222.use_clamp = False
    
    #node Combine XYZ.022
    combine_xyz_022 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_022.label = "r9.xyz"
    combine_xyz_022.name = "Combine XYZ.022"
    #X
    combine_xyz_022.inputs[0].default_value = 0.0
    #Y
    combine_xyz_022.inputs[1].default_value = 0.0
    #Z
    combine_xyz_022.inputs[2].default_value = 0.0
    
    #node Value
    value = HD2_Shader.nodes.new("ShaderNodeValue")
    value.label = "r9.w"
    value.name = "Value"
    
    value.outputs[0].default_value = 0.0
    #node Combine XYZ.025
    combine_xyz_025 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_025.label = "r7.xy"
    combine_xyz_025.name = "Combine XYZ.025"
    combine_xyz_025.inputs[2].hide = True
    #X
    combine_xyz_025.inputs[0].default_value = 0.0
    #Y
    combine_xyz_025.inputs[1].default_value = 1.0
    #Z
    combine_xyz_025.inputs[2].default_value = 0.0
    
    #node Vector Math.077
    vector_math_077 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_077.label = "-r8.xyz"
    vector_math_077.name = "Vector Math.077"
    vector_math_077.operation = 'MULTIPLY'
    #Vector_001
    vector_math_077.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_077.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_077.inputs[3].default_value = 1.0
    
    #node Vector Math.063
    vector_math_063 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_063.label = "r10.xyw (#569)"
    vector_math_063.name = "Vector Math.063"
    vector_math_063.operation = 'ADD'
    #Vector_002
    vector_math_063.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_063.inputs[3].default_value = 1.0
    
    #node Math.178
    math_178 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_178.label = "r6.y (#571)"
    math_178.name = "Math.178"
    math_178.operation = 'MULTIPLY'
    math_178.use_clamp = False
    #Value_002
    math_178.inputs[2].default_value = 0.5
    
    #node Math.179
    math_179 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_179.label = "r7.z (#572)"
    math_179.name = "Math.179"
    math_179.operation = 'ADD'
    math_179.use_clamp = False
    #Value_002
    math_179.inputs[2].default_value = 0.5
    
    #node Math.209
    math_209 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_209.label = "-r6.z"
    math_209.name = "Math.209"
    math_209.operation = 'MULTIPLY'
    math_209.use_clamp = False
    #Value_001
    math_209.inputs[1].default_value = -1.0
    #Value_002
    math_209.inputs[2].default_value = 0.5
    
    #node Vector Math.066
    vector_math_066 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_066.label = "r10.xyw (#575)"
    vector_math_066.name = "Vector Math.066"
    vector_math_066.operation = 'ADD'
    #Vector_002
    vector_math_066.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_066.inputs[3].default_value = 1.0
    
    #node Math.182
    math_182 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_182.label = "r6.z (#574)"
    math_182.name = "Math.182"
    math_182.operation = 'MULTIPLY'
    math_182.use_clamp = False
    #Value_002
    math_182.inputs[2].default_value = 0.5
    
    #node Vector Math.067
    vector_math_067 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_067.label = "r8.xyz (#576)"
    vector_math_067.name = "Vector Math.067"
    vector_math_067.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_067.inputs[3].default_value = 1.0
    
    #node Vector Math.064
    vector_math_064 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_064.label = "-r8.xyz"
    vector_math_064.name = "Vector Math.064"
    vector_math_064.operation = 'MULTIPLY'
    #Vector_001
    vector_math_064.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_064.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_064.inputs[3].default_value = 1.0
    
    #node Math.210
    math_210 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_210.name = "Math.210"
    math_210.operation = 'COMPARE'
    math_210.use_clamp = False
    #Value_001
    math_210.inputs[1].default_value = 0.0
    #Value_002
    math_210.inputs[2].default_value = 0.0
    
    #node Separate XYZ.067
    separate_xyz_067 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_067.name = "Separate XYZ.067"
    
    #node Combine XYZ.042
    combine_xyz_042 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_042.label = "r10.xyw (#582)"
    combine_xyz_042.name = "Combine XYZ.042"
    
    #node Vector Math.071
    vector_math_071 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_071.label = "r8.xyz (#581)"
    vector_math_071.name = "Vector Math.071"
    vector_math_071.operation = 'MAXIMUM'
    #Vector
    vector_math_071.inputs[0].default_value = (6.0999998822808266e-05, 6.0999998822808266e-05, 6.0999998822808266e-05)
    #Vector_002
    vector_math_071.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_071.inputs[3].default_value = 1.0
    
    #node Vector Math.069
    vector_math_069 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_069.label = "r8.xyz (#580)"
    vector_math_069.name = "Vector Math.069"
    vector_math_069.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_069.inputs[3].default_value = 1.0
    
    #node Math.183
    math_183 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_183.label = "r4.w (#577)"
    math_183.name = "Math.183"
    math_183.operation = 'MULTIPLY'
    math_183.use_clamp = False
    #Value_002
    math_183.inputs[2].default_value = 0.5
    
    #node Math.184
    math_184 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_184.label = "r4.w (#578)"
    math_184.name = "Math.184"
    math_184.operation = 'MULTIPLY'
    math_184.use_clamp = False
    #Value_002
    math_184.inputs[2].default_value = 0.5
    
    #node Separate XYZ.068
    separate_xyz_068 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_068.name = "Separate XYZ.068"
    
    #node Math.188
    math_188 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_188.name = "Math.188"
    math_188.operation = 'LOGARITHM'
    math_188.use_clamp = False
    #Value_001
    math_188.inputs[1].default_value = 2.0
    #Value_002
    math_188.inputs[2].default_value = 0.5
    
    #node Math.189
    math_189 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_189.name = "Math.189"
    math_189.operation = 'LOGARITHM'
    math_189.use_clamp = False
    #Value_001
    math_189.inputs[1].default_value = 2.0
    #Value_002
    math_189.inputs[2].default_value = 0.5
    
    #node Math.190
    math_190 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_190.name = "Math.190"
    math_190.operation = 'LOGARITHM'
    math_190.use_clamp = False
    #Value_001
    math_190.inputs[1].default_value = 2.0
    #Value_002
    math_190.inputs[2].default_value = 0.5
    
    #node Combine XYZ.048
    combine_xyz_048 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_048.label = "r13.xyz (#584)"
    combine_xyz_048.name = "Combine XYZ.048"
    
    #node Math.185
    math_185 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_185.name = "Math.185"
    math_185.operation = 'LESS_THAN'
    math_185.use_clamp = False
    #Value
    math_185.inputs[0].default_value = 0.040449999272823334
    #Value_002
    math_185.inputs[2].default_value = 0.5
    
    #node Math.186
    math_186 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_186.name = "Math.186"
    math_186.operation = 'LESS_THAN'
    math_186.use_clamp = False
    #Value
    math_186.inputs[0].default_value = 0.040449999272823334
    #Value_002
    math_186.inputs[2].default_value = 0.5
    
    #node Math.187
    math_187 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_187.name = "Math.187"
    math_187.operation = 'LESS_THAN'
    math_187.use_clamp = False
    #Value
    math_187.inputs[0].default_value = 0.040449999272823334
    #Value_002
    math_187.inputs[2].default_value = 0.5
    
    #node Vector Math.072
    vector_math_072 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_072.label = "r13.xyz (#583)"
    vector_math_072.name = "Vector Math.072"
    vector_math_072.operation = 'MULTIPLY_ADD'
    #Vector_001
    vector_math_072.inputs[1].default_value = (0.9478669762611389, 0.9478669762611389, 0.9478669762611389)
    #Vector_002
    vector_math_072.inputs[2].default_value = (0.05213300138711929, 0.05213300138711929, 0.05213300138711929)
    #Scale
    vector_math_072.inputs[3].default_value = 1.0
    
    #node Math.223
    math_223 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_223.name = "Math.223"
    math_223.operation = 'SUBTRACT'
    math_223.use_clamp = False
    #Value
    math_223.inputs[0].default_value = 1.0
    #Value_002
    math_223.inputs[2].default_value = 0.5
    
    #node Math.224
    math_224 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_224.name = "Math.224"
    math_224.operation = 'MULTIPLY'
    math_224.use_clamp = False
    #Value_001
    math_224.inputs[1].default_value = 0.5
    #Value_002
    math_224.inputs[2].default_value = 0.5
    
    #node Separate XYZ.041
    separate_xyz_041 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_041.name = "Separate XYZ.041"
    
    #node customization_camo_tiler_array
    customization_camo_tiler_array = HD2_Shader.nodes.new("ShaderNodeTexImage")
    customization_camo_tiler_array.label = "r7.yzw"
    customization_camo_tiler_array.name = "customization_camo_tiler_array"
    customization_camo_tiler_array.extension = 'EXTEND'
    customization_camo_tiler_array.image_user.frame_current = 0
    customization_camo_tiler_array.image_user.frame_duration = 1
    customization_camo_tiler_array.image_user.frame_offset = 22
    customization_camo_tiler_array.image_user.frame_start = 1
    customization_camo_tiler_array.image_user.tile = 0
    customization_camo_tiler_array.image_user.use_auto_refresh = False
    customization_camo_tiler_array.image_user.use_cyclic = False
    customization_camo_tiler_array.interpolation = 'Smart'
    customization_camo_tiler_array.projection = 'FLAT'
    customization_camo_tiler_array.projection_blend = 0.0

#get texture for camo array             
    try:
        customization_camo_tiler_array.image = bpy.data.images.get("customization_camo_tiler_array.png")
        customization_camo_tiler_array.image.colorspace_settings.name = "Non-Color"
        customization_camo_tiler_array.image.alpha_mode = "CHANNEL_PACKED"
    except:
        pass
    
    #node Vector Math.033
    vector_math_033 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_033.label = "r6.xy (#276)"
    vector_math_033.name = "Vector Math.033"
    vector_math_033.operation = 'ADD'
    #Vector
    vector_math_033.inputs[0].default_value = (-2.0, 2.0, 0.0)
    #Vector_002
    vector_math_033.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_033.inputs[3].default_value = 1.0
    
    #node Gamma.003
    gamma_003 = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma_003.name = "Gamma.003"
    #Gamma
    gamma_003.inputs[1].default_value = 2.200000047683716
    
    #node Vector Math.065
    vector_math_065 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_065.label = "r8.xyz (#570)"
    vector_math_065.name = "Vector Math.065"
    vector_math_065.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_065.inputs[3].default_value = 1.0
    
    #node pattern_lut 03
    pattern_lut_03 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    pattern_lut_03.label = "r10.x (#530)"
    pattern_lut_03.name = "pattern_lut 03"
    pattern_lut_03.extension = 'REPEAT'
    pattern_lut_03.image_user.frame_current = 0
    pattern_lut_03.image_user.frame_duration = 1
    pattern_lut_03.image_user.frame_offset = 495
    pattern_lut_03.image_user.frame_start = 1
    pattern_lut_03.image_user.tile = 0
    pattern_lut_03.image_user.use_auto_refresh = False
    pattern_lut_03.image_user.use_cyclic = False
    pattern_lut_03.interpolation = 'Closest'
    pattern_lut_03.projection = 'FLAT'
    pattern_lut_03.projection_blend = 0.0
    
    #node Math.225
    math_225 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_225.name = "Math.225"
    math_225.operation = 'MULTIPLY'
    math_225.use_clamp = False
    #Value_001
    math_225.inputs[1].default_value = 0.25
    #Value_002
    math_225.inputs[2].default_value = 0.5
    
    #node Vector Math.070
    vector_math_070 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_070.label = "-r8.xyz"
    vector_math_070.name = "Vector Math.070"
    vector_math_070.operation = 'MULTIPLY'
    #Vector_001
    vector_math_070.inputs[1].default_value = (-1.0, -1.0, -1.0)
    #Vector_002
    vector_math_070.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_070.inputs[3].default_value = 1.0
    
    #node Math.181
    math_181 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_181.label = "r6.y (#573)"
    math_181.name = "Math.181"
    math_181.operation = 'MULTIPLY_ADD'
    math_181.use_clamp = False
    
    #node Vector Math.068
    vector_math_068 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_068.label = "r10.xyw (#579)"
    vector_math_068.name = "Vector Math.068"
    vector_math_068.operation = 'ADD'
    #Vector_002
    vector_math_068.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_068.inputs[3].default_value = 1.0
    
    #node Gamma.002
    gamma_002 = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma_002.name = "Gamma.002"
    #Gamma
    gamma_002.inputs[1].default_value = 2.200000047683716
    
    #node Primary Material LUT_05
    primary_material_lut_05 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_05.label = "Primary Material LUT_05"
    primary_material_lut_05.name = "Primary Material LUT_05"
    primary_material_lut_05.use_custom_color = True
    primary_material_lut_05.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_05.extension = 'EXTEND'
    primary_material_lut_05.image_user.frame_current = 1
    primary_material_lut_05.image_user.frame_duration = 1
    primary_material_lut_05.image_user.frame_offset = 429
    primary_material_lut_05.image_user.frame_start = 1
    primary_material_lut_05.image_user.tile = 0
    primary_material_lut_05.image_user.use_auto_refresh = False
    primary_material_lut_05.image_user.use_cyclic = False
    primary_material_lut_05.interpolation = 'Closest'
    primary_material_lut_05.projection = 'FLAT'
    primary_material_lut_05.projection_blend = 0.0
    
    #node Combine XYZ.079
    combine_xyz_079 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_079.name = "Combine XYZ.079"
    #X
    combine_xyz_079.inputs[0].default_value = 0.5909090638160706
    #Z
    combine_xyz_079.inputs[2].default_value = 0.0
    
    #node Combine XYZ.082
    combine_xyz_082 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_082.name = "Combine XYZ.082"
    #X
    combine_xyz_082.inputs[0].default_value = 0.6818181872367859
    #Z
    combine_xyz_082.inputs[2].default_value = 0.0
    
    #node Vector Math.007
    vector_math_007 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_007.label = "r6.zw (#361)"
    vector_math_007.name = "Vector Math.007"
    vector_math_007.operation = 'MULTIPLY'
    #Vector_002
    vector_math_007.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_007.inputs[3].default_value = 1.0
    
    #node Separate XYZ.064
    separate_xyz_064 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_064.name = "Separate XYZ.064"
    
    #node Math.166
    math_166 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_166.name = "Math.166"
    math_166.operation = 'MULTIPLY'
    math_166.use_clamp = False
    #Value_001
    math_166.inputs[1].default_value = -1.0
    #Value_002
    math_166.inputs[2].default_value = 0.5
    
    #node Math.167
    math_167 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_167.label = "r5.x (#371)"
    math_167.name = "Math.167"
    math_167.operation = 'MULTIPLY_ADD'
    math_167.use_clamp = False
    #Value_002
    math_167.inputs[2].default_value = 1.0
    
    #node Math.173
    math_173 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_173.label = "r5.x (#376)"
    math_173.name = "Math.173"
    math_173.operation = 'SQRT'
    math_173.use_clamp = False
    #Value_001
    math_173.inputs[1].default_value = 0.5
    #Value_002
    math_173.inputs[2].default_value = 0.5
    
    #node Math.174
    math_174 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_174.label = "-r5.x"
    math_174.name = "Math.174"
    math_174.operation = 'MULTIPLY'
    math_174.use_clamp = False
    #Value_001
    math_174.inputs[1].default_value = -1.0
    #Value_002
    math_174.inputs[2].default_value = 0.5
    
    #node Vector Math.060
    vector_math_060 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_060.label = "r6.zw (#377)"
    vector_math_060.name = "Vector Math.060"
    vector_math_060.operation = 'MULTIPLY'
    #Vector_002
    vector_math_060.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_060.inputs[3].default_value = 1.0
    
    #node Vector Math.061
    vector_math_061 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_061.label = "r6.zw (#378)"
    vector_math_061.name = "Vector Math.061"
    vector_math_061.operation = 'MULTIPLY'
    #Vector_002
    vector_math_061.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_061.inputs[3].default_value = 1.0
    
    #node Vector Math.058
    vector_math_058 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_058.label = "r6.zw (#369)"
    vector_math_058.name = "Vector Math.058"
    vector_math_058.operation = 'ADD'
    #Vector_002
    vector_math_058.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_058.inputs[3].default_value = 1.0
    
    #node Mix.015
    mix_015 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_015.label = "r17.xy"
    mix_015.name = "Mix.015"
    mix_015.blend_type = 'MIX'
    mix_015.clamp_factor = True
    mix_015.clamp_result = False
    mix_015.data_type = 'VECTOR'
    mix_015.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_015.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_015.inputs[2].default_value = 0.0
    #B_Float
    mix_015.inputs[3].default_value = 0.0
    #A_Color
    mix_015.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_015.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_015.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_015.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Vector Math.003
    vector_math_003 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_003.label = "r6.zw (#351)"
    vector_math_003.name = "Vector Math.003"
    vector_math_003.operation = 'ADD'
    #Vector_002
    vector_math_003.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_003.inputs[3].default_value = 1.0
    
    #node Vector Math
    vector_math = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math.name = "Vector Math"
    vector_math.operation = 'ADD'
    #Vector
    vector_math.inputs[0].default_value = (-0.5, -0.5, -0.5)
    #Vector_002
    vector_math.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math.inputs[3].default_value = 1.0
    
    #node customization_material_detail_tiler_array.001
    customization_material_detail_tiler_array_001 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    customization_material_detail_tiler_array_001.label = "r16.xyzw"
    customization_material_detail_tiler_array_001.name = "customization_material_detail_tiler_array.001"
    customization_material_detail_tiler_array_001.extension = 'EXTEND'
    customization_material_detail_tiler_array_001.image_user.frame_current = 0
    customization_material_detail_tiler_array_001.image_user.frame_duration = 1
    customization_material_detail_tiler_array_001.image_user.frame_offset = -1
    customization_material_detail_tiler_array_001.image_user.frame_start = 1
    customization_material_detail_tiler_array_001.image_user.tile = 0
    customization_material_detail_tiler_array_001.image_user.use_auto_refresh = False
    customization_material_detail_tiler_array_001.image_user.use_cyclic = False
    customization_material_detail_tiler_array_001.interpolation = 'Smart'
    customization_material_detail_tiler_array_001.projection = 'FLAT'
    customization_material_detail_tiler_array_001.projection_blend = 0.0
    try:
        customization_material_detail_tiler_array_001.image = bpy.data.images.get("customization_material_detail_tiler_array.png")
        customization_material_detail_tiler_array_001.image.colorspace_settings.name = "Non-Color"
        customization_material_detail_tiler_array_001.image.alpha_mode = "CHANNEL_PACKED"
    except:
        pass
    
    #node Math.151
    math_151 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_151.label = "if (r7.y != 0)"
    math_151.name = "Math.151"
    math_151.operation = 'SUBTRACT'
    math_151.use_clamp = False
    #Value
    math_151.inputs[0].default_value = 1.0
    #Value_002
    math_151.inputs[2].default_value = 0.5
    
    #node Math.153
    math_153 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_153.label = "r6.x (#271)"
    math_153.name = "Math.153"
    math_153.operation = 'FLOOR'
    math_153.use_clamp = False
    #Value_001
    math_153.inputs[1].default_value = 0.5
    #Value_002
    math_153.inputs[2].default_value = 0.5
    
    #node Separate XYZ.048
    separate_xyz_048 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_048.name = "Separate XYZ.048"
    
    #node composite_array
    composite_array = HD2_Shader.nodes.new("ShaderNodeTexImage")
    composite_array.label = "r17.xyz"
    composite_array.name = "composite_array"
    composite_array.extension = 'EXTEND'
    composite_array.image_user.frame_current = 0
    composite_array.image_user.frame_duration = 1
    composite_array.image_user.frame_offset = 479
    composite_array.image_user.frame_start = 1
    composite_array.image_user.tile = 0
    composite_array.image_user.use_auto_refresh = False
    composite_array.image_user.use_cyclic = False
    composite_array.interpolation = 'Smart'
    composite_array.projection = 'FLAT'
    composite_array.projection_blend = 0.0
    try:
        composite_array.image = bpy.data.images.get("composite_array.png")
        composite_array.image.colorspace_settings.name = "Non-Color"
        composite_array.image.alpha_mode = "CHANNEL_PACKED"
    except:
        pass
    
    #node Vector Math.022
    vector_math_022 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_022.label = "r17.xyz (#309)"
    vector_math_022.name = "Vector Math.022"
    vector_math_022.operation = 'ADD'
    #Vector
    vector_math_022.inputs[0].default_value = (-0.5, -0.5, -0.5)
    #Vector_002
    vector_math_022.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_022.inputs[3].default_value = 1.0
    
    #node Clamp.019
    clamp_019 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_019.name = "Clamp.019"
    clamp_019.clamp_type = 'MINMAX'
    #Min
    clamp_019.inputs[1].default_value = 0.0
    #Max
    clamp_019.inputs[2].default_value = 1.0
    
    #node Vector Math.097
    vector_math_097 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_097.name = "Vector Math.097"
    vector_math_097.operation = 'ADD'
    #Vector_001
    vector_math_097.inputs[1].default_value = (0.5, 0.5, 0.5)
    #Vector_002
    vector_math_097.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_097.inputs[3].default_value = 1.0
    
    #node Vector Math.099
    vector_math_099 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_099.name = "Vector Math.099"
    vector_math_099.operation = 'ADD'
    #Vector_002
    vector_math_099.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_099.inputs[3].default_value = 1.0
    
    #node Vector Math.096
    vector_math_096 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_096.name = "Vector Math.096"
    vector_math_096.operation = 'MULTIPLY'
    #Vector_002
    vector_math_096.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_096.inputs[3].default_value = 1.0
    
    #node Vector Math.098
    vector_math_098 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_098.name = "Vector Math.098"
    vector_math_098.operation = 'ADD'
    #Vector_002
    vector_math_098.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_098.inputs[3].default_value = 1.0
    
    #node Vector Math.091
    vector_math_091 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_091.name = "Vector Math.091"
    vector_math_091.operation = 'ADD'
    #Vector_001
    vector_math_091.inputs[1].default_value = (0.5, 0.5, 0.5)
    #Vector_002
    vector_math_091.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_091.inputs[3].default_value = 1.0
    
    #node Combine XYZ.098
    combine_xyz_098 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_098.name = "Combine XYZ.098"
    #Z
    combine_xyz_098.inputs[2].default_value = 0.0
    
    #node Vector Math.113
    vector_math_113 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_113.name = "Vector Math.113"
    vector_math_113.operation = 'MULTIPLY'
    #Vector_001
    vector_math_113.inputs[1].default_value = (1.0, 1.0, 0.0)
    #Vector_002
    vector_math_113.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_113.inputs[3].default_value = 1.0
    
    #node Vector Math.112
    vector_math_112 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_112.name = "Vector Math.112"
    vector_math_112.operation = 'MULTIPLY'
    #Vector_001
    vector_math_112.inputs[1].default_value = (1.0, 1.0, 0.0)
    #Vector_002
    vector_math_112.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_112.inputs[3].default_value = 1.0
    
    #node Vector Math.021
    vector_math_021 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_021.name = "Vector Math.021"
    vector_math_021.operation = 'ADD'
    #Vector
    vector_math_021.inputs[0].default_value = (-0.5, -0.5, -0.5)
    #Vector_002
    vector_math_021.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_021.inputs[3].default_value = 1.0
    
    #node Mix.021
    mix_021 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_021.name = "Mix.021"
    mix_021.blend_type = 'MIX'
    mix_021.clamp_factor = True
    mix_021.clamp_result = False
    mix_021.data_type = 'VECTOR'
    mix_021.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_021.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_021.inputs[2].default_value = 0.0
    #B_Float
    mix_021.inputs[3].default_value = 0.0
    #A_Color
    mix_021.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_021.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_021.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_021.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Vector Math.114
    vector_math_114 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_114.name = "Vector Math.114"
    vector_math_114.operation = 'MULTIPLY'
    #Vector_001
    vector_math_114.inputs[1].default_value = (1.0, 1.0, 0.0)
    #Vector_002
    vector_math_114.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_114.inputs[3].default_value = 1.0
    
    #node Vector Math.111
    vector_math_111 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_111.name = "Vector Math.111"
    vector_math_111.operation = 'MULTIPLY'
    #Vector_002
    vector_math_111.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_111.inputs[3].default_value = 1.0
    
    #node Vector Math.115
    vector_math_115 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_115.name = "Vector Math.115"
    vector_math_115.operation = 'MULTIPLY'
    #Vector_002
    vector_math_115.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_115.inputs[3].default_value = 1.0
    
    #node Vector Math.089
    vector_math_089 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_089.name = "Vector Math.089"
    vector_math_089.operation = 'ADD'
    #Vector_002
    vector_math_089.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_089.inputs[3].default_value = 1.0
    
    #node Vector Math.116
    vector_math_116 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_116.name = "Vector Math.116"
    vector_math_116.operation = 'MULTIPLY'
    #Vector_002
    vector_math_116.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_116.inputs[3].default_value = 1.0
    
    #node Normal Map.001
    normal_map_001 = HD2_Shader.nodes.new("ShaderNodeNormalMap")
    normal_map_001.name = "Normal Map.001"
    normal_map_001.space = 'TANGENT'
    normal_map_001.uv_map = ""
    #Strength
    normal_map_001.inputs[0].default_value = 1.0
    
    #node Math.270
    math_270 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_270.name = "Math.270"
    math_270.operation = 'MULTIPLY_ADD'
    math_270.use_clamp = False
    #Value_001
    math_270.inputs[1].default_value = 2.0
    #Value_002
    math_270.inputs[2].default_value = -1.0
    
    #node Math.204
    math_204 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_204.name = "Math.204"
    math_204.operation = 'MULTIPLY'
    math_204.use_clamp = False
    #Value_002
    math_204.inputs[2].default_value = 0.0
    
    #node Math.205
    math_205 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_205.name = "Math.205"
    math_205.operation = 'MULTIPLY_ADD'
    math_205.use_clamp = True
    
    #node Math.268
    math_268 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_268.name = "Math.268"
    math_268.operation = 'SUBTRACT'
    math_268.use_clamp = False
    #Value
    math_268.inputs[0].default_value = 1.0
    #Value_002
    math_268.inputs[2].default_value = 0.0
    
    #node Math.269
    math_269 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_269.name = "Math.269"
    math_269.operation = 'SQRT'
    math_269.use_clamp = False
    #Value_001
    math_269.inputs[1].default_value = 0.5
    #Value_002
    math_269.inputs[2].default_value = 0.0
    
    #node Combine XYZ.099
    combine_xyz_099 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_099.name = "Combine XYZ.099"
    
    #node Vector Math.093
    vector_math_093 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_093.name = "Vector Math.093"
    vector_math_093.operation = 'ADD'
    #Vector_001
    vector_math_093.inputs[1].default_value = (1.0, 1.0, 1.0)
    #Vector_002
    vector_math_093.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_093.inputs[3].default_value = 1.0
    
    #node Vector Math.092
    vector_math_092 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_092.name = "Vector Math.092"
    vector_math_092.operation = 'NORMALIZE'
    #Vector_001
    vector_math_092.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_092.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_092.inputs[3].default_value = 1.0
    
    #node Vector Math.094
    vector_math_094 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_094.name = "Vector Math.094"
    vector_math_094.operation = 'DIVIDE'
    #Vector_001
    vector_math_094.inputs[1].default_value = (2.0, 2.0, 2.0)
    #Vector_002
    vector_math_094.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_094.inputs[3].default_value = 1.0
    
    #node Math.271
    math_271 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_271.name = "Math.271"
    math_271.operation = 'MULTIPLY_ADD'
    math_271.use_clamp = False
    #Value_001
    math_271.inputs[1].default_value = 2.0
    #Value_002
    math_271.inputs[2].default_value = -1.0
    
    #node Separate XYZ.056
    separate_xyz_056 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_056.name = "Separate XYZ.056"
    
    #node Math.272
    math_272 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_272.name = "Math.272"
    math_272.operation = 'MULTIPLY_ADD'
    math_272.use_clamp = False
    #Value_001
    math_272.inputs[1].default_value = 2.0
    #Value_002
    math_272.inputs[2].default_value = -1.0
    
    #node Math.273
    math_273 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_273.name = "Math.273"
    math_273.operation = 'MULTIPLY'
    math_273.use_clamp = False
    #Value_002
    math_273.inputs[2].default_value = 0.0
    
    #node Math.274
    math_274 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_274.name = "Math.274"
    math_274.operation = 'MULTIPLY_ADD'
    math_274.use_clamp = True
    
    #node Math.275
    math_275 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_275.name = "Math.275"
    math_275.operation = 'SUBTRACT'
    math_275.use_clamp = False
    #Value
    math_275.inputs[0].default_value = 1.0
    #Value_002
    math_275.inputs[2].default_value = 0.0
    
    #node Math.276
    math_276 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_276.name = "Math.276"
    math_276.operation = 'SQRT'
    math_276.use_clamp = False
    #Value_001
    math_276.inputs[1].default_value = 0.5
    #Value_002
    math_276.inputs[2].default_value = 0.0
    
    #node Combine XYZ.100
    combine_xyz_100 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_100.name = "Combine XYZ.100"
    
    #node Vector Math.110
    vector_math_110 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_110.name = "Vector Math.110"
    vector_math_110.operation = 'ADD'
    #Vector_001
    vector_math_110.inputs[1].default_value = (1.0, 1.0, 1.0)
    #Vector_002
    vector_math_110.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_110.inputs[3].default_value = 1.0
    
    #node Vector Math.117
    vector_math_117 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_117.name = "Vector Math.117"
    vector_math_117.operation = 'NORMALIZE'
    #Vector_001
    vector_math_117.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_117.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_117.inputs[3].default_value = 1.0
    
    #node Vector Math.118
    vector_math_118 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_118.name = "Vector Math.118"
    vector_math_118.operation = 'DIVIDE'
    #Vector_001
    vector_math_118.inputs[1].default_value = (2.0, 2.0, 2.0)
    #Vector_002
    vector_math_118.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_118.inputs[3].default_value = 1.0
    
    #node Math.277
    math_277 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_277.name = "Math.277"
    math_277.operation = 'MULTIPLY_ADD'
    math_277.use_clamp = False
    #Value_001
    math_277.inputs[1].default_value = 2.0
    #Value_002
    math_277.inputs[2].default_value = -1.0
    
    #node Separate XYZ.063
    separate_xyz_063 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_063.name = "Separate XYZ.063"
    
    #node Mix Shader
    mix_shader = HD2_Shader.nodes.new("ShaderNodeMixShader")
    mix_shader.name = "Mix Shader"
    
    #node transparency shader
    transparency_shader: ShaderNodeBsdfTransparent = HD2_Shader.nodes.new("ShaderNodeBsdfTransparent")
    transparency_shader.name = "Transparent"

    #node Mix Shader.001
    mix_shader_transparency = HD2_Shader.nodes.new("ShaderNodeMixShader")
    mix_shader_transparency.name = "Mix Shader.001"
    
    #node Separate XYZ.051
    separate_xyz_051 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_051.name = "Separate XYZ.051"

    #node Group Input
    group_input_1 = HD2_Shader.nodes.new("NodeGroupInput")
    group_input_1.name = "Group Input"

    # Cape Decal Node
    cape_decal_group_node: ShaderNodeGroup = HD2_Shader.nodes.new("ShaderNodeGroup")
    cape_decal_group_node.name = "Cape Decal Template"
    cape_decal_group_node.node_tree = cape_decal_tree

    decal_mix_node: ShaderNodeMix = HD2_Shader.nodes.new("ShaderNodeMix")
    decal_mix_node.data_type = "RGBA"
    decal_mix_node.name = "Decal Mix"
    HD2_Shader.links.new(cape_decal_group_node.outputs[0], decal_mix_node.inputs[7])
    HD2_Shader.links.new(cape_decal_group_node.outputs[1], decal_mix_node.inputs[0])

    alpha_cutoff_node: ShaderNodeMath = HD2_Shader.nodes.new("ShaderNodeMath")
    alpha_cutoff_node.name = "Alpha Clip"
    alpha_cutoff_node.operation = "GREATER_THAN"
    alpha_cutoff_node.inputs[1].default_value = 0.5

    #vector_math_069.Vector -> principled_bsdf_001.Base Color
    HD2_Shader.links.new(vector_math_069.outputs[0], decal_mix_node.inputs[6])

    #node Principled BSDF.001
    principled_bsdf_001 = HD2_Shader.nodes.new("ShaderNodeBsdfPrincipled")
    principled_bsdf_001.name = "Principled BSDF.001"
    principled_bsdf_001.distribution = 'MULTI_GGX'
    principled_bsdf_001.subsurface_method = 'RANDOM_WALK'
    #Decal Mix Color -> principled_bsdf_001.Base Color
    HD2_Shader.links.new(decal_mix_node.outputs[2], principled_bsdf_001.inputs[0])
    #IOR
    principled_bsdf_001.inputs[3].default_value = 1.4500000476837158
    #Cape Control Mask Alpha -> principled_bsdf_001 Alpha
    HD2_Shader.links.new(group_input_1.outputs[3], alpha_cutoff_node.inputs[0])
    HD2_Shader.links.new(alpha_cutoff_node.outputs[0], principled_bsdf_001.inputs[4])
    #Weight
    principled_bsdf_001.inputs[6].default_value = 0.0
    #Subsurface Weight
    principled_bsdf_001.inputs[7].default_value = 0.0
    #Subsurface Radius
    principled_bsdf_001.inputs[8].default_value = (1.0, 0.20000000298023224, 0.10000000149011612)
    #Subsurface Scale
    principled_bsdf_001.inputs[9].default_value = 0.05000000074505806
    #Subsurface IOR
    principled_bsdf_001.inputs[10].default_value = 1.399999976158142
    #Subsurface Anisotropy
    principled_bsdf_001.inputs[11].default_value = 0.0
    #Specular Tint
    principled_bsdf_001.inputs[13].default_value = (1.0, 1.0, 1.0, 1.0)
    #Anisotropic
    principled_bsdf_001.inputs[14].default_value = 0.0
    #Anisotropic Rotation
    principled_bsdf_001.inputs[15].default_value = 0.0
    #Tangent
    principled_bsdf_001.inputs[16].default_value = (0.0, 0.0, 0.0)
    #Transmission Weight
    principled_bsdf_001.inputs[17].default_value = 0.0
    #Coat IOR
    principled_bsdf_001.inputs[20].default_value = 1.4500000476837158
    #Coat Tint
    principled_bsdf_001.inputs[21].default_value = (1.0, 1.0, 1.0, 1.0)
    #Sheen Weight
    principled_bsdf_001.inputs[23].default_value = 0.0
    #Sheen Roughness
    principled_bsdf_001.inputs[24].default_value = 0.5
    #Sheen Tint
    principled_bsdf_001.inputs[25].default_value = (1.0, 1.0, 1.0, 1.0)
    #Emission Color
    principled_bsdf_001.inputs[26].default_value = (1.0, 1.0, 1.0, 1.0)
    #Emission Strength
    principled_bsdf_001.inputs[27].default_value = 0.0
    
    #node Normal Map
    normal_map = HD2_Shader.nodes.new("ShaderNodeNormalMap")
    normal_map.name = "Normal Map"
    normal_map.space = 'TANGENT'
    normal_map.uv_map = ""
    #Strength
    normal_map.inputs[0].default_value = 1.0
    
    #node Math.050
    math_050 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_050.name = "Math.050"
    math_050.operation = 'POWER'
    math_050.use_clamp = False
    #Value_001
    math_050.inputs[1].default_value = 2.200000047683716
    #Value_002
    math_050.inputs[2].default_value = 0.5
    
    #node Gamma.001
    gamma_001 = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma_001.name = "Gamma.001"
    #Gamma
    gamma_001.inputs[1].default_value = 2.200000047683716
    
    #node Gamma
    gamma = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma.name = "Gamma"
    #Gamma
    gamma.inputs[1].default_value = 2.200000047683716
    
    #node Math.278
    math_278 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_278.name = "Math.278"
    math_278.operation = 'POWER'
    math_278.use_clamp = False
    #Value_001
    math_278.inputs[1].default_value = 2.200000047683716
    #Value_002
    math_278.inputs[2].default_value = 0.5
    
    #node Clamp.018
    clamp_018 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_018.name = "Clamp.018"
    clamp_018.clamp_type = 'MINMAX'
    #Min
    clamp_018.inputs[1].default_value = 0.0
    #Max
    clamp_018.inputs[2].default_value = 1.0
    
    #node Mix.033
    mix_033 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_033.name = "Mix.033"
    mix_033.blend_type = 'MIX'
    mix_033.clamp_factor = True
    mix_033.clamp_result = False
    mix_033.data_type = 'RGBA'
    mix_033.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_033.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_033.inputs[2].default_value = 0.0
    #B_Float
    mix_033.inputs[3].default_value = 0.0
    #A_Vector
    mix_033.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_033.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_033.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_033.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.096
    combine_xyz_096 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_096.name = "Combine XYZ.096"
    #X
    combine_xyz_096.inputs[0].default_value = 1.0
    #Z
    combine_xyz_096.inputs[2].default_value = 0.0
    
    #node Combine XYZ.097
    combine_xyz_097 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_097.name = "Combine XYZ.097"
    #X
    combine_xyz_097.inputs[0].default_value = 1.0
    #Z
    combine_xyz_097.inputs[2].default_value = 0.0
    
    #node Separate XYZ.036
    separate_xyz_036 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_036.name = "Separate XYZ.036"
    
    #node Primary Material LUT_22
    primary_material_lut_22 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_22.label = "Primary Material LUT_22"
    primary_material_lut_22.name = "Primary Material LUT_22"
    primary_material_lut_22.use_custom_color = True
    primary_material_lut_22.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_22.extension = 'EXTEND'
    primary_material_lut_22.image_user.frame_current = 1
    primary_material_lut_22.image_user.frame_duration = 1
    primary_material_lut_22.image_user.frame_offset = 429
    primary_material_lut_22.image_user.frame_start = 1
    primary_material_lut_22.image_user.tile = 0
    primary_material_lut_22.image_user.use_auto_refresh = False
    primary_material_lut_22.image_user.use_cyclic = False
    primary_material_lut_22.interpolation = 'Closest'
    primary_material_lut_22.projection = 'FLAT'
    primary_material_lut_22.projection_blend = 0.0
    
    #node Math.084
    math_084 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_084.label = "r6.z (#253)"
    math_084.name = "Math.084"
    math_084.operation = 'FLOOR'
    math_084.use_clamp = False
    #Value_001
    math_084.inputs[1].default_value = 0.5
    #Value_002
    math_084.inputs[2].default_value = 0.5
    
    #node Value.006
    value_006 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_006.label = "r4.w (#244)"
    value_006.name = "Value.006"
    
    value_006.outputs[0].default_value = 26.0
    #node Math.085
    math_085 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_085.label = "r6.z (#254)"
    math_085.name = "Math.085"
    math_085.operation = 'FLOORED_MODULO'
    math_085.use_clamp = False
    #Value_002
    math_085.inputs[2].default_value = 0.5
    
    #node Object Info
    object_info = HD2_Shader.nodes.new("ShaderNodeObjectInfo")
    object_info.name = "Object Info"
    
    #node Mix.075
    mix_075 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_075.name = "Mix.075"
    mix_075.blend_type = 'MIX'
    mix_075.clamp_factor = True
    mix_075.clamp_result = False
    mix_075.data_type = 'RGBA'
    mix_075.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_075.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_075.inputs[2].default_value = 0.0
    #B_Float
    mix_075.inputs[3].default_value = 0.0
    #A_Vector
    mix_075.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_075.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_075.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_075.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Separate XYZ.024
    separate_xyz_024 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_024.name = "Separate XYZ.024"
    
    #node Math.086
    math_086 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_086.label = "r14.z (#255)"
    math_086.name = "Math.086"
    math_086.operation = 'FLOOR'
    math_086.use_clamp = False
    #Value_001
    math_086.inputs[1].default_value = 0.5
    #Value_002
    math_086.inputs[2].default_value = 0.5
    
    #node Math.229
    math_229 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_229.name = "Math.229"
    math_229.operation = 'MULTIPLY'
    math_229.use_clamp = False
    #Value_001
    math_229.inputs[1].default_value = 7.853950023651123
    #Value_002
    math_229.inputs[2].default_value = 0.5
    
    #node Math.228
    math_228 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_228.name = "Math.228"
    math_228.operation = 'MULTIPLY'
    math_228.use_clamp = False
    #Value_001
    math_228.inputs[1].default_value = 31.4158992767334
    #Value_002
    math_228.inputs[2].default_value = 0.5
    
    #node Math.226
    math_226 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_226.name = "Math.226"
    math_226.operation = 'FLOOR'
    math_226.use_clamp = False
    #Value_001
    math_226.inputs[1].default_value = 0.5
    #Value_002
    math_226.inputs[2].default_value = 0.5
    
    #node Detail UVs
    detail_uvs = HD2_Shader.nodes.new("ShaderNodeUVMap")
    detail_uvs.label = "Detail UVs"
    detail_uvs.name = "Detail UVs"
    detail_uvs.from_instancer = False
    detail_uvs.uv_map = "UVMap.001"
    
    #node Math.079
    math_079 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_079.label = "r6.z (#251)"
    math_079.name = "Math.079"
    math_079.operation = 'MULTIPLY'
    math_079.use_clamp = False
    #Value_002
    math_079.inputs[2].default_value = 0.5
    
    #node Combine XYZ.029
    combine_xyz_029 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_029.name = "Combine XYZ.029"
    #Z
    combine_xyz_029.inputs[2].default_value = 0.0
    
    #node Math.038
    math_038 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_038.label = "r7.z (#504)"
    math_038.name = "Math.038"
    math_038.operation = 'MULTIPLY'
    math_038.use_clamp = False
    #Value_002
    math_038.inputs[2].default_value = 0.5
    
    #node Math.041
    math_041 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_041.label = "r4.w (#320)"
    math_041.name = "Math.041"
    math_041.operation = 'MULTIPLY'
    math_041.use_clamp = False
    #Value_002
    math_041.inputs[2].default_value = 0.5
    
    #node Combine XYZ.086
    combine_xyz_086 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_086.name = "Combine XYZ.086"
    #X
    combine_xyz_086.inputs[0].default_value = 0.7727272510528564
    #Z
    combine_xyz_086.inputs[2].default_value = 0.0
    
    #node Combine XYZ.089
    combine_xyz_089 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_089.name = "Combine XYZ.089"
    #X
    combine_xyz_089.inputs[0].default_value = 0.8181818127632141
    #Z
    combine_xyz_089.inputs[2].default_value = 0.0
    
    #node Combine XYZ.088
    combine_xyz_088 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_088.name = "Combine XYZ.088"
    #X
    combine_xyz_088.inputs[0].default_value = 0.8181818127632141
    #Z
    combine_xyz_088.inputs[2].default_value = 0.0
    
    #node Combine XYZ.090
    combine_xyz_090 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_090.name = "Combine XYZ.090"
    #X
    combine_xyz_090.inputs[0].default_value = 0.8636363744735718
    #Z
    combine_xyz_090.inputs[2].default_value = 0.0
    
    #node Combine XYZ.091
    combine_xyz_091 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_091.name = "Combine XYZ.091"
    #X
    combine_xyz_091.inputs[0].default_value = 0.8636363744735718
    #Z
    combine_xyz_091.inputs[2].default_value = 0.0
    
    #node Math.227
    math_227 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_227.name = "Math.227"
    math_227.operation = 'SUBTRACT'
    math_227.use_clamp = False
    #Value
    math_227.inputs[0].default_value = 25.0
    #Value_002
    math_227.inputs[2].default_value = 0.5
    
    #node Vector Math.001
    vector_math_001 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_001.name = "Vector Math.001"
    vector_math_001.operation = 'MULTIPLY_ADD'
    #Scale
    vector_math_001.inputs[3].default_value = 1.0
    
    #node Math.042
    math_042 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_042.label = "r4.w (#321)"
    math_042.name = "Math.042"
    math_042.operation = 'MAXIMUM'
    math_042.use_clamp = False
    #Value
    math_042.inputs[0].default_value = 0.0
    #Value_002
    math_042.inputs[2].default_value = 0.5
    
    #node Math.025
    math_025 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_025.label = "r10.x (#469)"
    math_025.name = "Math.025"
    math_025.operation = 'MULTIPLY'
    math_025.use_clamp = False
    #Value_002
    math_025.inputs[2].default_value = 0.5
    
    #node Math.039
    math_039 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_039.label = "r7.z (#505)"
    math_039.name = "Math.039"
    math_039.operation = 'MAXIMUM'
    math_039.use_clamp = False
    #Value
    math_039.inputs[0].default_value = 0.0
    #Value_002
    math_039.inputs[2].default_value = 0.5
    
    #node Combine XYZ.087
    combine_xyz_087 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_087.name = "Combine XYZ.087"
    #X
    combine_xyz_087.inputs[0].default_value = 0.7727272510528564
    #Z
    combine_xyz_087.inputs[2].default_value = 0.0
    
    #node Separate XYZ.018
    separate_xyz_018 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_018.label = "r19.xyz"
    separate_xyz_018.name = "Separate XYZ.018"
    
    #node Primary Material LUT_17
    primary_material_lut_17 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_17.label = "Primary Material LUT_17"
    primary_material_lut_17.name = "Primary Material LUT_17"
    primary_material_lut_17.use_custom_color = True
    primary_material_lut_17.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_17.extension = 'EXTEND'
    primary_material_lut_17.image_user.frame_current = 1
    primary_material_lut_17.image_user.frame_duration = 1
    primary_material_lut_17.image_user.frame_offset = 429
    primary_material_lut_17.image_user.frame_start = 1
    primary_material_lut_17.image_user.tile = 0
    primary_material_lut_17.image_user.use_auto_refresh = False
    primary_material_lut_17.image_user.use_cyclic = False
    primary_material_lut_17.interpolation = 'Closest'
    primary_material_lut_17.projection = 'FLAT'
    primary_material_lut_17.projection_blend = 0.0
    
    #node Primary Material LUT_18
    primary_material_lut_18 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_18.label = "Primary Material LUT_18"
    primary_material_lut_18.name = "Primary Material LUT_18"
    primary_material_lut_18.use_custom_color = True
    primary_material_lut_18.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_18.extension = 'EXTEND'
    primary_material_lut_18.image_user.frame_current = 1
    primary_material_lut_18.image_user.frame_duration = 1
    primary_material_lut_18.image_user.frame_offset = 429
    primary_material_lut_18.image_user.frame_start = 1
    primary_material_lut_18.image_user.tile = 0
    primary_material_lut_18.image_user.use_auto_refresh = False
    primary_material_lut_18.image_user.use_cyclic = False
    primary_material_lut_18.interpolation = 'Closest'
    primary_material_lut_18.projection = 'FLAT'
    primary_material_lut_18.projection_blend = 0.0
    
    #node Primary Material LUT_19
    primary_material_lut_19 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_19.label = "Primary Material LUT_19"
    primary_material_lut_19.name = "Primary Material LUT_19"
    primary_material_lut_19.use_custom_color = True
    primary_material_lut_19.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_19.extension = 'EXTEND'
    primary_material_lut_19.image_user.frame_current = 1
    primary_material_lut_19.image_user.frame_duration = 1
    primary_material_lut_19.image_user.frame_offset = 429
    primary_material_lut_19.image_user.frame_start = 1
    primary_material_lut_19.image_user.tile = 0
    primary_material_lut_19.image_user.use_auto_refresh = False
    primary_material_lut_19.image_user.use_cyclic = False
    primary_material_lut_19.interpolation = 'Closest'
    primary_material_lut_19.projection = 'FLAT'
    primary_material_lut_19.projection_blend = 0.0
    
    #node Math.003
    math_003 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_003.label = "r2.x (#222)"
    math_003.name = "Math.003"
    math_003.operation = 'ADD'
    math_003.use_clamp = False
    #Value
    math_003.inputs[0].default_value = -0.5
    #Value_002
    math_003.inputs[2].default_value = 0.5
    
    #node Combine XYZ.026
    combine_xyz_026 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_026.name = "Combine XYZ.026"
    #X
    combine_xyz_026.inputs[0].default_value = 0.0
    #Z
    combine_xyz_026.inputs[2].default_value = 0.0
    
    #node Vector Math.002
    vector_math_002 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_002.name = "Vector Math.002"
    vector_math_002.operation = 'FRACTION'
    #Vector_001
    vector_math_002.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Vector_002
    vector_math_002.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_002.inputs[3].default_value = 1.0
    
    #node Math.043
    math_043 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_043.label = "r4.w (#322)"
    math_043.name = "Math.043"
    math_043.operation = 'MULTIPLY'
    math_043.use_clamp = False
    #Value
    math_043.inputs[0].default_value = 4.0
    #Value_002
    math_043.inputs[2].default_value = 0.5
    
    #node Combine XYZ.020
    combine_xyz_020 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_020.label = "r19.xy"
    combine_xyz_020.name = "Combine XYZ.020"
    #Z
    combine_xyz_020.inputs[2].default_value = 0.0
    
    #node Math.026
    math_026 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_026.label = "r10.x (#470)"
    math_026.name = "Math.026"
    math_026.operation = 'MAXIMUM'
    math_026.use_clamp = False
    #Value
    math_026.inputs[0].default_value = 0.0
    #Value_002
    math_026.inputs[2].default_value = 0.5
    
    #node Math.040
    math_040 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_040.label = "r7.z (#506)"
    math_040.name = "Math.040"
    math_040.operation = 'MULTIPLY'
    math_040.use_clamp = False
    #Value
    math_040.inputs[0].default_value = 4.0
    #Value_002
    math_040.inputs[2].default_value = 0.5
    
    #node Separate XYZ.016
    separate_xyz_016 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_016.name = "Separate XYZ.016"
    
    #node Separate XYZ.025
    separate_xyz_025 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_025.name = "Separate XYZ.025"
    
    #node Math.236
    math_236 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_236.name = "Math.236"
    math_236.operation = 'MULTIPLY_ADD'
    math_236.use_clamp = False
    #Value_001
    math_236.inputs[1].default_value = 10000000.0
    #Value_002
    math_236.inputs[2].default_value = -4960000.0
    
    #node Mix.066
    mix_066 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_066.name = "Mix.066"
    mix_066.blend_type = 'MIX'
    mix_066.clamp_factor = True
    mix_066.clamp_result = False
    mix_066.data_type = 'FLOAT'
    mix_066.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_066.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_066.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_066.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_066.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_066.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_066.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_066.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.068
    mix_068 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_068.name = "Mix.068"
    mix_068.blend_type = 'MIX'
    mix_068.clamp_factor = True
    mix_068.clamp_result = False
    mix_068.data_type = 'FLOAT'
    mix_068.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_068.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_068.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_068.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_068.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_068.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_068.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_068.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.069
    mix_069 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_069.name = "Mix.069"
    mix_069.blend_type = 'MIX'
    mix_069.clamp_factor = True
    mix_069.clamp_result = False
    mix_069.data_type = 'RGBA'
    mix_069.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_069.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_069.inputs[2].default_value = 0.0
    #B_Float
    mix_069.inputs[3].default_value = 0.0
    #A_Vector
    mix_069.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_069.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_069.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_069.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Vector Math.078
    vector_math_078 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_078.name = "Vector Math.078"
    vector_math_078.operation = 'ADD'
    #Vector_002
    vector_math_078.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_078.inputs[3].default_value = 1.0
    
    #node Vector Math.019
    vector_math_019 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_019.label = "r20.xy (#323)"
    vector_math_019.name = "Vector Math.019"
    vector_math_019.operation = 'MULTIPLY'
    #Vector_002
    vector_math_019.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_019.inputs[3].default_value = 1.0
    
    #node Math.060
    math_060 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_060.label = "r5.x (#388)"
    math_060.name = "Math.060"
    math_060.operation = 'MULTIPLY'
    math_060.use_clamp = False
    #Value_002
    math_060.inputs[2].default_value = 0.5
    
    #node Math.027
    math_027 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_027.label = "r10.x (#471)"
    math_027.name = "Math.027"
    math_027.operation = 'MULTIPLY'
    math_027.use_clamp = False
    #Value
    math_027.inputs[0].default_value = 4.0
    #Value_002
    math_027.inputs[2].default_value = 0.5
    
    #node Combine XYZ.018
    combine_xyz_018 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_018.name = "Combine XYZ.018"
    #Z
    combine_xyz_018.inputs[2].default_value = 0.0
    
    #node Combine XYZ.017
    combine_xyz_017 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_017.name = "Combine XYZ.017"
    #Z
    combine_xyz_017.inputs[2].default_value = 0.0
        
    #node Math.237
    math_237 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_237.name = "Math.237"
    math_237.operation = 'MULTIPLY_ADD'
    math_237.use_clamp = False
    #Value_001
    math_237.inputs[1].default_value = 10000000.0
    #Value_002
    math_237.inputs[2].default_value = -4960000.0
    
    #node Combine XYZ.093
    combine_xyz_093 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_093.name = "Combine XYZ.093"
    #X
    combine_xyz_093.inputs[0].default_value = 0.9090909361839294
    #Z
    combine_xyz_093.inputs[2].default_value = 0.0
    
    #node Combine XYZ.092
    combine_xyz_092 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_092.name = "Combine XYZ.092"
    #X
    combine_xyz_092.inputs[0].default_value = 0.9090909361839294
    #Z
    combine_xyz_092.inputs[2].default_value = 0.0
    
    #node Math.080
    math_080 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_080.name = "Math.080"
    math_080.operation = 'LESS_THAN'
    math_080.use_clamp = False
    #Value
    math_080.inputs[0].default_value = 3.0
    #Value_002
    math_080.inputs[2].default_value = 0.5
    
    #node Vector Math.083
    vector_math_083 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_083.name = "Vector Math.083"
    vector_math_083.operation = 'MULTIPLY'
    #Vector_001
    vector_math_083.inputs[1].default_value = (1.0, 0.03846200183033943, 0.0)
    #Vector_002
    vector_math_083.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_083.inputs[3].default_value = 1.0
    
    #node Math.061
    math_061 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_061.label = "r5.x (#389)"
    math_061.name = "Math.061"
    math_061.operation = 'MAXIMUM'
    math_061.use_clamp = False
    #Value
    math_061.inputs[0].default_value = 0.0
    #Value_002
    math_061.inputs[2].default_value = 0.5
    
    #node Value.008
    value_008 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_008.label = "r22.z (#476)"
    value_008.name = "Value.008"
    
    value_008.outputs[0].default_value = 0.0
    #node Math.028
    math_028 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_028.label = "r22.x (#472)"
    math_028.name = "Math.028"
    math_028.operation = 'MULTIPLY'
    math_028.use_clamp = False
    #Value_002
    math_028.inputs[2].default_value = 0.5
    
    #node Math.029
    math_029 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_029.label = "r22.y (#473)"
    math_029.name = "Math.029"
    math_029.operation = 'MULTIPLY'
    math_029.use_clamp = False
    #Value_002
    math_029.inputs[2].default_value = 0.5
    
    #node Reroute
    reroute = HD2_Shader.nodes.new("NodeReroute")
    reroute.label = "r27.x (#474)"
    reroute.name = "Reroute"
    #node Reroute.001
    reroute_001 = HD2_Shader.nodes.new("NodeReroute")
    reroute_001.label = "r27.y (#475)"
    reroute_001.name = "Reroute.001"
    #node Vector Math.016
    vector_math_016 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_016.label = "r24.xy (#507)"
    vector_math_016.name = "Vector Math.016"
    vector_math_016.operation = 'MULTIPLY'
    #Vector_002
    vector_math_016.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_016.inputs[3].default_value = 1.0
    
    #node Separate XYZ.020
    separate_xyz_020 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_020.name = "Separate XYZ.020"

    #node Mix.028
    Slot1and2 = HD2_Shader.nodes.new("ShaderNodeMix")
    Slot1and2.label = "Slot 1 & 2"
    Slot1and2.name = "Mix.028"
    Slot1and2.blend_type = 'MIX'
    Slot1and2.clamp_factor = True
    Slot1and2.clamp_result = False
    Slot1and2.data_type = 'FLOAT'
    Slot1and2.factor_mode = 'UNIFORM'
    
    #node Math.238
    math_238 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_238.name = "Math.238"
    math_238.operation = 'MULTIPLY_ADD'
    math_238.use_clamp = False
    #Value_001
    math_238.inputs[0].default_value = 0.0
    math_238.inputs[1].default_value = 10000000.0
    math_238.inputs[2].default_value = -4960000.0
    
    #node Primary Material LUT_20
    primary_material_lut_20 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_20.label = "Primary Material LUT_20"
    primary_material_lut_20.name = "Primary Material LUT_20"
    primary_material_lut_20.use_custom_color = True
    primary_material_lut_20.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_20.extension = 'EXTEND'
    primary_material_lut_20.image_user.frame_current = 1
    primary_material_lut_20.image_user.frame_duration = 1
    primary_material_lut_20.image_user.frame_offset = 429
    primary_material_lut_20.image_user.frame_start = 1
    primary_material_lut_20.image_user.tile = 0
    primary_material_lut_20.image_user.use_auto_refresh = False
    primary_material_lut_20.image_user.use_cyclic = False
    primary_material_lut_20.interpolation = 'Closest'
    primary_material_lut_20.projection = 'FLAT'
    primary_material_lut_20.projection_blend = 0.0
    
    #node Math.081
    math_081 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_081.label = "r5.x (#249)"
    math_081.name = "Math.081"
    math_081.operation = 'SUBTRACT'
    math_081.use_clamp = False
    #Value
    math_081.inputs[0].default_value = 1.0
    #Value_002
    math_081.inputs[2].default_value = 0.5
    
    #node Separate XYZ.035
    separate_xyz_035 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_035.name = "Separate XYZ.035"
    
    #node Separate XYZ.040
    separate_xyz_040 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_040.name = "Separate XYZ.040"
    
    #node Math.062
    math_062 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_062.label = "r5.x (#390)"
    math_062.name = "Math.062"
    math_062.operation = 'MULTIPLY'
    math_062.use_clamp = False
    #Value
    math_062.inputs[0].default_value = 4.0
    #Value_002
    math_062.inputs[2].default_value = 0.5
    
    #node Separate XYZ.032
    separate_xyz_032 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_032.name = "Separate XYZ.032"
    
    #node Combine XYZ.005
    combine_xyz_005 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_005.label = "r22.xyz"
    combine_xyz_005.name = "Combine XYZ.005"
    
    #node Combine XYZ.006
    combine_xyz_006 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_006.label = "r27.xyz"
    combine_xyz_006.name = "Combine XYZ.006"
    
    #node Separate XYZ.017
    separate_xyz_017 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_017.name = "Separate XYZ.017"
    
    #node Value.001
    value_001 = HD2_Shader.nodes.new("ShaderNodeValue")
    value_001.label = "r24.z (#508)"
    value_001.name = "Value.001"
    
    value_001.outputs[0].default_value = 0.0
    #node Mix.023
    Slot3 = HD2_Shader.nodes.new("ShaderNodeMix")
    Slot3.label = "Slot 3"
    Slot3.name = "Mix.023"
    Slot3.blend_type = 'MIX'
    Slot3.clamp_factor = True
    Slot3.clamp_result = False
    Slot3.data_type = 'FLOAT'
    Slot3.factor_mode = 'UNIFORM'
        
    #node customization_material_detail_tiler_array
    customization_material_detail_tiler_array = HD2_Shader.nodes.new("ShaderNodeTexImage")
    customization_material_detail_tiler_array.label = "r14.xyzw"
    customization_material_detail_tiler_array.name = "customization_material_detail_tiler_array"
    customization_material_detail_tiler_array.extension = 'EXTEND'
    customization_material_detail_tiler_array.image_user.frame_current = 0
    customization_material_detail_tiler_array.image_user.frame_duration = 1
    customization_material_detail_tiler_array.image_user.frame_offset = 5
    customization_material_detail_tiler_array.image_user.frame_start = 1
    customization_material_detail_tiler_array.image_user.tile = 0
    customization_material_detail_tiler_array.image_user.use_auto_refresh = False
    customization_material_detail_tiler_array.image_user.use_cyclic = False
    customization_material_detail_tiler_array.interpolation = 'Smart'
    customization_material_detail_tiler_array.projection = 'FLAT'
    customization_material_detail_tiler_array.projection_blend = 0.0
    try:
        customization_material_detail_tiler_array.image = bpy.data.images.get("customization_material_detail_tiler_array.png")
        customization_material_detail_tiler_array.image.colorspace_settings.name = "Non-Color"
        customization_material_detail_tiler_array.image.alpha_mode = "CHANNEL_PACKED"
    except:
        pass
    
    #node Mix.039
    mix_039 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_039.name = "Mix.039"
    mix_039.blend_type = 'MIX'
    mix_039.clamp_factor = True
    mix_039.clamp_result = False
    mix_039.data_type = 'RGBA'
    mix_039.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_039.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_039.inputs[2].default_value = 0.0
    #B_Float
    mix_039.inputs[3].default_value = 0.0
    #A_Vector
    mix_039.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_039.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_039.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_039.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.060
    combine_xyz_060 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_060.name = "Combine XYZ.060"
    #X
    combine_xyz_060.inputs[0].default_value = 0.1818181872367859
    #Z
    combine_xyz_060.inputs[2].default_value = 0.0
    
    #node Combine XYZ.061
    combine_xyz_061 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_061.name = "Combine XYZ.061"
    #X
    combine_xyz_061.inputs[0].default_value = 0.1818181872367859
    #Z
    combine_xyz_061.inputs[2].default_value = 0.0
    
    #node Combine XYZ.041
    combine_xyz_041 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_041.name = "Combine XYZ.041"
    #Z
    combine_xyz_041.inputs[2].default_value = 0.0
    
    #node Math.239
    math_239 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_239.name = "Math.239"
    math_239.operation = 'MULTIPLY_ADD'
    math_239.use_clamp = False
    #Value_001
    math_239.inputs[1].default_value = 10000000.0
    #Value_002
    math_239.inputs[2].default_value = -4960000.0
    
    #node Mix.072
    mix_072 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_072.name = "Mix.072"
    mix_072.blend_type = 'MIX'
    mix_072.clamp_factor = True
    mix_072.clamp_result = False
    mix_072.data_type = 'FLOAT'
    mix_072.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_072.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_072.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_072.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_072.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_072.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_072.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_072.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Separate XYZ.009
    separate_xyz_009 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_009.name = "Separate XYZ.009"
    
    #node Math.082
    math_082 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_082.name = "Math.082"
    math_082.operation = 'COMPARE'
    math_082.use_clamp = False
    #Value_001
    math_082.inputs[1].default_value = 0.0
    #Value_002
    math_082.inputs[2].default_value = 0.0
    
    #node Combine XYZ.039
    combine_xyz_039 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_039.name = "Combine XYZ.039"
    
    #node Combine XYZ.024
    combine_xyz_024 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_024.name = "Combine XYZ.024"
    
    #node Combine XYZ.030
    combine_xyz_030 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_030.label = "r5.xx"
    combine_xyz_030.name = "Combine XYZ.030"
    #Z
    combine_xyz_030.inputs[2].default_value = 0.0
    
    #node Combine XYZ.031
    combine_xyz_031 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_031.label = "r21.xy"
    combine_xyz_031.name = "Combine XYZ.031"
    #Z
    combine_xyz_031.inputs[2].default_value = 0.0
    
    #node Math.249
    math_249 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_249.name = "Math.249"
    math_249.operation = 'MULTIPLY'
    math_249.use_clamp = False
    #Value_002
    math_249.inputs[2].default_value = 0.5
    
    #node Vector Math.040
    vector_math_040 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_040.label = "r10.xyw (#477)"
    vector_math_040.name = "Vector Math.040"
    vector_math_040.operation = 'ADD'
    #Vector_002
    vector_math_040.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_040.inputs[3].default_value = 1.0
    
    #node Combine XYZ.019
    combine_xyz_019 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_019.label = "r24.xyz"
    combine_xyz_019.name = "Combine XYZ.019"
    
    #node Mix.018
    mix_018 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_018.name = "Mix.018"
    mix_018.blend_type = 'MIX'
    mix_018.clamp_factor = True
    mix_018.clamp_result = False
    mix_018.data_type = 'RGBA'
    mix_018.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_018.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_018.inputs[2].default_value = 0.0
    #B_Float
    mix_018.inputs[3].default_value = 0.0
    #A_Vector
    mix_018.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_018.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_018.inputs[6].default_value = (0.0, 0.0, 0.0, 1.0)
    #A_Rotation
    mix_018.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_018.inputs[9].default_value = (0.0, 0.0, 0.0)
        
    #node Mix.025
    Slot4 = HD2_Shader.nodes.new("ShaderNodeMix")
    Slot4.label = "Slot 4"
    Slot4.name = "Mix.025"
    Slot4.blend_type = 'MIX'
    Slot4.clamp_factor = True
    Slot4.clamp_result = False
    Slot4.data_type = 'FLOAT'
    Slot4.factor_mode = 'UNIFORM'
    
    #node Primary Material LUT_04
    primary_material_lut_04 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_04.label = "Primary Material LUT_04"
    primary_material_lut_04.name = "Primary Material LUT_04"
    primary_material_lut_04.use_custom_color = True
    primary_material_lut_04.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_04.extension = 'EXTEND'
    primary_material_lut_04.image_user.frame_current = 1
    primary_material_lut_04.image_user.frame_duration = 1
    primary_material_lut_04.image_user.frame_offset = 429
    primary_material_lut_04.image_user.frame_start = 1
    primary_material_lut_04.image_user.tile = 0
    primary_material_lut_04.image_user.use_auto_refresh = False
    primary_material_lut_04.image_user.use_cyclic = False
    primary_material_lut_04.interpolation = 'Closest'
    primary_material_lut_04.projection = 'FLAT'
    primary_material_lut_04.projection_blend = 0.0
    
    #node Vector Math.102
    vector_math_102 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_102.label = "r19.xyz (#325)"
    vector_math_102.name = "Vector Math.102"
    vector_math_102.operation = 'ADD'
    #Vector_002
    vector_math_102.inputs[2].default_value = (-0.5, 0.39989998936653137, 0.0)
    #Scale
    vector_math_102.inputs[3].default_value = 1.0
    
    #node Separate XYZ.043
    separate_xyz_043 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_043.name = "Separate XYZ.043"
    
    #node Math.240
    math_240 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_240.name = "Math.240"
    math_240.operation = 'MULTIPLY_ADD'
    math_240.use_clamp = False
    #Value_001
    math_240.inputs[1].default_value = 10000000.0
    #Value_002
    math_240.inputs[2].default_value = -4960000.0
    
    #node Combine XYZ.054
    combine_xyz_054 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_054.name = "Combine XYZ.054"
    #X
    combine_xyz_054.inputs[0].default_value = 0.04545454680919647
    #Z
    combine_xyz_054.inputs[2].default_value = 0.0
    
    #node Combine XYZ.055
    combine_xyz_055 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_055.name = "Combine XYZ.055"
    #X
    combine_xyz_055.inputs[0].default_value = 0.04545454680919647
    #Z
    combine_xyz_055.inputs[2].default_value = 0.0
    
    #node Math.083
    math_083 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_083.label = "if (r5.x != 0)"
    math_083.name = "Math.083"
    math_083.operation = 'SUBTRACT'
    math_083.use_clamp = False
    #Value
    math_083.inputs[0].default_value = 1.0
    #Value_002
    math_083.inputs[2].default_value = 0.5
    
    #node Combine XYZ.045
    combine_xyz_045 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_045.label = "r14.xy (#261)"
    combine_xyz_045.name = "Combine XYZ.045"
    combine_xyz_045.inputs[2].hide = True
    #X
    combine_xyz_045.inputs[0].default_value = 0.0
    #Y
    combine_xyz_045.inputs[1].default_value = 0.0
    #Z
    combine_xyz_045.inputs[2].default_value = 0.0
    
    #node Combine XYZ.046
    combine_xyz_046 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_046.label = "r19.xyz (#339)"
    combine_xyz_046.name = "Combine XYZ.046"
    
    #node Vector Math.027
    vector_math_027 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_027.label = "r22.xy (#391)"
    vector_math_027.name = "Vector Math.027"
    vector_math_027.operation = 'MULTIPLY'
    #Vector_002
    vector_math_027.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_027.inputs[3].default_value = 1.0
    
    #node Separate XYZ.037
    separate_xyz_037 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_037.name = "Separate XYZ.037"
    
    #node Math.248
    math_248 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_248.name = "Math.248"
    math_248.operation = 'MULTIPLY_ADD'
    math_248.use_clamp = False
    
    #node Math.031
    math_031 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_031.label = "-r6.w"
    math_031.name = "Math.031"
    math_031.operation = 'MULTIPLY'
    math_031.use_clamp = False
    #Value_001
    math_031.inputs[1].default_value = -1.0
    #Value_002
    math_031.inputs[2].default_value = 0.5
    
    #node Vector Math.041
    vector_math_041 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_041.label = "r10.x (#478)"
    vector_math_041.name = "Vector Math.041"
    vector_math_041.operation = 'DOT_PRODUCT'
    #Vector_002
    vector_math_041.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_041.inputs[3].default_value = 1.0
    
    #node Combine XYZ.003
    combine_xyz_003 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_003.name = "Combine XYZ.003"
    #X
    combine_xyz_003.inputs[0].default_value = 0.5
    #Y
    combine_xyz_003.inputs[1].default_value = 0.0
    #Z
    combine_xyz_003.inputs[2].default_value = 0.0
    
    #node Vector Math.017
    vector_math_017 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_017.label = "r10.xyw (#509)"
    vector_math_017.name = "Vector Math.017"
    vector_math_017.operation = 'ADD'
    #Vector_002
    vector_math_017.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_017.inputs[3].default_value = 1.0
    
    #node ID Mask Array 02
    id_mask_array_02 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    id_mask_array_02.label = "ID Mask Array 02"
    id_mask_array_02.name = "ID Mask Array 02"
    id_mask_array_02.use_custom_color = True
    id_mask_array_02.color = (0.6000000238418579, 0.0, 0.0)
    id_mask_array_02.extension = 'REPEAT'
    id_mask_array_02.image_user.frame_current = 1
    id_mask_array_02.image_user.frame_duration = 1
    id_mask_array_02.image_user.frame_offset = 6360
    id_mask_array_02.image_user.frame_start = 1
    id_mask_array_02.image_user.tile = 0
    id_mask_array_02.image_user.use_auto_refresh = False
    id_mask_array_02.image_user.use_cyclic = False
    id_mask_array_02.interpolation = 'Linear'
    id_mask_array_02.projection = 'FLAT'
    id_mask_array_02.projection_blend = 0.0

    #node Mix.024
    Slot5 = HD2_Shader.nodes.new("ShaderNodeMix")
    Slot5.label = "Slot 5"
    Slot5.name = "Mix.024"
    Slot5.blend_type = 'MIX'
    Slot5.clamp_factor = True
    Slot5.clamp_result = False
    Slot5.data_type = 'FLOAT'
    Slot5.factor_mode = 'UNIFORM'
    
    #node Vector Math.084
    vector_math_084 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_084.name = "Vector Math.084"
    vector_math_084.operation = 'ADD'
    #Vector
    vector_math_084.inputs[0].default_value = (-0.5, -0.5, -0.5)
    #Vector_002
    vector_math_084.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_084.inputs[3].default_value = 1.0
    
    #node Mix.040
    mix_040 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_040.name = "Mix.040"
    mix_040.blend_type = 'MIX'
    mix_040.clamp_factor = True
    mix_040.clamp_result = False
    mix_040.data_type = 'FLOAT'
    mix_040.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_040.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_040.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_040.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_040.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_040.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_040.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_040.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Separate XYZ.005
    separate_xyz_005 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_005.name = "Separate XYZ.005"
    
    #node Math.241
    math_241 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_241.name = "Math.241"
    math_241.operation = 'MULTIPLY_ADD'
    math_241.use_clamp = False
    #Value_001
    math_241.inputs[1].default_value = 10000000.0
    #Value_002
    math_241.inputs[2].default_value = -4960000.0
    
    #node Primary Material LUT_01
    primary_material_lut_01 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_01.label = "Primary Material LUT_01"
    primary_material_lut_01.name = "Primary Material LUT_01"
    primary_material_lut_01.use_custom_color = True
    primary_material_lut_01.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_01.extension = 'EXTEND'
    primary_material_lut_01.image_user.frame_current = 1
    primary_material_lut_01.image_user.frame_duration = 1
    primary_material_lut_01.image_user.frame_offset = 429
    primary_material_lut_01.image_user.frame_start = 1
    primary_material_lut_01.image_user.tile = 0
    primary_material_lut_01.image_user.use_auto_refresh = False
    primary_material_lut_01.image_user.use_cyclic = False
    primary_material_lut_01.interpolation = 'Closest'
    primary_material_lut_01.projection = 'FLAT'
    primary_material_lut_01.projection_blend = 0.0
    
    #node Mix.005
    mix_005 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_005.label = "r14.xy if (r5.x != 0)"
    mix_005.name = "Mix.005"
    mix_005.blend_type = 'MIX'
    mix_005.clamp_factor = True
    mix_005.clamp_result = False
    mix_005.data_type = 'VECTOR'
    mix_005.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_005.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_005.inputs[2].default_value = 0.0
    #B_Float
    mix_005.inputs[3].default_value = 0.0
    #A_Color
    mix_005.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_005.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_005.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_005.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.233
    math_233 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_233.label = "r19.w (#325)"
    math_233.name = "Math.233"
    math_233.operation = 'ADD'
    math_233.use_clamp = False
    #Value
    math_233.inputs[0].default_value = 0.0
    #Value_002
    math_233.inputs[2].default_value = 0.5
    
    #node Separate XYZ.033
    separate_xyz_033 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_033.name = "Separate XYZ.033"
    
    #node Math.234
    math_234 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_234.name = "Math.234"
    math_234.operation = 'MULTIPLY_ADD'
    math_234.use_clamp = False
    
    #node Math.030
    math_030 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_030.name = "Math.030"
    math_030.operation = 'ADD'
    math_030.use_clamp = False
    #Value_002
    math_030.inputs[2].default_value = 0.5
    
    #node Vector Math.018
    vector_math_018 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_018.label = "r7.z (#510)"
    vector_math_018.name = "Vector Math.018"
    vector_math_018.operation = 'DOT_PRODUCT'
    #Vector_002
    vector_math_018.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_018.inputs[3].default_value = 1.0
    
    #node Separate XYZ.059
    separate_xyz_059 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_059.name = "Separate XYZ.059"
    
    #node Mix.019
    mix_019 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_019.name = "Mix.019"
    mix_019.blend_type = 'MIX'
    mix_019.clamp_factor = True
    mix_019.clamp_result = False
    mix_019.data_type = 'FLOAT'
    mix_019.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_019.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_019.inputs[2].default_value = 0.0
    #A_Vector
    mix_019.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_019.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_019.inputs[6].default_value = (0.0, 0.0, 0.0, 1.0)
    #B_Color
    mix_019.inputs[7].default_value = (0.0, 0.0, 0.0, 1.0)
    #A_Rotation
    mix_019.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_019.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node UV Map
    uv_map = HD2_Shader.nodes.new("ShaderNodeUVMap")
    uv_map.name = "Base UV"
    uv_map.from_instancer = False
    uv_map.uv_map = "UVs for Baking"
    
    #node Mix.027
    Slot6 = HD2_Shader.nodes.new("ShaderNodeMix")
    Slot6.label = "Slot 6"
    Slot6.name = "Mix.027"
    Slot6.blend_type = 'MIX'
    Slot6.clamp_factor = True
    Slot6.clamp_result = False
    Slot6.data_type = 'FLOAT'
    Slot6.factor_mode = 'UNIFORM'
        
    #node pattern_lut 02
    pattern_lut_02 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    pattern_lut_02.label = "r23.xyzw 02"
    pattern_lut_02.name = "pattern_lut 02"
    pattern_lut_02.extension = 'REPEAT'
    pattern_lut_02.image_user.frame_current = 0
    pattern_lut_02.image_user.frame_duration = 1
    pattern_lut_02.image_user.frame_offset = 495
    pattern_lut_02.image_user.frame_start = 1
    pattern_lut_02.image_user.tile = 0
    pattern_lut_02.image_user.use_auto_refresh = False
    pattern_lut_02.image_user.use_cyclic = False
    pattern_lut_02.interpolation = 'Closest'
    pattern_lut_02.projection = 'FLAT'
    pattern_lut_02.projection_blend = 0.0

    #node Math.242
    math_242 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_242.name = "Math.242"
    math_242.operation = 'MULTIPLY_ADD'
    math_242.use_clamp = False
    #Value_001
    math_242.inputs[1].default_value = 10000000.0
    #Value_002
    math_242.inputs[2].default_value = -4960000.0
    
    #node Combine XYZ.094
    combine_xyz_094 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_094.name = "Combine XYZ.094"
    #X
    combine_xyz_094.inputs[0].default_value = 0.9545454382896423
    #Z
    combine_xyz_094.inputs[2].default_value = 0.0
    
    #node Combine XYZ.095
    combine_xyz_095 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_095.name = "Combine XYZ.095"
    #X
    combine_xyz_095.inputs[0].default_value = 0.9545454382896423
    #Z
    combine_xyz_095.inputs[2].default_value = 0.0
    
    #node Combine XYZ.052
    combine_xyz_052 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_052.name = "Combine XYZ.052"
    #X
    combine_xyz_052.inputs[0].default_value = 0.0
    #Z
    combine_xyz_052.inputs[2].default_value = 0.0
    
    #node Combine XYZ.053
    combine_xyz_053 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_053.name = "Combine XYZ.053"
    #X
    combine_xyz_053.inputs[0].default_value = 0.0
    #Z
    combine_xyz_053.inputs[2].default_value = 0.0
    
    #node Combine XYZ.067
    combine_xyz_067 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_067.name = "Combine XYZ.067"
    #X
    combine_xyz_067.inputs[0].default_value = 0.3181818127632141
    #Z
    combine_xyz_067.inputs[2].default_value = 0.0
    
    #node Mix.045
    mix_045 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_045.name = "Mix.045"
    mix_045.blend_type = 'MIX'
    mix_045.clamp_factor = True
    mix_045.clamp_result = False
    mix_045.data_type = 'RGBA'
    mix_045.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_045.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_045.inputs[2].default_value = 0.0
    #B_Float
    mix_045.inputs[3].default_value = 0.0
    #A_Vector
    mix_045.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_045.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Rotation
    mix_045.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_045.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.066
    combine_xyz_066 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_066.name = "Combine XYZ.066"
    #X
    combine_xyz_066.inputs[0].default_value = 0.3181818127632141
    #Z
    combine_xyz_066.inputs[2].default_value = 0.0
    
    #node Mix.034
    mix_034 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_034.name = "Mix.034"
    mix_034.blend_type = 'MIX'
    mix_034.clamp_factor = True
    mix_034.clamp_result = False
    mix_034.data_type = 'FLOAT'
    mix_034.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_034.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_034.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_034.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_034.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_034.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_034.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_034.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Separate XYZ.022
    separate_xyz_022 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_022.name = "Separate XYZ.022"
    
    #node Combine XYZ.032
    combine_xyz_032 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_032.name = "Combine XYZ.032"
    #Z
    combine_xyz_032.inputs[2].default_value = 0.0
    
    #node Separate XYZ.069
    separate_xyz_069 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_069.name = "Separate XYZ.069"
    
    #node Clamp.008
    clamp_008 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_008.label = "r6.w (#479)"
    clamp_008.name = "Clamp.008"
    clamp_008.clamp_type = 'MINMAX'
    #Min
    clamp_008.inputs[1].default_value = 0.0
    #Max
    clamp_008.inputs[2].default_value = 1.0
    
    #node Math.119
    math_119 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_119.label = "r7.y (#517)"
    math_119.name = "Math.119"
    math_119.operation = 'ADD'
    math_119.use_clamp = False
    #Value_001
    math_119.inputs[1].default_value = -0.5
    #Value_002
    math_119.inputs[2].default_value = 0.5
    
    #node Math.110
    math_110 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_110.name = "Math.110"
    math_110.operation = 'ADD'
    math_110.use_clamp = False
    #Value_002
    math_110.inputs[2].default_value = 0.5
    
    #node Combine XYZ.002
    combine_xyz_002 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_002.name = "Combine XYZ.002"
    #X
    combine_xyz_002.inputs[0].default_value = 0.0
    #Y
    combine_xyz_002.inputs[1].default_value = 0.0
    #Z
    combine_xyz_002.inputs[2].default_value = 0.0
    
    #node Mapping
    mapping = HD2_Shader.nodes.new("ShaderNodeMapping")
    mapping.name = "Mapping"
    mapping.vector_type = 'POINT'
    #Location
    mapping.inputs[1].default_value = (0.0, 0.0, 0.0)
    #Rotation
    mapping.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    mapping.inputs[3].default_value = (1.0, 0.5, 1.0)
    
    #node Mix.026
    Slot7 = HD2_Shader.nodes.new("ShaderNodeMix")
    Slot7.label = "Slot 7"
    Slot7.name = "Mix.026"
    Slot7.blend_type = 'MIX'
    Slot7.clamp_factor = True
    Slot7.clamp_result = False
    Slot7.data_type = 'FLOAT'
    Slot7.factor_mode = 'UNIFORM'

    update_slot_defaults(HD2_Shader, material)

    #node Math.250
    math_250 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_250.name = "Math.250"
    math_250.operation = 'MULTIPLY_ADD'
    math_250.use_clamp = False
    
    #node Primary Material LUT_21
    primary_material_lut_21 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_21.label = "Primary Material LUT_21"
    primary_material_lut_21.name = "Primary Material LUT_21"
    primary_material_lut_21.use_custom_color = True
    primary_material_lut_21.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_21.extension = 'EXTEND'
    primary_material_lut_21.image_user.frame_current = 1
    primary_material_lut_21.image_user.frame_duration = 1
    primary_material_lut_21.image_user.frame_offset = 429
    primary_material_lut_21.image_user.frame_start = 1
    primary_material_lut_21.image_user.tile = 0
    primary_material_lut_21.image_user.use_auto_refresh = False
    primary_material_lut_21.image_user.use_cyclic = False
    primary_material_lut_21.interpolation = 'Closest'
    primary_material_lut_21.projection = 'FLAT'
    primary_material_lut_21.projection_blend = 0.0
    
    #node Primary Material LUT_00
    primary_material_lut_00 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_00.label = "Primary Material LUT_00"
    primary_material_lut_00.name = "Primary Material LUT_00"
    primary_material_lut_00.use_custom_color = True
    primary_material_lut_00.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_00.extension = 'EXTEND'
    primary_material_lut_00.image_user.frame_current = 1
    primary_material_lut_00.image_user.frame_duration = 1
    primary_material_lut_00.image_user.frame_offset = 429
    primary_material_lut_00.image_user.frame_start = 1
    primary_material_lut_00.image_user.tile = 0
    primary_material_lut_00.image_user.use_auto_refresh = False
    primary_material_lut_00.image_user.use_cyclic = False
    primary_material_lut_00.interpolation = 'Closest'
    
    #node Primary Material LUT_07
    primary_material_lut_07 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_07.label = "Primary Material LUT_07"
    primary_material_lut_07.name = "Primary Material LUT_07"
    primary_material_lut_07.use_custom_color = True
    primary_material_lut_07.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_07.extension = 'EXTEND'
    primary_material_lut_07.image_user.frame_current = 1
    primary_material_lut_07.image_user.frame_duration = 1
    primary_material_lut_07.image_user.frame_offset = 429
    primary_material_lut_07.image_user.frame_start = 1
    primary_material_lut_07.image_user.tile = 0
    primary_material_lut_07.image_user.use_auto_refresh = False
    primary_material_lut_07.image_user.use_cyclic = False
    primary_material_lut_07.interpolation = 'Closest'
    primary_material_lut_07.projection = 'FLAT'
    primary_material_lut_07.projection_blend = 0.0
    
    #node Combine XYZ.023
    combine_xyz_023 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_023.label = "r7.wz"
    combine_xyz_023.name = "Combine XYZ.023"
    #Z
    combine_xyz_023.inputs[2].default_value = 0.0
    
    #node Math.067
    math_067 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_067.name = "Math.067"
    math_067.operation = 'ADD'
    math_067.use_clamp = False
    #Value_002
    math_067.inputs[2].default_value = 0.5
    
    #node Vector Math.104
    vector_math_104 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_104.label = "r21.xyz (#393)"
    vector_math_104.name = "Vector Math.104"
    vector_math_104.operation = 'ADD'
    #Vector_002
    vector_math_104.inputs[2].default_value = (0.0, 0.0, 0.0)
    #Scale
    vector_math_104.inputs[3].default_value = 1.0
    
    #node Math.259
    math_259 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_259.name = "Math.259"
    math_259.operation = 'MULTIPLY'
    math_259.use_clamp = False
    #Value_002
    math_259.inputs[2].default_value = 0.5
    
    #node Combine XYZ.033
    combine_xyz_033 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_033.name = "Combine XYZ.033"
    
    #node Math.033
    math_033 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_033.name = "Math.033"
    math_033.operation = 'MULTIPLY'
    math_033.use_clamp = False
    #Value_001
    math_033.inputs[1].default_value = -1.0
    #Value_002
    math_033.inputs[2].default_value = 0.5
    
    #node Math.120
    math_120 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_120.label = "r7.y (#518)"
    math_120.name = "Math.120"
    math_120.operation = 'MULTIPLY'
    math_120.use_clamp = True
    #Value_001
    math_120.inputs[1].default_value = 100.0
    #Value_002
    math_120.inputs[2].default_value = 0.5
    
    #node Clamp.014
    clamp_014 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_014.label = "r7.z (#511)"
    clamp_014.name = "Clamp.014"
    clamp_014.clamp_type = 'MINMAX'
    #Min
    clamp_014.inputs[1].default_value = 0.0
    #Max
    clamp_014.inputs[2].default_value = 1.0
    
    #node Pattern Mask Array 02
    pattern_mask_array_02 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    pattern_mask_array_02.label = "Pattern Mask Array 02"
    pattern_mask_array_02.name = "Pattern Mask Array 02"
    pattern_mask_array_02.use_custom_color = True
    pattern_mask_array_02.color = (0.6000000238418579, 0.6000000238418579, 0.6000000238418579)
    pattern_mask_array_02.extension = 'REPEAT'
    pattern_mask_array_02.image_user.frame_current = 1
    pattern_mask_array_02.image_user.frame_duration = 1
    pattern_mask_array_02.image_user.frame_offset = 6360
    pattern_mask_array_02.image_user.frame_start = 1
    pattern_mask_array_02.image_user.tile = 0
    pattern_mask_array_02.image_user.use_auto_refresh = False
    pattern_mask_array_02.image_user.use_cyclic = False
    pattern_mask_array_02.interpolation = 'Linear'
    pattern_mask_array_02.projection = 'FLAT'
    pattern_mask_array_02.projection_blend = 0.0

    #node Mix.029
    Slot8 = HD2_Shader.nodes.new("ShaderNodeMix")
    Slot8.label = "Slot 8"
    Slot8.name = "Mix.029"
    Slot8.blend_type = 'MIX'
    Slot8.clamp_factor = True
    Slot8.clamp_result = False
    Slot8.data_type = 'FLOAT'
    Slot8.factor_mode = 'UNIFORM'
    try:
        Slot8.inputs[3].default_value = (1-(0.5/PrimaryMaterialLUTSizeY*15))
        if PrimaryMaterialLUTSizeY < 7.1:
            Slot8.mute = True
    except:
        pass
    
    #node pattern_lut
    pattern_lut = HD2_Shader.nodes.new("ShaderNodeTexImage")
    pattern_lut.label = "r22.xyzw"
    pattern_lut.name = "pattern_lut"
    pattern_lut.extension = 'EXTEND'
    pattern_lut.interpolation = 'Closest'
        
    gamma_pattern = HD2_Shader.nodes.new("ShaderNodeGamma")
    gamma_pattern.name = "gamma pattern"
    #Gamma
    gamma_pattern.inputs[1].default_value = 2.2
    
    #node Mix.074
    mix_074 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_074.name = "Mix.074"
    mix_074.blend_type = 'MIX'
    mix_074.clamp_factor = True
    mix_074.clamp_result = False
    mix_074.data_type = 'FLOAT'
    mix_074.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_074.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_074.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_074.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_074.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_074.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_074.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_074.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Mix.031
    mix_031 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_031.name = "Mix.031"
    mix_031.blend_type = 'MIX'
    mix_031.clamp_factor = True
    mix_031.clamp_result = False
    mix_031.data_type = 'FLOAT'
    mix_031.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_031.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_031.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_031.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_031.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_031.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_031.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_031.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Combine XYZ.064
    combine_xyz_064 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_064.name = "Combine XYZ.064"
    #X
    combine_xyz_064.inputs[0].default_value = 0.27272728085517883
    #Z
    combine_xyz_064.inputs[2].default_value = 0.0
    
    #node Combine XYZ.065
    combine_xyz_065 = HD2_Shader.nodes.new("ShaderNodeCombineXYZ")
    combine_xyz_065.name = "Combine XYZ.065"
    #X
    combine_xyz_065.inputs[0].default_value = 0.27272728085517883
    #Z
    combine_xyz_065.inputs[2].default_value = 0.0
    
    #node Mix.046
    mix_046 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_046.name = "Mix.046"
    mix_046.blend_type = 'MIX'
    mix_046.clamp_factor = True
    mix_046.clamp_result = False
    mix_046.data_type = 'FLOAT'
    mix_046.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_046.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_046.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_046.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_046.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_046.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_046.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_046.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Vector Math.020
    vector_math_020 = HD2_Shader.nodes.new("ShaderNodeVectorMath")
    vector_math_020.label = "r14.zw (#326)"
    vector_math_020.name = "Vector Math.020"
    vector_math_020.operation = 'MULTIPLY_ADD'
    #Vector_001
    vector_math_020.inputs[1].default_value = (1.0, -0.4000000059604645, 0.0)
    #Vector_002
    vector_math_020.inputs[2].default_value = (-0.5, 0.39989998936653137, 0.0)
    #Scale
    vector_math_020.inputs[3].default_value = 1.0
    
    #node Clamp.005
    clamp_005 = HD2_Shader.nodes.new("ShaderNodeClamp")
    clamp_005.label = "r4.w (#328)"
    clamp_005.name = "Clamp.005"
    clamp_005.clamp_type = 'MINMAX'
    #Min
    clamp_005.inputs[1].default_value = 0.0
    #Max
    clamp_005.inputs[2].default_value = 1.0
    
    #node Math.258
    math_258 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_258.name = "Math.258"
    math_258.operation = 'MULTIPLY_ADD'
    math_258.use_clamp = False
    
    #node Separate XYZ.034
    separate_xyz_034 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_034.name = "Separate XYZ.034"
    
    #node Separate XYZ.049
    separate_xyz_049 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_049.name = "Separate XYZ.049"
    
    #node Math.032
    math_032 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_032.label = "r10.x (#482)"
    math_032.name = "Math.032"
    math_032.operation = 'ADD'
    math_032.use_clamp = False
    #Value
    math_032.inputs[0].default_value = 1.0
    #Value_002
    math_032.inputs[2].default_value = 0.5
    
    #node Math.115
    math_115 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_115.label = "r7.w (#519)"
    math_115.name = "Math.115"
    math_115.operation = 'MULTIPLY'
    math_115.use_clamp = False
    #Value_002
    math_115.inputs[2].default_value = 0.5
    
    #node Math.036
    math_036 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_036.name = "Math.036"
    math_036.operation = 'COMPARE'
    math_036.use_clamp = False
    #Value_001
    math_036.inputs[1].default_value = -1.0
    #Value_002
    math_036.inputs[2].default_value = 0.0
    
    #node Math.121
    math_121 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_121.name = "Math.121"
    math_121.operation = 'COMPARE'
    math_121.use_clamp = False
    #Value_001
    math_121.inputs[1].default_value = 1.0
    #Value_002
    math_121.inputs[2].default_value = 0.0
    
    #node Math.126
    math_126 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_126.label = "-r7.w"
    math_126.name = "Math.126"
    math_126.operation = 'MULTIPLY'
    math_126.use_clamp = False
    #Value_001
    math_126.inputs[1].default_value = -1.0
    #Value_002
    math_126.inputs[2].default_value = 0.5
    
    #node Math.116
    math_116 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_116.label = "r10.x (#520)"
    math_116.name = "Math.116"
    math_116.operation = 'LESS_THAN'
    math_116.use_clamp = False
    #Value
    math_116.inputs[0].default_value = 0.0
    #Value_002
    math_116.inputs[2].default_value = 0.5
    
    #node Math.034
    math_034 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_034.name = "Math.034"
    math_034.operation = 'COMPARE'
    math_034.use_clamp = False
    #Value_001
    math_034.inputs[1].default_value = 3.0
    #Value_002
    math_034.inputs[2].default_value = 0.0
    
    #node Mix.030
    mix_030 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_030.name = "Mix.030"
    mix_030.blend_type = 'MIX'
    mix_030.clamp_factor = True
    mix_030.clamp_result = False
    mix_030.data_type = 'FLOAT'
    mix_030.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_030.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_030.inputs[2].default_value = 0.0
    #A_Vector
    mix_030.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_030.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_030.inputs[6].default_value = (0.0, 0.0, 0.0, 1.0)
    #B_Color
    mix_030.inputs[7].default_value = (0.0, 0.0, 0.0, 1.0)
    #A_Rotation
    mix_030.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_030.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.243
    math_243 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_243.name = "Math.243"
    math_243.operation = 'MULTIPLY_ADD'
    math_243.use_clamp = False
    math_243.inputs[1].default_value = 0.000
    math_243.inputs[2].default_value = 0.000

#Pattern Mask 2 disabled for now till we learn more about it

#    try:
#        IDMaskArraySizeX = material.node_tree.nodes['Secondary Material LUT Texture'].inputs[0].node.image.size[0]
#        IDMaskArraySizeY = material.node_tree.nodes['Secondary Material LUT Texture'].inputs[0].node.image.size[1]
#        if IDMaskArraySizeY == 1:
#            math_243.inputs[1].default_value = 10000000.0
#            math_243.inputs[2].default_value = -4960000.0
#    except:
#        pass
    
    #node Separate XYZ.023
    separate_xyz_023 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_023.name = "Separate XYZ.023"
    
    #node Math.256
    math_256 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_256.label = "r21.w (#393)"
    math_256.name = "Math.256"
    math_256.operation = 'ADD'
    math_256.use_clamp = False
    #Value
    math_256.inputs[0].default_value = 0.0
    #Value_002
    math_256.inputs[2].default_value = 0.5
    
    #node Math.257
    math_257 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_257.name = "Math.257"
    math_257.operation = 'MULTIPLY_ADD'
    math_257.use_clamp = False
    
    #node Math.020
    math_020 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_020.name = "Math.020"
    math_020.operation = 'LESS_THAN'
    math_020.use_clamp = False
    #Value_001
    math_020.inputs[1].default_value = 0.0
    #Value_002
    math_020.inputs[2].default_value = 0.0
    
    #node Math.089
    math_089 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_089.name = "Math.089"
    math_089.operation = 'MULTIPLY'
    math_089.use_clamp = False
    #Value_001
    math_089.inputs[1].default_value = -1.0
    #Value_002
    math_089.inputs[2].default_value = 0.5
    
    #node Math.023
    math_023 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_023.name = "Math.023"
    math_023.operation = 'COMPARE'
    math_023.use_clamp = False
    #Value_001
    math_023.inputs[1].default_value = 1.0
    #Value_002
    math_023.inputs[2].default_value = 0.0
    
    #node Math.035
    math_035 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_035.label = "r7.y (#496)"
    math_035.name = "Math.035"
    math_035.operation = 'SUBTRACT'
    math_035.use_clamp = False
    #Value
    math_035.inputs[0].default_value = 1.0
    #Value_002
    math_035.inputs[2].default_value = 0.0
    
    #node Math.037
    math_037 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_037.label = "r7.y (#499)"
    math_037.name = "Math.037"
    math_037.operation = 'SUBTRACT'
    math_037.use_clamp = False
    #Value
    math_037.inputs[0].default_value = 1.0
    #Value_002
    math_037.inputs[2].default_value = 0.0
    
    #node Math.117
    math_117 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_117.name = "Math.117"
    math_117.operation = 'COMPARE'
    math_117.use_clamp = False
    #Value_001
    math_117.inputs[1].default_value = 0.0
    #Value_002
    math_117.inputs[2].default_value = 0.0
    
    #node Math.122
    math_122 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_122.label = "r10.x (#524)"
    math_122.name = "Math.122"
    math_122.operation = 'SUBTRACT'
    math_122.use_clamp = False
    #Value
    math_122.inputs[0].default_value = 1.0
    #Value_002
    math_122.inputs[2].default_value = 0.5
    
    #node Math.125
    math_125 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_125.label = "r10.x (#526)"
    math_125.name = "Math.125"
    math_125.operation = 'MULTIPLY_ADD'
    math_125.use_clamp = False
    
    #node Primary Material LUT_06
    primary_material_lut_06 = HD2_Shader.nodes.new("ShaderNodeTexImage")
    primary_material_lut_06.label = "Primary Material LUT_06"
    primary_material_lut_06.name = "Primary Material LUT_06"
    primary_material_lut_06.use_custom_color = True
    primary_material_lut_06.color = (0.0, 0.6000000238418579, 0.0)
    primary_material_lut_06.extension = 'EXTEND'
    primary_material_lut_06.image_user.frame_current = 1
    primary_material_lut_06.image_user.frame_duration = 1
    primary_material_lut_06.image_user.frame_offset = 429
    primary_material_lut_06.image_user.frame_start = 1
    primary_material_lut_06.image_user.tile = 0
    primary_material_lut_06.image_user.use_auto_refresh = False
    primary_material_lut_06.image_user.use_cyclic = False
    primary_material_lut_06.interpolation = 'Closest'
    primary_material_lut_06.projection = 'FLAT'
    primary_material_lut_06.projection_blend = 0.0
    
    #node Mix.044
    mix_044 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_044.name = "Mix.044"
    mix_044.blend_type = 'MIX'
    mix_044.clamp_factor = True
    mix_044.clamp_result = False
    mix_044.data_type = 'FLOAT'
    mix_044.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_044.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_044.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_044.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_044.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_044.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_044.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_044.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.260
    math_260 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_260.name = "Math.260"
    math_260.operation = 'MULTIPLY_ADD'
    math_260.use_clamp = False
    
    #node Math.021
    math_021 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_021.label = "r6.w (#442)"
    math_021.name = "Math.021"
    math_021.operation = 'SUBTRACT'
    math_021.use_clamp = False
    #Value
    math_021.inputs[0].default_value = 1.0
    #Value_002
    math_021.inputs[2].default_value = 0.0
    
    #node Math.090
    math_090 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_090.label = "r10.y (#483)"
    math_090.name = "Math.090"
    math_090.operation = 'ADD'
    math_090.use_clamp = False
    #Value_002
    math_090.inputs[2].default_value = 0.5
    
    #node Math.024
    math_024 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_024.label = "r10.x (#467)"
    math_024.name = "Math.024"
    math_024.operation = 'SUBTRACT'
    math_024.use_clamp = False
    #Value
    math_024.inputs[0].default_value = 1.0
    #Value_002
    math_024.inputs[2].default_value = 0.5
    
    #node Math.118
    math_118 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_118.label = "if (r10.x != 0) (#521)"
    math_118.name = "Math.118"
    math_118.operation = 'SUBTRACT'
    math_118.use_clamp = False
    #Value
    math_118.inputs[0].default_value = 1.0
    #Value_002
    math_118.inputs[2].default_value = 0.0
    
    #node Math.124
    math_124 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_124.name = "Math.124"
    math_124.operation = 'COMPARE'
    math_124.use_clamp = False
    #Value_001
    math_124.inputs[1].default_value = 1.0
    #Value_002
    math_124.inputs[2].default_value = 0.0
    
    #node Math.127
    math_127 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_127.label = "r10.x (#527)"
    math_127.name = "Math.127"
    math_127.operation = 'ADD'
    math_127.use_clamp = False
    #Value
    math_127.inputs[0].default_value = 1.0
    #Value_002
    math_127.inputs[2].default_value = 0.5
    
    #node Math.175
    math_175 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_175.name = "Math.175"
    math_175.operation = 'MULTIPLY'
    math_175.use_clamp = False
    #Value_002
    math_175.inputs[2].default_value = 0.5
    
    #node Math.063
    math_063 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_063.label = "r16.x (#395)"
    math_063.name = "Math.063"
    math_063.operation = 'ADD'
    math_063.use_clamp = True
    #Value_002
    math_063.inputs[2].default_value = 0.5
    
    #node Math.091
    math_091 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_091.label = "r16.x (#484)"
    math_091.name = "Math.091"
    math_091.operation = 'MAXIMUM'
    math_091.use_clamp = False
    #Value
    math_091.inputs[0].default_value = 0.0
    #Value_002
    math_091.inputs[2].default_value = 0.5
    
    #node Math.180
    math_180 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_180.name = "Math.180"
    math_180.operation = 'MULTIPLY'
    math_180.use_clamp = False
    #Value_002
    math_180.inputs[2].default_value = 0.5
    
    #node Math.123
    math_123 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_123.label = "if (r10.x != 0) (#525)"
    math_123.name = "Math.123"
    math_123.operation = 'SUBTRACT'
    math_123.use_clamp = False
    #Value
    math_123.inputs[0].default_value = 1.0
    #Value_002
    math_123.inputs[2].default_value = 0.5
    
    #node Math.128
    math_128 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_128.label = "r10.x (#528)"
    math_128.name = "Math.128"
    math_128.operation = 'MINIMUM'
    math_128.use_clamp = False
    #Value
    math_128.inputs[0].default_value = 1.0
    #Value_002
    math_128.inputs[2].default_value = 0.5
    
    #node Math.176
    math_176 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_176.name = "Math.176"
    math_176.operation = 'MULTIPLY'
    math_176.use_clamp = False
    #Value_002
    math_176.inputs[2].default_value = 0.5
    
    #node Mix.008
    mix_008 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_008.label = "r16.x if (r10.x != 0)"
    mix_008.name = "Mix.008"
    mix_008.blend_type = 'MIX'
    mix_008.clamp_factor = True
    mix_008.clamp_result = False
    mix_008.data_type = 'FLOAT'
    mix_008.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_008.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_008.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_008.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_008.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_008.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_008.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_008.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.129
    math_129 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_129.label = "r16.x (#529)"
    math_129.name = "Math.129"
    math_129.operation = 'MULTIPLY'
    math_129.use_clamp = False
    #Value_002
    math_129.inputs[2].default_value = 0.5
    
    #node Math.177
    math_177 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_177.name = "Math.177"
    math_177.operation = 'MULTIPLY'
    math_177.use_clamp = False
    #Value_002
    math_177.inputs[2].default_value = 0.5
    
    #node Mix.006
    mix_006 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_006.name = "Mix.006"
    mix_006.blend_type = 'MIX'
    mix_006.clamp_factor = True
    mix_006.clamp_result = False
    mix_006.data_type = 'FLOAT'
    mix_006.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_006.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_006.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_006.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_006.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_006.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_006.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_006.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Separate XYZ.060
    separate_xyz_060 = HD2_Shader.nodes.new("ShaderNodeSeparateXYZ")
    separate_xyz_060.name = "Separate XYZ.060"
    
    #node Mix.078
    mix_078 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_078.name = "Mix.078"
    mix_078.blend_type = 'MIX'
    mix_078.clamp_factor = True
    mix_078.clamp_result = False
    mix_078.data_type = 'FLOAT'
    mix_078.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_078.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Float
    mix_078.inputs[2].default_value = 0.0
    #A_Vector
    mix_078.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_078.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_078.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_078.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_078.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_078.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Math.145
    math_145 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_145.label = "r7.y (#543)"
    math_145.name = "Math.145"
    math_145.operation = 'MULTIPLY_ADD'
    math_145.use_clamp = False
    #Value_002
    math_145.inputs[2].default_value = 1.0
    
    #node Math.146
    math_146 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_146.label = "r7.y (#544)"
    math_146.name = "Math.146"
    math_146.operation = 'MULTIPLY'
    math_146.use_clamp = False
    #Value_002
    math_146.inputs[2].default_value = 1.0
    
    #node Mix.011
    mix_011 = HD2_Shader.nodes.new("ShaderNodeMix")
    mix_011.label = "r16.z (#545)"
    mix_011.name = "Mix.011"
    mix_011.blend_type = 'MIX'
    mix_011.clamp_factor = True
    mix_011.clamp_result = False
    mix_011.data_type = 'FLOAT'
    mix_011.factor_mode = 'UNIFORM'
    #Factor_Vector
    mix_011.inputs[1].default_value = (0.5, 0.5, 0.5)
    #A_Vector
    mix_011.inputs[4].default_value = (0.0, 0.0, 0.0)
    #B_Vector
    mix_011.inputs[5].default_value = (0.0, 0.0, 0.0)
    #A_Color
    mix_011.inputs[6].default_value = (0.5, 0.5, 0.5, 1.0)
    #B_Color
    mix_011.inputs[7].default_value = (0.5, 0.5, 0.5, 1.0)
    #A_Rotation
    mix_011.inputs[8].default_value = (0.0, 0.0, 0.0)
    #B_Rotation
    mix_011.inputs[9].default_value = (0.0, 0.0, 0.0)
    
    #node Group Output
    group_output_1 = HD2_Shader.nodes.new("NodeGroupOutput")
    group_output_1.name = "Group Output"
    group_output_1.is_active_output = True
    
    #node Math.194
    math_194 = HD2_Shader.nodes.new("ShaderNodeMath")
    math_194.name = "Math.194"
    math_194.operation = 'POWER'
    math_194.use_clamp = False
    #Value_001
    math_194.inputs[1].default_value = 2.200000047683716
    #Value_002
    math_194.inputs[2].default_value = 0.5

    update_images(HD2_Shader, material)

    #Set parents
    frame_005.parent = frame_004
    frame_014.parent = frame_013
    frame_015.parent = frame_013
    frame_020.parent = frame_001
    frame_006.parent = frame
    frame_018.parent = frame_017
    frame_008.parent = frame_007
    frame_009.parent = frame_008
    frame_010.parent = frame_007
    frame_003.parent = frame_002
    frame_011.parent = frame_003
    frame_012.parent = frame_011
    frame_023.parent = frame_022
    combine_xyz_076.parent = frame_037
    combine_xyz_077.parent = frame_037
    mix_055.parent = frame_037
    mix_056.parent = frame_037
    primary_material_lut_12.parent = frame_037
    primary_material_lut_13.parent = frame_038
    combine_xyz_078.parent = frame_038
    combine_xyz_084.parent = frame_041
    combine_xyz_085.parent = frame_041
    primary_material_lut_16.parent = frame_041
    mix_071.parent = frame_045
    mix_057.parent = frame_038
    mix_058.parent = frame_038
    combine_xyz_080.parent = frame_039
    combine_xyz_081.parent = frame_039
    primary_material_lut_14.parent = frame_039
    mix_059.parent = frame_039
    mix_060.parent = frame_039
    mix_061.parent = frame_040
    mix_062.parent = frame_040
    combine_xyz_083.parent = frame_040
    primary_material_lut_15.parent = frame_040
    mix_063.parent = frame_041
    mix_064.parent = frame_041
    mix_070.parent = frame_044
    mix_073.parent = frame_046
    mix_076.parent = frame_047
    mix_032.parent = frame_025
    combine_xyz_056.parent = frame_027
    mix_035.parent = frame_027
    combine_xyz_057.parent = frame_027
    mix_036.parent = frame_027
    primary_material_lut_02.parent = frame_027
    combine_xyz_058.parent = frame_028
    primary_material_lut_03.parent = frame_028
    combine_xyz_059.parent = frame_028
    combine_xyz_062.parent = frame_030
    mix_041.parent = frame_030
    mix_042.parent = frame_030
    mix_043.parent = frame_031
    combine_xyz_068.parent = frame_033
    mix_047.parent = frame_033
    mix_048.parent = frame_033
    primary_material_lut_08.parent = frame_033
    combine_xyz_070.parent = frame_034
    combine_xyz_071.parent = frame_034
    mix_050.parent = frame_034
    primary_material_lut_09.parent = frame_034
    combine_xyz_072.parent = frame_035
    combine_xyz_073.parent = frame_035
    mix_051.parent = frame_035
    primary_material_lut_10.parent = frame_035
    mix_053.parent = frame_036
    mix_054.parent = frame_036
    primary_material_lut_11.parent = frame_036
    combine_xyz_063.parent = frame_030
    combine_xyz_069.parent = frame_033
    combine_xyz_075.parent = frame_036
    combine_xyz_074.parent = frame_036
    mix_049.parent = frame_034
    mix_052.parent = frame_035
    mix_065.parent = frame_042
    mix_067.parent = frame_043
    vector_math_032.parent = frame_016
    combine_xyz_043.parent = frame_004
    vector_math_035.parent = frame_004
    combine_xyz_044.parent = frame_005
    math_230.parent = frame_004
    math_150.parent = frame_013
    math_152.parent = frame_013
    math_154.parent = frame_013
    math_155.parent = frame_013
    vector_math_086.parent = frame_013
    vector_math_087.parent = frame_013
    vector_math_088.parent = frame_013
    math_235.parent = frame_013
    vector_math_085.parent = frame_013
    math_231.parent = frame_013
    math_232.parent = frame_013
    combine_xyz_040.parent = frame_013
    vector_math_034.parent = frame_013
    separate_xyz_061.parent = frame_013
    separate_xyz_047.parent = frame_013
    math_087.parent = frame_013
    math_156.parent = frame_013
    math_159.parent = frame_013
    math_157.parent = frame_013
    math_160.parent = frame_013
    math_161.parent = frame_013
    math_162.parent = frame_013
    math_163.parent = frame_013
    vector_math_055.parent = frame_013
    vector_math_056.parent = frame_013
    separate_xyz_062.parent = frame_013
    vector_math_057.parent = frame_014
    math_158.parent = frame_013
    math_077.parent = frame_001
    math_078.parent = frame_001
    math_046.parent = frame_001
    math_047.parent = frame_001
    math_048.parent = frame_001
    value_002.parent = frame_001
    math_246.parent = frame_001
    math_247.parent = frame_001
    combine_xyz_034.parent = frame_001
    separate_xyz_027.parent = frame_001
    math_049.parent = frame_001
    combine_xyz_027.parent = frame_001
    math_044.parent = frame_001
    math_045.parent = frame_001
    vector_math_090.parent = frame_001
    vector_math_095.parent = frame_001
    vector_math_100.parent = frame_001
    vector_math_101.parent = frame_001
    mix_037.parent = frame_028
    mix_038.parent = frame_028
    math_008.parent = frame
    vector_math_004.parent = frame
    separate_xyz_004.parent = frame
    separate_xyz_001.parent = frame
    math_011.parent = frame
    math_010.parent = frame
    math_013.parent = frame
    math_012.parent = frame
    math_016.parent = frame
    math_017.parent = frame
    math_018.parent = frame
    math_015.parent = frame
    math_014.parent = frame
    vector_math_006.parent = frame
    separate_xyz_006.parent = frame
    vector_math_005.parent = frame
    math_009.parent = frame
    vector_math_008.parent = frame_006
    math_164.parent = frame_017
    math_165.parent = frame_017
    math_169.parent = frame_017
    math_170.parent = frame_017
    math_171.parent = frame_017
    math_172.parent = frame_017
    separate_xyz_065.parent = frame_017
    math_168.parent = frame_017
    vector_math_059.parent = frame_017
    vector_math_062.parent = frame_018
    separate_xyz_012.parent = frame_048
    math_264.parent = frame_048
    math_263.parent = frame_048
    math_262.parent = frame_048
    math_265.parent = frame_048
    separate_xyz_050.parent = frame_048
    vector_math_106.parent = frame_007
    vector_math_107.parent = frame_007
    combine_xyz_010.parent = frame_007
    clamp_020.parent = frame_007
    separate_xyz_002.parent = frame_007
    vector_math_108.parent = frame_007
    vector_math_109.parent = frame_007
    math_019.parent = frame_007
    math_266.parent = frame_007
    math_267.parent = frame_007
    vector_math_009.parent = frame_007
    vector_math_010.parent = frame_007
    separate_xyz_007.parent = frame_007
    clamp_001.parent = frame_007
    clamp_002.parent = frame_007
    clamp.parent = frame_007
    combine_xyz_001.parent = frame_007
    separate_xyz_008.parent = frame_007
    vector_math_012.parent = frame_007
    vector_math_011.parent = frame_007
    vector_math_013.parent = frame_007
    vector_math_014.parent = frame_007
    vector_math_037.parent = frame_007
    vector_math_038.parent = frame_007
    vector_math_039.parent = frame_007
    gamma_006.parent = frame_007
    vector_math_015.parent = frame_007
    vector_math_036.parent = frame_007
    combine_xyz_051.parent = frame_007
    value_007.parent = frame_009
    mix_003.parent = frame_009
    math_107.parent = frame_008
    math_092.parent = frame_008
    math_093.parent = frame_008
    vector_math_043.parent = frame_008
    vector_math_044.parent = frame_008
    vector_math_042.parent = frame_008
    math_108.parent = frame_008
    mix_007.parent = frame_009
    math_109.parent = frame_008
    mix_002.parent = frame_010
    value_011.parent = frame_010
    math_114.parent = frame_003
    math_111.parent = frame_003
    math_112.parent = frame_003
    math_113.parent = frame_003
    vector_math_053.parent = frame_011
    vector_math_052.parent = frame_011
    vector_math_054.parent = frame_011
    combine_xyz_016.parent = frame_012
    math_130.parent = frame_012
    math_131.parent = frame_012
    math_132.parent = frame_012
    math_134.parent = frame_012
    math_136.parent = frame_012
    math_138.parent = frame_012
    math_133.parent = frame_012
    math_135.parent = frame_012
    math_137.parent = frame_012
    math_140.parent = frame_012
    math_139.parent = frame_012
    math_141.parent = frame_012
    math_142.parent = frame_012
    math_144.parent = frame_012
    math_143.parent = frame_012
    math_196.parent = frame_021
    clamp_015.parent = frame_021
    math_197.parent = frame_021
    math_198.parent = frame_021
    math_195.parent = frame_021
    math_193.parent = frame_021
    vector_math_074.parent = frame_021
    vector_math_075.parent = frame_021
    vector_math_076.parent = frame_021
    math_200.parent = frame_021
    math_199.parent = frame_021
    math_202.parent = frame_021
    math_206.parent = frame_021
    math_207.parent = frame_021
    math_201.parent = frame_021
    math_203.parent = frame_021
    math_208.parent = frame_021
    math_192.parent = frame_021
    math_022.parent = frame_021
    math_211.parent = frame_022
    separate_xyz_019.parent = frame_022
    clamp_016.parent = frame_022
    clamp_017.parent = frame_022
    math_212.parent = frame_022
    combine_xyz_008.parent = frame_022
    math_213.parent = frame_022
    separate_xyz_021.parent = frame_022
    math_214.parent = frame_022
    math_218.parent = frame_022
    math_219.parent = frame_022
    math_221.parent = frame_022
    math_220.parent = frame_022
    separate_xyz_026.parent = frame_022
    math_215.parent = frame_022
    math_216.parent = frame_022
    math_217.parent = frame_022
    combine_xyz_021.parent = frame_022
    vector_math_080.parent = frame_022
    vector_math_079.parent = frame_022
    vector_math_082.parent = frame_022
    vector_math_081.parent = frame_022
    math_222.parent = frame_022
    combine_xyz_022.parent = frame_023
    value.parent = frame_023
    combine_xyz_025.parent = frame_023
    math_210.parent = frame_022
    math_223.parent = frame_049
    math_224.parent = frame_049
    customization_camo_tiler_array.parent = frame_007
    vector_math_033.parent = frame_013
    pattern_lut_03.parent = frame_012
    gamma_pattern.parent = frame_012
    math_225.parent = frame_049
    primary_material_lut_05.parent = frame_030
    combine_xyz_079.parent = frame_038
    combine_xyz_082.parent = frame_040
    vector_math_007.parent = frame
    separate_xyz_064.parent = frame_017
    math_166.parent = frame_017
    math_167.parent = frame_017
    math_173.parent = frame_017
    math_174.parent = frame_017
    vector_math_060.parent = frame_017
    vector_math_061.parent = frame_017
    vector_math_058.parent = frame_017
    mix_015.parent = frame_020
    vector_math_003.parent = frame
    vector_math.parent = frame_013
    customization_material_detail_tiler_array_001.parent = frame_013
    math_151.parent = frame_013
    math_153.parent = frame_013
    separate_xyz_048.parent = frame_004
    composite_array.parent = frame_001
    vector_math_022.parent = frame_001
    clamp_019.parent = frame
    vector_math_097.parent = frame_051
    vector_math_099.parent = frame_051
    vector_math_096.parent = frame_051
    vector_math_098.parent = frame_051
    vector_math_091.parent = frame_050
    combine_xyz_098.parent = frame_050
    vector_math_113.parent = frame_050
    vector_math_112.parent = frame_050
    vector_math_021.parent = frame_050
    mix_021.parent = frame_050
    vector_math_114.parent = frame_051
    vector_math_111.parent = frame_051
    vector_math_115.parent = frame_051
    vector_math_089.parent = frame_050
    vector_math_116.parent = frame_050
    normal_map_001.parent = frame_051
    math_270.parent = frame_051
    math_204.parent = frame_051
    math_205.parent = frame_051
    math_268.parent = frame_051
    math_269.parent = frame_051
    combine_xyz_099.parent = frame_051
    vector_math_093.parent = frame_051
    vector_math_092.parent = frame_051
    vector_math_094.parent = frame_051
    math_271.parent = frame_051
    separate_xyz_056.parent = frame_051
    math_272.parent = frame_050
    math_273.parent = frame_050
    math_274.parent = frame_050
    math_275.parent = frame_050
    math_276.parent = frame_050
    combine_xyz_100.parent = frame_050
    vector_math_110.parent = frame_050
    vector_math_117.parent = frame_050
    vector_math_118.parent = frame_050
    math_277.parent = frame_050
    separate_xyz_063.parent = frame_050
    mix_shader.parent = frame_053
    separate_xyz_051.parent = frame_053
    principled_bsdf_001.parent = frame_052
    normal_map.parent = frame_050
    clamp_018.parent = frame_049
    mix_033.parent = frame_026
    combine_xyz_096.parent = frame_047
    combine_xyz_097.parent = frame_047
    primary_material_lut_22.parent = frame_047
    math_084.parent = frame_004
    value_006.parent = frame_004
    math_085.parent = frame_004
    object_info.parent = frame_004
    mix_075.parent = frame_047
    math_086.parent = frame_004
    math_229.parent = frame_004
    math_228.parent = frame_004
    math_226.parent = frame_004
    detail_uvs.parent = frame_004
    math_079.parent = frame_004
    combine_xyz_029.parent = frame_004
    math_038.parent = frame_003
    combine_xyz_086.parent = frame_042
    combine_xyz_089.parent = frame_043
    combine_xyz_088.parent = frame_043
    combine_xyz_090.parent = frame_044
    combine_xyz_091.parent = frame_044
    math_227.parent = frame_004
    vector_math_001.parent = frame_004
    math_025.parent = frame_008
    math_039.parent = frame_003
    combine_xyz_087.parent = frame_042
    primary_material_lut_17.parent = frame_042
    primary_material_lut_18.parent = frame_043
    primary_material_lut_19.parent = frame_044
    combine_xyz_026.parent = frame_004
    vector_math_002.parent = frame_004
    math_026.parent = frame_008
    math_040.parent = frame_003
    separate_xyz_016.parent = frame_003
    math_236.parent = frame_024
    mix_066.parent = frame_042
    mix_068.parent = frame_043
    mix_069.parent = frame_044
    vector_math_078.parent = frame_004
    math_027.parent = frame_008
    combine_xyz_018.parent = frame_003
    combine_xyz_017.parent = frame_003
    math_237.parent = frame_024
    combine_xyz_093.parent = frame_045
    combine_xyz_092.parent = frame_045
    vector_math_083.parent = frame_004
    value_008.parent = frame_008
    math_028.parent = frame_008
    math_029.parent = frame_008
    reroute.parent = frame_008
    reroute_001.parent = frame_008
    vector_math_016.parent = frame_003
    Slot1and2.parent = frame_024
    math_238.parent = frame_024
    primary_material_lut_20.parent = frame_045
    separate_xyz_035.parent = frame_004
    combine_xyz_005.parent = frame_008
    combine_xyz_006.parent = frame_008
    separate_xyz_017.parent = frame_003
    value_001.parent = frame_003
    Slot3.parent = frame_024
    customization_material_detail_tiler_array.parent = frame_004
    mix_039.parent = frame_029
    combine_xyz_060.parent = frame_029
    combine_xyz_061.parent = frame_029
    math_239.parent = frame_024
    mix_072.parent = frame_045
    math_082.parent = frame_004
    combine_xyz_039.parent = frame_004
    math_249.parent = frame_019
    vector_math_040.parent = frame_008
    combine_xyz_019.parent = frame_003
    Slot4.parent = frame_024
    primary_material_lut_04.parent = frame_029
    math_240.parent = frame_024
    combine_xyz_054.parent = frame_026
    combine_xyz_055.parent = frame_026
    math_083.parent = frame_004
    combine_xyz_045.parent = frame_005
    separate_xyz_037.parent = frame_019
    math_248.parent = frame_019
    math_031.parent = frame_008
    vector_math_041.parent = frame_008
    combine_xyz_003.parent = frame_003
    vector_math_017.parent = frame_003
    Slot5.parent = frame_024
    vector_math_084.parent = frame_004
    mix_040.parent = frame_029
    separate_xyz_005.parent = frame_019
    math_241.parent = frame_024
    primary_material_lut_01.parent = frame_026
    math_234.parent = frame_019
    math_030.parent = frame_008
    vector_math_018.parent = frame_003
    separate_xyz_059.parent = frame_003
    Slot6.parent = frame_024
    pattern_lut_02.parent = frame_003
    gamma_pattern.parent = frame_003
    math_242.parent = frame_024
    combine_xyz_094.parent = frame_046
    combine_xyz_095.parent = frame_046
    combine_xyz_052.parent = frame_025
    combine_xyz_053.parent = frame_025
    combine_xyz_067.parent = frame_032
    mix_045.parent = frame_032
    combine_xyz_066.parent = frame_032
    mix_034.parent = frame_026
    clamp_008.parent = frame_008
    math_119.parent = frame_003
    math_110.parent = frame_003
    combine_xyz_002.parent = frame_003
    Slot7.parent = frame_024
    math_250.parent = frame_019
    primary_material_lut_21.parent = frame_046
    primary_material_lut_00.parent = frame_025
    primary_material_lut_07.parent = frame_032
    math_033.parent = frame_008
    math_120.parent = frame_003
    clamp_014.parent = frame_003
    Slot8.parent = frame_024
    pattern_lut.parent = frame_003
    gamma_pattern.parent = frame_003
    mix_074.parent = frame_046
    mix_031.parent = frame_025
    combine_xyz_064.parent = frame_031
    combine_xyz_065.parent = frame_031
    mix_046.parent = frame_032
    math_032.parent = frame_008
    math_115.parent = frame_003
    math_036.parent = frame_003
    math_121.parent = frame_011
    math_126.parent = frame_012
    math_116.parent = frame_003
    math_034.parent = frame_002
    math_243.parent = frame_024
    math_020.parent = frame_007
    math_089.parent = frame_008
    math_023.parent = frame_008
    math_035.parent = frame_002
    math_037.parent = frame_003
    math_117.parent = frame_011
    math_122.parent = frame_011
    math_125.parent = frame_012
    primary_material_lut_06.parent = frame_031
    mix_044.parent = frame_031
    math_021.parent = frame_007
    math_090.parent = frame_008
    math_024.parent = frame_008
    math_118.parent = frame_011
    math_124.parent = frame_012
    math_127.parent = frame_012
    math_091.parent = frame_008
    math_123.parent = frame_012
    math_128.parent = frame_012
    math_129.parent = frame_012
    mix_006.parent = frame_049
    mix_078.parent = frame_015
    math_145.parent = frame_011
    math_146.parent = frame_011
    mix_011.parent = frame_011
    math_194.parent = frame_053
    
    #Set locations
    frame_024.location = (-4780.0, 5400.0)
    frame_047.location = (-3152.0, -8487.0)
    frame_046.location = (-3152.0, -7847.0)
    frame_045.location = (-3152.0, -7207.0)
    frame_044.location = (-3152.0, -6567.0)
    frame_043.location = (-3152.0, -5927.0)
    frame_042.location = (-3152.0, -5287.0)
    frame_041.location = (-3152.0, -4647.0)
    frame_040.location = (-3152.0, -4007.0)
    frame_039.location = (-3152.0, -3367.0)
    frame_038.location = (-3152.0, -2727.0)
    frame_037.location = (-3152.0, -2087.0)
    frame_036.location = (-3152.0, -1447.0)
    frame_035.location = (-3152.0, -807.0)
    frame_034.location = (-3152.0, -167.0)
    frame_033.location = (-3152.0, 473.0)
    frame_032.location = (-3152.0, 1113.0)
    frame_031.location = (-3152.0, 1753.0)
    frame_030.location = (-3152.0, 2393.0)
    frame_028.location = (-3152.0, 3673.0)
    frame_027.location = (-3152.0, 4313.0)
    frame_026.location = (-3152.0, 4953.0)
    frame_025.location = (-3152.0, 5593.0)
    frame_016.location = (420.0, 4280.0)
    frame_004.location = (-563.0, 5480.0)
    frame_005.location = (3670.0, -340.0)
    frame_013.location = (4180.0, 5480.0)
    frame_014.location = (3170.0, -480.0)
    frame_015.location = (3180.0, -860.0)
    frame_001.location = (8740.0, 4747.0)
    frame_020.location = (920.0, 180.0)
    frame.location = (15830.0, 4548.0)
    frame_006.location = (2170.0, -320.0)
    frame_017.location = (18370.0, 4527.0)
    frame_018.location = (2060.0, -467.0)
    frame_019.location = (11970.0, 4780.0)
    frame_048.location = (25333.0, 3590.0)
    frame_007.location = (28930.0, 3655.0)
    frame_008.location = (4230.0, -688.0)
    frame_009.location = (1760.0, -427.0)
    frame_010.location = (6848.0, -2203.0)
    frame_002.location = (36490.0, 3848.0)
    frame_003.location = (600.0, -313.0)
    frame_011.location = (3070.0, -48.0)
    frame_012.location = (630.0, 20.0)
    frame_021.location = (45380.0, 3574.0)
    frame_022.location = (46700.0, 3540.0)
    frame_023.location = (1120.0, -800.0)
    frame_049.location = (50400.0, 4660.0)
    frame_050.location = (49820.0, 4060.0)
    frame_051.location = (49770.0, 3770.0)
    frame_053.location = (52480.0, 5180.0)
    cape_decal_group_node.location = (51450.0, 4950.0)
    decal_mix_node.location = (51800.0, 4900.0)
    alpha_cutoff_node.location = (51800.0, 4600.0)
    transparency_shader.location = (52880.0, 5040.0)
    mix_shader_transparency.location = (53080.0, 5100.0)
    frame_052.location = (52390.0, 4900.0)
    frame_029.location = (-3152.0, 3033.0)
    combine_xyz_076.location = (-288.0, 187.0)
    combine_xyz_077.location = (-288.0, -93.0)
    mix_055.location = (152.0, 187.0)
    mix_056.location = (152.0, -33.0)
    primary_material_lut_12.location = (-128.0, 187.0)
    primary_material_lut_13.location = (-128.0, 187.0)
    combine_xyz_078.location = (-288.0, 187.0)
    combine_xyz_084.location = (-288.0, -93.0)
    combine_xyz_085.location = (-288.0, 187.0)
    primary_material_lut_16.location = (-128.0, 187.0)
    mix_071.location = (152.0, 187.0)
    mix_057.location = (152.0, 187.0)
    mix_058.location = (152.0, -33.0)
    combine_xyz_080.location = (-288.0, -93.0)
    combine_xyz_081.location = (-288.0, 187.0)
    primary_material_lut_14.location = (-128.0, 187.0)
    mix_059.location = (152.0, 187.0)
    mix_060.location = (152.0, -33.0)
    mix_061.location = (152.0, 187.0)
    mix_062.location = (152.0, -33.0)
    combine_xyz_083.location = (-288.0, 187.0)
    primary_material_lut_15.location = (-128.0, 187.0)
    mix_063.location = (152.0, 187.0)
    mix_064.location = (152.0, -33.0)
    mix_070.location = (152.0, -33.0)
    mix_073.location = (152.0, 187.0)
    mix_076.location = (152.0, -33.0)
    separate_xyz_052.location = (-2800.0, -7660.0)
    mix_032.location = (152.0, 187.0)
    combine_xyz_056.location = (-288.0, 187.0)
    mix_035.location = (152.0, 187.0)
    combine_xyz_057.location = (-288.0, -93.0)
    mix_036.location = (152.0, -33.0)
    primary_material_lut_02.location = (-128.0, 187.0)
    combine_xyz_058.location = (-288.0, 187.0)
    primary_material_lut_03.location = (-128.0, 187.0)
    combine_xyz_059.location = (-288.0, -93.0)
    combine_xyz_062.location = (-288.0, 187.0)
    mix_041.location = (152.0, 187.0)
    mix_042.location = (152.0, -33.0)
    mix_043.location = (152.0, 187.0)
    combine_xyz_068.location = (-288.0, 187.0)
    mix_047.location = (152.0, 187.0)
    mix_048.location = (152.0, -33.0)
    primary_material_lut_08.location = (-128.0, 187.0)
    combine_xyz_070.location = (-288.0, 187.0)
    combine_xyz_071.location = (-288.0, -93.0)
    mix_050.location = (152.0, -33.0)
    primary_material_lut_09.location = (-128.0, 187.0)
    combine_xyz_072.location = (-288.0, 187.0)
    combine_xyz_073.location = (-288.0, -93.0)
    mix_051.location = (152.0, 187.0)
    primary_material_lut_10.location = (-128.0, 187.0)
    mix_053.location = (152.0, 187.0)
    mix_054.location = (152.0, -33.0)
    primary_material_lut_11.location = (-128.0, 187.0)
    combine_xyz_063.location = (-288.0, -93.0)
    combine_xyz_069.location = (-288.0, -93.0)
    combine_xyz_075.location = (-288.0, -93.0)
    combine_xyz_074.location = (-288.0, 187.0)
    mix_049.location = (152.0, 187.0)
    mix_052.location = (152.0, -33.0)
    gamma_005.location = (-2800.0, 1940.0)
    gamma_004.location = (-2800.0, 4500.0)
    separate_xyz_038.location = (-2800.0, 660.0)
    separate_xyz_057.location = (-2800.0, -3180.0)
    separate_xyz_058.location = (-2800.0, -3820.0)
    mix_065.location = (152.0, 187.0)
    mix_067.location = (152.0, 187.0)
    math_191.location = (-2280.0, 5620.0)
    clamp_004.location = (-2120.0, 5620.0)
    separate_xyz_028.location = (-2280.0, 5460.0)
    combine_xyz_028.location = (-2120.0, 5460.0)
    vector_math_023.location = (-1960.0, 5460.0)
    math_075.location = (-1480.0, 5460.0)
    separate_xyz_045.location = (-1640.0, 5460.0)
    math_074.location = (-1320.0, 5460.0)
    vector_math_024.location = (-1800.0, 5460.0)
    clamp_003.location = (-2280.0, 5780.0)
    separate_xyz_029.location = (-1320.0, 5160.0)
    vector_math_025.location = (-1320.0, 5300.0)
    separate_xyz_030.location = (-1160.0, 5300.0)
    math_051.location = (-1160.0, 5160.0)
    math_054.location = (-1160.0, 4980.0)
    math_052.location = (-1000.0, 5160.0)
    math_055.location = (-840.0, 5160.0)
    math_056.location = (-840.0, 5320.0)
    math_053.location = (-680.0, 5320.0)
    math_057.location = (-520.0, 5320.0)
    math_058.location = (-360.0, 5320.0)
    vector_math_026.location = (-40.0, 5320.0)
    separate_xyz_031.location = (120.0, 5320.0)
    vector_math_032.location = (-110.0, 1040.0)
    mix_013.location = (3040.0, 4860.0)
    math_059.location = (-200.0, 5320.0)
    math_147.location = (3040.0, 5500.0)
    math_148.location = (3200.0, 5500.0)
    mix_004.location = (3360.0, 5500.0)
    math_149.location = (3200.0, 5320.0)
    combine_xyz_043.location = (3263.0, -320.0)
    vector_math_035.location = (3423.0, -320.0)
    combine_xyz_044.location = (-287.0, -320.0)
    math_230.location = (2943.0, -520.0)
    math_150.location = (-640.0, 20.0)
    math_152.location = (-320.0, -160.0)
    math_154.location = (-160.0, -320.0)
    math_155.location = (0.0, -320.0)
    vector_math_086.location = (480.0, -480.0)
    vector_math_087.location = (640.0, -480.0)
    vector_math_088.location = (800.0, -480.0)
    math_235.location = (1220.0, -680.0)
    vector_math_085.location = (320.0, -480.0)
    math_231.location = (160.0, -320.0)
    math_232.location = (320.0, -320.0)
    combine_xyz_040.location = (480.0, -320.0)
    vector_math_034.location = (1540.0, -480.0)
    separate_xyz_061.location = (1700.0, -480.0)
    separate_xyz_047.location = (1540.0, -640.0)
    math_087.location = (1700.0, -640.0)
    math_156.location = (1860.0, -640.0)
    math_159.location = (1860.0, -480.0)
    math_157.location = (2020.0, -640.0)
    math_160.location = (2020.0, -480.0)
    math_161.location = (2180.0, -480.0)
    math_162.location = (2340.0, -480.0)
    math_163.location = (2500.0, -480.0)
    vector_math_055.location = (2660.0, -480.0)
    vector_math_056.location = (2820.0, -480.0)
    separate_xyz_062.location = (2980.0, -480.0)
    vector_math_057.location = (-10.0, -20.0)
    math_158.location = (1700.0, -820.0)
    math_077.location = (-980.0, 753.0)
    math_078.location = (-820.0, 753.0)
    math_244.location = (7560.0, 5500.0)
    math_046.location = (-680.0, 573.0)
    math_047.location = (-520.0, 573.0)
    math_048.location = (-360.0, 573.0)
    math_245.location = (7560.0, 5320.0)
    value_002.location = (-680.0, 413.0)
    math_246.location = (-200.0, 573.0)
    math_247.location = (-40.0, 573.0)
    combine_xyz_034.location = (120.0, 573.0)
    separate_xyz_027.location = (1020.0, 413.0)
    math_049.location = (1180.0, 413.0)
    combine_xyz_027.location = (1340.0, 413.0)
    math_044.location = (-360.0, 413.0)
    math_045.location = (-200.0, 413.0)
    vector_math_090.location = (-40.0, 413.0)
    vector_math_095.location = (120.0, 413.0)
    vector_math_100.location = (280.0, 413.0)
    vector_math_101.location = (440.0, 413.0)
    mix_012.location = (10640.0, 5500.0)
    value_003.location = (10480.0, 5500.0)
    combine_xyz_038.location = (13420.0, 4400.0)
    math_068.location = (13100.0, 4560.0)
    math_069.location = (13260.0, 4560.0)
    math_070.location = (13420.0, 4560.0)
    vector_math_031.location = (13580.0, 4560.0)
    mix_037.location = (152.0, 187.0)
    mix_038.location = (152.0, -33.0)
    separate_xyz_044.location = (13740.0, 4560.0)
    combine_xyz_049.location = (13900.0, 4560.0)
    vector_math_103.location = (14060.0, 4560.0)
    math_251.location = (14060.0, 4400.0)
    math_252.location = (14700.0, 4400.0)
    math_253.location = (14540.0, 4400.0)
    math_254.location = (14380.0, 4400.0)
    math_255.location = (14860.0, 4400.0)
    separate_xyz_042.location = (14220.0, 4400.0)
    math_076.location = (14060.0, 4220.0)
    math_072.location = (13740.0, 4220.0)
    math_073.location = (13900.0, 4220.0)
    math_071.location = (13420.0, 4220.0)
    clamp_006.location = (13580.0, 4220.0)
    math_002.location = (15180.0, 4240.0)
    math_006.location = (15500.0, 4400.0)
    math_005.location = (15660.0, 4400.0)
    math_004.location = (15340.0, 4400.0)
    math_007.location = (15820.0, 4400.0)
    clamp_007.location = (15180.0, 4400.0)
    math_088.location = (15020.0, 4400.0)
    separate_xyz_046.location = (14220.0, 4220.0)
    math_008.location = (190.0, -148.0)
    vector_math_004.location = (350.0, -328.0)
    separate_xyz_004.location = (510.0, -328.0)
    separate_xyz_001.location = (350.0, -488.0)
    math_011.location = (510.0, -488.0)
    math_010.location = (670.0, -488.0)
    math_013.location = (510.0, -668.0)
    math_012.location = (830.0, -488.0)
    math_016.location = (990.0, -328.0)
    math_017.location = (1150.0, -328.0)
    math_018.location = (1310.0, -328.0)
    math_015.location = (830.0, -328.0)
    math_014.location = (670.0, -328.0)
    vector_math_006.location = (1630.0, -328.0)
    separate_xyz_006.location = (1950.0, -328.0)
    vector_math_005.location = (1470.0, -328.0)
    math_009.location = (350.0, -148.0)
    vector_math_008.location = (-30.0, -8.0)
    math_164.location = (-150.0, -127.0)
    math_165.location = (10.0, -127.0)
    math_169.location = (790.0, -307.0)
    math_170.location = (950.0, -487.0)
    math_171.location = (1110.0, -487.0)
    math_172.location = (1270.0, -487.0)
    separate_xyz_065.location = (790.0, -487.0)
    math_168.location = (470.0, -487.0)
    vector_math_059.location = (150.0, -487.0)
    vector_math_062.location = (30.0, -20.0)
    vector_math_029.location = (24000.0, 3720.0)
    vector_math_030.location = (24160.0, 3720.0)
    math_094.location = (24320.0, 3720.0)
    separate_xyz_039.location = (23680.0, 3720.0)
    math_064.location = (22880.0, 3720.0)
    math_065.location = (23040.0, 3720.0)
    math_066.location = (23200.0, 3720.0)
    combine_xyz_035.location = (22880.0, 3560.0)
    combine_xyz_036.location = (23360.0, 3720.0)
    vector_math_028.location = (23520.0, 3720.0)
    value_005.location = (23340.0, 3580.0)
    combine_xyz_037.location = (23840.0, 3720.0)
    math_095.location = (24320.0, 3540.0)
    math_096.location = (24480.0, 3540.0)
    math_097.location = (24640.0, 3540.0)
    combine_xyz_007.location = (24640.0, 3380.0)
    vector_math_045.location = (24800.0, 3380.0)
    separate_xyz_010.location = (24960.0, 3380.0)
    combine_xyz_050.location = (25120.0, 3500.0)
    math_261.location = (25280.0, 3360.0)
    vector_math_105.location = (25280.0, 3500.0)
    separate_xyz_011.location = (25120.0, 3720.0)
    separate_xyz_055.location = (24160.0, 3540.0)
    combine_xyz_009.location = (25280.0, 3720.0)
    separate_xyz_012.location = (147.0, 130.0)
    math_264.location = (307.0, 130.0)
    math_263.location = (467.0, 130.0)
    math_262.location = (627.0, 130.0)
    math_265.location = (787.0, 130.0)
    math_098.location = (26320.0, 3720.0)
    clamp_009.location = (26480.0, 3720.0)
    clamp_010.location = (28080.0, 3720.0)
    vector_math_048.location = (27760.0, 3720.0)
    math_102.location = (27920.0, 3720.0)
    combine_xyz_012.location = (27440.0, 3720.0)
    vector_math_047.location = (27600.0, 3720.0)
    math_100.location = (26800.0, 3720.0)
    math_099.location = (26640.0, 3720.0)
    math_101.location = (26960.0, 3720.0)
    combine_xyz_011.location = (26960.0, 3560.0)
    vector_math_046.location = (27120.0, 3720.0)
    separate_xyz_013.location = (27280.0, 3720.0)
    separate_xyz_050.location = (147.0, -90.0)
    vector_math_051.location = (27760.0, 3420.0)
    clamp_011.location = (28080.0, 3420.0)
    math_106.location = (27920.0, 3420.0)
    combine_xyz_014.location = (27440.0, 3420.0)
    vector_math_050.location = (27600.0, 3420.0)
    math_103.location = (26640.0, 3420.0)
    math_104.location = (26800.0, 3420.0)
    combine_xyz_013.location = (26960.0, 3260.0)
    math_105.location = (26960.0, 3420.0)
    vector_math_049.location = (27120.0, 3420.0)
    separate_xyz_014.location = (27280.0, 3420.0)
    value_009.location = (27280.0, 3560.0)
    combine_xyz_015.location = (28240.0, 3260.0)
    separate_xyz_053.location = (28400.0, 3260.0)
    separate_xyz_015.location = (27920.0, 3260.0)
    clamp_013.location = (28080.0, 3260.0)
    clamp_012.location = (28080.0, 3100.0)
    value_010.location = (27280.0, 3260.0)
    vector_math_106.location = (630.0, -555.0)
    vector_math_107.location = (790.0, -555.0)
    combine_xyz_010.location = (470.0, -555.0)
    clamp_020.location = (310.0, -555.0)
    separate_xyz_002.location = (150.0, -555.0)
    vector_math_108.location = (-10.0, -555.0)
    vector_math_109.location = (-170.0, -555.0)
    math_019.location = (-330.0, -555.0)
    math_266.location = (-170.0, -715.0)
    math_267.location = (-10.0, -715.0)
    vector_math_009.location = (1210.0, -555.0)
    vector_math_010.location = (1370.0, -555.0)
    separate_xyz_007.location = (1530.0, -555.0)
    clamp_001.location = (1690.0, -715.0)
    clamp_002.location = (1690.0, -875.0)
    clamp.location = (1690.0, -555.0)
    combine_xyz_001.location = (1850.0, -555.0)
    separate_xyz_008.location = (2010.0, -555.0)
    vector_math_012.location = (1690.0, -1035.0)
    vector_math_011.location = (1850.0, -1035.0)
    vector_math_013.location = (2170.0, -555.0)
    vector_math_014.location = (2330.0, -555.0)
    vector_math_037.location = (2970.0, -555.0)
    vector_math_038.location = (2810.0, -555.0)
    vector_math_039.location = (3130.0, -555.0)
    gamma_006.location = (3330.0, -555.0)
    vector_math_015.location = (2490.0, -555.0)
    vector_math_036.location = (2650.0, -555.0)
    combine_xyz_051.location = (150.0, -715.0)
    value_007.location = (-100.0, -260.0)
    mix_003.location = (60.0, -260.0)
    math_107.location = (1540.0, -107.0)
    math_092.location = (1700.0, 53.0)
    math_093.location = (1700.0, -107.0)
    vector_math_043.location = (900.0, 413.0)
    vector_math_044.location = (1220.0, 413.0)
    vector_math_042.location = (1060.0, 413.0)
    math_108.location = (1860.0, 53.0)
    mix_007.location = (60.0, -60.0)
    math_109.location = (1860.0, -267.0)
    mix_002.location = (22.0, 808.0)
    value_011.location = (-138.0, 808.0)
    mix_009.location = (36180.0, 2380.0)
    mix_010.location = (36180.0, 2200.0)
    mix_001.location = (35220.0, 2000.0)
    math_114.location = (2770.0, -255.0)
    math_111.location = (2450.0, -255.0)
    math_112.location = (2610.0, -255.0)
    math_113.location = (2450.0, -395.0)
    vector_math_053.location = (660.0, 53.0)
    vector_math_052.location = (820.0, 53.0)
    vector_math_054.location = (980.0, 53.0)
    combine_xyz_016.location = (1350.0, -147.0)
    math_130.location = (1510.0, -427.0)
    math_131.location = (1350.0, -427.0)
    math_132.location = (1670.0, -427.0)
    math_134.location = (1670.0, -607.0)
    math_136.location = (1990.0, -607.0)
    math_138.location = (2150.0, -607.0)
    math_133.location = (1510.0, -607.0)
    math_135.location = (1830.0, -607.0)
    math_137.location = (1990.0, -787.0)
    math_140.location = (1990.0, -947.0)
    math_139.location = (2150.0, -947.0)
    math_141.location = (2310.0, -947.0)
    math_142.location = (2150.0, -1127.0)
    math_144.location = (2470.0, -1127.0)
    math_143.location = (2310.0, -1127.0)
    separate_xyz.location = (44380.0, 3720.0)
    separate_xyz_003.location = (44860.0, 3720.0)
    combine_xyz.location = (44700.0, 3720.0)
    vector_math_073.location = (44220.0, 3720.0)
    uv_map_002.location = (44060.0, 3720.0)
    math_1.location = (44540.0, 3720.0)
    math_001.location = (44540.0, 3540.0)
    mix_014.location = (44020.0, 2260.0)
    mix_016.location = (44020.0, 2460.0)
    math_196.location = (160.0, -34.0)
    clamp_015.location = (0.0, -34.0)
    math_197.location = (320.0, -34.0)
    math_198.location = (480.0, -34.0)
    math_195.location = (-160.0, -34.0)
    math_193.location = (-320.0, -34.0)
    vector_math_074.location = (-160.0, -194.0)
    vector_math_075.location = (0.0, -194.0)
    vector_math_076.location = (160.0, -194.0)
    math_200.location = (0.0, -354.0)
    math_199.location = (160.0, -354.0)
    math_202.location = (160.0, -534.0)
    math_206.location = (160.0, -694.0)
    math_207.location = (320.0, -694.0)
    math_201.location = (320.0, -534.0)
    math_203.location = (480.0, -534.0)
    math_208.location = (640.0, -534.0)
    math_192.location = (-160.0, 146.0)
    math_022.location = (-320.0, 146.0)
    mix_1.location = (46220.0, 3080.0)
    mix_017.location = (46220.0, 2880.0)
    math_211.location = (-120.0, 180.0)
    separate_xyz_019.location = (-280.0, 0.0)
    clamp_016.location = (-120.0, 0.0)
    clamp_017.location = (-120.0, -160.0)
    math_212.location = (-120.0, -320.0)
    combine_xyz_008.location = (40.0, 0.0)
    math_213.location = (40.0, -320.0)
    separate_xyz_021.location = (200.0, 0.0)
    math_214.location = (360.0, 0.0)
    math_218.location = (520.0, 0.0)
    math_219.location = (680.0, 0.0)
    math_221.location = (840.0, 0.0)
    math_220.location = (1000.0, 0.0)
    separate_xyz_026.location = (40.0, -480.0)
    math_215.location = (200.0, -480.0)
    math_216.location = (200.0, -620.0)
    math_217.location = (200.0, -760.0)
    combine_xyz_021.location = (360.0, -480.0)
    vector_math_080.location = (520.0, -480.0)
    vector_math_079.location = (200.0, -900.0)
    vector_math_082.location = (680.0, -640.0)
    vector_math_081.location = (680.0, -480.0)
    math_222.location = (1160.0, 0.0)
    combine_xyz_022.location = (0.0, 80.0)
    value.location = (0.0, -40.0)
    combine_xyz_025.location = (0.0, -140.0)
    vector_math_077.location = (48060.0, 3340.0)
    vector_math_063.location = (48220.0, 3340.0)
    math_178.location = (48380.0, 3180.0)
    math_179.location = (48380.0, 3020.0)
    math_209.location = (48220.0, 3020.0)
    vector_math_066.location = (48380.0, 2700.0)
    math_182.location = (48380.0, 2860.0)
    vector_math_067.location = (48540.0, 2700.0)
    vector_math_064.location = (48220.0, 2700.0)
    math_210.location = (-280.0, 180.0)
    separate_xyz_067.location = (49500.0, 2500.0)
    combine_xyz_042.location = (49820.0, 2500.0)
    vector_math_071.location = (49340.0, 2500.0)
    vector_math_069.location = (49180.0, 2500.0)
    math_183.location = (48700.0, 2500.0)
    math_184.location = (48860.0, 2500.0)
    separate_xyz_068.location = (49820.0, 2020.0)
    math_188.location = (49980.0, 2020.0)
    math_189.location = (49980.0, 1860.0)
    math_190.location = (49980.0, 1700.0)
    combine_xyz_048.location = (50140.0, 2020.0)
    math_185.location = (49660.0, 2500.0)
    math_186.location = (49660.0, 2340.0)
    math_187.location = (49660.0, 2180.0)
    vector_math_072.location = (49660.0, 2020.0)
    math_223.location = (260.0, -180.0)
    math_224.location = (420.0, -180.0)
    separate_xyz_041.location = (-2800.0, 3860.0)
    customization_camo_tiler_array.location = (950.0, -555.0)
    vector_math_033.location = (1380.0, -480.0)
    gamma_003.location = (-2800.0, 5780.0)
    vector_math_065.location = (48380.0, 3340.0)
    pattern_lut_03.location = (1510.0, -147.0)
    math_225.location = (420.0, -20.0)
    vector_math_070.location = (48700.0, 2700.0)
    math_181.location = (48540.0, 3180.0)
    vector_math_068.location = (49020.0, 2500.0)
    gamma_002.location = (-2800.0, 2580.0)
    primary_material_lut_05.location = (-128.0, 187.0)
    combine_xyz_079.location = (-288.0, -93.0)
    combine_xyz_082.location = (-288.0, -93.0)
    vector_math_007.location = (1790.0, -328.0)
    separate_xyz_064.location = (310.0, -307.0)
    math_166.location = (470.0, -307.0)
    math_167.location = (630.0, -307.0)
    math_173.location = (1430.0, -487.0)
    math_174.location = (1590.0, -487.0)
    vector_math_060.location = (1750.0, -487.0)
    vector_math_061.location = (1910.0, -487.0)
    vector_math_058.location = (150.0, -307.0)
    mix_015.location = (600.0, 53.0)
    vector_math_003.location = (190.0, -328.0)
    vector_math.location = (1220.0, -480.0)
    customization_material_detail_tiler_array_001.location = (960.0, -480.0)
    math_151.location = (-480.0, 20.0)
    math_153.location = (-320.0, -320.0)
    separate_xyz_048.location = (3103.0, -320.0)
    composite_array.location = (600.0, 413.0)
    vector_math_022.location = (860.0, 413.0)
    clamp_019.location = (1470.0, -488.0)
    vector_math_097.location = (730.0, -30.0)
    vector_math_099.location = (570.0, -30.0)
    vector_math_096.location = (250.0, -30.0)
    vector_math_098.location = (410.0, -30.0)
    vector_math_091.location = (680.0, 160.0)
    combine_xyz_098.location = (-120.0, -40.0)
    vector_math_113.location = (40.0, -40.0)
    vector_math_112.location = (40.0, 160.0)
    vector_math_021.location = (360.0, 20.0)
    mix_021.location = (200.0, 160.0)
    vector_math_114.location = (-230.0, -30.0)
    vector_math_111.location = (-70.0, -30.0)
    vector_math_115.location = (90.0, -30.0)
    vector_math_089.location = (520.0, 160.0)
    vector_math_116.location = (360.0, 160.0)
    normal_map_001.location = (2330.0, -30.0)
    math_270.location = (1050.0, -210.0)
    math_204.location = (1210.0, -210.0)
    math_205.location = (1210.0, -30.0)
    math_268.location = (1370.0, -30.0)
    math_269.location = (1530.0, -30.0)
    combine_xyz_099.location = (1690.0, -30.0)
    vector_math_093.location = (2010.0, -30.0)
    vector_math_092.location = (1850.0, -30.0)
    vector_math_094.location = (2170.0, -30.0)
    math_271.location = (1050.0, -30.0)
    separate_xyz_056.location = (890.0, -30.0)
    math_272.location = (1000.0, -20.0)
    math_273.location = (1160.0, -20.0)
    math_274.location = (1160.0, 160.0)
    math_275.location = (1320.0, 160.0)
    math_276.location = (1480.0, 160.0)
    combine_xyz_100.location = (1640.0, 160.0)
    vector_math_110.location = (1960.0, 160.0)
    vector_math_117.location = (1800.0, 160.0)
    vector_math_118.location = (2120.0, 160.0)
    math_277.location = (1000.0, 160.0)
    separate_xyz_063.location = (840.0, 160.0)
    mix_shader.location = (190.0, -80.0)
    separate_xyz_051.location = (-130.0, -80.0)
    principled_bsdf_001.location = (-20.0, -40.0)
    normal_map.location = (2280.0, 160.0)
    math_050.location = (52460.0, 4200.0)
    gamma_001.location = (52460.0, 4040.0)
    gamma.location = (52460.0, 3940.0)
    math_278.location = (52460.0, 4360.0)
    clamp_018.location = (100.0, -180.0)
    mix_033.location = (152.0, 187.0)
    combine_xyz_096.location = (-288.0, -93.0)
    combine_xyz_097.location = (-288.0, 187.0)
    separate_xyz_036.location = (-2800.0, 5140.0)
    primary_material_lut_22.location = (-128.0, 187.0)
    math_084.location = (1083.0, -320.0)
    value_006.location = (1083.0, -460.0)
    math_085.location = (1243.0, -320.0)
    object_info.location = (1083.0, -580.0)
    mix_075.location = (152.0, 187.0)
    separate_xyz_024.location = (-2800.0, -8300.0)
    math_086.location = (1403.0, -320.0)
    math_229.location = (1243.0, -580.0)
    math_228.location = (1243.0, -740.0)
    math_226.location = (1563.0, -320.0)
    detail_uvs.location = (1403.0, -460.0)
    math_079.location = (1403.0, -720.0)
    combine_xyz_029.location = (1403.0, -580.0)
    math_038.location = (990.0, -95.0)
    math_041.location = (10640.0, 5320.0)
    combine_xyz_086.location = (-288.0, -93.0)
    combine_xyz_089.location = (-288.0, 187.0)
    combine_xyz_088.location = (-288.0, -93.0)
    combine_xyz_090.location = (-288.0, -93.0)
    combine_xyz_091.location = (-288.0, 187.0)
    math_227.location = (1723.0, -320.0)
    vector_math_001.location = (1563.0, -600.0)
    math_042.location = (10800.0, 5320.0)
    math_025.location = (-380.0, 573.0)
    math_039.location = (1150.0, -95.0)
    combine_xyz_087.location = (-288.0, 187.0)
    separate_xyz_018.location = (-2800.0, 3220.0)
    primary_material_lut_17.location = (-128.0, 187.0)
    primary_material_lut_18.location = (-128.0, 187.0)
    primary_material_lut_19.location = (-128.0, 187.0)
    math_003.location = (-2440.0, 5780.0)
    combine_xyz_026.location = (1883.0, -320.0)
    vector_math_002.location = (1723.0, -600.0)
    math_043.location = (10960.0, 5320.0)
    combine_xyz_020.location = (10960.0, 5160.0)
    math_026.location = (-220.0, 573.0)
    math_040.location = (1310.0, -95.0)
    separate_xyz_016.location = (830.0, -95.0)
    separate_xyz_025.location = (-4960.0, 5780.0)
    math_236.location = (0.0, 380.0)
    mix_066.location = (152.0, -33.0)
    mix_068.location = (152.0, -33.0)
    mix_069.location = (152.0, 187.0)
    vector_math_078.location = (2043.0, -320.0)
    vector_math_019.location = (11120.0, 5320.0)
    math_060.location = (20800.0, 3880.0)
    math_027.location = (-60.0, 573.0)
    combine_xyz_018.location = (1490.0, -95.0)
    combine_xyz_017.location = (1490.0, -275.0)
    math_237.location = (0.0, 200.0)
    combine_xyz_093.location = (-288.0, 187.0)
    combine_xyz_092.location = (-288.0, -93.0)
    math_080.location = (-1480.0, 5620.0)
    vector_math_083.location = (2203.0, -320.0)
    math_061.location = (20960.0, 3880.0)
    value_008.location = (100.0, 253.0)
    math_028.location = (100.0, 573.0)
    math_029.location = (100.0, 413.0)
    reroute.location = (100.0, 133.0)
    reroute_001.location = (100.0, 93.0)
    vector_math_016.location = (1650.0, -95.0)
    separate_xyz_020.location = (11280.0, 5320.0)
    Slot1and2.location = (160.0, 380.0)
    math_238.location = (0.0, 20.0)
    primary_material_lut_20.location = (-128.0, 187.0)
    math_081.location = (-1320.0, 5620.0)
    separate_xyz_035.location = (2623.0, -320.0)
    separate_xyz_040.location = (11760.0, 4580.0)
    math_062.location = (21120.0, 3880.0)
    separate_xyz_032.location = (20640.0, 3880.0)
    combine_xyz_005.location = (260.0, 573.0)
    combine_xyz_006.location = (260.0, 413.0)
    separate_xyz_017.location = (1810.0, -95.0)
    value_001.location = (1810.0, -275.0)
    Slot3.location = (320.0, 380.0)
    customization_material_detail_tiler_array.location = (2363.0, -320.0)
    mix_039.location = (152.0, 187.0)
    combine_xyz_060.location = (-288.0, 187.0)
    combine_xyz_061.location = (-288.0, -93.0)
    combine_xyz_041.location = (11440.0, 5320.0)
    math_239.location = (0.0, -160.0)
    mix_072.location = (152.0, -33.0)
    separate_xyz_009.location = (3200.0, 4660.0)
    math_082.location = (1083.0, 20.0)
    combine_xyz_039.location = (2783.0, -320.0)
    combine_xyz_024.location = (11920.0, 4720.0)
    combine_xyz_030.location = (21280.0, 3880.0)
    combine_xyz_031.location = (21280.0, 3740.0)
    math_249.location = (290.0, -60.0)
    vector_math_040.location = (420.0, 573.0)
    combine_xyz_019.location = (1970.0, -95.0)
    mix_018.location = (-5120.0, 5540.0)
    Slot4.location = (480.0, 380.0)
    primary_material_lut_04.location = (-128.0, 187.0)
    vector_math_102.location = (11280.0, 5020.0)
    separate_xyz_043.location = (-4960.0, 5540.0)
    math_240.location = (0.0, -340.0)
    combine_xyz_054.location = (-288.0, 187.0)
    combine_xyz_055.location = (-288.0, -93.0)
    math_083.location = (1243.0, 20.0)
    combine_xyz_045.location = (-287.0, -420.0)
    combine_xyz_046.location = (14060.0, 4720.0)
    vector_math_027.location = (21440.0, 3880.0)
    separate_xyz_037.location = (130.0, -220.0)
    math_248.location = (450.0, -60.0)
    math_031.location = (580.0, 433.0)
    vector_math_041.location = (580.0, 573.0)
    combine_xyz_003.location = (410.0, -95.0)
    vector_math_017.location = (2130.0, -95.0)
    id_mask_array_02.location = (-5380.0, 5780.0)
    Slot5.location = (640.0, 380.0)
    vector_math_084.location = (2943.0, -320.0)
    mix_040.location = (152.0, -33.0)
    separate_xyz_005.location = (130.0, -60.0)
    math_241.location = (0.0, -520.0)
    primary_material_lut_01.location = (-128.0, 187.0)
    mix_005.location = (3040.0, 4660.0)
    math_233.location = (11280.0, 4880.0)
    separate_xyz_033.location = (21600.0, 3880.0)
    math_234.location = (610.0, -60.0)
    math_030.location = (740.0, 573.0)
    vector_math_018.location = (2290.0, -95.0)
    separate_xyz_059.location = (2610.0, -535.0)
    mix_019.location = (-5120.0, 5120.0)
    uv_map.location = (-5700.0, 5780.0)
    Slot6.location = (800.0, 380.0)
    pattern_lut_02.location = (570.0, -95.0)
    math_242.location = (0.0, -700.0)
    combine_xyz_094.location = (-288.0, -93.0)
    combine_xyz_095.location = (-288.0, 187.0)
    combine_xyz_052.location = (-288.0, 187.0)
    combine_xyz_053.location = (-288.0, -93.0)
    combine_xyz_067.location = (-288.0, -93.0)
    mix_045.location = (152.0, 187.0)
    combine_xyz_066.location = (-288.0, 187.0)
    mix_034.location = (152.0, -33.0)
    separate_xyz_022.location = (11280.0, 4720.0)
    combine_xyz_032.location = (21760.0, 3880.0)
    separate_xyz_069.location = (21760.0, 3580.0)
    clamp_008.location = (900.0, 573.0)
    math_119.location = (2770.0, -535.0)
    math_110.location = (2450.0, -95.0)
    combine_xyz_002.location = (-330.0, 185.0)
    mapping.location = (-5540.0, 5780.0)
    Slot7.location = (960.0, 380.0)
    group_input_1.location = (-5920.0, 5680.0)
    math_250.location = (770.0, -60.0)
    primary_material_lut_21.location = (-128.0, 187.0)
    primary_material_lut_00.location = (-128.0, 187.0)
    primary_material_lut_07.location = (-128.0, 187.0)
    combine_xyz_023.location = (11440.0, 4720.0)
    math_067.location = (12940.0, 4720.0)
    vector_math_104.location = (21920.0, 3880.0)
    math_259.location = (22240.0, 3880.0)
    combine_xyz_033.location = (21920.0, 3580.0)
    math_033.location = (1060.0, 213.0)
    math_120.location = (2930.0, -535.0)
    clamp_014.location = (2610.0, -95.0)
    pattern_mask_array_02.location = (-5380.0, 5500.0)
    Slot8.location = (1120.0, 380.0)
    pattern_lut.location = (-170.0, 185.0)
    gamma_pattern.location = (100, 300.0)
    mix_074.location = (152.0, -33.0)
    mix_031.location = (152.0, -33.0)
    combine_xyz_064.location = (-288.0, 187.0)
    combine_xyz_065.location = (-288.0, -93.0)
    mix_046.location = (152.0, -33.0)
    vector_math_020.location = (11600.0, 4720.0)
    clamp_005.location = (13100.0, 4720.0)
    math_258.location = (22400.0, 3880.0)
    separate_xyz_034.location = (22080.0, 3880.0)
    separate_xyz_049.location = (22080.0, 3580.0)
    math_032.location = (1220.0, 213.0)
    math_115.location = (3090.0, -535.0)
    math_036.location = (90.0, 185.0)
    math_121.location = (820.0, -147.0)
    math_126.location = (870.0, 33.0)
    math_116.location = (3250.0, -535.0)
    math_034.location = (-70.0, -128.0)
    mix_030.location = (-5120.0, 4760.0)
    math_243.location = (0.0, -880.0)
    separate_xyz_023.location = (11760.0, 4720.0)
    math_256.location = (21920.0, 3740.0)
    math_257.location = (22560.0, 3880.0)
    math_020.location = (-330.0, 65.0)
    math_089.location = (1380.0, 213.0)
    math_023.location = (-700.0, 753.0)
    math_035.location = (90.0, -128.0)
    math_037.location = (250.0, 185.0)
    math_117.location = (340.0, 233.0)
    math_122.location = (980.0, -147.0)
    math_125.location = (1030.0, 33.0)
    primary_material_lut_06.location = (-128.0, 187.0)
    mix_044.location = (152.0, -33.0)
    math_260.location = (22720.0, 3880.0)
    math_021.location = (-170.0, 65.0)
    math_090.location = (1540.0, 213.0)
    math_024.location = (-540.0, 753.0)
    math_118.location = (500.0, 233.0)
    math_124.location = (550.0, 213.0)
    math_127.location = (1190.0, 33.0)
    math_175.location = (37540.0, 4100.0)
    math_063.location = (22880.0, 3880.0)
    math_091.location = (1700.0, 213.0)
    math_180.location = (36020.0, 2560.0)
    math_123.location = (710.0, 213.0)
    math_128.location = (1350.0, 33.0)
    math_176.location = (41000.0, 4100.0)
    mix_008.location = (36180.0, 2560.0)
    math_129.location = (1510.0, 33.0)
    math_177.location = (41720.0, 4100.0)
    mix_006.location = (-60.0, -180.0)
    separate_xyz_060.location = (-2800.0, -620.0)
    mix_078.location = (-20.0, 60.0)
    math_145.location = (3300.0, 53.0)
    math_146.location = (3460.0, 53.0)
    mix_011.location = (3620.0, 53.0)
    group_output_1.location = (53320.0, 5100.0)
    math_194.location = (30.0, -80.0)
    
    #initialize HD2_Shader links
    connect_input_links(HD2_Shader)

    #combine_xyz_002.Vector -> pattern_lut.Vector
    HD2_Shader.links.new(combine_xyz_002.outputs[0], pattern_lut.inputs[0])
    #combine_xyz_003.Vector -> pattern_lut_02.Vector
    HD2_Shader.links.new(combine_xyz_003.outputs[0], pattern_lut_02.inputs[0])
    #mix_031.Result -> math_034.Value
    HD2_Shader.links.new(mix_031.outputs[0], math_034.inputs[0])
    #math_034.Value -> math_035.Value
    HD2_Shader.links.new(math_034.outputs[0], math_035.inputs[1])
    #math_036.Value -> math_037.Value
    HD2_Shader.links.new(math_036.outputs[0], math_037.inputs[1])
    #pattern_lut.Alpha -> math_036.Value
    HD2_Shader.links.new(pattern_lut.outputs[1], math_036.inputs[0])
    #pattern_lut_02.Color -> separate_xyz_016.Vector
    HD2_Shader.links.new(pattern_lut_02.outputs[0], separate_xyz_016.inputs[0])
    #separate_xyz_016.Z -> math_038.Value
    HD2_Shader.links.new(separate_xyz_016.outputs[2], math_038.inputs[0])
    #math_003.Value -> math_038.Value
    HD2_Shader.links.new(math_003.outputs[0], math_038.inputs[1])
    #math_038.Value -> math_039.Value
    HD2_Shader.links.new(math_038.outputs[0], math_039.inputs[1])
    #math_039.Value -> math_040.Value
    HD2_Shader.links.new(math_039.outputs[0], math_040.inputs[1])
    #separate_xyz_016.X -> combine_xyz_017.X
    HD2_Shader.links.new(separate_xyz_016.outputs[0], combine_xyz_017.inputs[0])
    #separate_xyz_016.Y -> combine_xyz_017.Y
    HD2_Shader.links.new(separate_xyz_016.outputs[1], combine_xyz_017.inputs[1])
    #combine_xyz_017.Vector -> vector_math_016.Vector
    HD2_Shader.links.new(combine_xyz_017.outputs[0], vector_math_016.inputs[0])
    #math_040.Value -> combine_xyz_018.X
    HD2_Shader.links.new(math_040.outputs[0], combine_xyz_018.inputs[0])
    #math_040.Value -> combine_xyz_018.Y
    HD2_Shader.links.new(math_040.outputs[0], combine_xyz_018.inputs[1])
    #combine_xyz_018.Vector -> vector_math_016.Vector
    HD2_Shader.links.new(combine_xyz_018.outputs[0], vector_math_016.inputs[1])
    #vector_math_016.Vector -> separate_xyz_017.Vector
    HD2_Shader.links.new(vector_math_016.outputs[0], separate_xyz_017.inputs[0])
    #separate_xyz_017.X -> combine_xyz_019.X
    HD2_Shader.links.new(separate_xyz_017.outputs[0], combine_xyz_019.inputs[0])
    #separate_xyz_017.Y -> combine_xyz_019.Y
    HD2_Shader.links.new(separate_xyz_017.outputs[1], combine_xyz_019.inputs[1])
    #value_001.Value -> combine_xyz_019.Z
    HD2_Shader.links.new(value_001.outputs[0], combine_xyz_019.inputs[2])
    #combine_xyz_019.Vector -> vector_math_017.Vector
    HD2_Shader.links.new(combine_xyz_019.outputs[0], vector_math_017.inputs[0])
    #pattern_lut_02.Color -> vector_math_017.Vector
    HD2_Shader.links.new(pattern_lut_02.outputs[0], vector_math_017.inputs[1])
    #vector_math_017.Vector -> vector_math_018.Vector
    HD2_Shader.links.new(vector_math_017.outputs[0], vector_math_018.inputs[0])
    #separate_xyz_018.Z -> math_041.Value
    HD2_Shader.links.new(separate_xyz_018.outputs[2], math_041.inputs[0])
    #math_003.Value -> math_041.Value
    HD2_Shader.links.new(math_003.outputs[0], math_041.inputs[1])
    #math_041.Value -> math_042.Value
    HD2_Shader.links.new(math_041.outputs[0], math_042.inputs[1])
    #math_042.Value -> math_043.Value
    HD2_Shader.links.new(math_042.outputs[0], math_043.inputs[1])
    #separate_xyz_018.X -> combine_xyz_020.X
    HD2_Shader.links.new(separate_xyz_018.outputs[0], combine_xyz_020.inputs[0])
    #separate_xyz_018.Y -> combine_xyz_020.Y
    HD2_Shader.links.new(separate_xyz_018.outputs[1], combine_xyz_020.inputs[1])
    #combine_xyz_020.Vector -> vector_math_019.Vector
    HD2_Shader.links.new(combine_xyz_020.outputs[0], vector_math_019.inputs[0])
    #vector_math_019.Vector -> separate_xyz_020.Vector
    HD2_Shader.links.new(vector_math_019.outputs[0], separate_xyz_020.inputs[0])
    #separate_xyz_022.Z -> combine_xyz_023.Y
    HD2_Shader.links.new(separate_xyz_022.outputs[2], combine_xyz_023.inputs[1])
    #combine_xyz_023.Vector -> vector_math_020.Vector
    HD2_Shader.links.new(combine_xyz_023.outputs[0], vector_math_020.inputs[0])
    #vector_math_020.Vector -> separate_xyz_023.Vector
    HD2_Shader.links.new(vector_math_020.outputs[0], separate_xyz_023.inputs[0])
    #separate_xyz_023.X -> combine_xyz_024.Z
    HD2_Shader.links.new(separate_xyz_023.outputs[0], combine_xyz_024.inputs[2])
    #separate_xyz_024.X -> math_044.Value
    HD2_Shader.links.new(separate_xyz_024.outputs[0], math_044.inputs[1])
    #math_044.Value -> math_045.Value
    HD2_Shader.links.new(math_044.outputs[0], math_045.inputs[1])
    #math_046.Value -> math_047.Value
    HD2_Shader.links.new(math_046.outputs[0], math_047.inputs[0])
    #value_002.Value -> math_047.Value
    HD2_Shader.links.new(value_002.outputs[0], math_047.inputs[1])
    #math_047.Value -> math_048.Value
    HD2_Shader.links.new(math_047.outputs[0], math_048.inputs[0])
    #vector_math_022.Vector -> separate_xyz_027.Vector
    HD2_Shader.links.new(vector_math_022.outputs[0], separate_xyz_027.inputs[0])
    #separate_xyz_027.Y -> math_049.Value
    HD2_Shader.links.new(separate_xyz_027.outputs[1], math_049.inputs[0])
    #separate_xyz_027.Z -> combine_xyz_027.Y
    HD2_Shader.links.new(separate_xyz_027.outputs[2], combine_xyz_027.inputs[1])
    #math_003.Value -> clamp_003.Value
    HD2_Shader.links.new(math_003.outputs[0], clamp_003.inputs[0])
    #math_191.Value -> clamp_004.Value
    HD2_Shader.links.new(math_191.outputs[0], clamp_004.inputs[0])
    #separate_xyz_028.Z -> combine_xyz_028.X
    HD2_Shader.links.new(separate_xyz_028.outputs[2], combine_xyz_028.inputs[0])
    #separate_xyz_028.Y -> combine_xyz_028.Y
    HD2_Shader.links.new(separate_xyz_028.outputs[1], combine_xyz_028.inputs[1])
    #separate_xyz_028.X -> combine_xyz_028.Z
    HD2_Shader.links.new(separate_xyz_028.outputs[0], combine_xyz_028.inputs[2])
    #combine_xyz_028.Vector -> vector_math_023.Vector
    HD2_Shader.links.new(combine_xyz_028.outputs[0], vector_math_023.inputs[0])
    #vector_math_023.Vector -> vector_math_024.Vector
    HD2_Shader.links.new(vector_math_023.outputs[0], vector_math_024.inputs[0])
    #vector_math_024.Vector -> vector_math_025.Vector
    HD2_Shader.links.new(vector_math_024.outputs[0], vector_math_025.inputs[0])
    #vector_math_024.Vector -> vector_math_025.Vector
    HD2_Shader.links.new(vector_math_024.outputs[0], vector_math_025.inputs[1])
    #vector_math_024.Vector -> separate_xyz_029.Vector
    HD2_Shader.links.new(vector_math_024.outputs[0], separate_xyz_029.inputs[0])
    #separate_xyz_029.X -> math_051.Value
    HD2_Shader.links.new(separate_xyz_029.outputs[0], math_051.inputs[0])
    #separate_xyz_029.X -> math_052.Value
    HD2_Shader.links.new(separate_xyz_029.outputs[0], math_052.inputs[1])
    #math_051.Value -> math_052.Value
    HD2_Shader.links.new(math_051.outputs[0], math_052.inputs[0])
    #separate_xyz_029.Y -> math_054.Value
    HD2_Shader.links.new(separate_xyz_029.outputs[1], math_054.inputs[0])
    #separate_xyz_029.Y -> math_055.Value
    HD2_Shader.links.new(separate_xyz_029.outputs[1], math_055.inputs[1])
    #math_052.Value -> math_055.Value
    HD2_Shader.links.new(math_052.outputs[0], math_055.inputs[2])
    #vector_math_025.Vector -> separate_xyz_030.Vector
    HD2_Shader.links.new(vector_math_025.outputs[0], separate_xyz_030.inputs[0])
    #separate_xyz_030.X -> math_056.Value
    HD2_Shader.links.new(separate_xyz_030.outputs[0], math_056.inputs[0])
    #separate_xyz_030.Y -> math_056.Value
    HD2_Shader.links.new(separate_xyz_030.outputs[1], math_056.inputs[1])
    #math_056.Value -> math_053.Value
    HD2_Shader.links.new(math_056.outputs[0], math_053.inputs[1])
    #math_053.Value -> math_057.Value
    HD2_Shader.links.new(math_053.outputs[0], math_057.inputs[0])
    #math_055.Value -> math_057.Value
    HD2_Shader.links.new(math_055.outputs[0], math_057.inputs[1])
    #math_057.Value -> math_058.Value
    HD2_Shader.links.new(math_057.outputs[0], math_058.inputs[0])
    #math_058.Value -> math_059.Value
    HD2_Shader.links.new(math_058.outputs[0], math_059.inputs[0])
    #math_059.Value -> vector_math_026.Vector
    HD2_Shader.links.new(math_059.outputs[0], vector_math_026.inputs[0])
    #vector_math_024.Vector -> vector_math_026.Vector
    HD2_Shader.links.new(vector_math_024.outputs[0], vector_math_026.inputs[1])
    #vector_math_026.Vector -> separate_xyz_031.Vector
    HD2_Shader.links.new(vector_math_026.outputs[0], separate_xyz_031.inputs[0])
    #mix_045.Result -> separate_xyz_032.Vector
    HD2_Shader.links.new(mix_045.outputs[2], separate_xyz_032.inputs[0])
    #separate_xyz_032.Z -> math_060.Value
    HD2_Shader.links.new(separate_xyz_032.outputs[2], math_060.inputs[0])
    #math_003.Value -> math_060.Value
    HD2_Shader.links.new(math_003.outputs[0], math_060.inputs[1])
    #math_060.Value -> math_061.Value
    HD2_Shader.links.new(math_060.outputs[0], math_061.inputs[1])
    #math_061.Value -> math_062.Value
    HD2_Shader.links.new(math_061.outputs[0], math_062.inputs[1])
    #math_062.Value -> combine_xyz_030.X
    HD2_Shader.links.new(math_062.outputs[0], combine_xyz_030.inputs[0])
    #math_062.Value -> combine_xyz_030.Y
    HD2_Shader.links.new(math_062.outputs[0], combine_xyz_030.inputs[1])
    #separate_xyz_032.X -> combine_xyz_031.X
    HD2_Shader.links.new(separate_xyz_032.outputs[0], combine_xyz_031.inputs[0])
    #separate_xyz_032.Y -> combine_xyz_031.Y
    HD2_Shader.links.new(separate_xyz_032.outputs[1], combine_xyz_031.inputs[1])
    #combine_xyz_031.Vector -> vector_math_027.Vector
    HD2_Shader.links.new(combine_xyz_031.outputs[0], vector_math_027.inputs[0])
    #combine_xyz_030.Vector -> vector_math_027.Vector
    HD2_Shader.links.new(combine_xyz_030.outputs[0], vector_math_027.inputs[1])
    #vector_math_027.Vector -> separate_xyz_033.Vector
    HD2_Shader.links.new(vector_math_027.outputs[0], separate_xyz_033.inputs[0])
    #separate_xyz_023.X -> combine_xyz_033.Z
    HD2_Shader.links.new(separate_xyz_023.outputs[0], combine_xyz_033.inputs[2])
    #separate_xyz_069.X -> combine_xyz_033.X
    HD2_Shader.links.new(separate_xyz_069.outputs[0], combine_xyz_033.inputs[0])
    #separate_xyz_069.Y -> combine_xyz_033.Y
    HD2_Shader.links.new(separate_xyz_069.outputs[1], combine_xyz_033.inputs[1])
    #mix_044.Result -> math_063.Value
    HD2_Shader.links.new(mix_044.outputs[0], math_063.inputs[1])
    #math_003.Value -> math_064.Value
    HD2_Shader.links.new(math_003.outputs[0], math_064.inputs[1])
    #separate_xyz_038.Z -> math_064.Value
    HD2_Shader.links.new(separate_xyz_038.outputs[2], math_064.inputs[0])
    #math_064.Value -> math_065.Value
    HD2_Shader.links.new(math_064.outputs[0], math_065.inputs[1])
    #math_065.Value -> math_066.Value
    HD2_Shader.links.new(math_065.outputs[0], math_066.inputs[1])
    #separate_xyz_038.X -> combine_xyz_035.X
    HD2_Shader.links.new(separate_xyz_038.outputs[0], combine_xyz_035.inputs[0])
    #separate_xyz_038.Y -> combine_xyz_035.Y
    HD2_Shader.links.new(separate_xyz_038.outputs[1], combine_xyz_035.inputs[1])
    #combine_xyz_035.Vector -> vector_math_028.Vector
    HD2_Shader.links.new(combine_xyz_035.outputs[0], vector_math_028.inputs[0])
    #math_066.Value -> combine_xyz_036.X
    HD2_Shader.links.new(math_066.outputs[0], combine_xyz_036.inputs[0])
    #math_066.Value -> combine_xyz_036.Y
    HD2_Shader.links.new(math_066.outputs[0], combine_xyz_036.inputs[1])
    #combine_xyz_036.Vector -> vector_math_028.Vector
    HD2_Shader.links.new(combine_xyz_036.outputs[0], vector_math_028.inputs[1])
    #vector_math_028.Vector -> separate_xyz_039.Vector
    HD2_Shader.links.new(vector_math_028.outputs[0], separate_xyz_039.inputs[0])
    #separate_xyz_039.X -> combine_xyz_037.X
    HD2_Shader.links.new(separate_xyz_039.outputs[0], combine_xyz_037.inputs[0])
    #separate_xyz_039.Y -> combine_xyz_037.Y
    HD2_Shader.links.new(separate_xyz_039.outputs[1], combine_xyz_037.inputs[1])
    #value_005.Value -> combine_xyz_037.Z
    HD2_Shader.links.new(value_005.outputs[0], combine_xyz_037.inputs[2])
    #combine_xyz_037.Vector -> vector_math_029.Vector
    HD2_Shader.links.new(combine_xyz_037.outputs[0], vector_math_029.inputs[0])
    #mix_047.Result -> vector_math_029.Vector
    HD2_Shader.links.new(mix_047.outputs[2], vector_math_029.inputs[1])
    #vector_math_029.Vector -> vector_math_030.Vector
    HD2_Shader.links.new(vector_math_029.outputs[0], vector_math_030.inputs[0])
    #separate_xyz_040.X -> combine_xyz_024.X
    HD2_Shader.links.new(separate_xyz_040.outputs[0], combine_xyz_024.inputs[0])
    #separate_xyz_040.Y -> combine_xyz_024.Y
    HD2_Shader.links.new(separate_xyz_040.outputs[1], combine_xyz_024.inputs[1])
    #math_067.Value -> clamp_005.Value
    HD2_Shader.links.new(math_067.outputs[0], clamp_005.inputs[0])
    #mix_037.Result -> separate_xyz_041.Vector
    HD2_Shader.links.new(mix_037.outputs[2], separate_xyz_041.inputs[0])
    #separate_xyz_041.Z -> math_068.Value
    HD2_Shader.links.new(separate_xyz_041.outputs[2], math_068.inputs[0])
    #math_003.Value -> math_068.Value
    HD2_Shader.links.new(math_003.outputs[0], math_068.inputs[1])
    #math_068.Value -> math_069.Value
    HD2_Shader.links.new(math_068.outputs[0], math_069.inputs[1])
    #math_069.Value -> math_070.Value
    HD2_Shader.links.new(math_069.outputs[0], math_070.inputs[1])
    #math_070.Value -> vector_math_031.Vector
    HD2_Shader.links.new(math_070.outputs[0], vector_math_031.inputs[1])
    #combine_xyz_038.Vector -> vector_math_031.Vector
    HD2_Shader.links.new(combine_xyz_038.outputs[0], vector_math_031.inputs[0])
    #vector_math_031.Vector -> separate_xyz_044.Vector
    HD2_Shader.links.new(vector_math_031.outputs[0], separate_xyz_044.inputs[0])
    #math_071.Value -> clamp_006.Value
    HD2_Shader.links.new(math_071.outputs[0], clamp_006.inputs[0])
    #clamp_006.Result -> math_072.Value
    HD2_Shader.links.new(clamp_006.outputs[0], math_072.inputs[1])
    #math_072.Value -> math_073.Value
    HD2_Shader.links.new(math_072.outputs[0], math_073.inputs[1])
    #clamp_006.Result -> math_073.Value
    HD2_Shader.links.new(clamp_006.outputs[0], math_073.inputs[2])
    #separate_xyz_031.Y -> vector_math_032.Vector
    HD2_Shader.links.new(separate_xyz_031.outputs[1], vector_math_032.inputs[0])
    #math_075.Value -> math_074.Value
    HD2_Shader.links.new(math_075.outputs[0], math_074.inputs[1])
    #vector_math_023.Vector -> separate_xyz_045.Vector
    HD2_Shader.links.new(vector_math_023.outputs[0], separate_xyz_045.inputs[0])
    #separate_xyz_045.X -> math_075.Value
    HD2_Shader.links.new(separate_xyz_045.outputs[0], math_075.inputs[0])
    #math_074.Value -> math_073.Value
    HD2_Shader.links.new(math_074.outputs[0], math_073.inputs[0])
    #math_073.Value -> math_076.Value
    HD2_Shader.links.new(math_073.outputs[0], math_076.inputs[0])
    #math_074.Value -> math_076.Value
    HD2_Shader.links.new(math_074.outputs[0], math_076.inputs[1])
    #math_244.Value -> math_077.Value
    HD2_Shader.links.new(math_244.outputs[0], math_077.inputs[0])
    #math_077.Value -> math_078.Value
    HD2_Shader.links.new(math_077.outputs[0], math_078.inputs[1])
    #math_049.Value -> combine_xyz_027.X
    HD2_Shader.links.new(math_049.outputs[0], combine_xyz_027.inputs[0])
    #separate_xyz_024.X -> math_079.Value
    HD2_Shader.links.new(separate_xyz_024.outputs[0], math_079.inputs[1])
    #mix_031.Result -> math_080.Value
    HD2_Shader.links.new(mix_031.outputs[0], math_080.inputs[1])
    #math_080.Value -> math_081.Value
    HD2_Shader.links.new(math_080.outputs[0], math_081.inputs[1])
    #math_081.Value -> math_082.Value
    HD2_Shader.links.new(math_081.outputs[0], math_082.inputs[0])
    #math_082.Value -> math_083.Value
    HD2_Shader.links.new(math_082.outputs[0], math_083.inputs[1])
    #math_084.Value -> math_085.Value
    HD2_Shader.links.new(math_084.outputs[0], math_085.inputs[0])
    #value_006.Value -> math_085.Value
    HD2_Shader.links.new(value_006.outputs[0], math_085.inputs[1])
    #math_085.Value -> math_086.Value
    HD2_Shader.links.new(math_085.outputs[0], math_086.inputs[0])
    #separate_xyz_048.Z -> combine_xyz_043.X
    HD2_Shader.links.new(separate_xyz_048.outputs[2], combine_xyz_043.inputs[0])
    #combine_xyz_043.Vector -> vector_math_035.Vector
    HD2_Shader.links.new(combine_xyz_043.outputs[0], vector_math_035.inputs[1])
    #separate_xyz_023.X -> combine_xyz_046.Z
    HD2_Shader.links.new(separate_xyz_023.outputs[0], combine_xyz_046.inputs[2])
    #math_083.Value -> mix_005.Factor
    HD2_Shader.links.new(math_083.outputs[0], mix_005.inputs[0])
    #combine_xyz_045.Vector -> mix_005.A
    HD2_Shader.links.new(combine_xyz_045.outputs[0], mix_005.inputs[4])
    #math_088.Value -> clamp_007.Value
    HD2_Shader.links.new(math_088.outputs[0], clamp_007.inputs[0])
    #mix_036.Result -> math_002.Value
    HD2_Shader.links.new(mix_036.outputs[0], math_002.inputs[1])
    #math_002.Value -> math_004.Value
    HD2_Shader.links.new(math_002.outputs[0], math_004.inputs[1])
    #clamp_007.Result -> math_004.Value
    HD2_Shader.links.new(clamp_007.outputs[0], math_004.inputs[0])
    #math_006.Value -> math_005.Value
    HD2_Shader.links.new(math_006.outputs[0], math_005.inputs[1])
    #math_004.Value -> math_006.Value
    HD2_Shader.links.new(math_004.outputs[0], math_006.inputs[0])
    #mix_042.Result -> math_005.Value
    HD2_Shader.links.new(mix_042.outputs[0], math_005.inputs[0])
    #math_005.Value -> math_007.Value
    HD2_Shader.links.new(math_005.outputs[0], math_007.inputs[1])
    #clamp_005.Result -> math_007.Value
    HD2_Shader.links.new(clamp_005.outputs[0], math_007.inputs[0])
    #math_004.Value -> math_007.Value
    HD2_Shader.links.new(math_004.outputs[0], math_007.inputs[2])
    #math_007.Value -> math_008.Value
    HD2_Shader.links.new(math_007.outputs[0], math_008.inputs[0])
    #math_008.Value -> math_009.Value
    HD2_Shader.links.new(math_008.outputs[0], math_009.inputs[1])
    #mix_013.Result -> vector_math_003.Vector
    HD2_Shader.links.new(mix_013.outputs[1], vector_math_003.inputs[0])
    #mix_013.Result -> vector_math_003.Vector
    HD2_Shader.links.new(mix_013.outputs[1], vector_math_003.inputs[1])
    #vector_math_003.Vector -> vector_math_004.Vector
    HD2_Shader.links.new(vector_math_003.outputs[0], vector_math_004.inputs[0])
    #vector_math_003.Vector -> vector_math_004.Vector
    HD2_Shader.links.new(vector_math_003.outputs[0], vector_math_004.inputs[1])
    #math_011.Value -> math_010.Value
    HD2_Shader.links.new(math_011.outputs[0], math_010.inputs[0])
    #separate_xyz_001.X -> math_011.Value
    HD2_Shader.links.new(separate_xyz_001.outputs[0], math_011.inputs[0])
    #separate_xyz_001.X -> math_010.Value
    HD2_Shader.links.new(separate_xyz_001.outputs[0], math_010.inputs[1])
    #separate_xyz_001.Y -> math_013.Value
    HD2_Shader.links.new(separate_xyz_001.outputs[1], math_013.inputs[0])
    #math_013.Value -> math_012.Value
    HD2_Shader.links.new(math_013.outputs[0], math_012.inputs[0])
    #math_010.Value -> math_012.Value
    HD2_Shader.links.new(math_010.outputs[0], math_012.inputs[2])
    #separate_xyz_001.Y -> math_012.Value
    HD2_Shader.links.new(separate_xyz_001.outputs[1], math_012.inputs[1])
    #vector_math_003.Vector -> separate_xyz_001.Vector
    HD2_Shader.links.new(vector_math_003.outputs[0], separate_xyz_001.inputs[0])
    #vector_math_004.Vector -> separate_xyz_004.Vector
    HD2_Shader.links.new(vector_math_004.outputs[0], separate_xyz_004.inputs[0])
    #separate_xyz_004.X -> math_014.Value
    HD2_Shader.links.new(separate_xyz_004.outputs[0], math_014.inputs[0])
    #separate_xyz_004.Y -> math_014.Value
    HD2_Shader.links.new(separate_xyz_004.outputs[1], math_014.inputs[1])
    #math_014.Value -> math_015.Value
    HD2_Shader.links.new(math_014.outputs[0], math_015.inputs[1])
    #math_015.Value -> math_016.Value
    HD2_Shader.links.new(math_015.outputs[0], math_016.inputs[0])
    #math_012.Value -> math_016.Value
    HD2_Shader.links.new(math_012.outputs[0], math_016.inputs[1])
    #math_016.Value -> math_017.Value
    HD2_Shader.links.new(math_016.outputs[0], math_017.inputs[0])
    #math_017.Value -> math_018.Value
    HD2_Shader.links.new(math_017.outputs[0], math_018.inputs[0])
    #vector_math_003.Vector -> vector_math_005.Vector
    HD2_Shader.links.new(vector_math_003.outputs[0], vector_math_005.inputs[1])
    #math_018.Value -> vector_math_005.Vector
    HD2_Shader.links.new(math_018.outputs[0], vector_math_005.inputs[0])
    #vector_math_005.Vector -> vector_math_006.Vector
    HD2_Shader.links.new(vector_math_005.outputs[0], vector_math_006.inputs[0])
    #clamp_019.Result -> vector_math_006.Vector
    HD2_Shader.links.new(clamp_019.outputs[0], vector_math_006.inputs[1])
    #vector_math_006.Vector -> vector_math_007.Vector
    HD2_Shader.links.new(vector_math_006.outputs[0], vector_math_007.inputs[0])
    #math_007.Value -> vector_math_007.Vector
    HD2_Shader.links.new(math_007.outputs[0], vector_math_007.inputs[1])
    #vector_math_007.Vector -> separate_xyz_006.Vector
    HD2_Shader.links.new(vector_math_007.outputs[0], separate_xyz_006.inputs[0])
    #separate_xyz_006.Y -> vector_math_008.Vector
    HD2_Shader.links.new(separate_xyz_006.outputs[1], vector_math_008.inputs[0])
    #separate_xyz_052.Z -> math_019.Value
    HD2_Shader.links.new(separate_xyz_052.outputs[2], math_019.inputs[1])
    #mix_074.Result -> math_020.Value
    HD2_Shader.links.new(mix_074.outputs[0], math_020.inputs[0])
    #math_020.Value -> math_021.Value
    HD2_Shader.links.new(math_020.outputs[0], math_021.inputs[1])
    #vector_math_009.Vector -> vector_math_010.Vector
    HD2_Shader.links.new(vector_math_009.outputs[0], vector_math_010.inputs[0])
    #separate_xyz_052.X -> vector_math_010.Vector
    HD2_Shader.links.new(separate_xyz_052.outputs[0], vector_math_010.inputs[1])
    #separate_xyz_052.Y -> vector_math_010.Vector
    HD2_Shader.links.new(separate_xyz_052.outputs[1], vector_math_010.inputs[2])
    #vector_math_010.Vector -> separate_xyz_007.Vector
    HD2_Shader.links.new(vector_math_010.outputs[0], separate_xyz_007.inputs[0])
    #separate_xyz_007.X -> clamp.Value
    HD2_Shader.links.new(separate_xyz_007.outputs[0], clamp.inputs[0])
    #separate_xyz_007.Y -> clamp_001.Value
    HD2_Shader.links.new(separate_xyz_007.outputs[1], clamp_001.inputs[0])
    #separate_xyz_007.Z -> clamp_002.Value
    HD2_Shader.links.new(separate_xyz_007.outputs[2], clamp_002.inputs[0])
    #clamp.Result -> combine_xyz_001.X
    HD2_Shader.links.new(clamp.outputs[0], combine_xyz_001.inputs[0])
    #clamp_001.Result -> combine_xyz_001.Y
    HD2_Shader.links.new(clamp_001.outputs[0], combine_xyz_001.inputs[1])
    #clamp_002.Result -> combine_xyz_001.Z
    HD2_Shader.links.new(clamp_002.outputs[0], combine_xyz_001.inputs[2])
    #mix_065.Result -> vector_math_011.Vector
    HD2_Shader.links.new(mix_065.outputs[2], vector_math_011.inputs[0])
    #mix_063.Result -> vector_math_012.Vector
    HD2_Shader.links.new(mix_063.outputs[2], vector_math_012.inputs[0])
    #vector_math_012.Vector -> vector_math_011.Vector
    HD2_Shader.links.new(vector_math_012.outputs[0], vector_math_011.inputs[1])
    #combine_xyz_001.Vector -> separate_xyz_008.Vector
    HD2_Shader.links.new(combine_xyz_001.outputs[0], separate_xyz_008.inputs[0])
    #vector_math_011.Vector -> vector_math_013.Vector
    HD2_Shader.links.new(vector_math_011.outputs[0], vector_math_013.inputs[1])
    #mix_063.Result -> vector_math_013.Vector
    HD2_Shader.links.new(mix_063.outputs[2], vector_math_013.inputs[2])
    #vector_math_013.Vector -> vector_math_014.Vector
    HD2_Shader.links.new(vector_math_013.outputs[0], vector_math_014.inputs[0])
    #mix_067.Result -> vector_math_015.Vector
    HD2_Shader.links.new(mix_067.outputs[2], vector_math_015.inputs[0])
    #vector_math_014.Vector -> vector_math_015.Vector
    HD2_Shader.links.new(vector_math_014.outputs[0], vector_math_015.inputs[1])
    #vector_math_015.Vector -> vector_math_036.Vector
    HD2_Shader.links.new(vector_math_015.outputs[0], vector_math_036.inputs[1])
    #separate_xyz_008.X -> vector_math_013.Vector
    HD2_Shader.links.new(separate_xyz_008.outputs[0], vector_math_013.inputs[0])
    #separate_xyz_008.Y -> vector_math_036.Vector
    HD2_Shader.links.new(separate_xyz_008.outputs[1], vector_math_036.inputs[0])
    #vector_math_013.Vector -> vector_math_036.Vector
    HD2_Shader.links.new(vector_math_013.outputs[0], vector_math_036.inputs[2])
    #vector_math_036.Vector -> vector_math_038.Vector
    HD2_Shader.links.new(vector_math_036.outputs[0], vector_math_038.inputs[0])
    #vector_math_038.Vector -> vector_math_037.Vector
    HD2_Shader.links.new(vector_math_038.outputs[0], vector_math_037.inputs[1])
    #mix_069.Result -> vector_math_037.Vector
    HD2_Shader.links.new(mix_069.outputs[2], vector_math_037.inputs[0])
    #vector_math_037.Vector -> vector_math_039.Vector
    HD2_Shader.links.new(vector_math_037.outputs[0], vector_math_039.inputs[1])
    #separate_xyz_008.Z -> vector_math_039.Vector
    HD2_Shader.links.new(separate_xyz_008.outputs[2], vector_math_039.inputs[0])
    #vector_math_036.Vector -> vector_math_039.Vector
    HD2_Shader.links.new(vector_math_036.outputs[0], vector_math_039.inputs[2])
    #mix_031.Result -> math_023.Value
    HD2_Shader.links.new(mix_031.outputs[0], math_023.inputs[0])
    #math_023.Value -> math_024.Value
    HD2_Shader.links.new(math_023.outputs[0], math_024.inputs[1])
    #math_003.Value -> math_025.Value
    HD2_Shader.links.new(math_003.outputs[0], math_025.inputs[0])
    #mix_070.Result -> math_025.Value
    HD2_Shader.links.new(mix_070.outputs[0], math_025.inputs[1])
    #math_025.Value -> math_026.Value
    HD2_Shader.links.new(math_025.outputs[0], math_026.inputs[1])
    #math_026.Value -> math_027.Value
    HD2_Shader.links.new(math_026.outputs[0], math_027.inputs[1])
    #math_027.Value -> math_028.Value
    HD2_Shader.links.new(math_027.outputs[0], math_028.inputs[1])
    #mix_066.Result -> math_028.Value
    HD2_Shader.links.new(mix_066.outputs[0], math_028.inputs[0])
    #math_027.Value -> math_029.Value
    HD2_Shader.links.new(math_027.outputs[0], math_029.inputs[1])
    #mix_068.Result -> math_029.Value
    HD2_Shader.links.new(mix_068.outputs[0], math_029.inputs[0])
    #mix_066.Result -> reroute.Input
    HD2_Shader.links.new(mix_066.outputs[0], reroute.inputs[0])
    #mix_068.Result -> reroute_001.Input
    HD2_Shader.links.new(mix_068.outputs[0], reroute_001.inputs[0])
    #math_028.Value -> combine_xyz_005.X
    HD2_Shader.links.new(math_028.outputs[0], combine_xyz_005.inputs[0])
    #math_029.Value -> combine_xyz_005.Y
    HD2_Shader.links.new(math_029.outputs[0], combine_xyz_005.inputs[1])
    #value_008.Value -> combine_xyz_005.Z
    HD2_Shader.links.new(value_008.outputs[0], combine_xyz_005.inputs[2])
    #reroute.Output -> combine_xyz_006.X
    HD2_Shader.links.new(reroute.outputs[0], combine_xyz_006.inputs[0])
    #reroute_001.Output -> combine_xyz_006.Y
    HD2_Shader.links.new(reroute_001.outputs[0], combine_xyz_006.inputs[1])
    #mix_070.Value -> combine_xyz_006.Z
    HD2_Shader.links.new(mix_070.outputs[0], combine_xyz_006.inputs[2])
    #combine_xyz_006.Vector -> vector_math_040.Vector
    HD2_Shader.links.new(combine_xyz_006.outputs[0], vector_math_040.inputs[0])
    #combine_xyz_005.Vector -> vector_math_040.Vector
    HD2_Shader.links.new(combine_xyz_005.outputs[0], vector_math_040.inputs[1])
    #vector_math_040.Vector -> vector_math_041.Vector
    HD2_Shader.links.new(vector_math_040.outputs[0], vector_math_041.inputs[0])
    #combine_xyz_046.Vector -> vector_math_041.Vector
    HD2_Shader.links.new(combine_xyz_046.outputs[0], vector_math_041.inputs[1])
    #vector_math_041.Value -> math_030.Value
    HD2_Shader.links.new(vector_math_041.outputs[1], math_030.inputs[0])
    #mix_072.Result -> math_031.Value
    HD2_Shader.links.new(mix_072.outputs[0], math_031.inputs[0])
    #math_031.Value -> math_030.Value
    HD2_Shader.links.new(math_031.outputs[0], math_030.inputs[1])
    #math_030.Value -> clamp_008.Value
    HD2_Shader.links.new(math_030.outputs[0], clamp_008.inputs[0])
    #vector_math_039.Vector -> vector_math_043.Vector
    HD2_Shader.links.new(vector_math_039.outputs[0], gamma_006.inputs[0])
    #vector_math_039.Vector -> vector_math_043.Vector
    HD2_Shader.links.new(gamma_006.outputs[0], vector_math_043.inputs[0])
    #vector_math_043.Vector -> vector_math_042.Vector
    HD2_Shader.links.new(vector_math_043.outputs[0], vector_math_042.inputs[1])
    #gamma_003.Color -> vector_math_042.Vector
    HD2_Shader.links.new(gamma_003.outputs[0], vector_math_042.inputs[0])
    #vector_math_039.Vector -> vector_math_044.Vector
    HD2_Shader.links.new(gamma_006.outputs[0], vector_math_044.inputs[2])
    #clamp_008.Result -> math_033.Value
    HD2_Shader.links.new(clamp_008.outputs[0], math_033.inputs[0])
    #math_033.Value -> math_032.Value
    HD2_Shader.links.new(math_033.outputs[0], math_032.inputs[1])
    #math_032.Value -> math_089.Value
    HD2_Shader.links.new(math_032.outputs[0], math_089.inputs[0])
    #math_089.Value -> math_090.Value
    HD2_Shader.links.new(math_089.outputs[0], math_090.inputs[1])
    #math_063.Value -> math_090.Value
    HD2_Shader.links.new(math_063.outputs[0], math_090.inputs[0])
    #math_090.Value -> math_091.Value
    HD2_Shader.links.new(math_090.outputs[0], math_091.inputs[1])
    #math_032.Value -> math_092.Value
    HD2_Shader.links.new(math_032.outputs[0], math_092.inputs[0])
    #mix_076.Result -> math_092.Value
    HD2_Shader.links.new(mix_076.outputs[0], math_092.inputs[1])
    #mix_064.Result -> math_093.Value
    HD2_Shader.links.new(mix_064.outputs[0], math_093.inputs[0])
    #combine_xyz_046.Vector -> vector_math_030.Vector
    HD2_Shader.links.new(combine_xyz_046.outputs[0], vector_math_030.inputs[1])
    #combine_xyz_046.Vector -> vector_math_018.Vector
    HD2_Shader.links.new(combine_xyz_046.outputs[0], vector_math_018.inputs[1])
    #vector_math_030.Value -> math_094.Value
    HD2_Shader.links.new(vector_math_030.outputs[1], math_094.inputs[0])
    #mix_048.Result -> math_094.Value
    HD2_Shader.links.new(mix_048.outputs[0], math_094.inputs[1])
    #mix_049.Result -> separate_xyz_055.Vector
    HD2_Shader.links.new(mix_049.outputs[2], separate_xyz_055.inputs[0])
    #separate_xyz_055.Z -> math_095.Value
    HD2_Shader.links.new(separate_xyz_055.outputs[2], math_095.inputs[0])
    #math_003.Value -> math_095.Value
    HD2_Shader.links.new(math_003.outputs[0], math_095.inputs[1])
    #math_095.Value -> math_096.Value
    HD2_Shader.links.new(math_095.outputs[0], math_096.inputs[1])
    #math_096.Value -> math_097.Value
    HD2_Shader.links.new(math_096.outputs[0], math_097.inputs[1])
    #separate_xyz_055.X -> combine_xyz_007.X
    HD2_Shader.links.new(separate_xyz_055.outputs[0], combine_xyz_007.inputs[0])
    #separate_xyz_055.Y -> combine_xyz_007.Y
    HD2_Shader.links.new(separate_xyz_055.outputs[1], combine_xyz_007.inputs[1])
    #combine_xyz_007.Vector -> vector_math_045.Vector
    HD2_Shader.links.new(combine_xyz_007.outputs[0], vector_math_045.inputs[0])
    #vector_math_045.Vector -> separate_xyz_010.Vector
    HD2_Shader.links.new(vector_math_045.outputs[0], separate_xyz_010.inputs[0])
    #combine_xyz_046.Vector -> separate_xyz_011.Vector
    HD2_Shader.links.new(combine_xyz_046.outputs[0], separate_xyz_011.inputs[0])
    #separate_xyz_011.X -> combine_xyz_009.X
    HD2_Shader.links.new(separate_xyz_011.outputs[0], combine_xyz_009.inputs[0])
    #separate_xyz_011.Y -> combine_xyz_009.Y
    HD2_Shader.links.new(separate_xyz_011.outputs[1], combine_xyz_009.inputs[1])
    #math_098.Value -> clamp_009.Value
    HD2_Shader.links.new(math_098.outputs[0], clamp_009.inputs[0])
    #mix_059.Result -> separate_xyz_057.Vector
    HD2_Shader.links.new(mix_059.outputs[2], separate_xyz_057.inputs[0])
    #math_003.Value -> math_099.Value
    HD2_Shader.links.new(math_003.outputs[0], math_099.inputs[0])
    #math_099.Value -> math_100.Value
    HD2_Shader.links.new(math_099.outputs[0], math_100.inputs[1])
    #math_100.Value -> math_101.Value
    HD2_Shader.links.new(math_100.outputs[0], math_101.inputs[1])
    #separate_xyz_057.X -> combine_xyz_011.X
    HD2_Shader.links.new(separate_xyz_057.outputs[0], combine_xyz_011.inputs[0])
    #separate_xyz_057.Y -> combine_xyz_011.Y
    HD2_Shader.links.new(separate_xyz_057.outputs[1], combine_xyz_011.inputs[1])
    #combine_xyz_011.Vector -> vector_math_046.Vector
    HD2_Shader.links.new(combine_xyz_011.outputs[0], vector_math_046.inputs[0])
    #math_101.Value -> vector_math_046.Vector
    HD2_Shader.links.new(math_101.outputs[0], vector_math_046.inputs[1])
    #vector_math_046.Vector -> separate_xyz_013.Vector
    HD2_Shader.links.new(vector_math_046.outputs[0], separate_xyz_013.inputs[0])
    #separate_xyz_013.X -> combine_xyz_012.X
    HD2_Shader.links.new(separate_xyz_013.outputs[0], combine_xyz_012.inputs[0])
    #separate_xyz_013.Y -> combine_xyz_012.Y
    HD2_Shader.links.new(separate_xyz_013.outputs[1], combine_xyz_012.inputs[1])
    #value_009.Value -> combine_xyz_012.Z
    HD2_Shader.links.new(value_009.outputs[0], combine_xyz_012.inputs[2])
    #combine_xyz_012.Vector -> vector_math_047.Vector
    HD2_Shader.links.new(combine_xyz_012.outputs[0], vector_math_047.inputs[0])
    #mix_059.Result -> vector_math_047.Vector
    HD2_Shader.links.new(mix_059.outputs[2], vector_math_047.inputs[1])
    #vector_math_047.Vector -> vector_math_048.Vector
    HD2_Shader.links.new(vector_math_047.outputs[0], vector_math_048.inputs[0])
    #combine_xyz_046.Vector -> vector_math_048.Vector
    HD2_Shader.links.new(combine_xyz_046.outputs[0], vector_math_048.inputs[1])
    #vector_math_048.Value -> math_102.Value
    HD2_Shader.links.new(vector_math_048.outputs[1], math_102.inputs[0])
    #mix_060.Result -> math_102.Value
    HD2_Shader.links.new(mix_060.outputs[0], math_102.inputs[1])
    #math_102.Value -> clamp_010.Value
    HD2_Shader.links.new(math_102.outputs[0], clamp_010.inputs[0])
    #mix_061.Result -> separate_xyz_058.Vector
    HD2_Shader.links.new(mix_061.outputs[2], separate_xyz_058.inputs[0])
    #separate_xyz_058.Z -> math_103.Value
    HD2_Shader.links.new(separate_xyz_058.outputs[2], math_103.inputs[0])
    #math_003.Value -> math_103.Value
    HD2_Shader.links.new(math_003.outputs[0], math_103.inputs[1])
    #math_103.Value -> math_104.Value
    HD2_Shader.links.new(math_103.outputs[0], math_104.inputs[1])
    #math_104.Value -> math_105.Value
    HD2_Shader.links.new(math_104.outputs[0], math_105.inputs[1])
    #separate_xyz_058.X -> combine_xyz_013.X
    HD2_Shader.links.new(separate_xyz_058.outputs[0], combine_xyz_013.inputs[0])
    #separate_xyz_058.Y -> combine_xyz_013.Y
    HD2_Shader.links.new(separate_xyz_058.outputs[1], combine_xyz_013.inputs[1])
    #combine_xyz_013.Vector -> vector_math_049.Vector
    HD2_Shader.links.new(combine_xyz_013.outputs[0], vector_math_049.inputs[0])
    #math_105.Value -> vector_math_049.Vector
    HD2_Shader.links.new(math_105.outputs[0], vector_math_049.inputs[1])
    #vector_math_049.Vector -> separate_xyz_014.Vector
    HD2_Shader.links.new(vector_math_049.outputs[0], separate_xyz_014.inputs[0])
    #separate_xyz_014.X -> combine_xyz_014.X
    HD2_Shader.links.new(separate_xyz_014.outputs[0], combine_xyz_014.inputs[0])
    #separate_xyz_014.Y -> combine_xyz_014.Y
    HD2_Shader.links.new(separate_xyz_014.outputs[1], combine_xyz_014.inputs[1])
    #value_010.Value -> combine_xyz_014.Z
    HD2_Shader.links.new(value_010.outputs[0], combine_xyz_014.inputs[2])
    #combine_xyz_014.Vector -> vector_math_050.Vector
    HD2_Shader.links.new(combine_xyz_014.outputs[0], vector_math_050.inputs[0])
    #mix_061.Result -> vector_math_050.Vector
    HD2_Shader.links.new(mix_061.outputs[2], vector_math_050.inputs[1])
    #vector_math_050.Vector -> vector_math_051.Vector
    HD2_Shader.links.new(vector_math_050.outputs[0], vector_math_051.inputs[0])
    #combine_xyz_046.Vector -> vector_math_051.Vector
    HD2_Shader.links.new(combine_xyz_046.outputs[0], vector_math_051.inputs[1])
    #vector_math_051.Value -> math_106.Value
    HD2_Shader.links.new(vector_math_051.outputs[1], math_106.inputs[0])
    #mix_062.Result -> math_106.Value
    HD2_Shader.links.new(mix_062.outputs[0], math_106.inputs[1])
    #math_106.Value -> clamp_011.Value
    HD2_Shader.links.new(math_106.outputs[0], clamp_011.inputs[0])
    #mix_057.Result -> separate_xyz_015.Vector
    HD2_Shader.links.new(mix_057.outputs[2], separate_xyz_015.inputs[0])
    #clamp_013.Result -> combine_xyz_015.X
    HD2_Shader.links.new(clamp_013.outputs[0], combine_xyz_015.inputs[0])
    #clamp_012.Result -> combine_xyz_015.Y
    HD2_Shader.links.new(clamp_012.outputs[0], combine_xyz_015.inputs[1])
    #mix_058.Result -> clamp_012.Value
    HD2_Shader.links.new(mix_058.outputs[0], clamp_012.inputs[0])
    #separate_xyz_015.Z -> clamp_013.Value
    HD2_Shader.links.new(separate_xyz_015.outputs[2], clamp_013.inputs[0])
    #clamp_009.Result -> math_107.Value
    HD2_Shader.links.new(clamp_009.outputs[0], math_107.inputs[0])
    #math_107.Value -> math_093.Value
    HD2_Shader.links.new(math_107.outputs[0], math_093.inputs[1])
    #math_092.Value -> math_108.Value
    HD2_Shader.links.new(math_092.outputs[0], math_108.inputs[0])
    #math_093.Value -> math_108.Value
    HD2_Shader.links.new(math_093.outputs[0], math_108.inputs[1])
    #clamp_009.Result -> math_108.Value
    HD2_Shader.links.new(clamp_009.outputs[0], math_108.inputs[2])
    #clamp_008.Result -> math_109.Value
    HD2_Shader.links.new(clamp_008.outputs[0], math_109.inputs[0])
    #clamp_010.Result -> math_109.Value
    HD2_Shader.links.new(clamp_010.outputs[0], math_109.inputs[1])
    #math_021.Value -> mix_002.Factor
    HD2_Shader.links.new(math_021.outputs[0], mix_002.inputs[0])
    #value_011.Value -> mix_002.A
    HD2_Shader.links.new(value_011.outputs[0], mix_002.inputs[2])
    #value_007.Value -> mix_003.A
    HD2_Shader.links.new(value_007.outputs[0], mix_003.inputs[2])
    #vector_math_039.Vector -> mix_007.A
    HD2_Shader.links.new(gamma_006.outputs[0], mix_007.inputs[4])
    #vector_math_044.Vector -> mix_007.B
    HD2_Shader.links.new(vector_math_044.outputs[0], mix_007.inputs[5])
    #vector_math_051.Value -> mix_003.B
    HD2_Shader.links.new(vector_math_051.outputs[1], mix_003.inputs[3])
    #mix_003.Result -> mix_002.B
    HD2_Shader.links.new(mix_003.outputs[0], mix_002.inputs[3])
    #vector_math_018.Value -> math_110.Value
    HD2_Shader.links.new(vector_math_018.outputs[1], math_110.inputs[0])
    #pattern_lut_02.Alpha -> math_110.Value
    HD2_Shader.links.new(pattern_lut_02.outputs[1], math_110.inputs[1])
    #pattern_lut.Alpha -> math_111.Value
    HD2_Shader.links.new(pattern_lut.outputs[1], math_111.inputs[0])
    #math_113.Value -> math_112.Value
    HD2_Shader.links.new(math_113.outputs[0], math_112.inputs[1])
    #math_111.Value -> math_112.Value
    HD2_Shader.links.new(math_111.outputs[0], math_112.inputs[0])
    #math_112.Value -> math_114.Value
    HD2_Shader.links.new(math_112.outputs[0], math_114.inputs[0])
    #math_115.Value -> math_116.Value
    HD2_Shader.links.new(math_115.outputs[0], math_116.inputs[1])
    #math_116.Value -> math_117.Value
    HD2_Shader.links.new(math_116.outputs[0], math_117.inputs[0])
    #math_117.Value -> math_118.Value
    HD2_Shader.links.new(math_117.outputs[0], math_118.inputs[1])
    try:
        PatternLUTSizeX = material.node_tree.nodes['Pattern LUT Texture'].inputs[0].node.image.size[0]
        if PatternLUTSizeX >= 0:
            HD2_Shader.links.new(pattern_lut.outputs[0], gamma_pattern.inputs[0])
    except:
        pass
    
    HD2_Shader.links.new(gamma_pattern.outputs[0], vector_math_052.inputs[0])
    #separate_xyz_059.X -> math_119.Value
    HD2_Shader.links.new(separate_xyz_059.outputs[0], math_119.inputs[0])
    #math_119.Value -> math_120.Value
    HD2_Shader.links.new(math_119.outputs[0], math_120.inputs[0])
    #mix_077.Result -> separate_xyz_059.Vector

    #math_120.Value -> math_115.Value
    HD2_Shader.links.new(math_120.outputs[0], math_115.inputs[1])
    #mix_001.Result -> vector_math_053.Vector
    HD2_Shader.links.new(mix_001.outputs[1], vector_math_053.inputs[0])
    #vector_math_053.Vector -> vector_math_052.Vector
    HD2_Shader.links.new(vector_math_053.outputs[0], vector_math_052.inputs[1])
    #math_115.Value -> vector_math_054.Vector
    HD2_Shader.links.new(math_115.outputs[0], vector_math_054.inputs[0])
    #vector_math_052.Vector -> vector_math_054.Vector
    HD2_Shader.links.new(vector_math_052.outputs[0], vector_math_054.inputs[1])
    #mix_001.Result -> vector_math_054.Vector
    HD2_Shader.links.new(mix_001.outputs[1], vector_math_054.inputs[2])
    #mix_031.Result -> math_121.Value
    HD2_Shader.links.new(mix_031.outputs[0], math_121.inputs[0])
    #math_121.Value -> math_122.Value
    HD2_Shader.links.new(math_121.outputs[0], math_122.inputs[1])
    #math_124.Value -> math_123.Value
    HD2_Shader.links.new(math_124.outputs[0], math_123.inputs[1])
    #math_122.Value -> math_124.Value
    HD2_Shader.links.new(math_122.outputs[0], math_124.inputs[0])
    #math_126.Value -> math_125.Value
    HD2_Shader.links.new(math_126.outputs[0], math_125.inputs[0])
    #math_115.Value -> math_126.Value
    HD2_Shader.links.new(math_115.outputs[0], math_126.inputs[0])
    #math_115.Value -> math_125.Value
    HD2_Shader.links.new(math_115.outputs[0], math_125.inputs[1])
    #clamp_005.Result -> math_125.Value
    HD2_Shader.links.new(clamp_005.outputs[0], math_125.inputs[2])
    #math_125.Value -> math_127.Value
    HD2_Shader.links.new(math_125.outputs[0], math_127.inputs[1])
    #math_127.Value -> math_128.Value
    HD2_Shader.links.new(math_127.outputs[0], math_128.inputs[1])
    #math_128.Value -> math_129.Value
    HD2_Shader.links.new(math_128.outputs[0], math_129.inputs[1])
    #math_091.Value -> mix_008.B
    HD2_Shader.links.new(math_091.outputs[0], mix_008.inputs[3])
    #math_063.Value -> mix_008.A
    HD2_Shader.links.new(math_063.outputs[0], mix_008.inputs[2])
    #mix_008.Result -> math_129.Value
    HD2_Shader.links.new(mix_008.outputs[0], math_129.inputs[0])
    #combine_xyz_016.Vector -> pattern_lut_03.Vector
    HD2_Shader.links.new(combine_xyz_016.outputs[0], pattern_lut_03.inputs[0])
    #math_040.Value -> math_130.Value
    HD2_Shader.links.new(math_040.outputs[0], math_130.inputs[0])
    #math_120.Value -> math_130.Value
    HD2_Shader.links.new(math_120.outputs[0], math_130.inputs[1])
    #clamp_005.Result -> math_131.Value
    HD2_Shader.links.new(clamp_005.outputs[0], math_131.inputs[0])
    #math_131.Value -> math_130.Value
    HD2_Shader.links.new(math_131.outputs[0], math_130.inputs[2])
    #math_130.Value -> math_132.Value
    HD2_Shader.links.new(math_130.outputs[0], math_132.inputs[1])
    #math_108.Value -> mix_009.B
    HD2_Shader.links.new(math_108.outputs[0], mix_009.inputs[3])
    #clamp_009.Result -> mix_009.A
    HD2_Shader.links.new(clamp_009.outputs[0], mix_009.inputs[2])
    #mix_009.Result -> math_133.Value
    HD2_Shader.links.new(mix_009.outputs[0], math_133.inputs[0])
    #math_133.Value -> math_134.Value
    HD2_Shader.links.new(math_133.outputs[0], math_134.inputs[1])
    #math_134.Value -> math_135.Value
    HD2_Shader.links.new(math_134.outputs[0], math_135.inputs[1])
    #math_135.Value -> math_136.Value
    HD2_Shader.links.new(math_135.outputs[0], math_136.inputs[1])
    #math_132.Value -> math_136.Value
    HD2_Shader.links.new(math_132.outputs[0], math_136.inputs[0])
    #pattern_lut_03.Alpha -> math_137.Value
    HD2_Shader.links.new(pattern_lut_03.outputs[1], math_137.inputs[0])
    #math_133.Value -> math_137.Value
    HD2_Shader.links.new(math_133.outputs[0], math_137.inputs[1])
    #math_136.Value -> math_138.Value
    HD2_Shader.links.new(math_136.outputs[0], math_138.inputs[0])
    #math_137.Value -> math_138.Value
    HD2_Shader.links.new(math_137.outputs[0], math_138.inputs[1])
    #mix_009.Result -> math_138.Value
    HD2_Shader.links.new(mix_009.outputs[0], math_138.inputs[2])
    #clamp_014.Result -> math_140.Value
    HD2_Shader.links.new(clamp_014.outputs[0], math_140.inputs[0])
    #math_140.Value -> math_139.Value
    HD2_Shader.links.new(math_140.outputs[0], math_139.inputs[0])
    #math_120.Value -> math_139.Value
    HD2_Shader.links.new(math_120.outputs[0], math_139.inputs[1])
    #math_109.Value -> mix_010.B
    HD2_Shader.links.new(math_109.outputs[0], mix_010.inputs[3])
    #math_180.Value -> mix_008.Factor
    HD2_Shader.links.new(math_180.outputs[0], mix_008.inputs[0])
    #math_180.Value -> mix_009.Factor
    HD2_Shader.links.new(math_180.outputs[0], mix_009.inputs[0])
    #math_180.Value -> mix_010.Factor
    HD2_Shader.links.new(math_180.outputs[0], mix_010.inputs[0])
    #clamp_010.Result -> mix_010.A
    HD2_Shader.links.new(clamp_010.outputs[0], mix_010.inputs[2])
    #math_139.Value -> math_141.Value
    HD2_Shader.links.new(math_139.outputs[0], math_141.inputs[1])
    #mix_010.Result -> math_141.Value
    HD2_Shader.links.new(mix_010.outputs[0], math_141.inputs[0])
    #math_094.Value -> math_142.Value
    HD2_Shader.links.new(math_094.outputs[0], math_142.inputs[0])
    #math_142.Value -> math_143.Value
    HD2_Shader.links.new(math_142.outputs[0], math_143.inputs[1])
    #math_115.Value -> math_144.Value
    HD2_Shader.links.new(math_115.outputs[0], math_144.inputs[0])
    #math_143.Value -> math_144.Value
    HD2_Shader.links.new(math_143.outputs[0], math_144.inputs[1])
    #math_094.Value -> math_144.Value
    HD2_Shader.links.new(math_094.outputs[0], math_144.inputs[2])
    #math_140.Value -> math_145.Value
    HD2_Shader.links.new(math_140.outputs[0], math_145.inputs[0])
    #math_120.Value -> math_145.Value
    HD2_Shader.links.new(math_120.outputs[0], math_145.inputs[1])
    #math_145.Value -> math_146.Value
    HD2_Shader.links.new(math_145.outputs[0], math_146.inputs[1])
    #mix_031.Result -> math_147.Value
    HD2_Shader.links.new(mix_031.outputs[0], math_147.inputs[0])
    #math_147.Value -> math_148.Value
    HD2_Shader.links.new(math_147.outputs[0], math_148.inputs[1])
    #mix_051.Result -> separate_xyz_060.Vector
    HD2_Shader.links.new(mix_051.outputs[2], separate_xyz_060.inputs[0])
    #separate_xyz_060.Z -> math_149.Value
    HD2_Shader.links.new(separate_xyz_060.outputs[2], math_149.inputs[1])
    #math_149.Value -> mix_004.Factor
    HD2_Shader.links.new(math_149.outputs[0], mix_004.inputs[0])
    #math_148.Value -> mix_004.B
    HD2_Shader.links.new(math_148.outputs[0], mix_004.inputs[3])
    #mix_004.Result -> math_150.Value
    HD2_Shader.links.new(mix_004.outputs[0], math_150.inputs[0])
    #math_150.Value -> math_151.Value
    HD2_Shader.links.new(math_150.outputs[0], math_151.inputs[1])
    #separate_xyz_024.X -> math_152.Value
    HD2_Shader.links.new(separate_xyz_024.outputs[0], math_152.inputs[1])
    #separate_xyz_060.Y -> math_153.Value
    HD2_Shader.links.new(separate_xyz_060.outputs[1], math_153.inputs[0])
    #math_153.Value -> math_154.Value
    HD2_Shader.links.new(math_153.outputs[0], math_154.inputs[0])
    #math_154.Value -> math_155.Value
    HD2_Shader.links.new(math_154.outputs[0], math_155.inputs[0])
    #vector_math_033.Vector -> vector_math_034.Vector
    HD2_Shader.links.new(vector_math_033.outputs[0], vector_math_034.inputs[0])
    #vector_math_033.Vector -> vector_math_034.Vector
    HD2_Shader.links.new(vector_math_033.outputs[0], vector_math_034.inputs[1])
    #vector_math_033.Vector -> separate_xyz_047.Vector
    HD2_Shader.links.new(vector_math_033.outputs[0], separate_xyz_047.inputs[0])
    #separate_xyz_047.X -> math_087.Value
    HD2_Shader.links.new(separate_xyz_047.outputs[0], math_087.inputs[0])
    #separate_xyz_047.X -> math_156.Value
    HD2_Shader.links.new(separate_xyz_047.outputs[0], math_156.inputs[1])
    #math_087.Value -> math_156.Value
    HD2_Shader.links.new(math_087.outputs[0], math_156.inputs[0])
    #math_156.Value -> math_157.Value
    HD2_Shader.links.new(math_156.outputs[0], math_157.inputs[2])
    #separate_xyz_047.Y -> math_158.Value
    HD2_Shader.links.new(separate_xyz_047.outputs[1], math_158.inputs[0])
    #math_158.Value -> math_157.Value
    HD2_Shader.links.new(math_158.outputs[0], math_157.inputs[0])
    #separate_xyz_047.Y -> math_157.Value
    HD2_Shader.links.new(separate_xyz_047.outputs[1], math_157.inputs[1])
    #vector_math_034.Vector -> separate_xyz_061.Vector
    HD2_Shader.links.new(vector_math_034.outputs[0], separate_xyz_061.inputs[0])
    #separate_xyz_061.X -> math_159.Value
    HD2_Shader.links.new(separate_xyz_061.outputs[0], math_159.inputs[0])
    #separate_xyz_061.Y -> math_159.Value
    HD2_Shader.links.new(separate_xyz_061.outputs[1], math_159.inputs[1])
    #math_159.Value -> math_160.Value
    HD2_Shader.links.new(math_159.outputs[0], math_160.inputs[1])
    #math_160.Value -> math_161.Value
    HD2_Shader.links.new(math_160.outputs[0], math_161.inputs[0])
    #math_157.Value -> math_161.Value
    HD2_Shader.links.new(math_157.outputs[0], math_161.inputs[1])
    #math_161.Value -> math_162.Value
    HD2_Shader.links.new(math_161.outputs[0], math_162.inputs[0])
    #math_162.Value -> math_163.Value
    HD2_Shader.links.new(math_162.outputs[0], math_163.inputs[0])
    #math_163.Value -> vector_math_055.Vector
    HD2_Shader.links.new(math_163.outputs[0], vector_math_055.inputs[0])
    #vector_math_033.Vector -> vector_math_055.Vector
    HD2_Shader.links.new(vector_math_033.outputs[0], vector_math_055.inputs[1])
    #vector_math_055.Vector -> vector_math_056.Vector
    HD2_Shader.links.new(vector_math_055.outputs[0], vector_math_056.inputs[0])
    #mix_052.Result -> vector_math_056.Vector
    HD2_Shader.links.new(mix_052.outputs[0], vector_math_056.inputs[1])
    #vector_math_056.Vector -> separate_xyz_062.Vector
    HD2_Shader.links.new(vector_math_056.outputs[0], separate_xyz_062.inputs[0])
    #separate_xyz_062.Y -> vector_math_057.Vector
    HD2_Shader.links.new(separate_xyz_062.outputs[1], vector_math_057.inputs[0])
    #math_146.Value -> mix_011.B
    HD2_Shader.links.new(math_146.outputs[0], mix_011.inputs[3])
    #separate_xyz_030.X -> mix_011.Factor
    HD2_Shader.links.new(separate_xyz_030.outputs[0], mix_011.inputs[0])
    #math_164.Value -> math_165.Value
    HD2_Shader.links.new(math_164.outputs[0], math_165.inputs[1])
    #combine_xyz_044.Vector -> mix_013.A
    HD2_Shader.links.new(combine_xyz_044.outputs[0], mix_013.inputs[4])
    #vector_math_035.Vector -> mix_013.B
    HD2_Shader.links.new(vector_math_035.outputs[0], mix_013.inputs[5])
    #math_083.Value -> mix_013.Factor
    HD2_Shader.links.new(math_083.outputs[0], mix_013.inputs[0])
    #mix_013.Result -> mix_015.A
    HD2_Shader.links.new(mix_013.outputs[1], mix_015.inputs[4])
    #combine_xyz_027.Vector -> mix_015.B
    HD2_Shader.links.new(combine_xyz_027.outputs[0], mix_015.inputs[5])
    #math_078.Value -> mix_015.Factor
    HD2_Shader.links.new(math_078.outputs[0], mix_015.inputs[0])
    #mix_015.Result -> vector_math_058.Vector
    HD2_Shader.links.new(mix_015.outputs[1], vector_math_058.inputs[0])
    #mix_015.Result -> vector_math_058.Vector
    HD2_Shader.links.new(mix_015.outputs[1], vector_math_058.inputs[1])
    #mix_013.Result -> vector_math_059.Vector
    HD2_Shader.links.new(mix_013.outputs[1], vector_math_059.inputs[0])
    #mix_013.Result -> vector_math_059.Vector
    HD2_Shader.links.new(mix_013.outputs[1], vector_math_059.inputs[1])
    #vector_math_058.Vector -> separate_xyz_064.Vector
    HD2_Shader.links.new(vector_math_058.outputs[0], separate_xyz_064.inputs[0])
    #math_166.Value -> math_167.Value
    HD2_Shader.links.new(math_166.outputs[0], math_167.inputs[0])
    #separate_xyz_064.X -> math_166.Value
    HD2_Shader.links.new(separate_xyz_064.outputs[0], math_166.inputs[0])
    #separate_xyz_064.X -> math_167.Value
    HD2_Shader.links.new(separate_xyz_064.outputs[0], math_167.inputs[1])
    #separate_xyz_064.Y -> math_168.Value
    HD2_Shader.links.new(separate_xyz_064.outputs[1], math_168.inputs[0])
    #math_168.Value -> math_169.Value
    HD2_Shader.links.new(math_168.outputs[0], math_169.inputs[0])
    #separate_xyz_064.Y -> math_169.Value
    HD2_Shader.links.new(separate_xyz_064.outputs[1], math_169.inputs[1])
    #math_167.Value -> math_169.Value
    HD2_Shader.links.new(math_167.outputs[0], math_169.inputs[2])
    #vector_math_059.Vector -> separate_xyz_065.Vector
    HD2_Shader.links.new(vector_math_059.outputs[0], separate_xyz_065.inputs[0])
    #separate_xyz_065.X -> math_170.Value
    HD2_Shader.links.new(separate_xyz_065.outputs[0], math_170.inputs[0])
    #separate_xyz_065.Y -> math_170.Value
    HD2_Shader.links.new(separate_xyz_065.outputs[1], math_170.inputs[1])
    #math_170.Value -> math_171.Value
    HD2_Shader.links.new(math_170.outputs[0], math_171.inputs[1])
    #math_171.Value -> math_172.Value
    HD2_Shader.links.new(math_171.outputs[0], math_172.inputs[1])
    #math_169.Value -> math_172.Value
    HD2_Shader.links.new(math_169.outputs[0], math_172.inputs[0])
    #math_172.Value -> math_173.Value
    HD2_Shader.links.new(math_172.outputs[0], math_173.inputs[0])
    #math_173.Value -> math_174.Value
    HD2_Shader.links.new(math_173.outputs[0], math_174.inputs[0])
    #math_174.Value -> vector_math_060.Vector
    HD2_Shader.links.new(math_174.outputs[0], vector_math_060.inputs[0])
    #vector_math_058.Vector -> vector_math_060.Vector
    HD2_Shader.links.new(vector_math_058.outputs[0], vector_math_060.inputs[1])
    #vector_math_060.Vector -> vector_math_061.Vector
    HD2_Shader.links.new(vector_math_060.outputs[0], vector_math_061.inputs[0])
    #vector_math_061.Vector -> vector_math_062.Vector
    HD2_Shader.links.new(vector_math_061.outputs[0], vector_math_062.inputs[0])
    #math_007.Value -> vector_math_062.Vector
    HD2_Shader.links.new(math_007.outputs[0], vector_math_062.inputs[1])
    #math_035.Value -> math_175.Value
    HD2_Shader.links.new(math_035.outputs[0], math_175.inputs[0])
    #math_037.Value -> math_175.Value
    HD2_Shader.links.new(math_037.outputs[0], math_175.inputs[1])
    #math_175.Value -> math_176.Value
    HD2_Shader.links.new(math_175.outputs[0], math_176.inputs[0])
    #math_118.Value -> math_176.Value
    HD2_Shader.links.new(math_118.outputs[0], math_176.inputs[1])
    #vector_math_054.Vector -> mix_016.B
    HD2_Shader.links.new(vector_math_054.outputs[0], mix_016.inputs[5])
    #math_176.Value -> mix_016.Factor
    try:
        PatternLUTSizeX = material.node_tree.nodes['Pattern LUT Texture'].inputs[0].node.image.size[0]
        if PatternLUTSizeX >= 0:
            HD2_Shader.links.new(math_176.outputs[0], mix_016.inputs[0])
    except:
        pass
    #mix_001.Result -> mix_016.A
    HD2_Shader.links.new(mix_001.outputs[1], mix_016.inputs[4])
    #vector_math_063.Vector -> vector_math_065.Vector
    HD2_Shader.links.new(vector_math_063.outputs[0], vector_math_065.inputs[1])
    #clamp_007.Result -> vector_math_065.Vector
    HD2_Shader.links.new(clamp_007.outputs[0], vector_math_065.inputs[0])
    #clamp_007.Result -> math_178.Value
    HD2_Shader.links.new(clamp_007.outputs[0], math_178.inputs[0])
    #separate_xyz_024.Z -> math_178.Value
    HD2_Shader.links.new(separate_xyz_024.outputs[2], math_178.inputs[1])
    #separate_xyz_024.Y -> math_179.Value
    HD2_Shader.links.new(separate_xyz_024.outputs[1], math_179.inputs[0])
    #math_024.Value -> math_180.Value
    HD2_Shader.links.new(math_024.outputs[0], math_180.inputs[1])
    #math_021.Value -> math_180.Value
    HD2_Shader.links.new(math_021.outputs[0], math_180.inputs[0])
    #math_178.Value -> math_181.Value
    HD2_Shader.links.new(math_178.outputs[0], math_181.inputs[0])
    #math_179.Value -> math_181.Value
    HD2_Shader.links.new(math_179.outputs[0], math_181.inputs[1])
    #mix_002.Result -> math_182.Value
    HD2_Shader.links.new(mix_002.outputs[0], math_182.inputs[0])
    #clamp_005.Result -> math_182.Value
    HD2_Shader.links.new(clamp_005.outputs[0], math_182.inputs[1])
    #gamma_005.Color -> vector_math_066.Vector
    HD2_Shader.links.new(gamma_005.outputs[0], vector_math_066.inputs[0])
    #vector_math_064.Vector -> vector_math_066.Vector
    HD2_Shader.links.new(vector_math_064.outputs[0], vector_math_066.inputs[1])
    #math_182.Value -> vector_math_067.Vector
    HD2_Shader.links.new(math_182.outputs[0], vector_math_067.inputs[0])
    #vector_math_066.Vector -> vector_math_067.Vector
    HD2_Shader.links.new(vector_math_066.outputs[0], vector_math_067.inputs[1])
    #vector_math_065.Vector -> vector_math_067.Vector
    HD2_Shader.links.new(vector_math_065.outputs[0], vector_math_067.inputs[2])
    #clamp_005.Result -> math_183.Value
    HD2_Shader.links.new(clamp_005.outputs[0], math_183.inputs[0])
    #clamp_005.Result -> math_183.Value
    HD2_Shader.links.new(clamp_005.outputs[0], math_183.inputs[1])
    #math_183.Value -> math_184.Value
    HD2_Shader.links.new(math_183.outputs[0], math_184.inputs[0])
    #mix_002.Result -> math_184.Value
    HD2_Shader.links.new(mix_002.outputs[0], math_184.inputs[1])
    #gamma_002.Color -> vector_math_068.Vector
    HD2_Shader.links.new(gamma_002.outputs[0], vector_math_068.inputs[0])
    #math_184.Value -> vector_math_069.Vector
    HD2_Shader.links.new(math_184.outputs[0], vector_math_069.inputs[0])
    #vector_math_068.Vector -> vector_math_069.Vector
    HD2_Shader.links.new(vector_math_068.outputs[0], vector_math_069.inputs[1])
    #vector_math_067.Vector -> vector_math_069.Vector
    HD2_Shader.links.new(vector_math_067.outputs[0], vector_math_069.inputs[2])
    #vector_math_067.Vector -> vector_math_070.Vector
    HD2_Shader.links.new(vector_math_067.outputs[0], vector_math_070.inputs[0])
    #vector_math_070.Vector -> vector_math_068.Vector
    HD2_Shader.links.new(vector_math_070.outputs[0], vector_math_068.inputs[1])
    #vector_math_069.Vector -> vector_math_071.Vector
    HD2_Shader.links.new(vector_math_069.outputs[0], vector_math_071.inputs[1])
    #vector_math_071.Vector -> separate_xyz_067.Vector
    HD2_Shader.links.new(vector_math_071.outputs[0], separate_xyz_067.inputs[0])
    #separate_xyz_067.X -> math_185.Value
    HD2_Shader.links.new(separate_xyz_067.outputs[0], math_185.inputs[1])
    #separate_xyz_067.Y -> math_186.Value
    HD2_Shader.links.new(separate_xyz_067.outputs[1], math_186.inputs[1])
    #separate_xyz_067.Z -> math_187.Value
    HD2_Shader.links.new(separate_xyz_067.outputs[2], math_187.inputs[1])
    #math_185.Value -> combine_xyz_042.X
    HD2_Shader.links.new(math_185.outputs[0], combine_xyz_042.inputs[0])
    #math_186.Value -> combine_xyz_042.Y
    HD2_Shader.links.new(math_186.outputs[0], combine_xyz_042.inputs[1])
    #math_187.Value -> combine_xyz_042.Z
    HD2_Shader.links.new(math_187.outputs[0], combine_xyz_042.inputs[2])
    #vector_math_071.Vector -> vector_math_072.Vector
    HD2_Shader.links.new(vector_math_071.outputs[0], vector_math_072.inputs[0])
    #vector_math_072.Vector -> separate_xyz_068.Vector
    HD2_Shader.links.new(vector_math_072.outputs[0], separate_xyz_068.inputs[0])
    #separate_xyz_068.X -> math_188.Value
    HD2_Shader.links.new(separate_xyz_068.outputs[0], math_188.inputs[0])
    #math_188.Value -> combine_xyz_048.X
    HD2_Shader.links.new(math_188.outputs[0], combine_xyz_048.inputs[0])
    #separate_xyz_068.Y -> math_189.Value
    HD2_Shader.links.new(separate_xyz_068.outputs[1], math_189.inputs[0])
    #separate_xyz_068.Z -> math_190.Value
    HD2_Shader.links.new(separate_xyz_068.outputs[2], math_190.inputs[0])
    #math_189.Value -> combine_xyz_048.Y
    HD2_Shader.links.new(math_189.outputs[0], combine_xyz_048.inputs[1])
    #math_190.Value -> combine_xyz_048.Z
    HD2_Shader.links.new(math_190.outputs[0], combine_xyz_048.inputs[2])
    #math_003.Value -> math_191.Value
    HD2_Shader.links.new(math_003.outputs[0], math_191.inputs[0])
    #math_054.Value -> math_055.Value
    HD2_Shader.links.new(math_054.outputs[0], math_055.inputs[0])
    #math_181.Value -> math_050.Value
    HD2_Shader.links.new(math_181.outputs[0], math_050.inputs[0])
    #math_097.Value -> vector_math_045.Vector
    HD2_Shader.links.new(math_097.outputs[0], vector_math_045.inputs[1])
    #math_110.Value -> clamp_014.Value
    HD2_Shader.links.new(math_110.outputs[0], clamp_014.inputs[0])
    #clamp_014.Result -> math_115.Value
    HD2_Shader.links.new(clamp_014.outputs[0], math_115.inputs[0])
    #math_024.Value -> mix_003.Factor
    HD2_Shader.links.new(math_024.outputs[0], mix_003.inputs[0])
    #mix_034.Result -> math_067.Value
    HD2_Shader.links.new(mix_034.outputs[0], math_067.inputs[1])
    #value_003.Value -> mix_012.B
    HD2_Shader.links.new(value_003.outputs[0], mix_012.inputs[3])
    #math_078.Value -> mix_012.Factor
    HD2_Shader.links.new(math_078.outputs[0], mix_012.inputs[0])
    #mix_005.Result -> separate_xyz_009.Vector
    HD2_Shader.links.new(mix_005.outputs[1], separate_xyz_009.inputs[0])
    #separate_xyz_009.Y -> mix_012.A
    HD2_Shader.links.new(separate_xyz_009.outputs[1], mix_012.inputs[2])
    #mix_012.Result -> math_071.Value
    HD2_Shader.links.new(mix_012.outputs[0], math_071.inputs[1])
    #uv_map_002.UV -> vector_math_073.Vector
    HD2_Shader.links.new(uv_map_002.outputs[0], vector_math_073.inputs[0])
    #vector_math_073.Vector -> separate_xyz.Vector
    HD2_Shader.links.new(vector_math_073.outputs[0], separate_xyz.inputs[0])
    #math_1.Value -> combine_xyz.X
    HD2_Shader.links.new(math_1.outputs[0], combine_xyz.inputs[0])
    #math_001.Value -> combine_xyz.Y
    HD2_Shader.links.new(math_001.outputs[0], combine_xyz.inputs[1])
    #separate_xyz.X -> math_1.Value
    HD2_Shader.links.new(separate_xyz.outputs[0], math_1.inputs[0])
    #separate_xyz.Y -> math_001.Value
    HD2_Shader.links.new(separate_xyz.outputs[1], math_001.inputs[0])
    #combine_xyz.Vector -> separate_xyz_003.Vector
    HD2_Shader.links.new(combine_xyz.outputs[0], separate_xyz_003.inputs[0])
    #separate_xyz_003.X -> math_022.Value
    HD2_Shader.links.new(separate_xyz_003.outputs[0], math_022.inputs[0])
    #math_022.Value -> math_192.Value
    HD2_Shader.links.new(math_022.outputs[0], math_192.inputs[1])
    #clamp_005.Result -> math_193.Value
    HD2_Shader.links.new(clamp_005.outputs[0], math_193.inputs[0])
    #math_193.Value -> math_195.Value
    HD2_Shader.links.new(math_193.outputs[0], math_195.inputs[0])
    #mix_016.Result -> vector_math_074.Vector
    HD2_Shader.links.new(mix_016.outputs[1], vector_math_074.inputs[0])
    #vector_math_074.Vector -> vector_math_075.Vector
    HD2_Shader.links.new(vector_math_074.outputs[0], vector_math_075.inputs[0])
    #math_195.Value -> clamp_015.Value
    HD2_Shader.links.new(math_195.outputs[0], clamp_015.inputs[0])
    #clamp_015.Result -> vector_math_076.Vector
    HD2_Shader.links.new(clamp_015.outputs[0], vector_math_076.inputs[0])
    #vector_math_075.Vector -> vector_math_076.Vector
    HD2_Shader.links.new(vector_math_075.outputs[0], vector_math_076.inputs[1])
    #mix_016.Result -> vector_math_076.Vector
    HD2_Shader.links.new(mix_016.outputs[1], vector_math_076.inputs[2])
    #clamp_015.Result -> math_196.Value
    HD2_Shader.links.new(clamp_015.outputs[0], math_196.inputs[0])
    #math_196.Value -> math_197.Value
    HD2_Shader.links.new(math_196.outputs[0], math_197.inputs[0])
    #math_197.Value -> math_198.Value
    HD2_Shader.links.new(math_197.outputs[0], math_198.inputs[0])
    #combine_xyz_015.Vector -> separate_xyz_053.Vector
    HD2_Shader.links.new(combine_xyz_015.outputs[0], separate_xyz_053.inputs[0])
    #separate_xyz_053.X -> math_198.Value
    HD2_Shader.links.new(separate_xyz_053.outputs[0], math_198.inputs[1])
    #separate_xyz_053.Y -> math_200.Value
    HD2_Shader.links.new(separate_xyz_053.outputs[1], math_200.inputs[0])
    #clamp_015.Result -> math_199.Value
    HD2_Shader.links.new(clamp_015.outputs[0], math_199.inputs[0])
    #separate_xyz_053.Y -> math_199.Value
    HD2_Shader.links.new(separate_xyz_053.outputs[1], math_199.inputs[2])
    #math_200.Value -> math_199.Value
    HD2_Shader.links.new(math_200.outputs[0], math_199.inputs[1])
    #clamp_015.Result -> math_201.Value
    HD2_Shader.links.new(clamp_015.outputs[0], math_201.inputs[1])
    #math_202.Value -> math_201.Value
    HD2_Shader.links.new(math_202.outputs[0], math_201.inputs[0])
    #clamp_007.Result -> math_202.Value
    HD2_Shader.links.new(clamp_007.outputs[0], math_202.inputs[0])
    #math_201.Value -> math_203.Value
    HD2_Shader.links.new(math_201.outputs[0], math_203.inputs[1])
    #math_138.Value -> mix_014.B
    HD2_Shader.links.new(math_138.outputs[0], mix_014.inputs[3])
    #mix_009.Result -> mix_014.A
    HD2_Shader.links.new(mix_009.outputs[0], mix_014.inputs[2])
    #mix_014.Result -> math_206.Value
    HD2_Shader.links.new(mix_014.outputs[0], math_206.inputs[0])
    #math_206.Value -> math_207.Value
    HD2_Shader.links.new(math_206.outputs[0], math_207.inputs[0])
    #math_203.Value -> math_208.Value
    HD2_Shader.links.new(math_203.outputs[0], math_208.inputs[0])
    #math_207.Value -> math_208.Value
    HD2_Shader.links.new(math_207.outputs[0], math_208.inputs[1])
    #mix_014.Result -> math_208.Value
    HD2_Shader.links.new(mix_014.outputs[0], math_208.inputs[2])
    #math_209.Value -> math_179.Value
    HD2_Shader.links.new(math_209.outputs[0], math_179.inputs[1])
    #clamp_008.Result -> vector_math_044.Vector
    HD2_Shader.links.new(clamp_008.outputs[0], vector_math_044.inputs[0])
    #vector_math_042.Vector -> vector_math_044.Vector
    HD2_Shader.links.new(vector_math_042.outputs[0], vector_math_044.inputs[1])
    #vector_math_077.Vector -> vector_math_063.Vector
    HD2_Shader.links.new(vector_math_077.outputs[0], vector_math_063.inputs[1])
    #vector_math_076.Vector -> mix_1.B
    HD2_Shader.links.new(vector_math_076.outputs[0], mix_1.inputs[5])
    #math_192.Value -> mix_1.Factor
    HD2_Shader.links.new(math_192.outputs[0], mix_1.inputs[0])
    #mix_016.Result -> mix_1.A
    HD2_Shader.links.new(mix_016.outputs[1], mix_1.inputs[4])
    #mix_1.Result -> vector_math_077.Vector
    HD2_Shader.links.new(mix_1.outputs[1], vector_math_077.inputs[0])
    #mix_1.Result -> vector_math_065.Vector
    HD2_Shader.links.new(mix_1.outputs[1], vector_math_065.inputs[2])
    #vector_math_065.Vector -> vector_math_064.Vector
    HD2_Shader.links.new(vector_math_065.outputs[0], vector_math_064.inputs[0])
    #separate_xyz_057.Z -> math_099.Value
    HD2_Shader.links.new(separate_xyz_057.outputs[2], math_099.inputs[1])
    #mix_007.Result -> mix_001.B
    HD2_Shader.links.new(mix_007.outputs[1], mix_001.inputs[5])
    #math_024.Value -> mix_007.Factor
    HD2_Shader.links.new(math_024.outputs[0], mix_007.inputs[0])
    #math_021.Value -> mix_001.Factor
    HD2_Shader.links.new(math_021.outputs[0], mix_001.inputs[0])
    #gamma_003.Color -> mix_001.A
    HD2_Shader.links.new(gamma_003.outputs[0], mix_001.inputs[4])
    #gamma_004.Color -> vector_math_063.Vector
    HD2_Shader.links.new(gamma_004.outputs[0], vector_math_063.inputs[0])
    #math_208.Value -> mix_017.B
    HD2_Shader.links.new(math_208.outputs[0], mix_017.inputs[3])
    #math_192.Value -> mix_017.Factor
    HD2_Shader.links.new(math_192.outputs[0], mix_017.inputs[0])
    #mix_014.Result -> mix_017.A
    HD2_Shader.links.new(mix_014.outputs[0], mix_017.inputs[2])
    #mix_017.Result -> math_209.Value
    HD2_Shader.links.new(mix_017.outputs[0], math_209.inputs[0])
    #mix_017.Result -> math_181.Value
    HD2_Shader.links.new(mix_017.outputs[0], math_181.inputs[2])
    #math_043.Value -> vector_math_019.Vector
    HD2_Shader.links.new(math_043.outputs[0], vector_math_019.inputs[1])
    #mix_005.Result -> separate_xyz_069.Vector
    HD2_Shader.links.new(mix_005.outputs[1], separate_xyz_069.inputs[0])
    #math_210.Value -> math_211.Value
    HD2_Shader.links.new(math_210.outputs[0], math_211.inputs[1])
    #math_245.Value -> math_210.Value
    HD2_Shader.links.new(math_245.outputs[0], math_210.inputs[0])
    #math_244.Value -> math_164.Value
    HD2_Shader.links.new(math_244.outputs[0], math_164.inputs[0])
    #mix_057.Result -> separate_xyz_019.Vector
    HD2_Shader.links.new(mix_057.outputs[2], separate_xyz_019.inputs[0])
    #clamp_017.Result -> combine_xyz_008.X
    HD2_Shader.links.new(clamp_017.outputs[0], combine_xyz_008.inputs[0])
    #clamp_016.Result -> combine_xyz_008.Y
    HD2_Shader.links.new(clamp_016.outputs[0], combine_xyz_008.inputs[1])
    #separate_xyz_019.X -> clamp_016.Value
    HD2_Shader.links.new(separate_xyz_019.outputs[0], clamp_016.inputs[0])
    #separate_xyz_019.Y -> clamp_017.Value
    HD2_Shader.links.new(separate_xyz_019.outputs[1], clamp_017.inputs[0])
    #clamp_007.Result -> math_212.Value
    HD2_Shader.links.new(clamp_007.outputs[0], math_212.inputs[0])
    #math_212.Value -> math_213.Value
    HD2_Shader.links.new(math_212.outputs[0], math_213.inputs[1])
    #combine_xyz_008.Vector -> separate_xyz_021.Vector
    HD2_Shader.links.new(combine_xyz_008.outputs[0], separate_xyz_021.inputs[0])
    #separate_xyz_021.X -> math_214.Value
    HD2_Shader.links.new(separate_xyz_021.outputs[0], math_214.inputs[0])
    #math_213.Value -> math_214.Value
    HD2_Shader.links.new(math_213.outputs[0], math_214.inputs[1])
    #mix_1.Result -> separate_xyz_026.Vector
    HD2_Shader.links.new(mix_1.outputs[1], separate_xyz_026.inputs[0])
    #separate_xyz_026.X -> math_215.Value
    HD2_Shader.links.new(separate_xyz_026.outputs[0], math_215.inputs[0])
    #separate_xyz_026.Y -> math_216.Value
    HD2_Shader.links.new(separate_xyz_026.outputs[1], math_216.inputs[0])
    #separate_xyz_026.Z -> math_217.Value
    HD2_Shader.links.new(separate_xyz_026.outputs[2], math_217.inputs[0])
    #math_215.Value -> combine_xyz_021.X
    HD2_Shader.links.new(math_215.outputs[0], combine_xyz_021.inputs[0])
    #math_216.Value -> combine_xyz_021.Y
    HD2_Shader.links.new(math_216.outputs[0], combine_xyz_021.inputs[1])
    #math_217.Value -> combine_xyz_021.Z
    HD2_Shader.links.new(math_217.outputs[0], combine_xyz_021.inputs[2])
    #mix_055.Result -> vector_math_079.Vector
    HD2_Shader.links.new(mix_055.outputs[2], vector_math_079.inputs[0])
    #combine_xyz_021.Vector -> vector_math_080.Vector
    HD2_Shader.links.new(combine_xyz_021.outputs[0], vector_math_080.inputs[0])
    #vector_math_079.Vector -> vector_math_080.Vector
    HD2_Shader.links.new(vector_math_079.outputs[0], vector_math_080.inputs[1])
    #vector_math_080.Vector -> vector_math_081.Vector
    HD2_Shader.links.new(vector_math_080.outputs[0], vector_math_081.inputs[1])
    #separate_xyz_021.Y -> vector_math_081.Vector
    HD2_Shader.links.new(separate_xyz_021.outputs[1], vector_math_081.inputs[0])
    #mix_055.Result -> vector_math_081.Vector
    HD2_Shader.links.new(mix_055.outputs[2], vector_math_081.inputs[2])
    #math_214.Value -> math_218.Value
    HD2_Shader.links.new(math_214.outputs[0], math_218.inputs[0])
    #math_218.Value -> math_219.Value
    HD2_Shader.links.new(math_218.outputs[0], math_219.inputs[0])
    #mix_014.Result -> math_219.Value
    HD2_Shader.links.new(mix_014.outputs[0], math_219.inputs[1])
    #mix_1.Result -> vector_math_082.Vector
    HD2_Shader.links.new(mix_1.outputs[1], vector_math_082.inputs[0])
    #math_219.Value -> math_221.Value
    HD2_Shader.links.new(math_219.outputs[0], math_221.inputs[0])
    #math_221.Value -> math_220.Value
    HD2_Shader.links.new(math_221.outputs[0], math_220.inputs[1])
    #math_220.Value -> math_222.Value
    HD2_Shader.links.new(math_220.outputs[0], math_222.inputs[1])
    #vector_math_082.Value -> math_222.Value
    HD2_Shader.links.new(vector_math_082.outputs[1], math_222.inputs[0])
    #math_219.Value -> math_222.Value
    HD2_Shader.links.new(math_219.outputs[0], math_222.inputs[2])
    #math_129.Value -> mix_006.B
    HD2_Shader.links.new(math_129.outputs[0], mix_006.inputs[3])
    #math_176.Value -> math_177.Value
    HD2_Shader.links.new(math_176.outputs[0], math_177.inputs[0])
    #math_123.Value -> math_177.Value
    HD2_Shader.links.new(math_123.outputs[0], math_177.inputs[1])
    #math_177.Value -> mix_006.Factor
    HD2_Shader.links.new(math_177.outputs[0], mix_006.inputs[0])
    #mix_008.Result -> mix_006.A
    HD2_Shader.links.new(mix_008.outputs[0], mix_006.inputs[2])
    #math_223.Value -> math_224.Value
    HD2_Shader.links.new(math_223.outputs[0], math_224.inputs[0])
    #mix_006.Result -> clamp_018.Value
    HD2_Shader.links.new(mix_006.outputs[0], clamp_018.inputs[0])
    #clamp_018.Result -> math_223.Value
    HD2_Shader.links.new(clamp_018.outputs[0], math_223.inputs[1])
    #clamp_018.Result -> principled_bsdf_001.Metallic
    HD2_Shader.links.new(clamp_018.outputs[0], principled_bsdf_001.inputs[1])
    #mix_033.Result -> separate_xyz_036.Vector
    HD2_Shader.links.new(mix_033.outputs[2], separate_xyz_036.inputs[0])
    #normal_map_001.Normal -> principled_bsdf_001.Normal
    HD2_Shader.links.new(normal_map_001.outputs[0], principled_bsdf_001.inputs[5])
    #normal_map.Normal -> principled_bsdf_001.Coat Normal
    HD2_Shader.links.new(normal_map.outputs[0], principled_bsdf_001.inputs[22])
    #math_181.Value -> principled_bsdf_001.Coat Roughness
    HD2_Shader.links.new(math_181.outputs[0], principled_bsdf_001.inputs[19])
    #math_181.Value -> principled_bsdf_001.Roughness
    HD2_Shader.links.new(math_181.outputs[0], principled_bsdf_001.inputs[2])
    #math_224.Value -> principled_bsdf_001.Coat Weight
    HD2_Shader.links.new(math_224.outputs[0], principled_bsdf_001.inputs[18])
    #math_223.Value -> math_225.Value
    HD2_Shader.links.new(math_223.outputs[0], math_225.inputs[0])
    #math_225.Value -> principled_bsdf_001.Specular IOR Level
    HD2_Shader.links.new(math_225.outputs[0], principled_bsdf_001.inputs[12])
    #vector_math_021.Vector -> vector_math_089.Vector
    HD2_Shader.links.new(vector_math_021.outputs[0], vector_math_089.inputs[1])
    #alpha_cutoff_node -> Transparency mix factor
    HD2_Shader.links.new(alpha_cutoff_node.outputs[0], mix_shader_transparency.inputs[0])
    #transparency_shader.Shader -> Transparency mix
    HD2_Shader.links.new(transparency_shader.outputs[0], mix_shader_transparency.inputs[1])
    #mix_shader.Shader -> Transparency mix
    HD2_Shader.links.new(mix_shader.outputs[0], mix_shader_transparency.inputs[2])
    #mix_shader_transparency.Shader -> group_output_1.Output
    HD2_Shader.links.new(mix_shader_transparency.outputs[0], group_output_1.inputs[0])
    #vector_math_089.Vector -> vector_math_091.Vector
    HD2_Shader.links.new(vector_math_089.outputs[0], vector_math_091.inputs[0])
    #vector_math_115.Vector -> vector_math_096.Vector
    HD2_Shader.links.new(vector_math_115.outputs[0], vector_math_096.inputs[0])
    #math_165.Value -> vector_math_096.Vector
    HD2_Shader.links.new(math_165.outputs[0], vector_math_096.inputs[1])
    #separate_xyz_036.X -> math_084.Value
    HD2_Shader.links.new(separate_xyz_036.outputs[0], math_084.inputs[0])
    #separate_xyz_036.X -> math_046.Value
    HD2_Shader.links.new(separate_xyz_036.outputs[0], math_046.inputs[0])
    #separate_xyz_009.X -> combine_xyz_046.X
    HD2_Shader.links.new(separate_xyz_009.outputs[0], combine_xyz_046.inputs[0])
    #separate_xyz_009.Y -> combine_xyz_046.Y
    HD2_Shader.links.new(separate_xyz_009.outputs[1], combine_xyz_046.inputs[1])
    #vector_math_099.Vector -> vector_math_097.Vector
    HD2_Shader.links.new(vector_math_099.outputs[0], vector_math_097.inputs[0])
    #vector_math_096.Vector -> vector_math_098.Vector
    HD2_Shader.links.new(vector_math_096.outputs[0], vector_math_098.inputs[0])
    #vector_math_021.Vector -> vector_math_098.Vector
    HD2_Shader.links.new(vector_math_021.outputs[0], vector_math_098.inputs[1])
    #vector_math_098.Vector -> vector_math_099.Vector
    HD2_Shader.links.new(vector_math_098.outputs[0], vector_math_099.inputs[0])
    #vector_math_007.Vector -> vector_math_099.Vector
    HD2_Shader.links.new(vector_math_116.outputs[0], vector_math_099.inputs[1])

    #Slot3.Result -> Slot4.A
    HD2_Shader.links.new(Slot3.outputs[0], Slot4.inputs[2])
    #Slot1and2.Result -> Slot3.A
    HD2_Shader.links.new(Slot1and2.outputs[0], Slot3.inputs[2])
    #Slot4.Result -> Slot5.A
    HD2_Shader.links.new(Slot4.outputs[0], Slot5.inputs[2])
    #Slot6.Result -> Slot7.A
    HD2_Shader.links.new(Slot6.outputs[0], Slot7.inputs[2])
    #Slot5.Result -> Slot6.A
    HD2_Shader.links.new(Slot5.outputs[0], Slot6.inputs[2])
    #Slot7.Result -> Slot8.A
    HD2_Shader.links.new(Slot7.outputs[0], Slot8.inputs[2])

    #mix_018.Result -> separate_xyz_043.Vector
    HD2_Shader.links.new(mix_018.outputs[2], separate_xyz_043.inputs[0])
    #combine_xyz_052.Vector -> primary_material_lut_00.Vector
    HD2_Shader.links.new(combine_xyz_052.outputs[0], primary_material_lut_00.inputs[0])
    #primary_material_lut_00.Color -> mix_032.A
    HD2_Shader.links.new(primary_material_lut_00.outputs[0], mix_032.inputs[6])
    #primary_material_lut_00.Alpha -> mix_031.A
    HD2_Shader.links.new(primary_material_lut_00.outputs[1], mix_031.inputs[2])
    #math_243.Value -> mix_032.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_032.inputs[0])
    #Slot8.Result -> combine_xyz_052.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_052.inputs[1])
    #math_243.Value -> mix_031.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_031.inputs[0])
    #Slot8.Result -> combine_xyz_053.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_053.inputs[1])
    #mix_032.Result -> gamma_003.Color
    HD2_Shader.links.new(mix_032.outputs[2], gamma_003.inputs[0])
    #combine_xyz_054.Vector -> primary_material_lut_01.Vector
    HD2_Shader.links.new(combine_xyz_054.outputs[0], primary_material_lut_01.inputs[0])
    #primary_material_lut_01.Color -> mix_033.A
    HD2_Shader.links.new(primary_material_lut_01.outputs[0], mix_033.inputs[6])
    #primary_material_lut_01.Alpha -> mix_034.A
    HD2_Shader.links.new(primary_material_lut_01.outputs[1], mix_034.inputs[2])
    #math_243.Value -> mix_033.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_033.inputs[0])
    #Slot8.Result -> combine_xyz_054.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_054.inputs[1])
    #math_243.Value -> mix_034.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_034.inputs[0])
    #Slot8.Result -> combine_xyz_055.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_055.inputs[1])
    #combine_xyz_056.Vector -> primary_material_lut_02.Vector
    HD2_Shader.links.new(combine_xyz_056.outputs[0], primary_material_lut_02.inputs[0])
    #primary_material_lut_02.Color -> mix_035.A
    HD2_Shader.links.new(primary_material_lut_02.outputs[0], mix_035.inputs[6])
    #primary_material_lut_02.Alpha -> mix_036.A
    HD2_Shader.links.new(primary_material_lut_02.outputs[1], mix_036.inputs[2])
    #math_243.Value -> mix_035.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_035.inputs[0])
    #Slot8.Result -> combine_xyz_056.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_056.inputs[1])
    #math_243.Value -> mix_036.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_036.inputs[0])
    #Slot8.Result -> combine_xyz_057.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_057.inputs[1])
    #mix_035.Result -> gamma_004.Color
    HD2_Shader.links.new(mix_035.outputs[2], gamma_004.inputs[0])
    #combine_xyz_058.Vector -> primary_material_lut_03.Vector
    HD2_Shader.links.new(combine_xyz_058.outputs[0], primary_material_lut_03.inputs[0])
    #primary_material_lut_03.Color -> mix_037.A
    HD2_Shader.links.new(primary_material_lut_03.outputs[0], mix_037.inputs[6])
    #primary_material_lut_03.Alpha -> mix_038.A
    HD2_Shader.links.new(primary_material_lut_03.outputs[1], mix_038.inputs[2])
    #math_243.Value -> mix_037.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_037.inputs[0])
    #Slot8.Result -> combine_xyz_058.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_058.inputs[1])
    #math_243.Value -> mix_038.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_038.inputs[0])
    #Slot8.Result -> combine_xyz_059.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_059.inputs[1])
    #combine_xyz_060.Vector -> primary_material_lut_04.Vector
    HD2_Shader.links.new(combine_xyz_060.outputs[0], primary_material_lut_04.inputs[0])
    #primary_material_lut_04.Color -> mix_039.A
    HD2_Shader.links.new(primary_material_lut_04.outputs[0], mix_039.inputs[6])
    #primary_material_lut_04.Alpha -> mix_040.A
    HD2_Shader.links.new(primary_material_lut_04.outputs[1], mix_040.inputs[2])
    #math_243.Value -> mix_039.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_039.inputs[0])
    #Slot8.Result -> combine_xyz_060.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_060.inputs[1])
    #math_243.Value -> mix_040.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_040.inputs[0])
    #Slot8.Result -> combine_xyz_061.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_061.inputs[1])
    #mix_039.Result -> separate_xyz_018.Vector
    HD2_Shader.links.new(mix_039.outputs[2], separate_xyz_018.inputs[0])
    #combine_xyz_062.Vector -> primary_material_lut_05.Vector
    HD2_Shader.links.new(combine_xyz_062.outputs[0], primary_material_lut_05.inputs[0])
    #primary_material_lut_05.Color -> mix_041.A
    HD2_Shader.links.new(primary_material_lut_05.outputs[0], mix_041.inputs[6])
    #primary_material_lut_05.Alpha -> mix_042.A
    HD2_Shader.links.new(primary_material_lut_05.outputs[1], mix_042.inputs[2])
    #math_243.Value -> mix_041.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_041.inputs[0])
    #Slot8.Result -> combine_xyz_062.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_062.inputs[1])
    #math_243.Value -> mix_042.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_042.inputs[0])
    #Slot8.Result -> combine_xyz_063.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_063.inputs[1])
    #mix_041.Result -> gamma_002.Color
    HD2_Shader.links.new(mix_041.outputs[2], gamma_002.inputs[0])
    #combine_xyz_064.Vector -> primary_material_lut_06.Vector
    HD2_Shader.links.new(combine_xyz_064.outputs[0], primary_material_lut_06.inputs[0])
    #primary_material_lut_06.Color -> mix_043.A
    HD2_Shader.links.new(primary_material_lut_06.outputs[0], mix_043.inputs[6])
    #primary_material_lut_06.Alpha -> mix_044.A
    HD2_Shader.links.new(primary_material_lut_06.outputs[1], mix_044.inputs[2])
    #math_243.Value -> mix_043.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_043.inputs[0])
    #Slot8.Result -> combine_xyz_064.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_064.inputs[1])
    #math_243.Value -> mix_044.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_044.inputs[0])
    #Slot8.Result -> combine_xyz_065.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_065.inputs[1])
    #mix_043.Result -> gamma_005.Color
    HD2_Shader.links.new(mix_043.outputs[2], gamma_005.inputs[0])
    #combine_xyz_066.Vector -> primary_material_lut_07.Vector
    HD2_Shader.links.new(combine_xyz_066.outputs[0], primary_material_lut_07.inputs[0])
    #primary_material_lut_07.Color -> mix_045.A
    HD2_Shader.links.new(primary_material_lut_07.outputs[0], mix_045.inputs[6])
    #primary_material_lut_07.Alpha -> mix_046.A
    HD2_Shader.links.new(primary_material_lut_07.outputs[1], mix_046.inputs[2])
    #math_243.Value -> mix_045.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_045.inputs[0])
    #Slot8.Result -> combine_xyz_066.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_066.inputs[1])
    #math_243.Value -> mix_046.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_046.inputs[0])
    #Slot8.Result -> combine_xyz_067.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_067.inputs[1])
    #combine_xyz_068.Vector -> primary_material_lut_08.Vector
    HD2_Shader.links.new(combine_xyz_068.outputs[0], primary_material_lut_08.inputs[0])
    #primary_material_lut_08.Color -> mix_047.A
    HD2_Shader.links.new(primary_material_lut_08.outputs[0], mix_047.inputs[6])
    #primary_material_lut_08.Alpha -> mix_048.A
    HD2_Shader.links.new(primary_material_lut_08.outputs[1], mix_048.inputs[2])
    #math_243.Value -> mix_047.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_047.inputs[0])
    #Slot8.Result -> combine_xyz_068.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_068.inputs[1])
    #math_243.Value -> mix_048.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_048.inputs[0])
    #Slot8.Result -> combine_xyz_069.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_069.inputs[1])
    #mix_047.Result -> separate_xyz_038.Vector
    HD2_Shader.links.new(mix_047.outputs[2], separate_xyz_038.inputs[0])
    #combine_xyz_070.Vector -> primary_material_lut_09.Vector
    HD2_Shader.links.new(combine_xyz_070.outputs[0], primary_material_lut_09.inputs[0])
    #primary_material_lut_09.Color -> mix_049.A
    HD2_Shader.links.new(primary_material_lut_09.outputs[0], mix_049.inputs[6])
    #primary_material_lut_09.Alpha -> mix_050.A
    HD2_Shader.links.new(primary_material_lut_09.outputs[1], mix_050.inputs[2])
    #math_243.Value -> mix_049.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_049.inputs[0])
    #Slot8.Result -> combine_xyz_070.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_070.inputs[1])
    #math_243.Value -> mix_050.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_050.inputs[0])
    #Slot8.Result -> combine_xyz_071.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_071.inputs[1])
    #combine_xyz_072.Vector -> primary_material_lut_10.Vector
    HD2_Shader.links.new(combine_xyz_072.outputs[0], primary_material_lut_10.inputs[0])
    #primary_material_lut_10.Color -> mix_051.A
    HD2_Shader.links.new(primary_material_lut_10.outputs[0], mix_051.inputs[6])
    #primary_material_lut_10.Alpha -> mix_052.A
    HD2_Shader.links.new(primary_material_lut_10.outputs[1], mix_052.inputs[2])
    #math_243.Value -> mix_051.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_051.inputs[0])
    #Slot8.Result -> combine_xyz_072.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_072.inputs[1])
    #math_243.Value -> mix_052.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_052.inputs[0])
    #Slot8.Result -> combine_xyz_073.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_073.inputs[1])
    #combine_xyz_074.Vector -> primary_material_lut_11.Vector
    HD2_Shader.links.new(combine_xyz_074.outputs[0], primary_material_lut_11.inputs[0])
    #primary_material_lut_11.Color -> mix_053.A
    HD2_Shader.links.new(primary_material_lut_11.outputs[0], mix_053.inputs[6])
    #primary_material_lut_11.Alpha -> mix_054.A
    HD2_Shader.links.new(primary_material_lut_11.outputs[1], mix_054.inputs[2])
    #math_243.Value -> mix_053.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_053.inputs[0])
    #Slot8.Result -> combine_xyz_074.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_074.inputs[1])
    #math_243.Value -> mix_054.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_054.inputs[0])
    #Slot8.Result -> combine_xyz_075.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_075.inputs[1])
    #combine_xyz_076.Vector -> primary_material_lut_12.Vector
    HD2_Shader.links.new(combine_xyz_076.outputs[0], primary_material_lut_12.inputs[0])
    #primary_material_lut_12.Color -> mix_055.A
    HD2_Shader.links.new(primary_material_lut_12.outputs[0], mix_055.inputs[6])
    #primary_material_lut_12.Alpha -> mix_056.A
    HD2_Shader.links.new(primary_material_lut_12.outputs[1], mix_056.inputs[2])
    #math_243.Value -> mix_055.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_055.inputs[0])
    #Slot8.Result -> combine_xyz_076.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_076.inputs[1])
    #math_243.Value -> mix_056.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_056.inputs[0])
    #Slot8.Result -> combine_xyz_077.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_077.inputs[1])
    #combine_xyz_078.Vector -> primary_material_lut_13.Vector
    HD2_Shader.links.new(combine_xyz_078.outputs[0], primary_material_lut_13.inputs[0])
    #primary_material_lut_13.Color -> mix_057.A
    HD2_Shader.links.new(primary_material_lut_13.outputs[0], mix_057.inputs[6])
    #primary_material_lut_13.Alpha -> mix_058.A
    HD2_Shader.links.new(primary_material_lut_13.outputs[1], mix_058.inputs[2])
    #math_243.Value -> mix_057.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_057.inputs[0])
    #Slot8.Result -> combine_xyz_078.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_078.inputs[1])
    #math_243.Value -> mix_058.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_058.inputs[0])
    #Slot8.Result -> combine_xyz_079.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_079.inputs[1])
    #combine_xyz_081.Vector -> primary_material_lut_14.Vector
    HD2_Shader.links.new(combine_xyz_081.outputs[0], primary_material_lut_14.inputs[0])
    #primary_material_lut_14.Color -> mix_059.A
    HD2_Shader.links.new(primary_material_lut_14.outputs[0], mix_059.inputs[6])
    #primary_material_lut_14.Alpha -> mix_060.A
    HD2_Shader.links.new(primary_material_lut_14.outputs[1], mix_060.inputs[2])
    #math_243.Value -> mix_059.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_059.inputs[0])
    #Slot8.Result -> combine_xyz_081.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_081.inputs[1])
    #math_243.Value -> mix_060.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_060.inputs[0])
    #Slot8.Result -> combine_xyz_080.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_080.inputs[1])
    #combine_xyz_083.Vector -> primary_material_lut_15.Vector
    HD2_Shader.links.new(combine_xyz_083.outputs[0], primary_material_lut_15.inputs[0])
    #primary_material_lut_15.Color -> mix_061.A
    HD2_Shader.links.new(primary_material_lut_15.outputs[0], mix_061.inputs[6])
    #primary_material_lut_15.Alpha -> mix_062.A
    HD2_Shader.links.new(primary_material_lut_15.outputs[1], mix_062.inputs[2])
    #math_243.Value -> mix_061.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_061.inputs[0])
    #Slot8.Result -> combine_xyz_083.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_083.inputs[1])
    #math_243.Value -> mix_062.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_062.inputs[0])
    #Slot8.Result -> combine_xyz_082.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_082.inputs[1])
    #combine_xyz_085.Vector -> primary_material_lut_16.Vector
    HD2_Shader.links.new(combine_xyz_085.outputs[0], primary_material_lut_16.inputs[0])
    #primary_material_lut_16.Color -> mix_063.A
    HD2_Shader.links.new(primary_material_lut_16.outputs[0], mix_063.inputs[6])
    #primary_material_lut_16.Alpha -> mix_064.A
    HD2_Shader.links.new(primary_material_lut_16.outputs[1], mix_064.inputs[2])
    #math_243.Value -> mix_063.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_063.inputs[0])
    #Slot8.Result -> combine_xyz_085.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_085.inputs[1])
    #math_243.Value -> mix_064.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_064.inputs[0])
    #Slot8.Result -> combine_xyz_084.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_084.inputs[1])
    #combine_xyz_087.Vector -> primary_material_lut_17.Vector
    HD2_Shader.links.new(combine_xyz_087.outputs[0], primary_material_lut_17.inputs[0])
    #primary_material_lut_17.Color -> mix_065.A
    HD2_Shader.links.new(primary_material_lut_17.outputs[0], mix_065.inputs[6])
    #primary_material_lut_17.Alpha -> mix_066.A
    HD2_Shader.links.new(primary_material_lut_17.outputs[1], mix_066.inputs[2])
    #math_243.Value -> mix_065.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_065.inputs[0])
    #Slot8.Result -> combine_xyz_087.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_087.inputs[1])
    #math_243.Value -> mix_066.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_066.inputs[0])
    #Slot8.Result -> combine_xyz_086.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_086.inputs[1])
    #combine_xyz_089.Vector -> primary_material_lut_18.Vector
    HD2_Shader.links.new(combine_xyz_089.outputs[0], primary_material_lut_18.inputs[0])
    #primary_material_lut_18.Color -> mix_067.A
    HD2_Shader.links.new(primary_material_lut_18.outputs[0], mix_067.inputs[6])
    #primary_material_lut_18.Alpha -> mix_068.A
    HD2_Shader.links.new(primary_material_lut_18.outputs[1], mix_068.inputs[2])
    #math_243.Value -> mix_067.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_067.inputs[0])
    #Slot8.Result -> combine_xyz_089.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_089.inputs[1])
    #math_243.Value -> mix_068.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_068.inputs[0])
    #Slot8.Result -> combine_xyz_088.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_088.inputs[1])
    #combine_xyz_091.Vector -> primary_material_lut_19.Vector
    HD2_Shader.links.new(combine_xyz_091.outputs[0], primary_material_lut_19.inputs[0])
    #primary_material_lut_19.Color -> mix_069.A
    HD2_Shader.links.new(primary_material_lut_19.outputs[0], mix_069.inputs[6])
    #primary_material_lut_19.Alpha -> mix_070.A
    HD2_Shader.links.new(primary_material_lut_19.outputs[1], mix_070.inputs[2])
    #math_243.Value -> mix_069.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_069.inputs[0])
    #Slot8.Result -> combine_xyz_091.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_091.inputs[1])
    #math_243.Value -> mix_070.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_070.inputs[0])
    #Slot8.Result -> combine_xyz_090.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_090.inputs[1])
    #combine_xyz_093.Vector -> primary_material_lut_20.Vector
    HD2_Shader.links.new(combine_xyz_093.outputs[0], primary_material_lut_20.inputs[0])
    #primary_material_lut_20.Color -> mix_071.A
    HD2_Shader.links.new(primary_material_lut_20.outputs[0], mix_071.inputs[6])
    #primary_material_lut_20.Alpha -> mix_072.A
    HD2_Shader.links.new(primary_material_lut_20.outputs[1], mix_072.inputs[2])
    #math_243.Value -> mix_071.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_071.inputs[0])
    #Slot8.Result -> combine_xyz_093.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_093.inputs[1])
    #math_243.Value -> mix_072.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_072.inputs[0])
    #Slot8.Result -> combine_xyz_092.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_092.inputs[1])
    #combine_xyz_095.Vector -> primary_material_lut_21.Vector
    HD2_Shader.links.new(combine_xyz_095.outputs[0], primary_material_lut_21.inputs[0])
    #primary_material_lut_21.Color -> mix_073.A
    HD2_Shader.links.new(primary_material_lut_21.outputs[0], mix_073.inputs[6])
    #primary_material_lut_21.Alpha -> mix_074.A
    HD2_Shader.links.new(primary_material_lut_21.outputs[1], mix_074.inputs[2])
    #math_243.Value -> mix_073.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_073.inputs[0])
    #Slot8.Result -> combine_xyz_095.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_095.inputs[1])
    #math_243.Value -> mix_074.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_074.inputs[0])
    #Slot8.Result -> combine_xyz_094.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_094.inputs[1])
    #combine_xyz_097.Vector -> primary_material_lut_22.Vector
    HD2_Shader.links.new(combine_xyz_097.outputs[0], primary_material_lut_22.inputs[0])
    #primary_material_lut_22.Color -> mix_075.A
    HD2_Shader.links.new(primary_material_lut_22.outputs[0], mix_075.inputs[6])
    #primary_material_lut_22.Alpha -> mix_076.A
    HD2_Shader.links.new(primary_material_lut_22.outputs[1], mix_076.inputs[2])
    #math_243.Value -> mix_075.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_075.inputs[0])
    #Slot8.Result -> combine_xyz_097.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_097.inputs[1])
    #math_243.Value -> mix_076.Factor
    HD2_Shader.links.new(math_243.outputs[0], mix_076.inputs[0])
    #Slot8.Result -> combine_xyz_096.Y
    HD2_Shader.links.new(Slot8.outputs[0], combine_xyz_096.inputs[1])
    #mix_073.Result -> separate_xyz_052.Vector
    HD2_Shader.links.new(mix_073.outputs[2], separate_xyz_052.inputs[0])
    #mix_075.Result -> separate_xyz_024.Vector
    HD2_Shader.links.new(mix_075.outputs[2], separate_xyz_024.inputs[0])
    #separate_xyz_025.X -> math_236.Value
    HD2_Shader.links.new(separate_xyz_025.outputs[0], math_236.inputs[0])
    #math_236.Value -> Slot1and2.Factor
    HD2_Shader.links.new(math_236.outputs[0], Slot1and2.inputs[0])
    #separate_xyz_025.Y -> math_237.Value
    HD2_Shader.links.new(separate_xyz_025.outputs[1], math_237.inputs[0])
    #separate_xyz_025.Z -> math_238.Value
    HD2_Shader.links.new(separate_xyz_025.outputs[2], math_238.inputs[0])
    #math_237.Value -> Slot3.Factor
    HD2_Shader.links.new(math_237.outputs[0], Slot3.inputs[0])
    #math_238.Value -> Slot4.Factor
    HD2_Shader.links.new(math_238.outputs[0], Slot4.inputs[0])
    #separate_xyz_043.X -> math_239.Value
    HD2_Shader.links.new(separate_xyz_043.outputs[0], math_239.inputs[0])
    #math_239.Value -> Slot5.Factor
    HD2_Shader.links.new(math_239.outputs[0], Slot5.inputs[0])
    #separate_xyz_043.Y -> math_240.Value
    HD2_Shader.links.new(separate_xyz_043.outputs[1], math_240.inputs[0])
    #separate_xyz_043.Z -> math_241.Value
    HD2_Shader.links.new(separate_xyz_043.outputs[2], math_241.inputs[0])
    #mix_019.Result -> math_242.Value
    HD2_Shader.links.new(mix_019.outputs[0], math_242.inputs[0])
    #math_240.Value -> Slot6.Factor
    HD2_Shader.links.new(math_240.outputs[0], Slot6.inputs[0])
    #math_241.Value -> Slot7.Factor
    HD2_Shader.links.new(math_241.outputs[0], Slot7.inputs[0])
    #math_242.Value -> Slot8.Factor
    HD2_Shader.links.new(math_242.outputs[0], Slot8.inputs[0])
    #mix_030.Result -> math_243.Value
    HD2_Shader.links.new(mix_030.outputs[0], math_243.inputs[0])
    #mapping.Vector -> id_mask_array_02.Vector
    HD2_Shader.links.new(mapping.outputs[0], id_mask_array_02.inputs[0])
    #uv_map.UV -> mapping.Vector
    HD2_Shader.links.new(uv_map.outputs[0], mapping.inputs[0])
    #mapping.Vector -> pattern_mask_array_02.Vector
    HD2_Shader.links.new(mapping.outputs[0], pattern_mask_array_02.inputs[0])
    #id_mask_array_02.Color -> mix_018.B
    HD2_Shader.links.new(id_mask_array_02.outputs[0], mix_018.inputs[7])
    #id_mask_array_02.Alpha -> mix_019.B
    HD2_Shader.links.new(id_mask_array_02.outputs[1], mix_019.inputs[3])
    #pattern_mask_array_02.Color -> mix_030.B
    HD2_Shader.links.new(pattern_mask_array_02.outputs[0], mix_030.inputs[3])
    #separate_xyz_036.Z -> math_088.Value
    HD2_Shader.links.new(separate_xyz_036.outputs[2], math_088.inputs[1])
    #separate_xyz_036.Y -> clamp_019.Value
    HD2_Shader.links.new(separate_xyz_036.outputs[1], clamp_019.inputs[0])
    #separate_xyz_041.X -> combine_xyz_038.X
    HD2_Shader.links.new(separate_xyz_041.outputs[0], combine_xyz_038.inputs[0])
    #separate_xyz_041.Y -> combine_xyz_038.Y
    HD2_Shader.links.new(separate_xyz_041.outputs[1], combine_xyz_038.inputs[1])
    #separate_xyz_060.X -> math_098.Value
    HD2_Shader.links.new(separate_xyz_060.outputs[0], math_098.inputs[1])
    #separate_xyz_060.Z -> mix_011.A
    HD2_Shader.links.new(separate_xyz_060.outputs[2], mix_011.inputs[2])
    #combine_xyz_026.Vector -> vector_math_078.Vector
    HD2_Shader.links.new(combine_xyz_026.outputs[0], vector_math_078.inputs[1])
    #vector_math_083.Vector -> customization_material_detail_tiler_array.Vector
    HD2_Shader.links.new(vector_math_083.outputs[0], customization_material_detail_tiler_array.inputs[0])
    #vector_math_001.Vector -> vector_math_002.Vector
    HD2_Shader.links.new(vector_math_001.outputs[0], vector_math_002.inputs[0])
    #math_226.Value -> math_227.Value
    HD2_Shader.links.new(math_226.outputs[0], math_227.inputs[1])
    #math_227.Value -> combine_xyz_026.Y
    HD2_Shader.links.new(math_227.outputs[0], combine_xyz_026.inputs[1])
    #vector_math_078.Vector -> vector_math_083.Vector
    HD2_Shader.links.new(vector_math_078.outputs[0], vector_math_083.inputs[0])
    #vector_math_002.Vector -> vector_math_078.Vector
    HD2_Shader.links.new(vector_math_002.outputs[0], vector_math_078.inputs[0])
    #object_info.Random -> math_228.Value
    HD2_Shader.links.new(object_info.outputs[5], math_228.inputs[0])
    #combine_xyz_029.Vector -> vector_math_001.Vector
    HD2_Shader.links.new(combine_xyz_029.outputs[0], vector_math_001.inputs[2])
    #math_086.Value -> math_226.Value
    HD2_Shader.links.new(math_086.outputs[0], math_226.inputs[0])
    #math_079.Value -> vector_math_001.Vector
    HD2_Shader.links.new(math_079.outputs[0], vector_math_001.inputs[1])
    #detail_uvs.UV -> vector_math_001.Vector
    HD2_Shader.links.new(detail_uvs.outputs[0], vector_math_001.inputs[0])
    #customization_material_detail_tiler_array.Color -> separate_xyz_035.Vector
    HD2_Shader.links.new(customization_material_detail_tiler_array.outputs[0], separate_xyz_035.inputs[0])
    #math_228.Value -> combine_xyz_029.Y
    HD2_Shader.links.new(math_228.outputs[0], combine_xyz_029.inputs[1])
    #math_229.Value -> combine_xyz_029.X
    HD2_Shader.links.new(math_229.outputs[0], combine_xyz_029.inputs[0])
    #object_info.Random -> math_229.Value
    HD2_Shader.links.new(object_info.outputs[5], math_229.inputs[0])
    #combine_xyz_039.Vector -> vector_math_084.Vector
    HD2_Shader.links.new(combine_xyz_039.outputs[0], vector_math_084.inputs[1])
    #separate_xyz_035.Z -> combine_xyz_039.X
    HD2_Shader.links.new(separate_xyz_035.outputs[2], combine_xyz_039.inputs[0])
    #customization_material_detail_tiler_array.Alpha -> combine_xyz_039.Y
    HD2_Shader.links.new(customization_material_detail_tiler_array.outputs[1], combine_xyz_039.inputs[1])
    #separate_xyz_035.X -> combine_xyz_039.Z
    HD2_Shader.links.new(separate_xyz_035.outputs[0], combine_xyz_039.inputs[2])
    #separate_xyz_035.Y -> math_230.Value
    HD2_Shader.links.new(separate_xyz_035.outputs[1], math_230.inputs[1])
    #vector_math_084.Vector -> separate_xyz_048.Vector
    HD2_Shader.links.new(vector_math_084.outputs[0], separate_xyz_048.inputs[0])
    #math_230.Value -> combine_xyz_043.Y
    HD2_Shader.links.new(math_230.outputs[0], combine_xyz_043.inputs[1])
    #vector_math_084.Vector -> mix_005.B
    HD2_Shader.links.new(vector_math_084.outputs[0], mix_005.inputs[5])
    #vector_math_084.Vector -> separate_xyz_040.Vector
    HD2_Shader.links.new(vector_math_084.outputs[0], separate_xyz_040.inputs[0])
    #combine_xyz_040.Vector -> vector_math_087.Vector
    HD2_Shader.links.new(combine_xyz_040.outputs[0], vector_math_087.inputs[1])
    #vector_math_088.Vector -> customization_material_detail_tiler_array_001.Vector
    HD2_Shader.links.new(vector_math_088.outputs[0], customization_material_detail_tiler_array_001.inputs[0])
    #vector_math_085.Vector -> vector_math_086.Vector
    HD2_Shader.links.new(vector_math_085.outputs[0], vector_math_086.inputs[0])
    #math_231.Value -> math_232.Value
    HD2_Shader.links.new(math_231.outputs[0], math_232.inputs[1])
    #math_232.Value -> combine_xyz_040.Y
    HD2_Shader.links.new(math_232.outputs[0], combine_xyz_040.inputs[1])
    #vector_math_087.Vector -> vector_math_088.Vector
    HD2_Shader.links.new(vector_math_087.outputs[0], vector_math_088.inputs[0])
    #vector_math_086.Vector -> vector_math_087.Vector
    HD2_Shader.links.new(vector_math_086.outputs[0], vector_math_087.inputs[0])
    #detail_uvs.UV -> vector_math_085.Vector
    HD2_Shader.links.new(detail_uvs.outputs[0], vector_math_085.inputs[0])
    #math_152.Value -> vector_math_085.Vector
    HD2_Shader.links.new(math_152.outputs[0], vector_math_085.inputs[1])
    #math_155.Value -> math_231.Value
    HD2_Shader.links.new(math_155.outputs[0], math_231.inputs[0])
    #customization_material_detail_tiler_array_001.Color -> vector_math.Vector
    HD2_Shader.links.new(customization_material_detail_tiler_array_001.outputs[0], vector_math.inputs[1])
    #customization_material_detail_tiler_array_001.Alpha -> math_235.Value
    HD2_Shader.links.new(customization_material_detail_tiler_array_001.outputs[1], math_235.inputs[1])
    #vector_math.Vector -> vector_math_033.Vector
    HD2_Shader.links.new(vector_math.outputs[0], vector_math_033.inputs[1])
    #mix_031.Result -> math_245.Value
    HD2_Shader.links.new(mix_031.outputs[0], math_245.inputs[0])
    #mix_031.Result -> math_244.Value
    HD2_Shader.links.new(mix_031.outputs[0], math_244.inputs[0])
    #combine_xyz_034.Vector -> vector_math_100.Vector
    HD2_Shader.links.new(combine_xyz_034.outputs[0], vector_math_100.inputs[1])
    #vector_math_101.Vector -> composite_array.Vector
    HD2_Shader.links.new(vector_math_101.outputs[0], composite_array.inputs[0])
    #vector_math_090.Vector -> vector_math_095.Vector
    HD2_Shader.links.new(vector_math_090.outputs[0], vector_math_095.inputs[0])
    #math_246.Value -> math_247.Value
    HD2_Shader.links.new(math_246.outputs[0], math_247.inputs[1])
    #math_247.Value -> combine_xyz_034.Y
    HD2_Shader.links.new(math_247.outputs[0], combine_xyz_034.inputs[1])
    #vector_math_100.Vector -> vector_math_101.Vector
    HD2_Shader.links.new(vector_math_100.outputs[0], vector_math_101.inputs[0])
    #vector_math_095.Vector -> vector_math_100.Vector
    HD2_Shader.links.new(vector_math_095.outputs[0], vector_math_100.inputs[0])
    #math_048.Value -> math_246.Value
    HD2_Shader.links.new(math_048.outputs[0], math_246.inputs[0])
    #detail_uvs.UV -> vector_math_090.Vector
    HD2_Shader.links.new(detail_uvs.outputs[0], vector_math_090.inputs[0])
    #math_045.Value -> vector_math_090.Vector
    HD2_Shader.links.new(math_045.outputs[0], vector_math_090.inputs[1])
    #composite_array.Color -> vector_math_022.Vector
    HD2_Shader.links.new(composite_array.outputs[0], vector_math_022.inputs[1])
    #combine_xyz_029.Vector -> vector_math_085.Vector
    HD2_Shader.links.new(combine_xyz_029.outputs[0], vector_math_085.inputs[2])
    #combine_xyz_029.Vector -> vector_math_090.Vector
    HD2_Shader.links.new(combine_xyz_029.outputs[0], vector_math_090.inputs[2])
    #separate_xyz_020.X -> combine_xyz_041.X
    HD2_Shader.links.new(separate_xyz_020.outputs[0], combine_xyz_041.inputs[0])
    #separate_xyz_020.Y -> combine_xyz_041.Y
    HD2_Shader.links.new(separate_xyz_020.outputs[1], combine_xyz_041.inputs[1])
    #combine_xyz_041.Vector -> vector_math_102.Vector
    HD2_Shader.links.new(combine_xyz_041.outputs[0], vector_math_102.inputs[0])
    #mix_039.Result -> vector_math_102.Vector
    HD2_Shader.links.new(mix_039.outputs[2], vector_math_102.inputs[1])
    #mix_040.Result -> math_233.Value
    HD2_Shader.links.new(mix_040.outputs[0], math_233.inputs[1])
    #math_248.Value -> math_234.Value
    HD2_Shader.links.new(math_248.outputs[0], math_234.inputs[2])
    #math_234.Value -> math_250.Value
    HD2_Shader.links.new(math_234.outputs[0], math_250.inputs[2])
    #math_249.Value -> math_248.Value
    HD2_Shader.links.new(math_249.outputs[0], math_248.inputs[2])
    #math_233.Value -> math_250.Value
    HD2_Shader.links.new(math_233.outputs[0], math_250.inputs[0])
    #separate_xyz_037.Z -> math_234.Value
    HD2_Shader.links.new(separate_xyz_037.outputs[2], math_234.inputs[1])
    #separate_xyz_023.Y -> math_250.Value
    HD2_Shader.links.new(separate_xyz_023.outputs[1], math_250.inputs[1])
    #separate_xyz_005.Y -> math_248.Value
    HD2_Shader.links.new(separate_xyz_005.outputs[1], math_248.inputs[0])
    #separate_xyz_037.X -> math_249.Value
    HD2_Shader.links.new(separate_xyz_037.outputs[0], math_249.inputs[1])
    #separate_xyz_005.X -> math_249.Value
    HD2_Shader.links.new(separate_xyz_005.outputs[0], math_249.inputs[0])
    #separate_xyz_005.Z -> math_234.Value
    HD2_Shader.links.new(separate_xyz_005.outputs[2], math_234.inputs[0])
    #separate_xyz_037.Y -> math_248.Value
    HD2_Shader.links.new(separate_xyz_037.outputs[1], math_248.inputs[1])
    #vector_math_102.Vector -> separate_xyz_005.Vector
    HD2_Shader.links.new(vector_math_102.outputs[0], separate_xyz_005.inputs[0])
    #combine_xyz_024.Vector -> separate_xyz_037.Vector
    HD2_Shader.links.new(combine_xyz_024.outputs[0], separate_xyz_037.inputs[0])
    #math_250.Value -> math_067.Value
    HD2_Shader.links.new(math_250.outputs[0], math_067.inputs[0])
    #separate_xyz_044.X -> combine_xyz_049.X
    HD2_Shader.links.new(separate_xyz_044.outputs[0], combine_xyz_049.inputs[0])
    #separate_xyz_044.Y -> combine_xyz_049.Y
    HD2_Shader.links.new(separate_xyz_044.outputs[1], combine_xyz_049.inputs[1])
    #combine_xyz_049.Vector -> vector_math_103.Vector
    HD2_Shader.links.new(combine_xyz_049.outputs[0], vector_math_103.inputs[0])
    #mix_037.Result -> vector_math_103.Vector
    HD2_Shader.links.new(mix_037.outputs[2], vector_math_103.inputs[1])
    #mix_038.Result -> math_251.Value
    HD2_Shader.links.new(mix_038.outputs[0], math_251.inputs[1])
    #math_253.Value -> math_252.Value
    HD2_Shader.links.new(math_253.outputs[0], math_252.inputs[2])
    #math_252.Value -> math_255.Value
    HD2_Shader.links.new(math_252.outputs[0], math_255.inputs[2])
    #math_254.Value -> math_253.Value
    HD2_Shader.links.new(math_254.outputs[0], math_253.inputs[2])
    #math_251.Value -> math_255.Value
    HD2_Shader.links.new(math_251.outputs[0], math_255.inputs[0])
    #separate_xyz_046.Z -> math_252.Value
    HD2_Shader.links.new(separate_xyz_046.outputs[2], math_252.inputs[1])
    #math_076.Value -> math_255.Value
    HD2_Shader.links.new(math_076.outputs[0], math_255.inputs[1])
    #separate_xyz_042.Y -> math_253.Value
    HD2_Shader.links.new(separate_xyz_042.outputs[1], math_253.inputs[0])
    #separate_xyz_046.X -> math_254.Value
    HD2_Shader.links.new(separate_xyz_046.outputs[0], math_254.inputs[1])
    #separate_xyz_042.X -> math_254.Value
    HD2_Shader.links.new(separate_xyz_042.outputs[0], math_254.inputs[0])
    #separate_xyz_042.Z -> math_252.Value
    HD2_Shader.links.new(separate_xyz_042.outputs[2], math_252.inputs[0])
    #separate_xyz_046.Y -> math_253.Value
    HD2_Shader.links.new(separate_xyz_046.outputs[1], math_253.inputs[1])
    #vector_math_103.Vector -> separate_xyz_042.Vector
    HD2_Shader.links.new(vector_math_103.outputs[0], separate_xyz_042.inputs[0])
    #combine_xyz_046.Vector -> separate_xyz_046.Vector
    HD2_Shader.links.new(combine_xyz_046.outputs[0], separate_xyz_046.inputs[0])
    #math_255.Value -> math_088.Value
    HD2_Shader.links.new(math_255.outputs[0], math_088.inputs[0])
    #mix_045.Result -> vector_math_104.Vector
    HD2_Shader.links.new(mix_045.outputs[2], vector_math_104.inputs[1])
    #separate_xyz_033.X -> combine_xyz_032.X
    HD2_Shader.links.new(separate_xyz_033.outputs[0], combine_xyz_032.inputs[0])
    #separate_xyz_033.Y -> combine_xyz_032.Y
    HD2_Shader.links.new(separate_xyz_033.outputs[1], combine_xyz_032.inputs[1])
    #combine_xyz_032.Vector -> vector_math_104.Vector
    HD2_Shader.links.new(combine_xyz_032.outputs[0], vector_math_104.inputs[0])
    #mix_046.Result -> math_256.Value
    HD2_Shader.links.new(mix_046.outputs[0], math_256.inputs[1])
    #math_258.Value -> math_257.Value
    HD2_Shader.links.new(math_258.outputs[0], math_257.inputs[2])
    #math_257.Value -> math_260.Value
    HD2_Shader.links.new(math_257.outputs[0], math_260.inputs[2])
    #math_259.Value -> math_258.Value
    HD2_Shader.links.new(math_259.outputs[0], math_258.inputs[2])
    #math_256.Value -> math_260.Value
    HD2_Shader.links.new(math_256.outputs[0], math_260.inputs[0])
    #separate_xyz_049.Z -> math_257.Value
    HD2_Shader.links.new(separate_xyz_049.outputs[2], math_257.inputs[1])
    #separate_xyz_023.Y -> math_260.Value
    HD2_Shader.links.new(separate_xyz_023.outputs[1], math_260.inputs[1])
    #separate_xyz_034.Y -> math_258.Value
    HD2_Shader.links.new(separate_xyz_034.outputs[1], math_258.inputs[0])
    #separate_xyz_049.X -> math_259.Value
    HD2_Shader.links.new(separate_xyz_049.outputs[0], math_259.inputs[1])
    #separate_xyz_034.X -> math_259.Value
    HD2_Shader.links.new(separate_xyz_034.outputs[0], math_259.inputs[0])
    #separate_xyz_034.Z -> math_257.Value
    HD2_Shader.links.new(separate_xyz_034.outputs[2], math_257.inputs[0])
    #separate_xyz_049.Y -> math_258.Value
    HD2_Shader.links.new(separate_xyz_049.outputs[1], math_258.inputs[1])
    #vector_math_104.Vector -> separate_xyz_034.Vector
    HD2_Shader.links.new(vector_math_104.outputs[0], separate_xyz_034.inputs[0])
    #combine_xyz_033.Vector -> separate_xyz_049.Vector
    HD2_Shader.links.new(combine_xyz_033.outputs[0], separate_xyz_049.inputs[0])
    #math_260.Value -> math_063.Value
    HD2_Shader.links.new(math_260.outputs[0], math_063.inputs[0])
    #mix_049.Result -> vector_math_105.Vector
    HD2_Shader.links.new(mix_049.outputs[2], vector_math_105.inputs[1])
    #separate_xyz_010.X -> combine_xyz_050.X
    HD2_Shader.links.new(separate_xyz_010.outputs[0], combine_xyz_050.inputs[0])
    #separate_xyz_010.Y -> combine_xyz_050.Y
    HD2_Shader.links.new(separate_xyz_010.outputs[1], combine_xyz_050.inputs[1])
    #combine_xyz_050.Vector -> vector_math_105.Vector
    HD2_Shader.links.new(combine_xyz_050.outputs[0], vector_math_105.inputs[0])
    #mix_050.Result -> math_261.Value
    HD2_Shader.links.new(mix_050.outputs[0], math_261.inputs[1])
    #clamp_003.Result -> combine_xyz_009.Z
    HD2_Shader.links.new(clamp_003.outputs[0], combine_xyz_009.inputs[2])
    #math_263.Value -> math_262.Value
    HD2_Shader.links.new(math_263.outputs[0], math_262.inputs[2])
    #math_262.Value -> math_265.Value
    HD2_Shader.links.new(math_262.outputs[0], math_265.inputs[2])
    #math_264.Value -> math_263.Value
    HD2_Shader.links.new(math_264.outputs[0], math_263.inputs[2])
    #clamp_004.Result -> math_265.Value
    HD2_Shader.links.new(clamp_004.outputs[0], math_265.inputs[0])
    #separate_xyz_050.Z -> math_262.Value
    HD2_Shader.links.new(separate_xyz_050.outputs[2], math_262.inputs[1])
    #math_261.Value -> math_265.Value
    HD2_Shader.links.new(math_261.outputs[0], math_265.inputs[1])
    #separate_xyz_012.Y -> math_263.Value
    HD2_Shader.links.new(separate_xyz_012.outputs[1], math_263.inputs[0])
    #separate_xyz_050.X -> math_264.Value
    HD2_Shader.links.new(separate_xyz_050.outputs[0], math_264.inputs[1])
    #separate_xyz_012.X -> math_264.Value
    HD2_Shader.links.new(separate_xyz_012.outputs[0], math_264.inputs[0])
    #separate_xyz_012.Z -> math_262.Value
    HD2_Shader.links.new(separate_xyz_012.outputs[2], math_262.inputs[0])
    #separate_xyz_050.Y -> math_263.Value
    HD2_Shader.links.new(separate_xyz_050.outputs[1], math_263.inputs[1])
    #combine_xyz_009.Vector -> separate_xyz_012.Vector
    HD2_Shader.links.new(combine_xyz_009.outputs[0], separate_xyz_012.inputs[0])
    #vector_math_105.Vector -> separate_xyz_050.Vector
    HD2_Shader.links.new(vector_math_105.outputs[0], separate_xyz_050.inputs[0])
    #math_265.Value -> math_098.Value
    HD2_Shader.links.new(math_265.outputs[0], math_098.inputs[0])
    #combine_xyz_051.Vector -> vector_math_106.Vector
    HD2_Shader.links.new(combine_xyz_051.outputs[0], vector_math_106.inputs[1])
    #vector_math_109.Vector -> vector_math_108.Vector
    HD2_Shader.links.new(vector_math_109.outputs[0], vector_math_108.inputs[0])
    #math_266.Value -> math_267.Value
    HD2_Shader.links.new(math_266.outputs[0], math_267.inputs[1])
    #math_267.Value -> combine_xyz_051.Y
    HD2_Shader.links.new(math_267.outputs[0], combine_xyz_051.inputs[1])
    #vector_math_106.Vector -> vector_math_107.Vector
    HD2_Shader.links.new(vector_math_106.outputs[0], vector_math_107.inputs[0])
    #combine_xyz_010.Vector -> vector_math_106.Vector
    HD2_Shader.links.new(combine_xyz_010.outputs[0], vector_math_106.inputs[0])
    #vector_math_108.Vector -> separate_xyz_002.Vector
    HD2_Shader.links.new(vector_math_108.outputs[0], separate_xyz_002.inputs[0])
    #separate_xyz_002.X -> combine_xyz_010.X
    HD2_Shader.links.new(separate_xyz_002.outputs[0], combine_xyz_010.inputs[0])
    #clamp_020.Result -> combine_xyz_010.Y
    HD2_Shader.links.new(clamp_020.outputs[0], combine_xyz_010.inputs[1])
    #separate_xyz_002.Y -> clamp_020.Value
    HD2_Shader.links.new(separate_xyz_002.outputs[1], clamp_020.inputs[0])
    #vector_math_107.Vector -> customization_camo_tiler_array.Vector
    HD2_Shader.links.new(vector_math_107.outputs[0], customization_camo_tiler_array.inputs[0])
    #detail_uvs.UV -> vector_math_109.Vector
    HD2_Shader.links.new(detail_uvs.outputs[0], vector_math_109.inputs[0])
    #combine_xyz_029.Vector -> vector_math_109.Vector
    HD2_Shader.links.new(combine_xyz_029.outputs[0], vector_math_109.inputs[2])
    #math_019.Value -> vector_math_109.Vector
    HD2_Shader.links.new(math_019.outputs[0], vector_math_109.inputs[1])
    #mix_074.Result -> math_266.Value
    HD2_Shader.links.new(mix_074.outputs[0], math_266.inputs[0])
    #customization_camo_tiler_array.Color -> vector_math_009.Vector
    HD2_Shader.links.new(customization_camo_tiler_array.outputs[0], vector_math_009.inputs[1])
    #pattern_lut.Alpha -> math_113.Value
    HD2_Shader.links.new(pattern_lut.outputs[1], math_113.inputs[0])
    #math_177.Value -> mix_014.Factor
    HD2_Shader.links.new(math_177.outputs[0], mix_014.inputs[0])
    #vector_math_071.Vector -> group_output_1.Color
    #HD2_Shader.links.new(vector_math_071.outputs[0], group_output_1.inputs[1])
    HD2_Shader.links.new(decal_mix_node.outputs[2], group_output_1.inputs[1])
    #math_278.Value -> group_output_1.Metallic
    HD2_Shader.links.new(math_278.outputs[0], group_output_1.inputs[3])
    #math_050.Value -> group_output_1.Roughness
    HD2_Shader.links.new(math_050.outputs[0], group_output_1.inputs[4])
    #gamma.Color -> group_output_1.Clearcoat Normal
    HD2_Shader.links.new(gamma.outputs[0], group_output_1.inputs[6])
    #alpha_cutoff_node.value -> group_output_1.Alpha
    HD2_Shader.links.new(alpha_cutoff_node.outputs[0], group_output_1.inputs[7])
    #gamma_001.Color -> group_output_1.Normal
    HD2_Shader.links.new(gamma_001.outputs[0], group_output_1.inputs[5])
    #separate_xyz_051.Z -> math_194.Value
    HD2_Shader.links.new(separate_xyz_051.outputs[2], math_194.inputs[0])
    #principled_bsdf_001.BSDF -> mix_shader.Shader
    HD2_Shader.links.new(principled_bsdf_001.outputs[0], mix_shader.inputs[2])
    #math_194.Value -> mix_shader.Fac
    HD2_Shader.links.new(math_194.outputs[0], mix_shader.inputs[0])
    #clamp_019.Result -> vector_math_061.Vector
    HD2_Shader.links.new(clamp_019.outputs[0], vector_math_061.inputs[1])
    #vector_math_114.Vector -> vector_math_111.Vector
    HD2_Shader.links.new(vector_math_114.outputs[0], vector_math_111.inputs[0])
    #vector_math_113.Vector -> mix_021.A
    HD2_Shader.links.new(vector_math_113.outputs[0], mix_021.inputs[4])
    #math_151.Value -> mix_021.Factor
    HD2_Shader.links.new(math_151.outputs[0], mix_021.inputs[0])
    #math_078.Value -> vector_math_111.Vector
    HD2_Shader.links.new(math_078.outputs[0], vector_math_111.inputs[1])
    #vector_math_112.Vector -> mix_021.B
    HD2_Shader.links.new(vector_math_112.outputs[0], mix_021.inputs[5])
    #vector_math.Vector -> vector_math_112.Vector
    HD2_Shader.links.new(vector_math.outputs[0], vector_math_112.inputs[0])
    #combine_xyz_098.Vector -> vector_math_113.Vector
    HD2_Shader.links.new(combine_xyz_098.outputs[0], vector_math_113.inputs[0])
    #separate_xyz_048.Z -> combine_xyz_098.X
    HD2_Shader.links.new(separate_xyz_048.outputs[2], combine_xyz_098.inputs[0])
    #math_230.Value -> combine_xyz_098.Y
    HD2_Shader.links.new(math_230.outputs[0], combine_xyz_098.inputs[1])
    #vector_math_022.Vector -> vector_math_114.Vector
    HD2_Shader.links.new(vector_math_022.outputs[0], vector_math_114.inputs[0])
    #vector_math_111.Vector -> vector_math_115.Vector
    HD2_Shader.links.new(vector_math_111.outputs[0], vector_math_115.inputs[0])
    #mix_021.Result -> vector_math_116.Vector
    HD2_Shader.links.new(mix_021.outputs[1], vector_math_116.inputs[0])
    #clamp_019.Result -> vector_math_115.Vector
    HD2_Shader.links.new(clamp_019.outputs[0], vector_math_115.inputs[1])
    #clamp_019.Result -> vector_math_116.Vector
    HD2_Shader.links.new(clamp_019.outputs[0], vector_math_116.inputs[1])
    #vector_math_116.Vector -> vector_math_089.Vector
    HD2_Shader.links.new(vector_math_116.outputs[0], vector_math_089.inputs[0])
    #combine_xyz_099.Vector -> vector_math_092.Vector
    HD2_Shader.links.new(combine_xyz_099.outputs[0], vector_math_092.inputs[0])
    #math_205.Value -> math_268.Value
    HD2_Shader.links.new(math_205.outputs[0], math_268.inputs[1])
    #math_270.Value -> combine_xyz_099.Y
    HD2_Shader.links.new(math_270.outputs[0], combine_xyz_099.inputs[1])
    #math_268.Value -> math_269.Value
    HD2_Shader.links.new(math_268.outputs[0], math_269.inputs[0])
    #math_269.Value -> combine_xyz_099.Z
    HD2_Shader.links.new(math_269.outputs[0], combine_xyz_099.inputs[2])
    #math_204.Value -> math_205.Value
    HD2_Shader.links.new(math_204.outputs[0], math_205.inputs[2])
    #math_271.Value -> math_205.Value
    HD2_Shader.links.new(math_271.outputs[0], math_205.inputs[0])
    #math_271.Value -> math_205.Value
    HD2_Shader.links.new(math_271.outputs[0], math_205.inputs[1])
    #math_270.Value -> math_204.Value
    HD2_Shader.links.new(math_270.outputs[0], math_204.inputs[1])
    #math_270.Value -> math_204.Value
    HD2_Shader.links.new(math_270.outputs[0], math_204.inputs[0])
    #math_271.Value -> combine_xyz_099.X
    HD2_Shader.links.new(math_271.outputs[0], combine_xyz_099.inputs[0])
    #separate_xyz_056.X -> math_271.Value
    HD2_Shader.links.new(separate_xyz_056.outputs[0], math_271.inputs[0])
    #separate_xyz_056.Y -> math_270.Value
    HD2_Shader.links.new(separate_xyz_056.outputs[1], math_270.inputs[0])
    #vector_math_092.Vector -> vector_math_093.Vector
    HD2_Shader.links.new(vector_math_092.outputs[0], vector_math_093.inputs[0])
    #vector_math_093.Vector -> vector_math_094.Vector
    HD2_Shader.links.new(vector_math_093.outputs[0], vector_math_094.inputs[0])
    #vector_math_097.Vector -> separate_xyz_056.Vector
    HD2_Shader.links.new(vector_math_097.outputs[0], separate_xyz_056.inputs[0])
    #vector_math_094.Vector -> normal_map_001.Color
    HD2_Shader.links.new(vector_math_094.outputs[0], normal_map_001.inputs[1])
    #vector_math_094.Vector -> gamma.Color
    HD2_Shader.links.new(vector_math_094.outputs[0], gamma.inputs[0])
    #combine_xyz_100.Vector -> vector_math_117.Vector
    HD2_Shader.links.new(combine_xyz_100.outputs[0], vector_math_117.inputs[0])
    #math_274.Value -> math_275.Value
    HD2_Shader.links.new(math_274.outputs[0], math_275.inputs[1])
    #math_272.Value -> combine_xyz_100.Y
    HD2_Shader.links.new(math_272.outputs[0], combine_xyz_100.inputs[1])
    #math_275.Value -> math_276.Value
    HD2_Shader.links.new(math_275.outputs[0], math_276.inputs[0])
    #math_276.Value -> combine_xyz_100.Z
    HD2_Shader.links.new(math_276.outputs[0], combine_xyz_100.inputs[2])
    #math_273.Value -> math_274.Value
    HD2_Shader.links.new(math_273.outputs[0], math_274.inputs[2])
    #math_277.Value -> math_274.Value
    HD2_Shader.links.new(math_277.outputs[0], math_274.inputs[0])
    #math_277.Value -> math_274.Value
    HD2_Shader.links.new(math_277.outputs[0], math_274.inputs[1])
    #math_272.Value -> math_273.Value
    HD2_Shader.links.new(math_272.outputs[0], math_273.inputs[1])
    #math_272.Value -> math_273.Value
    HD2_Shader.links.new(math_272.outputs[0], math_273.inputs[0])
    #math_277.Value -> combine_xyz_100.X
    HD2_Shader.links.new(math_277.outputs[0], combine_xyz_100.inputs[0])
    #separate_xyz_063.X -> math_277.Value
    HD2_Shader.links.new(separate_xyz_063.outputs[0], math_277.inputs[0])
    #separate_xyz_063.Y -> math_272.Value
    HD2_Shader.links.new(separate_xyz_063.outputs[1], math_272.inputs[0])
    #vector_math_117.Vector -> vector_math_110.Vector
    HD2_Shader.links.new(vector_math_117.outputs[0], vector_math_110.inputs[0])
    #vector_math_110.Vector -> vector_math_118.Vector
    HD2_Shader.links.new(vector_math_110.outputs[0], vector_math_118.inputs[0])
    #vector_math_118.Vector -> normal_map.Color
    HD2_Shader.links.new(vector_math_118.outputs[0], normal_map.inputs[1])
    #vector_math_091.Vector -> separate_xyz_063.Vector
    HD2_Shader.links.new(vector_math_091.outputs[0], separate_xyz_063.inputs[0])
    #vector_math_118.Vector -> gamma_001.Color
    HD2_Shader.links.new(vector_math_118.outputs[0], gamma_001.inputs[0])
    #clamp_018.Result -> math_278.Value
    HD2_Shader.links.new(clamp_018.outputs[0], math_278.inputs[0])
    #separate_xyz_060.Z -> mix_078.B
    HD2_Shader.links.new(separate_xyz_060.outputs[2], mix_078.inputs[3])
    #math_151.Value -> mix_078.Factor
    HD2_Shader.links.new(math_151.outputs[0], mix_078.inputs[0])
    #mix_078.Result -> math_146.Value
    HD2_Shader.links.new(mix_078.outputs[0], math_146.inputs[0])
    #math_194.Value -> group_output_1.Ambient Occlusion
    HD2_Shader.links.new(math_194.outputs[0], group_output_1.inputs[2])
    
    return HD2_Shader

def create_cape_decal_template() -> NodeTree:
    node_tree: NodeTree = bpy.data.node_groups.new("Cape Decal Template", type="ShaderNodeTree")
    node_tree.interface.new_socket("Color", in_out="OUTPUT", socket_type="NodeSocketColor")
    node_tree.interface.new_socket("Alpha", in_out="OUTPUT", socket_type="NodeSocketFloat")
    cape_layer_tree = create_cape_layer()
    
    cape_layer_node: ShaderNodeGroup = node_tree.nodes.new("ShaderNodeGroup")
    cape_layer_node.name = "Layer 1"
    cape_layer_node.node_tree = cape_layer_tree
    cape_layer_node.update()
    cape_layer_node.inputs[0].default_value = 4
    cape_layer_node.inputs[1].default_value = 5

    mask_mix_node: ShaderNodeMix = node_tree.nodes.new("ShaderNodeMix")
    mask_mix_node.data_type = "VECTOR"
    node_tree.links.new(cape_layer_node.outputs[2], mask_mix_node.inputs[0])
    node_tree.links.new(cape_layer_node.outputs[3], mask_mix_node.inputs[5])
    mask_mix_node.inputs[4].default_value = (0, 0, 0)
    mask_mix_node.location = cape_layer_node.location + Vector((300, -100))

    alpha_mult_node: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    alpha_mult_node.operation = "MULTIPLY"
    node_tree.links.new(cape_layer_node.outputs[1], alpha_mult_node.inputs[0])
    node_tree.links.new(mask_mix_node.outputs[1], alpha_mult_node.inputs[1])
    alpha_mult_node.location = mask_mix_node.location + Vector((300, 0))
    alpha_mult_node.hide = True

    greater_than_node: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    greater_than_node.operation = "GREATER_THAN"
    greater_than_node.inputs[1].default_value = 0.5
    node_tree.links.new(alpha_mult_node.outputs[0], greater_than_node.inputs[0])
    greater_than_node.location = alpha_mult_node.location + Vector((0, -50))
    greater_than_node.hide = True

    group_output: NodeGroupOutput = node_tree.nodes.new("NodeGroupOutput")
    group_output.location = cape_layer_node.location + Vector((900, 0))
    node_tree.links.new(cape_layer_node.outputs[0], group_output.inputs[0])
    node_tree.links.new(greater_than_node.outputs[0], group_output.inputs[1])

    return node_tree

def create_cape_layer() -> NodeTree:
    node_tree: NodeTree = bpy.data.node_groups.new("Cape Decal Layer", "ShaderNodeTree")
    node_tree.interface.new_socket("Cape X Coord 1", socket_type="NodeSocketInt")
    node_tree.interface.new_socket("Cape X Coord 2", socket_type="NodeSocketInt")
    node_tree.interface.new_socket("Color", in_out="OUTPUT", socket_type="NodeSocketColor")
    node_tree.interface.new_socket("Alpha", in_out="OUTPUT", socket_type="NodeSocketFloat")
    node_tree.interface.new_socket("Factor", in_out="OUTPUT", socket_type="NodeSocketFloat")
    node_tree.interface.new_socket("Mask", in_out="OUTPUT", socket_type="NodeSocketVector")

    if "Group Input" not in node_tree.nodes:
        node_tree.nodes.new("NodeGroupInput").name = "Group Input"
    group_input: NodeGroupInput = node_tree.nodes["Group Input"]

    VEC_HORIZ = Vector((300, 0))
    VEC_VERT = Vector((0, 50))

    cape_lut_tree = create_cape_lut_tree()
    cape_lut_node_1: ShaderNodeGroup = node_tree.nodes.new("ShaderNodeGroup")
    cape_lut_node_1.node_tree = cape_lut_tree
    node_tree.links.new(group_input.outputs[0], cape_lut_node_1.inputs[0])
    cape_lut_node_1.name = "Cape LUT"
    cape_lut_node_1.location = group_input.location + VEC_HORIZ + (VEC_VERT * 3.5)

    cape_lut_node_2: ShaderNodeGroup = node_tree.nodes.new("ShaderNodeGroup")
    cape_lut_node_2.node_tree = cape_lut_tree
    node_tree.links.new(group_input.outputs[1], cape_lut_node_2.inputs[0])
    cape_lut_node_2.location = group_input.location + VEC_HORIZ - (VEC_VERT * 3.5)

    separate_color_1: ShaderNodeSeparateColor = node_tree.nodes.new("ShaderNodeSeparateColor")
    separate_color_1.mode = "RGB"
    separate_color_1.location = cape_lut_node_1.location + VEC_HORIZ
    separate_color_1.hide = True
    node_tree.links.new(cape_lut_node_1.outputs[0], separate_color_1.inputs[0])
    r9_xyzw: List[NodeSocketFloat] = [*list(separate_color_1.outputs), cape_lut_node_1.outputs[1]]

    combine_r9_xy_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    node_tree.links.new(r9_xyzw[0], combine_r9_xy_1.inputs[0])
    node_tree.links.new(r9_xyzw[1], combine_r9_xy_1.inputs[1])
    combine_r9_xy_1.location = separate_color_1.location + VEC_HORIZ + VEC_VERT
    combine_r9_xy_1.hide = True

    combine_r9_zw_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    node_tree.links.new(r9_xyzw[2], combine_r9_zw_1.inputs[0])
    node_tree.links.new(r9_xyzw[3], combine_r9_zw_1.inputs[1])
    combine_r9_zw_1.location = separate_color_1.location + VEC_HORIZ - VEC_VERT
    combine_r9_zw_1.hide = True

    separate_color_2: ShaderNodeSeparateColor = node_tree.nodes.new("ShaderNodeSeparateColor")
    separate_color_2.mode = "RGB"
    separate_color_2.location = cape_lut_node_2.location + VEC_HORIZ
    separate_color_2.hide = True
    node_tree.links.new(cape_lut_node_2.outputs[0], separate_color_2.inputs[0])
    r20_xyzw: List[NodeSocketFloat] = [*list(separate_color_2.outputs), cape_lut_node_2.outputs[1]]

    # r13.x r25.z
    sine_node: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    sine_node.operation = "SINE"
    node_tree.links.new(r20_xyzw[2], sine_node.inputs[0])
    sine_node.hide = True
    sine_node.location = separate_color_2.location + VEC_HORIZ + VEC_VERT

    # r14.x, r25.y
    cosine_node: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    cosine_node.operation = "COSINE"
    node_tree.links.new(r20_xyzw[2], cosine_node.inputs[0])
    cosine_node.hide = True
    cosine_node.location = separate_color_2.location + VEC_HORIZ

    # r25.x
    negate_node_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    negate_node_1.operation = "MULTIPLY"
    node_tree.links.new(sine_node.outputs[0], negate_node_1.inputs[0])
    negate_node_1.inputs[1].default_value = -1.0
    negate_node_1.hide = True
    negate_node_1.location = separate_color_2.location + VEC_HORIZ - VEC_VERT

    zw_mult_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    zw_mult_1.operation = 'MULTIPLY_ADD'
    zw_mult_1.inputs[1].default_value = (-0.5, -0.5, 0.0)
    zw_mult_1.hide = True
    node_tree.links.new(combine_r9_zw_1.outputs[0], zw_mult_1.inputs[0])
    node_tree.links.new(combine_r9_xy_1.outputs[0], zw_mult_1.inputs[2])
    zw_mult_1.location = combine_r9_xy_1.location + VEC_HORIZ - VEC_VERT

    # r11.zw
    zw_max_node_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    zw_max_node_1.operation = "MAXIMUM"
    node_tree.links.new(combine_r9_zw_1.outputs[0], zw_max_node_1.inputs[0])
    zw_max_node_1.inputs[1].default_value = (0.0001, 0.0001, 0.0000)
    zw_max_node_1.hide = True
    zw_max_node_1.location = combine_r9_xy_1.location + VEC_HORIZ - (VEC_VERT * 2)

    decal_uv_map: ShaderNodeUVMap = node_tree.nodes.new("ShaderNodeUVMap")
    decal_uv_map.location = combine_r9_xy_1.location + VEC_HORIZ + (VEC_VERT * 3)
    decal_uv_map.uv_map = "UVMap.002"

    combine_r25_yz_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    node_tree.links.new(cosine_node.outputs[0], combine_r25_yz_1.inputs[0])
    node_tree.links.new(sine_node.outputs[0], combine_r25_yz_1.inputs[1])
    combine_r25_yz_1.location = sine_node.location + VEC_HORIZ - (VEC_VERT / 2)
    combine_r25_yz_1.hide = True

    combine_r25_xy_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    node_tree.links.new(cosine_node.outputs[0], combine_r25_xy_1.inputs[1])
    node_tree.links.new(negate_node_1.outputs[0], combine_r25_xy_1.inputs[0])
    combine_r25_xy_1.location = sine_node.location + VEC_HORIZ - (VEC_VERT * 1.5)
    combine_r25_xy_1.hide = True

    uv_map_separate: ShaderNodeSeparateXYZ = node_tree.nodes.new("ShaderNodeSeparateXYZ")
    uv_map_separate.location = decal_uv_map.location + VEC_HORIZ
    node_tree.links.new(decal_uv_map.outputs[0], uv_map_separate.inputs[0])
    uv_map_separate.hide = True

    uv_map_invert_y: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    uv_map_invert_y.operation = "SUBTRACT"
    uv_map_invert_y.inputs[0].default_value = 1.0
    node_tree.links.new(uv_map_separate.outputs[1], uv_map_invert_y.inputs[1])
    uv_map_invert_y.location = decal_uv_map.location + VEC_HORIZ - VEC_VERT
    uv_map_invert_y.hide = True

    uv_map_combine: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    uv_map_combine.location = decal_uv_map.location + VEC_HORIZ - (VEC_VERT * 2)
    uv_map_combine.hide = True
    node_tree.links.new(uv_map_separate.outputs[0], uv_map_combine.inputs[0])
    node_tree.links.new(uv_map_invert_y.outputs[0], uv_map_combine.inputs[1])

    subtract_vec_node_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    subtract_vec_node_1.location = zw_mult_1.location + VEC_HORIZ
    subtract_vec_node_1.operation = "SUBTRACT"
    subtract_vec_node_1.hide = True
    node_tree.links.new(uv_map_combine.outputs[0], subtract_vec_node_1.inputs[0])
    node_tree.links.new(zw_mult_1.outputs[0], subtract_vec_node_1.inputs[1])

    # r9.xy
    divide_vec_node_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    divide_vec_node_1.location = subtract_vec_node_1.location + VEC_HORIZ
    divide_vec_node_1.operation = "DIVIDE"
    divide_vec_node_1.hide = True
    node_tree.links.new(subtract_vec_node_1.outputs[0], divide_vec_node_1.inputs[0])
    node_tree.links.new(zw_max_node_1.outputs[0], divide_vec_node_1.inputs[1])

    add_vec_node_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    add_vec_node_1.location = subtract_vec_node_1.location + VEC_HORIZ - VEC_VERT
    add_vec_node_1.operation = "ADD"
    add_vec_node_1.inputs[1].default_value = (-0.5, -0.5, 0.0)
    node_tree.links.new(divide_vec_node_1.outputs[0], add_vec_node_1.inputs[0])
    add_vec_node_1.hide = True

    # r14.yz
    mult_vec_node_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    mult_vec_node_1.location = add_vec_node_1.location + VEC_HORIZ - VEC_VERT
    mult_vec_node_1.operation = "MULTIPLY"
    mult_vec_node_1.hide = True
    node_tree.links.new(add_vec_node_1.outputs[0], mult_vec_node_1.inputs[0])
    node_tree.links.new(combine_r9_zw_1.outputs[0], mult_vec_node_1.inputs[1])

    mult_add_vec_node_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    mult_add_vec_node_1.location = divide_vec_node_1.location + VEC_HORIZ
    mult_add_vec_node_1.operation = "MULTIPLY_ADD"
    mult_add_vec_node_1.inputs[1].default_value = (2.0, 2.0, 0.0)
    mult_add_vec_node_1.inputs[2].default_value = (-1.0, -1.0, 0.0)
    node_tree.links.new(divide_vec_node_1.outputs[0], mult_add_vec_node_1.inputs[0])
    mult_add_vec_node_1.hide = True

    # r9.xy
    mult_vec_node_2: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    mult_vec_node_2.location = divide_vec_node_1.location + VEC_HORIZ - VEC_VERT
    mult_vec_node_2.operation = "MULTIPLY"
    mult_vec_node_2.hide = True
    node_tree.links.new(mult_add_vec_node_1.outputs[0], mult_vec_node_2.inputs[0])
    node_tree.links.new(combine_r9_zw_1.outputs[0], mult_vec_node_2.inputs[1])

    # r14.y
    dot_product_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    dot_product_1.location = mult_vec_node_2.location + VEC_HORIZ
    dot_product_1.operation = "DOT_PRODUCT"
    dot_product_1.hide = True
    node_tree.links.new(mult_vec_node_2.outputs[0], dot_product_1.inputs[0])
    node_tree.links.new(combine_r25_yz_1.outputs[0], dot_product_1.inputs[1])

    # r14.z
    dot_product_2: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    dot_product_2.location = mult_vec_node_2.location + VEC_HORIZ - VEC_VERT
    dot_product_2.operation = "DOT_PRODUCT"
    dot_product_2.hide = True
    node_tree.links.new(mult_vec_node_2.outputs[0], dot_product_2.inputs[0])
    node_tree.links.new(combine_r25_xy_1.outputs[0], dot_product_2.inputs[1])

    # r14.x
    mult_node_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    mult_node_1.operation = "MULTIPLY"
    mult_node_1.location = mult_vec_node_2.location + VEC_HORIZ + VEC_VERT
    mult_node_1.inputs[1].default_value = -1.0
    node_tree.links.new(dot_product_2.outputs[1], mult_node_1.inputs[0])
    mult_node_1.hide = True

    # r7.w
    less_than_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    less_than_1.operation = "LESS_THAN"
    less_than_1.location = mult_vec_node_2.location + VEC_HORIZ - (VEC_VERT * 6)
    less_than_1.hide = True
    less_than_1.inputs[1].default_value = 0.0
    node_tree.links.new(r20_xyzw[3], less_than_1.inputs[0])

    absolute_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    absolute_1.operation = "ABSOLUTE"
    absolute_1.location = mult_vec_node_2.location + VEC_HORIZ - (VEC_VERT * 7)
    absolute_1.hide = True
    node_tree.links.new(r20_xyzw[3], absolute_1.inputs[0])

    combine_r14_xy_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_r14_xy_1.location = mult_node_1.location + VEC_HORIZ - (VEC_VERT / 2)
    combine_r14_xy_1.hide = True
    node_tree.links.new(mult_node_1.outputs[0], combine_r14_xy_1.inputs[0])
    node_tree.links.new(dot_product_1.outputs[1], combine_r14_xy_1.inputs[1])
    
    combine_r14_yz_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_r14_yz_1.location = mult_node_1.location + VEC_HORIZ - (VEC_VERT * 1.5)
    combine_r14_yz_1.hide = True
    node_tree.links.new(dot_product_1.outputs[1], combine_r14_yz_1.inputs[0])
    node_tree.links.new(dot_product_2.outputs[1], combine_r14_yz_1.inputs[1])

    cape_lut_node_3: ShaderNodeGroup = node_tree.nodes.new("ShaderNodeGroup")
    cape_lut_node_3.node_tree = cape_lut_tree
    cape_lut_node_3.inputs[0].default_value = 6
    cape_lut_node_3.location = absolute_1.location + VEC_HORIZ - VEC_VERT

    # r9.xy
    mix_vec_node_1: ShaderNodeMix = node_tree.nodes.new("ShaderNodeMix")
    mix_vec_node_1.data_type = "VECTOR"
    mix_vec_node_1.location = combine_r14_xy_1.location + VEC_HORIZ - (VEC_VERT / 2)
    mix_vec_node_1.hide = True
    node_tree.links.new(less_than_1.outputs[0], mix_vec_node_1.inputs[0])
    node_tree.links.new(combine_r14_yz_1.outputs[0], mix_vec_node_1.inputs[4])
    node_tree.links.new(combine_r14_xy_1.outputs[0], mix_vec_node_1.inputs[5])

    max_node_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    max_node_1.location = combine_r14_xy_1.location + VEC_HORIZ - (VEC_VERT * 2)
    max_node_1.operation = "MAXIMUM"
    max_node_1.inputs[1].default_value = 0.0001
    max_node_1.hide = True
    node_tree.links.new(absolute_1.outputs[0], max_node_1.inputs[0])

    divide_node_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    divide_node_1.location = combine_r14_xy_1.location + VEC_HORIZ - (VEC_VERT * 3)
    divide_node_1.operation = "DIVIDE"
    divide_node_1.inputs[0].default_value = 1.0
    divide_node_1.hide = True
    node_tree.links.new(max_node_1.outputs[0], divide_node_1.inputs[1])

    separate_color_3: ShaderNodeSeparateColor = node_tree.nodes.new("ShaderNodeSeparateColor")
    separate_color_3.mode = "RGB"
    separate_color_3.location = cape_lut_node_3.location + VEC_HORIZ
    separate_color_3.hide = True
    node_tree.links.new(cape_lut_node_3.outputs[0], separate_color_3.inputs[0])
    r22_xyzw: List[NodeSocketFloat] = [*list(separate_color_3.outputs), cape_lut_node_3.outputs[1]]

    # r21.x
    dot_product_3: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    dot_product_3.location = cape_lut_node_3.location + VEC_HORIZ - (VEC_VERT * 3)
    dot_product_3.operation = "DOT_PRODUCT"
    dot_product_3.hide = True
    node_tree.links.new(mult_vec_node_1.outputs[0], dot_product_3.inputs[0])
    node_tree.links.new(combine_r25_yz_1.outputs[0], dot_product_3.inputs[1])

    # r21.y
    dot_product_4: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    dot_product_4.location = cape_lut_node_3.location + VEC_HORIZ - (VEC_VERT * 4)
    dot_product_4.operation = "DOT_PRODUCT"
    dot_product_4.hide = True
    node_tree.links.new(mult_vec_node_1.outputs[0], dot_product_4.inputs[0])
    node_tree.links.new(combine_r25_xy_1.outputs[0], dot_product_4.inputs[1])

    # r21.xy
    combine_r21_xy_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_r21_xy_1.location = dot_product_4.location + VEC_HORIZ - (VEC_VERT * 0.5)
    combine_r21_xy_1.hide = True
    node_tree.links.new(dot_product_3.outputs[1], combine_r21_xy_1.inputs[0])
    node_tree.links.new(dot_product_4.outputs[1], combine_r21_xy_1.inputs[1])

    # r9.xz
    scale_vec_node_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    scale_vec_node_1.location = divide_node_1.location + VEC_HORIZ + VEC_VERT
    scale_vec_node_1.operation = "SCALE"
    scale_vec_node_1.hide = True
    node_tree.links.new(mix_vec_node_1.outputs[1], scale_vec_node_1.inputs[0])
    node_tree.links.new(divide_node_1.outputs[0], scale_vec_node_1.inputs[3])

    separate_r9_xz_1: ShaderNodeSeparateXYZ = node_tree.nodes.new("ShaderNodeSeparateXYZ")
    separate_r9_xz_1.location = divide_node_1.location + VEC_HORIZ + (VEC_VERT * 2)
    separate_r9_xz_1.hide = True
    node_tree.links.new(scale_vec_node_1.outputs[0], separate_r9_xz_1.inputs[0])

    # r7.w
    less_than_2: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    less_than_2.operation = "LESS_THAN"
    less_than_2.location = separate_color_3.location + VEC_HORIZ
    less_than_2.hide = True
    less_than_2.inputs[1].default_value = 0.0
    node_tree.links.new(r22_xyzw[1], less_than_2.inputs[0])

    combine_r9_xy_2: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_r9_xy_2.location = less_than_2.location + VEC_HORIZ + (VEC_VERT * 2)
    combine_r9_xy_2.hide = True
    node_tree.links.new(separate_r9_xz_1.outputs[0], combine_r9_xy_2.inputs[0])
    node_tree.links.new(r22_xyzw[1], combine_r9_xy_2.inputs[1])

    negate_node_2: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    negate_node_2.operation = "MULTIPLY"
    node_tree.links.new(separate_r9_xz_1.outputs[0], negate_node_2.inputs[0])
    negate_node_2.inputs[1].default_value = -1.0
    negate_node_2.hide = True
    negate_node_2.location = less_than_2.location + VEC_HORIZ

    negate_node_3: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    negate_node_3.operation = "MULTIPLY"
    node_tree.links.new(r22_xyzw[1], negate_node_3.inputs[0])
    negate_node_3.inputs[1].default_value = -1.0
    negate_node_3.hide = True
    negate_node_3.location = less_than_2.location + VEC_HORIZ - VEC_VERT

    combine_r14_xy_2: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_r14_xy_2.location = less_than_2.location + VEC_HORIZ + VEC_VERT
    combine_r14_xy_2.hide = True
    node_tree.links.new(negate_node_2.outputs[0], combine_r14_xy_2.inputs[0])
    node_tree.links.new(negate_node_3.outputs[0], combine_r14_xy_2.inputs[1])

    # r11.zw
    divide_vec_node_2: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    divide_vec_node_2.location = combine_r21_xy_1.location + VEC_HORIZ
    divide_vec_node_2.operation = "DIVIDE"
    divide_vec_node_2.hide = True
    node_tree.links.new(combine_r21_xy_1.outputs[0], divide_vec_node_2.inputs[0])
    node_tree.links.new(zw_max_node_1.outputs[0], divide_vec_node_2.inputs[1])

    absolute_vec_1: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    absolute_vec_1.location = divide_vec_node_2.location + VEC_HORIZ
    absolute_vec_1.operation = "ABSOLUTE"
    absolute_vec_1.hide = True
    node_tree.links.new(divide_vec_node_2.outputs[0], absolute_vec_1.inputs[0])

    # r9.xy
    mix_vec_node_2: ShaderNodeMix = node_tree.nodes.new("ShaderNodeMix")
    mix_vec_node_2.data_type = "VECTOR"
    mix_vec_node_2.location = combine_r14_xy_2.location + VEC_HORIZ + (VEC_VERT / 2)
    mix_vec_node_2.hide = True
    node_tree.links.new(less_than_2.outputs[0], mix_vec_node_2.inputs[0])
    node_tree.links.new(combine_r9_xy_2.outputs[0], mix_vec_node_2.inputs[4])
    node_tree.links.new(combine_r14_xy_2.outputs[0], mix_vec_node_2.inputs[5])

    separate_r9_xy_1: ShaderNodeSeparateXYZ = node_tree.nodes.new("ShaderNodeSeparateXYZ")
    separate_r9_xy_1.location = combine_r14_xy_2.location + VEC_HORIZ + (VEC_VERT * 1.5)
    separate_r9_xy_1.hide = True
    node_tree.links.new(mix_vec_node_2.outputs[1], separate_r9_xy_1.inputs[0])

    combine_r9_xz_2: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_r9_xz_2.location = separate_r9_xy_1.location + VEC_VERT
    combine_r9_xz_2.hide = True
    node_tree.links.new(separate_r9_xy_1.outputs[0], combine_r9_xz_2.inputs[0])
    node_tree.links.new(separate_r9_xz_1.outputs[1], combine_r9_xz_2.inputs[1])

    combine_r20_xy_1: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_r20_xy_1.location = combine_r9_xz_2.location + VEC_VERT
    combine_r20_xy_1.hide = True
    node_tree.links.new(r20_xyzw[0], combine_r20_xy_1.inputs[0])
    node_tree.links.new(r20_xyzw[1], combine_r20_xy_1.inputs[1])

    # r9.xz
    add_vec_node_2: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    add_vec_node_2.location = combine_r20_xy_1.location + VEC_HORIZ - (VEC_VERT * 0.5)
    add_vec_node_2.operation = "ADD"
    add_vec_node_2.hide = True
    node_tree.links.new(combine_r20_xy_1.outputs[0], add_vec_node_2.inputs[0])
    node_tree.links.new(combine_r9_xz_2.outputs[0], add_vec_node_2.inputs[1])

    separate_r11_zw_1: ShaderNodeSeparateXYZ = node_tree.nodes.new("ShaderNodeSeparateXYZ")
    separate_r11_zw_1.location = absolute_vec_1.location + VEC_HORIZ
    separate_r11_zw_1.hide = True
    node_tree.links.new(absolute_vec_1.outputs[0], separate_r11_zw_1.inputs[0])

    less_than_3: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    less_than_3.operation = "LESS_THAN"
    less_than_3.location = separate_r11_zw_1.location + VEC_HORIZ + VEC_VERT
    less_than_3.hide = True
    less_than_3.inputs[1].default_value = 0.5
    node_tree.links.new(separate_r11_zw_1.outputs[0], less_than_3.inputs[0])

    less_than_4: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    less_than_4.operation = "LESS_THAN"
    less_than_4.location = separate_r11_zw_1.location + VEC_HORIZ - VEC_VERT
    less_than_4.hide = True
    less_than_4.inputs[1].default_value = 0.5
    node_tree.links.new(separate_r11_zw_1.outputs[1], less_than_4.inputs[0])

    # r11.zw
    mult_vec_node_3: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    mult_vec_node_3.location = add_vec_node_2.location + VEC_HORIZ
    mult_vec_node_3.operation = "MULTIPLY"
    mult_vec_node_3.hide = True
    node_tree.links.new(add_vec_node_2.outputs[0], mult_vec_node_3.inputs[0])
    mult_vec_node_3.inputs[1].default_value = (0.5, 0.5, 0.0)

    absolute_vec_2: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    absolute_vec_2.location = mult_vec_node_3.location + VEC_HORIZ + VEC_VERT
    absolute_vec_2.operation = "ABSOLUTE"
    absolute_vec_2.hide = True
    node_tree.links.new(mult_vec_node_3.outputs[0], absolute_vec_2.inputs[0])

    # r9.xz
    add_vec_node_3: ShaderNodeVectorMath = node_tree.nodes.new("ShaderNodeVectorMath")
    add_vec_node_3.location = mult_vec_node_3.location + VEC_HORIZ - VEC_VERT
    add_vec_node_3.operation = "ADD"
    add_vec_node_3.hide = True
    node_tree.links.new(mult_vec_node_3.outputs[0], add_vec_node_3.inputs[0])
    add_vec_node_3.inputs[1].default_value = (0.5, 0.5, 0.0)

    # r7.w
    mult_node_2: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    mult_node_2.operation = "MULTIPLY"
    mult_node_2.location = less_than_3.location + VEC_HORIZ - (VEC_VERT * 0.5)
    node_tree.links.new(less_than_3.outputs[0], mult_node_2.inputs[0])
    node_tree.links.new(less_than_4.outputs[0], mult_node_2.inputs[1])
    mult_node_2.hide = True

    separate_r11_zw_2: ShaderNodeSeparateXYZ = node_tree.nodes.new("ShaderNodeSeparateXYZ")
    separate_r11_zw_2.location = absolute_vec_2.location + VEC_HORIZ
    separate_r11_zw_2.hide = True
    node_tree.links.new(absolute_vec_2.outputs[0], separate_r11_zw_2.inputs[0])

    separate_r9_xz_2: ShaderNodeSeparateXYZ = node_tree.nodes.new("ShaderNodeSeparateXYZ")
    separate_r9_xz_2.location = add_vec_node_3.location + VEC_HORIZ - (VEC_VERT * 2)
    separate_r9_xz_2.hide = True
    node_tree.links.new(add_vec_node_3.outputs[0], separate_r9_xz_2.inputs[0])

    subtract_node_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    subtract_node_1.operation = "SUBTRACT"
    subtract_node_1.location = separate_r9_xz_2.location - VEC_VERT
    subtract_node_1.inputs[0].default_value = 1.0
    node_tree.links.new(separate_r9_xz_2.outputs[1], subtract_node_1.inputs[1])
    subtract_node_1.hide = True

    combine_decal_uv_coords: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_decal_uv_coords.location = subtract_node_1.location - VEC_VERT
    combine_decal_uv_coords.hide = True
    node_tree.links.new(separate_r9_xz_2.outputs[0], combine_decal_uv_coords.inputs[0])
    node_tree.links.new(subtract_node_1.outputs[0], combine_decal_uv_coords.inputs[1])

    # Image added when update images called
    decal_texture: ShaderNodeTexImage = node_tree.nodes.new("ShaderNodeTexImage")
    decal_texture.name = "Cape Decal"
    decal_texture.location = separate_r9_xz_2.location + VEC_HORIZ
    node_tree.links.new(combine_decal_uv_coords.outputs[0], decal_texture.inputs[0])

    # r11.z
    less_than_5: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    less_than_5.operation = "LESS_THAN"
    less_than_5.location = separate_r11_zw_2.location + VEC_HORIZ + VEC_VERT
    less_than_5.hide = True
    less_than_5.inputs[1].default_value = 0.5
    node_tree.links.new(separate_r11_zw_2.outputs[0], less_than_5.inputs[0])

    # r11.w
    less_than_6: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    less_than_6.operation = "LESS_THAN"
    less_than_6.location = separate_r11_zw_2.location + VEC_HORIZ - VEC_VERT
    less_than_6.hide = True
    less_than_6.inputs[1].default_value = 0.5
    node_tree.links.new(separate_r11_zw_2.outputs[1], less_than_6.inputs[0])

    # r7.w
    mult_node_3: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    mult_node_3.operation = "MULTIPLY"
    mult_node_3.location = less_than_5.location + VEC_HORIZ - (VEC_VERT * 0.5)
    node_tree.links.new(less_than_5.outputs[0], mult_node_3.inputs[0])
    node_tree.links.new(less_than_6.outputs[0], mult_node_3.inputs[1])
    mult_node_3.hide = True

    mult_node_4: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    mult_node_4.operation = "MULTIPLY"
    mult_node_4.location = mult_node_3.location + VEC_HORIZ
    node_tree.links.new(mult_node_3.outputs[0], mult_node_4.inputs[0])
    node_tree.links.new(mult_node_2.outputs[0], mult_node_4.inputs[1])
    mult_node_4.hide = True

    value_node: ShaderNodeValue = node_tree.nodes.new("ShaderNodeValue")
    value_node.outputs[0].default_value = 1.0
    value_node.location = decal_texture.location + VEC_HORIZ * 2 - VEC_VERT * 2

    if "Group Output" not in node_tree.nodes:
        node_tree.nodes.new("NodeGroupOutput").name = "Group Output"
    group_output: NodeGroupOutput = node_tree.nodes["Group Output"]
    group_output.location = decal_texture.location + VEC_HORIZ * 3

    node_tree.links.new(decal_texture.outputs[0], group_output.inputs[0])
    node_tree.links.new(decal_texture.outputs[1], group_output.inputs[1])
    node_tree.links.new(mult_node_4.outputs[0], group_output.inputs[2])
    node_tree.links.new(value_node.outputs[0], group_output.inputs[3])

    return node_tree

def create_cape_lut_tree() -> NodeTree:
    node_tree: NodeTree = bpy.data.node_groups.new("Cape LUT", "ShaderNodeTree")
    node_tree.interface.new_socket("Cape X Coord", socket_type="NodeSocketInt")
    node_tree.interface.new_socket("Color", in_out="OUTPUT", socket_type="NodeSocketColor")
    node_tree.interface.new_socket("Alpha", in_out="OUTPUT", socket_type="NodeSocketFloat")


    VEC_HORIZ = Vector((300, 0))
    VEC_VERT = Vector((0, 50))

    if "Group Input" not in node_tree.nodes:
        node_tree.nodes.new("NodeGroupInput").name = "Group Input"
    group_input: NodeGroupInput = node_tree.nodes["Group Input"]

    geometry: ShaderNodeNewGeometry = node_tree.nodes.new("ShaderNodeNewGeometry")
    geometry.location = group_input.location - VEC_VERT * 3
    geometry.hide = True

    add_node_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    add_node_1.operation = "ADD"
    add_node_1.location = group_input.location + VEC_HORIZ
    node_tree.links.new(group_input.outputs[0], add_node_1.inputs[0])
    add_node_1.inputs[1].default_value = 0.5
    add_node_1.hide = True

    divide_node_1: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    divide_node_1.operation = "DIVIDE"
    divide_node_1.location = add_node_1.location - VEC_VERT
    node_tree.links.new(add_node_1.outputs[0], divide_node_1.inputs[0])
    divide_node_1.inputs[1].default_value = 16.0
    divide_node_1.hide = True

    front_face_node: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    front_face_node.operation = "SUBTRACT"
    front_face_node.location = geometry.location + VEC_HORIZ
    front_face_node.inputs[0].default_value = 1.0
    node_tree.links.new(geometry.outputs[6], front_face_node.inputs[1])
    front_face_node.hide = True

    clamp_node: ShaderNodeClamp = node_tree.nodes.new("ShaderNodeClamp")
    clamp_node.inputs[1].default_value = 0.1
    clamp_node.inputs[2].default_value = 0.9
    clamp_node.hide = True
    node_tree.links.new(front_face_node.outputs[0], clamp_node.inputs[0])
    clamp_node.location = front_face_node.location - VEC_VERT

    combine_uv: ShaderNodeCombineXYZ = node_tree.nodes.new("ShaderNodeCombineXYZ")
    combine_uv.location = divide_node_1.location + VEC_HORIZ
    combine_uv.hide = True
    node_tree.links.new(divide_node_1.outputs[0], combine_uv.inputs[0])
    node_tree.links.new(clamp_node.outputs[0], combine_uv.inputs[1])

    lut_node: ShaderNodeTexImage = node_tree.nodes.new("ShaderNodeTexImage")
    lut_node.name = "Cape LUT Texture"
    lut_node.extension = "CLIP"
    lut_node.interpolation = "Closest"
    node_tree.links.new(combine_uv.outputs[0], lut_node.inputs[0])
    lut_node.location = combine_uv.location + VEC_HORIZ

    alpha_mult_node: ShaderNodeMath = node_tree.nodes.new("ShaderNodeMath")
    alpha_mult_node.operation = "MULTIPLY"
    alpha_mult_node.location = lut_node.location + VEC_HORIZ
    node_tree.links.new(lut_node.outputs[1], alpha_mult_node.inputs[0])
    node_tree.links.new(front_face_node.outputs[0], alpha_mult_node.inputs[1])
    alpha_mult_node.hide = True

    if "Group Output" not in node_tree.nodes:
        node_tree.nodes.new("NodeGroupOutput").name = "Group Output"
    group_output: NodeGroupOutput = node_tree.nodes["Group Output"]
    group_output.location = lut_node.location + VEC_HORIZ
    node_tree.links.new(lut_node.outputs[0], group_output.inputs[0])
    node_tree.links.new(alpha_mult_node.outputs[0], group_output.inputs[1])

    return node_tree


def update_images(HD2_Shader: NodeTree, material: Material):
    node_tree = material.node_tree
    cape_decal_node: ShaderNodeGroup = HD2_Shader.nodes["Cape Decal Template"]
    cape_decal_template: NodeTree = cape_decal_node.node_tree
    cape_layer_node: ShaderNodeGroup = cape_decal_template.nodes["Layer 1"]
    cape_layer_template: NodeTree = cape_layer_node.node_tree
    cape_decal_texture: ShaderNodeTexImage = cape_layer_template.nodes["Cape Decal"]
    cape_lut_node: ShaderNodeGroup = cape_layer_template.nodes["Cape LUT"]
    cape_lut_template: NodeTree = cape_lut_node.node_tree
    cape_lut_texture: ShaderNodeTexImage = cape_lut_template.nodes["Cape LUT Texture"]
    
    pattern_mask_array_02 = HD2_Shader.nodes["Pattern Mask Array 02"]
    id_mask_array_02 = HD2_Shader.nodes["ID Mask Array 02"]
    pattern_luts = [HD2_Shader.nodes[f"pattern_lut {i:02d}"] for i in range(2, 4)] + [HD2_Shader.nodes["pattern_lut"]]
    primary_material_luts = [HD2_Shader.nodes[f"Primary Material LUT_{i:02d}"] for i in range(23)]

    #Get Pattern LUT from external texture node for shader repetition
    pattern_lut_image = node_tree.nodes["Pattern LUT Texture"].image
    
    #Get Primary Material LUT from external texture node for shader repetition
    primary_lut_image = node_tree.nodes["Primary Material LUT Texture"].image
    primary_lut_image.colorspace_settings.name = "Non-Color"
    primary_lut_image.alpha_mode = "CHANNEL_PACKED"

    #Get Cape LUT from external texture node for shader repetition
    cape_lut_texture.image = node_tree.nodes["Cape LUT Texture"].image
    cape_lut_texture.image.colorspace_settings.name = "Non-Color"
    cape_lut_texture.image.alpha_mode = "CHANNEL_PACKED"

    #Get Decal texture from external texture node
    cape_decal_texture.image = node_tree.nodes["Decal Texture"].image

    #Get Pattern Mask from external texture node
    pattern_mask_array_02.image = node_tree.nodes["Pattern Mask Array"].image

    #Get ID Mask Array from external texture node for shader repetition
    id_mask_array_02.image = node_tree.nodes["ID Mask Array Texture"].image

    for node in pattern_luts:
        node.image = pattern_lut_image

    for node in primary_material_luts:
        node.image = primary_lut_image

def add_bake_uvs(obj: Object):
    if "UVs for Baking" in obj.data.uv_layers:
        return
    bake_uvs = obj.data.uv_layers.new(name="UVs for Baking")
    for loop_index in range(len(obj.data.uv_layers["UVMap"].data)):
        bake_uvs.data[loop_index].uv = obj.data.uv_layers["UVMap"].data[loop_index].uv
    
    for uv_data in bake_uvs.data:
        if uv_data.uv.x > 1.0:
            uv_data.uv.x %= 1
        if uv_data.uv.y > 1.0:
            uv_data.uv.y %= 1
        if uv_data.uv.x < 0.0:
            uv_data.uv.x = uv_data.uv.x % 1
        if uv_data.uv.y < 0.0:
            uv_data.uv.y = uv_data.uv.y % 1

    obj.data.update()

def update_array_uvs(material: Material):
    #set UVs for arrays
    material.node_tree.nodes['ID Mask UV'].inputs[0].default_value = (0.000)
    material.node_tree.nodes['Pattern Mask UV'].inputs[0].default_value = (0.000)
    try:
        IDMaskArraySizeX = material.node_tree.nodes['ID Mask Array Texture'].inputs[0].node.image.size[0]
        IDMaskArraySizeY = material.node_tree.nodes['ID Mask Array Texture'].inputs[0].node.image.size[1]
        if (IDMaskArraySizeY/IDMaskArraySizeX) >= 2.0:
            material.node_tree.nodes['ID Mask UV'].inputs[0].default_value = (1.000)
    except:
        pass
    
    try:
        PatternMaskArraySizeX = material.node_tree.nodes['Pattern Mask Array'].inputs[0].node.image.size[0]
        PatternMaskArraySizeY = material.node_tree.nodes['Pattern Mask Array'].inputs[0].node.image.size[1]
        if (PatternMaskArraySizeY/PatternMaskArraySizeX) >= 2.0:
            material.node_tree.nodes['Pattern Mask UV'].inputs[0].default_value = (1.000)
    except:
        pass

def connect_input_links(hd2_shader: NodeTree):
    group_input_1 = hd2_shader.nodes["Group Input"]
    #id_mask_array_02_off_on.Value -> mix_018.Factor
    hd2_shader.links.new(group_input_1.outputs[0], hd2_shader.nodes["Mix.018"].inputs[0])
    #id_mask_array_02_off_on.Value -> mix_019.Factor
    hd2_shader.links.new(group_input_1.outputs[0], hd2_shader.nodes["Mix.019"].inputs[0])
    #pattern_mask_array_02_off_on.Value -> mix_030.Factor
    hd2_shader.links.new(group_input_1.outputs[1], hd2_shader.nodes["Mix.030"].inputs[0])
    hd2_shader.links.new(group_input_1.outputs[2], hd2_shader.nodes["Separate XYZ.025"].inputs[0])
    #hd2_shader.links.new(group_input_1.outputs[3], hd2_shader.nodes["Math.238"].inputs[0])
    hd2_shader.links.new(group_input_1.outputs[4], hd2_shader.nodes["Separate XYZ.059"].inputs[0])
    #group_input_1.Decal Texture -> vector_math_075.Vector
    #hd2_shader.links.new(group_input_1.outputs[5], hd2_shader.nodes["Vector Math.075"].inputs[1])
    #group_input_1.Decal Texture(alpha) -> math_195.Value
    #hd2_shader.links.new(group_input_1.outputs[6], hd2_shader.nodes["Math.195"].inputs[1])
    #group_input_1.Normal Map -> vector_math_021.Vector
    hd2_shader.links.new(group_input_1.outputs[10], hd2_shader.nodes["Vector Math.021"].inputs[1])
    #group_input_1.Normal Map -> separate_xyz_051.Vector
    hd2_shader.links.new(group_input_1.outputs[10], hd2_shader.nodes["Separate XYZ.051"].inputs[0])
    #group_input_1.Normal Map -> separate_xyz_022.Vector
    hd2_shader.links.new(group_input_1.outputs[10], hd2_shader.nodes["Separate XYZ.022"].inputs[0])
    #group_input_1.Normal Map -> separate_xyz_028.Vector
    hd2_shader.links.new(group_input_1.outputs[10], hd2_shader.nodes["Separate XYZ.028"].inputs[0])
    #group_input_1.Normal Map_Alpha -> math_003.Value
    hd2_shader.links.new(group_input_1.outputs[11], hd2_shader.nodes["Math.003"].inputs[1])
    #group_input_1.Normal Map_Alpha -> combine_xyz_023.X
    hd2_shader.links.new(group_input_1.outputs[11], hd2_shader.nodes["Combine XYZ.023"].inputs[0])
    #group_input_1.detail_tile_factor_mult -> math_044.Value
    hd2_shader.links.new(group_input_1.outputs[12], hd2_shader.nodes["Math.044"].inputs[0])
    #group_input_1.detail_tile_factor_mult -> math_079.Value
    hd2_shader.links.new(group_input_1.outputs[12], hd2_shader.nodes["Math.079"].inputs[0])
    #group_input_1.detail_tile_factor_mult -> math_019.Value
    hd2_shader.links.new(group_input_1.outputs[12], hd2_shader.nodes["Math.019"].inputs[0])
    #group_input_1.detail_tile_factor_mult -> math_152.Value
    hd2_shader.links.new(group_input_1.outputs[12], hd2_shader.nodes["Math.152"].inputs[0])

def update_slot_defaults(hd2_shader: NodeTree, material: Material):
    PrimaryMaterialLUTSizeX = material.node_tree.nodes['Primary Material LUT Texture'].inputs[0].node.image.size[0]
    PrimaryMaterialLUTSizeY = material.node_tree.nodes['Primary Material LUT Texture'].inputs[0].node.image.size[1]
    Slot1and2: ShaderNodeMix = hd2_shader.nodes["Mix.028"]
    Slot1and2.inputs[2].default_value = (1-(0.5/PrimaryMaterialLUTSizeY))
    Slot1and2.inputs[3].default_value = (1-(0.5/PrimaryMaterialLUTSizeY*3))

    Slot3: ShaderNodeMix = hd2_shader.nodes["Mix.023"]
    Slot3.inputs[3].default_value = (1-(0.5/PrimaryMaterialLUTSizeY*5))
    if PrimaryMaterialLUTSizeY < 2.1:
        Slot3.mute = True
    
    Slot4: ShaderNodeMix = hd2_shader.nodes["Mix.025"]
    Slot4.inputs[3].default_value = (1-(0.5/PrimaryMaterialLUTSizeY*7))
    if PrimaryMaterialLUTSizeY < 3.1:
        Slot4.mute = True
    
    Slot5: ShaderNodeMix = hd2_shader.nodes["Mix.024"]
    Slot5.inputs[3].default_value = (1-(0.5/PrimaryMaterialLUTSizeY*9))
    if PrimaryMaterialLUTSizeY < 4.1:
        Slot5.mute = True

    Slot6: ShaderNodeMix = hd2_shader.nodes["Mix.027"]
    Slot6.inputs[3].default_value = (1-(0.5/PrimaryMaterialLUTSizeY*11))
    if PrimaryMaterialLUTSizeY < 5.1:
        Slot6.mute = True
    
    Slot7: ShaderNodeMix = hd2_shader.nodes["Mix.026"]
    Slot7.inputs[3].default_value = (1-(0.5/PrimaryMaterialLUTSizeY*13))
    if PrimaryMaterialLUTSizeY < 6.1:
        Slot7.mute = True

class CreateShader(bpy.types.Operator):
    bl_label = ("(Re)Build Shader")
    bl_idname = "node.create_operator"

    def execute(self, context):
        custom_node_name = ("HD2 Shader Template File Do Not Name Anything Else This Name")
        GroupNode = create_HD2_Shader(self, context, custom_node_name)

        return {'FINISHED'}

class UpdateShader(bpy.types.Operator):
    bl_label = ("Update Images")
    bl_idname = "node.update_operator"

    def execute(self, context):
        node_tree = bpy.context.active_object.active_material.node_tree.nodes["HD2 Shader Template"].node_tree
        #GroupNode = bpy.data.node_groups[custom_node_name]
        update_images(node_tree, bpy.context.active_object.active_material)
        add_bake_uvs(bpy.context.active_object)
        update_slot_defaults(node_tree, bpy.context.active_object.active_material)
        return {'FINISHED'}

def register():
    bpy.utils.register_class(NODE_PT_MAINPANEL)
    bpy.utils.register_class(CreateShader)
    bpy.utils.register_class(UpdateShader)


def unregister():
    bpy.utils.unregister_class(NODE_PT_MAINPANEL)
    bpy.utils.unregister_class(CreateShader)
    bpy.utils.unregister_class(UpdateShader)

if __name__ == "__main__":
    register()