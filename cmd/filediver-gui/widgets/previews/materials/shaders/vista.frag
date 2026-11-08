#version 430 core

#include "lighting.frag"
#include "triplanar.frag"
#include "grading.frag"

out vec4 fragColor;

in vec3 objPosition;
in vec2 fragUV0;
in mat4 instanceModel;
in mat3 dbg_fragITBN;

uniform sampler2D base_normal_ao;
uniform sampler2D detail_mask_;
uniform sampler2D emissive_texture;
uniform sampler2D ground_albedo;
uniform sampler2D ground_nar;
uniform sampler2D vista_detail_albedo;
uniform sampler2D vista_detail_nar;
uniform sampler2DArray triplanar_detail_albedo;
uniform sampler2DArray triplanar_detail_data;

uniform mat4 projection;

layout(shared, binding = 0) uniform VistaBlock {
    uint vista_type;
    float albedo_intensity_ground;
    float albedo_intensity_rock;
    float albedo_intensity_vista_detail;
    float base_ao_intensity;
    float base_normal_intensity;
    float emissive_intensity;
    float grading_group_id_ground;
    float grading_group_id_rock;
    float grading_group_id_vista_detail;
    float material_wetness;
    float normal_intensity_ground;
    float normal_intensity_vista_detail;
    float overlay_mask_amount;
    float triplanar_tiling;
    float tiling_ground;
    float tiling_vista_detail;
    float triplanar_mask_sharpen;
    float triplanar_normal_intensity;
    float triplanar_vertex_normal_lerp;
};

layout(shared, binding = 1) uniform DebugBlock {
    uint debug;
};

// vec3 emissive(vec3 base_color, float roughness, float ao, float obj_height) {
//     float luminance = dot(base_color, vec3(0.2127, 0.7152, 0.0721)) - emissive_base_color_luminance_range.x;
//     luminance = clamp(luminance / max(emissive_base_color_luminance_range.y - emissive_base_color_luminance_range.x, 0.001), 0.0, 1.0);
//     vec3 emissive_color = luminance * base_color * emissive_base_color_tint;

//     float emissive_roughness = 1.0 - clamp((emissive_roughness_range.x - roughness) / (emissive_roughness_range.x - emissive_roughness_range.y), 0.0, 1.0);
//     float emissive_ao = 1.0 - clamp((emissive_ao_range.x - ao) / (emissive_ao_range.x - emissive_ao_range.y), 0.0, 1.0);
//     vec3 emissive_roughness_ao_color = emissive_ao * emissive_roughness * emissive_base_color_ao_roughness;
//     emissive_roughness_ao_color = emissive_ao_roughness_intensity * emissive_roughness_ao_color;

//     emissive_color = gamma(abs(emissive_color), emissive_base_color_curve);
//     emissive_color = emissive_color * emissive_base_color_intensity + emissive_roughness_ao_color;
//     emissive_color = gamma(abs(emissive_color), emissive_power);

//     float emissive_height_val = clamp((obj_height - emissive_height_fade_start) / max(emissive_height_fade_distance, 0.001), 0.0, 1.0);
//     emissive_height_val = pow(emissive_height_val, max(emissive_height_fade_exponent, 0.001));

//     emissive_color = min((1.0 - emissive_height_val) * emissive_color, vec3(1.0, 1.0, 1.0));
//     return gamma(emissive_color, 0.8);
// }

