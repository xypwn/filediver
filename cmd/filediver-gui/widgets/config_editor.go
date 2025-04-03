package widgets

import (
	"errors"
	"maps"
	"slices"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/app"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
)

func ConfigEditor(template app.ConfigTemplate, config *app.Config) {
	reset := func() {
		*config = app.Config{}
		for extrName := range template.Extractors {
			(*config)[extrName] = map[string]string{}
		}
	}
	if *config == nil {
		reset()
	}

	editOption := func(convName, optName string) bool {
		changed := false
		strID := "##" + convName + optName
		opt := template.Extractors[convName].Options[optName]
		if opt.Type == app.ConfigValueEnum {
			if len(opt.Enum) == 2 && slices.Contains(opt.Enum, "false") && slices.Contains(opt.Enum, "true") {
				var val bool
				if v := (*config)[convName][optName]; v != "" {
					val = v == "true"
				} else {
					val = opt.Enum[0] == "true"
				}
				if imgui.Checkbox(strID, &val) {
					if val {
						(*config)[convName][optName] = "true"
					} else {
						(*config)[convName][optName] = "false"
					}
					changed = true
				}
			} else {
				selected := (*config)[convName][optName]
				if selected == "" {
					selected = opt.Enum[0]
				}

				if imgui.BeginCombo(strID, selected) {
					for _, val := range opt.Enum {
						isSelected := selected == val
						if imgui.SelectableBoolPtr(val, &isSelected) {
							(*config)[convName][optName] = val
						}
						if isSelected {
							imgui.SetItemDefaultFocus()
						}
					}
					imgui.EndCombo()
				}
			}
		} else if opt.Type == app.ConfigValueIntRange {
			enabled := (*config)[convName][optName] != ""
			imgui.BeginDisabledV(!enabled)
			var val int32
			sliderFormat := "no value"
			if enabled {
				i, _ := strconv.Atoi((*config)[convName][optName])
				val = int32(i)
				sliderFormat = "%d"
			}
			if imgui.SliderIntV(strID+" slider", &val, int32(opt.IntRangeMin), int32(opt.IntRangeMax), sliderFormat, imgui.SliderFlagsAlwaysClamp) {
				(*config)[convName][optName] = strconv.Itoa(int(val))
			}
			if enabled {
				imgui.SetItemTooltip(fnt.I("Lightbulb_2") + " Use Ctrl+Click to type in a value")
			} else {
				imgui.SetItemTooltip("Enable the checkbox first to set a value")
			}
			imgui.EndDisabled()

			imgui.SameLine()
			if imgui.Checkbox(strID+" check", &enabled) {
				if enabled {
					(*config)[convName][optName] = strconv.Itoa(opt.IntRangeMin)
				} else {
					(*config)[convName][optName] = ""
				}
			}
		} else {
			imutils.TextError(errors.New("unsupported option type"))
		}

		return changed
	}

	if imgui.Button("Reset config##Config editor") {
		reset()
	}
	const tableFlags = imgui.TableFlagsResizable | imgui.TableFlagsBorders | imgui.TableFlagsRowBg
	if imgui.BeginTableV("##Config editor", 3, tableFlags, imgui.NewVec2(0, 0), 0) {
		imgui.TableSetupColumn("Option")
		imgui.TableSetupColumn("Value")
		imgui.TableSetupColumnV("", imgui.TableColumnFlagsNoResize|imgui.TableColumnFlagsWidthFixed,
			imgui.CalcTextSize(fnt.I("Undo")).X+imgui.CurrentStyle().ItemSpacing().X,
			0)
		imgui.TableHeadersRow()

		for _, convName := range slices.Sorted(maps.Keys(template.Extractors)) {
			conv := template.Extractors[convName]
			imgui.TableNextColumn()
			open := imgui.TreeNodeExStrV(convName, imgui.TreeNodeFlagsSpanFullWidth|imgui.TreeNodeFlagsSpanAllColumns)
			imgui.TableNextColumn()
			imgui.TableNextColumn()
			if open {
				for _, optName := range slices.Sorted(maps.Keys(conv.Options)) {
					imgui.TableNextColumn()
					imgui.TreeNodeExStrV(optName, imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen)
					imgui.TableNextColumn()
					editOption(convName, optName)
					imgui.TableNextColumn()
					if (*config)[convName][optName] != "" {
						if imgui.Button(fnt.I("Undo") + "##" + convName + " " + optName) {
							(*config)[convName][optName] = ""
						}
						imgui.SetItemTooltip("Reset")
					}
				}
				imgui.TreePop()
			}
		}
		imgui.EndTable()
	}
}
