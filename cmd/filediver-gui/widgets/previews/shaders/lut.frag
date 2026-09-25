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
uniform sampler2DArray customization_camo_tiler_array;
uniform sampler2DArray customization_material_detail_tiler_array;
uniform sampler2D pattern_lut;
uniform sampler2D base_data;
uniform sampler2D material_lut;
uniform sampler2DArray pattern_masks_array;
uniform sampler2DArray id_masks_array;

layout(shared, binding = 0) uniform LutSettingsBlock {
    uint seed;
    bool use_decals;
    float detail_tiler_factor_mult;
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

void main() {
    float detail_tiler_factor_mult = 1.0;

    vec4 base_data_sample = texture(base_data, fragUV0);
    vec3 normal = vec3(base_data_sample.xy * 2.0 - 1.0, 0);
    normal.z = reconstructNormalZ(normal.xy);
    normal.x = -normal.x;
    vec3 normal_modified = normal;
    float ao = 1 - (base_data_sample.z * 0.4 + 0.6);
    float roughness = base_data_sample.w;

    uint material_lut_row = getMaterialLutRow();

    vec4 base_color = texelFetch(material_lut, ivec2(0, material_lut_row), 0);
    vec4 detail_layer = texelFetch(material_lut, ivec2(1, material_lut_row), 0);
    vec4 lut_2 = texelFetch(material_lut, ivec2(2, material_lut_row), 0);
    vec4 lut_3 = texelFetch(material_lut, ivec2(3, material_lut_row), 0);
    vec4 lut_4 = texelFetch(material_lut, ivec2(4, material_lut_row), 0);
    vec4 lut_5 = texelFetch(material_lut, ivec2(5, material_lut_row), 0);
    float lut_10_x = texelFetch(material_lut, ivec2(10, material_lut_row), 0).x;
    vec4 camo_controls = texelFetch(material_lut, ivec2(21, material_lut_row), 0);
    vec4 detail_tiling = texelFetch(material_lut, ivec2(22, material_lut_row), 0);

    vec4 material_detailing_xy = vec4(0);
    vec4 material_detailing_zw = vec4(0);
    vec4 material_detail_sample = vec4(0);

    if (base_color.w <= 3.0) {
        uint material_detail_len = textureSize(customization_material_detail_tiler_array, 0).z;
        uvec2 random_detail_offset = uvec2(seed, seed >> 9);
        uvec2 bitmask = (((uvec2(1) << uvec2(23)) - uvec2(1)) << uvec2(0)) & 0xffffffff;
        random_detail_offset = (random_detail_offset & bitmask.xy) | (uvec2(0x3f800000) & ~bitmask.xy);
        vec2 detail_offset = uintBitsToFloat(random_detail_offset) - 1;
        float detail_scaling = detail_tiling.x * detail_tiler_factor_mult;
        vec3 detail_uvs = vec3(fragUV1 * detail_scaling + detail_offset, float(uint(detail_layer.x) % material_detail_len));
        material_detail_sample = texture(customization_material_detail_tiler_array, detail_uvs) - 0.5;
        material_detailing_xy.xy = material_detail_sample.zw;
        material_detailing_zw.xy = material_detail_sample.xy * vec2(-1, 1);
    }

    vec4 scaled_lut_4 = vec4(lut_4.xy * max((roughness - 0.5) * lut_4.z, 0.0) * 4, 0.0, 0.0);
    lut_4 = lut_4 + scaled_lut_4;
    material_detailing_xy.zw = base_data_sample.wz * vec2(1, -0.4) + vec2(-0.5, 0.4);
    float detail_roughness = clamp(dot(lut_4, material_detailing_xy) + detail_layer.w, 0.0, 1.0);
    
    vec4 scaled_lut_3 = vec4(lut_3.xy * max((roughness - 0.5) * lut_3.z, 0.0) * 4, 0.0, 0.0);
    lut_3 = lut_3 + scaled_lut_3;
    vec4 modified_material_detail_tiling = material_detailing_xy;
    float temp_r15w = clamp(material_detailing_xy.y + 0.5, 0.0, 1.0);
    modified_material_detail_tiling.w = ao * (ao * (1 - temp_r15w) + temp_r15w);
    // r15.w
    float detail_intensity = clamp(dot(lut_3, modified_material_detail_tiling) + detail_layer.z, 0.0, 1.0);

    if (base_color.w <= 3.0) {
        lut_2.w = detail_intensity * (lut_2.w - 1.0) + 1.0;
        lut_2.w = detail_roughness * (lut_5.w - lut_2.w) + lut_2.w;

        vec2 temp_detailing = material_detailing_zw.xy + material_detailing_zw.xy;
        vec2 temp_detailing_sq = temp_detailing * temp_detailing;
        float detailing_z = recoverInverseZ(temp_detailing);
        temp_detailing = temp_detailing * -detailing_z * detail_layer.y * lut_2.w;

        normal_modified = normalize(dbg_fragITBN * (temp_detailing.y * dbg_fragTBN[1] + (temp_detailing.y * dbg_fragTBN[0])) + normal_modified);
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

        vec3 camo_uvs = vec3(detail_tiler_factor_mult * camo_controls.z * fragUV1, camo_controls.w);
        camo_sample = clamp((texture(customization_camo_tiler_array, camo_uvs).xyz - 0.5) * camo_controls.xxx + camo_controls.yyy, 0.0, 1.0);
        camo_result = mix(camo_color_1, camo_color_2, camo_sample.x);
        camo_result = mix(camo_result, camo_color_3, camo_sample.y);
        camo_result = mix(camo_result, camo_color_4, camo_sample.z);

        if (base_color.w != 1.0) {
            float camo_edge_wear = max((roughness - 0.5) * camo_color_4.w, 0.0) * 4;
            camo_combined_1 = vec3(camo_edge_wear * camo_color_2.w, camo_edge_wear * camo_color_3.w, 0.0);
            vec3 camo_combined_2 = vec3(camo_color_2.w, camo_color_3.w, camo_color_4.w);
            camo_combined_1 = camo_combined_1 + camo_combined_2;
            camo_mix = clamp(dot(camo_combined_1, material_detailing_xy.xyz) - camo_control_extra, 0.0, 1.0);
            base_color.xyz = mix(camo_result.xyz, base_color.xyz, camo_mix);
        } else {
            base_color.xyz = camo_result.xyz;
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
    }

    // perform basic lighting calcs

    vec3 ambient = vec3(1.0);
    vec3 lightDirection = normalize(fragTangentLightPosition - fragTangentFragmentPosition);
    vec3 lightColor = vec3(0.7);
    vec3 diffuse = max(dot(normal_modified, lightDirection), 0) * lightColor;

    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    vec3 reflectDirection = reflect(-lightDirection, normal_modified);
    vec3 halfwayDirection = normalize(lightDirection + viewDirection);
    vec3 specular = pow(max(dot(normal_modified, halfwayDirection), 0.0), 32.0) * lightColor;

    //fragColor = vec4(vec3(camo_mix), 1.0);
    //fragColor = vec4(reflectDirection, 1.0);
    //fragColor = vec4(material_detail_sample.xyz, 1.0);
    fragColor = vec4(base_color.rgb * (mix(ambient, diffuse, 0.6) + 0.5 * specular), 1.0);
}