package pattern

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	cl "github.com/CyberChainXyz/go-opencl"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

type codeBuilder struct {
	buf       bytes.Buffer
	B         bytes.Buffer
	IndentStr string
	Ind       int
}

func (cb *codeBuilder) line(format string, text []byte) {
	extraUnind := 0
	if strings.HasPrefix(format, "case ") || strings.HasSuffix(format, ":") {
		extraUnind++
	}
	if format != "" {
		for range cb.Ind - extraUnind {
			cb.B.WriteString(cb.IndentStr)
		}
	}
	cb.B.Write(text)
	cb.B.WriteByte('\n')
}

// L writes one or multiple formatted lines of code.
//
// Automatically indents for every "{" and unindents for
// every "}".
//
// "case" expressions and labels are indented 1 less without changing current
// indentation.
func (cb *codeBuilder) L(format string, args ...any) {
	ind := strings.Count(format, "{") - strings.Count(format, "}")
	if ind < 0 {
		cb.Ind += ind
	}
	argi := 0
	for ln := range strings.SplitSeq(format, "\n") {
		ln = strings.TrimSuffix(ln, "\r")
		argn := 0
		for i := 0; i < len(ln)-1; i++ {
			// NOTE: Non-ASCII not handled.
			if ln[i] == '%' && ln[i+1] != '%' {
				argn++
				i++
			}
		}
		fmt.Fprintf(&cb.buf, ln, args[argi:argi+argn]...)
		cb.line(ln, cb.buf.Bytes())
		cb.buf.Reset()
		argi += argn
	}
	if ind > 0 {
		cb.Ind += ind
	}
}

// Returns the number of elements needed in the index
// for the given program.
func getSegmentIdxLen(s segment) int {
	n := len(s.segs)
	for _, segs := range s.segs {
		for _, seg := range segs {
			n += getSegmentIdxLen(seg)
		}
	}
	return n
}

func readIdx(dest []int64, s segment, idx segIdx) int {
	p := 0
	for i := range len(idx.idxs) {
		dest[p] = int64(idx.idxs[i])
		p++
	}
	for i, segs := range s.segs {
		for j, seg := range segs {
			p += readIdx(dest[p:], seg, idx.segs[i][j])
		}
	}
	return p
}

type clBuffers struct {
	strArrs          [][]string
	strArrsFirstIdxs []int
	numWorkers       int
	idxLen           int
	maxMatchBufLen   int
	maxCandidateLen  int
	data             struct {
		tries        []uint32
		idxs         []int64
		strs         []byte
		strsOffsets  []uint32
		strLens      []uint32
		hashBitmap   []uint32
		targetHashes []uint64
		matches      []byte
		matchesLens  []uint32
	}
	cl struct {
		tries        *cl.Buffer
		idxs         *cl.Buffer
		strs         *cl.Buffer
		strsOffsets  *cl.Buffer
		strLens      *cl.Buffer
		hashBitmap   *cl.Buffer
		targetHashes *cl.Buffer
		matches      *cl.Buffer
		matchesLens  *cl.Buffer
	}
}

