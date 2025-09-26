package previews

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"path"
	"slices"

	"github.com/ebitengine/oto/v3"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
	stingray_wwise "github.com/xypwn/filediver/stingray/wwise"
)

type AutoPreviewType int

const (
	AutoPreviewEmpty AutoPreviewType = iota
	AutoPreviewUnit
	AutoPreviewAudio
	AutoPreviewVideo
	AutoPreviewTexture
	AutoPreviewStrings
	AutoPreviewMaterial
)

type AutoPreviewState struct {
	activeType AutoPreviewType
	activeID   stingray.FileID
	state      struct {
		unit     *UnitPreviewState
		audio    *WwisePreviewState
		video    *BikPreviewState
		texture  *DDSPreviewState
		strings  *StringsPreviewState
		material *MaterialPreviewState
	}

	hashes      map[stingray.Hash]string
	thinhashes  map[stingray.ThinHash]string
	getResource GetResourceFunc

	err error
}

func NewAutoPreview(otoCtx *oto.Context, audioSampleRate int, hashes map[stingray.Hash]string, thinhashes map[stingray.ThinHash]string, getResource GetResourceFunc, runner *exec.Runner) (*AutoPreviewState, error) {
	var err error
	pv := &AutoPreviewState{
		hashes:      hashes,
		thinhashes:  thinhashes,
		getResource: getResource,
	}
	pv.state.unit, err = NewUnitPreview()
	if err != nil {
		return nil, err
	}
	pv.state.audio = NewWwisePreview(otoCtx, audioSampleRate)
	pv.state.video = NewBikPreview(runner)
	pv.state.texture = NewDDSPreview()
	pv.state.strings = NewStringsPreview()
	pv.state.material = NewMaterialPreview()
	return pv, nil
}

func (pv *AutoPreviewState) Delete() {
	pv.state.unit.Delete()
	pv.state.audio.Delete()
	pv.state.video.Delete()
	pv.state.texture.Delete()
}

func (pv *AutoPreviewState) ActiveID() stingray.FileID {
	return pv.activeID
}

func (pv *AutoPreviewState) NeedCJKFont() bool {
	return pv.state.strings.NeedCJKFont()
}

