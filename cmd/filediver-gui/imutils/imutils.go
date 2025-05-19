// ImGui utilities.
package imutils

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
)

func TextError(err error) {
	imgui.PushTextWrapPos()
	CopyableTextcfV(
		imgui.NewVec4(0.8, 0.5, 0.5, 1),
		"Click to copy error to clipboard",
		"Error: %v",
		err,
	)
}

func CopyableTextf(format string, args ...any) {
	CopyableTextcf(imgui.Vec4{}, format, args...)
}

func CopyableTextcf(color imgui.Vec4, format string, args ...any) {
	CopyableTextcfV(color, "Click to copy to clipboard", format, args...)
}

func CopyableTextfV(tooltip string, format string, args ...any) {
	CopyableTextcfV(imgui.Vec4{}, tooltip, format, args...)
}

func CopyableTextcfV(color imgui.Vec4, tooltip string, format string, args ...any) {
	imgui.PushIDStr(format)
	imgui.PopID()

	ctx := imgui.CurrentContext()
	if color.Z != 0 {
		imgui.PushStyleColorVec4(imgui.ColText, color)
	}
	textPos := imgui.CursorScreenPos()
	Textf(format, args...)
	imgui.SetCursorScreenPos(textPos)
	imgui.SetNextItemAllowOverlap()
	textBtnID := imgui.IDStr("##Copyable text")
	clicked := imgui.InvisibleButton("##Copyable text", imgui.ItemRectSize())
	if color.Z != 0 {
		imgui.PopStyleColor()
	}
	if imgui.BeginItemTooltip() {
		if ctx.LastActiveId() == textBtnID && ctx.LastActiveIdTimer() < 1 {
			Textf(fnt.I("Check") + " Copied")
		} else {
			Textf(fnt.I("Content_copy") + " " + tooltip)
		}
		imgui.EndTooltip()
	}
	if clicked {
		imgui.SetClipboardText(fmt.Sprintf(format, args...))
	}
}

func Textf(format string, args ...any) {
	imgui.PushIDStr(format)
	imgui.TextUnformatted(fmt.Sprintf(format, args...))
	imgui.PopID()
}

func Textcf(color imgui.Vec4, format string, args ...any) {
	imgui.PushStyleColorVec4(imgui.ColText, color)
	Textf(format, args...)
	imgui.PopStyleColor()
}

func CheckboxHeight() float32 {
	// HACK: This is probably not accurate, but it seems
	// good enough so it's not noticeable for the user.
	style := imgui.CurrentStyle()
	return imgui.FrameHeight() + style.FramePadding().Y + style.ItemSpacing().Y
}

func ComboHeight() float32 {
	// HACK: This is probably not accurate, but it seems
	// good enough so it's not noticeable for the user.
	style := imgui.CurrentStyle()
	return imgui.FrameHeight() + style.ItemSpacing().Y
}

// ComboChoice creates a combo box where you can choose a value from choices.
// Uses fmt.Sprint to turn values into strings.
func ComboChoice[T comparable](label string, selected *T, choices []T) bool {
	changed := false
	if imgui.BeginCombo(label, fmt.Sprint(*selected)) {
		for _, val := range choices {
			isSelected := val == *selected
			if imgui.SelectableBoolPtr(fmt.Sprint(val), &isSelected) {
				*selected = val
				changed = true
			}
			if isSelected {
				imgui.SetItemDefaultFocus()
			}
		}
		imgui.EndCombo()
	}
	return changed
}
