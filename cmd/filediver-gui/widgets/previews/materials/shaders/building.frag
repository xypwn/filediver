#version 430 core

#include "lighting.frag"

out vec4 fragColor;

in vec3 fragInstPosition;
in vec2 fragUV0;
in vec2 fragUV1;
in mat4 instanceModel;

uniform sampler2D texture_lut;
uniform sampler2D decal_sheet;
uniform sampler2D Detail_Data;
uniform sampler2D NAC;
uniform sampler2DArray surface_data_array;

layout(shared, binding = 0) uniform BuildingSettingsBlock {
    vec3 decal_color_tint;
    float decal_metallic;
    float decal_metallic_blend;
    float decal_roughness;
    float decal_roughness_blend;
    float decal_wear;
    float fd_decals_enabled;
    float fd_decal_only;
    
    float global_detail_tile;
    float global_surface_tile;
    vec3 wear_colour;
    float wear_metallic_shift;
    float wear_multiply;
    float wear_power;
    float wear_roughness;
    float wear_roughness_blend;
    float wear_shift;

    float material_1_metallic;
    float material_1_roughness_base;
    vec4 material_1_roughness_build_up;
    float material_1_surface;
    float material_1_surface_normal;
    float material_1_surface_roughness;
    float material_1_surface_value;
    float material_01_tile_multiplier;
    vec3 material_1_wear_cavity_edge;

    float material_2_metallic;
    float material_2_roughness_base;
    vec4 material_2_roughness_build_up;
    float material_2_surface;
    float material_2_surface_normal;
    float material_2_surface_roughness;
    float material_2_surface_value;
    vec3 material_2_wear_cavity_edge;
    float material_02_tile_multiplier;

    float material_3_metallic;
    float material_3_roughness_base;
    vec4 material_3_roughness_build_up;
    float material_3_surface;
    float material_3_surface_normal;
    float material_3_surface_roughness;
    float material_3_surface_value;
    float material_03_tile_multiplier;
    vec3 material_3_wear_cavity_edge;

    float material_4_metallic;
    float material_4_roughness_base;
    vec4 material_4_roughness_build_up;
    float material_4_surface;
    float material_4_surface_normal;
    float material_4_surface_roughness;
    float material_4_surface_value;
    float material_04_tile_multiplier;
    vec3 material_4_wear_cavity_edge;

    float material_5_metallic;
    float material_5_roughness_base;
    vec4 material_5_roughness_build_up;
    float material_5_surface;
    float material_5_surface_normal;
    float material_5_surface_roughness;
    float material_5_surface_value;
    float material_05_tile_multiplier;
    vec3 material_5_wear_cavity_edge;

    float material_6_metallic;
    float material_6_roughness_base;
    vec4 material_6_roughness_build_up;
    float material_6_surface;
    float material_6_surface_normal;
    float material_6_surface_roughness;
    float material_6_surface_value;
    float material_06_tile_multiplier;
    vec3 material_6_wear_cavity_edge;

    float material_7_metallic;
    float material_7_roughness_base;
    vec4 material_7_roughness_build_up;
    float material_7_surface;
    float material_7_surface_normal;
    float material_7_surface_roughness;
    float material_7_surface_value;
    vec3 material_7_wear_cavity_edge;
    float material_07_tile_multiplier;

    float material_8_metallic;
    float material_8_roughness_base;
    vec4 material_8_roughness_build_up;
    float material_8_surface;
    float material_8_surface_normal;
    float material_8_surface_roughness;
    float material_8_surface_value;
    float material_08_tile_multiplier;
    vec3 material_8_wear_cavity_edge;
};

layout(shared, binding = 1) uniform DebugBlock {
    uint debug;
};

layout(shared, binding = 2) uniform PostProcessBlock {
    float contrast;
    float saturation;
};

in mat3 dbg_fragTBN;
vec2 getTriplanarUVs() {
    vec3 normal = dbg_fragTBN[2];
    vec2 triplanarUVs;
    vec4 temp;
    if (abs(normal.z) < abs(normal.x)) {
        temp.x = fragInstPosition.y;
        temp.y = fragInstPosition.z;
    } else {
        temp.x = fragInstPosition.x;
        temp.y = fragInstPosition.y;
    }
    if (abs(normal.z) < abs(normal.y)) {
        temp.z = fragInstPosition.x;
        temp.w = fragInstPosition.z;
    } else {
        temp.z = fragInstPosition.x;
        temp.w = fragInstPosition.y;
    }
    if (abs(normal.y) < abs(normal.x)) {
        triplanarUVs.x = temp.x;
        triplanarUVs.y = temp.y;
    } else {
        triplanarUVs.x = temp.z;
        triplanarUVs.y = temp.w;
    }
    return triplanarUVs;
}

