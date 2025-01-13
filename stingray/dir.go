package stingray

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xypwn/filediver/util"
)

const errPfx = "stingray: "

type File struct {
	triad  *Triad
	index  int
	exists [3]bool
}

func (f *File) ID() FileID {
	return f.triad.Files[f.index].ID
}

func (f *File) TriadName() string {
	return strings.TrimSuffix(filepath.Base(f.triad.MainPath), filepath.Ext(f.triad.MainPath))
}

func (f *File) TriadID() Hash {
	return f.triad.ID
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
	fileR, err := f.triad.OpenFile(f.index, typ)
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
	// This is for use when exporting by Triad, since even if the file is a duplicate,
	// if its included in the triad it should be exported with the rest. Armors will share geometry,
	// for example
	Duplicates map[FileID]map[Hash]*File
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
		Files:      make(map[FileID]*File),
		Duplicates: make(map[FileID]map[Hash]*File),
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
			if _, contains := dd.Duplicates[v.ID]; !contains {
				dd.Duplicates[v.ID] = make(map[Hash]*File)
			}
			newFile := &File{
				triad: tr,
				index: i,
			}
			for typ := DataType(0); typ < NumDataType; typ++ {
				has, err := tr.HasFile(i, typ)
				if err != nil {
					return nil, err
				}
				newFile.exists[typ] = has
			}
			// According to my testing, files that share the same ID always have the same
			// contents. Uncomment this to re-test that.
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
			dd.Files[v.ID] = newFile
			dd.Duplicates[v.ID][tr.ID] = newFile
		}
	}

	return dd, nil
}

// Opens the "data" game directory, reading all file metadata. Ctx allows for granular cancellation (before each triad open).
// onProgress is optional.
func OpenTriadData(ctx context.Context, triadPath string, onProgress func(curr, total int)) (*DataDir, error) {
	const errPfx = errPfx + "OpenTriad: "

	dd := &DataDir{
		Files: make(map[FileID]*File),
	}

	path := triadPath
	tr, err := OpenTriad(path)
	if err != nil {
		return nil, fmt.Errorf("%v%w", errPfx, err)
	}
	for i, v := range tr.Files {
		newFile := &File{
			triad: tr,
			index: i,
		}
		for typ := DataType(0); typ < NumDataType; typ++ {
			has, err := tr.HasFile(i, typ)
			if err != nil {
				return nil, err
			}
			newFile.exists[typ] = has
		}
		// According to my testing, files that share the same ID always have the same
		// contents. Uncomment this to re-test that.
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
		dd.Files[v.ID] = newFile
	}

	return dd, nil
}
