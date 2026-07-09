package xaml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strings"
)

var ErrNoResources = errors.New("xaml file does not contain a ResourceDictionary")

func (xaml *XAML) ToSVGs() (svgs map[string][]byte, err error) {
	svgs = make(map[string][]byte)
	if !strings.HasPrefix(string(xaml.Data), "<ResourceDictionary") {
		return nil, ErrNoResources
	}
	var resources ResourceDictionary
	err = xml.Unmarshal(xaml.Data, &resources)
	if err != nil {
		return nil, err
	}
	for _, element := range resources.DataTemplate {
		var b bytes.Buffer
		var paths []Path
		var canvas *Canvas
		if element.Viewbox != nil && element.Viewbox.Canvas != nil && len(element.Viewbox.Canvas.Path) > 0 {
			paths = element.Viewbox.Canvas.Path
			canvas = element.Viewbox.Canvas
		} else if element.Canvas != nil && len(element.Canvas.Path) > 0 {
			paths = element.Canvas.Path
			canvas = element.Canvas
		} else {
			paths = element.Path
			canvas = &Canvas{
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
			if width == 0 || height == 0 {
				// TODO: Properly figure out dims
				width, height = 512, 512
			}
		}

		fmt.Fprintf(&b, "<svg viewBox=\"0 0 %v %v\" xmlns=\"http://www.w3.org/2000/svg\">\n", *(canvas.Width), *(canvas.Height))

		for _, path := range paths {
			fillRule := "evenodd"
			if path.Data.FillRule == FillNonzero {
				fillRule = "nonzero"
			}
			fmt.Fprintf(&b, "  <path d=\"%s\"\n", path.Data.ToString(true))
			if path.FillAttr != nil {
				fmt.Fprintf(&b, "    fill=\"%s\"\n", *path.FillAttr)
			}
			if path.FillElem != nil && path.FillElem.SolidColorBrush != nil {
				fmt.Fprintf(&b, "    fill=\"%s\"\n", path.FillElem.SolidColorBrush.Color)
			}
			if fillRule != "nonzero" {
				fmt.Fprintf(&b, "    fill-rule=\"%v\"\n", fillRule)
			}
			if path.Stroke != nil {
				fmt.Fprintf(&b, "    stroke=\"%v\"\n", *path.Stroke)
			}
			if path.StrokeThickness != nil {
				fmt.Fprintf(&b, "    stroke-width=\"%v\"\n", *path.StrokeThickness)
			}
			if path.MatrixTransform != nil {
				m := path.MatrixTransform.Matrix
				m[4], m[5] = float64(path.Left), float64(path.Top)
				fmt.Fprintf(&b, "    transform=\"matrix(%s)\"\n", strings.ReplaceAll(m.String(), ",", " "))
			}
			fmt.Fprint(&b, "  />\n")
		}
		fmt.Fprint(&b, "</svg>\n")
		svgs[element.Key] = b.Bytes()
	}
	return
}
