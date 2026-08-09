package main

import (
	"bytes"
	"maps"
	"os"
	"slices"
)

// Sorts the given hashes file and removes duplicates.
// Puts all comments to the top of the file in their original order.
func sortHashFile(path string) error {
	var comments []string
	hashes := make(map[string]struct{})

	{
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for line := range bytes.SplitSeq(b, []byte("\n")) {
			line = bytes.TrimSuffix(line, []byte("\r"))
			if bytes.HasPrefix(line, []byte("#")) || bytes.HasPrefix(line, []byte("//")) {
				comments = append(comments, string(line))
				continue
			}
			if len(line) == 0 {
				continue
			}
			hashes[string(line)] = struct{}{}
		}
	}

	{
		var b bytes.Buffer
		for _, s := range comments {
			b.WriteString(s)
			b.WriteString("\n")
		}
		for _, s := range slices.Sorted(maps.Keys(hashes)) {
			b.WriteString(s)
			b.WriteString("\n")
		}
		if err := os.WriteFile(path, b.Bytes(), 0666); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := sortHashFile("cracked.txt"); err != nil {
		panic(err)
	}
}
