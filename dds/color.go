package dds

func mapBits1To8(x uint16) uint8 {
	return uint8(x * 255)
}

func mapBits2To8(x uint16) uint8 {
	return uint8((x * 340) >> 2)
}

func mapBits3To8(x uint16) uint8 {
	return uint8((x * 292) >> 3)
}

func mapBits4To8(x uint16) uint8 {
	return uint8(x * 17)
}

func mapBits5To8(x uint16) uint8 {
	return uint8((x*527 + 23) >> 6)
}

func mapBits6To8(x uint16) uint8 {
	return uint8((x*259 + 33) >> 6)
}

func mapBits7To8(x uint16) uint8 {
	return uint8((x * 129) >> 6)
}

func colorR5G6B5ToRGB(x uint16) (uint8, uint8, uint8) {
	r5 := (x & 0b1111_1000_0000_0000) >> 11
	g6 := (x & 0b0000_0111_1110_0000) >> 5
	b5 := x & 0b0000_0000_0001_1111
	return mapBits5To8(r5), mapBits6To8(g6), mapBits5To8(b5)
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
