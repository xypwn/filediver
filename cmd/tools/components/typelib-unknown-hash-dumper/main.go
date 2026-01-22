package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	datalib "github.com/xypwn/filediver/datalibrary"
)

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	parser := argparse.NewParser("typelib-json-dumper", "", nil)
	path := parser.String("", "path", &argparse.Option{
		Help:       "The path to the .dl_typelib to dump",
		Positional: true,
	})
	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("parser: %v", err)
	}

	typelibFile, err := os.OpenFile(*path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		prt.Fatalf("open %v: %v", *path, err)
	}

	typelibData, err := io.ReadAll(typelibFile)
	if err != nil {
		prt.Fatalf("readall %v: %v", *path, err)
	}

	typelib, err := datalib.ParseTypeLib(typelibData)
	if err != nil {
		prt.Fatalf("parse typelib %v: %v", *path, err)
	}

	for hash := range typelib.Types {
		if _, ok := datalib.DLHashesToStrings[hash]; ok {
			continue
		}
		fmt.Printf("%08x\n", uint32(hash))
	}

	for hash := range typelib.Enums {
		if _, ok := datalib.DLHashesToStrings[hash]; ok {
			continue
		}
		fmt.Printf("%08x\n", uint32(hash))
	}
}
