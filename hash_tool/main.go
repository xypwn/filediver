package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/hellflame/argparse"

	"github.com/xypwn/filediver/stingray"
)

func hash(s string, thin bool) string {
	h := stingray.Sum64([]byte(s))
	if thin {
		return h.Thin().String()
	} else {
		return h.String()
	}
}

// var charset = []byte{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "_"}
var wordlistChars = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "_"}

func crack(hashStrs []string, thin bool, wordlist []string, delim string, maxNumWords int) error {
	var hashBs [][]byte
	for _, s := range hashStrs {
		b, err := hex.DecodeString(s)
		if err != nil {
			return err
		}
		hashBs = append(hashBs, b)
		if thin {
			if len(b)*2 != 8 {
				return fmt.Errorf("expected thin hash to be of hex length 8, but got length %v", len(b)*2)
			}
		} else {
			if len(b)*2 != 16 {
				return fmt.Errorf("expected hash to be of hex length 16, but got length %v", len(b)*2)
			}
		}
	}
	var findMatch func([]byte) int
	if thin {
		var hashes []stingray.ThinHash
		for _, b := range hashBs {
			hashes = append(hashes, stingray.ThinHash{Value: binary.LittleEndian.Uint32(b)})
		}
		findMatch = func(b []byte) int {
			for i, h := range hashes {
				if stingray.Sum64(b).Thin() == h {
					return i
				}
			}
			return -1
		}
	} else {
		var hashes []stingray.Hash
		for _, b := range hashBs {
			hashes = append(hashes, stingray.Hash{Value: binary.LittleEndian.Uint64(b)})
		}
		findMatch = func(b []byte) int {
			for i, h := range hashes {
				if stingray.Sum64(b) == h {
					return i
				}
			}
			return -1
		}
	}

	var buf []byte
	ensureBufLen := func(cap int) {
		if cap > len(buf) {
			newBuf := make([]byte, cap*2)
			copy(newBuf, buf)
			buf = newBuf
		}
	}
	var tryCombinations func(i, wordsLeft int)
	tryCombinations = func(i, wordsLeft int) {
		if wordsLeft == 0 {
			if idx := findMatch(buf[:i]); idx != -1 {
				fmt.Printf(
					"String found: %v = \"%v\"\n",
					hex.EncodeToString(hashBs[idx]),
					string(buf[:i]),
				)
			}
			return
		}

		if i != 0 {
			ensureBufLen(i + len(delim))
			copy(buf[i:], delim)
			i += len(delim)
		}

		for _, word := range wordlist {
			ensureBufLen(i + len(word))
			copy(buf[i:], word)

			tryCombinations(i+len(word), wordsLeft-1)
		}
	}
	for numWords := 1; numWords <= maxNumWords; numWords++ {
		fmt.Printf("Trying %v words\n", numWords)
		tryCombinations(0, numWords)
	}

	return nil
}

func fileToStrings(path string) ([]string, error) {
	var res []string
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		s := sc.Text()
		if s != "" && !strings.HasPrefix(s, "//") {
			res = append(res, s)
		}
	}
	if sc.Err() != nil {
		return nil, err
	}
	return res, nil
}

func main() {
	parser := argparse.NewParser("filediver_hash_tool", "Simple tool for calculating and cracking murmur64a hashes.", nil)
	thin := parser.Flag("t", "thin", &argparse.Option{Help: "Use \"thin\" 32-bit hashes instead of 64-bit"})
	inputStrs := parser.Strings("", "input", &argparse.Option{Positional: true, Help: "Strings to hash / hashes to crack"})
	inputPath := parser.String("i", "input_file", &argparse.Option{Help: "Path to file containing strings to hash / hashes to crack"})
	modeCrack := parser.Flag("c", "crack", &argparse.Option{Help: "Attempt to crack a hash using an optional word list and brute-force"})
	wordlistPath := parser.String("w", "wordlist", &argparse.Option{Help: "Path to word list file"})
	maxWords := parser.Int("n", "max_words", &argparse.Option{Help: "Maximum number of words to try in a sequence", Default: "-1"})
	delim := parser.String("d", "delimiter", &argparse.Option{Help: "Delimiter to separate words by (default: none)", Default: ""})
	if err := parser.Parse(nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		}
		os.Exit(1)
	}
	if *maxWords <= 0 && *maxWords != -1 {
		fmt.Fprintln(os.Stderr, "max_words must be at least 1")
		os.Exit(1)
	}
	if *inputPath != "" {
		if len(*inputStrs) > 0 {
			fmt.Fprintln(os.Stderr, "can only use one of input or input_file")
			os.Exit(1)
		}
		var err error
		*inputStrs, err = fileToStrings(*inputPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if *modeCrack {
		wordlist := wordlistChars

		if *wordlistPath != "" {
			var err error
			wordlist, err = fileToStrings(*wordlistPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		if *maxWords == -1 {
			*maxWords = 6
		}

		if err := crack(*inputStrs, *thin, wordlist, *delim, *maxWords); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if *wordlistPath != "" {
			fmt.Fprintln(os.Stderr, "wordlist only available for \"crack\" mode")
			os.Exit(1)
		}

		if *maxWords != -1 {
			fmt.Fprintln(os.Stderr, "max_words only available for \"crack\" mode")
			os.Exit(1)
		}

		if len(*inputStrs) == 1 {
			fmt.Println(hash((*inputStrs)[0], *thin))
		} else {
			for _, s := range *inputStrs {
				fmt.Printf("\"%v\" = %v\n", s, hash(s, *thin))
			}
		}
	}
}
