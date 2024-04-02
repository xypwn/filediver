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
					lum := c[get3DcBits(data[:], &startBit, 3)]

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
