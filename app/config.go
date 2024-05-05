package app

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type ConfigValueType int

const (
	ConfigValueEnum ConfigValueType = iota
	ConfigValueIntRange
)

type ConfigTemplateOption struct {
	Type        ConfigValueType
	Enum        []string
	IntRangeMin int
	IntRangeMax int
}

type ConfigTemplateExtractor struct {
	Category        string
	Options         map[string]ConfigTemplateOption
	DefaultDisabled bool
}

type ConfigTemplate struct {
	Extractors map[string]ConfigTemplateExtractor
	Fallback   string
}

func parseExtractorOptions(optsStr string) (map[string]string, error) {
	res := make(map[string]string)
	opts := strings.Split(optsStr, ",")
	for _, opt := range opts {
		k, v, ok := strings.Cut(opt, "=")
		if !ok {
			v = "true"
		}

		res[k] = v
	}
	return res, nil
}

func parseExtractorConfig(cfgStr string) (map[string]map[string]string, error) {
	res := make(map[string]map[string]string)
	if cfgStr == "" {
		return res, nil
	}

	sp := strings.Split(cfgStr, " ")
	for _, s := range sp {
		name, optsStr, ok := strings.Cut(s, ":")
		if !ok {
			return nil, fmt.Errorf("expected \":\" to separate extractor name and options")
		}

		cfg, err := parseExtractorOptions(optsStr)
		if err != nil {
			return nil, err
		}

		res[name] = cfg
	}

	return res, nil
}

func validateExtractorOptions(template ConfigTemplateExtractor, opts map[string]string) error {
	for k, v := range opts {
		errPfx := fmt.Sprintf("\"%v=%v\": ", k, v)
		opt, ok := template.Options[k]
		if !ok {
			var availKs []string
			for k := range template.Options {
				availKs = append(availKs, "\""+k+"\"")
			}
			slices.Sort(availKs)
			return fmt.Errorf("%vunrecognized option: \"%v\", available options: %v", errPfx, k, strings.Join(availKs, ", "))
		}
		switch opt.Type {
		case ConfigValueEnum:
			if !slices.Contains(opt.Enum, v) {
				var availVs []string
				for _, v := range opt.Enum {
					availVs = append(availVs, "\""+v+"\"")
				}
				return fmt.Errorf("%vunrecognized value: \"%v\", available values: %v", errPfx, v, strings.Join(availVs, ", "))
			}
		case ConfigValueIntRange:
			vInt, err := strconv.Atoi(v)
			if err != nil || vInt < opt.IntRangeMin || vInt > opt.IntRangeMax {
				return fmt.Errorf("%vinvalid value: \"%v\", expected whole number from %v to %v", errPfx, v, opt.IntRangeMin, opt.IntRangeMax)
			}
		default:
			panic("invalid config ConfigValueType")
		}
	}
	return nil
}

func validateExtractorConfig(template ConfigTemplate, cfg map[string]map[string]string, shorthands map[string][]string) error {
	validExtractorName := func(name string) (isShorthand bool, ok bool) {
		if _, ok := template.Extractors[name]; ok {
			return false, true
		}
		if _, ok := shorthands[name]; ok {
			return true, true
		}
		return false, false
	}

	var availExtrsStr string
	{
		var availExtrs []string
		for k := range template.Extractors {
			availExtrs = append(availExtrs, "\""+k+"\"")
		}
		slices.Sort(availExtrs)
		availExtrsStr = fmt.Sprintf("available extractors: %v", strings.Join(availExtrs, ", "))
		var availShorthands []string
		for k := range shorthands {
			availShorthands = append(availShorthands, "\""+k+"\"")
		}
		slices.Sort(availShorthands)
		if len(availShorthands) > 0 {
			availExtrsStr += fmt.Sprintf(", available shorthands: %v", strings.Join(availShorthands, ", "))
		}
	}

	for name, opts := range cfg {
		if isShorthand, ok := validExtractorName(name); ok {
			if isShorthand {
				for _, name := range shorthands[name] {
					if err := validateExtractorOptions(template.Extractors[name], opts); err != nil {
						return fmt.Errorf("%v: %w", name, err)
					}
				}
			} else {
				if err := validateExtractorOptions(template.Extractors[name], opts); err != nil {
					return fmt.Errorf("%v: %w", name, err)
				}
			}
			continue
		}
		if name == "enable" || name == "disable" {
			for k, v := range opts {
				errPfx := fmt.Sprintf("\"%v:%v\": ", name, k)
				if _, ok := validExtractorName(k); !ok {
					return fmt.Errorf("%vunrecognized extractor name: \"%v\", %v", errPfx, k, availExtrsStr)
				}
				if v != "true" && v != "false" {
					return fmt.Errorf("%vexpected \"true\" or \"false\", but got: \"%v\"", errPfx, v)
				}
			}
			continue
		}
		return fmt.Errorf("unrecognized extractor name: \"%v\", %v", name, availExtrsStr)
	}
	return nil
}

