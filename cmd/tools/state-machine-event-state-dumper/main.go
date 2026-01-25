package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/state_machine"
)

type eventAnimationInfo struct {
	Name string `json:"name"`
	Hash string `json:"hex_id"`
	ID   uint64 `json:"id"`
}

type eventAnimations struct {
	Event       string               `json:"event"`
	Animations  []eventAnimationInfo `json:"animations"`
	AddedHashes []stingray.Hash      `json:"-"`
}

func dumpStateMachineStates(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	stateMachine, err := state_machine.LoadStateMachine(r)
	if err != nil {
		return err
	}

	layers := make(map[string][]eventAnimations, 0)
	for idx, layer := range stateMachine.Layers {
		layerAnimations := make([]eventAnimations, 0)
		eventOrder := make([]string, 0)
		eventAnimationMap := make(map[string]eventAnimations)
		for _, state := range layer.States {
			for _, event := range stateMachine.AnimationEventHashes {
				transition, contains := state.StateTransitions[event]
				if !contains {
					continue
				}
				transitionState := layer.States[transition.Index]
				resolvedEvent := ctx.LookupThinHash(event)
				eventAnimation, contains := eventAnimationMap[resolvedEvent]
				if !contains {
					eventAnimation = eventAnimations{
						Event:       resolvedEvent,
						Animations:  make([]eventAnimationInfo, 0),
						AddedHashes: make([]stingray.Hash, 0),
					}
					eventOrder = append(eventOrder, resolvedEvent)
				}
				for _, animation := range transitionState.AnimationHashes {
					if slices.Contains(eventAnimation.AddedHashes, animation) {
						continue
					}
					eventAnimation.AddedHashes = append(eventAnimation.AddedHashes, animation)
					eventAnimation.Animations = append(eventAnimation.Animations, eventAnimationInfo{
						Name: ctx.LookupHash(animation),
						ID:   animation.Value,
						Hash: animation.String(),
					})
				}
				eventAnimationMap[resolvedEvent] = eventAnimation
			}
		}
		for _, event := range eventOrder {
			layerAnimations = append(layerAnimations, eventAnimationMap[event])
		}
		layers[fmt.Sprintf("layer_%v", idx)] = layerAnimations
	}

	f, err := ctx.CreateFile(".animation_events.json")
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write([]byte("{\n")); err != nil {
		return err
	}

	for i := 0; i < len(layers); i++ {
		layerName := fmt.Sprintf("layer_%v", i)
		layer, contains := layers[layerName]
		if !contains {
			continue
		}
		result, err := json.MarshalIndent(layer, "    ", "    ")
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(fmt.Sprintf("    \"%v\": %v", layerName, string(result))))
		if err != nil {
			return err
		}
		if i < len(layers)-1 {
			_, err = f.Write([]byte(","))
			if err != nil {
				return err
			}
		}
		_, err = f.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	_, err = f.Write([]byte("}\n"))
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

	argp := argparse.NewParser("state-machine-state-dumper", "", nil)
	outputDirectory := argp.String("", "output", &argparse.Option{
		Positional: true,
	})
	filenameGlob := argp.String("", "filename_glob", &argparse.Option{
		Positional: true,
	})
	if err := argp.Parse(nil); err != nil {
		prt.Fatalf("argparse: %v", err)
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
			prt.Warnf("Animation name dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	files, err := a.MatchingFiles(*filenameGlob, "", []string{"state_machine"}, nil, "")
	if err != nil {
		prt.Fatalf("%v", err)
	}

	for fileID := range files {
		name, ok := a.Hashes[fileID.Name]
		if !ok {
			name = fileID.Name.String()
		}
		name = filepath.Base(name)
		var cfg appconfig.Config
		config.InitDefault(&cfg)
		extrCtx, _ := extractor.NewContext(
			ctx,
			fileID,
			a.Hashes,
			a.ThinHashes,
			a.ArmorSets,
			a.SkinOverrideGroups,
			a.WeaponPaintSchemes,
			a.GameBuildInfo,
			a.LanguageMap,
			a.DataDir,
			nil,
			cfg,
			filepath.Join(*outputDirectory, name),
			[]stingray.Hash{},
			prt.Warnf,
		)
		if err := dumpStateMachineStates(extrCtx); err != nil {
			if errors.Is(err, context.Canceled) {
				prt.NoStatus()
				prt.Warnf("State dump canceled, exiting cleanly")
				return
			} else {
				prt.Errorf("%v", err)
			}
		}
	}
}
