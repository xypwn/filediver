package d3dops

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type DclGlobalFlags struct {
	opcode uint32
}

const (
	refactoringAllowed            = (1 << 11)
	enableDoubles                 = (1 << 12)
	earlyDepthStencil             = (1 << 13)
	enableRawStructuredBuffers    = (1 << 14)
	skipOptimizations             = (1 << 15)
	enableMinPrecision            = (1 << 16)
	enable11_1DoubleExtensions    = (1 << 17)
	enable11_1NonDoubleExtensions = (1 << 18)
	res19                         = (1 << 19)
	res20                         = (1 << 20)
	res21                         = (1 << 21)
	res22                         = (1 << 22)
	res23                         = (1 << 23)
)

func (glob *DclGlobalFlags) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	toReturn := "/* Global Flags:\n"
	if glob.opcode&refactoringAllowed != 0 {
		toReturn += " * refactoringAllowed\n"
	}
	if glob.opcode&enableDoubles != 0 {
		toReturn += " * enableDoubles\n"
	}
	if glob.opcode&earlyDepthStencil != 0 {
		toReturn += " * earlyDepthStencil\n"
	}
	if glob.opcode&enableRawStructuredBuffers != 0 {
		toReturn += " * enableRawStructuredBuffers\n"
	}
	if glob.opcode&skipOptimizations != 0 {
		toReturn += " * skipOptimizations\n"
	}
	if glob.opcode&enableMinPrecision != 0 {
		toReturn += " * enableMinPrecision\n"
	}
	if glob.opcode&enable11_1DoubleExtensions != 0 {
		toReturn += " * enable11_1DoubleExtensions\n"
	}
	if glob.opcode&enable11_1NonDoubleExtensions != 0 {
		toReturn += " * enable11_1NonDoubleExtensions\n"
	}
	if glob.opcode&res19 != 0 {
		toReturn += " * res19\n"
	}
	if glob.opcode&res20 != 0 {
		toReturn += " * res20\n"
	}
	if glob.opcode&res21 != 0 {
		toReturn += " * res21\n"
	}
	if glob.opcode&res22 != 0 {
		toReturn += " * res22\n"
	}
	if glob.opcode&res23 != 0 {
		toReturn += " * res23\n"
	}
	return toReturn + " */\n\n"
}

type DclConstantBuffer struct {
	opcode uint32
	data   []uint8
}

const CONSTANT_BUFFER_ACCESS_PATTERN_MASK uint32 = 0x00000800

func (cb *DclConstantBuffer) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	operand, err := ParseOperand(bytes.NewReader(cb.data), ShaderOpcodeType(cb.opcode&TYPE_MASK))
	if err != nil {
		panic(err)
	}

	toReturn := "// Declare Constant Buffer "
	if cb.opcode&CONSTANT_BUFFER_ACCESS_PATTERN_MASK != 0 {
		toReturn += "Dynamic"
	} else {
		toReturn += "Immediate"
	}
	toReturn += fmt.Sprintf(" indexed, register cb%v, size %v", operand.Indices[0].Value, operand.Indices[1].Value)

	return toReturn + "\n"
}

type DclSampler struct {
	opcode uint32
	data   []uint8
}

const SAMPLER_MODE_MASK uint32 = 0x00007800
const SAMPLER_MODE_SHIFT = 11

type SAMPLER_MODE uint8

const (
	SAMPLER_MODE_DEFAULT SAMPLER_MODE = iota
	SAMPLER_MODE_COMPARISON
	SAMPLER_MODE_MONO
)

func (m SAMPLER_MODE) ToString() string {
	switch m {
	case SAMPLER_MODE_DEFAULT:
		return "DEFAULT"
	case SAMPLER_MODE_COMPARISON:
		return "COMPARISON"
	case SAMPLER_MODE_MONO:
		return "MONO"
	}
	return "unknown sampler mode"
}

func (s *DclSampler) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	operand, err := ParseOperand(bytes.NewReader(s.data), ShaderOpcodeType(s.opcode&TYPE_MASK))
	if err != nil {
		panic(err)
	}
	mode := SAMPLER_MODE((s.opcode & SAMPLER_MODE_MASK) >> SAMPLER_MODE_SHIFT)

	toReturn := fmt.Sprintf("// Declare Sampler s%v mode %v", operand.Indices[0].Value, mode.ToString())

	return toReturn + "\n"
}