func substituteShorthandKeys[T any](shorthands map[string][]string, cfg map[string]T) {
	for k, v := range cfg {
		if shs, ok := shorthands[k]; ok {
			for _, sh := range shs {
				cfg[sh] = v
			}
			continue
		}
	}
}

func getShorthandsMap(template ConfigTemplate) map[string][]string {
	res := make(map[string][]string)
	for extrName, extrTmpl := range template.Extractors {
		category := extrTmpl.Category
		if category != "" {
			res[category] = append(res[category], extrName)
		}
		res["all"] = append(res["all"], extrName)
	}
	return res
}

func ExtractorConfigHelpMessage(template ConfigTemplate) string {
	shorthands := getShorthandsMap(template)

	var res strings.Builder
	fmt.Fprintf(&res, "extractors:\n")

	var sortedExtrKeys []string
	for name := range template.Extractors {
		sortedExtrKeys = append(sortedExtrKeys, name)
	}
	slices.Sort(sortedExtrKeys)

	for _, name := range sortedExtrKeys {
		extr := template.Extractors[name]

		fmt.Fprintf(&res, "  \"%v\"", name)
		var notes []string
		if extr.DefaultDisabled {
			notes = append(notes, "disabled by default")
		}
		if name == template.Fallback {
			notes = append(notes, "fallback")
		}
		if len(notes) > 0 {
			fmt.Fprintf(&res, " (%v)", strings.Join(notes, ", "))
		}
		if len(extr.Options) > 0 {
			fmt.Fprintf(&res, ", options:")
		}
		fmt.Fprintf(&res, "\n")
		for name, opt := range extr.Options {
			fmt.Fprintf(&res, "    \"%v\"", name)
			switch opt.Type {
			case ConfigValueEnum:
				var enumVals []string
				for _, v := range opt.Enum {
					enumVals = append(enumVals, "\""+v+"\"")
				}
				if len(enumVals) > 0 {
					fmt.Fprintf(&res, " (any of: %v)", strings.Join(enumVals, ", "))
				}
			case ConfigValueIntRange:
				fmt.Fprintf(&res, " (whole number from %v to %v)", opt.IntRangeMin, opt.IntRangeMax)
			}
			fmt.Fprintf(&res, "\n")
		}
	}

	var sortedShorthandKeys []string
	for name := range shorthands {
		sortedShorthandKeys = append(sortedShorthandKeys, name)
	}
	slices.Sort(sortedShorthandKeys)

	fmt.Fprintf(&res, "extractor shorthands:\n")
	for _, name := range sortedShorthandKeys {
		extrs := shorthands[name]

		fmt.Fprintf(&res, "  \"%v\" (shorthand for: %v)\n", name, strings.Join(extrs, ", "))
	}

	return res.String()
}

func ParseExtractorConfig(template ConfigTemplate, cfgStr string) (map[string]map[string]string, error) {
	res, err := parseExtractorConfig(cfgStr)
	if err != nil {
		return nil, fmt.Errorf("extractor config: %w", err)
	}

	shorthands := getShorthandsMap(template)

	if err := validateExtractorConfig(template, res, shorthands); err != nil {
		return nil, fmt.Errorf("extractor config: %w", err)
	}

	substituteShorthandKeys(shorthands, res)
	substituteShorthandKeys(shorthands, res["enable"])
	substituteShorthandKeys(shorthands, res["disable"])

	return res, nil
}
