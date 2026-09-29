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

void main() {
    uint udim = min(getUDIM(inVertices[0].fragUV0), min(getUDIM(inVertices[1].fragUV0), getUDIM(inVertices[2].fragUV0)));
    if (hasVisibilityMasks && !isShown(udim)) {
        return;
    }
    gl_Position = gl_in[0].gl_Position;
    fragPosition = inVertices[0].fragPosition;
    fragUV0 = inVertices[0].fragUV0;
    fragUV1 = inVertices[0].fragUV1;
    fragUV2 = inVertices[0].fragUV2;
    fragTangentLightPosition = inVertices[0].fragTangentLightPosition;
    fragTangentViewPosition = inVertices[0].fragTangentViewPosition;
    fragTangentFragmentPosition = inVertices[0].fragTangentFragmentPosition;
    dbg_fragTBN = inVertices[0].dbg_fragTBN;
    dbg_fragITBN = inVertices[0].dbg_fragITBN;
    EmitVertex();
    gl_Position = gl_in[1].gl_Position;
    fragPosition = inVertices[1].fragPosition;
    fragUV0 = inVertices[1].fragUV0;
    fragUV1 = inVertices[1].fragUV1;
    fragUV2 = inVertices[1].fragUV2;
    fragTangentLightPosition = inVertices[1].fragTangentLightPosition;
    fragTangentViewPosition = inVertices[1].fragTangentViewPosition;
    fragTangentFragmentPosition = inVertices[1].fragTangentFragmentPosition;
    dbg_fragTBN = inVertices[1].dbg_fragTBN;
    dbg_fragITBN = inVertices[1].dbg_fragITBN;
    EmitVertex();
    gl_Position = gl_in[2].gl_Position;
    fragPosition = inVertices[2].fragPosition;
    fragUV0 = inVertices[2].fragUV0;
    fragUV1 = inVertices[2].fragUV1;
    fragUV2 = inVertices[2].fragUV2;
    fragTangentLightPosition = inVertices[2].fragTangentLightPosition;
    fragTangentViewPosition = inVertices[2].fragTangentViewPosition;
    fragTangentFragmentPosition = inVertices[2].fragTangentFragmentPosition;
    dbg_fragTBN = inVertices[2].dbg_fragTBN;
    dbg_fragITBN = inVertices[2].dbg_fragITBN;
    EmitVertex();
    EndPrimitive();
}