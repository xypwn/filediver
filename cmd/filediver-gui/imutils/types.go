package imutils

import (
	"reflect"

	"github.com/AllenDang/cimgui-go/imgui"
)

var dtypMapFromGo = map[reflect.Kind]imgui.DataType{
	reflect.Int8:    imgui.DataTypeS8,
	reflect.Uint8:   imgui.DataTypeU8,
	reflect.Int16:   imgui.DataTypeS16,
	reflect.Uint16:  imgui.DataTypeU16,
	reflect.Int32:   imgui.DataTypeS32,
	reflect.Uint32:  imgui.DataTypeU32,
	reflect.Int64:   imgui.DataTypeS64,
	reflect.Uint64:  imgui.DataTypeU64,
	reflect.Float32: imgui.DataTypeFloat,
	reflect.Float64: imgui.DataTypeDouble,
	// Bool and String intentionally left out,
	// since they're supported by scalar widgets
	// anyway.
}

func init() {
	if (^uint(0) >> 63) != 0 {
		dtypMapFromGo[reflect.Int] = imgui.DataTypeS64
		dtypMapFromGo[reflect.Uint] = imgui.DataTypeU64
	} else {
		dtypMapFromGo[reflect.Int] = imgui.DataTypeS32
		dtypMapFromGo[reflect.Uint] = imgui.DataTypeU32
	}
}

func DataTypeFromGo(kind reflect.Kind) (dtyp imgui.DataType, ok bool) {
	dtyp, ok = dtypMapFromGo[kind]
	if !ok {
		dtyp = -1
	}
	return
}
