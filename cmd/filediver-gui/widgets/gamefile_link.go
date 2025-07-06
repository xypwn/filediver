package widgets

import (
	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/stingray"
)

var gamefileLinkFile stingray.FileID

// PopGamefileLinkFile gets the last clicked gamefile link and
// resets it.
// ok is returned true if and only if there is a game
// file to be opened.
func PopGamefileLinkFile() (_ stingray.FileID, ok bool) {
	if (gamefileLinkFile == stingray.FileID{}) {
		return stingray.FileID{}, false
	}
	defer func() { gamefileLinkFile = stingray.FileID{} }()
	return gamefileLinkFile, true
}

// GamefileLinkText is like GamefileLinkTextF, but displays
// the file name (format name.type) as text.
func GamefileLinkText(file stingray.FileID) {
	GamefileLinkTextF(file, "%v.%v", file.Name, file.Type)
}

// GamefileLinkTextF draws text that can be right-clicked to copy the given
// game file hash to clipboard, or left-clicked to jump to that file in the
// browser.
func GamefileLinkTextF(file stingray.FileID, format string, args ...any) {
	imgui.PushIDStr(format)
	defer imgui.PopID()

	imutils.CopyableTextfV(imutils.CopyableTextOptions{
		TooltipCopied: fnt.I("Check") + "Hash copied",
		TooltipHovered: "File hash: " + file.Name.String() + "\n" +
			fnt.I("Jump_to_element") + "Left-click to jump to this file in browser\n" +
			fnt.I("Content_copy") + " Right-click to copy to clipboard",
		Btn:           imgui.MouseButtonRight,
		ClipboardText: file.Name.String(),
	}, format, args...)
	if imgui.IsItemClicked() {
		gamefileLinkFile = file
	}
}
