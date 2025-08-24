package stingray

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Locus represents the location
// of a single partial game file
// (i.e. just main, stream or GPU)
// within an archive.
type Locus struct {
	Offset uint64
	Size   uint32
}

// Exists returns whether the referenced
// partial game file has any contents.
func (l Locus) Exists() bool {
	return l.Size != 0
}

// FileInfo represent the component locations
// of a single game file.
type FileInfo struct {
	ArchiveID Hash
	Files     [NumDataType]Locus
}

func (f FileInfo) Exists(typ DataType) bool {
	return f.Files[typ].Exists()
}

// DataDir represents the collection of game files.
type DataDir struct {
	// Base directory path
	Path string
	// Archive ID to files in that archive
	Archives map[Hash][]FileID
	// File ID to all file info (according to testing,
	// all file info structs in the slice should refer
	// to the same data).
	Files map[FileID][]FileInfo
}

// Opens the "data" game directory, reading all file metadata. Ctx allows for granular cancellation (before each archive open).
// onProgress is optional.
func OpenDataDir(ctx context.Context, dirPath string, onProgress func(curr, total int)) (_ *DataDir, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("stingray: OpenDataDir: %w", err)
		}
	}()

	ents, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	dd := &DataDir{
		Path:     dirPath,
		Archives: make(map[Hash][]FileID),
		Files:    make(map[FileID][]FileInfo),
	}

	for i, ent := range ents {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if onProgress != nil {
			onProgress(i, len(ents))
		}
		if !ent.Type().IsRegular() {
			continue
		}
		if filepath.Ext(ent.Name()) != "" {
			continue
		}
		path := filepath.Join(dirPath, ent.Name())
		ar, err := OpenArchive(path)
		if err != nil {
			return nil, err
		}
		for _, fileData := range ar.Files {
			var file FileInfo
			file.ArchiveID = ar.ID
			for typ := range NumDataType {
				file.Files[typ] = Locus{
					Offset: fileData.Offsets[typ],
					Size:   fileData.Sizes[typ],
				}
			}
			dd.Files[fileData.ID] = append(dd.Files[fileData.ID], file)
			dd.Archives[ar.ID] = append(dd.Archives[ar.ID], fileData.ID)
		}
	}

	return dd, nil
}

// Attempts to read the given file.
// Returns [ErrFileNotExist] if id doesn't exist and
// [ErrFileDataTypeNotExist] if the file exists, but
// doesn't have the requested data type.
func (d *DataDir) Read(id FileID, typ DataType) ([]byte, error) {
	files := d.Files[id]
	if len(files) == 0 {
		return nil, ErrFileNotExist
	}
	file := files[0]
	if !file.Files[typ].Exists() {
		return nil, ErrFileDataTypeNotExist
	}

	af, err := os.Open(filepath.Join(d.Path, fmt.Sprintf("%016x%v", file.ArchiveID.Value, typ.ArchiveFileExtension())))
	if err != nil {
		return nil, err
	}
	defer af.Close()
	if _, err := af.Seek(int64(file.Files[typ].Offset), io.SeekStart); err != nil {
		return nil, err
	}

	b := make([]byte, file.Files[typ].Size)
	if _, err := io.ReadFull(af, b); err != nil {
		return nil, err
	}
	return b, nil
}
