package widgets

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/iancoleman/strcase"
	"github.com/xypwn/filediver/app/appconfig"
	fnt "github.com/xypwn/filediver/cmd/filediver-gui/fonts"
	"github.com/xypwn/filediver/cmd/filediver-gui/imutils"
	"github.com/xypwn/filediver/config"
)

func ConfigEditor(cfg *appconfig.Config, showAdvanced *bool) (changed bool) {
	dependsSatisfied, err := config.DependsSatisfied(cfg)
	if err != nil {
		imutils.TextError(err)
		return
	}

	configV := reflect.ValueOf(cfg).Elem()
	category := ""

	imgui.Checkbox("Show advanced options", showAdvanced)

	for _, field := range appconfig.ConfigFields.Fields {
		var tooltip string
		if dependsSatisfied[field.Name] {
			var affectedTypes []string
			for _, tag := range field.Tags {
				if after, ok := strings.CutPrefix(tag, "t:"); ok {
					affectedTypes = append(affectedTypes, after)
				}
			}
			help := field.Help
			if len(affectedTypes) != 0 {
				if help != "" {
					help += ", "
				}
				help += "affects: " + strings.Join(affectedTypes, ", ")
			}
			if help == "" {
				help = strcase.ToDelimited(field.Name, ' ')
			}
			tooltip = help
			switch field.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64:
				tooltip += " [ctrl+click to type in a value]"
			}
			if tooltip != "" {
				tooltip = strings.ToUpper(string(tooltip[0])) + tooltip[1:]
			}
		} else {
			var b strings.Builder
			fmt.Fprintf(&b, "Depends on: ")
			for i, dep := range field.Depends {
				if i != 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%v=%v", dep.Field, dep.Value)
			}
			tooltip = b.String()
		}

		if field.IsCategory {
			category = field.Name
			imgui.SeparatorText(category)
			imgui.SetItemTooltip(tooltip)
			continue
		} else if category == "" {
			category = "General"
			imgui.SeparatorText(category)
		}
		isAdvanced := slices.Contains(field.Tags, "advanced")
		if isAdvanced && !*showAdvanced {
			continue
		}

		imgui.PushIDStr(field.Name)
		imgui.BeginDisabledV(!dependsSatisfied[field.Name])

		valV := configV
		for name := range strings.SplitSeq(field.Name, ".") {
			valV = valV.FieldByName(name)
		}

		defaultVal := field.DefaultValue()
		imgui.BeginDisabledV(valV.Equal(reflect.ValueOf(defaultVal)))
		if imgui.Button(fnt.I("Undo")) {
			valV.Set(reflect.ValueOf(defaultVal))
			changed = true
		}
		imgui.EndDisabled()
		if imgui.BeginItemTooltip() {
			imutils.Textf("Reset to default (%v)", defaultVal)
			imgui.EndTooltip()
		}
		imgui.SameLine()

		label := strings.TrimPrefix(field.Name, category+".")
		width := max(0, imgui.ContentRegionAvail().X-imgui.CalcTextSize(label).X-imgui.CurrentStyle().FramePadding().X)
		imgui.PushItemWidth(width)
		switch field.Type.Kind() {
		case reflect.String:
			val := valV.String()
			if slices.Contains(field.Tags, "directory") {
				if imutils.FilePicker(label, &val, true) {
					valV.SetString(val)
					changed = true
				}
			} else if len(field.Options) > 0 {
				if imutils.ComboChoice(label, &val, field.Options) {
					valV.SetString(val)
					changed = true
				}
			} else {
				if imgui.InputTextWithHint(label, "", &val, 0, nil) {
					valV.SetString(val)
					changed = true
				}
			}
			imgui.SetItemTooltip(tooltip)
		case reflect.Bool:
			val := valV.Bool()
			if imgui.Checkbox(label, &val) {
				valV.SetBool(val)
				changed = true
			}
			imgui.SetItemTooltip(tooltip)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			min := int32(reflect.ValueOf(field.RangeMin).Int())
			max := int32(reflect.ValueOf(field.RangeMax).Int())
			val := int32(valV.Int())
			if imgui.SliderIntV(label, &val, min, max, "%d", imgui.SliderFlagsAlwaysClamp) {
				valV.SetInt(int64(val))
				changed = true
			}
			imgui.SetItemTooltip(tooltip)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			min := int32(reflect.ValueOf(field.RangeMin).Uint())
			max := int32(reflect.ValueOf(field.RangeMax).Uint())
			val := int32(valV.Uint())
			if imgui.SliderIntV(label, &val, min, max, "%d", imgui.SliderFlagsAlwaysClamp) {
				valV.SetUint(uint64(val))
				changed = true
			}
			imgui.SetItemTooltip(tooltip)
		case reflect.Float32, reflect.Float64:
			min := float32(reflect.ValueOf(field.RangeMin).Float())
			max := float32(reflect.ValueOf(field.RangeMax).Float())
			val := float32(valV.Uint())
			if imgui.SliderFloatV(label, &val, min, max, "%d", imgui.SliderFlagsAlwaysClamp) {
				valV.SetFloat(float64(val))
				changed = true
			}
			imgui.SetItemTooltip(tooltip)
		}
		imgui.PopItemWidth()

		imgui.EndDisabled()
		imgui.PopID()
	}

	return

	/*reset := func() {
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
				defaultVal := opt.Enum[0] == "true"
				if v := (*config)[convName][optName]; v != "" {
					val = v == "true"
				} else {
					val = defaultVal
				}
				if imgui.Checkbox(strID, &val) {
					if val == defaultVal {
						delete((*config)[convName], optName)
					} else {
						if val {
							(*config)[convName][optName] = "true"
						} else {
							(*config)[convName][optName] = "false"
						}
					}
					changed = true
				}
			} else {
				selectedVal := (*config)[convName][optName]
				defaultVal := opt.Enum[0]
				if selectedVal == "" {
					selectedVal = defaultVal
				}

				if imutils.ComboChoice(strID, &selectedVal, opt.Enum) {
					if selectedVal == defaultVal {
						delete((*config)[convName], optName)
					} else {
						(*config)[convName][optName] = selectedVal
					}
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
					delete((*config)[convName], optName)
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
	}*/
}
