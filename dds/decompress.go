package dds

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math/bits"

	"github.com/x448/float16"
)

// https://github.com/ImageMagick/ImageMagick/blob/main/coders/dds.c

var bc7ModeInfo = [8]struct {
	PartitionBits   uint8
	NumSubsets      uint8
	ColorPrecision  uint8
	AlphaPrecision  uint8
	NumPBits        uint8
	IndexPrecision  uint8
	Index2Precision uint8
}{
	{4, 3, 4, 0, 6, 3, 0},
	{6, 2, 6, 0, 2, 3, 0},
	{6, 3, 5, 0, 0, 2, 0},
	{6, 2, 7, 0, 4, 2, 0},
	{0, 1, 5, 6, 0, 2, 3},
	{0, 1, 7, 8, 0, 2, 2},
	{0, 1, 7, 7, 2, 4, 0},
	{6, 2, 5, 5, 4, 2, 0},
}

var bc7PartitionTable = [2][64][16]uint8{
	{ // Partition set for 2 subsets
		{0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1},
		{0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1},
		{0, 1, 1, 1, 0, 1, 1, 1, 0, 1, 1, 1, 0, 1, 1, 1},
		{0, 0, 0, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 1, 1},
		{0, 0, 1, 1, 0, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 1, 0, 0, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1},
		{0, 0, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1},
		{0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1},
		{0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 0, 1, 1, 1, 1},
		{0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 0},
		{0, 1, 1, 1, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0},
		{0, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 0, 1},
		{0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0},
		{0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0},
		{0, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 0},
		{0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0},
		{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 1, 1, 1, 0},
		{0, 0, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 1, 0, 0},
		{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
		{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1},
		{0, 1, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0},
		{0, 0, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0},
		{0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0},
		{0, 1, 0, 1, 0, 1, 0, 1, 1, 0, 1, 0, 1, 0, 1, 0},
		{0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 1},
		{0, 1, 0, 1, 1, 0, 1, 0, 1, 0, 1, 0, 0, 1, 0, 1},
		{0, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 1, 0},
		{0, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 0, 0, 0},
		{0, 0, 1, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 1, 0, 0},
		{0, 0, 1, 1, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 0, 0},
		{0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0},
		{0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 0, 0, 1, 1},
		{0, 1, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1},
		{0, 0, 0, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 0, 0, 0},
		{0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 0},
		{0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0},
		{0, 1, 1, 0, 1, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 1},
		{0, 0, 1, 1, 0, 1, 1, 0, 1, 1, 0, 0, 1, 0, 0, 1},
		{0, 1, 1, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0},
		{0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 0, 1, 1, 0},
		{0, 1, 1, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 0, 1},
		{0, 1, 1, 0, 0, 0, 1, 1, 0, 0, 1, 1, 1, 0, 0, 1},
		{0, 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 1},
		{0, 0, 0, 1, 1, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1},
		{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1},
		{0, 0, 1, 1, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 0, 1, 1, 1, 0},
		{0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 0, 1, 1, 1},
	},

	{ // Partition set for 3 subsets
		{0, 0, 1, 1, 0, 0, 1, 1, 0, 2, 2, 1, 2, 2, 2, 2},
		{0, 0, 0, 1, 0, 0, 1, 1, 2, 2, 1, 1, 2, 2, 2, 1},
		{0, 0, 0, 0, 2, 0, 0, 1, 2, 2, 1, 1, 2, 2, 1, 1},
		{0, 2, 2, 2, 0, 0, 2, 2, 0, 0, 1, 1, 0, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 1, 1, 2, 2},
		{0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 2, 2, 0, 0, 2, 2},
		{0, 0, 2, 2, 0, 0, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 1, 1, 0, 0, 1, 1, 2, 2, 1, 1, 2, 2, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2},
		{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2},
		{0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
		{0, 0, 1, 2, 0, 0, 1, 2, 0, 0, 1, 2, 0, 0, 1, 2},
		{0, 1, 1, 2, 0, 1, 1, 2, 0, 1, 1, 2, 0, 1, 1, 2},
		{0, 1, 2, 2, 0, 1, 2, 2, 0, 1, 2, 2, 0, 1, 2, 2},
		{0, 0, 1, 1, 0, 1, 1, 2, 1, 1, 2, 2, 1, 2, 2, 2},
		{0, 0, 1, 1, 2, 0, 0, 1, 2, 2, 0, 0, 2, 2, 2, 0},
		{0, 0, 0, 1, 0, 0, 1, 1, 0, 1, 1, 2, 1, 1, 2, 2},
		{0, 1, 1, 1, 0, 0, 1, 1, 2, 0, 0, 1, 2, 2, 0, 0},
		{0, 0, 0, 0, 1, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2, 2},
		{0, 0, 2, 2, 0, 0, 2, 2, 0, 0, 2, 2, 1, 1, 1, 1},
		{0, 1, 1, 1, 0, 1, 1, 1, 0, 2, 2, 2, 0, 2, 2, 2},
		{0, 0, 0, 1, 0, 0, 0, 1, 2, 2, 2, 1, 2, 2, 2, 1},
		{0, 0, 0, 0, 0, 0, 1, 1, 0, 1, 2, 2, 0, 1, 2, 2},
		{0, 0, 0, 0, 1, 1, 0, 0, 2, 2, 1, 0, 2, 2, 1, 0},
		{0, 1, 2, 2, 0, 1, 2, 2, 0, 0, 1, 1, 0, 0, 0, 0},
		{0, 0, 1, 2, 0, 0, 1, 2, 1, 1, 2, 2, 2, 2, 2, 2},
		{0, 1, 1, 0, 1, 2, 2, 1, 1, 2, 2, 1, 0, 1, 1, 0},
		{0, 0, 0, 0, 0, 1, 1, 0, 1, 2, 2, 1, 1, 2, 2, 1},
		{0, 0, 2, 2, 1, 1, 0, 2, 1, 1, 0, 2, 0, 0, 2, 2},
		{0, 1, 1, 0, 0, 1, 1, 0, 2, 0, 0, 2, 2, 2, 2, 2},
		{0, 0, 1, 1, 0, 1, 2, 2, 0, 1, 2, 2, 0, 0, 1, 1},
		{0, 0, 0, 0, 2, 0, 0, 0, 2, 2, 1, 1, 2, 2, 2, 1},
		{0, 0, 0, 0, 0, 0, 0, 2, 1, 1, 2, 2, 1, 2, 2, 2},
		{0, 2, 2, 2, 0, 0, 2, 2, 0, 0, 1, 2, 0, 0, 1, 1},
		{0, 0, 1, 1, 0, 0, 1, 2, 0, 0, 2, 2, 0, 2, 2, 2},
		{0, 1, 2, 0, 0, 1, 2, 0, 0, 1, 2, 0, 0, 1, 2, 0},
		{0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 0, 0, 0, 0},
		{0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0},
		{0, 1, 2, 0, 2, 0, 1, 2, 1, 2, 0, 1, 0, 1, 2, 0},
		{0, 0, 1, 1, 2, 2, 0, 0, 1, 1, 2, 2, 0, 0, 1, 1},
		{0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 0, 0, 0, 0, 1, 1},
		{0, 1, 0, 1, 0, 1, 0, 1, 2, 2, 2, 2, 2, 2, 2, 2},
		{0, 0, 0, 0, 0, 0, 0, 0, 2, 1, 2, 1, 2, 1, 2, 1},
		{0, 0, 2, 2, 1, 1, 2, 2, 0, 0, 2, 2, 1, 1, 2, 2},
		{0, 0, 2, 2, 0, 0, 1, 1, 0, 0, 2, 2, 0, 0, 1, 1},
		{0, 2, 2, 0, 1, 2, 2, 1, 0, 2, 2, 0, 1, 2, 2, 1},
		{0, 1, 0, 1, 2, 2, 2, 2, 2, 2, 2, 2, 0, 1, 0, 1},
		{0, 0, 0, 0, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1},
		{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 2, 2, 2, 2},
		{0, 2, 2, 2, 0, 1, 1, 1, 0, 2, 2, 2, 0, 1, 1, 1},
		{0, 0, 0, 2, 1, 1, 1, 2, 0, 0, 0, 2, 1, 1, 1, 2},
		{0, 0, 0, 0, 2, 1, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2},
		{0, 2, 2, 2, 0, 1, 1, 1, 0, 1, 1, 1, 0, 2, 2, 2},
		{0, 0, 0, 2, 1, 1, 1, 2, 1, 1, 1, 2, 0, 0, 0, 2},
		{0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 2, 2, 2, 2},
		{0, 0, 0, 0, 0, 0, 0, 0, 2, 1, 1, 2, 2, 1, 1, 2},
		{0, 1, 1, 0, 0, 1, 1, 0, 2, 2, 2, 2, 2, 2, 2, 2},
		{0, 0, 2, 2, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 2, 2},
		{0, 0, 2, 2, 1, 1, 2, 2, 1, 1, 2, 2, 0, 0, 2, 2},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 1, 1, 2},
		{0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 1},
		{0, 2, 2, 2, 1, 2, 2, 2, 0, 2, 2, 2, 1, 2, 2, 2},
		{0, 1, 0, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		{0, 1, 1, 1, 2, 0, 1, 1, 2, 2, 0, 1, 2, 2, 2, 0},
	},
}

var bc7AnchorIndexTable = [4][64]uint8{
	// Anchor index values for the first subset
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
	// Anchor index values for the second subset of two-subset partitioning
	{
		15, 15, 15, 15, 15, 15, 15, 15,
		15, 15, 15, 15, 15, 15, 15, 15,
		15, 2, 8, 2, 2, 8, 8, 15,
		2, 8, 2, 2, 8, 8, 2, 2,
		15, 15, 6, 8, 2, 8, 15, 15,
		2, 8, 2, 2, 2, 15, 15, 6,
		6, 2, 6, 8, 15, 15, 2, 2,
		15, 15, 15, 15, 15, 2, 2, 15,
	},
	// Anchor index values for the second subset of three-subset partitioning
	{
		3, 3, 15, 15, 8, 3, 15, 15,
		8, 8, 6, 6, 6, 5, 3, 3,
		3, 3, 8, 15, 3, 3, 6, 10,
		5, 8, 8, 6, 8, 5, 15, 15,
		8, 15, 3, 5, 6, 10, 8, 15,
		15, 3, 15, 5, 15, 15, 15, 15,
		3, 15, 5, 5, 5, 8, 5, 10,
		5, 10, 8, 13, 15, 12, 3, 3,
	},
	// Anchor index values for the third subset of three-subset partitioning
	{
		15, 8, 8, 3, 15, 15, 3, 8,
		15, 15, 15, 15, 15, 15, 15, 8,
		15, 8, 15, 3, 15, 8, 15, 8,
		3, 15, 6, 10, 15, 15, 10, 8,
		15, 3, 15, 10, 10, 8, 9, 10,
		6, 15, 8, 15, 3, 6, 6, 8,
		15, 3, 15, 15, 15, 15, 15, 15,
		15, 15, 15, 15, 3, 15, 15, 8,
	},
}

var bc7Weight2 = []uint8{0, 21, 43, 64}
var bc7Weight3 = []uint8{0, 9, 18, 27, 37, 46, 55, 64}
var bc7Weight4 = []uint8{0, 4, 9, 13, 17, 21, 26, 30, 34,
	38, 43, 47, 51, 55, 60, 64}

type DecompressFunc func(buf []uint8, r io.Reader, width, height int, info Info) error

func avgU8(xs ...uint8) uint8 {
	var sum int
	for _, x := range xs {
		sum += int(x)
	}
	return uint8(sum / len(xs))
}

func DecompressUncompressed(buf []uint8, r io.Reader, width, height int, info Info) error {
	bitMasks := [4]uint32{info.Header.PixelFormat.RBitMask, info.Header.PixelFormat.GBitMask, info.Header.PixelFormat.BBitMask, info.Header.PixelFormat.ABitMask}
	var bitMaskTZs [4]int
	for i := range bitMasks {
		bitMaskTZs[i] = bits.TrailingZeros32(bitMasks[i])
	}
	var bitMaskBits [4]int
	for i := range bitMasks {
		bitMaskBits[i] = bits.Len32(bitMasks[i] >> bitMaskTZs[i])
	}

	if info.Header.PixelFormat.RGBBitCount%8 != 0 {
		return fmt.Errorf("invalid RGB bit count: %v (must be multiple of 8)", info.Header.PixelFormat.RGBBitCount)
	}
	byteCount := info.Header.PixelFormat.RGBBitCount / 8
	if byteCount > 32 {
		return fmt.Errorf("invalid RGB bit count: %v (must be at most 32)", byteCount)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var stride int
			switch info.ColorModel {
			case color.GrayModel:
				stride = 1
			case color.Gray16Model:
				stride = 2
			case color.NRGBAModel:
				stride = 4
			case color.NRGBA64Model:
				stride = 8
			default:
				return errors.New("uncompressed image: unexpected color model")
			}
			idx := stride * (y*width + x)

			if info.Header.PixelFormat.Flags&PixelFormatFlagAlphaPixels == 0 {
				if info.ColorModel == color.NRGBAModel {
					buf[idx+3] = 0xff
				} else if info.ColorModel == color.NRGBA64Model {
					binary.BigEndian.PutUint16(buf[idx+6:], 0xffff)
				}
			}

			var data [4]uint8
			if _, err := io.ReadFull(r, data[:byteCount]); err != nil {
				return err
			}
			dataU32 := binary.LittleEndian.Uint32(data[:])

			offs := 0
			for i := range bitMasks {
				c := (dataU32 & bitMasks[i]) >> bitMaskTZs[i]
				if bitMaskBits[i] == 0 {
				} else if bitMaskBits[i] <= 8 {
					var v uint8
					switch bitMaskBits[i] {
					case 1:
						v = mapBits1To8(uint16(c))
					case 2:
						v = mapBits2To8(uint16(c))
					case 3:
						v = mapBits3To8(uint16(c))
					case 4:
						v = mapBits4To8(uint16(c))
					case 5:
						v = mapBits5To8(uint16(c))
					case 6:
						v = mapBits6To8(uint16(c))
					case 7:
						v = mapBits7To8(uint16(c))
					case 8:
						v = uint8(c)
					default:
						return fmt.Errorf("unsupported number of bits: %v", bitMaskBits[i])
					}
					buf[idx+offs] = v
					offs += 1
				} else if bitMaskBits[i] <= 16 {
					var v uint16
					switch bitMaskBits[i] {
					case 16:
						v = uint16(c)
					default:
						return fmt.Errorf("unsupported number of bits: %v", bitMaskBits[i])
					}
					binary.BigEndian.PutUint16(buf[idx+offs:], v)
					offs += 2
				} else {
					return fmt.Errorf("unsupported number of bits: %v", bitMaskBits[i])
				}
				if offs > stride {
					return fmt.Errorf("offset (%v) is larger than stride (%v)", offs, stride)
				}
			}
		}
	}
	return nil
}

func DecompressUncompressedDXT10(buf []uint8, r io.Reader, width, height int, info Info) error {
	if info.DXT10Header == nil {
		return errors.New("uncompressed DXT 10: expected DXT10 header")
	}
	var translatePixel func(idx int) error
	switch info.DXT10Header.DXGIFormat {
	case DXGIFormatR32G32B32A32Float:
		if info.ColorModel != color.NRGBA64Model {
			return errors.New("expected NRGBA64 model for R32G32B32A32Float")
		}
		translatePixel = func(idx int) error {
			for i := 0; i < 4; i++ {
				var v float32
				if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
					return err
				}
				binary.LittleEndian.PutUint16(buf[idx+2*i:], uint16(v*0xffff))
			}
			return nil
		}
	case DXGIFormatR16G16B16A16Float:
		if info.ColorModel != color.NRGBA64Model {
			return errors.New("expected NRGBA64 model for R16G16B16A16Float")
		}
		translatePixel = func(idx int) error {
			for i := 0; i < 4; i++ {
				var v uint16
				if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
					return err
				}
				binary.LittleEndian.PutUint16(buf[idx+2*i:], uint16(float16.Frombits(v).Float32()*0xffff))
			}
			return nil
		}
	case DXGIFormatR32Float:
		if info.ColorModel != color.Gray16Model {
			return errors.New("expected Gray16 model for R32Float")
		}
		translatePixel = func(idx int) error {
			var v float32
			if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
				return err
			}
			binary.LittleEndian.PutUint16(buf[idx:], uint16(v*0xffff))
			return nil
		}
	case DXGIFormatR8G8B8A8UNorm:
		if info.ColorModel != color.NRGBAModel {
			return errors.New("expected NRGBA model for R8G8B8A8UNorm")
		}
		translatePixel = func(idx int) error {
			if _, err := io.ReadFull(r, buf[idx:idx+4]); err != nil {
				return err
			}
			return nil
		}
	case DXGIFormatR16UNorm:
		if info.ColorModel != color.Gray16Model {
			return errors.New("expected Gray16 model model for R16UNorm")
		}
		translatePixel = func(idx int) error {
			var v uint16
			if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
				return err
			}
			binary.BigEndian.PutUint16(buf[idx:], v)
			return nil
		}
	case DXGIFormatR8UNorm:
		if info.ColorModel != color.GrayModel {
			return errors.New("expected Gray model model for R8UNorm")
		}
		translatePixel = func(idx int) error {
			if _, err := io.ReadFull(r, buf[idx:idx+1]); err != nil {
				return err
			}
			return nil
		}
	default:
		return fmt.Errorf("uncompressed image: unsupported DXGI format: %v", info.DXT10Header.DXGIFormat)
	}
	var stride int
	switch info.ColorModel {
	case color.GrayModel:
		stride = 1
	case color.Gray16Model:
		stride = 2
	case color.NRGBAModel:
		stride = 4
	case color.NRGBA64Model:
		stride = 8
	default:
		return errors.New("uncompressed image: unexpected color model")
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := stride * (y*width + x)
			if err := translatePixel(idx); err != nil {
				return err
			}
		}
	}
	return nil
}

