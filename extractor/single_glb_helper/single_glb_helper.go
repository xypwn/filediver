package single_glb_helper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor/blend_helper"
)

func CreateCloseableGltfDocument(outDir string, name string, formatBlend bool, runner *exec.Runner) (*gltf.Document, func(doc *gltf.Document) error) {
	document := gltf.NewDocument()
	document.Asset.Generator = "https://github.com/xypwn/filediver"
	document.Samplers = append(document.Samplers, &gltf.Sampler{
		MagFilter: gltf.MagLinear,
		MinFilter: gltf.MinLinear,
		WrapS:     gltf.WrapRepeat,
		WrapT:     gltf.WrapRepeat,
	})
	closeGLB := func(doc *gltf.Document) error {
		outPath := filepath.Join(outDir, name)
		if len(document.Buffers) == 0 {
			return nil
		}
		if formatBlend {
			err := blend_helper.ExportBlend(doc, outPath+".blend", runner)
			if err != nil {
				return fmt.Errorf("closing %v.blend: %w", outPath, err)
			}
		} else {
			err := exportGLB(doc, outPath+".glb")
			if err != nil {
				return fmt.Errorf("closing %v.glb: %w", outPath, err)
			}
		}
		return nil
	}
	return document, closeGLB
}

func exportGLB(doc *gltf.Document, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	enc := gltf.NewEncoder(out)
	if err := enc.Encode(doc); err != nil {
		return err
	}
	return nil
}
