#version 330 core

precision mediump float;
precision mediump int;

out vec4 fragColor;

uniform vec4 color;

void main() {
    fragColor = color;
}