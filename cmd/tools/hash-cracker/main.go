package main

import (
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
	"time"

	"github.com/hellflame/argparse"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/tools/fdtools-common"
	"github.com/xypwn/filediver/stingray"
)

var reAudioPath = regexp.MustCompile(`^content/audio(/[a-z]{2})?/[0-9]+$`)

type generatorContext struct {
	hashes      map[stingray.Hash]struct{}
	knownHashes map[string]struct{}
}

// generator is a hash string generator.
type generator struct {
	Name string
	// ctx is used for cancellation and the function MUST return within a reasonable time frame
	// after a cancellation is issued. Use the return value of try to poll ctx.Err() only some times
	// as to not impact generator performance too much.
	//
	// progressStore (TODO) can be used to store arbitrary data (GOB) before shutdown and retrieve
	// it again on the next startup.
	//
	// try should be called to try a hash string. doHousekeeping is returned as true every hundred
	// thousand or so iterations, which is when e.g. the context should be polled.
	Fn func(ctx context.Context, aCtx generatorContext, progressStore *any, try func([]byte) (doHousekeeping bool)) error
}

type worker struct {
	name     string
	active   atomic.Bool
	err      error
	triesCnt int
	_padding [128]byte // prevent false sharing (yes this much is a bit wasteful, but we don't have many workers so who cares)
}

type cracker struct {
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

func newCracker(prt app.Printer, a *app.App) *cracker {
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
	c := &cracker{
		prt:         prt,
		hashes:      hashes,
		knownHashes: knownHashes,
	}
	c.updateCliStatus()
	return c
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

func (c *cracker) start(ctx context.Context, generators []generator) error {
	var wg sync.WaitGroup
	c.workers = make([]worker, len(generators))
	for i := range c.workers {
		c.workers[i].name = generators[i].Name
		var ps any // TODO: Implement progress store!!!
		wg.Go(func() {
			c.msgFromWorker(i, "Starting")
			c.workers[i].active.Store(true)
			c.workers[i].err = generators[i].Fn(
				ctx,
				generatorContext{
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
	for i := range c.workers {
		err := c.workers[i].err
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
	return nil
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
func (c *cracker) try(s []byte, workerId int) (doHousekeeping bool) {
	w := &c.workers[workerId]
	w.triesCnt++
	if w.triesCnt >= 65536 {
		c.mu.Lock()
		c.numTriesSinceUpdate += w.triesCnt
		c.mu.Unlock()
		w.triesCnt = 0
		doHousekeeping = true
	}

	if _, exists := c.hashes[stingray.Sum(s)]; !exists {
		return
	}
	if _, exists := c.knownHashes[string(s)]; exists {
		return
	}
	c.mu.Lock()
	c.newHashes = append(c.newHashes, string(s))
	c.mu.Unlock()
	c.msgFromWorker(workerId, "%s=%s", stingray.Sum(s), s)
	return
}

func main() {
	genByName := make(map[string]generator)
	var genNames []string
	for _, g := range generators {
		genByName[g.Name] = g
		genNames = append(genNames, g.Name)
	}
	slices.Sort(genNames)
	genNamesStr := strings.Join(genNames, ",")

	argp := argparse.NewParser("hd2-hash-cracker", "HD2 hash cracking tool", &argparse.ParserConfig{})
	optGenerators := argp.String("G", "generators", &argparse.Option{
		Help: "crack using comma-separated generators (options: " + genNamesStr + ")",
	})
	prt, a := fdtools.Init(argp)
	var selectedGens []generator
	for name := range strings.SplitSeq(*optGenerators, ",") {
		name = strings.TrimSpace(name)
		gen, ok := genByName[name]
		if !ok {
			prt.Fatalf("Unknown generator %q (options: %s)", name, genNamesStr)
		}
		selectedGens = append(selectedGens, gen)
	}
	prt.Infof("Ctrl+C to quit")
	c := newCracker(prt, a)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := func() error {
		return c.start(ctx, selectedGens)
	}(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
