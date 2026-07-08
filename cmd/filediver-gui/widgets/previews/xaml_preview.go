package previews

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"math"

	"git.sr.ht/~sbinet/gg"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/xypwn/filediver/stingray/xaml"
)

type XamlPreview struct {
	*ImagePreview
}

func NewXamlPreview() *XamlPreview {
	return &XamlPreview{
		ImagePreview: NewImagePreview(),
	}
}

func (pv *XamlPreview) LoadXaml(src []byte) error {
	pv.Flags = MultipleImages

	srXaml, err := xaml.LoadXAML(bytes.NewReader(src))
	if err != nil {
		return err
	}

	var resources xaml.ResourceDictionary
	err = xml.Unmarshal(srXaml.Data, &resources)
	if err != nil {
		return err
	}
	//os.WriteFile("data.xaml", srXaml.Data, 0666)

	var imgs []image.Image
	var imgInfoTexts []string
	for elementIdx, element := range resources.DataTemplate {
		var paths []xaml.Path
		var canvas *xaml.Canvas
		if element.Viewbox != nil && element.Viewbox.Canvas != nil && len(element.Viewbox.Canvas.Path) > 0 {
			paths = element.Viewbox.Canvas.Path
			canvas = element.Viewbox.Canvas
		} else if element.Canvas != nil && len(element.Canvas.Path) > 0 {
			paths = element.Canvas.Path
			canvas = element.Canvas
		} else {
			paths = element.Path
			canvas = &xaml.Canvas{
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

		var pathsData []xaml.PathData
		for _, path := range paths {
			pathsData = append(pathsData, path.Data.Simplify())
		}

		haveMinMax := false
		var minX, minY, maxX, maxY float64
		updateMinMax := func(x, y float64) {
			if !haveMinMax {
				minX, maxX, minY, maxY = x, x, y, y
				haveMinMax = true
				return
			}
			minX = min(minX, x)
			minY = min(minY, y)
			maxX = max(maxX, x)
			maxY = max(maxY, y)
		}
		for _, path := range pathsData {
			for _, cmd := range path.Cmds {
				p := cmd.Params
				switch cmd.Cmd {
				case xaml.PathMove:
					updateMinMax(p[0], p[1])
				case xaml.PathLine:
					updateMinMax(p[0], p[1])
				case xaml.PathCubicBezier:
					updateMinMax(p[4], p[5])
				case xaml.PathQuadraticBezier:
					updateMinMax(p[2], p[3])
				case xaml.PathEllipticalArc:
					updateMinMax(p[5], p[6])
				default:
					return fmt.Errorf("unhandled command type %q", cmd.Cmd.ToByte(cmd.Relative))
				}
			}
		}

		// Add some padding
		{
			padX := (maxX - minX) * 0.1
			padY := (maxY - minY) * 0.1
			minX -= padX
			minY -= padY
			maxX += padX
			maxY += padY
		}

		rasterWidth, rasterHeight := int(maxX-minX), int(maxY-minY)
		scale := 1.0
		const minDim = 512
		if sz := min(rasterWidth, rasterHeight); sz > 0 && sz < minDim {
			scale = minDim / float64(sz)
			rasterWidth = int(float64(rasterWidth) * scale)
			rasterHeight = int(float64(rasterHeight) * scale)
		}
		dc := gg.NewContext(rasterWidth, rasterHeight)
		dc.Scale(scale, scale)
		dc.Translate(-minX, -minY)
		for i, path := range paths {
			pathData := pathsData[i]
			for _, cmd := range pathData.Cmds {
				p := cmd.Params
				switch cmd.Cmd {
				case xaml.PathMove:
					dc.MoveTo(p[0], p[1])
				case xaml.PathLine:
					dc.LineTo(p[0], p[1])
				case xaml.PathCubicBezier:
					dc.CubicTo(p[0], p[1], p[2], p[3], p[4], p[5])
				case xaml.PathQuadraticBezier:
					dc.QuadraticTo(p[0], p[1], p[2], p[3])
				case xaml.PathEllipticalArc:
					// TODO
					//dc.DrawEllipticalArc()
				default:
					return fmt.Errorf("unhandled command type %q", cmd.Cmd.ToByte(cmd.Relative))
				}
			}
			if path.StrokeThickness != nil {
				dc.SetLineWidth(float64(*path.StrokeThickness))
			} else {
				dc.SetLineWidth(1)
			}
			if pathData.FillRule == xaml.FillNonzero {
				dc.SetFillRuleEvenOdd() // TODO: Nonzero might not be an option in gg
				//dc.SetFillRuleWinding()
			} else {
				dc.SetFillRuleEvenOdd()
			}
			if path.Fill.A != 0 {
				dc.SetColor(path.Fill)
				dc.FillPreserve()
			}
			dc.SetColor(path.Stroke)
			dc.Stroke()
		}
		imgInfoTexts = append(imgInfoTexts, fmt.Sprintf(
			"image: %q (%v/%v)\nsize=%vx%v",
			element.Key,
			elementIdx+1, len(resources.DataTemplate),
			*canvas.Width, *canvas.Height))
		imgs = append(imgs, dc.Image())
	}

	pv.LoadImages(imgs)
	for i, s := range imgInfoTexts {
		pv.Images[i].DrawInfo = func() { imgui.TextUnformatted(s) }
	}
	return nil
}
