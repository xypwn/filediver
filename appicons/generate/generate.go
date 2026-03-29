package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"

	"golang.org/x/image/draw"
)

func scale(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func main() {
	srcB, err := os.ReadFile("filedivericon-cropped.png")
	if err != nil {
		panic(err)
	}
	src, err := png.Decode(bytes.NewReader(srcB))
	if err != nil {
		panic(err)
	}
	for _, size := range []int{16, 24, 32, 48, 64, 128} {
		var b bytes.Buffer
		if err := png.Encode(&b, scale(src, size, size)); err != nil {
			panic(err)
		}
		if err := os.WriteFile(fmt.Sprintf("filedivericon-cropped-%d.png", size), b.Bytes(), 0666); err != nil {
			panic(err)
		}
	}
}
