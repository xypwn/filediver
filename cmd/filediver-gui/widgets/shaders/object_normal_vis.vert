#version 330 core

precision mediump float;
precision mediump int;

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;

out vec4 normalEndPosition;

uniform mat4 mvp; // projection*view*model
uniform float len; // normal length

void main() {
    normalEndPosition = mvp * vec4(inPosition + inNormal * len, 1.0);
    gl_Position = mvp * vec4(inPosition, 1.0);
}