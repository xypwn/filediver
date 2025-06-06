package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	dlbin "github.com/xypwn/filediver/stingray/dl_bin"
	"github.com/xypwn/filediver/stingray/state_machine"
)

func dumpAnimationNames(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()
	stateMachine, err := state_machine.LoadStateMachine(r)
	if err != nil {
		return err
	}
	for _, group := range stateMachine.Groups {
		for _, animation := range group.Animations {
			knownName, ok := ctx.Hashes()[animation.Name]
			if ok {
				fmt.Println(knownName)
			} else {
				fmt.Println(animation.Name.String())
			}
		}
	}
	return nil
}

type stateMachineContext struct {
	ctx     context.Context
	file    *stingray.File
	app     *app.App
	printer app.Printer
}

func (c *stateMachineContext) OutPath() (string, error)  { return "", nil }
func (c *stateMachineContext) OutDir() (string, error)   { return "", nil }
func (c *stateMachineContext) AddFile()                  {}
func (c *stateMachineContext) File() *stingray.File      { return c.file }
func (c *stateMachineContext) Runner() *exec.Runner      { return nil }
func (c *stateMachineContext) Config() map[string]string { return nil }
func (c *stateMachineContext) GetResource(_, _ stingray.Hash) (*stingray.File, bool) {
	return nil, false
}
func (c *stateMachineContext) CreateFile(_ string) (io.WriteCloser, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *stateMachineContext) AllocateFile(_ string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
func (c *stateMachineContext) Ctx() context.Context                        { return c.ctx }
func (c *stateMachineContext) Files() []string                             { return nil }
func (c *stateMachineContext) Hashes() map[stingray.Hash]string            { return c.app.Hashes }
func (c *stateMachineContext) ThinHashes() map[stingray.ThinHash]string    { return c.app.ThinHashes }
func (c *stateMachineContext) TriadIDs() []stingray.Hash                   { return c.app.TriadIDs }
func (c *stateMachineContext) ArmorSets() map[stingray.Hash]dlbin.ArmorSet { return c.app.ArmorSets }
func (c *stateMachineContext) Warnf(f string, a ...any) {
	name, typ := c.app.LookupHash(c.file.ID().Name), c.app.LookupHash(c.file.ID().Type)
	c.printer.Warnf("dump %v.%v: %v", name, typ, fmt.Sprintf(f, a...))
}
func (c *stateMachineContext) LookupHash(hash stingray.Hash) string { return c.LookupHash(hash) }

var physicsConfigFormat = app.ConfigTemplate{
	Extractors: map[string]app.ConfigTemplateExtractor{
		"state_machine": {
			Category: "animations",
			Options: map[string]app.ConfigTemplateOption{
				"format": {
					Type: app.ConfigValueEnum,
					Enum: []string{"json", "source"},
				},
			},
		},
	},
	Fallback: "raw",
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

	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, make([]stingray.Hash, 0), stingray.Hash{Value: 0x0}, func(curr int, total int) {
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

	files, err := a.MatchingFiles("*.state_machine", "", a.TriadIDs, physicsConfigFormat, map[string]map[string]string{})
	if err != nil {
		prt.Fatalf("%v", err)
	}

	for _, file := range files {
		dumpCtx := &stateMachineContext{
			ctx:     ctx,
			file:    file,
			app:     a,
			printer: prt,
		}
		if err := dumpAnimationNames(dumpCtx); err != nil {
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
