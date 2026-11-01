#version 430 core

#include "lighting.frag"
#include "triplanar.frag"
#include "grading.frag"

out vec4 fragColor;

in vec3 objPosition;
in vec2 fragUV0;
in mat4 instanceModel;
in mat3 dbg_fragITBN;

uniform sampler2D base_normal_ao_dirt;
uniform sampler2D detail_mask_;
uniform sampler2D snow_glints_tiler;
uniform sampler2D weathering_albedo;
uniform sampler2D weathering_nar;
uniform sampler2DArray triplanar_detail_albedo;
uniform sampler2DArray triplanar_detail_data;
uniform sampler2DArray snow_pnrb_array;

uniform mat4 inv_view;

layout(shared, binding = 0) uniform RockBlock {
    uint rock_type;
    float albedo_intensity;
    float base_ao_intensity;
    float base_normal_intensity;
    vec3  bottom_color_tint;
    float bottom_dirt_value;
    float dirt_ao;
    float dirt_ao_sharpness;
    vec3  dirt_color;
    float dirt_detail_ao;
    float dirt_detail_sharpness;
    float dirt_detail_mask;
    float dirt_metallic;
    float dirt_roughness;
    float dirt_roughness_blend;
    vec2 emissive_ao_range;
    float emissive_ao_roughness_intensity;
    vec3 emissive_base_color_ao_roughness;
    float emissive_base_color_curve;
    float emissive_base_color_intensity;
    vec2 emissive_base_color_luminance_range;
    vec3 emissive_base_color_tint;
    float emissive_height_fade_distance;
    float emissive_height_fade_exponent;
    float emissive_height_fade_start;
    float emissive_power;
    vec2 emissive_roughness_range;
    float gradient_dirt_maxheight;
    float gradient_dirt_minheight;
    float grading_group_id;
    float grading_group_id_weathering;
    float grading_secondary_group_id;
    float heightmap_normals;
    vec3 ice_fuzz_color;
    float ice_fuzz_intensity;
    vec3 ice_subsurface_color;
    float ice_subsurface_diffusion;
    float ice_subsurface_intensity;
    float ice_subsurface_thickness;
    float ice_subsurface_wrap;
    float weathering_input;
    float material_wetness;
    float snow_amount;
    vec3 snow_base_color;
    vec3 snow_fuzz_color;
    float snow_glint_amount;
    float snow_glint_intensity;
    float snow_glint_roughness;
    float snow_glint_size;
    float snow_index0;
    float snow_normal_displacement;
    float snow_roughness;
    float snow_specular;
    float snow_ss_thickness;
    vec3 snow_subsurface_color;
    float snow_subsurface_diffusion;
    float snow_subsurface_intensity;
    float snow_subsurface_thickness;
    float snow_subsurface_wrap;
    float snow_up_displacement;
    float subsurface_occlusion_intensity;
    float subsurface_occlusion_maskcurve;
    float tiler_ao_intensity;
    vec3  top_color_tint;
    float top_dirt_value;
    vec3  triplanar_detail_tiler;
    float triplanar_normal_intensity;
    float triplanar_vertex_normal_lerp;
    float up_fade_curve;
    float up_fade_distance;
    float weathering_ao_amount;
    float weathering_ao_cutoff;
    float weathering_coverage;
    float weathering_coverage_multiplier;
    float weathering_height;
    float weathering_height_curve;
    float weathering_normal_mult;
    float weathering_on_dirt_amount;
    float weathering_thickness;
    float weathering_tiling;
};

layout(shared, binding = 1) uniform DebugBlock {
    uint debug;
};

