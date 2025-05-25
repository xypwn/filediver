#version 330 core

layout(location = 0) in vec3 inPosition;

uniform mat4 mvp; // projection*view*model

void main() {
    gl_Position = mvp * vec4(inPosition, 1.0);
}