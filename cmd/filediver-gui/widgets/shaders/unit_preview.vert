#version 320 es

precision mediump float;

layout(location = 0) in vec3 position;

layout(location = 0) uniform mat4 mvp;

void main() {
    gl_Position = mvp * vec4(position, 1.0);
}