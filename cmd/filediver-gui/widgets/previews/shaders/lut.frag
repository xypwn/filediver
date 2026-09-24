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

float reconstructNormalZ(vec2 xy) {
    return sqrt(1.0 - xy.x*xy.x - xy.y*xy.y);
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
    vec4 base_data_sample = texture(base_data, fragUV0);
    vec3 normal = vec3(base_data_sample.xy * 2.0 - 1.0, 0);
    normal.z = reconstructNormalZ(normal.xy);
    normal.x = -normal.x;
    float ao = base_data_sample.z * 0.4 + 0.6;
    float roughness = base_data_sample.w;

    uint material_lut_row = getMaterialLutRow();

    vec4 base_color = texelFetch(material_lut, ivec2(0, material_lut_row), 0);

    // perform basic lighting calcs

    vec3 ambient = vec3(1.0);
    vec3 lightDirection = normalize(fragTangentLightPosition - fragTangentFragmentPosition);
    vec3 lightColor = vec3(0.7);
    vec3 diffuse = max(dot(normal, lightDirection), 0) * lightColor;

    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    vec3 reflectDirection = reflect(-lightDirection, normal);
    vec3 halfwayDirection = normalize(lightDirection + viewDirection);
    vec3 specular = pow(max(dot(normal, halfwayDirection), 0.0), 32.0) * lightColor;

    fragColor = vec4(base_color.rgb * (mix(ambient, diffuse, 0.6) + 0.5 * specular), 1.0);
}