package stingray_package

import (
	"encoding/json"
	"errors"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_package "github.com/xypwn/filediver/stingray/package"
)

func ExtractPackageJSON(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().Open(ctx.Ctx(), stingray.DataMain)
	if err != nil {
		return err
	}
	defer r.Close()
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
