package widgets

import (
	"fmt"
	"math"
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

func ConfigEditor(cfg *appconfig.Config, showAdvanced *bool, queryBuf *string) (changed bool) {
	imgui.PushIDStr("##config editor")
	defer imgui.PopID()

	imgui.SetNextItemWidth(-math.SmallestNonzeroFloat32)
	imgui.InputTextWithHint("##search", fnt.I("Search")+" Filter options...", queryBuf, 0, nil)
	imgui.Checkbox("Show advanced options", showAdvanced)

	defer imgui.EndChild()
	if !imgui.BeginChildStr("##options") {
		return false
	}

	dependsSatisfied, err := config.DependsSatisfied(cfg)
	if err != nil {
		imutils.TextError(err)
		return
	}

	configV := reflect.ValueOf(cfg).Elem()
	category := ""

	isFieldShown := func(field *config.Field) bool {
		if field.IsCategory {
			return false
		}
		isAdvanced := slices.Contains(field.Tags, "advanced")
		if isAdvanced && !*showAdvanced {
			return false
		}
		match := func(s string) bool {
			return strings.Contains(
				strings.ToLower(s),
				strings.ToLower(*queryBuf),
			)
		}
		matchesQuery := *queryBuf == "" ||
			match(field.Name) ||
			slices.ContainsFunc(field.Options, match)
		return matchesQuery
	}

	// Show a category if any of its children are shown
	shownCategories := map[string]bool{}
	for _, field := range appconfig.ConfigFields.Fields {
		category, _, ok := strings.Cut(field.Name, ".")
		if ok && isFieldShown(field) {
			shownCategories[category] = true
		}
	}

	for _, field := range appconfig.ConfigFields.Fields {
		if !isFieldShown(field) && !(field.IsCategory && shownCategories[field.Name]) {
			continue
		}

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
			pos := imgui.CursorScreenPos()
			pos.Y -= imgui.CurrentStyle().ItemSpacing().Y
			imgui.SetCursorScreenPos(pos)
			imgui.PushTextWrapPos()
			imutils.Textcf(imgui.NewVec4(0.8, 0.8, 0.8, 1), "%v", tooltip)
			imgui.Spacing()
			continue
		} else if category == "" {
			category = "General"
			imgui.SeparatorText(category)
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
}
