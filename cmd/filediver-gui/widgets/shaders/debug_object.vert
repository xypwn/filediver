#version 320 es

precision mediump float;

layout(location = 0) in vec3 inPosition;

uniform mat4 mvp; // projection*view*model
uniform vec4 color;

void main() {
    gl_Position = mvp * vec4(inPosition, 1.0);
}