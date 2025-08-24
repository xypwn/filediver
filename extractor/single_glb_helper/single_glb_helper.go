package single_glb_helper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor/blend_helper"
)

func CreateCloseableGltfDocument(outDir string, triad string, formatBlend bool, runner *exec.Runner) (*gltf.Document, func(doc *gltf.Document) error) {
	document := gltf.NewDocument()
	document.Asset.Generator = "https://github.com/xypwn/filediver"
	document.Samplers = append(document.Samplers, &gltf.Sampler{
		MagFilter: gltf.MagLinear,
		MinFilter: gltf.MinLinear,
		WrapS:     gltf.WrapRepeat,
		WrapT:     gltf.WrapRepeat,
	})
	closeGLB := func(doc *gltf.Document) error {
		outPath := filepath.Join(outDir, triad)
		if len(document.Buffers) == 0 {
			return nil
		}
		if formatBlend {
			err := blend_helper.ExportBlend(doc, outPath, runner)
			if err != nil {
				return fmt.Errorf("closing %v.blend: %v", outPath, err)
			}
		} else {
			err := exportGLB(doc, outPath)
			if err != nil {
				return fmt.Errorf("closing %v.glb: %v", outPath, err)
			}
		}
		return nil
	}
	return document, closeGLB
}

func exportGLB(doc *gltf.Document, outPath string) error {
	path := outPath + ".glb"
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := gltf.NewEncoder(out)
	if err := enc.Encode(doc); err != nil {
		return err
	}
	return nil
}
