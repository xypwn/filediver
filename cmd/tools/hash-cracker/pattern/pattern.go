package pattern

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrNotAscii                       = errors.New("pattern must be ASCII-only")
	ErrNullCharacter                  = errors.New("pattern must not contain the null ('\\0') character")
	ErrUnexpectedEOF                  = errors.New("unexpected end of file")
	ErrNumberTooLarge                 = errors.New("number too large")
	ErrExpectedNonnegativeInteger     = errors.New("expected nonnegative whole number")
	ErrExpectedExpressionBeforeRepeat = errors.New("expected expression before repeat expression")
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

type (
	ir struct {
		names map[string]irSegment
		body  irSegment
	}
	irSegment       interface{ String() string }
	irSegmentChoice []irSegment // <a|b|c>
	irSegmentConcat []irSegment // <a><b><c>
	irSegmentStr    string      // <someString>
	irSegmentName   string      // <$name>
	irSegmentRepeat struct {
		Seg      irSegment
		Min, Max int
	} // <something{1,2}>
)

func (s irSegmentChoice) String() string {
	var b strings.Builder
	b.WriteString("<")
	for i, seg := range s {
		if i != 0 {
			b.WriteString("|")
		}
		b.WriteString(seg.String())
	}
	b.WriteString(">")
	return b.String()
}
func (s irSegmentConcat) String() string {
	var b strings.Builder
	for _, seg := range s {
		b.WriteString(seg.String())
	}
	return b.String()
}
func (s irSegmentStr) String() string {
	return strconv.Quote(string(s))
}
func (s irSegmentName) String() string {
	return "$" + strconv.Quote(string(s))
}
func (s irSegmentRepeat) String() string {
	return fmt.Sprintf("%s{%d,%d}", s.Seg, s.Min, s.Max)
}

type parser struct {
	src []byte
	i   int

	ir ir
}

// Makes the parser error at the last position.
//
// BREAKS CONTROL FLOW VIA PANIC!
func (p *parser) err(e error) {
	panic(&Error{Position: p.i - 1, Err: e})
}

// If charset is empty, errors unconditionally with
// message "unexpected '<last char>'".
//
// If charset is nonempty and the last char is not in charset,
// errors with message "expected 'a', 'b' or 'c', but got '<last char>'".
//
// MAY BREAK CONTROL FLOW VIA PANIC!
func (p *parser) expectLast(charset string) {
	charStr := func(c byte) string {
		if c == 0 {
			return "end of file"
		} else {
			return strconv.QuoteRune(rune(c))
		}
	}
	c := p.last()
	if charset == "" {
		p.err(fmt.Errorf("unexpected %s", charStr(c)))
	} else {
		if !strings.ContainsRune(charset, rune(c)) {
			var msg strings.Builder
			msg.WriteString("expected ")
			for i := range charset {
				msg.WriteString(charStr(charset[i]))
				if i < len(charset)-2 {
					msg.WriteString(", ")
				} else if i == len(charset)-2 {
					msg.WriteString(" or ")
				}
			}
			msg.WriteString(", but got ")
			msg.WriteString(charStr(c))
			p.err(errors.New(msg.String()))
		}
	}
}

// 0 if EOF or BOF
func (p *parser) last() byte {
	if p.i == 0 || p.i >= len(p.src)+1 {
		return 0
	}
	return p.src[p.i-1]
}

// 0 if EOF
func (p *parser) next() byte {
	if p.i >= len(p.src) {
		p.i++
		return 0
	}
	b := p.src[p.i]
	p.i++
	return b
}

// -1 for empty int
func (p *parser) parseNonegativeInt() (number int, delim byte) {
	n := -1
	for {
		c := p.next()
		if c < '0' || c > '9' {
			return n, c
		}
		if n == -1 {
			n = 0
		} else {
			if n > math.MaxInt/10 {
				p.err(ErrNumberTooLarge)
			}
			n *= 10
		}
		n += int(c - '0')
	}
}

