#version 330 core

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV;

void main() {
    fragColor = vec4(fragUV, 1.0, 1.0);
}