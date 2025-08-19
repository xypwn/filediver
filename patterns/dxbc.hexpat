#include <std/ptr.pat>

fn relative_to_parents_parents_parents_parents_parent(u128 offset) {
    return addressof(parent.parent.parent.parent.parent.parent);
};

fn relative_to_parents_parents_parents_parent(u128 offset) {
    return addressof(parent.parent.parent.parent.parent);
};

fn relative_to_parents_parents_parent(u128 offset) {
    return addressof(parent.parent.parent.parent);
};

fn relative_to_parents_parent(u128 offset) {
    return addressof(parent.parent.parent);
};

fn relative_to_parent(u128 offset) {
    return addressof(parent.parent);
};

enum D3D_SHADER_VARIABLE_CLASS : u16 {
    D3D_SVC_SCALAR = 0,
    D3D_SVC_VECTOR,
    D3D_SVC_MATRIX_ROWS,
    D3D_SVC_MATRIX_COLUMNS,
    D3D_SVC_OBJECT,
    D3D_SVC_STRUCT,
    D3D_SVC_INTERFACE_CLASS,
    D3D_SVC_INTERFACE_POINTER,
    D3D10_SVC_SCALAR,
    D3D10_SVC_VECTOR,
    D3D10_SVC_MATRIX_ROWS,
    D3D10_SVC_MATRIX_COLUMNS,
    D3D10_SVC_OBJECT,
    D3D10_SVC_STRUCT,
    D3D11_SVC_INTERFACE_CLASS,
    D3D11_SVC_INTERFACE_POINTER,
};

