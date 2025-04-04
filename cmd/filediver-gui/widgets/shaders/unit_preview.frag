#version 320 es

precision mediump float;
precision mediump int;

out vec4 fragColor;

in vec3 fragPosition;
in vec2 fragUV;
in vec3 fragTangentLightPosition; // tangent meaning in tangent space
in vec3 fragTangentViewPosition;
in vec3 fragTangentFragmentPosition;
in mat3 dbg_fragTBN;
in mat3 dbg_fragITBN;

#define FLAG_NORMAL_TEXTURE 0x1u

uniform mat4 mvp; // projection*view*model
uniform mat4 model;
uniform mat4 normalMat; // normal matrix = transpose(inverse(model)) // actually mat3
uniform vec3 viewPosition;

uniform sampler2D texAlbedo;
uniform sampler2D texNormal;

// Returns the Z value if truncated from a normalized XY.
float reconstructNormalZ(vec2 xy) {
    return sqrt(1.0 - xy.x*xy.x - xy.y*xy.y);
}

void main() {
    vec3 normal = texture(texNormal, fragUV).xyz;
    normal = normalize(normal * 2.0 - 1.0); // in tangent space
    normal.z = reconstructNormalZ(normal.xy);

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

    //fragColor = vec4(normalize(fragTangentLightPosition) * 0.5 + 0.5, 1.0);
    //fragColor = vec4(normalize(dbg_fragTBN * normal) * 0.5 + 0.5, 1.0);

    // Normal debugging
    //fragColor = vec4(normal * 0.5 + 0.5, 1.0);
    //fragColor = vec4(n, 1.0);
}