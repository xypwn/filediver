// This file is meant to be included in other shader fragments, hence the missing #version directive

in mat3 dbg_fragTBN;
vec2 getTriplanarUVs() {
    vec3 normal = dbg_fragTBN[2];
    vec2 triplanarUVs;
    vec4 temp;
    if (abs(normal.z) < abs(normal.x)) {
        temp.x = fragInstPosition.y;
        temp.y = fragInstPosition.z;
    } else {
        temp.x = fragInstPosition.x;
        temp.y = fragInstPosition.y;
    }
    if (abs(normal.z) < abs(normal.y)) {
        temp.z = fragInstPosition.x;
        temp.w = fragInstPosition.z;
    } else {
        temp.z = fragInstPosition.x;
        temp.w = fragInstPosition.y;
    }
    if (abs(normal.y) < abs(normal.x)) {
        triplanarUVs.x = temp.x;
        triplanarUVs.y = temp.y;
    } else {
        triplanarUVs.x = temp.z;
        triplanarUVs.y = temp.w;
    }
    return triplanarUVs;
}

vec3 getTriplanarNormal(vec3 tangent_space_normal) {
    vec3 normal = dbg_fragTBN[2];
    vec4 temp;
    temp.w = 1.0 - tangent_space_normal.x;
    vec3 result;

    if (abs(normal.z) < abs(normal.x)) {
        temp.x = tangent_space_normal.z;
        temp.y = temp.w;
        temp.z = tangent_space_normal.y;
    } else {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.y;
        temp.z = tangent_space_normal.z;
    }
    result.xyz = temp.xyz;
    if (abs(normal.z) < abs(normal.y)) {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.z;
        temp.z = tangent_space_normal.y;
    } else {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.y;
        temp.z = tangent_space_normal.z;
    }
    if (abs(normal.y) < abs(normal.x)) {
        result.x = result.x;
        result.y = result.y;
        result.z = result.z;
    } else {
        result.x = temp.x;
        result.y = temp.y;
        result.z = temp.z;
    }

    // temp.xyz = result * 2.0 - 1.0 + normal;
    // result.x = dot(temp.xyz, dbg_fragTBN[0]);
    // result.y = dot(temp.xyz, dbg_fragTBN[1]);
    // result.z = dot(temp.xyz, dbg_fragTBN[2]);
    return result * 2.0 - 1.0;
}
