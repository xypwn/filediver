package d3dops

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets/previews/shaders"
	"github.com/xypwn/filediver/stingray/unit/material/glsl"
)

type ShaderVariableClass uint16

const (
	SVC_SCALAR ShaderVariableClass = iota
	SVC_VECTOR
	SVC_MATRIX_ROWS
	SVC_MATRIX_COLUMNS
	SVC_OBJECT
	SVC_STRUCT
	SVC_INTERFACE_CLASS
	SVC_INTERFACE_POINTER
	SVC_V10_SCALAR
	SVC_V10_VECTOR
	SVC_V10_MATRIX_ROWS
	SVC_V10_MATRIX_COLUMNS
	SVC_V10_OBJECT
	SVC_V10_STRUCT
	SVC_V11_INTERFACE_CLASS
	SVC_V11_INTERFACE_POINTER
)

func (svc ShaderVariableClass) GLSLClass() string {
	switch svc {
	case SVC_VECTOR, SVC_V10_VECTOR:
		return "vec"
	case SVC_MATRIX_ROWS, SVC_MATRIX_COLUMNS, SVC_V10_MATRIX_ROWS, SVC_V10_MATRIX_COLUMNS:
		return "mat"
	case SVC_STRUCT, SVC_V10_STRUCT, SVC_OBJECT, SVC_V10_OBJECT:
		return "struct"
	default:
		return ""
	}
}

type ShaderVariableType uint16

const (
	SVT_VOID ShaderVariableType = iota
	SVT_BOOL
	SVT_INT
	SVT_FLOAT
	SVT_STRING
	SVT_TEXTURE
	SVT_TEXTURE1D
	SVT_TEXTURE2D
	SVT_TEXTURE3D
	SVT_TEXTURECUBE
	SVT_SAMPLER
	SVT_SAMPLER1D
	SVT_SAMPLER2D
	SVT_SAMPLER3D
	SVT_SAMPLERCUBE
	SVT_PIXELSHADER
	SVT_VERTEXSHADER
	SVT_PIXELFRAGMENT
	SVT_VERTEXFRAGMENT
	SVT_UINT
	SVT_UINT8
	SVT_GEOMETRYSHADER
	SVT_RASTERIZER
	SVT_DEPTHSTENCIL
	SVT_BLEND
	SVT_BUFFER
	SVT_CBUFFER
	SVT_TBUFFER
	SVT_TEXTURE1DARRAY
	SVT_TEXTURE2DARRAY
	SVT_RENDERTARGETVIEW
	SVT_DEPTHSTENCILVIEW
	SVT_TEXTURE2DMS
	SVT_TEXTURE2DMSARRAY
	SVT_TEXTURECUBEARRAY
	SVT_HULLSHADER
	SVT_DOMAINSHADER
	SVT_INTERFACE_POINTER
	SVT_COMPUTESHADER
	SVT_DOUBLE
	SVT_RWTEXTURE1D
	SVT_RWTEXTURE1DARRAY
	SVT_RWTEXTURE2D
	SVT_RWTEXTURE2DARRAY
	SVT_RWTEXTURE3D
	SVT_RWBUFFER
	SVT_BYTEADDRESS_BUFFER
	SVT_RWBYTEADDRESS_BUFFER
	SVT_STRUCTURED_BUFFER
	SVT_RWSTRUCTURED_BUFFER
	SVT_APPEND_STRUCTURED_BUFFER
	SVT_CONSUME_STRUCTURED_BUFFER
	SVT_MIN8FLOAT
	SVT_MIN10FLOAT
	SVT_MIN16FLOAT
	SVT_MIN12INT
	SVT_MIN16INT
	SVT_MIN16UINT
	SVT_INT16
	SVT_UINT16
	SVT_FLOAT16
	SVT_INT64
	SVT_UINT64
	SVT_V10_VOID
	SVT_V10_BOOL
	SVT_V10_INT
	SVT_V10_FLOAT
	SVT_V10_STRING
	SVT_V10_TEXTURE
	SVT_V10_TEXTURE1D
	SVT_V10_TEXTURE2D
	SVT_V10_TEXTURE3D
	SVT_V10_TEXTURECUBE
	SVT_V10_SAMPLER
	SVT_V10_SAMPLER1D
	SVT_V10_SAMPLER2D
	SVT_V10_SAMPLER3D
	SVT_V10_SAMPLERCUBE
	SVT_V10_PIXELSHADER
	SVT_V10_VERTEXSHADER
	SVT_V10_PIXELFRAGMENT
	SVT_V10_VERTEXFRAGMENT
	SVT_V10_UINT
	SVT_V10_UINT8
	SVT_V10_GEOMETRYSHADER
	SVT_V10_RASTERIZER
	SVT_V10_DEPTHSTENCIL
	SVT_V10_BLEND
	SVT_V10_BUFFER
	SVT_V10_CBUFFER
	SVT_V10_TBUFFER
	SVT_V10_TEXTURE1DARRAY
	SVT_V10_TEXTURE2DARRAY
	SVT_V10_RENDERTARGETVIEW
	SVT_V10_DEPTHSTENCILVIEW
	SVT_V10_TEXTURE2DMS
	SVT_V10_TEXTURE2DMSARRAY
	SVT_V10_TEXTURECUBEARRAY
	SVT_V11_HULLSHADER
	SVT_V11_DOMAINSHADER
	SVT_V11_INTERFACE_POINTER
	SVT_V11_COMPUTESHADER
	SVT_V11_DOUBLE
	SVT_V11_RWTEXTURE1D
	SVT_V11_RWTEXTURE1DARRAY
	SVT_V11_RWTEXTURE2D
	SVT_V11_RWTEXTURE2DARRAY
	SVT_V11_RWTEXTURE3D
	SVT_V11_RWBUFFER
	SVT_V11_BYTEADDRESS_BUFFER
	SVT_V11_RWBYTEADDRESS_BUFFER
	SVT_V11_STRUCTURED_BUFFER
	SVT_V11_RWSTRUCTURED_BUFFER
	SVT_V11_APPEND_STRUCTURED_BUFFER
	SVT_V11_CONSUME_STRUCTURED_BUFFER
)

