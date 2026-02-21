package imgui_wrapper

import "github.com/AllenDang/cimgui-go/imgui"

func makeStyle(guiScale float32) *imgui.Style {
	style := imgui.NewStyle()
	imgui.StyleColorsDarkV(style)
	style.SetFrameRounding(2)
	style.SetWindowRounding(4)
	style.SetChildRounding(4)
	style.SetPopupRounding(4)
	style.SetGrabRounding(2)
	style.SetScrollbarRounding(2)

	style.SetFontScaleDpi(guiScale)
	style.ScaleAllSizes(guiScale)
	return style
}
