package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material"
)

var MissingGPUData error = errors.New("no gpu data")

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Unable to detect game install directory.")
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray.ThinHash{}, func(curr int, total int) {
		prt.Statusf("Opening game directory %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			prt.NoStatus()
			prt.Warnf("Physics dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	files, err := a.MatchingFiles("", "", []string{"material"}, []stingray.Hash{}, "")
	if err != nil {
		prt.Fatalf("%v", err)
	}

	var cfg appconfig.Config
	config.InitDefault(&cfg)
	for fileId := range files {
		mainData, err := a.DataDir.Read(fileId, stingray.DataMain)
		if err != nil {
			prt.Warnf("failed to load %v: %v", a.LookupHash(fileId.Name), err)
			continue
		}
		material, err := material.LoadMain(bytes.NewReader(mainData))
		if err != nil {
			prt.Warnf("failed to parse %v: %v", a.LookupHash(fileId.Name), err)
		}
		detailData, ok := material.Textures[stingray.Sum("surface_data_array").Thin()]
		if !ok {
			continue
		}
		fmt.Println(a.LookupHash(detailData))
	}
}