func (svt ShaderVariableType) GLSLPrefix() string {
	switch svt {
	case SVT_BOOL, SVT_V10_BOOL:
		return "b"
	case SVT_INT, SVT_INT16, SVT_INT64, SVT_V10_INT, SVT_MIN12INT, SVT_MIN16INT:
		return "i"
	case SVT_UINT, SVT_UINT8, SVT_UINT16, SVT_UINT64, SVT_V10_UINT8, SVT_V10_UINT, SVT_MIN16UINT:
		return "u"
	case SVT_DOUBLE, SVT_V11_DOUBLE:
		return "d"
	default:
		return ""
	}
}

func (svt ShaderVariableType) GLSLType() string {
	switch svt {
	case SVT_BOOL, SVT_V10_BOOL:
		return "bool"
	case SVT_INT, SVT_INT16, SVT_INT64, SVT_V10_INT, SVT_MIN12INT, SVT_MIN16INT:
		return "int"
	case SVT_UINT, SVT_UINT8, SVT_UINT16, SVT_UINT64, SVT_V10_UINT8, SVT_V10_UINT, SVT_MIN16UINT:
		return "uint"
	case SVT_DOUBLE, SVT_V11_DOUBLE:
		return "double"
	case SVT_FLOAT, SVT_MIN8FLOAT, SVT_MIN10FLOAT, SVT_MIN16FLOAT, SVT_FLOAT16, SVT_V10_FLOAT:
		return "float"
	case SVT_SAMPLER, SVT_V10_SAMPLER:
		panic("Found a sampler example!")
	case SVT_SAMPLER1D, SVT_V10_SAMPLER1D:
		return "sampler1D"
	case SVT_SAMPLER2D, SVT_V10_SAMPLER2D:
		return "sampler2D"
	case SVT_SAMPLER3D, SVT_V10_SAMPLER3D:
		return "sampler3D"
	case SVT_SAMPLERCUBE, SVT_V10_SAMPLERCUBE:
		return "samplerCube"
	default:
		return fmt.Sprintf("Unimplemented GLSL type! Got variable type: %x", svt)
	}
}

func (svt ShaderVariableType) GLSLPrecision() string {
	switch svt {
	case SVT_UINT8, SVT_MIN8FLOAT, SVT_V10_UINT8:
		return "lowp "
	case SVT_MIN10FLOAT, SVT_MIN12INT, SVT_MIN16FLOAT, SVT_MIN16INT, SVT_MIN16UINT, SVT_INT16, SVT_UINT16, SVT_FLOAT16:
		return "mediump "
	case SVT_DOUBLE, SVT_INT64, SVT_UINT64, SVT_V11_DOUBLE:
		return "highp "
	default:
		return ""
	}
}

