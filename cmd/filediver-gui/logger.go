package main

import (
	"fmt"
	"math"
	"runtime/debug"
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
	items        []logItem
	status       string
	numErrs      int
	numWarns     int
	haveFatalErr bool

	scrollToBottom bool
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

// NOTE: Fatalf can't stop control flow by itself here. That has to be done externally.
// Fatalf will also print a stack trace.
func (l *Logger) Infof(f string, a ...any)  { l.add("INFO", 0.8, 0.8, 0.8, f, a...) }
func (l *Logger) Warnf(f string, a ...any)  { l.add("WARNING", 0.8, 0.8, 0, f, a...); l.numWarns++ }
func (l *Logger) Errorf(f string, a ...any) { l.add("ERROR", 0.8, 0.5, 0.5, f, a...); l.numErrs++ }
func (l *Logger) Fatalf(f string, a ...any) {
	l.add("FATAL ERROR", 0.9, 0.3, 0.3, f, a...)
	l.add("STACK TRACE", 0.9, 0.3, 0.3, "%s", debug.Stack())
	l.numErrs++
	l.haveFatalErr = true
}
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

func (l *Logger) HaveFatalErr() bool {
	l.Lock()
	defer l.Unlock()

	return l.haveFatalErr
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
	l.haveFatalErr = false
}

func LogView(l *Logger) {
	l.Lock()
	defer l.Unlock()

	avail := imgui.ContentRegionAvail()

	var statusText string
	if l.status != "" {
		statusText = "STATUS: " + l.status
	}

	{
		style := imgui.CurrentStyle()
		avail.Y -= 1 // separator
		avail.Y -= style.ItemSpacing().Y
		if l.status != "" {
			avail.Y -= imgui.CalcTextSizeV(statusText, false, avail.X).Y
			avail.Y -= style.ItemSpacing().Y
		}
		avail.Y -= imgui.FrameHeight()
		avail.Y -= style.ItemSpacing().Y
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
		if imgui.ScrollY() == imgui.ScrollMaxY() {
			l.scrollToBottom = true
		}
		{
			scrolledUp := imgui.IsWindowHovered() && imgui.CurrentIO().MouseWheel() > 0
			win := imgui.InternalCurrentWindow()
			activeID := imgui.InternalActiveID()
			scrollbarActive := activeID != 0 && activeID == imgui.InternalWindowScrollbarID(win, imgui.AxisY)
			if scrolledUp || scrollbarActive {
				l.scrollToBottom = false
			}
		}
		if l.scrollToBottom {
			imgui.SetScrollYFloat(imgui.ScrollMaxY())
		}
	}
	imgui.EndChild()

	imgui.Separator()
	if l.status != "" {
		imgui.PushTextWrapPos()
		imgui.TextUnformatted(statusText)
	}
	imgui.BeginDisabledV(l.scrollToBottom)
	if imgui.ButtonV("Scroll to bottom", imgui.NewVec2(-math.SmallestNonzeroFloat32, 0)) {
		l.scrollToBottom = true
	}
	if l.scrollToBottom {
		imgui.SetItemTooltip("(scroll up in the logs to stop auto-scrolling)")
	} else {
		imgui.SetItemTooltip("Click to scroll to bottom and keep auto-scrolling")
	}
	imgui.EndDisabled()
}