func makeClBuffers(runner *cl.OpenCLRunner, s segment, targetHashes []stingray.Hash, numWorkers, maxMatchBufLen int) (*clBuffers, error) {
	b := &clBuffers{
		numWorkers:      numWorkers,
		idxLen:          getSegmentIdxLen(s),
		maxMatchBufLen:  maxMatchBufLen,
		maxCandidateLen: s.MaxLen(),
	}

	if b.maxMatchBufLen < b.maxCandidateLen+1 {
		return nil, fmt.Errorf("maxMatchBufLen (%d) must be at least 1 more than the maximum candidate string length (%d)", b.maxMatchBufLen, b.maxCandidateLen)
	}

	// Create tries buffer.
	b.data.tries = make([]uint32, numWorkers)

	// Create index buffer.
	b.data.idxs = make([]int64, b.idxLen*numWorkers)

	// Collect string arrays for all cases
	// of unions of single string operands.
	{
		var addStrArrs func(s segment)
		addStrArrs = func(s segment) {
			if s.str != "" {
				return
			}
			for i, segs := range s.segs {
				if len(segs) == s.comps[i] {
					strs := make([]string, len(segs))
					for j, seg := range segs {
						strs[j] = seg.str
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
	b.data.matches = make([]byte, b.maxMatchBufLen*numWorkers)
	b.data.matchesLens = make([]uint32, numWorkers)

	// Create according OpenCL Buffers
	{
		var err error
		b.cl.tries, err = runner.CreateEmptyBuffer(cl.READ_WRITE, binary.Size(b.data.tries))
		if err != nil {
			return nil, err
		}
		b.cl.idxs, err = runner.CreateEmptyBuffer(cl.READ_WRITE, binary.Size(b.data.idxs))
		if err != nil {
			return nil, err
		}
		b.cl.strs, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.strs)
		if err != nil {
			return nil, err
		}
		b.cl.strsOffsets, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.strsOffsets)
		if err != nil {
			return nil, err
		}
		b.cl.strLens, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.strLens)
		if err != nil {
			return nil, err
		}
		b.cl.hashBitmap, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.hashBitmap)
		if err != nil {
			return nil, err
		}
		b.cl.targetHashes, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.targetHashes)
		if err != nil {
			return nil, err
		}
		b.cl.matches, err = runner.CreateEmptyBuffer(cl.WRITE_ONLY, binary.Size(b.data.matches))
		if err != nil {
			return nil, err
		}
		b.cl.matchesLens, err = runner.CreateEmptyBuffer(cl.READ_WRITE, binary.Size(b.data.matchesLens))
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (b *clBuffers) strArrOffset(s segment, unionIdx int) int {
	i := unionIdx
	if s.comps[i] != len(s.segs[i]) {
		panic("expected strArrOffset to be called on union of only strings")
	}
	arr := make([]string, len(s.segs[i]))
	for j := range s.segs[i] {
		arr[j] = s.segs[i][j].str
	}
	strArrIdx, found := slices.BinarySearchFunc(b.strArrs, arr, slices.Compare)
	if !found {
		panic("expected to find string array")
	}
	return b.strArrsFirstIdxs[strArrIdx]
}

func (b *clBuffers) read(runner *cl.OpenCLRunner) error {
	if err := cl.ReadBuffer(runner, 0, b.cl.tries, b.data.tries); err != nil {
		return err
	}
	if err := cl.ReadBuffer(runner, 0, b.cl.idxs, b.data.idxs); err != nil {
		return err
	}
	if err := cl.ReadBuffer(runner, 0, b.cl.matches, b.data.matches); err != nil {
		return err
	}
	if err := cl.ReadBuffer(runner, 0, b.cl.matchesLens, b.data.matchesLens); err != nil {
		return err
	}
	return nil
}

func (b *clBuffers) write(runner *cl.OpenCLRunner) error {
	if err := cl.WriteBuffer(runner, 0, b.cl.tries, b.data.tries, true); err != nil {
		return err
	}
	if err := cl.WriteBuffer(runner, 0, b.cl.idxs, b.data.idxs, true); err != nil {
		return err
	}
	if err := cl.WriteBuffer(runner, 0, b.cl.matches, b.data.matches, true); err != nil {
		return err
	}
	if err := cl.WriteBuffer(runner, 0, b.cl.matchesLens, b.data.matchesLens, true); err != nil {
		return err
	}
	return nil
}

func (b *clBuffers) drainMatches() (matches []string) {
	var s strings.Builder
	for i := range b.numWorkers {
		offs := i * b.maxMatchBufLen
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

func generateClCode(s segment, bufs *clBuffers) (code []byte) {
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

	var cb codeBuilder
	cb.IndentStr = "  "
	cb.L("#define MAX_CANDIDATE_LEN %d", bufs.maxCandidateLen)
	cb.L("#define MAX_MATCH_BUF_LEN %d", bufs.maxMatchBufLen)
	cb.L("#define NUM_TARGET_HASHES %d", len(bufs.data.targetHashes))
	cb.L("#define IDX_LEN %d", bufs.idxLen)
	cb.L(`
ulong murmur64a_sum(const char *d, size_t n) {
#define SEED 0
#define MIX 0xc6a4a7935bd1e995ul
#define SHIFTS 47

  ulong hash = SEED ^ (n * MIX);
  
  while (n >= 8) {
    ulong key = *((ulong*)d);
	d += 8;
	n -= 8;

	key *= MIX;
	key ^= key >> SHIFTS;
	key *= MIX;

	hash ^= key;
	hash *= MIX;
  }

  switch (n&7) {
  case 7: hash ^= (ulong)d[6] << (8*6);
  case 6: hash ^= (ulong)d[5] << (8*5);
  case 5: hash ^= (ulong)d[4] << (8*4);
  case 4: hash ^= (ulong)d[3] << (8*3);
  case 3: hash ^= (ulong)d[2] << (8*2);
  case 2: hash ^= (ulong)d[1] << (8*1);
  case 1: hash ^= (ulong)d[0] << (8*0);
    hash *= MIX;
  }

  hash ^= hash >> SHIFTS;

  hash *= MIX;
  hash ^= hash >> SHIFTS;

  return hash;

#undef SEED
#undef MIX
#undef SHIFTS
}

void* memcpy_lc(void* dest, __constant void* src, size_t n) {
  char *d = dest;
  __constant char *s = src;
  while (n--) *d++ = *s++;
  return dest;
}
void* memcpy_lg(void* dest, __global const void* src, size_t n) {
  char *d = dest;
  __global const char *s = src;
  while (n--) *d++ = *s++;
  return dest;
}
__global void* memcpy_gl(__global void* dest, const void* src, size_t n) {
  __global char *d = dest;
  const char *s = src;
  while (n--) *d++ = *s++;
  return dest;
}

// Bloom filter to rule out most candidates.
bool bitmap_test(ulong hash, __global const uint *hash_bitmap) {
#define BITS 24
  uint hl = (uint)hash;
  uint hh = (uint)(hash >> 32);
  uint idxl = hl >> (32-(BITS-4));
  uint bitl = 1 << (hl & 31);
  if (!(hash_bitmap[idxl] & bitl)) return false;
  uint idxh = hh >> (32-(BITS-4));
  uint bith = 1 << (hh & 31);
  if (!(hash_bitmap[idxh] & bith)) return false;
  return true;
#undef BITS
}

// More expensive check if a candidate actually matches
// a target hash.
bool binary_search(ulong hash, __global const ulong *target_hashes) {
  uint lo = 0;
  uint hi = NUM_TARGET_HASHES;
  while (lo < hi) {
    uint mid = (lo + hi) >> 1;
	if (target_hashes[mid] < hash)
	  lo = mid + 1;
	else
	  hi = mid;
  }
  return lo < NUM_TARGET_HASHES && target_hashes[lo] == hash;
}

// Tries the string as a match for a target hash.
// If it is a match, it is written to the match buffer.
//
// Returns false if the try filled up the match buffer
// fully and the kernel function should exit.
bool try(const char *s, size_t n, const size_t id, __global const uint *hash_bitmap, __global const ulong *target_hashes, __global char *matches, __global uint *matches_lens) {
  ulong h = murmur64a_sum(s, n);
  if (!bitmap_test(h, hash_bitmap)) return true;
  if (!binary_search(h, target_hashes)) return true;
  if (matches_lens[id]+n+1 > MAX_MATCH_BUF_LEN)
    return false;
  memcpy_gl(matches + id*MAX_MATCH_BUF_LEN + matches_lens[id], s, n);
  matches_lens[id] += n;
  matches[matches_lens[id]++] = 0;
  return true;
}
`)
	cb.L(`__kernel void kmain(__global uint* tries, __global long* idxs, __global const char *strs, __global const uint *strs_offsets, __global const uint *str_lens, __global const uint *hash_bitmap, __global const ulong *target_hashes, __global char *matches, __global uint *matches_lens) {`)
	cb.L("size_t id = get_global_id(0);")
	cb.L("uint n = tries[id]; // number of tries left")
	cb.L("char candidate[MAX_CANDIDATE_LEN];")
	cb.L("int candidate_len = 0;")
	cb.L("long i[IDX_LEN];")
	cb.L("memcpy_lg(i, idxs + id*IDX_LEN, sizeof(i));")
	totalIdxNum := 0
	var genCode func(s segment, canTry bool)
	genCode = func(s segment, canTry bool) {
		writeExprPushStr := func(memcpyFn, str string) {
			cb.L("%s(candidate+candidate_len, %s, str_len);", memcpyFn, str)
			cb.L("candidate_len += str_len;")
		}
		const exprTry = "if (!n || !try(candidate, candidate_len, id, hash_bitmap, target_hashes, matches, matches_lens)) goto ret; n--;"
		if s.str != "" { // fallback; this case shouldn't happen
			cb.L("int str_len = %d;", len(s.str))
			writeExprPushStr("memcpy_lc", quote(s.str))
			if canTry {
				cb.L(exprTry)
			}
			cb.L("candidate_len -= str_len;")
			return
		}
		idxNumOffs := totalIdxNum
		totalIdxNum += len(s.segs)
		for i, segs := range s.segs {
			idxNum := idxNumOffs + i
			cb.L("for (; i[%d] < %d; i[%d]++) {", idxNum, len(segs), idxNum)
			shouldTry := canTry && i == len(s.segs)-1
			if len(segs) == s.comps[i] {
				offs := bufs.strArrOffset(s, i)
				cb.L("const uint str_idx = %d+i[%d];", offs, idxNum)
				cb.L("const uint str_len = str_lens[str_idx];", offs, idxNum)
				writeExprPushStr("memcpy_lg", "strs + strs_offsets[str_idx]")
				if shouldTry {
					cb.L(exprTry)
				}
			} else {
				cb.L("uint str_len = 0;")
				cb.L("switch (i[%d]) {", idxNum)
				for j, seg := range segs {
					cb.L("case %d:", j)
					if seg.str != "" {
						cb.L("str_len = %d;", len(s.str))
						writeExprPushStr("memcpy_lc", quote(s.str))
						if shouldTry {
							cb.L(exprTry)
						}
					} else {
						genCode(seg, shouldTry)
					}
					cb.L("break;")
				}
				cb.L("}")
			}
		}
		for i := range slices.Backward(s.segs) {
			idxNum := idxNumOffs + i
			cb.L("candidate_len -= str_len;")
			cb.L("}")
			cb.L("i[%d] = 0;", idxNum)
		}
	}
	genCode(s, true)
	cb.L("ret:")
	cb.L("memcpy_gl(idxs + id*IDX_LEN, i, sizeof(i));")
	cb.L("tries[id] = n;")
	cb.L("return;")
	cb.L("}")
	return cb.B.Bytes()
}
