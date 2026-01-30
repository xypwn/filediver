package main

import (
	"bytes"
	"fmt"
	"io"
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

	parser := argparse.NewParser("dxbc-dumper", "DXBC dumper for Filediver", &argparse.ParserConfig{
		DisableHelp:            false,
		DisableDefaultShowHelp: true,
		WithHint:               true,
	})

	input := parser.String("i", "input", &argparse.Option{
		Help: "Path to a material.gpu file",
	})
	output := parser.String("o", "output", &argparse.Option{
		Help: "Folder path to write the output dxbc/glsl file(s)",
	})
	glsl := parser.Flag("g", "glsl", &argparse.Option{
		Help: "Write GLSL files rather than DXBC files",
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
	data, err := io.ReadAll(f)
	if err != nil {
		prt.Fatalf("%v", err)
	}
	var offset int = 0
	fileCountMap := make(map[d3d.ShaderProgramType]int)
	suffix := "dxbc"
	if *glsl {
		suffix = "glsl"
	}
	for true {
		if offset >= len(data) {
			break
		}
		idx := bytes.Index(data[offset:], d3d.MAGIC[:])
		if idx < 0 {
			break
		}
		dxbc, err := d3d.ParseDXBC(bytes.NewReader(data[offset+idx:]))
		if err != nil {
			prt.Errorf("ParseDXBC: %v (%#08x)", err, idx)
			idx = idx + 4
			break
		}

		if _, ok := fileCountMap[dxbc.ShaderCode.ProgramType]; !ok {
			fileCountMap[dxbc.ShaderCode.ProgramType] = 0
		}
		filename := *output + string(os.PathSeparator) + fmt.Sprintf("%v.%v", fileCountMap[dxbc.ShaderCode.ProgramType], suffix)

		switch dxbc.ShaderCode.ProgramType {
		case d3d.PIXEL_SHADER:
			filename += ".frag"
		case d3d.VERTEX_SHADER:
			filename += ".vert"
		case d3d.GEOMETRY_SHADER:
			filename += ".geom"
		case d3d.HULL_SHADER:
			filename += ".tesc"
		case d3d.DOMAIN_SHADER:
			filename += ".tese"
		case d3d.COMPUTE_SHADER:
			filename += ".comp"
		}

		prt.Infof("Offset %#08x", offset+idx)
		prt.Infof("Version %v.%v", dxbc.ShaderCode.Version.Major, dxbc.ShaderCode.Version.Minor)
		prt.Infof("Program type %v", dxbc.ShaderCode.ProgramType)

		out, err := os.Create(filename)
		if os.IsNotExist(err) {
			err = os.Mkdir(*output, os.ModeDir|os.ModePerm)
			if err == nil {
				out, err = os.Create(filename)
			}
		}
		if err != nil {
			prt.Fatalf("%v", err)
		}
		defer out.Close()

		prt.Infof("Writing %v...", filename)
		if *glsl {
			out.WriteString(dxbc.ToGLSL())
		} else {
			out.Write(data[offset+idx : offset+idx+int(dxbc.Size)])
		}
		offset = offset + idx + int(dxbc.Size)
		fileCountMap[dxbc.ShaderCode.ProgramType]++
		prt.Infof("")
	}

	prt.Infof("Done.")
}
