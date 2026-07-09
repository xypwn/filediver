package xaml

import (
	"fmt"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_xaml "github.com/xypwn/filediver/stingray/xaml"
)

func ExtractXAML(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	xaml, err := stingray_xaml.LoadXAML(r)
	if err != nil {
		return err
	}
	out, err := ctx.CreateFile(".xaml")
	if err != nil {
		return err
	}
	defer out.Close()
	xamlData, err := xaml.MarshalText()
	if err != nil {
		return err
	}
	_, err = out.Write(xamlData)
	return err
}

func ExtractSVG(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	xaml, err := stingray_xaml.LoadXAML(r)
	if err != nil {
		return err
	}
	svgs, err := xaml.ToSVGs()
	if err != nil {
		return err
	}
	for name, svg := range svgs {
		f, err := ctx.CreateFile(fmt.Sprintf(".svgs/%s.svg", name))
		if err != nil {
			return err
		}
		_, err = f.Write(svg)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
