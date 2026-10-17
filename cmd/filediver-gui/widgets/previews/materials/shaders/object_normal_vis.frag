#version 430 core

out vec4 fragColor;

in vec4 lineColor;

void main() {
    fragColor = lineColor;
}