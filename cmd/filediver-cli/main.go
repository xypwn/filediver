package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"

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
	"github.com/xypwn/filediver/stingray/unit"
)

// Aliases available in -T option.
var typeAliases = []string{
	"audio_stream", "wwise_stream",
	"audio_bank", "wwise_bank",
	"video", "bik",
	"model", "unit,geometry_group",
	"text", "strings,package,bones",
	"image", "texture",
	"animation_set", "state_machine",
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
	var optInclTriads *string
	var optArmorStringsFile *string
	var optKnownHashesPath *string
	var optThinHashListMode *string
	// Config common to CLI and GUI
	cfg := appconfig.Config{}

	if dontExit, err := cliHandleArgs(&cfg, func(argp *argparse.Parser) {
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
		if len(typeAliases)%2 != 0 {
			panic("expected typeAliases to be in chunks of 2")
		}
		var typeAliasStrs []string
		for i := 0; i < len(typeAliases); i += 2 {
			typeAliasStrs = append(typeAliasStrs,
				typeAliases[i+0]+"->"+typeAliases[i+1],
			)
		}
		optInclOnlyTypes = argp.String("T", "types", &argparse.Option{
			Default: "all",
			Help:    "only include comma-separated type name(s) (aliases: " + strings.Join(typeAliasStrs, ", ") + ")",
		})
		optInclTriads = argp.String("t", "triads", &argparse.Option{
			Help: "include comma-separated triad name(s) as found in game data directory; aka Archive ID, e.g. 0x9ba626afa44a3aa3",
		})
		optArmorStringsFile = argp.String("s", "strings", &argparse.Option{
			Default: "0x7c7587b563f10985",
			Help:    `strings file to use to map armor set string IDs to names (default: "0x7c7587b563f10985" - en-us)`,
		})
		optKnownHashesPath = argp.String("", "hashes-file", &argparse.Option{
			Help: "path to a text file containing known file and type names, will use built-in hash list if none is given",
		})
		optThinHashListMode = argp.String("b", "list-thins", &argparse.Option{
			Default: "none",
			Choices: []any{"none", "unknown", "known", "bone", "material", "all"},
			Help:    "if not none, list [option] thin hashes referenced in included unit files, then exit"},
		)
	}); err != nil {
		log.Fatal(err)
	} else if !dontExit {
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
		types := strings.NewReplacer(typeAliases...).
			Replace(*optInclOnlyTypes)
		inclOnlyTypes = strings.Split(types, ",")
	}
	var inclTriadIDs []stingray.Hash
	if *optInclTriads != "" {
		split := strings.Split(*optInclTriads, ",")
		for _, triad := range split {
			hash, err := stingray.ParseHash(triad)
			if err != nil {
				prt.Fatalf("parsing triad name: %v", err)
			}
			inclTriadIDs = append(inclTriadIDs, hash)
		}
	}

	armorStringsHash, err := stingray.ParseHash(*optArmorStringsFile)
	if err != nil {
		hashVal, err := strconv.ParseUint(*optArmorStringsFile, 10, 64)
		if err != nil {
			prt.Warnf("unable to parse armor strings hash (%v), using default of en-us", err)
			hashVal = 0x7c7587b563f10985
		}
		armorStringsHash = stingray.Hash{Value: hashVal}
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

	a, err := app.OpenGameDir(ctx, gamedir, knownHashes, knownThinHashes, armorStringsHash, func(curr, total int) {
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

	files, err := a.MatchingFiles(*optInclGlob, *optExclGlob, inclOnlyTypes, inclTriadIDs)
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
			triadIDs := a.DataDir.Files[id].TriadIDs()
			triadIDStrings := make([]string, len(triadIDs))
			for i := range triadIDs {
				// Triad/Archive filenames are not 0x prefixed, nor are they left padded with zeroes
				triadIDStrings[i] = fmt.Sprintf("%x", triadIDs[i].Value)
			}
			fmt.Printf("%v.%v, %v.%v <- %v\n",
				a.Hashes[id.Name], a.Hashes[id.Type],
				id.Name.String(), id.Type.String(),
				strings.Join(triadIDStrings, ", "),
			)
		}
	} else if *optThinHashListMode != "none" || *optThinToFind != "" {
		knownBone := make(map[string]bool)
		knownMat := make(map[string]bool)
		unknownBone := make(map[string]bool)
		unknownMat := make(map[string]bool)
		unitCount := 0
		for _, id := range sortedFileIDs {
			if id.Type != stingray.Sum64([]byte("unit")) {
				continue
			}
			r, err := a.DataDir.Files[id].Open(ctx, stingray.DataMain)
			if err != nil {
				prt.Errorf("opening %v.unit's main file: %v", err)
				continue
			}
			defer r.Close()

			unitInfo, err := unit.LoadInfo(r)
			if err != nil {
				prt.Errorf("loading info from %v.unit: %v", id.Name.String(), err)
				continue
			}

			for _, bone := range unitInfo.Bones {
				if *optThinToFind != "" && stingray.Sum64([]byte(*optThinToFind)).Thin() == bone.NameHash {
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
				if *optThinToFind != "" && stingray.Sum64([]byte(*optThinToFind)).Thin() == mat {
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
			prt.Infof("Listed %v units with bone or material '%v' == 0x%08x", unitCount, *optThinToFind, stingray.Sum64([]byte(*optThinToFind)).Thin().Value)
		}
	} else {
		prt.Infof("Extracting files...")

		var documents map[string]*gltf.Document = make(map[string]*gltf.Document)
		var documentsToClose []func() error
		if cfg.Unit.SingleFile {
			for _, key := range []string{"unit", "geometry_group", "material"} {
				name := "combined_" + key
				if optInclTriads != nil && len(*optInclTriads) > 0 {
					name = fmt.Sprintf("%s_%s", strings.ReplaceAll(*optInclTriads, ",", "_"), key)
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
			if _, err := a.ExtractFile(ctx, id, *optOutDir, cfg, runner, document, inclTriadIDs, prt); err == nil {
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
