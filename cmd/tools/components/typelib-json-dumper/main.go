package main

import (
	"encoding/json"
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

	if typelib.TypeInfoStringsSize == 0 {
		prt.Warnf("No strings contained in type library! (AH removed strings after 24-02-23, so any dl_library from after that date will not have type names in it)")
		prt.Warnf("Output will only have known hashed strings, other strings will have their offsets output")
	}

	result, err := json.MarshalIndent(typelib, "", "    ")
	if err != nil {
		prt.Fatalf("marshal typelib %v as json: %v", *path, err)
	}
	fmt.Println(string(result))
}
