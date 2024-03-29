package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime/pprof"
	"sort"

	//"github.com/davecgh/go-spew/spew"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"

	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
)

//go:embed hashes.txt
var knownHashesStr string

func main() {
	prt := app.NewPrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	if false {
		f, err := os.Create("cpu.prof")
		if err != nil {
			prt.Fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			prt.Fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	parser := argparse.NewParser("filediver", "An unofficial Helldivers 2 game asset extractor.", &argparse.ParserConfig{
		EpiLog: `matching files:
  Syntax is Glob (meaning *, ** etc. are supported)
  Basic format being matched is: <file_path>.<file_type>
  where file_type is NOT the extension, but the DATA TYPE (e.g. wwise_stream, texture etc.).
  examples:
    "content/audio/**.wwise_stream"              extract all wwise_stream files in content/audio, or any subfolders
    "{**.bik,**.wwise_stream,**.wwise_bank}"     extract all video and audio files (though easier with extractor config)
    "content/audio/us/303183090.wwise_stream"    extract one particular audio file (NOTE that the extension here is NOT the final extracted extension, but rather the data type)

extractor config:
  basic format: filediver -c "<key1>:<opt1>=<val1>,<opt2>=<val2> <key2>:<opt1>,<opt2>"
  examples:
    filediver -c "enable:all"                   extract ALL files, including raw files (i.e. files that can't be converted)
    filediver -c "enable:audio"                 only extract audio
    filediver -c "enable:bik bik:conv=false"    only extract bik files, but don't convert them to mp4
    filediver -c "audio:format=ogg"             convert audio to ogg instead of wav
` + app.ExtractorConfigHelpMessage(app.ConfigFormat),
		DisableDefaultShowHelp: true,
	})
	gameDir := parser.String("g", "gamedir", &argparse.Option{Help: "Helldivers 2 game directory"})
	modeList := parser.Flag("l", "list", &argparse.Option{Help: "List all files without extracting anything"})
	outDir := parser.String("o", "out", &argparse.Option{Default: "extracted", Help: "Output directory (default: extracted)"})
	extrCfgStr := parser.String("c", "config", &argparse.Option{Help: "Configure extractors (see \"extractor config\" section)"})
	extrInclGlob := parser.String("i", "include", &argparse.Option{Help: "Select only matching files (glob syntax, see matching files section)"})
	extrExclGlob := parser.String("x", "exclude", &argparse.Option{Help: "Exclude matching files from selection (glob syntax, can be mixed with --include, see matching files section)"})
	//verbose := parser.Flag("v", "verbose", &argparse.Option{Help: "Provide more detailed status output"})
	knownHashesPath := parser.String("", "hashes_file", &argparse.Option{Help: "Path to a text file containing known file and type names"})
	if err := parser.Parse(nil); err != nil {
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		}
		prt.Fatalf("%v", err)
	}

	extrCfg, err := app.ParseExtractorConfig(app.ConfigFormat, *extrCfgStr)
	if err != nil {
		prt.Fatalf("%v", err)
	}

	runner := exec.NewRunner()
	if ok := runner.Add("ffmpeg", "-y", "-hide_banner", "-loglevel", "error"); !ok {
		prt.Warnf("FFmpeg not installed or found locally. Please install FFmpeg, or place ffmpeg.exe in the current folder to convert videos to MP4 and audio to a variety of formats. Without FFmpeg, videos will be saved as BIK and audio will be saved was WAV.")
	}
	if ok := runner.Add("magick"); !ok {
		prt.Warnf("ImageMagick not installed or found locally. Please install ImageMagick, or place magick.exe in the current folder to convert textures. Without magick, textures cannot be converted.")
	}

	a, err := app.New(prt)
	if err != nil {
		prt.Fatalf("%v", err)
	}

	if *gameDir == "" {
		if path, err := a.DetectGameDir(); err == nil {
			prt.Infof("Using game found at: \"%v\"", path)
		} else {
			prt.Errorf("Helldivers 2 Steam installation path not found: %v", err)
			prt.Fatalf("Unable to detect game install directory. Please specify the game directory manually using the '-g' option.")
		}
	} else {
		if err := a.SetGameDir(*gameDir); err != nil {
			prt.Fatalf("%v", err)
		}
		prt.Infof("Game directory: \"%v\"", *gameDir)
	}

	if *knownHashesPath == "" {
		a.AddHashesFromString(knownHashesStr)
	} else {
		if err := a.AddHashesFromFile(*knownHashesPath); err != nil {
			prt.Fatalf("%v", err)
		}
	}

	if !*modeList {
		prt.Infof("Output directory: \"%v\"", *outDir)
	}

	prt.Infof("Reading metadata...")
	if err := a.OpenGameDir(); err != nil {
		prt.Fatalf("%v", err)
	}
	files, err := a.MatchingFiles(*extrInclGlob, *extrExclGlob, app.ConfigFormat, extrCfg)
	if err != nil {
		prt.Fatalf("%v", err)
	}

	getFileName := func(id stingray.FileID) string {
		name, ok := a.Hashes[id.Name]
		if !ok {
			name = id.Name.String()
		}
		typ, ok := a.Hashes[id.Type]
		if !ok {
			typ = id.Type.String()
		}
		return name + "." + typ
	}

	var sortedFileIDs []stingray.FileID
	for id := range files {
		sortedFileIDs = append(sortedFileIDs, id)
	}
	sort.Slice(sortedFileIDs, func(i, j int) bool {
		return getFileName(sortedFileIDs[i]) < getFileName(sortedFileIDs[j])
	})

	if *modeList {
		for _, id := range sortedFileIDs {
			fmt.Println(getFileName(id))
		}
	} else {
		prt.Infof("Extracting files...")

		numExtrFiles := 0
		for i, id := range sortedFileIDs {
			truncName := getFileName(id)
			if len(truncName) > 40 {
				truncName = "..." + truncName[len(truncName)-37:]
			}
			prt.Statusf("File %v/%v: %v", i+1, len(files), truncName)
			if err := a.ExtractFile(id, *outDir, extrCfg, runner); err == nil {
				numExtrFiles++
			} else {
				prt.Errorf("%v", err)
			}
		}

		prt.NoStatus()
		prt.Infof("Extracted %v/%v matching files", numExtrFiles, len(files))
	}
}
