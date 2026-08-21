package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"runtime/pprof"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/hellflame/argparse"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/tools/fdtools-common"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

var reAudioPath = regexp.MustCompile(`^content/audio(/[a-z]{2})?/[0-9]+$`)

type HashInfo struct {
	// All hashes present in the game
	// (the ones we're trying to crack and the
	// ones aready known).
	AllHashes map[stingray.Hash]struct{}
	// Hashes not yet known.
	UnknownHashes map[stingray.Hash]struct{}
	// Known hashes.
	KnownHashes map[string]struct{}
}

type Context struct {
	workerId      int
	args          []string
	a             *application
	progressStore *any
}

func (c Context) Try(text []byte) (cont bool) {
	return c.a.try(c.workerId, text)
}

func (c Context) ReportTries(newTries int, found []string, nextStr []byte) (cont bool) {
	return c.a.reportTries(c.workerId, newTries, found, nextStr)
}

func (c Context) Args() []string {
	return c.args
}

func (c Context) Info() HashInfo {
	return c.a.hashInfo
}

func (c Context) ProgressStore() (any, bool) {
	if c.progressStore == nil {
		return nil, false
	}
	return *c.progressStore, true
}

func (c Context) SetProgressStore(v any) {
	*c.progressStore = v
}

func (c Context) Msg(format string, args ...any) {
	c.a.msgFromWorker(c.workerId, format, args...)
}

// Cracker is a program that cracks hashes.
type Cracker struct {
	Name     string
	Desc     string
	ArgNames []string
	Fn       func(c Context) error
}

type worker struct {
	name     string
	ctxErr   error
	err      error
	triesCnt int
	mu       sync.Mutex
	prot     struct { // protected by mu
		active              bool
		numTriesSinceUpdate int
		hashesPerSecond     float64
		messageBuf          []string
		nextStr             []byte
	}
	_padding [128]byte // prevent false sharing (yes this much is a bit wasteful, but we don't have many workers so who cares)
}

type application struct {
	ctx      context.Context
	prt      app.Printer
	hashInfo HashInfo

	workers []worker

	mu sync.Mutex
	// Guarded by mu //
	newHashes []string // newly cracked hashes
	// End           //

	lastUpdateTime time.Time
}

func crack(ctx context.Context, crackers []Cracker, crackerArgs [][]string, prt app.Printer, app *app.App) (newHashes []string, err error) {
	allHashes := make(map[stingray.Hash]struct{})
	for k := range app.DataDir.Files {
		allHashes[k.Name] = struct{}{}
		allHashes[k.Type] = struct{}{}
	}
	knownHashes := make(map[string]struct{})
	unknownHashes := make(map[stingray.Hash]struct{})
	for h := range allHashes {
		if s, exists := app.Hashes[h]; exists {
			knownHashes[s] = struct{}{}
		} else {
			unknownHashes[h] = struct{}{}
		}
	}
	a := application{
		ctx: ctx,
		prt: prt,
		hashInfo: HashInfo{
			AllHashes:     allHashes,
			UnknownHashes: unknownHashes,
			KnownHashes:   knownHashes,
		},
	}
	a.updateCliStatus()

	var wg sync.WaitGroup
	a.workers = make([]worker, len(crackers))
	for i := range a.workers {
		a.workers[i].name = crackers[i].Name
		var ps any // TODO: Implement progress store!!!
		wg.Go(func() {
			a.msgFromWorker(i, "Starting")
			a.workers[i].mu.Lock()
			a.workers[i].prot.active = true
			a.workers[i].mu.Unlock()
			a.workers[i].err = crackers[i].Fn(Context{
				workerId:      i,
				args:          crackerArgs[i],
				a:             &a,
				progressStore: &ps,
			})
			a.workers[i].mu.Lock()
			a.workers[i].prot.active = false
			a.workers[i].mu.Unlock()
			a.msgFromWorker(i, "Finished")
		})
	}
	for range time.Tick(500 * time.Millisecond) {
		a.updateCliStatus()
		if dt := time.Since(a.lastUpdateTime).Seconds(); dt >= 5 {
			now := time.Now()
			for i := range a.workers {
				w := &a.workers[i]
				w.mu.Lock()
				w.prot.hashesPerSecond = float64(w.prot.numTriesSinceUpdate) / dt
				w.prot.numTriesSinceUpdate = 0
				for _, s := range w.prot.messageBuf {
					a.prt.Infof("%s", s)
				}
				w.prot.messageBuf = w.prot.messageBuf[:0]
				w.mu.Unlock()
				a.lastUpdateTime = now
			}
		}
		for i := range a.workers {
			w := &a.workers[i]
			w.mu.Lock()
			for _, s := range w.prot.messageBuf {
				a.prt.Infof("%s", s)
			}
			w.prot.messageBuf = w.prot.messageBuf[:0]
			w.mu.Unlock()
		}
		if a.getNumActiveWorkers() == 0 {
			a.prt.Infof("All workers done")
			break
		}
		if ctx.Err() != nil {
			a.prt.Infof("Shutting down workers")
			break
		}
	}
	wg.Wait()
	var firstErr error
	var numErrs int
	for i := range a.workers {
		w := &a.workers[i]
		if w.ctxErr != nil && !errors.Is(w.ctxErr, context.Canceled) {
			if firstErr == nil {
				firstErr = fmt.Errorf("worker %d (%s): context: %w", i, w.name, w.ctxErr)
			}
			numErrs++
		}
		if w.err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("worker %d (%s): %w", i, w.name, w.err)
			}
			numErrs++
		}
	}
	switch numErrs {
	case 0:
		return a.newHashes, nil
	case 1:
		return nil, firstErr
	default:
		return nil, fmt.Errorf("%w (and %v more errors)", firstErr, numErrs-1)
	}
}

