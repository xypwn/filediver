package pattern

import (
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xypwn/filediver/stingray"

	cl "github.com/CyberChainXyz/go-opencl"
)

func TestGenerateOpenClCode(t *testing.T) {
	require := require.New(t)
	irSeg, err := parse([]byte("<a|b|c|d|e|f|g|h|i|j>{1,10}"))
	//irSeg, err := parse([]byte("aaaaa|aaaa"))
	require.NoError(err)
	prog := compile(irSeg)
	fmt.Println(prog)

	info, _ := cl.Info()

	if len(info.Platforms) < 1 {
		log.Fatal("No OpenCL Devices")
	}
	if len(info.Platforms[0].Devices) < 1 {
		log.Fatal("No OpenCL Devices")
	}

	device := info.Platforms[0].Devices[0]
	runner, err := device.InitRunner()
	if err != nil {
		log.Fatal("InitRunner err:", err)
	}
	defer runner.Free()

	targetHashes := []stingray.Hash{
		stingray.Sum("aba"),
		stingray.Sum("aaaa"),
		stingray.Sum("aaaaa"),
		stingray.Sum("aaaaaa"),
		stingray.Sum("abcdef"),
	}
	numWorkers := 4096
	//numWorkers := 4
	//maxMatchBufLen := 5
	maxMatchBufLen := 64
	bufs, err := makeClBuffers(runner, prog, targetHashes, numWorkers, maxMatchBufLen)
	if err != nil {
		log.Fatal(err)
	}
	code := generateClCode(prog, bufs)
	fmt.Println(string(code))

	codes := []string{string(code)}
	kernelNameList := []string{"kmain"}
	tStart := time.Now()
	err = runner.CompileKernels(codes, kernelNameList, "")
	if err != nil {
		log.Fatal("CompileKernels err:", err)
	}
	fmt.Printf("Compiled in %v\n", time.Since(tStart))
	fmt.Println("idx len:", bufs.idxLen)

	var matches []string
	triesPerDispatch := 65536
	//triesPerDispatch := 64
	//triesPerDispatch := prog.comp
	var totalIdx int
	idx := prog.makeIndex()
	for {
		done := true
		idx.Reset()
		for i := range numWorkers {
			if bufs.data.tries[i] != 0 {
				// Worker stopped early because result
				// buffer was full.
				done = false
				continue
			}
			idx.Reset()
			if totalIdx > 0 {
				idx.Add(prog, totalIdx)
			}
			readIdx(bufs.data.idxs[i*bufs.idxLen:], prog, idx)
			tries := triesPerDispatch
			if totalIdx+tries > prog.comp {
				tries = prog.comp - totalIdx
			}
			bufs.data.tries[i] = uint32(tries)
			//fmt.Println("w", i, totalIdx, tries)
			totalIdx += tries
			if tries > 0 {
				done = false
			}
		}
		if done {
			break
		}
		if err := bufs.write(runner); err != nil {
			log.Fatal(err)
		}
		err = runner.RunKernel("kmain", 1, nil, []uint64{uint64(numWorkers)}, nil, []cl.KernelParam{
			cl.BufferParam(bufs.cl.tries),
			cl.BufferParam(bufs.cl.idxs),
			cl.BufferParam(bufs.cl.strs),
			cl.BufferParam(bufs.cl.strsOffsets),
			cl.BufferParam(bufs.cl.strLens),
			cl.BufferParam(bufs.cl.hashBitmap),
			cl.BufferParam(bufs.cl.targetHashes),
			cl.BufferParam(bufs.cl.matches),
			cl.BufferParam(bufs.cl.matchesLens),
		}, true)
		if err != nil {
			log.Fatal("RunKernel err:", err)
		}
		if err := bufs.read(runner); err != nil {
			log.Fatal(err)
		}
		matches = append(matches, bufs.drainMatches()...)
		//fmt.Println("tries:", bufs.data.tries)
		//fmt.Println("idxs:", bufs.data.idxs)
		fmt.Println("totalIdx:", totalIdx, "/", prog.comp, math.Log10(float64(prog.comp)))
	}

	fmt.Println("matches:", matches)
	fmt.Println(bufs.data.targetHashes)
}