func (pv *AutoPreviewState) LoadFile(ctx context.Context, fileID stingray.FileID, maxVideoVerticalResolution int) {
	if fileID == (stingray.FileID{}) {
		pv.activeID.Name.Value = 0
		pv.activeID.Type.Value = 0
		pv.activeType = AutoPreviewEmpty
		return
	}

	pv.activeID = fileID
	pv.err = nil

	var data [3][]byte
	// Fills data with the files of the according
	// data types. If the requested type doesn't
	// exist, the data slice of the missing type
	// remains nil.
	loadFiles := func(types ...stingray.DataType) error {
		for _, typ := range types {
			if data[typ] != nil {
				panic("programmer error: duplicate data type")
			}
			b, exists, err := pv.getResource(fileID, typ)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			if exists {
				data[typ] = b
			}
		}
		return nil
	}

	switch fileID.Type {
	case stingray.Sum("unit"):
		pv.activeType = AutoPreviewUnit
		if err := loadFiles(stingray.DataMain, stingray.DataGPU); err != nil {
			pv.err = err
			return
		}
		if err := pv.state.unit.LoadUnit(
			fileID.Name,
			data[stingray.DataMain],
			data[stingray.DataGPU],
			pv.getResource,
			pv.thinhashes,
		); err != nil {
			pv.err = fmt.Errorf("loading unit: %w", err)
			return
		}
	case stingray.Sum("wwise_stream"):
		pv.state.audio.ClearStreams()
		pv.activeType = AutoPreviewAudio
		if err := loadFiles(stingray.DataStream); err != nil {
			pv.err = err
			return
		}
		pv.state.audio.Title = fileID.Name.String()
		pv.state.audio.LoadStream(fileID.Name.String(), data[stingray.DataStream], true)
	case stingray.Sum("wwise_bank"):
		pv.state.audio.ClearStreams()
		pv.activeType = AutoPreviewAudio
		if err := loadFiles(stingray.DataMain); err != nil {
			pv.err = err
			return
		}
		bnkFile, ok := pv.hashes[fileID.Name]
		if !ok {
			pv.err = fmt.Errorf("expected wwise bank file %v.wwise_bank to have a known name", fileID.Name)
			return
		}
		pv.state.audio.Title = bnkFile
		dir := path.Dir(bnkFile)
		streams, err := stingray_wwise.BnkGetAllReferencedStreamData(
			bytes.NewReader(data[stingray.DataMain]),
			func(id uint32) (data []byte, exists bool, err error) {
				fileID := stingray.FileID{
					Name: stingray.Sum(path.Join(dir, fmt.Sprint(id))),
					Type: stingray.Sum("wwise_stream"),
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
	case stingray.Sum("bik"):
		pv.activeType = AutoPreviewVideo
		if err := loadFiles(stingray.DataMain, stingray.DataStream, stingray.DataGPU); err != nil {
			pv.err = err
			return
		}
		rs := []io.Reader{bytes.NewReader(data[stingray.DataMain])}
		if data[stingray.DataStream] != nil {
			rs = append(rs, bytes.NewReader(data[stingray.DataStream]))
		} else {
			rs = append(rs, bytes.NewReader(data[stingray.DataGPU]))
		}
		r := io.MultiReader(rs...)

		// Skip stingray header
		if _, err := io.ReadFull(r, make([]byte, 16)); err != nil {
			pv.err = err
			return
		}

		if err := pv.state.video.Load(r, maxVideoVerticalResolution); err != nil {
			cmdNotRegisteredErr := &exec.CommandNotRegisteredError{}
			if errors.As(err, &cmdNotRegisteredErr) {
				pv.err = errors.New(`FFmpeg not found; go to Settings->Extensions to install FFmpeg`)
			} else {
				pv.err = err
			}
			return
		}
	case stingray.Sum("texture"):
		pv.activeType = AutoPreviewTexture
		if err := loadFiles(stingray.DataMain, stingray.DataStream, stingray.DataGPU); err != nil {
			pv.err = err
			return
		}
		r := io.MultiReader(
			bytes.NewReader(data[stingray.DataMain]),
			bytes.NewReader(data[stingray.DataStream]),
			bytes.NewReader(data[stingray.DataGPU]),
		)
		if _, err := texture.DecodeInfo(r); err != nil {
			pv.err = fmt.Errorf("loading stingray DDS info: %w", err)
			return
		}
		img, err := dds.Decode(r, false)
		if err != nil {
			pv.err = fmt.Errorf("loading DDS image: %w", err)
			return
		}
		pv.state.texture.LoadImage(img)
	case stingray.Sum("strings"):
		pv.activeType = AutoPreviewStrings
		if err := loadFiles(stingray.DataMain); err != nil {
			pv.err = err
			return
		}
		data, err := stingray_strings.Load(
			bytes.NewReader(data[stingray.DataMain]),
		)
		if err != nil {
			pv.err = fmt.Errorf("loading DDS image: %w", err)
			return
		}
		pv.state.strings.Load(data, pv.thinhashes)
	case stingray.Sum("material"):
		pv.activeType = AutoPreviewMaterial
		if err := loadFiles(stingray.DataMain); err != nil {
			pv.err = err
			return
		}
		data, err := material.Load(bytes.NewReader(data[stingray.DataMain]))
		if err != nil {
			pv.err = fmt.Errorf("loading material: %w", err)
			return
		}

		err = pv.state.material.LoadMaterial(data, pv.getResource, pv.hashes, pv.thinhashes)
		if err != nil {
			pv.err = fmt.Errorf("loading material state: %w", err)
			return
		}
	default:
		pv.activeType = AutoPreviewEmpty
	}
}

func AutoPreview(name string, pv *AutoPreviewState) bool {
	if pv.err != nil {
		imutils.TextError(pv.err)
		return true
	}
	switch pv.activeType {
	case AutoPreviewEmpty:
		return false
	case AutoPreviewUnit:
		UnitPreview(name, pv.state.unit)
	case AutoPreviewAudio:
		WwisePreview(name, pv.state.audio)
	case AutoPreviewVideo:
		BikPreview(pv.state.video)
	case AutoPreviewTexture:
		DDSPreview(name, pv.state.texture)
	case AutoPreviewStrings:
		StringsPreview(pv.state.strings)
	case AutoPreviewMaterial:
		MaterialPreview(name, pv.state.material)
	default:
		panic("unhandled case")
	}
	return true
}
