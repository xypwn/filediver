#version 330 core

layout(location = 0) in vec4 inPositionU;
layout(location = 7) in mat4 instModel;

out vec2 UV;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main() {
    gl_Position = projection * view * model * instModel * vec4(inPositionU.xyz, 1.0);
    UV = vec2(0.0);
}