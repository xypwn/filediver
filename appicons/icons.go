package appicons

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

//go:embed filedivericon-cropped.png
var Cropped []byte

//go:embed filedivericon-cropped-16.png
var Cropped16 []byte

//go:embed filedivericon-cropped-24.png
var Cropped24 []byte

//go:embed filedivericon-cropped-32.png
var Cropped32 []byte

//go:embed filedivericon-cropped-48.png
var Cropped48 []byte

//go:embed filedivericon-cropped-64.png
var Cropped64 []byte

//go:embed filedivericon-cropped-128.png
var Cropped128 []byte

//go:embed filedivericon-256.png
var Icon256 []byte

var (
	croppedImg    image.Image
	cropped16Img  image.Image
	cropped24Img  image.Image
	cropped32Img  image.Image
	cropped48Img  image.Image
	cropped64Img  image.Image
	cropped128Img image.Image
	icon256Img    image.Image
)

func instance(dst *image.Image, src []byte) image.Image {
	if *dst != nil {
		return *dst
	}
	newImg, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		panic(err)
	}
	*dst = newImg
	return *dst
}

func CroppedImg() image.Image    { return instance(&croppedImg, Cropped) }
func Cropped16Img() image.Image  { return instance(&cropped16Img, Cropped16) }
func Cropped24Img() image.Image  { return instance(&cropped24Img, Cropped24) }
func Cropped32Img() image.Image  { return instance(&cropped32Img, Cropped32) }
func Cropped48Img() image.Image  { return instance(&cropped48Img, Cropped48) }
func Cropped64Img() image.Image  { return instance(&cropped64Img, Cropped64) }
func Cropped128Img() image.Image { return instance(&cropped128Img, Cropped128) }
func Icon256Img() image.Image    { return instance(&icon256Img, Icon256) }
