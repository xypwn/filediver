#version 330 core

layout(location = 0) in vec4 inPositionU;
layout(location = 7) in mat4 model;

out vec2 UV;

uniform mat4 view; // projection*view*model
uniform mat4 projection; // projection*view*model

void main() {
    gl_Position = projection * view * model * vec4(inPositionU.xyz, 1.0);
    UV = vec2(0.0);
}