enum D3D_SHADER_VARIABLE_TYPE : u16 {
    D3D_SVT_VOID = 0,
    D3D_SVT_BOOL = 1,
    D3D_SVT_INT = 2,
    D3D_SVT_FLOAT = 3,
    D3D_SVT_STRING = 4,
    D3D_SVT_TEXTURE = 5,
    D3D_SVT_TEXTURE1D = 6,
    D3D_SVT_TEXTURE2D = 7,
    D3D_SVT_TEXTURE3D = 8,
    D3D_SVT_TEXTURECUBE = 9,
    D3D_SVT_SAMPLER = 10,
    D3D_SVT_SAMPLER1D = 11,
    D3D_SVT_SAMPLER2D = 12,
    D3D_SVT_SAMPLER3D = 13,
    D3D_SVT_SAMPLERCUBE = 14,
    D3D_SVT_PIXELSHADER = 15,
    D3D_SVT_VERTEXSHADER = 16,
    D3D_SVT_PIXELFRAGMENT = 17,
    D3D_SVT_VERTEXFRAGMENT = 18,
    D3D_SVT_UINT = 19,
    D3D_SVT_UINT8 = 20,
    D3D_SVT_GEOMETRYSHADER = 21,
    D3D_SVT_RASTERIZER = 22,
    D3D_SVT_DEPTHSTENCIL = 23,
    D3D_SVT_BLEND = 24,
    D3D_SVT_BUFFER = 25,
    D3D_SVT_CBUFFER = 26,
    D3D_SVT_TBUFFER = 27,
    D3D_SVT_TEXTURE1DARRAY = 28,
    D3D_SVT_TEXTURE2DARRAY = 29,
    D3D_SVT_RENDERTARGETVIEW = 30,
    D3D_SVT_DEPTHSTENCILVIEW = 31,
    D3D_SVT_TEXTURE2DMS = 32,
    D3D_SVT_TEXTURE2DMSARRAY = 33,
    D3D_SVT_TEXTURECUBEARRAY = 34,
    D3D_SVT_HULLSHADER = 35,
    D3D_SVT_DOMAINSHADER = 36,
    D3D_SVT_INTERFACE_POINTER = 37,
    D3D_SVT_COMPUTESHADER = 38,
    D3D_SVT_DOUBLE = 39,
    D3D_SVT_RWTEXTURE1D = 40,
    D3D_SVT_RWTEXTURE1DARRAY = 41,
    D3D_SVT_RWTEXTURE2D = 42,
    D3D_SVT_RWTEXTURE2DARRAY = 43,
    D3D_SVT_RWTEXTURE3D = 44,
    D3D_SVT_RWBUFFER = 45,
    D3D_SVT_BYTEADDRESS_BUFFER = 46,
    D3D_SVT_RWBYTEADDRESS_BUFFER = 47,
    D3D_SVT_STRUCTURED_BUFFER = 48,
    D3D_SVT_RWSTRUCTURED_BUFFER = 49,
    D3D_SVT_APPEND_STRUCTURED_BUFFER = 50,
    D3D_SVT_CONSUME_STRUCTURED_BUFFER = 51,
    D3D_SVT_MIN8FLOAT = 52,
    D3D_SVT_MIN10FLOAT = 53,
    D3D_SVT_MIN16FLOAT = 54,
    D3D_SVT_MIN12INT = 55,
    D3D_SVT_MIN16INT = 56,
    D3D_SVT_MIN16UINT = 57,
    D3D_SVT_INT16,
    D3D_SVT_UINT16,
    D3D_SVT_FLOAT16,
    D3D_SVT_INT64,
    D3D_SVT_UINT64,
    D3D10_SVT_VOID,
    D3D10_SVT_BOOL,
    D3D10_SVT_INT,
    D3D10_SVT_FLOAT,
    D3D10_SVT_STRING,
    D3D10_SVT_TEXTURE,
    D3D10_SVT_TEXTURE1D,
    D3D10_SVT_TEXTURE2D,
    D3D10_SVT_TEXTURE3D,
    D3D10_SVT_TEXTURECUBE,
    D3D10_SVT_SAMPLER,
    D3D10_SVT_SAMPLER1D,
    D3D10_SVT_SAMPLER2D,
    D3D10_SVT_SAMPLER3D,
    D3D10_SVT_SAMPLERCUBE,
    D3D10_SVT_PIXELSHADER,
    D3D10_SVT_VERTEXSHADER,
    D3D10_SVT_PIXELFRAGMENT,
    D3D10_SVT_VERTEXFRAGMENT,
    D3D10_SVT_UINT,
    D3D10_SVT_UINT8,
    D3D10_SVT_GEOMETRYSHADER,
    D3D10_SVT_RASTERIZER,
    D3D10_SVT_DEPTHSTENCIL,
    D3D10_SVT_BLEND,
    D3D10_SVT_BUFFER,
    D3D10_SVT_CBUFFER,
    D3D10_SVT_TBUFFER,
    D3D10_SVT_TEXTURE1DARRAY,
    D3D10_SVT_TEXTURE2DARRAY,
    D3D10_SVT_RENDERTARGETVIEW,
    D3D10_SVT_DEPTHSTENCILVIEW,
    D3D10_SVT_TEXTURE2DMS,
    D3D10_SVT_TEXTURE2DMSARRAY,
    D3D10_SVT_TEXTURECUBEARRAY,
    D3D11_SVT_HULLSHADER,
    D3D11_SVT_DOMAINSHADER,
    D3D11_SVT_INTERFACE_POINTER,
    D3D11_SVT_COMPUTESHADER,
    D3D11_SVT_DOUBLE,
    D3D11_SVT_RWTEXTURE1D,
    D3D11_SVT_RWTEXTURE1DARRAY,
    D3D11_SVT_RWTEXTURE2D,
    D3D11_SVT_RWTEXTURE2DARRAY,
    D3D11_SVT_RWTEXTURE3D,
    D3D11_SVT_RWBUFFER,
    D3D11_SVT_BYTEADDRESS_BUFFER,
    D3D11_SVT_RWBYTEADDRESS_BUFFER,
    D3D11_SVT_STRUCTURED_BUFFER,
    D3D11_SVT_RWSTRUCTURED_BUFFER,
    D3D11_SVT_APPEND_STRUCTURED_BUFFER,
    D3D11_SVT_CONSUME_STRUCTURED_BUFFER,
};

