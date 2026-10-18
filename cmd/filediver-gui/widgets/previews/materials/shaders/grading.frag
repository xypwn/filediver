// This file is meant to be included with other fragment shader code, hence the missing version directive
uniform samplerBuffer asset_grading_lut;

int gradingPixel(float gradingGroupId) {
    return int(floor(gradingGroupId+0.5)) * 4 - 4;
}

mat4 gradingMatrix(int groupPixel) {
    vec4 row0 = texelFetch(asset_grading_lut, groupPixel);
    vec4 row1 = texelFetch(asset_grading_lut, groupPixel + 1);
    vec4 row2 = texelFetch(asset_grading_lut, groupPixel + 2);
    vec4 row3 = texelFetch(asset_grading_lut, groupPixel + 3);
    return mat4(row0, row1, row2, row3);
}

vec3 gradeColor(vec3 color, mat4 matrix) {
    return (color.y * matrix[1].xyz) + (color.x * matrix[0].xyz) + (color.z * matrix[2].xyz) + matrix[3].xyz;
}
