#version 430 core

#include "lighting.frag"

out vec4 fragColor;

in vec2 fragUV0;
in vec2 fragUV1;

layout(shared, binding = 0) uniform BugSettingsBlock {
    float material_wetness;
    // vec3 blood_color0;
    // vec3 blood_color1;
    // float blood_roughness;
    // float blood_subsurface;
    // vec2 burn_scorch;
    float clearcoat_intensity;
    float clearcoat_normal_mix;
    float clearcoat_roughness;
    float detail_curv_intensity_1;
    float detail_curv_intensity_2;
    float detail_mix_weight;
    float detail_norm_clearcoat_intensity_1;
    float detail_norm_clearcoat_intensity_2;
    float detail_norm_intensity_1;
    float detail_norm_intensity_2;
    float detail_norm_tiler_1;
    float detail_norm_tiler_2;
    float detail_roughness_clearcoat_intensity_1;
    float detail_roughness_clearcoat_intensity_2;
    float emissive_intensity;
    float i_end;
    float i_intensity;
    float i_thickness;
    float i_start;
    float micro_ao_intensity;
    float opacity_offset;
    float opacity_sharpness;
    float sss_intensity;
    float subsurface_diffusion;
    float subsurface_thickness;
    float subsurface_wrap;
    // float use_large_wound_lookup;
    // float wound_tile_scale;
    // float weathering_tiling;
    // float woundable_id;
    // float wound_parallax_scale;
    // float wound_painting_enabled;
};

layout(shared, binding = 1) uniform DebugBlock {
    uint debug;
};

uniform sampler2D albedo_iridescence;
uniform sampler2D clearcoat_roughness_ao_spec;
uniform sampler2D normal;
uniform sampler2D subsurface_opacity;
uniform sampler2D detail_normal_1;
uniform sampler2D texture_map_02;
uniform sampler2D emissive_color;
uniform sampler2D sclar_field_opacity;

