package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/hellflame/argparse"

	"github.com/xypwn/filediver/stingray"
)

func hash(s string, thin bool) {
	h := stingray.Sum64([]byte(s))
	if thin {
		fmt.Println(h.Thin())
	} else {
		fmt.Println(h)
	}
}

// var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
var charset = []byte("abcdefghijklmnopqrstuvwxyz0123456789_")

func crack(hashStr string, thin bool, wordlist []string, bruteForceMaxLen int) error {
	hashB, err := hex.DecodeString(hashStr)
	if err != nil {
		return err
	}
	var matches func([]byte) bool
	if thin {
		if len(hashB)*2 != 8 {
			return fmt.Errorf("expected thin hash to be of hex length 8, but got length %v", len(hashB)*2)
		}
		cmp := stingray.ThinHash{Value: binary.LittleEndian.Uint32(hashB)}
		matches = func(b []byte) bool {
			return stingray.Sum64(b).Thin() == cmp
		}
	} else {
		if len(hashB)*2 != 16 {
			return fmt.Errorf("expected hash to be of hex length 16, but got length %v", len(hashB)*2)
		}
		cmp := stingray.Hash{Value: binary.LittleEndian.Uint64(hashB)}
		matches = func(b []byte) bool {
			return stingray.Sum64(b) == cmp
		}
	}
	var tryCombinations func(b []byte, n int) bool
	tryCombinations = func(b []byte, n int) bool {
		for _, c := range charset {
			b[n] = c
			if n > 0 {
				if !tryCombinations(b, n-1) {
					return false
				}
			} else {
				if matches(b) {
					fmt.Printf("String found: \"%v\"!\n", string(b))
				}
			}
		}
		return true
	}

	buf := make([]byte, bruteForceMaxLen)
	for length := 1; length <= bruteForceMaxLen; length++ {
		fmt.Printf("Trying length %v\n", length)
		if !tryCombinations(buf[:length], length-1) {
			break
		}
	}

	return nil
}

func main() {
	parser := argparse.NewParser("filediver_hash_tool", "Simple tool for calculating and cracking murmur64a hashes.", nil)
	thin := parser.Flag("t", "thin", &argparse.Option{Help: "Use \"thin\" 32-bit hashes instead of 64-bit"})
	inputStr := parser.String("", "input", &argparse.Option{Positional: true, Help: "String to hash / hash to crack"})
	modeCrack := parser.Flag("c", "crack", &argparse.Option{Help: "Attempt to crack a hash using an optional word list and brute-force"})
	wordlistPath := parser.String("w", "wordlist", &argparse.Option{Help: "Path to word list file"})
	bruteForceMaxLen := parser.Int("l", "brute_force_length", &argparse.Option{Help: "Maximum string length for brute-force cracking", Default: "-1"})
	if err := parser.Parse(nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if err == argparse.BreakAfterHelpError {
			os.Exit(0)
		}
		os.Exit(1)
	}
	if *bruteForceMaxLen <= 0 && *bruteForceMaxLen != -1 {
		fmt.Fprintln(os.Stderr, "brute_force_length must be at least 1")
		os.Exit(1)
	}
	if *modeCrack {
		if *wordlistPath != "" {
			fmt.Fprintln(os.Stderr, "wordlist not yet implemented")
			os.Exit(1)
		}

		if *bruteForceMaxLen == -1 {
			*bruteForceMaxLen = 6
		}

		if err := crack(*inputStr, *thin, nil, *bruteForceMaxLen); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if *wordlistPath != "" {
			fmt.Fprintln(os.Stderr, "wordlist only available for \"crack\" mode")
			os.Exit(1)
		}

		if *bruteForceMaxLen != -1 {
			fmt.Fprintln(os.Stderr, "brute_force_length only available for \"crack\" mode")
			os.Exit(1)
		}

		hash(*inputStr, *thin)
	}
}
