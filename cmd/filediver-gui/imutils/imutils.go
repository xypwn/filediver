// ImGui utilities.
package imutils

import (
	"fmt"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/adrg/xdg"
	"github.com/ncruces/zenity"
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
	defer imgui.PopID()

	if i := strings.Index(format, "##"); i != -1 {
		format = format[:i]
	}

	ctx := imgui.CurrentContext()
	if color.W != 0 {
		imgui.PushStyleColorVec4(imgui.ColText, color)
	}
	textPos := imgui.CursorScreenPos()
	Textf(format, args...)
	imgui.SetCursorScreenPos(textPos)
	imgui.SetNextItemAllowOverlap()
	textBtnID := imgui.IDStr("##Copyable text")
	size := imgui.ItemRectSize()
	var clicked bool
	if size.X > 0 && size.Y > 0 { // prevent crash in case of empty label
		clicked = imgui.InvisibleButton("##Copyable text", size)
	}
	if color.W != 0 {
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
	defer imgui.PopID()
	if i := strings.Index(format, "##"); i != -1 {
		format = format[:i]
	}
	imgui.TextUnformatted(fmt.Sprintf(format, args...))
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
	return ComboChoiceAny(
		label,
		selected,
		choices,
		func(a, b T) bool { return a == b },
		func(x T) string { return fmt.Sprint(x) },
	)
}

// ComboChoiceAny is like [ComboChoice], but it uses user-supplied isEqual and
// toString functions and therefore accepts arbitrary types.
func ComboChoiceAny[T any](label string, selected *T, choices []T, isEqual func(a, b T) bool, toString func(T) string) bool {
	changed := false
	if imgui.BeginCombo(label, toString(*selected)) {
		for _, val := range choices {
			isSelected := isEqual(val, *selected)
			if imgui.SelectableBoolPtr(toString(val), &isSelected) {
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

// PopupManager is for when you want the next popup to only open
// after the previous one is closed.
// Popup order is determined by the order [PopupManager.Popup]()
// initially is called in.
type PopupManager struct {
	// Name to whether open. Values may be freely
	// modified. Values may also be set before
	// the corresponding call to Popup() is made
	// without effecting popup order.
	Open map[string]bool

	order        map[string]int // name to position in order
	orderCounter int
}

func NewPopupManager() *PopupManager {
	return &PopupManager{
		Open:  map[string]bool{},
		order: map[string]int{},
	}
}

// Popup draws a popup window with the given name.
// Internally uses BeginPopupModal.
// Put content drawing code into the body() function body.
// Call close() to close the current popup.
func (m *PopupManager) Popup(name string, content func(close func()), flags imgui.WindowFlags, closeBtn bool) {
	if _, ok := m.order[name]; !ok {
		m.order[name] = m.orderCounter
		m.orderCounter++
	}

	if !m.Open[name] {
		return
	}

	selfPos := m.order[name]
	for n, pos := range m.order {
		if n != name && m.Open[n] && pos < selfPos {
			// If there's already a different popup with
			// a lower position open, don't show this one.
			return
		}
	}

	imgui.OpenPopupStr(name)
	isOpen := true
	var pOpen *bool
	if closeBtn {
		pOpen = &isOpen
	}
	if imgui.BeginPopupModalV(name, pOpen, flags) {
		close := func() {
			imgui.CloseCurrentPopup()
			m.Open[name] = false
		}
		content(close)
		imgui.EndPopup()
	}
	if !isOpen {
		m.Open[name] = false
	}
}

func FilePicker(label string, path *string, directory bool) (changed bool) {
	imgui.PushIDStr(label)
	defer imgui.PopID()

	pathName := *path
	if after, ok := strings.CutPrefix(pathName, xdg.Home); ok {
		pathName = "~" + after
	}

	var icon, tooltip string
	if directory {
		icon = fnt.I("Folder_open")
		tooltip = icon + " Choose folder"
	} else {
		icon = fnt.I("File_open")
		tooltip = icon + " Choose file"
	}

	imgui.PushStyleColorVec4(imgui.ColText, imgui.NewVec4(0.9, 0.9, 0.9, 1))
	if imgui.Button(icon + " " + pathName) {
		opts := []zenity.Option{
			zenity.Filename(*path),
		}
		if directory {
			opts = append(opts, zenity.Directory())
		}
		if newPath, err := zenity.SelectFile(opts...); err == nil {
			*path = newPath
			changed = true
		} else if err != zenity.ErrCanceled {
			panic("error creating file picker: " + err.Error())
		}
	}
	imgui.SetItemTooltip(tooltip)
	imgui.PopStyleColor()
	imgui.SameLine()
	imgui.TextUnformatted(label)

	return
}