void main() {
    vec4 alb_iri = texture(albedo_iridescence, fragUV0);
    vec4 ss_opacity = texture(subsurface_opacity, fragUV0);
    vec4 cras = texture(clearcoat_roughness_ao_spec, fragUV0);
    float clearcoat = cras.x;
    float roughness = cras.y;
    float ao = cras.z;
    float spec = cras.w;
    float opacity = texture(sclar_field_opacity, fragUV0).x;

    opacity = clamp((opacity - 0.5) * opacity_sharpness + 0.5 + opacity_offset, 0.0, 1.0);
    if (opacity < 0.5) {
        discard;
    }
    vec3 normal_sample;
    normal_sample.xy = texture(normal, fragUV0).xy * 2.0 - 1.0;
    float normal_w = sqrt(1.0 - min(dot(normal_sample.xy, normal_sample.xy), 1.0));

    vec3 det_normals_1 = texture(detail_normal_1, fragUV0 * detail_norm_tiler_1).xyz * 2.0 - 1.0;
    vec3 det_normals_2 = texture(texture_map_02, fragUV0 * detail_norm_tiler_2).xyz * 2.0 - 1.0;

    // r12.x
    float details_weight = clamp(detail_mix_weight * (clearcoat * ao - 0.5) + 0.5, 0.0, 1.0);
    vec3 combined_detail_normal = vec3(det_normals_1.xy * detail_norm_intensity_1, 0.0);
    vec3 combined_clearcoat_detail_normal = vec3(det_normals_1.xy * detail_norm_clearcoat_intensity_1, 0.0);

    combined_detail_normal.xy = mix(combined_detail_normal.xy, det_normals_2.xy * detail_norm_intensity_2, details_weight);
    combined_clearcoat_detail_normal.xy = mix(combined_clearcoat_detail_normal.xy, det_normals_2.xy * detail_norm_clearcoat_intensity_2, details_weight);

    combined_detail_normal.z = reconstructNormalZ(combined_detail_normal.xy);
    normal_sample.z = reconstructNormalZ(normal_sample.xy);
    vec3 output_normal = reorientNormalMaps(normal_sample, combined_detail_normal);

    float detail_curvature = mix(det_normals_1.z * detail_curv_intensity_1, det_normals_2.z * detail_curv_intensity_2, details_weight);
    float detail_clearcoat = mix(det_normals_1.z * detail_roughness_clearcoat_intensity_1, det_normals_2.z * detail_roughness_clearcoat_intensity_2, details_weight);
    
    float inv_detail_curvature = 1.5 - (detail_curvature * 0.5 + 0.5);
    detail_curvature = clamp(detail_curvature * 0.5 + 1.0, 0.0, 1.0);

    vec3 base_color = gamma(abs(alb_iri.xyz), inv_detail_curvature);
    float adjusted_spec = gamma(abs(vec3(spec)), inv_detail_curvature).x;
    spec = clamp(spec * 2, 0.0, 1.0);
    detail_curvature = 1.0 - (spec * detail_curvature);
    float micro_ao_spec = 1.0 - (detail_curvature * micro_ao_intensity);

    float iridescence = clamp(alb_iri.w * i_thickness * i_intensity, 0.0, 1.0);
    float iridescence_thickness = clamp(i_thickness, 0.0, 1.0);
    float iridescence_start = clamp(i_start, 0.0, 1.0);
    float iridescence_end = clamp(i_end, 0.0, 1.0);
    iridescence_end = iridescence_end * (1.0 - iridescence_start);
    float iridescence_weight = iridescence_end * iridescence_thickness + iridescence_start;
    iridescence_weight = iridescence_weight * iridescence_weight;

    float material_wetness_value = clamp(material_wetness, 0.0, 1.0);

    float clearcoat_modified = clamp(clearcoat * clearcoat_intensity, 0.0, 1.0);
    roughness = roughness * 0.955 + 0.045;

    if (ss_opacity.w > 0.0) {
        ss_opacity.xyz = clamp(ss_opacity.xyz, 0.0, 1.0);
        base_color.xyz = ss_opacity.w * (ss_opacity.xyz - base_color.xyz) + base_color.xyz;
    }
    combined_clearcoat_detail_normal.z = reconstructNormalZ(combined_clearcoat_detail_normal.xy);
    if (clearcoat_modified > 0.0) {
        vec3 clearcoat_normal = clearcoat_normal_mix * (vec3(0.0, 0.0, 1.0) - vec3(normal_sample.xy, normal_w)) + vec3(normal_sample.xy, normal_w);
        clearcoat_normal = reorientNormalMaps(clearcoat_normal, combined_clearcoat_detail_normal);
        float clearcoat_roughness_value = (clamp(clearcoat_roughness * cras.y - detail_clearcoat, 0.0, 1.0) * 0.955 + 0.045) - roughness;
        roughness = clearcoat_modified * clearcoat_roughness_value + roughness;
        clearcoat_normal.xy = clearcoat_normal.xy * clearcoat_modified;
        clearcoat_normal.z = reconstructNormalZ(clearcoat_normal.xy);
        output_normal = reorientNormalMaps(output_normal, clearcoat_normal);
    }

    vec3 emissive_color_sample = texture(emissive_color, fragUV0).xyz;
    float emissive_alignment = dot(emissive_color_sample, vec3(0.212, 0.715, 0.072));
    vec3 temp_emissive = emissive_alignment * base_color.xyz * emissive_color_sample.xyz;
    emissive_color_sample.xyz = emissive_intensity * (emissive_alignment * (emissive_color_sample - temp_emissive) + temp_emissive);
    emissive_color_sample.xyz = clamp(emissive_color_sample, 0.0, 65504.0);

    temp_emissive = gamma(min(emissive_color_sample.xyz / 128.0, vec3(1.0)), 0.8);

    if (gl_FrontFacing) {
        output_normal = -output_normal;
    }

    switch(debug) {
    case 1:
        fragColor = vec4(base_color.xyz, 1.0);
        break;
    case 2:
        fragColor = vec4(vec3(roughness), 1.0);
        break;
    case 3:
        fragColor = vec4(vec3(spec), 1.0);
        break;
    case 4:
        fragColor = vec4(vec3(adjusted_spec), 1.0);
        break;
    case 5:
        fragColor = vec4(vec3(ao), 1.0);
        break;
    case 6:
        fragColor = vec4(output_normal * 0.5 + 0.5, 1.0);
        break;
    case 7:
        fragColor = vec4(emissive_color_sample, 1.0);
        break;
    case 8:
        fragColor = vec4(temp_emissive, 1.0);
        break;
    case 9:
        fragColor = vec4(vec3(micro_ao_spec), 1.0);
        break;
    default:
        fragColor = calculateLighting(vec4(base_color.xyz, 1.0), roughness, 0.0, micro_ao_spec * 0.1, ao, output_normal, 1.0, 1.0);
        fragColor.xyz += temp_emissive;
    }
}