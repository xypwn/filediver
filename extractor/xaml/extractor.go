package xaml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	stingray_xaml "github.com/xypwn/filediver/stingray/xaml"
)

var ErrNoResources = errors.New("xaml file does not contain a ResourceDictionary")

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

var re *regexp.Regexp = regexp.MustCompile(`^M[\d\.,]*(\D)`)

func ExtractSVG(ctx *extractor.Context) error {
	r, err := ctx.Open(ctx.FileID(), stingray.DataMain)
	if err != nil {
		return err
	}
	xaml, err := stingray_xaml.LoadXAML(r)
	if err != nil {
		return err
	}
	xamlData, err := xaml.MarshalText()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(string(xamlData), "<ResourceDictionary") {
		return ErrNoResources
	}
	var resources stingray_xaml.ResourceDictionary
	err = xml.Unmarshal(xamlData, &resources)
	if err != nil {
		return err
	}
	for _, element := range resources.DataTemplate {
		// fmt.Println(element.Key)
		out, err := ctx.CreateFile(fmt.Sprintf("/%v.svg", element.Key))
		if err != nil {
			return err
		}
		defer out.Close()
		var paths []stingray_xaml.Path
		var canvas *stingray_xaml.Canvas
		if element.Viewbox != nil && element.Viewbox.Canvas != nil && len(element.Viewbox.Canvas.Path) > 0 {
			paths = element.Viewbox.Canvas.Path
			canvas = element.Viewbox.Canvas
		} else if element.Canvas != nil && len(element.Canvas.Path) > 0 {
			paths = element.Canvas.Path
			canvas = element.Canvas
		} else {
			paths = element.Path
			canvas = &stingray_xaml.Canvas{
				Width:  nil,
				Height: nil,
			}
		}
		if canvas.Width == nil || canvas.Height == nil {
			var width float32 = 0
			var height float32 = 0
			canvas.Width = &width
			canvas.Height = &height
			for _, path := range paths {
				*canvas.Width = float32(math.Max(float64(path.Width+path.Left), float64(*canvas.Width)))
				*canvas.Height = float32(math.Max(float64(path.Height+path.Top), float64(*canvas.Height)))
			}
		}

		_, err = fmt.Fprintf(out, "<svg viewBox=\"0 0 %v %v\" xmlns=\"http://www.w3.org/2000/svg\">\n", *(canvas.Width), *(canvas.Height))
		if err != nil {
			return err
		}

		for _, path := range paths {
			data := path.Data
			fillRule := "evenodd"
			if cut, found := strings.CutPrefix(data, "F1"); found {
				data = cut
				fillRule = "nonzero"
			}
			if path.MatrixTransform == nil {
				if _, err = fmt.Fprintf(out, "  <path d=\"%v\"\n", data); err != nil {
					return err
				}
			} else {
				loc := re.FindStringIndex(data)
				if loc == nil {
					return fmt.Errorf("unexpected path start: \"%v\"", data[:32])
				}
				if _, err = fmt.Fprintf(out, "  <path d=\"M0,0%v\"\n", data[loc[1]-1:]); err != nil {
					return err
				}
			}
			if len(path.Fill) > 0 {
				var fill string = path.Fill
				if cut, found := strings.CutPrefix(fill, "#FF"); found {
					fill = fmt.Sprintf("#%v", cut)
				}
				if _, err = fmt.Fprintf(out, "    fill=\"%v\"\n", fill); err != nil {
					return err
				}
			} else {
				if _, err = fmt.Fprint(out, "    fill=\"rgba(0,0,0,0)\"\n"); err != nil {
					return err
				}
			}
			if fillRule != "nonzero" {
				if _, err = fmt.Fprintf(out, "    fill-rule=\"%v\"\n", fillRule); err != nil {
					return err
				}
			}
			if len(path.Stroke) > 0 {
				stroke := path.Stroke
				if cut, found := strings.CutPrefix(stroke, "#FF"); found {
					stroke = fmt.Sprintf("#%v", cut)
				}
				if _, err = fmt.Fprintf(out, "    stroke=\"%v\"\n", stroke); err != nil {
					return err
				}
			}
			if path.StrokeThickness != nil {
				if _, err = fmt.Fprintf(out, "    stroke-width=\"%v\"\n", *path.StrokeThickness); err != nil {
					return err
				}
			}
			if path.MatrixTransform != nil {
				matrix := fmt.Sprintf("%v,%v,%v", strings.TrimSuffix(path.MatrixTransform.Matrix, ",0,0"), path.Left, path.Top)
				if _, err = fmt.Fprintf(out, "    transform=\"matrix(%v)\"\n", strings.ReplaceAll(matrix, ",", " ")); err != nil {
					return err
				}
			}
			if _, err = fmt.Fprint(out, "  />\n"); err != nil {
				return err
			}
		}
		if _, err = fmt.Fprint(out, "</svg>\n"); err != nil {
			return err
		}
	}
	return nil
}
