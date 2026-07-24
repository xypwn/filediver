package main

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/hellflame/argparse"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/cmd/tools/fdtools-common"
	"github.com/xypwn/filediver/stingray"
)

var reExcludeExample = regexp.MustCompile(`^content/audio(/[a-z]{2})?/[0-9]+$`)

type cracker struct {
	prt app.Printer
	// All hashes present in the game
	// (the ones we're trying to crack and the
	// ones aready known).
	hashes map[stingray.Hash]struct{}
	// Currently known hashes
	knownHashes map[string]struct{}
	// Newly cracked hashes (by the cracker)
	newHashes []string

	hashesPerSecond     float64
	lastUpdateTime      time.Time
	numTriesSinceUpdate int
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

func (c *cracker) updateCliStatus() {
	c.prt.Statusf("Cracked %d/%d new hashes (%.2fMH/s)", len(c.newHashes), len(c.hashes)-len(c.knownHashes), c.hashesPerSecond/1e6)
}

func (c *cracker) try(s []byte) {
	// We want to compromise the hash rate as little as possible,
	// so only measure time every ~1million tries and only update
	// everything at most once per second.
	c.numTriesSinceUpdate++
	if c.numTriesSinceUpdate&((1<<20)-1) /*~1million*/ == 0 {
		if dt := time.Since(c.lastUpdateTime).Seconds(); dt >= 1 {
			c.updateCliStatus()
			c.hashesPerSecond = float64(c.numTriesSinceUpdate) / dt
			c.lastUpdateTime = time.Now()
			c.numTriesSinceUpdate = 0
		}
	}

	if _, exists := c.hashes[stingray.Sum(s)]; !exists {
		return
	}
	if _, exists := c.knownHashes[string(s)]; exists {
		return
	}
	c.prt.Infof("%s=%s", stingray.Sum(s), string(s))
	c.newHashes = append(c.newHashes, string(s))
}

// Super basic brute force algorithm that
// reshuffles known path segments. I plan to
// add more sophisticated algorithms.
func makeGuesses(c *cracker, examples []string) error {
	var words []string
	seenWords := make(map[string]struct{})
	for _, example := range examples {
		if reExcludeExample.MatchString(example) {
			continue
		}
		for word := range strings.FieldsFuncSeq(example, func(r rune) bool {
			return slices.Contains([]rune("/"), r)
		}) {
			if _, seen := seenWords[word]; !seen {
				words = append(words, word)
				seenWords[word] = struct{}{}
			}
		}
	}

	// Zeroes the slice of ints.
	zero := func(s []int) {
		for i := range s {
			s[i] = 0
		}
	}
	// Increments an array of indexes by one,
	// where max is the (exclusive) maximum
	// value for each index. Returns false if
	// the value overflows.
	inc := func(s []int, max int) (ok bool) {
		i := len(s) - 1
		for {
			s[i]++
			if s[i] < max {
				return true
			} else {
				s[i] = 0
				i--
				if i < 0 {
					return false
				}
			}
		}
	}
	var buf bytes.Buffer
	var idxs []int
	for { // iterate over lengths
		zero(idxs)
		idxs = append(idxs, 0)

		for { // iterate over permutations with length
			for i, idx := range idxs {
				if i != 0 {
					buf.WriteString("/")
				}
				buf.WriteString(words[idx])
			}
			c.try(buf.Bytes())
			buf.Reset()
			if !inc(idxs, len(words)) {
				break
			}
		}
	}
}

func main() {
	argp := argparse.NewParser("hd2-hash-cracker", "HD2 hash cracking tool", &argparse.ParserConfig{
		DisableDefaultShowHelp: true,
	})
	prt, a := fdtools.Init(argp)
	c := newCracker(prt, a)
	if err := func() error {
		if err := makeGuesses(c, slices.Collect(maps.Keys(c.knownHashes))); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
