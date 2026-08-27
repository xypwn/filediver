package cl

import (
	_ "embed"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
	cl "github.com/xypwn/gocl/cl-3.1"
)

//go:embed prelude.cl
var preludeClCode string

// Returns the number of elements needed in the index
// for the given program.
func getSegmentIdxLen(s pattern.Segment) int {
	n := len(s.Segs)
	for _, segs := range s.Segs {
		for _, seg := range segs {
			n += getSegmentIdxLen(seg)
		}
	}
	return n
}

// Reads the SegIdx into a flat array.
func readIdx(dest []uint32, s pattern.Segment, idx pattern.SegIdx) int {
	p := 0
	for i := range len(idx.Idxs) {
		dest[p] = uint32(idx.Idxs[i])
		p++
	}
	for i, segs := range s.Segs {
		if s.Comps[i] != len(s.Segs[i]) {
			for j, seg := range segs {
				p += readIdx(dest[p:], seg, idx.Segs[i][j])
			}
		}
	}
	return p
}

type clBuffers struct {
	strArrs          [][]string
	strArrsFirstIdxs []int
	numWorkers       int
	idxLen           int
	matchBufLen      int
	maxCandidateLen  int
	data             struct {
		tries        []uint32
		idxs         []uint32
		strs         []byte
		strsOffsets  []uint32
		strLens      []uint32
		hashBitmap   []uint32
		targetHashes []uint64
		matches      []byte
		matchesLens  []uint32
	}
	cl struct {
		tries        cl.Mem
		idxs         cl.Mem
		strs         cl.Mem
		strsOffsets  cl.Mem
		strLens      cl.Mem
		hashBitmap   cl.Mem
		targetHashes cl.Mem
		matches      cl.Mem
		matchesLens  cl.Mem
	}
}

