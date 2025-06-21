package imgui_wrapper

import (
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"

	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
)

var baseGlyphRanges = [...]imgui.Wchar{
	0x0020, 0x00ff, // basic latin + supplement
	0x0100, 0x017f, // latin extended-A
	0x2000, 0x206f, // general punctuation
	0x0400, 0x052f, // cyrillic + cyrillic supplement
	0,
}

var iconGlyphRanges = [...]imgui.Wchar{
	imgui.Wchar(fnt.IconFontInfo.Min),
	imgui.Wchar(fnt.IconFontInfo.Max16),
	0,
}

func updateFonts(guiScale float32, needCJKFonts bool) {
	io := imgui.CurrentIO()
	fonts := io.Fonts()
	fonts.Clear()

	style := imgui.NewStyle()

	fontSize := 16 * guiScale
	type fontSpec struct {
		scale       float32
		glyphRange  *imgui.Wchar
		ttfData     []byte
		extraConfig func(*imgui.FontConfig)
	}
	var fontSpecs []fontSpec
	// Base font
	fontSpecs = append(fontSpecs, fontSpec{
		scale:      1,
		glyphRange: &baseGlyphRanges[0],
		ttfData:    fnt.TextFont,
	})
	if needCJKFonts {
		fontSpecs = append(fontSpecs,
			// Japanese
			fontSpec{
				scale:      1.2,
				glyphRange: (&imgui.FontAtlas{}).GlyphRangesJapanese(),
				ttfData:    fnt.TextFontJP,
			},
			// Korean
			fontSpec{
				scale:      1.2,
				glyphRange: (&imgui.FontAtlas{}).GlyphRangesKorean(),
				ttfData:    fnt.TextFontKR,
			},
			// Chinese
			fontSpec{
				scale:      1.2,
				glyphRange: (&imgui.FontAtlas{}).GlyphRangesChineseFull(),
				ttfData:    fnt.TextFontCN,
			},
		)
	}
	// Icons
	fontSpecs = append(fontSpecs, fontSpec{
		scale:      1,
		glyphRange: &iconGlyphRanges[0],
		ttfData:    fnt.IconFont,
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
		fonts.AddFontFromMemoryTTFV(
			uintptr(unsafe.Pointer(&spec.ttfData[0])),
			int32(len(spec.ttfData)),
			fontSize*spec.scale,
			fc,
			spec.glyphRange,
		)
		fc.Destroy()
	}

	imguiDestroyFontsTexture()

	io.SetFontGlobalScale(1)
	style.ScaleAllSizes(guiScale)
	io.Ctx().SetStyle(*style)
}
