#version 430 core

layout(location = 0) in vec4 inPositionU;
layout(location = 1) in vec4 grassOffsetAOV;
layout(location = 2) in uvec4 packedNormalTangentWindParams;

out vec3 fragPosition;
out vec2 fragUV;
out vec3 fragTangentLightPosition; // tangent meaning in tangent space
out vec3 fragTangentViewPosition;
out vec3 fragTangentFragmentPosition;
out mat3 dbg_fragTBN;
out mat3 dbg_fragITBN;

uniform mat4 mvp; // projection*view*model
uniform mat4 model;
uniform mat3 normalMat; // normal matrix = transpose(inverse(model))
uniform vec3 viewPosition;

uniform sampler2D fibonacci_normal_lut;

void main() {
    gl_Position = mvp * vec4(inPositionU.xyz, 1.0);
    fragPosition = vec3(model * vec4(inPositionU.xyz, 1.0));
    fragUV = vec2(inPositionU.w, grassOffsetAOV.w);

    uvec2 topBitsNT = packedNormalTangentWindParams.xy >> uvec2(4, 4);
    uvec2 lowBitsNT = packedNormalTangentWindParams.xy & uvec2(0xf, 0xf);
    vec3 normal = texelFetch(fibonacci_normal_lut, ivec2(lowBitsNT.x, topBitsNT.x), 0).xyz;
    vec3 tangent = texelFetch(fibonacci_normal_lut, ivec2(lowBitsNT.y, topBitsNT.y), 0).xyz;
    vec3 bitangent = cross(normal, tangent);
    {
        vec3 t = normalize(normalMat * tangent);
        vec3 n = normalize(normalMat * normal);
        vec3 b = normalize(normalMat * bitangent);
        //t = normalize(t - dot(t, n) * n);
        //vec3 b = cross(n, t);

        mat3 tbn = mat3(t, b, n);
        mat3 itbn = inverse(tbn);

        dbg_fragTBN = tbn;
        dbg_fragITBN = itbn;

        vec3 lightPosition = normalize(viewPosition) * 10000.0; // light behind camera

        fragTangentLightPosition    = itbn * lightPosition;
        fragTangentViewPosition     = itbn * viewPosition;
        fragTangentFragmentPosition = itbn * fragPosition;
    }
}