func calculateDXTColors(c0 uint16, c1 uint16, ignoreAlpha bool) (r [4]uint8, g [4]uint8, b [4]uint8, a [4]uint8) {
	r[0], g[0], b[0] = colorR5G6B5ToRGB(c0)
	r[1], g[1], b[1] = colorR5G6B5ToRGB(c1)

	if ignoreAlpha || c0 > c1 {
		r[2] = avgU8(r[0], r[0], r[1])
		g[2] = avgU8(g[0], g[0], g[1])
		b[2] = avgU8(b[0], b[0], b[1])

		r[3] = avgU8(r[0], r[1], r[1])
		g[3] = avgU8(g[0], g[1], g[1])
		b[3] = avgU8(b[0], b[1], b[1])
	} else {
		r[2] = avgU8(r[0], r[1])
		g[2] = avgU8(g[0], g[1])
		b[2] = avgU8(b[0], b[1])

		a[3] = 255
	}
	return
}

func DecompressDXT1(buf []uint8, r io.Reader, width, height int, info Info) error {
	if info.ColorModel != color.NRGBAModel {
		return errors.New("DXT1 compression expects NRGBA color model")
	}

	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			var data [8]uint8
			if _, err := io.ReadFull(r, data[:]); err != nil {
				return err
			}

			c0 := binary.LittleEndian.Uint16(data[:2])
			c1 := binary.LittleEndian.Uint16(data[2:4])
			bits := binary.LittleEndian.Uint32(data[4:8])

			cR, cG, cB, cA := calculateDXTColors(c0, c1, false)

			for j := 0; j < 4; j++ {
				for i := 0; i < 4; i++ {
					if x+i >= width || y+j >= height {
						continue
					}
					idx := 4 * ((y+j)*width + (x + i))

					code := (bits >> ((j*4 + i) * 2)) & 0x03
					buf[idx+0] = cR[code]
					buf[idx+1] = cG[code]
					buf[idx+2] = cB[code]
					buf[idx+3] = 255
					if cA[code] != 0 {
						return errors.New("expected alpha to be 0 in DXT 1 compressed image")
					}
				}
			}
		}
	}
	return nil
}

