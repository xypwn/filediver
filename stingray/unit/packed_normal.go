package unit

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

func sign[T ~float32 | float64](f T) T {
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
		x, y = (1-abs32(y))*sign(x), (1-abs32(x))*sign(y)
	}
	return (mgl32.Vec3{x, y, z}).Normalize()

	/* Stingray's implementation (should be equivalent):
	x, y := v.X(), v.Y()
	z := 1 - abs32(x) - abs32(y)
	a := mgl32.Clamp(-z, 0, 1)
	return (mgl32.Vec3{
		x - a*sign(x),
		y - a*sign(y),
		z,
	}).Normalize()*/
}

func encodeOctahedral(v mgl32.Vec3) mgl32.Vec2 {
	x, y := v.X(), v.Y()
	l1Norm := abs32(v.X()) + abs32(v.Y()) + abs32(v.Z()) // 1 norm a.k.a. taxicab length
	x /= l1Norm
	y /= l1Norm
	if v.Z() < 0 {
		x, y = (1-abs32(y))*sign(x), (1-abs32(x))*sign(y)
	}
	return mgl32.Vec2{x, y}
}

func DecodePackedNormal(v uint32) (normal, tangent, bitangent mgl32.Vec3) {
	// See https://www.jeremyong.com/graphics/2023/01/09/tangent-spaces-and-diamond-encoding/

	r10 := v & 0x3ff         // packed normal X
	g10 := (v >> 10) & 0x3ff // packed normal Y
	b10 := (v >> 20) & 0x3ff // packed tangent rotation
	a2 := v >> 30 & 0x3      // packed is tangent flipped

	normal = decodeOctahedral(mgl32.Vec2{
		// Normalize to [-1;+1]
		float32(r10)*(2.0/1023.0) - 1,
		float32(g10)*(2.0/1023.0) - 1,
	})

	rot := float32(b10) * (1 / 1023.0)

	// Pre-computed cos and (-sin) of the rotation angle
	var catheti mgl32.Vec2
	if rot < 0.5 {
		catheti[0] = 4*rot - 1             // cos
		catheti[1] = abs32(catheti[0]) - 1 // -sin
	} else {
		catheti[0] = 3 - 4*rot             // cos
		catheti[1] = 1 - abs32(catheti[0]) // -sin
	}
	catheti = catheti.Normalize()

	tangentBaseChoice := abs32(normal.Z()) < abs32(normal.Y())

	// The following if statement is NOT in the actual shader.
	// This is to correct some special cases, which probably
	// came to be because the shader uses float16 and we're
	// using float32.
	if mgl32.FloatEqualThreshold(normal.X(), -1, 1e-6) ||
		mgl32.FloatEqualThreshold(normal.X(), 1, 1e-6) {
		tangentBaseChoice = !tangentBaseChoice
	}

	var tangentBase mgl32.Vec3 // tangentBase is orthogonal to normal
	if tangentBaseChoice {
		tangentBase = (mgl32.Vec3{normal.Y(), -normal.X(), 0}).Normalize()
	} else {
		tangentBase = (mgl32.Vec3{normal.Z(), 0, -normal.X()}).Normalize()
	}

	// Rotate tangentBase around normal.
	// See https://en.wikipedia.org/wiki/Rodrigues%27_rotation_formula
	// (last term omitted, since tangentBase is orthogonal to normal).
	tangent = tangentBase.Mul(catheti.X()).
		Add(tangentBase.Cross(normal).Mul(catheti.Y()))

	bitangent = normal.Cross(tangent)

	if a2 == 3 {
		bitangent = (mgl32.Vec3{}).Sub(bitangent)
	}

	return
}
