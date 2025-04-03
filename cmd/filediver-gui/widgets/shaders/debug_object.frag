#version 320 es

precision mediump float;

layout(location = 0) out vec4 outColor;

uniform mat4 mvp; // projection*view*model
uniform vec4 color;

void main() {
    outColor = color;
}