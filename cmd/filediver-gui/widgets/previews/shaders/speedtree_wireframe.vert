#version 330 core

layout(location = 0) in vec4 inPositionU;

out vec2 UV;

uniform mat4 mvp; // projection*view*model

void main() {
    gl_Position = mvp * vec4(inPositionU.xyz, 1.0);
    UV = vec2(0.0);
}