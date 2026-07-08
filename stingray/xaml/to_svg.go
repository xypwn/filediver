package xaml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var ErrNoResources = errors.New("xaml file does not contain a ResourceDictionary")

var re *regexp.Regexp = regexp.MustCompile(`^M[\d\.,]*(\D)`)

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
		}

		fmt.Fprintf(&b, "<svg viewBox=\"0 0 %v %v\" xmlns=\"http://www.w3.org/2000/svg\">\n", *(canvas.Width), *(canvas.Height))

		for _, path := range paths {
			fillRule := "evenodd"
			if path.Data.FillRule == FillNonzero {
				fillRule = "nonzero"
			}
			fmt.Fprintf(&b, "  <path d=\"%s\"\n", path.Data.ToString(true))
			fmt.Fprintf(&b, "    fill=\"%s\"\n", path.Fill)
			if fillRule != "nonzero" {
				fmt.Fprintf(&b, "    fill-rule=\"%v\"\n", fillRule)
			}
			fmt.Fprintf(&b, "    stroke=\"%v\"\n", path.Stroke)
			if path.StrokeThickness != nil {
				fmt.Fprintf(&b, "    stroke-width=\"%v\"\n", *path.StrokeThickness)
			}
			if path.MatrixTransform != nil {
				matrix := fmt.Sprintf("%v,%v,%v", strings.TrimSuffix(path.MatrixTransform.Matrix, ",0,0"), path.Left, path.Top)
				fmt.Fprintf(&b, "    transform=\"matrix(%v)\"\n", strings.ReplaceAll(matrix, ",", " "))
			}
			fmt.Fprint(&b, "  />\n")
		}
		fmt.Fprint(&b, "</svg>\n")
		svgs[element.Key] = b.Bytes()
	}
	return
}
