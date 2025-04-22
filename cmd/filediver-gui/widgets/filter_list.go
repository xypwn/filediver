package widgets

import (
	"fmt"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
)

func FilterListButton[T comparable](title string, windowOpen *bool, sel map[T]struct{}) bool {
	imgui.PushIDStr(title)
	defer imgui.PopID()

	var label strings.Builder
	label.WriteString(fnt.I("Filter_list"))
	label.WriteString(" ")
	label.WriteString(title)
	if len(sel) > 0 {
		fmt.Fprintf(&label, " (%v)", len(sel))
	}
	label.WriteString(" ")
	if *windowOpen {
		label.WriteString(fnt.I("Close"))
	} else {
		label.WriteString(fnt.I("Open_in_new"))
	}
	pressed := imgui.Button(label.String())
	if pressed {
		*windowOpen = !*windowOpen
	}
	if *windowOpen {
		imgui.SetItemTooltip("Close " + title + " Filter Window")
	} else {
		imgui.SetItemTooltip("Open " + title + " Filter Window")
	}
	return pressed
}

// Returns true if a checkbox was changed.
// Searches through label and tooltip.
func FilterListWindow[T comparable](title string, windowOpen *bool, searchHint string, queryBuf *string, avail []T, headerIndices map[int]string, sel *map[T]struct{}, label func(x T) string, tooltip func(x T) string) bool {
	if !*windowOpen {
		return false
	}

	maxHeight := imgui.MainViewport().Size().Y
	imgui.SetNextWindowSizeV(imgui.NewVec2(0, maxHeight), imgui.CondOnce)
	defer imgui.End()
	if !imgui.BeginV(title, windowOpen, 0) {
		return false
	}

	changed := false

	imgui.BeginDisabledV(len(*sel) == 0)
	if imgui.Button("Reset") {
		for k := range *sel {
			delete(*sel, k)
		}
		changed = true
	}
	imgui.Separator()
	imgui.EndDisabled()

	imgui.InputTextWithHint("##search", fnt.I("Search")+" "+searchHint, queryBuf, 0, nil)
	for i, k := range avail {
		if headerIndices != nil {
			if header, ok := headerIndices[i]; ok {
				imgui.TextUnformatted(header)
			}
		}
		lab := label(k)
		var tt string
		if tooltip != nil {
			tt = tooltip(k)
		}
		if *queryBuf == "" || strings.Contains(strings.ToLower(lab+" "+tt), strings.ToLower(*queryBuf)) {
			_, checked := (*sel)[k]
			if imgui.Checkbox(lab, &checked) {
				if checked {
					(*sel)[k] = struct{}{}
				} else {
					delete(*sel, k)
				}
				changed = true
			}
			if tt != "" && imgui.IsItemHovered() {
				imgui.SetTooltip(tt)
			}
		}
	}
	return changed
}
