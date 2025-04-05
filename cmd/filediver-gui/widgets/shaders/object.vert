#version 320 es

precision mediump float;
precision mediump int;

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec2 inUV;
layout(location = 3) in vec3 inTangent;
layout(location = 4) in vec3 inBitangent;

out vec3 fragPosition;
out vec2 fragUV;
out vec3 fragTangentLightPosition; // tangent meaning in tangent space
out vec3 fragTangentViewPosition;
out vec3 fragTangentFragmentPosition;
out mat3 dbg_fragTBN;
out mat3 dbg_fragITBN;

uniform mat4 mvp; // projection*view*model
uniform mat4 model;
uniform mat3 normalMat; // normal matrix = transpose(inverse(model))
uniform vec3 viewPosition;

void main() {
    gl_Position = mvp * vec4(inPosition, 1.0);
    fragPosition = vec3(model * vec4(inPosition, 1.0));
    fragUV = inUV;

    {
        vec3 t = normalize(normalMat * inTangent);
        vec3 n = normalize(normalMat * inNormal);
        //vec3 b = normalize(normalMat * inBitangent);
        t = normalize(t - dot(t, n) * n);
        vec3 b = cross(n, t);

        mat3 tbn = mat3(t, b, n);
        mat3 itbn = transpose(tbn); // == inverse, because orthogonal matrix

        dbg_fragTBN = tbn;
        dbg_fragITBN = itbn;

        vec3 lightPosition = normalize(viewPosition) * 10000.0; // light behind camera

        fragTangentLightPosition    = itbn * lightPosition;
        fragTangentViewPosition     = itbn * viewPosition;
        fragTangentFragmentPosition = itbn * fragPosition;
    }
}