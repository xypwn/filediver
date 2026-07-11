package xaml

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// See https://learn.microsoft.com/en-us/windows/apps/develop/platform/xaml/move-draw-commands-syntax.

type FillRule uint8

const (
	FillEvenOdd FillRule = iota
	FillNonzero
)

var isValidPathCmdChar = [256]bool{
	'M': true, 'L': true, 'H': true, 'V': true, 'C': true, 'Q': true, 'S': true, 'T': true, 'A': true, 'Z': true,
	'm': true, 'l': true, 'h': true, 'v': true, 'c': true, 'q': true, 's': true, 't': true, 'a': true, 'z': true,
}

var pathCmdNumParams = [256]uint8{
	'M': 2, 'L': 2, 'H': 1, 'V': 1, 'C': 6, 'Q': 4, 'S': 4, 'T': 2, 'A': 7, 'Z': 0,
}

var isWhitespace = [256]bool{
	' ': true, '\t': true, '\n': true, '\v': true, '\f': true, '\r': true,
}

type PathCmdChar byte

const (
	PathMove                  PathCmdChar = 'M'
	PathLine                  PathCmdChar = 'L'
	PathHLine                 PathCmdChar = 'H'
	PathVLine                 PathCmdChar = 'V'
	PathCubicBezier           PathCmdChar = 'C'
	PathQuadraticBezier       PathCmdChar = 'Q'
	PathSmoothCubicBezier     PathCmdChar = 'S'
	PathSmoothQuadraticBezier PathCmdChar = 'T'
	PathEllipticalArc         PathCmdChar = 'A'
	PathClose                 PathCmdChar = 'Z'
)

func (c PathCmdChar) IsValid() bool {
	return c >= 'A' && c <= 'Z' && isValidPathCmdChar[c]
}

func (c PathCmdChar) ToByte(relative bool) byte {
	if !c.IsValid() {
		return 0
	}
	b := byte(c)
	if relative {
		b += 'a' - 'A'
	}
	return b
}

type PathCmd struct {
	Cmd      PathCmdChar // uppercase version of command char
	Relative bool        // whether position is relative, i.e. command char is uppercase
	Params   [7]float64
}

type PathData struct {
	FillRule FillRule
	Cmds     []PathCmd
}

// Returns a simplified copy of the path data for
// easier rendering or conversion.
//   - Converts all relative positions to absolute.
//   - Converts all commands into the following subset: M,L,C,Q,A,Z (removes H,V,S,T).
//   - In the Z command, the first point x, y are now the first 2 parameters for convenience.
func (pd PathData) Simplify() PathData {
	// Reflects x, y over px, py.
	reflect := func(x, y, px, py float64) (rx, ry float64) {
		return px - (x - px), py - (y - py)
	}
	// Converts all position params to absolute coords.
	paramsToAbsolute := func(cmd PathCmd, x, y float64) (r [7]float64) {
		if !cmd.Relative {
			return cmd.Params
		}
		r = cmd.Params
		firstCoordIdx := 0
		isX := true
		if cmd.Cmd == PathEllipticalArc {
			firstCoordIdx = 5 // A/a has its first coord parameter at index 5
		}
		if cmd.Cmd == PathVLine {
			isX = false // V/v takes a Y coordinate
		}
		for i := firstCoordIdx; i < int(pathCmdNumParams[cmd.Cmd]); i++ {
			if isX {
				r[i] += x
			} else {
				r[i] += y
			}
			isX = !isX
		}
		return
	}

	resPd := PathData{
		FillRule: pd.FillRule,
		Cmds:     make([]PathCmd, 0, len(pd.Cmds)),
	}
	haveFirstPoint := false
	var firstX, firstY float64 // first point or first point after "Z" command
	var x, y float64
	var ctrlPtX, ctrlPtY float64 // last control point for smooth bezier curves
	for _, cmd := range pd.Cmds {
		p := paramsToAbsolute(cmd, x, y)
		resCmd := PathCmd{Cmd: cmd.Cmd, Params: p}
		switch cmd.Cmd {
		case PathMove:
			x, y = p[0], p[1]
			if !haveFirstPoint {
				firstX, firstY = x, y
				haveFirstPoint = true
			}
		case PathLine:
			x, y = p[0], p[1]
		case PathHLine:
			resCmd = PathCmd{
				Cmd:    PathLine,
				Params: [7]float64{p[0], y},
			}
			x = p[0]
		case PathVLine:
			resCmd = PathCmd{
				Cmd:    PathLine,
				Params: [7]float64{x, p[0]},
			}
			y = p[0]
		case PathCubicBezier:
			x, y = p[4], p[5]
			ctrlPtX, ctrlPtY = p[2], p[3]
		case PathQuadraticBezier:
			x, y = p[2], p[3]
			ctrlPtX, ctrlPtY = p[0], p[1]
		case PathSmoothCubicBezier:
			cx, cy := reflect(ctrlPtX, ctrlPtY, x, y)
			resCmd = PathCmd{
				Cmd:    PathCubicBezier,
				Params: [7]float64{cx, cy, p[0], p[1], p[2], p[3]},
			}
			x, y = p[2], p[3]
			ctrlPtX, ctrlPtY = p[0], p[1]
		case PathSmoothQuadraticBezier:
			cx, cy := reflect(ctrlPtX, ctrlPtY, x, y)
			resCmd = PathCmd{
				Cmd:    PathQuadraticBezier,
				Params: [7]float64{cx, cy, p[0], p[1]},
			}
			x, y = p[0], p[1]
			ctrlPtX, ctrlPtY = p[0], p[1]
		case PathEllipticalArc:
			x, y = p[5], p[6]
		case PathClose:
			resCmd.Params = [7]float64{firstX, firstY}
			x, y = firstX, firstY
			haveFirstPoint = false
		}
		resPd.Cmds = append(resPd.Cmds, resCmd)
	}
	return resPd
}

