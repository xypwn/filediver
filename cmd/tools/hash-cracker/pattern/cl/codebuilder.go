package cl

import (
	"bytes"
	"fmt"
	"strings"
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
			if ln[i] == '%' {
				if ln[i+1] != '%' && ln[i+1] != ' ' {
					argn++
				}
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
