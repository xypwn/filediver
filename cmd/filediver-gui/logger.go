package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
)

type logItem struct {
	Color imgui.Vec4
	Text  string
}

type Logger struct {
	sync.Mutex
	items    []logItem
	status   string
	numErrs  int
	numWarns int
}

var _ = app.Printer(&Logger{})

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) add(prefix string, r, g, b float32, f string, a ...any) {
	l.Lock()
	l.items = append(l.items, logItem{imgui.NewVec4(r, g, b, 1), prefix + ": " + fmt.Sprintf(f, a...)})
	l.Unlock()
}

func (l *Logger) setStatus(s string) {
	l.Lock()
	l.status = s
	l.Unlock()
}

// NOTE: Fatalf isn't really fatal here, since that wouldn't make sense in the context of a GUI.
func (l *Logger) Infof(f string, a ...any)   { l.add("INFO", 0.8, 0.8, 0.8, f, a...) }
func (l *Logger) Warnf(f string, a ...any)   { l.add("WARNING", 0.8, 0.8, 0, f, a...); l.numWarns++ }
func (l *Logger) Errorf(f string, a ...any)  { l.add("ERROR", 0.8, 0.5, 0.5, f, a...); l.numErrs++ }
func (l *Logger) Fatalf(f string, a ...any)  { l.add("FATAL ERROR", 0.9, 0, 0, f, a...); l.numErrs++ }
func (l *Logger) Statusf(f string, a ...any) { l.setStatus(fmt.Sprintf(f, a...)) }
func (l *Logger) NoStatus()                  { l.setStatus("") }

func (l *Logger) NumItems() int {
	l.Lock()
	defer l.Unlock()

	return len(l.items)
}

func (l *Logger) NumErrs() int {
	l.Lock()
	defer l.Unlock()

	return l.numErrs
}

func (l *Logger) NumWarns() int {
	l.Lock()
	defer l.Unlock()

	return l.numWarns
}

func (l *Logger) String() string {
	l.Lock()
	defer l.Unlock()

	var s strings.Builder
	for _, item := range l.items {
		s.WriteString(item.Text)
		s.WriteByte('\n')
	}
	return s.String()
}

func (l *Logger) Reset() {
	l.Lock()
	defer l.Unlock()

	l.items = nil
	l.status = ""
	l.numErrs = 0
	l.numWarns = 0
}

func LogView(l *Logger) {
	l.Lock()
	defer l.Unlock()

	avail := imgui.ContentRegionAvail()

	var statusText string
	if l.status != "" {
		statusText = "STATUS: " + l.status
	}

	if l.status != "" {
		style := imgui.CurrentStyle()
		statusSize := imgui.CalcTextSizeV(statusText, false, avail.X)
		statusSize.Y += style.ItemSpacing().Y
		statusSize.Y += 1 // separator
		statusSize.Y += style.ItemSpacing().Y
		avail.Y -= statusSize.Y
	}
	imgui.SetNextWindowSize(avail)
	if imgui.BeginChildStr("Logs") {
		for i, item := range l.items {
			imgui.PushIDInt(int32(i))
			imgui.PushTextWrapPos()
			imutils.CopyableTextcfV(item.Color, "Click to copy this item to clipboard", "%v", item.Text)
			imgui.PopID()
		}
		if len(l.items) == 0 {
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "Nothing here yet")
		}
	}
	imgui.EndChild()

	if l.status != "" {
		imgui.Separator()
		imgui.PushTextWrapPos()
		imgui.TextUnformatted(statusText)
	}
}
