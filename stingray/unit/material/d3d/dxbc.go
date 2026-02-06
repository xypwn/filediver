package d3d

import (
	"encoding/binary"
	"fmt"
	"io"

	d3dops "github.com/xypwn/filediver/stingray/unit/material/d3d/opcodes"
)

var MAGIC = [4]byte{'D', 'X', 'B', 'C'}

type DXBCHeader struct {
	Magic      [4]byte
	Digest     [16]byte
	Major      uint16
	Minor      uint16
	Size       uint32
	ChunkCount uint32
}

type DXBC struct {
	DXBCHeader
	ResourceDefinitions RDEF
	InputSignature      ISG1
	OutputSignature     OSG1
	ShaderCode          SHEX
	Chunks              []Chunk
}

func (dxbc *DXBC) Serialize() ([]byte, error) {
	data, err := binary.Append(nil, binary.LittleEndian, dxbc.DXBCHeader)
	if err != nil {
		return nil, err
	}
	chunkOffsets := make([]uint32, len(dxbc.Chunks))
	offset := uint32(len(data) + binary.Size(chunkOffsets))

	for i, chunk := range dxbc.Chunks {
		chunkOffsets[i] = offset
		offset = offset + uint32(binary.Size(chunk.ChunkHeader)+len(chunk.Data))
	}
	data, err = binary.Append(data, binary.LittleEndian, chunkOffsets)
	if err != nil {
		return nil, err
	}
	for i, chunk := range dxbc.Chunks {
		if len(data) != int(chunkOffsets[i]) {
			return nil, fmt.Errorf("incorrect chunk offset! %v != %v", len(data), int(chunkOffsets[i]))
		}
		data, err = binary.Append(data, binary.LittleEndian, chunk.ChunkHeader)
		if err != nil {
			return nil, err
		}
		data = append(data, chunk.Data...)
	}

	return data, nil
}

func ParseDXBC(r io.ReadSeeker) (*DXBC, error) {
	startOffset, _ := r.Seek(0, io.SeekCurrent)
	var header DXBCHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("read header: %v", err)
	}
	if header.Magic != MAGIC {
		return nil, fmt.Errorf("not a DXBC file")
	}

	chunkOffsets := make([]uint32, header.ChunkCount)
	if err := binary.Read(r, binary.LittleEndian, &chunkOffsets); err != nil {
		return nil, fmt.Errorf("read chunk offsets: %v", err)
	}

	toReturn := &DXBC{
		DXBCHeader: header,
		Chunks:     make([]Chunk, 0),
	}

	for _, offset := range chunkOffsets {
		_, err := r.Seek(startOffset+int64(offset), io.SeekStart)
		//fmt.Printf("Seeking offset 0x%08x\n", startOffset+int64(offset))
		if err != nil {
			return nil, fmt.Errorf("seek offset 0x%08x: %v", startOffset+int64(offset), err)
		}
		chunk, err := ParseChunk(r)
		if err != nil {
			return nil, fmt.Errorf("ParseChunk: %v", err)
		}
		switch chunk.Name {
		case [4]byte{'R', 'D', 'E', 'F'}:
			rdef, err := RDEFFromChunk(chunk)
			if err != nil {
				return nil, fmt.Errorf("RDEFFromChunk: %v", err)
			}
			toReturn.ResourceDefinitions = *rdef
		case [4]byte{'S', 'H', 'E', 'X'}:
			shex, err := SHEXFromChunk(chunk)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				toReturn.Chunks = append(toReturn.Chunks, *chunk)
				continue
			}
			toReturn.ShaderCode = *shex
		case [4]byte{'I', 'S', 'G', '1'}:
			isg1, err := ISG1FromChunk(chunk)
			if err != nil {
				return nil, fmt.Errorf("ISG1FromChunk: %v", err)
			}
			toReturn.InputSignature = *isg1
		case [4]byte{'I', 'S', 'G', 'N'}:
			isg1, err := ISGNFromChunk(chunk)
			if err != nil {
				return nil, fmt.Errorf("ISGNFromChunk: %v", err)
			}
			toReturn.InputSignature = *isg1
		case [4]byte{'O', 'S', 'G', '1'}:
			osg1, err := OSG1FromChunk(chunk)
			if err != nil {
				return nil, fmt.Errorf("OSG1FromChunk: %v", err)
			}
			toReturn.OutputSignature = *osg1
		case [4]byte{'O', 'S', 'G', 'N'}:
			osg1, err := OSGNFromChunk(chunk)
			if err != nil {
				return nil, fmt.Errorf("OSGNFromChunk: %v", err)
			}
			toReturn.OutputSignature = *osg1
		}
		toReturn.Chunks = append(toReturn.Chunks, *chunk)
	}
	return toReturn, nil
}

func (d *DXBC) ToGLSL() string {
	d3dops.EncounteredBFI = false
	d3dops.CreatedTemp = false
	toReturn := "#version 420 core\n\n"

	for i, cbuf := range d.ResourceDefinitions.ConstantBuffers {
		toReturn += cbuf.ToGLSL(i)
		toReturn += "\n\n"
	}

	for _, rbind := range d.ResourceDefinitions.ResourceBindings {
		toReturn += rbind.ToGLSL()
	}

	toReturn += d.InputSignature.ToGLSL()
	toReturn += d.OutputSignature.ToGLSL()

	toReturn += fmt.Sprintf("// Program type: %v\n", d.ShaderCode.ProgramType.ToString())

	toReturn += "void main() {\n"
	for _, opcode := range d.ShaderCode.Opcodes {
		toReturn += getGLSL(opcode, d.ResourceDefinitions.ConstantBuffers, d.InputSignature.Elements, d.OutputSignature.Elements, d.ResourceDefinitions.ResourceBindings)
	}
	toReturn += "}\n"
	return toReturn
}

func getGLSL(opcode d3dops.Opcode, cbs []d3dops.ConstantBuffer, isg, osg []d3dops.Element, res []d3dops.ResourceBinding) string {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println(r)
	// 	}
	// }()
	return opcode.ToGLSL(cbs, isg, osg, res)
}
