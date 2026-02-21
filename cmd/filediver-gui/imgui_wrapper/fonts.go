package imgui_wrapper

import (
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"

	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
)

// Current fonts.
var FontDefault *imgui.Font
var FontMono *imgui.Font

func setupFonts(needCJKFonts bool) {
	io := imgui.CurrentIO()
	fonts := io.Fonts()
	fonts.Clear()

	fontSize := float32(16)
	type fontSpec struct {
		scale       float32
		glyphRange  *imgui.Wchar
		ttfData     []byte
		extraConfig func(*imgui.FontConfig)
	}
	var fontSpecs []fontSpec
	// Base font
	fontSpecs = append(fontSpecs, fontSpec{
		scale: 1,
		//glyphRange: &baseGlyphRanges[0],
		ttfData: fnt.TextFont,
	})
	if needCJKFonts {
		fontSpecs = append(fontSpecs,
			// Japanese
			fontSpec{
				scale: 1.2,
				//glyphRange: (&imgui.FontAtlas{}).GlyphRangesJapanese(),
				ttfData: fnt.TextFontJP,
			},
			// Korean
			fontSpec{
				scale: 1.2,
				//glyphRange: (&imgui.FontAtlas{}).GlyphRangesKorean(),
				ttfData: fnt.TextFontKR,
			},
			// Chinese
			fontSpec{
				scale: 1.2,
				//glyphRange: (&imgui.FontAtlas{}).GlyphRangesChineseFull(),
				ttfData: fnt.TextFontCN,
			},
		)
	}
	// Icons
	fontSpecs = append(fontSpecs, fontSpec{
		scale:   1,
		ttfData: fnt.IconFont,
		extraConfig: func(fc *imgui.FontConfig) {
			fc.SetGlyphOffset(imgui.NewVec2(0, (0.2)*fontSize))
			fc.SetGlyphMinAdvanceX(1 * fontSize)
		},
	})
	for i, spec := range fontSpecs {
		fc := imgui.NewFontConfig()
		if i != 0 {
			fc.SetMergeMode(true)
		}
		fc.SetFontDataOwnedByAtlas(false)
		if spec.extraConfig != nil {
			spec.extraConfig(fc)
		}
		newFont := fonts.AddFontFromMemoryTTFV(
			uintptr(unsafe.Pointer(&spec.ttfData[0])),
			int32(len(spec.ttfData)),
			fontSize*spec.scale,
			fc,
			spec.glyphRange,
		)
		fc.Destroy()
		if i == 0 {
			FontDefault = newFont
		}
	}

	{ // Monospace
		fc := imgui.NewFontConfig()
		fc.SetFontDataOwnedByAtlas(false)
		FontMono = fonts.AddFontFromMemoryTTFV(
			uintptr(unsafe.Pointer(&fnt.TextFontMono[0])),
			int32(len(fnt.TextFontMono)),
			fontSize,
			fc,
			fonts.GlyphRangesDefault(),
		)
		fc.Destroy()
	}
}
