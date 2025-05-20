package widgets

import (
	"fmt"
	"math"
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

type FilterListSection[T comparable] struct {
	Title string // empty string for no title
	Items []T
}

// FilterListWindow shows a searchable list
// of items with checkboxes and a "deselect all"
// button.
// You may use NextWindowXXX before this function.
// Returns true if a checkbox was changed.
// Searches through label and tooltip.
func FilterListWindow[T comparable](title string, windowOpen *bool, searchHint string, queryBuf *string, sections []FilterListSection[T], sel *map[T]struct{}, label func(x T) string, tooltip func(x T) string) bool {
	if !*windowOpen {
		imgui.CurrentContext().NextWindowData().SetHasFlags(0)
		return false
	}

	defer imgui.End()
	if !imgui.BeginV(title, windowOpen, 0) {
		return false
	}

	changed := false

	imgui.BeginDisabledV(len(*sel) == 0)
	if imgui.Button("Deselect all") {
		for k := range *sel {
			delete(*sel, k)
		}
		changed = true
	}
	imgui.Separator()
	imgui.EndDisabled()

	type match struct {
		item    T
		label   string
		tooltip string
	}
	var matches []match

	imgui.SetNextItemWidth(-math.SmallestNonzeroFloat32)
	imgui.InputTextWithHint("##search", fnt.I("Search")+" "+searchHint, queryBuf, 0, nil)

	defer imgui.EndChild()
	if !imgui.BeginChildStr("##list") {
		return false
	}

	for _, sec := range sections {
		matches = matches[:0]
		for _, item := range sec.Items {
			lab := label(item)
			var tt string
			if tooltip != nil {
				tt = tooltip(item)
			}
			if *queryBuf == "" || strings.Contains(strings.ToLower(lab+" "+tt), strings.ToLower(*queryBuf)) {
				matches = append(matches, match{item, lab, tt})
			}
		}

		if sec.Title != "" && len(matches) > 0 {
			imgui.TextUnformatted(sec.Title)
		}
		for _, item := range matches {
			_, checked := (*sel)[item.item]
			if imgui.Checkbox(item.label, &checked) {
				if checked {
					(*sel)[item.item] = struct{}{}
				} else {
					delete(*sel, item.item)
				}
				changed = true
			}
			if item.tooltip != "" && imgui.IsItemHovered() {
				imgui.SetTooltip(item.tooltip)
			}
		}
	}
	return changed
}
