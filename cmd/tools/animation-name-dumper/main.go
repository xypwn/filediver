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
	"github.com/xypwn/filediver/stingray/state_machine"
)

func dumpAnimationNames(a *app.App, fileID stingray.FileID) error {
	bs, err := a.DataDir.Read(fileID, stingray.DataMain)
	if err != nil {
		return err
	}
	stateMachine, err := state_machine.LoadStateMachine(bytes.NewReader(bs))
	if err != nil {
		return err
	}
	for _, group := range stateMachine.Groups {
		for _, animation := range group.Animations {
			knownName, ok := a.Hashes[animation.Name]
			if ok {
				fmt.Println(knownName)
			} else {
				fmt.Println(animation.Name.String())
			}
		}
	}
	return nil
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
			prt.Warnf("Animation name dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	files, err := a.MatchingFiles("", "", []string{"state_machine"}, nil)
	if err != nil {
		prt.Fatalf("%v", err)
	}

	for fileID := range files {
		var cfg appconfig.Config
		config.InitDefault(&cfg)
		if err := dumpAnimationNames(a, fileID); err != nil {
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
