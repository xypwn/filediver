#version 320 es

precision mediump float;

layout(location = 0) in vec3 position;

uniform mat4 mvp; // projection*view*model
uniform mat4 model;
uniform mat4 normal; // normal matrix = transpose(inverse(model)) // actually mat3
uniform vec3 viewPosition;

void main() {
    gl_Position = mvp * vec4(position, 1.0);
}