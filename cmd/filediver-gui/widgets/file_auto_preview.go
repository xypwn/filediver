package widgets

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"path"
	"slices"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/ebitengine/oto/v3"
	"github.com/xypwn/filediver/stingray"
	stingray_wwise "github.com/xypwn/filediver/stingray/wwise"
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
	activeID   stingray.FileID
	state      struct {
		unit  *UnitPreviewState
		audio *WwisePreviewState
	}

	hashes      map[stingray.Hash]string
	getResource GetResourceFunc

	err error
}

func NewFileAutoPreview(otoCtx *oto.Context, audioSampleRate int, hashes map[stingray.Hash]string, getResource GetResourceFunc) (*FileAutoPreviewState, error) {
	var err error
	pv := &FileAutoPreviewState{
		hashes:      hashes,
		getResource: getResource,
	}
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

func (pv *FileAutoPreviewState) ActiveID() stingray.FileID {
	return pv.activeID
}

func (pv *FileAutoPreviewState) LoadFile(ctx context.Context, file *stingray.File) {
	if file == nil {
		pv.activeID.Name.Value = 0
		pv.activeID.Type.Value = 0
		pv.activeType = FileAutoPreviewEmpty
		return
	}

	pv.activeID = file.ID()
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
		pv.state.audio.Title = file.ID().Name.String()
		pv.state.audio.LoadStream(file.ID().Name.String(), data[stingray.DataStream], true)
	case stingray.Sum64([]byte("wwise_bank")):
		pv.state.audio.ClearStreams()
		pv.activeType = FileAutoPreviewAudio
		if err := loadFiles(stingray.DataMain); err != nil {
			pv.err = err
			return
		}
		bnkFile, ok := pv.hashes[file.ID().Name]
		if !ok {
			pv.err = fmt.Errorf("expected wwise bank file %v.wwise_bank to have a known name", file.ID().Name)
			return
		}
		pv.state.audio.Title = bnkFile
		dir := path.Dir(bnkFile)
		streams, err := stingray_wwise.BnkGetAllReferencedStreamData(
			bytes.NewReader(data[stingray.DataMain]),
			func(id uint32) (data []byte, exists bool, err error) {
				fileID := stingray.FileID{
					Name: stingray.Sum64([]byte(path.Join(dir, fmt.Sprint(id)))),
					Type: stingray.Sum64([]byte("wwise_stream")),
				}
				return pv.getResource(fileID, stingray.DataStream)
			},
		)
		if err != nil {
			pv.err = fmt.Errorf("loading wwise bank: %w", err)
			return
		}
		for _, id := range slices.Sorted(maps.Keys(streams)) {
			stream := streams[id]
			pv.state.audio.LoadStream(fmt.Sprint(id), stream, false)
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
