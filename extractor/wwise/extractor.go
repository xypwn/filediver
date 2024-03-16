package wwise

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/mewkiz/flac"
	flac_frame "github.com/mewkiz/flac/frame"
	flac_meta "github.com/mewkiz/flac/meta"

	"github.com/xypwn/filediver/extractor"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/util"
	"github.com/xypwn/filediver/wwise"
)

type format int

const (
	formatWav format = iota
	formatFlac
)

func pcmFloat32ToIntS16(dst []int, src []float32) {
	if len(dst) != len(src) {
		panic("dst and src must be the same length")
	}

	for i := 0; i < len(dst); i++ {
		val := int(math.Floor(float64(src[i])*32767 + 0.5))
		if val > 32767 {
			val = 32767
		}
		if val < -32768 {
			val = 32768
		}
		dst[i] = val
	}
}

func pcmFloat32ToInt32S16P(dst []int32, src []float32, chanIdx, numChans int) {
	if len(dst)*numChans != len(src) {
		panic("dst length multiplied by number of channels must equal src length")
	}

	for i := 0; i < len(dst); i++ {
		val := int32(math.Floor(float64(src[chanIdx+i*numChans])*32767 + 0.5))
		if val > 32767 {
			val = 32767
		}
		if val < -32768 {
			val = 32768
		}
		dst[i] = val
	}
}

func getFlacChannels(nchannels int) (flac_frame.Channels, error) {
	switch nchannels {
	case 1:
		return flac_frame.ChannelsMono, nil
	case 2:
		return flac_frame.ChannelsLR, nil
	case 3:
		return flac_frame.ChannelsLRC, nil
	case 4:
		return flac_frame.ChannelsLRLsRs, nil
	case 5:
		return flac_frame.ChannelsLRCLsRs, nil
	case 6:
		return flac_frame.ChannelsLRCLfeLsRs, nil
	case 7:
		return flac_frame.ChannelsLRCLfeCsSlSr, nil
	case 8:
		return flac_frame.ChannelsLRCLfeLsRsSlSr, nil
	default:
		return 0, fmt.Errorf("unsupported number of channels: %d", nchannels)
	}
}

func convertWemStream(out io.WriteSeeker, in io.ReadSeeker, format format) error {
	dec, err := wwise.OpenWem(in)
	if err != nil {
		return err
	}
	switch format {
	case formatWav:
		enc := wav.NewEncoder(out, dec.SampleRate(), 16, dec.Channels(), 1)
		defer enc.Close()
		smpBuf := make([]int, dec.BufferSize())
		for {
			data, err := dec.Decode()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					return err
				}
			}

			pcmFloat32ToIntS16(smpBuf[:len(data)], data)

			if err := enc.Write(&audio.IntBuffer{
				Format: &audio.Format{
					NumChannels: dec.Channels(),
					SampleRate:  dec.SampleRate(),
				},
				Data:           smpBuf[:len(data)],
				SourceBitDepth: 16,
			}); err != nil {
				return err
			}
		}
	case formatFlac:
		// As I found out way too deep into using this FLAC library, it only supports uncompressed FLAC.
		// I still couldn't find a single compressing audio encoder in pure Go. Maybe one day...

		enc, err := flac.NewEncoder(out, &flac_meta.StreamInfo{
			BlockSizeMin:  16,
			BlockSizeMax:  65535,
			SampleRate:    uint32(dec.SampleRate()),
			NChannels:     uint8(dec.Channels()),
			BitsPerSample: 16,
		})
		if err != nil {
			return err
		}
		smpBuf := make([][]int32, dec.Channels())
		subframes := make([]*flac_frame.Subframe, dec.Channels())
		for i := range subframes {
			subframes[i] = &flac_frame.Subframe{}
			smpBuf[i] = make([]int32, dec.BufferSize()/dec.Channels())
		}
		channelLayout, err := getFlacChannels(dec.Channels())
		if err != nil {
			return err
		}

		for {
			data, err := dec.Decode()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					return err
				}
			}
			nSamplesPerCh := len(data) / dec.Channels()

			for i := range subframes {
				pcmFloat32ToInt32S16P(smpBuf[i][:nSamplesPerCh], data, i, dec.Channels())
				subframes[i].SubHeader = flac_frame.SubHeader{
					Pred: flac_frame.PredVerbatim,
				}
				subframes[i].NSamples = nSamplesPerCh
				subframes[i].Samples = smpBuf[i][:nSamplesPerCh]
			}

			if err := enc.WriteFrame(&flac_frame.Frame{
				Header: flac_frame.Header{
					HasFixedBlockSize: false,
					BlockSize:         uint16(nSamplesPerCh),
					SampleRate:        uint32(dec.SampleRate()),
					Channels:          channelLayout,
					BitsPerSample:     16,
				},
				Subframes: subframes,
			}); err != nil {
				return err
			}
		}
		if err := enc.Close(); err != nil {
			return err
		}
	}
	return nil
}

