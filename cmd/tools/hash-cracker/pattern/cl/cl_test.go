package cl

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern"
	"github.com/xypwn/filediver/stingray"

	cl "github.com/CyberChainXyz/go-opencl"
)

func TestCl(t *testing.T) {
	require := require.New(t)
	prog, err := pattern.Compile([]byte("<a|b|c|d|e|f|g|h|i|j>{1,6}"), pattern.CompileOptions{})
	require.NoError(err)
	//fmt.Println(prog)

	info, err := cl.Info()
	require.NoError(err)

	if len(info.Platforms) < 1 {
		require.Fail("No OpenCL Devices")
	}
	if len(info.Platforms[0].Devices) < 1 {
		require.Fail("No OpenCL Devices")
	}

	device := info.Platforms[0].Devices[0]
	runner, err := device.InitRunner()
	require.NoError(err)
	defer runner.Free()

	targetHashes := []stingray.Hash{
		stingray.Sum("aba"),
		stingray.Sum("aaaa"),
		stingray.Sum("aaaaa"),
		stingray.Sum("aaaaaa"),
		stingray.Sum("abcdef"),
	}

	var allMatches []string
	cr, err := NewCracker(runner, prog, targetHashes, Options{})
	require.NoError(err)
	for {
		//fmt.Println(cr.TotalIdx(), "/", prog.Comp)
		matches, err := cr.Dispatch()
		if err == Done {
			break
		}
		require.NoError(err)
		allMatches = append(allMatches, matches...)
	}
	require.Equal([]string{"aba", "aaaa", "aaaaa", "aaaaaa", "abcdef"}, allMatches)
	//fmt.Println(allMatches)
}