struct Header {
    char magic[4];
    u8 digest[16];
    u16 major;
    u16 minor;
    u32 size;
    u32 parts;
};

struct string {
    char data[];
};

struct Type {
    D3D_SHADER_VARIABLE_CLASS var_class;
    D3D_SHADER_VARIABLE_TYPE var_type;
    u16 rows;
    u16 cols;
    u16 elements;
    u16 members;
    u16 member_offset;
    u8 unk00[18];
    string *name : u32 [[pointer_base("relative_to_parents_parents_parents_parents_parent")]];
};

struct Variable {
    string *name : u32 [[pointer_base("relative_to_parents_parents_parents_parent")]];
    u32 cb_offset;
    u32 size;
    u32 flags;
    Type *type : u32 [[pointer_base("relative_to_parents_parents_parents_parent")]];
    u32 default_value_offset;
    u32 unk[4];
};

struct ConstantBuffer {
    string *name : u32 [[pointer_base("relative_to_parents_parents_parent")]];
    u32 variable_count;
    Variable *variables[variable_count] : u32 [[pointer_base("relative_to_parents_parents_parent")]];
    u32 size;
    u32 flags;
    u32 type;
};

struct ResourceBinding {
    string *name : u32 [[pointer_base("relative_to_parents_parents_parent")]];
    u32 input_type;
    u32 return_type;
    u32 view_dimension;
    u32 sample_count;
    u32 bind_point;
    u32 bind_count;
    u32 flags;
};

struct List<T, auto count> {
    T data[count] [[inline]];
};

struct RDEF {
    u32 cbuf_count;
    List<ConstantBuffer, cbuf_count> *cbuf : u32 [[pointer_base("relative_to_parent")]];
    u32 rbind_count;
    List<ResourceBinding, rbind_count> *rbind : u32 [[pointer_base("relative_to_parent")]];
    u8 minor;
    u8 major;
    u16 program_type;
    u32 flags;
    string *creator : u32 [[pointer_base("relative_to_parent")]];
};

struct Element {
    u32 unk0;
    string *name : u32 [[pointer_base("relative_to_parents_parents_parent")]];
    u32 semantic_index;
    u32 system_value_type;
    u32 component_type;
    u32 register;
    u8 mask;
    u8 rw_mask;
    padding[2];
    u32 unk1;
};

struct ISGN {
    u32 element_count;
    List<Element, element_count> *elements : u32 [[pointer_base("relative_to_parent")]];
};

struct ISG1 {
    u32 element_count;
    List<Element, element_count> *elements : u32 [[pointer_base("relative_to_parent")]];
};

struct OSG1 {
    u32 element_count;
    List<Element, element_count> *elements : u32 [[pointer_base("relative_to_parent")]];
};

enum D3D10_SB_TOKENIZED_PROGRAM_TYPE : u16 {
    D3D10_SB_PIXEL_SHADER       = 0,
    D3D10_SB_VERTEX_SHADER      = 1,
    D3D10_SB_GEOMETRY_SHADER    = 2,
    
    // D3D11 Shaders
    D3D11_SB_HULL_SHADER        = 3,
    D3D11_SB_DOMAIN_SHADER      = 4,
    D3D11_SB_COMPUTE_SHADER     = 5,

    D3D11_SB_RESERVED0          = 0xFFF0
};

bitfield Version {
    minor : 4;
    major : 4;
};