func (svt ShaderVariableType) GLSLSize() int {
	switch svt {
	case SVT_UINT8, SVT_MIN8FLOAT, SVT_V10_UINT8:
		return 1
	case SVT_MIN10FLOAT, SVT_MIN12INT, SVT_MIN16FLOAT, SVT_MIN16INT, SVT_MIN16UINT, SVT_INT16, SVT_UINT16, SVT_FLOAT16:
		return 2
	case SVT_DOUBLE, SVT_INT64, SVT_UINT64, SVT_V11_DOUBLE:
		return 8
	default:
		return 4
	}
}

type ConstantBufferFlags uint32

const (
	CBF_USERPACKED     ConstantBufferFlags = 0x1
	CBF_V10_USERPACKED ConstantBufferFlags = 0x2
)

func (f ConstantBufferFlags) ToString() string {
	flagList := make([]string, 0)
	if f&CBF_USERPACKED != 0 {
		flagList = append(flagList, "CBF_USERPACKED")
	}
	if f&CBF_V10_USERPACKED != 0 {
		flagList = append(flagList, "CBF_V10_USERPACKED")
	}
	return strings.Join(flagList, " | ")
}

type ConstantBufferType uint32

const (
	CT_CBUFFER ConstantBufferType = iota
	CT_TBUFFER
	CT_INTERFACE_POINTERS
	CT_RESOURCE_BIND_INFO
)

func (t ConstantBufferType) ToString() string {
	switch t {
	case CT_CBUFFER:
		return "CBUFFER"
	case CT_TBUFFER:
		return "TBUFFER"
	case CT_INTERFACE_POINTERS:
		return "INTERFACE_POINTERS"
	case CT_RESOURCE_BIND_INFO:
		return "RESOURCE_BIND_INFO"
	}
	return "unknown constant buffer type"
}

type VariableType struct {
	Class        ShaderVariableClass
	Type         ShaderVariableType
	Rows         uint16
	Cols         uint16
	Elements     uint16
	Members      uint16
	MemberOffset uint16
	Name         string
}

type Variable struct {
	Name         string
	BufferOffset uint32
	Size         uint32
	Flags        uint32 // D3D_SHADER_VARIABLE_FLAGS
	DefaultData  []byte
	VariableType
}

func (v Variable) ToGLSL() string {
	toReturn := ""
	switch v.Class {
	case SVC_SCALAR, SVC_V10_SCALAR:
		toReturn += fmt.Sprintf("%v%v", v.Type.GLSLPrecision(), v.Type.GLSLType())
	case SVC_VECTOR, SVC_V10_VECTOR:
		toReturn += fmt.Sprintf("%v%v%v%v", v.Type.GLSLPrecision(), v.Type.GLSLPrefix(), v.Class.GLSLClass(), v.Cols)
	case SVC_MATRIX_COLUMNS, SVC_MATRIX_ROWS, SVC_V10_MATRIX_COLUMNS, SVC_V10_MATRIX_ROWS:
		toReturn += fmt.Sprintf("%v%v%v", v.Type.GLSLPrecision(), v.Type.GLSLPrefix(), v.Class.GLSLClass())
		if v.Cols == v.Rows {
			toReturn += fmt.Sprintf("%v", v.Cols)
		} else {
			toReturn += fmt.Sprintf("%vx%v", v.Cols, v.Rows)
		}
	default:
		panic("Unimplemented variable class!")
	}

	toReturn += fmt.Sprintf(" %v", v.Name)
	if v.Elements > 0 {
		toReturn += fmt.Sprintf("[%v]", v.Elements)
	}
	toReturn += fmt.Sprintf("; // size: %v, offset: %v", v.Size, v.BufferOffset)
	if v.DefaultData != nil && len(v.DefaultData) > 0 {
		toReturn += fmt.Sprintf(", default data: %v", v.DefaultData)
	}
	switch v.Class {
	case SVC_MATRIX_ROWS, SVC_V10_MATRIX_ROWS:
		toReturn += ", row major"
	}
	return toReturn
}

func (v Variable) ToGLSLAlign() int {
	toReturn := 0
	switch v.Class {
	case SVC_SCALAR, SVC_V10_SCALAR:
		toReturn = v.Type.GLSLSize()
	case SVC_VECTOR, SVC_V10_VECTOR:
		cols := int(v.Cols)
		if cols == 3 {
			cols = 4
		}
		toReturn = v.Type.GLSLSize() * cols
	case SVC_MATRIX_COLUMNS, SVC_MATRIX_ROWS, SVC_V10_MATRIX_COLUMNS, SVC_V10_MATRIX_ROWS:
		toReturn = v.Type.GLSLSize() * 4 * int(v.Rows)
	default:
		panic("Unimplemented variable class!")
	}

	if v.Elements > 0 {
		toReturn *= int(v.Elements)
	}

	return toReturn
}

