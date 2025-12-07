package stingray

import (
	"context"
	"errors"
	"fmt"
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
	// Whether the edition of the
	// game is prod_slim.
	IsSlimEdition bool
	// === SLIM EDITION ONLY FIELDS === //
	DSAA                 *DSAA
	ArchiveDSAAIndices   map[Hash][NumDataType]int
	Bundles              []NXABundleInfo
	SingleArchiveBundles map[Hash][NumDataType]*DSARStructure // single-archive DSAR bundles that are non-NXA, e.g. localized audio
	// === END                      === //
}

// Opens the "data" game directory, reading all file metadata. Ctx allows for rough cancellation (before each archive open).
// onProgress is optional.
func OpenDataDir(ctx context.Context, dirPath string, onProgress func(curr, total int)) (_ *DataDir, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("stingray: OpenDataDir: %w", err)
		}
	}()
	if _, err := os.Stat(filepath.Join(dirPath, "9ba626afa44a3aa3")); err == nil {
		// package/boot archive exists -> fat edition
		return openDataDirFat(ctx, dirPath, onProgress)
	} else if errors.Is(err, os.ErrNotExist) {
		// doesn't exist -> is contained in NXA bundle -> slim edition
		return openDataDirSlim(ctx, dirPath, onProgress)
	} else {
		return nil, err
	}
}

// Attempts to read at most nBytes from the given file, or all bytes
// if nBytes is -1.
// Returns [ErrFileNotExist] if id doesn't exist and
// [ErrFileDataTypeNotExist] if the file exists, but
// doesn't have the requested data type.
func (d *DataDir) ReadAtMost(id FileID, typ DataType, nBytes int) ([]byte, error) {
	files := d.Files[id]
	if len(files) == 0 {
		return nil, ErrFileNotExist
	}
	file := files[0]
	if !file.Files[typ].Exists() {
		return nil, ErrFileDataTypeNotExist
	}

	if nBytes == -1 {
		nBytes = int(file.Files[typ].Size)
	} else {
		nBytes = min(int(file.Files[typ].Size), nBytes)
	}

	if d.IsSlimEdition {
		return readNBytesSlim(d, file, typ, nBytes)
	} else {
		return readNBytesFat(d, file, typ, nBytes)
	}
}

func (d *DataDir) Read(id FileID, typ DataType) ([]byte, error) {
	return d.ReadAtMost(id, typ, -1)
}
