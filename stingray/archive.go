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

// Relocates all file offsets and updates the archive's Files with the result
func (a *Archive) MarshalBinary() ([]byte, error) {
	a.Header.MagicNum = [4]byte{0x11, 0x00, 0x00, 0xf0}
	a.Header.NumFiles = uint32(len(a.Files))
	a.Header.NumTypes = uint32(len(a.Types))

	var err error
	result := make([]byte, 0)
	if result, err = binary.Append(result, binary.LittleEndian, a.Header); err != nil {
		return nil, err
	}
	if result, err = binary.Append(result, binary.LittleEndian, a.Types); err != nil {
		return nil, err
	}
	relocatedFiles := make([]FileData, len(a.Files))
	offset := uint64(len(result) + binary.Size(a.Files))
	for idx, fileData := range a.Files {
		if offset%uint64(fileData.MainAlignment) != 0 {
			offset += uint64(fileData.MainAlignment) - offset%uint64(fileData.MainAlignment)
		}
		fileData.Offsets[DataMain] = offset
		offset += uint64(fileData.Sizes[DataMain])
		relocatedFiles[idx] = fileData
	}
	a.Files = relocatedFiles
	if result, err = binary.Append(result, binary.LittleEndian, a.Files); err != nil {
		return nil, err
	}
	return result, nil
}

func (a Archive) TypeIdx(typeId Hash) int {
	for i, typ := range a.Types {
		if typ.Name == typeId {
			return i
		}
	}
	return -1
}

func (a Archive) FileIdx(fileId FileID) int {
	for i, file := range a.Files {
		if file.ID.Name == fileId.Name && file.ID.Type == fileId.Type {
			return i
		}
	}
	return -1
}

func (a *Archive) AddFile(fileId FileID, size, alignment uint32) {
	typeIdx := a.TypeIdx(fileId.Type)
	if typeIdx == -1 {
		typeIdx = len(a.Types)
		a.Types = append(a.Types, TypeData{
			Name:          fileId.Type,
			Count:         0,
			MainAlignment: alignment,
			GPUAlignment:  alignment,
		})
	}
	typeData := a.Types[typeIdx]
	typeData.Count += 1
	a.Types[typeIdx] = typeData

	fileIdx := a.FileIdx(fileId)
	if fileIdx == -1 {
		fileIdx = len(a.Files)
		a.Files = append(a.Files, FileData{
			ID:               fileId,
			Offsets:          [3]uint64{},
			MainBufferOffset: 0,
			GPUBufferOffset:  0,
			Sizes:            [3]uint32{},
			Index:            uint32(fileIdx),
			MainAlignment:    alignment,
			GPUAlignment:     alignment,
		})
	}
	fileData := a.Files[fileIdx]
	fileData.Sizes[DataMain] = size
	a.Files[fileIdx] = fileData
}

// Gets the archive size (excluding main data; meaning size of header, type info and file info)
// by reading only the header.
func LoadArchiveSize(mainR io.Reader) (int, error) {
	r := bufio.NewReader(mainR)

	var hdr HeaderData
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return -1, err
	}

	if hdr.MagicNum != [4]byte{0x11, 0x00, 0x00, 0xF0} {
		return -1, errors.New("invalid magic number")
	}

	return binary.Size(HeaderData{}) +
		binary.Size(TypeData{})*int(hdr.NumTypes) +
		binary.Size(FileData{})*int(hdr.NumFiles), nil
}

func LoadArchive(mainFilename string, mainR io.Reader) (*Archive, error) {
	r := bufio.NewReader(mainR)
	id, err := ParseHash(mainFilename)
	if err != nil {
		return nil, fmt.Errorf("parsing archive main, filename: %w", err)
	}

	var hdr HeaderData
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, fmt.Errorf("reading header data: %v", err)
	}

	if hdr.MagicNum != [4]byte{0x11, 0x00, 0x00, 0xF0} {
		return nil, errors.New("invalid magic number")
	}

	types := make([]TypeData, hdr.NumTypes)
	if err := binary.Read(r, binary.LittleEndian, types); err != nil {
		return nil, fmt.Errorf("reading types: %v", err)
	}

	files := make([]FileData, hdr.NumFiles)
	if err := binary.Read(r, binary.LittleEndian, files); err != nil {
		return nil, fmt.Errorf("reading files: %v", err)
	}

	return &Archive{
		ID:     id,
		Header: hdr,
		Types:  types,
		Files:  files,
	}, nil
}
