#version 330 core

layout(location = 0) in vec3 inPosition;
layout(location = 2) in vec2 inUV;

out vec3 fragPosition;
out vec2 fragUV;

uniform mat4 projection;
uniform mat4 model;
uniform mat4 view;

void main() {
    gl_Position = projection * view * model * vec4(inPosition, 1.0);
    fragPosition = vec3(model * vec4(inPosition, 1.0));
    fragUV = inUV;
}