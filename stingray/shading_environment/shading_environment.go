package shading_environment

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type VariableContainerType uint32

const (
	VariableContainerType_Scalar VariableContainerType = iota
	VariableContainerType_Vec2
	VariableContainerType_Vec3
	VariableContainerType_Vec4
	VariableContainerType_Matrix
	VariableContainerType_Unknown1
	VariableContainerType_Vec8
	VariableContainerType_Unknown2
	VariableContainerType_Unknown3
	VariableContainerType_Unknown4
	VariableContainerType_Hash
)

func (p VariableContainerType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=VariableContainerType

type VariableInfo1 struct {
	ContainerType VariableContainerType
	Unk1          uint32
	Name          stingray.ThinHash
	Offset        uint32
}

type VariableInfo2 struct {
	Index         uint32
	Index2        uint32
	ContainerType VariableContainerType
	Name          stingray.ThinHash
}

type rawDefaultData struct {
	Name        stingray.ThinHash
	_           [4]uint8
	DataLength  uint32
	DataOffset  uint32
	_           [8]uint8
	FloatCount  uint32
	FloatOffset uint32
}

type rawShadingEnvironment struct {
	Magic               uint32
	_                   [4]uint8
	Material            stingray.Hash
	VariableInfo1Count  uint32
	VariableInfo1Offset uint32
	_                   [8]uint8
	DefaultDataCount    uint32
	DefaultDataOffset   uint32
	_                   [8]uint8
	UnknownCount        uint32
	UnknownOffset       uint32
	_                   [8]uint8
	VariableInfo2Count  uint32
	VariableInfo2Offset uint32
}

type Variable struct {
	Type VariableContainerType
	Name stingray.ThinHash
	Data any
}

type ShadingEnvironment struct {
	Material       stingray.Hash
	Variables      []Variable
	VariableInfos1 []VariableInfo1
	VariableInfos2 []VariableInfo2
}

type ShaderVariableMapping struct {
	UnkInt1      uint32
	UnkInt2      uint32
	VariableName stingray.ThinHash
	UnkInt3      uint32
	UnkInt4      uint32
}

type EntitySettingMapping struct {
	UnkInt1      uint16
	Index        uint16
	SettingName  stingray.ThinHash
	SettingName2 stingray.ThinHash
	SettingName3 stingray.ThinHash
	SettingName4 stingray.ThinHash
}

type rawShadingEnvironmentMapping struct {
	Mapping1Count  uint32
	Mapping1Offset uint32
	HashesCount    uint32
	HashesOffset   uint32
	Mapping2Count  uint32
	Mapping2Offset uint32
}

type ShadingEnvironmentMapping struct {
	ShaderVariableMappings []ShaderVariableMapping
	TextureNames           []stingray.ThinHash
	EntitySettingMappings  []EntitySettingMapping
}

func LoadShadingEnvironment(r io.ReadSeeker) (*ShadingEnvironment, error) {
	var environment rawShadingEnvironment
	if err := binary.Read(r, binary.LittleEndian, &environment); err != nil {
		return nil, err
	}

	variableInfos1 := make([]VariableInfo1, environment.VariableInfo1Count)
	if _, err := r.Seek(int64(environment.VariableInfo1Offset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, variableInfos1); err != nil {
		return nil, err
	}

	variableInfos2 := make([]VariableInfo2, environment.VariableInfo2Count)
	if _, err := r.Seek(int64(environment.VariableInfo2Offset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, variableInfos2); err != nil {
		return nil, err
	}

	defaultData := make([]rawDefaultData, environment.DefaultDataCount)
	if _, err := r.Seek(int64(environment.DefaultDataOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, defaultData); err != nil {
		return nil, err
	}

	if environment.UnknownCount > 0 {
		fmt.Printf("shading_environment: unknown count was %v\n\n", environment.UnknownCount)
	}
	if environment.VariableInfo1Count != environment.VariableInfo2Count {
		return nil, fmt.Errorf("Unexpected mismatch between variable info counts in shading_environment: %v != %v", environment.VariableInfo1Count, environment.VariableInfo2Count)
	}

	variables := make([]Variable, 0)
	for i := range environment.VariableInfo1Count {
		variable := Variable{
			Type: variableInfos1[i].ContainerType,
			Name: variableInfos1[i].Name,
		}
		if _, err := r.Seek(int64(defaultData[0].DataOffset+variableInfos1[i].Offset), io.SeekStart); err != nil {
			return nil, err
		}
		switch variable.Type {
		case VariableContainerType_Scalar:
			var temp float32
			if err := binary.Read(r, binary.LittleEndian, &temp); err != nil {
				return nil, err
			}
			variable.Data = temp
		case VariableContainerType_Vec2:
			temp := make([]float32, 2)
			if err := binary.Read(r, binary.LittleEndian, temp); err != nil {
				return nil, err
			}
			variable.Data = temp
		case VariableContainerType_Vec3:
			temp := make([]float32, 3)
			if err := binary.Read(r, binary.LittleEndian, temp); err != nil {
				return nil, err
			}
			variable.Data = temp
		case VariableContainerType_Vec4:
			temp := make([]float32, 4)
			if err := binary.Read(r, binary.LittleEndian, temp); err != nil {
				return nil, err
			}
			variable.Data = temp
		case VariableContainerType_Vec8:
			temp := make([]float32, 8)
			if err := binary.Read(r, binary.LittleEndian, temp); err != nil {
				return nil, err
			}
			variable.Data = temp
		case VariableContainerType_Matrix:
			temp := make([]float32, 16)
			if err := binary.Read(r, binary.LittleEndian, temp); err != nil {
				return nil, err
			}
			variable.Data = temp
		case VariableContainerType_Hash:
			var temp stingray.Hash
			if err := binary.Read(r, binary.LittleEndian, &temp); err != nil {
				return nil, err
			}
			variable.Data = temp
		}
		variables = append(variables, variable)
	}

	return &ShadingEnvironment{
		Material:       environment.Material,
		Variables:      variables,
		VariableInfos1: variableInfos1,
		VariableInfos2: variableInfos2,
	}, nil
}

func LoadShadingEnvironmentMapping(r io.ReadSeeker) (*ShadingEnvironmentMapping, error) {
	var mapping rawShadingEnvironmentMapping
	if err := binary.Read(r, binary.LittleEndian, &mapping); err != nil {
		return nil, err
	}

	mappings1 := make([]ShaderVariableMapping, mapping.Mapping1Count)
	if _, err := r.Seek(int64(mapping.Mapping1Offset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, mappings1); err != nil {
		return nil, err
	}

	hashes := make([]stingray.ThinHash, mapping.HashesCount)
	if _, err := r.Seek(int64(mapping.HashesOffset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, hashes); err != nil {
		return nil, err
	}

	mappings2 := make([]EntitySettingMapping, mapping.Mapping2Count)
	if _, err := r.Seek(int64(mapping.Mapping2Offset), io.SeekStart); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, mappings2); err != nil {
		return nil, err
	}

	return &ShadingEnvironmentMapping{
		ShaderVariableMappings: mappings1,
		TextureNames:           hashes,
		EntitySettingMappings:  mappings2,
	}, nil
}