enum D3D10_SB_OPCODE_TYPE : u16 {
    D3D10_SB_OPCODE_ADD          ,
    D3D10_SB_OPCODE_AND          ,
    D3D10_SB_OPCODE_BREAK        ,
    D3D10_SB_OPCODE_BREAKC       ,
    D3D10_SB_OPCODE_CALL         ,
    D3D10_SB_OPCODE_CALLC        ,
    D3D10_SB_OPCODE_CASE         ,
    D3D10_SB_OPCODE_CONTINUE     ,
    D3D10_SB_OPCODE_CONTINUEC    ,
    D3D10_SB_OPCODE_CUT          ,
    D3D10_SB_OPCODE_DEFAULT      ,
    D3D10_SB_OPCODE_DERIV_RTX    ,
    D3D10_SB_OPCODE_DERIV_RTY    ,
    D3D10_SB_OPCODE_DISCARD      ,
    D3D10_SB_OPCODE_DIV          ,
    D3D10_SB_OPCODE_DP2          ,
    D3D10_SB_OPCODE_DP3          ,
    D3D10_SB_OPCODE_DP4          ,
    D3D10_SB_OPCODE_ELSE         ,
    D3D10_SB_OPCODE_EMIT         ,
    D3D10_SB_OPCODE_EMITTHENCUT  ,
    D3D10_SB_OPCODE_ENDIF        ,
    D3D10_SB_OPCODE_ENDLOOP      ,
    D3D10_SB_OPCODE_ENDSWITCH    ,
    D3D10_SB_OPCODE_EQ           ,
    D3D10_SB_OPCODE_EXP          ,
    D3D10_SB_OPCODE_FRC          ,
    D3D10_SB_OPCODE_FTOI         ,
    D3D10_SB_OPCODE_FTOU         ,
    D3D10_SB_OPCODE_GE           ,
    D3D10_SB_OPCODE_IADD         ,
    D3D10_SB_OPCODE_IF           ,
    D3D10_SB_OPCODE_IEQ          ,
    D3D10_SB_OPCODE_IGE          ,
    D3D10_SB_OPCODE_ILT          ,
    D3D10_SB_OPCODE_IMAD         ,
    D3D10_SB_OPCODE_IMAX         ,
    D3D10_SB_OPCODE_IMIN         ,
    D3D10_SB_OPCODE_IMUL         ,
    D3D10_SB_OPCODE_INE          ,
    D3D10_SB_OPCODE_INEG         ,
    D3D10_SB_OPCODE_ISHL         ,
    D3D10_SB_OPCODE_ISHR         ,
    D3D10_SB_OPCODE_ITOF         ,
    D3D10_SB_OPCODE_LABEL        ,
    D3D10_SB_OPCODE_LD           ,
    D3D10_SB_OPCODE_LD_MS        ,
    D3D10_SB_OPCODE_LOG          ,
    D3D10_SB_OPCODE_LOOP         ,
    D3D10_SB_OPCODE_LT           ,
    D3D10_SB_OPCODE_MAD          ,
    D3D10_SB_OPCODE_MIN          ,
    D3D10_SB_OPCODE_MAX          ,
    D3D10_SB_OPCODE_CUSTOMDATA   ,
    D3D10_SB_OPCODE_MOV          ,
    D3D10_SB_OPCODE_MOVC         ,
    D3D10_SB_OPCODE_MUL          ,
    D3D10_SB_OPCODE_NE           ,
    D3D10_SB_OPCODE_NOP          ,
    D3D10_SB_OPCODE_NOT          ,
    D3D10_SB_OPCODE_OR           ,
    D3D10_SB_OPCODE_RESINFO      ,
    D3D10_SB_OPCODE_RET          ,
    D3D10_SB_OPCODE_RETC         ,
    D3D10_SB_OPCODE_ROUND_NE     ,
    D3D10_SB_OPCODE_ROUND_NI     ,
    D3D10_SB_OPCODE_ROUND_PI     ,
    D3D10_SB_OPCODE_ROUND_Z      ,
    D3D10_SB_OPCODE_RSQ          ,
    D3D10_SB_OPCODE_SAMPLE       ,
    D3D10_SB_OPCODE_SAMPLE_C     ,
    D3D10_SB_OPCODE_SAMPLE_C_LZ  ,
    D3D10_SB_OPCODE_SAMPLE_L     ,
    D3D10_SB_OPCODE_SAMPLE_D     ,
    D3D10_SB_OPCODE_SAMPLE_B     ,
    D3D10_SB_OPCODE_SQRT         ,
    D3D10_SB_OPCODE_SWITCH       ,
    D3D10_SB_OPCODE_SINCOS       ,
    D3D10_SB_OPCODE_UDIV         ,
    D3D10_SB_OPCODE_ULT          ,
    D3D10_SB_OPCODE_UGE          ,
    D3D10_SB_OPCODE_UMUL         ,
    D3D10_SB_OPCODE_UMAD         ,
    D3D10_SB_OPCODE_UMAX         ,
    D3D10_SB_OPCODE_UMIN         ,
    D3D10_SB_OPCODE_USHR         ,
    D3D10_SB_OPCODE_UTOF         ,
    D3D10_SB_OPCODE_XOR          ,
    D3D10_SB_OPCODE_DCL_RESOURCE                     , // DCL* opcodes have
    D3D10_SB_OPCODE_DCL_CONSTANT_BUFFER              , // custom operand formats.
    D3D10_SB_OPCODE_DCL_SAMPLER                      ,
    D3D10_SB_OPCODE_DCL_INDEX_RANGE                  ,
    D3D10_SB_OPCODE_DCL_GS_OUTPUT_PRIMITIVE_TOPOLOGY ,
    D3D10_SB_OPCODE_DCL_GS_INPUT_PRIMITIVE           ,
    D3D10_SB_OPCODE_DCL_MAX_OUTPUT_VERTEX_COUNT      ,
    D3D10_SB_OPCODE_DCL_INPUT                        ,
    D3D10_SB_OPCODE_DCL_INPUT_SGV                    ,
    D3D10_SB_OPCODE_DCL_INPUT_SIV                    ,
    D3D10_SB_OPCODE_DCL_INPUT_PS                     ,
    D3D10_SB_OPCODE_DCL_INPUT_PS_SGV                 ,
    D3D10_SB_OPCODE_DCL_INPUT_PS_SIV                 ,
    D3D10_SB_OPCODE_DCL_OUTPUT                       ,
    D3D10_SB_OPCODE_DCL_OUTPUT_SGV                   ,
    D3D10_SB_OPCODE_DCL_OUTPUT_SIV                   ,
    D3D10_SB_OPCODE_DCL_TEMPS                        ,
    D3D10_SB_OPCODE_DCL_INDEXABLE_TEMP               ,
    D3D10_SB_OPCODE_DCL_GLOBAL_FLAGS                 ,

// -----------------------------------------------

