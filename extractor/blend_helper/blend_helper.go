package blend_helper

import (
	"bytes"
	"fmt"
	"os"
	os_exec "os/exec"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/exec"
)

type ScriptExitError struct {
	*os_exec.ExitError
	Stdout []byte
	Stderr []byte
}

// FullError returns a human-readable version with additional
// context. Contains line breaks (which regular Go error messages shouldn't).
func (e *ScriptExitError) FullError() string {
	return fmt.Sprintf("Blender exporter failed to execute: %v\n== Stdout Log==\n%s== Stderr Log ==\n%s", e.ExitError, e.Stdout, e.Stderr)
}

// ExportBlend runs the blender exporter script.
//
// May return a [ScriptExitError], which indicates that the script itself
// failed to execute, along with stdout and stderr for tracing.
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
	var stdout, stderr bytes.Buffer
	cmd, err := runner.Start(blendExporter, &stdout, &stderr, read, "-", outPath)
	if err != nil {
		return err
	}
	defer func() {
		err = cmd.Wait()
		if exiterr, ok := err.(*os_exec.ExitError); ok {
			if exiterr.ExitCode() == 0xC0000005 {
				err = nil
			} else {
				err = &ScriptExitError{
					ExitError: exiterr,
					Stdout:    stdout.Bytes(),
					Stderr:    stderr.Bytes(),
				}
			}
		}
	}()
	if err := enc.Encode(doc); err != nil {
		return err
	}

	return nil
}
