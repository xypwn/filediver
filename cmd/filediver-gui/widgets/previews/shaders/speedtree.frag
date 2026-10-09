#version 430 core

#include "lighting.frag"

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

layout(shared, binding = 0) uniform SpeedtreeBlock {
    float grading_group_id;
    float grading_group_id_secondworld;
    float grading_group_id_thirdworld;
    float grading_group_id_trunk;
    float world_grading_color_value_variation;
    float ss_intensity_mult;
    float world;
    float opacity_threshold;
};

uniform sampler2D tex0;
uniform sampler2D tex1;
uniform sampler2D tex2;
uniform samplerBuffer asset_grading_lut;

int group(vec3 ids) {
    float selectId = world < 0.667 ? ids.y : ids.z;
    selectId = world < 0.333 ? ids.x : selectId;
    return int(floor(selectId+0.5)) * 4 - 4;
}

mat4 gradingMatrix(int groupId) {
    vec4 row0 = texelFetch(asset_grading_lut, groupId);
    vec4 row1 = texelFetch(asset_grading_lut, groupId + 1);
    vec4 row2 = texelFetch(asset_grading_lut, groupId + 2);
    vec4 row3 = texelFetch(asset_grading_lut, groupId + 3);
    return mat4(row0, row1, row2, row3);
}

vec3 gradeColor(vec3 color, mat4 matrix) {
    return (color.y * matrix[1].xyz) + (color.x * matrix[0].xyz) + (color.z * matrix[2].xyz) + matrix[3].xyz;
}

vec3 graded(vec3 color, float tex2_alpha) {
    float worldval = (abs(fract(world*12) - 0.5) * 2 - 1) * world_grading_color_value_variation + 1;
    float subsurface = 1.0 - min(clamp(clamp(floor(tex2_alpha * 63.75) / 63.0, 0.0, 1.0) * ss_intensity_mult, 0.0, 1.0) * 100, 1.0);
    int leafGroup = group(vec3(grading_group_id, grading_group_id_secondworld, grading_group_id_thirdworld));
    int trunkGroup = group(vec3(grading_group_id_trunk));

    mat4 leafGrading = mat4(1.0);
    if (leafGroup >= 0) {
        leafGrading = gradingMatrix(leafGroup);
    }
    mat4 trunkGrading = mat4(1.0);
    if (trunkGroup >= 0) {
        trunkGrading = gradingMatrix(trunkGroup);
    }
    vec3 leafGraded = gradeColor(color, leafGrading);
    vec3 trunkGraded = gradeColor(color, trunkGrading);

    return (-leafGraded * worldval + trunkGraded) * subsurface + (leafGraded * worldval);
}

void main() {
    vec4 albedoOpacity = texture(tex0, fragUV);
    if(albedoOpacity.w < opacity_threshold) {
        discard;
    }

    vec4 tex2_sample = texture(tex2, fragUV);
    vec3 albedo = graded(albedoOpacity.xyz, tex2_sample.a);
    vec3 normal = texture(tex1, fragUV).xyz;

    normal = normal * 2.0 - 1.0; // in tangent space
    normal.z = reconstructNormalZ(normal.xy);

    normal.x = -normal.x;
    // winding order is different than directx I guess, so frontfacing gets the back faces
    if (gl_FrontFacing) {
        normal = -normal;
    }
    vec3 ambient = vec3(1.0);

    vec3 lightDirection = normalize(fragTangentLightPosition - fragTangentFragmentPosition);
    vec3 lightColor = vec3(0.7);
    vec3 diffuse = max(dot(normal, lightDirection), 0.0) * lightColor;

    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    vec3 reflectDirection = reflect(-lightDirection, normal);
    vec3 halfwayDirection = normalize(lightDirection + viewDirection);
    vec3 specular = pow(max(dot(normal, halfwayDirection), 0.0), 32.0) * lightColor;

    fragColor = vec4(albedo * (mix(ambient, diffuse, 0.6) + 0.5 * specular), 1.0);
    //fragColor = calculateLighting(vec4(albedo, 1.0), tex2_sample.g, tex2_sample.a, tex2_sample.b, normal);
}