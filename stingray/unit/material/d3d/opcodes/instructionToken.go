package d3dops

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type InstructionToken struct {
	opcode        uint32
	extensions    []uint32
	functionIndex uint32
	operands      []Operand
}

const SATURATE_MASK uint32 = 0x00002000

func (tok *InstructionToken) Saturate() bool {
	return tok.opcode&SATURATE_MASK != 0
}

func (tok *InstructionToken) ToGLSL(cbs []ConstantBuffer, isg, osg []Element, res []ResourceBinding) string {
	opType := ShaderOpcodeType(tok.opcode & TYPE_MASK)
	switch len(tok.operands) {
	case 2:
		return tok.unaryOpGLSL(opType, cbs, isg, osg, res)
	case 3:
		return tok.binaryOpGLSL(opType, cbs, isg, osg, res)
	case 4:
		return tok.trinaryOpGLSL(opType, cbs, isg, osg, res)
	default:
		return fmt.Sprintf("// %v instruction not implemented! %v extensions, %v operands\n", opType.ToString(), len(tok.extensions), len(tok.operands))
	}
}

func (tok *InstructionToken) unaryOpGLSL(opType ShaderOpcodeType, cbs []ConstantBuffer, isg, osg []Element, res []ResourceBinding) string {
	toReturn := fmt.Sprintf("// Unary Op %v\n", opType.ToString())
	masks := [2]uint8{tok.operands[0].Mask(), tok.operands[0].Mask()}
	var expr string
	switch opType {
	case OPCODE_11_DERIV_RTX_COARSE, OPCODE_11_DERIV_RTX_FINE, OPCODE_DERIV_RTX:
		expr = fmt.Sprintf(
			"dFdx(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_11_DERIV_RTY_COARSE, OPCODE_11_DERIV_RTY_FINE, OPCODE_DERIV_RTY:
		expr = fmt.Sprintf(
			"dFdy(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_FTOU:
		count := tok.operands[0].MaskCount()
		var constructor string
		if count == 1 {
			constructor = "uint"
		} else {
			constructor = fmt.Sprintf("uvec%v", count)
		}

		expr = fmt.Sprintf(
			"%v(%v)",
			constructor,
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_FTOI:
		count := tok.operands[0].MaskCount()
		var constructor string
		if count == 1 {
			constructor = "int"
		} else {
			constructor = fmt.Sprintf("ivec%v", count)
		}

		expr = fmt.Sprintf(
			"%v(%v)",
			constructor,
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_UTOF, OPCODE_ITOF:
		count := tok.operands[0].MaskCount()
		var constructor string
		if count == 1 {
			constructor = "float"
		} else {
			constructor = fmt.Sprintf("vec%v", count)
		}

		expr = fmt.Sprintf(
			"%v(%v)",
			constructor,
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_RSQ:
		expr = fmt.Sprintf(
			"inversesqrt(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_SQRT:
		expr = fmt.Sprintf(
			"sqrt(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_MOV:
		expr = fmt.Sprintf(
			"%v",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_11_RCP:
		expr = fmt.Sprintf(
			"1.0 / %v",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_LOG:
		expr = fmt.Sprintf(
			"log2(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_EXP:
		expr = fmt.Sprintf(
			"exp2(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_FRC:
		expr = fmt.Sprintf(
			"fract(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_ROUND_NI:
		expr = fmt.Sprintf(
			"floor(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_ROUND_NE:
		expr = fmt.Sprintf(
			"roundEven(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_ROUND_PI:
		expr = fmt.Sprintf(
			"ceil(%v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	case OPCODE_ROUND_Z:
		operand := tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false)
		expr = fmt.Sprintf(
			"floor(abs(%v)) * sign(%v)",
			operand,
			operand,
		)
	default:
		expr = fmt.Sprintf(
			"%v /* Unimplemented */",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
		)
	}

	if tok.Saturate() {
		expr = fmt.Sprintf("clamp(%v, 0.0, 1.0)", expr)
	}
	if opType.ReturnNumberType() != internalNumberTypeFloat {
		expr = fmt.Sprintf("%v(%v)", opType.ReturnNumberType().BitcastToFloat(), expr)
	}

	toReturn += fmt.Sprintf(
		"%v = %v;",
		tok.operands[0].ToGLSL(cbs, isg, osg, res, masks[0], true),
		expr,
	)

	return toReturn + "\n"
}

func (tok *InstructionToken) binaryOpGLSL(opType ShaderOpcodeType, cbs []ConstantBuffer, isg, osg []Element, res []ResourceBinding) string {
	toReturn := fmt.Sprintf("// Binary Op %v mask 0x%x:\n", opType.ToString(), tok.operands[0].Mask())
	if tok.operands[0].ComponentSelectionMode() != OPERAND_4_COMPONENT_MASK_MODE {
		toReturn += "// ------------ operand 0 was not in mask mode! ------------\n"
	}
	masks := [3]uint8{tok.operands[0].Mask(), tok.operands[0].Mask(), tok.operands[0].Mask()}
	// Binary op mask special cases
	switch opType {
	case OPCODE_DP2:
		masks[1] = 0x3
		masks[2] = 0x3
	case OPCODE_DP3:
		masks[1] = 0x7
		masks[2] = 0x7
	case OPCODE_DP4:
		masks[1] = 0xf
		masks[2] = 0xf
	}
	// for i, operand := range tok.operands {
	// 	// toReturn += operand.ToGLSL(cbs, tok.operands[0].Mask())
	// 	// if i+1 < len(tok.operands) {
	// 	// 	toReturn += " "
	// 	// }
	// 	toReturn += operand.ToGLSL(cbs, masks[i]) + "\n"
	// 	toReturn += operand.ToString()
	// }

	var expr string
	switch opType {
	case OPCODE_DP2, OPCODE_DP3, OPCODE_DP4:
		expr = fmt.Sprintf(
			"dot(%v, %v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
			tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
		)
	case OPCODE_MAX, OPCODE_IMAX, OPCODE_UMAX, OPCODE_11_ATOMIC_IMAX, OPCODE_11_ATOMIC_UMAX, OPCODE_11_DMAX, OPCODE_11_IMM_ATOMIC_IMAX, OPCODE_11_IMM_ATOMIC_UMAX:
		expr = fmt.Sprintf(
			"max(%v, %v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
			tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
		)
	case OPCODE_MIN, OPCODE_IMIN, OPCODE_UMIN, OPCODE_11_ATOMIC_IMIN, OPCODE_11_ATOMIC_UMIN, OPCODE_11_DMIN, OPCODE_11_IMM_ATOMIC_IMIN, OPCODE_11_IMM_ATOMIC_UMIN:
		expr = fmt.Sprintf(
			"min(%v, %v)",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
			tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
		)
	case OPCODE_LD:
		// ld dest[.mask], srcAddress[.swizzle], srcResource[.swizzle]
		// gvec texelFetch(gsampler sampler​, ivec texCoord);
		expr = fmt.Sprintf(
			"texelFetch(%v, floatBitsToInt(%v))%v",
			tok.operands[2].ToGLSL(cbs, isg, osg, res, 0x0, true),
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], true),
			tok.operands[2].SwizzleMask(masks[2]),
		)
	case OPCODE_RESINFO:
		expr = fmt.Sprintf(
			"textureSize(%v, %v)%v",
			tok.operands[2].ToGLSL(cbs, isg, osg, res, 0x0, true),
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], true),
			tok.operands[2].SwizzleMask(masks[2]),
		)
	default:
		expr = fmt.Sprintf(
			"%v %v %v",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
			opType.ToOperator(),
			tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
		)
	}
	if tok.Saturate() {
		expr = fmt.Sprintf("clamp(%v, 0.0, 1.0)", expr)
	}
	if opType.ReturnNumberType() != internalNumberTypeFloat && opType != OPCODE_LD {
		expr = fmt.Sprintf("%v(%v)", opType.ReturnNumberType().BitcastToFloat(), expr)
	}

	toReturn += fmt.Sprintf(
		"%v = %v;",
		tok.operands[0].ToGLSL(cbs, isg, osg, res, masks[0], true),
		expr,
	)

	return toReturn + "\n"
}

func (tok *InstructionToken) trinaryOpGLSL(opType ShaderOpcodeType, cbs []ConstantBuffer, isg, osg []Element, res []ResourceBinding) string {
	toReturn := fmt.Sprintf("// Trinary Op %v\n", opType.ToString())
	masks := [4]uint8{tok.operands[0].Mask(), tok.operands[0].Mask(), tok.operands[0].Mask(), tok.operands[0].Mask()}
	var expr string
	switch opType {
	case OPCODE_MAD, OPCODE_IMAD, OPCODE_UMAD:
		expr = fmt.Sprintf(
			"%v * %v + %v",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
			tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
			tok.operands[3].ToGLSL(cbs, isg, osg, res, masks[3], false),
		)
	case OPCODE_UDIV:
		if tok.operands[0].OperandToken0.Type() != OPERAND_TYPE_NULL {
			masks[0] = tok.operands[0].Mask()
			masks[2] = tok.operands[0].Mask()
			masks[3] = tok.operands[0].Mask()
			expr = fmt.Sprintf(
				"%v / %v",
				tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
				tok.operands[3].ToGLSL(cbs, isg, osg, res, masks[3], false),
			)
			if tok.Saturate() {
				expr = fmt.Sprintf("clamp(%v, 0.0, 1.0)", expr)
			}
			if opType.ReturnNumberType() != internalNumberTypeFloat {
				expr = fmt.Sprintf("%v(%v)", opType.ReturnNumberType().BitcastToFloat(), expr)
			}
			toReturn += fmt.Sprintf(
				"%v = %v;\n",
				tok.operands[0].ToGLSL(cbs, isg, osg, res, masks[0], true),
				expr,
			)
		}
		if tok.operands[1].OperandToken0.Type() != OPERAND_TYPE_NULL {
			masks[1] = tok.operands[1].Mask()
			masks[2] = tok.operands[1].Mask()
			masks[3] = tok.operands[1].Mask()
			expr = fmt.Sprintf(
				"%v %% %v",
				tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
				tok.operands[3].ToGLSL(cbs, isg, osg, res, masks[3], false),
			)
			if tok.Saturate() {
				expr = fmt.Sprintf("clamp(%v, 0.0, 1.0)", expr)
			}
			if opType.ReturnNumberType() != internalNumberTypeFloat {
				expr = fmt.Sprintf("%v(%v)", opType.ReturnNumberType().BitcastToFloat(), expr)
			}
			toReturn += fmt.Sprintf(
				"%v = %v;\n",
				tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], true),
				expr,
			)
		}
		return toReturn
	case OPCODE_MOVC:
		expr = fmt.Sprintf(
			"bool(%v) ? %v : %v",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
			tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
			tok.operands[3].ToGLSL(cbs, isg, osg, res, masks[3], false),
		)
	//case OPCODE_SAMPLE:

	default:
		expr = fmt.Sprintf(
			"%v %v %v /* Unimplemented */",
			tok.operands[1].ToGLSL(cbs, isg, osg, res, masks[1], false),
			tok.operands[2].ToGLSL(cbs, isg, osg, res, masks[2], false),
			tok.operands[3].ToGLSL(cbs, isg, osg, res, masks[3], false),
		)
	}

	if tok.Saturate() {
		expr = fmt.Sprintf("clamp(%v, 0.0, 1.0)", expr)
	}
	if opType.ReturnNumberType() != internalNumberTypeFloat {
		expr = fmt.Sprintf("%v(%v)", opType.ReturnNumberType().BitcastToFloat(), expr)
	}

	toReturn += fmt.Sprintf(
		"%v = %v;",
		tok.operands[0].ToGLSL(cbs, isg, osg, res, masks[0], true),
		expr,
	)

	return toReturn + "\n"
}

func ParseInstruction(opcode uint32, data []uint8) (Opcode, error) {
	r := bytes.NewReader(data)
	opType := ShaderOpcodeType(opcode & TYPE_MASK)
	extended := IsExtended(opcode)
	extensions := make([]uint32, 0)
	for extended {
		var extendedToken uint32
		if err := binary.Read(r, binary.LittleEndian, &extendedToken); err != nil {
			return nil, err
		}
		extensions = append(extensions, extendedToken)
		extended = IsExtended(extendedToken)
	}

	var functionIndex uint32
	if opType == OPCODE_11_INTERFACE_CALL {
		if err := binary.Read(r, binary.LittleEndian, &functionIndex); err != nil {
			return nil, err
		}
	}

	operands := make([]Operand, 0)
	for {
		operand, err := ParseOperand(r, opType)
		if err != nil {
			break
		}
		operands = append(operands, *operand)
	}
	return &InstructionToken{
		opcode:        opcode,
		extensions:    extensions,
		functionIndex: functionIndex,
		operands:      operands,
	}, nil
}
