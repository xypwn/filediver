package main

import (
	"bytes"
	"cmp"
	"context"
	"maps"
	"slices"
	"strings"
)

var generators = []generator{
	// Reshuffles known path segments.
	{
		Name: "reshuffle",
		Fn: func(ctx context.Context, aCtx generatorContext, progressStore *any, try func([]byte) (doHousekeeping bool)) error {
			examples := slices.Sorted(maps.Keys(aCtx.knownHashes))

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
			if *progressStore != nil {
				idxs, _ = (*progressStore).([]int)
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
					doHousekeeping := try(buf.Bytes())
					buf.Reset()
					if !inc(idxs, len(words)) {
						break
					}

					if doHousekeeping {
						if err := ctx.Err(); err != nil {
							*progressStore = idxs
							return err
						}
					}
				}
			}
		},
	},
	// Tries to concatenate together existing prefix and suffix segments.
	{
		Name: "prefix-suffix",
		Fn: func(ctx context.Context, aCtx generatorContext, progressStore *any, try func([]byte) (doHousekeeping bool)) error {
			examples := slices.Sorted(maps.Keys(aCtx.knownHashes))

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
					doHousekeeping := try(buf.Bytes())
					if doHousekeeping {
						if err := ctx.Err(); err != nil {
							return err
						}
					}
				}
			}
			return nil
		},
	},
}