type DclResource struct {
	opcode uint32
	data   []uint8
}

const RESOURCE_DIMENSION_MASK uint32 = 0x0000f800
const RESOURCE_DIMENSION_SHIFT = 11

type RESOURCE_DIMENSION uint8

const (
	RESOURCE_DIMENSION_UNKNOWN RESOURCE_DIMENSION = iota
	RESOURCE_DIMENSION_BUFFER
	RESOURCE_DIMENSION_TEXTURE1D
	RESOURCE_DIMENSION_TEXTURE2D
	RESOURCE_DIMENSION_TEXTURE2DMS
	RESOURCE_DIMENSION_TEXTURE3D
	RESOURCE_DIMENSION_TEXTURECUBE
	RESOURCE_DIMENSION_TEXTURE1DARRAY
	RESOURCE_DIMENSION_TEXTURE2DARRAY
	RESOURCE_DIMENSION_TEXTURE2DMSARRAY
	RESOURCE_DIMENSION_TEXTURECUBEARRAY
	RESOURCE_DIMENSION_RAW_BUFFER
	RESOURCE_DIMENSION_STRUCTURED_BUFFER
)

func (rd RESOURCE_DIMENSION) ToString() string {
	switch rd {
	case RESOURCE_DIMENSION_UNKNOWN:
		return "UNKNOWN"
	case RESOURCE_DIMENSION_BUFFER:
		return "BUFFER"
	case RESOURCE_DIMENSION_TEXTURE1D:
		return "TEXTURE1D"
	case RESOURCE_DIMENSION_TEXTURE2D:
		return "TEXTURE2D"
	case RESOURCE_DIMENSION_TEXTURE2DMS:
		return "TEXTURE2DMS"
	case RESOURCE_DIMENSION_TEXTURE3D:
		return "TEXTURE3D"
	case RESOURCE_DIMENSION_TEXTURECUBE:
		return "TEXTURECUBE"
	case RESOURCE_DIMENSION_TEXTURE1DARRAY:
		return "TEXTURE1DARRAY"
	case RESOURCE_DIMENSION_TEXTURE2DARRAY:
		return "TEXTURE2DARRAY"
	case RESOURCE_DIMENSION_TEXTURE2DMSARRAY:
		return "TEXTURE2DMSARRAY"
	case RESOURCE_DIMENSION_TEXTURECUBEARRAY:
		return "TEXTURECUBEARRAY"
	case RESOURCE_DIMENSION_RAW_BUFFER:
		return "RAW_BUFFER"
	case RESOURCE_DIMENSION_STRUCTURED_BUFFER:
		return "STRUCTURED_BUFFER"
	}
	return "unknown resource dimension"
}

type ResourceReturnType uint8

const (
	RETURN_TYPE_UNKNOWN ResourceReturnType = iota
	RETURN_TYPE_UNORM
	RETURN_TYPE_SNORM
	RETURN_TYPE_SINT
	RETURN_TYPE_UINT
	RETURN_TYPE_FLOAT
	RETURN_TYPE_MIXED
	RETURN_TYPE_DOUBLE
	RETURN_TYPE_CONTINUED
	RETURN_TYPE_UNUSED
)

func (rrt ResourceReturnType) ToString() string {
	switch rrt {
	case RETURN_TYPE_UNORM:
		return "UNORM"
	case RETURN_TYPE_SNORM:
		return "SNORM"
	case RETURN_TYPE_SINT:
		return "SINT"
	case RETURN_TYPE_UINT:
		return "UINT"
	case RETURN_TYPE_FLOAT:
		return "FLOAT"
	case RETURN_TYPE_MIXED:
		return "MIXED"
	case RETURN_TYPE_DOUBLE:
		return "DOUBLE"
	case RETURN_TYPE_CONTINUED:
		return "CONTINUED"
	case RETURN_TYPE_UNUSED:
		return "UNUSED"
	default:
		return "UNKNOWN"
	}
}

type ResourceReturnToken uint32

const (
	RRT_XMASK = 0x0000000f
	RRT_YMASK = 0x000000f0
	RRT_ZMASK = 0x00000f00
	RRT_WMASK = 0x0000f000
)