func (p *parser) parseSegmentRepeatMinMax() irSegmentRepeat {
	min, delim := p.parseNonegativeInt()
	if min == -1 {
		p.err(ErrExpectedNonnegativeInteger)
	}
	p.expectLast(",}")
	switch delim {
	case '}':
		return irSegmentRepeat{Min: min, Max: min}
	case ',':
	}
	max, delim := p.parseNonegativeInt()
	if max == -1 {
		p.err(ErrExpectedNonnegativeInteger)
	}
	p.expectLast("}")
	if max < min {
		p.err(fmt.Errorf("expected minimum repetitions (%d) to be less than maximum (%d) repetitions", min, max))
	}
	return irSegmentRepeat{Min: min, Max: max}
}

func (p *parser) parseExpr(toplevel bool) irSegment {
	var segIsName bool
	var segStr strings.Builder

	// General hierarchy constructed here is:
	// segment = choice of parts of segment
	// example: <a>b{2}|c
	//   (choice between (a concatenated with bb) and c)
	// where:
	//   choice is an ordered set
	//   parts is a concatenation
	var seg irSegment
	var segParts []irSegment
	var segChoices []irSegment
	flushSegStr := func() {
		if segStr.Len() == 0 {
			return
		}
		str := strings.Clone(segStr.String())
		if segIsName {
			seg = irSegmentName(str)
		} else {
			seg = irSegmentStr(str)
		}
		segIsName = false
		segStr.Reset()
	}
	flushSegPart := func() {
		flushSegStr()
		if seg != nil {
			segParts = append(segParts, seg)
		}
		seg = nil
	}
	flushSegChoice := func() {
		flushSegPart()
		if len(segParts) > 0 {
			if len(segParts) == 1 {
				segChoices = append(segChoices, segParts[0])
			} else {
				segChoices = append(segChoices, irSegmentConcat(segParts))
			}
		}
		segParts = nil
	}

loop:
	for {
		c := p.next()
		switch c {
		case '>', 0:
			if toplevel {
				p.expectLast("\x00")
			} else {
				p.expectLast(">")
			}
			flushSegChoice()
			break loop
		case '\\':
			c1 := p.next()
			if c1 == 0 {
				p.err(ErrUnexpectedEOF)
			}
			segStr.WriteByte(c1)
		case '|':
			flushSegChoice()
		case '<':
			flushSegPart()
			seg = p.parseExpr(false)
		case '$':
			if segStr.Len() == 0 {
				segIsName = true
			} else {
				p.expectLast("")
			}
		case '{':
			flushSegPart()
			if len(segParts) == 0 {
				p.err(ErrExpectedExpressionBeforeRepeat)
			}
			rs := p.parseSegmentRepeatMinMax()
			rs.Seg = segParts[len(segParts)-1]
			segParts[len(segParts)-1] = rs
		default:
			if seg != nil {
				flushSegPart()
			}
			segStr.WriteByte(c)
		}
	}
	if len(segChoices) == 1 {
		return segChoices[0]
	} else if len(segChoices) == 0 {
		return nil
	} else {
		return irSegmentChoice(segChoices)
	}
}

func parse(src []byte) (irSeg irSegment, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*Error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	for _, c := range src {
		if c > 0x7f {
			return nil, ErrNotAscii
		}
		if c == 0 {
			return nil, ErrNullCharacter
		}
	}

	p := parser{
		src: src,
	}
	irSeg = p.parseExpr(true)
	return
}

type segment struct {
	// text (if str != "")
	str string
	// cartesian product of ordered sets (if str == "")
	//
	// (note that this is in a way the opposite of how
	// parse models things, as parse has an ordered set
	// of cartesian products; the reason we're using
	// this model is for the compiled segments is better
	// data locality)
	segs [][]segment
	// Number of elements (in this segment and in
	// each cartesian product operand).
	// math.MaxInt if >= math.MaxInt
	comp  int
	comps []int
}

type segIdx struct {
	segs     [][]segIdx
	idxs     []int
	totalIdx int // total index in [0,comp-1]
}

// Reset zeroes out the index.
func (idx *segIdx) Reset() {
	for i := range idx.idxs {
		idx.idxs[i] = 0
		for j := range idx.segs[i] {
			idx.segs[i][j].Reset()
		}
	}
	idx.totalIdx = 0
}

func (idx segIdx) String() string {
	if idx.segs == nil && idx.idxs == nil {
		return "nil"
	}
	var res strings.Builder
	fmt.Fprintf(&res, "segs=%v,", idx.segs)
	fmt.Fprintf(&res, "idxs=%v", idx.idxs)
	return res.String()
}