    // This marks the end of D3D10.0 opcodes
    D3D10_SB_OPCODE_RESERVED0,
    
// ---------- DX 10.1 op codes---------------------

    D3D10_1_SB_OPCODE_LOD,
    D3D10_1_SB_OPCODE_GATHER4,
    D3D10_1_SB_OPCODE_SAMPLE_POS,
    D3D10_1_SB_OPCODE_SAMPLE_INFO,

// -----------------------------------------------

    // This marks the end of D3D10.1 opcodes
    D3D10_1_SB_OPCODE_RESERVED1,

// ---------- DX 11 op codes---------------------
    D3D11_SB_OPCODE_HS_DECLS                         , // token marks beginning of HS sub-shader
    D3D11_SB_OPCODE_HS_CONTROL_POINT_PHASE           , // token marks beginning of HS sub-shader
    D3D11_SB_OPCODE_HS_FORK_PHASE                    , // token marks beginning of HS sub-shader
    D3D11_SB_OPCODE_HS_JOIN_PHASE                    , // token marks beginning of HS sub-shader

    D3D11_SB_OPCODE_EMIT_STREAM                      ,
    D3D11_SB_OPCODE_CUT_STREAM                       ,
    D3D11_SB_OPCODE_EMITTHENCUT_STREAM               ,
    D3D11_SB_OPCODE_INTERFACE_CALL                   ,