func DecompressDXT5(buf []uint8, r io.Reader, width, height int, info Info) error {
	if info.ColorModel != color.NRGBAModel {
		return errors.New("DXT5 compression expects NRGBA color model")
	}

	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			var data [16]uint8
			if _, err := io.ReadFull(r, data[:]); err != nil {
				return err
			}
			a0, a1 := data[0], data[1]
			alphaBits := uint64(binary.LittleEndian.Uint32(data[2:6]))
			alphaBits |= uint64(binary.LittleEndian.Uint16(data[6:8])) << 32
			c0, c1 := binary.LittleEndian.Uint16(data[8:10]), binary.LittleEndian.Uint16(data[10:12])
			bits := binary.LittleEndian.Uint32(data[12:16])

			cR, cG, cB, _ := calculateDXTColors(c0, c1, true)

			for j := 0; j < 4; j++ {
				for i := 0; i < 4; i++ {
					if x+i >= width || y+j >= height {
						continue
					}
					idx := 4 * ((y+j)*width + (x + i))

					code := (bits >> ((4*j + i) * 2)) & 0x3
					buf[idx+0] = cR[code]
					buf[idx+1] = cG[code]
					buf[idx+2] = cB[code]

					alphaCode := (alphaBits >> (3 * (4*j + i))) & 0x7
					var alpha uint8
					if alphaCode == 0 {
						alpha = a0
					} else if alphaCode == 1 {
						alpha = a1
					} else if a0 > a1 {
						alpha = uint8(((8-alphaCode)*uint64(a0) + (alphaCode-1)*uint64(a1)) / 7)
					} else if alphaCode == 6 {
						alpha = 0
					} else if alphaCode == 7 {
						alpha = 255
					} else {
						alpha = uint8(((6-alphaCode)*uint64(a0) + (alphaCode)*uint64(a1)) / 5)
					}
					buf[idx+3] = alpha
				}
			}
		}
	}
	return nil
}

