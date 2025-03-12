#version 320 es

precision mediump float;

layout(location = 0) out vec4 fragColor;

layout(location = 0) uniform mat4 mvp;

void main() {
    fragColor = vec4(1, 1, 1, 1);
}