package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/hellflame/argparse"
	// "github.com/davecgh/go-spew/spew"

	"github.com/xypwn/filediver/converter"
	_ "github.com/xypwn/filediver/converter/bik"
	_ "github.com/xypwn/filediver/converter/texture"
	_ "github.com/xypwn/filediver/converter/wwise"
	"github.com/xypwn/filediver/stingray"
)

//go:embed files.txt
var knownFilesStr string

//go:embed types.txt
var knownTypesStr string

func convertStingrayFile(outDirPath string, file *stingray.File, name, typ string) error {
	convert, usedDataTypes, err := converter.Converter(typ)
	if err != nil {
		return err
	}
	var readers [3]io.ReadSeeker

	foundDataTypes := 0
	for dataType := stingray.DataType(0); dataType < stingray.NumDataType; dataType++ {
		if usedDataTypes&(1<<dataType) == 0 {
			continue
		}
		if !file.Exists(dataType) {
			continue
		}
		r, err := file.Open(dataType)
		if err != nil {
			return err
		}
		defer r.Close()
		readers[dataType] = r
		foundDataTypes++
	}
	if foundDataTypes == 0 {
		return nil
	}
	outPath := filepath.Join(outDirPath, name)
	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	if err := convert(outPath, readers); err != nil {
		return fmt.Errorf("convert %v (type %v): %w", name, typ, err)
	}

	return nil
}

func main() {
	prt := newPrinter()

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
		DisableDefaultShowHelp: true,
	})
	gameDir := parser.String("g", "gamedir", &argparse.Option{Help: "Helldivers 2 game directory"})
	outDir := parser.String("o", "out", &argparse.Option{Default: "extracted", Help: "Output directory (default: extracted)"})
	//verbose := parser.Flag("v", "verbose", &argparse.Option{Help: "Provide more detailed status output"})
	knownFilesPath := parser.String("", "files_file", &argparse.Option{Help: "Path to a text file containing known file names"})
	knownTypesPath := parser.String("", "types_file", &argparse.Option{Help: "Path to a text file containing known type names"})
	if err := parser.Parse(nil); err != nil {
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		}
		prt.Fatalf("%v", err)
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		if _, err := exec.LookPath("./ffmpeg"); err != nil {
			prt.Warnf("FFmpeg not installed or found locally. Please install FFmpeg, or place ffmpeg.exe in the current folder to convert videos to MP4. Without FFmpeg, videos will be saved as BIK.")
		}
	}

	if *gameDir == "" {
		hd2SteamPath, err := getSteamPath("553850", "Helldivers 2")
		if err == nil {
			prt.Infof("Using game found at: \"%v\"", hd2SteamPath)
			*gameDir = hd2SteamPath
		} else {
			if *gameDir == "" {
				prt.Errorf("Helldivers 2 Steam installation path not found: %v", err)
				prt.Fatalf("Unable to detect game install directory. Please specify the game directory manually using the '-g' option.")
			}
		}
	} else {
		prt.Infof("Game directory: \"%v\"", *gameDir)
	}

	prt.Infof("Output directory: \"%v\"", *outDir)

	if *knownFilesPath != "" {
		b, err := os.ReadFile(*knownFilesPath)
		if err != nil {
			prt.Fatalf("%v", err)
		}
		knownFilesStr = string(b)
	}
	if *knownTypesPath != "" {
		b, err := os.ReadFile(*knownTypesPath)
		if err != nil {
			prt.Fatalf("%v", err)
		}
		knownTypesStr = string(b)
	}

	createHashLookup := func(src string) map[stingray.Hash]string {
		res := make(map[stingray.Hash]string)
		sc := bufio.NewScanner(strings.NewReader(src))
		for sc.Scan() {
			s := strings.TrimSpace(sc.Text())
			if s != "" && !strings.HasPrefix(s, "//") {
				res[stingray.Sum64([]byte(s))] = s
			}
		}
		return res
	}

	knownFiles := createHashLookup(knownFilesStr)
	knownTypes := createHashLookup(knownTypesStr)

	prt.Infof("Reading metadata...")
	dataDir, err := stingray.OpenDataDir(filepath.Join(*gameDir, "data"))
	if err != nil {
		prt.Fatalf("%v", err)
	}
	prt.Infof("Extracting files...")
	numFile := 0
	numExtrFiles := 0
	for id, file := range dataDir.Files {
		name, ok := knownFiles[id.Name]
		if !ok {
			name = id.Name.String()
		}
		typ, ok := knownTypes[id.Type]
		if !ok {
			typ = id.Type.String()
		}
		truncName := name
		if len(truncName) > 40 {
			truncName = "..." + truncName[len(truncName)-37:]
		}
		prt.Statusf("File %v/%v: %v (%v)", numFile+1, len(dataDir.Files), truncName, typ)
		if err := convertStingrayFile(*outDir, file, name, typ); err != nil && err != converter.ErrFileType {
			prt.Errorf("%v", err)
		} else {
			numExtrFiles++
		}
		numFile++
	}
	prt.NoStatus()
	prt.Infof("Extracted %v/%v files", numExtrFiles, len(dataDir.Files))
}