func decode3DcBlock(data []uint8) [8]uint8 {
	if len(data) < 8 {
		panic("data must be of length 8 or more")
	}

	var c [8]uint8
	c[0], c[1] = data[0], data[1]

	mode := 4
	if c[0] > c[1] {
		mode = 6
	}
	for i := 0; i < mode; i++ {
		c[i+2] = uint8(
			(float64((mode-i))*float64(c[0]) + float64(i+1)*float64(c[1])) /
				float64(mode+1))
	}
	if mode == 4 {
		c[6] = 0
		c[7] = 255
	}

	return c
}

func getBit(data []uint8, startBit *uint64) bool {
	index := (*startBit) >> 3
	base := (*startBit) - (index << 3)
	(*startBit)++
	if index >= uint64(len(data)) {
		return false
	}
	return (data[index]>>base)&1 != 0
}

func getBits(data []uint8, startBit *uint64, numBits uint8) uint8 {
	index := (*startBit) >> 3
	base := (*startBit) - (index << 3)
	if index >= uint64(len(data)) {
		return 0
	}
	var res uint8
	if base+uint64(numBits) > 8 {
		firstBits := 8 - base
		nextBits := uint64(numBits) - firstBits
		res = ((data[index] >> base) |
			((data[index+1] & ((1 << nextBits) - 1)) << firstBits))
	} else {
		res = (data[index] >> base) & ((1 << numBits) - 1)
	}
	*startBit += uint64(numBits)
	return res
}

