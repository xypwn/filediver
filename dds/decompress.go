package dds

import (
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/exp/constraints"
)

// https://github.com/ImageMagick/ImageMagick/blob/main/coders/dds.c

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

type DecompressFunc func(buf []uint8, r io.Reader, width, height int, info Info) error

func avgU8(xs ...uint8) uint8 {
	var sum int
	for _, x := range xs {
		sum += int(x)
	}
	return uint8(sum / len(xs))
}

func DecompressUncompressed(buf []uint8, r io.Reader, width, height int, info Info) error {
	if info.Alpha {
		// Alpha channel exists

		rBM, gBM, bBM, aBM := info.Header.PixelFormat.RBitMask, info.Header.PixelFormat.GBitMask, info.Header.PixelFormat.BBitMask, info.Header.PixelFormat.ABitMask
		var alphaBits uint8
		if info.Header.PixelFormat.RGBBitCount == 16 {
			if rBM == 0x7c00 && gBM == 0x03e0 && bBM == 0x001f && aBM == 0x8000 {
				alphaBits = 1
			} else if (rBM == 0x00ff && gBM == 0x00ff && bBM == 0x00ff && aBM == 0xff00) ||
				(rBM == 0x00ff && gBM == 0x0000 && bBM == 0x0000 && aBM == 0xff00) {
				alphaBits = 8
			} else if rBM == 0x0f00 && gBM == 0x0f00 && bBM == 0x00f0 && aBM == 0xf000 {
				alphaBits = 4
			} else {
				return fmt.Errorf("unsupported RGBA bit mask: R %08x G %08x B %08x A %08x", rBM, gBM, bBM, aBM)
			}
		}
		if info.DXT10Header != nil && info.DXT10Header.DXGIFormat == DXGIFormatB5G5R5A1UNorm {
			alphaBits = 1
		}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				idx := 4 * (y*width + x)
				if info.Header.PixelFormat.RGBBitCount == 16 ||
					(info.DXT10Header != nil && info.DXT10Header.DXGIFormat == DXGIFormatB5G5R5A1UNorm) {
					var data [2]uint8
					if _, err := io.ReadFull(r, data[:]); err != nil {
						return err
					}
					c := binary.LittleEndian.Uint16(data[:])
					if alphaBits == 1 {
						buf[idx+0], buf[idx+1], buf[idx+2], buf[idx+3] = colorA1R5G5B5ToRGBA(c)
					} else if alphaBits == 8 {
						gy := uint8(c & 0x00FF)
						a := uint8((c & 0xFF00) >> 8)
						buf[idx+0], buf[idx+1], buf[idx+2], buf[idx+3] = gy, gy, gy, a
					} else if alphaBits == 4 {
						buf[idx+0], buf[idx+1], buf[idx+2], buf[idx+3] = colorA4R4G4B4ToRGBA(c)
					} else {
						return fmt.Errorf("invalid alpha bits: %v", alphaBits)
					}
				} else if (info.DXT10Header != nil && info.DXT10Header.DXGIFormat == DXGIFormatR8G8B8A8UNorm) ||
					rBM == 0x0000_00ff && gBM == 0x0000_ff00 && bBM == 0x00ff_0000 && aBM == 0xff00_0000 {
					if _, err := io.ReadFull(r, buf[idx:idx+4]); err != nil {
						return err
					}
				} else if (info.DXT10Header != nil && info.DXT10Header.DXGIFormat == DXGIFormatB8G8R8A8UNorm) ||
					rBM == 0x00ff_0000 && gBM == 0x0000_ff00 && bBM == 0x0000_00ff && aBM == 0xff00_0000 {
					if _, err := io.ReadFull(r, buf[idx:idx+4]); err != nil {
						return err
					}
					buf[idx+0], buf[idx+2] = buf[idx+2], buf[idx+0] // BGRA to RGBA
				} else {
					var dxgi string
					if info.DXT10Header != nil {
						dxgi = fmt.Sprintf("dxgi format=%v, ", info.DXT10Header.DXGIFormat)
					}
					return fmt.Errorf("unsupported RGBA format: %vbitmask: R %08x G %08x B %08x A %08x", dxgi, rBM, gBM, bBM, aBM)
				}
			}
		}
	} else {
		// No alpha channel exists

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				idx := 4 * (y*width + x)
				if info.Header.PixelFormat.RGBBitCount == 8 ||
					(info.DXT10Header != nil && info.DXT10Header.DXGIFormat == DXGIFormatR8UNorm) {
					var c uint8
					if err := binary.Read(r, binary.LittleEndian, &c); err != nil {
						return err
					}
					buf[idx+0] = c
					buf[idx+1] = c
					buf[idx+2] = c
				} else if info.Header.PixelFormat.RGBBitCount == 16 ||
					(info.DXT10Header != nil && info.DXT10Header.DXGIFormat == DXGIFormatB5G6R5UNorm) {
					var c uint16
					if err := binary.Read(r, binary.LittleEndian, &c); err != nil {
						return err
					}
					buf[idx+0], buf[idx+1], buf[idx+2] = colorR5G6B5ToRGB(c)
				} else if info.Header.PixelFormat.RGBBitCount == 24 {
					if _, err := io.ReadFull(r, buf[idx:idx+3]); err != nil {
						return err
					}
					buf[idx+0], buf[idx+2] = buf[idx+2], buf[idx+0] // BGR to RGB
				} else {
					return fmt.Errorf("unsupported RGB bit count: %v", info.Header.PixelFormat.RGBBitCount)
				}
				buf[idx+3] = 255
			}
		}
	}
	return nil
}

func calculateColors(c0 uint16, c1 uint16, ignoreAlpha bool) (r [4]uint8, g [4]uint8, b [4]uint8, a [4]uint8) {
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

func DecompressDXT5(buf []uint8, r io.Reader, width, height int, _ Info) error {
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

			cR, cG, cB, _ := calculateColors(c0, c1, true)

			for j := 0; j < min(height-y, 4); j++ {
				for i := 0; i < min(width-x, 4); i++ {
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

func get3DcBits(data []uint8, startBit *uint64, numBits uint8) uint8 {
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

func Decompress3DcPlus(buf []uint8, r io.Reader, width, height int, _ Info) error {
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
					lum := c[get3DcBits(data[:], &startBit, 3)]

					if x+i >= width || y+j >= height {
						continue
					}

					idx := 4 * ((y+j)*width + (x + i))
					buf[idx+0] = lum
					buf[idx+1] = lum
					buf[idx+2] = lum
					buf[idx+3] = 255
				}
			}
		}
	}
	return nil
}

func Decompress3Dc(buf []uint8, r io.Reader, width, height int, _ Info) error {
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
					r := cR[get3DcBits(data[:], &startBitR, 3)]
					g := cG[get3DcBits(data[:], &startBitG, 3)]

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
