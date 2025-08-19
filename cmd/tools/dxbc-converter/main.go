package main

import (
	"os"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/stingray/unit/material/d3d"
)

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	parser := argparse.NewParser("dxbc-converter", "DXBC to GLSL translator for Filediver", &argparse.ParserConfig{
		DisableHelp:            false,
		DisableDefaultShowHelp: true,
		WithHint:               true,
	})

	input := parser.String("i", "input", &argparse.Option{
		Help: "Path to a dxbc file",
	})
	output := parser.String("o", "output", &argparse.Option{
		Help: "Path to write the output glsl file",
	})

	if err := parser.Parse(nil); err != nil {
		prt.Fatalf("%v", err)
	}

	f, err := os.Open(*input)
	if err != nil {
		prt.Fatalf("%v", err)
	}
	defer f.Close()

	prt.Infof("Parsing %v...", *input)
	dxbc, err := d3d.ParseDXBC(f)
	if err != nil {
		prt.Fatalf("ParseDXBC: %v", err)
	}

	out, err := os.Create(*output)
	if err != nil {
		prt.Fatalf("%v", err)
	}
	defer out.Close()

	prt.Infof("Writing %v...", *output)
	for _, opcode := range dxbc.ShaderCode.Opcodes {
		out.WriteString(opcode.ToGLSL())
	}
	prt.Infof("Done.")
}