func Decompress3DcPlus(buf []uint8, r io.Reader, width, height int, info Info) error {
	if info.ColorModel != color.GrayModel {
		return errors.New("3Dc+ compression expects gray color model")
	}

	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			var data [8]uint8
			if _, err := io.ReadFull(r, data[:]); err != nil {
				return err
			}

			c := decode3DcBlock(data[:])

			startBit := uint64(16)
			for j := 0; j < 4; j++ {
				for i := 0; i < 4; i++ {
					lum := c[getBits(data[:], &startBit, 3)]

					if x+i >= width || y+j >= height {
						continue
					}

					idx := (y+j)*width + (x + i)
					buf[idx] = lum
				}
			}
		}
	}
	return nil
}

func Decompress3Dc(buf []uint8, r io.Reader, width, height int, info Info) error {
	if info.ColorModel != color.NRGBAModel {
		return errors.New("3Dc compression expects NRGBA color model")
	}

	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			var data [16]uint8
			if _, err := io.ReadFull(r, data[:]); err != nil {
				return err
			}

			cR := decode3DcBlock(data[:8])
			cG := decode3DcBlock(data[8:])

			startBitR := uint64(16)
			startBitG := uint64(80)
			for j := 0; j < 4; j++ {
				for i := 0; i < 4; i++ {
					r := cR[getBits(data[:], &startBitR, 3)]
					g := cG[getBits(data[:], &startBitG, 3)]

					if x+i >= width || y+j >= height {
						continue
					}

					idx := 4 * ((y+j)*width + (x + i))
					buf[idx+0] = r
					buf[idx+1] = g
					buf[idx+2] = 0
					buf[idx+3] = 255
				}
			}
		}
	}
	return nil
}

