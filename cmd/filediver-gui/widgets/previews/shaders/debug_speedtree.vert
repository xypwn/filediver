#version 430 core

layout(location = 0) in vec4 inPositionU;
layout(location = 1) in vec4 LODPositionV;
layout(location = 2) in uvec4 packedNormalTangentAOTwosided;
layout(location = 3) in vec4 windParams;

out vec3 fragPosition;
out vec4 dbg_PackedNormalTangent;

uniform mat4 mvp; // projection*view*model

void main() {
    gl_Position = mvp * vec4(inPositionU.xyz, 1.0);
    dbg_PackedNormalTangent.xy = vec2(packedNormalTangentAOTwosided.xy >> 4) / 16.0;
    dbg_PackedNormalTangent.zw = vec2(packedNormalTangentAOTwosided.xy & 0xf) / 16.0;
}