    D3D11_SB_OPCODE_BUFINFO                          ,
    D3D11_SB_OPCODE_DERIV_RTX_COARSE                 ,
    D3D11_SB_OPCODE_DERIV_RTX_FINE                   ,
    D3D11_SB_OPCODE_DERIV_RTY_COARSE                 ,
    D3D11_SB_OPCODE_DERIV_RTY_FINE                   ,
    D3D11_SB_OPCODE_GATHER4_C                        ,
    D3D11_SB_OPCODE_GATHER4_PO                       ,
    D3D11_SB_OPCODE_GATHER4_PO_C                     ,
    D3D11_SB_OPCODE_RCP                              ,
    D3D11_SB_OPCODE_F32TOF16                         ,
    D3D11_SB_OPCODE_F16TOF32                         ,
    D3D11_SB_OPCODE_UADDC                            ,
    D3D11_SB_OPCODE_USUBB                            ,
    D3D11_SB_OPCODE_COUNTBITS                        ,
    D3D11_SB_OPCODE_FIRSTBIT_HI                      ,
    D3D11_SB_OPCODE_FIRSTBIT_LO                      ,
    D3D11_SB_OPCODE_FIRSTBIT_SHI                     ,
    D3D11_SB_OPCODE_UBFE                             ,
    D3D11_SB_OPCODE_IBFE                             ,
    D3D11_SB_OPCODE_BFI                              ,
    D3D11_SB_OPCODE_BFREV                            ,
    D3D11_SB_OPCODE_SWAPC                            ,

    D3D11_SB_OPCODE_DCL_STREAM                       ,
    D3D11_SB_OPCODE_DCL_FUNCTION_BODY                ,
    D3D11_SB_OPCODE_DCL_FUNCTION_TABLE               ,
    D3D11_SB_OPCODE_DCL_INTERFACE                    ,
    
    D3D11_SB_OPCODE_DCL_INPUT_CONTROL_POINT_COUNT    ,
    D3D11_SB_OPCODE_DCL_OUTPUT_CONTROL_POINT_COUNT   ,
    D3D11_SB_OPCODE_DCL_TESS_DOMAIN                  ,
    D3D11_SB_OPCODE_DCL_TESS_PARTITIONING            ,
    D3D11_SB_OPCODE_DCL_TESS_OUTPUT_PRIMITIVE        ,
    D3D11_SB_OPCODE_DCL_HS_MAX_TESSFACTOR            ,
    D3D11_SB_OPCODE_DCL_HS_FORK_PHASE_INSTANCE_COUNT ,
    D3D11_SB_OPCODE_DCL_HS_JOIN_PHASE_INSTANCE_COUNT ,