func (rrt ResourceReturnToken) X() ResourceReturnType {
	return ResourceReturnType(uint32(rrt) & RRT_XMASK)
}

func (rrt ResourceReturnToken) Y() ResourceReturnType {
	return ResourceReturnType((uint32(rrt) & RRT_YMASK) >> 4)
}

func (rrt ResourceReturnToken) Z() ResourceReturnType {
	return ResourceReturnType((uint32(rrt) & RRT_ZMASK) >> 8)
}

func (rrt ResourceReturnToken) W() ResourceReturnType {
	return ResourceReturnType((uint32(rrt) & RRT_WMASK) >> 12)
}

func (rrt ResourceReturnToken) ToString() string {
	return fmt.Sprintf("(%v, %v, %v, %v)", rrt.X().ToString(), rrt.Y().ToString(), rrt.Z().ToString(), rrt.W().ToString())
}

func (s *DclResource) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	r := bytes.NewReader(s.data)
	operandReg, err := ParseOperand(r, ShaderOpcodeType(s.opcode&TYPE_MASK))
	if err != nil {
		panic(err)
	}
	var returnType ResourceReturnToken
	err = binary.Read(r, binary.LittleEndian, &returnType)
	if err != nil {
		panic(err)
	}

	dimension := RESOURCE_DIMENSION((s.opcode & RESOURCE_DIMENSION_MASK) >> RESOURCE_DIMENSION_SHIFT)

	toReturn := fmt.Sprintf("// Declare %v Resource t%v -> %v", dimension.ToString(), operandReg.Indices[0].Value, returnType.ToString())

	return toReturn + "\n"
}

type INTERPOLATION_MODE uint8

const (
	INTERPOLATION_UNDEFINED INTERPOLATION_MODE = iota
	INTERPOLATION_CONSTANT
	INTERPOLATION_LINEAR
	INTERPOLATION_LINEAR_CENTROID
	INTERPOLATION_LINEAR_NOPERSPECTIVE
	INTERPOLATION_LINEAR_NOPERSPECTIVE_CENTROID
	INTERPOLATION_LINEAR_SAMPLE
	INTERPOLATION_LINEAR_NOPERSPECTIVE_SAMPLE
)

func (i INTERPOLATION_MODE) ToString() string {
	switch i {
	case INTERPOLATION_UNDEFINED:
		return "UNDEFINED"
	case INTERPOLATION_CONSTANT:
		return "CONSTANT"
	case INTERPOLATION_LINEAR:
		return "LINEAR"
	case INTERPOLATION_LINEAR_CENTROID:
		return "LINEAR_CENTROID"
	case INTERPOLATION_LINEAR_NOPERSPECTIVE:
		return "LINEAR_NOPERSPECTIVE"
	case INTERPOLATION_LINEAR_NOPERSPECTIVE_CENTROID:
		return "LINEAR_NOPERSPECTIVE_CENTROID"
	case INTERPOLATION_LINEAR_SAMPLE:
		return "LINEAR_SAMPLE"
	case INTERPOLATION_LINEAR_NOPERSPECTIVE_SAMPLE:
		return "LINEAR_NOPERSPECTIVE_SAMPLE"
	}
	return "unknown interpolation mode"
}

type DclInputPS struct {
	opcode uint32
	data   []uint8
}

const INTERPOLATION_MODE_MASK uint32 = 0x00007800
const INTERPOLATION_MODE_SHIFT = 11

func (i *DclInputPS) InterpolationMode() INTERPOLATION_MODE {
	return INTERPOLATION_MODE((i.opcode & INTERPOLATION_MODE_MASK) >> INTERPOLATION_MODE_SHIFT)
}

func (i *DclInputPS) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	operand, err := ParseOperand(bytes.NewReader(i.data), ShaderOpcodeType(i.opcode&TYPE_MASK))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("// Declare Input Pixel Shader: %v v%v%v\n", i.InterpolationMode().ToString(), operand.Indices[0].Value, operand.Swizzle())
}

type NAME uint16