func getFormat(config extractor.Config) (format, string, error) {
	f, ok := config["format"]
	if !ok {
		return formatWav, ".wav", nil
	}
	switch f {
	case "wav":
		return formatWav, ".wav", nil
	case "flac":
		return formatFlac, ".flac", nil
	default:
		return 0, "", fmt.Errorf("invalid audio output format: \"%v\"", f)
	}
}

func ExtractWem(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
	out, err := os.Create(outPath + ".wem")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, ins[stingray.DataStream]); err != nil {
		return err
	}
	return nil
}

func ConvertWem(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
	format, ext, err := getFormat(config)
	if err != nil {
		return err
	}
	out, err := os.Create(outPath + ext)
	if err != nil {
		return err
	}
	defer out.Close()
	if err := convertWemStream(out, ins[stingray.DataStream], format); err != nil {
		return err
	}
	return nil
}

type stingrayBnkHeader struct {
	Unk00 [4]byte
	Size  uint32
	Name  stingray.Hash
}

func extractBnk(ins [stingray.NumDataType]io.ReadSeeker, _ extractor.Config) (io.ReadSeeker, error) {
	in := ins[stingray.DataMain]

	fileSize, err := in.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	var hdr stingrayBnkHeader
	if err := binary.Read(in, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if int64(hdr.Size+0x10) != fileSize {
		return nil, fmt.Errorf("size specified in header (%v) does not match actual file size (%v)", hdr.Size+0x10, fileSize)
	}

	return util.NewSectionReadSeeker(
		in,
		0x10,
		fileSize-0x10,
	)
}

func ExtractBnk(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
	r, err := extractBnk(ins, config)
	if err != nil {
		return err
	}
	out, err := os.Create(outPath + ".bnk")
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func ConvertBnk(outPath string, ins [stingray.NumDataType]io.ReadSeeker, config extractor.Config) error {
	format, ext, err := getFormat(config)
	if err != nil {
		return err
	}

	bnkR, err := extractBnk(ins, config)
	if err != nil {
		return err
	}

	bnk, err := wwise.OpenBnk(bnkR, &wwise.BkhdXorKey{
		/* https://github.com/Xaymar/Hellextractor/issues/25 */
		// "reverse-engineer" the key in code:
		Version: 0x0000008c ^ 0x9211bc20,
		ID:      0x50c63a23 ^ 0xf3d64a1b,
	})
	if err != nil {
		return err
	}

	dirPath := outPath + ".bnk"
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	for i := 0; i < bnk.NumFiles(); i++ {
		wemR, err := bnk.OpenFile(i)
		if err != nil {
			return err
		}
		if err := func() error {
			out, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("%03d%v", i, ext)))
			if err != nil {
				return err
			}
			defer out.Close()
			if err := convertWemStream(out, wemR, format); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	return nil
}
