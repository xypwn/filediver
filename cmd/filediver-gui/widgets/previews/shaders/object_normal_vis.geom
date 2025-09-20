#version 330 core

layout (points) in;
layout (line_strip, max_vertices = 6) out;

in vec4 normalEndPosition[];
in vec4 tangentEndPosition[];
in vec4 bitangentEndPosition[];

out vec4 lineColor;

uniform bool showTangentBitangent;

void drawLine(vec4 endPositions[1]) {
    gl_Position = gl_in[0].gl_Position;
    EmitVertex();
    gl_Position = endPositions[0];
    EmitVertex();
    EndPrimitive();
}

void main() {
    lineColor = vec4(0, 0, 1, 1);
    drawLine(normalEndPosition);
    if (showTangentBitangent) {
        lineColor = vec4(1, 0, 0, 1);
        drawLine(tangentEndPosition);
        lineColor = vec4(0, 1, 0, 1);
        drawLine(bitangentEndPosition);
    }
}