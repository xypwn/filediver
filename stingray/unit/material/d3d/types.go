package d3d

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	d3dops "github.com/xypwn/filediver/stingray/unit/material/d3d/opcodes"
	"github.com/xypwn/filediver/util"
)

type RawVariableType struct {
	Class        d3dops.ShaderVariableClass
	Type         d3dops.ShaderVariableType
	Rows         uint16
	Cols         uint16
	Elements     uint16
	Members      uint16
	MemberOffset uint16
	_            [18]byte
	NameOffset   uint32
}

type RawVariable struct {
	NameOffset    uint32
	BufferOffset  uint32
	Size          uint32
	Flags         uint32 // D3D_SHADER_VARIABLE_FLAGS
	TypeOffset    uint32
	DefaultOffset uint32
	_             [16]byte
}

type RawConstantBuffer struct {
	NameOffset     uint32
	VariableCount  uint32
	VariableOffset uint32
	Size           uint32
	Flags          d3dops.ConstantBufferFlags
	Type           d3dops.ConstantBufferType
}

type RawResourceBinding struct {
	NameOffset    uint32
	InputType     d3dops.ShaderInputType
	ReturnType    d3dops.ShaderResourceReturnType
	ViewDimension d3dops.ShaderResourceViewDimension
	SampleCount   uint32
	BindPoint     uint32
	BindCount     uint32
	Flags         d3dops.ShaderInputFlags
}

type ShaderVersion struct {
	Minor uint8
	Major uint8
}

type RawOldElement struct {
	_             uint32
	NameOffset    uint32
	SemanticIndex uint32
	SystemValue   d3dops.SystemValueType
	ComponentType d3dops.RegisterComponentType
	Register      uint32
	Mask          uint8
	RWMask        uint8
	_             [2]byte
	_             uint32
}

type RawElement struct {
	NameOffset    uint32
	SemanticIndex uint32
	SystemValue   d3dops.SystemValueType
	ComponentType d3dops.RegisterComponentType
	Register      uint32
	Mask          uint8
	RWMask        uint8
	_             [2]byte
}

type ChunkHeader struct {
	Name [4]byte
	Size uint32
}

type RDEF struct {
	ChunkHeader
	ConstantBuffers  []d3dops.ConstantBuffer
	ResourceBindings []d3dops.ResourceBinding
	Version          ShaderVersion
	ProgramType      ShaderProgramType
	Flags            ShaderFlags
	Creator          string
}

func (r RDEF) Count(typ d3dops.ShaderInputType) uint32 {
	toReturn := uint32(0)
	for _, rb := range r.ResourceBindings {
		if rb.InputType == typ {
			toReturn++
		}
	}
	return toReturn
}

type ISG1 struct {
	ChunkHeader
	Elements []d3dops.Element
}

func (i ISG1) ToGLSL() string {
	toReturn := "// Input Signature\n"
	for _, element := range i.Elements {
		toReturn += fmt.Sprintf("%v\n", element.ToGLSL(true))
	}
	return toReturn + "\n"
}

type OSG1 struct {
	ChunkHeader
	Elements []d3dops.Element
}

func (o OSG1) ToGLSL() string {
	toReturn := "// Output Signature\n"
	for _, element := range o.Elements {
		toReturn += fmt.Sprintf("%v\n", element.ToGLSL(false))
	}
	return toReturn + "\n"
}

type SHEX struct {
	ChunkHeader
	Version     ShaderVersion
	ProgramType ShaderProgramType
	Opcodes     []d3dops.Opcode
}

type Chunk struct {
	ChunkHeader
	Data []byte
}

func variableFromRawVariable(r io.ReadSeeker, rawVar RawVariable) (*d3dops.Variable, error) {
	r.Seek(int64(rawVar.NameOffset), io.SeekStart)
	varName, err := util.ReadCString(r)
	if err != nil {
		return nil, fmt.Errorf("util.ReadCString: %v", err)
	}
	r.Seek(int64(rawVar.TypeOffset), io.SeekStart)
	var rawType RawVariableType
	if err := binary.Read(r, binary.LittleEndian, &rawType); err != nil {
		return nil, fmt.Errorf("read raw type: %v", err)
	}
	r.Seek(int64(rawType.NameOffset), io.SeekStart)
	typeName, err := util.ReadCString(r)
	if err != nil {
		return nil, fmt.Errorf("util.ReadCString: %v", err)
	}
	var defaultData []byte
	if rawVar.DefaultOffset != 0 {
		r.Seek(int64(rawVar.DefaultOffset), io.SeekStart)
		defaultData := make([]byte, rawVar.Size)
		if err := binary.Read(r, binary.LittleEndian, &defaultData); err != nil {
			return nil, fmt.Errorf("read default data: %v", err)
		}
	}
	return &d3dops.Variable{
		Name:         *varName,
		BufferOffset: rawVar.BufferOffset,
		Size:         rawVar.Size,
		Flags:        rawVar.Flags,
		DefaultData:  defaultData,
		VariableType: d3dops.VariableType{
			Class:        rawType.Class,
			Type:         rawType.Type,
			Rows:         rawType.Rows,
			Cols:         rawType.Cols,
			Elements:     rawType.Elements,
			Members:      rawType.Members,
			MemberOffset: rawType.MemberOffset,
			Name:         *typeName,
		},
	}, nil
}

