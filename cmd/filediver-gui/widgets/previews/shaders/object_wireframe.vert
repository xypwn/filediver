#version 430 core

layout(location = 0) in vec3 inPosition;
layout(location = 2) in vec2 inUV;
layout(location = 7) in mat4 model;

out vec2 UV;

uniform mat4 projection; // projection*view*model
uniform mat4 view; // projection*view*model

void main() {
    gl_Position = projection * view * model * vec4(inPosition, 1.0);
    UV = inUV;
}