    D3D11_SB_OPCODE_DCL_THREAD_GROUP                 ,
    D3D11_SB_OPCODE_DCL_UNORDERED_ACCESS_VIEW_TYPED  ,
    D3D11_SB_OPCODE_DCL_UNORDERED_ACCESS_VIEW_RAW    ,
    D3D11_SB_OPCODE_DCL_UNORDERED_ACCESS_VIEW_STRUCTURED,
    D3D11_SB_OPCODE_DCL_THREAD_GROUP_SHARED_MEMORY_RAW,
    D3D11_SB_OPCODE_DCL_THREAD_GROUP_SHARED_MEMORY_STRUCTURED,
    D3D11_SB_OPCODE_DCL_RESOURCE_RAW                 ,
    D3D11_SB_OPCODE_DCL_RESOURCE_STRUCTURED          ,
    D3D11_SB_OPCODE_LD_UAV_TYPED                     ,
    D3D11_SB_OPCODE_STORE_UAV_TYPED                  ,
    D3D11_SB_OPCODE_LD_RAW                           ,
    D3D11_SB_OPCODE_STORE_RAW                        ,
    D3D11_SB_OPCODE_LD_STRUCTURED                    ,
    D3D11_SB_OPCODE_STORE_STRUCTURED                 ,
    D3D11_SB_OPCODE_ATOMIC_AND                       ,
    D3D11_SB_OPCODE_ATOMIC_OR                        ,
    D3D11_SB_OPCODE_ATOMIC_XOR                       ,
    D3D11_SB_OPCODE_ATOMIC_CMP_STORE                 ,
    D3D11_SB_OPCODE_ATOMIC_IADD                      ,
    D3D11_SB_OPCODE_ATOMIC_IMAX                      ,
    D3D11_SB_OPCODE_ATOMIC_IMIN                      ,
    D3D11_SB_OPCODE_ATOMIC_UMAX                      ,
    D3D11_SB_OPCODE_ATOMIC_UMIN                      ,
    D3D11_SB_OPCODE_IMM_ATOMIC_ALLOC                 ,
    D3D11_SB_OPCODE_IMM_ATOMIC_CONSUME               ,
    D3D11_SB_OPCODE_IMM_ATOMIC_IADD                  ,
    D3D11_SB_OPCODE_IMM_ATOMIC_AND                   ,
    D3D11_SB_OPCODE_IMM_ATOMIC_OR                    ,
    D3D11_SB_OPCODE_IMM_ATOMIC_XOR                   ,
    D3D11_SB_OPCODE_IMM_ATOMIC_EXCH                  ,
    D3D11_SB_OPCODE_IMM_ATOMIC_CMP_EXCH              ,
    D3D11_SB_OPCODE_IMM_ATOMIC_IMAX                  ,
    D3D11_SB_OPCODE_IMM_ATOMIC_IMIN                  ,
    D3D11_SB_OPCODE_IMM_ATOMIC_UMAX                  ,
    D3D11_SB_OPCODE_IMM_ATOMIC_UMIN                  ,   
    D3D11_SB_OPCODE_SYNC                             ,
    
    D3D11_SB_OPCODE_DADD                             ,
    D3D11_SB_OPCODE_DMAX                             ,
    D3D11_SB_OPCODE_DMIN                             ,
    D3D11_SB_OPCODE_DMUL                             ,
    D3D11_SB_OPCODE_DEQ                              ,
    D3D11_SB_OPCODE_DGE                              ,
    D3D11_SB_OPCODE_DLT                              ,
    D3D11_SB_OPCODE_DNE                              ,
    D3D11_SB_OPCODE_DMOV                             ,
    D3D11_SB_OPCODE_DMOVC                            ,
    D3D11_SB_OPCODE_DTOF                             ,
    D3D11_SB_OPCODE_FTOD                             ,

    D3D11_SB_OPCODE_EVAL_SNAPPED                     ,
    D3D11_SB_OPCODE_EVAL_SAMPLE_INDEX                ,
    D3D11_SB_OPCODE_EVAL_CENTROID                    ,
    
    D3D11_SB_OPCODE_DCL_GS_INSTANCE_COUNT            ,

    D3D11_SB_OPCODE_ABORT                            ,
    D3D11_SB_OPCODE_DEBUG_BREAK                      ,

// -----------------------------------------------

    // This marks the end of D3D11.0 opcodes
    D3D11_SB_OPCODE_RESERVED0,

    D3D11_1_SB_OPCODE_DDIV,
    D3D11_1_SB_OPCODE_DFMA,
    D3D11_1_SB_OPCODE_DRCP,

    D3D11_1_SB_OPCODE_MSAD,

    D3D11_1_SB_OPCODE_DTOI,
    D3D11_1_SB_OPCODE_DTOU,
    D3D11_1_SB_OPCODE_ITOD,
    D3D11_1_SB_OPCODE_UTOD,

// -----------------------------------------------

    // This marks the end of D3D11.1 opcodes
    D3D11_1_SB_OPCODE_RESERVED0,

    D3DWDDM1_3_SB_OPCODE_GATHER4_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_GATHER4_C_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_GATHER4_PO_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_GATHER4_PO_C_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_LD_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_LD_MS_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_LD_UAV_TYPED_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_LD_RAW_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_LD_STRUCTURED_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_SAMPLE_L_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_SAMPLE_C_LZ_FEEDBACK,

