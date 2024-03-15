package converter

import (
	"errors"
	"io"
)

var ErrFileType = errors.New("converter: unknown file type")

var converters = make(map[string]converter)

type DataTypeFlag int

const (
	FlagDataMain DataTypeFlag = 1 << iota
	FlagDataStream
	FlagDataGPU
)

type converter struct {
	DataTypes DataTypeFlag
	FileType  string
	Convert   ConvertFunc
}

type ConvertFunc func(outPath string, in [3]io.ReadSeeker) error

func RegisterConverter(usedDataTypes DataTypeFlag, fileType string, convert ConvertFunc) {
	converters[fileType] = converter{
		DataTypes: usedDataTypes,
		FileType:  fileType,
		Convert:   convert,
	}
}

func Converter(fileType string) (ConvertFunc, DataTypeFlag, error) {
	conv, ok := converters[fileType]
	if !ok {
		return nil, 0, ErrFileType
	}
	return conv.Convert, conv.DataTypes, nil
}
