package extractor

import (
	"context"
	"fmt"
	"io"
	"os"
	os_exec "os/exec"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/stingray"
	dlbin "github.com/xypwn/filediver/stingray/dl_bin"
)

type Context interface {
	Ctx() context.Context
	File() *stingray.File
	Runner() *exec.Runner
	Config() map[string]string
	GetResource(name, typ stingray.Hash) (file *stingray.File, exists bool)
	// Call WriteCloser.Close() when done.
	CreateFile(suffix string) (io.WriteCloser, error)
	// Returns path to file.
	AllocateFile(suffix string) (string, error)
	Hashes() map[stingray.Hash]string
	ThinHashes() map[stingray.ThinHash]string
	// Selected triad ID, if any (-t option).
	TriadIDs() []stingray.Hash
	ArmorSets() map[stingray.Hash]dlbin.ArmorSet
	// Prints a warning message.
	Warnf(f string, a ...any)
}

type ExtractFunc func(ctx Context) error

func extractByType(ctx Context, typ stingray.DataType, extension string) error {
	r, err := ctx.File().Open(ctx.Ctx(), typ)
	if err != nil {
		return err
	}
	defer r.Close()

	out, err := ctx.CreateFile(fmt.Sprintf(".%v.%v", extension, typ.Extension()))
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return err
	}

	return nil
}

func extractCombined(ctx Context, extension string) error {
	if !(ctx.File().Exists(stingray.DataMain) || ctx.File().Exists(stingray.DataStream) || ctx.File().Exists(stingray.DataGPU)) {
		return fmt.Errorf("extractCombined: no data to extract for file")
	}
	out, err := ctx.CreateFile(fmt.Sprintf(".%v", extension))
	if err != nil {
		return err
	}
	defer out.Close()

	for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
		if !ctx.File().Exists(typ) {
			continue
		}
		r, err := ctx.File().Open(ctx.Ctx(), typ)
		if err != nil {
			return err
		}
		defer r.Close()

		if _, err := io.Copy(out, r); err != nil {
			return err
		}
	}

	return nil
}

func ExtractFuncRaw(extension string) ExtractFunc {
	return func(ctx Context) error {
		for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
			if ctx.File().Exists(typ) {
				if err := extractByType(ctx, typ, extension); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func ExtractFuncRawSingleType(extension string, typ stingray.DataType) ExtractFunc {
	return func(ctx Context) error {
		if ctx.File().Exists(typ) {
			return extractByType(ctx, typ, extension)
		}
		return fmt.Errorf("No %v data found", typ.String())
	}
}

func ExtractFuncRawCombined(extension string) ExtractFunc {
	return func(ctx Context) error {
		return extractCombined(ctx, extension)
	}
}

func ExportBlend(doc *gltf.Document, outPath string, runner *exec.Runner) (err error) {
	read, write, err := os.Pipe()
	if err != nil {
		return err
	}
	var blendExporter string = "hd2_accurate_blender_importer"
	if !runner.Has(blendExporter) {
		return fmt.Errorf("cannot export as .blend: \"%v\" missing", blendExporter)
	}
	enc := gltf.NewEncoder(write)
	path := outPath + ".blend"
	cmd, err := runner.Start(blendExporter, nil, nil, read, "-", path)
	if err != nil {
		return err
	}
	defer func() {
		err = cmd.Wait()
		if exiterr, ok := err.(*os_exec.ExitError); ok && exiterr.ExitCode() == 0xC0000005 {
			err = nil
		}
	}()
	if err := enc.Encode(doc); err != nil {
		return err
	}

	return nil
}
