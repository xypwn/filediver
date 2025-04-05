#version 320 es

precision mediump float;
precision mediump int;

layout (triangles) in;
layout (line_strip, max_vertices = 6) out;

in vec4 normalEndPosition[];

void drawLine(int index) {
    gl_Position = gl_in[index].gl_Position;
    EmitVertex();
    gl_Position = normalEndPosition[index];
    EmitVertex();
    EndPrimitive();
}

void main() {
    drawLine(0);
    drawLine(1);
    drawLine(2);
}