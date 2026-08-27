package pattern

import (
	"errors"
	"fmt"
	"iter"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
)

type CompileOptions struct {
	// Additional functions.
	//
	// Parameters must be of type string or any [IrSegment] type.
	//
	// Return value must be an IrSegment and optionally an error.
	Funcs map[string]any
	//
	Vars map[string]IrSegment
	// Don't apply optimizations (mostly meant for debugging,
	// you usually want to leave this set to false).
	NoOptimize bool
}

var (
	ErrNotAscii                       = errors.New("pattern must be ASCII-only")
	ErrNullCharacter                  = errors.New("pattern must not contain the null ('\\0') character")
	ErrNumberTooLarge                 = errors.New("number too large")
	ErrExpectedNonnegativeInteger     = errors.New("expected nonnegative whole number")
	ErrExpectedExpressionBeforeRepeat = errors.New("expected expression before repeat expression")
	ErrInvalidCharClassRange          = errors.New("invalid character class range")
	ErrInvalidVarOrFuncNameChar       = errors.New("invalid variable or function name char (expected [A-Za-z0-9_-])")
	ErrEmptyVarOrFuncName             = errors.New("empty variable or function name")
	ErrTooManyArgsInAssignment        = errors.New("assignment needs exactly one parameter after \"=\"")
	ErrVarOrFuncAlreadyExists         = errors.New("variable or function already exists")
	ErrUnknownVarOrFunc               = errors.New("unknown variable or function")
	ErrInvalidVarType                 = errors.New("invalid variable type")
	ErrInvalidFuncType                = errors.New("invalid function type")
	ErrInvalidFuncCall                = errors.New("invalid function call")
	ErrCallingFunc                    = errors.New("calling function")
	ErrComplexityTooLarge             = errors.New("complexity (number of possibilities) must be less than the 64-bit signed integer limit (9223372036854775807)")
)

type Error struct {
	Position int
	Err      error
}

