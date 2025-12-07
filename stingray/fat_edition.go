package stingray

// This file contains data structures and
// code specific to the "fat" (a.k.a. non-slim)
// version of the game.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func openDataDirFat(ctx context.Context, dirPath string, onProgress func(curr, total int)) (_ *DataDir, err error) {
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
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		ar, err := LoadArchive(ent.Name(), f)
		f.Close()
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

func readNBytesFat(d *DataDir, file FileInfo, typ DataType, nBytes int) ([]byte, error) {
	af, err := os.Open(filepath.Join(d.Path, fmt.Sprintf("%016x%v", file.ArchiveID.Value, typ.ArchiveFileExtension())))
	if err != nil {
		return nil, err
	}
	defer af.Close()
	if _, err := af.Seek(int64(file.Files[typ].Offset), io.SeekStart); err != nil {
		return nil, err
	}

	b := make([]byte, nBytes)
	if _, err := io.ReadFull(af, b); err != nil {
		return nil, err
	}
	return b, nil
}
