#version 420 core

/* Constant Buffer 0: global_viewport
 * Type: CBUFFER
 * Size: 784
 * Flags: 
 */
layout(binding = 0) uniform global_viewport {
    vec3 camera_unprojection; // size: 12, offset: 0
    vec3 camera_center_pos; // size: 12, offset: 16
    vec3 cb_camera_pos; // size: 12, offset: 32
    mat4 camera_view; // size: 64, offset: 48
    mat4 camera_projection; // size: 64, offset: 112
    mat4 camera_inv_view; // size: 64, offset: 176
    mat4 camera_inv_projection; // size: 64, offset: 240
    mat4 camera_view_projection; // size: 64, offset: 304
    mat4 camera_last_view; // size: 64, offset: 368
    mat4 camera_last_projection; // size: 64, offset: 432
    mat4 camera_last_inv_view; // size: 64, offset: 496
    mat4 camera_last_inv_projection; // size: 64, offset: 560
    mat4 camera_last_view_projection; // size: 64, offset: 624
    vec3 camera_near_far; // size: 12, offset: 688
    float time; // size: 4, offset: 700
    float delta_time; // size: 4, offset: 704
    float frame_number; // size: 4, offset: 708
    vec2 vp_render_resolution; // size: 8, offset: 712
    vec2 raw_non_checkerboarded_target_size; // size: 8, offset: 720
    float taa_enabled; // size: 4, offset: 728
    float vrs_enabled; // size: 4, offset: 732
    float imp_transparent_override; // size: 4, offset: 736
    float debug_rendering; // size: 4, offset: 740
    float post_effects_enabled; // size: 4, offset: 744
    vec4 raw_non_checkerboarded_viewport; // size: 16, offset: 752
    float debug_lod; // size: 4, offset: 768
    float debug_shadow_lod; // size: 4, offset: 772
    float texture_density_visualization; // size: 4, offset: 776
};

/* Constant Buffer 1: c_per_object
 * Type: CBUFFER
 * Size: 32
 * Flags: 
 */
layout(binding = 1) uniform c_per_object {
    float exposure; // size: 4, offset: 0
    float ioffset; // size: 4, offset: 4
    float vertex_deformation_flags; // size: 4, offset: 8
    float material_wetness; // size: 4, offset: 12
    float detail_tile_factor_mult; // size: 4, offset: 16
    float debug_mode; // size: 4, offset: 20
    float weathering_variant; // size: 4, offset: 24
};

/* Constant Buffer 2: c_per_instance
 * Type: CBUFFER
 * Size: 160
 * Flags: 
 */
layout(binding = 2) uniform c_per_instance {
    mat4 cb_world; // size: 64, offset: 0
    vec2 lod_fade_level; // size: 8, offset: 64
    uint instance_seed; // size: 4, offset: 72
    mat4 cb_last_world; // size: 64, offset: 80
    uvec2 visibility_mask; // size: 8, offset: 144
    uint _instancing_zero; // size: 4, offset: 152
};

uniform samplerBuffer idata;
/* Resource Binding 0: global_viewport
 *   Input Type: CBUFFER
 *   Return Type: NO_RETURN
 *   View Dimension: UNKNOWN
 *   Sample Count: 0
 *   Bind Count: 1
 *   Flags: NONE
 */

/* Resource Binding 1: c_per_object
 *   Input Type: CBUFFER
 *   Return Type: NO_RETURN
 *   View Dimension: UNKNOWN
 *   Sample Count: 0
 *   Bind Count: 1
 *   Flags: NONE
 */

/* Resource Binding 2: c_per_instance
 *   Input Type: CBUFFER
 *   Return Type: NO_RETURN
 *   View Dimension: UNKNOWN
 *   Sample Count: 0
 *   Bind Count: 1
 *   Flags: NONE
 */

