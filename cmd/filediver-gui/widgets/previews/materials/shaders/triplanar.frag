// This file is meant to be included in other shader fragments, hence the missing #version directive

layout(shared, binding = 9) uniform TriplanarBlock {
    vec3 fd_world_offset;
};

in vec3 fragInstPosition;
in mat3 dbg_fragTBN;
vec2 getTriplanarUVs() {
    vec3 normal = dbg_fragTBN[2];
    vec2 triplanarUVs;
    vec4 temp;
    if (abs(normal.z) < abs(normal.x)) {
        temp.x = fragInstPosition.y + fd_world_offset.y;
        temp.y = fragInstPosition.z + fd_world_offset.z;
    } else {
        temp.x = fragInstPosition.x + fd_world_offset.x;
        temp.y = fragInstPosition.y + fd_world_offset.y;
    }
    if (abs(normal.z) < abs(normal.y)) {
        temp.z = fragInstPosition.x + fd_world_offset.x;
        temp.w = fragInstPosition.z + fd_world_offset.z;
    } else {
        temp.z = fragInstPosition.x + fd_world_offset.x;
        temp.w = fragInstPosition.y + fd_world_offset.y;
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

vec2 getCustomTriplanarUVs(vec3 normal) {
    vec2 triplanarUVs;
    vec4 temp;
    if (abs(normal.z) < abs(normal.x)) {
        temp.x = fragInstPosition.y + fd_world_offset.y;
        temp.y = fragInstPosition.z + fd_world_offset.z;
    } else {
        temp.x = fragInstPosition.x + fd_world_offset.x;
        temp.y = fragInstPosition.y + fd_world_offset.y;
    }
    if (abs(normal.z) < abs(normal.y)) {
        temp.z = fragInstPosition.x + fd_world_offset.x;
        temp.w = fragInstPosition.z + fd_world_offset.z;
    } else {
        temp.z = fragInstPosition.x + fd_world_offset.x;
        temp.w = fragInstPosition.y + fd_world_offset.y;
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

    return result * 2.0 - 1.0;
}

vec3 getCustomTriplanarNormal(vec3 normal, vec3 tangent_space_normal) {
    vec3 temp;
    vec3 result;

    if (abs(normal.z) < abs(normal.x)) {
        temp.x = tangent_space_normal.y;
        temp.y = tangent_space_normal.x;
        temp.z = tangent_space_normal.z;
    } else {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.z;
        temp.z = tangent_space_normal.y;
    }
    result.xyz = temp.xyz;
    if (abs(normal.z) < abs(normal.y)) {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.y;
        temp.z = tangent_space_normal.z;
    } else {
        temp.x = tangent_space_normal.x;
        temp.y = tangent_space_normal.z;
        temp.z = tangent_space_normal.y;
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

    return result;
}
