package unit

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

func sign32(f float32) float32 {
	if f >= 0 {
		return 1
	} else {
		return -1
	}
}

// 32-bit abs
func abs32(f float32) float32 {
	// See std math.Abs.
	return math.Float32frombits(math.Float32bits(f) &^ (1 << 31))
}

// Performs octahedral decoding of v.
// Input vector components must be in range [-1;+1].
func decodeOctahedral(v mgl32.Vec2) mgl32.Vec3 {
	x, y := v.X(), v.Y()
	z := 1 - abs32(x) - abs32(y)
	if z < 0 {
		x, y = (1-abs32(y))*sign32(x), (1-abs32(x))*sign32(y)
	}
	return (mgl32.Vec3{x, y, z}).Normalize()
}

// DecodePackedOctahedralNormal decodes an octahedrally
// encoded normal packed into the first 20 bits of v.
func DecodePackedOctahedralNormal(v uint32) mgl32.Vec3 {
	r10 := v & 0x3ff
	g10 := (v >> 10) & 0x3ff
	return decodeOctahedral(mgl32.Vec2{
		// Normalize to [-1;+1]
		float32(r10)*(2.0/1023.0) - 1,
		float32(g10)*(2.0/1023.0) - 1,
	})
}

func encodeOctahedral(v mgl32.Vec3) mgl32.Vec2 {
	x, y := v.X(), v.Y()
	l1Norm := abs32(v.X()) + abs32(v.Y()) + abs32(v.Z()) // 1 norm a.k.a. taxicab length
	x /= l1Norm
	y /= l1Norm
	if v.Z() < 0 {
		x, y = (1-abs32(y))*sign32(x), (1-abs32(x))*sign32(y)
	}
	return mgl32.Vec2{x, y}
}

// EncodePackedOctahedralNormal encodes an octahedrally
// encoded normal packed into the first 20 bits of the return value.
func EncodePackedOctahedralNormal(v mgl32.Vec3) uint32 {
	u := encodeOctahedral(v)
	return uint32((u.X()+1)*(1023.0/2.0)) |
		(uint32((u.Y()+1)*(1023.0/2.0)) << 10)
}
