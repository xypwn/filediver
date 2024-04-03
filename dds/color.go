package dds

func mapBits1To8(x uint16) uint8 {
	return uint8(x * 255)
}

func mapBits2To8(x uint16) uint8 {
	return uint8(x<<6 | x<<4 | x<<2 | x)
}

func mapBits3To8(x uint16) uint8 {
	return uint8(x<<5 | x<<2 | x>>1)
}

func mapBits4To8(x uint16) uint8 {
	return uint8(x<<4 | x)
}

func mapBits5To8(x uint16) uint8 {
	return uint8(x<<3 | x>>2)
}

func mapBits6To8(x uint16) uint8 {
	return uint8(x<<2 | x>>4)
}

func mapBits7To8(x uint16) uint8 {
	return uint8(x<<1 | x>>6)
}

func colorR5G6B5ToRGB(x uint16) (uint8, uint8, uint8) {
	r5 := (x & 0b1111_1000_0000_0000) >> 11
	g6 := (x & 0b0000_0111_1110_0000) >> 5
	b5 := x & 0b0000_0000_0001_1111
	return mapBits5To8(r5), mapBits6To8(g6), mapBits5To8(b5)
}
