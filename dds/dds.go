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
	Alpha       bool
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
		ColorModel: color.NRGBAModel,
		NumMipMaps: 1,
	}

	cubemap := hdr.Caps2&Caps2Cubemap != 0
	volume := hdr.Caps2&Caps2Volume != 0 && hdr.Depth > 0

	if hdr.PixelFormat.Flags&PixelFormatFlagRGB != 0 ||
		hdr.PixelFormat.Flags&PixelFormatFlagLuminance != 0 {
		info.Alpha = hdr.PixelFormat.Flags&PixelFormatFlagAlphaPixels != 0
		info.Decompress = DecompressUncompressed
	} else if hdr.PixelFormat.Flags&PixelFormatFlagFourCC != 0 {
		switch hdr.PixelFormat.FourCC {
		case [4]byte{'A', 'T', 'I', '2'}:
			info.Alpha = false
			info.Decompress = Decompress3Dc
		case [4]byte{'D', 'X', 'T', '1'}:
			return Info{}, errors.New("DXT1 compression unsupported")
		case [4]byte{'D', 'X', 'T', '3'}:
			return Info{}, errors.New("DXT3 compression unsupported")
		case [4]byte{'D', 'X', 'T', '5'}:
			info.Alpha = true
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
			case DXGIFormatR8UNorm:
				info.Alpha = false
				info.Decompress = DecompressUncompressed
			case DXGIFormatB5G6R5UNorm:
				info.Alpha = false
				info.Decompress = DecompressUncompressed
			case DXGIFormatB5G5R5A1UNorm:
				info.Alpha = true
				info.Decompress = DecompressUncompressed
			case DXGIFormatB8G8R8A8UNorm:
				info.Alpha = true
				info.Decompress = DecompressUncompressed
			case DXGIFormatR8G8B8A8UNorm:
				info.Alpha = true
				info.Decompress = DecompressUncompressed
			case DXGIFormatR10G10B10A2UNorm:
				info.Alpha = true
				info.Decompress = DecompressUncompressed
			case DXGIFormatB8G8R8X8UNorm:
				// X8 isn't really alpha, but we have to other channel to write it to
				info.Alpha = true
				info.Decompress = DecompressUncompressed
			case DXGIFormatBC1UNorm:
				return Info{}, errors.New("DXT1 compression unsupported")
			case DXGIFormatBC2UNorm:
				return Info{}, errors.New("DXT3 compression unsupported")
			case DXGIFormatBC3UNorm:
				info.Alpha = true
				info.Decompress = DecompressDXT5
			case DXGIFormatBC5UNorm:
				info.Alpha = false
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
		stride := 3
		if info.Alpha {
			stride++
		}
		width, height := int(info.Header.Width), int(info.Header.Height)
		mipMaps := make([]*DDSMipMap, 0, mipMapsToRead)
		for j := 0; j < mipMapsToRead; j++ {
			if width == 0 || height == 0 {
				break
			}

			var buf []uint8
			var img image.Image
			if info.ColorModel == color.NRGBAModel {
				newImg := image.NewNRGBA(image.Rect(0, 0, width, height))
				buf = newImg.Pix
				img = newImg
			} else {
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
			Image:   mipMaps[0],
			MipMaps: mipMaps,
		}
	}

	return &DDS{
		Image:  images[0],
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
