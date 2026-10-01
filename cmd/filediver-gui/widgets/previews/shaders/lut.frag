#version 430 core

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV0;
in vec2 fragUV1;
in vec2 fragUV2;
in vec3 fragTangentLightPosition; // tangent meaning in tangent space
in vec3 fragTangentViewPosition;
in vec3 fragTangentFragmentPosition;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

uniform sampler2D decal_sheet;
uniform sampler2DArray composite_array;
uniform sampler2DArray customization_camo_tiler_array;
uniform sampler2DArray customization_material_detail_tiler_array;
uniform sampler2D pattern_lut;
uniform sampler2D base_data;
uniform sampler2D material_lut;
uniform sampler2DArray pattern_masks_array;
uniform sampler2DArray id_masks_array;
uniform sampler2D ibl_brdf_lut;

layout(shared, binding = 0) uniform LutSettingsBlock {
    uint seed;
    bool use_decals;
    float damage_mask_selector;
    float detail_tile_factor_mult;
    float decal_id;
    float decal_id_offset_x;
    float decal_id_offset_y;
    float decal_size_x;
    float decal_size_y;
    float decal_lut_id;
    float decal_lut_size_x;
    float decal_lut_size_y;
    float decal_normal_intentity;
    float decal_normal_offset;
    vec2 decal_scalarfield_end;
    float decal_alpha_offset;
    float decal_alpha_sharpness;
    float decal_channel_selection;
    float use_decal_rgb_channels;
    float underlying_normal_behind_decal_opacity;
};

float reconstructNormalZ(vec2 xy) {
    return sqrt(1.0 - xy.x*xy.x - xy.y*xy.y);
}

float recoverInverseZ(vec2 normal_xy) {
    return inversesqrt(1 - (normal_xy.x * normal_xy.x) - (normal_xy.y * normal_xy.y));
}

vec3 rowColors[8] = {
    {0, 0, 0},
    {1, 0, 0},
    {0, 1, 0},
    {0, 0, 1},
    {1, 1, 0},
    {1, 0, 1},
    {0, 1, 1},
    {1, 1, 1},
};

uint getMaterialLutRow() {
    mat4 identity = mat4(1.0);
    uint id_masks_len = textureSize(id_masks_array, 0).z;
    uint id_mask_layer = 0;
    uint material_lut_row = 0;
    while(id_mask_layer < id_masks_len) {
        vec4 id_mask_sample = texture(id_masks_array, vec3(fragUV0, float(id_mask_layer)));
        uint row_base = id_mask_layer * 4;
        uint channel = 0;
        while(channel < 4) {
            if (dot(id_mask_sample, identity[channel]) >= 0.496) {
                material_lut_row = row_base + channel;
                break;
            }
            channel = channel + 1;
        }

        id_mask_layer = id_mask_layer + 1;
    }
    return material_lut_row;
}

vec3 sRGB(vec3 color) {
    vec3 temp = max(color, vec3(0.000061));
    color = temp * 12.92;
    temp = exp2(log2(max(temp, vec3(0.003131))) * 0.416667) * 1.055 - 0.055;
    return min(color, temp);
}

vec3 gamma(vec3 color, float value) {
    return exp2(log2(color) * vec3(value));
}

