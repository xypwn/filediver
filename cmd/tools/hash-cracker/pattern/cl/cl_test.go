package cl

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern"
	"github.com/xypwn/filediver/stingray"

	cl "github.com/xypwn/gocl/cl-3.1"
)

func TestCl(t *testing.T) {
	require := require.New(t)
	prog, err := pattern.Compile([]byte("<a|b|c|d|e|f|g|h|i|j>{1,6}"), pattern.CompileOptions{})
	require.NoError(err)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	platforms, err := cl.GetPlatformIDs()
	require.NoError(err)
	if len(platforms) == 0 {
		require.Fail("no OpenCL platforms")
	}
	platform := platforms[0]
	devices, err := cl.GetDeviceIDs(platform, cl.DEVICE_TYPE_GPU)
	require.NoError(err)
	if len(devices) == 0 {
		require.Fail("no OpenCL devices")
	}
	device := devices[0]

	targetHashes := []stingray.Hash{
		stingray.Sum("aba"),
		stingray.Sum("aaaa"),
		stingray.Sum("aaaaa"),
		stingray.Sum("aaaaaa"),
		stingray.Sum("abcdef"),
	}

	var allMatches []string
	cr, err := NewCracker(device, prog, targetHashes, Options{})
	require.NoError(err)
	defer cr.Delete()
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
