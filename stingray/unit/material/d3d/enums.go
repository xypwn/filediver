package d3d

type ShaderInstructionReturnType uint8

const (
	RET_FLOAT ShaderInstructionReturnType = iota
	RET_RCPFLOAT
	RET_UINT
)

type ShaderFlags uint32

const (
	SF_NONE                           ShaderFlags = 0x00000000
	SF_DEBUG                          ShaderFlags = 0x00000001
	SF_SKIP_VALIDATION                ShaderFlags = 0x00000002
	SF_SKIP_OPTIMIZATION              ShaderFlags = 0x00000004
	SF_PACK_MATRIX_ROW_MAJOR          ShaderFlags = 0x00000008
	SF_PACK_MATRIX_COLUMN_MAJOR       ShaderFlags = 0x00000010
	SF_PARTIAL_PRECISION              ShaderFlags = 0x00000020
	SF_FORCE_VS_SOFTWARE_NOOPT        ShaderFlags = 0x00000040
	SF_FORCE_PS_SOFTWARE_NOOPT        ShaderFlags = 0x00000080
	SF_NO_PRESHADER                   ShaderFlags = 0x00000100
	SF_AVOID_FLOW_CONTROL             ShaderFlags = 0x00000200
	SF_PREFER_FLOW_CONTROL            ShaderFlags = 0x00000400
	SF_ENABLE_STRICTNESS              ShaderFlags = 0x00000800
	SF_ENABLE_BACKWARDS_COMPATIBILITY ShaderFlags = 0x00001000
	SF_IEEE_STRICTNESS                ShaderFlags = 0x00002000
	SF_OPTIMIZATION_LEVEL0            ShaderFlags = 0x00004000
	SF_OPTIMIZATION_LEVEL1            ShaderFlags = 0
	SF_OPTIMIZATION_LEVEL2            ShaderFlags = 0x0000C000
	SF_OPTIMIZATION_LEVEL3            ShaderFlags = 0x00008000
	SF_RESERVED16                     ShaderFlags = 0x00010000
	SF_RESERVED17                     ShaderFlags = 0x00020000
	SF_WARNINGS_ARE_ERRORS            ShaderFlags = 0x00040000
)
