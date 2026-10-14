#version 430 core

layout (triangles) in;
layout (triangle_strip, max_vertices = 3) out;

in VertexOutput
{
    vec3 fragPosition;
    vec2 fragUV0;
    vec2 fragUV1;
    vec2 fragUV2;
    vec3 fragTangentLightPosition;
    vec3 fragTangentViewPosition;
    vec3 fragTangentFragmentPosition;
    mat3 dbg_fragTBN;
    mat3 dbg_fragITBN;
} inVertices[];

out vec3 fragPosition;
out vec2 fragUV0;
out vec2 fragUV1;
out vec2 fragUV2;
out vec3 fragTangentLightPosition;
out vec3 fragTangentViewPosition;
out vec3 fragTangentFragmentPosition;
out mat3 dbg_fragTBN;
out mat3 dbg_fragITBN;

uniform bool hasVisibilityMasks;
uniform bool udimShown[64];

bool isShown(uint udim) {
    return udim < 64 && udimShown[udim];
}

uint getUDIM(vec2 uv) {
    return uint(floor(clamp(uv.x, 0.0, 31.999)) + 32 * floor(clamp(0.99 - uv.y, 0.0, 1.0)));
}

void drawVertex(int idx) {
    gl_Position = gl_in[idx].gl_Position;
    fragPosition = inVertices[idx].fragPosition;
    fragUV0 = inVertices[idx].fragUV0;
    fragUV1 = inVertices[idx].fragUV1;
    fragUV2 = inVertices[idx].fragUV2;
    fragTangentLightPosition = inVertices[idx].fragTangentLightPosition;
    fragTangentViewPosition = inVertices[idx].fragTangentViewPosition;
    fragTangentFragmentPosition = inVertices[idx].fragTangentFragmentPosition;
    dbg_fragTBN = inVertices[idx].dbg_fragTBN;
    dbg_fragITBN = inVertices[idx].dbg_fragITBN;
    EmitVertex();
}

void main() {
    uint udim = min(getUDIM(inVertices[0].fragUV0), min(getUDIM(inVertices[1].fragUV0), getUDIM(inVertices[2].fragUV0)));
    if (hasVisibilityMasks && !isShown(udim)) {
        return;
    }
    drawVertex(0);
    drawVertex(1);
    drawVertex(2);
    EndPrimitive();
}