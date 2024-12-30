package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
)

const prolog = `// Copied from Hellextractor by Xaymar (https://github.com/Xaymar/Hellextractor),
// hd2-name-db by DTZxPorter (https://github.com/dtzxporter/hd2-name-db) and
// Helldivers 2 Hashlist by HW12Dev (https://github.com/HW12Dev/Helldivers2-Hashlist).
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
	// This one includes a lot of garbage (probably a lot is scraped from strings), but it does improve hash coverage by a few percent.
	if err := appendHTTPFile(&strs, "https://raw.githubusercontent.com/HW12Dev/Helldivers2-Hashlist/main/hashlist.txt", func(s string) string {
		lhs, _, _ := strings.Cut(s, ".")
		return lhs
	}); err != nil {
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
	if err := appendHTTPFile(&strs, "https://raw.githubusercontent.com/xypwn/filediver/master/hashes/cracked.txt", nil); err != nil {
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