vec3 getTriplanarNormal(vec3 tangent_space_normal) {
    vec3 normal = dbg_fragTBN[2];
    vec4 temp;
    temp.w = 1.0 - tangent_space_normal.x;
    vec3 result;

    if (abs(normal.z) < abs(normal.x)) {
        temp.x = tangent_space_normal.z;
        temp.y = temp.w;
        temp.z = tangent_space_normal.y;
    } else {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.y;
        temp.z = tangent_space_normal.z;
    }
    result.xyz = temp.xyz;
    if (abs(normal.z) < abs(normal.y)) {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.z;
        temp.z = tangent_space_normal.y;
    } else {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.y;
        temp.z = tangent_space_normal.z;
    }
    if (abs(normal.y) < abs(normal.x)) {
        result.x = result.x;
        result.y = result.y;
        result.z = result.z;
    } else {
        result.x = temp.x;
        result.y = temp.y;
        result.z = temp.z;
    }

    // temp.xyz = result * 2.0 - 1.0 + normal;
    // result.x = dot(temp.xyz, dbg_fragTBN[0]);
    // result.y = dot(temp.xyz, dbg_fragTBN[1]);
    // result.z = dot(temp.xyz, dbg_fragTBN[2]);
    return result * 2.0 - 1.0;
}

vec3 materialColor[8] = {
    {0, 0, 0},
    {1, 0, 0},
    {0, 1, 0},
    {0, 0, 1},
    {1, 1, 0},
    {1, 0, 1},
    {0, 1, 1},
    {1, 1, 1},
};

