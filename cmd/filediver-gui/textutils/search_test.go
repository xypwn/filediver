package textutils_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xypwn/filediver/cmd/filediver-gui/textutils"
)

func TestMatchesAny(t *testing.T) {
	knownName := "content/fac_helldivers/textures/helmet_albedo.texture"
	hashName := "0123456789abcdef.0123456789abcdef"

	cases := []struct {
		name  string
		query string
		words []string
		want  bool
	}{
		{"empty query", "", []string{knownName, hashName}, true},

		{"include", "helmet", []string{knownName, hashName}, true},
		{"include case-insensitive", "HELMET", []string{knownName, hashName}, true},
		{"include in hash name", "0123456789abcdef", []string{knownName, hashName}, true},
		{"include miss", "nonexistent", []string{knownName, hashName}, false},
		{"includes in same word", "helmet texture", []string{knownName, hashName}, true},
		{"includes split across words", "helmet 0123456789abcdef", []string{knownName, hashName}, false},

		{"exclusion in known name", "-helmet", []string{knownName, hashName}, false},
		{"exclusion in hash name", "-abcdef", []string{knownName, hashName}, false},
		{"exclusion miss", "-nonexistent", []string{knownName, hashName}, true},
		{"exclusion in other word than include", "helmet -abcdef", []string{knownName, hashName}, false},
		{"include with exclusion miss", "helmet -foo", []string{knownName, hashName}, true},
		{"escaped dash", `\-bar`, []string{"content/foo-bar.unit"}, true},

		{"diacritics", "cafe", []string{"content/café/helmet.texture"}, true},
		{"diacritics exclusion", "-cafe", []string{"content/café/helmet.texture"}, false},
		{"non-ascii query", "müller", []string{"content/cafe/muller.texture"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			query := textutils.NewQuery(c.query)
			require.Equal(t, c.want, query.MatchesAny(c.words...))
		})
	}
}

func BenchmarkMatchesAnyAscii(b *testing.B) {
	benchMatchesAny(b, "content/textures/helmet_%04d_albedo.texture")
}

func BenchmarkMatchesAnyNonAscii(b *testing.B) {
	benchMatchesAny(b, "content/textures/café_helmet_%04d_albedo.texture")
}

func benchMatchesAny(b *testing.B, nameFormat string) {
	queries := []string{"helmet", "nonexistent", "-nonexistent", "cafe"}
	rng := rand.New(rand.NewPCG(1, 2))
	words := make([][2]string, 1000)
	for i := range words {
		words[i] = [2]string{
			fmt.Sprintf(nameFormat, i),
			fmt.Sprintf("%016x.texture", rng.Uint64()),
		}
	}
	for _, q := range queries {
		query := textutils.NewQuery(q)
		b.Run(q, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				for _, w := range words {
					query.MatchesAny(w[0], w[1])
				}
			}
		})
	}
}
