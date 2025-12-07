package stingray

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	// Returned when the file exists, but doesn't have the
	// requested data type.
	ErrFileDataTypeNotExist = errors.New("file data type doesn't exist")
	// Returned when the file doesn't exist.
	ErrFileNotExist = errors.New("file doesn't exist")
)

type DataType int

const (
	DataMain DataType = iota
	DataStream
	DataGPU
	NumDataType
)

func (t DataType) String() string {
	switch t {
	case DataMain:
		return "main"
	case DataStream:
		return "stream"
	case DataGPU:
		return "GPU"
	default:
		panic("unhandled case")
	}
}

func (t DataType) ArchiveFileExtension() string {
	switch t {
	case DataMain:
		return ""
	case DataStream:
		return ".stream"
	case DataGPU:
		return ".gpu_resources"
	default:
		panic("unhandled case")
	}
}

type FileID struct {
	Name Hash
	Type Hash
}

func NewFileID(name Hash, typ Hash) FileID {
	return FileID{Name: name, Type: typ}
}

// Cmp compares two file IDs with order: name > type.
func (id FileID) Cmp(other FileID) int {
	if r := id.Name.Cmp(other.Name); r != 0 {
		return r
	}
	return id.Type.Cmp(other.Type)
}

type HeaderData struct {
	MagicNum [4]byte // 0x11 0x00 0x00 0xF0
	NumTypes uint32
	NumFiles uint32

	Unk00          [20]byte
	ApproxMainSize uint64 // aligned by 256 / weirdly offset
	ApproxGPUSize  uint64 // aligned by 256 / weirdly offset
	Unk01          [24]byte
}

type TypeData struct {
	Unk00         uint32
	Unk01         uint32
	Name          Hash
	Count         uint32
	Unk02         uint32
	MainAlignment uint32
	GPUAlignment  uint32
}

type FileData struct {
	ID               FileID
	Offsets          [NumDataType]uint64
	MainBufferOffset uint64
	GPUBufferOffset  uint64
	Sizes            [NumDataType]uint32
	MainAlignment    uint32
	GPUAlignment     uint32
	Index            uint32
}

type Archive struct {
	ID     Hash
	Header HeaderData
	Types  []TypeData
	Files  []FileData
}

func LoadArchive(mainFilename string, mainR io.Reader) (*Archive, error) {
	r := bufio.NewReader(mainR)
	id, err := ParseHash(mainFilename)
	if err != nil {
		return nil, fmt.Errorf("parsing archive main filename: %w", err)
	}

	var hdr HeaderData
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}

	if hdr.MagicNum != [4]byte{0x11, 0x00, 0x00, 0xF0} {
		return nil, errors.New("invalid magic number")
	}

	types := make([]TypeData, hdr.NumTypes)
	if err := binary.Read(r, binary.LittleEndian, types); err != nil {
		return nil, err
	}

	files := make([]FileData, hdr.NumFiles)
	if err := binary.Read(r, binary.LittleEndian, files); err != nil {
		return nil, err
	}

	return &Archive{
		ID:     id,
		Header: hdr,
		Types:  types,
		Files:  files,
	}, nil
}
