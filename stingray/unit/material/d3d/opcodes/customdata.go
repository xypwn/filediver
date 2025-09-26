package d3dops

import (
	"encoding/binary"
	"fmt"
)

type ShaderOpcodeCustomDataClass uint32

const (
	CUSTOMDATA_COMMENT ShaderOpcodeCustomDataClass = iota
	CUSTOMDATA_DEBUGINFO
	CUSTOMDATA_OPAQUE
	CUSTOMDATA_DCL_IMMEDIATE_CONSTANT_BUFFER
	CUSTOMDATA_SHADER_MESSAGE
	CUSTOMDATA_SHADER_CLIP_PLANE_CONSTANT_MAPPINGS_FOR_DX9
)

type CustomDataDCLImmediateConstantBuffer struct {
	Class     ShaderOpcodeCustomDataClass
	Constants [][4]float32
}

func (icb *CustomDataDCLImmediateConstantBuffer) ToGLSL(_ []ConstantBuffer, _, _ []Element, _ []ResourceBinding) string {
	toReturn := fmt.Sprintf("vec4 icb[%v];\n", len(icb.Constants))
	for i, constant := range icb.Constants {
		toReturn += fmt.Sprintf("icb[%v] = vec4(%v, %v, %v, %v);\n", i, constant[0], constant[1], constant[2], constant[3])
	}
	return toReturn
}

func ParseCustomData(class ShaderOpcodeCustomDataClass, data []byte) (Opcode, error) {
	if len(data)%16 != 0 {
		return nil, fmt.Errorf("data length was not divisible by 16")
	}
	constants := make([][4]float32, 0)
	offset := 0
	for i := 0; i < len(data)/16; i++ {
		var tuple [4]float32
		consumed, err := binary.Decode(data[offset:], binary.LittleEndian, &tuple)
		if err != nil {
			return nil, err
		}
		offset += consumed
		constants = append(constants, tuple)
	}
	return &CustomDataDCLImmediateConstantBuffer{
		Class:     class,
		Constants: constants,
	}, nil
}
