package texture

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"io"

	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

func ExtractDDSData(ctx context.Context, file *stingray.File) ([]byte, error) {
	if !file.Exists(stingray.DataMain) {
		return nil, errors.New("no main data")
	}
	r, err := file.OpenMulti(ctx, stingray.DataMain, stingray.DataStream, stingray.DataGPU)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if _, err := texture.DecodeInfo(r); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ExtractDDS(ctx extractor.Context) error {
	data, err := ExtractDDSData(ctx.Ctx(), ctx.File())
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".dds")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(data)
	return err
}

func ConvertToPNGData(ctx context.Context, file *stingray.File) ([]byte, error) {
	origTex, err := texture.Decode(ctx, file, false)
	if err != nil {
		return nil, err
	}

	tex := origTex
	if len(origTex.Images) > 1 {
		tex = dds.StackLayers(origTex)
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, tex)
	return buf.Bytes(), err
}

func ConvertToPNG(ctx extractor.Context) error {
	data, err := ConvertToPNGData(ctx.Ctx(), ctx.File())
	if err != nil {
		return err
	}

	out, err := ctx.CreateFile(".png")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(data)
	return err
}
