#version 430 core

#include "lighting.frag"

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV0;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

layout(shared, binding = 0) uniform EmissiveBlock {
    vec3 base_color;
    float emissive_intensity;
};

void main() {
    fragColor = vec4(emissive_intensity * base_color, 1.0);
}