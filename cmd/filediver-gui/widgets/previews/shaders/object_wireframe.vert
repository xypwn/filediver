#version 430 core

layout(location = 0) in vec3 inPosition;
layout(location = 2) in vec2 inUV;
layout(location = 7) in float udim;

uniform mat4 mvp; // projection*view*model
uniform bool hasVisibilityMasks;
uniform bool udimShown[64];

bool isShown() {
    return udim < 64 && udimShown[int(udim)];
}

void main() {
    if (hasVisibilityMasks && !isShown()) {
        gl_Position = vec4(vec3(0.0), 1.0);
        return;
    }
    gl_Position = mvp * vec4(inPosition, 1.0);
}