func (v Variable) ToNative() any {
	switch v.Class {
	case SVC_SCALAR, SVC_V10_SCALAR:
		switch v.Type.GLSLType() {
		case "float":
			if v.Elements > 1 {
				return make([]float32, v.Elements)
			}
			return float32(0.0)
		case "bool":
			if v.Elements > 1 {
				return make([]bool, v.Elements)
			}
			return false
		case "int":
			if v.Elements > 1 {
				return make([]int32, v.Elements)
			}
			return int32(0)
		case "uint":
			if v.Elements > 1 {
				return make([]uint32, v.Elements)
			}
			return uint32(0)
		case "double":
			if v.Elements > 1 {
				return make([]float64, v.Elements)
			}
			return float64(0.0)
		}
	case SVC_VECTOR, SVC_V10_VECTOR:
		switch v.Type.GLSLType() {
		case "float":
			switch v.Cols {
			case 2:
				if v.Elements > 1 {
					return make([]mgl32.Vec2, v.Elements)
				}
				return mgl32.Vec2{}
			case 3:
				if v.Elements > 1 {
					return make([]mgl32.Vec3, v.Elements)
				}
				return mgl32.Vec3{}
			case 4:
				if v.Elements > 1 {
					return make([]mgl32.Vec4, v.Elements)
				}
				return mgl32.Vec4{}
			}
		case "bool":
			switch v.Cols {
			case 2:
				if v.Elements > 1 {
					return make([][2]bool, v.Elements)
				}
				return [2]bool{}
			case 3:
				if v.Elements > 1 {
					return make([][3]bool, v.Elements)
				}
				return [3]bool{}
			case 4:
				if v.Elements > 1 {
					return make([][4]bool, v.Elements)
				}
				return [4]bool{}
			}
		case "int":
			switch v.Cols {
			case 2:
				if v.Elements > 1 {
					return make([][2]int32, v.Elements)
				}
				return [2]int32{}
			case 3:
				if v.Elements > 1 {
					return make([][3]int32, v.Elements)
				}
				return [3]int32{}
			case 4:
				if v.Elements > 1 {
					return make([][4]int32, v.Elements)
				}
				return [4]int32{}
			}
		case "uint":
			switch v.Cols {
			case 2:
				if v.Elements > 1 {
					return make([][2]uint32, v.Elements)
				}
				return [2]uint32{}
			case 3:
				if v.Elements > 1 {
					return make([][3]uint32, v.Elements)
				}
				return [3]uint32{}
			case 4:
				if v.Elements > 1 {
					return make([][4]uint32, v.Elements)
				}
				return [4]uint32{}
			}
		case "double":
			switch v.Cols {
			case 2:
				if v.Elements > 1 {
					return make([][2]float64, v.Elements)
				}
				return [2]float64{}
			case 3:
				if v.Elements > 1 {
					return make([][3]float64, v.Elements)
				}
				return [3]float64{}
			case 4:
				if v.Elements > 1 {
					return make([][4]float64, v.Elements)
				}
				return [4]float64{}
			}
		}
	case SVC_MATRIX_COLUMNS, SVC_MATRIX_ROWS, SVC_V10_MATRIX_COLUMNS, SVC_V10_MATRIX_ROWS:
		switch v.Type.GLSLType() {
		case "float":
			switch v.Cols {
			case 2:
				switch v.Rows {
				case 2:
					if v.Elements > 1 {
						return make([]mgl32.Mat2, v.Elements)
					}
					return mgl32.Mat2{}
				case 3:
					if v.Elements > 1 {
						return make([]mgl32.Mat2x3, v.Elements)
					}
					return mgl32.Mat2x3{}
				case 4:
					if v.Elements > 1 {
						return make([]mgl32.Mat2x4, v.Elements)
					}
					return mgl32.Mat2x4{}
				}
			case 3:
				switch v.Rows {
				case 2:
					if v.Elements > 1 {
						return make([]mgl32.Mat3x2, v.Elements)
					}
					return mgl32.Mat3x2{}
				case 3:
					if v.Elements > 1 {
						return make([]mgl32.Mat3, v.Elements)
					}
					return mgl32.Mat3{}
				case 4:
					if v.Elements > 1 {
						return make([]mgl32.Mat3x4, v.Elements)
					}
					return mgl32.Mat3x4{}
				}
			case 4:
				switch v.Rows {
				case 2:
					if v.Elements > 1 {
						return make([]mgl32.Mat4x2, v.Elements)
					}
					return mgl32.Mat4x2{}
				case 3:
					if v.Elements > 1 {
						return make([]mgl32.Mat4x3, v.Elements)
					}
					return mgl32.Mat4x3{}
				case 4:
					if v.Elements > 1 {
						return make([]mgl32.Mat4, v.Elements)
					}
					return mgl32.Mat4{}
				}
			}
		}
	default:
		panic("Unimplemented variable class!")
	}
	return nil
}

