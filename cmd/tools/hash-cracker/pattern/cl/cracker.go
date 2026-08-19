package cl

import (
	"errors"
	"fmt"

	cl "github.com/CyberChainXyz/go-opencl"
	"github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern"
	"github.com/xypwn/filediver/stingray"
)

var Done = errors.New("cracker done")

// Advanced options for the cracker.
//
// Any zeroed field will be treated as if set to
// its default value.
type Options struct {
	NumWorkers                int // default 4096
	MinMatchBufLen            int // default 128
	TriesPerWorkerPerDispatch int // default 65536
}

type Cracker struct {
	Runner    *cl.OpenCLRunner
	DebugInfo struct {
		OpenClCode string
	}

	opts     Options
	prog     pattern.Segment
	bufs     *clBuffers
	idx      pattern.SegIdx
	totalIdx int
}

func NewCracker(runner *cl.OpenCLRunner, prog pattern.Segment, targetHashes []stingray.Hash, opts Options) (*Cracker, error) {
	if opts.NumWorkers == 0 {
		opts.NumWorkers = 4096
	}
	if opts.MinMatchBufLen == 0 {
		opts.MinMatchBufLen = 128
	}
	if opts.TriesPerWorkerPerDispatch == 0 {
		opts.TriesPerWorkerPerDispatch = 65536
	}
	c := &Cracker{
		Runner: runner,
		prog:   prog,
		idx:    prog.MakeIndex(),
		opts:   opts,
	}
	matchBufLen := opts.MinMatchBufLen
	matchBufLen = max(matchBufLen, 2*(prog.MaxLen()+1))
	var err error
	c.bufs, err = makeClBuffers(runner, prog, targetHashes, opts.NumWorkers, matchBufLen)
	if err != nil {
		return nil, fmt.Errorf("creating OpenCL buffers: %w", err)
	}
	code := string(generateClCode(prog, c.bufs))
	c.DebugInfo.OpenClCode = code
	if err := runner.CompileKernels([]string{code}, []string{"kmain"}, ""); err != nil {
		return nil, fmt.Errorf("compiling OpenCL kernel: %w", err)
	}
	return c, nil
}

// TotalIdx returns the total iteration index,
// which is approximately the number of total
// strings checked.
func (c *Cracker) TotalIdx() int {
	return c.totalIdx
}

// Dispatch dispatches a single batch of workers.
//
// Returns a slice with any newly found matching strings.
//
// Returns nil, [Done] when done.
func (c *Cracker) Dispatch() (matches []string, err error) {
	done := true
	for i := range c.bufs.numWorkers {
		if c.bufs.data.tries[i] != 0 {
			// Worker stopped early because result
			// buffer was full.
			done = false
			continue
		}
		c.idx.Reset()
		if c.totalIdx > 0 {
			c.idx.Add(c.prog, c.totalIdx)
		}
		readIdx(c.bufs.data.idxs[i*c.bufs.idxLen:], c.prog, c.idx)
		tries := c.opts.TriesPerWorkerPerDispatch
		if c.totalIdx+tries > c.prog.Comp {
			tries = c.prog.Comp - c.totalIdx
		}
		c.bufs.data.tries[i] = uint32(tries)
		c.totalIdx += tries
		if tries > 0 {
			done = false
		}
	}
	if done {
		return nil, Done
	}
	//fmt.Println("tries:", c.bufs.data.tries)
	//fmt.Println("matches:", c.bufs.data.matchesLens)
	if err := c.bufs.write(c.Runner); err != nil {
		return nil, fmt.Errorf("writing OpenCL buffers: %w", err)
	}
	if err := c.Runner.RunKernel("kmain", 1, nil, []uint64{uint64(c.bufs.numWorkers)}, nil, []cl.KernelParam{
		cl.BufferParam(c.bufs.cl.tries),
		cl.BufferParam(c.bufs.cl.idxs),
		cl.BufferParam(c.bufs.cl.strs),
		cl.BufferParam(c.bufs.cl.strsOffsets),
		cl.BufferParam(c.bufs.cl.strLens),
		cl.BufferParam(c.bufs.cl.hashBitmap),
		cl.BufferParam(c.bufs.cl.targetHashes),
		cl.BufferParam(c.bufs.cl.matches),
		cl.BufferParam(c.bufs.cl.matchesLens),
	}, true); err != nil {
		return nil, fmt.Errorf("running OpenCL kernel: %w", err)
	}
	if err := c.bufs.read(c.Runner); err != nil {
		return nil, fmt.Errorf("reading back OpenCL buffers: %w", err)
	}
	matches = c.bufs.drainMatches()
	return
}