func (e *Error) Error() string {
	return fmt.Sprintf("pattern error at position %d: %v", e.Position+1, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

type SegmentType uint8

const (
	SegmentText SegmentType = iota
	SegmentProdOfSets
)

type Segment struct {
	// Text or cartesian product of sets
	Type SegmentType
	// Text (if Type == SegmentText)
	Str string
	// cartesian product of ordered sets (if Type == SegmentProdOfSets)
	//
	// (note that this is in a way the opposite of how
	// parse models things, as parse has an ordered set
	// of cartesian products; the reason we're using
	// this model is for the compiled segments is better
	// data locality)
	Segs [][]Segment
	// Number of elements (in this segment and in
	// each cartesian product operand).
	// math.MaxInt if >= math.MaxInt
	Comp  int
	Comps []int
}

type SegIdx struct {
	Segs     [][]SegIdx
	Idxs     []int
	TotalIdx int // total index in [0,comp-1]
}

// Reset zeroes out the index.
//
// Assumes [SegIdx.TotalIdx] is consistent
// with any inner indices for performance.
func (idx *SegIdx) Reset() {
	if idx.TotalIdx == 0 {
		return
	}
	for i := range idx.Idxs {
		idx.Idxs[i] = 0
		for j := range idx.Segs[i] {
			idx.Segs[i][j].Reset()
		}
	}
	idx.TotalIdx = 0
}

func (idx SegIdx) String() string {
	if idx.Segs == nil && idx.Idxs == nil {
		return "nil"
	}
	var res strings.Builder
	fmt.Fprintf(&res, "segs=%v,", idx.Segs)
	fmt.Fprintf(&res, "idxs=%v", idx.Idxs)
	return res.String()
}

// Add efficiently advances the index by n for the segment.
//
// n must be positive.
func (idx *SegIdx) Add(s Segment, n int) (carry int) {
	// The recursive algorithm is a bit like adding
	// n to the last digit of a number in decimal
	// addition.
	//
	// Instead, we add the number to the rightmost leaf
	// of the index. Each node then returns its addition
	// "carry" (how many times the addition made the set
	// "wrap around").
	//
	// For a cartesian product, the carry is then added
	// to the next place (1 further left) recursively.
	//
	// For a union, the carry is added to the currently
	// selected node. Once that node is "filled", we move
	// to the next node and add the remaining difference
	// and so on. The returned carry is again the number
	// of "wrap-arounds".
	// This is optimized by only adding the modulus with
	// the complexity of the entire union, so the actual
	// "wrap-around" can occur at most once.

	//fmt.Println(">add", n)
	//defer func() { fmt.Println("<add", n, carry) }()
	if n <= 0 {
		panic("expected offset n to be positive")
	}
	addToUnion := func(i int, n int) (carry int) {
		//fmt.Println(">addToUnion", i, n)
		//defer func() { fmt.Println("<addToUnion", i, n, carry) }()

		// Fast path: All union operands are single strings
		if s.Comps[i] == len(s.Segs[i]) {
			v := idx.Idxs[i] + n
			carry = v / s.Comps[i]
			idx.Idxs[i] = v % s.Comps[i]
			return
		}

		// Slow path: Some union operands are cartesian products
		carry = n / s.Comps[i]
		n %= s.Comps[i]
		j := idx.Idxs[i]
		if idx.Segs[i][j].TotalIdx >= s.Segs[i][j].Comp-n {
			// Fill first union operand fully if it would overflow
			n -= s.Segs[i][j].Comp - idx.Segs[i][j].TotalIdx
			idx.Segs[i][j].Reset() // zero the fully filled operand
			j++
			if j >= len(s.Segs[i]) {
				carry++
				j = 0
			}
			for n >= s.Segs[i][j].Comp {
				// Fill remaining union operands that would overflow
				// (we don't have to do anything to the index since
				// it's already zeroed and would remain that way)
				n -= s.Segs[i][j].Comp
				j++
				if j >= len(s.Segs[i]) {
					carry++
					j = 0
				}
			}
		}
		if n != 0 {
			// Fill the rest into the union operand that can't be
			// filled fully
			c := idx.Segs[i][j].Add(s.Segs[i][j], n)
			if c != 0 {
				panic(fmt.Sprintf("expected carry to be zero, but got %d", c))
			}
			n = 0
		}
		idx.Idxs[i] = j
		return
	}
	carry = n
	for i := range slices.Backward(s.Segs) { // cart. prod. terms
		if carry == 0 {
			break
		}
		carry = addToUnion(i, carry)
	}
	idx.TotalIdx = (idx.TotalIdx + n) % s.Comp
	return
}

// AppendAt appends the string at the given pattern index to
// the buffer.
func (s Segment) AppendAt(buf []byte, idx SegIdx) []byte {
	if s.Type == SegmentText {
		buf = append(buf, s.Str...)
	}
	for i := range s.Segs {
		j := idx.Idxs[i]
		buf = s.Segs[i][j].AppendAt(buf, idx.Segs[i][j])
	}
	return buf
}

// StringAt returns the string at the given pattern index.
func (s Segment) StringAt(idx SegIdx) string {
	return string(s.AppendAt(nil, idx))
}

func (s Segment) String() string {
	if s.Type == SegmentText {
		return strconv.Quote(s.Str)
	} else {
		var b strings.Builder
		for _, segs := range s.Segs {
			b.WriteString("<")
			for i, seg := range segs {
				if i != 0 {
					b.WriteString("|")
				}
				b.WriteString(seg.String())
			}
			b.WriteString(">")
		}
		return b.String()
	}
}

func (s *Segment) calculateComps() {
	if s.Type == SegmentText {
		s.Comp = 1
		return
	}
	if s.Comps != nil {
		return
	}
	s.Comps = make([]int, len(s.Segs))
	prod := 1
	for i := range s.Segs {
		sum := 0
		for j := range s.Segs[i] {
			s.Segs[i][j].calculateComps()
			seg := s.Segs[i][j]
			if seg.Comp == 0 {
				panic("segment complexity should be nonzero after updateComps()")
			}
			if sum > math.MaxInt-seg.Comp {
				sum = math.MaxInt // addition would overflow
			} else {
				sum += seg.Comp
			}
		}
		if prod > math.MaxInt/sum {
			prod = math.MaxInt // multiplication would overflow
		} else {
			prod *= sum
		}
		s.Comps[i] = sum
	}
	s.Comp = prod
}

// MakeIndex creates an index for the given segment.
func (s Segment) MakeIndex() (idx SegIdx) {
	if s.Type == SegmentText {
		return
	}
	idx.Idxs = make([]int, len(s.Segs))
	idx.Segs = make([][]SegIdx, len(s.Segs))
	for i, segs := range s.Segs {
		idx.Segs[i] = make([]SegIdx, len(segs))
		for j, seg := range segs {
			idx.Segs[i][j] = seg.MakeIndex()
		}
	}
	return
}

// CompBig calculates the complexity (number of elements)
// as a big int, or max if the number is bigger than max.
//
// Pass nil as max to ignore maximum.
func (s Segment) CompBig(max *big.Int) *big.Int {
	if s.Type == SegmentText {
		return big.NewInt(1)
	}
	prod := big.NewInt(1)
	sum := &big.Int{}
	for _, segs := range s.Segs {
		sum.SetInt64(0)
		for _, seg := range segs {
			c := seg.CompBig(max)
			if max != nil && c.Cmp(max) >= 0 {
				return max
			}
			sum.Add(sum, c)
			if max != nil && sum.Cmp(max) >= 0 {
				return max
			}
		}
		prod.Mul(prod, sum)
		if max != nil && prod.Cmp(max) >= 0 {
			return max
		}
	}
	return prod
}

// MaxLen returns the maximum possible string length
// generated from the pattern.
func (s Segment) MaxLen() int {
	if s.Type == SegmentText {
		return len(s.Str)
	}
	concatLen := 0
	for _, segs := range s.Segs {
		unionLen := 0
		for _, seg := range segs {
			unionLen = max(unionLen, seg.MaxLen())
		}
		concatLen += unionLen
	}
	return concatLen
}

func (s *Segment) optimize() {
	// Optimize inner segments first.
	for i := range s.Segs {
		for j := range s.Segs[i] {
			s.Segs[i][j].optimize()
		}
	}

	// Empty cartesian products or union elements can
	// be removed.
	if i := slices.IndexFunc(s.Segs, func(segs []Segment) bool {
		return len(segs) == 0
	}); i != -1 {
		var newSegs [][]Segment
		var newComps []int
		for ; i < len(s.Segs); i++ {
			if len(s.Segs[i]) != 0 {
				newSegs = append(newSegs, s.Segs[i])
				newComps = append(newComps, s.Comps[i])
			}
		}
		s.Segs = newSegs
		s.Comps = newComps
	}
	for i := range s.Segs {
		s.Segs[i] = slices.DeleteFunc(s.Segs[i], func(s Segment) bool {
			return s.Type == SegmentProdOfSets && len(s.Segs) == 0
		})
	}

	// Cartesian product of single-parameter unions
	// of cartesian product can be flattened.
	//
	// Example:
	// x(A, u(x(B, C)), u(x(D))) -> x(A, B, C, D)
	if i := slices.IndexFunc(s.Segs, func(segs []Segment) bool {
		return len(segs) == 1
	}); i != -1 {
		var newSegs [][]Segment
		var newComps []int
		for ; i < len(s.Segs); i++ {
			if len(s.Segs[i]) == 1 && s.Segs[i][0].Type == SegmentProdOfSets {
				newSegs = append(newSegs, s.Segs[i][0].Segs...)
				newComps = append(newComps, s.Segs[i][0].Comps...)
			} else {
				newSegs = append(newSegs, s.Segs[i])
				newComps = append(newComps, s.Comps[i])
			}
		}
		s.Segs = newSegs
		s.Comps = newComps
	}

	// Cartesian product with single parameter
	// of union with single parameter can be
	// flattened.
	//
	// Example:
	// x(u(A)) -> A
	if len(s.Segs) == 1 && len(s.Segs[0]) == 1 {
		*s = s.Segs[0][0]
	}
}

func compile(irSeg IrSegment, opts CompileOptions) Segment {
	if irSeg == nil {
		return Segment{}
	}
	var res Segment
	switch irSeg := irSeg.(type) {
	case IrSegmentChoice:
		if len(irSeg) == 0 {
			// Empty choice doesn't really make
			// sense mathematically, but it should
			// probably always be an empty string.
			res.Type = SegmentText
			res.Str = ""
			return res
		}
		res.Type = SegmentProdOfSets
		res.Segs = make([][]Segment, 1)
		res.Segs[0] = make([]Segment, len(irSeg))
		for i := range irSeg {
			res.Segs[0][i] = compile(irSeg[i], opts)
		}
	case IrSegmentConcat:
		res.Type = SegmentProdOfSets
		res.Segs = make([][]Segment, len(irSeg))
		for i := range irSeg {
			res.Segs[i] = []Segment{compile(irSeg[i], opts)}
		}
	case IrSegmentStr:
		res.Type = SegmentText
		res.Str = string(irSeg)
	case IrSegmentRepeat:
		// Expand repeat (e.g. a{1,2} -> a|aa)
		body := compile(irSeg.Seg, opts)
		if irSeg.Min == 1 && irSeg.Max == 1 {
			return body
		}
		var sep Segment
		if irSeg.Sep != nil {
			sep = compile(irSeg.Sep, opts)
		}
		res.Type = SegmentProdOfSets
		res.Segs = make([][]Segment, 1)
		res.Segs[0] = make([]Segment, irSeg.Max-irSeg.Min+1)
		for i := range res.Segs[0] {
			reps := irSeg.Min + i
			interspersed := 0
			if reps >= 2 && irSeg.Sep != nil {
				interspersed = reps - 1
			}
			repeated := Segment{
				Type: SegmentProdOfSets,
				Segs: make([][]Segment, 0, reps+interspersed),
			}
			for j := range reps {
				if irSeg.Sep != nil && j != 0 {
					repeated.Segs = append(repeated.Segs, []Segment{sep})
				}
				repeated.Segs = append(repeated.Segs, []Segment{body})
			}
			res.Segs[0][i] = repeated
		}
	default:
		panic("unhandled case")
	}
	return res
}

// Compile compiles the given expression.
//
// Pass opts as CompileOptions{} for default values.
//
// If err is of type [Error], it also has positional information.
func Compile(src []byte, opts CompileOptions) (prog Segment, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("compiling pattern: %w", err)
		}
	}()
	_, irSeg, err := parse(src, opts.Vars, opts.Funcs)
	if err != nil {
		return Segment{}, err
	}
	seg := compile(irSeg, opts)
	seg.calculateComps()
	if !opts.NoOptimize {
		seg.optimize()
	}
	if seg.Comp == math.MaxInt {
		comp, _ := seg.CompBig(nil).Float64()
		return Segment{}, fmt.Errorf("%w: complexity is %.0f, ~%.2e times too large",
			ErrComplexityTooLarge,
			comp,
			comp/float64(seg.Comp))
	}
	return seg, nil
}

func ChoiceFromStrings(strs iter.Seq[string]) IrSegmentChoice {
	var res IrSegmentChoice
	for s := range strs {
		res = append(res, IrSegmentStr(s))
	}
	return res
}
