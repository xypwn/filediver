#version 330 core

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec3 inUV;

out vec4 normalEndPosition;

uniform mat4 mvp; // projection*view*model
uniform float len; // normal length
uniform bool udimShown[64];

bool isShown() {
    int udim = int(inUV.x) | int(1-inUV.y)<<5;
    return udim < 64 && udimShown[udim];
}

void main() {
    if (!isShown()) return;
    normalEndPosition = mvp * vec4(inPosition + inNormal * len, 1.0);
    gl_Position = mvp * vec4(inPosition, 1.0);
}