#version 430 core

layout(location = 0) in vec3 inPosition;
layout(location = 7) in mat4 instModel;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main() {
    gl_Position = projection * view * model * instModel * vec4(inPosition, 1.0);
}