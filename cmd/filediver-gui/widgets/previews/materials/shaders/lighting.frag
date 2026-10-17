// note - this file is meant to be included in other shader source, not to be used standalone
// hence the missing #version directive

in vec3 fragTangentLightPosition; // tangent meaning in tangent space
in vec3 fragTangentViewPosition;
in vec3 fragTangentFragmentPosition;

uniform sampler2D ibl_brdf_lut;

float reconstructNormalZ(vec2 xy) {
    return sqrt(1.0 - xy.x*xy.x - xy.y*xy.y);
}

float recoverInverseZ(vec2 normal_xy) {
    return inversesqrt(1 - (normal_xy.x * normal_xy.x) - (normal_xy.y * normal_xy.y));
}

vec3 sRGBFromIntensity(vec3 color) {
    vec3 temp = max(color, vec3(0.000061));
    color = temp * 12.92;
    temp = exp2(log2(max(temp, vec3(0.003131))) * 0.416667) * 1.055 - 0.055;
    return min(color, temp);
}

vec3 sRGBToIntensity(vec3 color) {
    color = max(color, vec3(0.000061));
    bvec3 compared = lessThan(vec3(0.04045), color);
    vec3 temp = exp2(log2((color + 0.055) / 1.055) * 2.4);
    color = color / 12.92;
    color.x = compared.x ? temp.x : color.x;
    color.y = compared.y ? temp.y : color.y;
    color.z = compared.z ? temp.z : color.z;
    return color;
}

vec3 rgb2hsv(vec3 c)
{
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));

    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 hsv2rgb(vec3 c)
{
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

// base and detail should be unpacked already, and this will return an unpacked normal
vec3 reorientNormalMaps(vec3 base, vec3 detail) {
    vec3 t = base + vec3(0, 0, 1);
    vec3 u = detail * vec3(-1, -1, 1);
    return normalize(t*dot(t, u)/t.z - u);
}

vec3 gamma(vec3 color, float value) {
    return exp2(log2(color) * vec3(value));
}

// Normal is expected to be in tangent space
vec4 calculateLighting(vec4 base_color, float roughness, float metallic, float ambient_occlusion, vec3 normal, float saturation, float contrast) {
    vec3 ambient = vec3(0.5) * ambient_occlusion;
    vec3 lightDirection = normalize(fragTangentLightPosition - fragTangentFragmentPosition);
    vec3 lightColor = vec3(1.25);
    vec3 diffuse = max(dot(normal, lightDirection), 0) * lightColor;

    vec3 specularColor = mix(vec3(0.04), base_color.rgb, metallic);

    vec3 viewDirection = normalize(fragTangentViewPosition - fragTangentFragmentPosition);
    vec3 reflectDirection = reflect(-lightDirection, normal);
    vec3 halfwayDirection = normalize(lightDirection + viewDirection);

    // blinn specular reflectance
    float NdH = max(dot(normal, halfwayDirection), 0.001);
    float HdV = max(dot(halfwayDirection, viewDirection), 0.001);
    float NdL = max(dot(normal, lightDirection), 0.0);
    float NdV = max(dot(normal, viewDirection), 0.001);

    vec3 specularFresnel = mix(specularColor, vec3(0.7), pow(1.01 - HdV, 5.0));
    float k = 1.999 / (roughness * roughness);

    vec3 blinnSpecularRef = min(1.0, 3.0 * 0.0398 * k) * pow(NdH, min(10000.0, k)) * specularFresnel * vec3(NdL);
    vec3 diffRef = (vec3(1.0) - specularFresnel) * (1.0 / 3.1415926) * NdL;

    vec3 reflectedLight = blinnSpecularRef * lightColor;
    vec3 diffuseLight = diffRef * lightColor;

    // ibl lighting
    vec2 brdf = texture(ibl_brdf_lut, vec2(max(NdV, 0.0), roughness)).xy;
    vec3 iblspec = min(vec3(0.99), mix(specularColor, vec3(1.0), pow(1.01 - NdV, 5.0))) * brdf.x + brdf.y;
    reflectedLight += iblspec * ambient;
    diffuseLight += ambient * (1.0 / 3.1415926);

    vec3 specular = pow(NdH, 32.0) * lightColor;
    // adjust saturation
    base_color.rgb = rgb2hsv(base_color.rgb);
    base_color.g *= saturation;
    base_color.rgb = hsv2rgb(base_color.rgb); 
    // adjust contrast
    base_color.rgb = (base_color.rgb - 0.5) * contrast + 0.5;

    return vec4(diffuseLight * mix(base_color.rgb, vec3(0.0), metallic) + reflectedLight, 1.0);
}