func (pd *PathData) UnmarshalText(text []byte) error {
	*pd = PathData{}
	i := 0
	peek := func() byte {
		if i >= len(text) {
			return 0
		}
		return text[i]
	}
	skipWhitespaceComma := func() {
		for isWhitespace[peek()] || peek() == ',' {
			i++
		}
	}
	parseStr := func(s string) bool {
		if bytes.HasPrefix(text[i:], []byte(s)) {
			i += len(s)
			return true
		}
		return false
	}
	parseDouble := func() (float64, error) {
		if parseStr("Infinity") {
			return math.Inf(1), nil
		} else if parseStr("-Infinity") {
			return math.Inf(-1), nil
		} else if parseStr("NaN") {
			return math.NaN(), nil
		} else {
			start := i
			expectSign := true
			for {
				c := peek()
				ok := (c >= '0' && c <= '9') ||
					(expectSign && (c == '+' || c == '-')) ||
					c == '.' || c == 'e' || c == 'E'
				if !ok {
					break
				}
				i++
				expectSign = c == 'e' || c == 'E'
			}
			end := i
			return strconv.ParseFloat(string(text[start:end]), 64)
		}
	}
	skipWhitespaceComma()
	if parseStr("F0") {
	} else if parseStr("F1") {
		pd.FillRule = FillNonzero
	}
	for {
		skipWhitespaceComma()
		c := peek()
		i++
		if c == 0 {
			break
		}
		isRel := false
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
			isRel = true
		}
		cmd := PathCmdChar(c)
		if !cmd.IsValid() {
			return fmt.Errorf("unrecognized path command %q", cmd)
		}
		var params []float64
		for {
			skipWhitespaceComma()
			c := peek()
			if isValidPathCmdChar[c] || c == 0 {
				break
			}
			param, err := parseDouble()
			if err != nil {
				return err
			}
			params = append(params, param)
		}
		// It is valid to specify a multiple of the parameter number to repeat the same command.
		wantParams := int(pathCmdNumParams[cmd])
		var extraParams int
		if wantParams > 0 {
			extraParams = len(params) % wantParams
		} else {
			extraParams = len(params)
		}
		if extraParams != 0 {
			return fmt.Errorf("expected %d parameters for command %q, but got %d", pathCmdNumParams[cmd], cmd, extraParams)
		}
		if wantParams > 0 {
			for p := 0; p < len(params); p += wantParams {
				var cmdParams [7]float64
				copy(cmdParams[:], params[p:p+wantParams])
				pd.Cmds = append(pd.Cmds, PathCmd{cmd, isRel, cmdParams})
				if cmd == PathMove {
					// "M x0 y0 x1 y1 x2 y2" should draw a line from
					// the first point to the second point etc, i.e.
					// all but the first command are transformed into
					// "L".
					cmd = PathLine
				}
			}
		} else {
			pd.Cmds = append(pd.Cmds, PathCmd{cmd, isRel, [7]float64{}})
		}
	}
	return nil
}

func (pd PathData) ToString(svgMode bool) string {
	var b strings.Builder
	if !svgMode && pd.FillRule == FillNonzero {
		b.WriteString("F1 ")
	}
	var lastCmdByte byte
	for i, cmd := range pd.Cmds {
		if i != 0 {
			b.WriteString(" ")
		}
		cmdByte := cmd.Cmd.ToByte(cmd.Relative)
		if cmdByte != lastCmdByte {
			b.WriteByte(cmdByte)
		}

		for j := range pathCmdNumParams[cmd.Cmd] {
			if j != 0 {
				b.WriteString(",")
			}
			param := cmd.Params[j]
			switch {
			case math.IsInf(param, -1):
				if svgMode {
					b.WriteString("-Infinity")
				} else {
					b.WriteString("-1e308")
				}
			case math.IsInf(param, 1):
				if !svgMode {
					b.WriteString("Infinity")
				} else {
					b.WriteString("1e308")
				}
			case math.IsNaN(param):
				if !svgMode {
					b.WriteString("NaN")
				} else {
					b.WriteString("0")
				}
			default:
				fmt.Fprint(&b, param)
			}
		}
		lastCmdByte = cmdByte
	}
	return b.String()
}

// Always succeeds.
func (pd PathData) MarshalText() ([]byte, error) {
	return []byte(pd.ToString(false)), nil
}
