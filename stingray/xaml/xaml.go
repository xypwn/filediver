package xaml

import (
	"encoding/binary"
	"encoding/xml"
	"io"
)

type Header struct {
	Size uint32
	_    [12]uint8
}

type XAML struct {
	Header
	Data []byte
}

type MatrixTransform struct {
	XMLName xml.Name `xml:"MatrixTransform"`
	Matrix  string   `xml:",attr"`
}

type Path struct {
	XMLName         xml.Name         `xml:"Path"`
	Fill            Color            `xml:",attr"`
	Stroke          Color            `xml:",attr"`
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
