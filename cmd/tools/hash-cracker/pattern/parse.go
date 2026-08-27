package pattern

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/xypwn/filediver/util"
)

// IR (intermediate representation) is basically the AST
// for the pattern language. It is what any complex operations
// are based on and provides the data type variables, functions,
// character classes etc. operate on.
//
// The compiled language ([Segment]) is a flatter language
// tree comprised of only two operations.
type (
	Ir struct {
		names map[string]IrSegment
		body  IrSegment
	}
	IrSegment interface {
		String() string
		Clone() IrSegment
		private()
	}
	IrSegmentChoice []IrSegment // <a|b|c>
	IrSegmentConcat []IrSegment // <a><b><c>
	IrSegmentStr    string      // <someString>
	IrSegmentRepeat struct {
		Seg      IrSegment
		Sep      IrSegment // optional separator
		Min, Max int
	} // <something{1,2}>
)

func (s IrSegmentChoice) private() {}
func (s IrSegmentConcat) private() {}
func (s IrSegmentStr) private()    {}
func (s IrSegmentRepeat) private() {}

func (s IrSegmentChoice) String() string {
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
func (s IrSegmentChoice) Clone() IrSegment {
	sNew := slices.Clone(s)
	for i := range sNew {
		sNew[i] = sNew[i].Clone()
	}
	return sNew
}

func (s IrSegmentConcat) String() string {
	var b strings.Builder
	for _, seg := range s {
		b.WriteString(seg.String())
	}
	return b.String()
}
func (s IrSegmentConcat) Clone() IrSegment {
	sNew := slices.Clone(s)
	for i := range sNew {
		sNew[i] = sNew[i].Clone()
	}
	return sNew
}

func (s IrSegmentStr) String() string {
	return strconv.Quote(string(s))
}
func (s IrSegmentStr) Clone() IrSegment {
	return IrSegmentStr(strings.Clone(string(s)))
}

func (s IrSegmentRepeat) String() string {
	if s.Sep != nil {
		return fmt.Sprintf("%s{%d,%d,%s}", s.Seg, s.Min, s.Max, s.Sep)
	} else {
		return fmt.Sprintf("%s{%d,%d}", s.Seg, s.Min, s.Max)
	}
}
func (s IrSegmentRepeat) Clone() IrSegment {
	sNew := s
	sNew.Seg = sNew.Seg.Clone()
	return sNew
}

func isValidFuncOrVarNameChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '-'
}