// Adds n to the [segIdx] for the [segment]. The
// index MUST have been generated from the given
// segment via [segment.makeIndex].
func (idx segIdx) Add(s segment, n int) (carry int) {
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
		if s.comps[i] == len(s.segs[i]) {
			v := idx.idxs[i] + n
			carry = v / s.comps[i]
			idx.idxs[i] = v % s.comps[i]
			return
		}

		// Slow path: Some union operands are cartesian products
		carry = n / s.comps[i]
		n %= s.comps[i]
		j := idx.idxs[i]
		if idx.segs[i][j].totalIdx >= s.segs[i][j].comp-n {
			// Fill first union operand fully if it would overflow
			n -= s.segs[i][j].comp - idx.segs[i][j].totalIdx
			idx.segs[i][j].Reset() // zero the fully filled operand
			j++
			if j >= len(s.segs[i]) {
				carry++
				j = 0
			}
			for n >= s.segs[i][j].comp {
				// Fill remaining union operands that would overflow
				// (we don't have to do anything to the index since
				// it's already zeroed and would remain that way)
				n -= s.segs[i][j].comp
				j++
				if j >= len(s.segs[i]) {
					carry++
					j = 0
				}
			}
		}
		if n != 0 {
			// Fill the rest into the union operand that can't be
			// filled fully
			c := idx.segs[i][j].Add(s.segs[i][j], n)
			if c != 0 {
				panic("expected carry to be zero")
			}
			n = 0
		}
		idx.idxs[i] = j
		return
	}
	carry = n
	for i := range slices.Backward(s.segs) { // cart. prod. terms
		if carry == 0 {
			break
		}
		carry = addToUnion(i, carry)
	}
	idx.totalIdx = (idx.totalIdx + n) % s.comp
	return
}

func (s segment) AppendAt(buf []byte, idx segIdx) []byte {
	if s.str != "" {
		buf = append(buf, s.str...)
	}
	for i := range s.segs {
		j := idx.idxs[i]
		buf = s.segs[i][j].AppendAt(buf, idx.segs[i][j])
	}
	return buf
}

func (s segment) StringAt(idx segIdx) string {
	return string(s.AppendAt(nil, idx))
}

type Program struct {
	segment
}

