package extractor

import (
	"fmt"
	"io"

	"github.com/xypwn/filediver/stingray"
)

type ExtractFunc func(ctx *Context) error

func extractByType(ctx *Context, typ stingray.DataType, extension string) error {
	r, err := ctx.Open(ctx.FileID(), typ)
	if err != nil {
		return err
	}

	var typExtension string
	switch typ {
	case stingray.DataMain:
		typExtension = ".main"
	case stingray.DataStream:
		typExtension = ".stream"
	case stingray.DataGPU:
		typExtension = ".gpu"
	default:
		panic("unhandled case")
	}
	out, err := ctx.CreateFile("." + extension + typExtension)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return err
	}

	return nil
}

func extractCombined(ctx *Context, extension string) error {
	if !(ctx.Exists(ctx.FileID(), stingray.DataMain) || ctx.Exists(ctx.FileID(), stingray.DataStream) || ctx.Exists(ctx.FileID(), stingray.DataGPU)) {
		return fmt.Errorf("extractCombined: no data to extract for file")
	}
	out, err := ctx.CreateFile(fmt.Sprintf(".%v", extension))
	if err != nil {
		return err
	}
	defer out.Close()

	for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
		r, err := ctx.Open(ctx.FileID(), typ)
		if err == stingray.ErrFileDataTypeNotExist {
			continue
		}
		if err != nil {
			return err
		}

		if _, err := io.Copy(out, r); err != nil {
			return err
		}
	}

	return nil
}

func ExtractFuncRaw(extension string) ExtractFunc {
	return func(ctx *Context) error {
		for _, typ := range [3]stingray.DataType{stingray.DataMain, stingray.DataStream, stingray.DataGPU} {
			if ctx.Exists(ctx.FileID(), typ) {
				if err := extractByType(ctx, typ, extension); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func ExtractFuncRawSingleType(extension string, typ stingray.DataType) ExtractFunc {
	return func(ctx *Context) error {
		if ctx.Exists(ctx.FileID(), typ) {
			return extractByType(ctx, typ, extension)
		}
		return fmt.Errorf("no %v data found", typ.String())
	}
}

func ExtractFuncRawCombined(extension string) ExtractFunc {
	return func(ctx *Context) error {
		return extractCombined(ctx, extension)
	}
}
