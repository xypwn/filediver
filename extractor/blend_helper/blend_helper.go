package blend_helper

import (
	"fmt"
	"os"
	os_exec "os/exec"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/exec"
)

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