func (a *application) getNumActiveWorkers() int {
	numActiveWorkers := 0
	for i := range a.workers {
		a.workers[i].mu.Lock()
		if a.workers[i].prot.active {
			numActiveWorkers++
		}
		a.workers[i].mu.Unlock()
	}
	return numActiveWorkers
}

func (a *application) updateCliStatus() {
	var msg strings.Builder
	a.mu.Lock()
	fmt.Fprintf(&msg, "Total: %d/%d hashes, %d/%d workers",
		len(a.newHashes), len(a.hashInfo.UnknownHashes),
		a.getNumActiveWorkers(), len(a.workers))
	a.mu.Unlock()
	for i := range a.workers {
		w := &a.workers[i]
		w.mu.Lock()
		fmt.Fprintf(&msg, "; %s/%d: %.2fMH/s (%q)", w.name, i, w.prot.hashesPerSecond/1e6, w.prot.nextStr)
		w.mu.Unlock()
	}
	a.prt.Statusf("%s", msg.String())
}

// Must be called from the thread of the worker with the given ID.
func (a *application) msgFromWorker(workerId int, format string, args ...any) {
	w := &a.workers[workerId]
	msg := fmt.Sprintf(format, args...)
	msg = fmt.Sprintf("<%s/%d> %s", w.name, workerId, msg)
	w.mu.Lock()
	w.prot.messageBuf = append(w.prot.messageBuf, msg)
	w.mu.Unlock()
}

// Must be called from the thread of the worker with the given ID.
//
// If cont is returned false, the worker must do its cleanup and return
// within a reasonable time.
func (a *application) reportTries(workerId int, newTries int, found []string, nextStr []byte) (cont bool) {
	w := &a.workers[workerId]

	a.mu.Lock()
	a.newHashes = append(a.newHashes, found...)
	a.mu.Unlock()
	w.mu.Lock()
	w.prot.nextStr = bytes.Clone(nextStr)
	w.mu.Unlock()
	for _, s := range found {
		a.msgFromWorker(workerId, "%s=%s", stingray.Sum(s), s)
	}

	w.triesCnt += newTries
	if w.triesCnt >= 65536 {
		w.mu.Lock()
		w.prot.numTriesSinceUpdate += w.triesCnt
		w.mu.Unlock()
		w.triesCnt = 0
		if err := a.ctx.Err(); err != nil {
			w.ctxErr = err
			return false
		}
	}
	return true
}

// Must be called from the thread of the worker with the given ID.
//
// If cont is returned false, the worker must do its cleanup and return
// within a reasonable time.
func (a *application) try(workerId int, s []byte) (cont bool) {
	var found []string
	if _, exists := a.hashInfo.UnknownHashes[stingray.Sum(s)]; exists {
		found = append(found, string(s))
	}
	return a.reportTries(workerId, 1, found, s)
}