void main() {
    vec3 normal = normalize(dbg_fragTBN[2].xyz);
    vec3 mixed_vertex_normal = normal;

    // vec3 face_tangent = dFdx(objPosition);
    // vec3 face_bitangent = dFdy(objPosition);
    // vec3 face_mixed_tangent = face_tangent - dot(face_tangent, mixed_vertex_normal) * mixed_vertex_normal;
    // vec3 face_mixed_bitangent = face_bitangent - dot(face_bitangent, mixed_vertex_normal) * mixed_vertex_normal;
    // vec3 face_crossed_bitangent = cross(face_tangent, mixed_vertex_normal);
    // float face_handedness = dot(face_bitangent, face_crossed_bitangent) < 0.0 ? -1.0 : 1.0;

    // vec2 uv_dx = dFdx(fragUV0);
    // vec2 uv_dy = dFdy(fragUV0);

    // float uv_align = dot(uv_dx, vec2(uv_dy.y, -uv_dy.x));
    // float uv_handedness = uv_align < 0.0 ? -1.0 : 1.0;
    // vec3 uv_face_tangent = (uv_handedness * uv_dy.y) * face_mixed_tangent - ((uv_handedness * uv_dx.y) * face_mixed_bitangent);

    // uv_face_tangent = abs(uv_align) > 0.0 ? normalize(uv_face_tangent) : uv_face_tangent;
    // float handedness = face_handedness * uv_handedness;
    // vec3 uv_face_bitangent = handedness * cross(uv_face_tangent, mixed_vertex_normal);

    vec3 base_normal_ao_sample = texture(base_normal_ao, fragUV0).xyz;
    vec3 details = texture(detail_mask_, fragUV0).xyz;

    vec3 base_normal = vec3((base_normal_ao_sample.xy * 2.0 - 1.0) * vec2(-1.0, 1.0), 0.0);

    vec2 nxy_sq = base_normal.xy * base_normal.xy;

    base_normal.xy = base_normal_intensity * base_normal.xy;
    base_normal.z = max(max(nxy_sq.x, nxy_sq.y) / 16384.0, 1.0 - base_normal.x * base_normal.x - base_normal.y * base_normal.y);
    base_normal = dbg_fragTBN * base_normal;

    vec3 combined_normal = mixed_vertex_normal + base_normal;
    combined_normal = triplanar_vertex_normal_lerp * (normal - combined_normal) + combined_normal;
    float sqrlen = dot(combined_normal, combined_normal);
    combined_normal = normalize(combined_normal);
    if (sqrlen <= 0.0) {
        combined_normal = vec3(0.0, 0.0, 1.0);
    }

    vec3 sharpened_normal = gamma(abs(combined_normal), abs(triplanar_mask_sharpen) + 1.0);
    sharpened_normal = clamp((sharpened_normal / dot(sharpened_normal, vec3(1.0))) * 1.2 - 0.1, 0.0, 1.0);

    float projection_val = 1.0 / (ceil(12.0 / projection[0].x) / 6.0);
    projection_val = projection_val * (25.0 / 9.0);

    vec3 tangentTriplanarNormal = normalize(dbg_fragITBN * objPosition);
    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    float triplanar_view_align = sqrt(dot(tangentTriplanarNormal, viewDirection));
    triplanar_view_align = max(ceil(triplanar_view_align * 0.01) * 500, 1.0);

    vec3 heightmap_normal = fract(0.1031 * roundEven(projection_val / triplanar_view_align * vec3(0.0, 1.0, 0.0)));
    vec3 adjusted_normal = heightmap_normal.yzx + vec3(33.33);
    heightmap_normal = vec3(dot(heightmap_normal, adjusted_normal)) + heightmap_normal;
    float result = fract(heightmap_normal.z * (heightmap_normal.x + heightmap_normal.y)) - 0.5;
    result = 1.0 - 1.5 * result;

    vec2 detail_normal_x = vec2(0.0);
    vec2 detail_normal_y = vec2(0.0);
    vec2 detail_normal_z = vec2(0.0);
    vec3 detail_albedo = vec3(0.0);
    float detail_ao = 0.0;
    float detail_roughness = 0.0;
    if (sharpened_normal.x > 0.0) {
        vec3 triplanar_uvs = vec3(objPosition.zy * triplanar_tiling, 0.0);
        vec4 data_sample = texture(triplanar_detail_data, triplanar_uvs);
        vec3 albedo_sample = texture(triplanar_detail_albedo, triplanar_uvs).xyz;
        detail_albedo = albedo_sample * sharpened_normal.x;

        detail_normal_x = (data_sample.xy * vec2(-1.0, 1.0) + vec2(1.0, 0.0)) * 2 - 1;
        vec2 temp = detail_normal_x * detail_normal_x;
        float normal_z = inversesqrt(max(max(temp.x, temp.y) / 16384.0, 1.0 - detail_normal_x.x * detail_normal_x.x - detail_normal_x.y * detail_normal_x.y));
        detail_normal_x = sharpened_normal.x * detail_normal_x * -normal_z;

        detail_ao = sharpened_normal.x * data_sample.z;
        detail_roughness = sharpened_normal.x * data_sample.w;
    }

    if (sharpened_normal.y > 0.0) {
        vec3 triplanar_uvs = vec3(objPosition.xy * triplanar_tiling, 0.0);
        vec4 data_sample = texture(triplanar_detail_data, triplanar_uvs);
        vec3 albedo_sample = texture(triplanar_detail_albedo, triplanar_uvs).xyz;
        detail_albedo = albedo_sample * sharpened_normal.y + detail_albedo;

        detail_normal_y = (data_sample.xy * vec2(-1.0, 1.0) + vec2(1.0, 0.0)) * 2 - 1;
        vec2 temp = detail_normal_y * detail_normal_y;
        float normal_z = inversesqrt(max(max(temp.x, temp.y) / 16384.0, 1.0 - detail_normal_y.x * detail_normal_y.x - detail_normal_y.y * detail_normal_y.y));
        detail_normal_y = sharpened_normal.y * detail_normal_y * -normal_z;

        detail_ao = sharpened_normal.y * data_sample.z + detail_ao;
        detail_roughness = sharpened_normal.y * data_sample.w + detail_roughness;
    }

    if (sharpened_normal.z > 0.0) {
        vec3 triplanar_uvs = vec3(objPosition.zx * triplanar_tiling, 0.0);
        vec4 data_sample = texture(triplanar_detail_data, triplanar_uvs);
        vec3 albedo_sample = texture(triplanar_detail_albedo, triplanar_uvs).xyz;
        detail_albedo = albedo_sample * sharpened_normal.z + detail_albedo;

        detail_normal_z = (data_sample.xy * vec2(-1.0, 1.0) + vec2(1.0, 0.0)) * 2 - 1;
        vec2 temp = detail_normal_z * detail_normal_z;
        float normal_z = inversesqrt(max(max(temp.x, temp.y) / 16384.0, 1.0 - detail_normal_z.x * detail_normal_z.x - detail_normal_z.y * detail_normal_z.y));
        detail_normal_z = sharpened_normal.z * detail_normal_z * -normal_z;

        detail_ao = sharpened_normal.z * data_sample.z + detail_ao;
        detail_roughness = sharpened_normal.z * data_sample.w + detail_roughness;
    }
    
    vec3 combined_detail_normal = vec3(detail_normal_z.x + detail_normal_y.y, detail_normal_z.y + detail_normal_x.y, detail_normal_x.x + detail_normal_y.x);
    float normal_align = dot(combined_normal, combined_detail_normal);
    combined_normal = triplanar_normal_intensity * (combined_detail_normal - normal_align * combined_normal) + base_normal;

    vec3 albedo = detail_albedo * albedo_intensity_rock;
    float base_ao = base_ao_intensity * (base_normal_ao_sample.z - 1.0) + 1.0;
    float ao = base_ao * detail_ao;

    if (grading_group_id_rock > 0.0) {
        albedo = gradeColor(albedo, gradingMatrix(gradingPixel(grading_group_id_rock)));
    }
    albedo = clamp(albedo, 0.0, 1.0);

    vec3 ground_color = albedo_intensity_ground * texture(ground_albedo, fragUV0 * tiling_ground).xyz;
    vec4 ground_nar_sample = texture(ground_nar, fragUV0 * tiling_ground);

    if (grading_group_id_ground > 0.0) {
        ground_color = gradeColor(ground_color, gradingMatrix(gradingPixel(grading_group_id_ground)));
    }
    ground_color = clamp(ground_color, 0.0, 1.0);
    albedo = details.x * (ground_color - albedo) + albedo;
    float roughness = details.x * (ground_nar_sample.w - detail_roughness) + detail_roughness;
    ao = details.x * (base_ao * ground_nar_sample.z - ao) + ao;
    
    vec3 ground_normal = vec3((ground_nar_sample.xy * 2.0 - 1.0) * vec2(-1.0, 1.0), 0.0);
    vec2 temp = ground_normal.xy * ground_normal.xy;
    ground_normal.z = inversesqrt(max(max(temp.x, temp.y) / 16384.0, 1.0 - ground_normal.x * ground_normal.x - ground_normal.y * ground_normal.y));
    ground_normal.xy = ground_normal.xy * -ground_normal.z;
    ground_normal = normal_intensity_ground * (ground_normal.x * normalize(dbg_fragTBN[0]) + ground_normal.y * normalize(dbg_fragTBN[1])) + base_normal;
    combined_normal = details.x * (ground_normal - combined_normal) + combined_normal;

    vec3 vista_detail_color = albedo_intensity_vista_detail * texture(vista_detail_albedo, fragUV0 * tiling_vista_detail).xyz;
    vec4 vista_detail_nar_sample = texture(vista_detail_nar, fragUV0 * tiling_vista_detail);

    if (grading_group_id_vista_detail > 0.0) {
        vista_detail_color = gradeColor(vista_detail_color, gradingMatrix(gradingPixel(grading_group_id_vista_detail)));
    }
    vista_detail_color = clamp(vista_detail_color, 0.0, 1.0);
    albedo = details.y * (vista_detail_color - albedo) + albedo;
    roughness = details.y * (vista_detail_nar_sample.w - roughness) + roughness;
    ao = details.y * (base_ao * vista_detail_nar_sample.z - ao) + ao;
    
    vec3 vista_detail_normal = vec3((vista_detail_nar_sample.xy * 2.0 - 1.0) * vec2(-1.0, 1.0), 0.0);
    temp = vista_detail_normal.xy * vista_detail_normal.xy;
    vista_detail_normal.z = inversesqrt(max(max(temp.x, temp.y) / 16384.0, 1.0 - vista_detail_normal.x * vista_detail_normal.x - vista_detail_normal.y * vista_detail_normal.y));
    vista_detail_normal.xy = vista_detail_normal.xy * -vista_detail_normal.z;
    vista_detail_normal = normal_intensity_vista_detail * (vista_detail_normal.x * normalize(dbg_fragTBN[0]) + vista_detail_normal.y * normalize(dbg_fragTBN[1])) + base_normal;
    combined_normal = details.y * (vista_detail_normal - combined_normal) + combined_normal;

    albedo = overlay_mask_amount * (2.0 * details.z * albedo - albedo) + albedo;
    combined_normal = normalize(combined_normal);
    combined_normal.z = -combined_normal.z;
    roughness = roughness * 0.955 + 0.045;
    float metallic = material_wetness;

    vec3 tangentCombinedNormal = dbg_fragITBN * combined_normal;

    vec3 emissive_color = emissive_intensity * texture(emissive_texture, fragUV0).xyz;

    switch (debug) {
    case 1:
        fragColor = vec4(sqrt(albedo), 1.0);
        break;
    case 2:
        fragColor = vec4(vec3(roughness), 1.0);
        break;
    case 3:
        fragColor = vec4(vec3(ao), 1.0);
        break;
    case 4:
        fragColor = vec4(vec3(metallic), 1.0);
        break;
    case 5:
        fragColor = vec4(combined_normal, 1.0);
        break;
    case 6:
        fragColor = vec4(ground_normal, 1.0);
        break;
    case 7:
        fragColor = vec4(vista_detail_normal, 1.0);
        break;
    case 8:
        fragColor = vec4(base_normal, 1.0);
        break;
    case 9:
        fragColor = vec4(mixed_vertex_normal, 1.0);
        break;
    case 10:
        fragColor = vec4(combined_detail_normal, 1.0);
        break;
    case 11:
        fragColor = vec4(sharpened_normal, 1.0);
        break;
    case 12:
        fragColor = vec4(detail_normal_x, 0.0, 1.0);
        break;
    case 13:
        fragColor = vec4(detail_normal_y, 0.0, 1.0);
        break;
    case 14:
        fragColor = vec4(detail_normal_z, 0.0, 1.0);
        break;
    case 15:
        fragColor = vec4(fract(objPosition), 1.0);
        break;
    case 16:
        fragColor = vec4(vec3(result), 1.0);
        break;
    default:
        fragColor = calculateLighting(vec4(sqrt(albedo), 1.0), roughness, metallic, 0.0, ao, tangentCombinedNormal, 1.0, 1.0);
        fragColor += vec4(emissive_color, 0.0);
    }
}
