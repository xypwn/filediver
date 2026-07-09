package previews

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

type MaterialPreview struct {
	*ImagePreview
	settings         map[string][]float32
	settingKeys      []string
	settingsVisible  bool
	baseMaterial     stingray.Hash
	baseMaterialName string
}

func NewMaterialPreview() *MaterialPreview {
	return &MaterialPreview{
		ImagePreview:    NewImagePreview(),
		settings:        make(map[string][]float32),
		settingsVisible: true,
	}
}

func (pv *MaterialPreview) LoadMaterial(mat *material.Material, getResource GetResourceFunc, hashes map[stingray.Hash]string, thinhashes map[stingray.ThinHash]string) error {
	if mat == nil {
		return fmt.Errorf("attempted to load nil material")
	}

	clear(pv.settings)
	pv.settingKeys = nil
	pv.baseMaterial = stingray.Hash{}

	var imgsToLoad []image.Image
	var imgInfoTexts []string

	sortedTextureUsages := slices.SortedFunc(maps.Keys(mat.Textures), stingray.ThinHash.Cmp)
	for i, key := range sortedTextureUsages {
		path := mat.Textures[key]
		var imageName, pathName string
		var ok bool
		imageName, ok = thinhashes[key]
		if !ok {
			imageName = "Unknown texture usage: " + key.String()
		}
		if pathName, ok = hashes[path]; !ok {
			pathName = path.String()
		}

		var img *dds.DDS
		if path.Value != 0 {
			dataMain, fileExists, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum("texture")}, stingray.DataMain)
			if err != nil {
				return fmt.Errorf("material texture %v: %w", pathName, err)
			}
			if !fileExists {
				return fmt.Errorf("material texture %v: referenced texture does not exist", pathName)
			}

			dataGPU, _, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum("texture")}, stingray.DataGPU)
			if err != nil {
				return fmt.Errorf("material texture %v: %w", pathName, err)
			}

			dataStream, _, err := getResource(stingray.FileID{Name: path, Type: stingray.Sum("texture")}, stingray.DataStream)
			if err != nil && !errors.Is(err, stingray.ErrFileDataTypeNotExist) {
				return fmt.Errorf("material texture %v: %w", pathName, err)
			}

			r := io.MultiReader(
				bytes.NewReader(dataMain),
				bytes.NewReader(dataStream),
				bytes.NewReader(dataGPU),
			)
			if _, err := texture.DecodeInfo(r); err != nil {
				return fmt.Errorf("material texture %v: loading stingray DDS info: %w", pathName, err)
			}
			img, err = dds.Decode(r, false)
			if err != nil {
				return fmt.Errorf("material texture %v: loading DDS image: %w", pathName, err)
			}
		}

		var infoB strings.Builder
		fmt.Fprintf(&infoB, "Usage=%v (%v/%v)\n", imageName, i+1, len(sortedTextureUsages))
		if path.Value != 0 {
			fmt.Fprintf(&infoB, "Size=(%v,%v)\nFormat=%v\nPath=%v\n", img.Bounds().Dx(), img.Bounds().Dy(), img.Info.DXT10Header.DXGIFormat, path)
		} else {
			fmt.Fprintf(&infoB, "Usage=N/A (0/0)\nSize=N/A\nFormat=N/A\nPath=N/A\n")
		}
		if img != nil {
			imgsToLoad = append(imgsToLoad, img.Image)
		} else {
			imgsToLoad = append(imgsToLoad, nil)
		}
		imgInfoTexts = append(imgInfoTexts, infoB.String())
	}

	for key, value := range mat.Settings {
		keyName, ok := thinhashes[key]
		if !ok {
			keyName = "unknown setting: " + key.String()
		}
		pv.settings[keyName] = value
		pv.settingKeys = append(pv.settingKeys, keyName)
	}
	slices.Sort(pv.settingKeys)

	pv.baseMaterial = mat.BaseMaterial
	if bmName, ok := hashes[mat.BaseMaterial]; ok {
		pv.baseMaterialName = bmName
	} else {
		pv.baseMaterialName = mat.BaseMaterial.String()
	}

	pv.Flags = LinearFilteringButton | IgnoreAlphaButton | MultipleImages
	pv.LoadImages(imgsToLoad)
	pv.Alt = "<material has no textures>"
	infoText := fmt.Sprintf("Num Settings=%v\n", len(pv.settings))
	pv.DrawInfo = func() {
		imgui.TextUnformatted(infoText)
		imgui.TextUnformatted("Base material:")
		imgui.SameLine()
		if (pv.baseMaterial == stingray.Hash{}) {
			imgui.TextUnformatted("none")
		} else {
			fileID := stingray.FileID{
				Name: pv.baseMaterial,
				Type: stingray.Sum("material"),
			}
			widgets.GamefileLinkTextF(fileID, "%v", pv.baseMaterialName)
		}
	}
	for i := range pv.Images {
		pv.Images[i].DrawInfo = func() { imgui.TextUnformatted(imgInfoTexts[i]) }
		pv.Images[i].Alt = "<no texture>"
	}

	return nil
}

func (pv *MaterialPreview) SettingsEmpty() bool {
	return len(pv.settings) == 0
}

func (pv *MaterialPreview) SettingsVisible() bool {
	return pv.settingsVisible
}

func (pv *MaterialPreview) SetSettingsVisible(visible bool) {
	pv.settingsVisible = visible
}

func (pv *MaterialPreview) DrawSettings() {
	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY | imgui.TableFlagsRowBg
	if imgui.BeginTableV("##Material Settings", 2, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupColumnV("Value", imgui.TableColumnFlagsWidthStretch, 2, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		for _, id := range pv.settingKeys {
			imgui.PushIDStr(id)

			imgui.TableNextColumn()
			imutils.CopyableTextf("%v", id)

			imgui.TableNextColumn()
			settingValue := pv.settings[id]
			formatted := make([]string, len(settingValue))
			for i := range settingValue {
				formatted[i] = fmt.Sprintf("%.3f", settingValue[i])
			}
			settingString := strings.Join(formatted, ", ")
			if len(settingValue) > 1 {
				settingString = "(" + settingString + ")"
			}
			imgui.TextUnformatted(settingString)

			imgui.PopID()
		}
		imgui.EndTable()
	}
}
