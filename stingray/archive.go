package stingray

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	Paths  [NumDataType]string // stream and GPU are optional
	Header HeaderData
	Types  []TypeData
	Files  []FileData
}

func OpenArchive(mainPath string) (*Archive, error) {
	if filepath.Ext(mainPath) != "" {
		return nil, errors.New("expected path to file with no extension")
	}

	basename := strings.TrimSuffix(filepath.Base(mainPath), filepath.Ext(mainPath))
	id, err := ParseHash(basename)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(mainPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)

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

	var paths [NumDataType]string
	paths[DataMain] = mainPath
	for _, typ := range []DataType{DataStream, DataGPU} {
		path := mainPath + typ.ArchiveFileExtension()
		if _, err := os.Stat(path); err != nil {
			path = ""
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}

	return &Archive{
		ID:     id,
		Paths:  paths,
		Header: hdr,
		Types:  types,
		Files:  files,
	}, nil
}

func (ar *Archive) HasDataType(typ DataType) bool {
	return ar.Paths[typ] != ""
}

func (ar *Archive) fileInfo(fileIndex int, typ DataType) (path string, offset uint64, size uint32, err error) {
	if !ar.HasDataType(typ) {
		return "", 0, 0, fmt.Errorf("%v data: %w", typ, ErrFileDataTypeNotExist)
	}

	if fileIndex >= len(ar.Files) {
		return "", 0, 0, fmt.Errorf("file index out of range (got: %v, max: %v)", fileIndex, len(ar.Files)-1)
	}

	info := ar.Files[fileIndex]
	return ar.Paths[typ], info.Offsets[typ], info.Sizes[typ], nil
}

func (ar *Archive) HasFile(fileIndex int, typ DataType) (bool, error) {
	if !ar.HasDataType(typ) {
		return false, nil
	}

	_, _, size, err := ar.fileInfo(fileIndex, typ)
	if err != nil {
		return false, err
	}

	return size > 0, nil
}

// ReadFile attempts to read the file with the given index and type
// within the Archive.
// Returns a wrapped version of ErrFileDataTypeNotExist when the file exists,
// but doesn't have the requested data type.
// If maxBytes is zero, the entire file is read. If maxBytes is nonzero,
// maxBytes bytes or less are read.
func (ar *Archive) ReadFile(fileIndex int, typ DataType, maxBytes int) ([]byte, error) {
	path, offset, size, err := ar.fileInfo(fileIndex, typ)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, fmt.Errorf("%v data: %w", typ, ErrFileDataTypeNotExist)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, err
	}

	var data []byte
	if maxBytes == 0 {
		data = make([]byte, size)
	} else {
		data = make([]byte, min(int(size), maxBytes))
	}
	if _, err := io.ReadFull(f, data); err != nil {
		return nil, err
	}

	return data, nil
}
