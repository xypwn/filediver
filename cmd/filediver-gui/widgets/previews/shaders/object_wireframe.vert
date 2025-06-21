#version 330 core

layout(location = 0) in vec3 inPosition;
layout(location = 2) in vec2 inUV;

uniform mat4 mvp; // projection*view*model
uniform bool udimShown[64];

bool isShown() {
    int udim = int(inUV.x) | int(1-inUV.y)<<5;
    return udim < 64 && udimShown[udim];
}

void main() {
    if (!isShown()) return;
    gl_Position = mvp * vec4(inPosition, 1.0);
}