func main() {
	crackerByName := make(map[string]Cracker)
	var crackerNames []string
	for _, c := range crackers {
		crackerByName[c.Name] = c
		crackerNames = append(crackerNames, c.Name)
	}
	slices.Sort(crackerNames)

	var epilog strings.Builder
	{
		epilog.WriteString("crackers:\n")
		tw := tabwriter.NewWriter(&epilog, 0, 1, 2, ' ', 0)
		for _, c := range crackers {
			fmt.Fprintf(tw, "  %s", c.Name)
			for _, argName := range c.ArgNames {
				fmt.Fprintf(tw, " <%s>", argName)
			}
			fmt.Fprintf(tw, "\t%s\n", c.Desc)
		}
		tw.Flush()
	}

	argp := argparse.NewParser("hd2-hash-cracker", "HD2 hash cracking tool", &argparse.ParserConfig{
		EpiLog: epilog.String(),
	})
	optWrite := argp.Flag("w", "write", &argparse.Option{
		Help: "write newly cracked hashes to hashes/cracked.txt (app must be run from project root)",
	})
	optArgs := argp.Strings("", "crackers-and-args", &argparse.Option{
		Help:       "cracker name followed by args; separate multiple crackers by \"AND\"",
		Positional: true,
	})
	optCpuProfile := argp.Flag("", "cpuprofile", &argparse.Option{
		Help:      "write CPU profile",
		HideEntry: true,
	})
	prt, a := fdtools.Init(argp)
	if *optCpuProfile {
		const filename = "cpu.prof"
		f, err := os.Create(filename)
		if err != nil {
			prt.Fatalf("creating CPU profile file: %w", err)
		}
		prt.Infof("Starting CPU profile")
		if err := pprof.StartCPUProfile(f); err != nil {
			prt.Fatalf("starting CPU profile: %w", err)
		}
		defer func() {
			pprof.StopCPUProfile()
			prt.Infof("CPU profile written to " + filename)
		}()
	}
	var selCrackers []Cracker
	var selCrackerArgs [][]string
	{
		nextIsCracker := true
		for _, arg := range *optArgs {
			if arg == "AND" {
				if nextIsCracker {
					prt.Fatalf("unexpected AND")
				}
				nextIsCracker = true
			} else if nextIsCracker {
				c, ok := crackerByName[arg]
				if !ok {
					prt.Fatalf("Unknown cracker %q (options: %s [see --help for list of crackers])", arg, strings.Join(crackerNames, ", "))
				}
				selCrackers = append(selCrackers, c)
				selCrackerArgs = append(selCrackerArgs, nil)
				nextIsCracker = false
			} else {
				selCrackerArgs[len(selCrackerArgs)-1] = append(selCrackerArgs[len(selCrackerArgs)-1],
					arg)
			}
		}
		for i := range selCrackers {
			if len(selCrackerArgs[i]) != len(selCrackers[i].ArgNames) {
				prt.Fatalf("Cracker %q expects %d args, but got %d",
					selCrackers[i].Name, len(selCrackers[i].ArgNames),
					len(selCrackerArgs[i]))
			}
		}
	}
	prt.Infof("Ctrl+C to quit")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := func() error {
		newHashes, err := crack(ctx, selCrackers, selCrackerArgs, prt, a)
		if err != nil {
			return err
		}
		if *optWrite {
			crackedFile := "hashes/cracked.txt"
			if _, err := os.Stat(crackedFile); errors.Is(err, os.ErrNotExist) {
				crackedFile = "../../../" + crackedFile
			}
			if _, err := os.Stat(crackedFile); errors.Is(err, os.ErrNotExist) {
				prt.Errorf("Unable to write to cracked.txt because it could not be found")
				return nil
			}
			prt.Infof("Adding %d hashes to %s...", len(newHashes), crackedFile)
			fb, err := os.ReadFile(crackedFile)
			if err != nil {
				return err
			}
			b := bytes.NewBuffer(fb)
			for _, h := range newHashes {
				b.WriteString(h)
				b.WriteString("\n")
			}
			if err := os.WriteFile(crackedFile, hashes.TidyHashes(b.Bytes()), 0666); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		prt.Fatalf("%v\n", err)
	}
}
