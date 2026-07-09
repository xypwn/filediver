package xaml

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

type Header struct {
	Size uint32
	_    [12]uint8
}

type XAML struct {
	Header
	Data []byte
}

// Affine matrix of the following form,
// where a-f are the array elements:
//
//	|a c e|
//	|b d f|
//	|0 0 1|
type TransformMatrix [6]float64

func (m TransformMatrix) String() string {
	return fmt.Sprintf("%v,%v,%v,%v,%v,%v", m[0], m[1], m[2], m[3], m[4], m[5])
}

func (m *TransformMatrix) UnmarshalText(text []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse matrix %q: %w", text, err)
		}
	}()
	sp := bytes.SplitN(text, []byte(","), 6)
	if len(sp) != 6 {
		return fmt.Errorf("expected 6 comma-separated values")
	}
	var mNew TransformMatrix
	for i, s := range sp {
		f, err := strconv.ParseFloat(string(s), 64)
		if err != nil {
			return err
		}
		mNew[i] = f
	}
	*m = mNew
	return nil
}

// Always succeeds.
func (m TransformMatrix) MarshalText() (text []byte, err error) {
	return []byte(m.String()), nil
}

type MatrixTransform struct {
	XMLName xml.Name        `xml:"MatrixTransform"`
	Matrix  TransformMatrix `xml:",attr"`
}

type Point [2]float64

func (p *Point) UnmarshalText(text []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse point %q: %w", text, err)
		}
	}()
	sp := bytes.SplitN(text, []byte(","), 2)
	if len(sp) != 2 {
		return fmt.Errorf("expected 2 comma-separated values")
	}
	var pNew Point
	for i, s := range sp {
		f, err := strconv.ParseFloat(string(s), 64)
		if err != nil {
			return err
		}
		pNew[i] = f
	}
	*p = pNew
	return nil
}

func (p Point) String() string {
	return fmt.Sprintf("%v,%v", p[0], p[1])
}

// Always succeeds.
func (p Point) MarshalText() (text []byte, err error) {
	return []byte(p.String()), nil
}

type SolidColorBrush struct {
	Color Color `xml:",attr"`
}

type GradientStop struct {
	Color  Color   `xml:",attr"`
	Offset float64 `xml:",attr"`
}

type LinearGradientBrush struct {
	GradientStop []GradientStop
	StartPoint   Point `xml:",attr"`
	EndPoint     Point `xml:",attr"`
}

type ShapeFill struct {
	SolidColorBrush     *SolidColorBrush
	LinearGradientBrush *LinearGradientBrush
}

type Path struct {
	XMLName         xml.Name         `xml:"Path"`
	FillAttr        *Color           `xml:"Fill,attr"`
	FillElem        *ShapeFill       `xml:"Path.Fill"`
	Stroke          *Color           `xml:",attr"`
	StrokeThickness *uint32          `xml:",attr"`
	Data            PathData         `xml:",attr"`
	Stretch         string           `xml:",attr"`
	Height          float32          `xml:",attr"`
	Width           float32          `xml:",attr"`
	Left            float32          `xml:"Canvas.Left,attr"`
	Top             float32          `xml:"Canvas.Top,attr"`
	MatrixTransform *MatrixTransform `xml:"Path.RenderTransform>MatrixTransform"`
}

type Canvas struct {
	XMLName xml.Name `xml:"Canvas"`
	Width   *float32 `xml:",attr"`
	Height  *float32 `xml:",attr"`
	Path    []Path
}

type Viewbox struct {
	XMLName xml.Name `xml:"Viewbox"`
	Canvas  *Canvas
}

type DataTemplate struct {
	XMLName xml.Name `xml:"DataTemplate"`
	Key     string   `xml:",attr"`
	Viewbox *Viewbox
	Canvas  *Canvas
	Path    []Path
}

type ResourceDictionary struct {
	XMLName      xml.Name `xml:"ResourceDictionary"`
	DataTemplate []DataTemplate
}

func (x XAML) MarshalText() ([]byte, error) {
	return x.Data, nil
}

func LoadXAML(r io.Reader) (*XAML, error) {
	var hdr Header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	data := make([]byte, hdr.Size)
	if err := binary.Read(r, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return &XAML{
		Header: hdr,
		Data:   data,
	}, nil
}
