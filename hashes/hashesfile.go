package hashes

import (
	"bytes"
	"os"
	"slices"

	"github.com/xypwn/filediver/util"
)

func parseHashes(b []byte) (hashes []string, comments []string) {
	for line := range bytes.SplitSeq(b, []byte("\n")) {
		line = bytes.TrimSuffix(line, []byte("\r"))
		if bytes.HasPrefix(line, []byte("#")) || bytes.HasPrefix(line, []byte("//")) {
			comments = append(comments, string(line))
			continue
		}
		if len(line) == 0 {
			continue
		}
		hashes = append(hashes, string(line))
	}
	return
}

// Parses the given hashes file. Returns all non-empty
// non-commented lines.
//
// Lines are considered commented if they start with "#" or "//".
func ParseHashes(s string) (hashes []string) {
	hashes, _ = parseHashes([]byte(s))
	return
}

// Sorts the given hashes file and removes duplicates.
//
// Puts all comments to the top of the file in their original order.
//
// See [ParseHashes].
func TidyHashes(b []byte) []byte {
	var comments []string

	hashes, comments := parseHashes(b)
	slices.Sort(hashes)
	util.Uniq(hashes)

	var buf bytes.Buffer
	buf.Grow(len(b))
	for _, s := range comments {
		buf.WriteString(s)
		buf.WriteString("\n")
	}
	for _, s := range hashes {
		buf.WriteString(s)
		buf.WriteString("\n")
	}
	return buf.Bytes()
}

// Like [TidyHashes], but updates a file by path in-place. Returns the new file contents.
func TidyHashesFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b = TidyHashes(b)
	if err := os.WriteFile(path, b, info.Mode()); err != nil {
		return nil, err
	}
	return b, nil
}
