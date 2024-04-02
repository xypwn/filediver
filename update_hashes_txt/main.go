package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
)

const prolog = `// Copied from Hellextractor by Xaymar (https://github.com/Xaymar/Hellextractor)
// as well as from hd2-name-db by DTZxPorter (https://github.com/dtzxporter/hd2-name-db)
// This is a list of all known file type and file path strings.
// Their respective hashes will be generated automatically.
`

func appendHTTPFile(strs *map[string]struct{}, url string, transform func(string) string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rd := bufio.NewScanner(resp.Body)
	for rd.Scan() {
		s := rd.Text()
		if len(s) == 0 || strings.HasPrefix(s, "//") || strings.HasPrefix(s, "#") {
			continue
		}
		if transform != nil {
			s = transform(s)
		}
		(*strs)[s] = struct{}{}
	}
	return nil
}

func main() {
	strs := make(map[string]struct{})
	if err := appendHTTPFile(&strs, "https://raw.githubusercontent.com/Xaymar/Hellextractor/root/files.txt", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := appendHTTPFile(&strs, "https://raw.githubusercontent.com/Xaymar/Hellextractor/root/types.txt", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := appendHTTPFile(&strs, "https://raw.githubusercontent.com/dtzxporter/hd2-name-db/main/assets.txt", func(s string) string {
		_, rhs, _ := strings.Cut(s, ",")
		return rhs
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out, err := os.Create("hashes.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := out.Write([]byte(prolog)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	strsSorted := make([]string, 0, len(strs))
	for k := range strs {
		strsSorted = append(strsSorted, k)
	}
	slices.Sort(strsSorted)
	for _, line := range strsSorted {
		if _, err := fmt.Fprintln(out, line); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
