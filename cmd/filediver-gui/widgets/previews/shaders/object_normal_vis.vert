#version 330 core

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec3 inUV;
layout(location = 3) in vec3 inTangent;
layout(location = 4) in vec3 inBitangent;

out vec4 normalEndPosition;
out vec4 tangentEndPosition;
out vec4 bitangentEndPosition;

uniform mat4 mvp; // projection*view*model
uniform float len; // normal length
uniform bool udimShown[64];

bool isShown() {
    int udim = int(inUV.x) | int(1-inUV.y)<<5;
    return udim < 64 && udimShown[udim];
}

void main() {
    if (!isShown()) return;
    normalEndPosition    = mvp * vec4(inPosition + inNormal * len, 1.0);
    tangentEndPosition   = mvp * vec4(inPosition + inTangent * len, 1.0);
    bitangentEndPosition = mvp * vec4(inPosition + inBitangent * len, 1.0);
    gl_Position = mvp * vec4(inPosition, 1.0);
}