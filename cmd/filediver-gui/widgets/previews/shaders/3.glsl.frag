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

/* Resource Binding 0: __samp_decal_sheet
 *   Input Type: SAMPLER
 *   Return Type: NO_RETURN
 *   View Dimension: UNKNOWN
 *   Sample Count: 0
 *   Bind Count: 1
 *   Flags: NONE
 */

/* Resource Binding 1: __samp_pattern_masks_array
 *   Input Type: SAMPLER
 *   Return Type: NO_RETURN
 *   View Dimension: UNKNOWN
 *   Sample Count: 0
 *   Bind Count: 1
 *   Flags: NONE
 */

uniform samplerBuffer idata;
uniform sampler2D tex_decal_sheet;
uniform sampler2D tex_pattern_lut;
uniform sampler2DArray tex_pattern_masks_array;
uniform sampler2DArray tex_customization_camo_tiler_array;
uniform sampler2DArray tex_composite_array;
uniform sampler2D tex_base_data;
uniform sampler2D tex_material_lut;
uniform sampler2DArray tex_customization_material_detail_tiler_array;
uniform sampler2DArray tex_id_masks_array;
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

// Input Signature
layout(location = 0) in vec4 iSV_POSITION; // SV_POSITION
layout(location = 1) in vec4 iTEXCOORD0; // SV_UNDEFINED
layout(location = 2) in vec3 iTEXCOORD1; // SV_UNDEFINED
layout(location = 3) in vec4 iTEXCOORD2; // SV_UNDEFINED
layout(location = 4) in vec2 iTEXCOORD3; // SV_UNDEFINED
layout(location = 5) in vec3 iTEXCOORD4; // SV_UNDEFINED

// Output Signature
layout(location = 0) out vec4 oSV_TARGET0; // SV_UNDEFINED
layout(location = 1) out vec4 oSV_TARGET1; // SV_UNDEFINED
layout(location = 2) out vec4 oSV_TARGET2; // SV_UNDEFINED
layout(location = 3) out vec4 oSV_TARGET3; // SV_UNDEFINED

