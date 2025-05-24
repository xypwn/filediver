package widgets

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
)

type StringsPreviewState struct {
	queryBuf    string
	shownIDs    []uint32
	strings     map[uint32]string
	needCJKFont bool
}

func NewStringsPreview() *StringsPreviewState {
	return &StringsPreviewState{}
}

func (pv *StringsPreviewState) Load(data *stingray_strings.StingrayStrings) {
	pv.queryBuf = ""
	pv.strings = data.Strings
	pv.updateShownIDs()
	for _, v := range pv.strings {
		if strings.ContainsFunc(v, func(r rune) bool {
			// NOTE(xypwn): This isn't 100% accurate, but it should be good enough for our use case.
			return unicode.In(r, unicode.Unified_Ideograph, unicode.Hiragana, unicode.Katakana, unicode.Hangul)
		}) {
			pv.needCJKFont = true
		}
	}
}

func (pv *StringsPreviewState) NeedCJKFont() bool {
	return pv.needCJKFont
}

func (pv *StringsPreviewState) updateShownIDs() {
	pv.shownIDs = pv.shownIDs[:0]
	for _, id := range slices.Sorted(maps.Keys(pv.strings)) {
		match := fmt.Sprintf("%v %v", id, pv.strings[id])
		if strings.Contains(strings.ToLower(match), strings.ToLower(pv.queryBuf)) {
			pv.shownIDs = append(pv.shownIDs, id)
		}
	}
}

func StringsPreview(pv *StringsPreviewState) {
	imgui.SetNextItemWidth(imgui.ContentRegionAvail().X)
	if imgui.InputTextWithHint("##Search Strings", fnt.I("Search")+" Filter by ID or string...", &pv.queryBuf, 0, nil) {
		pv.updateShownIDs()
	}

	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY | imgui.TableFlagsRowBg
	if imgui.BeginTableV("##Strings Preview", 2, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("ID", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupColumnV("String", imgui.TableColumnFlagsWidthStretch, 2, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		for _, key := range pv.shownIDs {
			value := pv.strings[key]
			imgui.PushIDInt(int32(key))
			imgui.TableNextColumn()
			imutils.CopyableTextf("%v##key", key)
			imgui.TableNextColumn()
			imgui.PushTextWrapPos()
			imutils.CopyableTextf("%v##value", value)
			imgui.PopID()
		}
		imgui.EndTable()
	}
}
