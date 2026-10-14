#version 430 core

#include "lighting.frag"

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV0;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

uniform sampler2D ao_map;
uniform sampler2D color_map;
uniform sampler2D emissive_map;
uniform sampler2D metallic_map;
uniform sampler2D normal_map;
uniform sampler2D roughness_map;

layout(shared, binding = 0) uniform EmissiveColorBlock {
    vec3 base_color;
    vec3 emissive;
    float emissive_intensity;
    float metallic;
    float roughness;
    float use_ao_map;
    float use_color_map;
    float use_emissive_map;
    float use_metallic_map;
    float use_normal_map;
    float use_roughness_map;
    vec2 uv_offset;
    vec2 uv_scale;
};

void main() {
    vec2 uv = fragUV0 * uv_scale + uv_offset;
    float ao, metallicVal, roughnessVal;
    vec4 color;
    vec3 normal = vec3(0.5, 0.5, 1.0) * 2.0 - 1.0;
    if (use_ao_map == 1.0) {
        ao = texture(ao_map, uv).r;
    } else {
        ao = 1.0;
    }

    if (use_color_map == 1.0) {
        color = texture(color_map, uv);
    } else {
        color = vec4(base_color, 1.0);
    }

    if (use_metallic_map == 1.0) {
        metallicVal = texture(metallic_map, uv).r;
    } else {
        metallicVal = metallic;
    }

    if (use_roughness_map == 1.0) {
        roughnessVal = texture(roughness_map, uv).r;
    } else {
        roughnessVal = roughness;
    }

    if (use_normal_map == 1.0) {
        normal.xy = texture(normal_map, uv).rg * 2.0 - 1.0;
        normal.z = reconstructNormalZ(normal.xy);
        normal = normalize(normal);
    }

    fragColor = calculateLighting(color, roughnessVal, metallicVal, ao, normal);
    // tangent/bitangent appear to be incorrect for this material, or at least the unit (0xe176a69136bc9e6d) I was testing.
    // the packed normal decoding does appear to be correct, but for some reason the (bi)tangents*normalMat are not
    // identical for each vertex on the face, which leads to problems with lighting
    //fragColor = vec4(dbg_fragTBN[uint(debug_mode)], 1.0);
}