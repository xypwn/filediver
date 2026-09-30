#version 430 core

layout (points) in;
layout (line_strip, max_vertices = 6) out;

in vec4 normalEndPosition[];
in vec4 tangentEndPosition[];
in vec4 bitangentEndPosition[];
in vec2 vertexUV[];

out vec4 lineColor;

uniform bool showTangentBitangent;
uniform bool hasVisibilityMasks;
uniform bool udimShown[64];

uint getUDIM(vec2 uv) {
    return uint(floor(clamp(uv.x, 0.0, 31.999)) + 32 * floor(clamp(0.99 - uv.y, 0.0, 1.0)));
}

bool isShown(uint udim) {
    return udim < 64 && udimShown[udim];
}

void drawLine(vec4 endPosition) {
    gl_Position = gl_in[0].gl_Position;
    EmitVertex();
    gl_Position = endPosition;
    EmitVertex();
    EndPrimitive();
}

void main() {
    if (hasVisibilityMasks && !isShown(getUDIM(vertexUV[0]))) {
        return;
    }
    lineColor = vec4(0, 0, 1, 1);
    drawLine(normalEndPosition[0]);
    if (showTangentBitangent) {
        lineColor = vec4(1, 0, 0, 1);
        drawLine(tangentEndPosition[0]);
        lineColor = vec4(0, 1, 0, 1);
        drawLine(bitangentEndPosition[0]);
    }
}