func readBC7Endpoints(data [16]uint8, mode uint64, startBit *uint64) (r [6]uint8, g [6]uint8, b [6]uint8, a [6]uint8) {
	numSubsets := bc7ModeInfo[mode].NumSubsets
	colorBits := bc7ModeInfo[mode].ColorPrecision

	for i := 0; i < int(numSubsets)*2; i++ {
		r[i] = getBits(data[:], startBit, colorBits)
	}
	for i := 0; i < int(numSubsets)*2; i++ {
		g[i] = getBits(data[:], startBit, colorBits)
	}
	for i := 0; i < int(numSubsets)*2; i++ {
		b[i] = getBits(data[:], startBit, colorBits)
	}

	alphaBits := bc7ModeInfo[mode].AlphaPrecision
	hasAlpha := mode >= 4
	if hasAlpha {
		for i := 0; i < int(numSubsets)*2; i++ {
			a[i] = getBits(data[:], startBit, alphaBits)
		}
	} else {
		for i := 0; i < int(numSubsets)*2; i++ {
			a[i] = 255
		}
	}

	hasPBits := mode == 0 || mode == 1 || mode == 3 || mode == 6 || mode == 7

	if hasPBits {
		for i := 0; i < int(numSubsets)*2; i++ {
			r[i] <<= 1
			g[i] <<= 1
			b[i] <<= 1
			a[i] <<= 1
		}

		if mode == 1 {
			pBit0 := getBit(data[:], startBit)
			pBit1 := getBit(data[:], startBit)

			if pBit0 {
				r[0] |= 1
				g[0] |= 1
				b[0] |= 1
				r[1] |= 1
				g[1] |= 1
				b[1] |= 1
			}

			if pBit1 {
				r[2] |= 1
				g[2] |= 1
				b[2] |= 1
				r[3] |= 1
				g[3] |= 1
				b[3] |= 1
			}
		} else {
			for i := 0; i < int(numSubsets)*2; i++ {
				pBit := getBit(data[:], startBit)
				if pBit {
					r[i] |= 1
					g[i] |= 1
					b[i] |= 1
					a[i] |= 1
				}
			}
		}

		colorBits++
		alphaBits++
	}

	for i := 0; i < int(numSubsets)*2; i++ {
		r[i] <<= (8 - colorBits)
		g[i] <<= (8 - colorBits)
		b[i] <<= (8 - colorBits)
		a[i] <<= (8 - alphaBits)

		r[i] |= r[i] >> colorBits
		g[i] |= g[i] >> colorBits
		b[i] |= b[i] >> colorBits
		a[i] |= a[i] >> alphaBits
	}

	if !hasAlpha {
		for i := 0; i < int(numSubsets)*2; i++ {
			a[i] = 255
		}
	}

	return
}

