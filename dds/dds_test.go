package dds_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"image"
	_ "image/png"
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

func checkDDSImageEqual(t *testing.T, ddsPath, comparePath string, allowDelta int64) {
	var ddsImg image.Image
	var compareImg image.Image

	{
		r, err := os.Open(ddsPath)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()

		var name string
		ddsImg, name, err = image.Decode(r)
		if err != nil {
			t.Fatal(err)
		}

		if name != "dds" {
			t.Fatalf("expected \"dds\" image, but got \"%v\"", name)
		}
	}

	{
		r, err := os.Open(comparePath)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()

		compareImg, _, err = image.Decode(r)
		if err != nil {
			t.Fatal(err)
		}
	}

	if ddsImg.Bounds() != compareImg.Bounds() {
		t.Fatal("DDS image and compare image must have the same bounds")
	}

	for y := ddsImg.Bounds().Min.Y; y < ddsImg.Bounds().Max.Y; y++ {
		for x := ddsImg.Bounds().Min.X; y < ddsImg.Bounds().Max.X; y++ {
			r0, g0, b0, a0 := ddsImg.At(x, y).RGBA()
			r1, g1, b1, a1 := compareImg.At(x, y).RGBA()
			dR, dG, dB, dA := int64(r1)-int64(r0), int64(g1)-int64(g0), int64(b1)-int64(b0), int64(a1)-int64(a0)
			if dR < 0 {
				dR = -dR
			}
			if dG < 0 {
				dG = -dG
			}
			if dB < 0 {
				dB = -dB
			}
			if dA < 0 {
				dA = -dA
			}
			// Allow slight deviations due to different mappings
			if dR > allowDelta || dG > allowDelta || dB > allowDelta || dA > allowDelta {
				t.Fatalf("DDS image and compare image are not equal (x=%v, y=%v: dds: %v, compare: %v)", x, y, [4]uint32{r0, g0, b0, a0}, [4]uint32{r1, g1, b1, a1})
			}
		}
	}
}

func TestDDSImage(t *testing.T) {
	checkDDSImageEqual(t, "testimgs/dds/testimg-bc1.dds", "testimgs/compare/testimg-bc1.png", 257)
	checkDDSImageEqual(t, "testimgs/dds/testimg-bc3.dds", "testimgs/compare/testimg-bc3.png", 257)
	checkDDSImageEqual(t, "testimgs/dds/testimg-bc4.dds", "testimgs/compare/testimg-bc4.png", 257)
	checkDDSImageEqual(t, "testimgs/dds/testimg-bc5.dds", "testimgs/compare/testimg-bc5.png", 257)
	checkDDSImageEqual(t, "testimgs/dds/testimg-bc7.dds", "testimgs/compare/testimg-bc7.png", 0)
	checkDDSImageEqual(t, "testimgs/dds/testimg-rgb8.dds", "testimgs/compare/testimg.png", 0)
	checkDDSImageEqual(t, "testimgs/dds/testimg-rgba8.dds", "testimgs/compare/testimg.png", 0)
	checkDDSImageEqual(t, "testimgs/dds/testimg-r5g6r5.dds", "testimgs/compare/testimg-r5g6r5.png", 0)
	checkDDSImageEqual(t, "testimgs/dds/testimg-l8.dds", "testimgs/compare/testimg-l8.png", 0)
}

func TestDDSMipMaps(t *testing.T) {
	r, err := os.Open("testimgs/dds/testimg-bc3.dds")
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

	testImageChecksum(t, dds.Images[0].MipMaps[0], "17a28fb962d0277240418a5f14fb5b14b1c528fcda019d0c9f69de2426886402")
	testImageChecksum(t, dds.Images[0].MipMaps[1], "dacf23b70aa1422e232c2d496c6cf845e57f3c9e3132823edb603787e843800c")
	testImageChecksum(t, dds.Images[0].MipMaps[2], "db0772a48c675b7e1ee58a85c0b7438f8fb3b9b12bd5040fc5cdadb9d44b2324")
}
