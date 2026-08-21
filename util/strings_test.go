package util_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xypwn/filediver/util"
)

func TestSplitStringAny(t *testing.T) {
	cases := []struct {
		name  string
		s     string
		chars string
		want  []string
	}{
		{"simple", "a,bc,d", ",", []string{"a", "bc", "d"}},
		{"repeating_sep", "a,bc,,d", ",", []string{"a", "bc", "", "d"}},
		{"sep_at_beginning_and_end", ",,a,b,,", ",", []string{"", "", "a", "b", "", ""}},
		{"multiple_seps", "a,b;c:d", ",;:", []string{"a", "b", "c", "d"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, util.SplitStringAny(c.s, c.chars))
		})
		t.Run(c.name+"_seq", func(t *testing.T) {
			require.Equal(t, c.want, slices.Collect(util.SplitStringAnySeq(c.s, c.chars)))
		})
	}
}
