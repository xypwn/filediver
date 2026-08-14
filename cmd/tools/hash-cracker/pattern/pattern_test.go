package pattern

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		expr    string
		want    irSegment
		wantErr string
	}{
		{"basic", "<a|b>cd",
			irSegmentConcat{
				irSegmentChoice{
					irSegmentStr("a"),
					irSegmentStr("b"),
				},
				irSegmentStr("cd"),
			}, ""},
		{"repeat", "<a|b>{5,10}",
			irSegmentRepeat{
				Seg: irSegmentChoice{
					irSegmentStr("a"),
					irSegmentStr("b"),
				},
				Min: 5,
				Max: 10,
			}, ""},
		{"repeat 1", "<a|b>{10}",
			irSegmentRepeat{
				Seg: irSegmentChoice{
					irSegmentStr("a"),
					irSegmentStr("b"),
				},
				Min: 10,
				Max: 10,
			}, ""},
		{"repeat invalid range", "<a|b>{10,5}", nil,
			"pattern error at position 11: expected minimum repetitions (10) to be less than maximum (5) repetitions"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require := require.New(t)
			irSeg, err := parse([]byte(c.expr))
			if c.wantErr == "" {
				require.NoError(err)
			} else {
				require.EqualError(err, c.wantErr)
			}
			require.Equal(c.want, irSeg)
		})
	}
}

func TestHello(t *testing.T) {
	require := require.New(t)
	//irSeg, err := parse([]byte("<$hello|hi{1,2}>{1,2}"))
	irSeg, err := parse([]byte("<<0|1|2|3|4|5|6|7|8|9>{3}><0|1|2|3|4|5|6|7|8|9>{2}"))
	//irSeg, err := parse([]byte("<0|1|2|3|4|5|6|7|8|9>{6}"))
	//irSeg, err := parse([]byte("<0|1|2|3|4|5|6|7>{10}"))
	require.NoError(err)
	fmt.Println("ir", irSeg)

	prog := compile(irSeg)
	fmt.Println("prog", prog)
	//fmt.Println(prog.Comp())
	fmt.Println(prog.comp)
	idx := prog.makeIndex()
	//fmt.Println(idx)
	fmt.Println(prog.StringAt(idx))
	//fmt.Println("carry =", idx.Add(prog, 12345))
	fmt.Println("carry =", idx.Add(prog, 100))
	fmt.Println(prog.StringAt(idx))
	fmt.Println("carry =", idx.Add(prog, 100))
	fmt.Println(prog.StringAt(idx))
	fmt.Println(idx)
}