// Program type: PIXEL_SHADER
void main() {
/* Global Flags:
 * refactoringAllowed
 * enableMinPrecision
 */

vec4 icb[4];
icb[0] = vec4(1, 0, 0, 0);
icb[1] = vec4(0, 1, 0, 0);
icb[2] = vec4(0, 0, 1, 0);
icb[3] = vec4(0, 0, 0, 1);
// Declare Constant Buffer Immediate indexed, register cb0, size 46
// Declare Constant Buffer Immediate indexed, register cb1, size 2
// Declare Sampler s0 mode DEFAULT
// Declare Sampler s1 mode DEFAULT
// Declare BUFFER Resource t0 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2D Resource t1 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2D Resource t2 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2DARRAY Resource t3 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2DARRAY Resource t4 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2DARRAY Resource t5 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2D Resource t6 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2D Resource t7 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2DARRAY Resource t8 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare TEXTURE2DARRAY Resource t9 -> (FLOAT, FLOAT, FLOAT, FLOAT)
// Declare Input Pixel Shader SIV: LINEAR_NOPERSPECTIVE v0.xy, POSITION
// Declare Input Pixel Shader: LINEAR v1.xyzw
// Declare Input Pixel Shader: LINEAR v2.xyz
// Declare Input Pixel Shader: LINEAR v3.xyzw
// Declare Input Pixel Shader: LINEAR v4.xy
// Declare Input Pixel Shader: LINEAR v5.xyz
// Declare Output: o0.xyzw
// Declare Output: o1.xyzw
// Declare Output: o2.xyzw
// Declare Output: o3.xyzw
// Declare Temps: r0 - r23
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
vec4 r10;
vec4 r11;
vec4 r12;
vec4 r13;
vec4 r14;
vec4 r15;
vec4 r16;
vec4 r17;
vec4 r18;
vec4 r19;
vec4 r20;
vec4 r21;
vec4 r22;
vec4 r23;
// Unary Op FTOU
r0.x = uintBitsToFloat(uint(iTEXCOORD0.w));
// Unary Op FTOU
r0.y = uintBitsToFloat(uint(ioffset));
// Trinary Op IMAD
r0.x = intBitsToFloat(floatBitsToInt(r0.x) * 11 + floatBitsToInt(r0.y));
// Binary Op ADD mask 0xe:
r0.yzw = iTEXCOORD0.xyz + -cb_camera_pos.xyz;
// Binary Op DP3 mask 0x1:
r1.x = dot(iTEXCOORD1.xyz, iTEXCOORD1.xyz);
// Unary Op RSQ
r1.x = inversesqrt(r1.x);
// Binary Op MUL mask 0xe:
r1.yzw = r1.xxx * iTEXCOORD1.xyz;
// Unary Op DERIV_RTX_COARSE
r2.xyz = dFdx(iTEXCOORD0.yzx);
// Unary Op DERIV_RTY_COARSE
r3.xyz = dFdy(iTEXCOORD0.xyz);
// Binary Op DP3 mask 0x8:
r2.w = dot(r2.zxy, r1.yzw);
// Trinary Op MAD
r4.xyz = -r2.www * r1.yzw + r2.zxy;
// Binary Op DP3 mask 0x8:
r2.w = dot(r3.xyz, r1.yzw);
// Trinary Op MAD
r5.xyz = -r2.www * r1.yzw + r3.xyz;
// Binary Op MUL mask 0x7:
r6.xyz = r1.wyz * r2.xyz;
// Trinary Op MAD
r2.xyz = r1.zwy * r2.yzx + -r6.xyz;
// Binary Op DP3 mask 0x1:
r2.x = dot(r3.xyz, r2.xyz);
// Binary Op LT mask 0x1:
r2.x = intBitsToFloat(int(r2.x < 0.000000));
// Trinary Op MOVC
r7.x = bool(r2.x) ? -1.000000 : 1.000000;
// Unary Op DERIV_RTX_FINE
r2.xyzw = dFdx(iTEXCOORD2.xyzw);
// Unary Op DERIV_RTY_FINE
r3.xyzw = dFdy(iTEXCOORD2.xyzw);
// Binary Op MUL mask 0x3:
r6.xy = r3.yx * vec2(1.000000, -1.000000);
// Binary Op DP2 mask 0x1:
r2.x = dot(r2.xy, r6.xy);
// Binary Op LT mask 0x1:
r3.x = intBitsToFloat(int(r2.x < 0.000000));
// Trinary Op MOVC
r3.x = bool(r3.x) ? -1.000000 : 1.000000;
// Unary Op MOV
r6.zw = -r2.wy;
// Binary Op MUL mask 0x3:
r8.xy = r3.xx * r6.xw;
// Binary Op MUL mask 0xe:
r8.yzw = r5.xyz * r8.yyy;
// Trinary Op MAD
r8.xyz = r4.xyz * r8.xxx + r8.yzw;
// Binary Op LT mask 0x1:
r2.x = intBitsToFloat(int(0.000000 < abs(r2.x)));
// Binary Op DP3 mask 0x2:
r2.y = dot(r8.xyz, r8.xyz);
// Unary Op RSQ
r2.y = inversesqrt(r2.y);
// Binary Op MUL mask 0x7:
r9.xyz = r2.yyy * r8.xyz;
// Trinary Op MOVC
r8.xyz = bool(r2.xxx) ? r9.xyz : r8.xyz;
// Binary Op MUL mask 0x1:
r2.x = r7.x * r3.x;
// Binary Op MUL mask 0xe:
r7.yzw = r1.wyz * r8.yzx;
// Trinary Op MAD
r7.yzw = r1.zwy * r8.zxy + -r7.yzw;
// Binary Op MUL mask 0x7:
r9.xyz = r2.xxx * r7.yzw;
// Binary Op MUL mask 0x3:
r6.xy = r3.wz * vec2(1.000000, -1.000000);
// Binary Op DP2 mask 0x1:
r2.x = dot(r2.zw, r6.xy);
// Binary Op LT mask 0x2:
r2.y = intBitsToFloat(int(r2.x < 0.000000));
// Trinary Op MOVC
r2.y = bool(r2.y) ? -1.000000 : 1.000000;
// Binary Op MUL mask 0xc:
r2.zw = r2.yy * r6.xz;
// Binary Op MUL mask 0x7:
r3.xyz = r2.www * r5.xyz;
// Trinary Op MAD
r3.xyz = r4.xyz * r2.zzz + r3.xyz;
// Binary Op LT mask 0x1:
r2.x = intBitsToFloat(int(0.000000 < abs(r2.x)));
// Binary Op DP3 mask 0x4:
r2.z = dot(r3.xyz, r3.xyz);
// Unary Op RSQ
r2.z = inversesqrt(r2.z);
// Binary Op MUL mask 0x7:
r4.xyz = r2.zzz * r3.xyz;
// Trinary Op MOVC
r2.xzw = bool(r2.xxx) ? r4.xyz : r3.xyz;
// Binary Op MUL mask 0x2:
r2.y = r7.x * r2.y;
// Binary Op MUL mask 0x7:
r7.xyz = r1.wyz * r2.zwx;
// Trinary Op MAD
r7.xyz = r1.zwy * r2.wxz + -r7.xyz;
// Binary Op MUL mask 0x7:
r3.xyz = r2.yyy * r7.xyz;
// Binary Op IADD mask 0x1:
r0.x = intBitsToFloat(floatBitsToInt(r0.x) + 1);
// Binary Op LD mask 0x1:
r4.x = texelFetch(idata, floatBitsToInt(r0.x)).x;
// Binary Op RESINFO mask 0x1:
r0.x = intBitsToFloat(textureSize(tex_id_masks_array, 0).z);
// Unary Op MOV
r5.xy = iTEXCOORD2.xy;
// Unary Op MOV
r6.y = 0.000000;
// Unary Op MOV
r2.y = 0.000000;
// No Param Op LOOP
while(true) {
// Binary Op UGE mask 0x8:
r3.w = intBitsToFloat(int(floatBitsToUint(r2.y) >= floatBitsToUint(r0.x)));
// Single Param Op BREAKC
if (bool(r3.w)) { break; }
// Unary Op UTOF
r5.z = float(floatBitsToUint(r2.y));
// Trinary Op SAMPLE
r7.xyzw = texture(tex_id_masks_array, r5.xyz).xyzw;
// Binary Op ISHL mask 0x8:
r3.w = intBitsToFloat(floatBitsToInt(r2.y) << 2);
// Unary Op MOV
r4.z = r6.y;
// Unary Op MOV
r4.w = 0.000000;
// No Param Op LOOP
while(true) {
// Binary Op UGE mask 0x4:
r5.z = intBitsToFloat(int(floatBitsToUint(r4.w) >= 4));
// Single Param Op BREAKC
if (bool(r5.z)) { break; }
// Binary Op DP4 mask 0x4:
r5.z = dot(r7.xyzw, icb[floatBitsToInt(r4.w)].xyzw);
// Binary Op GE mask 0x4:
r5.z = intBitsToFloat(int(r5.z >= 0.496000));
// Single Param Op IF
if (bool(r5.z)) {
// Binary Op IADD mask 0x4:
r5.z = intBitsToFloat(floatBitsToInt(r3.w) + floatBitsToInt(r4.w));
// Unary Op MOV
r4.z = r5.z;
// No Param Op BREAK
break;
// No Param Op ENDIF
}
// Binary Op IADD mask 0x8:
r4.w = intBitsToFloat(floatBitsToInt(r4.w) + 1);
// No Param Op ENDLOOP
}
// Unary Op MOV
r6.y = r4.z;
// Binary Op IADD mask 0x2:
r2.y = intBitsToFloat(floatBitsToInt(r2.y) + 1);
// No Param Op ENDLOOP
}
// Trinary Op SAMPLE
r5.xyzw = texture(tex_base_data, iTEXCOORD2.xy).xyzw;
// Binary Op ADD mask 0x1:
r10.x = r5.w + -0.500000;
// Unary Op MOV
r7.z = clamp(r10.x, 0.0, 1.0);
// Unary Op MOV
r7.w = clamp(-r10.x, 0.0, 1.0);
// Trinary Op MAD
r10.yzw = r5.zxy * vec3(0.400000, -1.000000, 1.000000) + vec3(0.600000, 1.000000, 0.000000);
// Trinary Op MAD
r10.zw = r10.zw * vec2(2.000000, 2.000000) + vec2(-1.000000, -1.000000);
// Binary Op MUL mask 0x3:
r11.xy = r10.zw * r10.zw;
// Trinary Op MAD
r11.z = -r10.z * r10.z + 1.000000;
// Trinary Op MAD
r11.z = -r10.w * r10.w + r11.z;
// Binary Op MAX mask 0x1:
r11.x = max(r11.y, r11.x);
// Binary Op MUL mask 0x1:
r11.x = r11.x * 0.000061;
// Binary Op MAX mask 0x1:
r11.x = max(r11.x, r11.z);
// Unary Op RSQ
r11.x = inversesqrt(r11.x);
// Binary Op MUL mask 0xc:
r10.zw = r10.zw * -r11.xx;
// Binary Op MUL mask 0x7:
r11.xyz = r9.xyz * r10.www;
// Trinary Op MAD
r11.xyz = r10.zzz * r8.xyz + r11.xyz;
// Binary Op ADD mask 0x4:
r10.z = -r10.y + 1.000000;
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r8.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op LD mask 0xf:
r9.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.zy), 0).xyzw;
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r12.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op RESINFO mask 0x1:
r0.x = intBitsToFloat(textureSize(tex_customization_material_detail_tiler_array, 0).z);
// Binary Op USHR mask 0x2:
r4.y = uintBitsToFloat(floatBitsToUint(r4.x) >> 9);
// Quatenary Op BFI
uvec4 bitmask;
bitmask.xy = (((uvec2(1) << uvec2(23, 23))-uvec2(1)) << uvec2(0, 0)) & 0xffffffff;
r4.xy = ((floatBitsToUint(r4.xy) << uvec2(0, 0)) & bitmask.xy) | (uvec2(1065353216, 1065353216) & ~bitmask.xy);
// Binary Op ADD mask 0x3:
r4.xy = r4.xy + vec2(-1.000000, -1.000000);
// Binary Op GE mask 0x2:
r2.y = intBitsToFloat(int(3.000000 >= r8.w));
// Single Param Op IF
if (bool(r2.y)) {
// Binary Op MUL mask 0x8:
r3.w = r12.x * detail_tile_factor_mult;
// Trinary Op MAD
r13.xy = iTEXCOORD2.zw * r3.ww + r4.xy;
// Unary Op FTOU
r3.w = uintBitsToFloat(uint(r9.x));
// Trinary Op UDIV
r3.w = uintBitsToFloat(floatBitsToUint(r3.w) % floatBitsToUint(r0.x));
// Unary Op UTOF
r13.z = float(floatBitsToUint(r3.w));
// Trinary Op SAMPLE
r13.xyzw = texture(tex_customization_material_detail_tiler_array, r13.xyz).xyzw;
// Binary Op ADD mask 0xf:
r13.xyzw = r13.zwxy + vec4(-0.500000, -0.500000, -0.500000, -0.500000);
// Binary Op MUL mask 0x3:
r14.xy = r13.zw * vec2(-1.000000, 1.000000);
// No Param Op ELSE
} else {
// Unary Op MOV
r14.xy = vec2(0.000000, 0.000000);
// Unary Op MOV
r13.xy = vec2(0.000000, 0.000000);
// No Param Op ENDIF
}
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r15.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).zxyw;
// Binary Op GE mask 0x8:
r3.w = intBitsToFloat(int(r8.w >= 2.000000));
// Binary Op LT mask 0x4:
r4.z = intBitsToFloat(int(0.000000 < r15.x));
// Binary Op AND mask 0x4:
r4.z = uintBitsToFloat(floatBitsToUint(r3.w) & floatBitsToUint(r4.z));
// Single Param Op IF
if (bool(r4.z)) {
// Binary Op MUL mask 0x4:
r4.z = r12.x * detail_tile_factor_mult;
// Trinary Op MAD
r4.xy = iTEXCOORD2.zw * r4.zz + r4.xy;
// Unary Op FTOU
r4.w = uintBitsToFloat(uint(r15.z));
// Trinary Op UDIV
r4.w = uintBitsToFloat(floatBitsToUint(r4.w) % floatBitsToUint(r0.x));
// Unary Op UTOF
r4.z = float(floatBitsToUint(r4.w));
// Trinary Op SAMPLE
r4.xyzw = texture(tex_customization_material_detail_tiler_array, r4.xyz).xyzw;
// Binary Op ADD mask 0xf:
r4.xyzw = r4.xyzw + vec4(-0.500000, -0.500000, -0.500000, -0.500000);
// Binary Op MUL mask 0xc:
r14.zw = r4.xy * vec2(-2.000000, 2.000000);
// Binary Op MUL mask 0x3:
r16.xy = r14.zw * r14.zw;
// Trinary Op MAD
r10.w = -r14.z * r14.z + 1.000000;
// Trinary Op MAD
r10.w = -r14.w * r14.w + r10.w;
// Binary Op MAX mask 0x8:
r11.w = max(r16.y, r16.x);
// Binary Op MUL mask 0x8:
r11.w = r11.w * 0.000061;
// Binary Op MAX mask 0x8:
r10.w = max(r10.w, r11.w);
// Unary Op RSQ
r10.w = inversesqrt(r10.w);
// Binary Op MUL mask 0xc:
r14.zw = r14.zw * -r10.ww;
// Binary Op MUL mask 0xc:
r14.zw = r15.ww * r14.zw;
// Binary Op MUL mask 0x7:
r16.xyz = r3.xyz * r14.www;
// Trinary Op MAD
r16.xyz = r14.zzz * r2.xzw + r16.xyz;
// Binary Op ADD mask 0x7:
r16.xyz = r11.xyz + r16.xyz;
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r17.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Unary Op MOV
r7.xy = r4.zw;
// Binary Op DP3 mask 0x8:
r10.w = dot(r17.xyz, r7.xyz);
// Binary Op ADD mask 0x8:
r10.w = r17.w + r10.w;
// No Param Op ELSE
} else {
// Unary Op MOV
r16.xyz = r11.xyz;
// Unary Op MOV
r15.x = 0.000000;
// Unary Op MOV
r10.w = 0.000000;
// No Param Op ENDIF
}
// Binary Op EQ mask 0x3:
r18.xy = intBitsToFloat(ivec2(equal(r8.ww, vec2(4.000000, 1.000000))));
// Single Param Op IF
if (bool(r18.x)) {
// Binary Op MUL mask 0x4:
r18.z = r12.x * detail_tile_factor_mult;
// Binary Op MUL mask 0x4:
r18.z = r18.z * 4.000000;
// Unary Op FTOU
r18.w = uintBitsToFloat(uint(r9.x));
// Trinary Op UDIV
r0.x = uintBitsToFloat(floatBitsToUint(r18.w) % floatBitsToUint(r0.x));
// Binary Op MUL mask 0x3:
r19.xy = r18.zz * iTEXCOORD2.zw;
// Unary Op UTOF
r19.z = float(floatBitsToUint(r0.x));
// Trinary Op SAMPLE
r17.xyz = texture(tex_composite_array, r19.xyz).xyz;
// Binary Op ADD mask 0x7:
r13.xyz = r17.zxy + vec3(-0.500000, -0.500000, -0.500000);
// Unary Op MOV
r13.w = -r13.y;
// Unary Op MOV
r14.zw = r13.wz;
// Unary Op MOV
r13.y = 0.000000;
// No Param Op ELSE
} else {
// Unary Op MOV
r14.zw = r14.xy;
// No Param Op ENDIF
}
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r4.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r17.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op MUL mask 0x8:
r11.w = r10.x * r17.z;
// Binary Op MAX mask 0x8:
r11.w = max(r11.w, 0.000000);
// Binary Op MUL mask 0x8:
r11.w = r11.w * 4.000000;
// Binary Op MUL mask 0x3:
r19.xy = r11.ww * r17.xy;
// Unary Op MOV
r19.zw = vec2(0.000000, 0.000000);
// Binary Op ADD mask 0xf:
r17.xyzw = r17.xyzw + r19.xyzw;
// Trinary Op MAD
r13.zw = r5.wz * vec2(1.000000, -0.400000) + vec2(-0.500000, 0.400000);
// Binary Op DP4 mask 0x8:
r11.w = dot(r17.xyzw, r13.xyzw);
// Binary Op ADD mask 0x8:
r11.w = clamp(r9.w + r11.w, 0.0, 1.0);
// Binary Op MUL mask 0x1:
r12.x = r10.x * r4.z;
// Binary Op MAX mask 0x1:
r12.x = max(r12.x, 0.000000);
// Binary Op MUL mask 0x1:
r12.x = r12.x * 4.000000;
// Binary Op MUL mask 0x3:
r5.xy = r4.xy * r12.xx;
// Unary Op MOV
r5.zw = vec2(0.000000, 0.000000);
// Binary Op ADD mask 0xf:
r4.xyzw = r4.xyzw + r5.xyzw;
// Binary Op ADD mask 0x1:
r12.x = clamp(r13.y + 0.500000, 0.0, 1.0);
// Binary Op ADD mask 0x4:
r15.z = -r12.x + 1.000000;
// Trinary Op MAD
r12.x = r10.z * r15.z + r12.x;
// Binary Op MUL mask 0x8:
r5.w = r10.z * r12.x;
// Unary Op MOV
r5.xyz = r13.xyz;
// Binary Op DP4 mask 0x4:
r10.z = dot(r4.xyzw, r5.xyzw);
// Binary Op ADD mask 0x4:
r10.z = clamp(r9.z + r10.z, 0.0, 1.0);
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r4.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r17.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op ADD mask 0x1:
r12.x = r4.w + -1.000000;
// Trinary Op MAD
r12.x = r10.z * r12.x + 1.000000;
// Binary Op ADD mask 0x4:
r15.z = -r12.x + r17.w;
// Trinary Op MAD
r12.x = r11.w * r15.z + r12.x;
// Single Param Op IF
if (bool(r2.y)) {
// Binary Op ADD mask 0x3:
r14.xy = r14.xy + r14.xy;
// Binary Op MUL mask 0xc:
r15.zw = r14.xy * r14.xy;
// Trinary Op MAD
r16.w = -r14.x * r14.x + 1.000000;
// Trinary Op MAD
r16.w = -r14.y * r14.y + r16.w;
// Binary Op MAX mask 0x4:
r15.z = max(r15.w, r15.z);
// Binary Op MUL mask 0x4:
r15.z = r15.z * 0.000061;
// Binary Op MAX mask 0x4:
r15.z = max(r15.z, r16.w);
// Unary Op RSQ
r15.z = inversesqrt(r15.z);
// Binary Op MUL mask 0x3:
r14.xy = r14.xy * -r15.zz;
// Binary Op MUL mask 0x3:
r14.xy = r9.yy * r14.xy;
// Binary Op MUL mask 0x3:
r14.xy = r12.xx * r14.xy;
// Binary Op MUL mask 0x7:
r20.xyz = r3.xyz * r14.yyy;
// Trinary Op MAD
r20.xyz = r14.xxx * r2.xzw + r20.xyz;
// Binary Op ADD mask 0x7:
r11.xyz = r11.xyz + r20.xyz;
// No Param Op ENDIF
}
// Single Param Op IF
if (bool(r18.x)) {
// Binary Op ADD mask 0x3:
r14.xy = r14.zw + r14.zw;
// Binary Op MUL mask 0xc:
r14.zw = r14.xy * r14.xy;
// Trinary Op MAD
r15.z = -r14.x * r14.x + 1.000000;
// Trinary Op MAD
r15.z = -r14.y * r14.y + r15.z;
// Binary Op MAX mask 0x4:
r14.z = max(r14.w, r14.z);
// Binary Op MUL mask 0x4:
r14.z = r14.z * 0.000061;
// Binary Op MAX mask 0x4:
r14.z = max(r14.z, r15.z);
// Unary Op RSQ
r14.z = inversesqrt(r14.z);
// Binary Op MUL mask 0x3:
r14.xy = r14.xy * -r14.zz;
// Binary Op MUL mask 0x3:
r14.xy = r9.yy * r14.xy;
// Binary Op MUL mask 0x3:
r14.xy = r12.xx * r14.xy;
// Binary Op MUL mask 0xe:
r14.yzw = r3.xyz * r14.yyy;
// Trinary Op MAD
r14.xyz = r14.xxx * r2.xzw + r14.yzw;
// Binary Op ADD mask 0x7:
r11.xyz = r11.xyz + r14.xyz;
// No Param Op ENDIF
}
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r2.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r9.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op MUL mask 0x1:
r12.x = r10.x * r9.z;
// Binary Op MAX mask 0x1:
r12.x = max(r12.x, 0.000000);
// Binary Op MUL mask 0x1:
r12.x = r12.x * 4.000000;
// Binary Op MUL mask 0x3:
r14.xy = r9.xy * r12.xx;
// Unary Op MOV
r14.zw = vec2(0.000000, 0.000000);
// Binary Op ADD mask 0xf:
r9.xyzw = r9.xyzw + r14.xyzw;
// Binary Op DP4 mask 0x1:
r12.x = dot(r9.xyzw, r13.xyzw);
// Binary Op ADD mask 0x1:
r12.x = clamp(r2.w + r12.x, 0.0, 1.0);
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r9.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op MUL mask 0x1:
r14.x = r10.x * r9.z;
// Binary Op MAX mask 0x1:
r14.x = max(r14.x, 0.000000);
// Binary Op MUL mask 0x1:
r14.x = r14.x * 4.000000;
// Binary Op MUL mask 0x3:
r14.xy = r9.xy * r14.xx;
// Unary Op MOV
r14.z = 0.000000;
// Binary Op ADD mask 0x7:
r14.xyz = r9.xyz + r14.xyz;
// Binary Op DP3 mask 0x1:
r14.x = dot(r14.xyz, r5.xyz);
// Binary Op ADD mask 0x1:
r14.x = clamp(r9.w + r14.x, 0.0, 1.0);
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r9.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op MUL mask 0x2:
r14.y = r10.x * r9.z;
// Binary Op MAX mask 0x2:
r14.y = max(r14.y, 0.000000);
// Binary Op MUL mask 0x2:
r14.y = r14.y * 4.000000;
// Binary Op MUL mask 0x3:
r13.xy = r9.xy * r14.yy;
// Unary Op MOV
r13.zw = vec2(0.000000, 0.000000);
// Binary Op ADD mask 0xf:
r9.xyzw = r9.xyzw + r13.xyzw;
// Unary Op MOV
r7.xy = r5.xy;
// Binary Op DP4 mask 0x2:
r14.y = dot(r9.xyzw, r7.xyzw);
// Binary Op ADD mask 0x2:
r14.y = clamp(r15.y + r14.y, 0.0, 1.0);
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r7.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op MUL mask 0x4:
r14.z = r10.x * r7.z;
// Binary Op MAX mask 0x4:
r14.z = max(r14.z, 0.000000);
// Binary Op MUL mask 0x4:
r14.z = r14.z * 4.000000;
// Binary Op MUL mask 0x3:
r20.xy = r7.xy * r14.zz;
// Unary Op MOV
r20.z = 0.000000;
// Binary Op ADD mask 0xe:
r15.yzw = r7.xyz + r20.xyz;
// Binary Op DP3 mask 0x4:
r14.z = dot(r15.yzw, r5.xyz);
// Binary Op ADD mask 0x4:
r14.z = clamp(r7.w + r14.z, 0.0, 1.0);
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r7.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op MUL mask 0x8:
r14.w = r10.x * r7.z;
// Binary Op MAX mask 0x8:
r14.w = max(r14.w, 0.000000);
// Binary Op MUL mask 0x8:
r14.w = r14.w * 4.000000;
// Binary Op MUL mask 0x3:
r20.xy = r7.xy * r14.ww;
// Unary Op MOV
r20.z = 0.000000;
// Binary Op ADD mask 0xe:
r15.yzw = r7.xyz + r20.xyz;
// Binary Op DP3 mask 0x8:
r14.w = dot(r15.yzw, r5.xyz);
// Binary Op ADD mask 0x8:
r14.w = clamp(r7.w + r14.w, 0.0, 1.0);
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r7.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).zwxy;
// Unary Op MOV
r7.xy = clamp(r7.xy, 0.0, 1.0);
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r9.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).xyzw;
// Binary Op GE mask 0x1:
r0.x = intBitsToFloat(int(r9.w >= 0.000000));
// Single Param Op IF
if (bool(r0.x)) {
// Unary Op MOV
r13.xzw = vec3(0.000000, 0.000000, 0.000000);
// Unary Op MOV
r13.y = r6.y;
// Binary Op LD mask 0xf:
r19.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r13.xy), 0).xyzw;
// Unary Op MOV
r13.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r20.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r13.xy), 0).xyzw;
// Unary Op MOV
r13.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r21.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r13.xy), 0).xyzw;
// Unary Op MOV
r13.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r22.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r13.xy), 0).xywz;
// Unary Op MOV
r13.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0x2:
r15.y = texelFetch(tex_material_lut, floatBitsToInt(r13.xy), 0).w;
// Binary Op MUL mask 0x1:
r0.x = r9.z * detail_tile_factor_mult;
// Binary Op MUL mask 0x3:
r3.xy = r0.xx * iTEXCOORD2.zw;
// Unary Op MOV
r3.z = r9.w;
// Trinary Op SAMPLE
r23.xyz = texture(tex_customization_camo_tiler_array, r3.xyz).xyz;
// Binary Op ADD mask 0x7:
r23.xyz = r23.xyz + vec3(-0.500000, -0.500000, -0.500000);
// Trinary Op MAD
r23.xyz = clamp(r23.xyz * r9.xxx + r9.yyy, 0.0, 1.0);
// Binary Op ADD mask 0x7:
r20.xyz = -r19.xyz + r20.xyz;
// Trinary Op MAD
r20.xyz = r23.xxx * r20.xyz + r19.xyz;
// Binary Op ADD mask 0x7:
r21.xyz = -r20.xyz + r21.xyz;
// Trinary Op MAD
r20.xyz = r23.yyy * r21.xyz + r20.xyz;
// Binary Op ADD mask 0x7:
r21.xyz = -r20.xyz + r22.xyw;
// Trinary Op MAD
r20.xyz = r23.zzz * r21.xyz + r20.xyz;
// Binary Op NE mask 0x1:
r0.x = intBitsToFloat(int(r8.w != 1.000000));
// Single Param Op IF
if (bool(r0.x)) {
// Binary Op MUL mask 0x4:
r15.z = r10.x * r22.z;
// Binary Op MAX mask 0x4:
r15.z = max(r15.z, 0.000000);
// Binary Op MUL mask 0x4:
r15.z = r15.z * 4.000000;
// Binary Op MUL mask 0x1:
r21.x = r15.z * r20.w;
// Binary Op MUL mask 0x2:
r21.y = r15.z * r21.w;
// Unary Op MOV
r22.x = r20.w;
// Unary Op MOV
r22.y = r21.w;
// Unary Op MOV
r21.z = 0.000000;
// Binary Op ADD mask 0x7:
r21.xyz = r21.xyz + r22.xyz;
// Binary Op DP3 mask 0x4:
r15.z = dot(r21.xyz, r5.xyz);
// Binary Op ADD mask 0x2:
r15.y = clamp(-r15.y + r15.z, 0.0, 1.0);
// Binary Op ADD mask 0x7:
r21.xyz = r8.xyz + -r20.xyz;
// Trinary Op MAD
r8.xyz = r15.yyy * r21.xyz + r20.xyz;
// Binary Op ADD mask 0x4:
r15.z = -r15.y + 1.000000;
// Binary Op ADD mask 0x8:
r15.w = r12.x + -r15.z;
// Binary Op MAX mask 0x1:
r12.x = max(r15.w, 0.000000);
// Binary Op MUL mask 0x8:
r12.w = r12.w * r15.z;
// Binary Op ADD mask 0x4:
r15.z = -r14.y + r19.w;
// Trinary Op MAD
r14.y = r12.w * r15.z + r14.y;
// Binary Op MUL mask 0x4:
r14.z = r14.z * r15.y;
// No Param Op ELSE
} else {
// Unary Op MOV
r8.xyz = r20.xyz;
// Unary Op MOV
r15.y = 1.000000;
// No Param Op ENDIF
}
// No Param Op ELSE
} else {
// Unary Op MOV
r15.y = 1.000000;
// No Param Op ENDIF
}
// Binary Op NE mask 0x1:
r0.x = intBitsToFloat(int(r8.w != 3.000000));
// Single Param Op IF
if (bool(r0.x)) {
// Binary Op LD mask 0xf:
r9.xyzw = texelFetch(tex_pattern_lut, floatBitsToInt(uvec2(0, 0)), 0).xyzw;
// Binary Op NE mask 0x1:
r0.x = intBitsToFloat(int(r9.w != -1.000000));
// Single Param Op IF
if (bool(r0.x)) {
// Binary Op RESINFO mask 0x1:
r0.x = intBitsToFloat(textureSize(tex_pattern_masks_array, 0).z);
// Binary Op LD mask 0xf:
r13.xyzw = texelFetch(tex_pattern_lut, floatBitsToInt(uvec2(1, 0)), 0).xyzw;
// Binary Op MUL mask 0x1:
r10.x = r10.x * r13.z;
// Binary Op MAX mask 0x1:
r10.x = max(r10.x, 0.000000);
// Binary Op MUL mask 0x1:
r10.x = r10.x * 4.000000;
// Binary Op MUL mask 0x3:
r20.xy = r10.xx * r13.xy;
// Unary Op MOV
r20.z = 0.000000;
// Binary Op ADD mask 0x7:
r20.xyz = r13.xyz + r20.xyz;
// Binary Op DP3 mask 0x1:
r10.x = dot(r20.xyz, r5.xyz);
// Binary Op ADD mask 0x1:
r10.x = clamp(r13.w + r10.x, 0.0, 1.0);
// Unary Op FTOU
r3.x = uintBitsToFloat(uint(r9.w));
// Trinary Op UDIV
r0.x = uintBitsToFloat(floatBitsToUint(r3.x) % floatBitsToUint(r0.x));
// Unary Op UTOF
r3.z = float(floatBitsToUint(r0.x));
// Unary Op MOV
r3.xy = iTEXCOORD2.xy;
// Trinary Op SAMPLE
r0.x = texture(tex_pattern_masks_array, r3.xyz).x;
// Binary Op ADD mask 0x1:
r0.x = r0.x + -0.500000;
// Binary Op MUL mask 0x1:
r0.x = clamp(r0.x * 100.000000, 0.0, 1.0);
// Binary Op MUL mask 0x8:
r12.w = r0.x * r10.x;
// Binary Op LT mask 0x1:
r3.x = intBitsToFloat(int(0.000000 < r12.w));
// Single Param Op IF
if (bool(r3.x)) {
// Binary Op ADD mask 0x7:
r20.xyz = -r8.xyz + r9.xyz;
// Trinary Op MAD
r8.xyz = r12.www * r20.xyz + r8.xyz;
// Binary Op NE mask 0x1:
r3.x = intBitsToFloat(int(r8.w != 1.000000));
// Single Param Op IF
if (bool(r3.x)) {
// Trinary Op MAD
r15.z = -r12.w * r12.w + r11.w;
// Binary Op ADD mask 0x4:
r15.z = r15.z + 1.000000;
// Binary Op MIN mask 0x4:
r15.z = min(r15.z, 1.000000);
// Binary Op MUL mask 0x1:
r12.x = r12.x * r15.z;
// Binary Op LD mask 0x4:
r15.z = texelFetch(tex_pattern_lut, floatBitsToInt(uvec2(2, 0)), 0).w;
// Trinary Op MAD
r15.w = r10.x * r0.x + -r11.w;
// Binary Op MAX mask 0x8:
r15.w = max(r15.w, 0.000000);
// Binary Op ADD mask 0x8:
r16.w = -r14.y + 2.000000;
// Unary Op RCP
r16.w = 1.0 / r16.w;
// Binary Op MUL mask 0x8:
r15.w = r15.w * r16.w;
// Binary Op ADD mask 0x4:
r15.z = -r14.y + r15.z;
// Trinary Op MAD
r14.y = r15.w * r15.z + r14.y;
// Trinary Op MAD
r15.z = -r10.x * r0.x + 1.000000;
// Binary Op MUL mask 0x4:
r14.z = r14.z * r15.z;
// Binary Op ADD mask 0x4:
r15.z = -r14.x + 0.500000;
// Trinary Op MAD
r14.x = r12.w * r15.z + r14.x;
// No Param Op ENDIF
}
// Trinary Op MAD
r10.x = -r10.x * r0.x + 1.000000;
// Binary Op MUL mask 0x1:
r10.x = r10.x * r15.x;
// Trinary Op MOVC
r15.x = bool(r3.w) ? r10.x : r15.x;
// Binary Op ADD mask 0x7:
r20.xyz = -r11.xyz + r16.xyz;
// Trinary Op MAD
r20.xyz = r12.www * r20.xyz + r11.xyz;
// Trinary Op MOVC
r11.xyz = bool(r18.xxx) ? r20.xyz : r11.xyz;
// No Param Op ENDIF
}
// No Param Op ENDIF
}
// No Param Op ENDIF
}
// Binary Op ADD mask 0x3:
r3.xy = iTEXCOORD3.xy + vec2(-0.500000, -0.500000);
// Binary Op LT mask 0x3:
r3.xy = intBitsToFloat(ivec2(lessThan(abs(r3.xy), vec2(0.500000, 0.500000))));
// Binary Op AND mask 0x1:
r0.x = uintBitsToFloat(floatBitsToUint(r3.y) & floatBitsToUint(r3.x));
// Single Param Op IF
if (bool(r0.x)) {
// Trinary Op SAMPLE
r3.xyzw = texture(tex_decal_sheet, iTEXCOORD3.xy).xyzw;
// Binary Op ADD mask 0x1:
r10.x = clamp(-r11.w + r3.w, 0.0, 1.0);
// Binary Op ADD mask 0x7:
r20.xyz = -r8.xyz + r3.xyz;
// Trinary Op MAD
r8.xyz = r10.xxx * r20.xyz + r8.xyz;
// Binary Op ADD mask 0x8:
r12.w = -r10.x + 1.000000;
// Binary Op MUL mask 0x1:
r12.x = r12.w * r12.x;
// Trinary Op MAD
r14.z = r10.x * -r14.z + r14.z;
// Binary Op ADD mask 0x1:
r10.x = -r10.z + r10.x;
// Binary Op MAX mask 0x1:
r10.x = max(r10.x, 0.000000);
// Binary Op ADD mask 0x8:
r12.w = -r14.y + 0.400000;
// Trinary Op MAD
r14.y = r10.x * r12.w + r14.y;
// No Param Op ENDIF
}
// Single Param Op IF
if (bool(r18.y)) {
// Unary Op MOV
r7.zw = clamp(r7.zw, 0.0, 1.0);
// Binary Op ADD mask 0x1:
r10.x = -r10.z + 1.000000;
// Binary Op MUL mask 0x1:
r10.x = r7.w * r10.x;
// Unary Op MOV
r6.xzw = vec3(0.000000, 0.000000, 0.000000);
// Binary Op LD mask 0xf:
r3.xyzw = texelFetch(tex_material_lut, floatBitsToInt(r6.xy), 0).wxyz;
// Unary Op SQRT
r20.xyz = sqrt(r8.xyz);
// Binary Op ADD mask 0x7:
r20.xyz = -r3.yzw + r20.xyz;
// Trinary Op MAD
r20.xyz = r7.zzz * r20.xyz + r3.yzw;
// Trinary Op MAD
r12.w = -r10.x * r14.y + 1.000000;
// Binary Op DP3 mask 0x4:
r15.z = clamp(dot(r8.xyz, vec3(0.300000, 0.590000, 0.110000)), 0.0, 1.0);
// Binary Op ADD mask 0x8:
r15.w = -r12.w + 1.000000;
// Trinary Op MAD
r12.w = r15.z * r15.w + r12.w;
// No Param Op ELSE
} else {
// Unary Op MOV
r20.xyz = vec3(0.000000, 0.000000, 0.000000);
// Unary Op MOV
r3.x = 0.000000;
// Unary Op MOV
r10.x = 0.000000;
// Unary Op MOV
r12.w = 1.000000;
// No Param Op ENDIF
}
// Binary Op ADD mask 0x7:
r21.xyz = r4.xyz + -r8.xyz;
// Trinary Op MAD
r21.xyz = r10.zzz * r21.xyz + r8.xyz;
// Binary Op MUL mask 0x4:
r10.z = r12.z * r10.z;
// Binary Op ADD mask 0x2:
r12.y = r12.y + -r14.y;
// Trinary Op MAD
r10.z = clamp(r10.z * r12.y + r14.y, 0.0, 1.0);
// Binary Op MUL mask 0x2:
r12.y = r11.w * r15.y;
// Binary Op ADD mask 0x7:
r22.xyz = r2.xyz + -r21.xyz;
// Trinary Op MAD
r21.xyz = r12.yyy * r22.xyz + r21.xyz;
// Binary Op MUL mask 0x8:
r11.w = r11.w * r11.w;
// Binary Op MUL mask 0x8:
r11.w = r15.y * r11.w;
// Binary Op ADD mask 0xe:
r15.yzw = r17.xyz + -r21.xyz;
// Trinary Op MAD
r15.yzw = r11.www * r15.yzw + r21.xyz;
// Binary Op MAX mask 0xe:
r15.yzw = max(r15.yzw, vec3(0.000061, 0.000061, 0.000061));
// Binary Op LT mask 0x7:
r6.xyz = intBitsToFloat(ivec3(lessThan(vec3(0.040450, 0.040450, 0.040450), r15.yzw)));
// Trinary Op MAD
r17.xyz = r15.yzw * vec3(0.947867, 0.947867, 0.947867) + vec3(0.052133, 0.052133, 0.052133);
// Unary Op LOG
r17.xyz = log2(r17.xyz);
// Binary Op MUL mask 0x7:
r17.xyz = r17.xyz * vec3(2.400000, 2.400000, 2.400000);
// Unary Op EXP
r17.xyz = exp2(r17.xyz);
// Binary Op MUL mask 0xe:
r15.yzw = r15.yzw * vec3(0.077399, 0.077399, 0.077399);
// Trinary Op MOVC
r15.yzw = bool(r6.xyz) ? r17.xyz : r15.yzw;
// Trinary Op MAD
r11.xyz = iTEXCOORD1.xyz * r1.xxx + r11.xyz;
// Binary Op DP3 mask 0x8:
r11.w = dot(r11.xyz, r11.xyz);
// Unary Op RSQ
r11.w = inversesqrt(r11.w);
// Binary Op MUL mask 0x7:
r11.xyz = r11.www * r11.xyz;
// Binary Op DP3 mask 0x1:
r0.x = dot(r0.yzw, r0.yzw);
// Unary Op RSQ
r0.x = inversesqrt(r0.x);
// Binary Op MUL mask 0x7:
r0.xyz = r0.xxx * r0.yzw;
// Binary Op DIV mask 0x3:
r6.xy = iSV_POSITION.xy / vp_render_resolution.xy;
// Binary Op DIV mask 0xc:
r6.zw = iTEXCOORD4.xy / iTEXCOORD4.zz;
// Binary Op ADD mask 0x3:
r6.xy = -r6.zw + r6.xy;
// Binary Op MIN mask 0xe:
r15.yzw = min(r15.yzw, vec3(1.000000, 1.000000, 1.000000));
// Binary Op LT mask 0x8:
r0.w = intBitsToFloat(int(0.000000 < r12.x));
// Trinary Op MOVC
r10.x = bool(r0.w) ? 0.000000 : r10.x;
// Binary Op LT mask 0x8:
r0.w = intBitsToFloat(int(0.000000 < r10.x));
// Trinary Op MOVC
r11.w = bool(r0.w) ? 0.000000 : r15.x;
// Unary Op MOV
r14.x = clamp(r14.x, 0.0, 1.0);
// Trinary Op MAD
r10.z = r10.z * 0.955000 + 0.045000;
// Binary Op DP3 mask 0x2:
r1.y = dot(-r0.xyz, r1.yzw);
// Binary Op ADD mask 0x4:
r1.z = -abs(r1.y) + 1.000000;
// Trinary Op MAD
r1.y = abs(r1.y) * r12.w + r1.z;
// Binary Op MUL mask 0x2:
r1.y = r10.y * r1.y;
// Binary Op MUL mask 0x4:
r1.z = r14.w * r14.z;
// Binary Op MAX mask 0x8:
r0.w = max(r14.z, 0.000000);
// Binary Op ADD mask 0x4:
r6.z = -r7.x + 1.000000;
// Binary Op MUL mask 0x4:
r6.z = r7.y * r6.z;
// Trinary Op MAD
r0.w = r6.z * r0.w + r7.x;
// Binary Op MUL mask 0x8:
r0.w = r0.w * r0.w;
// Binary Op DP3 mask 0x8:
r1.w = dot(r11.xyz, -r0.xyz);
// Binary Op MAX mask 0x8:
r1.w = max(r1.w, 0.000100);
// Unary Op MOV
r20.xyz = clamp(r20.xyz, 0.0, 1.0);
// Unary Op MOV
r3.x = clamp(r3.x, 0.0, 1.0);
// Trinary Op MAD
r1.w = -r1.w * r3.x + r3.x;
// Binary Op MAX mask 0xc:
r1.zw = max(r1.zw, vec2(0.000000, 0.000000));
// Binary Op ADD mask 0xe:
r12.yzw = -r15.yzw + r20.xyz;
// Trinary Op MAD
r12.yzw = r1.www * r12.yzw + r15.yzw;
// Unary Op MOV
r1.w = clamp(material_wetness, 0.0, 1.0);
// Binary Op LT mask 0x1:
r0.x = intBitsToFloat(int(0.010000 < r11.w));
// Trinary Op MOVC
r10.y = bool(r0.x) ? r11.w : 0.000000;
// Unary Op FTOU
r0.xy = uintBitsToFloat(uvec2(iSV_POSITION.xy));
// Unary Op FTOU
r0.z = uintBitsToFloat(uint(frame_number));
// Binary Op AND mask 0x4:
r0.z = uintBitsToFloat(floatBitsToUint(r0.z) & 1);
// Unary Op FTOU
r6.z = uintBitsToFloat(uint(taa_enabled));
// Binary Op IADD mask 0x1:
r0.x = intBitsToFloat(floatBitsToInt(r0.y) + floatBitsToInt(r0.x));
// Trinary Op IMAD
r0.x = intBitsToFloat(floatBitsToInt(r0.z) * floatBitsToInt(r6.z) + floatBitsToInt(r0.x));
// Binary Op AND mask 0x1:
r0.x = uintBitsToFloat(floatBitsToUint(r0.x) & 1);
// Unary Op UTOF
r0.x = float(floatBitsToUint(r0.x));
// Binary Op LT mask 0x2:
r0.y = intBitsToFloat(int(0.000000 < r10.y));
// Single Param Op IF
if (bool(r0.y)) {
// Trinary Op MAD
r14.yzw = iTEXCOORD1.xyz * r1.xxx + r16.xyz;
// Binary Op DP3 mask 0x1:
r1.x = dot(r14.yzw, r14.yzw);
// Unary Op RSQ
r1.x = inversesqrt(r1.x);
// Unary Op MOV
r10.w = clamp(r10.w, 0.0, 1.0);
// Trinary Op MAD
r10.w = r10.w * 0.955000 + 0.045000;
// Binary Op ADD mask 0x8:
r10.w = -r10.z + r10.w;
// Trinary Op MAD
r10.z = r0.x * r10.w + r10.z;
// Trinary Op MAD
r14.yzw = r14.yzw * r1.xxx + -r11.xyz;
// Trinary Op MAD
r11.xyz = r0.xxx * r14.yzw + r11.xyz;
// No Param Op ENDIF
}
// Binary Op MUL mask 0x1:
r1.x = r1.y * r1.y;
// Trinary Op MAD
r0.y = r1.x * 63.000000 + 0.500000;
// Unary Op FTOU
r0.y = uintBitsToFloat(uint(r0.y));
// Quatenary Op BFI
bitmask.x = (((1 << 6)-1) << 0) & 0xffffffff;
r0.y = ((floatBitsToUint(r0.y) << 0) & bitmask.x) | (128 & ~bitmask.x);
// Unary Op UTOF
r6.z = float(floatBitsToUint(r0.y));
// Unary Op MOV
r6.w = 172.000000;
// Binary Op DIV mask 0x6:
r0.yz = r6.zw / vec2(255.000000, 255.000000);
// Trinary Op MAD
r6.z = r1.z * 31.000000 + 0.500000;
// Unary Op FTOU
r6.z = uintBitsToFloat(uint(r6.z));
// Trinary Op MAD
r0.w = r0.w * 7.000000 + 0.500000;
// Unary Op FTOU
r0.w = uintBitsToFloat(uint(r0.w));
// Quatenary Op BFI
bitmask.x = (((1 << 3)-1) << 5) & 0xffffffff;
r0.w = ((floatBitsToUint(r0.w) << 5) & bitmask.x) | (0 & ~bitmask.x);
// Binary Op IADD mask 0x8:
r0.w = intBitsToFloat(floatBitsToInt(r0.w) + floatBitsToInt(r6.z));
// Unary Op UTOF
r0.w = float(floatBitsToUint(r0.w));
// Binary Op MUL mask 0x8:
r0.w = r0.w * 0.003922;
// Unary Op SQRT
oSV_TARGET0.xyz = sqrt(r12.yzw);
// Trinary Op MAD
oSV_TARGET1.xyz = r11.xyz * vec3(0.500000, 0.500000, 0.500000) + vec3(0.500000, 0.500000, 0.500000);
// Trinary Op MAD
r1.x = r10.x * -0.500000 + 0.500000;
// Trinary Op MAD
oSV_TARGET2.x = clamp(r10.y * 0.500000 + r1.x, 0.0, 1.0);
// Trinary Op MAD
oSV_TARGET2.y = r12.x * 0.500000 + 0.500000;
// Binary Op ADD mask 0x1:
r1.x = -r14.x + r1.w;
// Trinary Op MAD
oSV_TARGET3.y = r0.x * r1.x + r14.x;
// Unary Op MOV
oSV_TARGET0.w = r0.w;
// Unary Op MOV
oSV_TARGET1.w = 0.666667;
// Unary Op MOV
oSV_TARGET2.zw = r6.xy;
// Unary Op MOV
oSV_TARGET3.x = r10.z;
// Unary Op MOV
oSV_TARGET3.zw = r0.yz;
// No Param Op RET
return;
}
