package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/hellflame/argparse"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/tools/fdtools-common"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

var reAudioPath = regexp.MustCompile(`^content/audio(/[a-z]{2})?/[0-9]+$`)

type hashInfo struct {
	hashes      map[stingray.Hash]struct{}
	knownHashes map[string]struct{}
}

// generator is a hash string generator.
type generator struct {
	// Generator name (lowercase, no spaces, "-" as word separator)
	Name string
	// Generator description
	Desc string
	// progressStore (TODO) can be used to store arbitrary data (GOB) before shutdown and retrieve
	// it again on the next startup.
	//
	// try should be called to try a hash string. If cont is returned as false, the generator must
	// return within a reasonable time frame.
	Fn func(info hashInfo, progressStore *any, try func([]byte) (cont bool)) error
}

type worker struct {
	name     string
	active   atomic.Bool
	ctxErr   error
	err      error
	triesCnt int
	_padding [128]byte // prevent false sharing (yes this much is a bit wasteful, but we don't have many workers so who cares)
}

type cracker struct {
	ctx context.Context
	prt app.Printer
	// All hashes present in the game
	// (the ones we're trying to crack and the
	// ones aready known).
	hashes map[stingray.Hash]struct{}
	// Currently known hashes
	knownHashes map[string]struct{}

	workers []worker

	mu sync.Mutex
	// Guarded by mu //
	newHashes         []string // newly cracked hashes
	workerMessagesBuf []string
	// End           //

	lastUpdateTime      time.Time
	numTriesSinceUpdate int
	hashesPerSecond     float64
}

