package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/stingray"
)

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	genWwiseStreams := os.Getenv("FILEDIVER_GEN_WWISE_STREAMS")
	if genWwiseStreams == "0" || genWwiseStreams == "" {
		prt.Infof("Set env FILEDIVER_GEN_WWISE_STREAMS=1 to generate wwise stream file names. This requires a HD2 steam installation with ALL LANGUAGE PACKS INSTALLED and takes a few minutes of times.")
		os.Exit(0)
	}

	prt.Infof("Generating wwise_stream hashes...")

	gameDir, err := app.DetectGameDir()
	if err == nil {
		prt.Infof("Using game found at: \"%v\"", gameDir)
	} else {
		prt.Errorf("Helldivers 2 Steam installation path not found: %v\n", err)
		prt.Fatalf("Command line option for installation path not implemented in wwise_streams generator. Please open an issue on GitHub")
	}
	ctx := context.Background() // no need to exit cleanly since we're only reading
	a, err := app.OpenGameDir(ctx, gameDir, nil, nil, stingray.Hash{}, func(curr, total int) {
		prt.Statusf("Reading metadata %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		prt.Fatalf("Error opening game dir: %v", err)
	}
	prt.NoStatus()

	numWwiseStreams := 0
	fileNames := make(map[stingray.Hash]struct{})
	for id := range a.DataDir.Files {
		fileNames[id.Name] = struct{}{}
		if id.Type == stingray.Sum("wwise_stream") {
			numWwiseStreams++
		}
	}

	type audioPack struct {
		Prefix         []byte
		FriendlyName   string
		IsLanguagePack bool
	}
	audioPacks := []audioPack{
		{[]byte("content/audio/"), "Core", false},
		{[]byte("content/audio/us/"), "US English", true},
		{[]byte("content/audio/de/"), "German", true},
		{[]byte("content/audio/es/"), "Spanish (Spain)", true},
		{[]byte("content/audio/bp/"), "Brazilian Portuguese", true},
		{[]byte("content/audio/fr/"), "French", true},
		{[]byte("content/audio/it/"), "Italian", true},
		{[]byte("content/audio/ms/"), "Spanish (Latin America)", true},
		{[]byte("content/audio/jp/"), "Japanese", true},
	}

	out, err := os.Create("wwise_streams.txt")
	if err != nil {
		prt.Fatalf("Error: %v", err)
	}
	defer out.Close()

	numFound := 0
	var buf []byte
	for _, pack := range audioPacks {
		buf = append(buf, pack.Prefix...)
		counter := 0
		numFoundInPack := 0
		for i := int64(0); i < 1<<30; i++ {
			buf = strconv.AppendInt(buf, i, 10)
			if _, ok := fileNames[stingray.Sum(buf)]; ok {
				fmt.Fprintln(out, string(buf))
				numFoundInPack++
				numFound++
			}
			if counter >= 1<<25 {
				prt.Statusf("%v - %v: %.0f%% - %v found", string(pack.Prefix), pack.FriendlyName, float64(i)/float64(1<<30)*100, numFoundInPack)
				counter = 0
			}
			counter++
			buf = buf[:len(pack.Prefix)]
		}
		buf = buf[:0]
		if numFoundInPack == 0 {
			prt.Warnf("Language pack \"%v\" not found. Please install it or names will be missing!", pack.FriendlyName)
		}
	}
	prt.NoStatus()

	prt.Infof("Done generating wwise_stream hashes. Names for %v/%v wwise streams found.", numFound, numWwiseStreams)
}
