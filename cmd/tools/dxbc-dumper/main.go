package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/hellflame/argparse"
	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/material/d3d"
	d3dops "github.com/xypwn/filediver/stingray/unit/material/d3d/opcodes"
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

	var knownThinHashes []string
	knownThinHashes = append(knownThinHashes, app.ParseHashes(hashes.ThinHashes)...)

	thinHashesMap := make(map[stingray.ThinHash]string)
	for _, h := range knownThinHashes {
		thinHashesMap[stingray.Sum(h).Thin()] = h
	}

	lookupThinHash := func(hash stingray.ThinHash) string {
		val, ok := thinHashesMap[hash]
		if ok {
			return val
		}
		return hash.String()
	}

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
	fileCountMap := make(map[d3dops.ShaderProgramType]int)
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
		reader := bytes.NewReader(data[offset+idx:])
		dxbc, err := d3d.ParseDXBC(reader)
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
		case d3dops.PIXEL_SHADER:
			filename += ".frag"
		case d3dops.VERTEX_SHADER:
			filename += ".vert"
		case d3dops.GEOMETRY_SHADER:
			filename += ".geom"
		case d3dops.HULL_SHADER:
			filename += ".tesc"
		case d3dops.DOMAIN_SHADER:
			filename += ".tese"
		case d3dops.COMPUTE_SHADER:
			filename += ".comp"
		}

		var thinHash stingray.ThinHash
		err = binary.Read(reader, binary.LittleEndian, &thinHash)
		if err == nil {
			fmt.Printf("%v - %v - %v\n", lookupThinHash(thinHash), filename, dxbc.ShaderCode.ProgramType.ToString())
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
