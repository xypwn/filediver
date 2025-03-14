package widgets

import (
	"context"
	"fmt"
	"io"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/ebitengine/oto/v3"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/wwise"
)

type FileAutoPreviewType int

const (
	FileAutoPreviewEmpty FileAutoPreviewType = iota
	FileAutoPreviewUnit
	FileAutoPreviewAudio
	FileAutoPreviewTexture
)

type FileAutoPreviewState struct {
	activeType FileAutoPreviewType
	state      struct {
		unit  *UnitPreviewState
		audio *WwisePreviewState
	}
	err error
}

func NewFileAutoPreview(otoCtx *oto.Context) (*FileAutoPreviewState, error) {
	var err error
	pv := &FileAutoPreviewState{}
	pv.state.unit, err = NewUnitPreview()
	if err != nil {
		return nil, err
	}
	pv.state.audio = NewWwisePreview(otoCtx)
	return pv, nil
}

func (pv *FileAutoPreviewState) Delete() {
	pv.state.unit.Delete()
	pv.state.audio.Delete()
}

func (pv *FileAutoPreviewState) LoadFile(ctx context.Context, file *stingray.File) {
	if file == nil {
		pv.activeType = FileAutoPreviewEmpty
		return
	}

	pv.err = nil

	// TODO: Evaluate if maybe we shouldn't pre-read all data all the time.
	// Currently this is necessary since we close the readers (although right
	// now the implementation is prealloc readers internally, but that may
	// change in the future).
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
	case stingray.Sum64([]byte("wwise_stream")):
		pv.state.audio.ClearStreams()
		pv.state.audio.PlayOnLoadStream = true
		pv.activeType = FileAutoPreviewAudio
		if err := loadFiles(stingray.DataStream); err != nil {
			pv.err = err
			return
		}
		wem, err := wwise.OpenWem(readers[stingray.DataStream])
		if err != nil {
			pv.err = fmt.Errorf("loading wwise stream: %w", err)
			return
		}
		pv.state.audio.Title = file.ID().Name.String()
		pv.state.audio.LoadStream(file.ID().Name.String(), wem)
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
	case FileAutoPreviewAudio:
		WwisePreview(name, pv.state.audio)
	default:
		panic("unhandled case")
	}
	return true
}
