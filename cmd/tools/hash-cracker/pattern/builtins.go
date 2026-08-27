package pattern

import (
	"fmt"
	"regexp"
	"strconv"
)

var builtinVars = map[string]IrSegment{}

var builtinFuncs = map[string]any{
	// limit limits the number of choices to at most n, where
	// n must be a non-negative number.
	"limit": func(s IrSegmentChoice, n string) (IrSegmentChoice, error) {
		nInt, err := strconv.Atoi(n)
		if err != nil {
			return nil, err
		}
		if len(s) > nInt {
			s = s[:nInt]
		}
		return s, nil
	},
	// filter only keeps the choices that fully match the given regex.
	//
	// All choices must be strings.
	"filter": func(s IrSegmentChoice, re string) (IrSegmentChoice, error) {
		r, err := regexp.Compile("^" + re + "$") // HACK
		if err != nil {
			return nil, err
		}
		var res IrSegmentChoice
		for _, c := range s {
			str, ok := c.(IrSegmentStr)
			if !ok {
				return nil, fmt.Errorf("expected choice to consist only of strings")
			}
			if r.MatchString(string(str)) {
				res = append(res, c)
			}
		}
		return res, nil
	},

	// =====================
	// Special functions
	// =====================
	// These have to be implemented in the parser, as they
	// modify the current parser state.

	// load loads the given pattern from a file.
	"load": (func(filename string) (IrSegment, error))(nil),
	// import imports any variables from the given
	// file.
	"import": (func(filename string) (IrSegment, error))(nil),

	// =====================
	// End special functions
	// =====================
}
