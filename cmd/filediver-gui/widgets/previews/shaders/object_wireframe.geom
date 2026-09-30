#version 430 core

layout (triangles) in;
layout (line_strip, max_vertices = 4) out;

in vec2 UV[];

uniform bool hasVisibilityMasks;
uniform bool udimShown[64];

bool isShown(uint udim) {
    return udim < 64 && udimShown[udim];
}

uint getUDIM(vec2 uv) {
    return uint(floor(clamp(uv.x, 0.0, 31.999)) + 32 * floor(clamp(0.99 - uv.y, 0.0, 1.0)));
}

void main() {
    uint udim = min(getUDIM(UV[0]), min(getUDIM(UV[1]), getUDIM(UV[2])));
    if (hasVisibilityMasks && !isShown(udim)) {
        return;
    }
    gl_Position = gl_in[0].gl_Position;
    EmitVertex();
    gl_Position = gl_in[1].gl_Position;
    EmitVertex();
    gl_Position = gl_in[2].gl_Position;
    EmitVertex();
    gl_Position = gl_in[0].gl_Position;
    EmitVertex();
    EndPrimitive();
}