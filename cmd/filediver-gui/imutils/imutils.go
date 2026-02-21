// ImGui utilities.
package imutils

import (
	"fmt"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/adrg/xdg"
	"github.com/ncruces/zenity"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"golang.org/x/exp/constraints"
)

func TextError(err error) {
	imgui.PushTextWrapPos()
	CopyableTextfV(
		CopyableTextOptions{
			TooltipHovered: fnt.I("Content_copy") + "Click to copy error to clipboard",
			Color:          imgui.NewVec4(0.8, 0.5, 0.5, 1),
		},
		"Error: %v",
		err,
	)
}

type CopyableTextOptions struct {
	// Tooltip shown when hovering normally.
	// Default: fonts.I("Content_copy") + " Click to copy to clipboard"
	TooltipHovered string
	// Tooltip shown for 1s after copying to clipboard.
	// Default: fonts.I("Check") + " Copied"
	TooltipCopied string
	// Mouse button to copy to clipboard.
	// Default: left mouse button
	Btn imgui.MouseButton
	// Text to be copied into clipboard.
	// Default: Formatted text passed in.
	ClipboardText string
	// Text color. Zero-value (default)
	// indicates default text color.
	Color imgui.Vec4
}

// CopyableTextf is the same as [CopyableTextfV]
// with default options.
func CopyableTextf(format string, args ...any) {
	CopyableTextfV(CopyableTextOptions{}, format, args...)
}

// CopyableTextfV creates a text item, which can be clicked
// to copy to clipboard.
// Pass zero value into opts for default.
func CopyableTextfV(opts CopyableTextOptions, format string, args ...any) {
	if opts.Color != (imgui.Vec4{}) {
		imgui.PushStyleColorVec4(imgui.ColText, opts.Color)
	}
	Textf(format, args...)
	if opts.Color != (imgui.Vec4{}) {
		imgui.PopStyleColor()
	}

	clicked := imgui.IsItemClickedV(opts.Btn)
	copied := StickyActivateID(imgui.IDStr(format), clicked, 1)
	if imgui.BeginItemTooltip() {
		if copied {
			s := opts.TooltipCopied
			if opts.TooltipCopied == "" {
				s = fnt.I("Check") + " Copied"
			}
			Textf("%v", s)
		} else {
			s := opts.TooltipHovered
			if opts.TooltipHovered == "" {
				s = fnt.I("Content_copy") + " Click to copy to clipboard"
			}
			Textf("%v", s)
		}
		imgui.EndTooltip()
	}

	if clicked {
		s := opts.ClipboardText
		if s == "" {
			if i := strings.Index(format, "##"); i != -1 {
				format = format[:i]
			}
			s = fmt.Sprintf(format, args...)
		}
		imgui.SetClipboardText(s)
	}
}

var stickyActivateItem = struct {
	ID   imgui.ID
	Time float64
}{}

// StickyActivateID will set the given ID to be sticky-active. The ID doesn't
// have to belong to an existing item.
// (Note that this is different from ImGUI's builtin concept of activation.)
// Once activate is true, this function will keep returning true for the given
// time, or until a different item is sticky-activated.
func StickyActivateID(id imgui.ID, activate bool, time float64) (active bool) {
	if activate {
		stickyActivateItem.ID = id
		stickyActivateItem.Time = imgui.Time()
		return true
	} else if stickyActivateItem.ID == id &&
		imgui.Time()-stickyActivateItem.Time < time {
		return true
	}
	return false
}

// StickyActivate is like [StickyActivateID], but will automatically use
// the last item's ID.
func StickyActivate(activate bool, time float64) (active bool) {
	lastID := imgui.ID(imgui.CurrentContext().LastItemData().CData.ID)
	return StickyActivateID(lastID, activate, time)
}

func Textf(format string, args ...any) {
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
	defer func() {
		// Make sure we always clear the SetNextWindowXXX options.
		ctx := imgui.CurrentContext()
		nwd := ctx.NextWindowData()
		nwd.SetHasFlags(imgui.NextWindowDataFlagsNone)
		ctx.SetNextWindowData(nwd)
	}()

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

// Used for [S] and [SVec2].
var GlobalScale float32 = 1

// S scales the given number by the current GUI scale
// after converting it to float32.
//
// Use [SVec2] to make a scaled imgui.Vec2.
//
// Requires [GlobalScale] to be set.
func S[T constraints.Float | constraints.Integer](x T) float32 {
	return GlobalScale * float32(x)
}

// SVec2 creates an imgui.Vec2 scaled by the current
// GUI scale.
//
// See [S].
//
// Requires [GlobalScale] to be set.
func SVec2[X, Y constraints.Float | constraints.Integer](x X, y Y) imgui.Vec2 {
	return imgui.NewVec2(S(x), S(y))
}

func calcButtonsWidth(buttonTexts ...string) float32 {
	if len(buttonTexts) == 0 {
		return 0
	}
	var width float32
	style := imgui.CurrentStyle()
	for _, buttonText := range buttonTexts {
		width += imgui.CalcTextSize(buttonText).X
	}
	width += float32(2*len(buttonTexts)) * style.FramePadding().X
	width += float32((len(buttonTexts) - 1)) * style.ItemSpacing().X
	return width
}

// RightAlignButtons sets the cursor X so that the next line of buttons
// drawn with the given texts will be right-aligned.
//
// Does nothing if len(buttonTexts) == 0.
func RightAlignButtons(buttonTexts ...string) {
	if len(buttonTexts) == 0 {
		return
	}
	cursor := imgui.CursorScreenPos()
	cursor.X += imgui.ContentRegionAvail().X
	cursor.X -= calcButtonsWidth(buttonTexts...)
	imgui.SetCursorScreenPos(cursor)
}

// CenterAlignButtons sets the cursor X so that the next line of buttons
// drawn with the given texts will be center-aligned.
//
// Does nothing if len(buttonTexts) == 0.
func CenterAlignButtons(buttonTexts ...string) {
	if len(buttonTexts) == 0 {
		return
	}
	cursor := imgui.CursorScreenPos()
	cursor.X += imgui.ContentRegionAvail().X / 2
	cursor.X -= calcButtonsWidth(buttonTexts...) / 2
	imgui.SetCursorScreenPos(cursor)
}