func getBC7SubsetIndex(numSubsets, partitionID uint8, pixelIndex int) uint8 {
	if numSubsets == 2 {
		return bc7PartitionTable[0][partitionID][pixelIndex]
	}
	if numSubsets == 3 {
		return bc7PartitionTable[1][partitionID][pixelIndex]
	}
	return 0
}

func isBC7PixelAnchorIndex(subsetIndex, numSubsets uint8, pixelIndex int, partitionID uint8) bool {
	tableIndex := 0
	if subsetIndex == 0 {
		tableIndex = 0
	} else if subsetIndex == 1 && numSubsets == 2 {
		tableIndex = 1
	} else if subsetIndex == 1 && numSubsets == 3 {
		tableIndex = 2
	} else {
		tableIndex = 3
	}

	return int(bc7AnchorIndexTable[tableIndex][partitionID]) == pixelIndex
}

func DecompressBC7(buf []uint8, r io.Reader, width, height int, info Info) error {
	if info.ColorModel != color.NRGBAModel {
		return errors.New("BC7 compression expects NRGBA color model")
	}

	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			var data [16]uint8
			if _, err := io.ReadFull(r, data[:]); err != nil {
				return err
			}

			startBit := uint64(0)
			for startBit <= 8 && !getBit(data[:], &startBit) {
			}
			mode := startBit - 1

			if mode > 7 {
				return fmt.Errorf("invalid mode: %v", mode)
			}

			numSubsets := bc7ModeInfo[mode].NumSubsets
			partitionID := uint8(0)

			if mode == 0 || mode == 1 || mode == 2 || mode == 3 || mode == 7 {
				partitionID = getBits(data[:], &startBit, bc7ModeInfo[mode].PartitionBits)
				if partitionID > 63 {
					return fmt.Errorf("invalid partition ID: %v", partitionID)
				}
			}

			rotation := uint8(0)
			if mode == 4 || mode == 5 {
				rotation = getBits(data[:], &startBit, 2)
			}

			selectorBit := false
			if mode == 4 {
				selectorBit = getBit(data[:], &startBit)
			}

			cR, cG, cB, cA := readBC7Endpoints(data, mode, &startBit)

			indexPrec := bc7ModeInfo[mode].IndexPrecision
			index2Prec := bc7ModeInfo[mode].Index2Precision

			var alphaIndices [16]uint8
			if mode == 4 && selectorBit {
				indexPrec = 3
				if getBit(data[:], &startBit) {
					alphaIndices[0] = 1
				}
				for i := 1; i < 16; i++ {
					alphaIndices[i] = getBits(data[:], &startBit, 2)
				}
			}

			var numBits uint8
			var subsetIndices [16]uint8
			var colorIndices [16]uint8
			for i := 0; i < 16; i++ {
				subsetIndices[i] = getBC7SubsetIndex(numSubsets, partitionID, i)
				numBits = indexPrec
				if isBC7PixelAnchorIndex(subsetIndices[i], numSubsets, i, partitionID) {
					numBits--
				}
				colorIndices[i] = getBits(data[:], &startBit, numBits)
			}

			if mode == 5 || (mode == 4 && !selectorBit) {
				alphaIndices[0] = getBits(data[:], &startBit, index2Prec-1)
				for i := 1; i < 16; i++ {
					alphaIndices[i] = getBits(data[:], &startBit, index2Prec)
				}
			}

			var areaW, areaH int
			if width-x < 4 {
				areaW = width - x
			} else {
				areaW = 4
			}
			if height-y < 4 {
				areaH = height - y
			} else {
				areaH = 4
			}
			for i := 0; i < areaW*areaH; i++ {
				c0 := 2 * subsetIndices[i]
				c1 := 2*subsetIndices[i] + 1
				c2 := colorIndices[i]

				weight := uint8(64)
				switch indexPrec {
				case 2:
					if int(c2) < len(bc7Weight2) {
						weight = bc7Weight2[c2]
					}
				case 3:
					if int(c2) < len(bc7Weight3) {
						weight = bc7Weight3[c2]
					}
				default:
					if int(c2) < len(bc7Weight4) {
						weight = bc7Weight4[c2]
					}
				}

				r := uint8(((64-int(weight))*int(cR[c0]) + int(weight)*int(cR[c1]) + 32) >> 6)
				g := uint8(((64-int(weight))*int(cG[c0]) + int(weight)*int(cG[c1]) + 32) >> 6)
				b := uint8(((64-int(weight))*int(cB[c0]) + int(weight)*int(cB[c1]) + 32) >> 6)
				a := uint8(((64-int(weight))*int(cA[c0]) + int(weight)*int(cA[c1]) + 32) >> 6)

				if mode == 4 || mode == 5 {
					a0 := alphaIndices[i]
					if int(a0) < len(bc7Weight2) {
						weight = bc7Weight2[a0]
					}
					if mode == 4 && !selectorBit && int(a0) < len(bc7Weight3) {
						weight = bc7Weight3[a0]
					}
					if c0 < 6 && c1 < 6 {
						a = uint8(((64-int(weight))*int(cA[c0]) + int(weight)*int(cA[c1]) + 32) >> 6)
					}
				}

				switch rotation {
				case 1:
					a, r = r, a
				case 2:
					a, g = g, a
				case 3:
					a, b = b, a
				}

				areaX, areaY := i%areaW, i/areaW
				idx := 4 * ((y+areaY)*width + (x + areaX))
				buf[idx+0] = r
				buf[idx+1] = g
				buf[idx+2] = b
				buf[idx+3] = a
			}
		}
	}
	return nil
}
