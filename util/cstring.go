package util

import (
	"io"
	"strings"
)

func readCString(r io.Reader, skipLeadingZeroes bool) (string, error) {
	var s strings.Builder
	s.Grow(12) // reasonable guesstimate to minimize reallocations
	var b [1]byte
	var n int
	var err error
	for {
		n, err = r.Read(b[:])
		if n >= 1 {
			if b[0] == 0 {
				if skipLeadingZeroes && s.Len() == 0 {
					// Leading zero found => ignore if requested
					continue
				}
				// Null terminator found
				break
			}
			s.WriteByte(b[0])
		}
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		return s.String(), io.ErrUnexpectedEOF
	} else if err != nil {
		return "", err
	}
	return s.String(), nil
}

// Reads a null-terminated string from r (including null terminator)
// and returns it as a Go string (without null terminator).
//
// If the reader hits EOF before a null terminator, returns ErrUnexpectedEOF
// along with the so far read string.
//
// Returns an empty string and error if any other error is encountered.
func ReadCString(r io.Reader) (string, error) {
	return readCString(r, false)
}

// Like [ReadCString], but skips any leading null bytes.
func ReadCStringSkipLeadingNulls(r io.Reader) (string, error) {
	return readCString(r, true)
}