func (v Variable) SwizzleFromSrc(swizzleSrc [4]int8, mask uint8) string {
	if v.Class == SVC_SCALAR || v.Class == SVC_V10_SCALAR {
		return ""
	}
	cmp := [4]string{"x", "y", "z", "w"}
	toReturn := "."
	base := int8(v.BufferOffset%16) / 4
	for i, src := range swizzleSrc {
		if (mask & (1 << i)) == 0 {
			continue
		}
		if src == -1 {
			break
		}
		toReturn += cmp[src-base]
	}
	return toReturn
}

type ConstantBuffer struct {
	Name      string
	Variables []Variable
	Size      uint32
	Flags     ConstantBufferFlags
	Type      ConstantBufferType
}

func (cb *ConstantBuffer) ToGLSL(idx int) string {
	toReturn := fmt.Sprintf("/* Constant Buffer %v: %v\n", idx, cb.Name)
	toReturn += fmt.Sprintf(" * Type: %v\n", cb.Type.ToString())
	toReturn += fmt.Sprintf(" * Size: %v\n", cb.Size)
	toReturn += fmt.Sprintf(" * Flags: %v\n", cb.Flags.ToString())
	toReturn += " */\n"
	bindingIdx := slices.Index(glsl.UniformBlockNames, cb.Name)
	if bindingIdx == -1 {
		toReturn += fmt.Sprintf("layout(std140) uniform %v {\n", cb.Name)
	} else {
		toReturn += fmt.Sprintf("layout(std140, binding = %v) uniform %v {\n", bindingIdx, cb.Name)
	}
	for _, variable := range cb.Variables {
		toReturn += fmt.Sprintf("    %v\n", variable.ToGLSL())
	}
	toReturn += "};"
	return toReturn
}

func (cb *ConstantBuffer) ToUniform() shaders.DynamicUniformBlock {
	toReturn := shaders.NewDynamicUniformBlock(cb.Name)
	for _, variable := range cb.Variables {
		toReturn.Append(variable.Name, variable.ToNative())
	}
	return toReturn
}

func (cb *ConstantBuffer) VariableFromOffset(offset uint32) (*Variable, uint32, error) {
	if offset >= cb.Size {
		return nil, 0, fmt.Errorf("offset out of range")
	}
	for i, variable := range cb.Variables {
		if variable.BufferOffset == offset || i+1 == len(cb.Variables) {
			return &cb.Variables[i], (offset - variable.BufferOffset) / 16, nil
		}
		nextVar := cb.Variables[i+1]
		if nextVar.BufferOffset <= offset {
			continue
		}
		return &cb.Variables[i], (offset - variable.BufferOffset) / 16, nil
	}
	return nil, 0, fmt.Errorf("should be unreachable")
}

type SystemValueType uint32

const (
	SV_UNDEFINED SystemValueType = iota
	SV_POSITION
	SV_CLIP_DISTANCE
	SV_CULL_DISTANCE
	SV_RENDER_TARGET_ARRAY_INDEX
	SV_VIEWPORT_ARRAY_INDEX
	SV_VERTEX_ID
	SV_PRIMITIVE_ID
	SV_INSTANCE_ID
	SV_IS_FRONT_FACE
	SV_SAMPLE_INDEX
	SV_FINAL_QUAD_EDGE_TESS_FACTOR
	SV_FINAL_QUAD_INSIDE_TESS_FACTOR
	SV_FINAL_TRI_EDGE_TESS_FACTOR
	SV_FINAL_TRI_INSIDE_TESS_FACTOR
	SV_FINAL_LINE_DETAIL_TESS_FACTOR
	SV_FINAL_LINE_DENSITY_TESS_FACTOR
	SV_TARGET              SystemValueType = 64
	SV_DEPTH               SystemValueType = 65
	SV_COVERAGE            SystemValueType = 66
	SV_DEPTH_GREATER_EQUAL SystemValueType = 67
	SV_DEPTH_LESS_EQUAL    SystemValueType = 68
)

