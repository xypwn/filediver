package single_glb_helper

import (
	"context"
	"path/filepath"

	"github.com/qmuntal/gltf"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/exec"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/ah_bin"
)

func CreateCloseableGltfDocument(ctx context.Context, statusf func(format string, args ...any), outDir string, name string, format string, runner *exec.Runner, buildInfo *ah_bin.BuildInfo) (*gltf.Document, func(doc *gltf.Document) error) {
	document := gltf.NewDocument()
	document.Asset.Generator = "https://github.com/xypwn/filediver"
	if buildInfo != nil {
		document.Scenes[0].Extras = map[string]any{"Helldivers 2 Version": buildInfo.Version}
	}
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
		closeCtx, _ := extractor.NewContext(
			ctx,
			stingray.NewFileID(stingray.Hash{}, stingray.Hash{}),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			buildInfo,
			nil,
			nil,
			runner,
			appconfig.Config{},
			outPath,
			nil,
			nil,
			statusf,
		)
		extractor.SaveDocument(closeCtx, doc, "combined", format)
		return nil
	}
	return document, closeGLB
}