    D3DWDDM1_3_SB_OPCODE_SAMPLE_CLAMP_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_SAMPLE_B_CLAMP_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_SAMPLE_D_CLAMP_FEEDBACK,
    D3DWDDM1_3_SB_OPCODE_SAMPLE_C_CLAMP_FEEDBACK,

    D3DWDDM1_3_SB_OPCODE_CHECK_ACCESS_FULLY_MAPPED,

// -----------------------------------------------

    // This marks the end of WDDM 1.3 opcodes
    D3DWDDM1_3_SB_OPCODE_RESERVED0,

    D3D10_SB_NUM_OPCODES                                     // Should be the last entry
};

enum D3D10_SB_RESINFO_INSTRUCTION_RETURN_TYPE : u8 {
    FLOAT      = 0,
    RCPFLOAT   = 1,
    UINT       = 2,
};

enum D3D10_SB_CUSTOMDATA_CLASS : u32 {
    D3D10_SB_CUSTOMDATA_COMMENT = 0,
    D3D10_SB_CUSTOMDATA_DEBUGINFO,
    D3D10_SB_CUSTOMDATA_OPAQUE,
    D3D10_SB_CUSTOMDATA_DCL_IMMEDIATE_CONSTANT_BUFFER,
    D3D11_SB_CUSTOMDATA_SHADER_MESSAGE,
    D3D11_SB_CUSTOMDATA_SHADER_CLIP_PLANE_CONSTANT_MAPPINGS_FOR_DX9,
};

struct Opcode {
    u32 opcode [[hidden]];
    D3D10_SB_OPCODE_TYPE type = opcode & 0x07ff [[export]];
    if (type == D3D10_SB_OPCODE_TYPE::D3D10_SB_OPCODE_CUSTOMDATA) {
        D3D10_SB_CUSTOMDATA_CLASS class = (opcode & 0xfffff800) >> 11 [[export]];
        u32 size;
        u32 data[size-2];
    } else {
        u8 length = (opcode & 0x7f000000) >> 24 [[export]];
        bool saturate = (opcode & 0x00002000) != 0 [[export]];
        u8 test_boolean = (opcode & 0x00040000) >> 18 [[export]];
        u8 precise_mask = (opcode & 0x00780000) >> 19 [[export]];
        D3D10_SB_RESINFO_INSTRUCTION_RETURN_TYPE return_type = (opcode & 0x00001800) >> 11 [[export]];
        bool sync_threads = (opcode & 0x00000800) != 0 [[export]];
        bool sync_thread_group_shm = (opcode & 0x00001000) != 0 [[export]];
        bool sync_unordered_access_memory_group = (opcode & 0x00002000) != 0 [[export]];
        bool sync_unordered_access_memory_global = (opcode & 0x00004000) != 0 [[export]];
        bool extended = (opcode & 0x80000000) != 0 [[export]];
        if (length > 1)
            u32 data[length-1];
    }
};

struct SHEX {
    u64 begin = $;
    Version version;
    u8 reserved;
    D3D10_SB_TOKENIZED_PROGRAM_TYPE program_type;
    u32 num_dwords;
    Opcode opcodes[while($ < begin+(num_dwords-2)*4)];
};

struct PartHeader {
    char name[4];
    u32 size;
};

struct Part {
    PartHeader header;
    if(header.name == "RDEF") {
        RDEF data;
        padding[header.size - sizeof(data)];
    } else if(header.name == "ISGN") {
        ISGN data;
        padding[header.size - sizeof(data)];
    } else if(header.name == "ISG1") {
        ISG1 data;
        padding[header.size - sizeof(data)];
    } else if(header.name == "OSG1") {
        OSG1 data;
        padding[header.size - sizeof(data)];
    } else if(header.name == "SHEX") {
        SHEX data;
        padding[header.size - sizeof(data)];
    } else {
        u8 data[header.size];
    }
};

struct Pointer<T> {
    T *data : u32 [[pointer_base("relative_to_parents_parent")]];
};

struct DXBC {
    Header header;
    Pointer<Part> parts[header.parts];
};

DXBC dxbc @0x9ed90;