func (svt SystemValueType) ToString() string {
	switch svt {
	case SV_UNDEFINED:
		return "SV_UNDEFINED"
	case SV_POSITION:
		return "SV_POSITION"
	case SV_CLIP_DISTANCE:
		return "SV_CLIP_DISTANCE"
	case SV_CULL_DISTANCE:
		return "SV_CULL_DISTANCE"
	case SV_RENDER_TARGET_ARRAY_INDEX:
		return "SV_RENDER_TARGET_ARRAY_INDEX"
	case SV_VIEWPORT_ARRAY_INDEX:
		return "SV_VIEWPORT_ARRAY_INDEX"
	case SV_VERTEX_ID:
		return "SV_VERTEX_ID"
	case SV_PRIMITIVE_ID:
		return "SV_PRIMITIVE_ID"
	case SV_INSTANCE_ID:
		return "SV_INSTANCE_ID"
	case SV_IS_FRONT_FACE:
		return "SV_IS_FRONT_FACE"
	case SV_SAMPLE_INDEX:
		return "SV_SAMPLE_INDEX"
	case SV_FINAL_QUAD_EDGE_TESS_FACTOR:
		return "SV_FINAL_QUAD_EDGE_TESS_FACTOR"
	case SV_FINAL_QUAD_INSIDE_TESS_FACTOR:
		return "SV_FINAL_QUAD_INSIDE_TESS_FACTOR"
	case SV_FINAL_TRI_EDGE_TESS_FACTOR:
		return "SV_FINAL_TRI_EDGE_TESS_FACTOR"
	case SV_FINAL_TRI_INSIDE_TESS_FACTOR:
		return "SV_FINAL_TRI_INSIDE_TESS_FACTOR"
	case SV_FINAL_LINE_DETAIL_TESS_FACTOR:
		return "SV_FINAL_LINE_DETAIL_TESS_FACTOR"
	case SV_FINAL_LINE_DENSITY_TESS_FACTOR:
		return "SV_FINAL_LINE_DENSITY_TESS_FACTOR"
	case SV_TARGET:
		return "SV_TARGET"
	case SV_DEPTH:
		return "SV_DEPTH"
	case SV_COVERAGE:
		return "SV_COVERAGE"
	case SV_DEPTH_GREATER_EQUAL:
		return "SV_DEPTH_GREATER_EQUAL"
	case SV_DEPTH_LESS_EQUAL:
		return "SV_DEPTH_LESS_EQUAL"
	}
	return "Unknown system value type!"
}

type RegisterComponentType uint32

const (
	RCT_UNKNOWN RegisterComponentType = iota
	RCT_UINT32
	RCT_SINT32
	RCT_FLOAT32
)

func (rct RegisterComponentType) GLSLPrefix() string {
	switch rct {
	case RCT_UINT32:
		return "u"
	case RCT_SINT32:
		return "s"
	default:
		return ""
	}
}

func (rct RegisterComponentType) GLSLType() string {
	switch rct {
	case RCT_UINT32:
		return "uint"
	case RCT_SINT32:
		return "int"
	case RCT_FLOAT32:
		return "float"
	default:
		return ""
	}
}

type Element struct {
	Name          string
	SemanticIndex uint32
	SystemValue   SystemValueType
	ComponentType RegisterComponentType
	Register      uint32
	Mask          uint8
	RWMask        uint8
}

func (e Element) ToGLSL(isInput bool) string {
	direction := "out"
	if isInput {
		direction = "in"
	}
	result := fmt.Sprintf("layout(location = %v) %v ", e.Register, direction)

	switch e.Mask {
	case 0x1:
		result += e.ComponentType.GLSLType()
	case 0x2, 0x3:
		result += fmt.Sprintf("%vvec2", e.ComponentType.GLSLPrefix())
	case 0x4, 0x5, 0x6, 0x7:
		result += fmt.Sprintf("%vvec3", e.ComponentType.GLSLPrefix())
	case 0x8, 0x9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF:
		result += fmt.Sprintf("%vvec4", e.ComponentType.GLSLPrefix())
	default:
		panic(fmt.Sprintf("Element with unexpected mask %x - %v", e.Mask, e.ComponentType))
	}

	result += fmt.Sprintf(" %v; // %v", e.NameWithIndex(isInput), e.SystemValue.ToString())
	return result
}