void main() {
    uint material = uint(clamp(7.0 - (floor(fragUV0.y) + 3), 0.0, 7.0));
    float metallic, roughness_base, surface, surface_normal, surface_roughness, surface_value, tile_multiplier;
    vec4 roughness_build_up;
    vec3 wear_cavity_edge;
    switch (material) {
    case 0:
        metallic = material_1_metallic;
        roughness_base = material_1_roughness_base;
        surface = material_1_surface;
        surface_normal = material_1_surface_normal;
        surface_roughness = material_1_surface_roughness;
        surface_value = material_1_surface_value;
        wear_cavity_edge = material_1_wear_cavity_edge;
        roughness_build_up = material_1_roughness_build_up;
        tile_multiplier = material_01_tile_multiplier;
        break;
    case 1:
        metallic = material_2_metallic;
        roughness_base = material_2_roughness_base;
        surface = material_2_surface;
        surface_normal = material_2_surface_normal;
        surface_roughness = material_2_surface_roughness;
        surface_value = material_2_surface_value;
        wear_cavity_edge = material_2_wear_cavity_edge;
        roughness_build_up = material_2_roughness_build_up;
        tile_multiplier = material_02_tile_multiplier;
        break;
    case 2:
        metallic = material_3_metallic;
        roughness_base = material_3_roughness_base;
        surface = material_3_surface;
        surface_normal = material_3_surface_normal;
        surface_roughness = material_3_surface_roughness;
        surface_value = material_3_surface_value;
        wear_cavity_edge = material_3_wear_cavity_edge;
        roughness_build_up = material_3_roughness_build_up;
        tile_multiplier = material_03_tile_multiplier;
        break;
    case 3:
        metallic = material_4_metallic;
        roughness_base = material_4_roughness_base;
        surface = material_4_surface;
        surface_normal = material_4_surface_normal;
        surface_roughness = material_4_surface_roughness;
        surface_value = material_4_surface_value;
        wear_cavity_edge = material_4_wear_cavity_edge;
        roughness_build_up = material_4_roughness_build_up;
        tile_multiplier = material_04_tile_multiplier;
        break;
    case 4:
        metallic = material_5_metallic;
        roughness_base = material_5_roughness_base;
        surface = material_5_surface;
        surface_normal = material_5_surface_normal;
        surface_roughness = material_5_surface_roughness;
        surface_value = material_5_surface_value;
        wear_cavity_edge = material_5_wear_cavity_edge;
        roughness_build_up = material_5_roughness_build_up;
        tile_multiplier = material_05_tile_multiplier;
        break;
    case 5:
        metallic = material_6_metallic;
        roughness_base = material_6_roughness_base;
        surface = material_6_surface;
        surface_normal = material_6_surface_normal;
        surface_roughness = material_6_surface_roughness;
        surface_value = material_6_surface_value;
        wear_cavity_edge = material_6_wear_cavity_edge;
        roughness_build_up = material_6_roughness_build_up;
        tile_multiplier = material_06_tile_multiplier;
        break;
    case 6:
        metallic = material_7_metallic;
        roughness_base = material_7_roughness_base;
        surface = material_7_surface;
        surface_normal = material_7_surface_normal;
        surface_roughness = material_7_surface_roughness;
        surface_value = material_7_surface_value;
        wear_cavity_edge = material_7_wear_cavity_edge;
        roughness_build_up = material_7_roughness_build_up;
        tile_multiplier = material_07_tile_multiplier;
        break;
    case 7:
        metallic = material_8_metallic;
        roughness_base = material_8_roughness_base;
        surface = material_8_surface;
        surface_normal = material_8_surface_normal;
        surface_roughness = material_8_surface_roughness;
        surface_value = material_8_surface_value;
        wear_cavity_edge = material_8_wear_cavity_edge;
        roughness_build_up = material_8_roughness_build_up;
        tile_multiplier = material_08_tile_multiplier;
        break;
    }

    // Get the object location and use it to select a base color for each material
    vec2 lut_uvs = vec2((instanceModel[3].x + instanceModel[3].z) / 150.0, float(material) / 8.0);
    vec4 base_color = textureLod(texture_lut, lut_uvs, 0.0);
    float surface_layer = roundEven(surface) - 1.0;

    vec2 triplanar_uvs = getTriplanarUVs() * tile_multiplier;

    float surface_data_value, surface_data_roughness;
    vec3 surface_data_normal;
    if (surface_layer >= 0.0) {
        vec4 temp = texture(surface_data_array, vec3(triplanar_uvs / global_surface_tile, surface_layer));
        surface_data_value = temp.z;
        surface_data_roughness = temp.w;
        surface_data_normal.xy = temp.xy;
    } else {
        surface_data_value = 0.5;
        surface_data_roughness = 0.5;
        surface_data_normal.xy = vec2(0.5);
    }
    surface_data_normal.z = 0.5;

    surface_data_normal = getTriplanarNormal(surface_data_normal);
    surface_data_normal.xy *= surface_normal;
    surface_data_normal.z = reconstructNormalZ(surface_data_normal.xy);

    float surface_value_orig = surface_value;
    surface_value = surface_value * (surface_data_value - 0.5);
    float surface_roughness_orig = surface_roughness;
    surface_roughness = surface_roughness * (surface_data_roughness - 0.5);

    vec4 detail_sample = texture(Detail_Data, triplanar_uvs / global_detail_tile) * 2.0 - 1.0;
    vec4 nac_sample = texture(NAC, fragUV0);

    vec3 normal = vec3(nac_sample.xy * 2.0 - 1.0, 0.0);
    normal.z = sqrt(abs(1.0 - normal.x * normal.x - normal.y * normal.y));
    vec3 nac_normal = normal;
    normal = reorientNormalMaps(normal, surface_data_normal);

    float curvature = clamp(nac_sample.w * 2.0 - 1.0, 0.0, 1.0);
    float negative_curvature = clamp((1.0 - nac_sample.w) * 2.0 - 1.0, 0.0, 1.0);

    vec3 cavity_weights = 1.0 - wear_cavity_edge.x * (roughness_build_up.xyw * detail_sample.xyw);
    vec3 before_cavity_weights = base_color.xyz;
    base_color.xyz = base_color.xyz * cavity_weights.x * cavity_weights.y * cavity_weights.x * cavity_weights.z;// * 0.75;
    vec3 before_negative_curvature = base_color.xyz;
    float negative_curvature_color_mult = (wear_cavity_edge.y * negative_curvature + 1.0);
    base_color.xyz = base_color.xyz * negative_curvature_color_mult;
    vec3 before_positive_curvature_and_sv = base_color.xyz;
    float positive_curvature_color_mult = (wear_cavity_edge.z * curvature + 1.0);
    base_color.xyz = clamp(base_color.xyz * positive_curvature_color_mult + (2 * surface_value), 0.0, 1.0);
    vec3 before_wear_combined = base_color.xyz;

    float roughness = (detail_sample.x * roughness_build_up.x) + (detail_sample.y * roughness_build_up.y) + (detail_sample.z * roughness_build_up.z) + (detail_sample.w * roughness_build_up.w) + roughness_base;
    roughness = clamp(surface_roughness * 2.0 + roughness, 0.0, 1.0);

    vec4 decal_sample;
    float decal_wear_roughness;
    vec2 decal_uvs = fd_decal_only != 0.0 ? fragUV0 : fragUV1;
    if (fd_decals_enabled != 0.0) {
        decal_sample = texture(decal_sheet, decal_uvs);
        float temp = clamp(floor(decal_uvs.x + 1.0), 0.0, 1.0);
        temp = temp - clamp(fragUV0.x, 0.0, 1.0);
        decal_wear_roughness = decal_wear * max(roughness * 2.0 - 1.0, 0.0);
        decal_wear_roughness = clamp(decal_sample.w * temp - decal_wear_roughness, 0.0, 1.0);
        float inv_decal_wear_roughness = 1.0 - decal_wear_roughness;
        
        decal_sample.xyz = decal_sample.xyz * decal_color_tint * decal_wear_roughness;
        base_color.xyz = base_color.xyz * inv_decal_wear_roughness + decal_sample.xyz;
        
        float metallic_blend = decal_metallic_blend * decal_wear_roughness;
        metallic = metallic_blend * (decal_metallic - metallic) + metallic;

        normal = normalize(decal_wear_roughness * (vec3(0.0, 0.0, 1.0) - normal) + normal);

        roughness = decal_wear_roughness * decal_roughness_blend * (decal_roughness - roughness) + roughness;
    }

    float wear_combined = clamp(exp2(log2(roughness) * wear_power) * wear_multiply + wear_shift, 0.0, 1.0) * base_color.w;
    base_color.xyz = sqrt(clamp(wear_combined * (wear_colour - base_color.xyz) + base_color.xyz, 0.0, 1.0));

    roughness = clamp((wear_combined * wear_roughness_blend) * (wear_roughness - roughness) + roughness, 0.0, 1.0) * 0.955 + 0.045;
    metallic = clamp(wear_combined * wear_metallic_shift + metallic, 0.0, 1.0); // * 0.5 + 0.5;
    float ao = clamp(nac_sample.z, 0.0, 1.0);

    switch (debug) {
    case 1:
        fragColor = vec4(vec3(roughness), 1.0);
        break;
    case 2:
        fragColor = vec4(vec3(metallic), 1.0);
        break;
    case 3:
        fragColor = vec4(vec3(ao), 1.0);
        break;
    case 4:
        fragColor = vec4(normal, 1.0);
        break;
    case 5:
        fragColor = vec4(mod(getTriplanarUVs(), 1.0), 0.0, 1.0);
        break;
    case 6:
        fragColor = vec4(getTriplanarNormal(surface_data_normal), 1.0);
        break;
    case 7:
        fragColor = vec4(dbg_fragTBN[2], 1.0);
        break;
    case 8:
        fragColor = vec4(lut_uvs, 0.0, 1.0);
        break;
    case 9:
        fragColor = vec4(vec3(decal_wear_roughness), 1.0);
        break;
    case 10:
        fragColor = vec4(decal_sample.xyz, 1.0);
        break;
    case 11:
        fragColor = vec4(texture(decal_sheet, decal_uvs).xyz, 1.0);
        break;
    case 12:
        fragColor = vec4(decal_uvs, 0.0, 1.0);
        break;
    case 13:
        fragColor = vec4(vec3(clamp(floor(decal_uvs.x + 1.0), 0.0, 1.0) - clamp(fragUV0.x, 0.0, 1.0)), 1.0);
        break;
    case 14:
        fragColor = vec4(vec3(clamp(floor(decal_uvs.x + 1.0), 0.0, 1.0)), 1.0);
        break;
    case 15:
        fragColor = vec4(vec3(clamp(fragUV0.x, 0.0, 1.0)), 1.0);
        break;
    default:
        fragColor = calculateLighting(base_color, roughness, metallic, ao, normal, max(saturation + 1.5, 0.0), max(contrast + 1.2, 0.0));
        break;
    }
}