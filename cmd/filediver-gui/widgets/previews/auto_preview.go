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
	AutoPreviewTree
	AutoPreviewAudio
	AutoPreviewVideo
	AutoPreviewTexture
	AutoPreviewStrings
	AutoPreviewMaterial
	AutoPreviewXAML
)

type AutoPreview struct {
	activeType AutoPreviewType
	activeID   stingray.FileID
	previews   struct {
		unit      *UnitPreviewState
		speedtree *SpeedtreePreviewState
		audio     *WwisePreview
		video     *BinkPreview
		texture   *ImagePreview
		strings   *StringsPreview
		material  *MaterialPreview
		xaml      *XamlPreview
	}

	hashes      map[stingray.Hash]string
	thinhashes  map[stingray.ThinHash]string
	getResource GetResourceFunc

	err error
}

func NewAutoPreview(otoCtx *oto.Context, audioSampleRate int, hashes map[stingray.Hash]string, thinhashes map[stingray.ThinHash]string, getResource GetResourceFunc, runner *exec.Runner) (*AutoPreview, error) {
	var err error
	pv := &AutoPreview{
		hashes:      hashes,
		thinhashes:  thinhashes,
		getResource: getResource,
	}
	pv.previews.unit, err = NewUnitPreview()
	if err != nil {
		return nil, err
	}
	pv.previews.speedtree, err = NewSpeedtreePreview()
	if err != nil {
		return nil, err
	}
	pv.previews.audio = NewWwisePreview(otoCtx, audioSampleRate)
	pv.previews.video = NewBinkPreview(runner)
	pv.previews.texture = NewImagePreview()
	pv.previews.texture.Flags = LinearFilteringButton | IgnoreAlphaButton
	pv.previews.strings = NewStringsPreview()
	pv.previews.material = NewMaterialPreview()
	pv.previews.xaml = NewXamlPreview()
	return pv, nil
}

func (pv *AutoPreview) Delete() {
	pv.previews.unit.Delete()
	pv.previews.speedtree.Delete()
	pv.previews.audio.Delete()
	pv.previews.video.Delete()
	pv.previews.texture.Delete()
	pv.previews.xaml.Delete()
}

func (pv *AutoPreview) ActiveID() stingray.FileID {
	return pv.activeID
}

func (pv *AutoPreview) ActiveType() AutoPreviewType {
	return pv.activeType
}

func (pv *AutoPreview) NeedCJKFont() bool {
	return pv.previews.strings.NeedCJKFont()
}

func (pv *AutoPreview) LoadFile(ctx context.Context, fileID stingray.FileID, maxVideoVerticalResolution int) {
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
		if err := pv.previews.unit.LoadUnit(
			fileID.Name,
			data[stingray.DataMain],
			data[stingray.DataGPU],
			pv.getResource,
			pv.thinhashes,
		); err != nil {
			pv.err = fmt.Errorf("loading unit: %w", err)
			return
		}
	case stingray.Sum("speedtree"):
		pv.activeType = AutoPreviewTree
		if err := loadFiles(stingray.DataMain, stingray.DataGPU); err != nil {
			pv.err = err
			return
		}
		if err := pv.previews.speedtree.LoadSpeedtree(
			fileID.Name,
			data[stingray.DataMain],
			data[stingray.DataGPU],
			pv.getResource,
			pv.thinhashes,
		); err != nil {
			pv.err = fmt.Errorf("loading speedtree: %w", err)
			return
		}
	case stingray.Sum("wwise_stream"):
		pv.previews.audio.ClearStreams()
		pv.activeType = AutoPreviewAudio
		if err := loadFiles(stingray.DataStream); err != nil {
			pv.err = err
			return
		}
		pv.previews.audio.Title = fileID.Name.String()
		pv.previews.audio.LoadStream(fileID.Name.String(), data[stingray.DataStream], nil, true)
	case stingray.Sum("wwise_bank"):
		pv.previews.audio.ClearStreams()
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
		pv.previews.audio.Title = bnkFile
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
			pv.previews.audio.LoadStream(fmt.Sprint(id), stream.Data, stream.Err, false)
		}
	case stingray.Sum("bik"), stingray.Sum("bk2"):
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

		if err := pv.previews.video.Load(r, maxVideoVerticalResolution); err != nil {
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
		ImagePreviewLoadDDS(pv.previews.texture, img)
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
		pv.previews.strings.Load(data, pv.thinhashes)
	case stingray.Sum("material"):
		pv.activeType = AutoPreviewMaterial
		if err := loadFiles(stingray.DataMain); err != nil {
			pv.err = err
			return
		}
		data, err := material.LoadMain(bytes.NewReader(data[stingray.DataMain]))
		if err != nil {
			pv.err = fmt.Errorf("loading material: %w", err)
			return
		}

		err = pv.previews.material.LoadMaterial(data, pv.getResource, pv.hashes, pv.thinhashes)
		if err != nil {
			pv.err = fmt.Errorf("loading material: %w", err)
			return
		}
	case stingray.Sum("xaml"):
		pv.activeType = AutoPreviewXAML
		if err := loadFiles(stingray.DataMain); err != nil {
			pv.err = err
			return
		}
		if err := pv.previews.xaml.LoadXaml(data[stingray.DataMain]); err != nil {
			pv.err = fmt.Errorf("loading xaml: %w", err)
			return
		}
	default:
		pv.activeType = AutoPreviewEmpty
	}
}

func (pv *AutoPreview) MaterialSettingsEmpty() bool {
	return pv.previews.material.SettingsEmpty()
}

func (pv *AutoPreview) MaterialSettingsVisible() bool {
	return pv.previews.material.SettingsVisible()
}

func (pv *AutoPreview) SetMaterialSettingsVisible(visible bool) {
	pv.previews.material.SetSettingsVisible(visible)
}

func (pv *AutoPreview) Draw(name string) bool {
	if pv.err != nil {
		imutils.TextError(pv.err)
		return true
	}
	switch pv.activeType {
	case AutoPreviewEmpty:
		return false
	case AutoPreviewUnit:
		UnitPreview(name, pv.previews.unit)
	case AutoPreviewTree:
		SpeedtreePreview(name, pv.previews.speedtree)
	case AutoPreviewAudio:
		pv.previews.audio.Draw(name)
	case AutoPreviewVideo:
		pv.previews.video.Draw()
	case AutoPreviewTexture:
		pv.previews.texture.Draw(name)
	case AutoPreviewStrings:
		pv.previews.strings.Draw()
	case AutoPreviewMaterial:
		pv.previews.material.Draw(name)
	case AutoPreviewXAML:
		pv.previews.xaml.Draw(name)
	default:
		panic("unhandled case")
	}
	return true
}

func (pv *AutoPreview) DrawMaterialSettings() bool {
	if pv.err != nil {
		return true
	}
	switch pv.activeType {
	case AutoPreviewEmpty:
		return false
	case AutoPreviewMaterial:
		pv.previews.material.DrawSettings()
	}
	return true
}