func makeClBuffers(context cl.Context, s pattern.Segment, targetHashes []stingray.Hash, numWorkers, matchBufLen int) (*clBuffers, error) {
	b := &clBuffers{
		numWorkers:      numWorkers,
		idxLen:          getSegmentIdxLen(s),
		matchBufLen:     matchBufLen,
		maxCandidateLen: s.MaxLen(),
	}

	if b.matchBufLen < b.maxCandidateLen+1 {
		return nil, fmt.Errorf("matchBufLen (%d) must be at least 1 more than the maximum candidate string length (%d)", b.matchBufLen, b.maxCandidateLen)
	}

	var check func(s pattern.Segment) error
	check = func(s pattern.Segment) error {
		if s.Comp == math.MaxInt {
			return fmt.Errorf("pattern: complexity must be less than 64-bit signed integer limit")
		}
		for _, segs := range s.Segs {
			if len(segs) > math.MaxUint32 {
				return fmt.Errorf("pattern: operands of union must be no more than the 32-bit unsigned integer limit (4294967295)")
			}
			for _, seg := range segs {
				if err := check(seg); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := check(s); err != nil {
		return nil, err
	}

	// Create tries buffer.
	b.data.tries = make([]uint32, numWorkers)

	// Create index buffer.
	b.data.idxs = make([]uint32, b.idxLen*numWorkers)

	// Collect string arrays for all cases
	// of unions of single string operands.
	{
		var addStrArrs func(s pattern.Segment)
		addStrArrs = func(s pattern.Segment) {
			if s.Type == pattern.SegmentText {
				return
			}
			for i, segs := range s.Segs {
				if len(segs) == s.Comps[i] {
					strs := make([]string, len(segs))
					for j, seg := range segs {
						strs[j] = seg.Str
					}
					b.strArrs = append(b.strArrs, strs)
				} else {
					for _, seg := range segs {
						addStrArrs(seg)
					}
				}
			}
		}
		addStrArrs(s)
		slices.SortFunc(b.strArrs, slices.Compare)
		b.strArrs = util.UniqFunc(b.strArrs, slices.Equal)

		b.strArrsFirstIdxs = make([]int, len(b.strArrs))
		var strArr []string
		for i := range b.strArrs {
			b.strArrsFirstIdxs[i] = len(strArr)
			strArr = append(strArr, b.strArrs[i]...)
		}

		b.data.strsOffsets = make([]uint32, len(strArr))
		b.data.strLens = make([]uint32, len(strArr))
		for i := range strArr {
			b.data.strsOffsets[i] = uint32(len(b.data.strs))
			b.data.strLens[i] = uint32(len(strArr[i]))
			b.data.strs = append(b.data.strs, strArr[i]...)
		}
	}

	// Create hash bitmap (bloom filter)
	{
		const BITS = 24
		b.data.hashBitmap = make([]uint32, 1<<(BITS-4))
		for _, h := range targetHashes {
			hl := uint32(h.Value)
			hh := uint32(h.Value >> 32)
			idxl := hl >> (32 - (BITS - 4))
			bitl := uint32(1) << (hl & 31)
			b.data.hashBitmap[idxl] |= bitl
			idxh := hh >> (32 - (BITS - 4))
			bith := uint32(1) << (hh & 31)
			b.data.hashBitmap[idxh] |= bith
		}
	}

	// Target hash array.
	{
		b.data.targetHashes = make([]uint64, len(targetHashes))
		for i := range targetHashes {
			b.data.targetHashes[i] = targetHashes[i].Value
		}

		// Sort and dedupe
		slices.Sort(b.data.targetHashes)
		b.data.targetHashes = util.Uniq(b.data.targetHashes)
	}

	// Match buffers
	b.data.matches = make([]byte, b.matchBufLen*numWorkers)
	b.data.matchesLens = make([]uint32, numWorkers)

	// Prevent errors from empty buffer
	if len(b.data.idxs) == 0 {
		b.data.idxs = []uint32{0}
	}
	if len(b.data.strs) == 0 {
		b.data.strsOffsets = []uint32{0}
		b.data.strLens = []uint32{0}
		b.data.strs = []byte{0}
	}
	if len(b.data.targetHashes) == 0 {
		b.data.targetHashes = []uint64{0}
	}

	// Create according OpenCL Buffers
	{
		var err error
		b.cl.tries, err = cl.CreateBufferSlice(context, cl.MEM_READ_WRITE, b.data.tries)
		if err != nil {
			return nil, fmt.Errorf("creating tries buffer: %w", err)
		}
		b.cl.idxs, err = cl.CreateBufferSlice(context, cl.MEM_READ_WRITE, b.data.idxs)
		if err != nil {
			return nil, fmt.Errorf("creating indices buffer: %w", err)
		}
		b.cl.strs, err = cl.CreateBufferSlice(context, cl.MEM_READ_ONLY|cl.MEM_COPY_HOST_PTR, b.data.strs)
		if err != nil {
			return nil, fmt.Errorf("creating strings buffer: %w", err)
		}
		b.cl.strsOffsets, err = cl.CreateBufferSlice(context, cl.MEM_READ_ONLY|cl.MEM_COPY_HOST_PTR, b.data.strsOffsets)
		if err != nil {
			return nil, fmt.Errorf("creating strings offsets buffer: %w", err)
		}
		b.cl.strLens, err = cl.CreateBufferSlice(context, cl.MEM_READ_ONLY|cl.MEM_COPY_HOST_PTR, b.data.strLens)
		if err != nil {
			return nil, fmt.Errorf("creating string lengths buffer: %w", err)
		}
		b.cl.hashBitmap, err = cl.CreateBufferSlice(context, cl.MEM_READ_ONLY|cl.MEM_COPY_HOST_PTR, b.data.hashBitmap)
		if err != nil {
			return nil, fmt.Errorf("creating hash bitmap buffer: %w", err)
		}
		b.cl.targetHashes, err = cl.CreateBufferSlice(context, cl.MEM_READ_ONLY|cl.MEM_COPY_HOST_PTR, b.data.targetHashes)
		if err != nil {
			return nil, fmt.Errorf("creating target hash buffer: %w", err)
		}
		b.cl.matches, err = cl.CreateBufferSlice(context, cl.MEM_WRITE_ONLY, b.data.matches)
		if err != nil {
			return nil, fmt.Errorf("creating match buffer: %w", err)
		}
		b.cl.matchesLens, err = cl.CreateBufferSlice(context, cl.MEM_READ_WRITE, b.data.matchesLens)
		if err != nil {
			return nil, fmt.Errorf("creating matches length buffer: %w", err)
		}
	}

	return b, nil
}

func (b *clBuffers) Delete() {
	cl.ReleaseMemObject(b.cl.tries)
	cl.ReleaseMemObject(b.cl.idxs)
	cl.ReleaseMemObject(b.cl.strs)
	cl.ReleaseMemObject(b.cl.strsOffsets)
	cl.ReleaseMemObject(b.cl.strLens)
	cl.ReleaseMemObject(b.cl.hashBitmap)
	cl.ReleaseMemObject(b.cl.targetHashes)
	cl.ReleaseMemObject(b.cl.matches)
	cl.ReleaseMemObject(b.cl.matchesLens)
}

func (b *clBuffers) strArrOffset(s pattern.Segment, unionIdx int) int {
	i := unionIdx
	if s.Comps[i] != len(s.Segs[i]) {
		panic("expected strArrOffset to be called on union of only strings")
	}
	arr := make([]string, len(s.Segs[i]))
	for j := range s.Segs[i] {
		arr[j] = s.Segs[i][j].Str
	}
	strArrIdx, found := slices.BinarySearchFunc(b.strArrs, arr, slices.Compare)
	if !found {
		panic("expected to find string array")
	}
	return b.strArrsFirstIdxs[strArrIdx]
}

func (b *clBuffers) read(queue cl.CommandQueue) error {
	if err := cl.EnqueueReadBufferSlice(queue, b.cl.tries, true, 0, b.data.tries, nil, nil); err != nil {
		return err
	}
	if err := cl.EnqueueReadBufferSlice(queue, b.cl.idxs, true, 0, b.data.idxs, nil, nil); err != nil {
		return err
	}
	if err := cl.EnqueueReadBufferSlice(queue, b.cl.matches, true, 0, b.data.matches, nil, nil); err != nil {
		return err
	}
	if err := cl.EnqueueReadBufferSlice(queue, b.cl.matchesLens, true, 0, b.data.matchesLens, nil, nil); err != nil {
		return err
	}
	return nil
}

func (b *clBuffers) write(queue cl.CommandQueue) error {
	if err := cl.EnqueueWriteBufferSlice(queue, b.cl.tries, true, 0, b.data.tries, nil, nil); err != nil {
		return err
	}
	if err := cl.EnqueueWriteBufferSlice(queue, b.cl.idxs, true, 0, b.data.idxs, nil, nil); err != nil {
		return err
	}
	if err := cl.EnqueueReadBufferSlice(queue, b.cl.matchesLens, true, 0, b.data.matchesLens, nil, nil); err != nil {
		return err
	}
	return nil
}

func (b *clBuffers) drainMatches() (matches []string) {
	var s strings.Builder
	for i := range b.numWorkers {
		offs := i * b.matchBufLen
		for j := range int(b.data.matchesLens[i]) {
			c := b.data.matches[offs+j]
			if c != 0 {
				s.WriteByte(c)
			} else {
				matches = append(matches, strings.Clone(s.String()))
				s.Reset()
			}
		}
	}
	for i := range b.data.matchesLens {
		b.data.matchesLens[i] = 0
	}
	return
}

func generateClCode(s pattern.Segment, bufs *clBuffers) (code []byte) {
	// Quotes an OpenCL string (doesn't cover
	// all escape sequences).
	quote := func(s string) string {
		var b strings.Builder
		b.WriteByte('"')
		for _, r := range s {
			switch r {
			case '"':
				b.WriteString(`\"`)
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\\':
				b.WriteString(`\\`)
			default:
				if r >= 0x20 && r <= 0x7e {
					b.WriteByte(byte(r))
				} else {
					var p [4]byte
					for _, c := range p[:utf8.EncodeRune(p[:], r)] {
						fmt.Fprintf(&b, `\x%02x`, c)
					}
				}
			}
		}
		b.WriteByte('"')
		return b.String()
	}

	cb := &codeBuilder{IndentStr: "  "}
	cb.L("#define MAX_CANDIDATE_LEN %d", bufs.maxCandidateLen)
	cb.L("#define MAX_MATCH_BUF_LEN %d", bufs.matchBufLen)
	cb.L("#define NUM_TARGET_HASHES %d", len(bufs.data.targetHashes))
	cb.L("#define IDX_LEN %d", bufs.idxLen)
	cb.L("%s", preludeClCode)

	writeExprPushStr := func(cb *codeBuilder, memcpyFn, str string) {
		cb.L("%s((char*)candidate + candidate_len, %s, str_len);", memcpyFn, str)
		cb.L("candidate_len += str_len;")
	}
	writeExprPopStr := func(cb *codeBuilder) {
		cb.L("candidate_len -= str_len;")
	}
	var guessFns []*codeBuilder
	pushGuessFn := func() (cb *codeBuilder, name string) {
		fnIdx := len(guessFns)
		name = fmt.Sprintf("guess_%d", fnIdx)
		cb = &codeBuilder{IndentStr: "  "}
		cb.L("bool %s(size_t id, u32 *n, u32 *i, u64 *candidate, i32 candidate_len, __global const char *strs, __global const u32 *strs_offsets, __global const u32 *str_lens, __global const u32 *hash_bitmap, __global const u64 *target_hashes, __global char *matches, __global u32 *matches_lens) {", name)
		guessFns = append(guessFns, cb)
		return
	}
	popGuessFn := func(cb *codeBuilder) {
		cb.L("return true;")
		cb.L("}")
	}
	writeExprCallGuessFn := func(cb *codeBuilder, name string) {
		if name == "try" {
			cb.L("if (!(*n) || !try(candidate, candidate_len, id, hash_bitmap, target_hashes, matches, matches_lens)) return false;")
			cb.L("(*n)--;")
		} else {
			cb.L("if (!%s(id, n, i, candidate, candidate_len, strs, strs_offsets, str_lens, hash_bitmap, target_hashes, matches, matches_lens)) return false;", name)
		}
	}

	//fmt.Println("strs:", bufs.strArrs)

	cIdxCounter := 0
	// guessFn is the function to call after the current segment has reached
	// its last cartesian product term. For the root node, it's "try", since we
	// always want to try the string produced on after the rightmost cartesian
	// product.
	//
	// For any union, each union operand should have the same guessFn. To avoid
	// repetition, we generate the next cartesian product term outside the union
	// as a new guessing function and then make the union guess via that.
	var genCode func(cb *codeBuilder, s pattern.Segment, idx int, guessFn string)
	genCode = func(cb *codeBuilder, s pattern.Segment, idx int, guessFn string) {
		if s.Type == pattern.SegmentText {
			cb.L("const u32 str_len = %d;", len(s.Str))
			writeExprPushStr(cb, "memcpy_pg", quote(s.Str))
			writeExprCallGuessFn(cb, guessFn)
			writeExprPopStr(cb)
			return
		}

		segs := s.Segs[idx]

		cIdx := cIdxCounter
		cIdxCounter++

		if len(segs) == s.Comps[idx] { // List of single strings
			if len(segs) != 1 {
				cb.L("for (; i[%d] < %d; i[%d]++) {", cIdx, len(segs), cIdx)
			} else {
				cb.L("{")
			}
			offs := bufs.strArrOffset(s, idx)
			cb.L("const u32 str_idx = %d+i[%d];", offs, cIdx)
			cb.L("const u32 str_len = str_lens[str_idx];", offs, cIdx)
			writeExprPushStr(cb, "memcpy_pg", "strs + strs_offsets[str_idx]")
			if idx == len(s.Segs)-1 {
				writeExprCallGuessFn(cb, guessFn)
			} else {
				genCode(cb, s, idx+1, guessFn)
			}
			writeExprPopStr(cb)
			cb.L("}")
		} else { // Some union operand with more nesting exists
			var fname string
			if idx == len(s.Segs)-1 {
				fname = guessFn
			} else {
				fcb, name := pushGuessFn()
				genCode(fcb, s, idx+1, guessFn)
				popGuessFn(fcb)
				fname = name
			}
			cb.L("switch (i[%d]) {", cIdx)
			for j, seg := range segs {
				cb.L("case %d:", j)
				genCode(cb, seg, 0, fname)
				if j != len(segs)-1 {
 					cb.L("i[%d]++;", cIdx)
					cb.L("//fallthrough")
				}
			}
			cb.L("}")
		}
		if len(segs) != 1 {
			cb.L("i[%d] = 0;", cIdx)
		}
	}

	var guessEntryFnName string
	{
		var fcb *codeBuilder
		fcb, guessEntryFnName = pushGuessFn()
		genCode(fcb, s, 0, "try")
		popGuessFn(fcb)
	}

	for _, fcb := range slices.Backward(guessFns) {
		cb.L("%s", fcb.B.String())
	}

	cb.L(`__kernel void kmain(__global u32* tries, __global u32* idxs, __global const char *strs, __global const u32 *strs_offsets, __global const u32 *str_lens, __global const u32 *hash_bitmap, __global const u64 *target_hashes, __global char *matches, __global u32 *matches_lens) {`)
	cb.L("size_t id = get_global_id(0);")
	cb.L("u32 n = tries[id]; // number of tries left")
	cb.L("u64 candidate[(MAX_CANDIDATE_LEN+7)/8]; // using u64 to force correct alignment to be able to speed up murmurhash algorithm")
	cb.L("i32 candidate_len = 0;")
	cb.L("u32 i[IDX_LEN];")
	cb.L("memcpy_pg(i, idxs + id*IDX_LEN, sizeof(i));")
	cb.L("%s(id, &n, i, candidate, candidate_len, strs, strs_offsets, str_lens, hash_bitmap, target_hashes, matches, matches_lens);", guessEntryFnName)
	cb.L("memcpy_gp(idxs + id*IDX_LEN, i, sizeof(i));")
	cb.L("tries[id] = n;")
	cb.L("}")
	return cb.B.Bytes()
}
