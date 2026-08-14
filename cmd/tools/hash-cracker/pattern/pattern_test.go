package pattern

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
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