func (e Element) NameWithIndex(isInput bool) string {
	semantic := ""
	switch e.SystemValue {
	case SV_POSITION:
		break
	default:
		semantic = strconv.Itoa(int(e.SemanticIndex))
	}
	prefix := "i"
	if !isInput {
		prefix = "o"
	}
	return fmt.Sprintf("%v%v%v", prefix, e.Name, semantic)
}

type ShaderInputType uint32

const (
	CBUFFER ShaderInputType = iota
	TBUFFER
	TEXTURE
	SAMPLER
	UAV_RW_TYPED
	STRUCTURED
	UAV_RW_STRUCTURED
	BYTE_ADDRESS
	UAV_RW_BYTE_ADDRESS
	UAV_APPEND_STRUCTURED
	UAV_CONSUME_STRUCTURED
	UAV_RW_STRUCTURED_WITH_COUNTER
)

func (sit ShaderInputType) ToString() string {
	switch sit {
	case CBUFFER:
		return "CBUFFER"
	case TBUFFER:
		return "TBUFFER"
	case TEXTURE:
		return "TEXTURE"
	case SAMPLER:
		return "SAMPLER"
	case UAV_RW_TYPED:
		return "UAV_RW_TYPED"
	case STRUCTURED:
		return "STRUCTURED"
	case UAV_RW_STRUCTURED:
		return "UAV_RW_STRUCTURED"
	case BYTE_ADDRESS:
		return "BYTE_ADDRESS"
	case UAV_RW_BYTE_ADDRESS:
		return "UAV_RW_BYTE_ADDRESS"
	case UAV_APPEND_STRUCTURED:
		return "UAV_APPEND_STRUCTURED"
	case UAV_CONSUME_STRUCTURED:
		return "UAV_CONSUME_STRUCTURED"
	case UAV_RW_STRUCTURED_WITH_COUNTER:
		return "UAV_RW_STRUCTURED_WITH_COUNTER"
	}
	return "unknown shader input type"
}

type ShaderResourceReturnType uint32

const (
	NO_RETURN ShaderResourceReturnType = iota
	UNORM
	SNORM
	SINT
	UINT
	FLOAT
	MIXED
	DOUBLE
	CONTINUED
)

func (srrt ShaderResourceReturnType) ToString() string {
	switch srrt {
	case NO_RETURN:
		return "NO_RETURN"
	case UNORM:
		return "UNORM"
	case SNORM:
		return "SNORM"
	case SINT:
		return "SINT"
	case UINT:
		return "UINT"
	case FLOAT:
		return "FLOAT"
	case MIXED:
		return "MIXED"
	case DOUBLE:
		return "DOUBLE"
	case CONTINUED:
		return "CONTINUED"
	}
	return "unknown shader resource return type"
}

func (srrt ShaderResourceReturnType) ToOpcodeNumberType() opcodeNumberType {
	switch srrt {
	case SINT:
		return internalNumberTypeInt
	case UINT:
		return internalNumberTypeUInt
	case UNORM, SNORM, FLOAT:
		return internalNumberTypeFloat
	case DOUBLE:
		return internalNumberTypeDouble
	}
	return internalNumberTypeUnknown
}

func (srrt ShaderResourceReturnType) GLSLPrefix() string {
	switch srrt {
	case SINT:
		return "i"
	case UINT:
		return "u"
	default:
		return ""
	}
}

type ShaderResourceViewDimension uint32

const (
	UNKNOWN ShaderResourceViewDimension = iota
	BUFFER
	TEXTURE_1D
	TEXTURE_1D_ARRAY
	TEXTURE_2D
	TEXTURE_2D_ARRAY
	TEXTURE_2D_MULTISAMPLED
	TEXTURE_3D
	TEXTURE_CUBE
	TEXTURE_CUBE_ARRAY
	EXTENDED_BUFFER
)

func (srvd ShaderResourceViewDimension) ToString() string {
	switch srvd {
	case UNKNOWN:
		return "UNKNOWN"
	case BUFFER:
		return "BUFFER"
	case TEXTURE_1D:
		return "TEXTURE_1D"
	case TEXTURE_1D_ARRAY:
		return "TEXTURE_1D_ARRAY"
	case TEXTURE_2D:
		return "TEXTURE_2D"
	case TEXTURE_2D_ARRAY:
		return "TEXTURE_2D_ARRAY"
	case TEXTURE_2D_MULTISAMPLED:
		return "TEXTURE_2D_MULTISAMPLED"
	case TEXTURE_3D:
		return "TEXTURE_3D"
	case TEXTURE_CUBE:
		return "TEXTURE_CUBE"
	case TEXTURE_CUBE_ARRAY:
		return "TEXTURE_CUBE_ARRAY"
	case EXTENDED_BUFFER:
		return "EXTENDED_BUFFER"
	}
	return "unknown shader resource view dimension"
}