func crack(ctx context.Context, generators []generator, prt app.Printer, a *app.App) (newHashes []string, err error) {
	hashes := make(map[stingray.Hash]struct{})
	for k := range a.DataDir.Files {
		hashes[k.Name] = struct{}{}
		hashes[k.Type] = struct{}{}
	}
	knownHashes := make(map[string]struct{})
	for h, s := range a.Hashes {
		if _, exists := hashes[h]; exists {
			knownHashes[s] = struct{}{}
		}
	}
	c := cracker{
		ctx:         ctx,
		prt:         prt,
		hashes:      hashes,
		knownHashes: knownHashes,
	}
	c.updateCliStatus()

	var wg sync.WaitGroup
	c.workers = make([]worker, len(generators))
	for i := range c.workers {
		c.workers[i].name = generators[i].Name
		var ps any // TODO: Implement progress store!!!
		wg.Go(func() {
			c.msgFromWorker(i, "Starting")
			c.workers[i].active.Store(true)
			c.workers[i].err = generators[i].Fn(
				hashInfo{
					hashes:      c.hashes,
					knownHashes: c.knownHashes,
				},
				&ps,
				func(s []byte) bool { return c.try(s, i) },
			)
			c.workers[i].active.Store(false)
			c.msgFromWorker(i, "Finished")
		})
	}
	for range time.Tick(500 * time.Millisecond) {
		if c.getNumActiveWorkers() == 0 {
			c.prt.Infof("All workers done")
			break
		}
		if ctx.Err() != nil {
			c.prt.Infof("Shutting down workers...")
			break
		}
		c.updateCliStatus()
		c.mu.Lock()
		now := time.Now()
		dt := now.Sub(c.lastUpdateTime).Seconds()
		c.hashesPerSecond = float64(c.numTriesSinceUpdate) / dt
		c.lastUpdateTime = now
		c.numTriesSinceUpdate = 0
		if len(c.workerMessagesBuf) > 0 {
			for _, s := range c.workerMessagesBuf {
				c.prt.Infof("%s", s)
			}
			c.workerMessagesBuf = c.workerMessagesBuf[:0]
		}
		c.mu.Unlock()
	}
	wg.Wait()
	var firstErr error
	var numErrs int
	for i := range c.workers {
		w := &c.workers[i]
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
	if numErrs <= 1 {
		return nil, firstErr
	} else {
		return nil, fmt.Errorf("%w (and %v more errors)", firstErr, numErrs-1)
	}
}

func (c *cracker) getNumActiveWorkers() int {
	numActiveWorkers := 0
	for i := range c.workers {
		if c.workers[i].active.Load() {
			numActiveWorkers++
		}
	}
	return numActiveWorkers
}

func (c *cracker) updateCliStatus() {
	c.mu.Lock()
	c.prt.Statusf("Cracked %d/%d new hashes (%v/%v workers, %.2fMH/s)", len(c.newHashes), len(c.hashes)-len(c.knownHashes), c.getNumActiveWorkers(), len(c.workers), c.hashesPerSecond/1e6)
	c.mu.Unlock()
}

// Must be called from the thread of the worker with the given ID.
func (c *cracker) msgFromWorker(workerId int, format string, args ...any) {
	c.mu.Lock()
	msg := fmt.Sprintf(format, args...)
	msg = fmt.Sprintf("<worker %d (%s)> %s", workerId, c.workers[workerId].name, msg)
	c.workerMessagesBuf = append(c.workerMessagesBuf, msg)
	c.mu.Unlock()
}

// Must be called from the thread of the worker with the given ID.
//
// If cont is returned false, the worker must do its cleanup and return
// within a reasonable time.
func (c *cracker) try(s []byte, workerId int) (cont bool) {
	w := &c.workers[workerId]
	w.triesCnt++
	if w.triesCnt >= 65536 {
		c.mu.Lock()
		c.numTriesSinceUpdate += w.triesCnt
		c.mu.Unlock()
		w.triesCnt = 0
		if err := c.ctx.Err(); err != nil {
			w.ctxErr = err
			return false
		}
	}

	if _, exists := c.hashes[stingray.Sum(s)]; !exists {
		return true
	}
	if _, exists := c.knownHashes[string(s)]; exists {
		return true
	}
	c.mu.Lock()
	c.newHashes = append(c.newHashes, string(s))
	c.mu.Unlock()
	c.msgFromWorker(workerId, "%s=%s", stingray.Sum(s), s)
	return true
}

func main() {
	genByName := make(map[string]generator)
	var genNames []string
	for _, g := range generators {
		genByName[g.Name] = g
		genNames = append(genNames, g.Name)
	}
	slices.Sort(genNames)

	var epilog strings.Builder
	{
		epilog.WriteString("generators:\n")
		tw := tabwriter.NewWriter(&epilog, 0, 1, 2, ' ', 0)
		for _, g := range generators {
			fmt.Fprintf(tw, "  %s\t%s\n", g.Name, g.Desc)
		}
		tw.Flush()
	}

	argp := argparse.NewParser("hd2-hash-cracker", "HD2 hash cracking tool", &argparse.ParserConfig{
		EpiLog: epilog.String(),
	})
	optGenerators := argp.String("G", "generators", &argparse.Option{
		Help: "crack using comma-separated generators (options: see 'generators' section below)",
	})
	optWrite := argp.Flag("w", "write", &argparse.Option{
		Help: "write newly cracked hashes to hashes/cracked.txt (program must be run from project root)",
	})
	prt, a := fdtools.Init(argp)
	var selectedGens []generator
	for name := range strings.SplitSeq(*optGenerators, ",") {
		name = strings.TrimSpace(name)
		gen, ok := genByName[name]
		if !ok {
			prt.Fatalf("Unknown generator %q (options: %s [see --help for descriptions])", name, strings.Join(genNames, ", "))
		}
		selectedGens = append(selectedGens, gen)
	}
	prt.Infof("Ctrl+C to quit")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := func() error {
		newHashes, err := crack(ctx, selectedGens, prt, a)
		if err != nil {
			return err
		}
		if *optWrite {
			const crackedFile = "hashes/cracked.txt"
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
