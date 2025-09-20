#version 330 core

out vec4 fragColor;

in vec4 lineColor;

void main() {
    fragColor = lineColor;
}