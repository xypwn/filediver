package widgets

import (
	"context"
	"fmt"
	"io"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/stingray"
)

type FileAutoPreviewType int

const (
	FileAutoPreviewEmpty FileAutoPreviewType = iota
	FileAutoPreviewUnit
	FileAutoPreviewTexture
)

type FileAutoPreviewState struct {
	activeType FileAutoPreviewType
	state      struct {
		unit *UnitPreviewState
	}
	err error

	IsUsing bool
}

func NewFileAutoPreview() (*FileAutoPreviewState, error) {
	var err error
	pv := &FileAutoPreviewState{}
	pv.state.unit, err = NewUnitPreview()
	if err != nil {
		return nil, err
	}
	return pv, nil
}

func (pv *FileAutoPreviewState) Delete() {
	pv.state.unit.Delete()
}

func (pv *FileAutoPreviewState) LoadFile(ctx context.Context, file *stingray.File) {
	if file == nil {
		pv.activeType = FileAutoPreviewEmpty
		return
	}

	pv.err = nil

	var readers [3]io.ReadSeekCloser
	loadFiles := func(types ...stingray.DataType) error {
		for _, typ := range types {
			if readers[typ] != nil {
				panic("programmer error: duplicate data type")
			}
			rd, err := file.Open(ctx, typ)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			readers[typ] = rd
		}
		return nil
	}

	defer func() {
		for _, rd := range readers {
			if rd != nil {
				rd.Close()
			}
		}
	}()

	switch file.ID().Type {
	case stingray.Sum64([]byte("unit")):
		pv.activeType = FileAutoPreviewUnit
		if err := loadFiles(stingray.DataMain, stingray.DataGPU); err != nil {
			pv.err = err
			return
		}
		if err := pv.state.unit.LoadUnit(
			readers[stingray.DataMain],
			readers[stingray.DataGPU],
		); err != nil {
			pv.err = fmt.Errorf("loading unit: %w", err)
			return
		}
	default:
		pv.activeType = FileAutoPreviewEmpty
	}
}

func FileAutoPreview(name string, pv *FileAutoPreviewState) bool {
	if pv.err != nil {
		imgui.PushStyleColorVec4(imgui.ColText, imgui.NewVec4(0.8, 0.5, 0.5, 1))
		imgui.TextUnformatted(fmt.Sprintf("Error: %v", pv.err))
		imgui.PopStyleColor()
		return true
	}
	switch pv.activeType {
	case FileAutoPreviewEmpty:
		return false
	case FileAutoPreviewUnit:
		UnitPreview(name, pv.state.unit)
		pv.IsUsing = pv.state.unit.IsUsing
	default:
		panic("unhandled case")
	}
	return true
}
