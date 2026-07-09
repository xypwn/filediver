package previews

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/srwiley/rasterx"
	"github.com/xypwn/filediver/stingray/xaml"
	"golang.org/x/image/math/fixed"
)

type extentScanner struct {
	haveExtent bool
	extent     fixed.Rectangle26_6
	clipRect   fixed.Rectangle26_6
}

func (s *extentScanner) update(p fixed.Point26_6) {
	if !s.clipRect.Empty() && !p.In(s.clipRect) {
		return
	}
	if !s.haveExtent {
		s.extent.Min, s.extent.Max = p, p
		s.haveExtent = true
	}
	s.extent.Min.X = min(s.extent.Min.X, p.X)
	s.extent.Min.Y = min(s.extent.Min.Y, p.Y)
	s.extent.Max.X = max(s.extent.Max.X, p.X)
	s.extent.Max.Y = max(s.extent.Max.Y, p.Y)
}
func (s *extentScanner) Start(p fixed.Point26_6) {
	s.update(p)
}
func (s *extentScanner) Line(p fixed.Point26_6) {
	s.update(p)
}
func (s *extentScanner) Draw() {}
func (s *extentScanner) GetPathExtent() fixed.Rectangle26_6 {
	return s.extent
}
func (s *extentScanner) SetBounds(int, int) {}
func (s *extentScanner) SetColor(any)       {}
func (s *extentScanner) SetWinding(bool)    {}
func (s *extentScanner) Clear()             {}
func (s *extentScanner) SetClip(r image.Rectangle) {
	s.clipRect = fixed.R(r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
}

func sToFixed(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64)
}
func pToFixed(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{X: sToFixed(x), Y: sToFixed(y)}
}
func sFromFixed(x fixed.Int26_6) float64 {
	return float64(x) / 64
}
func rFromFixed(r fixed.Rectangle26_6) (x0, y0, x1, y1 float64) {
	return sFromFixed(r.Min.X), sFromFixed(r.Min.Y), sFromFixed(r.Max.X), sFromFixed(r.Max.Y)
}

/*type imguiScanner struct {
	pos    imgui.Vec2
	dl     *imgui.DrawList
	paths  [][]imgui.Vec2
	extSc  extentScanner
	extent fixed.Rectangle26_6
	color  uint32
}

func (s *imguiScanner) addPoint(p fixed.Point26_6) {
	i := len(s.paths) - 1
	if i == -1 {
		return
	}
	v := imgui.NewVec2(float32(sFromFixed(p.X)), float32(sFromFixed(p.Y)))
	v = v.Add(s.pos)
	s.paths[i] = append(s.paths[i], v)
	s.extSc.update(p)
}
func (s *imguiScanner) Start(p fixed.Point26_6) {
	s.paths = append(s.paths, []imgui.Vec2{})
	s.addPoint(p)
}
func (s *imguiScanner) Line(p fixed.Point26_6) {
	s.addPoint(p)
}
func (s *imguiScanner) Draw() {
	for _, path := range s.paths {
		if len(path) == 0 {
			continue
		}
		for _, p := range path {
			s.dl.PathLineTo(p)
		}
		s.dl.PathLineTo(path[0])
		s.dl.PathFillConcave(s.color)
		//for i := len(path) - 1; i >= 0; i-- {
		//	s.dl.PathLineTo(path[i])
		//}
		//s.dl.PathLineTo(path[len(path)-1])
		//s.dl.PathFillConcave(s.color)
	}
}
func (s *imguiScanner) GetPathExtent() fixed.Rectangle26_6 {
	return s.extSc.GetPathExtent()
}
func (s *imguiScanner) SetBounds(int, int) {}
func (s *imguiScanner) SetColor(col any) {
	switch col := col.(type) {
	case color.Color:
		c := color.NRGBAModel.Convert(col).(color.NRGBA)
		s.color = uint32(c.R) | uint32(c.G)<<8 | uint32(c.B)<<16 | uint32(c.A)<<24
	default:
		panic(fmt.Sprintf("unsupported color type %T", col))
	}
}
func (s *imguiScanner) SetWinding(bool) {}
func (s *imguiScanner) Clear() {
	s.paths = nil
}
func (s *imguiScanner) SetClip(r image.Rectangle) {
	if s.dl.ClipRectStack().Size > 0 {
		s.dl.PopClipRect()
	}
	if r.Empty() {
		s.dl.PushClipRectFullScreen()
		return
	}
	rmin := imgui.NewVec2(float32(r.Min.X), float32(r.Min.Y))
	rmax := imgui.NewVec2(float32(r.Min.X), float32(r.Min.Y))
	s.dl.PushClipRect(rmin, rmax)
}*/

