package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"os/signal"
	"reflect"
	"runtime/pprof"
	"slices"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"

	//"github.com/davecgh/go-spew/spew"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/mattn/go-shellwords"
	"github.com/qmuntal/gltf"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor/single_glb_helper"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	stingray_strings "github.com/xypwn/filediver/stingray/strings"
	"github.com/xypwn/filediver/stingray/unit"
)

// Aliases available in -T option.
var typeAliases = map[string][]string{
	"audio_stream":  {"wwise_stream"},
	"audio_bank":    {"wwise_bank"},
	"video":         {"bik"},
	"model":         {"unit", "geometry_group"},
	"text":          {"strings", "package", "bones"},
	"image":         {"texture"},
	"animation_set": {"state_machine"},
}

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	// CLI options
	var optList *bool
	var optThinToFind *string
	var optOutDir *string
	var optInclGlob *string
	var optExclGlob *string
	var optInclOnlyTypes *string
	var optInclArchives *string
	var optStringsLanguage *string
	var optMetadataFilter *string
	var optKnownHashesPath *string
	var optThinHashListMode *string
	var optHelpMetadata *bool
	// Config common to CLI and GUI
	cfg := appconfig.Config{}

	if argp, dontExit, err := cliHandleArgs(&cfg, func(argp *argparse.Parser) {
		optList = argp.Flag("l", "list", &argparse.Option{
			Help: "list all files without extracting anything; format: known_name.known_type, name_hash.type_hash <- archives...",
		})
		optThinToFind = argp.String("f", "find-thin", &argparse.Option{
			Help: "search for given thinhash (bone or material) name and print the unit file(s) containing it, then exit",
		})
		optOutDir = argp.String("o", "out", &argparse.Option{
			Default: "extracted",
			Help:    "output directory",
		})
		optInclGlob = argp.String("i", "include", &argparse.Option{
			Help: "select only matching files (glob syntax)",
		})
		optExclGlob = argp.String("x", "exclude", &argparse.Option{
			Help: "exclude matching files from selection (glob syntax, can be mixed with --include)",
		})
		var typeAliasStrs []string
		for _, t := range slices.Sorted(maps.Keys(typeAliases)) {
			typeAliasStrs = append(typeAliasStrs,
				t+"->"+strings.Join(typeAliases[t], ","),
			)
		}
		optInclOnlyTypes = argp.String("T", "types", &argparse.Option{
			Default: "all",
			Help:    "only include comma-separated type name(s) (aliases: " + strings.Join(typeAliasStrs, ", ") + ")",
		})
		optInclArchives = argp.String("t", "triads", &argparse.Option{
			Help: "include comma-separated archive name(s) [formerly triads] as found in game data directory, e.g. 0x9ba626afa44a3aa3",
		})
		langs := make([]any, len(stingray_strings.LanguageFriendlyNames))
		for i := range langs {
			langs[i] = stingray_strings.LanguageFriendlyNames[i]
		}
		optStringsLanguage = argp.String("s", "strings-language", &argparse.Option{
			Default: "English (US)",
			Choices: langs,
			Help:    "Language to use when exporting names and descriptions",
		})
		optMetadataFilter = argp.String("m", "filter-metadata", &argparse.Option{
			Help: `metadata search filter (see --help-metadata)`,
		})
		optKnownHashesPath = argp.String("", "hashes-file", &argparse.Option{
			Help: "path to a text file containing known file and type names, will use built-in hash list if none is given",
		})
		optThinHashListMode = argp.String("b", "list-thins", &argparse.Option{
			Default: "none",
			Choices: []any{"none", "unknown", "known", "bone", "material", "all"},
			Help:    "if not none, list [option] thin hashes referenced in included unit files, then exit"},
		)
		optHelpMetadata = argp.Flag("", "help-metadata", &argparse.Option{
			Help: `show metadata filter syntax help`,
		})
	}); err != nil {
		log.Fatal(err)
	} else if !dontExit {
		os.Exit(0)
	} else if *optInclGlob == "" && *optInclArchives == "" && *optMetadataFilter == "" {
		cliShowHelp(argp)
		fmt.Println("\nExpected some specifier of which files to extract/list/search (--include, --triads or --filter-metadata).\nIf you wish to select all files, just pass -i \"*\".")
		os.Exit(1)
	}

	if *optHelpMetadata {
		fmt.Println(`Example:
  filediver -T texture --filter-metadata "width == 512 && height == 1024 && format == 'BC1UNorm'"
  (extracts all textures with width 512, height 1024 and BC1UNorm image format [mostly cape textures])

Syntax:
  - expr-lang (see https://expr-lang.org)
  - hashes must be passed as strings
  - value name casing is ignored
  - casing is ignored when checking if strings are equal

Options:`)
		typ := reflect.TypeFor[app.FileMetadata]()
		tabw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tabw, "  =Name=\t=Type=\t=Description=\n")
		for i := range typ.NumField() {
			field := typ.Field(i)
			if field.Tag.Get("meta") == "true" {
				continue
			}
			var exampleStr string
			if example, ok := field.Tag.Lookup("example"); ok {
				exampleStr = fmt.Sprintf(" (e.g. %s)", example)
			}
			fmt.Fprintf(tabw, "  %s\t%s\t%s%s\n", field.Name, app.FileMetadataTypeName(field.Name), field.Tag.Get("help"), exampleStr)
		}
		tabw.Flush()
		os.Exit(0)
	}

	if value, ok := os.LookupEnv("FILEDIVER_CPU_PROFILE"); ok && value != "" && value != "0" {
		f, err := os.Create(value)
		if err != nil {
			prt.Fatalf("%v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			prt.Fatalf("%v", err)
		}
		defer pprof.StopCPUProfile()
	}

	runner := exec.NewRunner()
	if ok := runner.Add("ffmpeg", "-y", "-hide_banner", "-loglevel", "error"); !ok {
		if cfg.Video.Format != "bik" && cfg.Video.Format != "raw" {
			cfg.Video.Format = "bik"
		}
		if cfg.Audio.Format != "wav" && cfg.Audio.Format != "raw" {
			cfg.Video.Format = "wav"
		}
		prt.Warnf("FFmpeg not installed or found locally. Please install FFmpeg, or place ffmpeg.exe in the current folder to convert videos to MP4 and audio to a variety of formats. Without FFmpeg, videos will be saved as BIK and audio will be saved was WAV.")
	}
	blenderImporterCommand := []string{"scripts_dist/hd2_accurate_blender_importer/hd2_accurate_blender_importer"}
	if value := os.Getenv("FILEDIVER_BLENDER_IMPORTER_COMMAND"); value != "" {
		if args, err := shellwords.Parse(value); err == nil && len(args) >= 1 {
			prt.Infof("Using blender importer command: %v", args)
			blenderImporterCommand = args
		}
	}
	if ok := runner.AddWithName("hd2_accurate_blender_importer", blenderImporterCommand[0], blenderImporterCommand[1:]...); !ok {
		if cfg.Model.Format == "blend" {
			cfg.Model.Format = "glb"
		}
		if cfg.Material.Format == "blend" {
			cfg.Material.Format = "glb"
		}
		prt.Warnf("Blender importer not found. Exporting directly to .blend is not available. Please download the scripts_dist archive and place its contents into the same folder as filediver (see https://github.com/xypwn/filediver?tab=readme-ov-file#helper-scripts-scripts_dist). Without blender importer, models will be saved as GLB.")
	}

	var inclOnlyTypes []string
	if *optInclOnlyTypes != "all" {
		for typeName := range strings.SplitSeq(*optInclOnlyTypes, ",") {
			if replace, ok := typeAliases[typeName]; ok {
				inclOnlyTypes = append(inclOnlyTypes, replace...)
			} else {
				inclOnlyTypes = append(inclOnlyTypes, typeName)
			}
		}
	}
	var inclArchiveIDs []stingray.Hash
	if *optInclArchives != "" {
		split := strings.Split(*optInclArchives, ",")
		for _, archive := range split {
			hash, err := stingray.ParseHash(archive)
			if err != nil {
				prt.Fatalf("parsing archive name: %v", err)
			}
			inclArchiveIDs = append(inclArchiveIDs, hash)
		}
	}

	var gamedir string
	if cfg.Gamedir == "<auto-detect>" {
		var err error
		gamedir, err = app.DetectGameDir()
		if err == nil {
			prt.Infof("Using game found at: \"%v\"", gamedir)
		} else {
			prt.Errorf("Helldivers 2 Steam installation path not found: %v", err)
			prt.Fatalf("Unable to detect game install directory. Please specify the game directory manually using the '-g' option.")
		}
	} else {
		gamedir = cfg.Gamedir
		prt.Infof("Game directory: \"%v\"", gamedir)
	}

	var knownHashes []string
	knownHashes = append(knownHashes, app.ParseHashes(hashes.Hashes)...)
	if *optKnownHashesPath != "" {
		b, err := os.ReadFile(*optKnownHashesPath)
		if err != nil {
			prt.Fatalf("%v", err)
		}
		knownHashes = append(knownHashes, app.ParseHashes(string(b))...)
	}

	var knownThinHashes []string
	knownThinHashes = append(knownThinHashes, app.ParseHashes(hashes.ThinHashes)...)

	if !(*optList || *optThinHashListMode != "none" || *optThinToFind != "") {
		prt.Infof("Output directory: \"%v\"", *optOutDir)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	a, err := app.OpenGameDir(ctx, gamedir, knownHashes, knownThinHashes, stingray_strings.LanguageFriendlyNameToHash[*optStringsLanguage], func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			prt.NoStatus()
			prt.Warnf("Metadata read canceled, exiting")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	files, err := a.MatchingFiles(*optInclGlob, *optExclGlob, inclOnlyTypes, inclArchiveIDs, *optMetadataFilter)
	if err != nil {
		prt.Fatalf("%v", err)
	}

	getFileName := func(id stingray.FileID) string {
		return a.LookupHash(id.Name) + "." + a.LookupHash(id.Type)
	}

	var sortedFileIDs []stingray.FileID
	for id := range files {
		sortedFileIDs = append(sortedFileIDs, id)
	}
	sort.Slice(sortedFileIDs, func(i, j int) bool {
		return getFileName(sortedFileIDs[i]) < getFileName(sortedFileIDs[j])
	})

	{
		names := make(map[stingray.Hash]struct{})
		types := make(map[stingray.Hash]struct{})
		for id := range a.DataDir.Files {
			names[id.Name] = struct{}{}
			types[id.Type] = struct{}{}
		}
		numKnownNames := 0
		numKnownTypes := 0
		for k := range names {
			if _, ok := a.Hashes[k]; ok {
				numKnownNames++
			}
		}
		for k := range types {
			if _, ok := a.Hashes[k]; ok {
				numKnownTypes++
			}
		}
		prt.Infof(
			"Known hashes: names %.2f%%, types %.2f%%",
			float64(numKnownNames)/float64(len(names))*100,
			float64(numKnownTypes)/float64(len(types))*100,
		)
	}

	if *optList {
		for _, id := range sortedFileIDs {
			var archiveIDStrings []string
			for _, file := range a.DataDir.Files[id] {
				// Archive filenames are not 0x prefixed
				archiveIDStrings = append(archiveIDStrings,
					fmt.Sprintf("%016x", file.ArchiveID.Value))
			}
			fmt.Printf("%v.%v, %v.%v <- %v\n",
				a.Hashes[id.Name], a.Hashes[id.Type],
				id.Name.String(), id.Type.String(),
				strings.Join(archiveIDStrings, ", "),
			)
		}
	} else if *optThinHashListMode != "none" || *optThinToFind != "" {
		knownBone := make(map[string]bool)
		knownMat := make(map[string]bool)
		unknownBone := make(map[string]bool)
		unknownMat := make(map[string]bool)
		unitCount := 0
		for _, id := range sortedFileIDs {
			if id.Type != stingray.Sum("unit") {
				continue
			}
			b, err := a.DataDir.Read(id, stingray.DataMain)
			if err != nil {
				prt.Errorf("opening %v.unit's main file: %v", err)
				continue
			}

			unitInfo, err := unit.LoadInfo(bytes.NewReader(b))
			if err != nil {
				prt.Errorf("loading info from %v.unit: %v", id.Name.String(), err)
				continue
			}

			for _, bone := range unitInfo.Bones {
				if *optThinToFind != "" && stingray.Sum(*optThinToFind).Thin() == bone.NameHash {
					unitName, exists := a.Hashes[id.Name]
					if !exists {
						unitName = id.Name.String()
					}
					fmt.Printf("%v.unit\n", unitName)
					unitCount++
					break
				} else if *optThinToFind != "" {
					continue
				}

				if name, exists := a.ThinHashes[bone.NameHash]; exists {
					knownBone[name] = true
				} else {
					unknownBone[bone.NameHash.String()] = true
				}
			}
			for mat := range unitInfo.Materials {
				if *optThinToFind != "" && stingray.Sum(*optThinToFind).Thin() == mat {
					unitName, exists := a.Hashes[id.Name]
					if !exists {
						unitName = id.Name.String()
					}
					fmt.Printf("%v.unit\n", unitName)
					unitCount++
					break
				} else if *optThinToFind != "" {
					continue
				}

				if name, exists := a.ThinHashes[mat]; exists {
					knownMat[name] = true
				} else {
					unknownMat[mat.String()] = true
				}
			}
		}

		knownSorted := make([]string, len(knownBone)+len(knownMat))
		i := 0
		for name := range knownBone {
			knownSorted[i] = name
			i++
		}
		for name := range knownMat {
			knownSorted[i] = name
			i++
		}

		unknownSorted := make([]string, len(unknownBone)+len(unknownMat))
		i = 0
		for name := range unknownBone {
			unknownSorted[i] = name
			i++
		}
		for name := range unknownMat {
			unknownSorted[i] = name
			i++
		}

		stdoutStat, _ := os.Stdout.Stat()
		showRedirectHint := (stdoutStat.Mode() & os.ModeCharDevice) != 0
		var printed int
		switch *optThinHashListMode {
		case "known":
			slices.Sort(knownSorted)
			for _, bone := range knownSorted {
				fmt.Println(bone)
			}
			printed = len(knownSorted)
		case "unknown":
			slices.Sort(unknownSorted)
			for _, bone := range unknownSorted {
				fmt.Println(bone)
			}
			printed = len(unknownSorted)
		case "bone":
			slices.Sort(knownSorted[:len(knownBone)])
			slices.Sort(unknownSorted[:len(unknownBone)])
			for _, bone := range knownSorted[:len(knownBone)] {
				fmt.Println(bone)
			}
			for _, bone := range unknownSorted[:len(unknownBone)] {
				fmt.Println(bone)
			}
			printed = len(unknownBone) + len(knownBone)
		case "material":
			slices.Sort(knownSorted[len(knownBone):])
			slices.Sort(unknownSorted[len(unknownBone):])
			for _, mat := range knownSorted[len(knownBone):] {
				fmt.Println(mat)
			}
			for _, mat := range unknownSorted[len(unknownBone):] {
				fmt.Println(mat)
			}
			printed = len(unknownMat) + len(knownMat)
		case "all":
			slices.Sort(knownSorted)
			slices.Sort(unknownSorted)
			for _, bone := range knownSorted {
				fmt.Println(bone)
			}
			for _, bone := range unknownSorted {
				fmt.Println(bone)
			}
			printed = len(unknownSorted) + len(knownSorted)
		}

		if showRedirectHint && printed > 127 {
			prt.Infof("Listed %v bones or materials (you should probably redirect this to a file)", printed)
		}
		if *optThinToFind != "" {
			prt.Infof("Listed %v units with bone or material '%v' == 0x%08x", unitCount, *optThinToFind, stingray.Sum(*optThinToFind).Thin().Value)
		}
	} else {
		prt.Infof("Extracting files...")

		var documents map[string]*gltf.Document = make(map[string]*gltf.Document)
		var documentsToClose []func() error
		if cfg.Unit.SingleFile {
			for _, key := range []string{"unit", "geometry_group", "material"} {
				name := "combined_" + key
				if optInclArchives != nil && len(*optInclArchives) > 0 {
					name = fmt.Sprintf("%s_%s", strings.ReplaceAll(*optInclArchives, ",", "_"), key)
				}
				var formatBlend bool
				switch key {
				case "unit", "geometry_group":
					formatBlend = cfg.Model.Format == "blend"
				case "material":
					formatBlend = cfg.Material.Format == "blend"
				default:
					panic("unknown format: " + key)
				}
				doc, close := single_glb_helper.CreateCloseableGltfDocument(*optOutDir, name, formatBlend, runner)
				documents[key] = doc
				documentsToClose = append(documentsToClose, func() error { return close(doc) })
			}
		}

		numExtrFiles := 0
		for i, id := range sortedFileIDs {
			truncName := getFileName(id)
			if len(truncName) > 40 {
				truncName = "..." + truncName[len(truncName)-37:]
			}
			typ, ok := a.Hashes[id.Type]
			if !ok {
				typ = id.Type.String()
			}
			prt.Statusf("File %v/%v: %v", i+1, len(files), truncName)
			document := documents[typ]
			if _, err := a.ExtractFile(ctx, id, *optOutDir, cfg, runner, document, inclArchiveIDs, prt); err == nil {
				numExtrFiles++
			} else {
				if errors.Is(err, context.Canceled) {
					prt.NoStatus()
					prt.Warnf("Extraction canceled, exiting cleanly")
					return
				} else {
					prt.Errorf("%v", err)
				}
			}
		}

		for _, close := range documentsToClose {
			if err := close(); err != nil {
				prt.Errorf("%v", err)
			}
		}

		prt.NoStatus()
		prt.Infof("Extracted %v/%v matching files", numExtrFiles, len(files))
	}
}
