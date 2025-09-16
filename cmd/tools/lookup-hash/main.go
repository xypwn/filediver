package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
)

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	parser := argparse.NewParser("lookup-hash", "", nil)
	thin := parser.Flag("t", "thin", &argparse.Option{
		Help: "Given hash is 32bit",
	})
	base := parser.Int("b", "base", &argparse.Option{
		Help:    "Base the hash is given in",
		Choices: []any{2, 8, 10, 16},
		Default: "16",
	})
	hash := parser.String("", "hash", &argparse.Option{
		Help:       "The hash to look up",
		Positional: true,
	})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("%v", err)
	}

	bitsize := 64
	if *thin {
		bitsize = 32
	}

	value, err := strconv.ParseUint(*hash, *base, bitsize)
	if err != nil {
		prt.Fatalf("Could not parse hash %v: %v", *hash, err)
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	hashesMap := make(map[stingray.Hash]string)
	for _, h := range knownHashes {
		hashesMap[stingray.Sum(h)] = h
	}
	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range knownThinHashes {
		thinHashesMap[stingray.Sum(h).Thin()] = h
	}

	lookupThinHash := func(hash stingray.ThinHash) string {
		if name, ok := thinHashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	lookupHash := func(hash stingray.Hash) string {
		if name, ok := hashesMap[hash]; ok {
			return name
		}
		return hash.String()
	}

	prt.NoStatus()

	var toPrint string
	if *thin {
		toPrint = lookupThinHash(stingray.ThinHash{Value: uint32(value)})
	} else {
		toPrint = lookupHash(stingray.Hash{Value: value})
	}
	fmt.Println(toPrint)
}
