package main

import (
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/stingray"
)

type HashBulkImporterMode int

const (
	HashBulkImporterHex HashBulkImporterMode = iota
	HashBulkImporterDecimal
)

type HashBulkImporter struct {
	Text               string
	Hashes             []stingray.Hash
	DetectedHashesText string
	Mode               HashBulkImporterMode
}

// Returns true if w.Hashes should be imported.
func (w *HashBulkImporter) Draw(justGotOpened bool) bool {
	if justGotOpened {
		w.Text = imgui.ClipboardText()
	}

	changed := justGotOpened
	changed = imutils.ComboChoiceAny("Mode", &w.Mode,
		[]HashBulkImporterMode{HashBulkImporterHex, HashBulkImporterDecimal},
		func(a, b HashBulkImporterMode) bool { return a == b },
		func(m HashBulkImporterMode) string {
			switch m {
			case HashBulkImporterHex:
				return "Hex (e.g. 123abc... or 0x123abc...)"
			case HashBulkImporterDecimal:
				return "Decimal (e.g. 123456789000000000)"
			default:
				return "unknown"
			}
		}) || changed
	imgui.Separator()
	size := imutils.SVec2(400, 200)
	changed = imgui.InputTextMultiline("##Input", &w.Text, size, imgui.InputTextFlagsWordWrap, nil) || changed
	if changed {
		var digits string
		switch w.Mode {
		case HashBulkImporterHex:
			digits = "0123456789ABCDEFabcdef"
		case HashBulkImporterDecimal:
			digits = "0123456789"
		}
		w.Hashes = w.Hashes[:0]
		for s := range strings.FieldsFuncSeq(w.Text, func(r rune) bool { return !strings.ContainsRune(digits, r) }) {
			if len(s) > 2 {
				h, err := stingray.ParseHash(s)
				if err != nil {
					continue
				}
				w.Hashes = append(w.Hashes, h)
			}
		}
		var detected strings.Builder
		for i, h := range w.Hashes {
			if i != 0 {
				detected.WriteString(", ")
			}
			detected.WriteString(h.String())
		}
		w.DetectedHashesText = detected.String()
	}
	imgui.PushTextWrapPos()
	if len(w.Hashes) > 0 {
		imutils.Textf("Detected %d hashes: %s", len(w.Hashes), w.DetectedHashesText)
	} else {
		imutils.Textf("No hashes detected")
	}
	imgui.PopTextWrapPos()
	imgui.Separator()
	return imgui.ButtonV("Import", imgui.NewVec2(imgui.ContentRegionAvail().X, 0))
}
