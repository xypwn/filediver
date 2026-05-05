#version 430 core

out vec4 fragColor;

in vec3 fragPosition;
in vec4 dbg_PackedNormalTangent;

void main() {
    fragColor = vec4(dbg_PackedNormalTangent.xyzw);
    return;
}