// cmds must be from a simplified path.
func addSimplePath(cmds []xaml.PathCmd, adder rasterx.Adder) {
	started := false
	var x, y float64
	for _, cmd := range cmds {
		p := cmd.Params
		switch cmd.Cmd {
		case xaml.PathMove:
			if started {
				adder.Stop(false)
			}
			adder.Start(pToFixed(p[0], p[1]))
			started = true
			x, y = p[0], p[1]
		case xaml.PathLine:
			adder.Line(pToFixed(p[0], p[1]))
			x, y = p[0], p[1]
		case xaml.PathCubicBezier:
			adder.CubeBezier(pToFixed(p[0], p[1]), pToFixed(p[2], p[3]), pToFixed(p[4], p[5]))
			x, y = p[4], p[5]
		case xaml.PathQuadraticBezier:
			adder.QuadBezier(pToFixed(p[0], p[1]), pToFixed(p[2], p[3]))
			x, y = p[2], p[3]
		case xaml.PathEllipticalArc:
			cx, cy := rasterx.FindEllipseCenter(
				&p[0], &p[1], p[2]*math.Pi/180,
				x, y, p[5], p[6],
				p[4] == 0, p[3] == 0,
			)
			x, y = rasterx.AddArc(p[:], cx, cy, x, y, adder)
		default:
			panic(fmt.Sprintf("unhandled command type %q (expected simplified command, meaning one of M,L,C,Q,A)", cmd.Cmd.ToByte(cmd.Relative)))
		}
	}
	if started {
		adder.Stop(false)
	}
}

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

	if !bytes.HasPrefix(srXaml.Data, []byte("<ResourceDictionary")) {
		return xaml.ErrNoResources
	}

	var resources xaml.ResourceDictionary
	err = xml.Unmarshal(srXaml.Data, &resources)
	if err != nil {
		return err
	}

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

		var extent fixed.Rectangle26_6
		var pathExtents []fixed.Rectangle26_6
		{
			for i, path := range pathsData {
				sc := &extentScanner{}
				filler := rasterx.NewFiller(0, 0, sc)
				addSimplePath(path.Cmds, filler)
				ext := sc.GetPathExtent()
				pathExtents = append(pathExtents, ext)
				if i == 0 {
					extent = ext
				} else {
					extent = extent.Union(ext)
				}
			}
		}
		minX, minY, maxX, maxY := rFromFixed(extent)

		// Add some padding
		if !extent.Empty() {
			padX := (maxX-minX)*0.1 + 10
			padY := (maxY-minY)*0.1 + 10
			minX -= padX
			minY -= padY
			maxX += padX
			maxY += padY
		}
		width, height := int(maxX-minX), int(maxY-minY)

		rasterWidth, rasterHeight := width, height
		scale := 1.0
		const minDim = 256
		if sz := min(rasterWidth, rasterHeight); sz > 0 && sz < minDim {
			scale = minDim / float64(sz)
			rasterWidth = int(float64(rasterWidth) * scale)
			rasterHeight = int(float64(rasterHeight) * scale)
		}
		img := image.NewNRGBA(image.Rect(0, 0, rasterWidth, rasterHeight))
		scanner := rasterx.NewScannerGV(img.Bounds().Dx(), img.Bounds().Dy(), img, img.Bounds())
		dasher := rasterx.NewDasher(img.Bounds().Dx(), img.Bounds().Dy(), scanner)
		mAdder := &rasterx.MatrixAdder{
			Adder: dasher,
		}
		var defaultMat rasterx.Matrix2D
		{
			m := rasterx.Identity
			m = m.Scale(scale, scale)
			m = m.Translate(-minX, -minY)
			defaultMat = m
		}
		for i, path := range paths {
			pathData := pathsData[i]
			strokeWidth := 1.0
			if path.StrokeThickness != nil {
				strokeWidth = float64(*path.StrokeThickness)
			}

			mAdder.M = defaultMat
			if path.MatrixTransform != nil {
				// TODO: Add proper layouting functionality:
				//  - Width, Height
				//  - Stretch="Fill"
				//	- Canvas.Top/Left
				// This probably needs some kind of intermediate representation to be feasible for SVG and and preview.
				tm := path.MatrixTransform.Matrix
				m := rasterx.Matrix2D{A: tm[0], B: tm[1], C: tm[2], D: tm[3], E: tm[4], F: tm[5]}
				m = m.Translate(float64(path.Left /*+path.Width/2*/), float64(path.Top /*+path.Height/2*/)) // this is broken here, but seems to work on exported SVGs?
				m = mAdder.M.Mult(m)
				mAdder.M = m
			}

			drawFill := false
			var fill any
			if path.FillAttr != nil {
				fill = *path.FillAttr
				drawFill = path.FillAttr.Color.A != 0
			}
			if path.FillElem != nil {
				if path.FillElem.SolidColorBrush != nil {
					fill = path.FillElem.SolidColorBrush.Color
					drawFill = true
				}
				if lg := path.FillElem.LinearGradientBrush; lg != nil {
					x0, y0, x1, y1 := rFromFixed(pathExtents[i])
					//fmt.Println(x0, y0, x1, y1, lg) // TODO: Not working
					g := rasterx.Gradient{
						Points: [5]float64{lg.StartPoint[0], lg.StartPoint[1], lg.EndPoint[0], lg.EndPoint[1]},
						Bounds: struct {
							X float64
							Y float64
							W float64
							H float64
						}{x0, y0, x1 - x0, y1 - y0},
						Matrix:   rasterx.Identity,
						Spread:   rasterx.PadSpread,
						Units:    rasterx.ObjectBoundingBox,
						IsRadial: false,
					}
					for _, gs := range lg.GradientStop {
						g.Stops = append(g.Stops, rasterx.GradStop{
							StopColor: gs.Color.Color,
							Offset:    gs.Offset,
							Opacity:   1,
						})
					}
					fill = g.GetColorFunction(1)
					drawFill = true
				}
			}
			drawStroke := false
			var stroke any
			if path.Stroke != nil {
				stroke = *path.Stroke
				drawStroke = path.Stroke.Color.A != 0
			}
			if fill == nil && stroke == nil {
				fill = color.RGBA{R: 255, G: 255, B: 255, A: 255}
				drawFill = true
			}

			if drawFill && fill != nil {
				dasher.Clear()
				dasher.Filler.SetColor(fill)
				mAdder.Adder = &dasher.Filler
				addSimplePath(pathData.Cmds, mAdder)
				dasher.Filler.Draw()
			}
			if drawStroke && stroke != nil {
				dasher.Clear()
				dasher.SetColor(stroke)
				dasher.SetStroke(
					sToFixed(strokeWidth*scale),
					sToFixed(4*scale),
					rasterx.ButtCap, rasterx.ButtCap,
					nil, rasterx.Bevel,
					nil, 0)
				mAdder.Adder = dasher
				addSimplePath(pathData.Cmds, mAdder)
				dasher.Draw()
			}

			//if pathData.FillRule == xaml.FillNonzero { } else { } // Fill rules don't seem to be supported
		}
		imgInfoTexts = append(imgInfoTexts, fmt.Sprintf(
			"image: %q (%v/%v)\nsize=%vx%v",
			element.Key,
			elementIdx+1, len(resources.DataTemplate),
			width, height))
		imgs = append(imgs, img)
	}

	pv.LoadImages(imgs)
	for i, s := range imgInfoTexts {
		pv.Images[i].DrawInfo = func() { imgui.TextUnformatted(s) }
	}
	return nil
}
