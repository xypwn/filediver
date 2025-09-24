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
	"github.com/xypwn/filediver/stingray/unit/material/d3d"
)

var MissingGPUData error = errors.New("no gpu data")

func dumpMaterialStrings(fileId stingray.FileID, a *app.App) error {
	files := a.DataDir.Files[fileId]
	if !(len(files) > 0 && files[0].Exists(stingray.DataGPU)) {
		return MissingGPUData
	}
	gpuData, err := a.DataDir.Read(fileId, stingray.DataGPU)
	if err != nil {
		return err
	}

	idx := bytes.Index(gpuData, []byte("DXBC"))
	for idx != -1 {
		gpuData = gpuData[idx:]
		r := bytes.NewReader(gpuData)
		dxbc, err := d3d.ParseDXBC(r)
		if err != nil {
			return err
		}

		for _, rbind := range dxbc.ResourceDefinitions.ResourceBindings {
			fmt.Println(rbind.Name)
		}
		for _, cbuf := range dxbc.ResourceDefinitions.ConstantBuffers {
			for _, variable := range cbuf.Variables {
				fmt.Println(variable.Name)
			}
		}

		if len(gpuData) < 5 {
			break
		}
		gpuData = gpuData[4:]
		idx = bytes.Index(gpuData, []byte("DXBC"))
	}
	return err
}

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
		if err := dumpMaterialStrings(fileId, a); err != nil && !errors.Is(err, MissingGPUData) {
			if errors.Is(err, context.Canceled) {
				prt.NoStatus()
				prt.Warnf("Name dump canceled, exiting cleanly")
				return
			} else {
				prt.Errorf("%v", err)
			}
		}
	}
}