vec3 emissive(vec3 base_color, float roughness, float ao, float obj_height) {
    float luminance = dot(base_color, vec3(0.2127, 0.7152, 0.0721)) - emissive_base_color_luminance_range.x;
    luminance = clamp(luminance / max(emissive_base_color_luminance_range.y - emissive_base_color_luminance_range.x, 0.001), 0.0, 1.0);
    vec3 emissive_color = luminance * base_color * emissive_base_color_tint;

    float emissive_roughness = 1.0 - clamp((emissive_roughness_range.x - roughness) / (emissive_roughness_range.x - emissive_roughness_range.y), 0.0, 1.0);
    float emissive_ao = 1.0 - clamp((emissive_ao_range.x - ao) / (emissive_ao_range.x - emissive_ao_range.y), 0.0, 1.0);
    vec3 emissive_roughness_ao_color = emissive_ao * emissive_roughness * emissive_base_color_ao_roughness;
    emissive_roughness_ao_color = emissive_ao_roughness_intensity * emissive_roughness_ao_color;

    emissive_color = gamma(abs(emissive_color), emissive_base_color_curve);
    emissive_color = emissive_color * emissive_base_color_intensity + emissive_roughness_ao_color;
    emissive_color = gamma(abs(emissive_color), emissive_power);

    float emissive_height_val = clamp((obj_height - emissive_height_fade_start) / max(emissive_height_fade_distance, 0.001), 0.0, 1.0);
    emissive_height_val = pow(emissive_height_val, max(emissive_height_fade_exponent, 0.001));

    emissive_color = min((1.0 - emissive_height_val) * emissive_color, vec3(1.0, 1.0, 1.0));
    return gamma(emissive_color, 0.8);
}

vec3 ice_fuzz_subsurface(vec3 base_color, float view_align, float inv_ao) {
    vec3 fuzz_color = clamp(ice_fuzz_color, 0.0, 1.0);
    float fuzz_intensity = clamp(ice_fuzz_intensity, 0.0, 1.0);
    fuzz_intensity = max(fuzz_intensity - view_align * fuzz_intensity, 0.0);
    base_color = fuzz_intensity * (fuzz_color - base_color) + base_color;

    float subsurface_occlusion = clamp(pow(inv_ao, max(subsurface_occlusion_maskcurve, 0.001)) * subsurface_occlusion_intensity + ice_subsurface_intensity, 0.0, 1.0);
    float subsurface_wrap = clamp(ice_subsurface_wrap, 0.0, 1.0);
    float subsurface_wrap_val = subsurface_wrap * (1.0 - 2.0 * view_align) + view_align;
    subsurface_wrap = subsurface_wrap - 0.5;
    subsurface_wrap_val = subsurface_occlusion * (pow(abs(subsurface_wrap_val), abs(subsurface_wrap) * 2.0 + 1.0) - abs(subsurface_wrap) + 0.5);
    return subsurface_wrap_val * (ice_subsurface_color - base_color) + base_color;
}

