package dds

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
)

type Info struct {
	Header      Header
	DXT10Header *DXT10Header
	Decompress  DecompressFunc
	ColorModel  color.Model
	NumMipMaps  int
	NumImages   int
}

func DecodeInfo(r io.Reader) (Info, error) {
	hdr, err := DecodeHeader(r)
	if err != nil {
		return Info{}, err
	}

	info := Info{
		Header:     hdr,
		NumMipMaps: 1,
	}

	cubemap := hdr.Caps2&Caps2Cubemap != 0
	volume := hdr.Caps2&Caps2Volume != 0 && hdr.Depth > 0

	if hdr.PixelFormat.Flags&PixelFormatFlagRGB != 0 {
		info.ColorModel = color.NRGBAModel
		info.Decompress = DecompressUncompressed
	} else if hdr.PixelFormat.Flags&PixelFormatFlagYUV != 0 {
		if hdr.PixelFormat.Flags&PixelFormatFlagAlphaPixels == 0 {
			info.ColorModel = color.YCbCrModel
		} else {
			info.ColorModel = color.NYCbCrAModel
		}
		info.Decompress = DecompressUncompressed
	} else if hdr.PixelFormat.Flags&PixelFormatFlagLuminance != 0 {
		if hdr.PixelFormat.Flags&PixelFormatFlagAlphaPixels == 0 {
			if hdr.PixelFormat.GBitMask == 0 && hdr.PixelFormat.BBitMask == 0 {
				if hdr.PixelFormat.RGBBitCount > 8 {
					info.ColorModel = color.Gray16Model
				} else {
					info.ColorModel = color.GrayModel
				}
			} else {
				info.ColorModel = color.NRGBAModel
			}
		} else {
			info.ColorModel = color.NRGBAModel
		}
		info.Decompress = DecompressUncompressed
	} else if hdr.PixelFormat.Flags&PixelFormatFlagFourCC != 0 {
		switch hdr.PixelFormat.FourCC {
		case [4]byte{'A', 'T', 'I', '1'}:
			info.ColorModel = color.GrayModel
			info.Decompress = Decompress3DcPlus
		case [4]byte{'A', 'T', 'I', '2'}:
			info.ColorModel = color.NRGBAModel
			info.Decompress = Decompress3Dc
		case [4]byte{'D', 'X', 'T', '1'}:
			info.ColorModel = color.NRGBAModel
			info.Decompress = DecompressDXT1
		case [4]byte{'D', 'X', 'T', '3'}:
			return Info{}, errors.New("DXT3 compression unsupported")
		case [4]byte{'D', 'X', 'T', '5'}:
			info.ColorModel = color.NRGBAModel
			info.Decompress = DecompressDXT5
		case [4]byte{'D', 'X', '1', '0'}:
			dx10, err := DecodeDXT10Header(r)
			if err != nil {
				return Info{}, err
			}
			info.DXT10Header = &dx10

			if dx10.ResourceDimension != D3D10ResourceDimensionTexture2D {
				return Info{}, errors.New("unsupported DXT10 resource dimension")
			}

			switch dx10.DXGIFormat {
			case DXGIFormatR32G32B32A32Float,
				DXGIFormatR32G32B32Float,
				DXGIFormatR16G16B16A16Float,
				DXGIFormatR16G16B16A16UNorm,
				DXGIFormatR32G32Float:
				info.ColorModel = color.NRGBA64Model
				info.Decompress = DecompressUncompressedDXT10
			case DXGIFormatR32Float, DXGIFormatR16UNorm:
				info.ColorModel = color.Gray16Model
				info.Decompress = DecompressUncompressedDXT10
			case DXGIFormatR8UNorm:
				info.ColorModel = color.GrayModel
				info.Decompress = DecompressUncompressedDXT10
			case DXGIFormatR8G8B8A8UNorm:
				info.ColorModel = color.NRGBAModel
				info.Decompress = DecompressUncompressedDXT10
			case DXGIFormatBC1UNorm:
				info.ColorModel = color.NRGBAModel
				info.Decompress = DecompressDXT1
			case DXGIFormatBC2UNorm:
				return Info{}, errors.New("DXT3 compression unsupported")
			case DXGIFormatBC3UNorm:
				info.ColorModel = color.NRGBAModel
				info.Decompress = DecompressDXT5
			case DXGIFormatBC4UNorm:
				info.ColorModel = color.GrayModel
				info.Decompress = Decompress3DcPlus
			case DXGIFormatBC5UNorm:
				info.ColorModel = color.NRGBAModel
				info.Decompress = Decompress3Dc
			case DXGIFormatBC7UNorm, DXGIFormatBC7UNormSRGB:
				return Info{}, errors.New("BC7 compression unsupported")
			default:
				return Info{}, fmt.Errorf("unsupported DXGI format: %v", dx10.DXGIFormat)
			}

			if dx10.MiscFlag&D3D10ResourceMiscFlagTextureCube != 0 {
				cubemap = true
			}
		default:
			return Info{}, fmt.Errorf("unsupported cmpression format: unknown fourCC: %v", string(hdr.PixelFormat.FourCC[:]))
		}
	}

	info.NumImages = 1
	if cubemap {
		info.NumImages = 0
		if hdr.Caps2&Caps2CubemapPlusX != 0 {
			info.NumImages++
		}
		if hdr.Caps2&Caps2CubemapMinusX != 0 {
			info.NumImages++
		}
		if hdr.Caps2&Caps2CubemapPlusY != 0 {
			info.NumImages++
		}
		if hdr.Caps2&Caps2CubemapMinusY != 0 {
			info.NumImages++
		}
		if hdr.Caps2&Caps2CubemapPlusZ != 0 {
			info.NumImages++
		}
		if hdr.Caps2&Caps2CubemapMinusZ != 0 {
			info.NumImages++
		}
	}

	if volume {
		info.NumImages = int(hdr.Depth)
	}

	if info.NumImages == 0 {
		return Info{}, errors.New("invalid image header: no images")
	}

	if hdr.Caps&CapsMipMap != 0 &&
		(hdr.Caps&CapsTexture != 0 || hdr.Caps2&Caps2Cubemap != 0) {
		info.NumMipMaps = int(hdr.MipMapCount)
	}

	if info.NumMipMaps == 0 {
		return Info{}, errors.New("invalid image header: base image mipmap (mip 0) missing")
	}

	return info, nil
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	info, err := DecodeInfo(r)
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{
		ColorModel: info.ColorModel,
		Width:      int(info.Header.Width),
		Height:     int(info.Header.Height),
	}, nil
}

