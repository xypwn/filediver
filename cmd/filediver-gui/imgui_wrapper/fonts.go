package imgui_wrapper

import (
	"cmp"
	"runtime"
	"slices"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"

	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
)

type FontSpec struct {
	Scale        float32 // 1 is 100%; as an exception, 0 is treated like 1
	Order        int     // lower is loaded earlier in merge component
	TtfData      []byte
	ExtraConfig  func(fc *imgui.FontConfig, fontSize float32)
	MergeWithIds []string
}

// Current fonts.
func FontDefault() *imgui.Font {
	f, _ := GetFont("default")
	return f
}
func FontMono() *imgui.Font {
	f, _ := GetFont("monospace")
	return f
}

type font struct {
	Name string
	Spec FontSpec
	Font *imgui.Font
	// Merge ID adjacency list. Kept consistent with other font nodes.
	merge map[string]struct{}
}

var fontsInitialized bool
var fonts map[string]*font = make(map[string]*font)
var fontsToAdd map[string]*font = make(map[string]*font)
var fontsToRemove map[string]*font = make(map[string]*font)

func AddFont(id string, spec FontSpec) (overwrote bool) {
	overwrote = RemoveFont(id)
	f := &font{id, spec, nil, make(map[string]struct{})}
	for _, id := range spec.MergeWithIds {
		f.merge[id] = struct{}{}
	}
	fontsToAdd[id] = f
	return
}

func RemoveFont(id string) (removed bool) {
	if _, ok := fontsToAdd[id]; ok {
		delete(fontsToAdd, id)
		return true
	}
	if _, ok := fontsToRemove[id]; ok {
		return false
	}
	f, ok := fonts[id]
	if !ok {
		return false
	}
	fontsToRemove[id] = f
	return true
}

func GetFont(id string) (font *imgui.Font, ok bool) {
	f, ok := fonts[id]
	if !ok {
		return nil, false
	}
	return f.Font, true
}

func addDefaultFonts() {
	AddFont("default", FontSpec{
		Order:   -999,
		TtfData: fnt.TextFont,
	})
	AddFont("icons", FontSpec{
		TtfData: fnt.IconFont,
		ExtraConfig: func(fc *imgui.FontConfig, fontSize float32) {
			fc.SetGlyphOffset(imgui.NewVec2(0, (0.2)*fontSize))
			fc.SetGlyphMinAdvanceX(1 * fontSize)
		},
		MergeWithIds: []string{"default"},
	})
	AddFont("monospace", FontSpec{
		TtfData: fnt.TextFontMono,
	})
}

func addCjkFonts() {
	AddFont("japanese", FontSpec{
		Scale:        1.2,
		TtfData:      fnt.TextFontJP,
		MergeWithIds: []string{"default"},
	})
	AddFont("korean", FontSpec{
		Scale:        1.2,
		TtfData:      fnt.TextFontKR,
		MergeWithIds: []string{"default"},
	})
	AddFont("chinese", FontSpec{
		Scale:        1.2,
		TtfData:      fnt.TextFontCN,
		MergeWithIds: []string{"default"},
	})
}

func removeCjkFonts() {
	RemoveFont("japanese")
	RemoveFont("korean")
	RemoveFont("chinese")
}

func updateFonts() {
	if len(fontsToAdd) == 0 && len(fontsToRemove) == 0 {
		return
	}
	io := imgui.CurrentIO()
	igfonts := io.Fonts()
	if !fontsInitialized {
		igfonts.Clear()
		fontsInitialized = true
	}

	// Remove all fonts to be removed
	for _, f := range fontsToRemove {
		igfonts.RemoveFont(f.Font)
		delete(fonts, f.Name)
	}
	clear(fontsToRemove)

	// Add fonts to be added (which are partial at this point)
	// to be updated, which will rebuild them
	updateRoots := make(map[string]struct{})
	for _, f := range fontsToAdd {
		fonts[f.Name] = f
		updateRoots[f.Name] = struct{}{}
	}
	clear(fontsToAdd)

	// Complete any incomplete font merge graph adjacency lists
	for _, f := range fonts {
		for id := range f.merge {
			if _, ok := fonts[id]; ok {
				fonts[id].merge[f.Name] = struct{}{}
			}
		}
	}

	// Determine components of the merge graph that need updating
	var updateComponents [][]string
	{
		var component []string
		seen := make(map[string]struct{})
		var dfs func(id string)
		dfs = func(id string) {
			if _, ok := seen[id]; ok {
				return
			}
			component = append(component, id)
			seen[id] = struct{}{}
			for n := range fonts[id].merge {
				if _, ok := fonts[n]; ok {
					dfs(n)
				}
			}
		}
		for id := range updateRoots {
			component = nil
			dfs(id)
			if len(component) > 0 {
				updateComponents = append(updateComponents, component)
			}
		}
	}

	// We want a stable font order
	for _, component := range updateComponents {
		slices.SortFunc(component, func(a, b string) int {
			return cmp.Or(
				cmp.Compare(fonts[a].Spec.Order, fonts[b].Spec.Order),
				cmp.Compare(a, b),
			)
		})
	}
	slices.SortFunc(updateComponents, slices.Compare)

	for _, component := range updateComponents {
		// Delete first font in component
		// since they're merged, they're all
		// the same font.
		for _, n := range component {
			f := fonts[n]
			if f.Font != nil {
				igfonts.RemoveFont(f.Font)
				runtime.KeepAlive(f.Spec.TtfData)
				break
			}
		}
		for _, n := range component {
			f := fonts[n]
			f.Font = nil
		}

		for i, node := range component {
			f := fonts[node]
			fontSize := float32(16)
			fc := imgui.NewFontConfig()
			if i != 0 {
				fc.SetMergeMode(true)
			}
			fc.SetFontDataOwnedByAtlas(false)
			if f.Spec.ExtraConfig != nil {
				f.Spec.ExtraConfig(fc, fontSize)
			}
			scale := f.Spec.Scale
			if scale == 0 {
				scale = 1
			}
			f.Font = igfonts.AddFontFromMemoryTTFV(
				uintptr(unsafe.Pointer(&f.Spec.TtfData[0])),
				int32(len(f.Spec.TtfData)),
				fontSize*scale,
				fc,
				nil,
			)
			fc.Destroy()
		}
	}
}
