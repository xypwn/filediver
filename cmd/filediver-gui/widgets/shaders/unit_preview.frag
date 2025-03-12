#version 320 es

precision mediump float;

layout(location = 0) out vec4 fragColor;

uniform mat4 mvp; // projection*view*model
uniform mat4 model;
uniform mat4 normal; // normal matrix = transpose(inverse(model)) // actually mat3
uniform vec3 viewPosition;

void main() {
    fragColor = vec4(1, 1, 1, 1);
}