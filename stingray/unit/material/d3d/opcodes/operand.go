package d3dops

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type OPERAND_NUM_COMPONENTS uint8

const (
	OPERAND_0_COMPONENT OPERAND_NUM_COMPONENTS = iota
	OPERAND_1_COMPONENT
	OPERAND_4_COMPONENT
	OPERAND_N_COMPONENT
)

const OPERAND_NUM_COMPONENTS_MASK uint32 = 0x00000003

type OPERAND_4_COMPONENT_SELECTION_MODE uint8

const (
	OPERAND_4_COMPONENT_MASK_MODE OPERAND_4_COMPONENT_SELECTION_MODE = iota
	OPERAND_4_COMPONENT_SWIZZLE_MODE
	OPERAND_4_COMPONENT_SELECT_1_MODE
)

func (t OPERAND_4_COMPONENT_SELECTION_MODE) ToString() string {
	switch t {
	case OPERAND_4_COMPONENT_MASK_MODE:
		return "MASK_MODE"
	case OPERAND_4_COMPONENT_SWIZZLE_MODE:
		return "SWIZZLE_MODE"
	case OPERAND_4_COMPONENT_SELECT_1_MODE:
		return "SELECT_1_MODE"
	}
	return "Unknown operand 4 component selection mode!"
}

const OPERAND_4_COMPONENT_SELECTION_MODE_MASK uint32 = 0x0000000c
const OPERAND_4_COMPONENT_SELECTION_MODE_SHIFT uint32 = 2

const OPERAND_4_COMPONENT_MASK_MASK uint32 = 0x000000f0
const OPERAND_4_COMPONENT_MASK_SHIFT uint32 = 4
const OPERAND_4_COMPONENT_MASK_X uint32 = 0x00000010
const OPERAND_4_COMPONENT_MASK_Y uint32 = 0x00000020
const OPERAND_4_COMPONENT_MASK_Z uint32 = 0x00000040
const OPERAND_4_COMPONENT_MASK_W uint32 = 0x00000080

type OPERAND_TYPE uint8

const (
	OPERAND_TYPE_TEMP           OPERAND_TYPE = 0 // Temporary Register File
	OPERAND_TYPE_INPUT          OPERAND_TYPE = 1 // General Input Register File
	OPERAND_TYPE_OUTPUT         OPERAND_TYPE = 2 // General Output Register File
	OPERAND_TYPE_INDEXABLE_TEMP OPERAND_TYPE = 3 // Temporary Register File (indexable)
	OPERAND_TYPE_IMMEDIATE32    OPERAND_TYPE = 4 // 32bit/component immediate value(s)
	// If for example, operand token bits
	// [01:00]==D3D10_SB_OPERAND_4_COMPONENT,
	// this means that the operand type:
	// D3D10_SB_OPERAND_TYPE_IMMEDIATE32
	// results in 4 additional 32bit
	// DWORDS present for the operand.
	OPERAND_TYPE_IMMEDIATE64               OPERAND_TYPE = 5  // 64bit/comp.imm.val(s)HI:LO
	OPERAND_TYPE_SAMPLER                   OPERAND_TYPE = 6  // Reference to sampler state
	OPERAND_TYPE_RESOURCE                  OPERAND_TYPE = 7  // Reference to memory resource (e.g. texture)
	OPERAND_TYPE_CONSTANT_BUFFER           OPERAND_TYPE = 8  // Reference to constant buffer
	OPERAND_TYPE_IMMEDIATE_CONSTANT_BUFFER OPERAND_TYPE = 9  // Reference to immediate constant buffer
	OPERAND_TYPE_LABEL                     OPERAND_TYPE = 10 // Label
	OPERAND_TYPE_INPUT_PRIMITIVEID         OPERAND_TYPE = 11 // Input primitive ID
	OPERAND_TYPE_OUTPUT_DEPTH              OPERAND_TYPE = 12 // Output Depth
	OPERAND_TYPE_NULL                      OPERAND_TYPE = 13 // Null register, used to discard results of operations
	// Below Are operands new in DX 10.1
	OPERAND_TYPE_RASTERIZER           OPERAND_TYPE = 14 // DX10.1 Rasterizer register, used to denote the depth/stencil and render target resources
	OPERAND_TYPE_OUTPUT_COVERAGE_MASK OPERAND_TYPE = 15 // DX10.1 PS output MSAA coverage mask (scalar)
	// Below Are operands new in DX 11
	OPERAND_TYPE_STREAM                             OPERAND_TYPE = 16 // Reference to GS stream output resource
	OPERAND_TYPE_FUNCTION_BODY                      OPERAND_TYPE = 17 // Reference to a function definition
	OPERAND_TYPE_FUNCTION_TABLE                     OPERAND_TYPE = 18 // Reference to a set of functions used by a class
	OPERAND_TYPE_INTERFACE                          OPERAND_TYPE = 19 // Reference to an interface
	OPERAND_TYPE_FUNCTION_INPUT                     OPERAND_TYPE = 20 // Reference to an input parameter to a function
	OPERAND_TYPE_FUNCTION_OUTPUT                    OPERAND_TYPE = 21 // Reference to an output parameter to a function
	OPERAND_TYPE_OUTPUT_CONTROL_POINT_ID            OPERAND_TYPE = 22 // HS Control Point phase input saying which output control point ID this is
	OPERAND_TYPE_INPUT_FORK_INSTANCE_ID             OPERAND_TYPE = 23 // HS Fork Phase input instance ID
	OPERAND_TYPE_INPUT_JOIN_INSTANCE_ID             OPERAND_TYPE = 24 // HS Join Phase input instance ID
	OPERAND_TYPE_INPUT_CONTROL_POINT                OPERAND_TYPE = 25 // HS Fork+Join, DS phase input control points (array of them)
	OPERAND_TYPE_OUTPUT_CONTROL_POINT               OPERAND_TYPE = 26 // HS Fork+Join phase output control points (array of them)
	OPERAND_TYPE_INPUT_PATCH_CONSTANT               OPERAND_TYPE = 27 // DS+HSJoin Input Patch Constants (array of them)
	OPERAND_TYPE_INPUT_DOMAIN_POINT                 OPERAND_TYPE = 28 // DS Input Domain point
	OPERAND_TYPE_THIS_POINTER                       OPERAND_TYPE = 29 // Reference to an interface this pointer
	OPERAND_TYPE_UNORDERED_ACCESS_VIEW              OPERAND_TYPE = 30 // Reference to UAV u#
	OPERAND_TYPE_THREAD_GROUP_SHARED_MEMORY         OPERAND_TYPE = 31 // Reference to Thread Group Shared Memory g#
	OPERAND_TYPE_INPUT_THREAD_ID                    OPERAND_TYPE = 32 // Compute Shader Thread ID
	OPERAND_TYPE_INPUT_THREAD_GROUP_ID              OPERAND_TYPE = 33 // Compute Shader Thread Group ID
	OPERAND_TYPE_INPUT_THREAD_ID_IN_GROUP           OPERAND_TYPE = 34 // Compute Shader Thread ID In Thread Group
	OPERAND_TYPE_INPUT_COVERAGE_MASK                OPERAND_TYPE = 35 // Pixel shader coverage mask input
	OPERAND_TYPE_INPUT_THREAD_ID_IN_GROUP_FLATTENED OPERAND_TYPE = 36 // Compute Shader Thread ID In Group Flattened to a 1D value.
	OPERAND_TYPE_INPUT_GS_INSTANCE_ID               OPERAND_TYPE = 37 // Input GS instance ID
	OPERAND_TYPE_OUTPUT_DEPTH_GREATER_EQUAL         OPERAND_TYPE = 38 // Output Depth, forced to be greater than or equal than current depth
	OPERAND_TYPE_OUTPUT_DEPTH_LESS_EQUAL            OPERAND_TYPE = 39 // Output Depth, forced to be less than or equal to current depth
	OPERAND_TYPE_CYCLE_COUNTER                      OPERAND_TYPE = 40 // Cycle counter
)

