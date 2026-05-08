#version 430 core

layout(location = 0) in vec4 inPositionU;
layout(location = 1) in vec4 LODPositionV;
layout(location = 2) in uvec4 packedNormalTangentAOTwosided;
layout(location = 3) in vec4 windParams;

out vec4 normalEndPosition;
out vec4 tangentEndPosition;
out vec4 bitangentEndPosition;

uniform mat4 mvp; // projection*view*model
uniform float len;

uniform sampler2D fibonacci_normal_lut;

void main() {
    gl_Position = mvp * vec4(inPositionU.xyz, 1.0);

    uvec2 topBitsNT = packedNormalTangentAOTwosided.xy >> uvec2(4, 4);
    uvec2 lowBitsNT = packedNormalTangentAOTwosided.xy & uvec2(0xf, 0xf);
    vec3 normal = texelFetch(fibonacci_normal_lut, ivec2(lowBitsNT.x, topBitsNT.x), 0).xyz;
    vec3 tangent = texelFetch(fibonacci_normal_lut, ivec2(lowBitsNT.y, topBitsNT.y), 0).xyz;
    vec3 bitangent = cross(normal, tangent);

    normalEndPosition = mvp * vec4(inPositionU.xyz + normal * len, 1.0);
    tangentEndPosition = mvp * vec4(inPositionU.xyz + tangent * len, 1.0);
    bitangentEndPosition = mvp * vec4(inPositionU.xyz + bitangent * len, 1.0);
}