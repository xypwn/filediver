package appicons

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

//go:embed filedivericon.png
var Icon []byte

//go:embed filedivericon-16.png
var Icon16 []byte

//go:embed filedivericon-24.png
var Icon24 []byte

//go:embed filedivericon-32.png
var Icon32 []byte

//go:embed filedivericon-48.png
var Icon48 []byte

//go:embed filedivericon-64.png
var Icon64 []byte

//go:embed filedivericon-128.png
var Icon128 []byte

//go:embed filedivericon-256.png
var Icon256 []byte

var (
	iconImg    image.Image
	icon16Img  image.Image
	icon24Img  image.Image
	icon32Img  image.Image
	icon48Img  image.Image
	icon64Img  image.Image
	icon128Img image.Image
	icon256Img image.Image
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

func IconImg() image.Image    { return instance(&iconImg, Icon) }
func Icon16Img() image.Image  { return instance(&icon16Img, Icon16) }
func Icon24Img() image.Image  { return instance(&icon24Img, Icon24) }
func Icon32Img() image.Image  { return instance(&icon32Img, Icon32) }
func Icon48Img() image.Image  { return instance(&icon48Img, Icon48) }
func Icon64Img() image.Image  { return instance(&icon64Img, Icon64) }
func Icon128Img() image.Image { return instance(&icon128Img, Icon128) }
func Icon256Img() image.Image { return instance(&icon256Img, Icon256) }