func (t OPERAND_TYPE) ToString() string {
	switch t {
	case OPERAND_TYPE_TEMP:
		return "TEMP"
	case OPERAND_TYPE_INPUT:
		return "INPUT"
	case OPERAND_TYPE_OUTPUT:
		return "OUTPUT"
	case OPERAND_TYPE_INDEXABLE_TEMP:
		return "INDEXABLE_TEMP"
	case OPERAND_TYPE_IMMEDIATE32:
		return "IMMEDIATE32"
	case OPERAND_TYPE_IMMEDIATE64:
		return "IMMEDIATE64"
	case OPERAND_TYPE_SAMPLER:
		return "SAMPLER"
	case OPERAND_TYPE_RESOURCE:
		return "RESOURCE"
	case OPERAND_TYPE_CONSTANT_BUFFER:
		return "CONSTANT_BUFFER"
	case OPERAND_TYPE_IMMEDIATE_CONSTANT_BUFFER:
		return "IMMEDIATE_CONSTANT_BUFFER"
	case OPERAND_TYPE_LABEL:
		return "LABEL"
	case OPERAND_TYPE_INPUT_PRIMITIVEID:
		return "INPUT_PRIMITIVEID"
	case OPERAND_TYPE_OUTPUT_DEPTH:
		return "OUTPUT_DEPTH"
	case OPERAND_TYPE_NULL:
		return "NULL"
	case OPERAND_TYPE_RASTERIZER:
		return "RASTERIZER"
	case OPERAND_TYPE_OUTPUT_COVERAGE_MASK:
		return "OUTPUT_COVERAGE_MASK"
	case OPERAND_TYPE_STREAM:
		return "STREAM"
	case OPERAND_TYPE_FUNCTION_BODY:
		return "FUNCTION_BODY"
	case OPERAND_TYPE_FUNCTION_TABLE:
		return "FUNCTION_TABLE"
	case OPERAND_TYPE_INTERFACE:
		return "INTERFACE"
	case OPERAND_TYPE_FUNCTION_INPUT:
		return "FUNCTION_INPUT"
	case OPERAND_TYPE_FUNCTION_OUTPUT:
		return "FUNCTION_OUTPUT"
	case OPERAND_TYPE_OUTPUT_CONTROL_POINT_ID:
		return "OUTPUT_CONTROL_POINT_ID"
	case OPERAND_TYPE_INPUT_FORK_INSTANCE_ID:
		return "INPUT_FORK_INSTANCE_ID"
	case OPERAND_TYPE_INPUT_JOIN_INSTANCE_ID:
		return "INPUT_JOIN_INSTANCE_ID"
	case OPERAND_TYPE_INPUT_CONTROL_POINT:
		return "INPUT_CONTROL_POINT"
	case OPERAND_TYPE_OUTPUT_CONTROL_POINT:
		return "OUTPUT_CONTROL_POINT"
	case OPERAND_TYPE_INPUT_PATCH_CONSTANT:
		return "INPUT_PATCH_CONSTANT"
	case OPERAND_TYPE_INPUT_DOMAIN_POINT:
		return "INPUT_DOMAIN_POINT"
	case OPERAND_TYPE_THIS_POINTER:
		return "THIS_POINTER"
	case OPERAND_TYPE_UNORDERED_ACCESS_VIEW:
		return "UNORDERED_ACCESS_VIEW"
	case OPERAND_TYPE_THREAD_GROUP_SHARED_MEMORY:
		return "THREAD_GROUP_SHARED_MEMORY"
	case OPERAND_TYPE_INPUT_THREAD_ID:
		return "INPUT_THREAD_ID"
	case OPERAND_TYPE_INPUT_THREAD_GROUP_ID:
		return "INPUT_THREAD_GROUP_ID"
	case OPERAND_TYPE_INPUT_THREAD_ID_IN_GROUP:
		return "INPUT_THREAD_ID_IN_GROUP"
	case OPERAND_TYPE_INPUT_COVERAGE_MASK:
		return "INPUT_COVERAGE_MASK"
	case OPERAND_TYPE_INPUT_THREAD_ID_IN_GROUP_FLATTENED:
		return "INPUT_THREAD_ID_IN_GROUP_FLATTENED"
	case OPERAND_TYPE_INPUT_GS_INSTANCE_ID:
		return "INPUT_GS_INSTANCE_ID"
	case OPERAND_TYPE_OUTPUT_DEPTH_GREATER_EQUAL:
		return "OUTPUT_DEPTH_GREATER_EQUAL"
	case OPERAND_TYPE_OUTPUT_DEPTH_LESS_EQUAL:
		return "OUTPUT_DEPTH_LESS_EQUAL"
	case OPERAND_TYPE_CYCLE_COUNTER:
		return "CYCLE_COUNTER"
	}
	return "Unknown operand type!"
}

