package ttf

import (
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_ttf "github.com/xypwn/filediver/stingray/ttf"
)

func extract(ctx *extractor.Context, suffix string) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	ttf, err := stingray_ttf.LoadTTF(r)
	if err != nil {
		return err
	}
	out, err := ctx.CreateFile(suffix)
	if err != nil {
		return err
	}
	defer out.Close()
	ttfData, err := ttf.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = out.Write(ttfData)
	return err
}

func ExtractTTF(ctx *extractor.Context) error {
	return extract(ctx, ".ttf")
}

func ExtractOTF(ctx *extractor.Context) error {
	return extract(ctx, ".otf")
}
