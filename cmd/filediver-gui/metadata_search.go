package main

import (
	"reflect"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/stingray"
)

func DrawMetadataSearchHelp() {
	gray := imgui.NewVec4(0.8, 0.8, 0.8, 1)
	yellow := imgui.NewVec4(1, 0.9, 0.1, 1)

	imutils.Textf(`Anything following a ? in the search query will be interpreted as a metadata search expression.`)

	imutils.Textf(`Example:`)
	imgui.SameLine()
	imgui.PushStyleColorVec4(imgui.ColText, gray)
	imutils.CopyableTextf(`? width == 512 and format == "BC1UNorm"`)
	imgui.PopStyleColor()

	imutils.Textf("Syntax:")
	imgui.SameLine()
	imgui.TextLinkOpenURLV("expr-lang", "https://expr-lang.org/docs/language-definition")
	imutils.Textf(` - hashes must be passed as strings`)
	imutils.Textf(` - value name casing is ignored`)
	imutils.Textf(` - casing is ignored when checking if strings are equal`)

	imutils.Textf("Available values:\n")
	if imgui.BeginTableV("##AvailableValues", 3, imgui.TableFlagsBorders|imgui.TableFlagsRowBg, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Name", 0, 0, 0)
		imgui.TableSetupColumnV("Type", 0, 0, 0)
		imgui.TableSetupColumnV("Description", 0, 0, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()
		typ := reflect.TypeFor[app.FileMetadata]()
		for i := range typ.NumField() {
			field := typ.Field(i)
			imgui.PushIDInt(int32(i))
			var typStr string
			if t, ok := field.Tag.Lookup("type"); ok {
				typStr = t
			} else {
				switch field.Type {
				case reflect.TypeFor[[]stingray.Hash]():
					typStr = "hashes"
				case reflect.TypeFor[stingray.Hash]():
					typStr = "hash"
				case reflect.TypeFor[string]():
					typStr = "string"
				case reflect.TypeFor[int](), reflect.TypeFor[float64]():
					typStr = "number"
				default:
					panic("unknown type")
				}
			}
			imgui.TableNextColumn()
			imutils.Textcf(yellow, "%s", field.Name)
			imgui.TableNextColumn()
			imutils.Textcf(gray, "%s", typStr)
			imgui.TableNextColumn()
			imutils.Textcf(gray, "%s", field.Tag.Get("help"))
			if example, ok := field.Tag.Lookup("example"); ok {
				imgui.SameLine()
				imutils.Textcf(gray, "(e.g. %s)", example)
			}
			imgui.PopID()
		}
		imgui.EndTable()
	}
}
