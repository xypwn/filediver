package cl

import (
	"errors"
	"fmt"

	"github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern"
	"github.com/xypwn/filediver/stingray"
	cl "github.com/xypwn/gocl/cl-3.1"
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
	DebugInfo struct {
		OpenClCode string
	}

	device   cl.DeviceId
	context  cl.Context
	queue    cl.CommandQueue
	kernel   cl.Kernel
	opts     Options
	prog     pattern.Segment
	bufs     *clBuffers
	idx      pattern.SegIdx
	totalIdx int
}

func NewCracker(device cl.DeviceId, prog pattern.Segment, targetHashes []stingray.Hash, opts Options) (_ *Cracker, err error) {
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
		prog: prog,
		idx:  prog.MakeIndex(),
		opts: opts,
	}
	defer func() {
		if err != nil {
			if clErr := (*cl.Error)(nil); errors.As(err, &clErr) {
				err = fmt.Errorf("OpenCL: %w", err)
			}
			c.Delete()
		}
	}()

	c.context, err = cl.CreateContext(nil, []cl.DeviceId{device}, nil)
	if err != nil {
		return nil, err
	}

	matchBufLen := opts.MinMatchBufLen
	matchBufLen = max(matchBufLen, 2*(prog.MaxLen()+1))
	c.bufs, err = makeClBuffers(c.context, prog, targetHashes, opts.NumWorkers, matchBufLen)
	if err != nil {
		return nil, fmt.Errorf("creating buffers: %w", err)
	}

	code := string(generateClCode(prog, c.bufs))
	c.DebugInfo.OpenClCode = code
	program, err := cl.CreateProgramWithSource(c.context, []string{code})
	if err != nil {
		return nil, fmt.Errorf("creating program: %w", err)
	}
	defer cl.ReleaseProgram(program)
	if err := cl.BuildProgram(program, []cl.DeviceId{device}, "", nil); err != nil {
		var errLog string
		if err := cl.GetProgramBuildInfo(program, device, cl.PROGRAM_BUILD_LOG, &errLog); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("compiling program: %s", errLog)
	}
	c.kernel, err = cl.CreateKernel(program, "kmain")
	if err != nil {
		return nil, fmt.Errorf("creating kernel: %w", err)
	}
	if err := cl.SetKernelArgValues(c.kernel, 0,
		c.bufs.cl.tries, c.bufs.cl.idxs,
		c.bufs.cl.strs, c.bufs.cl.strsOffsets, c.bufs.cl.strLens,
		c.bufs.cl.hashBitmap, c.bufs.cl.targetHashes,
		c.bufs.cl.matches, c.bufs.cl.matchesLens); err != nil {
		return nil, fmt.Errorf("setting arg values: %w", err)
	}

	c.queue, err = cl.CreateCommandQueueWithProperties(c.context, device, nil)
	if err != nil {
		return nil, fmt.Errorf("creating command queue: %w", err)
	}
	return c, nil
}

func (c *Cracker) Delete() {
	if c.context != nil {
		cl.ReleaseContext(c.context)
	}
	if c.queue != nil {
		cl.ReleaseCommandQueue(c.queue)
	}
	if c.kernel != nil {
		cl.ReleaseKernel(c.kernel)
	}
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
	if err := c.bufs.write(c.queue); err != nil {
		return nil, fmt.Errorf("writing buffers: %w", err)
	}
	if err := cl.EnqueueNDRangeKernel(c.queue, c.kernel, 1, nil, []uint64{uint64(c.bufs.numWorkers)}, nil, nil, nil); err != nil {
		return nil, fmt.Errorf("running kernel: %w", err)
	}
	if err := cl.Finish(c.queue); err != nil {
		return nil, fmt.Errorf("finishing queue: %w", err)
	}
	if err := c.bufs.read(c.queue); err != nil {
		return nil, fmt.Errorf("reading back buffers: %w", err)
	}
	matches = c.bufs.drainMatches()
	return
}
