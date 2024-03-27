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

func hash(s string, thin bool, bigEndian bool) string {
	var pfx string
	var endian binary.ByteOrder
	if bigEndian {
		pfx = "0x"
		endian = binary.BigEndian
	} else {
		endian = binary.LittleEndian
	}

	h := stingray.Sum64([]byte(s))
	if thin {
		return pfx + h.Thin().StringEndian(endian)
	} else {
		return pfx + h.StringEndian(endian)
	}
}

func decodeHash(s string) (hash64 stingray.Hash, hash32 stingray.ThinHash, thin bool, err error) {
	endian := binary.ByteOrder(binary.LittleEndian)
	if sBE, ok := strings.CutPrefix(s, "0x"); ok {
		endian = binary.BigEndian
		s = sBE
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return stingray.Hash{}, stingray.ThinHash{}, false, err
	}
	if len(b)*2 != 8 && len(b)*2 != 16 {
		return stingray.Hash{}, stingray.ThinHash{}, false,
			fmt.Errorf("expected thin hash to be of hex length 8 (32 bytes a.k.a. thin hash), or hex length 16 (64 bytes a.k.a. normal hash), but got length %v", len(b)*2)
	}
	thin = len(b)*2 == 8
	if thin {
		hash := stingray.ThinHash{Value: endian.Uint32(b)}
		return stingray.Hash{}, hash, thin, err
	} else {
		hash := stingray.Hash{Value: endian.Uint64(b)}
		return hash, stingray.ThinHash{}, thin, err
	}
}

func crack(hashStrs []string, wordlist []string, prefixlist []string, delim string, maxNumWords int) error {
	var hashes64 []stingray.Hash
	var hashStrs64 []string
	var hashes32 []stingray.ThinHash
	var hashStrs32 []string
	for _, s := range hashStrs {
		h64, h32, thin, err := decodeHash(s)
		if err != nil {
			return err
		}
		if thin {
			hashes32 = append(hashes32, h32)
			hashStrs32 = append(hashStrs32, s)
		} else {
			hashes64 = append(hashes64, h64)
			hashStrs64 = append(hashStrs64, s)
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
	var tryCombinations func(i, wordsLeft int, firstWord bool)
	tryCombinations = func(i, wordsLeft int, firstWord bool) {
		if wordsLeft == 0 {
			for idx, h := range hashes32 {
				if stingray.Sum64(buf[:i]).Thin() == h {
					fmt.Printf("String found: %v = \"%v\"\n", hashStrs32[idx], string(buf[:i]))
				}
			}
			for idx, h := range hashes64 {
				if stingray.Sum64(buf[:i]) == h {
					fmt.Printf("String found: %v = \"%v\"\n", hashStrs64[idx], string(buf[:i]))
				}
			}
			return
		}

		if !firstWord {
			ensureBufLen(i + len(delim))
			copy(buf[i:], delim)
			i += len(delim)
		}

		if i == 0 {
			for _, prefix := range prefixlist {
				ensureBufLen(i + len(prefix))
				copy(buf[i:], prefix)

				tryCombinations(i+len(prefix), wordsLeft, firstWord)
			}
		}

		for _, word := range wordlist {
			ensureBufLen(i + len(word))
			copy(buf[i:], word)

			tryCombinations(i+len(word), wordsLeft-1, false)
		}
	}
	for numWords := 1; numWords <= maxNumWords; numWords++ {
		fmt.Printf("Trying %v words\n", numWords)
		tryCombinations(0, numWords, true)
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
	parser := argparse.NewParser("filediver_hash_tool", "Simple tool for calculating and cracking murmur64a hashes.", &argparse.ParserConfig{
		EpiLog: `Without prefix, input hashes are considered little endian (e.g. ddafccccf2172e9e).
With "0x" prefix, hashes are considered big endian (e.g. 0x9e2e17f2ccccafdd).
Hashes may be of length 32-bit (hex length 8, a.k.a. thin hash) or 64-bit (hex length 16, a.k.a. normal hash).
Different hash lengths and endianesses may be mixed in the input.`,
	})
	thin := parser.Flag("t", "thin", &argparse.Option{Help: "Output \"thin\" 32-bit hashes instead of 64-bit"})
	bigEndian := parser.Flag("b", "big_endian", &argparse.Option{Help: "Output hashes in big endian"})
	inputStrs := parser.Strings("", "input", &argparse.Option{Positional: true, Help: "Strings to hash / hashes to crack (see epilog)"})
	inputPath := parser.String("i", "input_file", &argparse.Option{Help: "Path to file containing strings to hash / hashes to crack"})
	modeCrack := parser.Flag("c", "crack", &argparse.Option{Help: "Attempt to crack a hash using an optional word list and brute-force"})
	wordlistPath := parser.String("w", "wordlist", &argparse.Option{Help: "Path to word list file"})
	prefixlistPath := parser.String("p", "prefixlist", &argparse.Option{Help: "Path to prefix list file (e.g. for known directories)"})
	maxWords := parser.Int("n", "max_words", &argparse.Option{Help: "Maximum number of words to try in a sequence", Default: "-1"})
	delim := parser.String("d", "delimiter", &argparse.Option{Help: "Delimiter to separate words by (default: \"_\")", Default: "_"})
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
		if *thin {
			fmt.Fprintln(os.Stderr, "\"thin\" option only available for \"hash\" mode")
			os.Exit(1)
		}

		if *bigEndian {
			fmt.Fprintln(os.Stderr, "\"big_endian\" option only available for \"hash\" mode")
			os.Exit(1)
		}

		if len(*inputStrs) > 0 {
			fmt.Fprintln(os.Stderr, "can only use one of \"input\" or \"input_file\"")
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
		if *wordlistPath == "" {
			fmt.Fprintln(os.Stderr, "need \"wordlist\" for \"crack\" mode")
			os.Exit(1)
		}

		wordlist, err := fileToStrings(*wordlistPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var prefixlist []string
		if *prefixlistPath != "" {
			var err error
			prefixlist, err = fileToStrings(*prefixlistPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		if *maxWords == -1 {
			*maxWords = 6
		}

		if err := crack(*inputStrs, wordlist, prefixlist, *delim, *maxWords); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if *wordlistPath != "" {
			fmt.Fprintln(os.Stderr, "\"wordlist\" option only available for \"crack\" mode")
			os.Exit(1)
		}

		if *maxWords != -1 {
			fmt.Fprintln(os.Stderr, "\"max_words\" option only available for \"crack\" mode")
			os.Exit(1)
		}

		if len(*inputStrs) == 1 {
			fmt.Println(hash((*inputStrs)[0], *thin, *bigEndian))
		} else {
			for _, s := range *inputStrs {
				fmt.Printf("\"%v\" = %v\n", s, hash(s, *thin, *bigEndian))
			}
		}
	}
}
