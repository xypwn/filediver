package widgets

import (
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/stingray"
)

type hashViewer struct {
	name   string
	format string
}

type HashConverterState struct {
	HaveValue    bool
	Value        stingray.Hash
	HexInput     string
	DecimalInput string
	inputErr     error
	viewers      []hashViewer
}

func NewHashConverter() *HashConverterState {
	return &HashConverterState{
		viewers: []hashViewer{
			{
				name:   "Decimal (HD2SDK/Audio Modder)",
				format: "%d",
			},
			{
				name:   "Hex (Filediver)",
				format: "0x%016x",
			},
			{
				name:   "Hex (Diver)",
				format: "0x%x",
			},
		},
	}
}

var isHex = [256]bool{
	'x': true, 'X': true,
	'0': true, '1': true, '2': true, '3': true, '4': true, '5': true, '6': true, '7': true, '8': true, '9': true,
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true,
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true,
}

var hexCharFilter = func(data imgui.InputTextCallbackData) int {
	if data.EventFlag()&imgui.InputTextFlagsCallbackCharFilter == 0 {
		return 0
	}
	if c := data.EventChar(); c <= 0x7f && isHex[c] {
		return 0
	} else {
		return 1
	}
}

func DrawHashConverter(s *HashConverterState) {
	gray := imgui.NewVec4(0.8, 0.8, 0.8, 1)
	red := imgui.NewVec4(0.8, 0.5, 0.5, 1)
	if imgui.InputTextWithHint("Hex", "e.g. 0x01234567abcdabcd (0x optional)", &s.HexInput, imgui.InputTextFlagsCallbackCharFilter, hexCharFilter) {
		s.DecimalInput = ""
		if s.HexInput == "" || s.HexInput == "0x" {
			s.HaveValue = false
			s.inputErr = nil
		} else {
			s.Value, s.inputErr = stingray.ParseHash(s.HexInput)
			s.HaveValue = s.inputErr == nil
		}
	}
	if imgui.InputTextWithHint("Decimal", "e.g. 123456789000000000", &s.DecimalInput, imgui.InputTextFlagsCharsDecimal, nil) {
		s.HexInput = ""
		if s.DecimalInput == "" {
			s.HaveValue = false
			s.inputErr = nil
		} else {
			s.Value.Value, s.inputErr = strconv.ParseUint(s.DecimalInput, 10, 64)
			s.HaveValue = s.inputErr == nil
		}
	}
	if s.inputErr != nil {
		imgui.PushTextWrapPos()
		imutils.Textcf(red, "Input error: %v", s.inputErr)
		imgui.PopTextWrapPos()
	}
	if s.HaveValue {
		imgui.Separator()
		if imgui.BeginTable("##HashConvResultTable", 2) {
			for i, viewer := range s.viewers {
				imgui.PushIDInt(int32(i))
				imgui.TableNextColumn()
				imutils.Textcf(gray, "%s", viewer.name)
				imgui.TableNextColumn()
				imutils.CopyableTextf(viewer.format, s.Value.Value)
				imgui.PopID()
			}
			imgui.EndTable()
		}
	}
}