void main() {
    vec3 normal = normalize(dbg_fragTBN[2].xyz);
    vec3 mixed_vertex_normal = heightmap_normals * (vec3(0.0, 0.0, 1.0) - normal) + normal;
    float sqrlen = dot(mixed_vertex_normal, mixed_vertex_normal);
    mixed_vertex_normal = mixed_vertex_normal * inversesqrt(sqrlen);
    if (sqrlen == 0.0) {
        mixed_vertex_normal = vec3(0.0, 0.0, 1.0);
    }

    vec3 heightmap_normal = /*heightmap_tbn **/ normal;

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

    vec4 nad = texture(base_normal_ao_dirt, fragUV0);
    vec2 details = texture(detail_mask_, fragUV0).xy;

    vec3 nad_normal = vec3((nad.xy * 2.0 - 1.0) * vec2(-1.0, 1.0), 0.0);

    vec2 nxy_sq = nad_normal.xy * nad_normal.xy;

    nad_normal.z = max(max(nxy_sq.x, nxy_sq.y) / 16384.0, 1.0 - nad_normal.x * nad_normal.x - nad_normal.y * nad_normal.y);
    nad_normal.z = inversesqrt(nad_normal.z);
    nad_normal.xy = nad_normal.xy * -nad_normal.z;
    nad_normal.xyz = base_normal_intensity * (nad_normal.x * normalize(dbg_fragTBN[0]) + nad_normal.y * normalize(dbg_fragTBN[1]));

    vec3 combined_normal = mixed_vertex_normal + nad_normal;
    combined_normal = triplanar_vertex_normal_lerp * (normal - combined_normal) + combined_normal;
    sqrlen = dot(combined_normal, combined_normal);
    combined_normal = normalize(combined_normal);
    if (sqrlen <= 0.0) {
        combined_normal = vec3(0.0, 0.0, 1.0);
    }

    vec2 uvs = getCustomTriplanarUVs(combined_normal);

    vec4 detail_albedo_x = texture(triplanar_detail_albedo, vec3(uvs * triplanar_detail_tiler.x, 0.0));
    vec4 detail_albedo_y = texture(triplanar_detail_albedo, vec3(uvs * triplanar_detail_tiler.y, 1.0));
    vec4 detail_albedo_z = texture(triplanar_detail_albedo, vec3(uvs * triplanar_detail_tiler.z, 2.0));

    float detail_mix_y = dot(details.xx, details.xx) - 3.0 * details.x + 2.0;
    detail_mix_y = dot(vec2(detail_mix_y), details.xx);
    detail_mix_y = clamp(((2.0 * detail_mix_y) + detail_albedo_y.w - detail_albedo_x.w - 2.0) * 1.5 + 0.5, 0.0, 1.0);

    vec3 combined_detail_albedo = detail_mix_y * (detail_albedo_y.xyz - detail_albedo_x.xyz) + detail_albedo_x.xyz;
    float combined_albedo_blend = details.x * (detail_albedo_y.w - detail_albedo_x.w) + detail_albedo_x.w;

    float detail_mix_z = dot(details.yy, details.yy) - 3.0 * details.y + 2.0;
    detail_mix_z = dot(vec2(detail_mix_z), details.yy);
    combined_albedo_blend = (combined_albedo_blend + 1.0) - detail_mix_z;
    detail_mix_z = detail_mix_z + detail_albedo_z.w - 1.0;
    detail_mix_z = clamp((detail_mix_z - combined_albedo_blend) * 1.5 + 0.5, 0.0, 1.0);
    combined_detail_albedo = detail_mix_z * (detail_albedo_z.xyz - combined_detail_albedo) + combined_detail_albedo;

    vec4 detail_data_x = texture(triplanar_detail_data, vec3(uvs * triplanar_detail_tiler.x, 0.0));
    vec4 detail_data_y = texture(triplanar_detail_data, vec3(uvs * triplanar_detail_tiler.y, 1.0));
    vec4 detail_data_z = texture(triplanar_detail_data, vec3(uvs * triplanar_detail_tiler.z, 2.0));

    vec4 combined_detail_data = detail_mix_y * (detail_data_y - detail_data_x) + detail_data_x;
    combined_detail_data = detail_mix_z * (detail_data_z - combined_detail_data) + combined_detail_data;

    float inv_ao = 1.0 - combined_detail_data.z;
    float half_ao = combined_detail_data.z * 0.5;

    vec3 detail_normal = vec3((combined_detail_data.xy * vec2(-1.0, 1.0) + vec2(1.0, 0.0)) * 2.0 - 1.0, 0.0);
    vec2 dn_sq = detail_normal.xy * detail_normal.xy;
    float dn_z = inversesqrt(max(max(dn_sq.x, dn_sq.y) / 16384.0, 1.0 - detail_normal.x * detail_normal.x - detail_normal.y * detail_normal.y));

    detail_normal.xz = detail_normal.xy * -dn_z;
    detail_normal.y = 0.0;

    detail_normal = getCustomTriplanarNormal(combined_normal, detail_normal);
    combined_normal = triplanar_normal_intensity * (detail_normal - dot(detail_normal, combined_normal) * combined_normal) + nad_normal;

    vec3 base_albedo = combined_detail_albedo * albedo_intensity;

    float obj_height = objPosition.z + fd_world_offset.y;
    float dirt_height = clamp((obj_height - gradient_dirt_maxheight) / (gradient_dirt_minheight - gradient_dirt_maxheight), 0.0, 1.0);
    dirt_height = (3.0 - dirt_height * 2.0) * dirt_height * dirt_height;

    float dirt_value = dirt_height * (bottom_dirt_value - top_dirt_value) + top_dirt_value;

    float inv_dirt_detail_sharpness = 1.0 - dirt_detail_sharpness;
    half_ao = min(pow(abs(half_ao), 4.0), 1.0);

    float inv_dirt_ao_sharpness = 1.0 - dirt_ao_sharpness;
    float inv_dirt_ao = 1.0 - dirt_ao;
    float inv_dirt_detail_ao = 1.0 - dirt_detail_ao;

    float min_value = min(dirt_value, 0.0);
    float dirt_combined_inv_ao = clamp(clamp(clamp(inv_ao + min_value, 0.0, 1.0) - inv_dirt_detail_ao, 0.0, 1.0) / inv_dirt_detail_sharpness, 0.0, 1.0);
    dirt_combined_inv_ao = (3.0 - dirt_combined_inv_ao * 2.0) * dirt_combined_inv_ao * dirt_combined_inv_ao;

    float dirt_combined_ao = max(clamp(1.0 + (min_value - nad.z) - half_ao, 0.0, 1.0) - clamp(inv_dirt_ao, 0.0, 1.0), 0.0);
    dirt_combined_ao = clamp(dirt_combined_ao / inv_dirt_ao_sharpness, 0.0, 1.0);
    dirt_combined_ao = (3.0 - dirt_combined_ao * 2.0) * dirt_combined_ao * dirt_combined_ao;
    dirt_combined_ao = clamp(inv_dirt_ao, 1.0, 4.0) * dirt_combined_ao;

    float dirt_mix = max(dirt_combined_inv_ao, dirt_combined_ao);
    dirt_mix = max(dirt_mix, clamp(dirt_value, 0.0, 1.0));
    dirt_mix = max(dirt_mix, nad.w * dirt_detail_mask);

    float roughness = (dirt_mix * dirt_roughness_blend) * (dirt_roughness - combined_detail_data.w) + combined_detail_data.w;
    vec3 temp_dirt_color = gamma(abs(dirt_color), 2.2);

    vec3 base_color = max(dirt_mix, 0.001) * (temp_dirt_color - base_albedo) + base_albedo;

    float up_fade = max(up_fade_distance, 0.001);
    float up_fade_crv = max(up_fade_curve, 0.001);

    // what's up_val?
    float up_val = pow(clamp(obj_height / up_fade, 0.0, 1.0), up_fade_crv);

    vec3 tint_color = up_val * (top_color_tint - bottom_color_tint) + bottom_color_tint;
    base_color = tint_color * base_color;

    float ao = tiler_ao_intensity * (combined_detail_data.z - 1.0) + 1.0;
    ao = (base_ao_intensity * (nad.z - 1.0) + 1.0) * ao;

    vec3 primary_graded_color = base_color;
    if (grading_group_id > 0.0) {
        primary_graded_color = gradeColor(base_color, gradingMatrix(gradingPixel(grading_group_id)));
    }
    primary_graded_color = clamp(primary_graded_color, 0.0, 1.0);

    vec3 secondary_graded_color = base_color;
    if (grading_secondary_group_id > 0.0) {
        secondary_graded_color = gradeColor(base_color, gradingMatrix(gradingPixel(grading_secondary_group_id)));
    }
    secondary_graded_color = clamp(secondary_graded_color, 0.0, 1.0);
    base_color = nad.w * (secondary_graded_color - primary_graded_color) + primary_graded_color;

    base_color = dirt_mix * dirt_metallic * (dirt_color - base_color) + base_color;

    inv_ao = abs(1.0 - ao);
    float weathering_coverage_amount = clamp(weathering_input * weathering_coverage_multiplier + weathering_coverage, 0.0, 1.0);
    float weathering_height_val = pow(1.0 - clamp(obj_height / max(weathering_height, 0.001), 0.0, 1.0), max(weathering_height_curve, 0.001));

    if (weathering_height > 99.9) {
        weathering_height_val = 1.0;
    }
    weathering_height_val = weathering_height_val * weathering_coverage_amount;

    float heightmap_up_amount = clamp(heightmap_normal.z, -0.05, 1.0) + 0.05;
    float weathering_amount = clamp((2.5 * (heightmap_up_amount / 1.05 - (1.0 - pow(weathering_height_val, 0.7))) - 0.1) / 1.1, 0.0, 1.0);
    weathering_amount = clamp((1.0 - clamp(dirt_mix * weathering_on_dirt_amount + weathering_amount, 0.0, 1.0)) * weathering_thickness - weathering_thickness + 1.0, 0.0, 1.0);
    combined_normal = normalize(mixed_vertex_normal + combined_normal);
    combined_normal = weathering_amount * (combined_normal - heightmap_normal) + heightmap_normal;
    float combined_len = dot(combined_normal, combined_normal);
    combined_normal = normalize(combined_normal);
    if (combined_len <= 0) {
        combined_normal = vec3(0.0, 0.0, 1.0);
    }

    float combined_normal_up_amount = 2.5 * (((clamp(combined_normal.y, -0.125, 1.0) + 0.125) / 1.125) - (1 - weathering_height_val * weathering_height_val));
    combined_normal_up_amount = clamp(combined_normal_up_amount / (weathering_height_val * weathering_height_val + 0.001), 0.0, 1.0);

    float weathering_ao = clamp(((1.0 - ao) * weathering_ao_amount) * clamp(combined_normal.y, 0.0, 1.0) - weathering_ao_cutoff, 0.0, 1.0);

    dirt_mix = clamp(weathering_ao + (dirt_mix * weathering_on_dirt_amount + combined_normal_up_amount), 0.0, 1.0);

    if (dirt_mix != 0.0 && rock_type != 2 /*not snowy*/) {
        vec2 weathering_uvs = getCustomTriplanarUVs(combined_normal) * weathering_tiling;
        vec3 weathering_color = texture(weathering_albedo, weathering_uvs).xyz;
        vec4 weathering_nar_sample = texture(weathering_nar, weathering_uvs);

        vec3 weathering_normal = vec3(weathering_nar_sample.xy * 2.0 - 1.0, 0.0);
        vec2 weathering_sqnorm = weathering_normal.xy * weathering_normal.xy;
        weathering_normal.z = max(max(weathering_sqnorm.x, weathering_sqnorm.y) / 16384.0, 1.0 - weathering_normal.x * weathering_normal.x - weathering_normal.y * weathering_normal.y);
        weathering_normal.xz = weathering_normal.xy * -inversesqrt(weathering_normal.z);
        weathering_normal.y = 0.0;
        weathering_normal = getCustomTriplanarNormal(combined_normal, weathering_normal);
        float normal_align = dot(combined_normal, weathering_normal);
        weathering_normal = normalize(weathering_normal_mult * (weathering_normal - normal_align * combined_normal) + combined_normal);

        vec3 weathering_graded_color = weathering_color;
        if (grading_group_id_weathering > 0.0) {
            weathering_graded_color = gradeColor(weathering_color, gradingMatrix(gradingPixel(grading_group_id_weathering)));
        }
        weathering_color = clamp(weathering_graded_color, 0.0, 1.0);

        base_color = dirt_mix * (weathering_color - base_color) + base_color;
        roughness = dirt_mix * (weathering_nar_sample.w - roughness) + roughness;
        combined_normal = normalize(dirt_mix * (weathering_normal - combined_normal) + combined_normal);
        ao = min(dirt_mix * (weathering_nar_sample.w - 1.0) + 1.0, ao);
    }

    vec3 fuzz_color = vec3(0.0);
    vec3 subsurface_color = vec3(0.0);
    float subsurface_wrap = 0.5;
    float subsurface_intensity = 0.0;
    float subsurface_diffusion = 0.6;
    float subsurface_thickness = 0.075;
    float specular = 0.5;
    vec3 perturbed_normal = vec3(0.0, 0.0, 1.0);
    vec3 perturbed_triplanar = perturbed_normal;
    vec4 pnrb_sample;
    if (dirt_mix != 0.0 && rock_type == 2 /*snowy*/) {
        vec3 snow_uvs = vec3(getCustomTriplanarUVs(combined_normal) * weathering_tiling, snow_index0);
        pnrb_sample = texture(snow_pnrb_array, snow_uvs);

        perturbed_normal = vec3(pnrb_sample.xy * 2.0 - 1.0, 0.0);
        vec2 pnrb_sq = perturbed_normal.xy * perturbed_normal.xy;
        perturbed_normal.z = max(max(pnrb_sq.x, pnrb_sq.y) / 16384.0, 1.0 - perturbed_normal.x * perturbed_normal.x - perturbed_normal.y * perturbed_normal.y);
        perturbed_normal.xz = perturbed_normal.xy * -inversesqrt(perturbed_normal.z);
        perturbed_normal.y = 0.0;
        perturbed_normal = getCustomTriplanarNormal(combined_normal, perturbed_normal);
        perturbed_triplanar = perturbed_normal;
        perturbed_normal = perturbed_normal - dot(combined_normal, perturbed_normal) * combined_normal;

        vec2 glitter_uvs = getCustomTriplanarUVs(combined_normal) * snow_glint_size;
        vec3 glints_sample = texture(snow_glints_tiler, glitter_uvs).xyz * 2.0 - 1.0;
        vec3 glint_normal = vec3(glints_sample.xy, sqrt(abs(1.0 - glints_sample.x * glints_sample.x - glints_sample.y * glints_sample.y)));
        float glint_align = dot(glint_normal, vec3(inv_view[0].y, inv_view[1].y, inv_view[2].y));
        glint_align = clamp(glint_align - (1.0 - snow_glint_amount), 0.0, 1.0);
        glint_align = clamp(glint_align / (1.0 - snow_glint_amount), 0.0, 1.0);

        vec3 snow_color = gamma(max(abs(snow_base_color), vec3(0.0001)), pnrb_sample.z + 1.0);
        base_color = dirt_mix * (snow_color - base_color) + base_color;

        subsurface_color = dirt_mix * gamma(max(abs(snow_subsurface_color), vec3(0.0001)), pnrb_sample.z - pnrb_sample.w * 0.25 + 1.0);
        fuzz_color = dirt_mix * snow_fuzz_color;

        subsurface_intensity = dirt_mix * snow_subsurface_intensity;
        subsurface_wrap = dirt_mix * (snow_subsurface_wrap - 0.5) + 0.5;
        subsurface_diffusion = dirt_mix * (snow_subsurface_diffusion - 0.6) + 0.6;
        subsurface_thickness = dirt_mix * (snow_subsurface_thickness - 0.075) + 0.075;
        specular = dirt_mix * (snow_specular * pnrb_sample.w - 0.5) + 0.5;

        roughness = dirt_mix * (clamp(glints_sample.z * snow_glint_roughness + snow_roughness - glint_align, 0.0, 1.0) - roughness) + roughness;
        combined_normal = normalize(dirt_mix * dirt_mix * (normalize(combined_normal + perturbed_normal) - combined_normal) + combined_normal);

        ao = dirt_mix * 0.5 * (1.0 - ao) + ao;
    }

    vec3 tangentCombinedNormal = normalize(dbg_fragITBN * combined_normal);
    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    float view_align = max(dot(tangentCombinedNormal, viewDirection), 0.0001);
    float presnow_view_align = view_align;

    float metallic = material_wetness;
    roughness = clamp(roughness, 0.0, 1.0) * 0.955 + 0.045;

    float subsurface = 0.0;
    vec3 emissive_color = vec3(0.0);
    // Rock type is based on the index of the specific rock base material in the rockMaterials slice
    switch(rock_type) {
    case 1: // emissive
        emissive_color = emissive(base_color, roughness, ao, obj_height);
        break;
    case 2: // snow
        base_color = max(dirt_mix - view_align * dirt_mix, 0.0) * (clamp(fuzz_color, 0.0, 1.0) - base_color) + base_color;
        subsurface_wrap = clamp(subsurface_wrap, 0.0, 1.0);
        subsurface_intensity = clamp(subsurface_intensity, 0.0, 1.0);
        subsurface = subsurface_wrap * (1.0 - 2.0 * view_align) + view_align;
        subsurface = pow(abs(subsurface), abs(subsurface_wrap - 0.5) * 2.0 + 1.0);
        subsurface = (subsurface - abs(subsurface_wrap - 0.5) + 0.5) * subsurface_intensity;
        base_color = subsurface * (subsurface_color - base_color) + base_color;
        metallic = subsurface * (metallic - specular) + specular;
        break;
    case 3: // ice
        base_color = ice_fuzz_subsurface(base_color, view_align, inv_ao);
        break;
    }

    switch (debug) {
    case 1:
        fragColor = vec4(sqrt(base_color), 1.0);
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
        fragColor = vec4(vec3(subsurface), 1.0);
        break;
    case 6:
        fragColor = vec4(vec3(view_align), 1.0);
        break;
    case 7:
        fragColor = vec4(combined_normal, 1.0);
        break;
    case 8:
        fragColor = vec4(viewDirection, 1.0);
        break;
    case 9:
        fragColor = vec4(vec3(obj_height), 1.0);
        break;
    default:
        fragColor = calculateLighting(vec4(sqrt(base_color), 1.0), roughness, metallic, 0.0, ao, tangentCombinedNormal, 1.0, 1.0);
        fragColor += vec4(emissive_color, 0.0);
    }
}