const (
	NAME_UNDEFINED NAME = iota
	NAME_POSITION
	NAME_CLIP_DISTANCE
	NAME_CULL_DISTANCE
	NAME_RENDER_TARGET_ARRAY_INDEX
	NAME_VIEWPORT_ARRAY_INDEX
	NAME_VERTEX_ID
	NAME_PRIMITIVE_ID
	NAME_INSTANCE_ID
	NAME_IS_FRONT_FACE
	NAME_SAMPLE_INDEX
	NAME_FINAL_QUAD_U_EQ_0_EDGE_TESSFACTOR
	NAME_FINAL_QUAD_V_EQ_0_EDGE_TESSFACTOR
	NAME_FINAL_QUAD_U_EQ_1_EDGE_TESSFACTOR
	NAME_FINAL_QUAD_V_EQ_1_EDGE_TESSFACTOR
	NAME_FINAL_QUAD_U_INSIDE_TESSFACTOR
	NAME_FINAL_QUAD_V_INSIDE_TESSFACTOR
	NAME_FINAL_TRI_U_EQ_0_EDGE_TESSFACTOR
	NAME_FINAL_TRI_V_EQ_0_EDGE_TESSFACTOR
	NAME_FINAL_TRI_W_EQ_0_EDGE_TESSFACTOR
	NAME_FINAL_TRI_INSIDE_TESSFACTOR
	NAME_FINAL_LINE_DETAIL_TESSFACTOR
	NAME_FINAL_LINE_DENSITY_TESSFACTOR
)

func (n NAME) ToString() string {
	switch n {
	case NAME_UNDEFINED:
		return "UNDEFINED"
	case NAME_POSITION:
		return "POSITION"
	case NAME_CLIP_DISTANCE:
		return "CLIP_DISTANCE"
	case NAME_CULL_DISTANCE:
		return "CULL_DISTANCE"
	case NAME_RENDER_TARGET_ARRAY_INDEX:
		return "RENDER_TARGET_ARRAY_INDEX"
	case NAME_VIEWPORT_ARRAY_INDEX:
		return "VIEWPORT_ARRAY_INDEX"
	case NAME_VERTEX_ID:
		return "VERTEX_ID"
	case NAME_PRIMITIVE_ID:
		return "PRIMITIVE_ID"
	case NAME_INSTANCE_ID:
		return "INSTANCE_ID"
	case NAME_IS_FRONT_FACE:
		return "IS_FRONT_FACE"
	case NAME_SAMPLE_INDEX:
		return "SAMPLE_INDEX"
	case NAME_FINAL_QUAD_U_EQ_0_EDGE_TESSFACTOR:
		return "FINAL_QUAD_U_EQ_0_EDGE_TESSFACTOR"
	case NAME_FINAL_QUAD_V_EQ_0_EDGE_TESSFACTOR:
		return "FINAL_QUAD_V_EQ_0_EDGE_TESSFACTOR"
	case NAME_FINAL_QUAD_U_EQ_1_EDGE_TESSFACTOR:
		return "FINAL_QUAD_U_EQ_1_EDGE_TESSFACTOR"
	case NAME_FINAL_QUAD_V_EQ_1_EDGE_TESSFACTOR:
		return "FINAL_QUAD_V_EQ_1_EDGE_TESSFACTOR"
	case NAME_FINAL_QUAD_U_INSIDE_TESSFACTOR:
		return "FINAL_QUAD_U_INSIDE_TESSFACTOR"
	case NAME_FINAL_QUAD_V_INSIDE_TESSFACTOR:
		return "FINAL_QUAD_V_INSIDE_TESSFACTOR"
	case NAME_FINAL_TRI_U_EQ_0_EDGE_TESSFACTOR:
		return "FINAL_TRI_U_EQ_0_EDGE_TESSFACTOR"
	case NAME_FINAL_TRI_V_EQ_0_EDGE_TESSFACTOR:
		return "FINAL_TRI_V_EQ_0_EDGE_TESSFACTOR"
	case NAME_FINAL_TRI_W_EQ_0_EDGE_TESSFACTOR:
		return "FINAL_TRI_W_EQ_0_EDGE_TESSFACTOR"
	case NAME_FINAL_TRI_INSIDE_TESSFACTOR:
		return "FINAL_TRI_INSIDE_TESSFACTOR"
	case NAME_FINAL_LINE_DETAIL_TESSFACTOR:
		return "FINAL_LINE_DETAIL_TESSFACTOR"
	case NAME_FINAL_LINE_DENSITY_TESSFACTOR:
		return "FINAL_LINE_DENSITY_TESSFACTOR"
	}
	return "unknown name value"
}

const NAME_MASK = 0x0000ffff