func (srvd ShaderResourceViewDimension) ToGLSL() string {
	switch srvd {
	case BUFFER:
		return "samplerBuffer"
	case TEXTURE_1D:
		return "sampler1D"
	case TEXTURE_1D_ARRAY:
		return "sampler1DArray"
	case TEXTURE_2D:
		return "sampler2D"
	case TEXTURE_2D_ARRAY:
		return "sampler2DArray"
	case TEXTURE_2D_MULTISAMPLED:
		return "sampler2DMS"
	case TEXTURE_3D:
		return "sampler3D"
	case TEXTURE_CUBE:
		return "samplerCube"
	case TEXTURE_CUBE_ARRAY:
		return "samplerCubeArray"
	}
	return "unknown sampler type"
}

func (srvd ShaderResourceViewDimension) Dimensions() int {
	switch srvd {
	case BUFFER:
		return 1
	case TEXTURE_1D:
		return 1
	case TEXTURE_1D_ARRAY:
		return 2
	case TEXTURE_2D:
		return 2
	case TEXTURE_2D_ARRAY:
		return 3
	case TEXTURE_2D_MULTISAMPLED:
		return 2
	case TEXTURE_3D:
		return 3
	case TEXTURE_CUBE:
		return 3
	case TEXTURE_CUBE_ARRAY:
		return 4
	}
	return -1
}

type ShaderInputFlags uint32

const (
	SIF_NONE                ShaderInputFlags = 0x0
	SIF_USERPACKED          ShaderInputFlags = 0x1
	SIF_COMPARISON_SAMPLER  ShaderInputFlags = 0x2
	SIF_TEXTURE_COMPONENT_0 ShaderInputFlags = 0x4
	SIF_TEXTURE_COMPONENT_1 ShaderInputFlags = 0x8
	SIF_UNUSED              ShaderInputFlags = 0x10
)

func (sif ShaderInputFlags) ToString() string {
	if sif == SIF_NONE {
		return "NONE"
	}
	flags := make([]string, 0)
	if sif&SIF_USERPACKED != SIF_NONE {
		flags = append(flags, "USERPACKED")
	}
	if sif&SIF_COMPARISON_SAMPLER != SIF_NONE {
		flags = append(flags, "COMPARISON_SAMPLER")
	}
	if sif&SIF_TEXTURE_COMPONENT_0 != SIF_NONE {
		flags = append(flags, "TEXTURE_COMPONENT_0")
	}
	if sif&SIF_TEXTURE_COMPONENT_1 != SIF_NONE {
		flags = append(flags, "TEXTURE_COMPONENT_1")
	}
	if sif&SIF_UNUSED != SIF_NONE {
		flags = append(flags, "UNUSED")
	}
	return strings.Join(flags, " | ")
}

type ResourceBinding struct {
	Name          string
	InputType     ShaderInputType
	ReturnType    ShaderResourceReturnType
	ViewDimension ShaderResourceViewDimension
	SampleCount   uint32
	BindPoint     uint32
	BindCount     uint32
	Flags         ShaderInputFlags
}

func (rb ResourceBinding) ToString() string {
	toReturn := fmt.Sprintf("/* Resource Binding %v: %v\n", rb.BindPoint, rb.Name)
	toReturn += fmt.Sprintf(" *   Input Type: %v\n", rb.InputType.ToString())
	toReturn += fmt.Sprintf(" *   Return Type: %v\n", rb.ReturnType.ToString())
	toReturn += fmt.Sprintf(" *   View Dimension: %v\n", rb.ViewDimension.ToString())
	toReturn += fmt.Sprintf(" *   Sample Count: %v\n", rb.SampleCount)
	toReturn += fmt.Sprintf(" *   Bind Count: %v\n", rb.BindCount)
	toReturn += fmt.Sprintf(" *   Flags: %v\n", rb.Flags.ToString())
	return toReturn + " */\n\n"
}

func (rb ResourceBinding) ToGLSL() string {
	if rb.InputType != TEXTURE {
		return rb.ToString()
	}
	return fmt.Sprintf("uniform %v%v %v;\n", rb.ReturnType.GLSLPrefix(), rb.ViewDimension.ToGLSL(), strings.TrimLeft(rb.Name, "_"))
}
