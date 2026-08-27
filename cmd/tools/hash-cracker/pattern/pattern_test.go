package pattern

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		expr    string
		want    IrSegment
		wantErr string
	}{
		{"basic", "<a|b>cd",
			IrSegmentConcat{
				IrSegmentChoice{
					IrSegmentStr("a"),
					IrSegmentStr("b"),
				},
				IrSegmentStr("cd"),
			}, ""},
		{"repeat", "<a|b>{5,10}",
			IrSegmentRepeat{
				Seg: IrSegmentChoice{
					IrSegmentStr("a"),
					IrSegmentStr("b"),
				},
				Min: 5,
				Max: 10,
			}, ""},
		{"repeat 1", "<a|b>{10}",
			IrSegmentRepeat{
				Seg: IrSegmentChoice{
					IrSegmentStr("a"),
					IrSegmentStr("b"),
				},
				Min: 10,
				Max: 10,
			}, ""},
		{"repeat sep", "<a|b>{5,10,<c|d>}",
			IrSegmentRepeat{
				Seg: IrSegmentChoice{
					IrSegmentStr("a"),
					IrSegmentStr("b"),
				},
				Min: 5,
				Max: 10,
				Sep: IrSegmentChoice{
					IrSegmentStr("c"),
					IrSegmentStr("d"),
				},
			}, ""},
		{"repeat invalid range", "<a|b>{10,5}", nil,
			"pattern error at position 11: expected minimum repetitions (10) to be less than maximum (5) repetitions"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require := require.New(t)
			_, irSeg, err := parse([]byte(c.expr), nil, nil)
			if c.wantErr == "" {
				require.NoError(err)
			} else {
				require.EqualError(err, c.wantErr)
			}
			require.Equal(c.want, irSeg)
		})
	}
}

func TestCompile(t *testing.T) {
	cases := []struct {
		name    string
		expr    string
		opts    CompileOptions
		want    Segment
		wantErr string
	}{
		{"basic", "a|b",
			CompileOptions{},
			Segment{
				Type: SegmentProdOfSets,
				Segs: [][]Segment{
					{
						{Type: SegmentText, Str: "a", Comp: 1},
						{Type: SegmentText, Str: "b", Comp: 1},
					},
				},
				Comp:  2,
				Comps: []int{2},
			}, "",
		},
		{"concat", "c<a|b>d",
			CompileOptions{},
			Segment{
				Type: SegmentProdOfSets,
				Segs: [][]Segment{
					{
						{Type: SegmentText, Str: "c", Comp: 1},
					},
					{
						{Type: SegmentText, Str: "a", Comp: 1},
						{Type: SegmentText, Str: "b", Comp: 1},
					},
					{
						{Type: SegmentText, Str: "d", Comp: 1},
					},
				},
				Comp:  2,
				Comps: []int{1, 2, 1},
			}, "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require := require.New(t)
			prog, err := Compile([]byte(c.expr), c.opts)
			if c.wantErr == "" {
				require.NoError(err)
			} else {
				require.EqualError(err, c.wantErr)
			}
			require.Equal(c.want, prog)
		})
	}
}

func TestAdd(t *testing.T) {
	cases := []struct {
		name  string
		expr  string
		nwant []any
	}{
		{"basic", "<0|1|2|3|4|5|6|7|8|9>{10}",
			[]any{
				69420, "0000069420",
				12345678, "0012415098",
			},
		},
		{"nested_or", "<0|1|2|3|4|5|6|7|8|9>{3}|one_thousand",
			[]any{
				999, "999",
				1, "one_thousand",
			},
		},
	}
	for _, c := range cases {
		for _, optimize := range []bool{false, true} {
			name := c.name
			if optimize {
				name += "_opt"
			} else {
				name += "_no_opt"
			}
			t.Run(name, func(t *testing.T) {
				require := require.New(t)
				prog, err := Compile([]byte(c.expr), CompileOptions{NoOptimize: !optimize})
				require.NoError(err)
				idx := prog.MakeIndex()
				for i := 0; i < len(c.nwant); i += 2 {
					n := c.nwant[i].(int)
					want := c.nwant[i+1].(string)
					idx.Add(prog, n)
					require.Equal(want, prog.StringAt(idx))
				}
			})
		}
	}
}