type DclInputPSSIV struct {
	opcode uint32
	data   []uint8
}

func (i *DclInputPSSIV) InterpolationMode() INTERPOLATION_MODE {
	return INTERPOLATION_MODE((i.opcode & INTERPOLATION_MODE_MASK) >> INTERPOLATION_MODE_SHIFT)
}

func (i *DclInputPSSIV) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	r := bytes.NewReader(i.data)
	operand, err := ParseOperand(r, ShaderOpcodeType(i.opcode&TYPE_MASK))
	if err != nil {
		panic(err)
	}
	var nameToken uint32
	err = binary.Read(r, binary.LittleEndian, &nameToken)
	if err != nil {
		panic(err)
	}
	name := NAME(nameToken & NAME_MASK)
	return fmt.Sprintf("// Declare Input Pixel Shader SIV: %v v%v%v, %v\n", i.InterpolationMode().ToString(), operand.Indices[0].Value, operand.Swizzle(), name.ToString())
}

type DclOutput struct {
	opcode uint32
	data   []uint8
}

func (i *DclOutput) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	operand, err := ParseOperand(bytes.NewReader(i.data), ShaderOpcodeType(i.opcode&TYPE_MASK))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("// Declare Output: o%v%v\n", operand.Indices[0].Value, operand.Swizzle())
}

type DclTemps struct {
	opcode uint32
	data   []uint8
}

func (i *DclTemps) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	var tempCount uint32
	err := binary.Read(bytes.NewReader(i.data), binary.LittleEndian, &tempCount)
	if err != nil {
		panic(err)
	}
	toReturn := fmt.Sprintf("// Declare Temps: r0 - r%v\n", tempCount-1)
	for i := uint32(0); i < tempCount; i++ {
		toReturn += fmt.Sprintf("vec4 r%v;\n", i)
	}
	return toReturn
}

type DclIndexableTemp struct {
	opcode uint32
	data   []uint8
}

func (i *DclIndexableTemp) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	var index, count, components uint32
	r := bytes.NewReader(i.data)
	err := binary.Read(r, binary.LittleEndian, &index)
	if err != nil {
		panic(err)
	}
	err = binary.Read(r, binary.LittleEndian, &count)
	if err != nil {
		panic(err)
	}
	err = binary.Read(r, binary.LittleEndian, &components)
	if err != nil {
		panic(err)
	}
	var swizzle string
	switch components {
	case 1:
		swizzle = ".x"
	case 2:
		swizzle = ".xy"
	case 3:
		swizzle = ".xyz"
	case 4:
		swizzle = ".xyzw"
	}
	return fmt.Sprintf("// Declare Indexable Temp: x%v[%v]%v\nvec%v x%v[%v];\n", index, count, swizzle, components, index, count)
}

func ParseDeclaration(opcode uint32, data []uint8) (Opcode, error) {
	opType := ShaderOpcodeType(opcode & TYPE_MASK)
	switch opType {
	case OPCODE_DCL_GLOBAL_FLAGS:
		return &DclGlobalFlags{
			opcode: opcode,
		}, nil
	case OPCODE_DCL_CONSTANT_BUFFER:
		return &DclConstantBuffer{
			opcode: opcode,
			data:   data,
		}, nil
	case OPCODE_DCL_SAMPLER:
		return &DclSampler{
			opcode: opcode,
			data:   data,
		}, nil
	case OPCODE_DCL_RESOURCE:
		return &DclResource{
			opcode: opcode,
			data:   data,
		}, nil
	case OPCODE_DCL_INPUT_PS:
		return &DclInputPS{
			opcode: opcode,
			data:   data,
		}, nil
	case OPCODE_DCL_INPUT_PS_SIV:
		return &DclInputPSSIV{
			opcode: opcode,
			data:   data,
		}, nil
	case OPCODE_DCL_OUTPUT:
		return &DclOutput{
			opcode: opcode,
			data:   data,
		}, nil
	case OPCODE_DCL_TEMPS:
		return &DclTemps{
			opcode: opcode,
			data:   data,
		}, nil
	case OPCODE_DCL_INDEXABLE_TEMP:
		return &DclIndexableTemp{
			opcode: opcode,
			data:   data,
		}, nil
	}
	return nil, fmt.Errorf("unimplemented declaration opcode")
}