func constantBufferFromRawConstantBuffer(r io.ReadSeeker, rawCBuf RawConstantBuffer) (*d3dops.ConstantBuffer, error) {
	r.Seek(int64(rawCBuf.NameOffset), io.SeekStart)
	cbName, err := util.ReadCString(r)
	if err != nil {
		return nil, fmt.Errorf("util.ReadCString: %v", err)
	}
	r.Seek(int64(rawCBuf.VariableOffset), io.SeekStart)
	rawVariables := make([]RawVariable, rawCBuf.VariableCount)
	if err := binary.Read(r, binary.LittleEndian, &rawVariables); err != nil {
		return nil, err
	}

	variables := make([]d3dops.Variable, 0)
	for _, rawVar := range rawVariables {
		variable, err := variableFromRawVariable(r, rawVar)
		if err != nil {
			return nil, fmt.Errorf("variableFromRawVariable: %v", err)
		}
		variables = append(variables, *variable)
	}
	return &d3dops.ConstantBuffer{
		Name:      *cbName,
		Variables: variables,
		Size:      rawCBuf.Size,
		Flags:     rawCBuf.Flags,
		Type:      rawCBuf.Type,
	}, nil
}

func RDEFFromChunk(chunk *Chunk) (*RDEF, error) {
	r := bytes.NewReader(chunk.Data)
	var constantBufferCount uint32
	if err := binary.Read(r, binary.LittleEndian, &constantBufferCount); err != nil {
		return nil, err
	}
	var constantBufferArrayOffset uint32
	if err := binary.Read(r, binary.LittleEndian, &constantBufferArrayOffset); err != nil {
		return nil, err
	}
	var resourceBindingCount uint32
	if err := binary.Read(r, binary.LittleEndian, &resourceBindingCount); err != nil {
		return nil, err
	}
	var resourceBindingArrayOffset uint32
	if err := binary.Read(r, binary.LittleEndian, &resourceBindingArrayOffset); err != nil {
		return nil, err
	}
	var version ShaderVersion
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	var programType ShaderProgramType
	if err := binary.Read(r, binary.LittleEndian, &programType); err != nil {
		return nil, err
	}
	var flags ShaderFlags
	if err := binary.Read(r, binary.LittleEndian, &flags); err != nil {
		return nil, err
	}
	var creatorOffset uint32
	if err := binary.Read(r, binary.LittleEndian, &creatorOffset); err != nil {
		return nil, err
	}

	r.Seek(int64(constantBufferArrayOffset), io.SeekStart)
	rawConstantBuffers := make([]RawConstantBuffer, constantBufferCount)
	if err := binary.Read(r, binary.LittleEndian, &rawConstantBuffers); err != nil {
		return nil, err
	}
	constantBuffers := make([]d3dops.ConstantBuffer, 0)
	for _, rawCBuf := range rawConstantBuffers {
		cbuf, err := constantBufferFromRawConstantBuffer(r, rawCBuf)
		if err != nil {
			return nil, fmt.Errorf("constantBufferFromRawConstantBuffer: %v", err)
		}
		constantBuffers = append(constantBuffers, *cbuf)
	}

	r.Seek(int64(resourceBindingArrayOffset), io.SeekStart)
	rawResourceBindings := make([]RawResourceBinding, resourceBindingCount)
	if err := binary.Read(r, binary.LittleEndian, &rawResourceBindings); err != nil {
		return nil, err
	}
	resourceBindings := make([]d3dops.ResourceBinding, 0)
	for _, rawRB := range rawResourceBindings {
		r.Seek(int64(rawRB.NameOffset), io.SeekStart)
		rbName, err := util.ReadCString(r)
		if err != nil {
			return nil, fmt.Errorf("util.ReadCString: %v", err)
		}
		resourceBindings = append(resourceBindings, d3dops.ResourceBinding{
			Name:          *rbName,
			InputType:     rawRB.InputType,
			ReturnType:    rawRB.ReturnType,
			ViewDimension: rawRB.ViewDimension,
			SampleCount:   rawRB.SampleCount,
			BindPoint:     rawRB.BindPoint,
			BindCount:     rawRB.BindCount,
			Flags:         rawRB.Flags,
		})
	}

	r.Seek(int64(creatorOffset), io.SeekStart)
	creatorName, err := util.ReadCString(r)
	if err != nil {
		return nil, fmt.Errorf("util.ReadCString: %v", err)
	}
	return &RDEF{
		ChunkHeader:      chunk.ChunkHeader,
		ConstantBuffers:  constantBuffers,
		ResourceBindings: resourceBindings,
		Version:          version,
		ProgramType:      programType,
		Flags:            flags,
		Creator:          *creatorName,
	}, nil
}

