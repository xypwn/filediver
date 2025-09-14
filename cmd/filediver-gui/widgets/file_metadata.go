package widgets

import (
	"fmt"
	"reflect"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/stingray"
)

func FileMetadata(meta app.FileMetadata) {
	drawValue := func(val reflect.Value) {
		drawHash := func(h stingray.Hash) {
			imutils.CopyableTextf("%v", h)
		}
		switch val := val.Interface().(type) {
		case []stingray.Hash:
			for i, h := range val {
				if i != 0 {
					imgui.SameLineV(0, 0)
					imutils.Textf(",")
				}
				drawHash(h)
			}
		case stingray.Hash:
			drawHash(val)
		case string, int, float64:
			imutils.Textf("%v", val)
		default:
			imutils.TextError(fmt.Errorf("unknown type %T", val))
		}
	}

	if imgui.BeginTableV("##FileMetadata", 3, imgui.TableFlagsBorders|imgui.TableFlagsRowBg|imgui.TableFlagsSizingFixedFit, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumnV("Name", 0, 0, 0)
		imgui.TableSetupColumnV("Description", 0, imutils.S(100), 0)
		imgui.TableSetupColumnV("Value", imgui.TableColumnFlagsWidthStretch, 0, 0)
		imgui.TableSetupScrollFreeze(0, 1)
		imgui.TableHeadersRow()
		val := reflect.ValueOf(meta)
		typ := val.Type()
		for i := range typ.NumField() {
			field := typ.Field(i)
			imgui.PushIDInt(int32(i))
			imgui.TableNextColumn()
			imutils.Textf("%s", field.Name)
			imgui.TableNextColumn()
			imgui.PushTextWrapPos()
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%s", field.Tag.Get("help"))
			imgui.TableNextColumn()
			drawValue(val.Field(i))
			imgui.PopTextWrapPos()
			imgui.PopID()
		}
		imgui.EndTable()
	}
}
