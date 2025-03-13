#version 320 es

precision mediump float;

layout(location = 0) out vec4 fragColor;

layout(location = 0) in vec3 fragPosition;
layout(location = 1) in vec3 fragNormal;

uniform mat4 mvp; // projection*view*model
uniform mat4 model;
uniform mat4 normal; // normal matrix = transpose(inverse(model)) // actually mat3
uniform vec3 viewPosition;

void main() {
    vec3 lightPosition = vec3(0.0, 1000.0, 1000.0);
    vec3 lightDirection = normalize(lightPosition - fragPosition);

    vec3 lightColor = vec3(0.4);

    vec3 albedo = vec3(1.0);
    vec3 ambient = vec3(1.0);
    vec3 diffuse = max(dot(fragNormal, lightDirection), 0.0) * lightColor;

    vec3 viewDirection = normalize(viewPosition - fragPosition);
    vec3 reflectDirection = reflect(-lightDirection, fragNormal);
    //vec3 specular = pow(max(dot(viewDirection, reflectDirection), 0.0), 32) * lightColor;
    vec3 specular = pow(max(dot(viewDirection, reflectDirection), 0.0), 256) * lightColor;

    fragColor = vec4(albedo * (mix(ambient, diffuse, 0.9) + 0.5 * specular), 1.0);

    // Normal debugging
    //fragColor = vec4(fragNormal * 0.5 + 0.5, 1.0);
    //fragColor = vec4(fragNormal, 1.0);
}