package dds

func mapBits4To8(x uint16) uint8 {
	return uint8(x * 17)
}

func mapBits5To8(x uint16) uint8 {
	return uint8((x*527 + 23) >> 6)
}

func mapBits6To8(x uint16) uint8 {
	return uint8((x*259 + 33) >> 6)
}

func colorR5G6B5ToRGB(x uint16) (uint8, uint8, uint8) {
	r5 := (x & 0b1111_1000_0000_0000) >> 11
	g6 := (x & 0b0000_0111_1110_0000) >> 5
	b5 := x & 0b0000_0000_0001_1111
	return mapBits5To8(r5), mapBits6To8(g6), mapBits5To8(b5)
}

func colorA1R5G5B5ToRGBA(x uint16) (uint8, uint8, uint8, uint8) {
	a1 := (x & 0b1000_0000_0000_0000) >> 15
	r5 := (x & 0b0111_1100_0000_0000) >> 10
	g5 := (x & 0b0000_0011_1110_0000) >> 5
	b5 := x & 0b0000_0000_0001_1111
	return mapBits5To8(r5), mapBits5To8(g5), mapBits5To8(b5), uint8(a1 * 255)
}

func colorA4R4G4B4ToRGBA(x uint16) (uint8, uint8, uint8, uint8) {
	a4 := (x & 0b1111_0000_0000_0000) >> 12
	r4 := (x & 0b0000_1111_0000_0000) >> 8
	g4 := (x & 0b0000_0000_1111_0000) >> 4
	b4 := x & 0b0000_0000_0000_1111
	return mapBits4To8(r4), mapBits4To8(g4), mapBits4To8(b4), mapBits4To8(a4)
}

// Program to brute force the mapBitsXTo8 parameters
/*
package main

import (
	"fmt"
	"math"
)

func main() {
	const startBits = 5
	const rsh = 6

	const n = 1 << startBits

	for a := 0; a < 800; a++ {
		for b := 0; b < 128; b++ {
			bad := false
			for i := 0; i < n; i++ {
				expect := int(math.Floor(float64(i)*255/(n-1) + 0.5))
				try := (i*a + b) >> rsh
				if try != expect {
					bad = true
					break
				}
			}
			if !bad {
				fmt.Printf("mapBits%vTo8(x uint16) uint8 {\n", startBits)
				fmt.Printf("	return uint8((x*%v + %v) >> %v)\n", a, b, rsh)
				fmt.Printf("}\n")
			}
		}
	}
}
*/
