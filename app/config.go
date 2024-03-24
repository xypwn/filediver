package app

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xypwn/filediver/extractor"
)

type ConfigTemplateOption struct {
	PossibleValues []string
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

func parseExtractorOptions(optsStr string) (extractor.Config, error) {
	res := make(extractor.Config)
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

func parseExtractorConfig(cfgStr string) (map[string]extractor.Config, error) {
	res := make(map[string]extractor.Config)
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

func validateExtractorOptions(template ConfigTemplateExtractor, opts extractor.Config) error {
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
		if !slices.Contains(opt.PossibleValues, v) {
			var availVs []string
			for _, v := range opt.PossibleValues {
				availVs = append(availVs, "\""+v+"\"")
			}
			return fmt.Errorf("%vunrecognized value: \"%v\", available values: %v", errPfx, v, strings.Join(availVs, ", "))
		}
	}
	return nil
}

func validateExtractorConfig(template ConfigTemplate, cfg map[string]extractor.Config, shorthands map[string][]string) error {
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
			var possibleVs []string
			for _, v := range opt.PossibleValues {
				possibleVs = append(possibleVs, "\""+v+"\"")
			}
			if len(possibleVs) > 0 {
				fmt.Fprintf(&res, " (possible values: %v)", strings.Join(possibleVs, ", "))
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

func ParseExtractorConfig(template ConfigTemplate, cfgStr string) (map[string]extractor.Config, error) {
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
