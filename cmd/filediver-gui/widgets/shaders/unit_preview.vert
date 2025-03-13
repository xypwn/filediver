#version 320 es

precision mediump float;

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;

layout(location = 0) out vec3 fragPosition;
layout(location = 1) out vec3 fragNormal;

uniform mat4 mvp; // projection*view*model
uniform mat4 model;
uniform mat4 normal; // normal matrix = transpose(inverse(model)) // actually mat3
uniform vec3 viewPosition;

void main() {
    gl_Position = mvp * vec4(inPosition, 1.0);
    fragPosition = vec3(model * vec4(inPosition, 1.0));
    fragNormal = normalize(mat3(normal) * inNormal);
}