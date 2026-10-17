#version 430 core

out vec2 fragUV0;

layout (location = 0) in vec3 inPos;
layout (location = 1) in vec2 inUV0;

void main() 
{
    gl_Position = vec4(inPos, 1.0);
    fragUV0 = inUV0;
}