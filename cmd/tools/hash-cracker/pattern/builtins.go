package pattern

import (
	"fmt"
	"regexp"
	"strconv"
)

var builtinVars = map[string]IrSegment{}

var builtinFuncs = map[string]any{
	// sep separates each part in the given repeat segment by the given
	// separator. Doesn't change the repeat segment's complexity if the
	// separator has a complexity of 1.
	//
	// Example: ${sep [ab]{0,2} _} -> "", "a", "b", "a_a", "a_b", "b_a", "b_b"
	"sep": func(s IrSegmentRepeat, sep IrSegment) IrSegment {
		ch := IrSegmentChoice{}
		for i := s.Min; i <= s.Max; i++ {
			switch i {
			case 0:
				ch = append(ch, IrSegmentStr(""))
			case 1:
				ch = append(ch, s.Seg)
			default:
				ch = append(ch,
					IrSegmentConcat{
						s.Seg,
						IrSegmentRepeat{
							Min: i - 1, Max: i - 1,
							Seg: IrSegmentConcat{
								sep,
								s.Seg,
							},
						},
					})
			}
		}
		return ch
	},
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
