package main

import (
	"bytes"
	"cmp"
	"fmt"
	"maps"
	"os"
	"runtime"
	"slices"
	"strings"

	cl "github.com/CyberChainXyz/go-opencl"
	"github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern"
	pcl "github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern/cl"
)

var crackers = []Cracker{
	{
		Name: "reshuffle",
		Desc: "reshuffle known path segments",
		Fn: func(c Context) error {
			examples := slices.Sorted(maps.Keys(c.Info().KnownHashes))

			var words []string
			seenWords := make(map[string]struct{})
			for _, example := range examples {
				if reAudioPath.MatchString(example) {
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
			slices.Sort(words)

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
			if ps, ok := c.ProgressStore(); ok {
				idxs, _ = ps.([]int)
			}
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
					cont := c.Try(buf.Bytes())
					buf.Reset()
					if !inc(idxs, len(words)) {
						break
					}

					if !cont {
						c.SetProgressStore(idxs)
						return nil
					}
				}
			}
		},
	},
	{
		Name: "prefix-suffix",
		Desc: "concatenate together existing prefix and suffix segments",
		Fn: func(c Context) error {
			examples := slices.Sorted(maps.Keys(c.Info().KnownHashes))

			prefixesByN := make(map[string]int) // prefix to number of occurrences
			suffixesByN := make(map[string]int) // suffix to number of occurrences
			for _, example := range examples {
				if reAudioPath.MatchString(example) {
					continue
				}
				var words []string // filename split before prefixes
				{
					i := 0
					for {
						var j int
						if i < len(example) {
							j = strings.IndexAny(example[i+1:], "/_:")
						} else {
							j = len(example)
						}
						if j == -1 {
							words = append(words, example[i:])
							break
						} else {
							words = append(words, example[i:i+1+j])
							i += 1 + j
						}
					}
				}
				for i := 0; i <= len(words); i++ {
					pfx := strings.Join(words[:i], "")
					sfx := strings.Join(words[i:], "")
					if len(sfx) > 0 && slices.Contains([]byte("/_:"), sfx[0]) {
						sfx = sfx[1:]
					}
					prefixesByN[pfx]++
					suffixesByN[sfx]++
				}
			}
			// Sort descending by number of occurrences
			pfxs := slices.Collect(maps.Keys(prefixesByN))
			slices.SortFunc(pfxs, func(a, b string) int {
				return cmp.Compare(prefixesByN[b], prefixesByN[a])
			})
			sfxs := slices.Collect(maps.Keys(suffixesByN))
			slices.SortFunc(sfxs, func(a, b string) int {
				return cmp.Compare(suffixesByN[b], suffixesByN[a])
			})

			var buf bytes.Buffer
			for _, pfx := range pfxs {
				for _, sfx := range sfxs {
					buf.Reset()
					buf.WriteString(pfx)
					buf.WriteString(sfx)
					if !c.Try(buf.Bytes()) {
						return nil
					}
				}
			}
			return nil
		},
	},
	{
		Name:     "pattern",
		Desc:     "OpenCL-based cracker using pattern syntax. Use file:<filename> to load pattern from file path.",
		ArgNames: []string{"pattern"},
		Fn: func(c Context) error {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			argPat := c.Args()[0]
			var patSrc []byte
			if file, ok := strings.CutPrefix(argPat, "file:"); ok {
				var err error
				patSrc, err = os.ReadFile(file)
				if err != nil {
					return err
				}
			} else {
				patSrc = []byte(argPat)
			}
			prog, err := pattern.Compile(patSrc, pattern.CompileOptions{})
			if err != nil {
				return err
			}

			info, err := cl.Info()
			if err != nil {
				return err
			}

			if len(info.Platforms) == 0 {
				return fmt.Errorf("no OpenCL platforms")
			}
			if len(info.Platforms[0].Devices) == 0 {
				return fmt.Errorf("no OpenCL devices")
			}

			device := info.Platforms[0].Devices[0]
			runner, err := device.InitRunner()
			c.Msg("Initializing buffers and compiling OpenCL kernel")
			cr, err := pcl.NewCracker(runner, prog,
				slices.Collect(maps.Keys(c.Info().UnknownHashes)),
				pcl.Options{})
			if err != nil {
				return err
			}

			c.Msg("Making guesses")
			prevTotalIdx := 0
			idx := prog.MakeIndex()
			for {
				matches, err := cr.Dispatch()
				if err == pcl.Done {
					break
				} else if err != nil {
					return err
				}
				idx.Reset()
				if cr.TotalIdx() > 0 {
					idx.Add(prog, cr.TotalIdx())
				}
				if !c.ReportTries(cr.TotalIdx()-prevTotalIdx, matches, prog.AppendAt(nil, idx)) {
					break
				}
				prevTotalIdx = cr.TotalIdx()
			}
			return nil
		},
	},
}
