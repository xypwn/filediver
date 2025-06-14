package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/hellflame/argparse"
	"github.com/iancoleman/strcase"
	"github.com/jwalton/go-supportscolor"

	"github.com/xypwn/filediver/config"
)

func cliHandleArgs(configStruct any, addExtraArgs func(argp *argparse.Parser)) (dontExit bool, err error) {
	supportsColor := supportscolor.Stdout().SupportsColor
	args := os.Args[1:]

	if slices.Contains(args, "-c") || slices.Contains(args, "--config") {
		fmt.Println(`-c option is deprecated; see https://github.com/xypwn/filediver/wiki/10-CLI-Basics`)
		return false, nil
	}

	showHelp := false
	showAdvancedHelp := false
	if slices.Contains(args, "-h") || slices.Contains(args, "--help") {
		showHelp = true
	}
	if slices.Contains(args, "--help-all") {
		showHelp = true
		showAdvancedHelp = true
	}

	argpCfg := &argparse.ParserConfig{
		DisableHelp:            true,
		DisableDefaultShowHelp: true,
		WithHint:               true,
	}
	if showHelp && !showAdvancedHelp {
		if argpCfg.EpiLog != "" {
			argpCfg.EpiLog += "\n"
		}
		argpCfg.EpiLog += "Use --help-all to show all options, including advanced options."
	}
	argp := argparse.NewParser("filediver", "Helldivers 2 game asset extractor. https://github.com/xypwn/filediver", argpCfg)
	argp.Flag("h", "help", &argparse.Option{Help: "show help page"})
	argp.Flag("", "help-all", &argparse.Option{Help: "show help including advanced options"})

	if addExtraArgs != nil {
		addExtraArgs(argp)
	}

	fs, err := config.Fields(configStruct)
	if err != nil {
		return false, err
	}

	formatFieldName := func(field string) string {
		return "--" + strcase.ToKebab(field)
	}

	category := ""
	categoryHelp := ""
	values := map[string]*string{}
	flags := map[string]*bool{}
	for _, field := range fs.Fields {
		isAdvanced := slices.Contains(field.Tags, "advanced")
		if field.IsCategory {
			category = strcase.ToDelimited(field.Name, ' ') + " options"
			var affectedTypes []string
			for _, tag := range field.Tags {
				if after, ok := strings.CutPrefix(tag, "t:"); ok {
					affectedTypes = append(affectedTypes, after)
				}
			}
			categoryHelp = field.Help
			if len(affectedTypes) != 0 {
				if categoryHelp != "" {
					categoryHelp += ", "
				}
				categoryHelp += "affects: " + strings.Join(affectedTypes, ", ")
			}
			continue
		}
		if isAdvanced && showHelp && !showAdvancedHelp {
			continue
		}

		opts := &argparse.Option{
			Help:  field.Help,
			Group: category,
		}
		if isAdvanced {
			opts.Group = "advanced " + opts.Group
		}
		if !isAdvanced && categoryHelp != "" {
			opts.Group = opts.Group + " (" + categoryHelp + ")"
		}
		opts.HintInfo = fs.FieldFormatHint(field.Name, formatFieldName)
		var short string
		if field.Short != 0 {
			short = string(field.Short)
		}
		if field.Type.Kind() == reflect.Bool {
			flags[field.Name] = argp.Flag(short, strcase.ToKebab(field.Name), opts)
		} else {
			switch field.Type.Kind() {
			case reflect.String:
				opts.Meta = "STRING"
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				opts.Meta = "WHOLE-NUMBER"
			case reflect.Float32, reflect.Float64:
				opts.Meta = "DECIMAL-NUMBER"
			}
			values[field.Name] = argp.String(short, strcase.ToKebab(field.Name), opts)
		}
	}

	if showHelp {
		color := argparse.NoColor
		if supportsColor {
			color = argparse.DefaultColor
		}
		fmt.Println(argp.FormatHelpWithColor(color))
		return false, nil
	}

	if err := argp.Parse(args); err != nil {
		return false, err
	}

	err = config.MarshalFunc(configStruct, func(name string) (string, bool) {
		b, ok := flags[name]
		if ok {
			return strconv.FormatBool(*b), true
		}
		s, ok := values[name]
		if ok && *s != "" {
			return *s, true
		}
		return "", false
	})
	if err != nil {
		if merr, ok := err.(*config.MarshalErr); ok {
			log.Fatalf("parameter to %v %v", formatFieldName(merr.Field), merr.Err)
		} else {
			log.Fatal(err)
		}
	}

	return true, nil
}