func SHEXFromChunk(chunk *Chunk) (*SHEX, error) {
	r := bytes.NewReader(chunk.Data)
	var version uint8
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	// Skip a byte
	r.Seek(1, io.SeekCurrent)
	var programType ShaderProgramType
	if err := binary.Read(r, binary.LittleEndian, &programType); err != nil {
		return nil, err
	}
	var dwordCount uint32
	if err := binary.Read(r, binary.LittleEndian, &dwordCount); err != nil {
		return nil, err
	}
	opcodes := make([]d3dops.Opcode, 0)
	for {
		// offset, _ := r.Seek(0, io.SeekCurrent)
		// fmt.Printf("SHEX offset 0x%08x: ", offset)
		opcode, err := d3dops.ParseOpcode(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("d3dops.ParseOpcode: %v", err)
		}
		opcodes = append(opcodes, opcode)
	}
	return &SHEX{
		ChunkHeader: chunk.ChunkHeader,
		Version: ShaderVersion{
			Minor: version & 0x0F,
			Major: version >> 4,
		},
		ProgramType: programType,
		Opcodes:     opcodes,
	}, nil
}

func ISG1FromChunk(chunk *Chunk) (*ISG1, error) {
	r := bytes.NewReader(chunk.Data)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	var elementArrayOffset uint32
	if err := binary.Read(r, binary.LittleEndian, &elementArrayOffset); err != nil {
		return nil, err
	}

	r.Seek(int64(elementArrayOffset), io.SeekStart)
	rawElements := make([]RawOldElement, count)
	if err := binary.Read(r, binary.LittleEndian, &rawElements); err != nil {
		return nil, err
	}

	elements := make([]d3dops.Element, 0)
	for _, rawElem := range rawElements {
		r.Seek(int64(rawElem.NameOffset), io.SeekStart)
		name, err := util.ReadCString(r)
		if err != nil {
			return nil, fmt.Errorf("util.ReadCString: %v", err)
		}
		elements = append(elements, d3dops.Element{
			Name:          *name,
			SemanticIndex: rawElem.SemanticIndex,
			SystemValue:   rawElem.SystemValue,
			ComponentType: rawElem.ComponentType,
			Register:      rawElem.Register,
			Mask:          rawElem.Mask,
			RWMask:        rawElem.RWMask,
		})
	}
	return &ISG1{
		ChunkHeader: chunk.ChunkHeader,
		Elements:    elements,
	}, nil
}

func OSG1FromChunk(chunk *Chunk) (*OSG1, error) {
	isg1, err := ISG1FromChunk(chunk)
	if err != nil {
		return nil, err
	}
	return &OSG1{
		ChunkHeader: isg1.ChunkHeader,
		Elements:    isg1.Elements,
	}, nil
}

func ISGNFromChunk(chunk *Chunk) (*ISG1, error) {
	r := bytes.NewReader(chunk.Data)
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	var elementArrayOffset uint32
	if err := binary.Read(r, binary.LittleEndian, &elementArrayOffset); err != nil {
		return nil, err
	}

	r.Seek(int64(elementArrayOffset), io.SeekStart)
	rawElements := make([]RawElement, count)
	if err := binary.Read(r, binary.LittleEndian, &rawElements); err != nil {
		return nil, err
	}

	elements := make([]d3dops.Element, 0)
	for _, rawElem := range rawElements {
		r.Seek(int64(rawElem.NameOffset), io.SeekStart)
		name, err := util.ReadCString(r)
		if err != nil {
			return nil, fmt.Errorf("util.ReadCString: %v", err)
		}
		elements = append(elements, d3dops.Element{
			Name:          *name,
			SemanticIndex: rawElem.SemanticIndex,
			SystemValue:   rawElem.SystemValue,
			ComponentType: rawElem.ComponentType,
			Register:      rawElem.Register,
			Mask:          rawElem.Mask,
			RWMask:        rawElem.RWMask,
		})
	}
	return &ISG1{
		ChunkHeader: chunk.ChunkHeader,
		Elements:    elements,
	}, nil
}

func OSGNFromChunk(chunk *Chunk) (*OSG1, error) {
	isg1, err := ISGNFromChunk(chunk)
	if err != nil {
		return nil, err
	}
	return &OSG1{
		ChunkHeader: isg1.ChunkHeader,
		Elements:    isg1.Elements,
	}, nil
}

func ParseChunk(r io.Reader) (*Chunk, error) {
	var header ChunkHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("read header: %v", err)
	}
	// fmt.Printf("chunk header:\n    name: %v\n    size: %v\n", string(header.Name[:]), header.Size)
	data := make([]uint8, header.Size)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return nil, fmt.Errorf("read data: %v", err)
	}
	return &Chunk{
		ChunkHeader: header,
		Data:        data,
	}, nil
}
