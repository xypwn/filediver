#version 430 core

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV;
in vec3 fragTangentLightPosition; // tangent meaning in tangent space
in vec3 fragTangentViewPosition;
in vec3 fragTangentFragmentPosition;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

uniform float grading_group_id;
uniform float grading_group_id_secondworld;
uniform float grading_group_id_thirdworld;
uniform float grading_group_id_trunk;
uniform float world_grading_color_value_variation;
uniform float ss_intensity_mult;
uniform float world;

uniform sampler2D texAlbedo;
uniform sampler2D texNormal;
uniform sampler2D tex2;
uniform samplerBuffer texAssetGrading;
uniform float opacityThreshold;

// Reconstructs the Z value if Z was truncated from XYZ.
float reconstructNormalZ(vec2 xy) {
    return sqrt(1.0 - xy.x*xy.x - xy.y*xy.y);
}

int group(vec3 ids) {
    float selectId = world < 0.667 ? ids.y : ids.z;
    selectId = world < 0.333 ? ids.x : selectId;
    return int(floor(selectId+0.5)) * 4 - 4;
}

mat4 gradingMatrix(int groupId) {
    vec4 row0 = texelFetch(texAssetGrading, groupId);
    vec4 row1 = texelFetch(texAssetGrading, groupId + 1);
    vec4 row2 = texelFetch(texAssetGrading, groupId + 2);
    vec4 row3 = texelFetch(texAssetGrading, groupId + 3);
    return mat4(row0, row1, row2, row3);
}

vec3 gradeColor(vec3 color, mat4 matrix) {
    return (color.y * matrix[1].xyz) + (color.x * matrix[0].xyz) + (color.z * matrix[2].xyz) + matrix[3].xyz;
}

vec3 graded(vec3 color) {
    float tex2_alpha = texture(tex2, fragUV).w;
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
    vec4 albedoOpacity = texture(texAlbedo, fragUV);
    if(albedoOpacity.w < opacityThreshold) {
        discard;
    }

    vec3 albedo = graded(albedoOpacity.xyz);
    vec3 normal = texture(texNormal, fragUV).xyz;

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
}