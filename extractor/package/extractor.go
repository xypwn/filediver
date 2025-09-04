package stingray_package

import (
	"encoding/json"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_package "github.com/xypwn/filediver/stingray/package"
)

func ExtractPackageJSON(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	pkgData, err := stingray_package.LoadPackage(r)
	if err != nil {
		return err
	}
	outData := make([]struct {
		Type          string
		File          string
		KnownTypeName string
		KnownFileName string
	}, pkgData.FileCount)
	for i, item := range pkgData.Items {
		outData[i].Type = item.Type.String()
		outData[i].File = item.Name.String()
		outData[i].KnownTypeName = ctx.Hashes()[item.Type]
		outData[i].KnownFileName = ctx.Hashes()[item.Name]
	}
	out, err := ctx.CreateFile(".package.json")
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(outData); err != nil {
		return err
	}
	return nil
}
