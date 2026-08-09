package previews

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/cmd/filediver-gui/imgui_wrapper"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
)

func randomString() string {
	var buf [8]byte
	rand.Read(buf[:])
	return base64.StdEncoding.EncodeToString(buf[:])
}

type FontPreview struct {
	currFontId string
	inputText  string
}

func NewFontPreview() *FontPreview {
	pv := &FontPreview{}

	return pv
}

func (pv *FontPreview) Delete() {
	if pv.currFontId != "" {
		imgui_wrapper.RemoveFont(pv.currFontId)
	}
}

func (pv *FontPreview) Load(fontData []byte) error {
	imgui_wrapper.RemoveFont(pv.currFontId)

	fontId := "FontPreview#" + randomString()
	imgui_wrapper.AddFont(fontId, imgui_wrapper.FontSpec{TtfData: fontData})
	pv.currFontId = fontId
	return nil
}

func (pv *FontPreview) Draw(name string) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	font, ok := imgui_wrapper.GetFont(pv.currFontId)
	if !ok {
		return
	}
	//imutils.Textf("Debug name: %s", font.DebugName())
	imgui.PushFont(font, 0)
	imutils.PushFontScale(2)
	imgui.PushTextWrapPos()
	imgui.TextUnformatted("the quick brown fox jumps over the lazy dog")
	imgui.TextUnformatted("THE QUICK BROWN FOX JUMPS OVER THE LAZY DOG")
	imgui.SetNextItemWidth(imgui.ContentRegionAvail().X)
	imgui.InputTextWithHint(
		"##input",
		"Type something here",
		&pv.inputText,
		/*imgui.InputTextFlagsWordWrap|imgui.InputTextFlags(imgui.InputTextFlagsMultiline)*/ 0,
		nil,
	)
	imgui.PopTextWrapPos()
	imutils.PopFontScale()
	imgui.PopFont()
}
