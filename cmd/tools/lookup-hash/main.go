package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

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

	a := app.GenerateHashes(ctx, knownHashes, knownThinHashes, nil)
	prt.NoStatus()

	var toPrint string
	if *thin {
		toPrint = a.LookupThinHash(stingray.ThinHash{Value: uint32(value)})
	} else {
		toPrint = a.LookupHash(stingray.Hash{Value: value})
	}
	fmt.Println(toPrint)
}
