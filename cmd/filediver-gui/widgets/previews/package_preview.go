package previews

import (
	"regexp"
	"slices"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/textutils"
	"github.com/xypwn/filediver/cmd/filediver-gui/widgets"
	datalib "github.com/xypwn/filediver/datalibrary"
	"github.com/xypwn/filediver/stingray"
	stingray_package "github.com/xypwn/filediver/stingray/package"
)

type PackagePreview struct {
	*stingray_package.Package
	queryBuf     string
	shownFiles   []stingray.FileID
	stringNames  map[stingray.FileID]string
	stringHashes map[stingray.FileID]string
	packageId    stingray.FileID

	lookupHash  func(stingray.Hash) string
	getArmorSet func(x stingray.Hash) (datalib.ArmorSet, bool)
}

func NewPackagePreview(lookupHash func(stingray.Hash) string, getArmorSet func(stingray.Hash) (datalib.ArmorSet, bool)) *PackagePreview {
	return &PackagePreview{
		shownFiles:   make([]stingray.FileID, 0),
		stringNames:  make(map[stingray.FileID]string),
		stringHashes: make(map[stingray.FileID]string),
		lookupHash:   lookupHash,
		getArmorSet:  getArmorSet,
	}
}

func (pv *PackagePreview) updateShownFiles() {
	pv.shownFiles = pv.shownFiles[:0]
	for _, item := range pv.Package.Items {
		fileID := stingray.NewFileID(item.Name, item.Type)
		if textutils.QueryMatchesAny(pv.queryBuf, pv.stringNames[fileID], pv.stringHashes[fileID]) {
			pv.shownFiles = append(pv.shownFiles, fileID)
		}
	}

	slices.SortFunc(pv.shownFiles, func(a, b stingray.FileID) int {
		return strings.Compare(pv.stringNames[a], pv.stringNames[b])
	})
}

var colorRegex = regexp.MustCompile("<c[=#A-Z]+>(.*?)<\\/c>")

func (pv *PackagePreview) drawArmorSet(armorSet datalib.ArmorSet) {
	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsRowBg
	if imgui.BeginTableV("##Armor Set", 2, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV(armorSet.Type.String()+" Info", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupColumnV("", imgui.TableColumnFlagsWidthStretch, 2, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		imgui.PushIDStr("Armor Name Row")
		imgui.TableNextRow()
		imgui.TableSetColumnIndex(0)
		imgui.Text("Name")
		imgui.TableSetColumnIndex(1)
		imutils.CopyableTextf("%v##value", armorSet.Name)
		imgui.PopID()

		imgui.PushIDStr("Armor Type Row")
		imgui.TableNextRow()
		imgui.TableSetColumnIndex(0)
		imgui.Text("Type")
		imgui.TableSetColumnIndex(1)
		imutils.Textf("%v##value", armorSet.Type.String())
		imgui.PopID()

		imgui.PushIDStr("Armor Icon Row")
		imgui.TableNextRow()
		imgui.TableSetColumnIndex(0)
		imgui.Text("Icon")
		imgui.TableSetColumnIndex(1)
		widgets.GamefileLinkTextF(stingray.NewFileID(armorSet.Passive.IconHash, stingray.Sum("texture")), "%v.texture", pv.lookupHash(armorSet.Passive.IconHash))
		imgui.PopID()

		imgui.PushIDStr("Armor Passive Row")
		imgui.TableNextRow()
		imgui.TableSetColumnIndex(0)
		imgui.Text("Passive")
		imgui.TableSetColumnIndex(1)
		imutils.Textf("%v##value", armorSet.Passive.Name)

		imgui.Indent()
		for _, desc := range armorSet.Passive.ResolveDescription() {
			if len(desc) == 0 {
				continue
			}
			matches := colorRegex.FindAllStringSubmatchIndex(desc, -1)
			if matches == nil {
				imutils.Textf("%v", desc)
				continue
			}
			offset := 0
			for i := 0; i < len(matches[0]); i += 2 {
				match := matches[0][i : i+2]
				if i == 0 {
					imutils.Textf("%v", desc[offset:match[0]])
					offset = match[1]
					continue
				}
				imgui.SameLineV(0.0, 0.0)
				imutils.Textcf(imgui.ColorConvertU32ToFloat4(0xff10e7ff), "%v", desc[match[0]:match[1]])
			}
			if offset < len(desc) {
				imgui.SameLineV(0.0, 0.0)
				imutils.Textf("%v", desc[offset:])
			}
		}
		imgui.Unindent()
		imgui.PopID()

		imgui.EndTable()
	}
}

func (pv *PackagePreview) Draw(name string) {
	imgui.PushIDStr(name)
	defer imgui.PopID()

	widgets.ArchiveLinkTextF(pv.packageId, fnt.I.Package+" %v.%v", pv.lookupHash(pv.packageId.Name), pv.lookupHash(pv.packageId.Type))
	imutils.Textf("File Count: %v", pv.Package.FileCount)
	armorSet, ok := pv.getArmorSet(pv.packageId.Name)
	if ok {
		pv.drawArmorSet(armorSet)
	}
	imgui.SetNextItemWidth(imgui.ContentRegionAvail().X)
	if imgui.InputTextWithHint("##Search Files", fnt.I.Search+" Filter by hash or string...", &pv.queryBuf, 0, nil) {
		pv.updateShownFiles()
	}
	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsScrollY | imgui.TableFlagsRowBg
	if imgui.BeginTableV("##Files Preview", 1, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Files", imgui.TableColumnFlagsWidthStretch, 1, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()

		for _, id := range pv.shownFiles {
			value := pv.stringNames[id]
			imgui.PushIDStr(value)
			imgui.TableNextColumn()
			widgets.GamefileLinkTextF(id, "%v", value)
			imgui.PopID()
		}
		imgui.EndTable()
	}
}

func (pv *PackagePreview) Load(data *stingray_package.Package, name stingray.FileID) {
	pv.Package = data
	pv.packageId = name
	clear(pv.stringHashes)
	clear(pv.stringNames)
	for _, item := range pv.Package.Items {
		fileID := stingray.NewFileID(item.Name, item.Type)
		pv.stringNames[fileID] = strings.ToLower(pv.lookupHash(fileID.Name) + "." + pv.lookupHash(fileID.Type))
		pv.stringHashes[fileID] = fileID.Name.String() + "." + fileID.Type.String()
	}
	pv.updateShownFiles()
}
