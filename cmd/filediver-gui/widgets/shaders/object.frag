#version 330 core

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV;
in vec3 fragTangentLightPosition; // tangent meaning in tangent space
in vec3 fragTangentViewPosition;
in vec3 fragTangentFragmentPosition;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

uniform sampler2D texAlbedo;
uniform sampler2D texNormal;
uniform bool shouldReconstructNormalZ;

// Reconstructs the Z value if Z was truncated from XYZ.
float reconstructNormalZ(vec2 xy) {
    return sqrt(1.0 - xy.x*xy.x - xy.y*xy.y);
}

void main() {
    vec3 normal = texture(texNormal, fragUV).xyz;

    //fragColor = vec4(normal, 1.0); return;

    normal = normalize(normal * 2.0 - 1.0); // in tangent space
    if (shouldReconstructNormalZ) {
        normal.z = reconstructNormalZ(normal.xy);
    }
    normal.x = -normal.x;

    vec3 albedo = texture(texAlbedo, fragUV).xyz;
    vec3 ambient = vec3(1.0);

    vec3 lightDirection = normalize(fragTangentLightPosition - fragTangentFragmentPosition);
    vec3 lightColor = vec3(0.7);
    vec3 diffuse = max(dot(normal, lightDirection), 0.0) * lightColor;

    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    vec3 reflectDirection = reflect(-lightDirection, normal);
    vec3 halfwayDirection = normalize(lightDirection + viewDirection);
    vec3 specular = pow(max(dot(normal, halfwayDirection), 0.0), 32.0) * lightColor;

    fragColor = vec4(albedo * (mix(ambient, diffuse, 0.6) + 0.5 * specular), 1.0);

    // Normal debugging (ignoring normal map)
    //fragColor = vec4(normalize(dbg_fragTBN * vec3(0.0, 0.0, 1.0)), 1.0);
    // Normal debugging (world space)
    //fragColor = vec4(normalize(dbg_fragTBN * normal), 1.0);
    //fragColor = vec4(normalize(dbg_fragTBN * normal) * 0.5 + 0.5, 1.0);
    // Normal debugging (tangent space)
    //fragColor = vec4(normal, 1.0);
    //fragColor = vec4(normal * 0.5 + 0.5, 1.0);
}