// pattern parser.
//
// Uses an (always caught) panic for control flow on error.
type parser struct {
	src []byte
	i   int

	// All functions and variables
	funcs map[string]any
	vars  map[string]IrSegment
	// Functions and variables present during initialization
	initialFuncs map[string]any
	initialVars  map[string]IrSegment
	// Variables assigned during pattern
	// parsing
	newVars []string

	ir Ir
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

// panics if nothing was read yet
func (p *parser) unread() {
	if p.i == 0 {
		panic("unread called without reading anything before")
	}
	p.i--
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

func (p *parser) parseSegmentRepeatMinMaxSep() IrSegmentRepeat {
	var min, max int
	var delim byte
	isNextInt := func() bool {
		c := p.next()
		p.unread()
		return c >= '0' && c <= '9'
	}
	min, delim = p.parseNonegativeInt()
	if min == -1 {
		p.err(ErrExpectedNonnegativeInteger)
	}
	p.expectLast(",}")
	if delim == '}' {
		return IrSegmentRepeat{Min: min, Max: min}
	}
	if isNextInt() {
		max, delim = p.parseNonegativeInt()
		if max == -1 {
			p.err(ErrExpectedNonnegativeInteger)
		}
		p.expectLast(",}")
		if max < min {
			p.err(fmt.Errorf("expected minimum repetitions (%d) to be less than maximum (%d) repetitions", min, max))
		}
		if delim == '}' {
			return IrSegmentRepeat{Min: min, Max: max}
		}
	}
	sep := p.parseExpr("}")
	return IrSegmentRepeat{Min: min, Max: max, Sep: sep}
}

func (p *parser) parseCharClass() IrSegmentChoice {
	var complement bool // [^...]
	var chars []byte
loop:
	for {
		c := p.next()
		switch c {
		case 0:
			p.expectLast("")
		case ']':
			break loop
		case '\\':
			c1 := p.next()
			if c1 == 0 {
				p.expectLast("")
			}
			chars = append(chars, c1)
		default:
			if len(chars) == 0 && c == '^' {
				complement = true
			} else if len(chars) >= 2 && chars[len(chars)-1] == '-' {
				start, end := chars[len(chars)-2], c
				chars = chars[:len(chars)-2]
				if start > end || start < 0x20 || end > 0x7e {
					p.err(ErrInvalidCharClassRange)
				}
				for i := start; i <= end; i++ {
					chars = append(chars, i)
				}
			} else {
				chars = append(chars, c)
			}
		}
	}
	slices.Sort(chars)
	chars = util.Uniq(chars)
	var seg IrSegmentChoice
	if complement {
		// Iterating over chars and filling the gaps would be more efficient,
		// but this is probably good enough.
		for c := byte(0x20); c <= 0x7e; c++ {
			if _, ok := slices.BinarySearch(chars, c); !ok {
				seg = append(seg, IrSegmentStr(string(c)))
			}
		}
	} else {
		for _, c := range chars {
			seg = append(seg, IrSegmentStr(string(c)))
		}
	}
	return seg
}

// Leaves the parser after the next non-whitespace char
// (you probably want to call unread() after invoking
// this function).
func (p *parser) skipWhitespace() {
	for {
		c := p.next()
		if c != ' ' && c != '\t' {
			return
		}
	}
}

// Parses a function parameter string.
func (p *parser) parseString() string {
	var s strings.Builder
	parseUnquoted := func() {
		for {
			c := p.next()
			if c == 0 {
				p.expectLast("")
			} else if c == ' ' || c == '\t' || c == '}' {
				break
			} else if c == '\\' {
				c1 := p.next()
				if c1 == 0 {
					p.expectLast("")
				}
				s.WriteByte(c1)
			} else {
				s.WriteByte(c)
			}
		}
	}
	parseQuoted := func(quote rune) {
		for {
			c := p.next()
			if c == 0 {
				p.expectLast("")
			} else if rune(c) == quote {
				break
			} else if c == '\\' {
				c1 := p.next()
				if c1 == 0 {
					p.expectLast("")
				}
				s.WriteByte(c1)
			} else {
				s.WriteByte(c)
			}
		}
	}
	c := p.next()
	if c == 0 {
		p.expectLast("")
	}
	switch c {
	case '\'':
		parseQuoted('\'')
	case '"':
		parseQuoted('"')
	default:
		s.WriteByte(c)
		parseUnquoted()
	}
	return s.String()
}

// Evaluates an assignment, variable read or function
// call. Returns the resulting value.
//
// May return nil, indicating no produced value.
func (p *parser) parseHashtagExpr() IrSegment {
	var nameB strings.Builder
	for {
		c := p.next()
		if c == 0 {
			p.expectLast("")
		}
		if c == '}' {
			p.unread()
			break
		}
		if c == ' ' || c == '\t' {
			p.skipWhitespace()
			p.unread()
			break
		}
		if !isValidFuncOrVarNameChar(rune(c)) {
			p.err(ErrInvalidVarOrFuncNameChar)
		}
		nameB.WriteByte(c)
	}
	name := nameB.String()
	if len(name) == 0 {
		p.err(ErrEmptyVarOrFuncName)
	}

	fn, isFuncCall := p.funcs[name]
	tfn := reflect.TypeOf(fn)

	var args []any
	for i := 0; ; i++ {
		c := p.next()
		if c == '}' {
			break
		}
		p.unread()
		isStr := false
		if isFuncCall {
			var tin reflect.Type
			if i < tfn.NumIn()-1 {
				tin = tfn.In(i)
			} else {
				if tfn.IsVariadic() {
					tin = tfn.In(tfn.NumIn() - 1).Elem()
				} else if i < tfn.NumIn() {
					tin = tfn.In(i)
				}
			}
			if tin != nil && reflect.TypeFor[string]().AssignableTo(tin) {
				isStr = true
			}
		}
		var arg any
		if isStr {
			// Parse string
			arg = p.parseString()
		} else {
			// Parse [IrSegment]
			arg = p.parseExpr(" \t}")
			if arg == nil {
				p.err(fmt.Errorf("%w: %s: nil parameter passed as argument %d", ErrInvalidFuncCall, name, i+1))
			}
		}
		args = append(args, arg)
		c = p.last()
		if c == ' ' || c == '\t' {
			p.skipWhitespace()
		}
		p.unread() // next char is now non-whitespace
	}

	// Assignment
	if len(args) >= 2 {
		if s, ok := args[0].(IrSegmentStr); ok && s == "=" {
			if len(args) != 2 {
				p.err(ErrTooManyArgsInAssignment)
			}
			if _, exists := p.vars[name]; exists {
				p.err(fmt.Errorf("%w: %s", ErrVarOrFuncAlreadyExists, name))
			}
			if _, exists := p.funcs[name]; exists {
				p.err(fmt.Errorf("%w: %s", ErrVarOrFuncAlreadyExists, name))
			}
			p.vars[name] = args[1].(IrSegment)
			p.newVars = append(p.newVars, name)
			return nil
		}
	}

	// Variable
	if !isFuncCall && len(args) == 0 {
		val, ok := p.vars[name]
		if !ok {
			p.err(fmt.Errorf("%w: %s", ErrUnknownVarOrFunc, name))
		}
		return val
	}

	// Function call
	{
		fn, ok := p.funcs[name]
		if !ok {
			p.err(fmt.Errorf("%w: %s", ErrUnknownVarOrFunc, name))
		}
		vfn := reflect.ValueOf(fn)
		tfn := vfn.Type()
		if tfn.IsVariadic() {
			if len(args) < tfn.NumIn()-1 {
				p.err(fmt.Errorf("%w: %s: expected at least %d parameters, but got %d", ErrInvalidFuncCall, name, tfn.NumIn()-1, len(args)))
			}
		} else {
			if len(args) != tfn.NumIn() {
				p.err(fmt.Errorf("%w: %s: expected exactly %d parameters, but got %d", ErrInvalidFuncCall, name, tfn.NumIn(), len(args)))
			}
		}
		for i := range args {
			tgot := reflect.TypeOf(args[i])
			var twant reflect.Type
			if i < tfn.NumIn()-1 {
				twant = tfn.In(i)
			} else {
				if tfn.IsVariadic() {
					twant = tfn.In(tfn.NumIn() - 1).Elem()
				} else {
					twant = tfn.In(i)
				}
			}
			if !tgot.AssignableTo(twant) {
				p.err(fmt.Errorf("%w: %s: expected type of parameter %d to be assignable to %s, but got %s", ErrInvalidFuncCall, name, i+1, twant, tgot))
			}
		}
		funcCallErr := func(err error) {
			p.err(fmt.Errorf("%w: %s: %w", ErrCallingFunc, name, err))
		}
		if name == "load" || name == "import" {
			// Special function "load" can't be handled
			// like normal function, since it modifies the
			// current parser instance.
			filename := args[0].(string)
			data, err := os.ReadFile(filename)
			if err != nil {
				funcCallErr(err)
			}
			p1, seg, err := parse(data, p.initialVars, p.initialFuncs)
			if err != nil {
				funcCallErr(err)
			}
			switch name {
			case "load":
				if seg == nil {
					funcCallErr(fmt.Errorf("loading file %q: doesn't return any values (did you mean to use import?)", filename))
				}
				return seg
			case "import":
				for _, vn := range p1.newVars {
					if _, exists := p.vars[vn]; exists {
						funcCallErr(fmt.Errorf("cannot import file %q: variable %s: %w", filename, vn, ErrVarOrFuncAlreadyExists))
					}
					p.vars[vn] = p1.vars[vn]
					p.newVars = append(p.newVars, vn)
				}
				return nil
			default:
				panic("unhandled case")
			}
		} else {
			// Normal function call
			vins := make([]reflect.Value, len(args))
			for i := range vins {
				vins[i] = reflect.ValueOf(args[i])
			}
			vouts := vfn.Call(vins)
			if len(vouts) == 2 && !vouts[1].IsNil() {
				funcCallErr(vouts[1].Interface().(error))
			}
			if vouts[0].IsNil() {
				return nil
			}
			return vouts[0].Interface().(IrSegment)
		}
	}
}

func (p *parser) parseExpr(endingDelims string) IrSegment {
	var segStr strings.Builder

	// General hierarchy constructed here is:
	// segment = choice of parts of segment
	// example: <a>b{2}|c
	//   (choice between (a concatenated with bb) and c)
	// where:
	//   choice is an ordered set
	//   parts is a concatenation
	var seg IrSegment
	var segParts []IrSegment
	var segChoices []IrSegment
	flushSegStr := func() {
		if segStr.Len() == 0 {
			return
		}
		str := strings.Clone(segStr.String())
		seg = IrSegmentStr(str)
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
				segChoices = append(segChoices, IrSegmentConcat(segParts))
			}
		}
		segParts = nil
	}

	for {
		c := p.next()
		if strings.ContainsRune(endingDelims, rune(c)) || c == 0 {
			p.expectLast(endingDelims)
			flushSegChoice()
			break
		}
		switch c {
		case '\\':
			c1 := p.next()
			if c1 == 0 {
				p.expectLast("")
			}
			segStr.WriteByte(c1)
		case '|', '\n':
			flushSegChoice()
		case '<':
			flushSegPart()
			seg = p.parseExpr(">")
		case '[':
			flushSegPart()
			seg = p.parseCharClass()
		case '#':
			flushSegPart()
			p.next()
			p.expectLast("{")
			newSeg := p.parseHashtagExpr()
			if newSeg != nil {
				seg = newSeg
			}
		case '{':
			flushSegPart()
			if len(segParts) == 0 {
				p.err(ErrExpectedExpressionBeforeRepeat)
			}
			rs := p.parseSegmentRepeatMinMaxSep()
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
		return IrSegmentChoice(segChoices)
	}
}

func (p *parser) checkVarAndFuncNamesAndTypes() {
	for name, v := range p.vars {
		for _, r := range name {
			if !isValidFuncOrVarNameChar(r) {
				p.err(fmt.Errorf("%w: %s", ErrInvalidVarOrFuncNameChar, name))
			}
		}
		if v == nil {
			p.err(fmt.Errorf("%w: %s: expected variable to be non-nil", ErrInvalidVarType, name))
		}
	}
	for name, fn := range p.funcs {
		err := func(err error, format string, args ...any) {
			p.err(fmt.Errorf("%w: %s: %s", err, name, fmt.Sprintf(format, args...)))
		}
		for _, r := range name {
			if !isValidFuncOrVarNameChar(r) {
				p.err(fmt.Errorf("%w: %s", ErrInvalidVarOrFuncNameChar, name))
			}
		}
		tfn := reflect.TypeOf(fn)
		{
			nOut := tfn.NumOut()
			if nOut == 0 {
				err(ErrInvalidFuncType, "expected func to return at an IrSegment and an optional error (1 or 2 return parameters), but got none")
			}
			out0 := tfn.Out(0)
			if !out0.AssignableTo(reflect.TypeFor[IrSegment]()) {
				err(ErrInvalidFuncType, "expected func's first output parameter to be of type IrSegment, but got %s", out0.String())
			}
			if nOut >= 2 {
				out1 := tfn.Out(1)
				if !out1.AssignableTo(reflect.TypeFor[error]()) {
					err(ErrInvalidFuncType, "expected func's optional second output parameter to be of type error, but got %s", out1.String())
				}
			}
			if nOut >= 3 {
				err(ErrInvalidFuncType, "expected func to return at most 2 parameters, but got %d", nOut)
			}
		}
		checkInputType := func(typ reflect.Type, paramStr string) {
			if !typ.Implements(reflect.TypeFor[IrSegment]()) && !reflect.TypeFor[string]().AssignableTo(typ) {
				err(ErrInvalidFuncType, "expected %s to be of type IrSegment or string, but got %s", paramStr, typ.String())
			}
		}
		for i := range tfn.NumIn() {
			in := tfn.In(i)
			if tfn.IsVariadic() && i == tfn.NumIn()-1 {
				checkInputType(in.Elem(), "variadic parameter")
			} else {
				checkInputType(in, fmt.Sprintf("parameter %d", i+1))
			}
		}
	}
}

func parse(src []byte, extraVars map[string]IrSegment, extraFuncs map[string]any) (p *parser, irSeg IrSegment, err error) {
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
			return nil, nil, ErrNotAscii
		}
		if c == 0 {
			return nil, nil, ErrNullCharacter
		}
	}

	p = &parser{
		src:          src,
		initialFuncs: make(map[string]any),
		initialVars:  make(map[string]IrSegment),
	}
	maps.Copy(p.initialFuncs, builtinFuncs)
	maps.Copy(p.initialVars, builtinVars)
	maps.Copy(p.initialFuncs, extraFuncs)
	maps.Copy(p.initialVars, extraVars)
	p.funcs = maps.Clone(p.initialFuncs)
	p.vars = maps.Clone(p.initialVars)
	p.checkVarAndFuncNamesAndTypes()
	irSeg = p.parseExpr("\x00")
	return
}
