package widgets

import (
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/cmd/filediver-gui/tasks"
)

func DrawTask(name string, ts tasks.TaskState) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	text := name
	if i := strings.Index(text, "##"); i != -1 {
		text = text[:i]
	}
	if ts.Status != "" {
		if text != "" {
			text += ": "
		}
		text += ts.Status
	}
	imgui.TextUnformatted(text)
	imgui.ProgressBar(float32(ts.Prog))
}