func (t OPERAND_TYPE) ToGLSL() string {
	switch t {
	case OPERAND_TYPE_TEMP:
		return "r"
	case OPERAND_TYPE_INPUT:
		return "v"
	case OPERAND_TYPE_OUTPUT:
		return "o"
	case OPERAND_TYPE_INDEXABLE_TEMP:
		return "x"
	case OPERAND_TYPE_SAMPLER:
		return "s"
	case OPERAND_TYPE_CONSTANT_BUFFER:
		return "cb"
	case OPERAND_TYPE_IMMEDIATE_CONSTANT_BUFFER:
		return "icb"
	case OPERAND_TYPE_RESOURCE:
		return "t"
	}
	return "Unknown operand type!"
}

const OPERAND_TYPE_MASK uint32 = 0x000ff000
const OPERAND_TYPE_SHIFT uint32 = 12

type OPERAND_INDEX_DIMENSION uint8

const (
	OPERAND_INDEX_0D OPERAND_INDEX_DIMENSION = iota
	OPERAND_INDEX_1D
	OPERAND_INDEX_2D
	OPERAND_INDEX_3D
)

func (r OPERAND_INDEX_DIMENSION) ToString() string {
	switch r {
	case OPERAND_INDEX_0D:
		return "0D"
	case OPERAND_INDEX_1D:
		return "1D"
	case OPERAND_INDEX_2D:
		return "2D"
	case OPERAND_INDEX_3D:
		return "3D"
	}
	return "Unknown operand index dimension!"
}

const OPERAND_INDEX_DIMENSION_MASK uint32 = 0x00300000
const OPERAND_INDEX_DIMENSION_SHIFT uint32 = 20

type OPERAND_INDEX_REPRESENTATION uint8

const (
	OPERAND_INDEX_IMMEDIATE32               OPERAND_INDEX_REPRESENTATION = iota // 1 dword
	OPERAND_INDEX_IMMEDIATE64                                                   // 2 dwords (HI32:LO32)
	OPERAND_INDEX_RELATIVE                                                      // Extra operand
	OPERAND_INDEX_IMMEDIATE32_PLUS_RELATIVE                                     // Extra dword and extra operand
	OPERAND_INDEX_IMMEDIATE64_PLUS_RELATIVE                                     // 2 extra dwords and extra operand
)

func (r OPERAND_INDEX_REPRESENTATION) ToString() string {
	switch r {
	case OPERAND_INDEX_IMMEDIATE32:
		return "IMMEDIATE32"
	case OPERAND_INDEX_IMMEDIATE64:
		return "IMMEDIATE64"
	case OPERAND_INDEX_RELATIVE:
		return "RELATIVE"
	case OPERAND_INDEX_IMMEDIATE32_PLUS_RELATIVE:
		return "IMMEDIATE32_PLUS_RELATIVE"
	case OPERAND_INDEX_IMMEDIATE64_PLUS_RELATIVE:
		return "IMMEDIATE64_PLUS_RELATIVE"
	}
	return "Unknown operand index representation!"
}

type OperandToken0 struct {
	Token uint32
}

