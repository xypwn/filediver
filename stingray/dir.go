package stingray

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const errPfx = "stingray: "

type File struct {
	triad  *Triad
	index  int
	exists [3]bool
}

func (f *File) Exists(typ DataType) bool {
	return f.exists[typ]
}

// Call Close() on returned reader when done.
func (f *File) Open(typ DataType) (io.ReadSeekCloser, error) {
	return f.triad.OpenFile(f.index, typ)
}

// For testing purposes, takes a HUGE amount of time to execute.
func (a *File) contentEqual(b *File, dt DataType) (bool, error) {
	fa, err := a.Open(dt)
	if err != nil {
		return false, err
	}
	defer fa.Close()
	ba, err := io.ReadAll(fa)
	if err != nil {
		return false, err
	}
	fb, err := b.Open(dt)
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
}

func OpenDataDir(dirPath string) (*DataDir, error) {
	const errPfx = errPfx + "OpenDataDir: "

	ents, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("%v%w", errPfx, err)
	}

	dd := &DataDir{
		Files: make(map[FileID]*File),
	}

	for _, ent := range ents {
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
	}

	return dd, nil
}
