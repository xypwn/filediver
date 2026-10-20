#version 430 core

#include "lighting.frag"
#include "triplanar.frag"

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV0;
in mat3 dbg_fragITBN;
in mat4 instanceModel;

uniform sampler2D distortion_map;
uniform sampler2D dirt_map;
uniform sampler2D light_bleed_map;
uniform sampler2D normal_map;
uniform sampler2D roughness_map;
uniform sampler2D input_image;

layout(shared, binding = 0) uniform SignBlock {
    vec2 Atlas_size;
    float artifacts_amount;
    float dirt_amount;
    vec3 dirt_color;
    float distortion_amount;
    float emissive_amount;
    float light_bleed_amount;
    float material_wetness;
    float normal_map_amount;
    float normal_map_coloring_amount;
    vec2 resolution_multiplier;
    float screen_enabled;
    vec4 screen_rect;
    float use_dropshadow;
};

void main() {
    vec3 normal;
    normal.xy = texture(normal_map, fragUV0).xy * 2.0 - 1.0;
    normal.z = sqrt(1.0 - min(dot(normal.xy, normal.xy), 1.0)) - 1.0;
    normal = normal_map_amount * normal + vec3(0.0, 0.0, 1.0);

    float dirt;
    dirt = texture(dirt_map, fragUV0).x;
    //dirt = clamp(dirt.x * dirt_amount * dirt_color.xyz, 0.0, 1.0);

    float roughness = clamp(texture(roughness_map, fragUV0).x, 0.0, 1.0) * 0.955 + 0.045;
    float metallic = clamp(material_wetness - 0.5, 0.0, 1.0);

    vec2 screen_rect_uvs = ((fragUV0 - 0.5) * resolution_multiplier + 0.5) * screen_rect.zw + screen_rect.xy;
    vec2 atlas_uvs = floor((instanceModel[3].x + instanceModel[3].z + fd_world_offset.x + fd_world_offset.z) * vec2(123.321, 321.123)) / Atlas_size;
    bvec2 atlas_gte = greaterThanEqual(atlas_uvs, -atlas_uvs);
    atlas_uvs = fract(abs(atlas_uvs));

    if (!atlas_gte.x) {
        atlas_uvs.x = -atlas_uvs.x;
    }

    if (!atlas_gte.y) {
        atlas_uvs.y = -atlas_uvs.y;
    }

    screen_rect_uvs = (atlas_uvs * Atlas_size + screen_rect_uvs) / Atlas_size;
    // todo: add some distortion stuff here based on time
    vec3 color = texture(input_image, screen_rect_uvs).xyz;
    // color = noise * 0.01 + color;

    vec3 light_bleed = texture(light_bleed_map, fragUV0).xyz;
    color = light_bleed * light_bleed_amount + color;
    color = abs(normal.x) * normal_map_coloring_amount + color;
    color = color * screen_enabled;
    color = gamma(color, 2.2) * (1.0 - dirt * dirt_amount) * emissive_amount;

    if (use_dropshadow != 0.0) {
        vec3 dropshadow = texture(input_image, screen_rect_uvs + 0.01).xyz * emissive_amount * screen_enabled * 0.1;
        color = max(color, dropshadow);
    }

    fragColor = calculateLighting(vec4(color, 1.0), roughness, metallic, 1.0, normal, 1.0, 1.0);
}