package widgets

import (
	"bytes"
	"context"
	"fmt"

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

func NewFileAutoPreview(otoCtx *oto.Context, audioSampleRate int) (*FileAutoPreviewState, error) {
	var err error
	pv := &FileAutoPreviewState{}
	pv.state.unit, err = NewUnitPreview()
	if err != nil {
		return nil, err
	}
	pv.state.audio = NewWwisePreview(otoCtx, audioSampleRate)
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

	var data [3][]byte
	loadFiles := func(types ...stingray.DataType) error {
		for _, typ := range types {
			if data[typ] != nil {
				panic("programmer error: duplicate data type")
			}
			b, err := file.Read(typ)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			data[typ] = b
		}
		return nil
	}

	switch file.ID().Type {
	case stingray.Sum64([]byte("unit")):
		pv.activeType = FileAutoPreviewUnit
		if err := loadFiles(stingray.DataMain, stingray.DataGPU); err != nil {
			pv.err = err
			return
		}
		if err := pv.state.unit.LoadUnit(
			data[stingray.DataMain],
			data[stingray.DataGPU],
		); err != nil {
			pv.err = fmt.Errorf("loading unit: %w", err)
			return
		}
	case stingray.Sum64([]byte("wwise_stream")):
		pv.state.audio.ClearStreams()
		pv.activeType = FileAutoPreviewAudio
		if err := loadFiles(stingray.DataStream); err != nil {
			pv.err = err
			return
		}
		wem, err := wwise.OpenWem(bytes.NewReader(data[stingray.DataStream]))
		if err != nil {
			pv.err = fmt.Errorf("loading wwise stream: %w", err)
			return
		}
		pv.state.audio.Title = file.ID().Name.String()
		if err := pv.state.audio.LoadStream(file.ID().Name.String(), wem, true); err != nil {
			pv.err = err
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
	case FileAutoPreviewAudio:
		WwisePreview(name, pv.state.audio)
	default:
		panic("unhandled case")
	}
	return true
}