vec3 rgb2hsv(vec3 c)
{
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));

    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 hsv2rgb(vec3 c)
{
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

// base and detail should be unpacked already, and this will return an unpacked normal
vec3 reorientNormalMaps(vec3 base, vec3 detail) {
    vec3 t = base + vec3(0, 0, 1);
    vec3 u = detail * vec3(-1, -1, 1);
    return normalize(t*dot(t, u)/t.z - u);
}

void main() {
    float detail_tile_factor_mult = 1.0;

    vec4 base_data_sample = texture(base_data, fragUV0);
    vec3 normal = vec3(base_data_sample.xy * 2.0 - 1.0, 0);
    normal.z = reconstructNormalZ(normal.xy);
    normal.x = -normal.x;
    vec3 normal_modified = normal;
    float ao = 1 - (base_data_sample.z * 0.4 + 0.6);
    float base_data_roughness = base_data_sample.w;

    uint material_lut_row = getMaterialLutRow();

    vec4 base_color = texelFetch(material_lut, ivec2(0, material_lut_row), 0);
    vec4 detail_layer = texelFetch(material_lut, ivec2(1, material_lut_row), 0);
    vec4 lut_2 = texelFetch(material_lut, ivec2(2, material_lut_row), 0);
    vec4 lut_3 = texelFetch(material_lut, ivec2(3, material_lut_row), 0);
    vec4 lut_4 = texelFetch(material_lut, ivec2(4, material_lut_row), 0);
    vec4 lut_5 = texelFetch(material_lut, ivec2(5, material_lut_row), 0);
    vec4 lut_6 = texelFetch(material_lut, ivec2(6, material_lut_row), 0);
    vec4 metallic_detail_controls = texelFetch(material_lut, ivec2(7, material_lut_row), 0);
    vec4 specular_detail_controls = texelFetch(material_lut, ivec2(8, material_lut_row), 0);
    vec4 lut_10 = texelFetch(material_lut, ivec2(10, material_lut_row), 0);
    vec4 lut_11 = texelFetch(material_lut, ivec2(11, material_lut_row), 0);
    float roughness = lut_10.x;
    vec4 camo_controls = texelFetch(material_lut, ivec2(21, material_lut_row), 0);
    vec4 detail_tiling = texelFetch(material_lut, ivec2(22, material_lut_row), 0);

    // overwritten by composite_array, if used
    vec4 material_detailing_xy = vec4(0);

    vec4 material_detailing_zw = vec4(0);
    vec4 material_detail_sample = vec4(0);

    // r12.w
    float detail_normal_2_alignment;
    vec4 material_detail_sample_2 = vec4(0);
    uvec2 random_detail_offset = uvec2(seed, seed >> 9);
    uvec2 bitmask = (((uvec2(1) << uvec2(23)) - uvec2(1)) << uvec2(0)) & 0xffffffff;
    random_detail_offset = (random_detail_offset & bitmask.xy) | (uvec2(0x3f800000) & ~bitmask.xy);
    vec2 detail_offset = uintBitsToFloat(random_detail_offset) - 1;

    uint material_detail_len = textureSize(customization_material_detail_tiler_array, 0).z;
    if (base_color.w <= 3.0) {
        float detail_scaling = detail_tiling.x * detail_tile_factor_mult;
        vec3 detail_uvs = vec3(fragUV1 * detail_scaling + detail_offset, float(uint(detail_layer.x) % material_detail_len));
        material_detail_sample = texture(customization_material_detail_tiler_array, detail_uvs) - 0.5;
        material_detailing_xy.xy = material_detail_sample.zw;
        material_detailing_zw.xy = material_detail_sample.xy * vec2(-1, 1);
    }

    if (base_color.w >= 2.0 && lut_10.z > 0.0) {
        vec3 detail_uvs_2 = vec3(detail_tiling.x * detail_tile_factor_mult * fragUV1 + detail_offset, float(uint(lut_10.y) % material_detail_len));
        material_detail_sample_2 = texture(customization_material_detail_tiler_array, detail_uvs_2) - 0.5;
        // r16.zw
        material_detailing_zw.zw = material_detail_sample_2.xy * vec2(-2, 2);
        vec2 temp_sq = material_detailing_zw.zw * material_detailing_zw.zw;
        float material_detailing_normal_z = max(max(temp_sq.x, temp_sq.y) * 0.000061, reconstructNormalZ(material_detailing_zw.zw));
        material_detailing_zw.zw = -inversesqrt(material_detailing_normal_z) * material_detailing_zw.zw * lut_10.w;
        normal_modified = normalize(vec3(material_detailing_zw.zw + normal_modified.xy, normal_modified.z * material_detailing_normal_z));
        material_detail_sample_2.xyz = vec3(material_detail_sample_2.zw, clamp((base_data_roughness - 0.5), 0.0, 1.0));
        detail_normal_2_alignment = dot(lut_11.xyz, material_detail_sample_2.xyz) + lut_11.w;
    } else {
        lut_10.z = 0.0;
        detail_normal_2_alignment = 0.0;
    }

    if (base_color.w == 4.0) {
        vec3 composite_uvs = vec3(detail_tiling.x * detail_tile_factor_mult * 4.0 * fragUV1, float(uint(detail_layer.x) % material_detail_len /*not composite length, yes*/));
        vec3 composite_sample = texture(composite_array, composite_uvs).xyz;
        material_detailing_xy.xyz = composite_sample.zxy - 0.5;
        material_detailing_xy.w = -(composite_sample.x - 0.5);
        material_detailing_zw.zw = vec2(-(composite_sample.x - 0.5), composite_sample.y - 0.5);
        material_detailing_xy.y = 0.0;
    } else {
        material_detailing_zw.zw = material_detailing_zw.xy;
    }

    vec4 scaled_lut_4 = vec4(lut_4.xy * max((base_data_roughness - 0.5) * lut_4.z, 0.0) * 4, 0.0, 0.0);
    lut_4 = lut_4 + scaled_lut_4;
    material_detailing_xy.zw = base_data_sample.wz * vec2(1, -0.4) + vec2(-0.5, 0.4);
    // r15.z
    float detail_roughness = clamp(dot(lut_4, material_detailing_xy) + detail_layer.w, 0.0, 1.0);
    
    vec4 scaled_lut_3 = vec4(lut_3.xy * max((base_data_roughness - 0.5) * lut_3.z, 0.0) * 4, 0.0, 0.0);
    lut_3 = lut_3 + scaled_lut_3;
    vec4 modified_material_detail_tiling = material_detailing_xy;
    float temp_r15w = clamp(material_detailing_xy.y + 0.5, 0.0, 1.0);
    modified_material_detail_tiling.w = ao * (ao * (1 - temp_r15w) + temp_r15w);
    // r15.w
    float detail_intensity = clamp(dot(lut_3, modified_material_detail_tiling) + detail_layer.z, 0.0, 1.0);

    vec4 scaled_metallic_detail_controls = vec4(metallic_detail_controls.xy * max((base_data_roughness - 0.5) * metallic_detail_controls.z, 0.0) * 4, 0.0, 0.0);
    metallic_detail_controls = metallic_detail_controls + scaled_metallic_detail_controls;
    // r15.x
    float detail_metallic = clamp(lut_6.w + dot(metallic_detail_controls, material_detailing_xy), 0.0, 1.0);

    vec3 scaled_specular_detail_controls = vec3(specular_detail_controls.xy * max((base_data_roughness - 0.5) * specular_detail_controls.z, 0.0) * 4, 0.0);
    scaled_specular_detail_controls = scaled_specular_detail_controls + specular_detail_controls.xyz;
    // r15.y
    float detail_specular = clamp(specular_detail_controls.w + dot(scaled_specular_detail_controls, material_detailing_xy.xyz), 0.0, 1.0);

    vec2 temp_detailing;
    if (base_color.w <= 3.0) {
        lut_2.w = detail_intensity * (lut_2.w - 1.0) + 1.0;
        lut_2.w = detail_roughness * (lut_5.w - lut_2.w) + lut_2.w;

        temp_detailing = material_detailing_zw.xy + material_detailing_zw.xy;
        vec2 temp_detailing_sq = temp_detailing * temp_detailing;
        float detailing_z = 1 - temp_detailing.x * temp_detailing.x - temp_detailing.y * temp_detailing.y;
        detailing_z = inversesqrt(max(max(temp_detailing_sq.x, temp_detailing_sq.y) * 0.000061, detailing_z));
        temp_detailing = temp_detailing * -detailing_z * detail_layer.y * lut_2.w;

        normal_modified = normalize(vec3(temp_detailing + normal_modified.xy, normal_modified.z * detailing_z));
    }

    vec3 camo_sample = vec3(0);
    vec4 camo_result = vec4(0);
    vec3 camo_combined_1 = vec3(0);
    float camo_mix = 0;

    if (camo_controls.w >= 0.0) {
        vec4 camo_color_1 = texelFetch(material_lut, ivec2(16, material_lut_row), 0);
        vec4 camo_color_2 = texelFetch(material_lut, ivec2(17, material_lut_row), 0);
        vec4 camo_color_3 = texelFetch(material_lut, ivec2(18, material_lut_row), 0);
        vec4 camo_color_4 = texelFetch(material_lut, ivec2(19, material_lut_row), 0);
        float camo_control_extra = texelFetch(material_lut, ivec2(20, material_lut_row), 0).w;

        vec3 camo_uvs = vec3(detail_tile_factor_mult * camo_controls.z * fragUV1, camo_controls.w);
        camo_sample = clamp((texture(customization_camo_tiler_array, camo_uvs).xyz - 0.5) * camo_controls.xxx + camo_controls.yyy, 0.0, 1.0);
        camo_result = mix(camo_color_1, camo_color_2, camo_sample.x);
        camo_result = mix(camo_result, camo_color_3, camo_sample.y);
        camo_result = mix(camo_result, camo_color_4, camo_sample.z);

        if (base_color.w != 1.0) {
            float camo_edge_wear = max((base_data_roughness - 0.5) * camo_color_4.w, 0.0) * 4;
            camo_combined_1 = vec3(camo_edge_wear * camo_color_2.w, camo_edge_wear * camo_color_3.w, 0.0);
            vec3 camo_combined_2 = vec3(camo_color_2.w, camo_color_3.w, camo_color_4.w);
            camo_combined_1 = camo_combined_1 + camo_combined_2;
            camo_mix = clamp(dot(camo_combined_1, material_detailing_xy.xyz) - camo_control_extra, 0.0, 1.0);
            base_color.xyz = mix(camo_result.xyz, base_color.xyz, camo_mix);
            detail_metallic = max(detail_metallic - (1.0 - camo_mix), 0.0);
            float camo_tiling = detail_tiling.w * (1.0 - camo_mix);
            roughness = camo_tiling * (camo_color_1.w - roughness) + roughness;
        } else {
            base_color.xyz = camo_result.xyz;
            camo_mix = 1.0;
        }
    } else {
        camo_mix = 1.0;
    }

    // pattern mask
    if (base_color.w != 3.0) {
        vec4 pattern_color = texelFetch(pattern_lut, ivec2(0, 0), 0);
        if (pattern_color.w != -1.0) {
            int depth = textureSize(pattern_masks_array, 0).z;
            vec4 pattern_controls_1 = texelFetch(pattern_lut, ivec2(1, 0), 0);
            float pattern_edge_wear = max((base_data_roughness - 0.5) * pattern_controls_1.z, 0.0) * 4.0;
            vec3 pattern_normal = vec3(pattern_controls_1.xy * pattern_edge_wear, 0.0) + pattern_controls_1.xyz;
            float pattern_mix = clamp(dot(pattern_normal, material_detailing_xy.xyz) + pattern_controls_1.w, 0.0, 1.0);
            vec3 pattern_mask_uv = vec3(fragUV0, float(int(pattern_color.w) % depth));
            float mask_sample = clamp(100 * (texture(pattern_masks_array, pattern_mask_uv).x - 0.5), 0.0, 1.0);
            float pattern_mix_masked = pattern_mix * mask_sample;
            if (pattern_mix_masked > 0.0) {
                base_color.xyz = mix(base_color.xyz, pattern_color.xyz, pattern_mix_masked);
                if (base_color.w != 1.0) {
                    float pattern_roughness = min(detail_roughness - (pattern_mix_masked * pattern_mix_masked) + 1.0, 1.0);
                    detail_metallic = detail_metallic * pattern_roughness;
                    float pattern_control_2 = texelFetch(pattern_lut, ivec2(2, 0), 0).w;
                    pattern_mix = max(pattern_mix * mask_sample - detail_roughness, 0.0) * (1.0 / (2.0 - roughness));
                    roughness = pattern_mix * (pattern_roughness - roughness) + roughness;
                    detail_specular = pattern_mix_masked * (0.5 - detail_specular) + detail_specular;
                }
            }
        }
    }

    bvec2 decal_range = lessThan(abs(fragUV2 - 0.5), vec2(0.5));
    if(use_decals && decal_range.x && decal_range.y) {
        vec4 decal_sheet_size = vec4(textureSize(decal_sheet, 0).xy, 0, 0);
        decal_sheet_size.zw = max(decal_sheet_size.xy, vec2(1.0));
        
        // note: spelled as "decal_normal_intentity" in files
        float decal_normal_intensity = decal_normal_intentity;
        float decal_id_val = max(floor(decal_id + 0.5), 0.0);

        // r27
        vec4 decal_size_xyzw = vec4(max(decal_size_x, 1.0), max(decal_size_y, 1.0), decal_sheet_size.zw * fragUV2);
        vec2 decal_offset = fragUV2 * decal_sheet_size.zw + (decal_id_val * vec2(decal_id_offset_x, -decal_id_offset_y));
        decal_offset = decal_offset / decal_sheet_size.zw;

        vec2 decal_lut_size = vec2(decal_lut_size_x, decal_lut_size_y);
        if (decal_lut_size_x > 1.0 || decal_lut_size_y > 1.0) {
            decal_lut_size = decal_lut_size  / decal_sheet_size.zw;
        }
        vec2 temp = min(max(vec2(1.0) / decal_sheet_size.zw, decal_lut_size), vec2(1.0));
        decal_lut_size = vec2(1) - temp;
        temp = temp * vec2(1.0/6.0, 1.0/3.0);

        float decal_lut_id_val = clamp(floor(decal_lut_id + 0.5), 0.0, 5.0);

        vec4 decal_sample = texture(decal_sheet, fragUV2);

        bvec2 decal_size_scaled_down = lessThan(decal_size_xyzw.zw, decal_size_xyzw.xy);
        if (decal_size_scaled_down.x && decal_size_scaled_down.y) {
            decal_sample = texture(decal_sheet, decal_offset);
            if (use_decal_rgb_channels > 0.5) {
                float channel = (decal_channel_selection == 2.0) ? decal_sample.z : decal_sample.x;
                channel = (decal_channel_selection == 1.0) ? decal_sample.y : channel;

                float decal_lut_id_plus_half = decal_lut_id_val + 0.5;
                vec4 decal_lut_offset = vec4(decal_lut_id_plus_half, 0.5, decal_lut_id_plus_half, 1.5);
                decal_lut_offset = decal_lut_offset * temp.xyxy + decal_lut_size.xyxy;

                vec2 decal_lut_offset_2 = vec2(decal_lut_id_plus_half, 2.5);
                decal_lut_offset_2 = decal_lut_offset_2 * temp.xy + decal_lut_size.xy;

                vec3 decal_red_sample = texture(decal_sheet, decal_lut_offset.xy).xyz;
                vec3 decal_green_sample = texture(decal_sheet, decal_lut_offset.zw).xyz;
                vec3 decal_blue_sample = texture(decal_sheet, decal_lut_offset_2.xy).xyz;

                float channel_mix_1 = (channel < 1.0/3.0) ? 1.0 : 0.0;
                float channel_mix_2 = (channel < 2.0/3.0 && channel >= 1.0/3.0) ? 1.0 : 0.0;
                float channel_mix_3 = (channel >= 2.0/3.0) ? 1.0 : 0.0;

                decal_sample.xyz = channel_mix_2 * decal_red_sample + (channel_mix_1 * decal_blue_sample + (channel_mix_3 * decal_green_sample));
            }
        }

        bvec2 decal_uv_in_scalarfield = lessThan(vec2(fragUV2.x, 1 - fragUV2.y), decal_scalarfield_end);
        if (decal_uv_in_scalarfield.x && decal_uv_in_scalarfield.y) {
            float inv_normal_offset = max(1 - decal_normal_offset, 0.001);

            vec2 offset = vec2(0.5) / decal_sheet_size.xy;

            float decal_sheet_alpha_neg_x = texture(decal_sheet, vec2(fragUV2.x - offset.x, fragUV2.y)).w;
            decal_sheet_alpha_neg_x = decal_sheet_alpha_neg_x - decal_normal_offset;
            decal_sheet_alpha_neg_x = clamp(decal_sheet_alpha_neg_x / inv_normal_offset, 0.0, 1.0);

            float decal_sheet_alpha_pos_x = texture(decal_sheet, vec2(fragUV2.x + offset.x, fragUV2.y)).w;
            decal_sheet_alpha_pos_x = decal_sheet_alpha_pos_x - decal_normal_offset;
            decal_sheet_alpha_pos_x = clamp(decal_sheet_alpha_pos_x / inv_normal_offset, 0.0, 1.0);

            float decal_sheet_alpha_neg_y = texture(decal_sheet, vec2(fragUV2.x, fragUV2.y - offset.y)).w;
            decal_sheet_alpha_neg_y = decal_sheet_alpha_neg_y - decal_normal_offset;
            decal_sheet_alpha_neg_y = clamp(decal_sheet_alpha_neg_y / inv_normal_offset, 0.0, 1.0);

            float decal_sheet_alpha_pos_y = texture(decal_sheet, vec2(fragUV2.x, fragUV2.y + offset.y)).w;
            decal_sheet_alpha_pos_y = decal_sheet_alpha_pos_y - decal_normal_offset;
            decal_sheet_alpha_pos_y = clamp(decal_sheet_alpha_pos_y / inv_normal_offset, 0.0, 1.0);

            float temp_alpha_x = decal_sheet_alpha_pos_x - decal_sheet_alpha_neg_x;
            float temp_alpha_y = decal_sheet_alpha_pos_y - decal_sheet_alpha_neg_y;
            vec2 temp_intensity = vec2(temp_alpha_x, temp_alpha_y) * decal_normal_intensity;

            float offset_alpha = decal_sample.w - decal_alpha_offset;
            decal_sample.w = clamp((offset_alpha / inv_normal_offset) * decal_alpha_sharpness, 0.0, 1.0);

            float normal_opacity = (underlying_normal_behind_decal_opacity - 1.0) * decal_sample.w + 1.0;
            vec3 temp_normal_modified = normal_modified * normal_opacity;

            float inv_normal_opacity = (1.0 - underlying_normal_behind_decal_opacity) * decal_sample.w;

            vec3 result_normal = -normal_modified * normal_opacity + normal;
            normal_modified = inv_normal_opacity * result_normal + temp_normal_modified;
        }

        float decal_mix = clamp(decal_sample.w - detail_roughness, 0.0, 1.0);
        vec3 temp_color = decal_sample.xyz - base_color.xyz;
        base_color.xyz = temp_color * decal_mix + base_color.xyz;
        detail_metallic = detail_metallic * (1.0 - decal_mix);
        roughness = max(decal_mix - detail_intensity, 0.0) * (0.4 - roughness) + roughness;
    }

    lut_2.xyz = detail_intensity * (lut_2.xyz - base_color.xyz) + base_color.xyz;
    roughness = detail_tiling.z * detail_intensity * (detail_tiling.y - roughness) + roughness;
    float camo_roughness = detail_roughness * camo_mix;
    lut_2.xyz = camo_roughness * (lut_6.xyz - lut_2.xyz) + lut_2.xyz;
    camo_roughness = camo_roughness * detail_roughness;
    lut_2.xyz = camo_roughness * (lut_5.xyz - lut_2.xyz) + lut_2.xyz;
    lut_2.xyz = max(lut_2.xyz, vec3(0.000061));
    bvec3 notVerySmall = lessThan(vec3(0.04045), lut_2.xyz);
    vec3 srgb_temp = lut_2.xyz * vec3(0.947867) + vec3(0.052133);
    srgb_temp = gamma(srgb_temp, 2.4);
    lut_2.xyz = lut_2.xyz * vec3(0.0774);
    vec3 temp;
    temp.x = notVerySmall.x ? srgb_temp.x : lut_2.x;
    temp.y = notVerySmall.y ? srgb_temp.y : lut_2.y;
    temp.z = notVerySmall.z ? srgb_temp.z : lut_2.z;
    lut_2.xyz = temp;

    // r17.z - 0.5
    float dirt_ao = 0.5;
    dirt_ao = clamp(dirt_ao * 2, 0.0, 1.0);
    vec3 dirty_color = lut_2.xyz * dirt_ao * vec3(0.7);
    dirty_color = lut_2.xyz * vec3(0.3) + dirty_color;

    float dirt_roughness = 0.7;
    vec4 weathering_special = vec4(0.5);
    float weathering_dirt_roughness = (clamp(weathering_special.z * 0.5 + (dirt_roughness - clamp(base_data_roughness - 0.5, 0.0, 1.0) * 5.0), 0.0, 1.0) - 0.5) * 5.0;
    float weathering_amount = 1.0;

    uint material = 1 << material_lut_row;
    bool damage_mask_selection = (uint(damage_mask_selector) & material) > 0;
    // r13.z (also probably only for vehicles, so...)
    float damage_weathering_amount = damage_mask_selection ? clamp(weathering_amount, 0.0, 1.0) : 0.0;

    weathering_dirt_roughness = damage_weathering_amount * min(clamp(weathering_dirt_roughness * clamp(2 * (weathering_amount - 0.5), 0.0, 1.0), 0.0, 1.0), 0.75);

    base_color.xyz = weathering_dirt_roughness * (dirty_color - lut_2.xyz) + lut_2.xyz;
    detail_specular = detail_specular - weathering_dirt_roughness * detail_specular;

    roughness = clamp(weathering_dirt_roughness * (0.8 - roughness) + roughness, 0.0, 1.0);

    base_color.xyz = sRGB(base_color.xyz);

    normal_modified = normalize(normal_modified);

    // perform basic lighting calcs

    vec3 ambient = vec3(0.5) * base_data_sample.z;
    vec3 lightDirection = normalize(fragTangentLightPosition - fragTangentFragmentPosition);
    vec3 lightColor = vec3(0.7);
    vec3 diffuse = max(dot(normal_modified, lightDirection), 0) * lightColor;

    vec3 specularColor = mix(vec3(0.04), base_color.rgb, detail_metallic);

    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    vec3 reflectDirection = reflect(-lightDirection, normal_modified);
    vec3 halfwayDirection = normalize(lightDirection + viewDirection);

    // blinn specular reflectance
    float NdH = max(dot(normal_modified, halfwayDirection), 0.001);
    float HdV = max(dot(halfwayDirection, viewDirection), 0.001);
    float NdL = max(dot(normal_modified, lightDirection), 0.0);
    float NdV = max(dot(normal_modified, viewDirection), 0.001);

    vec3 specularFresnel = mix(specularColor, vec3(0.7), pow(1.01 - HdV, 5.0));
    float k = 1.999 / (roughness * roughness);

    vec3 blinnSpecularRef = min(1.0, 3.0 * 0.0398 * k) * pow(NdH, min(10000.0, k)) * specularFresnel * vec3(NdL);
    vec3 diffRef = (vec3(1.0) - specularFresnel) * (1.0 / 3.1415926) * NdL;

    vec3 reflectedLight = blinnSpecularRef * lightColor;
    vec3 diffuseLight = diffRef * lightColor;

    // ibl lighting
    vec2 brdf = texture2D(ibl_brdf_lut, vec2(roughness, NdV)).xy * 0.66;
    vec3 iblspec = min(vec3(0.99), mix(specularColor, vec3(1.0), pow(1.01 - NdV, 5.0))) * brdf.x + brdf.y;
    reflectedLight += iblspec * ambient;
    diffuseLight += ambient * (1.0 / 3.1415926);

    vec3 specular = pow(NdH, 32.0) * lightColor;
    // adjust saturation
    base_color.rgb = rgb2hsv(base_color.rgb);
    base_color.g *= 1.30;
    base_color.rgb = hsv2rgb(base_color.rgb); 
    // adjust contrast
    base_color.rgb = (base_color.rgb - 0.5) * 1.6 + 0.5;

    //fragColor = vec4(vec3(base_color.w == 4.0), 1.0);
    //fragColor = vec4(reflectDirection, 1.0);
    //fragColor = vec4(material_detail_sample.xyz, 1.0);
    //fragColor = vec4(base_color.rgb * (mix(ambient, diffuse, 0.6) + 0.5 * specular), 1.0);
    fragColor = vec4(diffuseLight * mix(base_color.rgb, vec3(0.0), detail_metallic) + reflectedLight, 1.0);
}