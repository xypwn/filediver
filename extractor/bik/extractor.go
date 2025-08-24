package bik

import (
	"encoding/binary"
	"io"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
)

func extract(ctx extractor.Context, save func(ctx extractor.Context, r io.Reader) error) error {
	dataTypes := []stingray.DataType{stingray.DataMain}
	if ctx.Exists(ctx.FileID(), stingray.DataStream) {
		dataTypes = append(dataTypes, stingray.DataStream)
	} else {
		dataTypes = append(dataTypes, stingray.DataGPU)
	}

	var rs []io.Reader
	for _, dataType := range dataTypes {
		r, err := ctx.Open(ctx.FileID(), dataType)
		if err != nil {
			return err
		}
		rs = append(rs, r)
	}

	r := io.MultiReader(rs...)
	var hdr struct {
		Unk00 [16]byte
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return err
	}

	return save(ctx, r)
}

func ExtractBik(ctx extractor.Context) error {
	return extract(ctx, func(ctx extractor.Context, r io.Reader) error {
		out, err := ctx.CreateFile(".bik")
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, r); err != nil {
			return err
		}
		return nil
	})
}

func ConvertToMP4(ctx extractor.Context) error {
	if !ctx.Runner().Has("ffmpeg") {
		return ExtractBik(ctx)
	}

	return extract(ctx, func(ctx extractor.Context, r io.Reader) error {
		outPath, err := ctx.AllocateFile(".mp4")
		if err != nil {
			return err
		}
		return ctx.Runner().Run(
			"ffmpeg",
			nil,
			r,
			"-f", "bink",
			"-i", "pipe:",
			outPath,
		)
	})
}
