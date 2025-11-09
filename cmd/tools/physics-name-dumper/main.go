package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/physics"
)

func dumpPhysicsNames(a *app.App, fileID stingray.FileID, withHashes bool) error {
	bs, err := a.DataDir.Read(fileID, stingray.DataMain)
	if err != nil {
		return err
	}

	physics, err := physics.LoadPhysics(bytes.NewReader(bs))
	if err != nil {
		return err
	}
	physicsSuffix := string(physics.NameEnd[:23])
	knownName, ok := a.Hashes[fileID.Name]
	var toPrint string
	if ok {
		toPrint = knownName
	} else {
		toPrint = physicsSuffix
	}
	if withHashes {
		toPrint += ": " + fileID.Name.String()
	}
	fmt.Println(toPrint)
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

	parser := argparse.NewParser("physics-name-dumper", "", &argparse.ParserConfig{DisableDefaultShowHelp: true})
	withHashes := parser.Flag("i", "include-hashes", &argparse.Option{})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("argparser: %v", err)
	}

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

	files, err := a.MatchingFiles("", "", []string{"physics"}, nil, "")
	if err != nil {
		prt.Fatalf("%v", err)
	}

	for fileID := range files {
		var cfg appconfig.Config
		config.InitDefault(&cfg)
		if err := dumpPhysicsNames(a, fileID, *withHashes); err != nil {
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
