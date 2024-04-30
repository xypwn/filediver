package texture

import (
	"errors"
	"image/png"
	"io"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

func ExtractDDS(ctx extractor.Context) error {
	if !ctx.File().Exists(stingray.DataMain) {
		return errors.New("no main data")
	}
	r, err := ctx.File().OpenMulti(ctx.Ctx(), stingray.DataMain, stingray.DataStream, stingray.DataGPU)
	if err != nil {
		return err
	}
	defer r.Close()
	if _, err := texture.DecodeInfo(r); err != nil {
		return err
	}
	out, err := ctx.CreateFile(".dds")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func ConvertToPNG(ctx extractor.Context) error {
	tex, err := texture.Decode(ctx.Ctx(), ctx.File(), false)
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".png")
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, tex)
}
