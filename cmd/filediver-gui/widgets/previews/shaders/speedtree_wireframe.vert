#version 330 core

layout(location = 0) in vec4 inPositionU;

uniform mat4 mvp; // projection*view*model

void main() {
    gl_Position = mvp * vec4(inPositionU.xyz, 1.0);
}