func (o OperandToken0) NumComponents() int {
	switch OPERAND_NUM_COMPONENTS(o.Token & OPERAND_NUM_COMPONENTS_MASK) {
	case OPERAND_0_COMPONENT:
		return 0
	case OPERAND_1_COMPONENT:
		return 1
	case OPERAND_4_COMPONENT:
		return 4
	case OPERAND_N_COMPONENT:
		return -1
	}
	return -1
}

func (o OperandToken0) ComponentSelectionMode() OPERAND_4_COMPONENT_SELECTION_MODE {
	numComps := o.NumComponents()
	if numComps != 4 {
		return 0
	}
	return OPERAND_4_COMPONENT_SELECTION_MODE((o.Token & OPERAND_4_COMPONENT_SELECTION_MODE_MASK) >> OPERAND_4_COMPONENT_SELECTION_MODE_SHIFT)
}

func (o OperandToken0) SwizzleSrc() [4]int8 {
	toReturn := [4]int8{-1, -1, -1, -1}
	switch o.ComponentSelectionMode() {
	case OPERAND_4_COMPONENT_MASK_MODE:
		if o.Token&OPERAND_4_COMPONENT_MASK_X != 0 {
			toReturn[0] = 0
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_Y != 0 {
			toReturn[1] = 1
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_Z != 0 {
			toReturn[2] = 2
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_W != 0 {
			toReturn[3] = 3
		}
	case OPERAND_4_COMPONENT_SWIZZLE_MODE:
		xsrc := (o.Token & (0x3 << 4)) >> 4
		ysrc := (o.Token & (0x3 << 6)) >> 6
		zsrc := (o.Token & (0x3 << 8)) >> 8
		wsrc := (o.Token & (0x3 << 10)) >> 10
		toReturn[0] = int8(xsrc)
		toReturn[1] = int8(ysrc)
		toReturn[2] = int8(zsrc)
		toReturn[3] = int8(wsrc)
	case OPERAND_4_COMPONENT_SELECT_1_MODE:
		xsrc := (o.Token & (0x3 << 4)) >> 4
		toReturn[0] = int8(xsrc)
		toReturn[1] = int8(xsrc)
		toReturn[2] = int8(xsrc)
		toReturn[3] = int8(xsrc)
	}
	return toReturn
}

func (o OperandToken0) Mask() uint8 {
	numComps := o.NumComponents()
	if numComps != 4 {
		return 0x0
	}
	switch o.ComponentSelectionMode() {
	case OPERAND_4_COMPONENT_MASK_MODE:
		return uint8(o.Token >> 4)
	case OPERAND_4_COMPONENT_SWIZZLE_MODE:
		return 0xf
	case OPERAND_4_COMPONENT_SELECT_1_MODE:
		xsrc := (o.Token & (0x3 << 4)) >> 4
		return 1 << xsrc
	}
	return 0x0
}

func (o OperandToken0) MaskCount() int {
	mask := o.Mask()
	count := 0
	for i := 0; i < 4; i++ {
		if mask&(1<<i) != 0 {
			count++
		}
	}
	return count
}

func (o OperandToken0) Swizzle() string {
	numComps := o.NumComponents()
	if numComps != 4 {
		return ""
	}
	toReturn := "."
	switch o.ComponentSelectionMode() {
	case OPERAND_4_COMPONENT_MASK_MODE:
		if o.Token&OPERAND_4_COMPONENT_MASK_X != 0 {
			toReturn += "x"
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_Y != 0 {
			toReturn += "y"
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_Z != 0 {
			toReturn += "z"
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_W != 0 {
			toReturn += "w"
		}
	case OPERAND_4_COMPONENT_SWIZZLE_MODE:
		xsrc := (o.Token & (0x3 << 4)) >> 4
		ysrc := (o.Token & (0x3 << 6)) >> 6
		zsrc := (o.Token & (0x3 << 8)) >> 8
		wsrc := (o.Token & (0x3 << 10)) >> 10
		cmp := [4]string{"x", "y", "z", "w"}
		toReturn += cmp[xsrc] + cmp[ysrc] + cmp[zsrc] + cmp[wsrc]
	case OPERAND_4_COMPONENT_SELECT_1_MODE:
		xsrc := (o.Token & (0x3 << 4)) >> 4
		cmp := [4]string{"x", "y", "z", "w"}
		toReturn += cmp[xsrc]
	}
	return toReturn
}

func (o OperandToken0) SwizzleMask(mask uint8) string {
	numComps := o.NumComponents()
	if numComps != 4 || mask == 0x0 {
		return ""
	}
	toReturn := "."
	switch o.ComponentSelectionMode() {
	case OPERAND_4_COMPONENT_MASK_MODE:
		if o.Token&OPERAND_4_COMPONENT_MASK_X != 0 && (mask&0x1 != 0) {
			toReturn += "x"
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_Y != 0 && (mask&0x2 != 0) {
			toReturn += "y"
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_Z != 0 && (mask&0x4 != 0) {
			toReturn += "z"
		}
		if o.Token&OPERAND_4_COMPONENT_MASK_W != 0 && (mask&0x8 != 0) {
			toReturn += "w"
		}
	case OPERAND_4_COMPONENT_SWIZZLE_MODE:
		xsrc := (o.Token & (0x3 << 4)) >> 4
		ysrc := (o.Token & (0x3 << 6)) >> 6
		zsrc := (o.Token & (0x3 << 8)) >> 8
		wsrc := (o.Token & (0x3 << 10)) >> 10
		cmp := [4]string{"x", "y", "z", "w"}
		if mask&0x1 != 0 {
			toReturn += cmp[xsrc]
		}
		if mask&0x2 != 0 {
			toReturn += cmp[ysrc]
		}
		if mask&0x4 != 0 {
			toReturn += cmp[zsrc]
		}
		if mask&0x8 != 0 {
			toReturn += cmp[wsrc]
		}
	case OPERAND_4_COMPONENT_SELECT_1_MODE:
		xsrc := (o.Token & (0x3 << 4)) >> 4
		cmp := [4]string{"x", "y", "z", "w"}
		toReturn += cmp[xsrc]
	}
	return toReturn
}

func (o OperandToken0) Type() OPERAND_TYPE {
	return OPERAND_TYPE((o.Token & OPERAND_TYPE_MASK) >> OPERAND_TYPE_SHIFT)
}

func (o OperandToken0) IndexDimension() OPERAND_INDEX_DIMENSION {
	return OPERAND_INDEX_DIMENSION((o.Token & OPERAND_INDEX_DIMENSION_MASK) >> OPERAND_INDEX_DIMENSION_SHIFT)
}

func (o OperandToken0) IndexRepresentation(dimension uint8) OPERAND_INDEX_REPRESENTATION {
	if uint8(o.IndexDimension()) < dimension {
		return 0
	}
	return OPERAND_INDEX_REPRESENTATION((o.Token >> (22 + 3*(dimension))) & 0x7)
}

func (o OperandToken0) Extended() bool {
	return (o.Token & 0x80000000) != 0
}

func (o OperandToken0) ToString() string {
	toReturn := "/* Operand Token 0:\n"
	toReturn += fmt.Sprintf(" *   Components: %v\n", o.NumComponents())
	if o.NumComponents() == 4 {
		toReturn += fmt.Sprintf(" *     Component Selection Mode: %v\n", o.ComponentSelectionMode().ToString())
		toReturn += fmt.Sprintf(" *     Swizzle: %v\n", o.Swizzle())
	}
	toReturn += fmt.Sprintf(" *   Type: %v\n", o.Type().ToString())
	toReturn += fmt.Sprintf(" *   Index Dimension: %v\n", o.IndexDimension().ToString())
	if o.IndexDimension() >= OPERAND_INDEX_1D {
		toReturn += fmt.Sprintf(" *     Index Representation 1: %v\n", o.IndexRepresentation(1).ToString())
	}
	if o.IndexDimension() >= OPERAND_INDEX_2D {
		toReturn += fmt.Sprintf(" *     Index Representation 2: %v\n", o.IndexRepresentation(2).ToString())
	}
	if o.IndexDimension() == OPERAND_INDEX_3D {
		toReturn += fmt.Sprintf(" *     Index Representation 3: %v\n", o.IndexRepresentation(3).ToString())
	}
	if o.Extended() {
		toReturn += fmt.Sprintf(" *   Extended\n")
	}
	return toReturn + " */\n"
}

type EXTENDED_OPERAND_TYPE uint8

const (
	EXTENDED_OPERAND_EMPTY EXTENDED_OPERAND_TYPE = iota
	EXTENDED_OPERAND_MODIFIER
)

func (e EXTENDED_OPERAND_TYPE) ToString() string {
	switch e {
	case EXTENDED_OPERAND_EMPTY:
		return "EMPTY"
	case EXTENDED_OPERAND_MODIFIER:
		return "MODIFIER"
	}
	return "unknown extended operand type!"
}

const EXTENDED_OPERAND_TYPE_MASK uint32 = 0x0000003f

type OPERAND_MODIFIER uint8

const (
	OPERAND_MODIFIER_NONE OPERAND_MODIFIER = iota
	OPERAND_MODIFIER_NEG
	OPERAND_MODIFIER_ABS
	OPERAND_MODIFIER_ABSNEG
)

func (e OPERAND_MODIFIER) ToString() string {
	switch e {
	case OPERAND_MODIFIER_NONE:
		return "NONE"
	case OPERAND_MODIFIER_NEG:
		return "NEG"
	case OPERAND_MODIFIER_ABS:
		return "ABS"
	case OPERAND_MODIFIER_ABSNEG:
		return "ABSNEG"
	}
	return "unknown extended operand modifier!"
}

const OPERAND_MODIFIER_MASK uint32 = 0x00003fc0
const OPERAND_MODIFIER_SHIFT uint32 = 6

type MIN_PRECISION uint8

const (
	OPERAND_MIN_PRECISION_DEFAULT MIN_PRECISION = iota
	OPERAND_MIN_PRECISION_FLOAT_16
	OPERAND_MIN_PRECISION_FLOAT_2_8
	OPERAND_MIN_PRECISION_SINT_16
	OPERAND_MIN_PRECISION_UINT_16
)

func (m MIN_PRECISION) ToString() string {
	switch m {
	case OPERAND_MIN_PRECISION_DEFAULT:
		return "DEFAULT"
	case OPERAND_MIN_PRECISION_FLOAT_16:
		return "FLOAT_16"
	case OPERAND_MIN_PRECISION_FLOAT_2_8:
		return "FLOAT_2_8"
	case OPERAND_MIN_PRECISION_SINT_16:
		return "SINT_16"
	case OPERAND_MIN_PRECISION_UINT_16:
		return "UINT_16"
	}
	return "unknown min precision"
}

const MIN_PRECISION_MASK uint32 = 0x0001C000
const MIN_PRECISION_SHIFT uint32 = 14

type OperandToken1 struct {
	Token uint32
}

func (o OperandToken1) Type() EXTENDED_OPERAND_TYPE {
	return EXTENDED_OPERAND_TYPE(o.Token & EXTENDED_OPERAND_TYPE_MASK)
}

func (o OperandToken1) Modifier() OPERAND_MODIFIER {
	if o.Type() != EXTENDED_OPERAND_MODIFIER {
		return OPERAND_MODIFIER_NONE
	}
	return OPERAND_MODIFIER((o.Token & OPERAND_MODIFIER_MASK) >> OPERAND_MODIFIER_SHIFT)
}

func (o OperandToken1) MinPrecision() MIN_PRECISION {
	if o.Type() != EXTENDED_OPERAND_MODIFIER {
		return OPERAND_MIN_PRECISION_DEFAULT
	}
	return MIN_PRECISION((o.Token & MIN_PRECISION_MASK) >> MIN_PRECISION_SHIFT)
}

func (o OperandToken1) Extended() bool {
	return (o.Token & 0x80000000) != 0
}

func (o OperandToken1) ToString() string {
	toReturn := "/* Operand Token 1:\n"
	toReturn += fmt.Sprintf(" *   Type: %v\n", o.Type().ToString())
	toReturn += fmt.Sprintf(" *   Modifier: %v\n", o.Modifier().ToString())
	toReturn += fmt.Sprintf(" *   Minimum Precision: %v\n", o.MinPrecision().ToString())
	return toReturn + " */"
}

type Operand struct {
	OperandToken0
	OperandToken1
	parentOpcode ShaderOpcodeType
	Indices      []OperandIndex
	Immediate    [][]uint8
}

func ParseOperand(r io.Reader, parent ShaderOpcodeType) (*Operand, error) {
	var operand OperandToken0
	var extOperand OperandToken1
	err := binary.Read(r, binary.LittleEndian, &operand)
	if err != nil {
		return nil, fmt.Errorf("no operand present")
	}
	fmt.Printf("Read operand token 0 %v\n", strconv.FormatUint(uint64(operand.Token), 2))

	if operand.Extended() {
		err = binary.Read(r, binary.LittleEndian, &extOperand)
		if err != nil {
			return nil, fmt.Errorf("operand token 0 marked as extended, but no extended token was present")
		}
		fmt.Printf("Read operand token 1 %v\n", strconv.FormatUint(uint64(extOperand.Token), 2))
	}

	indices := make([]OperandIndex, 0)
	for i := uint8(0); i < uint8(operand.IndexDimension()); i++ {
		repr := operand.IndexRepresentation(i)
		fmt.Printf("Operand index %v - representation %v\n", i, repr.ToString())
		var value uint64
		var register *Operand
		switch repr {
		case OPERAND_INDEX_IMMEDIATE32:
			var temp uint32
			err := binary.Read(r, binary.LittleEndian, &temp)
			if err != nil {
				return nil, fmt.Errorf("unable to read imm32 index dimension %v of operand", i)
			}
			value = uint64(temp)
		case OPERAND_INDEX_IMMEDIATE64:
			err := binary.Read(r, binary.LittleEndian, &value)
			if err != nil {
				return nil, fmt.Errorf("unable to read imm64 index dimension %v of operand", i)
			}
		case OPERAND_INDEX_RELATIVE:
			register, err = ParseOperand(r, parent)
			if err != nil {
				return nil, fmt.Errorf("unable to read relative index dimension %v of operand", i)
			}
		case OPERAND_INDEX_IMMEDIATE32_PLUS_RELATIVE:
			var temp uint32
			err := binary.Read(r, binary.LittleEndian, &temp)
			if err != nil {
				return nil, fmt.Errorf("unable to read imm32 index dimension %v of operand", i)
			}
			value = uint64(temp)
			register, err = ParseOperand(r, parent)
			if err != nil {
				return nil, fmt.Errorf("unable to read relative index dimension %v of operand", i)
			}
		case OPERAND_INDEX_IMMEDIATE64_PLUS_RELATIVE:
			err := binary.Read(r, binary.LittleEndian, &value)
			if err != nil {
				return nil, fmt.Errorf("unable to read imm64 index dimension %v of operand", i)
			}
			register, err = ParseOperand(r, parent)
			if err != nil {
				return nil, fmt.Errorf("unable to read relative index dimension %v of operand", i)
			}
		}
		indices = append(indices, OperandIndex{
			Value:          value,
			Representation: repr,
			Register:       register,
		})
	}

	immediate := make([][]uint8, 0)
	componentByteSize := 4
	switch operand.Type() {
	case OPERAND_TYPE_IMMEDIATE64:
		componentByteSize = 8
	case OPERAND_TYPE_IMMEDIATE32:
		for i := 0; i < operand.NumComponents(); i++ {
			data := make([]uint8, componentByteSize)
			err := binary.Read(r, binary.LittleEndian, &data)
			if err != nil {
				return nil, fmt.Errorf("could not read operand immediate data")
			}
			immediate = append(immediate, data)
		}
	default:
		break
	}

	return &Operand{
		OperandToken0: operand,
		OperandToken1: extOperand,
		parentOpcode:  parent,
		Indices:       indices,
		Immediate:     immediate,
	}, nil
}

func (o *Operand) GetImmediateFloat() ([]float32, error) {
	toReturn := make([]float32, 0)
	for _, data := range o.Immediate {
		var val float32
		_, err := binary.Decode(data, binary.LittleEndian, &val)
		if err != nil {
			return nil, err
		}
		toReturn = append(toReturn, val)
	}
	return toReturn, nil
}

func (o *Operand) GetImmediateDouble() ([]float64, error) {
	toReturn := make([]float64, 0)
	for _, data := range o.Immediate {
		var val float64
		_, err := binary.Decode(data, binary.LittleEndian, &val)
		if err != nil {
			return nil, err
		}
		toReturn = append(toReturn, val)
	}
	return toReturn, nil
}

func (o *Operand) GetImmediateUInt() ([]uint32, error) {
	toReturn := make([]uint32, 0)
	for _, data := range o.Immediate {
		var val uint32
		_, err := binary.Decode(data, binary.LittleEndian, &val)
		if err != nil {
			return nil, err
		}
		toReturn = append(toReturn, val)
	}
	return toReturn, nil
}

func (o *Operand) GetImmediateInt() ([]int32, error) {
	toReturn := make([]int32, 0)
	for _, data := range o.Immediate {
		var val int32
		_, err := binary.Decode(data, binary.LittleEndian, &val)
		if err != nil {
			return nil, err
		}
		toReturn = append(toReturn, val)
	}
	return toReturn, nil
}

func (o *Operand) ToString() string {
	toReturn := o.OperandToken0.ToString()
	if o.OperandToken0.Extended() {
		toReturn += o.OperandToken1.ToString()
	}

	toReturn += "/* Indices:\n"
	for _, index := range o.Indices {
		switch index.Representation {
		case OPERAND_INDEX_IMMEDIATE32, OPERAND_INDEX_IMMEDIATE64:
			toReturn += fmt.Sprintf(" *   Immediate: %v\n", index.Value)
		case OPERAND_INDEX_RELATIVE:
			toReturn += fmt.Sprintf(" *   Relative: %v %v\n", index.Register.OperandToken0.Type().ToString(), index.Register.Indices[0].Value)
		case OPERAND_INDEX_IMMEDIATE32_PLUS_RELATIVE, OPERAND_INDEX_IMMEDIATE64_PLUS_RELATIVE:
			toReturn += fmt.Sprintf(" *   Immediate: %v\n", index.Value)
			toReturn += fmt.Sprintf(" *   + Relative: %v %v\n", index.Register.OperandToken0.Type().ToString(), index.Register.Indices[0].Value)
		}
	}
	toReturn += " */\n"

	toReturn += "/* Immediates:\n"
	switch o.parentOpcode.NumberType() {
	case internalNumberTypeDouble:
		imm, _ := o.GetImmediateDouble()
		for i, val := range imm {
			toReturn += fmt.Sprintf(" *   %v: %.3f\n", i, val)
		}
	case internalNumberTypeFloat:
		imm, _ := o.GetImmediateFloat()
		for i, val := range imm {
			toReturn += fmt.Sprintf(" *   %v: %.3f\n", i, val)
		}
	case internalNumberTypeUInt:
		imm, _ := o.GetImmediateUInt()
		for i, val := range imm {
			toReturn += fmt.Sprintf(" *   %v: %v\n", i, val)
		}
	case internalNumberTypeInt:
		imm, _ := o.GetImmediateInt()
		for i, val := range imm {
			toReturn += fmt.Sprintf(" *   %v: %v\n", i, val)
		}
	default:
		toReturn += " *   None\n"
	}
	toReturn += " */\n"

	return toReturn + "\n"
}

func (o *Operand) ToGLSL(cbs []ConstantBuffer, isg, osg []Element, res []ResourceBinding, mask uint8, isResult bool) string {
	operandType := o.OperandToken0.Type()
	modifier := o.Modifier()
	if operandType == OPERAND_TYPE_IMMEDIATE32 || operandType == OPERAND_TYPE_IMMEDIATE64 {
		immediate := ""
		maskCount := 0
		maskStart := -1
		for i := 0; i < 4; i++ {
			if mask&(1<<i) != 0 {
				if maskStart == -1 {
					maskStart = i
				}
				maskCount++
			}
		}
		if len(o.Immediate) > 1 && maskCount > 1 {
			immediate = fmt.Sprintf("%vvec%v(", o.parentOpcode.NumberType().Prefix(), maskCount)
		}
		added := 0
		switch o.parentOpcode.NumberType() {
		case internalNumberTypeDouble:
			imm, _ := o.GetImmediateDouble()
			if len(imm) == 1 {
				immediate += fmt.Sprintf("%.6f", imm[0])
			} else {
				for i, val := range imm {
					if mask&(1<<i) == 0 {
						continue
					}
					immediate += fmt.Sprintf("%.6f", val)
					added += 1
					if i+1 < len(imm) && added < maskCount {
						immediate += ", "
					}
				}
			}
		case internalNumberTypeFloat, internalNumberTypeUnknown:
			imm, _ := o.GetImmediateFloat()
			if len(imm) == 1 {
				immediate += fmt.Sprintf("%.6f", imm[0])
			} else {
				for i, val := range imm {
					if mask&(1<<i) == 0 {
						continue
					}
					immediate += fmt.Sprintf("%.6f", val)
					added += 1
					if i+1 < len(imm) && added < maskCount {
						immediate += ", "
					}
				}
			}
		case internalNumberTypeUInt:
			imm, _ := o.GetImmediateUInt()
			if len(imm) == 1 {
				immediate += fmt.Sprintf("%v", imm[0])
			} else {
				for i, val := range imm {
					if mask&(1<<i) == 0 {
						continue
					}
					immediate += fmt.Sprintf("%v", val)
					added += 1
					if i+1 < len(imm) && added < maskCount {
						immediate += ", "
					}
				}
			}
		case internalNumberTypeInt:
			imm, _ := o.GetImmediateInt()
			if len(imm) == 1 {
				immediate += fmt.Sprintf("%v", imm[0])
			} else {
				for i, val := range imm {
					if mask&(1<<i) == 0 {
						continue
					}
					immediate += fmt.Sprintf("%v", val)
					added += 1
					if i+1 < len(imm) && added < maskCount {
						immediate += ", "
					}
				}
			}
		default:
			immediate += o.parentOpcode.NumberType().ToString()
		}
		if len(o.Immediate) > 1 && maskCount > 1 {
			immediate += ")"
		}
		return immediate
	}

	var toReturn string
	if operandType == OPERAND_TYPE_CONSTANT_BUFFER {
		swizzle := o.SwizzleSrc()
		if swizzle[0] == -1 {
			panic("x swizzle should be set for constant buffer operand")
		}
		offset := uint32(o.Indices[1].Value*16) + uint32(swizzle[0])*4
		variable, err := cbs[o.Indices[0].Value].VariableFromOffset(offset)
		if err != nil {
			panic(err)
		}
		toReturn = variable.Name + variable.SwizzleFromSrc(swizzle, mask)
	} else if operandType == OPERAND_TYPE_IMMEDIATE_CONSTANT_BUFFER {
		toReturn = fmt.Sprintf("%v[floatBitsToInt(%v)]%v", operandType.ToGLSL(), o.Indices[0].ToGLSL(cbs, isg, osg, res, mask, true), o.SwizzleMask(mask))
	} else if operandType == OPERAND_TYPE_INPUT {
		toReturn = fmt.Sprintf("%v%v", isg[o.Indices[0].Value].NameWithIndex(), o.SwizzleMask(mask))
	} else if operandType == OPERAND_TYPE_OUTPUT {
		toReturn = fmt.Sprintf("%v%v", osg[o.Indices[0].Value].NameWithIndex(), o.SwizzleMask(mask))
	} else if operandType == OPERAND_TYPE_RESOURCE {
		var rb *ResourceBinding
		for i := range res {
			if res[i].InputType != TEXTURE {
				continue
			}
			if res[i].BindPoint == uint32(o.Indices[0].Value) {
				rb = &res[i]
				break
			}
		}
		if rb == nil {
			return fmt.Sprintf("resource binding %v not found", o.Indices[0].Value)
		}
		return rb.Name
	} else if operandType == OPERAND_TYPE_NULL {
		return "null"
	} else {
		toReturn = fmt.Sprintf("%v%v", operandType.ToGLSL(), o.Indices[0].ToGLSL(cbs, isg, osg, res, mask, true))
		if o.IndexDimension() >= OPERAND_INDEX_2D {
			toReturn += fmt.Sprintf("[%v]", o.Indices[1].ToGLSL(cbs, isg, osg, res, mask, true))
		}
		if o.IndexDimension() == OPERAND_INDEX_3D {
			toReturn += fmt.Sprintf("[%v]", o.Indices[2].ToGLSL(cbs, isg, osg, res, mask, true))
		}
		toReturn += o.SwizzleMask(mask)
	}
	if !isResult && o.parentOpcode.NumberType().BitcastFromFloat() != "" {
		toReturn = fmt.Sprintf("%v(%v)", o.parentOpcode.NumberType().BitcastFromFloat(), toReturn)
	}
	switch modifier {
	case OPERAND_MODIFIER_ABS:
		return fmt.Sprintf("abs(%v)", toReturn)
	case OPERAND_MODIFIER_ABSNEG:
		return fmt.Sprintf("-abs(%v)", toReturn)
	case OPERAND_MODIFIER_NEG:
		return fmt.Sprintf("-%v", toReturn)
	default:
		return toReturn
	}
}

type OperandIndex struct {
	Value          uint64
	Representation OPERAND_INDEX_REPRESENTATION
	Register       *Operand
}

func (o OperandIndex) ToGLSL(cbs []ConstantBuffer, isg, osg []Element, res []ResourceBinding, mask uint8, isResult bool) string {
	switch o.Representation {
	case OPERAND_INDEX_IMMEDIATE32, OPERAND_INDEX_IMMEDIATE64:
		return fmt.Sprintf("%v", o.Value)
	case OPERAND_INDEX_IMMEDIATE32_PLUS_RELATIVE, OPERAND_INDEX_IMMEDIATE64_PLUS_RELATIVE:
		return fmt.Sprintf("%v + %v", o.Value, o.Register.ToGLSL(cbs, isg, osg, res, mask, isResult))
	case OPERAND_INDEX_RELATIVE:
		return o.Register.ToGLSL(cbs, isg, osg, res, mask, isResult)
	}
	return "unknown operand index representation"
}
