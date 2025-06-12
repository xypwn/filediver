// "triad" describes a storage unit used in the game files.
// It consists of a "main" file (no extension), a "stream" file (.stream),
// and a "GPU" file (.gpu_resources). The stream and GPU files are optional.
package stingray

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func (t DataType) Extension() string {
	switch t {
	case DataMain:
		return "main"
	case DataStream:
		return "stream"
	case DataGPU:
		return "gpu"
	default:
		panic("unhandled case")
	}
}

type FileID struct {
	Name Hash
	Type Hash
}

// Cmp compares two file IDs with order: name > type.
func (id FileID) Cmp(other FileID) int {
	if r := id.Name.Cmp(other.Name); r != 0 {
		return r
	}
	return id.Type.Cmp(other.Type)
}

// Unk means the data's purpose is unknown.
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
	MainOffset       uint64
	StreamOffset     uint64
	GPUOffset        uint64
	MainBufferOffset uint64
	GPUBufferOffset  uint64
	MainSize         uint32
	StreamSize       uint32
	GPUSize          uint32
	MainAlignment    uint32
	GPUAlignment     uint32
	Index            uint32
}

// A triad consists of a main file, a stream file and a GPU resource file.
// The stream file and GPU resource file are optional.
type Triad struct {
	ID         Hash
	MainPath   string
	StreamPath string // optional
	GPUpath    string // optional
	Header     HeaderData
	Types      []TypeData
	Files      []FileData
}

func OpenTriad(mainPath string) (*Triad, error) {
	if filepath.Ext(mainPath) != "" {
		return nil, errors.New("expected path to file with no extension")
	}

	basename := strings.TrimSuffix(filepath.Base(mainPath), filepath.Ext(mainPath))
	id, err := ParseHash(basename)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(mainPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hdr HeaderData
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}

	if hdr.MagicNum != [4]byte{0x11, 0x00, 0x00, 0xF0} {
		return nil, errors.New("invalid magic number")
	}

	types := make([]TypeData, hdr.NumTypes)
	for i := 0; i < int(hdr.NumTypes); i++ {
		if err := binary.Read(f, binary.LittleEndian, &types[i]); err != nil {
			return nil, err
		}
	}

	files := make([]FileData, hdr.NumFiles)
	for i := 0; i < int(hdr.NumFiles); i++ {
		if err := binary.Read(f, binary.LittleEndian, &files[i]); err != nil {
			return nil, err
		}
	}

	streamPath := mainPath + ".stream"
	fStream, err := os.Open(streamPath)
	if err != nil {
		streamPath = ""
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	defer fStream.Close()

	gpuPath := mainPath + ".gpu_resources"
	fGPU, err := os.Open(gpuPath)
	if err != nil {
		gpuPath = ""
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	defer fGPU.Close()

	return &Triad{
		ID:         id,
		MainPath:   mainPath,
		StreamPath: streamPath,
		GPUpath:    gpuPath,
		Header:     hdr,
		Types:      types,
		Files:      files,
	}, nil
}

func (tr *Triad) HasDataType(typ DataType) bool {
	switch typ {
	case DataMain:
		return tr.MainPath != ""
	case DataStream:
		return tr.StreamPath != ""
	case DataGPU:
		return tr.GPUpath != ""
	default:
		panic("unhandled case")
	}
}

func (tr *Triad) fileInfo(fileIndex int, typ DataType) (path string, offset uint64, size uint32, err error) {
	if !tr.HasDataType(typ) {
		return "", 0, 0, fmt.Errorf("don't have %v file type", typ)
	}

	if fileIndex >= len(tr.Files) {
		return "", 0, 0, fmt.Errorf("file index out of range (got: %v, max: %v)", fileIndex, len(tr.Files)-1)
	}

	info := tr.Files[fileIndex]
	switch typ {
	case DataMain:
		path = tr.MainPath
		offset = info.MainOffset
		size = info.MainSize
	case DataStream:
		path = tr.StreamPath
		offset = info.StreamOffset
		size = info.StreamSize
	case DataGPU:
		path = tr.GPUpath
		offset = info.GPUOffset
		size = info.GPUSize
	default:
		panic("unhandled case")
	}

	return
}

func (tr *Triad) HasFile(fileIndex int, typ DataType) (bool, error) {
	if !tr.HasDataType(typ) {
		return false, nil
	}

	_, _, size, err := tr.fileInfo(fileIndex, typ)
	if err != nil {
		return false, err
	}

	return size > 0, nil
}

type containedFile struct {
	*io.SectionReader
	f *os.File
}

func (f *containedFile) Close() error {
	return f.f.Close()
}

// Call Close() on returned reader when done.
func (tr *Triad) OpenFile(fileIndex int, typ DataType) (io.ReadSeekCloser, error) {
	path, offset, size, err := tr.fileInfo(fileIndex, typ)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, errors.New("contained file does not exist")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fInfo, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	if d := int64(offset) + int64(size) - fInfo.Size(); d > 0 {
		f.Close()
		return nil, fmt.Errorf("contained file exceeds %v bytes beyond container file", d)
	}

	return &containedFile{
		SectionReader: io.NewSectionReader(f, int64(offset), int64(size)),
		f:             f,
	}, nil
}