// Input Signature
layout(location = 0) in vec4 iPOSITION0; // SV_UNDEFINED
layout(location = 1) in vec4 iNORMAL0; // SV_UNDEFINED
layout(location = 2) in vec2 iTEXCOORD0; // SV_UNDEFINED
layout(location = 3) in vec2 iTEXCOORD1; // SV_UNDEFINED
layout(location = 4) in vec2 iTEXCOORD2; // SV_UNDEFINED
layout(location = 5) in uint iTEXCOORD15; // SV_UNDEFINED
layout(location = 6) in uint iSV_VertexID0; // SV_VERTEX_ID
layout(location = 7) in uint iSV_InstanceID0; // SV_INSTANCE_ID

// Output Signature
layout(location = 0) out vec4 oSV_POSITION; // SV_POSITION
layout(location = 1) out vec4 oTEXCOORD0; // SV_UNDEFINED
layout(location = 2) out vec3 oTEXCOORD1; // SV_UNDEFINED
layout(location = 3) out vec4 oTEXCOORD2; // SV_UNDEFINED
layout(location = 4) out vec2 oTEXCOORD3; // SV_UNDEFINED
layout(location = 5) out vec3 oTEXCOORD4; // SV_UNDEFINED

// Program type: VERTEX_SHADER
void main() {
/* Global Flags:
 * refactoringAllowed
 * enableMinPrecision
 */

// Declare Constant Buffer Immediate indexed, register cb0, size 47
// Declare Constant Buffer Immediate indexed, register cb1, size 1
// Declare Constant Buffer Immediate indexed, register cb2, size 10
// Declare BUFFER Resource t0 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare Input: v0.xyzw
// Declare Input: v1.xy
// Declare Input: v2.xy
// Declare Input: v3.xy
// Declare Input: v4.xy
// Declare Input: v5.x
// Declare Output SIV: o0.xyzw, POSITION
// Declare Output: o1.xyzw
// Declare Output: o2.xyz
// Declare Output: o3.xyzw
// Declare Output: o4.xy
// Declare Output: o5.xyz
// Declare Temps: r0 - r9
vec4 r0;
vec4 r1;
vec4 r2;
vec4 r3;
vec4 r4;
vec4 r5;
vec4 r6;
vec4 r7;
vec4 r8;
vec4 r9;
// Unary Op FTOU
r0.x = uintBitsToFloat(uint(frame_number));
// Unary Op UTOF
r0.x = float(floatBitsToUint(r0.x));
// Trinary Op MAD
r0.xy = r0.xx * vec2(0.754878, 0.569840) + vec2(0.500000, 0.500000);
// Unary Op FRC
r0.xy = fract(r0.xy);
// Trinary Op MAD
r0.xy = r0.xy * vec2(2.000000, 2.000000) + vec2(-1.000000, -1.000000);
// Binary Op MUL mask 0x3:
r0.xy = r0.xy * taa_enabled;
// Binary Op NE mask 0x4:
r0.z = intBitsToFloat(int(0.000000 != debug_rendering));
// Trinary Op MOVC
r0.xy = bool(r0.zz) ? vec2(0.000000, 0.000000) : r0.xy;
// Binary Op DIV mask 0x3:
r0.xy = r0.xy / raw_non_checkerboarded_target_size.xy;
// Unary Op FTOU
r0.z = uintBitsToFloat(uint(ioffset));
// Trinary Op IMAD
r0.z = intBitsToFloat(floatBitsToInt(iTEXCOORD15.x) * 11 + floatBitsToInt(r0.z));
// Binary Op IADD mask 0x4:
r0.z = intBitsToFloat(floatBitsToInt(r0.z) + floatBitsToInt(_instancing_zero));
// Binary Op IADD mask 0xf:
r1.xyzw = intBitsToFloat(floatBitsToInt(r0.zzzz) + ivec4(2, 3, 4, 5));
// Binary Op LD mask 0xf:
r2.xyzw = texelFetch(idata, floatBitsToInt(r1.w)).xyzw;
// Binary Op LD mask 0xf:
r3.xyzw = texelFetch(idata, floatBitsToInt(r1.y)).xzyw;
// Binary Op MUL mask 0xf:
r4.xyzw = r3.xzyw * iPOSITION0.yyyy;
// Binary Op LD mask 0xf:
r5.xyzw = texelFetch(idata, floatBitsToInt(r1.x)).yxzw;
// Binary Op LD mask 0xf:
r1.xyzw = texelFetch(idata, floatBitsToInt(r1.z)).xyzw;
// Trinary Op MAD
r4.xyzw = iPOSITION0.xxxx * r5.yxzw + r4.xyzw;
// Trinary Op MAD
r4.xyzw = iPOSITION0.zzzz * r1.xyzw + r4.xyzw;
// Trinary Op MAD
r2.xyzw = iPOSITION0.wwww * r2.xyzw + r4.xyzw;
// Binary Op DP4 mask 0x1:
r4.x = dot(r2.xyzw, camera_view_projection[0].xyzw);
// Binary Op DP4 mask 0x2:
r4.y = dot(r2.xyzw, camera_view_projection[1].xyzw);
// Binary Op DP4 mask 0x8:
r0.w = dot(r2.xyzw, camera_view_projection[3].xyzw);
// Trinary Op MAD
oSV_POSITION.xy = r0.xy * r0.ww + r4.xy;
// Binary Op IADD mask 0x8:
r1.w = intBitsToFloat(floatBitsToInt(r0.z) + 10);
// Binary Op IADD mask 0xf:
r4.xyzw = intBitsToFloat(floatBitsToInt(r0.zzzz) + ivec4(6, 7, 8, 9));
// Binary Op LD mask 0x3:
r6.xy = texelFetch(idata, floatBitsToInt(r1.w)).xy;
// Trinary Op MAD
r6.zw = iTEXCOORD0.xy * vec2(1.000000, -1.000000) + vec2(0.000000, 1.000000);
// Unary Op FTOU
r6.zw = uintBitsToFloat(uvec2(r6.zw));
// Unary Op INEG
r0.z = intBitsToFloat(-floatBitsToInt(r6.w));
// Binary Op ULT mask 0x3:
r7.xy = intBitsToFloat(ivec2(lessThan(floatBitsToUint(r6.ww), uvec2(1, 2))));
// Binary Op ISHL mask 0x8:
r1.w = intBitsToFloat(1 << floatBitsToInt(r6.z));
// Binary Op AND mask 0x4:
r7.z = uintBitsToFloat(floatBitsToUint(r0.z) & floatBitsToUint(r7.y));
// Binary Op AND mask 0x3:
r6.xy = uintBitsToFloat(floatBitsToUint(r6.xy) & floatBitsToUint(r7.xz));
// Binary Op OR mask 0x4:
r0.z = uintBitsToFloat(floatBitsToUint(r6.y) | floatBitsToUint(r6.x));
// Binary Op AND mask 0x4:
r0.z = uintBitsToFloat(floatBitsToUint(r1.w) & floatBitsToUint(r0.z));
// Trinary Op MOVC
oSV_POSITION.w = bool(r0.z) ? uintBitsToFloat(0xFFC00000u) : r0.w;
// Binary Op DP4 mask 0x4:
oSV_POSITION.z = dot(r2.xyzw, camera_view_projection[2].xyzw);
// Unary Op MOV
oTEXCOORD0.xyz = r2.xyz;
// Unary Op UTOF
r0.z = float(floatBitsToUint(iTEXCOORD15.x));
// Binary Op ADD mask 0x8:
oTEXCOORD0.w = r0.z + 0.500000;
// Trinary Op MAD
r8.xy = iNORMAL0.xy * vec2(2.000000, 2.000000) + vec2(-1.000000, -1.000000);
// Binary Op ADD mask 0x4:
r8.z = -abs(r8.x) + 1.000000;
// Binary Op ADD mask 0x4:
r9.z = -abs(r8.y) + r8.z;
// Unary Op MOV
r8.z = clamp(-r9.z, 0.0, 1.0);
// Binary Op GE mask 0xc:
r0.zw = intBitsToFloat(ivec2(greaterThanEqual(r8.xy, vec2(0.000000, 0.000000))));
// Trinary Op MOVC
r8.zw = bool(r0.zw) ? -r8.zz : r8.zz;
// Binary Op ADD mask 0x3:
r9.xy = r8.zw + r8.xy;
// Binary Op DP3 mask 0x1:
r8.x = dot(r9.xyz, r9.xyz);
// Unary Op RSQ
r8.x = inversesqrt(r8.x);
// Binary Op MUL mask 0x7:
r8.xyz = r8.xxx * r9.xyz;
// Unary Op MOV
r2.x = r5.y;
// Unary Op MOV
r2.y = r3.x;
// Unary Op MOV
r2.z = r1.x;
// Binary Op DP3 mask 0x1:
r9.x = dot(r8.xyz, r2.xyz);
// Unary Op MOV
r3.x = r5.z;
// Unary Op MOV
r5.y = r3.z;
// Unary Op MOV
r5.z = r1.y;
// Unary Op MOV
r3.z = r1.z;
// Binary Op DP3 mask 0x4:
r9.z = dot(r8.xyz, r3.xyz);
// Binary Op DP3 mask 0x2:
r9.y = dot(r8.xyz, r5.xyz);
// Binary Op DP3 mask 0x1:
r8.x = dot(r9.xyz, r9.xyz);
// Unary Op RSQ
r8.x = inversesqrt(r8.x);
// Binary Op MUL mask 0x7:
oTEXCOORD1.xyz = r8.xxx * r9.xyz;
// Unary Op MOV
r1.xy = iTEXCOORD0.xy;
// Unary Op MOV
r1.zw = iTEXCOORD1.xy;
// Unary Op MOV
oTEXCOORD2.xyzw = r1.xyzw;
// Unary Op MOV
oTEXCOORD3.xy = iTEXCOORD2.xy;
// Binary Op LD mask 0x7:
r2.xyz = texelFetch(idata, floatBitsToInt(r4.y)).xyz;
// Binary Op MUL mask 0x7:
r2.xyz = r2.xyz * iPOSITION0.yyy;
// Binary Op LD mask 0x7:
r3.xyz = texelFetch(idata, floatBitsToInt(r4.x)).xyz;
// Trinary Op MAD
r2.xyz = iPOSITION0.xxx * r3.xyz + r2.xyz;
// Binary Op LD mask 0x7:
r3.xyz = texelFetch(idata, floatBitsToInt(r4.z)).xyz;
// Binary Op LD mask 0x7:
r4.xyz = texelFetch(idata, floatBitsToInt(r4.w)).xyz;
// Trinary Op MAD
r2.xyz = iPOSITION0.zzz * r3.xyz + r2.xyz;
// Trinary Op MAD
r1.xyz = iPOSITION0.www * r4.xyz + r2.xyz;
// Unary Op MOV
r1.w = 1.000000;
// Binary Op DP4 mask 0x1:
r2.x = dot(r1.xyzw, camera_last_view_projection[0].xyzw);
// Binary Op DP4 mask 0x2:
r2.y = dot(r1.xyzw, camera_last_view_projection[1].xyzw);
// Binary Op DP4 mask 0x4:
r2.z = dot(r1.xyzw, camera_last_view_projection[3].xyzw);
// Binary Op DIV mask 0xb:
r1.xyw = r2.xyz / r2.zzz;
// Binary Op ADD mask 0x3:
r0.xy = r0.xy + r1.xy;
// Trinary Op MAD
r1.xy = r0.xy * vec2(0.500000, 0.500000) + vec2(0.500000, 0.500000);
// Binary Op ADD mask 0x4:
r1.z = -r1.y + 1.000000;
// Binary Op MUL mask 0x7:
oTEXCOORD4.xyz = r2.zzz * r1.xzw;
// No Param Op RET
return;
}
