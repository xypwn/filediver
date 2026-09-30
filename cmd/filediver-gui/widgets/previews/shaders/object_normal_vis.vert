#version 430 core

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec2 inUV;
layout(location = 3) in vec4 inTangent;
layout(location = 4) in vec3 inBitangent;

out vec4 normalEndPosition;
out vec4 tangentEndPosition;
out vec4 bitangentEndPosition;
out vec2 vertexUV;

uniform mat4 mvp; // projection*view*model
uniform float len; // normal length

void main() {
    vertexUV = inUV;
    normalEndPosition    = mvp * vec4(inPosition + inNormal * len, 1.0);
    tangentEndPosition   = mvp * vec4(inPosition + inTangent.xyz * len, 1.0);
    bitangentEndPosition = mvp * vec4(inPosition + inBitangent * len, 1.0);
    gl_Position = mvp * vec4(inPosition, 1.0);
}