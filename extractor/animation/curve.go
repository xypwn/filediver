package animation

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type VectorKeyframe struct {
	Time   float32
	Vector mgl32.Vec3
}

type QuaternionKeyframe struct {
	Time       float32
	Quaternion mgl32.Quat
}

type VectorCurve struct {
	Keyframes     []VectorKeyframe
	Duration      float32
	HermiteFrames [4]VectorKeyframe
	CurrentIndex  int
}

func (c *VectorCurve) ShiftFrame() {
	c.HermiteFrames[0] = c.HermiteFrames[1]
	c.HermiteFrames[1] = c.HermiteFrames[2]
	c.HermiteFrames[2] = c.HermiteFrames[3]
	if c.CurrentIndex < len(c.Keyframes) {
		c.HermiteFrames[3] = c.Keyframes[c.CurrentIndex]
		c.CurrentIndex++
	}
}

// Samples the curve into duration * framerate keyframes, spaced evenly along the curve
func (c *VectorCurve) Sample(framerate int) []VectorKeyframe {
	toReturn := make([]VectorKeyframe, 0)
	toReturn = append(toReturn, c.Keyframes[0])
	if len(c.Keyframes) == 1 {
		toReturn = append(toReturn, VectorKeyframe{
			Time:   c.Duration,
			Vector: c.Keyframes[0].Vector,
		})
		return toReturn
	}
	// T[i]   = 0.5 * (P[i+1] - P[i-1])
	// T[i+1] = 0.5 * (P[i+2] - P[i])
	c.HermiteFrames = [4]VectorKeyframe{
		c.Keyframes[0], // P[i-1]
		c.Keyframes[0], // P[i]
		c.Keyframes[0], // P[i+1]
		c.Keyframes[0], // P[i+2]
	}
	c.CurrentIndex = 1
	c.ShiftFrame()
	c.ShiftFrame()
	totalCount := int(math.Ceil(float64(framerate) * float64(c.Duration)))
	frameTime := float32(1.0 / float64(framerate))
	for i := 0; i < totalCount; i++ {
		currentTime := float32(i) * frameTime
		for currentTime > c.HermiteFrames[2].Time && c.HermiteFrames[1].Time != c.HermiteFrames[2].Time {
			c.ShiftFrame()
		}
		var s float32
		if c.HermiteFrames[1].Time == c.HermiteFrames[2].Time {
			s = 1.0
		} else {
			s = (currentTime - c.HermiteFrames[1].Time) / (c.HermiteFrames[2].Time - c.HermiteFrames[1].Time)
		}
		p1 := c.HermiteFrames[1].Vector
		p2 := c.HermiteFrames[2].Vector
		t1 := c.HermiteFrames[2].Vector.Sub(c.HermiteFrames[0].Vector).Mul(0.5)
		t2 := c.HermiteFrames[3].Vector.Sub(c.HermiteFrames[1].Vector).Mul(0.5)

		s2 := s * s
		s3 := s * s2

		h1 := 2*s3 - 3*s2 + 1
		h2 := -2*s3 + 3*s2
		h3 := s3 - 2*s2 + s
		h4 := s3 - s2

		p := p1.Mul(h1).Add(p2.Mul(h2)).Add(t1.Mul(h3)).Add(t2.Mul(h4))
		toReturn = append(toReturn, VectorKeyframe{
			Time:   currentTime,
			Vector: p,
		})
	}
	return toReturn
}

type QuaternionCurve struct {
	Keyframes     []QuaternionKeyframe
	Duration      float32
	HermiteFrames [4]QuaternionKeyframe
	CurrentIndex  int
}

func (c *QuaternionCurve) ShiftFrame() {
	c.HermiteFrames[0] = c.HermiteFrames[1]
	c.HermiteFrames[1] = c.HermiteFrames[2]
	c.HermiteFrames[2] = c.HermiteFrames[3]
	if c.CurrentIndex < len(c.Keyframes) {
		c.HermiteFrames[3] = c.Keyframes[c.CurrentIndex]
		c.CurrentIndex++
	}
}

// Samples the curve into duration * framerate keyframes, spaced evenly along the curve
func (c *QuaternionCurve) Sample(framerate int) []QuaternionKeyframe {
	toReturn := make([]QuaternionKeyframe, 0)
	toReturn = append(toReturn, c.Keyframes[0])
	if len(c.Keyframes) == 1 {
		toReturn = append(toReturn, QuaternionKeyframe{
			Time:       c.Duration,
			Quaternion: c.Keyframes[0].Quaternion,
		})
		return toReturn
	}
	// T[i]   = 0.5 * (P[i+1] - P[i-1])
	// T[i+1] = 0.5 * (P[i+2] - P[i])
	c.HermiteFrames = [4]QuaternionKeyframe{
		c.Keyframes[0], // P[i-1]
		c.Keyframes[0], // P[i]
		c.Keyframes[0], // P[i+1]
		c.Keyframes[0], // P[i+2]
	}
	c.CurrentIndex = 1
	c.ShiftFrame()
	c.ShiftFrame()
	totalCount := int(math.Ceil(float64(framerate) * float64(c.Duration)))
	frameTime := float32(1.0 / float64(framerate))
	for i := 0; i < totalCount; i++ {
		currentTime := float32(i) * frameTime
		for currentTime > c.HermiteFrames[2].Time && c.HermiteFrames[1].Time != c.HermiteFrames[2].Time {
			c.ShiftFrame()
		}
		var s float32
		if c.HermiteFrames[1].Time == c.HermiteFrames[2].Time {
			s = 1.0
		} else {
			s = (currentTime - c.HermiteFrames[1].Time) / (c.HermiteFrames[2].Time - c.HermiteFrames[1].Time)
		}
		p1 := c.HermiteFrames[1].Quaternion
		p2 := c.HermiteFrames[2].Quaternion
		t1 := c.HermiteFrames[2].Quaternion.Sub(c.HermiteFrames[0].Quaternion).Scale(0.5)
		t2 := c.HermiteFrames[3].Quaternion.Sub(c.HermiteFrames[1].Quaternion).Scale(0.5)

		s2 := s * s
		s3 := s * s2

		h1 := 2*s3 - 3*s2 + 1
		h2 := -2*s3 + 3*s2
		h3 := s3 - 2*s2 + s
		h4 := s3 - s2

		p := p1.Scale(h1).Add(p2.Scale(h2)).Add(t1.Scale(h3)).Add(t2.Scale(h4))
		toReturn = append(toReturn, QuaternionKeyframe{
			Time:       currentTime,
			Quaternion: p,
		})
	}
	return toReturn
}
