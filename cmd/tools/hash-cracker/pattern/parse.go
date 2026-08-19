package pattern

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

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

// pattern parser.
//
// Uses an (always caught) panic for control flow on error.
type parser struct {
	src []byte
	i   int

	ir ir
}

// Makes the parser error at the last position.
func (p *parser) err(e error) {
	panic(&Error{Position: p.i - 1, Err: e})
}

// If charset is empty, errors unconditionally with
// message "unexpected '<last char>'".
//
// If charset is nonempty and the last char is not in charset,
// errors with message "expected 'a', 'b' or 'c', but got '<last char>'".
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
