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

func ExtractFuncRaw(suffix string, types ...stingray.DataType) ExtractFunc {
	return func(ctx Context) error {
		r, err := ctx.File().OpenMulti(ctx.Ctx(), types...)
		if err != nil {
			return err
		}
		defer r.Close()

		out, err := ctx.CreateFile(suffix)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, r); err != nil {
			return err
		}
		return nil
	}
}

func ExportBlend(doc *gltf.Document, outPath string, runner *exec.Runner) (err error) {
	read, write, err := os.Pipe()
	if err != nil {
		return err
	}
	var blendExporter string = "scripts_dist/hd2_accurate_blender_importer/hd2_accurate_blender_importer"
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
