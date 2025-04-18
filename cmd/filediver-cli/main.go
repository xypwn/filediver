package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"syscall"

	//"github.com/davecgh/go-spew/spew"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/qmuntal/gltf"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

func main() {
	prt := app.NewPrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	parser := argparse.NewParser("filediver", "An unofficial Helldivers 2 game asset extractor.", &argparse.ParserConfig{
		EpiLog: `matching files:
  Syntax is Glob (meaning * is supported)
  Basic format being matched is: file_path.file_type .
  file_path is the file path, or the file hash and
  file_type is the data type (see "extractors" section for a list of data types).
  examples:
    filediver -i "content/audio/*.wwise_stream"            extract all wwise_stream files in content/audio, or any subfolders
    filediver -i "{*.bik,*.wwise_stream,*.wwise_bank}"     extract all video and audio files (though easier with extractor config)
    filediver -i "content/audio/us/303183090.wwise_stream" extract one particular audio file

extractor config:
  basic format: filediver -c "key1:opt1=val1,opt2=val2 key2:opt1,opt2"
  examples:
    filediver -c "enable:all"                extract ALL files, including raw files (i.e. files that can't be converted)
    filediver -c "enable:audio"              only extract audio
    filediver -c "enable:bik bik:format=bik" only extract bik files, but don't convert them to mp4
    filediver -c "audio:format=wav"          convert audio to wav
` + app.ExtractorConfigHelpMessage(app.ConfigFormat),
		DisableDefaultShowHelp: true,
	})
	triads := parser.String("t", "triads", &argparse.Option{Help: "Include comma-separated triad name(s) as found in game data directory (aka Archive ID, eg 0x9ba626afa44a3aa3)"})
	gameDir := parser.String("g", "gamedir", &argparse.Option{Help: "Helldivers 2 game directory"})
	modeList := parser.Flag("l", "list", &argparse.Option{Help: "List all files without extracting anything. Format: known_name.known_type, name_hash.type_hash <- archives..."})
	outDir := parser.String("o", "out", &argparse.Option{Default: "extracted", Help: "Output directory (default: extracted)"})
	extrCfgStr := parser.String("c", "config", &argparse.Option{Help: "Configure extractors (see \"extractor config\" section)"})
	extrInclGlob := parser.String("i", "include", &argparse.Option{Help: "Select only matching files (glob syntax, see matching files section)"})
	extrExclGlob := parser.String("x", "exclude", &argparse.Option{Help: "Exclude matching files from selection (glob syntax, can be mixed with --include, see matching files section)"})
	armorStringsFile := parser.String("s", "strings", &argparse.Option{Default: "0x7c7587b563f10985", Help: "Strings file to use to map armor set string IDs to names (default: \"0x7c7587b563f10985\" - en-us)"})
	//verbose := parser.Flag("v", "verbose", &argparse.Option{Help: "Provide more detailed status output"})
	knownHashesPath := parser.String("", "hashes_file", &argparse.Option{Help: "Path to a text file containing known file and type names"})
	if err := parser.Parse(nil); err != nil {
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		}
		prt.Fatalf("%v", err)
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

	extrCfg, err := app.ParseExtractorConfig(app.ConfigFormat, *extrCfgStr)
	if err != nil {
		prt.Fatalf("%v", err)
	}

	runner := exec.NewRunner()
	if ok := runner.Add("ffmpeg", "-y", "-hide_banner", "-loglevel", "error"); !ok {
		prt.Warnf("FFmpeg not installed or found locally. Please install FFmpeg, or place ffmpeg.exe in the current folder to convert videos to MP4 and audio to a variety of formats. Without FFmpeg, videos will be saved as BIK and audio will be saved was WAV.")
	}
	if ok := runner.Add("scripts_dist/hd2_accurate_blender_importer/hd2_accurate_blender_importer"); !ok {
		prt.Warnf("Blender importer not found. Exporting directly to .blend is not available. Please download the scripts_dist archive and place its contents into the same folder as filediver (see https://github.com/xypwn/filediver?tab=readme-ov-file#helper-scripts-scripts_dist). Without blender importer, models will be saved as GLB.")
	}
	defer runner.Close()

	triadIDs := make([]stingray.Hash, 0)
	if *triads != "" {
		split := strings.Split(*triads, ",")
		for _, triad := range split {
			trimmed := strings.TrimPrefix(triad, "0x")
			value, err := strconv.ParseUint(trimmed, 16, 64)
			if err != nil {
				prt.Fatalf("parsing triad name: %v", err)
			}
			triadIDs = append(triadIDs, stingray.Hash{Value: value})
		}
	}

	armorStringsValue, err := strconv.ParseUint(strings.TrimPrefix(*armorStringsFile, "0x"), 16, 64)
	if err != nil {
		armorStringsValue, err = strconv.ParseUint(strings.TrimPrefix(*armorStringsFile, "0x"), 10, 64)
		if err != nil {
			prt.Warnf("unable to parse armor strings hash, using default of en-us")
			armorStringsValue = 0x7c7587b563f10985
		}
	}
	armorStringsHash := stingray.Hash{Value: armorStringsValue}

	if *gameDir == "" {
		var err error
		*gameDir, err = app.DetectGameDir()
		if err == nil {
			prt.Infof("Using game found at: \"%v\"", *gameDir)
		} else {
			prt.Errorf("Helldivers 2 Steam installation path not found: %v", err)
			prt.Fatalf("Unable to detect game install directory. Please specify the game directory manually using the '-g' option.")
		}
	} else {
		prt.Infof("Game directory: \"%v\"", *gameDir)
	}

	var knownHashes []string
	knownHashes = append(knownHashes, app.ParseHashes(hashes.Hashes)...)
	if *knownHashesPath != "" {
		b, err := os.ReadFile(*knownHashesPath)
		if err != nil {
			prt.Fatalf("%v", err)
		}
		knownHashes = append(knownHashes, app.ParseHashes(string(b))...)
	}

	var knownThinHashes []string
	knownThinHashes = append(knownThinHashes, app.ParseHashes(hashes.ThinHashes)...)

	if !*modeList {
		prt.Infof("Output directory: \"%v\"", *outDir)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	a, err := app.OpenGameDir(ctx, *gameDir, knownHashes, knownThinHashes, triadIDs, armorStringsHash, func(curr, total int) {
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

	files, err := a.MatchingFiles(*extrInclGlob, *extrExclGlob, triadIDs, app.ConfigFormat, extrCfg)
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

	if *modeList {
		for _, id := range sortedFileIDs {
			triadIDs := a.DataDir.Files[id].TriadIDs()
			triadIDStrings := make([]string, len(triadIDs))
			for i := range triadIDs {
				triadIDStrings[i] = triadIDs[i].String()
			}
			fmt.Printf("%v.%v, %v.%v <- %v\n",
				a.Hashes[id.Name], a.Hashes[id.Type],
				id.Name.String(), id.Type.String(),
				strings.Join(triadIDStrings, ", "),
			)
		}
	} else {
		prt.Infof("Extracting files...")

		var documents map[string]*gltf.Document = make(map[string]*gltf.Document)
		for key := range extrCfg {
			if value, contains := extrCfg[key]["single_glb"]; !contains || value == "false" {
				continue
			}
			var closeGLB func(doc *gltf.Document) error
			name := "combined_" + key
			if triads != nil && len(*triads) > 0 {
				name = fmt.Sprintf("%s_%s", strings.ReplaceAll(*triads, ",", "_"), key)
			}
			documents[key], closeGLB = createCloseableGltfDocument(*outDir, name, extrCfg[key], runner)
			defer closeGLB(documents[key])
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
			if _, err := a.ExtractFile(ctx, id, *outDir, extrCfg, runner, document, prt); err == nil {
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

		prt.NoStatus()
		prt.Infof("Extracted %v/%v matching files", numExtrFiles, len(files))
	}
}

func createCloseableGltfDocument(outDir string, triad string, cfg map[string]string, runner *exec.Runner) (*gltf.Document, func(doc *gltf.Document) error) {
	document := gltf.NewDocument()
	document.Asset.Generator = "https://github.com/xypwn/filediver"
	document.Samplers = append(document.Samplers, &gltf.Sampler{
		MagFilter: gltf.MagLinear,
		MinFilter: gltf.MinLinear,
		WrapS:     gltf.WrapRepeat,
		WrapT:     gltf.WrapRepeat,
	})
	closeGLB := func(doc *gltf.Document) error {
		outPath := filepath.Join(outDir, triad)
		formatIsBlend := cfg["format"] == "blend" && runner.Has("hd2_accurate_blender_importer")
		if formatIsBlend {
			err := extractor.ExportBlend(doc, outPath, runner)
			if err != nil {
				return err
			}
		} else {
			err := exportGLB(doc, outPath)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return document, closeGLB
}

func exportGLB(doc *gltf.Document, outPath string) error {
	path := outPath + ".glb"
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := gltf.NewEncoder(out)
	if err := enc.Encode(doc); err != nil {
		return err
	}
	return nil
}
