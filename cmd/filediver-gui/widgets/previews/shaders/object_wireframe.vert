#version 430 core

layout(location = 0) in vec3 inPosition;
layout(location = 2) in vec2 inUV;

out vec2 UV;

uniform mat4 mvp; // projection*view*model

void main() {
    gl_Position = mvp * vec4(inPosition, 1.0);
    UV = inUV;
}