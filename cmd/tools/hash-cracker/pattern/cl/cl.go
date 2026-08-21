package cl

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	cl "github.com/CyberChainXyz/go-opencl"
	"github.com/xypwn/filediver/cmd/tools/hash-cracker/pattern"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
)

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

func makeClBuffers(runner *cl.OpenCLRunner, s pattern.Segment, targetHashes []stingray.Hash, numWorkers, matchBufLen int) (*clBuffers, error) {
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
		b.cl.tries, err = runner.CreateEmptyBuffer(cl.READ_WRITE, binary.Size(b.data.tries))
		if err != nil {
			return nil, fmt.Errorf("creating tries buffer: %w", err)
		}
		b.cl.idxs, err = runner.CreateEmptyBuffer(cl.READ_WRITE, binary.Size(b.data.idxs))
		if err != nil {
			return nil, fmt.Errorf("creating indices buffer: %w", err)
		}
		b.cl.strs, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.strs)
		if err != nil {
			return nil, fmt.Errorf("creating strings buffer: %w", err)
		}
		b.cl.strsOffsets, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.strsOffsets)
		if err != nil {
			return nil, fmt.Errorf("creating strings offsets buffer: %w", err)
		}
		b.cl.strLens, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.strLens)
		if err != nil {
			return nil, fmt.Errorf("creating string lengths buffer: %w", err)
		}
		b.cl.hashBitmap, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.hashBitmap)
		if err != nil {
			return nil, fmt.Errorf("creating hash bitmap buffer: %w", err)
		}
		b.cl.targetHashes, err = cl.CreateBuffer(runner, cl.READ_ONLY|cl.COPY_HOST_PTR, b.data.targetHashes)
		if err != nil {
			return nil, fmt.Errorf("creating target hash buffer: %w", err)
		}
		b.cl.matches, err = runner.CreateEmptyBuffer(cl.WRITE_ONLY, binary.Size(b.data.matches))
		if err != nil {
			return nil, fmt.Errorf("creating match buffer: %w", err)
		}
		b.cl.matchesLens, err = runner.CreateEmptyBuffer(cl.READ_WRITE, binary.Size(b.data.matchesLens))
		if err != nil {
			return nil, fmt.Errorf("creating matches length buffer: %w", err)
		}
	}

	return b, nil
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
	if err := cl.WriteBuffer(runner, 0, b.cl.matchesLens, b.data.matchesLens, true); err != nil {
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

	var cb codeBuilder
	cb.IndentStr = "  "
	cb.L("#define MAX_CANDIDATE_LEN %d", bufs.maxCandidateLen)
	cb.L("#define MAX_MATCH_BUF_LEN %d", bufs.matchBufLen)
	cb.L("#define NUM_TARGET_HASHES %d", len(bufs.data.targetHashes))
	cb.L("#define IDX_LEN %d", bufs.idxLen)
	cb.L(`
typedef int i32;
typedef unsigned int u32;
typedef long i64;
typedef unsigned long u64;

// Calculates the murmurhash 64a of the given
// string (as u64 array). n is the number of bytes.
//
// We're using u64s so we are guaranteed to have the correct
// alignment when casting to u64 in the main loop. For unaligned
// data, the only alternative would be bitwise-oring the bytes
// together, which is quite a lot slower (~30%% in my testing).
u64 murmur64a_sum(const u64 *d, u32 n) {
#define SEED 0ul
#define MIX 0xc6a4a7935bd1e995ul
#define SHIFTS 47ul

  u64 hash = SEED ^ ((u64)n * MIX);
  
  while (n >= 8) {
    u64 key = *d;
    d += 1;
    n -= 8;

    key *= MIX;
    key ^= key >> SHIFTS;
    key *= MIX;

    hash ^= key;
    hash *= MIX;
  }

  if (n&7) {
    switch (n&7) {
    case 7: hash ^= *d & 0x00fffffffffffffful; break;
    case 6: hash ^= *d & 0x0000fffffffffffful; break;
    case 5: hash ^= *d & 0x000000fffffffffful; break;
    case 4: hash ^= *d & 0x00000000fffffffful; break;
    case 3: hash ^= *d & 0x0000000000fffffful; break;
    case 2: hash ^= *d & 0x000000000000fffful; break;
    case 1: hash ^= *d & 0x00000000000000fful; break;
    }
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

void* memcpy_pc(void* dest, __constant void* src, size_t n) {
  char *d = dest;
  __constant char *s = src;
  while (n--) *d++ = *s++;
  return dest;
}
void* memcpy_pg(void* dest, __global const void* src, size_t n) {
  char *d = dest;
  __global const char *s = src;
  while (n--) *d++ = *s++;
  return dest;
}
__global void* memcpy_gp(__global void* dest, const void* src, size_t n) {
  __global char *d = dest;
  const char *s = src;
  while (n--) *d++ = *s++;
  return dest;
}

// Bloom filter to rule out most candidates.
bool bitmap_test(u64 hash, __global const u32 *hash_bitmap) {
#define BITS 24
  u32 hl = (u32)hash;
  u32 hh = (u32)(hash >> 32);
  u32 idxl = hl >> (32-(BITS-4));
  u32 bitl = 1 << (hl & 31);
  if (!(hash_bitmap[idxl] & bitl)) return false;
  u32 idxh = hh >> (32-(BITS-4));
  u32 bith = 1 << (hh & 31);
  if (!(hash_bitmap[idxh] & bith)) return false;
  return true;
#undef BITS
}

// More expensive check if a candidate actually matches
// a target hash.
bool binary_search(u64 hash, __global const u64 *target_hashes) {
  u32 lo = 0;
  u32 hi = NUM_TARGET_HASHES;
  while (lo < hi) {
    u32 mid = (lo + hi) >> 1;
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
bool try(const u64 *s, u32 n, const size_t id, __global const u32 *hash_bitmap, __global const u64 *target_hashes, __global char *matches, __global u32 *matches_lens) {
  u64 h = murmur64a_sum(s, n);
  if (!bitmap_test(h, hash_bitmap)) return true;
  if (!binary_search(h, target_hashes)) return true;
  if (matches_lens[id]+n+1 > MAX_MATCH_BUF_LEN)
    return false;
  memcpy_gp(matches + id*MAX_MATCH_BUF_LEN + matches_lens[id], s, n);
  matches_lens[id] += n;
  matches[matches_lens[id]++] = 0;
  return true;
}
`)
	cb.L(`__kernel void kmain(__global u32* tries, __global u32* idxs, __global const char *strs, __global const u32 *strs_offsets, __global const u32 *str_lens, __global const u32 *hash_bitmap, __global const u64 *target_hashes, __global char *matches, __global u32 *matches_lens) {`)
	cb.L("size_t id = get_global_id(0);")
	cb.L("u32 n = tries[id]; // number of tries left")
	cb.L("u64 candidate[(MAX_CANDIDATE_LEN+7)/8]; // using u64 to force correct alignment to be able to speed up murmurhash algorithm")
	cb.L("i32 candidate_len = 0;")
	cb.L("u32 i[IDX_LEN];")
	cb.L("memcpy_pg(i, idxs + id*IDX_LEN, sizeof(i));")
	totalIdxNum := 0
	var genCode func(s pattern.Segment, canTry bool)
	genCode = func(s pattern.Segment, canTry bool) {
		writeExprPushStr := func(memcpyFn, str string) {
			cb.L("%s((char*)candidate + candidate_len, %s, str_len);", memcpyFn, str)
			cb.L("candidate_len += str_len;")
		}
		const exprTry = "if (!n || !try(candidate, candidate_len, id, hash_bitmap, target_hashes, matches, matches_lens)) goto ret; n--;"
		if s.Type == pattern.SegmentText { // fallback; this case shouldn't happen
			cb.L("const u32 str_len = %d;", len(s.Str))
			writeExprPushStr("memcpy_pc", quote(s.Str))
			if canTry {
				cb.L(exprTry)
			}
			cb.L("candidate_len -= str_len;")
			return
		}
		idxNumOffs := totalIdxNum
		totalIdxNum += len(s.Segs)
		for i, segs := range s.Segs {
			idxNum := idxNumOffs + i
			cb.L("for (; i[%d] < %d; i[%d]++) {", idxNum, len(segs), idxNum)
			shouldTry := canTry && i == len(s.Segs)-1
			if len(segs) == s.Comps[i] {
				offs := bufs.strArrOffset(s, i)
				cb.L("const u32 str_idx = %d+i[%d];", offs, idxNum)
				cb.L("const u32 str_len = str_lens[str_idx];", offs, idxNum)
				writeExprPushStr("memcpy_pg", "strs + strs_offsets[str_idx]")
				if shouldTry {
					cb.L(exprTry)
				}
			} else {
				cb.L("u32 str_len = 0;")
				cb.L("switch (i[%d]) {", idxNum)
				for j, seg := range segs {
					cb.L("case %d:", j)
					if seg.Type == pattern.SegmentText {
						cb.L("str_len = %d;", len(seg.Str))
						writeExprPushStr("memcpy_pc", quote(seg.Str))
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
		for i := range slices.Backward(s.Segs) {
			idxNum := idxNumOffs + i
			cb.L("candidate_len -= str_len;")
			cb.L("}")
			cb.L("i[%d] = 0;", idxNum)
		}
	}
	genCode(s, true)
	cb.L("ret:")
	cb.L("memcpy_gp(idxs + id*IDX_LEN, i, sizeof(i));")
	cb.L("tries[id] = n;")
	cb.L("return;")
	cb.L("}")
	return cb.B.Bytes()
}
