package texture

import (
	"bytes"
	"errors"
	"image/png"
	"io"

	"github.com/xypwn/filediver/dds"
	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/unit/texture"
)

func ExtractDDSData(ctx *extractor.Context, id stingray.FileID) ([]byte, error) {
	if !ctx.Exists(id, stingray.DataMain) {
		return nil, errors.New("no main data")
	}
	var rs []io.Reader
	for dataType := range stingray.NumDataType {
		r, err := ctx.Open(id, dataType)
		if err == stingray.ErrFileDataTypeNotExist {
			continue
		}
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}

	r := io.MultiReader(rs...)
	if _, err := texture.DecodeInfo(r); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ExtractDDS(ctx *extractor.Context) error {
	data, err := ExtractDDSData(ctx, ctx.FileID())
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

func ConvertToPNGData(ctx *extractor.Context, id stingray.FileID) ([]byte, error) {
	ddsData, err := ExtractDDSData(ctx, id)
	if err != nil {
		return nil, err
	}

	origTex, err := dds.Decode(bytes.NewReader(ddsData), false)
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

func ConvertToPNG(ctx *extractor.Context) error {
	data, err := ConvertToPNGData(ctx, ctx.FileID())
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