type DDSMipMap struct {
	image.Image
	Width, Height int
}

type DDSImage struct {
	image.Image
	// Length is guaranteed to be >= 1.
	MipMaps []*DDSMipMap
}

type DDS struct {
	image.Image
	Info Info
	// Length is guaranteed to be >= 1.
	Images []*DDSImage
}

// https://github.com/ImageMagick/ImageMagick/blob/main/coders/dds.c

func Decode(r io.Reader, readMipMaps bool) (*DDS, error) {
	info, err := DecodeInfo(r)
	if err != nil {
		return nil, err
	}

	mipMapsToRead := 1
	if readMipMaps {
		mipMapsToRead = info.NumMipMaps
	}

	images := make([]*DDSImage, info.NumImages)
	for i := 0; i < info.NumImages; i++ {
		width, height := int(info.Header.Width), int(info.Header.Height)
		mipMaps := make([]*DDSMipMap, 0, mipMapsToRead)
		for j := 0; j < mipMapsToRead; j++ {
			if width == 0 || height == 0 {
				break
			}

			var buf []uint8
			var img image.Image
			switch info.ColorModel {
			case color.GrayModel:
				newImg := image.NewGray(image.Rect(0, 0, width, height))
				buf = newImg.Pix
				img = newImg
			case color.Gray16Model:
				newImg := image.NewGray16(image.Rect(0, 0, width, height))
				buf = newImg.Pix
				img = newImg
			case color.NRGBAModel:
				newImg := image.NewNRGBA(image.Rect(0, 0, width, height))
				buf = newImg.Pix
				img = newImg
			case color.NRGBA64Model:
				newImg := image.NewNRGBA64(image.Rect(0, 0, width, height))
				buf = newImg.Pix
				img = newImg
			default:
				return nil, errors.New("invalid color model passed by info structure")
			}
			if err := info.Decompress(buf, r, width, height, info); err != nil {
				return nil, err
			}
			mipMaps = append(mipMaps, &DDSMipMap{
				Image:  img,
				Width:  width,
				Height: height,
			})

			width /= 2
			height /= 2
		}

		if len(mipMaps) == 0 {
			return nil, errors.New("no mipmaps written")
		}

		images[i] = &DDSImage{
			Image:   mipMaps[0].Image,
			MipMaps: mipMaps,
		}
	}

	return &DDS{
		Image:  images[0].Image,
		Info:   info,
		Images: images,
	}, nil
}

func init() {
	image.RegisterFormat(
		"dds",
		"DDS ",
		func(r io.Reader) (image.Image, error) {
			return Decode(r, false)
		},
		DecodeConfig,
	)
}
