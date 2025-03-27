// ImGui utilities.
package imutils

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
)

// This use of a global is kind of
// ugly, but eh, it works.
var errorCopiedTime float64

func TextError(err error) {
	imgui.PushStyleColorVec4(imgui.ColText, imgui.NewVec4(0.8, 0.5, 0.5, 1))
	imgui.PushTextWrapPos()
	textPos := imgui.CursorScreenPos()
	Textf(fnt.I("Error")+" Error: %v", err)
	imgui.SetCursorScreenPos(textPos)
	imgui.SetNextItemAllowOverlap()
	clicked := imgui.InvisibleButton("##Error text", imgui.ItemRectSize())
	imgui.PopStyleColor()
	if imgui.BeginItemTooltip() {
		if imgui.Time()-errorCopiedTime < 1 {
			Textf(fnt.I("Check") + " Copied")
		} else {
			Textf(fnt.I("Content_copy") + " Click to copy error to clipboard")
		}
		imgui.EndTooltip()
	}
	if clicked {
		imgui.SetClipboardText(fmt.Sprintf("Error: %v", err))
		errorCopiedTime = imgui.Time()
	}
}

func Textf(format string, args ...any) {
	imgui.TextUnformatted(fmt.Sprintf(format, args...))
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