func (s segment) String() string {
	if s.str != "" {
		return strconv.Quote(s.str)
	} else {
		var b strings.Builder
		for _, segs := range s.segs {
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

func (s *segment) calculateComps() {
	if s.str != "" {
		s.comp = 1
		return
	}
	if s.comps != nil {
		return
	}
	s.comps = make([]int, len(s.segs))
	prod := 1
	for i := range s.segs {
		sum := 0
		for j := range s.segs[i] {
			s.segs[i][j].calculateComps()
			seg := s.segs[i][j]
			if seg.comp == 0 {
				panic("segment complexity should be nonzero after updateComps()")
			}
			if sum > math.MaxInt-seg.comp {
				sum = math.MaxInt // addition would overflow
			} else {
				sum += seg.comp
			}
		}
		if prod > math.MaxInt/sum {
			prod = math.MaxInt // multiplication would overflow
		} else {
			prod *= sum
		}
		s.comps[i] = sum
	}
	s.comp = prod
}

func (s segment) makeIndex() (idx segIdx) {
	if s.str != "" {
		return
	}
	idx.idxs = make([]int, len(s.segs))
	idx.segs = make([][]segIdx, len(s.segs))
	for i, segs := range s.segs {
		idx.segs[i] = make([]segIdx, len(segs))
		for j, seg := range segs {
			idx.segs[i][j] = seg.makeIndex()
		}
	}
	return
}

// Comp calculates the complexity (number of elements),
// or math.MaxInt if it is >= math.MaxInt.
func (s segment) Comp() int {
	if s.str != "" {
		return 1
	}
	prod := 1
	for _, segs := range s.segs {
		sum := 0
		for _, seg := range segs {
			c := seg.Comp()
			if c == math.MaxInt {
				return math.MaxInt
			}
			// TODO: Fix possible overflow from addition
			sum += c
		}
		if prod > math.MaxInt/sum {
			return math.MaxInt // would overflow
		}
		prod *= sum
	}
	return prod
}

// CompBig calculates the complexity (number of elements)
// as a big int, or max if the number is bigger than max.
//
// Pass nil as max to ignore maximum.
func (s segment) CompBig(max *big.Int) *big.Int {
	if s.str != "" {
		return big.NewInt(1)
	}
	prod := big.NewInt(1)
	sum := &big.Int{}
	for _, segs := range s.segs {
		sum.SetInt64(0)
		for _, seg := range segs {
			c := seg.CompBig(max)
			if max != nil && c.Cmp(max) >= 0 {
				return max
			}
			// TODO: Fix possible overflow from addition
			sum.Add(sum, c)
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
func (s segment) MaxLen() int {
	if s.str != "" {
		return len(s.str)
	}
	concatLen := 0
	for _, segs := range s.segs {
		unionLen := 0
		for _, seg := range segs {
			unionLen = max(unionLen, seg.MaxLen())
		}
		concatLen += unionLen
	}
	return concatLen
}

func (s *segment) optimize() {
	// Optimize inner segments first.
	for i := range s.segs {
		for j := range s.segs[i] {
			s.segs[i][j].optimize()
		}
	}

	// Empty cartesian products or union elements can
	// be removed.
	if i := slices.IndexFunc(s.segs, func(segs []segment) bool {
		return len(segs) == 0
	}); i != -1 {
		var newSegs [][]segment
		var newComps []int
		for ; i < len(s.segs); i++ {
			if len(s.segs[i]) != 0 {
				newSegs = append(newSegs, s.segs[i])
				newComps = append(newComps, s.comps[i])
			}
		}
		s.segs = newSegs
		s.comps = newComps
	}
	for i := range s.segs {
		s.segs[i] = slices.DeleteFunc(s.segs[i], func(s segment) bool {
			return s.str == "" && len(s.segs) == 0
		})
	}

	// Cartesian product of single-parameter unions
	// can be flattened.
	//
	// Example:
	// x(A, u(x(B, C)), u(x(D))) -> x(A, B, C, D)
	if i := slices.IndexFunc(s.segs, func(segs []segment) bool {
		return len(segs) == 1
	}); i != -1 {
		var newSegs [][]segment
		var newComps []int
		for ; i < len(s.segs); i++ {
			if len(s.segs[i]) == 1 {
				newSegs = append(newSegs, s.segs[i][0].segs...)
				newComps = append(newComps, s.segs[i][0].comps...)
			} else {
				newSegs = append(newSegs, s.segs[i])
				newComps = append(newComps, s.comps[i])
			}
		}
		s.segs = newSegs
		s.comps = newComps
	}

	// Cartesian product with single parameter
	// of union with single parameter can be
	// flattened.
	//
	// Example:
	// x(u(A)) -> A
	if len(s.segs) == 1 && len(s.segs[0]) == 1 {
		*s = s.segs[0][0]
	}
}

func compile(irSeg irSegment) segment {
	var res segment
	switch irSeg := irSeg.(type) {
	case irSegmentChoice:
		res.segs = make([][]segment, 1)
		res.segs[0] = make([]segment, len(irSeg))
		for i := range irSeg {
			res.segs[0][i] = compile(irSeg[i])
		}
	case irSegmentConcat:
		res.segs = make([][]segment, len(irSeg))
		for i := range irSeg {
			res.segs[i] = []segment{compile(irSeg[i])}
		}
	case irSegmentStr:
		res.str = string(irSeg)
	case irSegmentRepeat:
		// Expand repeat (e.g. a{1,2} -> a|aa)
		body := compile(irSeg.Seg)
		if irSeg.Min == 1 && irSeg.Max == 1 {
			return body
		}
		res.segs = make([][]segment, 1)
		res.segs[0] = make([]segment, irSeg.Max-irSeg.Min+1)
		for i := range res.segs[0] {
			reps := irSeg.Min + i
			repeated := segment{segs: make([][]segment, reps)}
			for j := range repeated.segs {
				repeated.segs[j] = []segment{body}
			}
			res.segs[0][i] = repeated
		}
	case irSegmentName:
		panic("TODO")
	default:
		panic("unhandled case")
	}

	res.calculateComps()
	res.optimize()
	return res
}

func Compile(src []byte) (err error) {
	irSeg, err := parse(src)

	var ir ir
	ir.body = irSeg
	//TODO
	return nil
}
