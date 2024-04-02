package dds_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/xypwn/filediver/dds"
)

func testImageChecksum(t *testing.T, img image.Image, expectedSumHexStr string) {
	expectedSum, err := hex.DecodeString(expectedSumHexStr)
	if err != nil {
		t.Fatal(err)
	}

	hash := sha256.New()
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; y < img.Bounds().Max.X; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			if err := binary.Write(hash, binary.BigEndian, [4]uint32{r, g, b, a}); err != nil {
				t.Fatal(err)
			}
		}
	}
	sum := hash.Sum(nil)
	if !bytes.Equal(sum, expectedSum) {
		t.Fatalf("invalid image checksum: expected %x, but got %x", expectedSum, sum)
	}
}

func testDDSImage(t *testing.T, path string, checksumHex string, save bool) {
	r, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	img, name, err := image.Decode(r)
	if err != nil {
		t.Fatal(err)
	}

	if name != "dds" {
		t.Fatalf("expected \"dds\" image, but got \"%v\"", name)
	}

	if save {
		w, err := os.Create("out.png")
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(w, img); err != nil {
			t.Fatal(err)
		}
	}

	testImageChecksum(t, img, checksumHex)
}

func TestDDSImage(t *testing.T) {
	testDDSImage(t, "testimg-bc1.dds", "079b4749d42c07f36bc6daa7bb2f5476beca92f38f4798e4c98e86624a50d931", false)
	testDDSImage(t, "testimg-bc3.dds", "b8127ddcbddd112914bf0a70c8a7116ec311d3f17e5773177ccc403ff610ca6a", false)
	testDDSImage(t, "testimg-bc4.dds", "26587032b504ca06724a35e9cb437895ce6e6e491d3a6245089cd396888224c2", false)
	testDDSImage(t, "testimg-bc5.dds", "449e0bb16584f6174218c10d7401bd79feff5cffe71d7c28b9fea16d5e6e4daa", false)
	testDDSImage(t, "testimg-rgb8.dds", "17a28fb962d0277240418a5f14fb5b14b1c528fcda019d0c9f69de2426886402", false)
	testDDSImage(t, "testimg-rgba8.dds", "17a28fb962d0277240418a5f14fb5b14b1c528fcda019d0c9f69de2426886402", false)
	testDDSImage(t, "testimg-r5g6r5.dds", "dda7c4a7d79e36aa746929c88de36311797d6c64ef35e3b604366f7d8ee9dafc", false)
	testDDSImage(t, "testimg-l8.dds", "b2c503dfcccd074d59dd1fa344053250bec3c84eee8b689617097c8c90e28bbe", false)
}

func TestDDSMipMaps(t *testing.T) {
	r, err := os.Open("testimg-bc3.dds")
	if err != nil {
		t.Fatal(err)
	}

	dds, err := dds.Decode(r, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(dds.Images) != 1 {
		t.Fatalf("expected 1 image, but got %v", len(dds.Images))
	}

	if len(dds.Images[0].MipMaps) != 10 {
		t.Fatalf("expected 10 mipmap, but got %v", len(dds.Images[0].MipMaps))
	}

	testImageChecksum(t, dds.Images[0].MipMaps[0], "b8127ddcbddd112914bf0a70c8a7116ec311d3f17e5773177ccc403ff610ca6a")
	testImageChecksum(t, dds.Images[0].MipMaps[1], "293c4be8a6c13bdedf3b8ff5f81d5fc8fe82c468277737f32e6e6035cd4169a6")
	testImageChecksum(t, dds.Images[0].MipMaps[2], "1a7ab642e80bf2a3634b5df75b2b0d1ce4c15047626115a3f33c9b7db8a21ae3")
}
