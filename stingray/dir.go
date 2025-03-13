package stingray

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xypwn/filediver/util"
)

const errPfx = "stingray: "

type File struct {
	// Multiple triads may point to the same file.
	// Len is assumed to be >= 1.
	triads []*Triad
	// Actual file data == triads[0].Files[indexInTriad0].
	indexInTriad0 int
	// Existing file types (main/stream/GPU).
	exists [3]bool
}

func (f *File) ID() FileID {
	return f.triads[0].Files[f.indexInTriad0].ID
}

func (f *File) TriadIDs() []Hash {
	res := make([]Hash, len(f.triads))
	for i := range f.triads {
		res[i] = f.triads[i].ID
	}
	return res
}

func (f *File) Exists(typ DataType) bool {
	return f.exists[typ]
}

type preallocReader struct {
	*bytes.Reader
	b []byte
}

func newPreallocReader(r io.Reader) (*preallocReader, error) {
	pr := &preallocReader{}
	var err error
	pr.b, err = io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	pr.Reader = bytes.NewReader(pr.b)
	return pr, nil
}

func (r *preallocReader) Close() error {
	return nil
}

// Call Close() on returned reader when done.
func (f *File) Open(ctx context.Context, typ DataType) (io.ReadSeekCloser, error) {
	fileR, err := f.triads[0].OpenFile(f.indexInTriad0, typ)
	if err != nil {
		return nil, err
	}
	r, err := newPreallocReader(fileR)
	if err != nil {
		return nil, err
	}
	return util.NewContextReadSeekCloser(ctx, r), nil
}

type multiReadCloser struct {
	io.Reader
	underlying []io.ReadCloser
}

func (r *multiReadCloser) Close() error {
	var firstErr error
	for _, v := range r.underlying {
		err := v.Close()
		if err != nil && firstErr != nil {
			firstErr = err
		}
	}
	return firstErr
}

// Returns a MultiReader concatenating the given data stream types.
// Skips any specified types that don't exist.
// If you need seeking functionality, use Open().
// Call Close() on returned reader when done.
func (f *File) OpenMulti(ctx context.Context, types ...DataType) (io.ReadCloser, error) {
	var rdcs []io.ReadCloser
	var rds []io.Reader
	for _, dataType := range types {
		if !f.Exists(dataType) {
			continue
		}
		r, err := f.Open(ctx, dataType)
		if err != nil {
			for _, rdc := range rdcs {
				rdc.Close()
			}
			return nil, err
		}
		rdcs = append(rdcs, r)
		rds = append(rds, r)
	}
	return &multiReadCloser{
		Reader:     io.MultiReader(rds...),
		underlying: rdcs,
	}, nil
}

// For testing purposes, takes a HUGE amount of time to execute.
func (a *File) contentEqual(b *File, dt DataType) (bool, error) {
	ctx := context.Background()
	fa, err := a.Open(ctx, dt)
	if err != nil {
		return false, err
	}
	defer fa.Close()
	ba, err := io.ReadAll(fa)
	if err != nil {
		return false, err
	}
	fb, err := b.Open(ctx, dt)
	if err != nil {
		return false, err
	}
	defer fb.Close()
	bb, err := io.ReadAll(fb)
	if err != nil {
		return false, err
	}
	return bytes.Equal(ba, bb), nil
}

type DataDir struct {
	Files map[FileID]*File
	// Triad ID to map of files by ID.
	FilesByTriad map[Hash]map[FileID]*File
}

// Opens the "data" game directory, reading all file metadata. Ctx allows for granular cancellation (before each triad open).
// onProgress is optional.
func OpenDataDir(ctx context.Context, dirPath string, onProgress func(curr, total int)) (*DataDir, error) {
	const errPfx = errPfx + "OpenDataDir: "

	ents, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("%v%w", errPfx, err)
	}

	dd := &DataDir{
		Files:        make(map[FileID]*File),
		FilesByTriad: make(map[Hash]map[FileID]*File),
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
		tr, err := OpenTriad(path)
		if err != nil {
			return nil, fmt.Errorf("%v%w", errPfx, err)
		}
		for i, v := range tr.Files {
			file, fileExists := dd.Files[v.ID]
			if fileExists {
				file.triads = append(file.triads, tr)
			} else {
				file = &File{
					triads:        []*Triad{tr},
					indexInTriad0: i,
				}
				for typ := DataType(0); typ < NumDataType; typ++ {
					has, err := tr.HasFile(i, typ)
					if err != nil {
						return nil, err
					}
					file.exists[typ] = has
				}
				dd.Files[v.ID] = file
			}
			// According to my testing, files that share the same ID always have the same
			// contents, but may have multiple triads pointing to them.
			// Uncomment this to re-test that.
			/*if _, ok := dd.Files[v.ID]; ok {
				for _, typ := range []DataType{DataMain, DataStream, DataGPU} {
					if newFile.Exists(typ) && dd.Files[v.ID].Exists(typ) {
						if eq, err := newFile.contentEqual(dd.Files[v.ID], typ); err != nil {
							return nil, err
						} else if !eq {
							return nil, fmt.Errorf("%vduplicate file ID with different %v file contents", errPfx, typ)
						}
					}
				}
			}*/
			if _, ok := dd.FilesByTriad[tr.ID]; !ok {
				dd.FilesByTriad[tr.ID] = make(map[FileID]*File)
			}
			dd.FilesByTriad[tr.ID][v.ID] = file
		}
	}

	return dd, nil
}
